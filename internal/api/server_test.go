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

// TestListSignals_ScopeIn_Allowlist pins the multi-scope filter
// behaviour: a comma-separated scope_in returns only signals whose
// scope_id is in the allowlist, never signals from scopes outside it.
// Critical for tenant-safety: consumers with N entities owned by one
// user can fetch in one round-trip without trusting the client to
// post-filter.
func TestListSignals_ScopeIn_Allowlist(t *testing.T) {
	ts, mem := setupServer(t)
	defer ts.Close()

	scopeA := uuid.New()
	scopeB := uuid.New()
	scopeC := uuid.New() // not in the allowlist — must NOT appear in response
	now := time.Now()

	for _, sc := range []uuid.UUID{scopeA, scopeB, scopeC} {
		sig := domain.Signal{
			ID: uuid.New(), ScopeID: sc, Series: uuid.New(),
			Pattern:    domain.PatternTypeRecurrence,
			DetectedAt: now,
			Window:     domain.TimeWindow{Start: now.Add(-time.Hour), End: now},
			Strength:   0.9, Confidence: 0.9,
		}
		if err := mem.Signals.Save(context.Background(), sig); err != nil {
			t.Fatalf("Save: %v", err)
		}
	}

	url := ts.URL + "/v1/signals?scope_in=" + scopeA.String() + "," + scopeB.String()
	resp, err := http.Get(url)
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
		t.Errorf("count = %d, want 2 (only scopeA + scopeB allowed)", body.Count)
	}
	for _, s := range body.Signals {
		if s.ScopeID == scopeC {
			t.Errorf("response leaked signal from scope %s — outside allowlist", scopeC)
		}
		if s.ScopeID != scopeA && s.ScopeID != scopeB {
			t.Errorf("unexpected scope %s in response", s.ScopeID)
		}
	}
}

// TestListSignals_NeitherScopeIDNorScopeIn pins the input-validation
// contract: at least one of scope_id or scope_in must be provided.
// Fail-closed avoids a missing-filter footgun that would dump every
// tenant's signals to anyone who forgot the query param.
func TestListSignals_NeitherScopeIDNorScopeIn(t *testing.T) {
	ts, _ := setupServer(t)
	defer ts.Close()
	resp, err := http.Get(ts.URL + "/v1/signals")
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, want 400 when neither scope filter is set", resp.StatusCode)
	}
}

// TestListSignals_ScopeIn_BadUUID rejects malformed entries with a
// descriptive 400 rather than silently dropping them (which would
// degrade to a leaky filter).
func TestListSignals_ScopeIn_BadUUID(t *testing.T) {
	ts, _ := setupServer(t)
	defer ts.Close()
	resp, err := http.Get(ts.URL + "/v1/signals?scope_in=not-a-uuid")
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, want 400 for malformed scope_in entry", resp.StatusCode)
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

// fakeBroadcaster is a minimal SSEBroadcaster for handler-level tests.
// It returns a channel the test pre-loads with a signal so we can
// assert what handleStream writes onto the wire without standing up
// the real notify.SSE.
type fakeBroadcaster struct {
	ch          chan domain.Signal
	subCalls    int
	lastScope   uuid.UUID
	lastPattern string
}

func (f *fakeBroadcaster) Subscribe(scope uuid.UUID, pattern string) (uuid.UUID, <-chan domain.Signal) {
	f.subCalls++
	f.lastScope = scope
	f.lastPattern = pattern
	return uuid.New(), f.ch
}

func (f *fakeBroadcaster) Unsubscribe(uuid.UUID) {
	close(f.ch)
}

func TestStream_RejectsWithoutBroadcaster(t *testing.T) {
	ts, _ := setupServer(t) // setupServer does not attach SSE
	defer ts.Close()
	resp, err := http.Get(ts.URL + "/v1/signals/stream?scope_id=" + uuid.New().String())
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusNotImplemented {
		t.Errorf("status = %d, want 501", resp.StatusCode)
	}
}

func TestStream_RequiresScopeID(t *testing.T) {
	mem := memory.New()
	srv := NewServer(mem.EntityStates, mem.Signals, nil, nil).
		WithSSE(&fakeBroadcaster{ch: make(chan domain.Signal, 1)})
	mux := http.NewServeMux()
	srv.RegisterRoutes(mux)
	ts := httptest.NewServer(mux)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/v1/signals/stream")
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", resp.StatusCode)
	}
}

func TestStream_DeliversSignalAsSSEEvent(t *testing.T) {
	mem := memory.New()
	bc := &fakeBroadcaster{ch: make(chan domain.Signal, 1)}
	srv := NewServer(mem.EntityStates, mem.Signals, nil, nil).WithSSE(bc)
	mux := http.NewServeMux()
	srv.RegisterRoutes(mux)
	ts := httptest.NewServer(mux)
	defer ts.Close()

	scope := uuid.New()
	signalID := uuid.New()
	bc.ch <- domain.Signal{
		ID:         signalID,
		ScopeID:    scope,
		Series:     uuid.New(),
		Pattern:    domain.PatternTypeRecurrence,
		DetectedAt: time.Now(),
		Window:     domain.TimeWindow{Start: time.Now().Add(-time.Hour), End: time.Now()},
		Strength:   0.9,
		Confidence: 0.8,
	}

	// Use a context to bound the read; the handler holds the
	// connection open until we cancel.
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet,
		ts.URL+"/v1/signals/stream?scope_id="+scope.String()+"&pattern=recurrence", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	if got := resp.Header.Get("Content-Type"); got != "text/event-stream" {
		t.Errorf("Content-Type = %q, want text/event-stream", got)
	}

	// Read repeatedly until we capture both the heartbeat and the
	// first signal frame (the heartbeat may flush separately) or hit
	// the test deadline.
	var body bytes.Buffer
	buf := make([]byte, 1024)
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		n, _ := resp.Body.Read(buf)
		if n > 0 {
			body.Write(buf[:n])
		}
		if bytes.Contains(body.Bytes(), []byte(signalID.String())) {
			break
		}
	}
	if !bytes.Contains(body.Bytes(), []byte("event: signal")) {
		t.Errorf("missing event line:\n%s", body.String())
	}
	if !bytes.Contains(body.Bytes(), []byte(signalID.String())) {
		t.Errorf("missing signal id in payload:\n%s", body.String())
	}
	if bc.subCalls != 1 || bc.lastScope != scope || bc.lastPattern != "recurrence" {
		t.Errorf("subscribe called wrong: calls=%d scope=%v pattern=%q", bc.subCalls, bc.lastScope, bc.lastPattern)
	}
}
