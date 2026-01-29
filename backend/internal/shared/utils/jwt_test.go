package utils

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const testSecret = "test-secret-key-for-jwt-testing-12345"

func TestNewJWTManager(t *testing.T) {
	tests := []struct {
		name    string
		secret  string
		wantErr error
	}{
		{
			name:    "valid secret",
			secret:  testSecret,
			wantErr: nil,
		},
		{
			name:    "empty secret",
			secret:  "",
			wantErr: ErrEmptySecret,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager, err := NewJWTManager(tt.secret)

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("NewJWTManager() error = %v, wantErr %v", err, tt.wantErr)
				}
				if manager != nil {
					t.Error("NewJWTManager() should return nil manager on error")
				}
				return
			}

			if err != nil {
				t.Errorf("NewJWTManager() unexpected error = %v", err)
			}
			if manager == nil {
				t.Error("NewJWTManager() returned nil manager")
			}
		})
	}
}

func TestGenerateAccessToken(t *testing.T) {
	manager, _ := NewJWTManager(testSecret)

	tests := []struct {
		name   string
		userID string
		role   string
	}{
		{
			name:   "member user",
			userID: "user-123",
			role:   "member",
		},
		{
			name:   "admin user",
			userID: "admin-456",
			role:   "admin",
		},
		{
			name:   "uuid user id",
			userID: "550e8400-e29b-41d4-a716-446655440000",
			role:   "member",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := manager.GenerateAccessToken(tt.userID, tt.role)
			if err != nil {
				t.Fatalf("GenerateAccessToken() error = %v", err)
			}

			if token == "" {
				t.Error("GenerateAccessToken() returned empty token")
			}

			// Validate the token
			claims, err := manager.ValidateAccessToken(token)
			if err != nil {
				t.Fatalf("ValidateAccessToken() error = %v", err)
			}

			if claims.UserID != tt.userID {
				t.Errorf("claims.UserID = %v, want %v", claims.UserID, tt.userID)
			}
			if claims.Role != tt.role {
				t.Errorf("claims.Role = %v, want %v", claims.Role, tt.role)
			}
			if claims.TokenType != AccessToken {
				t.Errorf("claims.TokenType = %v, want %v", claims.TokenType, AccessToken)
			}

			// Verify expiry is approximately 15 minutes from now
			expectedExpiry := time.Now().Add(AccessTokenExpiry)
			actualExpiry := claims.ExpiresAt.Time
			diff := actualExpiry.Sub(expectedExpiry)
			if diff < -time.Second || diff > time.Second {
				t.Errorf("Token expiry differs by %v from expected", diff)
			}
		})
	}
}

