package service

import (
	"context"
	"fmt"
	"math"

	"github.com/ascend/api/internal/domain"
	"github.com/ascend/api/internal/dto"
	"github.com/ascend/api/internal/repository"
	"github.com/google/uuid"
)

// OneRepMaxServiceInterface defines the contract for one rep max operations
type OneRepMaxServiceInterface interface {
	CreateOneRepMax(ctx context.Context, userID uuid.UUID, req *dto.CreateOneRepMaxRequest) (*dto.OneRepMaxResponse, error)
	GetOneRepMax(ctx context.Context, userID uuid.UUID, id uuid.UUID) (*dto.OneRepMaxResponse, error)
	ListOneRepMaxes(ctx context.Context, userID uuid.UUID, page, pageSize int) (*dto.OneRepMaxListResponse, error)
	GetOneRepMaxHistory(ctx context.Context, userID uuid.UUID, exerciseName string) (*dto.OneRepMaxHistoryResponse, error)
	UpdateOneRepMax(ctx context.Context, userID uuid.UUID, id uuid.UUID, req *dto.UpdateOneRepMaxRequest) (*dto.OneRepMaxResponse, error)
	DeleteOneRepMax(ctx context.Context, userID uuid.UUID, id uuid.UUID) error
}

// Verify that OneRepMaxService implements OneRepMaxServiceInterface
var _ OneRepMaxServiceInterface = (*OneRepMaxService)(nil)

// OneRepMaxService handles one rep max business logic
type OneRepMaxService struct {
	oneRepMaxRepo repository.OneRepMaxRepository
}

// NewOneRepMaxService creates a new one rep max service
func NewOneRepMaxService(oneRepMaxRepo repository.OneRepMaxRepository) *OneRepMaxService {
	return &OneRepMaxService{
		oneRepMaxRepo: oneRepMaxRepo,
	}
}

// CreateOneRepMax creates a new 1RM record
func (s *OneRepMaxService) CreateOneRepMax(ctx context.Context, userID uuid.UUID, req *dto.CreateOneRepMaxRequest) (*dto.OneRepMaxResponse, error) {
	oneRepMax := &domain.OneRepMax{
		UserID:       userID,
		ExerciseName: req.ExerciseName,
		Weight:       req.Weight,
		Date:         req.Date,
	}

	if err := s.oneRepMaxRepo.Create(ctx, oneRepMax); err != nil {
		return nil, fmt.Errorf("failed to create one rep max: %w", err)
	}

	return s.toResponse(oneRepMax), nil
}

// GetOneRepMax retrieves a 1RM record by ID
func (s *OneRepMaxService) GetOneRepMax(ctx context.Context, userID uuid.UUID, id uuid.UUID) (*dto.OneRepMaxResponse, error) {
	oneRepMax, err := s.oneRepMaxRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("one rep max not found")
	}

	// Verify ownership
	if oneRepMax.UserID != userID {
		return nil, fmt.Errorf("unauthorized access to one rep max")
	}

	return s.toResponse(oneRepMax), nil
}

// ListOneRepMaxes retrieves all 1RM records for a user
func (s *OneRepMaxService) ListOneRepMaxes(ctx context.Context, userID uuid.UUID, page, pageSize int) (*dto.OneRepMaxListResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// Get all records for user
	records, err := s.oneRepMaxRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get one rep maxes: %w", err)
	}

	// Convert to responses
	responses := make([]dto.OneRepMaxResponse, len(records))
	for i, record := range records {
		responses[i] = *s.toResponse(record)
	}

	// Calculate pagination
	total := len(responses)
	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))

	// Apply pagination
	start := (page - 1) * pageSize
	end := start + pageSize
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	paginatedRecords := responses[start:end]

	return &dto.OneRepMaxListResponse{
		Records:    paginatedRecords,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// GetOneRepMaxHistory retrieves the history of 1RM for a specific exercise
func (s *OneRepMaxService) GetOneRepMaxHistory(ctx context.Context, userID uuid.UUID, exerciseName string) (*dto.OneRepMaxHistoryResponse, error) {
	records, err := s.oneRepMaxRepo.GetByUserIDAndExercise(ctx, userID, exerciseName)
	if err != nil {
		return nil, fmt.Errorf("failed to get one rep max history: %w", err)
	}

	if len(records) == 0 {
		return &dto.OneRepMaxHistoryResponse{
			ExerciseName:  exerciseName,
			CurrentRecord: nil,
			History:       []dto.OneRepMaxResponse{},
			PersonalBest:  0,
			Improvement:   0,
		}, nil
	}

	// Convert to responses
	history := make([]dto.OneRepMaxResponse, len(records))
	personalBest := 0.0
	var currentRecord *dto.OneRepMaxResponse

	for i, record := range records {
		response := s.toResponse(record)
		history[i] = *response

		// Track personal best
		if record.Weight > personalBest {
			personalBest = record.Weight
		}

		// Most recent record is current (records are sorted by date DESC)
		if i == 0 {
			currentRecord = response
		}
	}

	// Calculate improvement (from earliest to current)
	improvement := 0.0
	if len(records) > 1 {
		earliest := records[len(records)-1]
		current := records[0]
		if earliest.Weight > 0 {
			improvement = ((current.Weight - earliest.Weight) / earliest.Weight) * 100
		}
	}

	return &dto.OneRepMaxHistoryResponse{
		ExerciseName:  exerciseName,
		CurrentRecord: currentRecord,
		History:       history,
		PersonalBest:  personalBest,
		Improvement:   improvement,
	}, nil
}

// UpdateOneRepMax updates a 1RM record
func (s *OneRepMaxService) UpdateOneRepMax(ctx context.Context, userID uuid.UUID, id uuid.UUID, req *dto.UpdateOneRepMaxRequest) (*dto.OneRepMaxResponse, error) {
	// Get existing record
	oneRepMax, err := s.oneRepMaxRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("one rep max not found")
	}

	// Verify ownership
	if oneRepMax.UserID != userID {
		return nil, fmt.Errorf("unauthorized access to one rep max")
	}

	// Update fields
	if req.Weight != nil {
		oneRepMax.Weight = *req.Weight
	}
	if req.Date != nil {
		oneRepMax.Date = *req.Date
	}

	if err := s.oneRepMaxRepo.Update(ctx, oneRepMax); err != nil {
		return nil, fmt.Errorf("failed to update one rep max: %w", err)
	}

	return s.toResponse(oneRepMax), nil
}

// DeleteOneRepMax deletes a 1RM record
func (s *OneRepMaxService) DeleteOneRepMax(ctx context.Context, userID uuid.UUID, id uuid.UUID) error {
	// Get record to verify ownership
	oneRepMax, err := s.oneRepMaxRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("one rep max not found")
	}

	// Verify ownership
	if oneRepMax.UserID != userID {
		return fmt.Errorf("unauthorized access to one rep max")
	}

	if err := s.oneRepMaxRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete one rep max: %w", err)
	}

	return nil
}

// toResponse converts domain model to response DTO
func (s *OneRepMaxService) toResponse(oneRepMax *domain.OneRepMax) *dto.OneRepMaxResponse {
	return &dto.OneRepMaxResponse{
		ID:           oneRepMax.ID,
		UserID:       oneRepMax.UserID,
		ExerciseName: oneRepMax.ExerciseName,
		Weight:       oneRepMax.Weight,
		Date:         oneRepMax.Date,
		CreatedAt:    oneRepMax.CreatedAt,
		UpdatedAt:    oneRepMax.UpdatedAt,
	}
}
