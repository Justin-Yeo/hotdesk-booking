package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/justinyeo/hotdesk-booking/backend/internal/shared/utils"
)

// MockRepository is a mock implementation of RepositoryInterface
type MockRepository struct {
	CreateUserFunc                  func(ctx context.Context, input *CreateUserInput) (*User, error)
	GetUserByEmailFunc              func(ctx context.Context, email string) (*User, error)
	GetUserByIDFunc                 func(ctx context.Context, id string) (*User, error)
	CreateSessionFunc               func(ctx context.Context, input *CreateSessionInput) (*Session, error)
	GetSessionByRefreshTokenFunc    func(ctx context.Context, refreshToken string) (*Session, error)
	DeleteSessionFunc               func(ctx context.Context, sessionID string) error
	DeleteSessionByRefreshTokenFunc func(ctx context.Context, refreshToken string) error
	DeleteAllUserSessionsFunc       func(ctx context.Context, userID string) (int64, error)
}

func (m *MockRepository) CreateUser(ctx context.Context, input *CreateUserInput) (*User, error) {
	if m.CreateUserFunc != nil {
		return m.CreateUserFunc(ctx, input)
	}
	return nil, nil
}

func (m *MockRepository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	if m.GetUserByEmailFunc != nil {
		return m.GetUserByEmailFunc(ctx, email)
	}
	return nil, nil
}

