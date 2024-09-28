package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Config interface {
	GetDSN() string // Data Source Name
}

// NewConnect create new connection to DB.
func NewConnect(ctx context.Context, cfg Config) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, cfg.GetDSN())
	if err != nil {
		return nil, fmt.Errorf("pgxpool.New() err: %w", err)
	}

	// Проверка соединения
	err = pool.Ping(ctx)
	if err != nil {
		pool.Close()
		return nil, fmt.Errorf("pool.Ping() err: %w", err)
	}

	return pool, nil
}
