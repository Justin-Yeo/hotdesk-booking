package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	// AccessTokenExpiry is the duration for which access tokens are valid
	AccessTokenExpiry = 15 * time.Minute
	// RefreshTokenExpiry is the duration for which refresh tokens are valid
	RefreshTokenExpiry = 7 * 24 * time.Hour // 7 days
)

var (
	// ErrInvalidToken is returned when a token is invalid
	ErrInvalidToken = errors.New("invalid token")
	// ErrExpiredToken is returned when a token has expired
	ErrExpiredToken = errors.New("token has expired")
	// ErrEmptySecret is returned when the JWT secret is empty
	ErrEmptySecret = errors.New("JWT secret cannot be empty")
	// ErrInvalidClaims is returned when token claims are invalid
	ErrInvalidClaims = errors.New("invalid token claims")
)

// TokenType represents the type of JWT token
type TokenType string

const (
	// AccessToken represents an access token
	AccessToken TokenType = "access"
	// RefreshToken represents a refresh token
	RefreshToken TokenType = "refresh"
)

// Claims represents the JWT claims structure
type Claims struct {
	UserID    string    `json:"user_id"`
	Role      string    `json:"role"`
	TokenType TokenType `json:"token_type"`
	jwt.RegisteredClaims
}

// JWTManager handles JWT token operations
type JWTManager struct {
	secret []byte
}

// NewJWTManager creates a new JWTManager with the given secret
func NewJWTManager(secret string) (*JWTManager, error) {
	if secret == "" {
		return nil, ErrEmptySecret
	}
	return &JWTManager{
		secret: []byte(secret),
	}, nil
}

// GenerateAccessToken creates a new access token for the given user
func (m *JWTManager) GenerateAccessToken(userID, role string) (string, error) {
	return m.generateToken(userID, role, AccessToken, AccessTokenExpiry)
}

// GenerateRefreshToken creates a new refresh token for the given user
func (m *JWTManager) GenerateRefreshToken(userID, role string) (string, error) {
	return m.generateToken(userID, role, RefreshToken, RefreshTokenExpiry)
}

// generateToken creates a JWT token with the specified parameters
func (m *JWTManager) generateToken(userID, role string, tokenType TokenType, expiry time.Duration) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:    userID,
		Role:      role,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

// ValidateToken validates a JWT token and returns its claims
func (m *JWTManager) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return m.secret, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidClaims
	}

	return claims, nil
}

// ValidateAccessToken validates an access token
func (m *JWTManager) ValidateAccessToken(tokenString string) (*Claims, error) {
	claims, err := m.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	if claims.TokenType != AccessToken {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// ValidateRefreshToken validates a refresh token
func (m *JWTManager) ValidateRefreshToken(tokenString string) (*Claims, error) {
	claims, err := m.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	if claims.TokenType != RefreshToken {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// GenerateTokenPair generates both access and refresh tokens
func (m *JWTManager) GenerateTokenPair(userID, role string) (accessToken, refreshToken string, err error) {
	accessToken, err = m.GenerateAccessToken(userID, role)
	if err != nil {
		return "", "", err
	}

	refreshToken, err = m.GenerateRefreshToken(userID, role)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}
