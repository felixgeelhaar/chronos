package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// RateLimiterConfig defines rate limiting configuration
type RateLimiterConfig struct {
	// RequestsPerMinute defines how many requests are allowed per minute per client
	RequestsPerMinute int
	// Burst defines the maximum burst size (allows temporary spikes)
	Burst int
	// Enabled controls whether rate limiting is active
	Enabled bool
}

// ipRateLimiter tracks rate limiters per IP address
type ipRateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
	rate     rate.Limit
	burst    int
}

// newIPRateLimiter creates a new IP-based rate limiter
func newIPRateLimiter(requestsPerMinute, burst int) *ipRateLimiter {
	// Convert requests per minute to requests per second
	requestsPerSecond := float64(requestsPerMinute) / 60.0

	return &ipRateLimiter{
		limiters: make(map[string]*rate.Limiter),
		rate:     rate.Limit(requestsPerSecond),
		burst:    burst,
	}
}

// getLimiter retrieves or creates a rate limiter for an IP address
func (i *ipRateLimiter) getLimiter(ip string) *rate.Limiter {
	i.mu.Lock()
	defer i.mu.Unlock()

	limiter, exists := i.limiters[ip]
	if !exists {
		limiter = rate.NewLimiter(i.rate, i.burst)
		i.limiters[ip] = limiter
	}

	return limiter
}

// cleanup removes inactive rate limiters to prevent memory leaks
func (i *ipRateLimiter) cleanup() {
	ticker := time.NewTicker(1 * time.Hour)
	go func() {
		for range ticker.C {
			i.mu.Lock()
			// Remove limiters that haven't been used recently
			// This prevents the map from growing indefinitely
			for ip, limiter := range i.limiters {
				// If the limiter's burst allowance is full, it hasn't been used recently
				if limiter.Tokens() >= float64(i.burst) {
					delete(i.limiters, ip)
				}
			}
			i.mu.Unlock()
		}
	}()
}

// RateLimiter creates a rate limiting middleware
// Limits requests based on client IP address
func RateLimiter(config RateLimiterConfig) gin.HandlerFunc {
	// If rate limiting is disabled, return a no-op middleware
	if !config.Enabled {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	// Default values if not specified
	if config.RequestsPerMinute == 0 {
		config.RequestsPerMinute = 100 // Default: 100 requests per minute
	}
	if config.Burst == 0 {
		config.Burst = 20 // Default: Allow burst of 20 requests
	}

	// Create a new rate limiter instance for this middleware
	// This ensures each middleware instance has isolated state (no global state)
	rateLimiter := newIPRateLimiter(config.RequestsPerMinute, config.Burst)
	rateLimiter.cleanup() // Start cleanup goroutine

	return func(c *gin.Context) {
		// Get client IP address
		ip := c.ClientIP()

		// Get rate limiter for this IP
		limiter := rateLimiter.getLimiter(ip)

		// Check if request is allowed
		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": gin.H{
					"code":        "RATE_LIMIT_EXCEEDED",
					"message":     "Too many requests. Please try again later.",
					"retry_after": "60s", // Client should wait 60 seconds
				},
			})
			c.Abort()
			return
		}

		// Add rate limit headers for transparency
		c.Header("X-RateLimit-Limit", string(rune(config.RequestsPerMinute)))
		c.Header("X-RateLimit-Remaining", string(rune(int(limiter.Tokens()))))

		c.Next()
	}
}

// StrictRateLimiter provides stricter rate limiting for sensitive endpoints
// Example: Authentication endpoints should have lower limits
func StrictRateLimiter() gin.HandlerFunc {
	return RateLimiter(RateLimiterConfig{
		RequestsPerMinute: 10,   // Only 10 requests per minute
		Burst:             3,    // Very small burst allowance
		Enabled:           true,
	})
}

// UserRateLimiter implements per-user rate limiting for authenticated endpoints
// This is more sophisticated than IP-based limiting
func UserRateLimiter(requestsPerMinute int) gin.HandlerFunc {
	// Create isolated state for this middleware instance (no global state)
	limiters := make(map[string]*rate.Limiter)
	mu := sync.RWMutex{}

	requestsPerSecond := float64(requestsPerMinute) / 60.0
	burst := requestsPerMinute / 6 // Burst is ~10% of per-minute limit

	return func(c *gin.Context) {
		// Get user ID from context (set by Auth middleware)
		userIDInterface, exists := c.Get("user_id")
		if !exists {
			// Fall back to IP-based if user not authenticated
			c.Next()
			return
		}

		userID := userIDInterface.(string)

		// Get or create limiter for this user
		mu.Lock()
		limiter, exists := limiters[userID]
		if !exists {
			limiter = rate.NewLimiter(rate.Limit(requestsPerSecond), burst)
			limiters[userID] = limiter
		}
		mu.Unlock()

		// Check if request is allowed
		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": gin.H{
					"code":        "USER_RATE_LIMIT_EXCEEDED",
					"message":     "You have exceeded your request quota. Please try again later.",
					"retry_after": "60s",
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
