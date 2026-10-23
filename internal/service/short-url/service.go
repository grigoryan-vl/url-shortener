package service

import (
	"context"
	"errors"

	"github.com/grigoryan-vl/URL-shortener/internal/model/storage"
	generatecodeservice "github.com/grigoryan-vl/URL-shortener/internal/service/generate-code"
)

const codeLengthByExample = 8
const maxGenerateAttempts = 10

type URLService struct {
	storage storage.Storage
}

func NewURLService(s storage.Storage) *URLService {
	return &URLService{storage: s}
}

func (srv *URLService) CreateShortURL(ctx context.Context, URL string) (string, error) {
	for i := 0; i < maxGenerateAttempts; i++ {
		code, err := generatecodeservice.GenerateBase62RandomCode(codeLengthByExample)
		if err != nil {
			return "", err
		}

		existing, err := srv.storage.Save(ctx, code, URL)
		if errors.Is(err, storage.ErrConflict) {
			return existing, storage.ErrConflict
		}
		if errors.Is(err, storage.ErrShortConflict) {
			continue
		}
		if err != nil {
			return "", err
		}
		return code, nil
	}
	return "", errors.New("failed to generate unique code")
}

func (srv *URLService) CreateShortURLBatch(ctx context.Context, urls []string) ([]string, error) {
	codes := make([]string, len(urls))
	items := make([]storage.BatchItem, len(urls))

	for i, u := range urls {
		code, err := generatecodeservice.GenerateBase62RandomCode(codeLengthByExample)
		if err != nil {
			return nil, err
		}
		codes[i] = code
		items[i] = storage.BatchItem{Short: code, Original: u}
	}

	if err := srv.storage.SaveBatch(ctx, items); err != nil {
		return nil, err
	}
	return codes, nil
}

func (srv *URLService) GetURLByID(ctx context.Context, id string) (string, error) {
	return srv.storage.Get(ctx, id)
}
