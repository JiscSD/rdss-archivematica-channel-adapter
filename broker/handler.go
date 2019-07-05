package broker

import (
	"fmt"
	"sync"

	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/broker/message"

	"github.com/pkg/errors"
)

// MessageHandler is a callback function supplied by subscribers.
type MessageHandler func(msg *message.Message) error

var UnassignedHandlerErr = errors.New("unknown message handler")

// subscriptions associates message handlers to message types.
type subscriptions struct {
	s map[message.MessageTypeEnum]MessageHandler
	sync.RWMutex
}

// Subscribe a handler to a specific message type.
func (s *subscriptions) Subscribe(t message.MessageTypeEnum, h MessageHandler) {
	s.Lock()
	if s.s == nil {
		s.s = map[message.MessageTypeEnum]MessageHandler{}
	}
	s.s[t] = h
	defer s.Unlock()
}

// HandleMessage runs the registered handler according to the type of the message given.
// Expect a unassignedHandlerErr if subscriptions has not a handler registered.
func (s *subscriptions) HandleMessage(m *message.Message) error {
	s.RLock()
	h, ok := s.s[m.MessageHeader.MessageType]
	s.RUnlock()
	if !ok {
		return errors.Wrap(
			UnassignedHandlerErr,
			fmt.Sprintf("type %s", m.MessageHeader.MessageType.String()))
	}
	return h(m)
}
