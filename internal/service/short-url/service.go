package service

import (
	"errors"
	"sync"

	generatecodeservice "github.com/grigoryan-vl/URL-shortener/internal/service/generate-code"
)

type URLService struct {
	URLStorage map[string]string
	rwMutex    sync.RWMutex
}

const codeLengthByExample = 8

func NewURLService() *URLService {
	return &URLService{
		URLStorage: make(map[string]string),
		rwMutex:    sync.RWMutex{},
	}
}

func (srv *URLService) CreateShortURL(URL string) (string, error) {
	for {
		newCode, err := generatecodeservice.GenerateBase62RandomCode(codeLengthByExample)
		if err != nil {
			return "", err
		}

		srv.rwMutex.Lock()
		if _, exists := srv.URLStorage[newCode]; !exists {
			srv.URLStorage[newCode] = URL
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

	return URL, nil
}
