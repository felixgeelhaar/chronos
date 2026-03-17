package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ascend/api/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func setupRateLimiterTest() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	return router
}

func TestRateLimiter_Disabled(t *testing.T) {
	router := setupRateLimiterTest()

	router.Use(middleware.RateLimiter(middleware.RateLimiterConfig{
		Enabled: false,
	}))

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Make many requests - should all succeed since rate limiter is disabled
	for i := 0; i < 20; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code)
	}
}

func TestRateLimiter_AllowsRequestsUnderLimit(t *testing.T) {
	router := setupRateLimiterTest()

	// Very generous limits: 600 requests per minute (10 per second), burst of 100
	router.Use(middleware.RateLimiter(middleware.RateLimiterConfig{
		RequestsPerMinute: 600,
		Burst:             100,
		Enabled:           true,
	}))

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Make 10 requests quickly - should all succeed
	for i := 0; i < 10; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "192.168.1.1:1234" // Same IP
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code, "Request %d should succeed", i+1)
	}
}

func TestRateLimiter_BlocksRequestsOverLimit(t *testing.T) {
	router := setupRateLimiterTest()

	// Very restrictive: 1 request per minute, burst of 2
	router.Use(middleware.RateLimiter(middleware.RateLimiterConfig{
		RequestsPerMinute: 1,
		Burst:             2,
		Enabled:           true,
	}))

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// First 2 requests should succeed (burst allowance)
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "192.168.1.1:1234"
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code, "Request %d should succeed (within burst)", i+1)
	}

	// Third request should be blocked
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "192.168.1.1:1234"
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusTooManyRequests, resp.Code, "Request should be rate limited")
}

func TestRateLimiter_DifferentIPsHaveSeparateLimits(t *testing.T) {
	router := setupRateLimiterTest()

	// Restrictive: 1 request per minute, burst of 1
	router.Use(middleware.RateLimiter(middleware.RateLimiterConfig{
		RequestsPerMinute: 1,
		Burst:             1,
		Enabled:           true,
	}))

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// IP 1 makes a request - should succeed
	req1 := httptest.NewRequest(http.MethodGet, "/test", nil)
	req1.RemoteAddr = "192.168.1.1:1234"
	resp1 := httptest.NewRecorder()
	router.ServeHTTP(resp1, req1)
	assert.Equal(t, http.StatusOK, resp1.Code, "First IP should succeed")

	// IP 2 makes a request - should also succeed (different IP, different limiter)
	req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
	req2.RemoteAddr = "192.168.1.2:1234"
	resp2 := httptest.NewRecorder()
	router.ServeHTTP(resp2, req2)
	assert.Equal(t, http.StatusOK, resp2.Code, "Second IP should succeed (separate limiter)")

	// IP 1 makes another request - should be blocked
	req3 := httptest.NewRequest(http.MethodGet, "/test", nil)
	req3.RemoteAddr = "192.168.1.1:1234"
	resp3 := httptest.NewRecorder()
	router.ServeHTTP(resp3, req3)
	assert.Equal(t, http.StatusTooManyRequests, resp3.Code, "First IP should be rate limited on second request")
}

func TestRateLimiter_UsesDefaultsWhenNotSpecified(t *testing.T) {
	router := setupRateLimiterTest()

	// Don't specify RequestsPerMinute or Burst - should use defaults
	router.Use(middleware.RateLimiter(middleware.RateLimiterConfig{
		Enabled: true,
	}))

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Make 10 requests quickly - should all succeed with default generous limits
	for i := 0; i < 10; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "192.168.1.1:1234"
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code, "Request %d should succeed with default limits", i+1)
	}
}

func TestStrictRateLimiter(t *testing.T) {
	router := setupRateLimiterTest()

	// Use the StrictRateLimiter convenience function
	router.Use(middleware.StrictRateLimiter())

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// First 3 requests should succeed (burst of 3)
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "192.168.1.1:1234"
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code, "Request %d should succeed (within burst)", i+1)
	}

	// Fourth request should be blocked (burst exhausted)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "192.168.1.1:1234"
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusTooManyRequests, resp.Code, "Request should be rate limited after burst exhausted")
}

