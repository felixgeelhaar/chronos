package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/ascend/api/internal/dto"
	"github.com/ascend/api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// VideoHandler handles video HTTP requests
type VideoHandler struct {
	videoService *service.VideoService
}

// NewVideoHandler creates a new video handler
func NewVideoHandler(videoService *service.VideoService) *VideoHandler {
	return &VideoHandler{
		videoService: videoService,
	}
}

// UploadVideo uploads a new video file
// POST /v1/videos
func (h *VideoHandler) UploadVideo(c *gin.Context) {
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

	// Parse multipart form
	err = c.Request.ParseMultipartForm(100 << 20) // 100 MB max
	if err != nil {
		log.Error().Err(err).Msg("Failed to parse multipart form")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "INVALID_REQUEST",
				"message": "Failed to parse form data",
			},
		})
		return
	}

	// Get file from form
	file, header, err := c.Request.FormFile("video")
	if err != nil {
		log.Error().Err(err).Msg("Failed to get video file from form")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "INVALID_REQUEST",
				"message": "Video file is required",
			},
		})
		return
	}
	defer file.Close()

	// Get optional metadata from form
	var req dto.UploadVideoRequest

	// Parse session_id if provided
	if sessionIDStr := c.Request.FormValue("session_id"); sessionIDStr != "" {
		sessionID, err := uuid.Parse(sessionIDStr)
		if err == nil {
			req.SessionID = &sessionID
		}
	}

	// Parse exercise_name if provided
	if exerciseName := c.Request.FormValue("exercise_name"); exerciseName != "" {
		req.ExerciseName = &exerciseName
	}

	// Alternatively, accept metadata as JSON in a separate field
	if metadataJSON := c.Request.FormValue("metadata"); metadataJSON != "" {
		if err := json.Unmarshal([]byte(metadataJSON), &req); err != nil {
			log.Warn().Err(err).Msg("Failed to parse metadata JSON, using form values")
		}
	}

	// Detect content type
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "video/mp4" // Default
	}

	response, err := h.videoService.UploadVideo(c.Request.Context(), userID, file, header.Filename, contentType, &req)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("Failed to upload video")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "UPLOAD_FAILED",
				"message": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusCreated, response)
}

// GetVideo retrieves a video by ID
// GET /v1/videos/:id
func (h *VideoHandler) GetVideo(c *gin.Context) {
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
				"message": "Invalid video ID",
			},
		})
		return
	}

	response, err := h.videoService.GetVideo(c.Request.Context(), userID, id)
	if err != nil {
		log.Error().Err(err).
			Str("user_id", userID.String()).
			Str("id", id.String()).
			Msg("Failed to get video")
		c.JSON(http.StatusNotFound, gin.H{
			"error": gin.H{
				"code":    "NOT_FOUND",
				"message": "Video not found",
			},
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// ListVideos retrieves all videos for the authenticated user
// GET /v1/videos?page=1&page_size=20
func (h *VideoHandler) ListVideos(c *gin.Context) {
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

	response, err := h.videoService.ListVideos(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("Failed to list videos")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "FETCH_FAILED",
				"message": "Failed to retrieve videos",
			},
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// ListVideosBySession retrieves all videos for a specific session
// GET /v1/sessions/:id/videos
func (h *VideoHandler) ListVideosBySession(c *gin.Context) {
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

	response, err := h.videoService.ListVideosBySession(c.Request.Context(), userID, sessionID)
	if err != nil {
		log.Error().Err(err).
			Str("user_id", userID.String()).
			Str("session_id", sessionID.String()).
			Msg("Failed to list videos by session")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "FETCH_FAILED",
				"message": "Failed to retrieve session videos",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"videos": response,
	})
}

// UpdateVideo updates video metadata
// PUT /v1/videos/:id
func (h *VideoHandler) UpdateVideo(c *gin.Context) {
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
				"message": "Invalid video ID",
			},
		})
		return
	}

	var req dto.UpdateVideoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Error().Err(err).Msg("Invalid update video request")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "INVALID_REQUEST",
				"message": "Invalid request body",
				"details": err.Error(),
			},
		})
		return
	}

	response, err := h.videoService.UpdateVideo(c.Request.Context(), userID, id, &req)
	if err != nil {
		log.Error().Err(err).
			Str("user_id", userID.String()).
			Str("id", id.String()).
			Msg("Failed to update video")
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

// DeleteVideo deletes a video
// DELETE /v1/videos/:id
func (h *VideoHandler) DeleteVideo(c *gin.Context) {
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
				"message": "Invalid video ID",
			},
		})
		return
	}

	err = h.videoService.DeleteVideo(c.Request.Context(), userID, id)
	if err != nil {
		log.Error().Err(err).
			Str("user_id", userID.String()).
			Str("id", id.String()).
			Msg("Failed to delete video")
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

// GeneratePresignedURL generates a temporary presigned URL for video access
// POST /v1/videos/:id/presigned-url
func (h *VideoHandler) GeneratePresignedURL(c *gin.Context) {
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
				"message": "Invalid video ID",
			},
		})
		return
	}

	// Default expiration: 1 hour
	expiration := 1 * time.Hour

	// Allow custom expiration via query parameter (in seconds)
	if expirationStr := c.Query("expiration"); expirationStr != "" {
		if seconds, err := strconv.Atoi(expirationStr); err == nil {
			expiration = time.Duration(seconds) * time.Second
		}
	}

	response, err := h.videoService.GeneratePresignedURL(c.Request.Context(), userID, id, expiration)
	if err != nil {
		log.Error().Err(err).
			Str("user_id", userID.String()).
			Str("id", id.String()).
			Msg("Failed to generate presigned URL")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "GENERATION_FAILED",
				"message": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, response)
}
