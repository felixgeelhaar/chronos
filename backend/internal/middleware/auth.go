package middleware

import (
	"net/http"
	"strings"

	"github.com/ascend/api/pkg/auth"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// Auth creates an authentication middleware
func Auth(jwtService *auth.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			log.Warn().Msg("Missing authorization header")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"code":    "UNAUTHORIZED",
					"message": "Authorization header required",
				},
			})
			c.Abort()
			return
		}

		// Check if it's a Bearer token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			log.Warn().Str("auth_header", authHeader).Msg("Invalid authorization header format")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"code":    "UNAUTHORIZED",
					"message": "Invalid authorization header format. Expected 'Bearer <token>'",
				},
			})
			c.Abort()
			return
		}

		token := parts[1]

		// Validate token
		claims, err := jwtService.ValidateAccessToken(token)
		if err != nil {
			log.Warn().Err(err).Msg("Invalid or expired token")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"code":    "UNAUTHORIZED",
					"message": "Invalid or expired token",
				},
			})
			c.Abort()
			return
		}

		// Set user information in context
		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)

		log.Debug().
			Str("user_id", claims.UserID.String()).
			Str("email", claims.Email).
			Msg("User authenticated")

		c.Next()
	}
}
