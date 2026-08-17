package detect

import (
	"testing"
	"time"

	"github.com/felixgeelhaar/chronos"
	"github.com/google/uuid"
)

func TestExplainSeries_PopulatesEvolutionAndMetadata(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
	obs := []chronos.EntityState{
		{ID: uuid.New(), EntityID: uuid.New(), ScopeID: uuid.New(), Timestamp: now, Features: []float64{1, 10}},
		{ID: uuid.New(), EntityID: uuid.New(), ScopeID: uuid.New(), Timestamp: now.Add(time.Hour), Features: []float64{1, 12}},
	}
	got := explainSeries(obs, 4, 0.85, detectorVersionRecurrence)
	if got.IsZero() {
		t.Fatal("explanation is zero")
	}
	if got.DetectorVersion != detectorVersionRecurrence {
		t.Errorf("version = %q", got.DetectorVersion)
	}
	if got.ComparablePeers != 4 {
		t.Errorf("peers = %d", got.ComparablePeers)
	}
	if got.ThresholdUsed != 0.85 {
		t.Errorf("threshold = %v", got.ThresholdUsed)
	}
	if len(got.FeatureEvolution) != 2 || got.FeatureEvolution[1].Value != 12 {
		t.Errorf("evolution = %+v", got.FeatureEvolution)
	}
}
