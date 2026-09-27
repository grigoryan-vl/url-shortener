package service

import (
	"encoding/json"
	"errors"
	"os"
	"sync"

	"github.com/google/uuid"

	generatecodeservice "github.com/grigoryan-vl/URL-shortener/internal/service/generate-code"
)

type fileRecord struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type URLService struct {
	URLStorage map[string]fileRecord
	rwMutex    sync.RWMutex
	filePath   string
}

const codeLengthByExample = 8

func NewURLService(filePath string) (*URLService, error) {

	srv := &URLService{
		URLStorage: make(map[string]fileRecord),
		rwMutex:    sync.RWMutex{},
		filePath:   filePath,
	}

	if err := srv.load(); err != nil {
		return nil, err
	}

	return srv, nil
}

func (srv *URLService) CreateShortURL(URL string) (string, error) {
	for {
		newCode, err := generatecodeservice.GenerateBase62RandomCode(codeLengthByExample)
		if err != nil {
			return "", err
		}

		srv.rwMutex.Lock()
		if _, exists := srv.URLStorage[newCode]; !exists {
			srv.URLStorage[newCode] = fileRecord{UUID: uuid.New().String(), ShortURL: newCode, OriginalURL: URL}
			err = srv.save()
			if err != nil {
				srv.rwMutex.Unlock()
				return "", err
			}
			srv.rwMutex.Unlock()
			return newCode, nil
		}
		srv.rwMutex.Unlock()
	}
}

func (srv *URLService) GetURLByID(id string) (string, error) {
	srv.rwMutex.RLock()
	defer srv.rwMutex.RUnlock()

	URL, ok := srv.URLStorage[id]
	if !ok {
		return "", errors.New("this key does not exist")
	}

	return URL.OriginalURL, nil
}

func (srv *URLService) save() error {
	recs := make([]fileRecord, 0, len(srv.URLStorage))
	for _, url := range srv.URLStorage {
		recs = append(recs, url)
	}
	data, err := json.Marshal(recs)
	if err != nil {
		return err
	}
	return os.WriteFile(srv.filePath, data, 0o644)
}

func (srv *URLService) load() error {
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
		srv.URLStorage[r.ShortURL] = r
	}
	return nil
}
