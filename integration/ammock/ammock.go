package ammock

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"testing"
)

type Pipeline struct {
	URL         string
	User        string
	Key         string
	TransferDir string

	ln net.Listener

	t    *testing.T
	stop chan chan struct{}
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
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "OK")
	})
	err := http.Serve(p.ln, mux)
	if err != nil {
		p.t.Log("Pipeline server is now closed:", err)
	}
}

func (p *Pipeline) loop() {
	ch := <-p.stop
	p.ln.Close()
	close(ch)
}

func (p *Pipeline) AssertAPIUsed() {

}

func (p *Pipeline) AssertAPINotUsed() {

}

func (p *Pipeline) Stop() {
	ch := make(chan struct{})
	p.stop <- ch
	<-ch

	_ = os.RemoveAll(p.TransferDir)
}
