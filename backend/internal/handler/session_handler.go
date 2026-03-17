package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/ascend/api/internal/dto"
	"github.com/ascend/api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// SessionHandler handles session HTTP requests
type SessionHandler struct {
	sessionService service.SessionServiceInterface
}

// NewSessionHandler creates a new session handler
func NewSessionHandler(sessionService service.SessionServiceInterface) *SessionHandler {
	return &SessionHandler{
		sessionService: sessionService,
	}
}

// CreateSession creates a new training session
// POST /v1/sessions
func (h *SessionHandler) CreateSession(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{
				"code":    "UNAUTHORIZED",
				"message": "User not authenticated",
			},
		})
		return
	}

	var req dto.CreateSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Error().Err(err).Msg("Invalid create session request")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "INVALID_REQUEST",
				"message": "Invalid request body",
				"details": err.Error(),
			},
		})
		return
	}

	response, err := h.sessionService.CreateSession(c.Request.Context(), userID, &req)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("Failed to create session")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "CREATE_FAILED",
				"message": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusCreated, response)
}

// GetSession retrieves a session by ID
// GET /v1/sessions/:id
func (h *SessionHandler) GetSession(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{
				"code":    "UNAUTHORIZED",
				"message": "User not authenticated",
			},
		})
		return
	}

	sessionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "INVALID_ID",
				"message": "Invalid session ID",
			},
		})
		return
	}

	response, err := h.sessionService.GetSession(c.Request.Context(), userID, sessionID)
	if err != nil {
		log.Error().Err(err).
			Str("user_id", userID.String()).
			Str("session_id", sessionID.String()).
			Msg("Failed to get session")
		c.JSON(http.StatusNotFound, gin.H{
			"error": gin.H{
				"code":    "SESSION_NOT_FOUND",
				"message": "Session not found",
			},
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// ListSessions retrieves sessions for the authenticated user
// GET /v1/sessions?page=1&page_size=20
func (h *SessionHandler) ListSessions(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{
				"code":    "UNAUTHORIZED",
				"message": "User not authenticated",
			},
		})
		return
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	response, err := h.sessionService.ListSessions(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("Failed to list sessions")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "FETCH_FAILED",
				"message": "Failed to retrieve sessions",
			},
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// UpdateSession updates a session
// PUT /v1/sessions/:id
func (h *SessionHandler) UpdateSession(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{
				"code":    "UNAUTHORIZED",
				"message": "User not authenticated",
			},
		})
		return
	}

	sessionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "INVALID_ID",
				"message": "Invalid session ID",
			},
		})
		return
	}

	var req dto.UpdateSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Error().Err(err).Msg("Invalid update session request")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "INVALID_REQUEST",
				"message": "Invalid request body",
				"details": err.Error(),
			},
		})
		return
	}

	response, err := h.sessionService.UpdateSession(c.Request.Context(), userID, sessionID, &req)
	if err != nil {
		log.Error().Err(err).
			Str("user_id", userID.String()).
			Str("session_id", sessionID.String()).
			Msg("Failed to update session")
		c.JSON(http.StatusNotFound, gin.H{
			"error": gin.H{
				"code":    "SESSION_NOT_FOUND",
				"message": "Session not found",
			},
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// DeleteSession deletes a session
// DELETE /v1/sessions/:id
func (h *SessionHandler) DeleteSession(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{
				"code":    "UNAUTHORIZED",
				"message": "User not authenticated",
			},
		})
		return
	}

	sessionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "INVALID_ID",
				"message": "Invalid session ID",
			},
		})
		return
	}

	err = h.sessionService.DeleteSession(c.Request.Context(), userID, sessionID)
	if err != nil {
		log.Error().Err(err).
			Str("user_id", userID.String()).
			Str("session_id", sessionID.String()).
			Msg("Failed to delete session")
		c.JSON(http.StatusNotFound, gin.H{
			"error": gin.H{
				"code":    "SESSION_NOT_FOUND",
				"message": "Session not found",
			},
		})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// getUserIDFromContext extracts user ID from request context
func getUserIDFromContext(c *gin.Context) (uuid.UUID, error) {
	userIDValue, exists := c.Get("user_id")
	if !exists {
		return uuid.Nil, fmt.Errorf("user_id not found in context")
	}

	userID, ok := userIDValue.(uuid.UUID)
	if !ok {
		return uuid.Nil, fmt.Errorf("invalid user_id type in context")
	}

	return userID, nil
}
