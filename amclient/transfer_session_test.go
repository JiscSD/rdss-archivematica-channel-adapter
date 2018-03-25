package amclient

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/url"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/spf13/afero"
)

// tempAmSharedFs is a helper used to create temporary filesystems used as the
// Archivematica Shared Directory.
func tempAmSharedFs(t *testing.T) afero.Fs {
	dir, err := ioutil.TempDir("", "amShareDirTest")
	if err != nil {
		t.Fatalf("cannot create temporary directory: %s", err)
	}

	fs := afero.NewBasePathFs(afero.NewOsFs(), dir)

	// Typically tmpDir and standardTransferDir are pre-created so we're doing
	// the same in our tests.
	if err := fs.MkdirAll(tmpDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := fs.MkdirAll(standardTransferDir, 0755); err != nil {
		t.Fatal(err)
	}

	return fs
}

// newTransferSession is a helper used to created TransferSession objects.
func newTransferSession(t *testing.T, name string) *TransferSession {
	if name == "" {
		name = "MyTransfer"
	}
	c := getClient(t)
	amSharedFs := tempAmSharedFs(t)
	ts, err := NewTransferSession(c, name, amSharedFs)
	if err != nil {
		t.Fatalf("NewTransferSession() returned a non-nil error: %v", err)
	}
	return ts
}

// getClient is a helper used to build a client with a mocked TransferService.
func getClient(t *testing.T) *Client {
	url, _ := url.Parse("http://localhost")
	c := NewClient(nil, url.String(), "", "")
	c.Transfer = &transferServiceMock{t: t}
	return c
}

type transferServiceMock struct {
	t          *testing.T
	approveReq *TransferApproveRequest
}

func (tsm *transferServiceMock) Start(ctx context.Context, req *TransferStartRequest) (*TransferStartResponse, *Response, error) {
	return nil, nil, nil
}

func (tsm *transferServiceMock) Approve(ctx context.Context, req *TransferApproveRequest) (*TransferApproveResponse, *Response, error) {
	tsm.approveReq = req
	return &TransferApproveResponse{
		UUID: "13d9e74c-f88a-44c3-9657-c756eb3fa1c8",
	}, &Response{}, nil
}

func (tsm *transferServiceMock) Unapproved(ctx context.Context, req *TransferUnapprovedRequest) (*TransferUnapprovedResponse, *Response, error) {
	return &TransferUnapprovedResponse{
		Message: "Fetched unapproved transfers successfully.",
		Results: []*TransferUnapprovedResponseResult{
			&TransferUnapprovedResponseResult{Directory: "Test"},
		},
	}, &Response{}, nil
}

func Test_tempDir(t *testing.T) {
	fs := tempAmSharedFs(t)

	name, err := tempDir(fs)
	if err != nil {
		// No reason to fail.
		t.Fatalf("Unexpected non-nil error returned: %v", err)
	}

	stat, err := fs.Stat(name)
	if err != nil {
		t.Fatalf("fs.Stat(name) returned an error: %v", err)
	}
	if !stat.IsDir() {
		t.Fatal("Directory not created")
	}
	if !strings.HasPrefix(name, "tmp/amclient") {
		t.Fatalf("Unexpected path returned: %s", name)
	}
}

func TestNewTransferSession(t *testing.T) {
	var (
		name = "MyTransfer"
		ts   = newTransferSession(t, name)
	)

	if have, want := ts.originalName, name; have != want {
		t.Fatalf("NewTransferSession() unexpected originalName; have %s, want %s", have, want)
	}
	if ts.Metadata == nil || ts.ChecksumsMD5 == nil || ts.ChecksumsSHA1 == nil || ts.ChecksumsSHA256 == nil {
		t.Fatal("NewTransferSession() returned a TransferSession not initialized propery")
	}
	if exists, err := ts.fs.Exists("/"); err != nil || !exists {
		t.Fatal("NewTransferSession() has an unexpected filesystem")
	}
	if contents, err := afero.ReadDir(ts.fs, "/"); err != nil || len(contents) > 0 {
		t.Fatal("NewTransferSession() has an unexpected filesystem")
	}
}

func TestTransferSession_path(t *testing.T) {
	var regex = regexp.MustCompile(`^tmp/amclientTransfer\d+$`)
	ts := newTransferSession(t, "")
	name := ts.path()
	if !regex.MatchString(name) {
		t.Fatalf("TransferSession.path() returned an unexpected string: %s", name)
	}
}

func TestTransferSession_move(t *testing.T) {
	name := "MyTransfer"
	ts := newTransferSession(t, name)

	// Add a file to the transfer. We're going to check if it's moved.
	var (
		fileName = "foobar.jpg"
		fileBlob = []byte{1, 2, 3, 4}
	)
	f, err := ts.Create(fileName)
	if err != nil {
		t.Fatal(err)
	}
	f.Write(fileBlob)
	f.Close()

	if err := ts.move(); err != nil {
		t.Fatalf("TransferSession.move() returned a non-nil error: %v", err)
	}
	have, err := ts.fs.ReadFile(fileName)
	if err != nil || !bytes.Equal(fileBlob, have) {
		t.Fatalf("TransferSession.move() did not move the contents properly")
	}
}

func TestTransferSession_move_WithRetry(t *testing.T) {
	name := "MyTransfer"
	ts := newTransferSession(t, name)

	// Cause conflict by creating "MyTransfer" and "MyTransfer-1".
	if err := ts.amSharedFs.MkdirAll(filepath.Join(standardTransferDir, name), 0755); err != nil {
		t.Fatal(err)
	}
	if err := ts.amSharedFs.MkdirAll(filepath.Join(standardTransferDir, name)+"-1", 0755); err != nil {
		t.Fatal(err)
	}

	if err := ts.move(); err != nil {
		t.Fatalf("TransferSession.move() returned a non-nil error: %v", err)
	}

	// I'm expecting to see "MyTransfer-2".
	if have, want := ts.path(), filepath.Join(standardTransferDir, name)+"-2"; have != want {
		t.Fatalf("TransferSession.move() did not rename as expected; have %s, want %s", have, want)
	}
}

func TestTransferSession_move_WithRetryMaxReached(t *testing.T) {
	name := "MyTransfer"
	ts := newTransferSession(t, name)

	// Cause conflict by creating "MyTransfer".
	if err := ts.amSharedFs.MkdirAll(filepath.Join(standardTransferDir, name), 0755); err != nil {
		t.Fatal(err)
	}
	if err := ts.amSharedFs.MkdirAll(filepath.Join(standardTransferDir, name)+"-1", 0755); err != nil {
		t.Fatal(err)
	}

	// Override maxRenameAttempts. We would need two attempts but have only one!
	maxRenameAttempts = 1
	if err := ts.move(); err == nil {
		t.Fatalf("TransferSession.move() as expected to fail but it didn't")
	}
}

func TestTransferSession_Create(t *testing.T) {
	ts := newTransferSession(t, "")
	tsPath := ts.fullPath()

	tests := []struct {
		path string
		want string
	}{
		{"foobar.jpg", filepath.Join(tsPath, "foobar.jpg")},
		{"foo/bar.jpg", filepath.Join(tsPath, "foo/bar.jpg")},
		{"/foo/bar.jpg", filepath.Join(tsPath, "foo/bar.jpg")},
		{"/f/o/o/b/a/r.jpg", filepath.Join(tsPath, "f/o/o/b/a/r.jpg")},
	}
	for _, tt := range tests {
		// Create the file and write some bytes.
		f, err := ts.Create(tt.path)
		if err != nil {
			t.Fatal(err)
		}
		f.Write([]byte{1, 2, 3, 4})
		defer f.Close()

		// Check that the file exists!
		found, err := ts.fs.Exists(tt.path)
		if err != nil || !found {
			t.Fatalf("file check failed: err=%s found=%t", err, found)
		}
	}
}

func TestTransferSession_Contents(t *testing.T) {
	ts := newTransferSession(t, "")
	afero.TempFile(ts.fs, "/", "uno")
	afero.TempFile(ts.fs, "/", "dos")

	if want, have := 2, len(ts.Contents()); want != have {
		t.Fatalf("Created %d files, only %d found", want, have)
	}
}

func TestTransferSession_Destroy(t *testing.T) {
	var (
		ts   = newTransferSession(t, "")
		path = ts.path()
	)
	afero.TempFile(ts.fs, "/", "uno")
	afero.TempFile(ts.fs, "/", "dos")

	if err := ts.Destroy(); err != nil {
		t.Fatalf("TransferSession.Destroy() returned an error: %v", err)
	}
	if found, err := ts.amSharedFs.DirExists(path); err != nil || found {
		t.Fatalf("TransferSession.Destroy() didn't have the expected results; err=%v, found=%t", err, found)
	}
}

func TestTransferSession_createMetadataDir(t *testing.T) {
	ts := newTransferSession(t, "")

	if err := ts.createMetadataDir(); err != nil {
		t.Fatalf("createMetadataDir failed: %v", err)
	}

	info, err := ts.fs.Stat("metadata")
	if info == nil {
		t.Fatal("Metadata directory was not created")
	}
	if err != nil {
		t.Fatalf("Metadata directory was not created: %v", err)
	}
	if !info.IsDir() {
		t.Fatalf("/metadata is not a directory")
	}
}

func TestTransferSession_Start(t *testing.T) {
	ts := newTransferSession(t, "Test")

	id, err := ts.Start()
	if err != nil {
		t.Fatalf("TransferSession.Start() failed: %v", err)
	}
	if have, want := id, ts.id; have != want {
		t.Errorf("Have %s, want %s", have, want)
	}

	transferService := ts.c.Transfer.(*transferServiceMock)
	if have, want := transferService.approveReq.Directory, ts.path(); have != want {
		t.Errorf("Have %s, want %s", have, want)
	}
	if transferService.approveReq.Type != standardTransferType {
		t.Errorf("Have %s, want %s", transferService.approveReq.Type, standardTransferType)
	}
}

func TestTransferSession_MetadataSet(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewBasePathFs(afero.NewMemMapFs(), "/")}
	set := NewMetadataSet(fs)
	set.Add("objects/", "dc.contributor", "Zhang, Shiyu")
	set.Add("objects/", "dc.contributor", "Whittow, William")
	set.Add("objects/", "dc.contributor", "Seager, Rob")
	set.Add("objects/", "dc.contributor", "Chauraya, Alford")
	set.Add("objects/", "dc.contributor", "Vardaxoglou, Yiannis")
	set.Add("objects/", "dc.identifier", "10.17028/rd.lboro.4665448.v1")
	set.Add("objects/", "dc.publisher", "Loughborough University")
	set.Add("objects/", "dc.title", "Non-uniform Mesh for Embroidered Microstrip Antennas - Simulation files")
	set.Add("objects/", "dc.type", "Dataset")
	set.Add("objects/", "dcterms.issued", "2017-03-17")
	set.Add("objects/woodpigeon-pic.jpg", "dc.identifier", "1")
	set.Add("objects/woodpigeon-pic.jpg", "dc.tile", "woodpigeon-pic.jpg")
	set.Add("objects/bird-sounds.mp3", "dc.identifier", "2")
	set.Add("objects/bird-sounds.mp3", "dc.title", "bird-sounds.mp3")

	var (
		want = `filename,dc.contributor,dc.contributor,dc.contributor,dc.contributor,dc.contributor,dc.identifier,dc.publisher,dc.tile,dc.title,dc.type,dcterms.issued
objects/,"Zhang, Shiyu","Whittow, William","Seager, Rob","Chauraya, Alford","Vardaxoglou, Yiannis",10.17028/rd.lboro.4665448.v1,Loughborough University,,Non-uniform Mesh for Embroidered Microstrip Antennas - Simulation files,Dataset,2017-03-17
objects/bird-sounds.mp3,,,,,,2,,,bird-sounds.mp3,,
objects/woodpigeon-pic.jpg,,,,,,1,,woodpigeon-pic.jpg,,,
`
	)

	if err := set.Write(); err != nil {
		t.Fatal(err)
	}
	c, err := fs.ReadFile("/metadata/metadata.csv")
	if err != nil {
		t.Fatal(err)
	}
	have := string(c)
	if want != have {
		t.Fatalf("Unexpected content:\nhave:\n%s\nwant:\n%s", have, want)
	}
}

