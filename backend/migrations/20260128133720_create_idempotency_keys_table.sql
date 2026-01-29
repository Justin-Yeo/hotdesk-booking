-- +goose Up
-- +goose StatementBegin
-- Create idempotency_keys table for handling duplicate requests
CREATE TABLE IF NOT EXISTS idempotency_keys (
    key VARCHAR(255) PRIMARY KEY,
    response JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create index on created_at for cleanup queries (24-hour TTL)
CREATE INDEX idx_idempotency_keys_created_at ON idempotency_keys(created_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Drop index
DROP INDEX IF EXISTS idx_idempotency_keys_created_at;

-- Drop idempotency_keys table
DROP TABLE IF EXISTS idempotency_keys;
-- +goose StatementEnd
