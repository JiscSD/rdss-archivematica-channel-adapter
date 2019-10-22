package message

//go:generate go run generator.go

import (
	"encoding/json"
	"time"

	bErrors "github.com/JiscSD/rdss-archivematica-channel-adapter/broker/errors"
	"github.com/JiscSD/rdss-archivematica-channel-adapter/version"
)

// Message represents RDSS messages.
type Message struct {
	// MessageHeader carries the message headers.
	MessageHeader MessageHeader

	// MessageBody carries the message payload.
	MessageBody interface{}
}

// New returns a pointer to a new message with a new ID.
func New(t MessageTypeEnum, c MessageClassEnum) *Message {
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

func (m *Message) TagError(err error) {
	if err == nil {
		return
	}
	e, ok := err.(*bErrors.Error)
	if ok && e != nil {
		m.MessageHeader.ErrorCode = e.Kind.String()
		m.MessageHeader.ErrorDescription = e.Err.Error()
	} else if !ok {
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
	return json.Marshal(&messageAlias{
		MessageHeader: json.RawMessage(header),
		MessageBody:   json.RawMessage(body),
	})
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
	return json.Unmarshal(msg.MessageBody, m.MessageBody)
}

// typedBody returns an interface{} type where the type of the underlying value
// is chosen after the header message type.
func typedBody(t MessageTypeEnum, correlationID *UUID) interface{} {
	var body interface{}
	switch {
	case t == MessageTypeEnum_MetadataCreate:
		body = new(MetadataCreateRequest)
	case t == MessageTypeEnum_MetadataRead:
		if correlationID == nil {
			body = new(MetadataReadRequest)
		} else {
			body = new(MetadataReadResponse)
		}
	case t == MessageTypeEnum_MetadataUpdate:
		body = new(MetadataUpdateRequest)
	case t == MessageTypeEnum_MetadataDelete:
		body = new(MetadataDeleteRequest)
	case t == MessageTypeEnum_PreservationEvent:
		body = new(PreservationEventRequest)
	}
	return body
}
