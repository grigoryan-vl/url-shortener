package memory

import (
	"context"
	"sync"

	"github.com/grigoryan-vl/URL-shortener/internal/model/storage"
)

type record struct {
	Original string
	UserID   string
}

type Storage struct {
	mu   sync.RWMutex
	data map[string]record
}

func New() *Storage {
	return &Storage{data: make(map[string]record)}
}

func (s *Storage) Save(_ context.Context, short, original, userID string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for k, v := range s.data {
		if v.Original == original {
			return k, storage.ErrConflict
		}
	}
	if _, exists := s.data[short]; exists {
		return "", storage.ErrShortConflict
	}
	s.data[short] = record{Original: original, UserID: userID}
	return short, nil
}

func (s *Storage) SaveBatch(_ context.Context, items []storage.BatchItem, userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, it := range items {
		if _, exists := s.data[it.Short]; exists {
			return storage.ErrShortConflict
		}
	}
	for _, it := range items {
		s.data[it.Short] = record{Original: it.Original, UserID: userID}
	}
	return nil
}

func (s *Storage) Get(_ context.Context, short string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, ok := s.data[short]
	if !ok {
		return "", storage.ErrNotFound
	}
	return r.Original, nil
}

func (s *Storage) GetByUserID(_ context.Context, userID string) ([]storage.UserURL, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []storage.UserURL
	for short, r := range s.data {
		if r.UserID == userID {
			result = append(result, storage.UserURL{Short: short, Original: r.Original})
		}
	}
	return result, nil
}

func (s *Storage) Close() error { return nil }
