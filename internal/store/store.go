// Package store defines the generic storage interface for Chronos.
// All persistence implementations (SQLite, PostgreSQL, MySQL, etc.) must satisfy this interface.
package store

import (
	"context"
	"fmt"

	"github.com/felixgeelhaar/chronos/pkg/insight"
	"github.com/felixgeelhaar/chronos/pkg/vector"
	"github.com/google/uuid"
)

// Store is the generic persistence interface. All database backends implement this.
type Store interface {
	// Entity states
	SaveEntityStates(ctx context.Context, adapter string, states []vector.EntityState) error
	LoadEntityStates(ctx context.Context, scopeID uuid.UUID) ([]vector.EntityState, error)
	LoadEntityStatesByEntity(ctx context.Context, entityID uuid.UUID) ([]vector.EntityState, error)
	DeleteOldEntityStates(ctx context.Context, before string, adapter string) error

	// Insights
	SaveInsight(ctx context.Context, in insight.Insight) error
	LoadInsights(ctx context.Context, scopeID uuid.UUID) ([]insight.Insight, error)
	LoadInsightByID(ctx context.Context, id uuid.UUID) (insight.Insight, error)
	DismissInsight(ctx context.Context, insightID, dismissedBy uuid.UUID) error

	// Feedback
	SaveFeedback(ctx context.Context, insightID uuid.UUID, fb insight.Feedback) error
	LoadFeedback(ctx context.Context, insightID uuid.UUID) (insight.Feedback, error)

	// Metrics
	CountEntityStates(ctx context.Context, adapter string) (int64, error)
	CountInsights(ctx context.Context, scopeID uuid.UUID) (int64, error)

	// Lifecycle
	Close() error
}

// Factory creates a Store from a connection string and type.
type Factory struct{}

// New creates a Store based on the provided configuration.
func (f *Factory) New(dbType, connStr string) (Store, error) {
	switch dbType {
	case "sqlite", "sqlite3":
		return newSQLiteStore(connStr)
	case "postgres", "postgresql":
		return newPostgresStore(connStr)
	case "memory":
		return newMemoryStore(), nil
	default:
		return nil, fmt.Errorf("unsupported database type: %s (supported: sqlite, postgres, memory)", dbType)
	}
}

// SupportedTypes returns the list of supported database backends.
func SupportedTypes() []string {
	return []string{"sqlite", "postgres", "memory"}
}
