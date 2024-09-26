package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type Config interface {
	GetDSN() string // Data Source Name
}

// NewConnect create new connection to DB.
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
