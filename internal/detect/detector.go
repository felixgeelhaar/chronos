// Package detect contains Chronos's pattern detectors and the engine that
// fans observations out across them.
//
// Each detector implements [Detector] and emits zero or more
// domain.Signals tagged with its PatternType. Detectors are pure
// functions of their input plus a configuration snapshot — they do no
// I/O. The [Engine] orchestrates a slice of detectors so adding a new
// pattern type is a one-file change.
package detect

import (
	"context"

	"github.com/felixgeelhaar/chronos"
	"github.com/felixgeelhaar/chronos/internal/domain"
	"github.com/google/uuid"
)

// Detector is the contract every pattern detector implements. Each
// detector is responsible for one PatternType.
//
// Detect operates on the states for a single scope. The engine groups
// the input by scope before calling Detect. Detectors must not mutate
// the input slice. Returned signals must already be Validated by the
// detector — the engine asserts but does not fix invalid output.
type Detector interface {
	// Pattern returns the PatternType this detector emits.
	Pattern() domain.PatternType

	// Detect produces signals for the given scope. The states slice is
	// guaranteed to be sorted by Timestamp ascending and to share the
	// supplied scopeID. May return an empty slice.
	Detect(ctx context.Context, scopeID uuid.UUID, states []chronos.EntityState) []domain.Signal
}
