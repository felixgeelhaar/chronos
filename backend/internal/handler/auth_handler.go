package handler

import (
	"net/http"

	"github.com/ascend/api/internal/dto"
	apperrors "github.com/ascend/api/internal/errors"
	"github.com/ascend/api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// AuthHandler handles authentication HTTP requests
type AuthHandler struct {
	authService service.AuthServiceInterface
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService service.AuthServiceInterface) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Register handles user registration
// POST /v1/auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Error().Err(err).Msg("Invalid registration request")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "INVALID_REQUEST",
				"message": "Invalid request body",
				"details": err.Error(),
			},
		})
		return
	}

	response, err := h.authService.Register(c.Request.Context(), &req)
	if err != nil {
		log.Error().Err(err).Str("email", req.Email).Msg("Registration failed")

		// Check error type and return appropriate HTTP status
		statusCode := http.StatusBadRequest
		if apperrors.IsConflict(err) {
			statusCode = http.StatusConflict
		}

		c.JSON(statusCode, gin.H{
			"error": gin.H{
				"code":    "REGISTRATION_FAILED",
				"message": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusCreated, response)
}

// Login handles user login
// POST /v1/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Error().Err(err).Msg("Invalid login request")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "INVALID_REQUEST",
				"message": "Invalid request body",
				"details": err.Error(),
			},
		})
		return
	}

	response, err := h.authService.Login(c.Request.Context(), &req)
	if err != nil {
		log.Error().Err(err).Str("email", req.Email).Msg("Login failed")
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{
				"code":    "AUTHENTICATION_FAILED",
				"message": "Invalid email or password",
			},
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// RefreshToken handles token refresh
// POST /v1/auth/refresh
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Error().Err(err).Msg("Invalid refresh token request")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "INVALID_REQUEST",
				"message": "Invalid request body",
				"details": err.Error(),
			},
		})
		return
	}

	response, err := h.authService.RefreshToken(c.Request.Context(), &req)
	if err != nil {
		log.Error().Err(err).Msg("Token refresh failed")
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{
				"code":    "INVALID_REFRESH_TOKEN",
				"message": "Invalid or expired refresh token",
			},
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetMe returns the current authenticated user
// GET /v1/auth/me
func (h *AuthHandler) GetMe(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userIDValue, exists := c.Get("user_id")
	if !exists {
		log.Error().Msg("User ID not found in context")
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{
				"code":    "UNAUTHORIZED",
				"message": "Authentication required",
			},
		})
		return
	}

	userID, ok := userIDValue.(uuid.UUID)
	if !ok {
		log.Error().Msg("Invalid user ID in context")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_ERROR",
				"message": "Invalid user ID",
			},
		})
		return
	}

	user, err := h.authService.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("Failed to get user")
		c.JSON(http.StatusNotFound, gin.H{
			"error": gin.H{
				"code":    "USER_NOT_FOUND",
				"message": "User not found",
			},
		})
		return
	}

	c.JSON(http.StatusOK, user)
}
