// Package middleware provides HTTP middleware functions for the application
// including authentication, authorization, logging, and rate limiting.
package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/justinyeo/hotdesk-booking/backend/internal/shared/response"
	"github.com/justinyeo/hotdesk-booking/backend/internal/shared/utils"
)

// Context keys for user information
const (
	UserIDKey = "user_id"
	RoleKey   = "role"
)

// JWTValidator defines the interface for JWT token validation
type JWTValidator interface {
	ValidateAccessToken(tokenString string) (*utils.Claims, error)
}

// AuthConfig holds configuration for the auth middleware
type AuthConfig struct {
	JWTValidator JWTValidator
}

// RequireAuth creates a middleware that requires a valid access token
func RequireAuth(cfg AuthConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get the Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return response.Error(c, fiber.StatusUnauthorized, response.ErrCodeUnauthorized, "Missing authorization header")
		}

		// Check for Bearer scheme
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			return response.Error(c, fiber.StatusUnauthorized, response.ErrCodeUnauthorized, "Invalid authorization header format")
		}

		tokenString := parts[1]
		if tokenString == "" {
			return response.Error(c, fiber.StatusUnauthorized, response.ErrCodeUnauthorized, "Missing access token")
		}

		// Validate the token
		claims, err := cfg.JWTValidator.ValidateAccessToken(tokenString)
		if err != nil {
			if err == utils.ErrExpiredToken {
				return response.Error(c, fiber.StatusUnauthorized, response.ErrCodeUnauthorized, "Access token has expired")
			}
			return response.Error(c, fiber.StatusUnauthorized, response.ErrCodeUnauthorized, "Invalid access token")
		}

		// Store user information in context
		c.Locals(UserIDKey, claims.UserID)
		c.Locals(RoleKey, claims.Role)

		return c.Next()
	}
}

// RequireRole creates a middleware that requires a specific role
func RequireRole(allowedRoles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get role from context (set by RequireAuth middleware)
		role, ok := c.Locals(RoleKey).(string)
		if !ok || role == "" {
			return response.Error(c, fiber.StatusUnauthorized, response.ErrCodeUnauthorized, "Authentication required")
		}

		// Check if the user's role is in the allowed list
		for _, allowedRole := range allowedRoles {
			if role == allowedRole {
				return c.Next()
			}
		}

		return response.Error(c, fiber.StatusForbidden, response.ErrCodeForbidden, "Insufficient permissions")
	}
}

// GetUserID extracts the user ID from the request context
func GetUserID(c *fiber.Ctx) string {
	if userID, ok := c.Locals(UserIDKey).(string); ok {
		return userID
	}
	return ""
}

// GetRole extracts the role from the request context
func GetRole(c *fiber.Ctx) string {
	if role, ok := c.Locals(RoleKey).(string); ok {
		return role
	}
	return ""
}
