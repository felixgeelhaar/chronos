// Package adapter defines the interface between Chronos and external data sources.
package adapter

import (
	"context"
	"sync"

	"github.com/felixgeelhaar/chronos/pkg/vector"
)

var (
	globalRegistry     *Registry
	globalRegistryMu sync.RWMutex
)

// Source is the interface that all data adapters must implement.
// It maps domain-specific data into Chronos' generic EntityState model.
type Source interface {
	// Name returns the adapter identifier (e.g., "ascend", "prometheus").
	Name() string

	// Fetch retrieves entity states from the external source.
	// cfg contains adapter-specific parameters (e.g., coach-id, connection string).
	Fetch(ctx context.Context, cfg map[string]string) ([]vector.EntityState, error)
}

// Registry holds all registered adapters.
type Registry struct {
	sources map[string]Source
}

// NewRegistry creates an empty adapter registry.
func NewRegistry() *Registry {
	return &Registry{sources: make(map[string]Source)}
}

// GetRegistry returns the global adapter registry.
// Adapters register themselves via init() to make this work.
func GetRegistry() *Registry {
	globalRegistryMu.Lock()
	defer globalRegistryMu.Unlock()
	if globalRegistry == nil {
		globalRegistry = NewRegistry()
	}
	return globalRegistry
}

// Register adds an adapter to the global registry.
func Register(src Source) {
	globalRegistryMu.Lock()
	defer globalRegistryMu.Unlock()
	if globalRegistry == nil {
		globalRegistry = NewRegistry()
	}
	globalRegistry.sources[src.Name()] = src
}

// Get retrieves an adapter by name from the global registry.
func Get(name string) (Source, bool) {
	globalRegistryMu.RLock()
	defer globalRegistryMu.RUnlock()
	if globalRegistry == nil {
		return nil, false
	}
	src, ok := globalRegistry.sources[name]
	return src, ok
}
