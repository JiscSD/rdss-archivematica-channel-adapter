package message

import (
	"fmt"
)

// Metadata Create

// MetadataCreateRequest represents the body of the message.
type MetadataCreateRequest struct {
	ResearchObject
}

// MetadataCreateRequest returns the body of the message.
func (m Message) MetadataCreateRequest() (*MetadataCreateRequest, error) {
	body, ok := m.MessageBody.(*MetadataCreateRequest)
	if !ok {
		return nil, fmt.Errorf("MetadataCreateRequest(): interface conversion error")
	}
	return body, nil
}

// Metadata Read

// MetadataReadRequest represents the body of the message.
type MetadataReadRequest struct {
	ObjectUuid *UUID `json:"objectUuid"`
}

// MetadataReadRequest returns the body of the message.
func (m Message) MetadataReadRequest() (*MetadataReadRequest, error) {
	b, ok := m.MessageBody.(*MetadataReadRequest)
	if !ok {
		return nil, fmt.Errorf("MetadataReadRequest(): interface conversion error")
	}
	return b, nil
}

// MetadataReadResponse represents the body of the message.
type MetadataReadResponse struct {
	ResearchObject
}

// MetadataReadResponse returns the body of the message.
func (m Message) MetadataReadResponse() (*MetadataReadResponse, error) {
	b, ok := m.MessageBody.(*MetadataReadResponse)
	if !ok {
		return nil, fmt.Errorf("MetadataReadResponse(): interface conversion error")
	}
	return b, nil
}

// Metadata Update

// MetadataUpdateRequest represents the body of the message.
type MetadataUpdateRequest struct {
	ResearchObject
}

// MetadataUpdateRequest returns the body of the message.
func (m Message) MetadataUpdateRequest() (*MetadataUpdateRequest, error) {
	b, ok := m.MessageBody.(*MetadataUpdateRequest)
	if !ok {
		return nil, fmt.Errorf("MetadataUpdateRequest(): interface conversion error")
	}
	return b, nil
}

// Metadata Delete

// MetadataDeleteRequest represents the body of the message.
type MetadataDeleteRequest struct {
	ObjectUuid *UUID `json:"objectUuid"`
}

// MetadataDeleteRequest returns the body of the message.
func (m Message) MetadataDeleteRequest() (*MetadataDeleteRequest, error) {
	b, ok := m.MessageBody.(*MetadataDeleteRequest)
	if !ok {
		return nil, fmt.Errorf("MetadataDeleteRequest(): interface conversion error")
	}
	return b, nil
}
