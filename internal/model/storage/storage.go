package storage

import (
	"context"
	"errors"
)

var (
	ErrNotFound = errors.New("url not found")
	ErrConflict = errors.New("short code already exists")
)

type Storage interface {
	Save(ctx context.Context, short, original string) error
	Get(ctx context.Context, short string) (string, error)
	Close() error
}
