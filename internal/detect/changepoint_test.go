package detect

import (
	"context"
	"testing"
	"time"

	"github.com/felixgeelhaar/chronos"
	"github.com/felixgeelhaar/chronos/internal/config"
	"github.com/felixgeelhaar/chronos/internal/domain"
	"github.com/google/uuid"
)

func cpCfg() *config.Config {
	return &config.Config{
		ChangePointMinShift:  1.5,
		ChangePointMinPoints: 8,
	}
}

func cpSeries(scope, series uuid.UUID, values []float64) []chronos.EntityState {
	out := make([]chronos.EntityState, len(values))
	t0 := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	for i, v := range values {
		out[i] = chronos.EntityState{
			ID:        uuid.New(),
			EntityID:  series,
			ScopeID:   scope,
			Timestamp: t0.Add(time.Duration(i) * time.Hour),
			Features:  []float64{v}, // last feature = outcome
		}
	}
	return out
}

func TestChangePoint_DetectsStepChange(t *testing.T) {
	t.Parallel()
	scope := uuid.New()
	series := uuid.New()
	// Mean shifts from ~10 to ~20 at index 6.
	values := []float64{10, 10.2, 9.8, 10.1, 10, 9.9, 20, 20.1, 19.9, 20, 20.2, 19.8}
	d := NewChangePoint(cpCfg())
	signals := d.Detect(context.Background(), scope, cpSeries(scope, series, values))
	if len(signals) != 1 {
		t.Fatalf("len = %d, want 1; got %+v", len(signals), signals)
	}
	s := signals[0]
	if s.Pattern != domain.PatternTypeChangePoint {
		t.Errorf("pattern = %q", s.Pattern)
	}
	if got := s.Metrics["split_index"]; got != 6 {
		t.Errorf("split_index = %v, want 6", got)
	}
	if s.Metrics["mean_before"] >= s.Metrics["mean_after"] {
		t.Errorf("mean_before %v should be < mean_after %v",
			s.Metrics["mean_before"], s.Metrics["mean_after"])
	}
	if got := s.Metrics["shift"]; got < 1.5 {
		t.Errorf("shift = %v, want ≥ 1.5", got)
	}
	if len(s.Evidence) != 2 {
		t.Errorf("evidence kinds = %d", len(s.Evidence))
	}
}

func TestChangePoint_NoSignalForFlatSeries(t *testing.T) {
	t.Parallel()
	scope := uuid.New()
	values := []float64{10, 10.1, 9.9, 10, 10.2, 9.8, 10, 10.1}
	d := NewChangePoint(cpCfg())
	signals := d.Detect(context.Background(), scope, cpSeries(scope, uuid.New(), values))
	if len(signals) != 0 {
		t.Errorf("len = %d, want 0", len(signals))
	}
}

func TestChangePoint_NoSignalBelowMinPoints(t *testing.T) {
	t.Parallel()
	scope := uuid.New()
	values := []float64{10, 10, 20, 20}
	d := NewChangePoint(cpCfg())
	signals := d.Detect(context.Background(), scope, cpSeries(scope, uuid.New(), values))
	if len(signals) != 0 {
		t.Errorf("len = %d, want 0 (below ChangePointMinPoints)", len(signals))
	}
}

func TestChangePoint_PicksLargestShift(t *testing.T) {
	t.Parallel()
	scope := uuid.New()
	// Two candidate splits — only the larger crosses threshold.
	values := []float64{10, 10.2, 11, 10.8, 12, 12.1, 12, 30, 29.5, 30.5, 29.8, 30.2}
	d := NewChangePoint(cpCfg())
	signals := d.Detect(context.Background(), scope, cpSeries(scope, uuid.New(), values))
	if len(signals) != 1 {
		t.Fatalf("len = %d", len(signals))
	}
	if got := signals[0].Metrics["split_index"]; got != 7 {
		t.Errorf("split_index = %v, want 7", got)
	}
}
