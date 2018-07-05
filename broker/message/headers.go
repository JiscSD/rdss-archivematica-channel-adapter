package message

import (
	"encoding/json"
	"fmt"
	"strings"
)

// MessageHeader contains important metadata describing the Message itself,
// including the type of Message, routing information, timings, sequencing, and
// so forth.
type MessageHeader struct {
	ID               *UUID            `json:"messageId"`
	CorrelationID    *UUID            `json:"correlationId,omitempty"`
	MessageClass     MessageClass     `json:"messageClass"`
	MessageType      MessageType      `json:"messageType"`
	ReturnAddress    string           `json:"returnAddress,omitempty"`
	MessageTimings   MessageTimings   `json:"messageTimings"`
	MessageSequence  MessageSequence  `json:"messageSequence"`
	MessageHistory   []MessageHistory `json:"messageHistory,omitempty"`
	Version          string           `json:"version"`
	ErrorCode        string           `json:"errorCode,omitempty"`
	ErrorDescription string           `json:"errorDescription,omitempty"`
	Generator        string           `json:"generator"`
}

// MessageClass is one of Command, Event or Document.
type MessageClass int

const (
	MessageClassCommand MessageClass = iota
	MessageClassEvent
	MessageClassDocument
)

var (
	classMap = map[string]MessageClass{
		"Command":  MessageClassCommand,
		"Event":    MessageClassEvent,
		"Document": MessageClassDocument,
	}
	classMapInv = make(map[MessageClass]string)
)

func init() {
	for k, v := range classMap {
		classMapInv[v] = k
	}
}

func (c MessageClass) String() string {
	v, ok := classMapInv[c]
	if !ok {
		return "Unknown"
	}
	return v
}

// MarshalJSON implements Marshaler.
func (c MessageClass) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.String())
}

// UnmarshalJSON implements Unmarshaler.
func (c *MessageClass) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	have, ok := classMap[s]
	if !ok {
		return fmt.Errorf("invalid message class %q", s)
	}
	*c = have
	return nil
}

// MessageType describes the message type in the format of <Type><Operation> and
// is described in the Message Body section.
type MessageType int

const (
	MessageTypeMetadataCreate MessageType = iota
	MessageTypeMetadataRead
	MessageTypeMetadataUpdate
	MessageTypeMetadataDelete
	MessageTypeVocabularyRead
	MessageTypeVocabularyPatch
)

var (
	typeMap = map[string]MessageType{
		"MetadataCreate":  MessageTypeMetadataCreate,
		"MetadataRead":    MessageTypeMetadataRead,
		"MetadataUpdate":  MessageTypeMetadataUpdate,
		"MetadataDelete":  MessageTypeMetadataDelete,
		"VocabularyRead":  MessageTypeVocabularyRead,
		"VocabularyPatch": MessageTypeVocabularyPatch,
	}
	typeMapInv = make(map[MessageType]string)
)

func init() {
	for k, v := range typeMap {
		typeMapInv[v] = k
	}
}

func (t MessageType) String() string {
	v, ok := typeMapInv[t]
	if !ok {
		return "Unknown"
	}
	return v
}

// MarshalJSON implements Marshaler
func (t MessageType) MarshalJSON() ([]byte, error) {
	s := t.String()
	s = strings.TrimSuffix(s, "Request")
	s = strings.TrimSuffix(s, "Response")
	return json.Marshal(s)
}

// UnmarshalJSON implements Unmarshaler
func (t *MessageType) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	v, ok := typeMap[s]
	if !ok {
		return fmt.Errorf("invalid message type %q", s)
	}
	*t = v
	return nil
}

type MessageTimings struct {
	PublishedTimestamp  Timestamp `json:"publishedTimestamp"`
	ExpirationTimestamp Timestamp `json:"expirationTimestamp,omitempty"`
}

type MessageSequence struct {
	Sequence *UUID `json:"sequence"`
	Position int   `json:"position"`
	Total    int   `json:"total"`
}

type MessageHistory struct {
	MachineId      string    `json:"machineId"`
	MachineAddress string    `json:"machineAddress"`
	Timestamp      Timestamp `json:"timestamp"`
}
