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

func cscCfg() *config.Config {
	return &config.Config{
		CrossScopeMin:       0.8,
		CrossScopeMinPoints: 5,
	}
}

func cscSeries(scope, series uuid.UUID, values []float64) []chronos.EntityState {
	t0 := time.Date(2026, 5, 1, 10, 0, 0, 0, time.UTC)
	out := make([]chronos.EntityState, len(values))
	for i, v := range values {
		out[i] = chronos.EntityState{
			ID:        uuid.New(),
			EntityID:  series,
			ScopeID:   scope,
			Timestamp: t0.Add(time.Duration(i) * time.Minute),
			Features:  []float64{v},
		}
	}
	return out
}

func TestCrossScopeCorrelation_DetectsTwoScopesMovingTogether(t *testing.T) {
	t.Parallel()
	sa, sb := uuid.New(), uuid.New()
	a := cscSeries(sa, uuid.New(), []float64{1, 2, 3, 4, 5, 6, 7, 8})
	b := cscSeries(sb, uuid.New(), []float64{2, 4, 6, 8, 10, 12, 14, 16})
	states := append(a, b...)
	d := NewCrossScopeCorrelation(cscCfg())
	signals := d.CrossDetect(context.Background(), states)
	if len(signals) != 1 {
		t.Fatalf("len = %d, want 1; got %+v", len(signals), signals)
	}
	s := signals[0]
	if s.Pattern != domain.PatternTypeCrossScopeCorrelation {
		t.Errorf("pattern = %q", s.Pattern)
	}
	if got := s.Metrics["abs_r"]; got < 0.99 {
		t.Errorf("abs_r = %v, want close to 1", got)
	}
}

func TestCrossScopeCorrelation_DropsWithinScopePairs(t *testing.T) {
	t.Parallel()
	sa := uuid.New()
	a := cscSeries(sa, uuid.New(), []float64{1, 2, 3, 4, 5, 6, 7, 8})
	b := cscSeries(sa, uuid.New(), []float64{2, 4, 6, 8, 10, 12, 14, 16})
	d := NewCrossScopeCorrelation(cscCfg())
	signals := d.CrossDetect(context.Background(), append(a, b...))
	if len(signals) != 0 {
		t.Errorf("len = %d, want 0 (same-scope pair handled by Correlation)", len(signals))
	}
}

func TestCrossScopeCorrelation_NoSignalBelowThreshold(t *testing.T) {
	t.Parallel()
	sa, sb := uuid.New(), uuid.New()
	a := cscSeries(sa, uuid.New(), []float64{1, 5, 2, 7, 3, 6, 4, 8})
	b := cscSeries(sb, uuid.New(), []float64{8, 1, 9, 2, 7, 3, 6, 4})
	d := NewCrossScopeCorrelation(cscCfg())
	signals := d.CrossDetect(context.Background(), append(a, b...))
	if len(signals) != 0 {
		t.Errorf("len = %d, want 0", len(signals))
	}
}
