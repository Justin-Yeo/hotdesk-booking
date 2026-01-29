package auth

import (
	"errors"

	"github.com/gofiber/fiber/v2"

	"github.com/justinyeo/hotdesk-booking/backend/internal/shared/response"
)

// Handler handles HTTP requests for authentication
type Handler struct {
	service *Service
}

// NewHandler creates a new auth handler
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// RegisterRequest represents the request body for registration
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginRequest represents the request body for login
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RefreshRequest represents the request body for token refresh
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// LogoutRequest represents the request body for logout
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// UserResponse represents the user data in responses (without sensitive fields)
type UserResponse struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

// LoginResponse represents the response for register and login
type LoginResponse struct {
	User         UserResponse `json:"user"`
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
}

// TokenResponse represents the response for token refresh
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// LogoutAllResponse represents the response for logout all
type LogoutAllResponse struct {
	SessionsRevoked int64 `json:"sessions_revoked"`
}

// Register handles POST /api/v1/auth/register
func (h *Handler) Register(c *fiber.Ctx) error {
	var req RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, response.ErrCodeBadRequest, "Invalid request body")
	}

	// Validate required fields
	if req.Email == "" {
		return response.Error(c, fiber.StatusBadRequest, response.ErrCodeValidation, "Email is required")
	}
	if req.Password == "" {
		return response.Error(c, fiber.StatusBadRequest, response.ErrCodeValidation, "Password is required")
	}

	result, err := h.service.Register(c.Context(), &RegisterInput{
		Email:    req.Email,
		Password: req.Password,
	})

	if err != nil {
		return h.handleServiceError(c, err)
	}

	return response.Success(c, fiber.StatusCreated, LoginResponse{
		User:         toUserResponse(result.User),
		AccessToken:  result.Tokens.AccessToken,
		RefreshToken: result.Tokens.RefreshToken,
	})
}

// Login handles POST /api/v1/auth/login
func (h *Handler) Login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, response.ErrCodeBadRequest, "Invalid request body")
	}

	// Validate required fields
	if req.Email == "" {
		return response.Error(c, fiber.StatusBadRequest, response.ErrCodeValidation, "Email is required")
	}
	if req.Password == "" {
		return response.Error(c, fiber.StatusBadRequest, response.ErrCodeValidation, "Password is required")
	}

	result, err := h.service.Login(c.Context(), &LoginInput{
		Email:    req.Email,
		Password: req.Password,
	})

	if err != nil {
		return h.handleServiceError(c, err)
	}

	return response.Success(c, fiber.StatusOK, LoginResponse{
		User:         toUserResponse(result.User),
		AccessToken:  result.Tokens.AccessToken,
		RefreshToken: result.Tokens.RefreshToken,
	})
}

// Refresh handles POST /api/v1/auth/refresh
func (h *Handler) Refresh(c *fiber.Ctx) error {
	var req RefreshRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, response.ErrCodeBadRequest, "Invalid request body")
	}

	if req.RefreshToken == "" {
		return response.Error(c, fiber.StatusBadRequest, response.ErrCodeValidation, "Refresh token is required")
	}

	tokens, err := h.service.RefreshToken(c.Context(), req.RefreshToken)
	if err != nil {
		return h.handleServiceError(c, err)
	}

	return response.Success(c, fiber.StatusOK, TokenResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	})
}

// Logout handles POST /api/v1/auth/logout
func (h *Handler) Logout(c *fiber.Ctx) error {
	var req LogoutRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, response.ErrCodeBadRequest, "Invalid request body")
	}

	if req.RefreshToken == "" {
		return response.Error(c, fiber.StatusBadRequest, response.ErrCodeValidation, "Refresh token is required")
	}

	err := h.service.Logout(c.Context(), req.RefreshToken)
	if err != nil {
		// Session not found is acceptable for logout - client may have already logged out
		if errors.Is(err, ErrSessionNotFound) {
			return response.Success(c, fiber.StatusOK, nil)
		}
		return h.handleServiceError(c, err)
	}

	return response.Success(c, fiber.StatusOK, nil)
}

// LogoutAll handles POST /api/v1/auth/logout-all
// This endpoint requires authentication (user ID from context)
func (h *Handler) LogoutAll(c *fiber.Ctx) error {
	// Get user ID from context (set by auth middleware)
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return response.Error(c, fiber.StatusUnauthorized, response.ErrCodeUnauthorized, "Authentication required")
	}

	count, err := h.service.LogoutAll(c.Context(), userID)
	if err != nil {
		return h.handleServiceError(c, err)
	}

	return response.Success(c, fiber.StatusOK, LogoutAllResponse{
		SessionsRevoked: count,
	})
}

// handleServiceError converts service errors to HTTP responses
func (h *Handler) handleServiceError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, ErrInvalidEmail):
		return response.Error(c, fiber.StatusBadRequest, response.ErrCodeValidation, "Invalid email format")
	case errors.Is(err, ErrPasswordTooShort):
		return response.Error(c, fiber.StatusBadRequest, response.ErrCodeValidation, "Password must be at least 8 characters")
	case errors.Is(err, ErrUserAlreadyExists):
		return response.Error(c, fiber.StatusConflict, response.ErrCodeConflict, "User with this email already exists")
	case errors.Is(err, ErrInvalidCredentials):
		return response.Error(c, fiber.StatusUnauthorized, response.ErrCodeUnauthorized, "Invalid email or password")
	case errors.Is(err, ErrUserDisabled):
		return response.Error(c, fiber.StatusForbidden, response.ErrCodeForbidden, "User account is disabled")
	case errors.Is(err, ErrInvalidRefreshToken):
		return response.Error(c, fiber.StatusUnauthorized, response.ErrCodeUnauthorized, "Invalid or expired refresh token")
	case errors.Is(err, ErrSessionNotFound):
		return response.Error(c, fiber.StatusUnauthorized, response.ErrCodeUnauthorized, "Session not found")
	default:
		return response.Error(c, fiber.StatusInternalServerError, response.ErrCodeInternalServer, "An unexpected error occurred")
	}
}

// toUserResponse converts a User model to UserResponse
func toUserResponse(user *User) UserResponse {
	return UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		Role:      string(user.Role),
		Status:    string(user.Status),
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
