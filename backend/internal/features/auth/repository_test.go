package auth

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var testDB *pgxpool.Pool

func TestMain(m *testing.M) {
	// Setup test database connection
	ctx := context.Background()
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/hotdesk_booking?sslmode=disable"
	}

	var err error
	testDB, err = pgxpool.New(ctx, dbURL)
	if err != nil {
		panic("Failed to connect to test database: " + err.Error())
	}
	defer testDB.Close()

	// Run tests
	code := m.Run()
	os.Exit(code)
}

// cleanupTestUser removes a test user by email
func cleanupTestUser(_ *testing.T, email string) {
	ctx := context.Background()
	_, _ = testDB.Exec(ctx, "DELETE FROM users WHERE email = $1", email)
}

// cleanupTestSession removes test sessions by user ID
func cleanupTestSession(_ *testing.T, userID string) {
	ctx := context.Background()
	_, _ = testDB.Exec(ctx, "DELETE FROM sessions WHERE user_id = $1", userID)
}

// ============================================================================
// User Repository Tests
// ============================================================================

func TestCreateUser(t *testing.T) {
	repo := NewRepository(testDB)
	ctx := context.Background()
	testEmail := "test_create_user@example.com"

	// Cleanup before and after test
	cleanupTestUser(t, testEmail)
	defer cleanupTestUser(t, testEmail)

	input := &CreateUserInput{
		Email:        testEmail,
		PasswordHash: "$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/X4.",
		Role:         RoleMember,
	}

	user, err := repo.CreateUser(ctx, input)
	if err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}

	if user.ID == "" {
		t.Error("CreateUser() user.ID should not be empty")
	}
	if user.Email != testEmail {
		t.Errorf("CreateUser() user.Email = %v, want %v", user.Email, testEmail)
	}
	if user.Role != RoleMember {
		t.Errorf("CreateUser() user.Role = %v, want %v", user.Role, RoleMember)
	}
	if user.Status != StatusActive {
		t.Errorf("CreateUser() user.Status = %v, want %v", user.Status, StatusActive)
	}
	if user.CreatedAt.IsZero() {
		t.Error("CreateUser() user.CreatedAt should not be zero")
	}
}

func TestCreateUser_DuplicateEmail(t *testing.T) {
	repo := NewRepository(testDB)
	ctx := context.Background()
	testEmail := "test_duplicate@example.com"

	// Cleanup before and after test
	cleanupTestUser(t, testEmail)
	defer cleanupTestUser(t, testEmail)

	input := &CreateUserInput{
		Email:        testEmail,
		PasswordHash: "$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/X4.",
		Role:         RoleMember,
	}

	// Create first user
	_, err := repo.CreateUser(ctx, input)
	if err != nil {
		t.Fatalf("CreateUser() first call error = %v", err)
	}

	// Try to create duplicate
	_, err = repo.CreateUser(ctx, input)
	if err != ErrUserAlreadyExists {
		t.Errorf("CreateUser() duplicate error = %v, want %v", err, ErrUserAlreadyExists)
	}
}

func TestGetUserByEmail(t *testing.T) {
	repo := NewRepository(testDB)
	ctx := context.Background()
	testEmail := "test_get_by_email@example.com"

	// Cleanup before and after test
	cleanupTestUser(t, testEmail)
	defer cleanupTestUser(t, testEmail)

	// Create test user
	input := &CreateUserInput{
		Email:        testEmail,
		PasswordHash: "$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/X4.",
		Role:         RoleAdmin,
	}
	createdUser, _ := repo.CreateUser(ctx, input)

	// Get user by email
	user, err := repo.GetUserByEmail(ctx, testEmail)
	if err != nil {
		t.Fatalf("GetUserByEmail() error = %v", err)
	}

	if user.ID != createdUser.ID {
		t.Errorf("GetUserByEmail() user.ID = %v, want %v", user.ID, createdUser.ID)
	}
	if user.Email != testEmail {
		t.Errorf("GetUserByEmail() user.Email = %v, want %v", user.Email, testEmail)
	}
	if user.Role != RoleAdmin {
		t.Errorf("GetUserByEmail() user.Role = %v, want %v", user.Role, RoleAdmin)
	}
}

