package service

import (
	"context"
	"fmt"
	"math"

	"github.com/ascend/api/internal/domain"
	"github.com/ascend/api/internal/dto"
	"github.com/ascend/api/internal/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SessionServiceInterface defines the contract for session operations
type SessionServiceInterface interface {
	CreateSession(ctx context.Context, userID uuid.UUID, req *dto.CreateSessionRequest) (*dto.SessionResponse, error)
	GetSession(ctx context.Context, userID uuid.UUID, sessionID uuid.UUID) (*dto.SessionResponse, error)
	ListSessions(ctx context.Context, userID uuid.UUID, page, pageSize int) (*dto.SessionListResponse, error)
	UpdateSession(ctx context.Context, userID uuid.UUID, sessionID uuid.UUID, req *dto.UpdateSessionRequest) (*dto.SessionResponse, error)
	DeleteSession(ctx context.Context, userID uuid.UUID, sessionID uuid.UUID) error
}

// Verify that SessionService implements SessionServiceInterface
var _ SessionServiceInterface = (*SessionService)(nil)

// SessionService handles session business logic
type SessionService struct {
	sessionRepo repository.SessionRepository
	setRepo     repository.SetRepository
	db          *gorm.DB
}

// NewSessionService creates a new session service
func NewSessionService(
	sessionRepo repository.SessionRepository,
	setRepo repository.SetRepository,
	db *gorm.DB,
) *SessionService {
	return &SessionService{
		sessionRepo: sessionRepo,
		setRepo:     setRepo,
		db:          db,
	}
}

// CreateSession creates a new training session with sets
func (s *SessionService) CreateSession(ctx context.Context, userID uuid.UUID, req *dto.CreateSessionRequest) (*dto.SessionResponse, error) {
	// Create session
	session := &domain.Session{
		UserID: userID,
		Date:   req.Date,
		Notes:  req.Notes,
	}

	// Use transaction to ensure atomicity
	err := s.db.Transaction(func(tx *gorm.DB) error {
		// Create session
		if err := s.sessionRepo.Create(ctx, session); err != nil {
			return fmt.Errorf("failed to create session: %w", err)
		}

		// Create sets if provided
		if len(req.Sets) > 0 {
			sets := make([]*domain.Set, len(req.Sets))
			for i, setInput := range req.Sets {
				sets[i] = &domain.Set{
					SessionID:    session.ID,
					ExerciseName: setInput.ExerciseName,
					Weight:       setInput.Weight,
					Reps:         setInput.Reps,
					RPE:          setInput.RPE,
					Notes:        setInput.Notes,
					SetOrder:     setInput.SetOrder,
				}
			}

			if err := s.setRepo.CreateBulk(ctx, sets); err != nil {
				return fmt.Errorf("failed to create sets: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Fetch complete session with sets
	return s.GetSession(ctx, userID, session.ID)
}

// GetSession retrieves a session by ID
func (s *SessionService) GetSession(ctx context.Context, userID uuid.UUID, sessionID uuid.UUID) (*dto.SessionResponse, error) {
	session, err := s.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("session not found")
	}

	// Verify ownership
	if session.UserID != userID {
		return nil, fmt.Errorf("unauthorized access to session")
	}

	return s.sessionToResponse(session), nil
}

// ListSessions retrieves sessions for a user with pagination
func (s *SessionService) ListSessions(ctx context.Context, userID uuid.UUID, page, pageSize int) (*dto.SessionListResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	sessions, err := s.sessionRepo.GetByUserID(ctx, userID, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get sessions: %w", err)
	}

	// For each session, load its sets
	responses := make([]dto.SessionResponse, len(sessions))
	for i, session := range sessions {
		// Get sets for this session
		sets, err := s.setRepo.GetBySessionID(ctx, session.ID)
		if err == nil && len(sets) > 0 {
			// Convert []*Set to []Set
			session.Sets = make([]domain.Set, len(sets))
			for j, set := range sets {
				session.Sets[j] = *set
			}
		}
		responses[i] = *s.sessionToResponse(session)
	}

	// Get total count for pagination metadata
	totalCount, err := s.sessionRepo.CountByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to count sessions: %w", err)
	}

	totalPages := int(math.Ceil(float64(totalCount) / float64(pageSize)))

	return &dto.SessionListResponse{
		Sessions:   responses,
		Total:      int(totalCount),
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// UpdateSession updates a session
func (s *SessionService) UpdateSession(ctx context.Context, userID uuid.UUID, sessionID uuid.UUID, req *dto.UpdateSessionRequest) (*dto.SessionResponse, error) {
	// Get existing session
	session, err := s.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("session not found")
	}

	// Verify ownership
	if session.UserID != userID {
		return nil, fmt.Errorf("unauthorized access to session")
	}

	// Use transaction for atomicity
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// Update session fields
		if req.Date != nil {
			session.Date = *req.Date
		}
		if req.Notes != nil {
			session.Notes = req.Notes
		}

		if err := s.sessionRepo.Update(ctx, session); err != nil {
			return fmt.Errorf("failed to update session: %w", err)
		}

		// If sets are provided, replace all existing sets
		if req.Sets != nil {
			// Delete existing sets
			if err := s.setRepo.DeleteBySessionID(ctx, sessionID); err != nil {
				return fmt.Errorf("failed to delete existing sets: %w", err)
			}

			// Create new sets
			if len(req.Sets) > 0 {
				sets := make([]*domain.Set, len(req.Sets))
				for i, setInput := range req.Sets {
					sets[i] = &domain.Set{
						SessionID:    session.ID,
						ExerciseName: setInput.ExerciseName,
						Weight:       setInput.Weight,
						Reps:         setInput.Reps,
						RPE:          setInput.RPE,
						Notes:        setInput.Notes,
						SetOrder:     setInput.SetOrder,
					}
				}

				if err := s.setRepo.CreateBulk(ctx, sets); err != nil {
					return fmt.Errorf("failed to create sets: %w", err)
				}
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Fetch updated session with sets
	return s.GetSession(ctx, userID, session.ID)
}

// DeleteSession deletes a session
func (s *SessionService) DeleteSession(ctx context.Context, userID uuid.UUID, sessionID uuid.UUID) error {
	// Get session to verify ownership
	session, err := s.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("session not found")
	}

	// Verify ownership
	if session.UserID != userID {
		return fmt.Errorf("unauthorized access to session")
	}

	// Delete session (cascade will delete sets)
	if err := s.sessionRepo.Delete(ctx, sessionID); err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	return nil
}

// sessionToResponse converts domain session to response DTO
func (s *SessionService) sessionToResponse(session *domain.Session) *dto.SessionResponse {
	response := &dto.SessionResponse{
		ID:        session.ID,
		UserID:    session.UserID,
		Date:      session.Date,
		Notes:     session.Notes,
		CreatedAt: session.CreatedAt,
		UpdatedAt: session.UpdatedAt,
	}

	// Convert sets
	if len(session.Sets) > 0 {
		response.Sets = make([]dto.SetResponse, len(session.Sets))
		for i, set := range session.Sets {
			response.Sets[i] = dto.SetResponse{
				ID:           set.ID,
				SessionID:    set.SessionID,
				ExerciseName: set.ExerciseName,
				Weight:       set.Weight,
				Reps:         set.Reps,
				RPE:          set.RPE,
				Notes:        set.Notes,
				SetOrder:     set.SetOrder,
				CreatedAt:    set.CreatedAt,
				UpdatedAt:    set.UpdatedAt,
			}
		}
	}

	return response
}
