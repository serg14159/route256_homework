-- +goose Up
-- +goose StatementBegin
CREATE TABLE items (
    id BIGSERIAL PRIMARY KEY,
    order_id BIGINT REFERENCES orders(id) ON DELETE CASCADE,
    sku INTEGER NOT NULL,
    count SMALLINT NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE items;
-- +goose StatementEnd
