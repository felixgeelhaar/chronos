// Package adapter defines the interface between Chronos and external data sources.
package adapter

import (
	"context"

	"github.com/felixgeelhaar/chronos/pkg/vector"
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

// Register adds an adapter to the registry.
func (r *Registry) Register(src Source) {
	r.sources[src.Name()] = src
}

// Get retrieves an adapter by name.
func (r *Registry) Get(name string) (Source, bool) {
	src, ok := r.sources[name]
	return src, ok
}

// List returns all registered adapter names.
func (r *Registry) List() []string {
	names := make([]string, 0, len(r.sources))
	for name := range r.sources {
		names = append(names, name)
	}
	return names
}