func (m *MockRepository) GetUserByID(ctx context.Context, id string) (*User, error) {
	if m.GetUserByIDFunc != nil {
		return m.GetUserByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockRepository) CreateSession(ctx context.Context, input *CreateSessionInput) (*Session, error) {
	if m.CreateSessionFunc != nil {
		return m.CreateSessionFunc(ctx, input)
	}
	return nil, nil
}

func (m *MockRepository) GetSessionByRefreshToken(ctx context.Context, refreshToken string) (*Session, error) {
	if m.GetSessionByRefreshTokenFunc != nil {
		return m.GetSessionByRefreshTokenFunc(ctx, refreshToken)
	}
	return nil, nil
}

func (m *MockRepository) DeleteSession(ctx context.Context, sessionID string) error {
	if m.DeleteSessionFunc != nil {
		return m.DeleteSessionFunc(ctx, sessionID)
	}
	return nil
}

func (m *MockRepository) DeleteSessionByRefreshToken(ctx context.Context, refreshToken string) error {
	if m.DeleteSessionByRefreshTokenFunc != nil {
		return m.DeleteSessionByRefreshTokenFunc(ctx, refreshToken)
	}
	return nil
}

func (m *MockRepository) DeleteAllUserSessions(ctx context.Context, userID string) (int64, error) {
	if m.DeleteAllUserSessionsFunc != nil {
		return m.DeleteAllUserSessionsFunc(ctx, userID)
	}
	return 0, nil
}

// MockJWTManager is a mock implementation of JWTManagerInterface
type MockJWTManager struct {
	GenerateAccessTokenFunc  func(userID, role string) (string, error)
	GenerateRefreshTokenFunc func(userID, role string) (string, error)
	GenerateTokenPairFunc    func(userID, role string) (accessToken, refreshToken string, err error)
	ValidateRefreshTokenFunc func(tokenString string) (*utils.Claims, error)
}

func (m *MockJWTManager) GenerateAccessToken(userID, role string) (string, error) {
	if m.GenerateAccessTokenFunc != nil {
		return m.GenerateAccessTokenFunc(userID, role)
	}
	return "mock-access-token", nil
}

func (m *MockJWTManager) GenerateRefreshToken(userID, role string) (string, error) {
	if m.GenerateRefreshTokenFunc != nil {
		return m.GenerateRefreshTokenFunc(userID, role)
	}
	return "mock-refresh-token", nil
}

func (m *MockJWTManager) GenerateTokenPair(userID, role string) (accessToken, refreshToken string, err error) {
	if m.GenerateTokenPairFunc != nil {
		return m.GenerateTokenPairFunc(userID, role)
	}
	return "mock-access-token", "mock-refresh-token", nil
}

func (m *MockJWTManager) ValidateRefreshToken(tokenString string) (*utils.Claims, error) {
	if m.ValidateRefreshTokenFunc != nil {
		return m.ValidateRefreshTokenFunc(tokenString)
	}
	return &utils.Claims{UserID: "user-123", Role: "member"}, nil
}

// ============================================================================
// Register Tests
// ============================================================================

func TestRegister_Success(t *testing.T) {
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
	result, err := service.Register(context.Background(), &RegisterInput{
		Email:    "test@example.com",
		Password: "password123",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.User.Email != "test@example.com" {
		t.Errorf("expected email test@example.com, got %s", result.User.Email)
	}
	if result.Tokens.AccessToken == "" {
		t.Error("expected access token to be set")
	}
	if result.Tokens.RefreshToken == "" {
		t.Error("expected refresh token to be set")
	}
}

func TestRegister_InvalidEmail(t *testing.T) {
	service := NewService(&MockRepository{}, &MockJWTManager{})

	testCases := []string{
		"invalid",
		"invalid@",
		"@example.com",
		"invalid@.com",
		"",
	}

	for _, email := range testCases {
		_, err := service.Register(context.Background(), &RegisterInput{
			Email:    email,
			Password: "password123",
		})

		if !errors.Is(err, ErrInvalidEmail) {
			t.Errorf("expected ErrInvalidEmail for email %q, got %v", email, err)
		}
	}
}

func TestRegister_PasswordTooShort(t *testing.T) {
	service := NewService(&MockRepository{}, &MockJWTManager{})

	_, err := service.Register(context.Background(), &RegisterInput{
		Email:    "test@example.com",
		Password: "short",
	})

	if !errors.Is(err, ErrPasswordTooShort) {
		t.Errorf("expected ErrPasswordTooShort, got %v", err)
	}
}

func TestRegister_DuplicateEmail(t *testing.T) {
	mockRepo := &MockRepository{
		CreateUserFunc: func(_ context.Context, _ *CreateUserInput) (*User, error) {
			return nil, ErrUserAlreadyExists
		},
	}

	service := NewService(mockRepo, &MockJWTManager{})
	_, err := service.Register(context.Background(), &RegisterInput{
		Email:    "existing@example.com",
		Password: "password123",
	})

	if !errors.Is(err, ErrUserAlreadyExists) {
		t.Errorf("expected ErrUserAlreadyExists, got %v", err)
	}
}

// ============================================================================
// Login Tests
// ============================================================================

func TestLogin_Success(t *testing.T) {
	// Pre-hash the password for comparison
	hashedPassword, _ := utils.HashPassword("password123")

	mockRepo := &MockRepository{
		GetUserByEmailFunc: func(_ context.Context, _ string) (*User, error) {
			return &User{
				ID:           "user-123",
				Email:        "test@example.com",
				PasswordHash: hashedPassword,
				Role:         RoleMember,
				Status:       StatusActive,
			}, nil
		},
		CreateSessionFunc: func(_ context.Context, _ *CreateSessionInput) (*Session, error) {
			return &Session{ID: "session-123"}, nil
		},
	}
	mockJWT := &MockJWTManager{}

	service := NewService(mockRepo, mockJWT)
	result, err := service.Login(context.Background(), &LoginInput{
		Email:    "test@example.com",
		Password: "password123",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.User.Email != "test@example.com" {
		t.Errorf("expected email test@example.com, got %s", result.User.Email)
	}
	if result.Tokens.AccessToken == "" {
		t.Error("expected access token to be set")
	}
}

func TestLogin_UserNotFound(t *testing.T) {
	mockRepo := &MockRepository{
		GetUserByEmailFunc: func(_ context.Context, _ string) (*User, error) {
			return nil, ErrUserNotFound
		},
	}

	service := NewService(mockRepo, &MockJWTManager{})
	_, err := service.Login(context.Background(), &LoginInput{
		Email:    "nonexistent@example.com",
		Password: "password123",
	})

	if !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	hashedPassword, _ := utils.HashPassword("correctpassword")

	mockRepo := &MockRepository{
		GetUserByEmailFunc: func(_ context.Context, _ string) (*User, error) {
			return &User{
				ID:           "user-123",
				Email:        "test@example.com",
				PasswordHash: hashedPassword,
				Status:       StatusActive,
			}, nil
		},
	}

	service := NewService(mockRepo, &MockJWTManager{})
	_, err := service.Login(context.Background(), &LoginInput{
		Email:    "test@example.com",
		Password: "wrongpassword",
	})

	if !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestLogin_UserDisabled(t *testing.T) {
	hashedPassword, _ := utils.HashPassword("password123")

	mockRepo := &MockRepository{
		GetUserByEmailFunc: func(_ context.Context, _ string) (*User, error) {
			return &User{
				ID:           "user-123",
				Email:        "test@example.com",
				PasswordHash: hashedPassword,
				Status:       StatusDisabled,
			}, nil
		},
	}

	service := NewService(mockRepo, &MockJWTManager{})
	_, err := service.Login(context.Background(), &LoginInput{
		Email:    "test@example.com",
		Password: "password123",
	})

	if !errors.Is(err, ErrUserDisabled) {
		t.Errorf("expected ErrUserDisabled, got %v", err)
	}
}

// ============================================================================
// RefreshToken Tests
// ============================================================================

func TestRefreshToken_Success(t *testing.T) {
	mockRepo := &MockRepository{
		GetSessionByRefreshTokenFunc: func(_ context.Context, _ string) (*Session, error) {
			return &Session{
				ID:           "session-123",
				UserID:       "user-123",
				RefreshToken: "old-refresh-token",
				ExpiresAt:    time.Now().Add(time.Hour),
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
			return "new-access-token", "new-refresh-token", nil
		},
	}

	service := NewService(mockRepo, mockJWT)
	tokens, err := service.RefreshToken(context.Background(), "old-refresh-token")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if tokens.AccessToken != "new-access-token" {
		t.Errorf("expected new access token")
	}
	if tokens.RefreshToken != "new-refresh-token" {
		t.Errorf("expected new refresh token")
	}
}

func TestRefreshToken_InvalidToken(t *testing.T) {
	mockJWT := &MockJWTManager{
		ValidateRefreshTokenFunc: func(_ string) (*utils.Claims, error) {
			return nil, utils.ErrInvalidToken
		},
	}

	service := NewService(&MockRepository{}, mockJWT)
	_, err := service.RefreshToken(context.Background(), "invalid-token")

	if !errors.Is(err, ErrInvalidRefreshToken) {
		t.Errorf("expected ErrInvalidRefreshToken, got %v", err)
	}
}

func TestRefreshToken_SessionNotFound(t *testing.T) {
	mockRepo := &MockRepository{
		GetSessionByRefreshTokenFunc: func(_ context.Context, _ string) (*Session, error) {
			return nil, ErrSessionNotFound
		},
	}
	mockJWT := &MockJWTManager{
		ValidateRefreshTokenFunc: func(_ string) (*utils.Claims, error) {
			return &utils.Claims{UserID: "user-123", Role: "member"}, nil
		},
	}

	service := NewService(mockRepo, mockJWT)
	_, err := service.RefreshToken(context.Background(), "valid-but-revoked-token")

	if !errors.Is(err, ErrInvalidRefreshToken) {
		t.Errorf("expected ErrInvalidRefreshToken, got %v", err)
	}
}

func TestRefreshToken_UserDisabled(t *testing.T) {
	mockRepo := &MockRepository{
		GetSessionByRefreshTokenFunc: func(_ context.Context, _ string) (*Session, error) {
			return &Session{
				ID:     "session-123",
				UserID: "user-123",
			}, nil
		},
		GetUserByIDFunc: func(_ context.Context, _ string) (*User, error) {
			return &User{
				ID:     "user-123",
				Status: StatusDisabled,
			}, nil
		},
	}
	mockJWT := &MockJWTManager{
		ValidateRefreshTokenFunc: func(_ string) (*utils.Claims, error) {
			return &utils.Claims{UserID: "user-123", Role: "member"}, nil
		},
	}

	service := NewService(mockRepo, mockJWT)
	_, err := service.RefreshToken(context.Background(), "refresh-token")

	if !errors.Is(err, ErrUserDisabled) {
		t.Errorf("expected ErrUserDisabled, got %v", err)
	}
}

// ============================================================================
// Logout Tests
// ============================================================================

func TestLogout_Success(t *testing.T) {
	mockRepo := &MockRepository{
		DeleteSessionByRefreshTokenFunc: func(_ context.Context, _ string) error {
			return nil
		},
	}

	service := NewService(mockRepo, &MockJWTManager{})
	err := service.Logout(context.Background(), "refresh-token")

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestLogout_SessionNotFound(t *testing.T) {
	mockRepo := &MockRepository{
		DeleteSessionByRefreshTokenFunc: func(_ context.Context, _ string) error {
			return ErrSessionNotFound
		},
	}

	service := NewService(mockRepo, &MockJWTManager{})
	err := service.Logout(context.Background(), "nonexistent-token")

	if !errors.Is(err, ErrSessionNotFound) {
		t.Errorf("expected ErrSessionNotFound, got %v", err)
	}
}

// ============================================================================
// LogoutAll Tests
// ============================================================================

func TestLogoutAll_Success(t *testing.T) {
	mockRepo := &MockRepository{
		DeleteAllUserSessionsFunc: func(_ context.Context, _ string) (int64, error) {
			return 3, nil
		},
	}

	service := NewService(mockRepo, &MockJWTManager{})
	count, err := service.LogoutAll(context.Background(), "user-123")

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if count != 3 {
		t.Errorf("expected 3 sessions deleted, got %d", count)
	}
}

func TestLogoutAll_NoSessions(t *testing.T) {
	mockRepo := &MockRepository{
		DeleteAllUserSessionsFunc: func(_ context.Context, _ string) (int64, error) {
			return 0, nil
		},
	}

	service := NewService(mockRepo, &MockJWTManager{})
	count, err := service.LogoutAll(context.Background(), "user-123")

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 sessions deleted, got %d", count)
	}
}

// ============================================================================
// Email Validation Tests
// ============================================================================

func TestIsValidEmail(t *testing.T) {
	validEmails := []string{
		"test@example.com",
		"user.name@domain.org",
		"user+tag@domain.co.uk",
		"test123@test-domain.com",
	}

	for _, email := range validEmails {
		if !isValidEmail(email) {
			t.Errorf("expected %q to be valid", email)
		}
	}

	invalidEmails := []string{
		"invalid",
		"invalid@",
		"@domain.com",
		"invalid@.com",
		"",
		"spaces in@email.com",
	}

	for _, email := range invalidEmails {
		if isValidEmail(email) {
			t.Errorf("expected %q to be invalid", email)
		}
	}
}
