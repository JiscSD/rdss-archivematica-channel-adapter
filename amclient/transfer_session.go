package amclient

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
)

const (
	// standardTransferType is the value we need to pass to the Approve API
	// to indicate that we're creating a standard transfer.
	standardTransferType = "standard"

	// standardTransferDir is the location of the directory where transfers need
	// to be moved. Relative to the shared directory (see amSharedFs).
	standardTransferDir = "watchedDirectories/activeTransfers/standardTransfer"

	// Relative to the shared directory (see amSharedFs).
	tmpDir = "tmp"
)

// TransferSession lets you prepare a new transfer and submit it to
// Archivematica. It is a convenience tool around the transfer service.
type TransferSession struct {
	// Archivematica HTTP client.
	c *Client

	// Transfer's filesystem.
	fs *afero.Afero

	// Archivematica Shared Directory filesystem.
	amSharedFs *afero.Afero

	// Original name given to the transfer. The `move` method may change the
	// final name if the desired name is already taken in `standardTransferDir`.
	originalName string

	// ID of the transfer associated to this session. Its value is populated as
	// soon as it is communicated to us by Archivematica which typically happens
	// when we receive the repsonse to the transfer approval request - see
	// TransferApproveResponse for more details.
	id string

	Metadata        *MetadataSet
	ChecksumsMD5    *ChecksumSet
	ChecksumsSHA1   *ChecksumSet
	ChecksumsSHA256 *ChecksumSet
}

// tempDir creates a new temporary directory in the Acchivematica Shared Directory.
func tempDir(amSharedFs afero.Fs) (string, error) {
	return afero.TempDir(amSharedFs, tmpDir, "amclientTransfer")
}

// NewTransferSession returns a pointer to a new TransferSession.
func NewTransferSession(c *Client, name string, amSharedFs afero.Fs) (*TransferSession, error) {
	tmpDir, err := tempDir(amSharedFs)
	if err != nil {
		return nil, errors.Wrap(err, "temporary folder cannot be created")
	}
	ts := &TransferSession{
		c:            c,
		fs:           &afero.Afero{Fs: afero.NewBasePathFs(amSharedFs, tmpDir)},
		amSharedFs:   &afero.Afero{Fs: amSharedFs},
		originalName: name,
	}
	ts.Metadata = NewMetadataSet(ts.fs)
	ts.ChecksumsMD5 = NewChecksumSet("md5", ts.fs)
	ts.ChecksumsSHA1 = NewChecksumSet("sha1", ts.fs)
	ts.ChecksumsSHA256 = NewChecksumSet("sha256", ts.fs)
	return ts, nil
}

// TransferSession returns a new transfer session.
func (c *Client) TransferSession(name string, depositFs afero.Fs) (*TransferSession, error) {
	return NewTransferSession(c, name, depositFs)
}

// path returns the path of the transfer directory relative to amSharedFs.
func (s *TransferSession) path() string {
	name, err := filepath.Rel(
		afero.FullBaseFsPath(s.amSharedFs.Fs.(*afero.BasePathFs), "/"),
		s.fullPath(),
	)
	if err != nil {
		return ""
	}
	return name
}

// fullPath returns the absolute path of the transfer directory.
func (s *TransferSession) fullPath() string {
	return afero.FullBaseFsPath(s.fs.Fs.(*afero.BasePathFs), "/")
}

// maxRenameAttempts is the no. of attempts that move() performs before giving up.
var maxRenameAttempts = 20

// move moves the transfer directory to `standardTransfer`. If the target
// location already exists we'll attempt to add a numeric suffix corresponding
// to the number of attempts, e.g.: Test, Test-1, Test-2... up to `maxtTries`
// attempts. An error is returned when the maximum number of attempts is
// reached.
func (s *TransferSession) move() error {
	var (
		counter = 1
		name    = filepath.Join(standardTransferDir, safeFileName(s.originalName))
		nName   = name
	)
	for {
		err := s.amSharedFs.Rename(s.path(), nName)
		if err == nil {
			// Succeeded. Update s.fs before we return.
			s.fs = &afero.Afero{
				Fs: afero.NewBasePathFs(afero.NewOsFs(), filepath.Join(
					afero.FullBaseFsPath(s.amSharedFs.Fs.(*afero.BasePathFs), "/"),
					nName,
				)),
			}
			break
		}
		nName = fmt.Sprintf("%s-%d", name, counter)
		counter++
		if counter >= maxRenameAttempts {
			return fmt.Errorf("max. number of attempts to create the standard transfer directory reached (%d) - last error: %v", maxRenameAttempts, err)
		}
	}
	return nil
}

