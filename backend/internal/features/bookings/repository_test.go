package bookings

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var testDB *pgxpool.Pool

func TestMain(m *testing.M) {
	// Setup database connection for tests
	ctx := context.Background()
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/hotdesk_booking?sslmode=disable"
	}

	var err error
	testDB, err = pgxpool.New(ctx, dbURL)
	if err != nil {
		panic("failed to connect to test database: " + err.Error())
	}
	defer testDB.Close()

	// Run tests
	code := m.Run()
	os.Exit(code)
}

// testDeskCounter is used to generate unique desk numbers
var testDeskCounter int

// Helper functions for test setup and cleanup
func setupTestDesk(t *testing.T) int {
	ctx := context.Background()
	testDeskCounter++
	var deskID int
	err := testDB.QueryRow(ctx,
		"INSERT INTO desks (desk_number, wing, status) VALUES ($1, $2, $3) RETURNING id",
		fmt.Sprintf("T%d-%d", testDeskCounter, time.Now().UnixNano()%1000000), "East", "available",
	).Scan(&deskID)
	if err != nil {
		t.Fatalf("failed to create test desk: %v", err)
	}
	return deskID
}

func setupTestUser(t *testing.T) string {
	ctx := context.Background()
	var userID string
	err := testDB.QueryRow(ctx,
		"INSERT INTO users (email, password_hash, role) VALUES ($1, $2, $3) RETURNING id",
		"test-"+time.Now().Format("150405.000")+"@example.com", "hash", "member",
	).Scan(&userID)
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}
	return userID
}

func cleanupTestBooking(_ *testing.T, bookingID int) {
	ctx := context.Background()
	_, _ = testDB.Exec(ctx, "DELETE FROM bookings WHERE id = $1", bookingID)
}

func cleanupTestDesk(_ *testing.T, deskID int) {
	ctx := context.Background()
	_, _ = testDB.Exec(ctx, "DELETE FROM bookings WHERE desk_id = $1", deskID)
	_, _ = testDB.Exec(ctx, "DELETE FROM desks WHERE id = $1", deskID)
}

func cleanupTestUser(_ *testing.T, userID string) {
	ctx := context.Background()
	_, _ = testDB.Exec(ctx, "DELETE FROM bookings WHERE user_id = $1", userID)
	_, _ = testDB.Exec(ctx, "DELETE FROM sessions WHERE user_id = $1", userID)
	_, _ = testDB.Exec(ctx, "DELETE FROM users WHERE id = $1", userID)
}

// ============================================================================
// CreateBooking Tests
// ============================================================================

