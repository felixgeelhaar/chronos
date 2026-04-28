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

func anomalyCfg() *config.Config {
	return &config.Config{
		MaxSignalsPerRun:     100,
		AnomalyMaxSimilarity: 0.5,
		AnomalyMinPeers:      2,
	}
}

func TestAnomaly_IsolatedSubjectEmits(t *testing.T) {
	d := NewAnomaly(anomalyCfg())
	scope := uuid.New()
	now := time.Now()

	// Three peers cluster around (1,2,3,5); one outlier at (10,20,30,1).
	peers := []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}
	outlier := uuid.New()

	states := []chronos.EntityState{
		{ID: uuid.New(), EntityID: peers[0], ScopeID: scope, Timestamp: now, Features: []float64{1, 2, 3, 5}},
		{ID: uuid.New(), EntityID: peers[1], ScopeID: scope, Timestamp: now, Features: []float64{1.1, 2.1, 3.1, 5.5}},
		{ID: uuid.New(), EntityID: peers[2], ScopeID: scope, Timestamp: now, Features: []float64{0.9, 1.9, 2.9, 4.5}},
		{ID: uuid.New(), EntityID: outlier, ScopeID: scope, Timestamp: now, Features: []float64{-1, -2, -3, -5}},
	}

	got := d.Detect(context.Background(), scope, states)
	// At least the outlier should trigger.
	var outlierSig *domain.Signal
	for i := range got {
		if got[i].Series == outlier {
			outlierSig = &got[i]
		}
	}
	if outlierSig == nil {
		t.Fatalf("outlier did not produce an Anomaly signal; signals=%+v", got)
	}
	if outlierSig.Pattern != domain.PatternTypeAnomaly {
		t.Errorf("Pattern = %s", outlierSig.Pattern)
	}
	if outlierSig.Metrics["max_peer_similarity"] >= anomalyCfg().AnomalyMaxSimilarity {
		t.Errorf("max_peer_similarity = %f, want < threshold", outlierSig.Metrics["max_peer_similarity"])
	}
	if err := outlierSig.Validate(); err != nil {
		t.Errorf("invalid signal: %v", err)
	}
}

func TestAnomaly_ClusteredEntitiesNoSignal(t *testing.T) {
	d := NewAnomaly(anomalyCfg())
	scope := uuid.New()
	now := time.Now()
	states := []chronos.EntityState{
		{ID: uuid.New(), EntityID: uuid.New(), ScopeID: scope, Timestamp: now, Features: []float64{1, 2, 3, 5}},
		{ID: uuid.New(), EntityID: uuid.New(), ScopeID: scope, Timestamp: now, Features: []float64{1.1, 2.1, 3.1, 5.5}},
		{ID: uuid.New(), EntityID: uuid.New(), ScopeID: scope, Timestamp: now, Features: []float64{0.9, 1.9, 2.9, 4.5}},
	}
	got := d.Detect(context.Background(), scope, states)
	if len(got) != 0 {
		t.Errorf("got %d signals on a clustered scope, want 0", len(got))
	}
}

func TestAnomaly_RequiresMinPeers(t *testing.T) {
	d := NewAnomaly(anomalyCfg())
	scope := uuid.New()
	now := time.Now()
	// Only one peer in scope — fewer than AnomalyMinPeers (2 needed).
	states := []chronos.EntityState{
		{ID: uuid.New(), EntityID: uuid.New(), ScopeID: scope, Timestamp: now, Features: []float64{1, 2, 3, 5}},
		{ID: uuid.New(), EntityID: uuid.New(), ScopeID: scope, Timestamp: now, Features: []float64{-100, -100, -100, -100}},
	}
	got := d.Detect(context.Background(), scope, states)
	if len(got) != 0 {
		t.Errorf("got %d signals below min_peers, want 0", len(got))
	}
}
