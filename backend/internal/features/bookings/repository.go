package bookings

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	// ErrBookingNotFound is returned when a booking is not found
	ErrBookingNotFound = errors.New("booking not found")
	// ErrBookingConflict is returned when there's a time overlap with existing booking
	ErrBookingConflict = errors.New("booking conflicts with an existing reservation")
	// ErrDeskNotAvailable is returned when the desk is not available for booking
	ErrDeskNotAvailable = errors.New("desk is not available for the requested time slot")
)

// PostgreSQL error code for exclusion_violation
const pgExclusionViolation = "23P01"

// Repository provides database operations for booking-related entities
type Repository struct {
	db *pgxpool.Pool
}

// NewRepository creates a new bookings repository
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// CreateBooking inserts a new booking into the database
func (r *Repository) CreateBooking(ctx context.Context, input *CreateBookingInput) (*Booking, error) {
	query := `
		INSERT INTO bookings (desk_id, user_id, start_time, end_time)
		VALUES ($1, $2, $3, $4)
		RETURNING id, desk_id, user_id, start_time, end_time, status,
		          checked_in_at, actual_end_time, cancelled_at, created_at, updated_at
	`

	var booking Booking
	err := r.db.QueryRow(ctx, query,
		input.DeskID,
		input.UserID,
		input.StartTime,
		input.EndTime,
	).Scan(
		&booking.ID,
		&booking.DeskID,
		&booking.UserID,
		&booking.StartTime,
		&booking.EndTime,
		&booking.Status,
		&booking.CheckedInAt,
		&booking.ActualEndTime,
		&booking.CancelledAt,
		&booking.CreatedAt,
		&booking.UpdatedAt,
	)

	if err != nil {
		// Check for exclusion constraint violation (overlapping booking)
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgExclusionViolation {
			return nil, ErrBookingConflict
		}
		return nil, fmt.Errorf("failed to create booking: %w", err)
	}

	return &booking, nil
}

// GetBookingByID retrieves a booking by its ID
func (r *Repository) GetBookingByID(ctx context.Context, id int) (*Booking, error) {
	query := `
		SELECT id, desk_id, user_id, start_time, end_time, status,
		       checked_in_at, actual_end_time, cancelled_at, created_at, updated_at
		FROM bookings
		WHERE id = $1
	`

	var booking Booking
	err := r.db.QueryRow(ctx, query, id).Scan(
		&booking.ID,
		&booking.DeskID,
		&booking.UserID,
		&booking.StartTime,
		&booking.EndTime,
		&booking.Status,
		&booking.CheckedInAt,
		&booking.ActualEndTime,
		&booking.CancelledAt,
		&booking.CreatedAt,
		&booking.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrBookingNotFound
		}
		return nil, fmt.Errorf("failed to get booking: %w", err)
	}

	return &booking, nil
}

// GetUserBookings retrieves bookings for a user with optional filters
func (r *Repository) GetUserBookings(ctx context.Context, filter *BookingFilter) ([]*Booking, error) {
	query := `
		SELECT id, desk_id, user_id, start_time, end_time, status,
		       checked_in_at, actual_end_time, cancelled_at, created_at, updated_at
		FROM bookings
		WHERE 1=1
	`
	args := []interface{}{}
	argNum := 1

	if filter.UserID != nil {
		query += fmt.Sprintf(" AND user_id = $%d", argNum)
		args = append(args, *filter.UserID)
		argNum++
	}

	if filter.DeskID != nil {
		query += fmt.Sprintf(" AND desk_id = $%d", argNum)
		args = append(args, *filter.DeskID)
		argNum++
	}

	if filter.Status != nil {
		query += fmt.Sprintf(" AND status = $%d", argNum)
		args = append(args, *filter.Status)
		argNum++
	}

	if filter.StartDate != nil {
		query += fmt.Sprintf(" AND start_time >= $%d", argNum)
		args = append(args, *filter.StartDate)
		argNum++
	}

	if filter.EndDate != nil {
		query += fmt.Sprintf(" AND end_time <= $%d", argNum)
		args = append(args, *filter.EndDate)
		argNum++
	}

	query += " ORDER BY start_time DESC"

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argNum)
		args = append(args, filter.Limit)
		argNum++
	}

	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argNum)
		args = append(args, filter.Offset)
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query bookings: %w", err)
	}
	defer rows.Close()

	var bookings []*Booking
	for rows.Next() {
		var booking Booking
		err := rows.Scan(
			&booking.ID,
			&booking.DeskID,
			&booking.UserID,
			&booking.StartTime,
			&booking.EndTime,
			&booking.Status,
			&booking.CheckedInAt,
			&booking.ActualEndTime,
			&booking.CancelledAt,
			&booking.CreatedAt,
			&booking.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan booking: %w", err)
		}
		bookings = append(bookings, &booking)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating bookings: %w", err)
	}

	return bookings, nil
}

