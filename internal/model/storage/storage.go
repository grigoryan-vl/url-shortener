package storage

import (
	"context"
	"errors"
)

var (
	ErrNotFound      = errors.New("url not found")
	ErrConflict      = errors.New("original url already exists")
	ErrShortConflict = errors.New("short code already exists")
)

type BatchItem struct {
	Short    string
	Original string
}

type Storage interface {
	Save(ctx context.Context, short, original string) (string, error)
	Get(ctx context.Context, short string) (string, error)
	SaveBatch(ctx context.Context, items []BatchItem) error
	Close() error
}
