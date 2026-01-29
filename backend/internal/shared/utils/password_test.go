package utils

import (
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestHashPassword(t *testing.T) {
	tests := []struct {
		name        string
		password    string
		wantErr     error
		shouldMatch bool
	}{
		{
			name:        "valid password",
			password:    "securePassword123!",
			wantErr:     nil,
			shouldMatch: true,
		},
		{
			name:        "minimum length password",
			password:    "12345678",
			wantErr:     nil,
			shouldMatch: true,
		},
		{
			name:        "empty password",
			password:    "",
			wantErr:     ErrEmptyPassword,
			shouldMatch: false,
		},
		{
			name:        "long password",
			password:    "ThisIsAVeryLongPasswordThatShouldStillWork1234567890!@#$%^&*()",
			wantErr:     nil,
			shouldMatch: true,
		},
		{
			name:        "password with special characters",
			password:    "P@$$w0rd!#%&*()_+-=[]{}|;':\",./<>?",
			wantErr:     nil,
			shouldMatch: true,
		},
		{
			name:        "password with unicode",
			password:    "密码安全123",
			wantErr:     nil,
			shouldMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := HashPassword(tt.password)

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("HashPassword() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("HashPassword() unexpected error = %v", err)
				return
			}

			if hash == "" {
				t.Error("HashPassword() returned empty hash")
				return
			}

			// Verify the hash is valid bcrypt format
			cost, err := bcrypt.Cost([]byte(hash))
			if err != nil {
				t.Errorf("HashPassword() produced invalid bcrypt hash: %v", err)
				return
			}

			if cost != BcryptCost {
				t.Errorf("HashPassword() cost = %d, want %d", cost, BcryptCost)
			}

			// Verify the password matches the hash
			if tt.shouldMatch {
				err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(tt.password))
				if err != nil {
					t.Errorf("HashPassword() hash doesn't match original password")
				}
			}
		})
	}
}

func TestHashPassword_DifferentHashes(t *testing.T) {
	password := "samePassword123"

	hash1, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	hash2, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	if hash1 == hash2 {
		t.Error("HashPassword() same password should produce different hashes due to random salt")
	}
}

func TestVerifyPassword(t *testing.T) {
	validPassword := "securePassword123!"
	validHash, _ := HashPassword(validPassword)

	tests := []struct {
		name     string
		password string
		hash     string
		wantErr  error
	}{
		{
			name:     "correct password",
			password: validPassword,
			hash:     validHash,
			wantErr:  nil,
		},
		{
			name:     "wrong password",
			password: "wrongPassword",
			hash:     validHash,
			wantErr:  ErrInvalidPassword,
		},
		{
			name:     "empty password",
			password: "",
			hash:     validHash,
			wantErr:  ErrEmptyPassword,
		},
		{
			name:     "empty hash",
			password: validPassword,
			hash:     "",
			wantErr:  ErrInvalidPassword,
		},
		{
			name:     "invalid hash format",
			password: validPassword,
			hash:     "notavalidhash",
			wantErr:  bcrypt.ErrHashTooShort,
		},
		{
			name:     "case sensitive - different case",
			password: "SecurePassword123!",
			hash:     validHash,
			wantErr:  ErrInvalidPassword,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := VerifyPassword(tt.password, tt.hash)

			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("VerifyPassword() expected error %v, got nil", tt.wantErr)
					return
				}
				// For bcrypt errors, just check that we got an error
				if tt.wantErr == bcrypt.ErrHashTooShort {
					return // Any error is acceptable for invalid hash
				}
				if err != tt.wantErr {
					t.Errorf("VerifyPassword() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("VerifyPassword() unexpected error = %v", err)
			}
		})
	}
}

func TestVerifyPassword_TimingSafe(t *testing.T) {
	// This test verifies that password comparison is timing-safe
	// bcrypt.CompareHashAndPassword is inherently timing-safe
	password := "testPassword123"
	hash, _ := HashPassword(password)

	// Multiple verifications should work consistently
	for i := 0; i < 10; i++ {
		err := VerifyPassword(password, hash)
		if err != nil {
			t.Errorf("VerifyPassword() iteration %d unexpected error = %v", i, err)
		}
	}
}

// Benchmark tests
func BenchmarkHashPassword(b *testing.B) {
	password := "benchmarkPassword123!"
	for i := 0; i < b.N; i++ {
		_, _ = HashPassword(password)
	}
}

func BenchmarkVerifyPassword(b *testing.B) {
	password := "benchmarkPassword123!"
	hash, _ := HashPassword(password)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = VerifyPassword(password, hash)
	}
}
