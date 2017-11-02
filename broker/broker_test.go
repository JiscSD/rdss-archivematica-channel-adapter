package broker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	logtest "github.com/sirupsen/logrus/hooks/test"

	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/broker/backend"
	_ "github.com/JiscRDSS/rdss-archivematica-channel-adapter/broker/backend/backendmock"
	bErrors "github.com/JiscRDSS/rdss-archivematica-channel-adapter/broker/errors"
	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/broker/message"
)

func ExampleBroker() {
	// See the definition of newBroker for more details.
	var b, _ = newBroker(nil)

	// Subscribe to MetadataCreate messages.
	b.SubscribeType(message.MessageTypeMetadataCreate, func(msg *message.Message) error {
		fmt.Println("MetadataCreate message received!")
		return nil
	})

	// A publisher can publish a MetadataCreate request.
	_ = b.Metadata.Create(context.Background(), &message.MetadataCreateRequest{})

	// Output: MetadataCreate message received!
}

// TestPanickingSubscriber will panic if the broker doesn't regain control.
func TestPanickingSubscriber(t *testing.T) {
	var b, _ = newBroker(nil)
	b.Subscribe(func(msg *message.Message) error {
		panic("error")
	})
	_ = b.Metadata.Create(context.Background(), &message.MetadataCreateRequest{})
}

func TestCounter(t *testing.T) {
	var b, _ = newBroker(nil)
	_ = b.Metadata.Create(context.Background(), &message.MetadataCreateRequest{})
	if got := b.Count(); got != 1 {
		t.Fatalf("b.Count mismatch: got %d, want %d", got, 1)
	}
	_ = b.Metadata.Delete(context.Background(), &message.MetadataDeleteRequest{})
	_ = b.Vocabulary.Patch(context.Background(), &message.VocabularyPatchRequest{})
	if got := b.Count(); got != 3 {
		t.Fatalf("b.Count mismatch: got %d, want %d", got, 3)
	}
}

func Test_messageHandler_duplicated(t *testing.T) {
	var (
		b, _ = newBroker(nil)
		m    = message.New(message.MessageTypeMetadataCreate, message.MessageClassCommand)
	)
	send := func() { _ = b.Request(context.Background(), m) }
	// Send the same message twice.
	send()
	send()
	if got := b.Count(); got != 1 {
		t.Fatalf("Broker should count only 1 message, seen %d", got)
	}
}

func Test_messageHandler_malformedJSON(t *testing.T) {
	b, _ := newBroker(nil)
	b.backend.Publish(b.config.QueueMain, []byte("INVALID-JSON"))
	if got := b.Count(); got != 0 {
		t.Fatal("The broker should count zero messages received but got:", got)
	}
}

func Test_messageHandler_errorHandling(t *testing.T) {
	var (
		b, _     = newBroker(nil)
		timeout  = make(chan bool, 1)
		received = make(chan bool, 1)
	)
	// The user subscribes to the broker and returns an error.
	b.Subscribe(func(msg *message.Message) error {
		return bErrors.New(bErrors.GENERR001, "...")
	})
	// We plug a subscriber to the backend to inspect messages received on the
	// error queue.
	b.backend.Subscribe(b.config.QueueError, func(data []byte) error {
		msg := &message.Message{}
		if err := json.Unmarshal(data, msg); err != nil {
			t.Fatal(err)
		}
		if msg.MessageHeader.ErrorCode != bErrors.GENERR001.String() {
			t.Fatal("Unexpected error code received:", msg.MessageHeader.ErrorCode)
		}
		received <- true
		return nil
	})
	// Send a message
	_ = b.Metadata.Create(context.Background(), &message.MetadataCreateRequest{})
	// Wait up to a second, fail if the message is not received in time.
	go func() {
		time.Sleep(1 * time.Second)
		timeout <- true
	}()
	select {
	case <-timeout:
		t.Fatal("The error wasn't delivered (or not soon enough)")
	case <-received:
	}
}

func Test_exists(t *testing.T) {
	var (
		b, _ = newBroker(nil)
		m    = message.New(message.MessageTypeMetadataCreate, message.MessageClassCommand)
	)
	// It should return false because the message was never seen before.
	if b.exists(m) {
		t.Fatal("b.exists() should have returned false because the message *was* new to the system")
	}
	// It should return true because the message is now recorded in the repo.
	if !b.exists(m) {
		t.Fatal("b.exists() should have returned true because the message *was not* new to the system")
	}
}

