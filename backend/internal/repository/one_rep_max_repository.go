package repository

import (
	"context"
	"fmt"

	"github.com/ascend/api/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// OneRepMaxRepository defines the interface for one rep max data access
type OneRepMaxRepository interface {
	Create(ctx context.Context, orm *domain.OneRepMax) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.OneRepMax, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.OneRepMax, error)
	GetByUserIDAndExercise(ctx context.Context, userID uuid.UUID, exerciseName string) ([]*domain.OneRepMax, error)
	GetLatestByUserIDAndExercise(ctx context.Context, userID uuid.UUID, exerciseName string) (*domain.OneRepMax, error)
	Update(ctx context.Context, orm *domain.OneRepMax) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// oneRepMaxRepository implements OneRepMaxRepository
type oneRepMaxRepository struct {
	db *gorm.DB
}

// NewOneRepMaxRepository creates a new one rep max repository
func NewOneRepMaxRepository(db *gorm.DB) OneRepMaxRepository {
	return &oneRepMaxRepository{db: db}
}

// Create creates a new one rep max record
func (r *oneRepMaxRepository) Create(ctx context.Context, orm *domain.OneRepMax) error {
	if err := r.db.WithContext(ctx).Create(orm).Error; err != nil {
		return fmt.Errorf("failed to create one rep max: %w", err)
	}
	return nil
}

// GetByID retrieves a one rep max by ID
func (r *oneRepMaxRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.OneRepMax, error) {
	var orm domain.OneRepMax
	if err := r.db.WithContext(ctx).First(&orm, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("one rep max not found")
		}
		return nil, fmt.Errorf("failed to get one rep max: %w", err)
	}
	return &orm, nil
}

// GetByUserID retrieves all one rep maxes for a user
func (r *oneRepMaxRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.OneRepMax, error) {
	var orms []*domain.OneRepMax
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("date DESC").
		Find(&orms).Error; err != nil {
		return nil, fmt.Errorf("failed to get one rep maxes: %w", err)
	}
	return orms, nil
}

// GetByUserIDAndExercise retrieves all one rep max records for a user and exercise
func (r *oneRepMaxRepository) GetByUserIDAndExercise(ctx context.Context, userID uuid.UUID, exerciseName string) ([]*domain.OneRepMax, error) {
	var orms []*domain.OneRepMax
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND exercise_name = ?", userID, exerciseName).
		Order("date DESC").
		Find(&orms).Error; err != nil {
		return nil, fmt.Errorf("failed to get one rep maxes by exercise: %w", err)
	}
	return orms, nil
}

// GetLatestByUserIDAndExercise retrieves the most recent one rep max for a user and exercise
func (r *oneRepMaxRepository) GetLatestByUserIDAndExercise(ctx context.Context, userID uuid.UUID, exerciseName string) (*domain.OneRepMax, error) {
	var orm domain.OneRepMax
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND exercise_name = ?", userID, exerciseName).
		Order("date DESC").
		First(&orm).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("one rep max not found")
		}
		return nil, fmt.Errorf("failed to get latest one rep max: %w", err)
	}
	return &orm, nil
}

// Update updates a one rep max
func (r *oneRepMaxRepository) Update(ctx context.Context, orm *domain.OneRepMax) error {
	if err := r.db.WithContext(ctx).Save(orm).Error; err != nil {
		return fmt.Errorf("failed to update one rep max: %w", err)
	}
	return nil
}

// Delete soft deletes a one rep max
func (r *oneRepMaxRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&domain.OneRepMax{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("failed to delete one rep max: %w", err)
	}
	return nil
}
