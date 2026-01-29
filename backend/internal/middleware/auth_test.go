package middleware

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"

	"github.com/justinyeo/hotdesk-booking/backend/internal/shared/response"
	"github.com/justinyeo/hotdesk-booking/backend/internal/shared/utils"
)

// MockJWTValidator is a mock implementation of JWTValidator
type MockJWTValidator struct {
	ValidateFunc func(tokenString string) (*utils.Claims, error)
}

func (m *MockJWTValidator) ValidateAccessToken(tokenString string) (*utils.Claims, error) {
	if m.ValidateFunc != nil {
		return m.ValidateFunc(tokenString)
	}
	return nil, utils.ErrInvalidToken
}

// parseAPIResponse parses the API response from the response body
func parseAPIResponse(t *testing.T, body io.Reader) response.APIResponse {
	var resp response.APIResponse
	if err := json.NewDecoder(body).Decode(&resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	return resp
}

// ============================================================================
// RequireAuth Tests
// ============================================================================

func TestRequireAuth_Success(t *testing.T) {
	mockValidator := &MockJWTValidator{
		ValidateFunc: func(_ string) (*utils.Claims, error) {
			return &utils.Claims{
				UserID: "user-123",
				Role:   "member",
			}, nil
		},
	}

	app := fiber.New()
	app.Use(RequireAuth(AuthConfig{JWTValidator: mockValidator}))
	app.Get("/protected", func(c *fiber.Ctx) error {
		userID := GetUserID(c)
		role := GetRole(c)
		return c.JSON(fiber.Map{"user_id": userID, "role": role})
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer valid-token")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	var result map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if result["user_id"] != "user-123" {
		t.Errorf("expected user_id user-123, got %s", result["user_id"])
	}
	if result["role"] != "member" {
		t.Errorf("expected role member, got %s", result["role"])
	}
}

func TestRequireAuth_MissingHeader(t *testing.T) {
	mockValidator := &MockJWTValidator{}

	app := fiber.New()
	app.Use(RequireAuth(AuthConfig{JWTValidator: mockValidator}))
	app.Get("/protected", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/protected", nil)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", resp.StatusCode)
	}

	apiResp := parseAPIResponse(t, resp.Body)
	if apiResp.Success {
		t.Error("expected success to be false")
	}
	if apiResp.Error == nil || apiResp.Error.Message != "Missing authorization header" {
		t.Errorf("expected missing header error, got %v", apiResp.Error)
	}
}

func TestRequireAuth_InvalidFormat(t *testing.T) {
	mockValidator := &MockJWTValidator{}

	app := fiber.New()
	app.Use(RequireAuth(AuthConfig{JWTValidator: mockValidator}))
	app.Get("/protected", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	testCases := []struct {
		name   string
		header string
	}{
		{"no bearer", "token-only"},
		{"wrong scheme", "Basic token"},
		{"missing token", "Bearer"},
		{"extra spaces", "Bearer  token"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/protected", nil)
			req.Header.Set("Authorization", tc.header)

			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("failed to send request: %v", err)
			}
			resp.Body.Close()

			if resp.StatusCode != fiber.StatusUnauthorized {
				t.Errorf("expected status 401, got %d", resp.StatusCode)
			}
		})
	}
}

func TestRequireAuth_InvalidToken(t *testing.T) {
	mockValidator := &MockJWTValidator{
		ValidateFunc: func(_ string) (*utils.Claims, error) {
			return nil, utils.ErrInvalidToken
		},
	}

	app := fiber.New()
	app.Use(RequireAuth(AuthConfig{JWTValidator: mockValidator}))
	app.Get("/protected", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", resp.StatusCode)
	}

	apiResp := parseAPIResponse(t, resp.Body)
	if apiResp.Error == nil || apiResp.Error.Message != "Invalid access token" {
		t.Errorf("expected invalid token error, got %v", apiResp.Error)
	}
}

func TestRequireAuth_ExpiredToken(t *testing.T) {
	mockValidator := &MockJWTValidator{
		ValidateFunc: func(_ string) (*utils.Claims, error) {
			return nil, utils.ErrExpiredToken
		},
	}

	app := fiber.New()
	app.Use(RequireAuth(AuthConfig{JWTValidator: mockValidator}))
	app.Get("/protected", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer expired-token")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", resp.StatusCode)
	}

	apiResp := parseAPIResponse(t, resp.Body)
	if apiResp.Error == nil || apiResp.Error.Message != "Access token has expired" {
		t.Errorf("expected expired token error, got %v", apiResp.Error)
	}
}

