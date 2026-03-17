package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ascend/api/internal/dto"
	"github.com/ascend/api/internal/handler"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAuthService is a mock implementation of AuthService
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Register(ctx context.Context, req *dto.RegisterRequest) (*dto.AuthResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.AuthResponse), args.Error(1)
}

func (m *MockAuthService) Login(ctx context.Context, req *dto.LoginRequest) (*dto.AuthResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.AuthResponse), args.Error(1)
}

func (m *MockAuthService) RefreshToken(ctx context.Context, req *dto.RefreshTokenRequest) (*dto.AuthResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.AuthResponse), args.Error(1)
}

func (m *MockAuthService) GetUserByID(ctx context.Context, userID uuid.UUID) (*dto.UserResponse, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.UserResponse), args.Error(1)
}

func TestAuthHandler_Register(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		requestBody    interface{}
		mockResponse   *dto.AuthResponse
		mockError      error
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name: "successful registration",
			requestBody: dto.RegisterRequest{
				Email:    "test@example.com",
				Password: "SecurePass123!",
				Name:     "Test User",
			},
			mockResponse: &dto.AuthResponse{
				AccessToken:  "access_token_here",
				RefreshToken: "refresh_token_here",
				ExpiresIn:    900,
			},
			mockError:      nil,
			expectedStatus: http.StatusCreated,
		},
		{
			name: "invalid email format",
			requestBody: dto.RegisterRequest{
				Email:    "invalid-email",
				Password: "SecurePass123!",
				Name:     "Test User",
			},
			mockResponse:   nil,
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "password too short",
			requestBody: dto.RegisterRequest{
				Email:    "test@example.com",
				Password: "short",
				Name:     "Test User",
			},
			mockResponse:   nil,
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "missing required fields",
			requestBody: dto.RegisterRequest{
				Email: "test@example.com",
				// Missing password and name
			},
			mockResponse:   nil,
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock service
			mockService := new(MockAuthService)
			if tt.mockResponse != nil {
				mockService.On("Register", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(tt.mockResponse, tt.mockError)
			}

			// Create handler with mock service
			authHandler := handler.NewAuthHandler(mockService)

			// Create test context
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Prepare request body
			bodyBytes, _ := json.Marshal(tt.requestBody)
			c.Request = httptest.NewRequest(http.MethodPost, "/v1/auth/register", bytes.NewBuffer(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")

			// Call handler
			authHandler.Register(c)

			// Assert response
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Verify mock expectations
			if tt.mockResponse != nil {
				mockService.AssertExpectations(t)
			}
		})
	}
}

func TestAuthHandler_Login(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		requestBody    dto.LoginRequest
		mockResponse   *dto.AuthResponse
		mockError      error
		expectedStatus int
	}{
		{
			name: "successful login",
			requestBody: dto.LoginRequest{
				Email:    "test@example.com",
				Password: "SecurePass123!",
			},
			mockResponse: &dto.AuthResponse{
				AccessToken:  "access_token_here",
				RefreshToken: "refresh_token_here",
				ExpiresIn:    900,
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name: "invalid email format",
			requestBody: dto.LoginRequest{
				Email:    "invalid-email",
				Password: "SecurePass123!",
			},
			mockResponse:   nil,
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "missing password",
			requestBody: dto.LoginRequest{
				Email: "test@example.com",
			},
			mockResponse:   nil,
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock service
			mockService := new(MockAuthService)
			if tt.mockResponse != nil {
				mockService.On("Login", mock.Anything, mock.Anything).
					Return(tt.mockResponse, tt.mockError)
			}

			// Create handler
			authHandler := handler.NewAuthHandler(mockService)

			// Create test context
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Prepare request
			bodyBytes, _ := json.Marshal(tt.requestBody)
			c.Request = httptest.NewRequest(http.MethodPost, "/v1/auth/login", bytes.NewBuffer(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")

			// Call handler
			authHandler.Login(c)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.mockResponse != nil {
				mockService.AssertExpectations(t)
			}
		})
	}
}

func TestAuthHandler_RefreshToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		requestBody    dto.RefreshTokenRequest
		mockResponse   *dto.AuthResponse
		mockError      error
		expectedStatus int
	}{
		{
			name: "successful token refresh",
			requestBody: dto.RefreshTokenRequest{
				RefreshToken: "valid_refresh_token",
			},
			mockResponse: &dto.AuthResponse{
				AccessToken:  "new_access_token",
				RefreshToken: "new_refresh_token",
				ExpiresIn:    900,
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name: "missing refresh token",
			requestBody: dto.RefreshTokenRequest{
				RefreshToken: "",
			},
			mockResponse:   nil,
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock service
			mockService := new(MockAuthService)
			if tt.mockResponse != nil {
				mockService.On("RefreshToken", mock.Anything, mock.Anything).
					Return(tt.mockResponse, tt.mockError)
			}

			// Create handler
			authHandler := handler.NewAuthHandler(mockService)

			// Create test context
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Prepare request
			bodyBytes, _ := json.Marshal(tt.requestBody)
			c.Request = httptest.NewRequest(http.MethodPost, "/v1/auth/refresh", bytes.NewBuffer(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")

			// Call handler
			authHandler.RefreshToken(c)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.mockResponse != nil {
				mockService.AssertExpectations(t)
			}
		})
	}
}
