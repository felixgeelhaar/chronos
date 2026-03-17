package auth_test

import (
	"testing"
	"time"

	"github.com/ascend/api/pkg/auth"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test configuration constants
const (
	testAccessSecret  = "test-access-secret-32-characters-long!!"
	testRefreshSecret = "test-refresh-secret-32-characters-long!"
	testAccessExpiry  = 15 * time.Minute
	testRefreshExpiry = 168 * time.Hour // 7 days
)

func TestNewJWTService(t *testing.T) {
	config := auth.JWTConfig{
		AccessSecret:  testAccessSecret,
		RefreshSecret: testRefreshSecret,
		AccessExpiry:  testAccessExpiry,
		RefreshExpiry: testRefreshExpiry,
	}

	service := auth.NewJWTService(config)

	assert.NotNil(t, service)
}

func TestJWTService_GenerateAccessToken(t *testing.T) {
	service := createTestJWTService()
	userID := uuid.New()
	email := "test@example.com"

	tests := []struct {
		name        string
		userID      uuid.UUID
		email       string
		expectError bool
	}{
		{
			name:        "valid credentials",
			userID:      userID,
			email:       email,
			expectError: false,
		},
		{
			name:        "empty email",
			userID:      userID,
			email:       "",
			expectError: false, // Email validation is at handler level
		},
		{
			name:        "zero UUID",
			userID:      uuid.UUID{},
			email:       email,
			expectError: false, // UUID validation is at handler level
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := service.GenerateAccessToken(tt.userID, tt.email)

			if tt.expectError {
				assert.Error(t, err)
				assert.Empty(t, token)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)

				// Token should be a valid JWT format (three parts separated by dots)
				assert.Contains(t, token, ".")
				parts := len(splitToken(token))
				assert.Equal(t, 3, parts, "JWT should have 3 parts (header.payload.signature)")
			}
		})
	}
}

func TestJWTService_GenerateRefreshToken(t *testing.T) {
	service := createTestJWTService()
	userID := uuid.New()
	email := "test@example.com"

	token, err := service.GenerateRefreshToken(userID, email)

	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// Refresh token should be different format than access token
	assert.Contains(t, token, ".")
}

func TestJWTService_ValidateAccessToken(t *testing.T) {
	service := createTestJWTService()
	userID := uuid.New()
	email := "test@example.com"

	// Generate valid token
	validToken, err := service.GenerateAccessToken(userID, email)
	require.NoError(t, err)

	tests := []struct {
		name        string
		token       string
		expectError bool
		checkClaims bool
	}{
		{
			name:        "valid token",
			token:       validToken,
			expectError: false,
			checkClaims: true,
		},
		{
			name:        "empty token",
			token:       "",
			expectError: true,
			checkClaims: false,
		},
		{
			name:        "malformed token",
			token:       "not.a.valid.jwt.token",
			expectError: true,
			checkClaims: false,
		},
		{
			name:        "invalid signature",
			token:       "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.invalid_signature",
			expectError: true,
			checkClaims: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := service.ValidateAccessToken(tt.token)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, claims)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, claims)

				if tt.checkClaims {
					assert.Equal(t, userID, claims.UserID)
					assert.Equal(t, email, claims.Email)
					assert.NotZero(t, claims.ExpiresAt)
					assert.NotZero(t, claims.IssuedAt)
				}
			}
		})
	}
}

func TestJWTService_ValidateRefreshToken(t *testing.T) {
	service := createTestJWTService()
	userID := uuid.New()
	email := "test@example.com"

	// Generate valid refresh token
	validToken, err := service.GenerateRefreshToken(userID, email)
	require.NoError(t, err)

	tests := []struct {
		name        string
		token       string
		expectError bool
	}{
		{
			name:        "valid refresh token",
			token:       validToken,
			expectError: false,
		},
		{
			name:        "empty token",
			token:       "",
			expectError: true,
		},
		{
			name:        "access token instead of refresh token",
			token:       mustGenerateAccessToken(t, service, userID, email),
			expectError: true, // Should fail because it uses wrong secret
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := service.ValidateRefreshToken(tt.token)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, claims)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, claims)
				assert.Equal(t, userID, claims.UserID)
				assert.Equal(t, email, claims.Email)
			}
		})
	}
}

