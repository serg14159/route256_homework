// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: orders.sql

package sqlc

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const createOrder = `-- name: CreateOrder :one
INSERT INTO orders (id, user_id, status_id)
VALUES (nextval('order_id_manual_seq') + $1, $2, (SELECT id FROM statuses st WHERE st.name = $3))
RETURNING id
`

type CreateOrderParams struct {
	Column1 interface{}
	UserID  int64
	Name    string
}

func (q *Queries) CreateOrder(ctx context.Context, arg *CreateOrderParams) (int64, error) {
	row := q.db.QueryRow(ctx, createOrder, arg.Column1, arg.UserID, arg.Name)
	var id int64
	err := row.Scan(&id)
	return id, err
}

const createOrderItem = `-- name: CreateOrderItem :one
INSERT INTO items (order_id, sku, count)
VALUES ($1, $2, $3)
RETURNING id
`

type CreateOrderItemParams struct {
	OrderID *int64
	Sku     int32
	Count   int16
}

func (q *Queries) CreateOrderItem(ctx context.Context, arg *CreateOrderItemParams) (int64, error) {
	row := q.db.QueryRow(ctx, createOrderItem, arg.OrderID, arg.Sku, arg.Count)
	var id int64
	err := row.Scan(&id)
	return id, err
}

const getAllOrders = `-- name: GetAllOrders :many
SELECT o.id, o.user_id, s.name AS status, o.created_at
FROM orders o
JOIN statuses s ON o.status_id = s.id
ORDER BY o.id DESC
`

type GetAllOrdersRow struct {
	ID        int64
	UserID    int64
	Status    string
	CreatedAt pgtype.Timestamptz
}

func (q *Queries) GetAllOrders(ctx context.Context) ([]*GetAllOrdersRow, error) {
	rows, err := q.db.Query(ctx, getAllOrders)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*GetAllOrdersRow
	for rows.Next() {
		var i GetAllOrdersRow
		if err := rows.Scan(
			&i.ID,
			&i.UserID,
			&i.Status,
			&i.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, &i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getOrderByID = `-- name: GetOrderByID :one
SELECT o.id, o.user_id, s.name AS status, o.created_at
FROM orders o
JOIN statuses s ON o.status_id = s.id
WHERE o.id = $1
`

type GetOrderByIDRow struct {
	ID        int64
	UserID    int64
	Status    string
	CreatedAt pgtype.Timestamptz
}

func (q *Queries) GetOrderByID(ctx context.Context, id int64) (*GetOrderByIDRow, error) {
	row := q.db.QueryRow(ctx, getOrderByID, id)
	var i GetOrderByIDRow
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.Status,
		&i.CreatedAt,
	)
	return &i, err
}

const getOrderItems = `-- name: GetOrderItems :many
SELECT id, order_id, sku, count
FROM items
WHERE order_id = $1
`

func (q *Queries) GetOrderItems(ctx context.Context, orderID *int64) ([]*Item, error) {
	rows, err := q.db.Query(ctx, getOrderItems, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*Item
	for rows.Next() {
		var i Item
		if err := rows.Scan(
			&i.ID,
			&i.OrderID,
			&i.Sku,
			&i.Count,
		); err != nil {
			return nil, err
		}
		items = append(items, &i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const setOrderStatus = `-- name: SetOrderStatus :exec
UPDATE orders
SET status_id = (SELECT id FROM statuses st WHERE st.name = $2)
WHERE orders.id = $1
`

type SetOrderStatusParams struct {
	ID   int64
	Name string
}

func (q *Queries) SetOrderStatus(ctx context.Context, arg *SetOrderStatusParams) error {
	_, err := q.db.Exec(ctx, setOrderStatus, arg.ID, arg.Name)
	return err
}
