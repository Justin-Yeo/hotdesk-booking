-- +goose Up
-- +goose StatementBegin
-- Create settings table (single row for global settings)
CREATE TABLE IF NOT EXISTS settings (
    id INTEGER PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    opening_start TIME NOT NULL DEFAULT '08:00:00',
    opening_end TIME NOT NULL DEFAULT '22:00:00',
    daily_hour_limit INTEGER NOT NULL DEFAULT 10,
    check_in_grace_period_minutes INTEGER NOT NULL DEFAULT 15,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Insert default settings row
INSERT INTO settings (id) VALUES (1);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Drop settings table
DROP TABLE IF EXISTS settings;
-- +goose StatementEnd