func TestCreateBooking(t *testing.T) {
	deskID := setupTestDesk(t)
	userID := setupTestUser(t)
	defer cleanupTestDesk(t, deskID)
	defer cleanupTestUser(t, userID)

	repo := NewRepository(testDB)
	startTime := time.Now().Add(time.Hour).Truncate(time.Second)
	endTime := startTime.Add(2 * time.Hour)

	booking, err := repo.CreateBooking(context.Background(), &CreateBookingInput{
		DeskID:    deskID,
		UserID:    userID,
		StartTime: startTime,
		EndTime:   endTime,
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer cleanupTestBooking(t, booking.ID)

	if booking.DeskID != deskID {
		t.Errorf("expected desk_id %d, got %d", deskID, booking.DeskID)
	}
	if booking.UserID != userID {
		t.Errorf("expected user_id %s, got %s", userID, booking.UserID)
	}
	if booking.Status != StatusConfirmed {
		t.Errorf("expected status confirmed, got %s", booking.Status)
	}
}

func TestCreateBooking_Conflict(t *testing.T) {
	deskID := setupTestDesk(t)
	userID := setupTestUser(t)
	defer cleanupTestDesk(t, deskID)
	defer cleanupTestUser(t, userID)

	repo := NewRepository(testDB)
	startTime := time.Now().Add(time.Hour).Truncate(time.Second)
	endTime := startTime.Add(2 * time.Hour)

	// Create first booking
	booking1, err := repo.CreateBooking(context.Background(), &CreateBookingInput{
		DeskID:    deskID,
		UserID:    userID,
		StartTime: startTime,
		EndTime:   endTime,
	})
	if err != nil {
		t.Fatalf("failed to create first booking: %v", err)
	}
	defer cleanupTestBooking(t, booking1.ID)

	// Try to create overlapping booking
	_, err = repo.CreateBooking(context.Background(), &CreateBookingInput{
		DeskID:    deskID,
		UserID:    userID,
		StartTime: startTime.Add(30 * time.Minute), // Overlaps with first booking
		EndTime:   endTime.Add(30 * time.Minute),
	})

	if err != ErrBookingConflict {
		t.Errorf("expected ErrBookingConflict, got %v", err)
	}
}

// ============================================================================
// GetBookingByID Tests
// ============================================================================

func TestGetBookingByID(t *testing.T) {
	deskID := setupTestDesk(t)
	userID := setupTestUser(t)
	defer cleanupTestDesk(t, deskID)
	defer cleanupTestUser(t, userID)

	repo := NewRepository(testDB)
	startTime := time.Now().Add(time.Hour).Truncate(time.Second)
	endTime := startTime.Add(2 * time.Hour)

	created, err := repo.CreateBooking(context.Background(), &CreateBookingInput{
		DeskID:    deskID,
		UserID:    userID,
		StartTime: startTime,
		EndTime:   endTime,
	})
	if err != nil {
		t.Fatalf("failed to create booking: %v", err)
	}
	defer cleanupTestBooking(t, created.ID)

	booking, err := repo.GetBookingByID(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if booking.ID != created.ID {
		t.Errorf("expected id %d, got %d", created.ID, booking.ID)
	}
}

func TestGetBookingByID_NotFound(t *testing.T) {
	repo := NewRepository(testDB)

	_, err := repo.GetBookingByID(context.Background(), 999999)
	if err != ErrBookingNotFound {
		t.Errorf("expected ErrBookingNotFound, got %v", err)
	}
}

// ============================================================================
// GetUserBookings Tests
// ============================================================================

func TestGetUserBookings(t *testing.T) {
	deskID := setupTestDesk(t)
	userID := setupTestUser(t)
	defer cleanupTestDesk(t, deskID)
	defer cleanupTestUser(t, userID)

	repo := NewRepository(testDB)

	// Create multiple bookings
	for i := 0; i < 3; i++ {
		startTime := time.Now().Add(time.Duration(i+1) * 24 * time.Hour).Truncate(time.Second)
		endTime := startTime.Add(2 * time.Hour)
		booking, err := repo.CreateBooking(context.Background(), &CreateBookingInput{
			DeskID:    deskID,
			UserID:    userID,
			StartTime: startTime,
			EndTime:   endTime,
		})
		if err != nil {
			t.Fatalf("failed to create booking %d: %v", i, err)
		}
		defer cleanupTestBooking(t, booking.ID)
	}

	bookings, err := repo.GetUserBookings(context.Background(), &BookingFilter{
		UserID: &userID,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(bookings) != 3 {
		t.Errorf("expected 3 bookings, got %d", len(bookings))
	}
}

func TestGetUserBookings_WithFilters(t *testing.T) {
	deskID := setupTestDesk(t)
	userID := setupTestUser(t)
	defer cleanupTestDesk(t, deskID)
	defer cleanupTestUser(t, userID)

	repo := NewRepository(testDB)

	// Create booking
	startTime := time.Now().Add(time.Hour).Truncate(time.Second)
	endTime := startTime.Add(2 * time.Hour)
	booking, err := repo.CreateBooking(context.Background(), &CreateBookingInput{
		DeskID:    deskID,
		UserID:    userID,
		StartTime: startTime,
		EndTime:   endTime,
	})
	if err != nil {
		t.Fatalf("failed to create booking: %v", err)
	}
	defer cleanupTestBooking(t, booking.ID)

	// Filter by status
	status := StatusConfirmed
	bookings, err := repo.GetUserBookings(context.Background(), &BookingFilter{
		UserID: &userID,
		Status: &status,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(bookings) != 1 {
		t.Errorf("expected 1 booking, got %d", len(bookings))
	}

	// Filter by non-matching status
	cancelledStatus := StatusCancelled
	bookings, err = repo.GetUserBookings(context.Background(), &BookingFilter{
		UserID: &userID,
		Status: &cancelledStatus,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(bookings) != 0 {
		t.Errorf("expected 0 bookings, got %d", len(bookings))
	}
}

func TestGetUserBookings_WithLimitOffset(t *testing.T) {
	deskID := setupTestDesk(t)
	userID := setupTestUser(t)
	defer cleanupTestDesk(t, deskID)
	defer cleanupTestUser(t, userID)

	repo := NewRepository(testDB)

	// Create multiple bookings
	for i := 0; i < 5; i++ {
		startTime := time.Now().Add(time.Duration(i+1) * 24 * time.Hour).Truncate(time.Second)
		endTime := startTime.Add(2 * time.Hour)
		booking, err := repo.CreateBooking(context.Background(), &CreateBookingInput{
			DeskID:    deskID,
			UserID:    userID,
			StartTime: startTime,
			EndTime:   endTime,
		})
		if err != nil {
			t.Fatalf("failed to create booking %d: %v", i, err)
		}
		defer cleanupTestBooking(t, booking.ID)
	}

	// Test limit
	bookings, err := repo.GetUserBookings(context.Background(), &BookingFilter{
		UserID: &userID,
		Limit:  2,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(bookings) != 2 {
		t.Errorf("expected 2 bookings with limit, got %d", len(bookings))
	}

	// Test offset
	bookings, err = repo.GetUserBookings(context.Background(), &BookingFilter{
		UserID: &userID,
		Limit:  10,
		Offset: 3,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(bookings) != 2 {
		t.Errorf("expected 2 bookings with offset, got %d", len(bookings))
	}
}

// ============================================================================
// UpdateBooking Tests
// ============================================================================

func TestUpdateBooking(t *testing.T) {
	deskID := setupTestDesk(t)
	userID := setupTestUser(t)
	defer cleanupTestDesk(t, deskID)
	defer cleanupTestUser(t, userID)

	repo := NewRepository(testDB)
	startTime := time.Now().Add(time.Hour).Truncate(time.Second)
	endTime := startTime.Add(2 * time.Hour)

	created, err := repo.CreateBooking(context.Background(), &CreateBookingInput{
		DeskID:    deskID,
		UserID:    userID,
		StartTime: startTime,
		EndTime:   endTime,
	})
	if err != nil {
		t.Fatalf("failed to create booking: %v", err)
	}
	defer cleanupTestBooking(t, created.ID)

	// Update status
	newStatus := StatusCompleted
	checkedIn := time.Now().Truncate(time.Second)
	updated, err := repo.UpdateBooking(context.Background(), created.ID, &UpdateBookingInput{
		Status:      &newStatus,
		CheckedInAt: &checkedIn,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if updated.Status != StatusCompleted {
		t.Errorf("expected status completed, got %s", updated.Status)
	}
	if updated.CheckedInAt == nil {
		t.Error("expected checked_in_at to be set")
	}
}

func TestUpdateBooking_NotFound(t *testing.T) {
	repo := NewRepository(testDB)
	newStatus := StatusCancelled

	_, err := repo.UpdateBooking(context.Background(), 999999, &UpdateBookingInput{
		Status: &newStatus,
	})
	if err != ErrBookingNotFound {
		t.Errorf("expected ErrBookingNotFound, got %v", err)
	}
}

// ============================================================================
// DeleteBooking Tests
// ============================================================================

func TestDeleteBooking(t *testing.T) {
	deskID := setupTestDesk(t)
	userID := setupTestUser(t)
	defer cleanupTestDesk(t, deskID)
	defer cleanupTestUser(t, userID)

	repo := NewRepository(testDB)
	startTime := time.Now().Add(time.Hour).Truncate(time.Second)
	endTime := startTime.Add(2 * time.Hour)

	created, err := repo.CreateBooking(context.Background(), &CreateBookingInput{
		DeskID:    deskID,
		UserID:    userID,
		StartTime: startTime,
		EndTime:   endTime,
	})
	if err != nil {
		t.Fatalf("failed to create booking: %v", err)
	}
	defer cleanupTestBooking(t, created.ID)

	err = repo.DeleteBooking(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify booking is cancelled
	booking, err := repo.GetBookingByID(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("failed to get booking: %v", err)
	}

	if booking.Status != StatusCancelled {
		t.Errorf("expected status cancelled, got %s", booking.Status)
	}
	if booking.CancelledAt == nil {
		t.Error("expected cancelled_at to be set")
	}
}

func TestDeleteBooking_NotFound(t *testing.T) {
	repo := NewRepository(testDB)

	err := repo.DeleteBooking(context.Background(), 999999)
	if err != ErrBookingNotFound {
		t.Errorf("expected ErrBookingNotFound, got %v", err)
	}
}

func TestDeleteBooking_AlreadyCancelled(t *testing.T) {
	deskID := setupTestDesk(t)
	userID := setupTestUser(t)
	defer cleanupTestDesk(t, deskID)
	defer cleanupTestUser(t, userID)

	repo := NewRepository(testDB)
	startTime := time.Now().Add(time.Hour).Truncate(time.Second)
	endTime := startTime.Add(2 * time.Hour)

	created, err := repo.CreateBooking(context.Background(), &CreateBookingInput{
		DeskID:    deskID,
		UserID:    userID,
		StartTime: startTime,
		EndTime:   endTime,
	})
	if err != nil {
		t.Fatalf("failed to create booking: %v", err)
	}
	defer cleanupTestBooking(t, created.ID)

	// Cancel once
	err = repo.DeleteBooking(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("expected no error on first cancel, got %v", err)
	}

	// Cancel again - should succeed (idempotent)
	err = repo.DeleteBooking(context.Background(), created.ID)
	if err != nil {
		t.Errorf("expected no error on second cancel, got %v", err)
	}
}

// ============================================================================
// IsDeskAvailable Tests
// ============================================================================

func TestIsDeskAvailable_Available(t *testing.T) {
	deskID := setupTestDesk(t)
	userID := setupTestUser(t)
	defer cleanupTestDesk(t, deskID)
	defer cleanupTestUser(t, userID)

	repo := NewRepository(testDB)

	// Check availability when no bookings exist
	available, err := repo.IsDeskAvailable(context.Background(), &DeskAvailabilityCheck{
		DeskID:    deskID,
		StartTime: time.Now().Add(time.Hour),
		EndTime:   time.Now().Add(3 * time.Hour),
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !available {
		t.Error("expected desk to be available")
	}
}

func TestIsDeskAvailable_NotAvailable(t *testing.T) {
	deskID := setupTestDesk(t)
	userID := setupTestUser(t)
	defer cleanupTestDesk(t, deskID)
	defer cleanupTestUser(t, userID)

	repo := NewRepository(testDB)
	startTime := time.Now().Add(time.Hour).Truncate(time.Second)
	endTime := startTime.Add(2 * time.Hour)

	// Create a booking
	booking, err := repo.CreateBooking(context.Background(), &CreateBookingInput{
		DeskID:    deskID,
		UserID:    userID,
		StartTime: startTime,
		EndTime:   endTime,
	})
	if err != nil {
		t.Fatalf("failed to create booking: %v", err)
	}
	defer cleanupTestBooking(t, booking.ID)

	// Check overlapping time slot
	available, err := repo.IsDeskAvailable(context.Background(), &DeskAvailabilityCheck{
		DeskID:    deskID,
		StartTime: startTime.Add(30 * time.Minute),
		EndTime:   endTime.Add(30 * time.Minute),
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if available {
		t.Error("expected desk to not be available")
	}
}

func TestIsDeskAvailable_ExcludeBookingID(t *testing.T) {
	deskID := setupTestDesk(t)
	userID := setupTestUser(t)
	defer cleanupTestDesk(t, deskID)
	defer cleanupTestUser(t, userID)

	repo := NewRepository(testDB)
	startTime := time.Now().Add(time.Hour).Truncate(time.Second)
	endTime := startTime.Add(2 * time.Hour)

	// Create a booking
	booking, err := repo.CreateBooking(context.Background(), &CreateBookingInput{
		DeskID:    deskID,
		UserID:    userID,
		StartTime: startTime,
		EndTime:   endTime,
	})
	if err != nil {
		t.Fatalf("failed to create booking: %v", err)
	}
	defer cleanupTestBooking(t, booking.ID)

	// Check availability excluding the booking itself
	available, err := repo.IsDeskAvailable(context.Background(), &DeskAvailabilityCheck{
		DeskID:           deskID,
		StartTime:        startTime,
		EndTime:          endTime,
		ExcludeBookingID: &booking.ID,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !available {
		t.Error("expected desk to be available when excluding the same booking")
	}
}

func TestIsDeskAvailable_CancelledBookingIgnored(t *testing.T) {
	deskID := setupTestDesk(t)
	userID := setupTestUser(t)
	defer cleanupTestDesk(t, deskID)
	defer cleanupTestUser(t, userID)

	repo := NewRepository(testDB)
	startTime := time.Now().Add(time.Hour).Truncate(time.Second)
	endTime := startTime.Add(2 * time.Hour)

	// Create and cancel a booking
	booking, err := repo.CreateBooking(context.Background(), &CreateBookingInput{
		DeskID:    deskID,
		UserID:    userID,
		StartTime: startTime,
		EndTime:   endTime,
	})
	if err != nil {
		t.Fatalf("failed to create booking: %v", err)
	}
	defer cleanupTestBooking(t, booking.ID)

	err = repo.DeleteBooking(context.Background(), booking.ID)
	if err != nil {
		t.Fatalf("failed to cancel booking: %v", err)
	}

	// Check availability - should be available since booking is cancelled
	available, err := repo.IsDeskAvailable(context.Background(), &DeskAvailabilityCheck{
		DeskID:    deskID,
		StartTime: startTime,
		EndTime:   endTime,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !available {
		t.Error("expected desk to be available when booking is cancelled")
	}
}

// ============================================================================
// GetUserDailyHours Tests
// ============================================================================

func TestGetUserDailyHours(t *testing.T) {
	deskID1 := setupTestDesk(t)
	deskID2 := setupTestDesk(t)
	userID := setupTestUser(t)
	defer cleanupTestDesk(t, deskID1)
	defer cleanupTestDesk(t, deskID2)
	defer cleanupTestUser(t, userID)

	repo := NewRepository(testDB)

	// Use a future date to avoid timezone issues
	// Create date at noon UTC to ensure it's the same day everywhere
	futureDate := time.Now().AddDate(0, 0, 7).UTC()
	testDay := time.Date(futureDate.Year(), futureDate.Month(), futureDate.Day(), 0, 0, 0, 0, time.UTC)

	// Booking 1: 2 hours on desk 1 (9:00 - 11:00)
	booking1Start := testDay.Add(9 * time.Hour)
	booking1, err := repo.CreateBooking(context.Background(), &CreateBookingInput{
		DeskID:    deskID1,
		UserID:    userID,
		StartTime: booking1Start,
		EndTime:   booking1Start.Add(2 * time.Hour),
	})
	if err != nil {
		t.Fatalf("failed to create booking 1: %v", err)
	}
	defer cleanupTestBooking(t, booking1.ID)

	// Booking 2: 3 hours on desk 2 (14:00 - 17:00)
	booking2Start := testDay.Add(14 * time.Hour)
	booking2, err := repo.CreateBooking(context.Background(), &CreateBookingInput{
		DeskID:    deskID2,
		UserID:    userID,
		StartTime: booking2Start,
		EndTime:   booking2Start.Add(3 * time.Hour),
	})
	if err != nil {
		t.Fatalf("failed to create booking 2: %v", err)
	}
	defer cleanupTestBooking(t, booking2.ID)

	// Calculate total hours
	hours, err := repo.GetUserDailyHours(context.Background(), userID, testDay)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if hours != 5.0 {
		t.Errorf("expected 5.0 hours, got %f", hours)
	}
}

func TestGetUserDailyHours_ExcludesCancelled(t *testing.T) {
	deskID := setupTestDesk(t)
	userID := setupTestUser(t)
	defer cleanupTestDesk(t, deskID)
	defer cleanupTestUser(t, userID)

	repo := NewRepository(testDB)

	today := time.Now().Truncate(24 * time.Hour)

	// Create and cancel a booking
	bookingStart := today.Add(9 * time.Hour)
	booking, err := repo.CreateBooking(context.Background(), &CreateBookingInput{
		DeskID:    deskID,
		UserID:    userID,
		StartTime: bookingStart,
		EndTime:   bookingStart.Add(2 * time.Hour),
	})
	if err != nil {
		t.Fatalf("failed to create booking: %v", err)
	}
	defer cleanupTestBooking(t, booking.ID)

	err = repo.DeleteBooking(context.Background(), booking.ID)
	if err != nil {
		t.Fatalf("failed to cancel booking: %v", err)
	}

	// Calculate total hours - should be 0 since booking is cancelled
	hours, err := repo.GetUserDailyHours(context.Background(), userID, today)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if hours != 0.0 {
		t.Errorf("expected 0.0 hours for cancelled booking, got %f", hours)
	}
}

func TestGetUserDailyHours_NoBookings(t *testing.T) {
	userID := setupTestUser(t)
	defer cleanupTestUser(t, userID)

	repo := NewRepository(testDB)

	hours, err := repo.GetUserDailyHours(context.Background(), userID, time.Now())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if hours != 0.0 {
		t.Errorf("expected 0.0 hours, got %f", hours)
	}
}

// ============================================================================
// GetBookingsByTimeRange Tests
// ============================================================================

func TestGetBookingsByTimeRange_Empty(t *testing.T) {
	deskID := setupTestDesk(t)
	defer cleanupTestDesk(t, deskID)

	repo := NewRepository(testDB)

	// Query a time range with no bookings
	bookings, err := repo.GetBookingsByTimeRange(context.Background(), deskID,
		time.Now().Add(100*24*time.Hour),
		time.Now().Add(101*24*time.Hour))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(bookings) != 0 {
		t.Errorf("expected 0 bookings, got %d", len(bookings))
	}
}

func TestGetUserBookings_WithDateRange(t *testing.T) {
	deskID := setupTestDesk(t)
	userID := setupTestUser(t)
	defer cleanupTestDesk(t, deskID)
	defer cleanupTestUser(t, userID)

	repo := NewRepository(testDB)

	// Create a booking for tomorrow
	tomorrow := time.Now().AddDate(0, 0, 1).UTC()
	startTime := time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 10, 0, 0, 0, time.UTC)
	booking, err := repo.CreateBooking(context.Background(), &CreateBookingInput{
		DeskID:    deskID,
		UserID:    userID,
		StartTime: startTime,
		EndTime:   startTime.Add(2 * time.Hour),
	})
	if err != nil {
		t.Fatalf("failed to create booking: %v", err)
	}
	defer cleanupTestBooking(t, booking.ID)

	// Query with date range that includes the booking
	startDate := startTime.Add(-time.Hour)
	endDate := startTime.Add(3 * time.Hour)
	bookings, err := repo.GetUserBookings(context.Background(), &BookingFilter{
		UserID:    &userID,
		StartDate: &startDate,
		EndDate:   &endDate,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(bookings) != 1 {
		t.Errorf("expected 1 booking, got %d", len(bookings))
	}

	// Query with date range that excludes the booking
	excludeStart := startTime.Add(-48 * time.Hour)
	excludeEnd := startTime.Add(-24 * time.Hour)
	bookings, err = repo.GetUserBookings(context.Background(), &BookingFilter{
		UserID:    &userID,
		StartDate: &excludeStart,
		EndDate:   &excludeEnd,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(bookings) != 0 {
		t.Errorf("expected 0 bookings, got %d", len(bookings))
	}
}

func TestUpdateBooking_TimeChange(t *testing.T) {
	deskID := setupTestDesk(t)
	userID := setupTestUser(t)
	defer cleanupTestDesk(t, deskID)
	defer cleanupTestUser(t, userID)

	repo := NewRepository(testDB)
	startTime := time.Now().Add(time.Hour).Truncate(time.Second)
	endTime := startTime.Add(2 * time.Hour)

	created, err := repo.CreateBooking(context.Background(), &CreateBookingInput{
		DeskID:    deskID,
		UserID:    userID,
		StartTime: startTime,
		EndTime:   endTime,
	})
	if err != nil {
		t.Fatalf("failed to create booking: %v", err)
	}
	defer cleanupTestBooking(t, created.ID)

	// Update the time
	newStart := startTime.Add(30 * time.Minute)
	newEnd := endTime.Add(30 * time.Minute)
	updated, err := repo.UpdateBooking(context.Background(), created.ID, &UpdateBookingInput{
		StartTime: &newStart,
		EndTime:   &newEnd,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !updated.StartTime.Equal(newStart) {
		t.Errorf("expected start_time %v, got %v", newStart, updated.StartTime)
	}
	if !updated.EndTime.Equal(newEnd) {
		t.Errorf("expected end_time %v, got %v", newEnd, updated.EndTime)
	}
}

func TestGetUserBookings_WithDeskFilter(t *testing.T) {
	deskID1 := setupTestDesk(t)
	deskID2 := setupTestDesk(t)
	userID := setupTestUser(t)
	defer cleanupTestDesk(t, deskID1)
	defer cleanupTestDesk(t, deskID2)
	defer cleanupTestUser(t, userID)

	repo := NewRepository(testDB)

	// Create booking on desk 1
	startTime1 := time.Now().Add(time.Hour).Truncate(time.Second)
	booking1, err := repo.CreateBooking(context.Background(), &CreateBookingInput{
		DeskID:    deskID1,
		UserID:    userID,
		StartTime: startTime1,
		EndTime:   startTime1.Add(2 * time.Hour),
	})
	if err != nil {
		t.Fatalf("failed to create booking 1: %v", err)
	}
	defer cleanupTestBooking(t, booking1.ID)

	// Create booking on desk 2
	startTime2 := time.Now().Add(24 * time.Hour).Truncate(time.Second)
	booking2, err := repo.CreateBooking(context.Background(), &CreateBookingInput{
		DeskID:    deskID2,
		UserID:    userID,
		StartTime: startTime2,
		EndTime:   startTime2.Add(2 * time.Hour),
	})
	if err != nil {
		t.Fatalf("failed to create booking 2: %v", err)
	}
	defer cleanupTestBooking(t, booking2.ID)

	// Filter by desk 1
	bookings, err := repo.GetUserBookings(context.Background(), &BookingFilter{
		UserID: &userID,
		DeskID: &deskID1,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(bookings) != 1 {
		t.Errorf("expected 1 booking for desk 1, got %d", len(bookings))
	}
}

func TestGetBookingsByTimeRange(t *testing.T) {
	deskID := setupTestDesk(t)
	userID := setupTestUser(t)
	defer cleanupTestDesk(t, deskID)
	defer cleanupTestUser(t, userID)

	repo := NewRepository(testDB)

	today := time.Now().Truncate(24 * time.Hour)

	// Create bookings
	booking1Start := today.Add(9 * time.Hour)
	booking1, err := repo.CreateBooking(context.Background(), &CreateBookingInput{
		DeskID:    deskID,
		UserID:    userID,
		StartTime: booking1Start,
		EndTime:   booking1Start.Add(2 * time.Hour),
	})
	if err != nil {
		t.Fatalf("failed to create booking 1: %v", err)
	}
	defer cleanupTestBooking(t, booking1.ID)

	booking2Start := today.Add(14 * time.Hour)
	booking2, err := repo.CreateBooking(context.Background(), &CreateBookingInput{
		DeskID:    deskID,
		UserID:    userID,
		StartTime: booking2Start,
		EndTime:   booking2Start.Add(3 * time.Hour),
	})
	if err != nil {
		t.Fatalf("failed to create booking 2: %v", err)
	}
	defer cleanupTestBooking(t, booking2.ID)

	// Query full day
	bookings, err := repo.GetBookingsByTimeRange(context.Background(), deskID, today, today.Add(24*time.Hour))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(bookings) != 2 {
		t.Errorf("expected 2 bookings, got %d", len(bookings))
	}
}