// Create creates a file in the filesystem, returning the file and an error, if
// any happens.
func (s *TransferSession) Create(name string) (afero.File, error) {
	err := s.fs.MkdirAll(filepath.Dir(name), os.FileMode(0755))
	if err != nil {
		return nil, err
	}
	f, err := s.fs.Create(name)
	if err != nil {
		return nil, err
	}
	return f, nil
}

// ProcessingConfig includes a processing configuration (processingMCP.xml)
// given its name.
func (s *TransferSession) ProcessingConfig(name string) error {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	config, _, err := s.c.ProcessingConfig.Get(ctx, name)
	if err != nil {
		return err
	}

	return s.fs.SafeWriteReader("/processingMCP.xml", config)
}

// Start the transfer in Archivematica. It returns the ID of the transfer.
func (s *TransferSession) Start() (string, error) {
	ctx := context.Background()

	if err := s.createMetadataDir(); err != nil {
		return "", err
	}

	if err := s.Metadata.Write(); err != nil {
		return "", err
	}

	if err := s.createChecksumsFiles(); err != nil {
		return "", err
	}

	// Move to standard transfer.
	if err := s.move(); err != nil {
		return "", errors.Wrap(err, "transfer could not be moved to standard transfer")
	}

	return s.id, s.approve(ctx)
}

// approve requests Archivematica to approve a transfer that has been already
// moved to the corresponding watched directory. It first ensures that the
// transfer is listed as unapproved and waits until it does. If positive, it
// attempts to approve it. It uses retries and timeouts. It does not give up if
// the server returns an error.
func (s *TransferSession) approve(ctx context.Context) error {
	var (
		req              = &TransferUnapprovedRequest{}
		transferPath     = s.path()
		transferBasePath = path.Base(transferPath)
	)
	const (
		maxElapsedTime = 10 * time.Minute
		maxInterval    = 30 * time.Second
	)
	op := func() error {
		payload, _, err := s.c.Transfer.Unapproved(ctx, req)
		if err != nil {
			// This is bad because AM isn't able to list unapproved requests.
			// We're going to keep trying.
			return errors.Wrap(err, "unapproved request failed")
		}
		// Check if the transfer is listed as unapproved.
		for _, item := range payload.Results {
			if item.Directory == transferBasePath {
				resp, _, err := s.c.Transfer.Approve(ctx, &TransferApproveRequest{
					Directory: transferPath,
					Type:      standardTransferType,
				})
				if err != nil {
					// We've tried but it failed. Retry.
					return errors.Wrap(err, "approve request failed")
				}
				s.id = resp.UUID
				return nil
			}
		}
		// We're going to keep trying.
		return errors.New("transfer is not listed yet - keep trying")
	}

	bo := backoff.NewExponentialBackOff()
	bo.MaxElapsedTime = maxElapsedTime
	bo.MaxInterval = maxInterval

	return backoff.Retry(op, bo)
}

// Contents returns a list with all the files currently available in the
// temporary transfer filesystem.
func (s *TransferSession) Contents() []string {
	var paths []string
	s.fs.Walk("", func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		if err != nil {
			return err
		}
		paths = append(paths, path)
		return nil
	})
	return paths
}

// Destroy removes the transfer directory and its contents. The caller should
// not expect TransferSession to be in a usable state once this method has been
// called.
func (s *TransferSession) Destroy() error {
	return s.amSharedFs.RemoveAll(s.path())
}

// DescribeFile registers metadata of a file. It causes the transfer to include
// a `metadata.json` file with the metadata of each file described.
func (s *TransferSession) DescribeFile(name, field, value string) {
	s.Metadata.Add(name, field, value)
}

// Describe registers metadata of the whole dataset/transfer. It causes the
// transfer to include a `metadata.json` file with the metadata included.
func (s *TransferSession) Describe(field, value string) {
	s.Metadata.Add("objects/", field, value)
}

// ChecksumMD5 registers a MD5 checksum for a file.
func (s *TransferSession) ChecksumMD5(name, sum string) {
	s.ChecksumsMD5.Add(name, sum)
}

// ChecksumSHA1 registers a SHA1 checksum for a file.
func (s *TransferSession) ChecksumSHA1(name, sum string) {
	s.ChecksumsSHA1.Add(name, sum)
}

// ChecksumSHA256 registers a SHA256 checksum for a file.
func (s *TransferSession) ChecksumSHA256(name, sum string) {
	s.ChecksumsSHA256.Add(name, sum)
}

func (s *TransferSession) createMetadataDir() error {
	const path = "/metadata"
	if _, err := s.fs.Stat(path); err != nil {
		return s.fs.Mkdir(path, os.FileMode(0755))
	}
	return nil
}

func (s *TransferSession) createChecksumsFiles() error {
	if err := s.ChecksumsMD5.Write(); err != nil {
		return err
	}
	if err := s.ChecksumsSHA1.Write(); err != nil {
		return err
	}
	if err := s.ChecksumsSHA256.Write(); err != nil {
		return err
	}
	return nil
}