func TestRequireAuth_CaseInsensitiveBearer(t *testing.T) {
	mockValidator := &MockJWTValidator{
		ValidateFunc: func(_ string) (*utils.Claims, error) {
			return &utils.Claims{
				UserID: "user-123",
				Role:   "member",
			}, nil
		},
	}

	app := fiber.New()
	app.Use(RequireAuth(AuthConfig{JWTValidator: mockValidator}))
	app.Get("/protected", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	// Test lowercase "bearer"
	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "bearer valid-token")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("failed to send request: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("expected status 200 for lowercase bearer, got %d", resp.StatusCode)
	}

	// Test uppercase "BEARER"
	req = httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "BEARER valid-token")

	resp, err = app.Test(req)
	if err != nil {
		t.Fatalf("failed to send request: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("expected status 200 for uppercase BEARER, got %d", resp.StatusCode)
	}
}

// ============================================================================
// RequireRole Tests
// ============================================================================

func TestRequireRole_Success(t *testing.T) {
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals(UserIDKey, "user-123")
		c.Locals(RoleKey, "admin")
		return c.Next()
	})
	app.Use(RequireRole("admin"))
	app.Get("/admin", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/admin", nil)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestRequireRole_InsufficientPermissions(t *testing.T) {
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals(UserIDKey, "user-123")
		c.Locals(RoleKey, "member")
		return c.Next()
	})
	app.Use(RequireRole("admin"))
	app.Get("/admin", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/admin", nil)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusForbidden {
		t.Errorf("expected status 403, got %d", resp.StatusCode)
	}

	apiResp := parseAPIResponse(t, resp.Body)
	if apiResp.Error == nil || apiResp.Error.Code != response.ErrCodeForbidden {
		t.Error("expected forbidden error")
	}
}

func TestRequireRole_NoAuth(t *testing.T) {
	app := fiber.New()
	// No auth middleware - context locals not set
	app.Use(RequireRole("admin"))
	app.Get("/admin", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/admin", nil)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", resp.StatusCode)
	}
}

func TestRequireRole_MultipleAllowedRoles(t *testing.T) {
	testCases := []struct {
		userRole     string
		expectedCode int
	}{
		{"admin", fiber.StatusOK},
		{"moderator", fiber.StatusOK},
		{"member", fiber.StatusForbidden},
	}

	for _, tc := range testCases {
		t.Run(tc.userRole, func(t *testing.T) {
			app := fiber.New()
			app.Use(func(c *fiber.Ctx) error {
				c.Locals(UserIDKey, "user-123")
				c.Locals(RoleKey, tc.userRole)
				return c.Next()
			})
			app.Use(RequireRole("admin", "moderator"))
			app.Get("/admin", func(c *fiber.Ctx) error {
				return c.SendString("ok")
			})

			req := httptest.NewRequest("GET", "/admin", nil)

			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("failed to send request: %v", err)
			}
			resp.Body.Close()

			if resp.StatusCode != tc.expectedCode {
				t.Errorf("expected status %d, got %d", tc.expectedCode, resp.StatusCode)
			}
		})
	}
}

// ============================================================================
// Helper Function Tests
// ============================================================================

func TestGetUserID(t *testing.T) {
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		c.Locals(UserIDKey, "user-123")
		userID := GetUserID(c)
		return c.SendString(userID)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if string(body) != "user-123" {
		t.Errorf("expected user-123, got %s", string(body))
	}
}

func TestGetUserID_NotSet(t *testing.T) {
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		userID := GetUserID(c)
		return c.SendString(userID)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if string(body) != "" {
		t.Errorf("expected empty string, got %s", string(body))
	}
}

func TestGetRole(t *testing.T) {
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		c.Locals(RoleKey, "admin")
		role := GetRole(c)
		return c.SendString(role)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if string(body) != "admin" {
		t.Errorf("expected admin, got %s", string(body))
	}
}

func TestGetRole_NotSet(t *testing.T) {
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		role := GetRole(c)
		return c.SendString(role)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if string(body) != "" {
		t.Errorf("expected empty string, got %s", string(body))
	}
}
