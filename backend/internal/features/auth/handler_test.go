package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/justinyeo/hotdesk-booking/backend/internal/shared/response"
	"github.com/justinyeo/hotdesk-booking/backend/internal/shared/utils"
)

// setupTestApp creates a Fiber app for testing
func setupTestApp(handler *Handler) *fiber.App {
	app := fiber.New()
	api := app.Group("/api/v1/auth")
	api.Post("/register", handler.Register)
	api.Post("/login", handler.Login)
	api.Post("/refresh", handler.Refresh)
	api.Post("/logout", handler.Logout)
	api.Post("/logout-all", func(c *fiber.Ctx) error {
		// Simulate auth middleware setting user_id
		c.Locals("user_id", "user-123")
		return handler.LogoutAll(c)
	})
	return app
}

// parseResponse parses the API response
func parseResponse(t *testing.T, body io.Reader) response.APIResponse {
	var resp response.APIResponse
	if err := json.NewDecoder(body).Decode(&resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	return resp
}

// ============================================================================
// Register Handler Tests
// ============================================================================

func TestHandler_Register_Success(t *testing.T) {
	mockRepo := &MockRepository{
		CreateUserFunc: func(_ context.Context, input *CreateUserInput) (*User, error) {
			return &User{
				ID:        "user-123",
				Email:     input.Email,
				Role:      input.Role,
				Status:    StatusActive,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}, nil
		},
		CreateSessionFunc: func(_ context.Context, _ *CreateSessionInput) (*Session, error) {
			return &Session{ID: "session-123"}, nil
		},
	}
	mockJWT := &MockJWTManager{}
	service := NewService(mockRepo, mockJWT)
	handler := NewHandler(service)
	app := setupTestApp(handler)

	reqBody := `{"email":"test@example.com","password":"password123"}`
	req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusCreated {
		t.Errorf("expected status 201, got %d", resp.StatusCode)
	}

	apiResp := parseResponse(t, resp.Body)
	if !apiResp.Success {
		t.Error("expected success to be true")
	}
	if apiResp.Error != nil {
		t.Errorf("expected no error, got %v", apiResp.Error)
	}
}

func TestHandler_Register_InvalidEmail(t *testing.T) {
	service := NewService(&MockRepository{}, &MockJWTManager{})
	handler := NewHandler(service)
	app := setupTestApp(handler)

	reqBody := `{"email":"invalid-email","password":"password123"}`
	req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusBadRequest {
		t.Errorf("expected status 400, got %d", resp.StatusCode)
	}

	apiResp := parseResponse(t, resp.Body)
	if apiResp.Success {
		t.Error("expected success to be false")
	}
	if apiResp.Error == nil || apiResp.Error.Code != response.ErrCodeValidation {
		t.Error("expected validation error")
	}
}

func TestHandler_Register_MissingEmail(t *testing.T) {
	service := NewService(&MockRepository{}, &MockJWTManager{})
	handler := NewHandler(service)
	app := setupTestApp(handler)

	reqBody := `{"password":"password123"}`
	req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusBadRequest {
		t.Errorf("expected status 400, got %d", resp.StatusCode)
	}
}

func TestHandler_Register_MissingPassword(t *testing.T) {
	service := NewService(&MockRepository{}, &MockJWTManager{})
	handler := NewHandler(service)
	app := setupTestApp(handler)

	reqBody := `{"email":"test@example.com"}`
	req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusBadRequest {
		t.Errorf("expected status 400, got %d", resp.StatusCode)
	}
}

func TestHandler_Register_DuplicateEmail(t *testing.T) {
	mockRepo := &MockRepository{
		CreateUserFunc: func(_ context.Context, _ *CreateUserInput) (*User, error) {
			return nil, ErrUserAlreadyExists
		},
	}
	service := NewService(mockRepo, &MockJWTManager{})
	handler := NewHandler(service)
	app := setupTestApp(handler)

	reqBody := `{"email":"existing@example.com","password":"password123"}`
	req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusConflict {
		t.Errorf("expected status 409, got %d", resp.StatusCode)
	}
}

func TestHandler_Register_InvalidJSON(t *testing.T) {
	service := NewService(&MockRepository{}, &MockJWTManager{})
	handler := NewHandler(service)
	app := setupTestApp(handler)

	reqBody := `{invalid json}`
	req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusBadRequest {
		t.Errorf("expected status 400, got %d", resp.StatusCode)
	}
}

// ============================================================================
// Login Handler Tests
// ============================================================================

func TestHandler_Login_Success(t *testing.T) {
	hashedPassword, _ := utils.HashPassword("password123")
	mockRepo := &MockRepository{
		GetUserByEmailFunc: func(_ context.Context, _ string) (*User, error) {
			return &User{
				ID:           "user-123",
				Email:        "test@example.com",
				PasswordHash: hashedPassword,
				Role:         RoleMember,
				Status:       StatusActive,
				CreatedAt:    time.Now(),
			}, nil
		},
		CreateSessionFunc: func(_ context.Context, _ *CreateSessionInput) (*Session, error) {
			return &Session{ID: "session-123"}, nil
		},
	}
	mockJWT := &MockJWTManager{}
	service := NewService(mockRepo, mockJWT)
	handler := NewHandler(service)
	app := setupTestApp(handler)

	reqBody := `{"email":"test@example.com","password":"password123"}`
	req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	apiResp := parseResponse(t, resp.Body)
	if !apiResp.Success {
		t.Error("expected success to be true")
	}
}

func TestHandler_Login_InvalidCredentials(t *testing.T) {
	mockRepo := &MockRepository{
		GetUserByEmailFunc: func(_ context.Context, _ string) (*User, error) {
			return nil, ErrUserNotFound
		},
	}
	service := NewService(mockRepo, &MockJWTManager{})
	handler := NewHandler(service)
	app := setupTestApp(handler)

	reqBody := `{"email":"nonexistent@example.com","password":"password123"}`
	req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", resp.StatusCode)
	}
}

