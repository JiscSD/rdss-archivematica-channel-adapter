package broker

import (
	"context"

	"github.com/JiscSD/rdss-archivematica-channel-adapter/broker/message"
)

// PreservationService publishes Preservation-type messages.
type PreservationService interface {
	Event(context.Context, *message.PreservationEventRequest) error
}

// PreservationServiceOp implements PreservationService.
type PreservationServiceOp struct {
	broker *Broker
}

// Event publishes a PreservationEvent message.
func (s *PreservationServiceOp) Event(ctx context.Context, req *message.PreservationEventRequest) error {
	msg := message.New(message.MessageTypeEnum_PreservationEvent, message.MessageClassEnum_Event)
	msg.MessageBody = req

	return s.broker.Request(ctx, msg)
}
