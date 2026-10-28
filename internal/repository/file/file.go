package file

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"sync"

	"github.com/google/uuid"

	"github.com/grigoryan-vl/URL-shortener/internal/model/storage"
)

type fileRecord struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type Storage struct {
	urlStorage map[string]fileRecord
	rwMutex    sync.RWMutex
	filePath   string
}

func New(filePath string) (*Storage, error) {
	srv := &Storage{
		urlStorage: make(map[string]fileRecord),
		filePath:   filePath,
	}

	if err := srv.load(); err != nil {
		return nil, err
	}

	return srv, nil
}

func (srv *Storage) Save(_ context.Context, short, original string) (string, error) {
	srv.rwMutex.Lock()
	defer srv.rwMutex.Unlock()

	for k, v := range srv.urlStorage {
		if v.OriginalURL == original {
			return k, storage.ErrConflict
		}
	}
	if _, exists := srv.urlStorage[short]; exists {
		return "", storage.ErrShortConflict
	}

	srv.urlStorage[short] = fileRecord{
		UUID:        uuid.New().String(),
		ShortURL:    short,
		OriginalURL: original,
	}
	if err := srv.save(); err != nil {
		return "", err
	}
	return short, nil
}

func (srv *Storage) SaveBatch(_ context.Context, items []storage.BatchItem) error {
	srv.rwMutex.Lock()
	defer srv.rwMutex.Unlock()

	for _, it := range items {
		if _, exists := srv.urlStorage[it.Short]; exists {
			return storage.ErrConflict
		}
	}
	for _, it := range items {
		srv.urlStorage[it.Short] = fileRecord{
			UUID:        uuid.New().String(),
			ShortURL:    it.Short,
			OriginalURL: it.Original,
		}
	}
	return srv.save() // один раз, а не на каждую запись
}

func (srv *Storage) Get(_ context.Context, short string) (string, error) {
	srv.rwMutex.RLock()
	defer srv.rwMutex.RUnlock()

	r, ok := srv.urlStorage[short]
	if !ok {
		return "", storage.ErrNotFound
	}

	return r.OriginalURL, nil
}

func (srv *Storage) Close() error { return nil }

func (srv *Storage) save() error {
	recs := make([]fileRecord, 0, len(srv.urlStorage))
	for _, url := range srv.urlStorage {
		recs = append(recs, url)
	}
	data, err := json.Marshal(recs)
	if err != nil {
		return err
	}
	return os.WriteFile(srv.filePath, data, 0o644)
}

func (srv *Storage) load() error {
	if srv.filePath == "" {
		return nil
	}

	data, err := os.ReadFile(srv.filePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	if len(data) == 0 {
		return nil
	}

	var records []fileRecord
	if err := json.Unmarshal(data, &records); err != nil {
		return err
	}

	for _, r := range records {
		srv.urlStorage[r.ShortURL] = r
	}
	return nil
}
