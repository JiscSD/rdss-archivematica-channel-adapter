package broker

import (
	"reflect"
	"runtime"
	"testing"

	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/broker/message"

	"github.com/stretchr/testify/require"
)

func requireEqualHandlers(t *testing.T, x, y interface{}) {
	t.Helper()
	require.Equal(t,
		runtime.FuncForPC(reflect.ValueOf(x).Pointer()).Name(),
		runtime.FuncForPC(reflect.ValueOf(y).Pointer()).Name(),
	)
}

func TestSubscriptionsSubscribe(t *testing.T) {
	s := subscriptions{s: map[message.MessageTypeEnum]MessageHandler{}}

	metadataCreateHandler := func(m *message.Message) error { return nil }
	s.Subscribe(message.MessageTypeEnum_MetadataCreate, metadataCreateHandler)

	preservationEventHandler := func(m *message.Message) error { return nil }
	s.Subscribe(message.MessageTypeEnum_PreservationEvent, preservationEventHandler)

	require.Len(t, s.s, 2)
	requireEqualHandlers(t, metadataCreateHandler, s.s[message.MessageTypeEnum_MetadataCreate])
	requireEqualHandlers(t, preservationEventHandler, s.s[message.MessageTypeEnum_PreservationEvent])
}

func TestSubscriptionsHandleMessage_NotFound(t *testing.T) {
	s := subscriptions{s: map[message.MessageTypeEnum]MessageHandler{}}

	err := s.handleMessage(&message.Message{})

	require.Error(t, err)
}

func TestSubscriptionsHandleMessage_Found(t *testing.T) {
	var (
		executed bool
		count    int
		s        = subscriptions{s: map[message.MessageTypeEnum]MessageHandler{}}
	)
	s.Subscribe(message.MessageTypeEnum_MetadataCreate, func(m *message.Message) error {
		count++
		executed = true
		return nil
	})

	err1 := s.handleMessage(&message.Message{})
	err2 := s.handleMessage(&message.Message{})

	require.NoError(t, err1)
	require.NoError(t, err2)
	require.Equal(t, executed, true)
	require.Equal(t, count, 2)
}
