-- +goose Up
-- +goose StatementBegin
-- Create wing_type enum
CREATE TYPE wing_type AS ENUM ('East', 'West');

-- Create desk_status enum
CREATE TYPE desk_status AS ENUM ('available', 'maintenance');

-- Create desks table
CREATE TABLE IF NOT EXISTS desks (
    id SERIAL PRIMARY KEY,
    desk_number VARCHAR(20) NOT NULL UNIQUE,
    wing wing_type NOT NULL,
    status desk_status NOT NULL DEFAULT 'available',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create index on wing for filtering by wing
CREATE INDEX idx_desks_wing ON desks(wing);

-- Create index on status for filtering by availability
CREATE INDEX idx_desks_status ON desks(status);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Drop indexes
DROP INDEX IF EXISTS idx_desks_status;
DROP INDEX IF EXISTS idx_desks_wing;

-- Drop desks table
DROP TABLE IF EXISTS desks;

-- Drop enum types
DROP TYPE IF EXISTS desk_status;
DROP TYPE IF EXISTS wing_type;
-- +goose StatementEnd
