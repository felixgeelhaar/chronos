package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ascend/api/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupErrorTest() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	return router
}

func TestErrorHandler_CatchesPanic(t *testing.T) {
	router := setupErrorTest()

	// Set request_id in context for the error handler to use
	router.Use(func(c *gin.Context) {
		c.Set("request_id", "test-request-id-123")
		c.Next()
	})

	router.Use(middleware.ErrorHandler())

	router.GET("/panic", func(c *gin.Context) {
		panic("something went wrong!")
	})

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusInternalServerError, resp.Code)
	assert.Contains(t, resp.Body.String(), "INTERNAL_SERVER_ERROR")
	assert.Contains(t, resp.Body.String(), "An internal server error occurred")
	assert.Contains(t, resp.Body.String(), "test-request-id-123")
}

func TestErrorHandler_NoPanicContinuesNormally(t *testing.T) {
	router := setupErrorTest()

	router.Use(func(c *gin.Context) {
		c.Set("request_id", "test-request-id-456")
		c.Next()
	})

	router.Use(middleware.ErrorHandler())

	router.GET("/success", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest(http.MethodGet, "/success", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Contains(t, resp.Body.String(), "success")
}

func TestErrorHandler_PanicWithDifferentTypes(t *testing.T) {
	tests := []struct {
		name      string
		panicVal  interface{}
	}{
		{"string panic", "error occurred"},
		{"error panic", assert.AnError},
		{"int panic", 500},
		{"nil panic", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := setupErrorTest()

			router.Use(func(c *gin.Context) {
				c.Set("request_id", "test-request-id")
				c.Next()
			})

			router.Use(middleware.ErrorHandler())

			router.GET("/panic", func(c *gin.Context) {
				panic(tt.panicVal)
			})

			req := httptest.NewRequest(http.MethodGet, "/panic", nil)
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			assert.Equal(t, http.StatusInternalServerError, resp.Code)
			assert.Contains(t, resp.Body.String(), "INTERNAL_SERVER_ERROR")
		})
	}
}
