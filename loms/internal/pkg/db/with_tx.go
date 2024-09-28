package database

import (
	"context"
	"fmt"
	service "route256/loms/internal/service/loms"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// WithTxFunc type of function for performs a function inside a transaction
type WithTxFunc func(ctx context.Context, tx pgx.Tx) error

// TransactionManager
type TransactionManager struct {
	pool *pgxpool.Pool
}

// NewTransactionManager creates new instance TransactionManager.
func NewTransactionManager(pool *pgxpool.Pool) *TransactionManager {
	return &TransactionManager{pool: pool}
}

// WithTx performs a function inside a transaction.
func (tm *TransactionManager) WithTx(ctx context.Context, fn service.WithTxFunc) error {
	tx, err := tm.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("pool.Begin() err: %w", err)
	}

	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(ctx); rbErr != nil {
				err = fmt.Errorf("tx.Rollback() err: %v, original err: %w", rbErr, err)
			}
		}
	}()

	if err := fn(ctx, tx); err != nil {
		return fmt.Errorf("fn(ctx, tx) err: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("tx.Commit() err: %w", err)
	}

	return nil
}