func TestUserRateLimiter_WithAuthenticatedUser(t *testing.T) {
	router := setupRateLimiterTest()

	// Very restrictive: 1 request per minute, burst of 1 (burst calculated as 1/6 of 60/6 = 1)
	router.Use(middleware.UserRateLimiter(6)) // 6 requests per minute = 1 burst

	router.GET("/test", func(c *gin.Context) {
		// Simulate Auth middleware setting user_id
		c.Set("user_id", uuid.New().String())
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	userID := uuid.New().String()

	// First request should succeed
	req1 := httptest.NewRequest(http.MethodGet, "/test", nil)
	w1 := httptest.NewRecorder()
	c1, _ := gin.CreateTestContext(w1)
	c1.Request = req1
	c1.Set("user_id", userID)
	router.ServeHTTP(w1, req1)

	// Need to manually inject user_id since we're not using full middleware chain
	// Let's use a different approach - add handler that sets user_id
	router2 := setupRateLimiterTest()
	router2.Use(func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Next()
	})
	router2.Use(middleware.UserRateLimiter(6))
	router2.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// First request - should succeed
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	resp := httptest.NewRecorder()
	router2.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusOK, resp.Code, "First request should succeed")

	// Second request - should be blocked (very restrictive limit)
	time.Sleep(10 * time.Millisecond) // Small delay to ensure distinct requests
	req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
	resp2 := httptest.NewRecorder()
	router2.ServeHTTP(resp2, req2)
	assert.Equal(t, http.StatusTooManyRequests, resp2.Code, "Second request should be rate limited")
}

func TestUserRateLimiter_WithoutAuthenticatedUser(t *testing.T) {
	router := setupRateLimiterTest()

	// No user in context - should fall back to allowing request
	router.Use(middleware.UserRateLimiter(60))

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	// Should succeed because no user_id in context means fallback behavior
	assert.Equal(t, http.StatusOK, resp.Code)
}

func TestUserRateLimiter_DifferentUsersHaveSeparateLimits(t *testing.T) {
	userID1 := uuid.New().String()
	userID2 := uuid.New().String()

	router := setupRateLimiterTest()

	// Add middleware that sets user_id based on header
	router.Use(func(c *gin.Context) {
		if userHeader := c.GetHeader("X-User-ID"); userHeader != "" {
			c.Set("user_id", userHeader)
		}
		c.Next()
	})

	// Restrictive: 6 requests per minute = burst of 1
	router.Use(middleware.UserRateLimiter(6))

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// User 1 makes a request - should succeed
	req1 := httptest.NewRequest(http.MethodGet, "/test", nil)
	req1.Header.Set("X-User-ID", userID1)
	resp1 := httptest.NewRecorder()
	router.ServeHTTP(resp1, req1)
	assert.Equal(t, http.StatusOK, resp1.Code, "User 1 first request should succeed")

	// User 2 makes a request - should also succeed (different user, different limiter)
	req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
	req2.Header.Set("X-User-ID", userID2)
	resp2 := httptest.NewRecorder()
	router.ServeHTTP(resp2, req2)
	assert.Equal(t, http.StatusOK, resp2.Code, "User 2 first request should succeed (separate limiter)")

	// User 1 makes another request - should be blocked
	req3 := httptest.NewRequest(http.MethodGet, "/test", nil)
	req3.Header.Set("X-User-ID", userID1)
	resp3 := httptest.NewRecorder()
	router.ServeHTTP(resp3, req3)
	assert.Equal(t, http.StatusTooManyRequests, resp3.Code, "User 1 second request should be rate limited")

	// User 2 makes another request - should also be blocked
	req4 := httptest.NewRequest(http.MethodGet, "/test", nil)
	req4.Header.Set("X-User-ID", userID2)
	resp4 := httptest.NewRecorder()
	router.ServeHTTP(resp4, req4)
	assert.Equal(t, http.StatusTooManyRequests, resp4.Code, "User 2 second request should be rate limited")
}