func TestTransferSession_ChecksumSet(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewBasePathFs(afero.NewMemMapFs(), "/")}
	set := NewChecksumSet("md5", fs)
	set.Add("bird-sounds.mp3", "92c8ab01cecceb3bf0789c2cd8c7415a")
	set.Add("woodpigeon-pic.jpg", "53a64110e067b14394c142c09571bea0")

	var (
		want1 = `92c8ab01cecceb3bf0789c2cd8c7415a bird-sounds.mp3
53a64110e067b14394c142c09571bea0 woodpigeon-pic.jpg
`
		want2 = `53a64110e067b14394c142c09571bea0 woodpigeon-pic.jpg
92c8ab01cecceb3bf0789c2cd8c7415a bird-sounds.mp3
`
	)

	if err := set.Write(); err != nil {
		t.Fatal(err)
	}
	c, err := fs.ReadFile("/metadata/checksum.md5")
	if err != nil {
		t.Fatal(err)
	}
	have := string(c)
	if want1 != have && want2 != have {
		t.Fatalf("Unexpected content:\nhave:\n%s\nwant1:\n%s\nwant2:\n%s", have, want1, want2)
	}
}

func TestMetadataSet_Entries(t *testing.T) {
	entries := map[string][][2]string{
		"key": [][2]string{
			[2]string{"dc.title", "Título"},
			[2]string{"dc.identifier", "12345"},
		},
	}
	ms := MetadataSet{entries: entries}
	got := ms.Entries()
	if !reflect.DeepEqual(entries, got) {
		t.Error("Entries() method did not return the expected copy of the entries")
	}

	// Mutate internal data structure
	delete(ms.entries, "key")
	if reflect.DeepEqual(ms.Entries(), got) {
		t.Error("Entries() should return a copy of the internal map but it didn't")
	}
}
