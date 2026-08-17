package detect

import (
	"context"
	"testing"
	"time"

	"github.com/felixgeelhaar/chronos/internal/config"
	"github.com/felixgeelhaar/chronos/internal/domain"
	"github.com/google/uuid"
)

func spikeCfg() *config.Config {
	return &config.Config{
		MaxSignalsPerRun: 100,
		SpikeZScore:      2.5,
		DropZScore:       2.5,
		SpikeWindow:      5,
	}
}

func TestSpike_PositiveDeviationEmits(t *testing.T) {
	d := NewSpike(spikeCfg())
	scope := uuid.New()
	entity := uuid.New()
	now := time.Now()
	// Five baseline points around 10, then a 20 outlier.
	ys := []float64{10, 10.1, 9.9, 10.05, 9.95, 20}
	states := mkSeries(scope, entity, now, ys)

	got := d.Detect(context.Background(), scope, states)
	if len(got) != 1 {
		t.Fatalf("got %d spikes, want 1", len(got))
	}
	sig := got[0]
	if sig.Pattern != domain.PatternTypeSpike {
		t.Errorf("Pattern = %s", sig.Pattern)
	}
	if sig.Metrics["z"] <= spikeCfg().SpikeZScore {
		t.Errorf("z = %f, want > threshold %f", sig.Metrics["z"], spikeCfg().SpikeZScore)
	}
	if sig.Strength <= 0 || sig.Strength > 1 {
		t.Errorf("Strength = %f, out of range", sig.Strength)
	}
	if err := sig.Validate(); err != nil {
		t.Errorf("invalid signal: %v", err)
	}
	if sig.Explanation.DetectorVersion != detectorVersionSpike {
		t.Errorf("explanation version = %q, want %s", sig.Explanation.DetectorVersion, detectorVersionSpike)
	}
	if len(sig.Explanation.FeatureEvolution) == 0 {
		t.Error("explanation feature_evolution is empty")
	}
}

func TestSpike_NoBaselineNoSignal(t *testing.T) {
	d := NewSpike(spikeCfg())
	scope := uuid.New()
	entity := uuid.New()
	now := time.Now()
	// Only three points — fewer than window+1.
	states := mkSeries(scope, entity, now, []float64{1, 2, 100})
	got := d.Detect(context.Background(), scope, states)
	if len(got) != 0 {
		t.Errorf("got %d signals, want 0 (insufficient baseline)", len(got))
	}
}

func TestSpike_NegativeDeviationNotASpike(t *testing.T) {
	// A drop should NOT trigger Spike.
	d := NewSpike(spikeCfg())
	scope := uuid.New()
	entity := uuid.New()
	now := time.Now()
	states := mkSeries(scope, entity, now, []float64{10, 10, 10, 10, 10, -10})
	got := d.Detect(context.Background(), scope, states)
	if len(got) != 0 {
		t.Errorf("Spike fired on a drop: %d signals", len(got))
	}
}

func TestDrop_NegativeDeviationEmits(t *testing.T) {
	d := NewDrop(spikeCfg())
	scope := uuid.New()
	entity := uuid.New()
	now := time.Now()
	states := mkSeries(scope, entity, now, []float64{100, 100.5, 99.5, 100, 100.2, 50})
	got := d.Detect(context.Background(), scope, states)
	if len(got) != 1 {
		t.Fatalf("got %d drops, want 1", len(got))
	}
	if got[0].Pattern != domain.PatternTypeDrop {
		t.Errorf("Pattern = %s", got[0].Pattern)
	}
	if got[0].Metrics["z"] >= 0 {
		t.Errorf("drop z should be negative, got %f", got[0].Metrics["z"])
	}
}

func TestSpike_ZeroVarianceSkipped(t *testing.T) {
	// Constant baseline → stddev=0 → cannot compute z; must not panic
	// and must not emit.
	d := NewSpike(spikeCfg())
	scope := uuid.New()
	entity := uuid.New()
	now := time.Now()
	states := mkSeries(scope, entity, now, []float64{5, 5, 5, 5, 5, 100})
	got := d.Detect(context.Background(), scope, states)
	if len(got) != 0 {
		t.Errorf("got %d signals on zero-variance baseline, want 0", len(got))
	}
}
