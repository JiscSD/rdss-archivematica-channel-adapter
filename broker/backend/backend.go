package backend

import (
	"fmt"
	"strings"
	"time"
	"github.com/cenkalti/backoff"
	log "github.com/sirupsen/logrus"
)

// Backend is a low-level interface used to interact with RDSS brokers.
type Backend interface {
	Publish(topic string, data []byte) error

	// Subscribe associates a message handler function to a particular topic.
	// The backend must handle the messages to all the subscribers. The message
	// must only be regarded as consumed when all the handlers have returned and
	// none of them returned a non-nil value.
	Subscribe(topic string, cb Handler)

	// Verify the availability of a topic.
	Check(topic string) error

	Close() error
	
	SetLogger(logger log.FieldLogger)
}

// Handler is a message handler for a particular topic.
type Handler func([]byte) error

// Constructor is a function that initializes and returns a Backend
// implementation with the given options.
type Constructor func(*Opts) (Backend, error)

var registration = make(map[string]Constructor)

// Opts holds configuration for the broker backend.
// It is meant to be used by implementations of Storage
type Opts struct {
	Opts map[string]string
}

type MessagePublisher interface {
	sendMessage() error
	canRetry(err error) bool
}

// DialOpts is a daisy-chaining mechanism for setting options to a backend
// during Dial.
type DialOpts func(*Opts) error

// Register register a new broker backend under a name. It is tipically used in
// init functions.
func Register(name string, fn Constructor) error {
	if _, exists := registration[name]; exists {
		return fmt.Errorf("broker backend already exists")
	}
	registration[name] = fn
	return nil
}

// WithOptions parses a string in the format "key1=value1,key2=value2,..." where
// keys and values are specific to each storage backend. Neither key nor value
// may contain the characters "," or "=". Use WithKeyValue repeatedly if these
// characters need to be used.
func WithOptions(options string) DialOpts {
	return func(o *Opts) error {
		pairs := strings.Split(options, ",")
		for _, p := range pairs {
			kv := strings.SplitN(p, "=", 2)
			if len(kv) != 2 {
				return fmt.Errorf("error parsing option %s", kv)
			}
			o.Opts[kv[0]] = kv[1]
		}
		return nil
	}
}

// WithKeyValue sets a key-value pair as option. If called multiple times with
// the same key, the last one wins.
func WithKeyValue(key, value string) DialOpts {
	return func(o *Opts) error {
		o.Opts[key] = value
		return nil
	}
}

// Dial dials the named broker backend using the dial options opts.
func Dial(name string, opts ...DialOpts) (Backend, error) {
	fn, found := registration[name]
	if !found {
		return nil, fmt.Errorf("unknown broker backend type %q", name)
	}
	dOpts := &Opts{Opts: make(map[string]string)}
	var err error
	for _, o := range opts {
		if o != nil {
			err = o(dOpts)
			if err != nil {
				return nil, err
			}
		}
	}
	return fn(dOpts)
}

//
// Function that may be used by Backend implementations to provide backoff
// and retry for network problems.
func Publish(publishFunc func() error, canRetry func(err error) bool) error {
		
	retry := backoff.NewExponentialBackOff()
	
	var err error
	for {
		err = publishFunc()
		if err == nil {
			break
		}
		if canRetry(err) {
			duration := retry.NextBackOff()
			if duration == backoff.Stop {
				break;
			}
			time.Sleep(duration)
		} else {
			break
		}
	}
	return err
}