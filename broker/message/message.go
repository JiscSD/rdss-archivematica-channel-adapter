package message

import (
	"encoding/json"
	"reflect"
	"strings"
	"time"

	bErrors "github.com/JiscRDSS/rdss-archivematica-channel-adapter/broker/errors"
	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/version"
)

// Message represents RDSS messages.
type Message struct {
	// MessageHeader carries the message headers.
	MessageHeader MessageHeader

	// MessageBody carries the message payload.
	MessageBody interface{}

	// Raw payload streams using during validation (gojsonschema), e.g.:
	//	loader := gojsonschema.NewBytesLoader(msg.body)
	//	validator.Validate(loader)
	header []byte
	body   []byte
}

// New returns a pointer to a new message with a new ID.
func New(t MessageType, c MessageClass) *Message {
	now := time.Now()
	return &Message{
		MessageHeader: MessageHeader{
			ID:           NewUUID(),
			MessageClass: c,
			MessageType:  t,
			MessageTimings: MessageTimings{
				PublishedTimestamp:  Timestamp(now),
				ExpirationTimestamp: Timestamp(now.AddDate(0, 1, 0)), // One month later.
			},
			MessageSequence: MessageSequence{
				Sequence: NewUUID(),
				Position: 1,
				Total:    1,
			},
			Version:   Version,
			Generator: version.AppVersion(),
		},
		MessageBody: typedBody(t, nil),
	}
}

// messageAlias is proxy type for Message. Using json.RawMessage in order to:
// - Delay JSON decoding.
// - Precompute JSON encoding.
type messageAlias struct {
	MessageHeader json.RawMessage `json:"messageHeader"`
	MessageBody   json.RawMessage `json:"messageBody"`
}

func (m *Message) ID() string {
	if m.MessageHeader.ID == nil {
		return ""
	}
	return m.MessageHeader.ID.String()
}

func (m Message) Type() string {
	t := reflect.TypeOf(m.MessageBody)
	var name string
	if t.Kind() == reflect.Ptr {
		name = t.Elem().String()
	} else {
		name = t.String()
	}
	parts := strings.Split(name, ".")
	return parts[1]
}

func (m *Message) TagError(err error) {
	if err == nil {
		return
	}
	e, ok := err.(*bErrors.Error)
	if ok && e != nil {
		m.MessageHeader.ErrorCode = e.Kind.String()
		m.MessageHeader.ErrorDescription = e.Err.Error()
	} else if !ok && err != nil {
		m.MessageHeader.ErrorCode = "Unknown"
		m.MessageHeader.ErrorDescription = err.Error()
	}
}

// MarshalJSON implements Marshaler.
func (m *Message) MarshalJSON() ([]byte, error) {
	header, err := json.Marshal(m.MessageHeader)
	if err != nil {
		return nil, err
	}
	body, err := json.Marshal(m.MessageBody)
	if err != nil {
		return nil, err
	}
	return json.Marshal(&messageAlias{json.RawMessage(header), json.RawMessage(body)})
}

// UnmarshalJSON implements Unmarshaler.
func (m *Message) UnmarshalJSON(data []byte) error {
	msg := messageAlias{}
	if err := json.Unmarshal(data, &msg); err != nil {
		return err
	}
	if err := json.Unmarshal(msg.MessageHeader, &m.MessageHeader); err != nil {
		return err
	}
	m.MessageBody = typedBody(m.MessageHeader.MessageType, m.MessageHeader.CorrelationID)
	m.header = []byte(msg.MessageHeader)
	m.body = []byte(msg.MessageBody)
	return json.Unmarshal(msg.MessageBody, m.MessageBody)
}

// typedBody returns an interface{} type where the type of the underlying value
// is chosen after the header message type.
func typedBody(t MessageType, correlationID *UUID) interface{} {
	var body interface{}
	switch {
	case t == MessageTypeMetadataCreate:
		body = new(MetadataCreateRequest)
	case t == MessageTypeMetadataRead:
		if correlationID == nil {
			body = new(MetadataReadRequest)
		} else {
			body = new(MetadataReadResponse)
		}
	case t == MessageTypeMetadataUpdate:
		body = new(MetadataUpdateRequest)
	case t == MessageTypeMetadataDelete:
		body = new(MetadataDeleteRequest)
	case t == MessageTypeVocabularyRead:
		if correlationID == nil {
			body = new(VocabularyReadRequest)
		} else {
			body = new(VocabularyReadResponse)
		}
	case t == MessageTypeVocabularyPatch:
		body = new(VocabularyPatchRequest)
	}
	return body
}
