package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Set represents a single set of an exercise
type Set struct {
	ID           uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	SessionID    uuid.UUID      `gorm:"type:uuid;not null;index:idx_sets_session" json:"session_id"`
	ExerciseName string         `gorm:"type:varchar(255);not null;index:idx_sets_exercise" json:"exercise_name"`
	Weight       float64        `gorm:"type:decimal(6,2);not null" json:"weight"`
	Reps         int            `gorm:"type:integer;not null" json:"reps"`
	RPE          *float64       `gorm:"type:decimal(3,1)" json:"rpe,omitempty"`
	Notes        *string        `gorm:"type:text" json:"notes,omitempty"`
	SetOrder     int            `gorm:"type:integer;not null;default:1" json:"set_order"`
	CreatedAt    time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Session Session `gorm:"foreignKey:SessionID" json:"session,omitempty"`
}

// TableName specifies the table name for Set model
func (Set) TableName() string {
	return "sets"
}

// BeforeCreate hook to set UUID if not provided
func (s *Set) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}
