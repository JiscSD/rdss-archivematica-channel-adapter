package backendmock

import (
	"sync"
	"errors"

	log "github.com/sirupsen/logrus"
	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/broker/backend"
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

func (b *BackendImpl) SetLogger(logger log.FieldLogger) {
	b.Logger = logger
}

// Mock implementation to exercise backoff and retry.
type BackendWithRetry struct {
	BackendImpl
	
	retries int
}

func (b *BackendWithRetry) Publish(topic string, data []byte) error {
	return backend.Publish(func() error {   // Send message function
			if b.retries < 3 {
				b.retries = b.retries + 1
				if b.BackendImpl.Logger != nil {
					b.BackendImpl.Logger.Infof("BackendWithRetry backoff")
				}
				return errors.New("Waiting backoff")
			}
			return b.BackendImpl.Publish(topic, data)
		}, func(err error) bool {  // Can retry function
			return true
		})
}
