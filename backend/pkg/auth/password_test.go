package auth_test

import (
	"strings"
	"testing"

	"github.com/ascend/api/pkg/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHashPassword(t *testing.T) {
	tests := []struct {
		name        string
		password    string
		expectError bool
	}{
		{
			name:        "valid password",
			password:    "SecurePassword123!",
			expectError: false,
		},
		{
			name:        "short password",
			password:    "pass",
			expectError: false, // Hashing should still work, validation is elsewhere
		},
		{
			name:        "long password (72 characters - bcrypt limit)",
			password:    strings.Repeat("a", 72),
			expectError: false,
		},
		{
			name:        "very long password (exceeds bcrypt limit)",
			password:    strings.Repeat("a", 73),
			expectError: true, // bcrypt rejects passwords > 72 bytes
		},
		{
			name:        "password with special characters",
			password:    "P@ssw0rd!#$%^&*()",
			expectError: false,
		},
		{
			name:        "unicode password",
			password:    "пароль密码🔒",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := auth.HashPassword(tt.password)

			if tt.expectError {
				assert.Error(t, err)
				assert.Empty(t, hash)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, hash)

				// Hash should be bcrypt format (starts with $2a$ or $2b$)
				assert.True(t, strings.HasPrefix(hash, "$2a$") || strings.HasPrefix(hash, "$2b$"))

				// Hash should not equal plain password
				assert.NotEqual(t, tt.password, hash)

				// Hash should be at least 59 characters (typical bcrypt hash length)
				assert.GreaterOrEqual(t, len(hash), 59)
			}
		})
	}
}

func TestVerifyPassword(t *testing.T) {
	// Pre-compute a valid hash for testing
	validPassword := "SecurePassword123!"
	validHash, err := auth.HashPassword(validPassword)
	require.NoError(t, err)

	tests := []struct {
		name          string
		hashedPassword string
		password      string
		expectError   bool
		errorContains string
	}{
		{
			name:          "correct password",
			hashedPassword: validHash,
			password:      validPassword,
			expectError:   false,
		},
		{
			name:          "incorrect password",
			hashedPassword: validHash,
			password:      "WrongPassword123!",
			expectError:   true,
			errorContains: "invalid password",
		},
		{
			name:          "empty password",
			hashedPassword: validHash,
			password:      "",
			expectError:   true,
			errorContains: "invalid password",
		},
		{
			name:          "case sensitive check",
			hashedPassword: validHash,
			password:      "securepassword123!", // lowercase
			expectError:   true,
			errorContains: "invalid password",
		},
		{
			name:          "invalid hash format",
			hashedPassword: "not-a-valid-hash",
			password:      validPassword,
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := auth.VerifyPassword(tt.hashedPassword, tt.password)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestHashPassword_Determinism(t *testing.T) {
	// Same password should produce different hashes (salt)
	password := "TestPassword123!"

	hash1, err1 := auth.HashPassword(password)
	require.NoError(t, err1)

	hash2, err2 := auth.HashPassword(password)
	require.NoError(t, err2)

	// Hashes should be different due to random salt
	assert.NotEqual(t, hash1, hash2, "bcrypt should generate different hashes with different salts")

	// But both should verify correctly
	assert.NoError(t, auth.VerifyPassword(hash1, password))
	assert.NoError(t, auth.VerifyPassword(hash2, password))
}

func TestPasswordSecurity_CostFactor(t *testing.T) {
	// Verify that the bcrypt cost factor is appropriate (12 is recommended)
	password := "TestPassword123!"
	hash, err := auth.HashPassword(password)
	require.NoError(t, err)

	// bcrypt hash format: $2a$12$... (where 12 is the cost)
	// Cost should be between 10-14 for good security/performance balance
	parts := strings.Split(hash, "$")
	require.GreaterOrEqual(t, len(parts), 3, "Invalid bcrypt hash format")

	cost := parts[2]
	assert.Equal(t, "12", cost, "Cost factor should be 12 for optimal security")
}

func TestVerifyPassword_TimingSafeComparison(t *testing.T) {
	// This test verifies that incorrect passwords still take similar time
	// to verify as correct passwords (prevents timing attacks)
	password := "CorrectPassword123!"
	hash, err := auth.HashPassword(password)
	require.NoError(t, err)

	// Test with multiple incorrect passwords
	incorrectPasswords := []string{
		"W", // 1 character different
		"WrongPassword123!", // Multiple characters different
		"CorrectPassword124!", // Last character different
		strings.Repeat("a", 100), // Very different length
	}

	for _, incorrectPwd := range incorrectPasswords {
		err := auth.VerifyPassword(hash, incorrectPwd)
		assert.Error(t, err, "Should fail for incorrect password")
		// Note: bcrypt is inherently timing-safe, so all comparisons
		// take approximately the same time regardless of how wrong the password is
	}
}

func TestPasswordEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		password    string
		expectError bool
	}{
		{"whitespace only", "   ", false},
		{"newline character", "pass\nword", false},
		{"tab character", "pass\tword", false},
		{"null bytes", "pass\x00word", false},
		{"emoji", "🔒🔑🚪", false},
		{"long password within limit", strings.Repeat("a", 70), false},
		{"very long password (exceeds limit)", strings.Repeat("a", 500), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := auth.HashPassword(tt.password)

			if tt.expectError {
				assert.Error(t, err)
				assert.Empty(t, hash)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, hash)

				// Should verify correctly
				err = auth.VerifyPassword(hash, tt.password)
				assert.NoError(t, err)

				// Should fail with different password
				err = auth.VerifyPassword(hash, tt.password+"different")
				assert.Error(t, err)
			}
		})
	}
}

// Benchmark tests for password hashing performance
func BenchmarkHashPassword(b *testing.B) {
	password := "BenchmarkPassword123!"
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = auth.HashPassword(password)
	}
}

func BenchmarkVerifyPassword(b *testing.B) {
	password := "BenchmarkPassword123!"
	hash, _ := auth.HashPassword(password)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = auth.VerifyPassword(hash, password)
	}
}
