package postgres

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	pgxmigrate "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"

	"github.com/grigoryan-vl/URL-shortener/internal/model/storage"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

type Storage struct {
	pool *pgxpool.Pool
}

func New(ctx context.Context, pool *pgxpool.Pool) (*Storage, error) {
	db := stdlib.OpenDBFromPool(pool)
	if err := migrateUp(db); err != nil {
		return nil, err
	}
	return &Storage{pool: pool}, nil
}

func (s *Storage) Save(ctx context.Context, short, original string) error {
	_, err := s.pool.Exec(ctx,
		`INSERT INTO urls (short_url, original_url) VALUES ($1, $2)`,
		short, original)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return storage.ErrConflict
		}
		return err
	}
	return nil
}

func (s *Storage) SaveBatch(ctx context.Context, items []storage.BatchItem) error {
	if len(items) == 0 {
		return nil
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx) // no-op если Commit уже прошёл

	for _, it := range items {
		_, err := tx.Exec(ctx,
			`INSERT INTO urls (short_url, original_url) VALUES ($1, $2)`,
			it.Short, it.Original)
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				return storage.ErrConflict
			}
			return err
		}
	}
	return tx.Commit(ctx)
}

func (s *Storage) Get(ctx context.Context, short string) (string, error) {
	var original string
	err := s.pool.QueryRow(ctx,
		`SELECT original_url FROM urls WHERE short_url = $1`, short).Scan(&original)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", storage.ErrNotFound
	}
	if err != nil {
		return "", err
	}
	return original, nil
}

func (s *Storage) Close() error { return nil } // pool закроет main

func migrateUp(db *sql.DB) error {
	src, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return err
	}
	driver, err := pgxmigrate.WithInstance(db, &pgxmigrate.Config{})
	if err != nil {
		return err
	}
	m, err := migrate.NewWithInstance("iofs", src, "postgres", driver)
	if err != nil {
		return err
	}
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return nil
}