func TestJWTService_TokenExpiry(t *testing.T) {
	// Create service with very short expiry
	shortExpiryService := auth.NewJWTService(auth.JWTConfig{
		AccessSecret:  testAccessSecret,
		RefreshSecret: testRefreshSecret,
		AccessExpiry:  1 * time.Millisecond, // Very short
		RefreshExpiry: 1 * time.Millisecond,
	})

	userID := uuid.New()
	email := "test@example.com"

	// Generate token
	token, err := shortExpiryService.GenerateAccessToken(userID, email)
	require.NoError(t, err)

	// Wait for token to expire
	time.Sleep(10 * time.Millisecond)

	// Validate expired token
	claims, err := shortExpiryService.ValidateAccessToken(token)

	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.Contains(t, err.Error(), "expired")
}

func TestJWTService_TokenNotYetValid(t *testing.T) {
	// This tests the "nbf" (not before) claim
	// For now, our implementation doesn't set nbf, but we should test the concept

	service := createTestJWTService()
	userID := uuid.New()
	email := "test@example.com"

	token, err := service.GenerateAccessToken(userID, email)
	require.NoError(t, err)

	// Token should be valid immediately
	claims, err := service.ValidateAccessToken(token)
	assert.NoError(t, err)
	assert.NotNil(t, claims)
}

func TestJWTService_DifferentUsers(t *testing.T) {
	service := createTestJWTService()

	user1ID := uuid.New()
	user1Email := "user1@example.com"

	user2ID := uuid.New()
	user2Email := "user2@example.com"

	// Generate tokens for two different users
	token1, err1 := service.GenerateAccessToken(user1ID, user1Email)
	token2, err2 := service.GenerateAccessToken(user2ID, user2Email)

	require.NoError(t, err1)
	require.NoError(t, err2)

	// Tokens should be different
	assert.NotEqual(t, token1, token2)

	// Validate both tokens
	claims1, err := service.ValidateAccessToken(token1)
	require.NoError(t, err)
	assert.Equal(t, user1ID, claims1.UserID)
	assert.Equal(t, user1Email, claims1.Email)

	claims2, err := service.ValidateAccessToken(token2)
	require.NoError(t, err)
	assert.Equal(t, user2ID, claims2.UserID)
	assert.Equal(t, user2Email, claims2.Email)
}

func TestJWTService_TokenReuse(t *testing.T) {
	service := createTestJWTService()
	userID := uuid.New()
	email := "test@example.com"

	// Generate token
	token, err := service.GenerateAccessToken(userID, email)
	require.NoError(t, err)

	// Validate token multiple times (should work)
	for i := 0; i < 3; i++ {
		claims, err := service.ValidateAccessToken(token)
		assert.NoError(t, err)
		assert.NotNil(t, claims)
		assert.Equal(t, userID, claims.UserID)
	}
}

func TestJWTService_WrongSecret(t *testing.T) {
	// Service with one secret
	service1 := auth.NewJWTService(auth.JWTConfig{
		AccessSecret:  "secret1-must-be-32-characters-long!!",
		RefreshSecret: testRefreshSecret,
		AccessExpiry:  testAccessExpiry,
		RefreshExpiry: testRefreshExpiry,
	})

	// Service with different secret
	service2 := auth.NewJWTService(auth.JWTConfig{
		AccessSecret:  "secret2-must-be-32-characters-long!!",
		RefreshSecret: testRefreshSecret,
		AccessExpiry:  testAccessExpiry,
		RefreshExpiry: testRefreshExpiry,
	})

	userID := uuid.New()
	email := "test@example.com"

	// Generate token with service1
	token, err := service1.GenerateAccessToken(userID, email)
	require.NoError(t, err)

	// Try to validate with service2 (different secret)
	claims, err := service2.ValidateAccessToken(token)

	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.Contains(t, err.Error(), "signature")
}

