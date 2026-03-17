package service

import (
	"context"
	"fmt"
	"time"

	"github.com/ascend/api/internal/domain"
	"github.com/ascend/api/internal/dto"
	apperrors "github.com/ascend/api/internal/errors"
	"github.com/ascend/api/internal/repository"
	"github.com/ascend/api/pkg/auth"
	"github.com/google/uuid"
)

// AuthServiceInterface defines the contract for authentication operations
type AuthServiceInterface interface {
	Register(ctx context.Context, req *dto.RegisterRequest) (*dto.AuthResponse, error)
	Login(ctx context.Context, req *dto.LoginRequest) (*dto.AuthResponse, error)
	RefreshToken(ctx context.Context, req *dto.RefreshTokenRequest) (*dto.AuthResponse, error)
	GetUserByID(ctx context.Context, userID uuid.UUID) (*dto.UserResponse, error)
}

// Verify that AuthService implements AuthServiceInterface
var _ AuthServiceInterface = (*AuthService)(nil)

// AuthService handles authentication business logic
type AuthService struct {
	userRepo   repository.UserRepository
	jwtService *auth.JWTService
	accessExpiry time.Duration
}

// NewAuthService creates a new auth service
func NewAuthService(userRepo repository.UserRepository, jwtService *auth.JWTService, accessExpiry time.Duration) *AuthService {
	return &AuthService{
		userRepo:   userRepo,
		jwtService: jwtService,
		accessExpiry: accessExpiry,
	}
}

// Register creates a new user account
func (s *AuthService) Register(ctx context.Context, req *dto.RegisterRequest) (*dto.AuthResponse, error) {
	// Check if user already exists
	existingUser, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err == nil && existingUser != nil {
		return nil, apperrors.ConflictError(fmt.Sprintf("user with email %s already exists", req.Email))
	}

	// Hash password
	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &domain.User{
		Email:      req.Email,
		Password:   hashedPassword,
		Name:       req.Name,
		BodyWeight: req.BodyWeight,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Generate tokens
	return s.generateAuthResponse(user)
}

// Login authenticates a user and returns tokens
func (s *AuthService) Login(ctx context.Context, req *dto.LoginRequest) (*dto.AuthResponse, error) {
	// Get user by email
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("invalid email or password")
	}

	// Verify password
	if err := auth.VerifyPassword(user.Password, req.Password); err != nil {
		return nil, fmt.Errorf("invalid email or password")
	}

	// Generate tokens
	return s.generateAuthResponse(user)
}

// RefreshToken generates a new access token from a refresh token
func (s *AuthService) RefreshToken(ctx context.Context, req *dto.RefreshTokenRequest) (*dto.AuthResponse, error) {
	// Validate refresh token
	claims, err := s.jwtService.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	// Get user
	user, err := s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	// Generate new tokens
	return s.generateAuthResponse(user)
}

// GetUserByID retrieves a user by ID
func (s *AuthService) GetUserByID(ctx context.Context, userID uuid.UUID) (*dto.UserResponse, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	return &dto.UserResponse{
		ID:         user.ID,
		Email:      user.Email,
		Name:       user.Name,
		BodyWeight: user.BodyWeight,
	}, nil
}

// generateAuthResponse generates access and refresh tokens for a user
func (s *AuthService) generateAuthResponse(user *domain.User) (*dto.AuthResponse, error) {
	accessToken, err := s.jwtService.GenerateAccessToken(user.ID, user.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.jwtService.GenerateRefreshToken(user.ID, user.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &dto.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int(s.accessExpiry.Seconds()),
		User: dto.UserResponse{
			ID:         user.ID,
			Email:      user.Email,
			Name:       user.Name,
			BodyWeight: user.BodyWeight,
		},
	}, nil
}
