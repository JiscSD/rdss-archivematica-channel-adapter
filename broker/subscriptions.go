package broker

import (
	"fmt"
	"sync"

	"github.com/JiscSD/rdss-archivematica-channel-adapter/broker/message"
)

// MessageHandler is a function supplied by message subscribers.
type MessageHandler func(msg *message.Message) error

// subscriptions associates message handlers to message types.
type subscriptions struct {
	s map[message.MessageTypeEnum]MessageHandler
	sync.RWMutex
}

// Subscribe a handler to a specific message type.
func (s *subscriptions) Subscribe(t message.MessageTypeEnum, h MessageHandler) {
	s.Lock()
	defer s.Unlock()
	s.s[t] = h
}

// handleMessage runs the registered handler according to the message type.
func (s *subscriptions) handleMessage(m *message.Message) error {
	s.RLock()
	h, ok := s.s[m.MessageHeader.MessageType]
	s.RUnlock()
	if !ok {
		return fmt.Errorf("message handler not registered for type %s", m.MessageHeader.MessageType)
	}
	return h(m)
}