func TestGenerateRefreshToken(t *testing.T) {
	manager, _ := NewJWTManager(testSecret)

	userID := "user-123"
	role := "member"

	token, err := manager.GenerateRefreshToken(userID, role)
	if err != nil {
		t.Fatalf("GenerateRefreshToken() error = %v", err)
	}

	if token == "" {
		t.Error("GenerateRefreshToken() returned empty token")
	}

	// Validate the token
	claims, err := manager.ValidateRefreshToken(token)
	if err != nil {
		t.Fatalf("ValidateRefreshToken() error = %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("claims.UserID = %v, want %v", claims.UserID, userID)
	}
	if claims.Role != role {
		t.Errorf("claims.Role = %v, want %v", claims.Role, role)
	}
	if claims.TokenType != RefreshToken {
		t.Errorf("claims.TokenType = %v, want %v", claims.TokenType, RefreshToken)
	}

	// Verify expiry is approximately 7 days from now
	expectedExpiry := time.Now().Add(RefreshTokenExpiry)
	actualExpiry := claims.ExpiresAt.Time
	diff := actualExpiry.Sub(expectedExpiry)
	if diff < -time.Second || diff > time.Second {
		t.Errorf("Token expiry differs by %v from expected", diff)
	}
}

func TestValidateToken_Invalid(t *testing.T) {
	manager, _ := NewJWTManager(testSecret)

	tests := []struct {
		name    string
		token   string
		wantErr error
	}{
		{
			name:    "empty token",
			token:   "",
			wantErr: ErrInvalidToken,
		},
		{
			name:    "malformed token",
			token:   "not.a.valid.jwt.token",
			wantErr: ErrInvalidToken,
		},
		{
			name:    "invalid signature",
			token:   "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMTIzIiwicm9sZSI6Im1lbWJlciJ9.invalidsignature",
			wantErr: ErrInvalidToken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := manager.ValidateToken(tt.token)
			if err != tt.wantErr {
				t.Errorf("ValidateToken() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateToken_WrongSecret(t *testing.T) {
	manager1, _ := NewJWTManager("secret1")
	manager2, _ := NewJWTManager("secret2")

	token, _ := manager1.GenerateAccessToken("user-123", "member")

	_, err := manager2.ValidateToken(token)
	if err != ErrInvalidToken {
		t.Errorf("ValidateToken() with wrong secret error = %v, wantErr %v", err, ErrInvalidToken)
	}
}

func TestValidateToken_ExpiredToken(t *testing.T) {
	manager, _ := NewJWTManager(testSecret)

	// Create an expired token manually
	now := time.Now()
	claims := Claims{
		UserID:    "user-123",
		Role:      "member",
		TokenType: AccessToken,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(-1 * time.Hour)), // Expired 1 hour ago
			IssuedAt:  jwt.NewNumericDate(now.Add(-2 * time.Hour)),
			NotBefore: jwt.NewNumericDate(now.Add(-2 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(testSecret))

	_, err := manager.ValidateToken(tokenString)
	if err != ErrExpiredToken {
		t.Errorf("ValidateToken() with expired token error = %v, wantErr %v", err, ErrExpiredToken)
	}
}

func TestValidateAccessToken_RejectsRefreshToken(t *testing.T) {
	manager, _ := NewJWTManager(testSecret)

	refreshToken, _ := manager.GenerateRefreshToken("user-123", "member")

	_, err := manager.ValidateAccessToken(refreshToken)
	if err != ErrInvalidToken {
		t.Errorf("ValidateAccessToken() with refresh token error = %v, wantErr %v", err, ErrInvalidToken)
	}
}

func TestValidateRefreshToken_RejectsAccessToken(t *testing.T) {
	manager, _ := NewJWTManager(testSecret)

	accessToken, _ := manager.GenerateAccessToken("user-123", "member")

	_, err := manager.ValidateRefreshToken(accessToken)
	if err != ErrInvalidToken {
		t.Errorf("ValidateRefreshToken() with access token error = %v, wantErr %v", err, ErrInvalidToken)
	}
}

func TestGenerateTokenPair(t *testing.T) {
	manager, _ := NewJWTManager(testSecret)

	userID := "user-123"
	role := "admin"

	accessToken, refreshToken, err := manager.GenerateTokenPair(userID, role)
	if err != nil {
		t.Fatalf("GenerateTokenPair() error = %v", err)
	}

	if accessToken == "" {
		t.Error("GenerateTokenPair() returned empty access token")
	}
	if refreshToken == "" {
		t.Error("GenerateTokenPair() returned empty refresh token")
	}

	// Tokens should be different
	if accessToken == refreshToken {
		t.Error("GenerateTokenPair() access and refresh tokens should be different")
	}

	// Validate access token
	accessClaims, err := manager.ValidateAccessToken(accessToken)
	if err != nil {
		t.Fatalf("ValidateAccessToken() error = %v", err)
	}
	if accessClaims.UserID != userID {
		t.Errorf("access token UserID = %v, want %v", accessClaims.UserID, userID)
	}

	// Validate refresh token
	refreshClaims, err := manager.ValidateRefreshToken(refreshToken)
	if err != nil {
		t.Fatalf("ValidateRefreshToken() error = %v", err)
	}
	if refreshClaims.UserID != userID {
		t.Errorf("refresh token UserID = %v, want %v", refreshClaims.UserID, userID)
	}
}

func TestClaims_ContainsRequiredFields(t *testing.T) {
	manager, _ := NewJWTManager(testSecret)

	token, _ := manager.GenerateAccessToken("user-123", "member")
	claims, _ := manager.ValidateAccessToken(token)

	// Verify required fields are present
	if claims.UserID == "" {
		t.Error("claims.UserID should not be empty")
	}
	if claims.Role == "" {
		t.Error("claims.Role should not be empty")
	}
	if claims.ExpiresAt == nil {
		t.Error("claims.ExpiresAt should not be nil")
	}
	if claims.IssuedAt == nil {
		t.Error("claims.IssuedAt should not be nil")
	}
}

// Benchmark tests
func BenchmarkGenerateAccessToken(b *testing.B) {
	manager, _ := NewJWTManager(testSecret)
	for i := 0; i < b.N; i++ {
		_, _ = manager.GenerateAccessToken("user-123", "member")
	}
}

func BenchmarkValidateToken(b *testing.B) {
	manager, _ := NewJWTManager(testSecret)
	token, _ := manager.GenerateAccessToken("user-123", "member")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = manager.ValidateToken(token)
	}
}