func TestGetUserByEmail_NotFound(t *testing.T) {
	repo := NewRepository(testDB)
	ctx := context.Background()

	_, err := repo.GetUserByEmail(ctx, "nonexistent@example.com")
	if err != ErrUserNotFound {
		t.Errorf("GetUserByEmail() error = %v, want %v", err, ErrUserNotFound)
	}
}

func TestGetUserByID(t *testing.T) {
	repo := NewRepository(testDB)
	ctx := context.Background()
	testEmail := "test_get_by_id@example.com"

	// Cleanup before and after test
	cleanupTestUser(t, testEmail)
	defer cleanupTestUser(t, testEmail)

	// Create test user
	input := &CreateUserInput{
		Email:        testEmail,
		PasswordHash: "$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/X4.",
		Role:         RoleMember,
	}
	createdUser, _ := repo.CreateUser(ctx, input)

	// Get user by ID
	user, err := repo.GetUserByID(ctx, createdUser.ID)
	if err != nil {
		t.Fatalf("GetUserByID() error = %v", err)
	}

	if user.ID != createdUser.ID {
		t.Errorf("GetUserByID() user.ID = %v, want %v", user.ID, createdUser.ID)
	}
	if user.Email != testEmail {
		t.Errorf("GetUserByID() user.Email = %v, want %v", user.Email, testEmail)
	}
}

func TestGetUserByID_NotFound(t *testing.T) {
	repo := NewRepository(testDB)
	ctx := context.Background()

	_, err := repo.GetUserByID(ctx, "00000000-0000-0000-0000-000000000000")
	if err != ErrUserNotFound {
		t.Errorf("GetUserByID() error = %v, want %v", err, ErrUserNotFound)
	}
}

func TestUpdateUser(t *testing.T) {
	repo := NewRepository(testDB)
	ctx := context.Background()
	testEmail := "test_update_user@example.com"
	updatedEmail := "test_update_user_new@example.com"

	// Cleanup before and after test
	cleanupTestUser(t, testEmail)
	cleanupTestUser(t, updatedEmail)
	defer cleanupTestUser(t, testEmail)
	defer cleanupTestUser(t, updatedEmail)

	// Create test user
	input := &CreateUserInput{
		Email:        testEmail,
		PasswordHash: "$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/X4.",
		Role:         RoleMember,
	}
	createdUser, _ := repo.CreateUser(ctx, input)

	// Update user
	newRole := RoleAdmin
	newStatus := StatusDisabled
	updateInput := &UpdateUserInput{
		Email:  &updatedEmail,
		Role:   &newRole,
		Status: &newStatus,
	}

	user, err := repo.UpdateUser(ctx, createdUser.ID, updateInput)
	if err != nil {
		t.Fatalf("UpdateUser() error = %v", err)
	}

	if user.Email != updatedEmail {
		t.Errorf("UpdateUser() user.Email = %v, want %v", user.Email, updatedEmail)
	}
	if user.Role != RoleAdmin {
		t.Errorf("UpdateUser() user.Role = %v, want %v", user.Role, RoleAdmin)
	}
	if user.Status != StatusDisabled {
		t.Errorf("UpdateUser() user.Status = %v, want %v", user.Status, StatusDisabled)
	}
	if !user.UpdatedAt.After(user.CreatedAt) {
		t.Error("UpdateUser() user.UpdatedAt should be after CreatedAt")
	}
}

func TestUpdateUser_NotFound(t *testing.T) {
	repo := NewRepository(testDB)
	ctx := context.Background()

	newRole := RoleAdmin
	updateInput := &UpdateUserInput{
		Role: &newRole,
	}

	_, err := repo.UpdateUser(ctx, "00000000-0000-0000-0000-000000000000", updateInput)
	if err != ErrUserNotFound {
		t.Errorf("UpdateUser() error = %v, want %v", err, ErrUserNotFound)
	}
}

// ============================================================================
// Session Repository Tests
// ============================================================================

