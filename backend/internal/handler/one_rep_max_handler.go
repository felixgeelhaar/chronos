package handler

import (
	"net/http"
	"strconv"

	"github.com/ascend/api/internal/dto"
	"github.com/ascend/api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// OneRepMaxHandler handles one rep max HTTP requests
type OneRepMaxHandler struct {
	oneRepMaxService service.OneRepMaxServiceInterface
}

// NewOneRepMaxHandler creates a new one rep max handler
func NewOneRepMaxHandler(oneRepMaxService service.OneRepMaxServiceInterface) *OneRepMaxHandler {
	return &OneRepMaxHandler{
		oneRepMaxService: oneRepMaxService,
	}
}

// CreateOneRepMax creates a new 1RM record
// POST /v1/one-rep-maxes
func (h *OneRepMaxHandler) CreateOneRepMax(c *gin.Context) {
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

	var req dto.CreateOneRepMaxRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Error().Err(err).Msg("Invalid create one rep max request")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "INVALID_REQUEST",
				"message": "Invalid request body",
				"details": err.Error(),
			},
		})
		return
	}

	response, err := h.oneRepMaxService.CreateOneRepMax(c.Request.Context(), userID, &req)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("Failed to create one rep max")
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

// GetOneRepMax retrieves a 1RM record by ID
// GET /v1/one-rep-maxes/:id
func (h *OneRepMaxHandler) GetOneRepMax(c *gin.Context) {
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

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "INVALID_ID",
				"message": "Invalid one rep max ID",
			},
		})
		return
	}

	response, err := h.oneRepMaxService.GetOneRepMax(c.Request.Context(), userID, id)
	if err != nil {
		log.Error().Err(err).
			Str("user_id", userID.String()).
			Str("id", id.String()).
			Msg("Failed to get one rep max")
		c.JSON(http.StatusNotFound, gin.H{
			"error": gin.H{
				"code":    "NOT_FOUND",
				"message": "One rep max not found",
			},
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// ListOneRepMaxes retrieves all 1RM records for the authenticated user
// GET /v1/one-rep-maxes?page=1&page_size=20
func (h *OneRepMaxHandler) ListOneRepMaxes(c *gin.Context) {
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

	response, err := h.oneRepMaxService.ListOneRepMaxes(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("Failed to list one rep maxes")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "FETCH_FAILED",
				"message": "Failed to retrieve one rep maxes",
			},
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetOneRepMaxHistory retrieves the history of 1RM for a specific exercise
// GET /v1/one-rep-maxes/exercise/:name/history
func (h *OneRepMaxHandler) GetOneRepMaxHistory(c *gin.Context) {
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

	exerciseName := c.Param("name")
	if exerciseName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "INVALID_REQUEST",
				"message": "Exercise name is required",
			},
		})
		return
	}

	response, err := h.oneRepMaxService.GetOneRepMaxHistory(c.Request.Context(), userID, exerciseName)
	if err != nil {
		log.Error().Err(err).
			Str("user_id", userID.String()).
			Str("exercise", exerciseName).
			Msg("Failed to get one rep max history")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "FETCH_FAILED",
				"message": "Failed to retrieve one rep max history",
			},
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// UpdateOneRepMax updates a 1RM record
// PUT /v1/one-rep-maxes/:id
func (h *OneRepMaxHandler) UpdateOneRepMax(c *gin.Context) {
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

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "INVALID_ID",
				"message": "Invalid one rep max ID",
			},
		})
		return
	}

	var req dto.UpdateOneRepMaxRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Error().Err(err).Msg("Invalid update one rep max request")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "INVALID_REQUEST",
				"message": "Invalid request body",
				"details": err.Error(),
			},
		})
		return
	}

	response, err := h.oneRepMaxService.UpdateOneRepMax(c.Request.Context(), userID, id, &req)
	if err != nil {
		log.Error().Err(err).
			Str("user_id", userID.String()).
			Str("id", id.String()).
			Msg("Failed to update one rep max")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "UPDATE_FAILED",
				"message": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// DeleteOneRepMax deletes a 1RM record
// DELETE /v1/one-rep-maxes/:id
func (h *OneRepMaxHandler) DeleteOneRepMax(c *gin.Context) {
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

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "INVALID_ID",
				"message": "Invalid one rep max ID",
			},
		})
		return
	}

	err = h.oneRepMaxService.DeleteOneRepMax(c.Request.Context(), userID, id)
	if err != nil {
		log.Error().Err(err).
			Str("user_id", userID.String()).
			Str("id", id.String()).
			Msg("Failed to delete one rep max")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "DELETE_FAILED",
				"message": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
