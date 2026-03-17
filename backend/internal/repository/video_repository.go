package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/ascend/api/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// VideoRepository defines the interface for video data access
type VideoRepository interface {
	Create(ctx context.Context, video *domain.Video) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Video, error)
	GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.Video, error)
	GetBySessionID(ctx context.Context, sessionID uuid.UUID) ([]*domain.Video, error)
	GetByUserIDAndDateRange(ctx context.Context, userID uuid.UUID, startDate, endDate time.Time) ([]*domain.Video, error)
	Update(ctx context.Context, video *domain.Video) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// videoRepository implements VideoRepository
type videoRepository struct {
	db *gorm.DB
}

// NewVideoRepository creates a new video repository
func NewVideoRepository(db *gorm.DB) VideoRepository {
	return &videoRepository{db: db}
}

// Create creates a new video
func (r *videoRepository) Create(ctx context.Context, video *domain.Video) error {
	if err := r.db.WithContext(ctx).Create(video).Error; err != nil {
		return fmt.Errorf("failed to create video: %w", err)
	}
	return nil
}

// GetByID retrieves a video by ID
func (r *videoRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Video, error) {
	var video domain.Video
	if err := r.db.WithContext(ctx).First(&video, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("video not found")
		}
		return nil, fmt.Errorf("failed to get video: %w", err)
	}
	return &video, nil
}

// GetByUserID retrieves videos for a user with pagination
func (r *videoRepository) GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.Video, error) {
	var videos []*domain.Video
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("date DESC").
		Limit(limit).
		Offset(offset).
		Find(&videos).Error; err != nil {
		return nil, fmt.Errorf("failed to get videos: %w", err)
	}
	return videos, nil
}

// GetBySessionID retrieves all videos for a session
func (r *videoRepository) GetBySessionID(ctx context.Context, sessionID uuid.UUID) ([]*domain.Video, error) {
	var videos []*domain.Video
	if err := r.db.WithContext(ctx).
		Where("session_id = ?", sessionID).
		Order("created_at ASC").
		Find(&videos).Error; err != nil {
		return nil, fmt.Errorf("failed to get videos by session: %w", err)
	}
	return videos, nil
}

// GetByUserIDAndDateRange retrieves videos for a user within a date range
func (r *videoRepository) GetByUserIDAndDateRange(ctx context.Context, userID uuid.UUID, startDate, endDate time.Time) ([]*domain.Video, error) {
	var videos []*domain.Video
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND date >= ? AND date <= ?", userID, startDate, endDate).
		Order("date DESC").
		Find(&videos).Error; err != nil {
		return nil, fmt.Errorf("failed to get videos by date range: %w", err)
	}
	return videos, nil
}

// Update updates a video
func (r *videoRepository) Update(ctx context.Context, video *domain.Video) error {
	if err := r.db.WithContext(ctx).Save(video).Error; err != nil {
		return fmt.Errorf("failed to update video: %w", err)
	}
	return nil
}

// Delete soft deletes a video
func (r *videoRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&domain.Video{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("failed to delete video: %w", err)
	}
	return nil
}
