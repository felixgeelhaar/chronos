package detect

import (
	"context"
	"sort"

	"github.com/felixgeelhaar/chronos"
	"github.com/felixgeelhaar/chronos/internal/config"
	"github.com/felixgeelhaar/chronos/internal/domain"
	"github.com/google/uuid"
)

// Engine fans observations out to a set of detectors, deduplicates and
// sorts the resulting signals, and applies a global cap.
//
// Construction is explicit: callers register the detectors they want.
// The default set (see [DefaultDetectors]) includes Recurrence; new
// pattern detectors are added by appending to the slice.
type Engine struct {
	cfg       *config.Config
	detectors []Detector
}

// NewEngine builds an Engine. If detectors is empty, DefaultDetectors
// is used.
func NewEngine(cfg *config.Config, detectors ...Detector) *Engine {
	if len(detectors) == 0 {
		detectors = DefaultDetectors(cfg)
	}
	return &Engine{cfg: cfg, detectors: detectors}
}

// DefaultDetectors returns the standard detector set wired with cfg.
// The full vision vocabulary is implemented: Recurrence, Trend, Spike,
// Drop, Stall, Anomaly, Seasonality, Correlation. Callers wanting a
// custom subset construct an Engine directly with explicit detectors.
func DefaultDetectors(cfg *config.Config) []Detector {
	return []Detector{
		NewRecurrence(cfg),
		NewTrend(cfg),
		NewSpike(cfg),
		NewDrop(cfg),
		NewStall(cfg),
		NewAnomaly(cfg),
		NewSeasonality(cfg),
		NewCorrelation(cfg),
	}
}

// Detect groups states by scope and runs every detector against each
// group. The combined output is sorted by detected-at descending,
// confidence descending, and capped at cfg.MaxSignalsPerRun.
func (e *Engine) Detect(ctx context.Context, states []chronos.EntityState) []domain.Signal {
	if len(states) == 0 {
		return nil
	}
	byScope := make(map[uuid.UUID][]chronos.EntityState)
	for _, s := range states {
		byScope[s.ScopeID] = append(byScope[s.ScopeID], s)
	}

	var all []domain.Signal
	for scopeID, scoped := range byScope {
		sortByTimestampAsc(scoped)
		for _, d := range e.detectors {
			all = append(all, d.Detect(ctx, scopeID, scoped)...)
		}
	}

	sort.SliceStable(all, func(i, j int) bool {
		if !all[i].DetectedAt.Equal(all[j].DetectedAt) {
			return all[i].DetectedAt.After(all[j].DetectedAt)
		}
		return all[i].Confidence > all[j].Confidence
	})

	if e.cfg.MaxSignalsPerRun > 0 && len(all) > e.cfg.MaxSignalsPerRun {
		all = all[:e.cfg.MaxSignalsPerRun]
	}
	return all
}

func sortByTimestampAsc(states []chronos.EntityState) {
	sort.SliceStable(states, func(i, j int) bool {
		return states[i].Timestamp.Before(states[j].Timestamp)
	})
}
