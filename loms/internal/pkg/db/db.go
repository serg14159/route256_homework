package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type Config interface {
	GetDSN() string // Data Source Name
}

// Function NewConnect create new connection to DB
func NewConnect(ctx context.Context, cfg Config) (*pgx.Conn, error) {
	conn, err := pgx.Connect(ctx, cfg.GetDSN())
	if err != nil {
		return nil, fmt.Errorf("pgx.Connect() err: %w", err)
	}

	// Check connect
	err = conn.Ping(ctx)
	if err != nil {
		conn.Close(ctx)
		return nil, fmt.Errorf("conn.Ping() err: %w", err)
	}

	return conn, nil
}

// WithTxFunc type of function for performs a function inside a transaction
type WithTxFunc func(ctx context.Context, tx pgx.Tx) error

// Function WithTx performs a function inside a transaction
func WithTx(ctx context.Context, conn *pgx.Conn, fn WithTxFunc) error {
	tx, err := conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("conn.Begin() err: %w", err)
	}

	if err := fn(ctx, tx); err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return fmt.Errorf("tx.Rollback() err: %w", err)
		}
		return fmt.Errorf("fn(ctx, tx) err: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("tx.Commit() err: %w", err)
	}

	return nil
}
