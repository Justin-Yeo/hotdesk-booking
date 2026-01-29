package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	// ErrUserNotFound is returned when a user is not found
	ErrUserNotFound = errors.New("user not found")
	// ErrUserAlreadyExists is returned when attempting to create a user with an existing email
	ErrUserAlreadyExists = errors.New("user with this email already exists")
	// ErrSessionNotFound is returned when a session is not found
	ErrSessionNotFound = errors.New("session not found")
	// ErrSessionExpired is returned when a session has expired
	ErrSessionExpired = errors.New("session has expired")
)

// Repository provides database operations for auth-related entities
type Repository struct {
	db *pgxpool.Pool
}

// NewRepository creates a new auth repository
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// ============================================================================
// User Repository Methods
// ============================================================================

// CreateUser inserts a new user into the database
func (r *Repository) CreateUser(ctx context.Context, input *CreateUserInput) (*User, error) {
	query := `
		INSERT INTO users (email, password_hash, role)
		VALUES ($1, $2, $3)
		RETURNING id, email, password_hash, role, status, created_at, updated_at
	`

	var user User
	err := r.db.QueryRow(ctx, query, input.Email, input.PasswordHash, input.Role).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.Status,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		// Check for unique constraint violation
		if isDuplicateKeyError(err) {
			return nil, ErrUserAlreadyExists
		}
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &user, nil
}

// GetUserByEmail retrieves a user by their email address
func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	query := `
		SELECT id, email, password_hash, role, status, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	var user User
	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.Status,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return &user, nil
}

// GetUserByID retrieves a user by their ID
func (r *Repository) GetUserByID(ctx context.Context, id string) (*User, error) {
	query := `
		SELECT id, email, password_hash, role, status, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	var user User
	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.Status,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	return &user, nil
}

// UpdateUser updates an existing user's fields
func (r *Repository) UpdateUser(ctx context.Context, id string, input *UpdateUserInput) (*User, error) {
	// Build dynamic update query based on provided fields
	query := `
		UPDATE users
		SET updated_at = NOW()
	`
	args := []interface{}{}
	argNum := 1

	if input.Email != nil {
		query += fmt.Sprintf(", email = $%d", argNum)
		args = append(args, *input.Email)
		argNum++
	}

	if input.PasswordHash != nil {
		query += fmt.Sprintf(", password_hash = $%d", argNum)
		args = append(args, *input.PasswordHash)
		argNum++
	}

	if input.Role != nil {
		query += fmt.Sprintf(", role = $%d", argNum)
		args = append(args, *input.Role)
		argNum++
	}

	if input.Status != nil {
		query += fmt.Sprintf(", status = $%d", argNum)
		args = append(args, *input.Status)
		argNum++
	}

	query += fmt.Sprintf(`
		WHERE id = $%d
		RETURNING id, email, password_hash, role, status, created_at, updated_at
	`, argNum)
	args = append(args, id)

	var user User
	err := r.db.QueryRow(ctx, query, args...).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.Status,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		if isDuplicateKeyError(err) {
			return nil, ErrUserAlreadyExists
		}
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return &user, nil
}

// ============================================================================
// Session Repository Methods
// ============================================================================

// CreateSession inserts a new session into the database
func (r *Repository) CreateSession(ctx context.Context, input *CreateSessionInput) (*Session, error) {
	query := `
		INSERT INTO sessions (user_id, refresh_token, expires_at)
		VALUES ($1, $2, $3)
		RETURNING id, user_id, refresh_token, expires_at, created_at
	`

	var session Session
	err := r.db.QueryRow(ctx, query, input.UserID, input.RefreshToken, input.ExpiresAt).Scan(
		&session.ID,
		&session.UserID,
		&session.RefreshToken,
		&session.ExpiresAt,
		&session.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return &session, nil
}

// GetSessionByRefreshToken retrieves a session by its refresh token
func (r *Repository) GetSessionByRefreshToken(ctx context.Context, refreshToken string) (*Session, error) {
	query := `
		SELECT id, user_id, refresh_token, expires_at, created_at
		FROM sessions
		WHERE refresh_token = $1
	`

	var session Session
	err := r.db.QueryRow(ctx, query, refreshToken).Scan(
		&session.ID,
		&session.UserID,
		&session.RefreshToken,
		&session.ExpiresAt,
		&session.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrSessionNotFound
		}
		return nil, fmt.Errorf("failed to get session by refresh token: %w", err)
	}

	// Check if session has expired
	if time.Now().After(session.ExpiresAt) {
		return nil, ErrSessionExpired
	}

	return &session, nil
}

// DeleteSession removes a single session by its ID
func (r *Repository) DeleteSession(ctx context.Context, sessionID string) error {
	query := `DELETE FROM sessions WHERE id = $1`

	result, err := r.db.Exec(ctx, query, sessionID)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrSessionNotFound
	}

	return nil
}

// DeleteSessionByRefreshToken removes a session by its refresh token
func (r *Repository) DeleteSessionByRefreshToken(ctx context.Context, refreshToken string) error {
	query := `DELETE FROM sessions WHERE refresh_token = $1`

	result, err := r.db.Exec(ctx, query, refreshToken)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrSessionNotFound
	}

	return nil
}

// DeleteAllUserSessions removes all sessions for a given user
func (r *Repository) DeleteAllUserSessions(ctx context.Context, userID string) (int64, error) {
	query := `DELETE FROM sessions WHERE user_id = $1`

	result, err := r.db.Exec(ctx, query, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to delete all user sessions: %w", err)
	}

	return result.RowsAffected(), nil
}

// DeleteExpiredSessions removes all expired sessions (for cleanup jobs)
func (r *Repository) DeleteExpiredSessions(ctx context.Context) (int64, error) {
	query := `DELETE FROM sessions WHERE expires_at < NOW()`

	result, err := r.db.Exec(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("failed to delete expired sessions: %w", err)
	}

	return result.RowsAffected(), nil
}

// ============================================================================
// Helper Functions
// ============================================================================

// isDuplicateKeyError checks if the error is a PostgreSQL unique constraint violation
func isDuplicateKeyError(err error) bool {
	// PostgreSQL error code for unique_violation is 23505
	return err != nil && (contains(err.Error(), "23505") || contains(err.Error(), "duplicate key"))
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
