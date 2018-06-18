package consumer

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/spf13/afero"

	log "github.com/sirupsen/logrus"
	logt "github.com/sirupsen/logrus/hooks/test"

	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/amclient"
	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/broker/message"
	s3lib "github.com/JiscRDSS/rdss-archivematica-channel-adapter/s3"
)

func Test_handleMetadataCreateRequest_errMessageType(t *testing.T) {
	c, _ := getConsumer(t)
	if err := c.handleMetadataCreateRequest(&message.Message{}); err == nil {
		t.Fatal("Expected non-nil error, got nil")
	}
}
func Test_handleMetadataCreateRequest_emptyFiles(t *testing.T) {
	c, _ := getConsumer(t)
	msg := message.New(message.MessageTypeMetadataCreate, message.MessageClassCommand)
	msg.MessageBody.(*message.MetadataCreateRequest).ObjectUuid = message.MustUUID("a3c982c9-e035-4702-bd97-cef8ab618ad5")
	if err := c.handleMetadataCreateRequest(msg); err != nil {
		t.Fatalf("Expected nil error, got %s", err)
	}
}

func Test_describeDataset(t *testing.T) {
	var (
		ts   = getTransferSession(t)
		want = map[string][][2]string{
			"objects/": [][2]string{
				[2]string{"dc.title", "Title"},
				[2]string{"dc.type", "audio"},
				[2]string{"dc.identifier", "I1"},
				[2]string{"dc.identifier", "I2"},
				[2]string{"dcterms.issued", "date of publication"},
				[2]string{"dc.publicationYear", "date of publication"},
				[2]string{"dc.publisher", "orgname"},
				[2]string{"dc.creatorName", "person 2"},
				[2]string{"dc.publisher", "person 3"},
			},
		}
	)
	describeDataset(ts, &message.ResearchObject{
		ObjectTitle:        "Title",
		ObjectResourceType: message.ResourceTypeEnum_audio,
		ObjectIdentifier: []message.Identifier{
			message.Identifier{IdentifierValue: "I1"},
			message.Identifier{IdentifierValue: "I2"},
		},
		ObjectDate: []message.Date{
			message.Date{DateType: message.DateTypeEnum_published, DateValue: "date of publication"},
			message.Date{DateType: message.DateTypeEnum_accepted, DateValue: "date of ..."},
		},
		ObjectOrganisationRole: []message.OrganisationRole{
			message.OrganisationRole{
				Organisation: message.Organisation{OrganisationName: "orgname"},
			},
		},
		ObjectPersonRole: []message.PersonRole{
			message.PersonRole{Role: message.PersonRoleEnum_relatedPerson, Person: message.Person{PersonGivenNames: "person 1"}},
			message.PersonRole{Role: message.PersonRoleEnum_dataCreator, Person: message.Person{PersonGivenNames: "person 2"}},
			message.PersonRole{Role: message.PersonRoleEnum_publisher, Person: message.Person{PersonGivenNames: "person 3"}},
		},
	})
	entries := ts.Metadata.Entries()
	if !reflect.DeepEqual(want, entries) {
		t.Fatalf("describeFile(); unexpected result, want %v, got %v", want, entries)
	}
}

func Test_describeFile(t *testing.T) {
	var (
		ts   = getTransferSession(t)
		want = map[string][][2]string{
			"objects/foobar.jpg": [][2]string{
				[2]string{"dc.identifier", "I1"},
				[2]string{"dc.title", "foobar.jpg"},
			},
		}
	)
	describeFile(ts, "foobar.jpg", &message.File{FileIdentifier: "I1", FileName: "foobar.jpg"})
	entries := ts.Metadata.Entries()
	if !reflect.DeepEqual(want, entries) {
		t.Fatalf("describeFile(); unexpected result, want %v, got %v", want, entries)
	}
}

func Test_downloadFile_HTTP(t *testing.T) {
	body := []byte(`data`)
	downloadFile_HTTP(t, body, func(w http.ResponseWriter, r *http.Request) {
		w.Write(body)
	})
}

