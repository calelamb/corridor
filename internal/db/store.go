package db

import (
	"context"
	"errors"
	"fmt"

	"corridor/internal/db/generated"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SourceSummary = generated.PublicSource

type Store struct{ Pool *pgxpool.Pool }

func Open(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, errors.New("database configuration invalid")
	}
	cfg.MaxConns = 10
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, errors.New("database pool unavailable")
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, errors.New("database connection unavailable")
	}
	return pool, nil
}
func (s Store) EventCount(ctx context.Context) (int64, error) {
	count, err := generated.New(s.Pool).EventCount(ctx)
	if err != nil {
		return 0, fmt.Errorf("read public coverage: %w", err)
	}
	return count, nil
}
func (s Store) Sources(ctx context.Context, limit, offset int32) ([]SourceSummary, int64, error) {
	if limit < 1 || limit > 100 || offset < 0 || offset > 100000 {
		return nil, 0, errors.New("pagination out of range")
	}
	tx, err := s.Pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return nil, 0, fmt.Errorf("begin sources snapshot: %w", err)
	}
	defer tx.Rollback(ctx)
	q := generated.New(tx)
	total, err := q.SourceCount(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count public sources: %w", err)
	}
	rows, err := q.Sources(ctx, generated.SourcesParams{Limit: limit, Offset: offset})
	if err != nil {
		return nil, 0, fmt.Errorf("read public sources: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, 0, fmt.Errorf("finish sources snapshot: %w", err)
	}
	return rows, total, nil
}
func (s Store) Check(ctx context.Context) error {
	var healthy bool
	err := s.Pool.QueryRow(ctx, `SELECT
 (SELECT count(*)=3 FROM pg_extension WHERE extname IN ('postgis','h3','h3_postgis'))
 AND to_regclass('public.public_coverage') IS NOT NULL
 AND to_regclass('public.public_sources') IS NOT NULL`).Scan(&healthy)
	if err != nil || !healthy {
		return errors.New("database schema unavailable")
	}
	_, err = s.EventCount(ctx)
	return err
}
