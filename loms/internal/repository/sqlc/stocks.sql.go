// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: stocks.sql

package sqlc

import (
	"context"
)

const cancelReservedItems = `-- name: CancelReservedItems :exec
UPDATE stocks
SET reserved = reserved - $2
WHERE sku = $1
`

type CancelReservedItemsParams struct {
	Sku      int32
	Reserved int64
}

func (q *Queries) CancelReservedItems(ctx context.Context, arg *CancelReservedItemsParams) error {
	_, err := q.db.Exec(ctx, cancelReservedItems, arg.Sku, arg.Reserved)
	return err
}

const getAvailableStockBySKU = `-- name: GetAvailableStockBySKU :one
SELECT total_count - reserved AS available
FROM stocks
WHERE sku = $1
`

func (q *Queries) GetAvailableStockBySKU(ctx context.Context, sku int32) (int32, error) {
	row := q.db.QueryRow(ctx, getAvailableStockBySKU, sku)
	var available int32
	err := row.Scan(&available)
	return available, err
}

const getStockBySKU = `-- name: GetStockBySKU :one
SELECT sku, total_count, reserved
FROM stocks
WHERE sku = $1
`

func (q *Queries) GetStockBySKU(ctx context.Context, sku int32) (*Stock, error) {
	row := q.db.QueryRow(ctx, getStockBySKU, sku)
	var i Stock
	err := row.Scan(&i.Sku, &i.TotalCount, &i.Reserved)
	return &i, err
}

const removeReservedItems = `-- name: RemoveReservedItems :exec
UPDATE stocks
SET reserved = reserved - $2, total_count = total_count - $2
WHERE sku = $1
`

type RemoveReservedItemsParams struct {
	Sku      int32
	Reserved int64
}

func (q *Queries) RemoveReservedItems(ctx context.Context, arg *RemoveReservedItemsParams) error {
	_, err := q.db.Exec(ctx, removeReservedItems, arg.Sku, arg.Reserved)
	return err
}

const reserveItems = `-- name: ReserveItems :exec
UPDATE stocks
SET reserved = reserved + $2
WHERE sku = $1
`

type ReserveItemsParams struct {
	Sku      int32
	Reserved int64
}

func (q *Queries) ReserveItems(ctx context.Context, arg *ReserveItemsParams) error {
	_, err := q.db.Exec(ctx, reserveItems, arg.Sku, arg.Reserved)
	return err
}