type mockBackoff struct {
}

func (mb *mockBackoff) NextBackOff() time.Duration {
	return time.Duration(0)
}

func (mb *mockBackoff) Reset() {

}

func Test_downloadFile_HTTP_retry(t *testing.T) {
	body := []byte(`data`)
	count := 0
	downloadFile_HTTP(t, body, func(w http.ResponseWriter, r *http.Request) {
		if count < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			count = count + 1
		} else {
			w.Write(body)
		}
	})
	if count != 3 {
		t.Fatalf("Did not retry correct number of times")
	}
}

func downloadFile_HTTP(t *testing.T, body []byte, requestHandler func(http.ResponseWriter, *http.Request)) {
	ctx := context.Background()
	logger, hook := logt.NewNullLogger()
	logger.SetLevel(log.DebugLevel)
	_, f := getFile(t, "/foobar.jpg")

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	url, _ := url.Parse(server.URL)

	mux.HandleFunc("/", requestHandler)

	if err := downloadFile(logger, ctx, nil, http.DefaultClient, f, message.StorageTypeEnum_HTTP, url.String(), &mockBackoff{}); err != nil {
		t.Fatal(err)
	}
	if info, err := f.Stat(); err != nil {
		t.Fatal(err)
	} else {
		if got, want := info.Size(), len(body); int(got) != want {
			t.Fatalf("Returned file has unexpected size; want %d, got %d", want, got)
		}
	}

	entries := hook.AllEntries()
	if len(entries) != 2 {
		t.Fatal("Unexpected number of log entries")
	}
	if entries[1].Message != fmt.Sprintf("Downloaded %s - %d bytes written", url.String(), len(body)) {
		t.Fatal("Unexpected log entry")
	}
}

func Test_downloadFile_S3(t *testing.T) {
	ctx := context.Background()
	logger, _ := logt.NewNullLogger()
	logger.SetLevel(log.DebugLevel)
	_, f := getFile(t, "foobar.jpg")

	body := []byte(`data`)
	s3c := &mockS3Client{t: t, data: ioutil.NopCloser(bytes.NewReader(body))}
	s3d := s3manager.NewDownloaderWithClient(s3c)
	client := &s3lib.ObjectStorageImpl{S3: s3c, S3Downloader: s3d}

	if err := downloadFile(logger, ctx, client, nil, f, message.StorageTypeEnum_S3, "s3://bucket/foobar.jpg", nil); err != nil {
		t.Fatal(err)
	}
	if info, err := f.Stat(); err != nil {
		t.Fatal(err)
	} else {
		if got, want := info.Size(), len(body); int(got) != want {
			t.Fatalf("Returned file has unexpected size; want %d, got %d", want, got)
		}
	}
}

type mockS3Client struct {
	s3iface.S3API
	t    *testing.T
	data io.ReadCloser
}

func (c *mockS3Client) GetObjectWithContext(ctx aws.Context, input *s3.GetObjectInput, opts ...request.Option) (*s3.GetObjectOutput, error) {
	return &s3.GetObjectOutput{
		Body:         c.data,
		ContentRange: aws.String("1"),
	}, nil
}

func getFile(t *testing.T, name string) (afero.Afero, afero.File) {
	fs := afero.Afero{Fs: afero.NewBasePathFs(afero.NewMemMapFs(), "/")}
	f, err := fs.Create(name)
	if err != nil {
		t.Fatal(err)
	}
	return fs, f
}

func getConsumer(t *testing.T) (*ConsumerImpl, *logt.Hook) {
	logger, hook := logt.NewNullLogger()
	c := &ConsumerImpl{
		ctx:     context.Background(),
		logger:  logger,
		storage: NewStorageInMemory(),
	}
	return c, hook
}

func getTransferSession(t *testing.T) *amclient.TransferSession {
	return &amclient.TransferSession{
		Metadata: amclient.NewMetadataSet(nil),
	}
}
