package detect

import (
	"context"
	"testing"
	"time"

	"github.com/felixgeelhaar/chronos/internal/config"
	"github.com/felixgeelhaar/chronos/internal/domain"
	"github.com/google/uuid"
)

func stallCfg() *config.Config {
	return &config.Config{
		MaxSignalsPerRun: 100,
		StallMaxStdDev:   0.05,
		StallMinPoints:   4,
	}
}

func TestStall_FlatSeriesEmits(t *testing.T) {
	d := NewStall(stallCfg())
	scope := uuid.New()
	entity := uuid.New()
	now := time.Now()
	// Tightly clustered values — relative stddev well below 0.05.
	states := mkSeries(scope, entity, now, []float64{100, 100.1, 99.9, 100.05, 100.02})

	got := d.Detect(context.Background(), scope, states)
	if len(got) != 1 {
		t.Fatalf("got %d signals, want 1", len(got))
	}
	sig := got[0]
	if sig.Pattern != domain.PatternTypeStall {
		t.Errorf("Pattern = %s", sig.Pattern)
	}
	if sig.Strength <= 0 || sig.Strength > 1 {
		t.Errorf("Strength = %f, out of range", sig.Strength)
	}
	if sig.Metrics["normalised_stddev"] >= stallCfg().StallMaxStdDev {
		t.Errorf("normalised_stddev = %f, want < threshold", sig.Metrics["normalised_stddev"])
	}
	if err := sig.Validate(); err != nil {
		t.Errorf("invalid signal: %v", err)
	}
}

func TestStall_NoisySeriesNoSignal(t *testing.T) {
	d := NewStall(stallCfg())
	scope := uuid.New()
	entity := uuid.New()
	now := time.Now()
	states := mkSeries(scope, entity, now, []float64{1, 2, 3, 4, 5})
	got := d.Detect(context.Background(), scope, states)
	if len(got) != 0 {
		t.Errorf("got %d signals on rising series, want 0", len(got))
	}
}

func TestStall_RespectsMinPoints(t *testing.T) {
	d := NewStall(stallCfg())
	scope := uuid.New()
	entity := uuid.New()
	now := time.Now()
	states := mkSeries(scope, entity, now, []float64{100, 100.01, 100.02}) // below min
	got := d.Detect(context.Background(), scope, states)
	if len(got) != 0 {
		t.Errorf("got %d signals below min_points, want 0", len(got))
	}
}

func TestStall_AllZeroOutcomesSkipped(t *testing.T) {
	// All zeros means we cannot normalise; the detector must not panic
	// or emit a NaN-laden signal.
	d := NewStall(stallCfg())
	scope := uuid.New()
	entity := uuid.New()
	now := time.Now()
	states := mkSeries(scope, entity, now, []float64{0, 0, 0, 0, 0})
	got := d.Detect(context.Background(), scope, states)
	if len(got) != 0 {
		t.Errorf("got %d signals on all-zero outcomes, want 0", len(got))
	}
}
