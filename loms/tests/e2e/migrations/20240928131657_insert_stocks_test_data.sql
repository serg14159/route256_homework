-- +goose Up
-- +goose StatementBegin
INSERT INTO stocks (sku, total_count, reserved) VALUES
    (1, 100, 10),
    (2, 200, 20);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM stocks WHERE sku IN (1, 2);
-- +goose StatementEnd