func Test_exists_putFails(t *testing.T) {
	var (
		b, h = newBroker(nil)
		m    = message.New(message.MessageTypeMetadataCreate, message.MessageClassCommand)
	)
	b.repository = &putErrRepo{b.repository}
	// It should return false because the message was never seen before.
	if b.exists(m) {
		t.Fatal("b.exists() should have returned false because the message *was* new to the system")
	}
	// It should return false because the message was never seen before, just
	// because putErrRepo failed to persist it.
	if b.exists(m) {
		t.Fatal("b.exists() should have returned false because the message *was* new to the system")
	}
	if entries := h.AllEntries(); len(entries) != 2 {
		t.Fatalf("Expected 2 log entries in the logger, %d seen", len(entries))
	}
}

func TestErrorQueues(t *testing.T) {
	b, _ := newBroker(t)
	defer b.Close()

	tests := []struct {
		name             string
		msg              *message.Message
		errorCode        bErrors.Kind
		errorDescription string
		queue            string
	}{
		{
			"Invalid queue",
			message.New(message.MessageTypeMetadataCreate, message.MessageClassCommand),
			bErrors.GENERR001,
			"mandatory field `foobar` is missing",
			b.config.QueueInvalid,
		},
		{
			"Error queue",
			message.New(message.MessageTypeMetadataCreate, message.MessageClassCommand),
			bErrors.GENERR006,
			"local repository is not accessible",
			b.config.QueueError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var recv bool
			b.backend.Subscribe(tt.queue, func(data []byte) error {
				msg := &message.Message{}
				err := json.Unmarshal(data, msg)
				if err != nil {
					t.Fatal("error unmarshaling bytes", err)
				}
				if msg.MessageHeader.ErrorCode != tt.errorCode.String() {
					t.Fatalf("b.handleErrMessage(): ErrorCode got %s, want %s", msg.MessageHeader.ErrorCode, tt.errorCode)
				}
				if msg.MessageHeader.ErrorDescription != tt.errorDescription {
					t.Fatalf("b.handleErrMessage(): ErrorDescription got %s, want %s", msg.MessageHeader.ErrorDescription, tt.errorDescription)
				}
				recv = true
				return nil
			})

			b.handleErrMessage(tt.msg, bErrors.New(tt.errorCode, tt.errorDescription), tt.queue)

			if !recv {
				t.Error("message not seen in the invalid queue")
			}
		})
	}
}

func TestMessageRetry(t *testing.T) {
	var (
		b, h = newRetryBroker(nil)
		m    = message.New(message.MessageTypeMetadataCreate, message.MessageClassCommand)
	)
	b.Request(context.Background(), m)
	var retryCount int
	for _, log := range h.AllEntries() {
		if log.Message == "BackendWithRetry backoff" {
			retryCount = retryCount + 1
		}
	}
	if retryCount != 3 {
		t.Error("Wrong number of network retries")
	}
}

func newBroker(t *testing.T) (*Broker, *logtest.Hook) {
	return newBroker_(t, "backendmock")
}

func newBroker_(t *testing.T, brokerName string) (*Broker, *logtest.Hook) {
	if t == nil {
		t = &testing.T{}
	}
	logger, logh := logtest.NewNullLogger()
	bc, err := backend.Dial(brokerName, []backend.DialOpts{}...)
	if err != nil {
		t.Fatal("newBroker() backend creation failed:", err)
	}
	b, err := New(bc, logger, &Config{
		QueueError:   "error",
		QueueInvalid: "invalid",
		QueueMain:    "main",
		RepositoryConfig: &RepositoryConfig{
			Backend: "builtin",
		},
	})
	if err != nil {
		t.Fatal("newBroker() broker creation failed:", err)
	}
	return b, logh
}

func newRetryBroker(t *testing.T) (*Broker, *logtest.Hook) {
	return newBroker_(t, "backendmockretry")
}

// putErrRepo is a Repository that fails to Put messages.
type putErrRepo struct {
	Repository
}

func (r *putErrRepo) Put(msg *message.Message) error {
	return errors.New("put failed")
}
