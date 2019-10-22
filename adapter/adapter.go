package adapter

import (
	"context"

	"github.com/JiscSD/rdss-archivematica-channel-adapter/broker"
	"github.com/JiscSD/rdss-archivematica-channel-adapter/broker/message"
	"github.com/JiscSD/rdss-archivematica-channel-adapter/s3"

	"github.com/sirupsen/logrus"
)

// Archivematica processing configuration preferred by this adapter.
const archivematicaProcessingConfig = "automated"

// Adapter is the core of the adapter.
//
// It uses a broker to subscribe to the queues, receive messages, forward
// operations to Archivematica and return the results. It employs an internal
// storage.
type Adapter struct {
	logger   logrus.FieldLogger
	broker   *broker.Broker
	s3       s3.ObjectStorage
	storage  Storage
	registry *Registry
	ctx      context.Context
	cancel   context.CancelFunc
	stop     chan chan struct{}
}

func New(
	logger logrus.FieldLogger,
	broker *broker.Broker,
	s3 s3.ObjectStorage,
	storage Storage,
	registry *Registry,
) *Adapter {

	c := &Adapter{
		logger:   logger,
		broker:   broker,
		s3:       s3,
		storage:  storage,
		registry: registry,
		stop:     make(chan chan struct{}),
	}

	c.ctx, c.cancel = context.WithCancel(context.Background())

	c.broker.Subscribe(message.MessageTypeEnum_MetadataCreate, c.handleMetadataCreateRequest)
	c.broker.Subscribe(message.MessageTypeEnum_MetadataUpdate, c.handleMetadataUpdateRequest)

	return c
}

func (c *Adapter) Run() {
	go c.broker.Run()
	c.loop()
}

func (c *Adapter) loop() {
	ch := <-c.stop
	c.cancel()
	c.registry.Stop()
	c.broker.Stop()
	close(ch)
}

func (c *Adapter) Stop() {
	ch := make(chan struct{})
	c.stop <- ch
	<-ch
}
