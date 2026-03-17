package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/ascend/api/internal/domain"
	"github.com/ascend/api/internal/dto"
	"github.com/ascend/api/internal/repository"
	"github.com/ascend/api/internal/service"
	"github.com/ascend/api/pkg/auth"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	sqlDB, err := db.DB()
	require.NoError(t, err)

	// Enable foreign key constraints
	_, err = sqlDB.Exec("PRAGMA foreign_keys = ON")
	require.NoError(t, err)

	// Create users table
	_, err = sqlDB.Exec(`
		CREATE TABLE users (
			id TEXT PRIMARY KEY NOT NULL,
			email TEXT NOT NULL UNIQUE,
			password TEXT NOT NULL,
			name TEXT NOT NULL,
			body_weight REAL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			deleted_at DATETIME
		)
	`)
	require.NoError(t, err)

	return db
}

// teardownTestDB closes the database connection
func teardownTestDB(t *testing.T, db *gorm.DB) {
	t.Helper()
	sqlDB, err := db.DB()
	require.NoError(t, err)
	require.NoError(t, sqlDB.Close())
}

func TestAuthService_Register(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	userRepo := repository.NewUserRepository(db)
	jwtService := auth.NewJWTService(auth.JWTConfig{
		AccessSecret:  "test-access-secret",
		RefreshSecret: "test-refresh-secret",
		AccessExpiry:  15 * time.Minute,
		RefreshExpiry: 24 * time.Hour,
	})
	authService := service.NewAuthService(userRepo, jwtService, 15*time.Minute)
	ctx := context.Background()

	tests := []struct {
		name        string
		request     *dto.RegisterRequest
		setupFunc   func()
		expectError bool
		errorMsg    string
	}{
		{
			name: "successful registration",
			request: &dto.RegisterRequest{
				Email:      "newuser@example.com",
				Password:   "password123",
				Name:       "New User",
				BodyWeight: floatPtr(75.5),
			},
			setupFunc:   func() {},
			expectError: false,
		},
		{
			name: "registration without body weight",
			request: &dto.RegisterRequest{
				Email:    "another@example.com",
				Password: "password123",
				Name:     "Another User",
			},
			setupFunc:   func() {},
			expectError: false,
		},
		{
			name: "duplicate email",
			request: &dto.RegisterRequest{
				Email:    "duplicate@example.com",
				Password: "password123",
				Name:     "Duplicate User",
			},
			setupFunc: func() {
				// Pre-create user with same email
				user := &domain.User{
					Email:    "duplicate@example.com",
					Password: "existingpassword",
					Name:     "Existing User",
				}
				_ = userRepo.Create(ctx, user)
			},
			expectError: true,
			errorMsg:    "already exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupFunc()

			response, err := authService.Register(ctx, tt.request)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, response)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
				require.NotNil(t, response)
				assert.NotEmpty(t, response.AccessToken)
				assert.NotEmpty(t, response.RefreshToken)
				assert.Greater(t, response.ExpiresIn, 0)
				assert.Equal(t, tt.request.Email, response.User.Email)
				assert.Equal(t, tt.request.Name, response.User.Name)
				assert.NotEqual(t, uuid.Nil, response.User.ID)

				// Verify password is hashed, not stored as plain text
				user, err := userRepo.GetByEmail(ctx, tt.request.Email)
				require.NoError(t, err)
				assert.NotEqual(t, tt.request.Password, user.Password, "Password should be hashed")
			}
		})
	}
}

