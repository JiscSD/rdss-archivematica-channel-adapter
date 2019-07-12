package ammock

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strings"
	"testing"
)

type Pipeline struct {
	URL         string
	User        string
	Key         string
	TransferDir string

	t    *testing.T
	ln   net.Listener
	stop chan chan struct{}
	used bool
}

func New(t *testing.T) *Pipeline {
	p := &Pipeline{
		User: "test",
		Key:  "test",
		t:    t,
		stop: make(chan chan struct{}),
	}

	// Transfer directory.
	name, err := ioutil.TempDir("", "adapter-integration-test-transfer-dir")
	if err != nil {
		p.t.Fatal("Cannot create temporary directory:", err)
	}
	p.TransferDir = name

	// Network listener
	ln, err := net.Listen("tcp4", "localhost:")
	if err != nil {
		p.t.Fatal("Cannot create network listener:", err)
	}
	p.ln = ln
	p.URL = fmt.Sprintf("http://%s/api", ln.Addr().String())

	go p.createServer()
	go p.loop()

	return p
}

func (p *Pipeline) createServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", p.apiHandler)
	err := http.Serve(p.ln, mux)
	if err != nil {
		p.t.Log("Pipeline server is now closed:", err)
	}
}

func (p *Pipeline) apiHandler(w http.ResponseWriter, r *http.Request) {
	const (
		transferID  = "096a284d-5067-4de0-a0a4-a684018cd6df"
		sipID       = "41699e73-ec9e-4240-b153-71f4155e7da4"
		storeLinkID = "20515483-25ed-4133-b23e-5bb14cab8e22"
	)

	p.used = true

	path := strings.TrimSuffix(r.URL.Path, "/")

	// The adapter creates a package.
	if path == "/api/v2beta/package" {
		fmt.Fprint(w, `{"id": "`+transferID+`"}`)
		return
	}

	// The adapter waits until it completes.
	if path == "/api/transfer/status/"+transferID {
		fmt.Fprint(w, `{
			"status": "COMPLETE",
			"name": "...",
			"sip_uuid": "`+sipID+`",
			"microservice": "...",
			"directory": "...",
			"path": "...",
			"message": "...",
			"type": "...",
			"uuid": "..."
		}`)
		return
	}

	// The adapter waits until it is stored.
	if path == "/api/v2beta/jobs/"+sipID {
		fmt.Fprint(w, `[
			{
				"uuid": "...",
				"name": "Store the AIP",
				"status": "COMPLETE",
				"microservice": "...",
				"link_uuid": "`+storeLinkID+`",
				"tasks": []
			}
		]`)
		return
	}

	http.NotFound(w, r)
}

func (p *Pipeline) loop() {
	ch := <-p.stop
	p.ln.Close()
	close(ch)
}

func (p *Pipeline) AssertAPIUsed() {
	p.t.Helper()
	if !p.used {
		p.t.Fatal("Pipeline seems unused")
	}
}

func (p *Pipeline) AssertAPINotUsed() {
	p.t.Helper()
	if p.used {
		p.t.Fatal("Pipeline seems used")
	}
}

func (p *Pipeline) AssertTransferDirIsEmpty() {
	p.t.Helper()
	if !isEmptyDir(p.t, p.TransferDir) {
		p.t.Fatal("Transfer directory is not empty")
	}
}

func (p *Pipeline) AssertTransferDirIsNotEmpty() {
	p.t.Helper()
	if isEmptyDir(p.t, p.TransferDir) {
		p.t.Fatal("Transfer directory is empty")
	}
}

func (p *Pipeline) Stop() {
	ch := make(chan struct{})
	p.stop <- ch
	<-ch

	_ = os.RemoveAll(p.TransferDir)
}

func isEmptyDir(t *testing.T, name string) bool {
	entries, err := ioutil.ReadDir(name)
	if err != nil {
		t.Fatal("Empty dir check failed in:", name)
	}
	return len(entries) == 0
}
