package dto

import (
	"time"

	"github.com/google/uuid"
)

// CreateSessionRequest represents the request to create a new session
type CreateSessionRequest struct {
	Date  time.Time        `json:"date" binding:"required"`
	Notes *string          `json:"notes,omitempty"`
	Sets  []CreateSetInput `json:"sets,omitempty"`
}

// UpdateSessionRequest represents the request to update a session
type UpdateSessionRequest struct {
	Date  *time.Time       `json:"date,omitempty"`
	Notes *string          `json:"notes,omitempty"`
	Sets  []CreateSetInput `json:"sets,omitempty"`
}

// CreateSetInput represents a set input when creating/updating a session
type CreateSetInput struct {
	ExerciseName string   `json:"exercise_name" binding:"required"`
	Weight       float64  `json:"weight" binding:"required,min=0"`
	Reps         int      `json:"reps" binding:"required,min=1"`
	RPE          *float64 `json:"rpe,omitempty"`
	Notes        *string  `json:"notes,omitempty"`
	SetOrder     int      `json:"set_order" binding:"required,min=1"`
}

// SessionResponse represents a session with all its data
type SessionResponse struct {
	ID        uuid.UUID     `json:"id"`
	UserID    uuid.UUID     `json:"user_id"`
	Date      time.Time     `json:"date"`
	Notes     *string       `json:"notes,omitempty"`
	Sets      []SetResponse `json:"sets,omitempty"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
}

// SessionListResponse represents a paginated list of sessions
type SessionListResponse struct {
	Sessions   []SessionResponse `json:"sessions"`
	Total      int               `json:"total"`
	Page       int               `json:"page"`
	PageSize   int               `json:"page_size"`
	TotalPages int               `json:"total_pages"`
}

// SetResponse represents a set in a session
type SetResponse struct {
	ID           uuid.UUID `json:"id"`
	SessionID    uuid.UUID `json:"session_id"`
	ExerciseName string    `json:"exercise_name"`
	Weight       float64   `json:"weight"`
	Reps         int       `json:"reps"`
	RPE          *float64  `json:"rpe,omitempty"`
	Notes        *string   `json:"notes,omitempty"`
	SetOrder     int       `json:"set_order"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
