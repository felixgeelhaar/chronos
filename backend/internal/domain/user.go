package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User represents an athlete in the system
type User struct {
	ID         uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Email      string         `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	Password   string         `gorm:"type:varchar(255);not null" json:"-"` // Never serialize password
	Name       string         `gorm:"type:varchar(255);not null" json:"name"`
	BodyWeight *float64       `gorm:"type:decimal(5,2)" json:"body_weight,omitempty"`
	CreatedAt  time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt  time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Sessions    []Session    `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"sessions,omitempty"`
	OneRepMaxes []OneRepMax  `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"one_rep_maxes,omitempty"`
	Videos      []Video      `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"videos,omitempty"`
}

// TableName specifies the table name for User model
func (User) TableName() string {
	return "users"
}

// BeforeCreate hook to set UUID if not provided
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}
