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

// TestEngine_FanOutAcrossDetectors exercises the engine with the full
// default detector set against a synthetic dataset crafted to trigger
// at least three orthogonal pattern types in a single Detect call.
//
// The point is not to validate detector internals (that's done in each
// detector's own _test.go file) but to confirm the engine wires the
// detectors together correctly and the resulting signals carry the
// expected PatternType tags.
func TestEngine_FanOutAcrossDetectors(t *testing.T) {
	cfg := &config.Config{
		MaxSignalsPerRun: 100,

		// Recurrence: tight similarity to keep peer matches focused.
		SimilarityThreshold: 0.95,
		MinSampleSize:       2,

		// Trend: easy threshold for the rising series.
		TrendMinSlope:  0.05,
		TrendMinPoints: 4,

		// Spike/Drop: rolling baseline of 4 points, z>=2.5.
		SpikeZScore: 2.5,
		DropZScore:  2.5,
		SpikeWindow: 4,

		// Stall: very tight stddev window so a flat-ish series qualifies.
		StallMaxStdDev: 0.01,
		StallMinPoints: 4,

		// Anomaly: leave it disabled (max sim 1.0 = anything counts as
		// "not isolated") to keep this test focused.
		AnomalyMaxSimilarity: 1.0,
		AnomalyMinPeers:      2,

		// Seasonality: high bar, more points than we provide here.
		SeasonalityMinAutocorr: 0.99,
		SeasonalityMinPoints:   100,
		SeasonalityMinPeriod:   2,

		// Correlation: high bar so it doesn't fire incidentally.
		CorrelationMin:       0.99,
		CorrelationMinPoints: 100,
	}
	e := NewEngine(cfg)

	scope := uuid.New()
	rising := uuid.New()
	flat := uuid.New()
	spiker := uuid.New()
	now := time.Now()

	var states []chronos.EntityState

	// Rising series — should trigger Trend.
	for i := 0; i < 8; i++ {
		states = append(states, chronos.EntityState{
			ID: uuid.New(), EntityID: rising, ScopeID: scope,
			Timestamp: now.Add(time.Duration(i) * time.Hour),
			Features:  []float64{1, 2, 3, float64(i + 1)},
		})
	}

	// Flat series — should trigger Stall (relative stddev tiny).
	for i := 0; i < 6; i++ {
		states = append(states, chronos.EntityState{
			ID: uuid.New(), EntityID: flat, ScopeID: scope,
			Timestamp: now.Add(time.Duration(i) * time.Hour),
			Features:  []float64{1, 2, 3, 100.0 + float64(i)*0.0001},
		})
	}

	// Spiker series — flat baseline followed by an outlier; should
	// trigger Spike.
	for i := 0; i < 5; i++ {
		states = append(states, chronos.EntityState{
			ID: uuid.New(), EntityID: spiker, ScopeID: scope,
			Timestamp: now.Add(time.Duration(i) * time.Hour),
			Features:  []float64{1, 2, 3, 10.0 + float64(i)*0.01},
		})
	}
	states = append(states, chronos.EntityState{
		ID: uuid.New(), EntityID: spiker, ScopeID: scope,
		Timestamp: now.Add(5 * time.Hour),
		Features:  []float64{1, 2, 3, 50.0},
	})

	signals := e.Detect(context.Background(), states)
	if len(signals) == 0 {
		t.Fatal("no signals emitted")
	}

	patterns := make(map[domain.PatternType]bool)
	bySeries := make(map[domain.PatternType][]uuid.UUID)
	for _, s := range signals {
		patterns[s.Pattern] = true
		bySeries[s.Pattern] = append(bySeries[s.Pattern], s.Series)
	}

	if !patterns[domain.PatternTypeTrend] {
		t.Errorf("Trend not emitted (signals: %v)", patterns)
	} else {
		// Trend should be on the rising series specifically.
		seenRising := false
		for _, id := range bySeries[domain.PatternTypeTrend] {
			if id == rising {
				seenRising = true
			}
		}
		if !seenRising {
			t.Errorf("Trend did not fire on rising series; saw %v", bySeries[domain.PatternTypeTrend])
		}
	}
	if !patterns[domain.PatternTypeStall] {
		t.Errorf("Stall not emitted")
	}
	if !patterns[domain.PatternTypeSpike] {
		t.Errorf("Spike not emitted")
	}
}

// TestEngine_RespectsMaxSignalsCap ensures the global cap is applied
// across detectors, not just within one detector.
func TestEngine_RespectsMaxSignalsCap(t *testing.T) {
	cfg := &config.Config{
		MaxSignalsPerRun: 1,

		SimilarityThreshold: 0.5,
		MinSampleSize:       2,

		TrendMinSlope:  0.05,
		TrendMinPoints: 4,

		// Disable detectors that need lots of points or peers to keep
		// the test focused on the cap.
		SpikeWindow:            10,
		StallMinPoints:         100,
		AnomalyMaxSimilarity:   -2,
		AnomalyMinPeers:        100,
		SeasonalityMinPoints:   100,
		SeasonalityMinAutocorr: 2,
		CorrelationMinPoints:   100,
		CorrelationMin:         2,
	}
	e := NewEngine(cfg)

	scope := uuid.New()
	now := time.Now()
	var states []chronos.EntityState
	// Two rising series — Trend should match both, Recurrence may also
	// fire. We expect MaxSignalsPerRun=1 to cap the output to one.
	for _, ent := range []uuid.UUID{uuid.New(), uuid.New()} {
		for i := 0; i < 8; i++ {
			states = append(states, chronos.EntityState{
				ID: uuid.New(), EntityID: ent, ScopeID: scope,
				Timestamp: now.Add(time.Duration(i) * time.Hour),
				Features:  []float64{1, 2, 3, float64(i + 1)},
			})
		}
	}

	got := e.Detect(context.Background(), states)
	if len(got) != 1 {
		t.Errorf("MaxSignalsPerRun=1 returned %d signals", len(got))
	}
}
