package middleware

import (
	"github.com/gin-gonic/gin"
)

// CORS middleware handles Cross-Origin Resource Sharing
// Configured for EMEA-specific origins and mobile app access
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Allow mobile app and web origins (localhost for development)
		origin := c.Request.Header.Get("Origin")
		allowedOrigins := map[string]bool{
			"http://localhost:3000":          true, // Local development
			"http://localhost:19006":         true, // Expo dev
			"https://ascend.app":             true, // Production web
			"https://www.ascend.app":         true, // Production web (www)
			"capacitor://localhost":          true, // iOS
			"http://localhost":               true, // Android
			"ionic://localhost":              true, // Ionic
		}

		if allowedOrigins[origin] {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		}

		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")
		c.Writer.Header().Set("Access-Control-Max-Age", "86400") // 24 hours

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