func TestAuthService_Login(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	userRepo := repository.NewUserRepository(db)
	jwtService := auth.NewJWTService(auth.JWTConfig{
		AccessSecret:  "test-access-secret",
		RefreshSecret: "test-refresh-secret",
		AccessExpiry:  15 * time.Minute,
		RefreshExpiry: 24 * time.Hour,
	})
	authService := service.NewAuthService(userRepo, jwtService, 15*time.Minute)
	ctx := context.Background()

	// Create test user
	hashedPassword, err := auth.HashPassword("correctpassword")
	require.NoError(t, err)

	user := &domain.User{
		Email:      "login@example.com",
		Password:   hashedPassword,
		Name:       "Login User",
		BodyWeight: floatPtr(80.0),
	}
	err = userRepo.Create(ctx, user)
	require.NoError(t, err)

	tests := []struct {
		name        string
		request     *dto.LoginRequest
		expectError bool
		errorMsg    string
	}{
		{
			name: "successful login",
			request: &dto.LoginRequest{
				Email:    "login@example.com",
				Password: "correctpassword",
			},
			expectError: false,
		},
		{
			name: "incorrect password",
			request: &dto.LoginRequest{
				Email:    "login@example.com",
				Password: "wrongpassword",
			},
			expectError: true,
			errorMsg:    "invalid email or password",
		},
		{
			name: "non-existent email",
			request: &dto.LoginRequest{
				Email:    "nonexistent@example.com",
				Password: "anypassword",
			},
			expectError: true,
			errorMsg:    "invalid email or password",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := authService.Login(ctx, tt.request)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, response)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
				require.NotNil(t, response)
				assert.NotEmpty(t, response.AccessToken)
				assert.NotEmpty(t, response.RefreshToken)
				assert.Greater(t, response.ExpiresIn, 0)
				assert.Equal(t, user.Email, response.User.Email)
				assert.Equal(t, user.Name, response.User.Name)
				assert.Equal(t, user.ID, response.User.ID)
			}
		})
	}
}

func TestAuthService_RefreshToken(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	userRepo := repository.NewUserRepository(db)
	jwtService := auth.NewJWTService(auth.JWTConfig{
		AccessSecret:  "test-access-secret",
		RefreshSecret: "test-refresh-secret",
		AccessExpiry:  15 * time.Minute,
		RefreshExpiry: 24 * time.Hour,
	})
	authService := service.NewAuthService(userRepo, jwtService, 15*time.Minute)
	ctx := context.Background()

	// Create test user
	user := &domain.User{
		Email:    "refresh@example.com",
		Password: "hashedpassword",
		Name:     "Refresh User",
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	// Generate valid refresh token
	validRefreshToken, err := jwtService.GenerateRefreshToken(user.ID, user.Email)
	require.NoError(t, err)

	tests := []struct {
		name         string
		request      *dto.RefreshTokenRequest
		expectError  bool
		errorMsg     string
	}{
		{
			name: "valid refresh token",
			request: &dto.RefreshTokenRequest{
				RefreshToken: validRefreshToken,
			},
			expectError: false,
		},
		{
			name: "invalid refresh token",
			request: &dto.RefreshTokenRequest{
				RefreshToken: "invalid.token.here",
			},
			expectError: true,
			errorMsg:    "invalid refresh token",
		},
		{
			name: "empty refresh token",
			request: &dto.RefreshTokenRequest{
				RefreshToken: "",
			},
			expectError: true,
			errorMsg:    "invalid refresh token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := authService.RefreshToken(ctx, tt.request)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, response)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
				require.NotNil(t, response)
				assert.NotEmpty(t, response.AccessToken)
				assert.NotEmpty(t, response.RefreshToken)
				assert.NotEqual(t, validRefreshToken, response.AccessToken, "New access token should be generated")
			}
		})
	}
}

func TestAuthService_GetUserByID(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	userRepo := repository.NewUserRepository(db)
	jwtService := auth.NewJWTService(auth.JWTConfig{
		AccessSecret:  "test-access-secret",
		RefreshSecret: "test-refresh-secret",
		AccessExpiry:  15 * time.Minute,
		RefreshExpiry: 24 * time.Hour,
	})
	authService := service.NewAuthService(userRepo, jwtService, 15*time.Minute)
	ctx := context.Background()

	// Create test user
	user := &domain.User{
		Email:      "getuser@example.com",
		Password:   "hashedpassword",
		Name:       "Get User",
		BodyWeight: floatPtr(70.5),
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	tests := []struct {
		name        string
		userID      uuid.UUID
		expectError bool
		errorMsg    string
	}{
		{
			name:        "existing user",
			userID:      user.ID,
			expectError: false,
		},
		{
			name:        "non-existent user",
			userID:      uuid.New(),
			expectError: true,
			errorMsg:    "user not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := authService.GetUserByID(ctx, tt.userID)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, response)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
				require.NotNil(t, response)
				assert.Equal(t, user.ID, response.ID)
				assert.Equal(t, user.Email, response.Email)
				assert.Equal(t, user.Name, response.Name)
				assert.Equal(t, user.BodyWeight, response.BodyWeight)
			}
		})
	}
}

// Helper function for float pointers
func floatPtr(f float64) *float64 {
	return &f
}
