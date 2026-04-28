package detect

import (
	"context"
	"math"
	"math/rand/v2"
	"testing"
	"time"

	"github.com/felixgeelhaar/chronos"
	"github.com/felixgeelhaar/chronos/internal/config"
	"github.com/google/uuid"
)

// Benchmark notes:
//
// Sizes here are chosen to represent a meaningful production scope —
// 50 entities × 50 observations is a "team" or "fleet" level of data.
// Run with: go test -bench=. -benchmem ./internal/detect.
//
// Use these to spot regressions: a 10× slowdown on Recurrence at this
// size is the kind of thing that turns a 1-second compute into a
// 10-second compute on a real workload, and that is the boundary
// where users feel it.

const (
	benchEntities        = 50
	benchPointsPerSeries = 50
	benchSeed1           = 1
	benchSeed2           = 2
)

// generateSeries deterministically builds numEntities × pointsPerEntity
// observations under scope. Feature values are drawn from a seeded
// PRNG so every run sees the same distribution.
func generateSeries(seed1, seed2 uint64, scope uuid.UUID, numEntities, pointsPerEntity int) []chronos.EntityState {
	rng := rand.New(rand.NewPCG(seed1, seed2))
	base := time.Now()
	out := make([]chronos.EntityState, 0, numEntities*pointsPerEntity)
	for e := 0; e < numEntities; e++ {
		ent := uuid.New()
		// Each entity has a constant baseline plus a small per-point
		// random walk; this gives detectors a meaningful structure
		// rather than pure noise.
		baseline := rng.Float64() * 10
		for p := 0; p < pointsPerEntity; p++ {
			out = append(out, chronos.EntityState{
				ID:        uuid.New(),
				EntityID:  ent,
				ScopeID:   scope,
				Timestamp: base.Add(time.Duration(p) * time.Hour),
				Features: []float64{
					baseline + rng.NormFloat64(),
					baseline*0.5 + rng.NormFloat64(),
					baseline*1.5 + rng.NormFloat64(),
					baseline + float64(p)*0.1 + rng.NormFloat64(),
				},
			})
		}
	}
	return out
}

// generatePeriodicSeries is used by Seasonality benchmarks: outcome
// follows a sine wave at lag-4 so the detector has something concrete
// to find.
func generatePeriodicSeries(scope, ent uuid.UUID, n int) []chronos.EntityState {
	base := time.Now()
	out := make([]chronos.EntityState, n)
	for i := 0; i < n; i++ {
		out[i] = chronos.EntityState{
			ID:        uuid.New(),
			EntityID:  ent,
			ScopeID:   scope,
			Timestamp: base.Add(time.Duration(i) * time.Hour),
			Features:  []float64{1, 2, 3, math.Sin(float64(i) * math.Pi / 2)},
		}
	}
	return out
}

func benchCfg() *config.Config {
	return &config.Config{
		MaxSignalsPerRun:       10_000,
		SimilarityThreshold:    0.5,
		MinSampleSize:          2,
		TrendMinSlope:          0.01,
		TrendMinPoints:         4,
		SpikeZScore:            2.5,
		DropZScore:             2.5,
		SpikeWindow:            5,
		StallMaxStdDev:         0.2,
		StallMinPoints:         4,
		AnomalyMaxSimilarity:   0.5,
		AnomalyMinPeers:        2,
		SeasonalityMinAutocorr: 0.5,
		SeasonalityMinPoints:   12,
		SeasonalityMinPeriod:   2,
		CorrelationMin:         0.5,
		CorrelationMinPoints:   5,
	}
}

func BenchmarkRecurrence(b *testing.B) {
	cfg := benchCfg()
	d := NewRecurrence(cfg)
	scope := uuid.New()
	states := generateSeries(benchSeed1, benchSeed2, scope, benchEntities, benchPointsPerSeries)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = d.Detect(context.Background(), scope, states)
	}
}

func BenchmarkTrend(b *testing.B) {
	cfg := benchCfg()
	d := NewTrend(cfg)
	scope := uuid.New()
	states := generateSeries(benchSeed1, benchSeed2, scope, benchEntities, benchPointsPerSeries)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = d.Detect(context.Background(), scope, states)
	}
}

func BenchmarkSpike(b *testing.B) {
	cfg := benchCfg()
	d := NewSpike(cfg)
	scope := uuid.New()
	states := generateSeries(benchSeed1, benchSeed2, scope, benchEntities, benchPointsPerSeries)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = d.Detect(context.Background(), scope, states)
	}
}

func BenchmarkStall(b *testing.B) {
	cfg := benchCfg()
	d := NewStall(cfg)
	scope := uuid.New()
	states := generateSeries(benchSeed1, benchSeed2, scope, benchEntities, benchPointsPerSeries)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = d.Detect(context.Background(), scope, states)
	}
}

func BenchmarkAnomaly(b *testing.B) {
	cfg := benchCfg()
	d := NewAnomaly(cfg)
	scope := uuid.New()
	states := generateSeries(benchSeed1, benchSeed2, scope, benchEntities, benchPointsPerSeries)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = d.Detect(context.Background(), scope, states)
	}
}

func BenchmarkSeasonality(b *testing.B) {
	cfg := benchCfg()
	d := NewSeasonality(cfg)
	scope := uuid.New()
	// Seasonality wants long, periodic series.
	var states []chronos.EntityState
	for i := 0; i < benchEntities; i++ {
		states = append(states, generatePeriodicSeries(scope, uuid.New(), 64)...)
	}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = d.Detect(context.Background(), scope, states)
	}
}

func BenchmarkCorrelation(b *testing.B) {
	cfg := benchCfg()
	d := NewCorrelation(cfg)
	scope := uuid.New()
	states := generateSeries(benchSeed1, benchSeed2, scope, benchEntities, benchPointsPerSeries)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = d.Detect(context.Background(), scope, states)
	}
}

func BenchmarkEngine_FullDefaultSet(b *testing.B) {
	cfg := benchCfg()
	e := NewEngine(cfg)
	scope := uuid.New()
	states := generateSeries(benchSeed1, benchSeed2, scope, benchEntities, benchPointsPerSeries)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = e.Detect(context.Background(), states)
	}
}
