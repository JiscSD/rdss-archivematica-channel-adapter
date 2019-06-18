package broker

import (
	"context"

	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/broker/message"
)

type PreservationService interface {
	Event(context.Context, *message.PreservationEventRequest) error
}

type PreservationServiceOp struct {
	broker *Broker
}

func (s *PreservationServiceOp) Event(ctx context.Context, req *message.PreservationEventRequest) error {
	msg := message.New(message.MessageTypeEnum_PreservationEvent, message.MessageClassEnum_Event)
	msg.MessageBody = req

	return s.broker.Request(ctx, msg)
}
