package repository

import (
	"context"
	"encoding/json"
	"route256/loms/internal/models"
	"route256/loms/internal/repository/sqlc"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OutboxRepository struct {
	queries sqlc.Querier
	pool    *pgxpool.Pool
}

// NewOutboxRepository creates new instance of OutboxRepository.
func NewOutboxRepository(pool *pgxpool.Pool) *OutboxRepository {
	return &OutboxRepository{
		queries: sqlc.New(pool),
		pool:    pool,
	}
}

// CreateEvent inserts new outbox event into database.
func (r *OutboxRepository) CreateEvent(ctx context.Context, tx pgx.Tx, eventType string, payload interface{}) error {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	q := r.getQuerier(tx)

	_, err = q.InsertOutboxEvent(ctx, &sqlc.InsertOutboxEventParams{
		EventType: eventType,
		Payload:   string(payloadBytes),
	})

	return err
}

// FetchNextMsg retrieves next unprocessed outbox event from database.
func (r *OutboxRepository) FetchNextMsg(ctx context.Context, tx pgx.Tx) (*models.OutboxEvent, error) {
	q := r.getQuerier(tx)

	row, err := q.FetchNextOutboxEvent(ctx)
	if err != nil {
		return nil, err
	}

	event := &models.OutboxEvent{
		ID:        int64(row.ID),
		EventType: row.EventType,
		Payload:   row.Payload,
	}

	return event, nil
}

// MarkAsSent marks outbox event as processed in database.
func (r *OutboxRepository) MarkAsSent(ctx context.Context, tx pgx.Tx, eventID int64) error {
	q := r.getQuerier(tx)
	return q.MarkOutboxEventAsProcessed(ctx, int32(eventID))
}

// getQuerier returns sqlc.Querier based on provided transaction.
func (r *OutboxRepository) getQuerier(tx pgx.Tx) sqlc.Querier {
	if tx != nil {
		return sqlc.New(tx)
	}
	return r.queries
}
