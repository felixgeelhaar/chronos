package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// Logger middleware logs HTTP requests with structured logging
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Generate request ID for tracing
		requestID := uuid.New().String()
		c.Set("request_id", requestID)

		// Start timer
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Get status code
		statusCode := c.Writer.Status()

		// Build full path with query
		if raw != "" {
			path = path + "?" + raw
		}

		// Log request with structured fields
		logEvent := log.Info()
		if statusCode >= 500 {
			logEvent = log.Error()
		} else if statusCode >= 400 {
			logEvent = log.Warn()
		}

		logEvent.
			Str("request_id", requestID).
			Str("method", c.Request.Method).
			Str("path", path).
			Int("status", statusCode).
			Dur("latency", latency).
			Str("client_ip", c.ClientIP()).
			Str("user_agent", c.Request.UserAgent()).
			Int("body_size", c.Writer.Size()).
			Msg("HTTP request")
	}
}
