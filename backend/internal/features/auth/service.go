package auth

import (
	"context"
	"errors"
	"regexp"
	"time"

	"github.com/justinyeo/hotdesk-booking/backend/internal/shared/utils"
)

const (
	// MinPasswordLength is the minimum required password length
	MinPasswordLength = 8
)

var (
	// ErrInvalidEmail is returned when email format is invalid
	ErrInvalidEmail = errors.New("invalid email format")
	// ErrPasswordTooShort is returned when password is too short
	ErrPasswordTooShort = errors.New("password must be at least 8 characters")
	// ErrInvalidCredentials is returned when login credentials are wrong
	ErrInvalidCredentials = errors.New("invalid email or password")
	// ErrUserDisabled is returned when user account is disabled
	ErrUserDisabled = errors.New("user account is disabled")
	// ErrInvalidRefreshToken is returned when refresh token is invalid
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
)

// emailRegex validates email format
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// RepositoryInterface defines the methods required from the repository
type RepositoryInterface interface {
	CreateUser(ctx context.Context, input *CreateUserInput) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	GetUserByID(ctx context.Context, id string) (*User, error)
	CreateSession(ctx context.Context, input *CreateSessionInput) (*Session, error)
	GetSessionByRefreshToken(ctx context.Context, refreshToken string) (*Session, error)
	DeleteSession(ctx context.Context, sessionID string) error
	DeleteSessionByRefreshToken(ctx context.Context, refreshToken string) error
	DeleteAllUserSessions(ctx context.Context, userID string) (int64, error)
}

// JWTManagerInterface defines the methods required from the JWT manager
type JWTManagerInterface interface {
	GenerateAccessToken(userID, role string) (string, error)
	GenerateRefreshToken(userID, role string) (string, error)
	GenerateTokenPair(userID, role string) (accessToken, refreshToken string, err error)
	ValidateRefreshToken(tokenString string) (*utils.Claims, error)
}

// Service provides authentication business logic
type Service struct {
	repo RepositoryInterface
	jwt  JWTManagerInterface
}

// NewService creates a new auth service
func NewService(repo RepositoryInterface, jwt JWTManagerInterface) *Service {
	return &Service{
		repo: repo,
		jwt:  jwt,
	}
}

// Tokens represents the tokens returned after successful authentication
type Tokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// RegisterInput represents the input for user registration
type RegisterInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginInput represents the input for user login
type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RegisterResult represents the result of a successful registration
type RegisterResult struct {
	User   *User  `json:"user"`
	Tokens Tokens `json:"tokens"`
}

// LoginResult represents the result of a successful login
type LoginResult struct {
	User   *User  `json:"user"`
	Tokens Tokens `json:"tokens"`
}

// Register creates a new user account
func (s *Service) Register(ctx context.Context, input *RegisterInput) (*RegisterResult, error) {
	// Validate email format
	if !isValidEmail(input.Email) {
		return nil, ErrInvalidEmail
	}

	// Validate password length
	if len(input.Password) < MinPasswordLength {
		return nil, ErrPasswordTooShort
	}

	// Hash the password
	passwordHash, err := utils.HashPassword(input.Password)
	if err != nil {
		return nil, err
	}

	// Create the user
	user, err := s.repo.CreateUser(ctx, &CreateUserInput{
		Email:        input.Email,
		PasswordHash: passwordHash,
		Role:         RoleMember,
	})
	if err != nil {
		return nil, err
	}

	// Generate tokens
	accessToken, refreshToken, err := s.jwt.GenerateTokenPair(user.ID, string(user.Role))
	if err != nil {
		return nil, err
	}

	// Create session with refresh token
	_, err = s.repo.CreateSession(ctx, &CreateSessionInput{
		UserID:       user.ID,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(utils.RefreshTokenExpiry),
	})
	if err != nil {
		return nil, err
	}

	return &RegisterResult{
		User: user,
		Tokens: Tokens{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		},
	}, nil
}

// Login authenticates a user and returns tokens
func (s *Service) Login(ctx context.Context, input *LoginInput) (*LoginResult, error) {
	// Get user by email
	user, err := s.repo.GetUserByEmail(ctx, input.Email)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	// Check if user is disabled
	if user.Status == StatusDisabled {
		return nil, ErrUserDisabled
	}

	// Verify password
	if err := utils.VerifyPassword(input.Password, user.PasswordHash); err != nil {
		return nil, ErrInvalidCredentials
	}

	// Generate tokens
	accessToken, refreshToken, err := s.jwt.GenerateTokenPair(user.ID, string(user.Role))
	if err != nil {
		return nil, err
	}

	// Create session with refresh token
	_, err = s.repo.CreateSession(ctx, &CreateSessionInput{
		UserID:       user.ID,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(utils.RefreshTokenExpiry),
	})
	if err != nil {
		return nil, err
	}

	return &LoginResult{
		User: user,
		Tokens: Tokens{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		},
	}, nil
}

// RefreshToken validates a refresh token and issues a new token pair
func (s *Service) RefreshToken(ctx context.Context, refreshToken string) (*Tokens, error) {
	// Validate the refresh token JWT
	claims, err := s.jwt.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, ErrInvalidRefreshToken
	}

	// Verify session exists in database
	session, err := s.repo.GetSessionByRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, ErrInvalidRefreshToken
	}

	// Get user to verify they still exist and are active
	user, err := s.repo.GetUserByID(ctx, session.UserID)
	if err != nil {
		return nil, ErrInvalidRefreshToken
	}

	// Check if user is disabled
	if user.Status == StatusDisabled {
		return nil, ErrUserDisabled
	}

	// Delete old session (token rotation)
	if err := s.repo.DeleteSession(ctx, session.ID); err != nil {
		return nil, err
	}

	// Generate new token pair
	newAccessToken, newRefreshToken, err := s.jwt.GenerateTokenPair(claims.UserID, claims.Role)
	if err != nil {
		return nil, err
	}

	// Create new session with new refresh token
	_, err = s.repo.CreateSession(ctx, &CreateSessionInput{
		UserID:       claims.UserID,
		RefreshToken: newRefreshToken,
		ExpiresAt:    time.Now().Add(utils.RefreshTokenExpiry),
	})
	if err != nil {
		return nil, err
	}

	return &Tokens{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

// Logout invalidates a single session by refresh token
func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	return s.repo.DeleteSessionByRefreshToken(ctx, refreshToken)
}

// LogoutAll invalidates all sessions for a user
func (s *Service) LogoutAll(ctx context.Context, userID string) (int64, error) {
	return s.repo.DeleteAllUserSessions(ctx, userID)
}

// isValidEmail validates email format using regex
func isValidEmail(email string) bool {
	return emailRegex.MatchString(email)
}
