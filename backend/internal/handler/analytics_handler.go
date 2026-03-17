package handler

import (
	"net/http"
	"time"

	"github.com/ascend/api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// AnalyticsHandler handles analytics HTTP requests
type AnalyticsHandler struct {
	analyticsService service.AnalyticsServiceInterface
}

// NewAnalyticsHandler creates a new analytics handler
func NewAnalyticsHandler(analyticsService service.AnalyticsServiceInterface) *AnalyticsHandler {
	return &AnalyticsHandler{
		analyticsService: analyticsService,
	}
}

// GetExerciseHistory retrieves the history of a specific exercise
// GET /v1/analytics/exercise/:name?start_date=2024-01-01&end_date=2024-12-31
func (h *AnalyticsHandler) GetExerciseHistory(c *gin.Context) {
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

	// Parse date range (default to last 90 days)
	endDate := time.Now()
	startDate := endDate.AddDate(0, -3, 0)

	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if parsed, err := time.Parse("2006-01-02", startDateStr); err == nil {
			startDate = parsed
		}
	}

	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if parsed, err := time.Parse("2006-01-02", endDateStr); err == nil {
			endDate = parsed
		}
	}

	response, err := h.analyticsService.GetExerciseHistory(c.Request.Context(), userID, exerciseName, startDate, endDate)
	if err != nil {
		log.Error().Err(err).Str("exercise", exerciseName).Msg("Failed to get exercise history")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "FETCH_FAILED",
				"message": "Failed to retrieve exercise history",
			},
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetACWR retrieves Acute:Chronic Workload Ratio
// GET /v1/analytics/acwr?exercise=Squat
func (h *AnalyticsHandler) GetACWR(c *gin.Context) {
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

	// Exercise name is optional - if not provided, calculates for all exercises
	exerciseName := c.Query("exercise")

	response, err := h.analyticsService.CalculateACWR(c.Request.Context(), userID, exerciseName)
	if err != nil {
		log.Error().Err(err).Str("exercise", exerciseName).Msg("Failed to calculate ACWR")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "CALCULATION_FAILED",
				"message": "Failed to calculate ACWR",
			},
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetVolumeProgress retrieves volume progress over time
// GET /v1/analytics/volume?exercise=Squat&period=month
func (h *AnalyticsHandler) GetVolumeProgress(c *gin.Context) {
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

	exerciseName := c.Query("exercise")
	period := c.DefaultQuery("period", "month")

	// Validate period
	if period != "week" && period != "month" && period != "year" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "INVALID_REQUEST",
				"message": "Period must be one of: week, month, year",
			},
		})
		return
	}

	response, err := h.analyticsService.GetVolumeProgress(c.Request.Context(), userID, exerciseName, period)
	if err != nil {
		log.Error().Err(err).
			Str("exercise", exerciseName).
			Str("period", period).
			Msg("Failed to get volume progress")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "FETCH_FAILED",
				"message": "Failed to retrieve volume progress",
			},
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetProgressSummary retrieves overall progress summary
// GET /v1/analytics/summary?start_date=2024-01-01&end_date=2024-12-31
func (h *AnalyticsHandler) GetProgressSummary(c *gin.Context) {
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

	// Parse date range (default to last 30 days)
	endDate := time.Now()
	startDate := endDate.AddDate(0, -1, 0)

	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if parsed, err := time.Parse("2006-01-02", startDateStr); err == nil {
			startDate = parsed
		}
	}

	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if parsed, err := time.Parse("2006-01-02", endDateStr); err == nil {
			endDate = parsed
		}
	}

	response, err := h.analyticsService.GetProgressSummary(c.Request.Context(), userID, startDate, endDate)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get progress summary")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "FETCH_FAILED",
				"message": "Failed to retrieve progress summary",
			},
		})
		return
	}

	c.JSON(http.StatusOK, response)
}
