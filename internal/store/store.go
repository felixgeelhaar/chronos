// Package store wires the configured persistence backend into the engine.
//
// Each backend (memory, sqlite, postgres) lives in a subpackage and
// exposes a Conn type that bundles its repositories. This package's only
// job is to dispatch on the configured database type and return a
// uniform port-typed handle to the rest of the engine.
package store

import (
	"context"
	"fmt"

	"github.com/felixgeelhaar/chronos/internal/ports"
	"github.com/felixgeelhaar/chronos/internal/store/memory"
	"github.com/felixgeelhaar/chronos/internal/store/postgres"
	"github.com/felixgeelhaar/chronos/internal/store/sqlite"
)

// Conn is the engine's view of a persistence backend: two port-typed
// repositories plus a Close method.
type Conn struct {
	EntityStates ports.EntityStateRepository
	Signals      ports.SignalRepository

	close func() error
}

// Close releases backend resources. Repository pointers must not be
// used after Close returns.
func (c *Conn) Close() error {
	if c.close == nil {
		return nil
	}
	return c.close()
}

// Open returns a Conn for the configured backend. Supported types are
// "memory", "sqlite" / "sqlite3", and "postgres" / "postgresql".
//
// Open performs a connectivity check against external backends so
// configuration errors surface at startup rather than first use.
func Open(_ context.Context, dbType, connStr string) (*Conn, error) {
	switch dbType {
	case "memory":
		c := memory.New()
		return &Conn{EntityStates: c.EntityStates, Signals: c.Signals, close: c.Close}, nil
	case "sqlite", "sqlite3":
		c, err := sqlite.Open(connStr)
		if err != nil {
			return nil, err
		}
		return &Conn{EntityStates: c.EntityStates, Signals: c.Signals, close: c.Close}, nil
	case "postgres", "postgresql":
		c, err := postgres.Open(connStr)
		if err != nil {
			return nil, err
		}
		return &Conn{EntityStates: c.EntityStates, Signals: c.Signals, close: c.Close}, nil
	default:
		return nil, fmt.Errorf("store: unsupported database type %q (supported: memory, sqlite, postgres)", dbType)
	}
}

// SupportedTypes lists the backends this build understands.
func SupportedTypes() []string {
	return []string{"memory", "sqlite", "postgres"}
}
