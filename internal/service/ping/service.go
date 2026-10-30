package ping

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PingService struct {
	pool *pgxpool.Pool
}

func NewPingService(pool *pgxpool.Pool) *PingService {
	return &PingService{pool: pool}
}

func (s *PingService) Ping(ctx context.Context) error {
	if s.pool == nil {
		return nil
	}
	return s.pool.Ping(ctx)
}
