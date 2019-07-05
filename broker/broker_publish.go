package broker

import (
	"context"
	"errors"

	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/broker/message"
)

// Request sends a fire-and-forget request to RDSS.
func (b *Broker) Request(_ context.Context, msg *message.Message) error {
	payload, err := msg.MarshalJSON()
	if err != nil {
		return err
	}
	return b.publishMessage(b.snsTopicMainARN, string(payload))
}

// RequestResponse sends a request and waits until a response is received.
func (b *Broker) RequestResponse(context.Context, *message.Message) (*message.Message, error) {
	return nil, errors.New("not implemented yet")
}
