-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS migration_test (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS migration_test;
-- +goose StatementEnd
