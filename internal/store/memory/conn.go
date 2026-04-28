// Package memory provides in-memory implementations of the persistence
// ports defined in internal/ports. It is the canonical backend used by
// tests; production code should use the SQLite or Postgres backends.
package memory

import (
	"sync"

	"github.com/felixgeelhaar/chronos"
	"github.com/felixgeelhaar/chronos/internal/domain"
	"github.com/google/uuid"
)

// Conn bundles the in-memory repositories. The zero value is unusable;
// use [New].
type Conn struct {
	mu sync.RWMutex

	entityStates map[uuid.UUID][]storedState // keyed by ScopeID
	signals      []domain.Signal             // flat slice; queries scan

	EntityStates *EntityStateRepository
	Signals      *SignalRepository
}

// storedState couples an EntityState with the adapter that produced it.
// Adapter attribution stays out of the public chronos.EntityState shape.
type storedState struct {
	state   chronos.EntityState
	adapter string
}

// New creates a fresh in-memory Conn with all repositories wired up.
func New() *Conn {
	c := &Conn{
		entityStates: make(map[uuid.UUID][]storedState),
	}
	c.EntityStates = &EntityStateRepository{conn: c}
	c.Signals = &SignalRepository{conn: c}
	return c
}

// Close is a no-op; the in-memory store has no resources to release.
// It satisfies io.Closer so the same factory wiring works as for the
// SQL backends.
func (c *Conn) Close() error { return nil }