func TestCreateSession(t *testing.T) {
	repo := NewRepository(testDB)
	ctx := context.Background()
	testEmail := "test_create_session@example.com"

	// Cleanup before and after test
	cleanupTestUser(t, testEmail)
	defer cleanupTestUser(t, testEmail)

	// Create test user first
	userInput := &CreateUserInput{
		Email:        testEmail,
		PasswordHash: "$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/X4.",
		Role:         RoleMember,
	}
	user, _ := repo.CreateUser(ctx, userInput)
	defer cleanupTestSession(t, user.ID)

	// Create session
	sessionInput := &CreateSessionInput{
		UserID:       user.ID,
		RefreshToken: "test_refresh_token_12345",
		ExpiresAt:    time.Now().Add(7 * 24 * time.Hour),
	}

	session, err := repo.CreateSession(ctx, sessionInput)
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}

	if session.ID == "" {
		t.Error("CreateSession() session.ID should not be empty")
	}
	if session.UserID != user.ID {
		t.Errorf("CreateSession() session.UserID = %v, want %v", session.UserID, user.ID)
	}
	if session.RefreshToken != sessionInput.RefreshToken {
		t.Errorf("CreateSession() session.RefreshToken = %v, want %v", session.RefreshToken, sessionInput.RefreshToken)
	}
}

func TestGetSessionByRefreshToken(t *testing.T) {
	repo := NewRepository(testDB)
	ctx := context.Background()
	testEmail := "test_get_session@example.com"
	refreshToken := "test_get_session_refresh_token"

	// Cleanup before and after test
	cleanupTestUser(t, testEmail)
	defer cleanupTestUser(t, testEmail)

	// Create test user
	userInput := &CreateUserInput{
		Email:        testEmail,
		PasswordHash: "$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/X4.",
		Role:         RoleMember,
	}
	user, _ := repo.CreateUser(ctx, userInput)
	defer cleanupTestSession(t, user.ID)

	// Create session
	sessionInput := &CreateSessionInput{
		UserID:       user.ID,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(7 * 24 * time.Hour),
	}
	createdSession, _ := repo.CreateSession(ctx, sessionInput)

	// Get session by refresh token
	session, err := repo.GetSessionByRefreshToken(ctx, refreshToken)
	if err != nil {
		t.Fatalf("GetSessionByRefreshToken() error = %v", err)
	}

	if session.ID != createdSession.ID {
		t.Errorf("GetSessionByRefreshToken() session.ID = %v, want %v", session.ID, createdSession.ID)
	}
	if session.UserID != user.ID {
		t.Errorf("GetSessionByRefreshToken() session.UserID = %v, want %v", session.UserID, user.ID)
	}
}

func TestGetSessionByRefreshToken_NotFound(t *testing.T) {
	repo := NewRepository(testDB)
	ctx := context.Background()

	_, err := repo.GetSessionByRefreshToken(ctx, "nonexistent_token")
	if err != ErrSessionNotFound {
		t.Errorf("GetSessionByRefreshToken() error = %v, want %v", err, ErrSessionNotFound)
	}
}

func TestGetSessionByRefreshToken_Expired(t *testing.T) {
	repo := NewRepository(testDB)
	ctx := context.Background()
	testEmail := "test_expired_session@example.com"
	refreshToken := "test_expired_refresh_token"

	// Cleanup before and after test
	cleanupTestUser(t, testEmail)
	defer cleanupTestUser(t, testEmail)

	// Create test user
	userInput := &CreateUserInput{
		Email:        testEmail,
		PasswordHash: "$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/X4.",
		Role:         RoleMember,
	}
	user, _ := repo.CreateUser(ctx, userInput)
	defer cleanupTestSession(t, user.ID)

	// Create expired session
	sessionInput := &CreateSessionInput{
		UserID:       user.ID,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(-1 * time.Hour), // Already expired
	}
	_, _ = repo.CreateSession(ctx, sessionInput)

	// Try to get expired session
	_, err := repo.GetSessionByRefreshToken(ctx, refreshToken)
	if err != ErrSessionExpired {
		t.Errorf("GetSessionByRefreshToken() expired error = %v, want %v", err, ErrSessionExpired)
	}
}

