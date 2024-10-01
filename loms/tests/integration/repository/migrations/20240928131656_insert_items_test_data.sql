-- +goose Up
-- +goose StatementBegin
INSERT INTO items (order_id, sku, count)
VALUES (1, 1, 5);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM items WHERE order_id IN (1);
-- +goose StatementEnd