package dto

import (
	"time"

	"github.com/google/uuid"
)

// UploadVideoRequest represents the request to upload a new video
// Note: Actual file upload is handled via multipart/form-data
type UploadVideoRequest struct {
	SessionID    *uuid.UUID `json:"session_id,omitempty"`
	ExerciseName *string    `json:"exercise_name,omitempty"`
}

// UpdateVideoRequest represents the request to update video metadata
type UpdateVideoRequest struct {
	SessionID    *uuid.UUID `json:"session_id,omitempty"`
	ExerciseName *string    `json:"exercise_name,omitempty"`
}

// VideoResponse represents a single video record
type VideoResponse struct {
	ID           uuid.UUID  `json:"id"`
	UserID       uuid.UUID  `json:"user_id"`
	SessionID    *uuid.UUID `json:"session_id,omitempty"`
	URL          string     `json:"url"`
	ThumbnailURL *string    `json:"thumbnail_url,omitempty"`
	Duration     *int       `json:"duration,omitempty"` // Duration in seconds
	FileSize     int64      `json:"file_size"`          // Size in bytes
	ExerciseName *string    `json:"exercise_name,omitempty"`
	Date         time.Time  `json:"date"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// VideoListResponse represents a list of videos
type VideoListResponse struct {
	Videos     []VideoResponse `json:"videos"`
	Total      int             `json:"total"`
	Page       int             `json:"page"`
	PageSize   int             `json:"page_size"`
	TotalPages int             `json:"total_pages"`
}

// GeneratePresignedURLResponse contains a presigned URL for video access
type GeneratePresignedURLResponse struct {
	URL       string    `json:"url"`
	ExpiresAt time.Time `json:"expires_at"`
}
