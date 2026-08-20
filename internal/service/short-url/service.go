package service

import (
	"errors"
	"sync"

	generatecodeservice "github.com/grigoryan-vl/url-shortener/internal/service/generate-code"
)

type UrlService struct {
	urlStorage map[string]string
	mutex      sync.RWMutex
}

const codeLengthByExample = 8

func NewUrlService() *UrlService {
	return &UrlService{
		urlStorage: make(map[string]string),
		mutex:      sync.RWMutex{},
	}
}

func (srv *UrlService) CreateShortUrl(url string) (string, error) {
	code, err := generatecodeservice.GenerateBase62RandomCode(codeLengthByExample)
	if err != nil {
		return "", err
	}

	srv.mutex.Lock()
	defer srv.mutex.Unlock()
	srv.urlStorage[code] = url

	return code, nil
}

func (srv *UrlService) GetUrlById(id string) (string, error) {
	srv.mutex.RLock()
	defer srv.mutex.RUnlock()

	url, ok := srv.urlStorage[id]
	if !ok {
		return "", errors.New("This key does not exist")
	}

	return url, nil
}
