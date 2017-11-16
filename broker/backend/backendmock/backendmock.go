package backendmock

import (
	"errors"
	"strconv"
	"sync"
	"time"

	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/broker/backend"
)

// New returns a backendmock backend.
func New(opts *backend.Opts) (backend.Backend, error) {
	b := &BackendImpl{}

	return b, nil
}

func NewWithRetry(opts *backend.Opts) (backend.Backend, error) {
	b := &BackendWithRetry{maxRetries: 3}
	if valueString, ok := opts.Opts["maxRetries"]; ok {
		value, err := strconv.ParseUint(valueString, 0, 0)
		if err != nil {
			return b, err
		}
		b.maxRetries = int(value)
	}
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

// BackendWithRetry is a mock implementation to exercise backoff and retry.
type BackendWithRetry struct {
	BackendImpl

	maxRetries int
	Retries    int
}

func (b *BackendWithRetry) Publish(topic string, data []byte) error {
	return backend.Publish(func() error { // Send message function
		if b.Retries < b.maxRetries {
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
