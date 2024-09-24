-- name: CreateOrder :one
INSERT INTO orders (user_id, status)
VALUES ($1, $2)
RETURNING id, user_id, status, created_at;

-- name: GetOrderByID :one
SELECT id, user_id, status, created_at
FROM orders
WHERE id = $1;

-- name: SetOrderStatus :exec
UPDATE orders
SET status = $2
WHERE id = $1;

-- name: CreateOrderItem :one
INSERT INTO items (order_id, sku, count)
VALUES ($1, $2, $3)
RETURNING id;

-- name: GetOrderItems :many
SELECT id, order_id, sku, count
FROM items
WHERE order_id = $1;