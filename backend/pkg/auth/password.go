package auth

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const (
	// bcrypt cost factor (default: 10, range: 4-31)
	bcryptCost = 12
)

// HashPassword hashes a plain text password using bcrypt
func HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hashedBytes), nil
}

// VerifyPassword compares a plain text password with a hashed password
func VerifyPassword(hashedPassword, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return fmt.Errorf("invalid password")
		}
		return fmt.Errorf("failed to verify password: %w", err)
	}
	return nil
}
