package adapter

import (
	"context"

	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/amclient"
	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/broker"
	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/broker/message"
	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/s3"

	"github.com/sirupsen/logrus"
)

// Archivematica processing configuration preferred by this adapter.
const archivematicaProcessingConfig = "automated"

// Consumer is the core of the adapter.
//
// It uses a broker to subscribe to the queues, receive messages, forward
// operations to Archivematica and return the results. It employs an internal
// storage.
type Consumer struct {
	logger  logrus.FieldLogger
	broker  *broker.Broker
	amc     *amclient.Client
	s3      s3.ObjectStorage
	storage Storage

	ctx    context.Context
	cancel context.CancelFunc
	stop   chan chan struct{}
}

func New(
	logger logrus.FieldLogger,
	broker *broker.Broker,
	amc *amclient.Client,
	s3 s3.ObjectStorage,
	storage Storage) *Consumer {

	c := &Consumer{
		logger:  logger,
		broker:  broker,
		amc:     amc,
		s3:      s3,
		storage: storage,
		stop:    make(chan chan struct{}),
	}

	c.ctx, c.cancel = context.WithCancel(context.Background())

	c.broker.Subscribe(message.MessageTypeEnum_MetadataCreate, c.handleMetadataCreateRequest)
	c.broker.Subscribe(message.MessageTypeEnum_MetadataUpdate, c.handleMetadataUpdateRequest)

	return c
}

func (c *Consumer) Run() {
	go c.broker.Run()
	c.loop()
}

func (c *Consumer) loop() {
	select {
	case ch := <-c.stop:
		c.cancel()
		c.broker.Stop()
		close(ch)
		return
	}
}

func (c *Consumer) Stop() {
	ch := make(chan struct{})
	c.stop <- ch
	<-ch
}
