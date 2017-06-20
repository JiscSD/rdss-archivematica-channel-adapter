package consumer_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/spf13/afero"

	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/amclient"
	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/broker"
	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/broker/backend"
	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/broker/backend/backendmock"
	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/broker/message"
	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/consumer"
)

var (
	ctx    context.Context
	cancel context.CancelFunc
	stop   = make(chan struct{})
	c      consumer.Consumer
	ba     backend.Backend
	br     *broker.Broker
	fs     afero.Fs
	mux    *http.ServeMux
	server *httptest.Server
)

func TestMain(m *testing.M) {
	os.Exit(testMainWrapper(m))
}

func testMainWrapper(m *testing.M) int {
	tearUp()
	defer tearDown()
	return m.Run()
}

func tearUp() {
	logger := log.New()

	var err error
	ba, err = backend.Dial("backendmock")
	if err != nil {
		panic(err)
	}
	defer ba.Close()
	br, err = broker.New(ba, logger, &broker.Config{QueueError: "f", QueueInvalid: "o", QueueMain: "o"})
	if err != nil {
		panic(err)
	}

	fs = afero.NewMemMapFs()

	// Archivematica client with HTTP server mock using httptest
	mux = http.NewServeMux()
	server = httptest.NewServer(mux)
	url, _ := url.Parse(server.URL)
	amc, _ := amclient.New(nil, url.String(), "", "", amclient.SetFs(fs))
	fs := afero.NewBasePathFs(afero.NewMemMapFs(), "/")

	ctx, cancel = context.WithCancel(context.Background())

	// Consumer with mocks
	c := consumer.MakeConsumer(
		ctx,
		logger,
		br,
		amc,
		&RandomObjectStorage{},
		fs)

	go func() {
		c.Start()
		stop <- struct{}{}
		cancel() // just to make vet happy
	}()
}

func tearDown() {
	cancel()
	<-stop
	server.Close()
}

func TestValidCreateMetadataMessage(t *testing.T) {
	// Build message MetadataCreate
	msg := message.New(message.TypeMetadataCreate, message.ClassCommand)
	body := &message.MetadataCreateRequest{
		UUID:  "a90652dd-6abd-424c-b7ce-d6728c7f3f9f",
		Title: "Research about birds in Doñana National Park",
		Files: []*message.MetadataFile{
			&message.MetadataFile{
				UUID:            "One",
				StorageLocation: "s3://bucket-01/one.mp3",
			},
			&message.MetadataFile{
				UUID:            "Two",
				StorageLocation: "s3://bucket-01/two.wav",
			},
		},
	}
	msg.Body = body

	// Install our custom HTTP handler.
	mux.HandleFunc("/api/transfer/start_transfer/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Request method should be POST, got: %s", r.Method)
		}

		decoder := json.NewDecoder(r.Body)
		msg := &amclient.TransferStartRequest{}
		err := decoder.Decode(msg)
		if err != nil {
			t.Errorf("Response cannot be decoded: %s", err)
		}
		defer r.Body.Close()

		if len(msg.Paths) < len(body.Files) {
			t.Errorf("Response does not include two files")
		}

		fmt.Println(body.Files)

		fmt.Fprint(w, `{"message": "Copy successful.", "path": "/foobar"`)
	})

	t.Run("Publish message", func(t *testing.T) {
		data, err := json.Marshal(msg)
		if err != nil {
			panic(err)
		}

		bmock := ba.(*backendmock.BackendImpl)
		bmock.Publish("", data)

		if br.Count != 1 {
			t.Errorf("Backend does not count 1 message sent")
		}

		const total = 4
		for i := 1; i <= total; i++ {
			bmock.Publish("", data)
		}
		if br.Count != total+1 {
			t.Errorf("Backend does not count 1 message sent")
		}
	})
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

// RandomObjectStorage is an implementation of ObjectStorage that downloads any
// file requested with random content. This is used in consumer_test.go.
type RandomObjectStorage struct{}

// Download implements ObjectStorage
func (s *RandomObjectStorage) Download(_ context.Context, w io.WriterAt, _ string) (int64, error) {
	data := make([]byte, 8)
	l, err := rand.Read(data)
	if err != nil {
		return 0, err
	}
	l, err = w.WriteAt(data, 0)
	return int64(l), err
}
