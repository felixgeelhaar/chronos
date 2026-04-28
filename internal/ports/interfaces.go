// Package ports declares the outbound interfaces ("ports") the engine drives.
//
// Two aggregates: entity-state observations and signals. Per the cognitive-
// stack vision Chronos does not own reviewer feedback (Mnemos's territory)
// or dismissal/decision lifecycle (Nous's territory) — those interfaces are
// deliberately absent.
//
// Per-aggregate repositories follow the Interface Segregation Principle:
// each backend only implements the slice it supports, and consumers depend
// on the narrowest interface they need.
package ports

import (
	"context"
	"time"

	"github.com/felixgeelhaar/chronos"
	"github.com/felixgeelhaar/chronos/internal/domain"
	"github.com/google/uuid"
)

// EntityStateRepository persists and loads time-series observations. The
// vision describes Chronos's intake as "Ingest(ctx, TimeSeriesPoint)";
// this interface offers both a single-point Ingest (for streaming
// adapters) and a batch Save (for pull-based adapters that fetch many
// points in one call).
type EntityStateRepository interface {
	// Ingest persists a single observation. Implementations should treat
	// repeated calls with the same ID as idempotent updates rather than
	// errors. This is the streaming entry point named to match the
	// cognitive-stack vocabulary.
	Ingest(ctx context.Context, adapterName string, state chronos.EntityState) error

	// Save persists a batch of observations. Equivalent in semantics to
	// calling Ingest once per state but transactionally guarded so a
	// partial write never leaves the store inconsistent.
	Save(ctx context.Context, adapterName string, states []chronos.EntityState) error

	// ListByScope returns all states for the given scope, most recent
	// first. Detectors load history through this method.
	ListByScope(ctx context.Context, scopeID uuid.UUID) ([]chronos.EntityState, error)

	// ListByEntity returns all observations of a single entity, most
	// recent first.
	ListByEntity(ctx context.Context, entityID uuid.UUID) ([]chronos.EntityState, error)

	// DeleteOlderThan removes states observed before the cutoff for the
	// given adapter. Used for retention.
	DeleteOlderThan(ctx context.Context, cutoff time.Time, adapterName string) error

	// Count returns the number of states recorded by the named adapter.
	Count(ctx context.Context, adapterName string) (int64, error)

	// ListScopes returns the set of distinct ScopeIDs that have at
	// least one observation. Order is unspecified. Used by the
	// in-process detection scheduler to know which scopes to detect
	// over without requiring callers to maintain a parallel index.
	ListScopes(ctx context.Context) ([]uuid.UUID, error)
}

// SignalFilter is a structured query against the signals store. All
// fields are optional; an empty filter matches everything in the scope
// (ScopeID is required by repositories that use this filter).
type SignalFilter struct {
	// ScopeID restricts results to a single scope. Required by callers.
	ScopeID uuid.UUID

	// Series, when set, restricts results to signals about a single
	// entity.
	Series *uuid.UUID

	// Pattern, when set, restricts results to a single PatternType.
	Pattern *domain.PatternType

	// Since and Until, when set, restrict results to signals detected
	// within the half-open interval [Since, Until).
	Since *time.Time
	Until *time.Time

	// MinConfidence, when set, drops signals below the threshold.
	MinConfidence *float64

	// Limit caps the number of returned signals; 0 means no limit.
	Limit int
}

// Notifier is the outbound port for pushing newly-persisted signals to
// downstream consumers (webhooks, SSE clients, in-process listeners).
//
// Implementations are fire-and-forget: failures must NOT propagate to
// the persistence path, must NOT panic the caller, and must respect
// ctx cancellation. Delivery semantics are at-most-once per consumer;
// consumers de-duplicate by Signal.ID. Per the cognitive-stack vision,
// notification is a transport concern only — the decision of what a
// signal means lives in Nous, never here.
type Notifier interface {
	Notify(ctx context.Context, sig domain.Signal)
}

// SignalRepository persists and queries signals. There is no Dismiss or
// IsActive concept here — once a signal is detected and persisted it is
// immutable. Decisions about whether to act on it (or to suppress it
// in a UI) live in Nous.
type SignalRepository interface {
	// Save persists a single signal including its evidence. Idempotent
	// on Signal.ID.
	Save(ctx context.Context, sig domain.Signal) error

	// List returns signals matching filter, ordered detected-at
	// descending (then confidence descending) by convention.
	List(ctx context.Context, filter SignalFilter) ([]domain.Signal, error)

	// Get returns a single signal by ID. Returns
	// domain.ErrSignalNotFound when no row exists.
	Get(ctx context.Context, id uuid.UUID) (domain.Signal, error)

	// Count returns the number of signals matching filter.
	Count(ctx context.Context, filter SignalFilter) (int64, error)
}
