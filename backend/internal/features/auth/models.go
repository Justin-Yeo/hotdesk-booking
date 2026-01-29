// Package auth provides authentication and authorization functionality
// including user management, session handling, and JWT token operations.
package auth

import (
	"time"
)

// UserRole represents the role of a user in the system
type UserRole string

const (
	// RoleMember is the default role for regular users
	RoleMember UserRole = "member"
	// RoleAdmin is the role for administrative users
	RoleAdmin UserRole = "admin"
)

// UserStatus represents the account status of a user
type UserStatus string

const (
	// StatusActive indicates an active user account
	StatusActive UserStatus = "active"
	// StatusDisabled indicates a disabled user account
	StatusDisabled UserStatus = "disabled"
)

// User represents a user in the system
type User struct {
	ID           string     `json:"id"`
	Email        string     `json:"email"`
	PasswordHash string     `json:"-"` // Never expose in JSON
	Role         UserRole   `json:"role"`
	Status       UserStatus `json:"status"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// Session represents a user session with a refresh token
type Session struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	RefreshToken string    `json:"-"` // Never expose in JSON
	ExpiresAt    time.Time `json:"expires_at"`
	CreatedAt    time.Time `json:"created_at"`
}

// CreateUserInput represents the input for creating a new user
type CreateUserInput struct {
	Email        string
	PasswordHash string
	Role         UserRole
}

// UpdateUserInput represents the input for updating an existing user
type UpdateUserInput struct {
	Email        *string
	PasswordHash *string
	Role         *UserRole
	Status       *UserStatus
}

// CreateSessionInput represents the input for creating a new session
type CreateSessionInput struct {
	UserID       string
	RefreshToken string
	ExpiresAt    time.Time
}
