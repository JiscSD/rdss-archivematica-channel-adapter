package backendmock

import (
	"errors"
	"sync"
	"time"

	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/broker/backend"
	log "github.com/sirupsen/logrus"
)

// New returns a backendmock backend.
func New(opts *backend.Opts) (backend.Backend, error) {
	b := &BackendImpl{}

	return b, nil
}

func NewWithRetry(opts *backend.Opts) (backend.Backend, error) {
	b := &BackendWithRetry{}

	return b, nil
}

func init() {
	backend.Register("backendmock", New)
	backend.Register("backendmockretry", NewWithRetry)
}

// BackendImpl is a mock implementation of broker.Backend. It's not safe to use
// from multiple goroutines (concurrent map access going on right now).
type BackendImpl struct {
	Subscriptions []subscription
	mu            sync.RWMutex
	Logger        log.FieldLogger
}

type subscription struct {
	topic string
	cb    backend.Handler
}

var _ backend.Backend = (*BackendImpl)(nil)

func (b *BackendImpl) Publish(topic string, data []byte) error {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for _, s := range b.Subscriptions {
		if topic != "" && s.topic != topic {
			continue
		}
		s.cb(data)
	}
	return nil
}

func (b *BackendImpl) Subscribe(topic string, cb backend.Handler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	subscription := subscription{topic: topic, cb: cb}
	b.Subscriptions = append(b.Subscriptions, subscription)
}

func (b *BackendImpl) Check(topic string) error {
	return nil
}

// Close implements broker.Backend
func (b *BackendImpl) Close() error {
	return nil
}

// Mock implementation to exercise backoff and retry.
type BackendWithRetry struct {
	BackendImpl

	Retries int
}

func (b *BackendWithRetry) Publish(topic string, data []byte) error {
	return backend.Publish(func() error { // Send message function
		if b.Retries < 3 {
			b.Retries = b.Retries + 1
			return errors.New("Waiting backoff")
		}
		return b.BackendImpl.Publish(topic, data)
	}, func(err error) bool { // Can retry function
		return true
	}, &mockBackoff{})
}

type mockBackoff struct {
}

func (mb *mockBackoff) NextBackOff() time.Duration {
	return time.Duration(0)
}
