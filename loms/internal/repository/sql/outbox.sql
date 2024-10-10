-- name: InsertOutboxEvent :one
INSERT INTO outbox (event_type, payload)
VALUES ($1, $2)
RETURNING id, event_type, payload, created_at, processed;

-- name: FetchNextOutboxEvent :one
SELECT id, event_type, payload
FROM outbox
WHERE processed = FALSE
ORDER BY created_at
LIMIT 1
FOR UPDATE SKIP LOCKED;

-- name: MarkOutboxEventAsProcessed :exec
UPDATE outbox
SET processed = TRUE
WHERE id = $1;