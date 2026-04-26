// Package memory provides an in-memory implementation of store.Store for testing.
package memory

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/felixgeelhaar/chronos/pkg/insight"
	"github.com/felixgeelhaar/chronos/pkg/vector"
	"github.com/google/uuid"
)

// Store is a thread-safe in-memory implementation of store.Store.
type Store struct {
	mu            sync.RWMutex
	entityStates  map[uuid.UUID][]vector.EntityState // keyed by scopeID
	insights      map[uuid.UUID][]insight.Insight    // keyed by scopeID
	feedback      map[uuid.UUID]insight.Feedback     // keyed by insightID
}

// New creates a new in-memory store.
func New() *Store {
	return &Store{
		entityStates: make(map[uuid.UUID][]vector.EntityState),
		insights:     make(map[uuid.UUID][]insight.Insight),
		feedback:     make(map[uuid.UUID]insight.Feedback),
	}
}

// Close is a no-op for the in-memory store.
func (s *Store) Close() error {
	return nil
}

// SaveEntityStates persists entity states.
func (s *Store) SaveEntityStates(_ context.Context, adapter string, states []vector.EntityState) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, st := range states {
		st := st
		st.Meta["adapter"] = adapter
		s.entityStates[st.ScopeID] = append(s.entityStates[st.ScopeID], st)
	}
	return nil
}

// LoadEntityStates retrieves states for a scope.
func (s *Store) LoadEntityStates(_ context.Context, scopeID uuid.UUID) ([]vector.EntityState, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	states := make([]vector.EntityState, len(s.entityStates[scopeID]))
	copy(states, s.entityStates[scopeID])
	return states, nil
}

// LoadEntityStatesByEntity retrieves states for a specific entity.
func (s *Store) LoadEntityStatesByEntity(_ context.Context, entityID uuid.UUID) ([]vector.EntityState, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []vector.EntityState
	for _, states := range s.entityStates {
		for _, st := range states {
			if st.EntityID == entityID {
				result = append(result, st)
			}
		}
	}
	return result, nil
}

// DeleteOldEntityStates removes states older than the given threshold.
func (s *Store) DeleteOldEntityStates(_ context.Context, before string, adapter string) error {
	beforeTime, err := time.Parse(time.RFC3339, before)
	if err != nil {
		return fmt.Errorf("invalid before time: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for scopeID, states := range s.entityStates {
		var filtered []vector.EntityState
		for _, st := range states {
			if st.Timestamp.Before(beforeTime) && st.Meta["adapter"] == adapter {
				continue
			}
			filtered = append(filtered, st)
		}
		s.entityStates[scopeID] = filtered
	}
	return nil
}

// SaveInsight persists an insight.
func (s *Store) SaveInsight(_ context.Context, in insight.Insight) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.insights[in.ScopeID] = append(s.insights[in.ScopeID], in)
	return nil
}

// LoadInsights retrieves active insights for a scope.
func (s *Store) LoadInsights(_ context.Context, scopeID uuid.UUID) ([]insight.Insight, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var active []insight.Insight
	for _, in := range s.insights[scopeID] {
		if in.IsActive() {
			active = append(active, in)
		}
	}
	return active, nil
}

// LoadInsightByID retrieves a single insight by ID.
func (s *Store) LoadInsightByID(_ context.Context, id uuid.UUID) (insight.Insight, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, insights := range s.insights {
		for _, in := range insights {
			if in.ID == id {
				return in, nil
			}
		}
	}
	return insight.Insight{}, fmt.Errorf("insight not found: %s", id)
}

// DismissInsight marks an insight as dismissed.
func (s *Store) DismissInsight(_ context.Context, insightID, dismissedBy uuid.UUID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for scopeID, insights := range s.insights {
		for i := range insights {
			if insights[i].ID == insightID {
				insights[i].DismissedAt = &now
				s.insights[scopeID] = insights
				return nil
			}
		}
	}
	return fmt.Errorf("insight not found: %s", insightID)
}

// SaveFeedback records feedback.
func (s *Store) SaveFeedback(_ context.Context, insightID uuid.UUID, fb insight.Feedback) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.feedback[insightID] = fb
	return nil
}

// LoadFeedback retrieves feedback for an insight.
func (s *Store) LoadFeedback(_ context.Context, insightID uuid.UUID) (insight.Feedback, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	fb, ok := s.feedback[insightID]
	if !ok {
		return insight.Feedback{}, nil
	}
	return fb, nil
}

// CountEntityStates returns the count of entity states for an adapter.
func (s *Store) CountEntityStates(_ context.Context, adapter string) (int64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var count int64
	for _, states := range s.entityStates {
		for _, st := range states {
			if st.Meta["adapter"] == adapter {
				count++
			}
		}
	}
	return count, nil
}

// CountInsights returns the count of insights for a scope.
func (s *Store) CountInsights(_ context.Context, scopeID uuid.UUID) (int64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return int64(len(s.insights[scopeID])), nil
}
