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
	"github.com/jackc/pgerrcode"
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

func (s *Storage) Save(ctx context.Context, short, original, userID string) (string, error) {
	var inserted string
	err := s.pool.QueryRow(ctx,
		`INSERT INTO urls (short_url, original_url, user_id) VALUES ($1, $2, $3)
		 ON CONFLICT (original_url) DO NOTHING
		 RETURNING short_url`, short, original, userID).Scan(&inserted)

	switch {
	case err == nil:
		return inserted, nil

	case errors.Is(err, pgx.ErrNoRows):
		// конфликт по original_url — достаём существующий код
		var existing string
		if err := s.pool.QueryRow(ctx,
			`SELECT short_url FROM urls WHERE original_url = $1`, original).Scan(&existing); err != nil {
			return "", err
		}
		return existing, storage.ErrConflict
	}

	// сюда попадаем при конфликте по PK short_url
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
		return "", storage.ErrShortConflict
	}
	return "", err
}

func (s *Storage) SaveBatch(ctx context.Context, items []storage.BatchItem, userID string) error {
	if len(items) == 0 {
		return nil
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	for _, it := range items {
		_, err := tx.Exec(ctx,
			`INSERT INTO urls (short_url, original_url, user_id) VALUES ($1, $2, $3)
			 ON CONFLICT (original_url) DO NOTHING`,
			it.Short, it.Original, userID)
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
				return storage.ErrShortConflict
			}
			return err
		}
	}
	return tx.Commit(ctx)
}

func (s *Storage) Get(ctx context.Context, short string) (string, error) {
	var original string
	var isDeleted bool
	err := s.pool.QueryRow(ctx,
		`SELECT original_url, is_deleted FROM urls WHERE short_url = $1`, short).
		Scan(&original, &isDeleted)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", storage.ErrNotFound
	}
	if err != nil {
		return "", err
	}
	if isDeleted {
		return "", storage.ErrDeleted
	}
	return original, nil
}

func (s *Storage) DeleteBatch(ctx context.Context, userID string, shorts []string) error {
	if len(shorts) == 0 {
		return nil
	}
	_, err := s.pool.Exec(ctx,
		`UPDATE urls SET is_deleted = true
		 WHERE user_id = $1 AND short_url = ANY($2)`,
		userID, shorts)
	return err
}

func (s *Storage) GetByUserID(ctx context.Context, userID string) ([]storage.UserURL, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT short_url, original_url FROM urls
	 WHERE user_id = $1 AND is_deleted = false
	 ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []storage.UserURL
	for rows.Next() {
		var u storage.UserURL
		if err := rows.Scan(&u.Short, &u.Original); err != nil {
			return nil, err
		}
		result = append(result, u)
	}
	return result, rows.Err()
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
