package memory

import (
	"context"
	"sync"

	"github.com/grigoryan-vl/URL-shortener/internal/model/storage"
)

type Storage struct {
	mu   sync.RWMutex
	data map[string]string
}

func New() *Storage {
	return &Storage{data: make(map[string]string)}
}

func (s *Storage) Save(_ context.Context, short, original string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// уже есть такой original?
	for k, v := range s.data {
		if v == original {
			return k, storage.ErrConflict
		}
	}
	if _, exists := s.data[short]; exists {
		return "", storage.ErrShortConflict
	}
	s.data[short] = original
	return short, nil
}

func (s *Storage) SaveBatch(_ context.Context, items []storage.BatchItem) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// сначала проверяем все конфликты — атомарность на уровне мапы
	for _, it := range items {
		if _, exists := s.data[it.Short]; exists {
			return storage.ErrConflict
		}
	}
	for _, it := range items {
		s.data[it.Short] = it.Original
	}
	return nil
}

func (s *Storage) Get(_ context.Context, short string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.data[short]
	if !ok {
		return "", storage.ErrNotFound
	}
	return v, nil
}

func (s *Storage) Close() error { return nil }
