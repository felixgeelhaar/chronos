package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ascend/api/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupCORSTest() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.CORS())
	return router
}

func TestCORS_AllowedOrigin(t *testing.T) {
	router := setupCORSTest()
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	allowedOrigins := []string{
		"http://localhost:3000",
		"http://localhost:19006",
		"https://ascend.app",
		"https://www.ascend.app",
		"capacitor://localhost",
		"http://localhost",
		"ionic://localhost",
	}

	for _, origin := range allowedOrigins {
		t.Run("origin_"+origin, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set("Origin", origin)
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			assert.Equal(t, http.StatusOK, resp.Code)
			assert.Equal(t, origin, resp.Header().Get("Access-Control-Allow-Origin"))
			assert.Equal(t, "true", resp.Header().Get("Access-Control-Allow-Credentials"))
			assert.Contains(t, resp.Header().Get("Access-Control-Allow-Headers"), "Authorization")
			assert.Contains(t, resp.Header().Get("Access-Control-Allow-Methods"), "GET")
			assert.Contains(t, resp.Header().Get("Access-Control-Allow-Methods"), "POST")
			assert.Equal(t, "86400", resp.Header().Get("Access-Control-Max-Age"))
		})
	}
}

func TestCORS_DisallowedOrigin(t *testing.T) {
	router := setupCORSTest()
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	disallowedOrigins := []string{
		"https://evil.com",
		"http://malicious-site.com",
		"http://localhost:8080", // Not in allowed list
	}

	for _, origin := range disallowedOrigins {
		t.Run("origin_"+origin, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set("Origin", origin)
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			assert.Equal(t, http.StatusOK, resp.Code)
			// Should not set Access-Control-Allow-Origin for disallowed origins
			assert.Empty(t, resp.Header().Get("Access-Control-Allow-Origin"))
			// But should still set credentials and other headers
			assert.Equal(t, "true", resp.Header().Get("Access-Control-Allow-Credentials"))
		})
	}
}

func TestCORS_PreflightRequest(t *testing.T) {
	router := setupCORSTest()
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest(http.MethodOptions, "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	// OPTIONS requests should return 204 No Content
	assert.Equal(t, http.StatusNoContent, resp.Code)
	assert.Equal(t, "http://localhost:3000", resp.Header().Get("Access-Control-Allow-Origin"))
	assert.Contains(t, resp.Header().Get("Access-Control-Allow-Methods"), "OPTIONS")
	assert.Empty(t, resp.Body.String())
}

func TestCORS_NoOriginHeader(t *testing.T) {
	router := setupCORSTest()
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	// No Origin header set
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
	// Should not set Access-Control-Allow-Origin when no Origin header
	assert.Empty(t, resp.Header().Get("Access-Control-Allow-Origin"))
	// But should still set credentials and other headers
	assert.Equal(t, "true", resp.Header().Get("Access-Control-Allow-Credentials"))
}
