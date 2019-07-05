package broker

import (
	"context"
	"errors"

	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/broker/message"
)

// MetadataService generates Metadata-type messages.
type MetadataService interface {
	Create(context.Context, *message.MetadataCreateRequest) error
	Read(context.Context, *message.MetadataReadRequest) (*message.MetadataReadResponse, error)
	Update(context.Context, *message.MetadataUpdateRequest) error
	Delete(context.Context, *message.MetadataDeleteRequest) error
}

// MetadataServiceOp implements MetadataService.
type MetadataServiceOp struct {
	broker *Broker
}

// Create publishes a MetadataCreate message.
func (s *MetadataServiceOp) Create(ctx context.Context, req *message.MetadataCreateRequest) error {
	msg := message.New(message.MessageTypeEnum_MetadataCreate, message.MessageClassEnum_Command)
	msg.MessageBody = req

	return s.broker.Request(ctx, msg)
}

// Read publishes a MetadataRead message.
func (s *MetadataServiceOp) Read(ctx context.Context, req *message.MetadataReadRequest) (*message.MetadataReadResponse, error) {
	msg := message.New(message.MessageTypeEnum_MetadataRead, message.MessageClassEnum_Command)
	msg.MessageBody = req

	resp, err := s.broker.RequestResponse(ctx, msg)
	r, ok := resp.MessageBody.(*message.MetadataReadResponse)
	if !ok {
		return nil, errors.New("unexpected")
	}

	return r, err
}

// Update publishes a MetadataUpdate message.
func (s *MetadataServiceOp) Update(ctx context.Context, req *message.MetadataUpdateRequest) error {
	msg := message.New(message.MessageTypeEnum_MetadataUpdate, message.MessageClassEnum_Command)
	msg.MessageBody = req

	return s.broker.Request(ctx, msg)
}

// Delete publishes a MetadataDelete message.
func (s *MetadataServiceOp) Delete(ctx context.Context, req *message.MetadataDeleteRequest) error {
	msg := message.New(message.MessageTypeEnum_MetadataDelete, message.MessageClassEnum_Command)
	msg.MessageBody = req

	return s.broker.Request(ctx, msg)
}
