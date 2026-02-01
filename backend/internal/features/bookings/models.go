// Package bookings provides booking management functionality
// including desk reservations, availability checking, and booking history.
package bookings

import (
	"time"
)

// BookingStatus represents the status of a booking
type BookingStatus string

const (
	// StatusConfirmed indicates a confirmed booking
	StatusConfirmed BookingStatus = "confirmed"
	// StatusCancelled indicates a cancelled booking
	StatusCancelled BookingStatus = "cancelled"
	// StatusCompleted indicates a completed booking
	StatusCompleted BookingStatus = "completed"
	// StatusNoShow indicates the user did not show up
	StatusNoShow BookingStatus = "no_show"
)

// Booking represents a desk booking in the system
type Booking struct {
	ID            int           `json:"id"`
	DeskID        int           `json:"desk_id"`
	UserID        string        `json:"user_id"`
	StartTime     time.Time     `json:"start_time"`
	EndTime       time.Time     `json:"end_time"`
	Status        BookingStatus `json:"status"`
	CheckedInAt   *time.Time    `json:"checked_in_at,omitempty"`
	ActualEndTime *time.Time    `json:"actual_end_time,omitempty"`
	CancelledAt   *time.Time    `json:"cancelled_at,omitempty"`
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
}

// CreateBookingInput represents the input for creating a new booking
type CreateBookingInput struct {
	DeskID    int
	UserID    string
	StartTime time.Time
	EndTime   time.Time
}

// UpdateBookingInput represents the input for updating an existing booking
type UpdateBookingInput struct {
	StartTime     *time.Time
	EndTime       *time.Time
	Status        *BookingStatus
	CheckedInAt   *time.Time
	ActualEndTime *time.Time
	CancelledAt   *time.Time
}

// BookingFilter represents filters for querying bookings
type BookingFilter struct {
	UserID    *string
	DeskID    *int
	Status    *BookingStatus
	StartDate *time.Time
	EndDate   *time.Time
	Limit     int
	Offset    int
}

// DeskAvailabilityCheck represents the parameters for checking desk availability
type DeskAvailabilityCheck struct {
	DeskID           int
	StartTime        time.Time
	EndTime          time.Time
	ExcludeBookingID *int // Exclude this booking ID when checking (for updates)
}
