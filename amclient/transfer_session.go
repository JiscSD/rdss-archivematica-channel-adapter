package amclient

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/afero"
)

// TransferSession is a convenience tool to make it easier to create and submit
// transfers.
//
// A transfer session can be created using `NewTransferSession`. Before you
// create one, make sure that `TransferDir` is invoked indicating the absolute
// path of the transfer source location.
type TransferSession struct {
	// Archivematica HTTP client.
	c *Client

	// Transfer's filesystem. This is based off a temporary directory that it's
	// automatically created inside `c.fs`.
	fs *afero.Afero

	// Name of the transfer.
	name string

	Metadata        *MetadataSet
	ChecksumsMD5    *ChecksumSet
	ChecksumsSHA1   *ChecksumSet
	ChecksumsSHA256 *ChecksumSet
}

// tmpfs creates a new temporary directory on the given filesystem and returns
// it as a new filesystem.
func tmpfs(fs afero.Fs) (*afero.Afero, error) {
	if fs == nil {
		return nil, errors.New("filesystem is nil")
	}

	// Try to guess the full path.
	basePath := "unknown"
	basePathFs, ok := fs.(*afero.BasePathFs)
	if ok {
		basePath = afero.FullBaseFsPath(basePathFs, "")
	}

	ok, err := afero.DirExists(fs, "/")
	if err != nil {
		return nil, fmt.Errorf("filesystem is not accesible: %v (%s)", err, basePath)
	}
	if !ok {
		return nil, fmt.Errorf("filesystem is not accesible (%s)", basePath)
	}

	const prefix = "amclientTransfer"
	tmpdir, err := afero.TempDir(fs, "/", prefix)
	if err != nil {
		return nil, errors.Wrap(err, "error creating temporary directory")
	}

	return &afero.Afero{Fs: afero.NewBasePathFs(fs, tmpdir)}, nil
}

// NewTransferSession returns a pointer to a new TransferSession.
func NewTransferSession(c *Client, name string) (*TransferSession, error) {
	if c.fs == nil {
		return nil, errors.New("transfer session not created because the base filesystem is not assigned")
	}
	fs, err := tmpfs(c.fs)
	if err != nil {
		return nil, errors.Wrap(err, "transfer session cannot initialize temporary directory")
	}
	ts := &TransferSession{c: c, fs: fs, name: name}
	ts.Metadata = NewMetadataSet(ts.fs)
	ts.ChecksumsMD5 = NewChecksumSet("md5", ts.fs)
	ts.ChecksumsSHA1 = NewChecksumSet("sha1", ts.fs)
	ts.ChecksumsSHA256 = NewChecksumSet("sha256", ts.fs)
	return ts, nil
}

// fullPath returns the absolute path of the transfer directory.
func (s *TransferSession) fullPath() string {
	return afero.FullBaseFsPath(s.fs.Fs.(*afero.BasePathFs), "")
}

func (s *TransferSession) path() string {
	rel, err := filepath.Rel(
		afero.FullBaseFsPath(s.c.fs.(*afero.BasePathFs), "/"),
		s.fullPath(),
	)
	if err != nil {
		return ""
	}
	return rel
}

// Create returns a new file created in the transfer directory.
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

// ProcessingConfig includes a processing configuration in the transfer.
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

// Start the transfer using the Package API endpoint. This API is still in beta.
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

	req := &PackageCreateRequest{Name: s.name, Path: s.path()}
	payload, _, err := s.c.Package.Create(ctx, req)
	if err != nil {
		return "", err
	}

	return payload.ID, nil
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
	return os.RemoveAll(s.fullPath())
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
		name = "objects/" + name // See https://github.com/artefactual/archivematica/pull/1333.
		if err := writer.Write([]string{sum, name}); err != nil {
			return err
		}
	}

	return nil
}
