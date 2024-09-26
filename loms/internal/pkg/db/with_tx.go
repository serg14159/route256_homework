package database

import (
	"context"
	"fmt"
	service "route256/loms/internal/service/loms"

	"github.com/jackc/pgx/v5"
)

// WithTxFunc type of function for performs a function inside a transaction
type WithTxFunc func(ctx context.Context, tx *pgx.Tx) error

// TransactionManager
type TransactionManager struct {
	conn *pgx.Conn
}

// NewTransactionManager creates new instance TransactionManager.
func NewTransactionManager(conn *pgx.Conn) *TransactionManager {
	return &TransactionManager{conn: conn}
}

// WithTx performs a function inside a transaction.
func (tm *TransactionManager) WithTx(ctx context.Context, fn service.WithTxFunc) error {
	tx, err := tm.conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("conn.Begin() err: %w", err)
	}

	if err := fn(ctx, &tx); err != nil {
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