func TestDeleteSession(t *testing.T) {
	repo := NewRepository(testDB)
	ctx := context.Background()
	testEmail := "test_delete_session@example.com"

	// Cleanup before and after test
	cleanupTestUser(t, testEmail)
	defer cleanupTestUser(t, testEmail)

	// Create test user
	userInput := &CreateUserInput{
		Email:        testEmail,
		PasswordHash: "$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/X4.",
		Role:         RoleMember,
	}
	user, _ := repo.CreateUser(ctx, userInput)
	defer cleanupTestSession(t, user.ID)

	// Create session
	sessionInput := &CreateSessionInput{
		UserID:       user.ID,
		RefreshToken: "test_delete_session_token",
		ExpiresAt:    time.Now().Add(7 * 24 * time.Hour),
	}
	session, _ := repo.CreateSession(ctx, sessionInput)

	// Delete session
	err := repo.DeleteSession(ctx, session.ID)
	if err != nil {
		t.Fatalf("DeleteSession() error = %v", err)
	}

	// Verify session is deleted
	_, err = repo.GetSessionByRefreshToken(ctx, sessionInput.RefreshToken)
	if err != ErrSessionNotFound {
		t.Errorf("GetSessionByRefreshToken() after delete error = %v, want %v", err, ErrSessionNotFound)
	}
}

func TestDeleteSession_NotFound(t *testing.T) {
	repo := NewRepository(testDB)
	ctx := context.Background()

	err := repo.DeleteSession(ctx, "00000000-0000-0000-0000-000000000000")
	if err != ErrSessionNotFound {
		t.Errorf("DeleteSession() error = %v, want %v", err, ErrSessionNotFound)
	}
}

func TestDeleteAllUserSessions(t *testing.T) {
	repo := NewRepository(testDB)
	ctx := context.Background()
	testEmail := "test_delete_all_sessions@example.com"

	// Cleanup before and after test
	cleanupTestUser(t, testEmail)
	defer cleanupTestUser(t, testEmail)

	// Create test user
	userInput := &CreateUserInput{
		Email:        testEmail,
		PasswordHash: "$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/X4.",
		Role:         RoleMember,
	}
	user, _ := repo.CreateUser(ctx, userInput)
	defer cleanupTestSession(t, user.ID)

	// Create multiple sessions
	for i := 0; i < 3; i++ {
		sessionInput := &CreateSessionInput{
			UserID:       user.ID,
			RefreshToken: "test_delete_all_token_" + string(rune('a'+i)),
			ExpiresAt:    time.Now().Add(7 * 24 * time.Hour),
		}
		_, _ = repo.CreateSession(ctx, sessionInput)
	}

	// Delete all sessions
	deleted, err := repo.DeleteAllUserSessions(ctx, user.ID)
	if err != nil {
		t.Fatalf("DeleteAllUserSessions() error = %v", err)
	}

	if deleted != 3 {
		t.Errorf("DeleteAllUserSessions() deleted = %v, want 3", deleted)
	}
}

func TestDeleteSessionByRefreshToken(t *testing.T) {
	repo := NewRepository(testDB)
	ctx := context.Background()
	testEmail := "test_delete_by_token@example.com"
	refreshToken := "test_delete_by_token_refresh"

	// Cleanup before and after test
	cleanupTestUser(t, testEmail)
	defer cleanupTestUser(t, testEmail)

	// Create test user
	userInput := &CreateUserInput{
		Email:        testEmail,
		PasswordHash: "$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/X4.",
		Role:         RoleMember,
	}
	user, _ := repo.CreateUser(ctx, userInput)
	defer cleanupTestSession(t, user.ID)

	// Create session
	sessionInput := &CreateSessionInput{
		UserID:       user.ID,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(7 * 24 * time.Hour),
	}
	_, _ = repo.CreateSession(ctx, sessionInput)

	// Delete session by refresh token
	err := repo.DeleteSessionByRefreshToken(ctx, refreshToken)
	if err != nil {
		t.Fatalf("DeleteSessionByRefreshToken() error = %v", err)
	}

	// Verify session is deleted
	_, err = repo.GetSessionByRefreshToken(ctx, refreshToken)
	if err != ErrSessionNotFound {
		t.Errorf("GetSessionByRefreshToken() after delete error = %v, want %v", err, ErrSessionNotFound)
	}
}

func TestDeleteSessionByRefreshToken_NotFound(t *testing.T) {
	repo := NewRepository(testDB)
	ctx := context.Background()

	err := repo.DeleteSessionByRefreshToken(ctx, "nonexistent_token")
	if err != ErrSessionNotFound {
		t.Errorf("DeleteSessionByRefreshToken() error = %v, want %v", err, ErrSessionNotFound)
	}
}