func TestHandler_Login_MissingFields(t *testing.T) {
	service := NewService(&MockRepository{}, &MockJWTManager{})
	handler := NewHandler(service)
	app := setupTestApp(handler)

	// Test missing email
	reqBody := `{"password":"password123"}`
	req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("failed to send request: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != fiber.StatusBadRequest {
		t.Errorf("expected status 400, got %d", resp.StatusCode)
	}
}

// ============================================================================
// Refresh Handler Tests
// ============================================================================

func TestHandler_Refresh_Success(t *testing.T) {
	mockRepo := &MockRepository{
		GetSessionByRefreshTokenFunc: func(_ context.Context, _ string) (*Session, error) {
			return &Session{
				ID:        "session-123",
				UserID:    "user-123",
				ExpiresAt: time.Now().Add(time.Hour),
			}, nil
		},
		GetUserByIDFunc: func(_ context.Context, _ string) (*User, error) {
			return &User{
				ID:     "user-123",
				Status: StatusActive,
			}, nil
		},
		DeleteSessionFunc: func(_ context.Context, _ string) error {
			return nil
		},
		CreateSessionFunc: func(_ context.Context, _ *CreateSessionInput) (*Session, error) {
			return &Session{ID: "session-456"}, nil
		},
	}
	mockJWT := &MockJWTManager{
		ValidateRefreshTokenFunc: func(_ string) (*utils.Claims, error) {
			return &utils.Claims{UserID: "user-123", Role: "member"}, nil
		},
		GenerateTokenPairFunc: func(_, _ string) (string, string, error) {
			return "new-access", "new-refresh", nil
		},
	}
	service := NewService(mockRepo, mockJWT)
	handler := NewHandler(service)
	app := setupTestApp(handler)

	reqBody := `{"refresh_token":"valid-refresh-token"}`
	req := httptest.NewRequest("POST", "/api/v1/auth/refresh", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestHandler_Refresh_InvalidToken(t *testing.T) {
	mockJWT := &MockJWTManager{
		ValidateRefreshTokenFunc: func(_ string) (*utils.Claims, error) {
			return nil, utils.ErrInvalidToken
		},
	}
	service := NewService(&MockRepository{}, mockJWT)
	handler := NewHandler(service)
	app := setupTestApp(handler)

	reqBody := `{"refresh_token":"invalid-token"}`
	req := httptest.NewRequest("POST", "/api/v1/auth/refresh", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", resp.StatusCode)
	}
}

func TestHandler_Refresh_MissingToken(t *testing.T) {
	service := NewService(&MockRepository{}, &MockJWTManager{})
	handler := NewHandler(service)
	app := setupTestApp(handler)

	reqBody := `{}`
	req := httptest.NewRequest("POST", "/api/v1/auth/refresh", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusBadRequest {
		t.Errorf("expected status 400, got %d", resp.StatusCode)
	}
}

// ============================================================================
// Logout Handler Tests
// ============================================================================

func TestHandler_Logout_Success(t *testing.T) {
	mockRepo := &MockRepository{
		DeleteSessionByRefreshTokenFunc: func(_ context.Context, _ string) error {
			return nil
		},
	}
	service := NewService(mockRepo, &MockJWTManager{})
	handler := NewHandler(service)
	app := setupTestApp(handler)

	reqBody := `{"refresh_token":"valid-refresh-token"}`
	req := httptest.NewRequest("POST", "/api/v1/auth/logout", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestHandler_Logout_SessionNotFound(t *testing.T) {
	mockRepo := &MockRepository{
		DeleteSessionByRefreshTokenFunc: func(_ context.Context, _ string) error {
			return ErrSessionNotFound
		},
	}
	service := NewService(mockRepo, &MockJWTManager{})
	handler := NewHandler(service)
	app := setupTestApp(handler)

	reqBody := `{"refresh_token":"invalid-token"}`
	req := httptest.NewRequest("POST", "/api/v1/auth/logout", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Logout should succeed even if session not found (idempotent)
	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestHandler_Logout_MissingToken(t *testing.T) {
	service := NewService(&MockRepository{}, &MockJWTManager{})
	handler := NewHandler(service)
	app := setupTestApp(handler)

	reqBody := `{}`
	req := httptest.NewRequest("POST", "/api/v1/auth/logout", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusBadRequest {
		t.Errorf("expected status 400, got %d", resp.StatusCode)
	}
}

// ============================================================================
// LogoutAll Handler Tests
// ============================================================================

func TestHandler_LogoutAll_Success(t *testing.T) {
	mockRepo := &MockRepository{
		DeleteAllUserSessionsFunc: func(_ context.Context, _ string) (int64, error) {
			return 3, nil
		},
	}
	service := NewService(mockRepo, &MockJWTManager{})
	handler := NewHandler(service)
	app := setupTestApp(handler)

	req := httptest.NewRequest("POST", "/api/v1/auth/logout-all", nil)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	apiResp := parseResponse(t, resp.Body)
	if !apiResp.Success {
		t.Error("expected success to be true")
	}
}

func TestHandler_LogoutAll_NoAuth(t *testing.T) {
	mockRepo := &MockRepository{}
	service := NewService(mockRepo, &MockJWTManager{})
	handler := NewHandler(service)

	// Create app without setting user_id in context
	app := fiber.New()
	app.Post("/api/v1/auth/logout-all", handler.LogoutAll)

	req := httptest.NewRequest("POST", "/api/v1/auth/logout-all", nil)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", resp.StatusCode)
	}
}
