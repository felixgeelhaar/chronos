package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ascend/api/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestSecurityHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name             string
		environment      string
		expectedHeaders  map[string]string
		notExpectedHeaders []string
	}{
		{
			name:        "development environment",
			environment: "development",
			expectedHeaders: map[string]string{
				"X-Frame-Options":                   "DENY",
				"X-Content-Type-Options":            "nosniff",
				"X-XSS-Protection":                  "1; mode=block",
				"Content-Security-Policy":           "default-src 'none'; frame-ancestors 'none'; base-uri 'self'; form-action 'self'",
				"Referrer-Policy":                   "no-referrer",
				"X-Permitted-Cross-Domain-Policies": "none",
				"Cache-Control":                     "no-store, no-cache, must-revalidate, proxy-revalidate",
				"Pragma":                            "no-cache",
				"Expires":                           "0",
				"Server":                            "Ascend-API",
			},
			notExpectedHeaders: []string{"Strict-Transport-Security"}, // HSTS only in production
		},
		{
			name:        "production environment",
			environment: "production",
			expectedHeaders: map[string]string{
				"X-Frame-Options":                   "DENY",
				"X-Content-Type-Options":            "nosniff",
				"X-XSS-Protection":                  "1; mode=block",
				"Strict-Transport-Security":         "max-age=31536000; includeSubDomains; preload",
				"Content-Security-Policy":           "default-src 'none'; frame-ancestors 'none'; base-uri 'self'; form-action 'self'",
				"Referrer-Policy":                   "no-referrer",
				"X-Permitted-Cross-Domain-Policies": "none",
				"Cache-Control":                     "no-store, no-cache, must-revalidate, proxy-revalidate",
				"Server":                            "Ascend-API",
			},
			notExpectedHeaders: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create router with security headers middleware
			router := gin.New()
			router.Use(middleware.SecurityHeaders(tt.environment))

			// Add test endpoint
			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "test"})
			})

			// Create request
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/test", nil)

			// Serve request
			router.ServeHTTP(w, req)

			// Assert expected headers
			for header, expectedValue := range tt.expectedHeaders {
				actualValue := w.Header().Get(header)
				assert.Equal(t, expectedValue, actualValue, "Header %s should be %s", header, expectedValue)
			}

			// Assert headers that should NOT be present
			for _, header := range tt.notExpectedHeaders {
				actualValue := w.Header().Get(header)
				assert.Empty(t, actualValue, "Header %s should not be present in %s", header, tt.environment)
			}
		})
	}
}

func TestSecurityHeaders_XFrameOptions(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(middleware.SecurityHeaders("development"))
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	router.ServeHTTP(w, req)

	// X-Frame-Options should prevent clickjacking
	assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
}

func TestSecurityHeaders_CSP(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(middleware.SecurityHeaders("development"))
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	router.ServeHTTP(w, req)

	// CSP should be restrictive for API backend
	csp := w.Header().Get("Content-Security-Policy")
	assert.Contains(t, csp, "default-src 'none'")
	assert.Contains(t, csp, "frame-ancestors 'none'")
}

func TestSecurityHeaders_NoServerIdentification(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(middleware.SecurityHeaders("development"))
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	router.ServeHTTP(w, req)

	// Server header should be generic, not revealing technology stack
	server := w.Header().Get("Server")
	assert.Equal(t, "Ascend-API", server)
	assert.NotContains(t, server, "gin", "Should not reveal Gin framework")
	assert.NotContains(t, server, "go", "Should not reveal Go version")
}

func TestSecurityHeaders_CacheControl(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(middleware.SecurityHeaders("development"))
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	router.ServeHTTP(w, req)

	// Cache headers should prevent sensitive data caching
	assert.Equal(t, "no-store, no-cache, must-revalidate, proxy-revalidate", w.Header().Get("Cache-Control"))
	assert.Equal(t, "no-cache", w.Header().Get("Pragma"))
	assert.Equal(t, "0", w.Header().Get("Expires"))
}

func TestSecurityHeaders_PermissionsPolicy(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(middleware.SecurityHeaders("development"))
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	router.ServeHTTP(w, req)

	// Permissions-Policy should disable all features for API
	permissions := w.Header().Get("Permissions-Policy")
	assert.Contains(t, permissions, "camera=()")
	assert.Contains(t, permissions, "microphone=()")
	assert.Contains(t, permissions, "geolocation=()")
}

func TestSecurityHeaders_HSTS_ProductionOnly(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name        string
		environment string
		expectHSTS  bool
	}{
		{
			name:        "HSTS in production",
			environment: "production",
			expectHSTS:  true,
		},
		{
			name:        "No HSTS in development",
			environment: "development",
			expectHSTS:  false,
		},
		{
			name:        "No HSTS in staging",
			environment: "staging",
			expectHSTS:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(middleware.SecurityHeaders(tt.environment))
			router.GET("/test", func(c *gin.Context) {
				c.Status(http.StatusOK)
			})

			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			router.ServeHTTP(w, req)

			hsts := w.Header().Get("Strict-Transport-Security")
			if tt.expectHSTS {
				assert.NotEmpty(t, hsts)
				assert.Contains(t, hsts, "max-age=31536000")
				assert.Contains(t, hsts, "includeSubDomains")
				assert.Contains(t, hsts, "preload")
			} else {
				assert.Empty(t, hsts, "HSTS should not be set in non-production environments")
			}
		})
	}
}

func TestSecurityHeaders_AllEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(middleware.SecurityHeaders("development"))

	// Add multiple endpoints
	router.GET("/api/users", func(c *gin.Context) { c.Status(http.StatusOK) })
	router.POST("/api/auth/login", func(c *gin.Context) { c.Status(http.StatusOK) })
	router.PUT("/api/sessions/123", func(c *gin.Context) { c.Status(http.StatusOK) })
	router.DELETE("/api/videos/456", func(c *gin.Context) { c.Status(http.StatusOK) })

	endpoints := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/api/users"},
		{http.MethodPost, "/api/auth/login"},
		{http.MethodPut, "/api/sessions/123"},
		{http.MethodDelete, "/api/videos/456"},
	}

	for _, endpoint := range endpoints {
		t.Run(endpoint.method+"_"+endpoint.path, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(endpoint.method, endpoint.path, nil)
			router.ServeHTTP(w, req)

			// All endpoints should have security headers
			assert.NotEmpty(t, w.Header().Get("X-Frame-Options"))
			assert.NotEmpty(t, w.Header().Get("X-Content-Type-Options"))
			assert.NotEmpty(t, w.Header().Get("Content-Security-Policy"))
		})
	}
}

// Test that SecureHeaders alias works
func TestSecureHeaders_Alias(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(middleware.SecureHeaders("development")) // Using alias
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	router.ServeHTTP(w, req)

	// Should work identically to SecurityHeaders()
	assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
}
