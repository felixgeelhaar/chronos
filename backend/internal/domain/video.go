package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Video represents a training video
type Video struct {
	ID           uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID       uuid.UUID      `gorm:"type:uuid;not null;index:idx_videos_user" json:"user_id"`
	SessionID    *uuid.UUID     `gorm:"type:uuid;index:idx_videos_session" json:"session_id,omitempty"`
	SetID        *uuid.UUID     `gorm:"type:uuid;index:idx_videos_set" json:"set_id,omitempty"`
	URL          string         `gorm:"type:varchar(500);not null" json:"url"`
	ThumbnailURL *string        `gorm:"type:varchar(500)" json:"thumbnail_url,omitempty"`
	Duration     *int           `gorm:"type:integer" json:"duration,omitempty"` // in seconds
	FileSize     *int64         `gorm:"type:bigint" json:"file_size,omitempty"` // in bytes
	ExerciseName *string        `gorm:"type:varchar(255);index:idx_videos_exercise" json:"exercise_name,omitempty"`
	Date         time.Time      `gorm:"type:date;not null;index:idx_videos_date" json:"date"`
	CreatedAt    time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	User    User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Session *Session `gorm:"foreignKey:SessionID" json:"session,omitempty"`
	Set     *Set     `gorm:"foreignKey:SetID" json:"set,omitempty"`
}

// TableName specifies the table name for Video model
func (Video) TableName() string {
	return "videos"
}

// BeforeCreate hook to set UUID if not provided
func (v *Video) BeforeCreate(tx *gorm.DB) error {
	if v.ID == uuid.Nil {
		v.ID = uuid.New()
	}
	return nil
}