// MetadataSet holds the metadata entries of the transfer.
type MetadataSet struct {
	entries map[string][][2]string
	fs      afero.Fs
}

// NewMetadataSet returns a new MetadataSet.
func NewMetadataSet(fs afero.Fs) *MetadataSet {
	return &MetadataSet{
		entries: make(map[string][][2]string),
		fs:      fs,
	}
}

// Entries returns all entries that were created.
func (m *MetadataSet) Entries() map[string][][2]string {
	// MetadataSet doesn't have a mutex yet but once it's used, it should be
	// locked right here.

	// Make a copy so the returned value won't race with future log requests.
	entries := make(map[string][][2]string)
	for k, v := range m.entries {
		entries[k] = v
	}
	return entries
}

func (m *MetadataSet) Add(name, field, value string) {
	m.entries[name] = append(m.entries[name], [2]string{field, value})
}

func (m *MetadataSet) Write() error {
	const (
		path = "/metadata/metadata.csv"
		sep  = ','
	)
	if len(m.entries) == 0 {
		return nil
	}
	f, err := m.fs.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	writer := csv.NewWriter(f)
	writer.Comma = sep
	writer.UseCRLF = false
	defer writer.Flush()

	// Build a list of fields with max. total of occurrences found
	occurrences := map[string]int{}
	for _, entry := range m.entries {
		for _, pair := range entry { // Pair ("dc.title", "title 1")
			var o int
			for _, p := range entry {
				if pair[0] == p[0] {
					o++
				}
			}
			if c, ok := occurrences[pair[0]]; !ok || (ok && o > c) {
				occurrences[pair[0]] = o
			}
		}
	}

	// Build a list of fields
	fields := []string{}
	for field, o := range occurrences {
		for i := 0; i < o; i++ {
			fields = append(fields, field)
		}
	}
	sort.Strings(fields)

	// Write header row in CSV.
	writer.Write(append([]string{"filename"}, fields...))

	// Create an slice of filenames sorted alphabetically. We're going to use it
	// so we can iterate over the files in order to generate CSV output in a
	// predicable way.
	names := make([]string, len(m.entries))
	for filename := range m.entries {
		names = append(names, filename)
	}
	sort.Strings(names)

	for _, filename := range names {
		entry, ok := m.entries[filename]
		if !ok {
			continue
		}
		var (
			values  = []string{filename}
			cursors = make(map[string]int)
		)
		// For each known field we either populate a value or an empty string.
		for _, field := range fields {
			var (
				value  string
				subset = entry
				offset = 0
			)
			if pos, ok := cursors[field]; ok {
				pos++
				subset = entry[pos:] // Continue at the next value
				offset = pos
			}
			for index, pair := range subset {
				if pair[0] == field {
					value = pair[1]                 // We have a match
					cursors[field] = index + offset // Memorize position
					break
				}
			}
			values = append(values, value)
		}
		if len(values) > 1 {
			writer.Write(values)
		}
	}

	return nil
}

// ChecksumSet holds the checksums of the files for a sum algorithm.
type ChecksumSet struct {
	sumType string
	values  map[string]string
	fs      afero.Fs
}

func NewChecksumSet(sumType string, fs afero.Fs) *ChecksumSet {
	return &ChecksumSet{
		sumType: sumType,
		values:  make(map[string]string),
		fs:      fs,
	}
}

func (c *ChecksumSet) Add(name, sum string) {
	c.values[name] = sum
}

func (c *ChecksumSet) Write() error {
	const (
		path = "/metadata/checksum.%s"
		sep  = ' '
	)
	if len(c.values) == 0 {
		return nil
	}
	f, err := c.fs.Create(fmt.Sprintf(path, c.sumType))
	if err != nil {
		return err
	}
	defer f.Close()
	writer := csv.NewWriter(f)
	writer.Comma = sep
	writer.UseCRLF = false
	defer writer.Flush()

	for name, sum := range c.values {
		if err := writer.Write([]string{sum, name}); err != nil {
			return err
		}
	}

	return nil
}

var (
	regexSeparators = regexp.MustCompile(`[ &_=+:]`)
	regexLegal      = regexp.MustCompile(`[^[:alnum:]-.]`)
)

// safeFileName returns safe string that can be used in file names
func safeFileName(str string) string {
	name := strings.Replace(str, "/", "-", -1)
	name = strings.Trim(name, " ")
	name = regexSeparators.ReplaceAllString(name, "-")
	name = regexLegal.ReplaceAllString(name, "")
	for strings.Contains(name, "--") {
		name = strings.Replace(name, "--", "-", -1)
	}
	return name
}
