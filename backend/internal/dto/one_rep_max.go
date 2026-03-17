package dto

import (
	"time"

	"github.com/google/uuid"
)

// CreateOneRepMaxRequest represents the request to create a new 1RM record
type CreateOneRepMaxRequest struct {
	ExerciseName string    `json:"exercise_name" binding:"required"`
	Weight       float64   `json:"weight" binding:"required,min=0"`
	Date         time.Time `json:"date" binding:"required"`
}

// UpdateOneRepMaxRequest represents the request to update a 1RM record
type UpdateOneRepMaxRequest struct {
	Weight *float64   `json:"weight,omitempty" binding:"omitempty,min=0"`
	Date   *time.Time `json:"date,omitempty"`
}

// OneRepMaxResponse represents a single 1RM record
type OneRepMaxResponse struct {
	ID           uuid.UUID `json:"id"`
	UserID       uuid.UUID `json:"user_id"`
	ExerciseName string    `json:"exercise_name"`
	Weight       float64   `json:"weight"`
	Date         time.Time `json:"date"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// OneRepMaxListResponse represents a list of 1RM records
type OneRepMaxListResponse struct {
	Records    []OneRepMaxResponse `json:"records"`
	Total      int                 `json:"total"`
	Page       int                 `json:"page"`
	PageSize   int                 `json:"page_size"`
	TotalPages int                 `json:"total_pages"`
}

// OneRepMaxHistoryResponse represents the history of 1RM for an exercise
type OneRepMaxHistoryResponse struct {
	ExerciseName  string                `json:"exercise_name"`
	CurrentRecord *OneRepMaxResponse    `json:"current_record,omitempty"`
	History       []OneRepMaxResponse   `json:"history"`
	PersonalBest  float64               `json:"personal_best"`
	Improvement   float64               `json:"improvement"` // Percentage improvement from first to current
}
