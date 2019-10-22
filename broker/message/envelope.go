package message

import (
	"encoding/json"
	"fmt"
)

// Envelop reads messages superficially only to extract the core attributes.
// It enables users to look up enough information to perform validation and/or
// conversion while avoding other potential decoding issues that we prefer to
// defer.
type Envelope struct {
	MessageHeader json.RawMessage `json:"messageHeader"`
	MessageBody   json.RawMessage `json:"messageBody"`
	Attributes    Attributes      `json:"-"`
}

// Open returns the Envelope of a message stream.
func Open(stream []byte) (*Envelope, error) {
	e := Envelope{Attributes: Attributes{}}

	if err := json.Unmarshal(stream, &e); err != nil {
		return nil, fmt.Errorf("error decoding header/body streams: %v", err)
	}

	if err := e.inspect(); err != nil {
		return nil, fmt.Errorf("error decoding envelope attributes: %v", err)
	}

	return &e, nil
}

type Attributes struct {
	Version       string `json:"version"`
	MessageType   string `json:"messageType"`
	CorrelationID string `json:"correlationId"`
}

// inspect extracts the main headers (version, type, correlation).
func (e *Envelope) inspect() error {
	if err := json.Unmarshal(e.MessageHeader, &e.Attributes); err != nil {
		return err
	}

	if e.Attributes.Version == "" {
		return fmt.Errorf("version header is empty or missing")
	}

	if e.Attributes.MessageType == "" {
		return fmt.Errorf("message type header is empty or missing")
	}

	return nil
}

const (
	SchemaIDMetadataCreateRequest    = "https://www.jisc.ac.uk/rdss/schema/message/metadata/create_request.json/#/definitions/MetadataCreateRequest"
	SchemaIDMetadataUpdateRequest    = "https://www.jisc.ac.uk/rdss/schema/message/metadata/update_request.json/#/definitions/MetadataUpdateRequest"
	SchemaIDMetadataDeleteRequest    = "https://www.jisc.ac.uk/rdss/schema/message/metadata/delete_request.json/#/definitions/MetadataDeleteRequest"
	SchemaIDMetadataReadRequest      = "https://www.jisc.ac.uk/rdss/schema/message/metadata/read_request.json/#/definitions/MetadataReadRequest"
	SchemaIDMetadataReadResponse     = "https://www.jisc.ac.uk/rdss/schema/message/metadata/read_response.json/#/definitions/MetadataReadResponse"
	SchemaIDPreservationEventRequest = "https://www.jisc.ac.uk/rdss/schema/message/preservation/preservation_event_request.json/#/definitions/PreservationEventRequest"
)

// SchemaDefinition returns the ID of the schema definition of the message.
func (e *Envelope) SchemaDefinition() string {
	switch e.Attributes.MessageType {
	case MessageTypeEnum_MetadataCreate.String():
		return SchemaIDMetadataCreateRequest
	case MessageTypeEnum_MetadataUpdate.String():
		return SchemaIDMetadataUpdateRequest
	case MessageTypeEnum_MetadataDelete.String():
		return SchemaIDMetadataDeleteRequest
	case MessageTypeEnum_MetadataRead.String():
		if e.Attributes.CorrelationID == "" {
			return SchemaIDMetadataReadRequest
		} else {
			return SchemaIDMetadataReadResponse
		}
	case MessageTypeEnum_PreservationEvent.String():
		return SchemaIDPreservationEventRequest
	}

	return ""
}
