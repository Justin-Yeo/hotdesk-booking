-- +goose Up
-- +goose StatementBegin
-- Create sessions table for refresh token storage
CREATE TABLE IF NOT EXISTS sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    refresh_token VARCHAR(255) NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create index on user_id for efficient user session lookups
CREATE INDEX idx_sessions_user_id ON sessions(user_id);

-- Create index on refresh_token for token validation
CREATE INDEX idx_sessions_refresh_token ON sessions(refresh_token);

-- Create index on expires_at for cleanup queries
CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Drop indexes
DROP INDEX IF EXISTS idx_sessions_expires_at;
DROP INDEX IF EXISTS idx_sessions_refresh_token;
DROP INDEX IF EXISTS idx_sessions_user_id;

-- Drop sessions table
DROP TABLE IF EXISTS sessions;
-- +goose StatementEnd
