package consumer

import (
	"context"
	"sync"
)

type storage interface {
	AssociateResearchObject(ctx context.Context, objectUUID string, transferID string) error
	GetResearchObject(ctx context.Context, objectUUID string) (string, error)
}

var _ storage = &storageInMemoryImpl{}

func newStorageInMemory() *storageInMemoryImpl {
	return &storageInMemoryImpl{
		h: make(map[string]string),
	}
}

type storageInMemoryImpl struct {
	h map[string]string
	sync.RWMutex
}

func (s *storageInMemoryImpl) AssociateResearchObject(ctx context.Context, objectUUID string, transferID string) error {
	s.Lock()
	defer s.Unlock()
	s.h[objectUUID] = transferID
	return nil
}

func (s *storageInMemoryImpl) GetResearchObject(ctx context.Context, objectUUID string) (string, error) {
	s.RLock()
	defer s.RUnlock()
	ret, ok := s.h[objectUUID]
	if !ok {
		return "", nil
	}
	return ret, nil
}
