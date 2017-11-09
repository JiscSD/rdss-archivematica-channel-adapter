package amclient

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/spf13/afero"
)

func TestNewTransferSession(t *testing.T) {
	c := getClient(t)
	fs := afero.NewBasePathFs(afero.NewMemMapFs(), "/")

	tests := []struct {
		name string
		want string
	}{
		{"Test1", "/Test1"},
		{"_Test2", "/-Test2"},
		{"Test/With/Slashes////", "/Test-With-Slashes-"},
		{"new work-added with framework/α-alumina   ", "/new-work-added-with-framework-alumina"},
		{"Test:foobar", "/Test-foobar"},
	}
	for _, tc := range tests {
		sess, err := NewTransferSession(c, tc.name, fs)
		if err != nil {
			t.Error(err)
		}
		if sess.Path != tc.want {
			t.Errorf("Want %s, have %s", tc.want, sess.Path)
		}
	}
}

func TestNewTransferSession_FolderAlreadyExists(t *testing.T) {
	c := getClient(t)
	fs := afero.NewBasePathFs(afero.NewMemMapFs(), "/")
	fs.Mkdir("/Test", os.FileMode(755))
	fs.Mkdir("/Test-1", os.FileMode(755))
	sess, err := NewTransferSession(c, "Test", fs)
	if err != nil {
		t.Fatal(err)
	}

	want := "/Test-2"
	if sess.Path != want {
		t.Fatalf("Want %s, have %s", want, sess.Path)
	}
}

func TestNewTransferSession_MaxAttemptsError(t *testing.T) {
	c := getClient(t)
	fs := afero.NewBasePathFs(afero.NewMemMapFs(), "/")
	fs.Mkdir("/Test", os.FileMode(755))
	for i := 1; i <= 20; i++ {
		fs.Mkdir(fmt.Sprintf("/Test-%d", i), os.FileMode(755))
	}
	sess, err := NewTransferSession(c, "Test", fs)
	if sess != nil {
		t.Fatal("TransferSession returned should be nil")
	}
	if err == nil {
		t.Fatal("An error was expected!")
	}
}

func TestTransferSession_Create(t *testing.T) {
	c := getClient(t)
	fs := afero.Afero{Fs: afero.NewBasePathFs(afero.NewMemMapFs(), "/")}
	sess, _ := NewTransferSession(c, "Test", fs)

	tests := []struct {
		path string
		want string
	}{
		{"foobar.jpg", "/Test/foobar.jpg"},
		{"foo/bar.jpg", "/Test/foo/bar.jpg"},
		{"/foo/bar.jpg", "/Test/foo/bar.jpg"},
		{"/f/o/o/b/a/r.jpg", "/Test/f/o/o/b/a/r.jpg"},
	}
	for _, tt := range tests {
		f, err := sess.Create(tt.path)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		have := f.Name()
		if tt.want != have {
			t.Fatalf("Want %s, have %s", tt.want, have)
		}
		dirPath := filepath.Dir(tt.want)
		if exists, err := fs.DirExists(dirPath); err != nil {
			t.Fatalf("DirExists failed: %v", err)
		} else if !exists {
			t.Fatalf("Directory %s not created", dirPath)
		}
	}
}

func TestTransferSession_Contents(t *testing.T) {
	c := getClient(t)
	fs := afero.NewBasePathFs(afero.NewMemMapFs(), "/")
	sess, _ := NewTransferSession(c, "Test", fs)
	afero.TempFile(sess.fs, "/", "uno")
	afero.TempFile(sess.fs, "/", "dos")

	want := 2
	have := len(sess.Contents())
	if want != have {
		t.Fatalf("Created %d files, only %d found", want, have)
	}
}

func TestTransferSession_createMetadataDir(t *testing.T) {
	c := getClient(t)
	fs := afero.NewBasePathFs(afero.NewMemMapFs(), "/")
	sess, _ := NewTransferSession(c, "Test", fs)

	if err := sess.createMetadataDir(); err != nil {
		t.Fatalf("createMetadataDir failed: %v", err)
	}
	info, err := sess.fs.Stat("metadata")
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
	c := getClient(t)
	fs := afero.NewBasePathFs(afero.NewMemMapFs(), "/")
	sess, _ := NewTransferSession(c, "Test", fs)
	if err := sess.Start(); err != nil {
		t.Fatal(err)
	}

	ts := c.Transfer.(*ts)
	if ts.approveReq.Directory != sess.Path {
		t.Errorf("Have %s, want %s", ts.approveReq.Directory, sess.Path)
	}
	if ts.approveReq.Type != "standard" {
		t.Errorf("Have %s, want \"standard\"", ts.approveReq.Type)
	}
}

func getClient(t *testing.T) *Client {
	url, _ := url.Parse("http://localhost")
	c := NewClient(nil, url.String(), "", "")
	c.Transfer = &ts{t: t}
	return c
}

type ts struct {
	t          *testing.T
	startReq   *TransferStartRequest
	approveReq *TransferApproveRequest
}

func (ts *ts) Start(ctx context.Context, req *TransferStartRequest) (*TransferStartResponse, *Response, error) {
	return nil, nil, nil
}

func (ts *ts) Approve(ctx context.Context, req *TransferApproveRequest) (*TransferApproveResponse, *Response, error) {
	ts.approveReq = req
	return &TransferApproveResponse{}, &Response{}, nil
}

func (ts *ts) Unapproved(ctx context.Context, req *TransferUnapprovedRequest) (*TransferUnapprovedResponse, *Response, error) {
	return &TransferUnapprovedResponse{
		Message: "Fetched unapproved transfers successfully.",
		Results: []*TransferUnapprovedResponseResult{
			&TransferUnapprovedResponseResult{Directory: "Test"},
		},
	}, &Response{}, nil
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
