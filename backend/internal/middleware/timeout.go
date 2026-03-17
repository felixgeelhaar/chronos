package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Timeout middleware sets a timeout for request processing
func Timeout(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Create context with timeout
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		// Replace request context
		c.Request = c.Request.WithContext(ctx)

		// Channel to track if handler finished
		finished := make(chan struct{})

		// Run handler in goroutine
		go func() {
			c.Next()
			finished <- struct{}{}
		}()

		// Wait for handler to finish or timeout
		select {
		case <-finished:
			// Handler completed successfully
			return
		case <-ctx.Done():
			// Timeout occurred
			c.JSON(http.StatusGatewayTimeout, gin.H{
				"error": gin.H{
					"code":    "REQUEST_TIMEOUT",
					"message": "Request processing timeout",
				},
			})
			c.Abort()
		}
	}
}
