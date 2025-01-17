// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: outbox.sql

package sqlc

import (
	"context"
)

const fetchNextOutboxEvent = `-- name: FetchNextOutboxEvent :one
SELECT id, event_type, payload
FROM outbox
WHERE processed = FALSE
ORDER BY created_at
LIMIT 1
FOR UPDATE SKIP LOCKED
`

type FetchNextOutboxEventRow struct {
	ID        int32
	EventType string
	Payload   string
}

func (q *Queries) FetchNextOutboxEvent(ctx context.Context) (*FetchNextOutboxEventRow, error) {
	row := q.db.QueryRow(ctx, fetchNextOutboxEvent)
	var i FetchNextOutboxEventRow
	err := row.Scan(&i.ID, &i.EventType, &i.Payload)
	return &i, err
}

const insertOutboxEvent = `-- name: InsertOutboxEvent :one
INSERT INTO outbox (event_type, payload)
VALUES ($1, $2)
RETURNING id, event_type, payload, created_at, processed
`

type InsertOutboxEventParams struct {
	EventType string
	Payload   string
}

func (q *Queries) InsertOutboxEvent(ctx context.Context, arg *InsertOutboxEventParams) (*Outbox, error) {
	row := q.db.QueryRow(ctx, insertOutboxEvent, arg.EventType, arg.Payload)
	var i Outbox
	err := row.Scan(
		&i.ID,
		&i.EventType,
		&i.Payload,
		&i.CreatedAt,
		&i.Processed,
	)
	return &i, err
}

const markOutboxEventAsProcessed = `-- name: MarkOutboxEventAsProcessed :exec
UPDATE outbox
SET processed = TRUE
WHERE id = $1
`

func (q *Queries) MarkOutboxEventAsProcessed(ctx context.Context, id int32) error {
	_, err := q.db.Exec(ctx, markOutboxEventAsProcessed, id)
	return err
}
