// Package broker implements a low-level interface for communicating with RDSS
// via message brokers such Amazon Kinesis Stream or RabbitMQ.
//
// The goal of this package is to be reusable by any user. It's not coupled to
// Archivematica.
package broker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"

	log "github.com/sirupsen/logrus"

	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/broker/backend"
	bErrors "github.com/JiscRDSS/rdss-archivematica-channel-adapter/broker/errors"
	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/broker/message"
)

// Broker is a broker client conforming to the RDSS messaging API.
type Broker struct {
	backend backend.Backend
	logger  log.FieldLogger
	config  *Config

	Metadata   MetadataService
	Vocabulary VocabularyService

	repository Repository

	// Number of messages received.
	count uint64

	// List of subscribers
	subs []subscription
	mu   sync.RWMutex
}

// subscription is a subscription with a handler to a particular topic. If the
// value of all is true then the handler will be invoked for every message.
type subscription struct {
	// Whether the subscriber wants to be subscribed to receive every message.
	all bool

	// The type of message that this subscriber is listening.
	mType message.MessageType

	// The callback associated to this particular subscriber.
	cb MessageHandler
}

// MessageHandler is a callback function supplied by subscribers.
type MessageHandler func(msg *message.Message) error

// New returns a new Broker.
func New(backend backend.Backend, logger log.FieldLogger, config *Config) (*Broker, error) {
	if logger == nil {
		logger = log.New()
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("Configuration error: %s", err)
	}

	b := &Broker{
		backend: backend,
		logger:  logger,
		config:  config,
	}
	if config.RepositoryConfig != nil {
		b.repository = MustRepository(NewRepository(config.RepositoryConfig))
	}
	b.Metadata = &MetadataServiceOp{broker: b}
	b.Vocabulary = &VocabularyServiceOp{broker: b}

	// Check queues
	if err := b.checkQueues(); err != nil {
		return nil, err
	}

	b.backend.SetLogger(logger)

	// Set up message router
	b.backend.Subscribe(config.QueueMain, b.messageHandler)
	return b, nil
}

// messageHandler implents backend.Handler. It runs in a separate goroutine.
func (b *Broker) messageHandler(data []byte) error {
	msg := &message.Message{}
	err := json.Unmarshal(data, msg)
	if err != nil {
		b.queueInvalidMessage(data, bErrors.NewWithError(bErrors.GENERR001, err))
		return nil
	}

	if b.exists(msg) {
		b.logger.Debugf("Message discarded (already seen): %s", msg.ID())
		return nil
	}

	atomic.AddUint64(&b.count, 1)
	b.logger.WithFields(log.Fields{"count": b.Count(), "type": msg.MessageHeader.MessageType}).Infoln("Message received")

	// Dispatch the message to its subscribers in parallel. Block until they're
	// all done and handle
	var (
		wg       sync.WaitGroup
		appError error
	)
	b.mu.RLock()
	defer b.mu.RUnlock()
	for _, s := range b.subs {
		if !s.all && msg.MessageHeader.MessageType != s.mType {
			continue
		}
		// Here is our goroutine that is going to dispatch the message to each
		// subscriber.
		wg.Add(1)
		go func(s subscription) {
			defer wg.Done()
			// We want to regain control of a panicking goroutine if that was
			// the case. The idea is not to let the application crash if the
			// user making use of this package implemented a faulty handler.
			defer func() {
				if r := recover(); r != nil {
					err = fmt.Errorf("recovered from panic in subscriber: %s", err)
				}
			}()
			if handlerErr := s.cb(msg); handlerErr != nil {
				appError = handlerErr
			}
		}(s)
	}

	// Wait until all the goroutines are done
	wg.Wait()

	if appError != nil {
		b.queueErrorMessage(msg, appError)
	}

	return err
}

// exists returns whether the message is already in the repository. As a side
// effect, the message is cached in the repo when it wasn't there so the second
// time this function is called for the same message the returned value should
// be true.
func (b *Broker) exists(msg *message.Message) bool {
	if item := b.repository.Get(msg.ID()); item != nil {
		return true
	}
	if err := b.repository.Put(msg); err != nil {
		b.logger.Error("Error trying to put the message in the local repository:", msg.ID())
	}
	return false
}

// checkQueues verifies access and availability of the queues being used.
func (b *Broker) checkQueues() (err error) {
	queues := []string{b.config.QueueMain, b.config.QueueError, b.config.QueueInvalid}
	for _, queue := range queues {
		if err = b.backend.Check(queue); err != nil {
			return err
		}
	}
	return
}

// handleErrMessage places a message on the Error Message Queue or the Invalid
// Message Queue. The message is tagged with error headers before published.
func (b *Broker) handleErrMessage(in interface{}, e error, queue string) {
	if queue == "" {
		return
	}

	logf := b.logger.WithFields(log.Fields{"messageId": "unknown", "target": queue, "err": e.Error()})
	bErr, ok := e.(*bErrors.Error)
	if ok {
		logf = logf.WithFields(log.Fields{"code": bErr.Kind, "err": bErr.Err.Error()})
	}

	var (
		data []byte
		err  error
	)
	switch msg := in.(type) {
	case *message.Message:
		logf = logf.WithFields(log.Fields{"messageId": msg.ID()})
		msg.TagError(bErr)
		data, err = msg.MarshalJSON()
		if err != nil {
			logf.Error("Error encoding message:", err)
			return
		}
	case []byte:
		data = msg
	default:
		logf.Error("Message ignored because its type is not recognized")
		return
	}

	logf.Warnf("A message is being placed on the Error Message Queue (%s)", queue)
	if err = b.backend.Publish(queue, data); err != nil {
		logf.Error("The message could not be published: ", err)
	}
}

func (b *Broker) queueInvalidMessage(in interface{}, bErr error) {
	b.handleErrMessage(in, bErr, b.config.QueueInvalid)
}

func (b *Broker) queueErrorMessage(in interface{}, bErr error) {
	b.handleErrMessage(in, bErr, b.config.QueueError)
}

// Subscribe creates a new subscription associated to every message received.
func (b *Broker) Subscribe(cb MessageHandler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.subs = append(b.subs, subscription{all: true, cb: cb})
}

// SubscribeType creates a new subscription associated to a particular message type.
func (b *Broker) SubscribeType(t message.MessageType, cb MessageHandler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.subs = append(b.subs, subscription{mType: t, cb: cb})
}

// Request sends a fire-and-forget request to RDSS.
func (b *Broker) Request(_ context.Context, msg *message.Message) error {
	data, err := msg.MarshalJSON()
	if err != nil {
		return err
	}
	return b.backend.Publish(b.config.QueueMain, data)
}

// RequestResponse sends a request and waits until a response is received.
func (b *Broker) RequestResponse(context.Context, *message.Message) (*message.Message, error) {
	return nil, errors.New("not implemented yet")
}

// Count returns the total number of messages received by this broker since it
// was executed.
func (b *Broker) Count() uint64 {
	return atomic.LoadUint64(&b.count)
}

func (b *Broker) Close() error {
	return b.backend.Close()
}
