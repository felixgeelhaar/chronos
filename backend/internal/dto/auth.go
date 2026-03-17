package dto

import "github.com/google/uuid"

// RegisterRequest represents the registration request payload
type RegisterRequest struct {
	Email    string   `json:"email" binding:"required,email"`
	Password string   `json:"password" binding:"required,min=8"`
	Name     string   `json:"name" binding:"required,min=2"`
	BodyWeight *float64 `json:"body_weight,omitempty"`
}

// LoginRequest represents the login request payload
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// RefreshTokenRequest represents the refresh token request payload
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// AuthResponse represents the authentication response
type AuthResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresIn    int       `json:"expires_in"` // seconds
	User         UserResponse `json:"user"`
}

// UserResponse represents the user data in auth response
type UserResponse struct {
	ID         uuid.UUID `json:"id"`
	Email      string    `json:"email"`
	Name       string    `json:"name"`
	BodyWeight *float64  `json:"body_weight,omitempty"`
}
