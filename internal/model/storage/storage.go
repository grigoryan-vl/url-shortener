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

type UserURL struct {
	Short    string
	Original string
}

type Storage interface {
	Save(ctx context.Context, short, original, userID string) (string, error)
	Get(ctx context.Context, short string) (string, error)
	GetByUserID(ctx context.Context, userID string) ([]UserURL, error)
	SaveBatch(ctx context.Context, items []BatchItem, userID string) error
	Close() error
}
