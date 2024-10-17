-- +goose Up
-- +goose StatementBegin
INSERT INTO orders (user_id, status_id)
VALUES (1, (SELECT id FROM statuses st WHERE st.name = 'new'));
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM orders WHERE user_id = 1;
-- +goose StatementEnd