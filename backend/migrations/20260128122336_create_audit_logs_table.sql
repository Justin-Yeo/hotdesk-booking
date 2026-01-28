-- +goose Up
-- +goose StatementBegin
-- Create audit_logs table for tracking entity changes
CREATE TABLE IF NOT EXISTS audit_logs (
    id SERIAL PRIMARY KEY,
    user_id UUID REFERENCES users(id),
    entity_type VARCHAR(50) NOT NULL,
    entity_id INTEGER NOT NULL,
    action VARCHAR(50) NOT NULL,
    changes JSONB,
    metadata JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create index on user_id for user activity queries
CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);

-- Create composite index on entity_type and entity_id for entity history
CREATE INDEX idx_audit_logs_entity ON audit_logs(entity_type, entity_id);

-- Create index on created_at for time-based queries
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Drop indexes
DROP INDEX IF EXISTS idx_audit_logs_created_at;
DROP INDEX IF EXISTS idx_audit_logs_entity;
DROP INDEX IF EXISTS idx_audit_logs_user_id;

-- Drop audit_logs table
DROP TABLE IF EXISTS audit_logs;
-- +goose StatementEnd
