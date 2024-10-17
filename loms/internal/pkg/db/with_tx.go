package database

import (
	"context"
	"fmt"
	service "route256/loms/internal/service/loms"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// WithTxFunc type of function for performs a function inside a transaction.
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
func (tm *TransactionManager) WithTx(ctx context.Context, fn service.WithTxFunc) (err error) {
	tx, err := tm.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("pool.Begin() err: %w", err)
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		} else if err != nil {
			_ = tx.Rollback(ctx)
		} else {
			err = tx.Commit(ctx)
		}
	}()

	err = fn(ctx, tx)
	return err
}
