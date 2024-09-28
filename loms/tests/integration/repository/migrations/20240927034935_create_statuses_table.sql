-- +goose Up
-- +goose StatementBegin
CREATE TABLE statuses (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE statuses;
-- +goose StatementEnd
