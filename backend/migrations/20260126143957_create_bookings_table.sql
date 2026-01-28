-- +goose Up
-- +goose StatementBegin
-- Enable btree_gist extension for exclusion constraint
CREATE EXTENSION IF NOT EXISTS btree_gist;

-- Create booking_status enum
CREATE TYPE booking_status AS ENUM ('confirmed', 'cancelled', 'completed', 'no_show');

-- Create bookings table with exclusion constraint for overlap detection
CREATE TABLE IF NOT EXISTS bookings (
    id SERIAL PRIMARY KEY,
    desk_id INTEGER NOT NULL REFERENCES desks(id),
    user_id UUID NOT NULL REFERENCES users(id),
    start_time TIMESTAMPTZ NOT NULL,
    end_time TIMESTAMPTZ NOT NULL,
    time_range TSTZRANGE GENERATED ALWAYS AS (tstzrange(start_time, end_time)) STORED,
    status booking_status NOT NULL DEFAULT 'confirmed',
    checked_in_at TIMESTAMPTZ,
    actual_end_time TIMESTAMPTZ,
    cancelled_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT no_overlapping_bookings
        EXCLUDE USING GIST (
            desk_id WITH =,
            time_range WITH &&
        ) WHERE (status NOT IN ('cancelled', 'no_show'))
);

-- Create index on desk_id for desk-based queries
CREATE INDEX idx_bookings_desk_id ON bookings(desk_id);

-- Create index on user_id for user booking history
CREATE INDEX idx_bookings_user_id ON bookings(user_id);

-- Create GIST index on time_range for range queries
CREATE INDEX idx_bookings_time_range ON bookings USING GIST(time_range);

-- Create index on status for filtering by booking status
CREATE INDEX idx_bookings_status ON bookings(status);

-- Create index on start_time for no-show queries
CREATE INDEX idx_bookings_start_time ON bookings(start_time);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Drop indexes
DROP INDEX IF EXISTS idx_bookings_start_time;
DROP INDEX IF EXISTS idx_bookings_status;
DROP INDEX IF EXISTS idx_bookings_time_range;
DROP INDEX IF EXISTS idx_bookings_user_id;
DROP INDEX IF EXISTS idx_bookings_desk_id;

-- Drop bookings table
DROP TABLE IF EXISTS bookings;

-- Drop enum type
DROP TYPE IF EXISTS booking_status;

-- Note: We don't drop btree_gist extension as it might be used by other tables
-- +goose StatementEnd
