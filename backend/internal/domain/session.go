package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Session represents a training session
type Session struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID      `gorm:"type:uuid;not null;index:idx_sessions_user_date" json:"user_id"`
	Date      time.Time      `gorm:"type:date;not null;index:idx_sessions_user_date" json:"date"`
	Notes     *string        `gorm:"type:text" json:"notes,omitempty"`
	CreatedAt time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	User   User    `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Sets   []Set   `gorm:"foreignKey:SessionID;constraint:OnDelete:CASCADE" json:"sets,omitempty"`
	Videos []Video `gorm:"foreignKey:SessionID;constraint:OnDelete:CASCADE" json:"videos,omitempty"`
}

// TableName specifies the table name for Session model
func (Session) TableName() string {
	return "sessions"
}

// BeforeCreate hook to set UUID if not provided
func (s *Session) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}
