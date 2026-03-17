package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/ascend/api/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SessionRepository defines the interface for session data access
type SessionRepository interface {
	Create(ctx context.Context, session *domain.Session) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Session, error)
	GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.Session, error)
	GetByUserIDAndDateRange(ctx context.Context, userID uuid.UUID, startDate, endDate time.Time) ([]*domain.Session, error)
	CountByUserID(ctx context.Context, userID uuid.UUID) (int64, error)
	Update(ctx context.Context, session *domain.Session) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// sessionRepository implements SessionRepository
type sessionRepository struct {
	db *gorm.DB
}

// NewSessionRepository creates a new session repository
func NewSessionRepository(db *gorm.DB) SessionRepository {
	return &sessionRepository{db: db}
}

// Create creates a new session
func (r *sessionRepository) Create(ctx context.Context, session *domain.Session) error {
	if err := r.db.WithContext(ctx).Create(session).Error; err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	return nil
}

// GetByID retrieves a session by ID with sets and videos
func (r *sessionRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Session, error) {
	var session domain.Session
	if err := r.db.WithContext(ctx).
		Preload("Sets", func(db *gorm.DB) *gorm.DB {
			return db.Order("set_order ASC")
		}).
		Preload("Videos").
		First(&session, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("session not found")
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}
	return &session, nil
}

// GetByUserID retrieves sessions for a user with pagination
func (r *sessionRepository) GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.Session, error) {
	var sessions []*domain.Session
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("date DESC").
		Limit(limit).
		Offset(offset).
		Find(&sessions).Error; err != nil {
		return nil, fmt.Errorf("failed to get sessions: %w", err)
	}
	return sessions, nil
}

// GetByUserIDAndDateRange retrieves sessions for a user within a date range
func (r *sessionRepository) GetByUserIDAndDateRange(ctx context.Context, userID uuid.UUID, startDate, endDate time.Time) ([]*domain.Session, error) {
	var sessions []*domain.Session
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND date >= ? AND date <= ?", userID, startDate, endDate).
		Order("date DESC").
		Preload("Sets", func(db *gorm.DB) *gorm.DB {
			return db.Order("set_order ASC")
		}).
		Find(&sessions).Error; err != nil {
		return nil, fmt.Errorf("failed to get sessions by date range: %w", err)
	}
	return sessions, nil
}

// CountByUserID counts the total number of sessions for a user
func (r *sessionRepository) CountByUserID(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&domain.Session{}).
		Where("user_id = ?", userID).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count sessions: %w", err)
	}
	return count, nil
}

// Update updates a session
func (r *sessionRepository) Update(ctx context.Context, session *domain.Session) error {
	if err := r.db.WithContext(ctx).Save(session).Error; err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}
	return nil
}

// Delete soft deletes a session and cascades to sets and videos
func (r *sessionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&domain.Session{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}
	return nil
}
