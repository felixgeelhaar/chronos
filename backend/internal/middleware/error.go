package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// ErrorHandler middleware catches panics and converts them to proper error responses
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Log the panic
				requestID, _ := c.Get("request_id")
				log.Error().
					Str("request_id", requestID.(string)).
					Interface("error", err).
					Str("path", c.Request.URL.Path).
					Msg("Panic recovered")

				// Return 500 Internal Server Error
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": gin.H{
						"code":      "INTERNAL_SERVER_ERROR",
						"message":   "An internal server error occurred",
						"requestId": requestID,
					},
				})
				c.Abort()
			}
		}()

		c.Next()
	}
}
