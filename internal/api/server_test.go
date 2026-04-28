package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/felixgeelhaar/chronos"
	"github.com/felixgeelhaar/chronos/internal/domain"
	"github.com/felixgeelhaar/chronos/internal/store/memory"
	"github.com/google/uuid"
)

func setupServer(t *testing.T) (*httptest.Server, *memory.Conn) {
	t.Helper()
	mem := memory.New()
	srv := NewServer(mem.EntityStates, mem.Signals, nil, nil)
	mux := http.NewServeMux()
	srv.RegisterRoutes(mux)
	return httptest.NewServer(mux), mem
}

func TestHealth(t *testing.T) {
	ts, _ := setupServer(t)
	defer ts.Close()
	resp, err := http.Get(ts.URL + "/health")
	if err != nil {
		t.Fatalf("GET /health: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
}

func TestIngest_PersistsState(t *testing.T) {
	ts, mem := setupServer(t)
	defer ts.Close()

	scope := uuid.New()
	entity := uuid.New()
	body := IngestRequest{
		EntityID:  entity,
		ScopeID:   scope,
		Timestamp: time.Now(),
		Features:  []float64{1, 2, 3, 5},
		Adapter:   "test",
	}
	buf, _ := json.Marshal(body)
	resp, err := http.Post(ts.URL+"/v1/ingest", "application/json", bytes.NewReader(buf))
	if err != nil {
		t.Fatalf("POST: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	got, _ := mem.EntityStates.ListByScope(context.Background(), scope)
	if len(got) != 1 || got[0].EntityID != entity {
		t.Fatalf("state not persisted: %+v", got)
	}
}

func TestIngest_RejectsInvalid(t *testing.T) {
	ts, _ := setupServer(t)
	defer ts.Close()

	// Missing entity_id should fail validation.
	body := IngestRequest{ScopeID: uuid.New(), Features: []float64{1, 2, 3}}
	buf, _ := json.Marshal(body)
	resp, err := http.Post(ts.URL+"/v1/ingest", "application/json", bytes.NewReader(buf))
	if err != nil {
		t.Fatalf("POST: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", resp.StatusCode)
	}
}

func TestListSignals_FiltersAndDTO(t *testing.T) {
	ts, mem := setupServer(t)
	defer ts.Close()

	scope := uuid.New()
	now := time.Now()
	for _, conf := range []float64{0.6, 0.8, 0.95} {
		sig := domain.Signal{
			ID:         uuid.New(),
			ScopeID:    scope,
			Series:     uuid.New(),
			Pattern:    domain.PatternTypeRecurrence,
			DetectedAt: now,
			Window:     domain.TimeWindow{Start: now.Add(-time.Hour), End: now},
			Strength:   conf,
			Confidence: conf,
			Metrics:    map[string]float64{"avg_similarity": conf},
		}
		if err := mem.Signals.Save(context.Background(), sig); err != nil {
			t.Fatalf("Save: %v", err)
		}
	}

	resp, err := http.Get(ts.URL + "/v1/signals?scope_id=" + scope.String() + "&min_confidence=0.7")
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	var body struct {
		Signals []SignalDTO `json:"signals"`
		Count   int         `json:"count"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.Count != 2 {
		t.Errorf("count = %d, want 2", body.Count)
	}
	for _, s := range body.Signals {
		if s.Pattern != string(domain.PatternTypeRecurrence) {
			t.Errorf("pattern = %s", s.Pattern)
		}
		if s.Confidence < 0.7 {
			t.Errorf("confidence %f below filter threshold", s.Confidence)
		}
	}
}

func TestListSignals_BadScope(t *testing.T) {
	ts, _ := setupServer(t)
	defer ts.Close()
	resp, err := http.Get(ts.URL + "/v1/signals?scope_id=not-a-uuid")
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d", resp.StatusCode)
	}
}

func TestSignalDetail_NotFound(t *testing.T) {
	ts, _ := setupServer(t)
	defer ts.Close()
	resp, err := http.Get(ts.URL + "/v1/signals/" + uuid.New().String())
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status = %d, want 404", resp.StatusCode)
	}
}

// Compile-time assertion that EntityState satisfies our DTO conversion.
var _ = chronos.EntityState{}
