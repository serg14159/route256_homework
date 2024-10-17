-- +goose Up
-- +goose StatementBegin
CREATE TABLE stocks (
    sku INTEGER PRIMARY KEY,
    total_count BIGINT NOT NULL,
    reserved BIGINT NOT NULL DEFAULT 0
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE stocks;
-- +goose StatementEnd
