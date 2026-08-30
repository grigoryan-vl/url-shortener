package service

import (
	"errors"
	"sync"

	generatecodeservice "github.com/grigoryan-vl/url-shortener/internal/service/generate-code"
)

type UrlService struct {
	urlStorage map[string]string
	rwMutex    sync.RWMutex
}

const codeLengthByExample = 8

func NewUrlService() *UrlService {
	return &UrlService{
		urlStorage: make(map[string]string),
		rwMutex:    sync.RWMutex{},
	}
}

func (srv *UrlService) CreateShortUrl(url string) (string, error) {
	for {
		newCode, err := generatecodeservice.GenerateBase62RandomCode(codeLengthByExample)
		if err != nil {
			return "", err
		}

		srv.rwMutex.Lock()
		if _, exists := srv.urlStorage[newCode]; !exists {
			srv.urlStorage[newCode] = url
			srv.rwMutex.Unlock()
			return newCode, nil
		}

		srv.rwMutex.Unlock()
	}
}

func (srv *UrlService) GetUrlById(id string) (string, error) {
	srv.rwMutex.RLock()
	defer srv.rwMutex.RUnlock()

	url, ok := srv.urlStorage[id]
	if !ok {
		return "", errors.New("This key does not exist")
	}

	return url, nil
}