// UpdateBooking updates an existing booking's fields
func (r *Repository) UpdateBooking(ctx context.Context, id int, input *UpdateBookingInput) (*Booking, error) {
	query := `UPDATE bookings SET updated_at = NOW()`
	args := []interface{}{}
	argNum := 1

	if input.StartTime != nil {
		query += fmt.Sprintf(", start_time = $%d", argNum)
		args = append(args, *input.StartTime)
		argNum++
	}

	if input.EndTime != nil {
		query += fmt.Sprintf(", end_time = $%d", argNum)
		args = append(args, *input.EndTime)
		argNum++
	}

	if input.Status != nil {
		query += fmt.Sprintf(", status = $%d", argNum)
		args = append(args, *input.Status)
		argNum++
	}

	if input.CheckedInAt != nil {
		query += fmt.Sprintf(", checked_in_at = $%d", argNum)
		args = append(args, *input.CheckedInAt)
		argNum++
	}

	if input.ActualEndTime != nil {
		query += fmt.Sprintf(", actual_end_time = $%d", argNum)
		args = append(args, *input.ActualEndTime)
		argNum++
	}

	if input.CancelledAt != nil {
		query += fmt.Sprintf(", cancelled_at = $%d", argNum)
		args = append(args, *input.CancelledAt)
		argNum++
	}

	query += fmt.Sprintf(`
		WHERE id = $%d
		RETURNING id, desk_id, user_id, start_time, end_time, status,
		          checked_in_at, actual_end_time, cancelled_at, created_at, updated_at
	`, argNum)
	args = append(args, id)

	var booking Booking
	err := r.db.QueryRow(ctx, query, args...).Scan(
		&booking.ID,
		&booking.DeskID,
		&booking.UserID,
		&booking.StartTime,
		&booking.EndTime,
		&booking.Status,
		&booking.CheckedInAt,
		&booking.ActualEndTime,
		&booking.CancelledAt,
		&booking.CreatedAt,
		&booking.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrBookingNotFound
		}
		// Check for exclusion constraint violation
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgExclusionViolation {
			return nil, ErrBookingConflict
		}
		return nil, fmt.Errorf("failed to update booking: %w", err)
	}

	return &booking, nil
}

// DeleteBooking performs a soft delete by setting status to 'cancelled'
func (r *Repository) DeleteBooking(ctx context.Context, id int) error {
	now := time.Now()
	query := `
		UPDATE bookings
		SET status = 'cancelled', cancelled_at = $1, updated_at = NOW()
		WHERE id = $2 AND status NOT IN ('cancelled', 'completed', 'no_show')
	`

	result, err := r.db.Exec(ctx, query, now, id)
	if err != nil {
		return fmt.Errorf("failed to cancel booking: %w", err)
	}

	if result.RowsAffected() == 0 {
		// Check if booking exists
		var exists bool
		checkQuery := `SELECT EXISTS(SELECT 1 FROM bookings WHERE id = $1)`
		if err := r.db.QueryRow(ctx, checkQuery, id).Scan(&exists); err != nil {
			return fmt.Errorf("failed to check booking existence: %w", err)
		}
		if !exists {
			return ErrBookingNotFound
		}
		// Booking exists but couldn't be cancelled (already cancelled/completed/no_show)
		return nil
	}

	return nil
}

// IsDeskAvailable checks if a desk is available for the specified time range
func (r *Repository) IsDeskAvailable(ctx context.Context, check *DeskAvailabilityCheck) (bool, error) {
	// Use the time_range column and GIST index for efficient overlap detection
	// Only check against active bookings (not cancelled or no_show)
	query := `
		SELECT NOT EXISTS (
			SELECT 1 FROM bookings
			WHERE desk_id = $1
			  AND status NOT IN ('cancelled', 'no_show')
			  AND time_range && tstzrange($2, $3)
	`
	args := []interface{}{check.DeskID, check.StartTime, check.EndTime}
	argNum := 4

	// Exclude a specific booking ID (useful for update operations)
	if check.ExcludeBookingID != nil {
		query += fmt.Sprintf(" AND id != $%d", argNum)
		args = append(args, *check.ExcludeBookingID)
	}

	query += ")"

	var available bool
	err := r.db.QueryRow(ctx, query, args...).Scan(&available)
	if err != nil {
		return false, fmt.Errorf("failed to check desk availability: %w", err)
	}

	return available, nil
}

// GetUserDailyHours calculates total booked hours for a user on a specific day
// Only counts confirmed and completed bookings (excludes cancelled and no_show)
func (r *Repository) GetUserDailyHours(ctx context.Context, userID string, date time.Time) (float64, error) {
	// Calculate the start and end of the day in the same timezone as the date
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	query := `
		SELECT COALESCE(
			SUM(
				EXTRACT(EPOCH FROM (
					LEAST(end_time, $3) - GREATEST(start_time, $2)
				)) / 3600.0
			),
			0
		)
		FROM bookings
		WHERE user_id = $1
		  AND status IN ('confirmed', 'completed')
		  AND start_time < $3
		  AND end_time > $2
	`

	var totalHours float64
	err := r.db.QueryRow(ctx, query, userID, startOfDay, endOfDay).Scan(&totalHours)
	if err != nil {
		return 0, fmt.Errorf("failed to calculate daily hours: %w", err)
	}

	return totalHours, nil
}

// GetBookingsByTimeRange retrieves all bookings that overlap with the given time range
func (r *Repository) GetBookingsByTimeRange(ctx context.Context, deskID int, startTime, endTime time.Time) ([]*Booking, error) {
	query := `
		SELECT id, desk_id, user_id, start_time, end_time, status,
		       checked_in_at, actual_end_time, cancelled_at, created_at, updated_at
		FROM bookings
		WHERE desk_id = $1
		  AND status NOT IN ('cancelled', 'no_show')
		  AND time_range && tstzrange($2, $3)
		ORDER BY start_time
	`

	rows, err := r.db.Query(ctx, query, deskID, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to query bookings by time range: %w", err)
	}
	defer rows.Close()

	var bookings []*Booking
	for rows.Next() {
		var booking Booking
		err := rows.Scan(
			&booking.ID,
			&booking.DeskID,
			&booking.UserID,
			&booking.StartTime,
			&booking.EndTime,
			&booking.Status,
			&booking.CheckedInAt,
			&booking.ActualEndTime,
			&booking.CancelledAt,
			&booking.CreatedAt,
			&booking.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan booking: %w", err)
		}
		bookings = append(bookings, &booking)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating bookings: %w", err)
	}

	return bookings, nil
}
