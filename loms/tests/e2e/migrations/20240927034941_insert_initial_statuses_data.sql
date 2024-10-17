-- +goose Up
-- +goose StatementBegin
INSERT INTO statuses (name) VALUES
    ('new'),
    ('awaiting payment'),
    ('failed'),
    ('paid'),
    ('cancelled');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM statuses WHERE name IN ('new', 'awaiting payment', 'failed', 'paid', 'cancelled');
-- +goose StatementEnd