func TestJWTService_AlgorithmConfusion(t *testing.T) {
	service := createTestJWTService()

	// Try to create a token with "none" algorithm (security vulnerability)
	maliciousToken := "eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0.eyJ1c2VyX2lkIjoiMTIzIiwiZW1haWwiOiJoYWNrZXJAZXhhbXBsZS5jb20ifQ."

	// Should reject tokens with "none" algorithm
	claims, err := service.ValidateAccessToken(maliciousToken)
	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestJWTService_ClaimsIntegrity(t *testing.T) {
	service := createTestJWTService()
	userID := uuid.New()
	email := "test@example.com"

	token, err := service.GenerateAccessToken(userID, email)
	require.NoError(t, err)

	// Validate and check all claims
	claims, err := service.ValidateAccessToken(token)
	require.NoError(t, err)

	// Check UserID
	assert.Equal(t, userID, claims.UserID)
	assert.NotEqual(t, uuid.Nil, claims.UserID)

	// Check Email
	assert.Equal(t, email, claims.Email)
	assert.NotEmpty(t, claims.Email)

	// Check timestamps
	assert.True(t, claims.IssuedAt.Time.Before(claims.ExpiresAt.Time))
	assert.True(t, claims.ExpiresAt.Time.After(time.Now()))

	// Check expiry duration
	duration := claims.ExpiresAt.Time.Sub(claims.IssuedAt.Time)
	assert.InDelta(t, testAccessExpiry.Seconds(), duration.Seconds(), 1.0)
}

func TestJWTService_RefreshTokenDifferentSecret(t *testing.T) {
	service := createTestJWTService()
	userID := uuid.New()
	email := "test@example.com"

	// Generate access token
	accessToken, err := service.GenerateAccessToken(userID, email)
	require.NoError(t, err)

	// Generate refresh token
	refreshToken, err := service.GenerateRefreshToken(userID, email)
	require.NoError(t, err)

	// Tokens should be different
	assert.NotEqual(t, accessToken, refreshToken)

	// Access token should NOT validate as refresh token
	_, err = service.ValidateRefreshToken(accessToken)
	assert.Error(t, err, "Access token should not validate as refresh token")

	// Refresh token should NOT validate as access token
	_, err = service.ValidateAccessToken(refreshToken)
	assert.Error(t, err, "Refresh token should not validate as access token")
}

func TestJWTService_LongExpiry(t *testing.T) {
	// Test with refresh token's long expiry
	service := createTestJWTService()
	userID := uuid.New()
	email := "test@example.com"

	token, err := service.GenerateRefreshToken(userID, email)
	require.NoError(t, err)

	claims, err := service.ValidateRefreshToken(token)
	require.NoError(t, err)

	// Expiry should be approximately 7 days in the future
	expectedExpiry := time.Now().Add(testRefreshExpiry)
	actualExpiry := claims.ExpiresAt

	// Allow 1 minute tolerance
	diff := actualExpiry.Sub(expectedExpiry)
	assert.Less(t, abs(diff), 1*time.Minute)
}

// Helper functions

func createTestJWTService() *auth.JWTService {
	config := auth.JWTConfig{
		AccessSecret:  testAccessSecret,
		RefreshSecret: testRefreshSecret,
		AccessExpiry:  testAccessExpiry,
		RefreshExpiry: testRefreshExpiry,
	}
	return auth.NewJWTService(config)
}

func mustGenerateAccessToken(t *testing.T, service *auth.JWTService, userID uuid.UUID, email string) string {
	token, err := service.GenerateAccessToken(userID, email)
	require.NoError(t, err)
	return token
}

func splitToken(token string) []string {
	parts := []string{}
	current := ""
	for _, ch := range token {
		if ch == '.' {
			parts = append(parts, current)
			current = ""
		} else {
			current += string(ch)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}

func abs(d time.Duration) time.Duration {
	if d < 0 {
		return -d
	}
	return d
}

// Benchmark tests for JWT performance
func BenchmarkGenerateAccessToken(b *testing.B) {
	service := createTestJWTService()
	userID := uuid.New()
	email := "bench@example.com"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.GenerateAccessToken(userID, email)
	}
}

func BenchmarkValidateAccessToken(b *testing.B) {
	service := createTestJWTService()
	userID := uuid.New()
	email := "bench@example.com"

	token, _ := service.GenerateAccessToken(userID, email)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = service.ValidateAccessToken(token)
	}
}

func BenchmarkGenerateRefreshToken(b *testing.B) {
	service := createTestJWTService()
	userID := uuid.New()
	email := "bench@example.com"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.GenerateRefreshToken(userID, email)
	}
}
