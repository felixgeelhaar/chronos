package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ascend/api/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupLoggerTest() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	return router
}

func TestLogger_SetsRequestID(t *testing.T) {
	router := setupLoggerTest()

	var capturedRequestID string

	router.Use(middleware.Logger())
	router.GET("/test", func(c *gin.Context) {
		requestID, exists := c.Get("request_id")
		require.True(t, exists)
		capturedRequestID = requestID.(string)
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.NotEmpty(t, capturedRequestID)
	// UUID format check (36 characters with dashes)
	assert.Len(t, capturedRequestID, 36)
}

func TestLogger_HandlesSuccessfulRequest(t *testing.T) {
	router := setupLoggerTest()

	router.Use(middleware.Logger())
	router.GET("/success", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest(http.MethodGet, "/success", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
}

func TestLogger_HandlesClientErrors(t *testing.T) {
	router := setupLoggerTest()

	router.Use(middleware.Logger())
	router.GET("/bad-request", func(c *gin.Context) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
	})

	req := httptest.NewRequest(http.MethodGet, "/bad-request", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusBadRequest, resp.Code)
}

func TestLogger_HandlesServerErrors(t *testing.T) {
	router := setupLoggerTest()

	router.Use(middleware.Logger())
	router.GET("/server-error", func(c *gin.Context) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server error"})
	})

	req := httptest.NewRequest(http.MethodGet, "/server-error", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusInternalServerError, resp.Code)
}

func TestLogger_IncludesQueryString(t *testing.T) {
	router := setupLoggerTest()

	router.Use(middleware.Logger())
	router.GET("/search", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"query": c.Query("q")})
	})

	req := httptest.NewRequest(http.MethodGet, "/search?q=test&limit=10", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Contains(t, resp.Body.String(), "test")
}

func TestLogger_HandlesMultipleStatusCodes(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
	}{
		{"200 OK", http.StatusOK},
		{"201 Created", http.StatusCreated},
		{"400 Bad Request", http.StatusBadRequest},
		{"401 Unauthorized", http.StatusUnauthorized},
		{"404 Not Found", http.StatusNotFound},
		{"500 Internal Server Error", http.StatusInternalServerError},
		{"503 Service Unavailable", http.StatusServiceUnavailable},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := setupLoggerTest()

			router.Use(middleware.Logger())
			router.GET("/test", func(c *gin.Context) {
				c.JSON(tt.statusCode, gin.H{"status": tt.statusCode})
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			assert.Equal(t, tt.statusCode, resp.Code)
		})
	}
}
