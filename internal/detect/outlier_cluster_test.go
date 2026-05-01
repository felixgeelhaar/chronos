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

func ocCfg() *config.Config {
	return &config.Config{
		SpikeWindow:              5,
		OutlierClusterMinSeries:  3,
		OutlierClusterZ:          2.5,
		OutlierClusterTimeWindow: 5 * time.Minute,
	}
}

func ocSeries(scope, series uuid.UUID, base float64, spikeIdx int, spike float64) []chronos.EntityState {
	t0 := time.Date(2026, 5, 1, 10, 0, 0, 0, time.UTC)
	out := make([]chronos.EntityState, 8)
	for i := range out {
		v := base
		if i == spikeIdx {
			v = spike
		}
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

func TestOutlierCluster_DetectsCohortAnomaly(t *testing.T) {
	t.Parallel()
	scope := uuid.New()
	s1, s2, s3 := uuid.New(), uuid.New(), uuid.New()
	// Three series all spike at the same index → one cluster.
	states := append(append(
		ocSeries(scope, s1, 10, 6, 100),
		ocSeries(scope, s2, 20, 6, 200)...),
		ocSeries(scope, s3, 5, 6, 80)...)
	d := NewOutlierCluster(ocCfg())
	signals := d.Detect(context.Background(), scope, states)
	if len(signals) != 1 {
		t.Fatalf("len = %d, want 1", len(signals))
	}
	s := signals[0]
	if s.Pattern != domain.PatternTypeOutlierCluster {
		t.Errorf("pattern = %q", s.Pattern)
	}
	if s.Series != uuid.Nil {
		t.Errorf("Series = %v, want nil (cohort-level)", s.Series)
	}
	if got := s.Metrics["member_count"]; got != 3 {
		t.Errorf("member_count = %v, want 3", got)
	}
	if len(s.Evidence) != 3 {
		t.Errorf("evidence rows = %d, want 3", len(s.Evidence))
	}
}

func TestOutlierCluster_NoSignalForSingleAnomaly(t *testing.T) {
	t.Parallel()
	scope := uuid.New()
	s1 := uuid.New()
	d := NewOutlierCluster(ocCfg())
	signals := d.Detect(context.Background(), scope, ocSeries(scope, s1, 10, 6, 100))
	if len(signals) != 0 {
		t.Errorf("len = %d, want 0", len(signals))
	}
}

func TestOutlierCluster_NoSignalForFlatSeries(t *testing.T) {
	t.Parallel()
	scope := uuid.New()
	s1, s2, s3 := uuid.New(), uuid.New(), uuid.New()
	states := append(append(
		ocSeries(scope, s1, 10, -1, 0),
		ocSeries(scope, s2, 20, -1, 0)...),
		ocSeries(scope, s3, 5, -1, 0)...)
	d := NewOutlierCluster(ocCfg())
	if signals := d.Detect(context.Background(), scope, states); len(signals) != 0 {
		t.Errorf("flat series produced %d signals", len(signals))
	}
}
