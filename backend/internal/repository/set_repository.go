package repository

import (
	"context"
	"fmt"

	"github.com/ascend/api/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SetRepository defines the interface for set data access
type SetRepository interface {
	Create(ctx context.Context, set *domain.Set) error
	CreateBulk(ctx context.Context, sets []*domain.Set) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Set, error)
	GetBySessionID(ctx context.Context, sessionID uuid.UUID) ([]*domain.Set, error)
	Update(ctx context.Context, set *domain.Set) error
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteBySessionID(ctx context.Context, sessionID uuid.UUID) error
}

// setRepository implements SetRepository
type setRepository struct {
	db *gorm.DB
}

// NewSetRepository creates a new set repository
func NewSetRepository(db *gorm.DB) SetRepository {
	return &setRepository{db: db}
}

// Create creates a new set
func (r *setRepository) Create(ctx context.Context, set *domain.Set) error {
	if err := r.db.WithContext(ctx).Create(set).Error; err != nil {
		return fmt.Errorf("failed to create set: %w", err)
	}
	return nil
}

// CreateBulk creates multiple sets in a transaction
func (r *setRepository) CreateBulk(ctx context.Context, sets []*domain.Set) error {
	if len(sets) == 0 {
		return nil
	}

	if err := r.db.WithContext(ctx).CreateInBatches(sets, 100).Error; err != nil {
		return fmt.Errorf("failed to create sets in bulk: %w", err)
	}
	return nil
}

// GetByID retrieves a set by ID
func (r *setRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Set, error) {
	var set domain.Set
	if err := r.db.WithContext(ctx).First(&set, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("set not found")
		}
		return nil, fmt.Errorf("failed to get set: %w", err)
	}
	return &set, nil
}

// GetBySessionID retrieves all sets for a session ordered by set_order
func (r *setRepository) GetBySessionID(ctx context.Context, sessionID uuid.UUID) ([]*domain.Set, error) {
	var sets []*domain.Set
	if err := r.db.WithContext(ctx).
		Where("session_id = ?", sessionID).
		Order("set_order ASC").
		Find(&sets).Error; err != nil {
		return nil, fmt.Errorf("failed to get sets by session: %w", err)
	}
	return sets, nil
}

// Update updates a set
func (r *setRepository) Update(ctx context.Context, set *domain.Set) error {
	if err := r.db.WithContext(ctx).Save(set).Error; err != nil {
		return fmt.Errorf("failed to update set: %w", err)
	}
	return nil
}

// Delete soft deletes a set
func (r *setRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&domain.Set{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("failed to delete set: %w", err)
	}
	return nil
}

// DeleteBySessionID deletes all sets for a session
func (r *setRepository) DeleteBySessionID(ctx context.Context, sessionID uuid.UUID) error {
	if err := r.db.WithContext(ctx).
		Where("session_id = ?", sessionID).
		Delete(&domain.Set{}).Error; err != nil {
		return fmt.Errorf("failed to delete sets by session: %w", err)
	}
	return nil
}
