package ping

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNoDatabase = errors.New("database not configured")

type PingService struct {
	pool *pgxpool.Pool
}

func NewPingService(pool *pgxpool.Pool) *PingService {
	return &PingService{pool: pool}
}

func (s *PingService) Ping(ctx context.Context) error {
	if s.pool == nil {
		return ErrNoDatabase
	}
	return s.pool.Ping(ctx)
}
