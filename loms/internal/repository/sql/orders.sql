-- name: CreateOrder :one
INSERT INTO orders (id, user_id, status_id)
VALUES (nextval('order_id_manual_seq') + $1, $2, (SELECT id FROM statuses st WHERE st.name = $3))
RETURNING id;

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

-- name: GetAllOrders :many
SELECT o.id, o.user_id, s.name AS status, o.created_at
FROM orders o
JOIN statuses s ON o.status_id = s.id
ORDER BY o.id DESC;