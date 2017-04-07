package broker

import (
	"context"

	log "github.com/Sirupsen/logrus"
)

type BrokerAPI interface {
	ReadMetadata(ctx context.Context) error
	CreateMetadata(ctx context.Context) error
	UpdateMetadata(ctx context.Context) error
	DeleteMetadata(ctx context.Context) error
}

type Broker struct {
	backend Backend
	logger  *log.Entry
}

func New(backend Backend, logger *log.Entry) BrokerAPI {
	return &Broker{
		backend: backend,
		logger:  logger,
	}
}

func (b Broker) ReadMetadata(ctx context.Context) error {
	return b.backend.Request(ctx, &Message{})
}

func (b Broker) CreateMetadata(ctx context.Context) error {
	return b.backend.Request(ctx, &Message{})
}

func (b Broker) UpdateMetadata(ctx context.Context) error {
	return b.backend.Request(ctx, &Message{})
}

func (b Broker) DeleteMetadata(ctx context.Context) error {
	return b.backend.Request(ctx, &Message{})
}
