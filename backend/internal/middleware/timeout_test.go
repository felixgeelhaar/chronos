package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ascend/api/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupTimeoutTest() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	return router
}

func TestTimeout_RequestCompletesWithinTimeout(t *testing.T) {
	router := setupTimeoutTest()

	router.Use(middleware.Timeout(2 * time.Second))

	router.GET("/fast", func(c *gin.Context) {
		// Fast handler - completes immediately
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest(http.MethodGet, "/fast", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Contains(t, resp.Body.String(), "success")
}

func TestTimeout_RequestExceedsTimeout(t *testing.T) {
	router := setupTimeoutTest()

	// Very short timeout: 50ms
	router.Use(middleware.Timeout(50 * time.Millisecond))

	router.GET("/slow", func(c *gin.Context) {
		// Slow handler - takes 200ms
		time.Sleep(200 * time.Millisecond)
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest(http.MethodGet, "/slow", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusGatewayTimeout, resp.Code)
	assert.Contains(t, resp.Body.String(), "REQUEST_TIMEOUT")
	assert.Contains(t, resp.Body.String(), "Request processing timeout")
}

func TestTimeout_ContextCancellationPropagates(t *testing.T) {
	router := setupTimeoutTest()

	router.Use(middleware.Timeout(100 * time.Millisecond))

	contextCancelled := false

	router.GET("/check-context", func(c *gin.Context) {
		// Simulate a long-running operation that checks context
		select {
		case <-time.After(200 * time.Millisecond):
			c.JSON(http.StatusOK, gin.H{"message": "completed"})
		case <-c.Request.Context().Done():
			contextCancelled = true
			// Don't write response - timeout middleware will handle it
			return
		}
	})

	req := httptest.NewRequest(http.MethodGet, "/check-context", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	// The timeout should trigger, cancelling the context
	assert.True(t, contextCancelled, "Context should be cancelled")
	assert.Equal(t, http.StatusGatewayTimeout, resp.Code)
}

func TestTimeout_MultipleRequestsIndependent(t *testing.T) {
	router := setupTimeoutTest()

	router.Use(middleware.Timeout(100 * time.Millisecond))

	router.GET("/variable", func(c *gin.Context) {
		// Variable delay based on query parameter
		delay := c.Query("delay")
		if delay == "long" {
			time.Sleep(200 * time.Millisecond) // Exceeds timeout
		} else {
			time.Sleep(50 * time.Millisecond) // Within timeout
		}
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// First request: fast (should succeed)
	req1 := httptest.NewRequest(http.MethodGet, "/variable?delay=short", nil)
	resp1 := httptest.NewRecorder()

	router.ServeHTTP(resp1, req1)

	assert.Equal(t, http.StatusOK, resp1.Code)

	// Second request: slow (should timeout)
	req2 := httptest.NewRequest(http.MethodGet, "/variable?delay=long", nil)
	resp2 := httptest.NewRecorder()

	router.ServeHTTP(resp2, req2)

	assert.Equal(t, http.StatusGatewayTimeout, resp2.Code)

	// Third request: fast again (should succeed)
	req3 := httptest.NewRequest(http.MethodGet, "/variable?delay=short", nil)
	resp3 := httptest.NewRecorder()

	router.ServeHTTP(resp3, req3)

	assert.Equal(t, http.StatusOK, resp3.Code)
}

func TestTimeout_DifferentTimeoutDurations(t *testing.T) {
	tests := []struct {
		name           string
		timeout        time.Duration
		handlerDelay   time.Duration
		expectedStatus int
	}{
		{"fast handler, generous timeout", 500 * time.Millisecond, 50 * time.Millisecond, http.StatusOK},
		{"slow handler, short timeout", 50 * time.Millisecond, 200 * time.Millisecond, http.StatusGatewayTimeout},
		{"exact timing - just under", 100 * time.Millisecond, 90 * time.Millisecond, http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := setupTimeoutTest()

			router.Use(middleware.Timeout(tt.timeout))

			router.GET("/test", func(c *gin.Context) {
				time.Sleep(tt.handlerDelay)
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			assert.Equal(t, tt.expectedStatus, resp.Code)
		})
	}
}
