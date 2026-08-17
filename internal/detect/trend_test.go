package detect

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/felixgeelhaar/chronos"
	"github.com/felixgeelhaar/chronos/internal/config"
	"github.com/felixgeelhaar/chronos/internal/domain"
	"github.com/google/uuid"
)

func trendCfg() *config.Config {
	return &config.Config{
		MaxSignalsPerRun: 100,
		TrendMinSlope:    0.05,
		TrendMinPoints:   4,
	}
}

func mkSeries(scope, entity uuid.UUID, start time.Time, ys []float64) []chronos.EntityState {
	out := make([]chronos.EntityState, len(ys))
	for i, y := range ys {
		out[i] = chronos.EntityState{
			ID: uuid.New(), EntityID: entity, ScopeID: scope,
			Timestamp: start.Add(time.Duration(i) * time.Hour),
			Features:  []float64{1, 2, 3, y},
		}
	}
	return out
}

func TestTrend_RisingSeriesEmits(t *testing.T) {
	d := NewTrend(trendCfg())
	scope := uuid.New()
	entity := uuid.New()
	now := time.Now()
	states := mkSeries(scope, entity, now, []float64{1.0, 2.0, 3.0, 4.0, 5.0, 6.0})

	got := d.Detect(context.Background(), scope, states)
	if len(got) != 1 {
		t.Fatalf("got %d signals, want 1", len(got))
	}
	sig := got[0]
	if sig.Pattern != domain.PatternTypeTrend {
		t.Errorf("Pattern = %s", sig.Pattern)
	}
	if sig.Series != entity {
		t.Errorf("Series = %v", sig.Series)
	}
	if sig.Metrics["slope"] <= 0 {
		t.Errorf("slope = %f, want > 0 (rising)", sig.Metrics["slope"])
	}
	if sig.Metrics["r2"] < 0.99 {
		t.Errorf("r2 = %f, want ~1 for clean linear data", sig.Metrics["r2"])
	}
	if err := sig.Validate(); err != nil {
		t.Errorf("emitted signal invalid: %v", err)
	}
	if sig.Explanation.DetectorVersion != detectorVersionTrend {
		t.Errorf("explanation version = %q, want %s", sig.Explanation.DetectorVersion, detectorVersionTrend)
	}
	if len(sig.Explanation.FeatureEvolution) != 6 {
		t.Errorf("feature_evolution len = %d, want 6", len(sig.Explanation.FeatureEvolution))
	}
}

func TestTrend_NoisySeriesNoSignal(t *testing.T) {
	d := NewTrend(trendCfg())
	scope := uuid.New()
	entity := uuid.New()
	now := time.Now()
	// Random-ish values with no consistent direction.
	states := mkSeries(scope, entity, now, []float64{5, 1, 5, 1, 5, 1})

	got := d.Detect(context.Background(), scope, states)
	if len(got) != 0 {
		t.Fatalf("got %d signals on noise, want 0", len(got))
	}
}

func TestTrend_FlatSeriesNoSignal(t *testing.T) {
	d := NewTrend(trendCfg())
	scope := uuid.New()
	entity := uuid.New()
	now := time.Now()
	states := mkSeries(scope, entity, now, []float64{5, 5, 5, 5, 5, 5})
	got := d.Detect(context.Background(), scope, states)
	if len(got) != 0 {
		t.Fatalf("got %d signals on flat series, want 0", len(got))
	}
}

func TestTrend_RespectsMinPoints(t *testing.T) {
	d := NewTrend(trendCfg())
	scope := uuid.New()
	entity := uuid.New()
	now := time.Now()
	states := mkSeries(scope, entity, now, []float64{1, 2, 3}) // below min
	got := d.Detect(context.Background(), scope, states)
	if len(got) != 0 {
		t.Errorf("got %d signals below min_points, want 0", len(got))
	}
}

func TestLinearRegression(t *testing.T) {
	xs := []float64{0, 1, 2, 3, 4}
	ys := []float64{1, 3, 5, 7, 9} // y = 2x + 1
	slope, intercept, r2 := linearRegression(xs, ys)
	if math.Abs(slope-2) > 1e-9 || math.Abs(intercept-1) > 1e-9 {
		t.Errorf("slope/intercept = %f/%f, want 2/1", slope, intercept)
	}
	if math.Abs(r2-1) > 1e-9 {
		t.Errorf("r2 = %f, want 1", r2)
	}
}
