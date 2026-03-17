package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ascend/api/internal/middleware"
	"github.com/ascend/api/pkg/auth"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupAuthTest() (*gin.Engine, *auth.JWTService) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	jwtService := auth.NewJWTService(auth.JWTConfig{
		AccessSecret:  "test-access-secret",
		RefreshSecret: "test-refresh-secret",
		AccessExpiry:  15 * time.Minute,
		RefreshExpiry: 24 * time.Hour,
	})

	return router, jwtService
}

func TestAuth_MissingAuthorizationHeader(t *testing.T) {
	router, jwtService := setupAuthTest()

	router.GET("/protected", middleware.Auth(jwtService), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusUnauthorized, resp.Code)
	assert.Contains(t, resp.Body.String(), "Authorization header required")
	assert.Contains(t, resp.Body.String(), "UNAUTHORIZED")
}

func TestAuth_InvalidAuthorizationFormat(t *testing.T) {
	router, jwtService := setupAuthTest()

	router.GET("/protected", middleware.Auth(jwtService), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	tests := []struct {
		name   string
		header string
	}{
		{"missing Bearer prefix", "some-token-here"},
		{"wrong prefix", "Basic some-token-here"},
		{"just Bearer", "Bearer"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			req.Header.Set("Authorization", tt.header)
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			assert.Equal(t, http.StatusUnauthorized, resp.Code)
			assert.Contains(t, resp.Body.String(), "Invalid authorization header format")
		})
	}
}

func TestAuth_InvalidOrExpiredToken(t *testing.T) {
	router, jwtService := setupAuthTest()

	router.GET("/protected", middleware.Auth(jwtService), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	tests := []struct {
		name  string
		token string
	}{
		{"invalid token", "invalid.token.here"},
		{"malformed token", "not-a-jwt-token"},
		{"empty token", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			req.Header.Set("Authorization", "Bearer "+tt.token)
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			assert.Equal(t, http.StatusUnauthorized, resp.Code)
			assert.Contains(t, resp.Body.String(), "Invalid or expired token")
		})
	}
}

func TestAuth_ValidToken(t *testing.T) {
	router, jwtService := setupAuthTest()

	var capturedUserID uuid.UUID
	var capturedEmail string

	router.GET("/protected", middleware.Auth(jwtService), func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		require.True(t, exists)
		capturedUserID = userID.(uuid.UUID)

		email, exists := c.Get("email")
		require.True(t, exists)
		capturedEmail = email.(string)

		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Generate valid token
	userID := uuid.New()
	email := "test@example.com"
	token, err := jwtService.GenerateAccessToken(userID, email)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, userID, capturedUserID)
	assert.Equal(t, email, capturedEmail)
	assert.Contains(t, resp.Body.String(), "success")
}
