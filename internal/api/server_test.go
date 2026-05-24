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

// TestIngestBatch_PersistsAllObservations is the happy path: a batch
// of N valid observations lands as N rows in the store. Without
// batch, the integrator would pay N round-trips; this test pins the
// round-trip-savings contract.
func TestIngestBatch_PersistsAllObservations(t *testing.T) {
	ts, mem := setupServer(t)
	defer ts.Close()

	scope := uuid.New()
	entity := uuid.New()
	obs := make([]IngestRequest, 0, 25)
	for i := 0; i < 25; i++ {
		obs = append(obs, IngestRequest{
			EntityID: entity, ScopeID: scope,
			Timestamp: time.Now().Add(time.Duration(i) * time.Second),
			Features:  []float64{float64(i)},
			Adapter:   "test",
		})
	}
	buf, _ := json.Marshal(IngestBatchRequest{Observations: obs})
	resp, err := http.Post(ts.URL+"/v1/ingest/batch", "application/json", bytes.NewReader(buf))
	if err != nil {
		t.Fatalf("POST: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	var body IngestBatchResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.Accepted != 25 {
		t.Errorf("Accepted = %d, want 25", body.Accepted)
	}
	stored, _ := mem.EntityStates.ListByScope(context.Background(), scope)
	if len(stored) != 25 {
		t.Errorf("stored = %d, want 25", len(stored))
	}
}

// TestIngestBatch_AllOrNothing pins that one bad row aborts the whole
// batch — partial writes would be hard to roll back and harder to
// reason about.
func TestIngestBatch_AllOrNothing(t *testing.T) {
	ts, mem := setupServer(t)
	defer ts.Close()
	scope := uuid.New()
	body := IngestBatchRequest{
		Observations: []IngestRequest{
			{EntityID: uuid.New(), ScopeID: scope, Features: []float64{1, 2, 3, 5}, Adapter: "test"},
			{EntityID: uuid.New(), ScopeID: scope, Features: []float64{}, Adapter: "test"}, // invalid: no features
			{EntityID: uuid.New(), ScopeID: scope, Features: []float64{4, 5, 6, 7}, Adapter: "test"},
		},
	}
	buf, _ := json.Marshal(body)
	resp, err := http.Post(ts.URL+"/v1/ingest/batch", "application/json", bytes.NewReader(buf))
	if err != nil {
		t.Fatalf("POST: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, want 400 (one invalid row aborts batch)", resp.StatusCode)
	}
	stored, _ := mem.EntityStates.ListByScope(context.Background(), scope)
	if len(stored) != 0 {
		t.Errorf("partial write: %d rows stored despite validation error", len(stored))
	}
}

// TestIngestBatch_RejectsEmpty fails fast on an empty array rather
// than silently 202'ing.
func TestIngestBatch_RejectsEmpty(t *testing.T) {
	ts, _ := setupServer(t)
	defer ts.Close()
	buf, _ := json.Marshal(IngestBatchRequest{})
	resp, err := http.Post(ts.URL+"/v1/ingest/batch", "application/json", bytes.NewReader(buf))
	if err != nil {
		t.Fatalf("POST: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", resp.StatusCode)
	}
}

// TestIngestBatch_RejectsOverLimit guards the memory + transaction-
// duration cap so a single misbehaving client can't pin a connection.
func TestIngestBatch_RejectsOverLimit(t *testing.T) {
	ts, _ := setupServer(t)
	defer ts.Close()
	scope := uuid.New()
	obs := make([]IngestRequest, MaxIngestBatchSize+1)
	for i := range obs {
		obs[i] = IngestRequest{EntityID: uuid.New(), ScopeID: scope, Features: []float64{1}, Adapter: "x"}
	}
	buf, _ := json.Marshal(IngestBatchRequest{Observations: obs})
	resp, err := http.Post(ts.URL+"/v1/ingest/batch", "application/json", bytes.NewReader(buf))
	if err != nil {
		t.Fatalf("POST: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, want 400 (over cap)", resp.StatusCode)
	}
}

// TestIngestBatch_DeferDetectionEchoed pins forward-compat contract
// for the defer_detection flag: it is accepted, persisted to the
// response, and the observations still land (chronos already
// separates ingest from detection so the flag is a no-op semantically).
func TestIngestBatch_DeferDetectionEchoed(t *testing.T) {
	ts, mem := setupServer(t)
	defer ts.Close()
	scope := uuid.New()
	body := IngestBatchRequest{
		DeferDetection: true,
		Observations: []IngestRequest{
			{EntityID: uuid.New(), ScopeID: scope, Features: []float64{1, 2, 3, 5}, Adapter: "test"},
		},
	}
	buf, _ := json.Marshal(body)
	resp, err := http.Post(ts.URL+"/v1/ingest/batch", "application/json", bytes.NewReader(buf))
	if err != nil {
		t.Fatalf("POST: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	var got IngestBatchResponse
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !got.DeferDetection {
		t.Errorf("DeferDetection not echoed: %+v", got)
	}
	stored, _ := mem.EntityStates.ListByScope(context.Background(), scope)
	if len(stored) != 1 {
		t.Errorf("stored = %d, want 1", len(stored))
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

// TestSignalDTO_ExplanationSurfaced pins detector-side context surfaces
// on wire shape so downstream narrators (Thor AI) explain WHY pattern
// fired without re-deriving the data.
func TestSignalDTO_ExplanationSurfaced(t *testing.T) {
	ts, mem := setupServer(t)
	defer ts.Close()
	scope := uuid.New()
	now := time.Now()

	sig := domain.Signal{
		ID:         uuid.New(),
		ScopeID:    scope,
		Series:     uuid.New(),
		Pattern:    domain.PatternTypeTrend,
		DetectedAt: now,
		Window:     domain.TimeWindow{Start: now.Add(-time.Hour), End: now},
		Strength:   0.8,
		Confidence: 0.8,
		Explanation: domain.Explanation{
			FeatureEvolution: []domain.FeatureSample{
				{At: now.Add(-2 * time.Hour), Value: 18.0},
				{At: now.Add(-time.Hour), Value: 22.0},
				{At: now, Value: 26.0},
			},
			ComparablePeers:    12,
			BaselineWindowDays: 90,
			ThresholdUsed:      2.5,
			DetectorVersion:    "trend-v2",
		},
	}
	if err := mem.Signals.Save(context.Background(), sig); err != nil {
		t.Fatalf("Save: %v", err)
	}

	resp, err := http.Get(ts.URL + "/v1/signals?scope_id=" + scope.String())
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	var body struct {
		Signals []SignalDTO `json:"signals"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(body.Signals) != 1 {
		t.Fatalf("expected 1 signal, got %d", len(body.Signals))
	}
	ex := body.Signals[0].Explanation
	if ex == nil {
		t.Fatal("Explanation missing on response — detector context must surface")
	}
	if ex.DetectorVersion != "trend-v2" {
		t.Errorf("detector_version = %q, want trend-v2", ex.DetectorVersion)
	}
	if ex.ComparablePeers != 12 {
		t.Errorf("comparable_peers = %d, want 12", ex.ComparablePeers)
	}
	if ex.BaselineWindowDays != 90 {
		t.Errorf("baseline_window_days = %d, want 90", ex.BaselineWindowDays)
	}
	if ex.ThresholdUsed != 2.5 {
		t.Errorf("threshold_used = %v, want 2.5", ex.ThresholdUsed)
	}
	if len(ex.FeatureEvolution) != 3 {
		t.Errorf("feature_evolution len = %d, want 3", len(ex.FeatureEvolution))
	}
}

// TestSignalDTO_NoExplanationOmitted pins the omitempty contract — a
// detector that did not surface an Explanation must yield a response
// without an explanation object, so consumers distinguish absence from
// zero.
func TestSignalDTO_NoExplanationOmitted(t *testing.T) {
	ts, mem := setupServer(t)
	defer ts.Close()
	scope := uuid.New()
	now := time.Now()

	if err := mem.Signals.Save(context.Background(), domain.Signal{
		ID: uuid.New(), ScopeID: scope, Series: uuid.New(),
		Pattern:    domain.PatternTypeRecurrence,
		DetectedAt: now,
		Window:     domain.TimeWindow{Start: now.Add(-time.Hour), End: now},
		Strength:   0.7, Confidence: 0.7,
	}); err != nil {
		t.Fatalf("Save: %v", err)
	}

	resp, err := http.Get(ts.URL + "/v1/signals?scope_id=" + scope.String())
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	var body struct {
		Signals []SignalDTO `json:"signals"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.Signals[0].Explanation != nil {
		t.Errorf("Explanation must be nil when detector did not surface one, got %+v", body.Signals[0].Explanation)
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
	lastScopes  []uuid.UUID
	lastPattern string
}

func (f *fakeBroadcaster) Subscribe(scopes []uuid.UUID, pattern string) (uuid.UUID, <-chan domain.Signal) {
	f.subCalls++
	f.lastScopes = append(f.lastScopes[:0], scopes...)
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
	if bc.subCalls != 1 || len(bc.lastScopes) != 1 || bc.lastScopes[0] != scope || bc.lastPattern != "recurrence" {
		t.Errorf("subscribe called wrong: calls=%d scopes=%v pattern=%q", bc.subCalls, bc.lastScopes, bc.lastPattern)
	}
}

// TestStream_ScopeInPassesAllowlist pins the load-bearing tenant
// safety property of #25: the server enforces the scope_in allowlist
// for the lifetime of the stream — clients cannot widen their own
// filter, and an outside-list signal never reaches them.
func TestStream_ScopeInPassesAllowlist(t *testing.T) {
	mem := memory.New()
	bc := &fakeBroadcaster{ch: make(chan domain.Signal, 1)}
	srv := NewServer(mem.EntityStates, mem.Signals, nil, nil).WithSSE(bc)
	mux := http.NewServeMux()
	srv.RegisterRoutes(mux)
	ts := httptest.NewServer(mux)
	defer ts.Close()

	scopeA, scopeB := uuid.New(), uuid.New()
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet,
		ts.URL+"/v1/signals/stream?scope_in="+scopeA.String()+","+scopeB.String(), nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	// Pump a few bytes so the handler actually invokes Subscribe.
	buf := make([]byte, 64)
	_, _ = resp.Body.Read(buf)

	if len(bc.lastScopes) != 2 {
		t.Fatalf("expected 2-scope allowlist, got %v", bc.lastScopes)
	}
	got := map[uuid.UUID]bool{bc.lastScopes[0]: true, bc.lastScopes[1]: true}
	if !got[scopeA] || !got[scopeB] {
		t.Errorf("allowlist missing one of {%s, %s}: got %v", scopeA, scopeB, bc.lastScopes)
	}
}

// TestStream_RejectsBothScopeIDAndScopeIn pins the input-validation
// contract: callers must pick one filter shape, not both. Accepting
// both silently would invite ambiguity about which one wins.
func TestStream_RejectsBothScopeIDAndScopeIn(t *testing.T) {
	mem := memory.New()
	srv := NewServer(mem.EntityStates, mem.Signals, nil, nil).
		WithSSE(&fakeBroadcaster{ch: make(chan domain.Signal, 1)})
	mux := http.NewServeMux()
	srv.RegisterRoutes(mux)
	ts := httptest.NewServer(mux)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/v1/signals/stream?scope_id=" + uuid.New().String() + "&scope_in=" + uuid.New().String())
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", resp.StatusCode)
	}
}

// TestStream_ScopeIn_BadUUID rejects malformed entries rather than
// silently dropping them — otherwise the allowlist would shrink
// without the caller noticing.
func TestStream_ScopeIn_BadUUID(t *testing.T) {
	mem := memory.New()
	srv := NewServer(mem.EntityStates, mem.Signals, nil, nil).
		WithSSE(&fakeBroadcaster{ch: make(chan domain.Signal, 1)})
	mux := http.NewServeMux()
	srv.RegisterRoutes(mux)
	ts := httptest.NewServer(mux)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/v1/signals/stream?scope_in=not-a-uuid")
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", resp.StatusCode)
	}
}
