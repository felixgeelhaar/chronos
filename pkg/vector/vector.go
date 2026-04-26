// Package vector provides generic types for time-series feature vectors.
// All domain knowledge is externalised through feature labels and metadata.
package vector

import (
	"time"

	"github.com/google/uuid"
)

// EntityState represents a single observation of an entity at a point in time.
// Adapters map their domain-specific data into this generic structure.
type EntityState struct {
	ID        uuid.UUID       // Unique observation ID
	EntityID  uuid.UUID       // The entity (athlete, server, sensor, etc.)
	ScopeID   uuid.UUID       // The scope (coach, team, tenant, etc.)
	Timestamp time.Time       // When this state was observed
	Features  []float64       // Normalised numeric features
	Labels    []string        // Human-readable feature names (optional)
	Meta      map[string]string // Adapter-specific metadata (not used for similarity)
}

// FeatureConfig defines how features are named and weighted.
type FeatureConfig struct {
	Names    []string  // Feature names, length must match Features
	Weights  []float64 // Optional weights for weighted similarity
}

// Validate checks that feature dimensions are consistent.
func (es *EntityState) Validate() error {
	if es.EntityID == uuid.Nil {
		return ErrMissingEntityID
	}
	if len(es.Features) == 0 {
		return ErrNoFeatures
	}
	return nil
}

// Normalise applies z-score normalisation to a slice of features.
func Normalise(features []float64) []float64 {
	if len(features) == 0 {
		return nil
	}

	mean := 0.0
	for _, v := range features {
		mean += v
	}
	mean /= float64(len(features))

	variance := 0.0
	for _, v := range features {
		d := v - mean
		variance += d * d
	}
	stddev := sqrt(variance / float64(len(features)))

	out := make([]float64, len(features))
	if stddev == 0 {
		return out // all zeros if no variance
	}
	for i, v := range features {
		out[i] = (v - mean) / stddev
	}
	return out
}

func sqrt(x float64) float64 {
	if x == 0 {
		return 0
	}
	z := x
	for i := 0; i < 10; i++ {
		z = (z + x/z) / 2
	}
	return z
}
