-- +goose Up
-- +goose StatementBegin
-- Create notifications table for in-app notifications
CREATE TABLE IF NOT EXISTS notifications (
    id SERIAL PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL,
    message TEXT NOT NULL,
    read BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create index on user_id for user-specific notification queries
CREATE INDEX idx_notifications_user_id ON notifications(user_id);

-- Create index on read status for filtering unread notifications
CREATE INDEX idx_notifications_read ON notifications(read);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Drop indexes
DROP INDEX IF EXISTS idx_notifications_read;
DROP INDEX IF EXISTS idx_notifications_user_id;

-- Drop notifications table
DROP TABLE IF EXISTS notifications;
-- +goose StatementEnd
