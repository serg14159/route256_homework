-- name: CreateOrder :one
INSERT INTO orders (user_id, status_id)
VALUES ($1, (SELECT id FROM statuses st WHERE st.name = $2))
RETURNING id, user_id, (SELECT name FROM statuses st WHERE st.id = orders.status_id) AS status, created_at;


-- name: GetOrderByID :one
SELECT o.id, o.user_id, s.name AS status, o.created_at
FROM orders o
JOIN statuses s ON o.status_id = s.id
WHERE o.id = $1;

-- name: SetOrderStatus :exec
UPDATE orders
SET status_id = (SELECT id FROM statuses st WHERE st.name = $2)
WHERE orders.id = $1;

-- name: CreateOrderItem :one
INSERT INTO items (order_id, sku, count)
VALUES ($1, $2, $3)
RETURNING id;

-- name: GetOrderItems :many
SELECT id, order_id, sku, count
FROM items
WHERE order_id = $1;