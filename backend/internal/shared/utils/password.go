// Package utils provides utility functions for authentication and security operations
// including password hashing, JWT token management, and other shared utilities.
package utils //nolint:revive // utils is acceptable for shared utility functions

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

// BcryptCost is the cost factor for bcrypt hashing (12 is recommended for security)
const BcryptCost = 12

var (
	// ErrEmptyPassword is returned when an empty password is provided
	ErrEmptyPassword = errors.New("password cannot be empty")
	// ErrHashingFailed is returned when bcrypt hashing fails
	ErrHashingFailed = errors.New("failed to hash password")
	// ErrInvalidPassword is returned when password verification fails
	ErrInvalidPassword = errors.New("invalid password")
)

// HashPassword generates a bcrypt hash from the provided password
// Returns the hashed password or an error if hashing fails
func HashPassword(password string) (string, error) {
	if password == "" {
		return "", ErrEmptyPassword
	}

	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), BcryptCost)
	if err != nil {
		return "", ErrHashingFailed
	}

	return string(hashedBytes), nil
}

// VerifyPassword compares a password with its hash
// Returns nil if the password matches, ErrInvalidPassword otherwise
func VerifyPassword(password, hash string) error {
	if password == "" {
		return ErrEmptyPassword
	}

	if hash == "" {
		return ErrInvalidPassword
	}

	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return ErrInvalidPassword
		}
		return err
	}

	return nil
}
