package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// OneRepMax represents a one-rep max record for an exercise
type OneRepMax struct {
	ID           uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID       uuid.UUID      `gorm:"type:uuid;not null;index:idx_one_rep_maxes_user_exercise" json:"user_id"`
	ExerciseName string         `gorm:"type:varchar(255);not null;index:idx_one_rep_maxes_user_exercise" json:"exercise_name"`
	Weight       float64        `gorm:"type:decimal(6,2);not null" json:"weight"`
	Date         time.Time      `gorm:"type:date;not null;index:idx_one_rep_maxes_date" json:"date"`
	CreatedAt    time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// TableName specifies the table name for OneRepMax model
func (OneRepMax) TableName() string {
	return "one_rep_maxes"
}

// BeforeCreate hook to set UUID if not provided
func (o *OneRepMax) BeforeCreate(tx *gorm.DB) error {
	if o.ID == uuid.Nil {
		o.ID = uuid.New()
	}
	return nil
}
