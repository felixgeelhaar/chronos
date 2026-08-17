package api

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/felixgeelhaar/chronos/internal/domain"
	"github.com/google/uuid"
)

// TestFederationExport_OptInRequired pins the load-bearing safety
// gate: without CHRONOS_FEDERATION_ENABLED=true the endpoint refuses
// to serve, even when signals exist. A forgotten env flag must NOT
// silently leak aggregate stats.
func TestFederationExport_OptInRequired(t *testing.T) {
	ts, _ := setupServer(t)
	defer ts.Close()
	resp, err := http.Get(ts.URL + "/v1/federation/export")
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusNotImplemented {
		t.Errorf("status = %d, want 501 (disabled)", resp.StatusCode)
	}
}

// TestFederationExport_NoTenantIdentifiersInResponse is the contract:
// nothing in the payload reveals which scope owned which signal. The
// test seeds signals across two scopes and inspects the marshalled
// JSON for either scope id — they MUST NOT appear.
func TestFederationExport_NoTenantIdentifiersInResponse(t *testing.T) {
	t.Setenv("CHRONOS_FEDERATION_ENABLED", "true")
	ts, mem := setupServer(t)
	defer ts.Close()

	scopeA, scopeB := uuid.New(), uuid.New()
	now := time.Now()
	for _, sc := range []uuid.UUID{scopeA, scopeB} {
		sig := domain.Signal{
			ID: uuid.New(), ScopeID: sc, Series: uuid.New(),
			Pattern: domain.PatternTypeTrend, DetectedAt: now,
			Window:   domain.TimeWindow{Start: now.Add(-time.Hour), End: now},
			Strength: 0.8, Confidence: 0.8,
		}
		if err := mem.Signals.Save(context.Background(), sig); err != nil {
			t.Fatalf("Save: %v", err)
		}
	}

	resp, err := http.Get(ts.URL + "/v1/federation/export")
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	body, _ := json.Marshal(decodeFederationResponse(t, resp))
	for _, sc := range []uuid.UUID{scopeA, scopeB} {
		if containsString(body, sc.String()) {
			t.Errorf("federation export leaks scope id %s", sc)
		}
	}
}

// TestAggregateFederationStats_GroupsByPattern pins the math: two
// trend signals at strength 0.8 → bucket count=2, avg=0.8, min/max
// match. Cross-pattern signals stay in distinct rows.
func TestAggregateFederationStats_GroupsByPattern(t *testing.T) {
	now := time.Now()
	sigs := []domain.Signal{
		mkFedSig(domain.PatternTypeTrend, 0.8, 0.8, now),
		mkFedSig(domain.PatternTypeTrend, 0.6, 0.7, now),
		mkFedSig(domain.PatternTypeRecurrence, 0.9, 0.95, now),
	}
	stats := AggregateFederationStats(sigs)
	if len(stats) != 2 {
		t.Fatalf("buckets = %d, want 2", len(stats))
	}
	// Stable ordering: recurrence < trend alphabetically.
	if stats[0].Pattern != string(domain.PatternTypeRecurrence) {
		t.Errorf("stats[0] = %s, want recurrence (alpha-sorted)", stats[0].Pattern)
	}
	for _, row := range stats {
		if row.Pattern == string(domain.PatternTypeTrend) {
			if row.Count != 2 {
				t.Errorf("trend count = %d, want 2", row.Count)
			}
			if !approxEqualFed(row.AvgStrength, 0.7, 1e-9) {
				t.Errorf("trend avg strength = %v, want 0.7", row.AvgStrength)
			}
			if row.MinStrength != 0.6 || row.MaxStrength != 0.8 {
				t.Errorf("trend strength bounds = %v/%v, want 0.6/0.8", row.MinStrength, row.MaxStrength)
			}
		}
	}
}

// TestAggregateFederationStats_CountsConfidenceClasses pins the
// per-class tally so a federated mirror can tell "10 strong + 200
// tentative" apart from "200 strong + 10 tentative" at a glance.
func TestAggregateFederationStats_CountsConfidenceClasses(t *testing.T) {
	now := time.Now()
	sigs := []domain.Signal{
		mkFedSigClass(domain.PatternTypeTrend, domain.ConfidenceClassStrong, now),
		mkFedSigClass(domain.PatternTypeTrend, domain.ConfidenceClassStrong, now),
		mkFedSigClass(domain.PatternTypeTrend, domain.ConfidenceClassEstablished, now),
		mkFedSigClass(domain.PatternTypeTrend, domain.ConfidenceClassTentative, now),
	}
	stats := AggregateFederationStats(sigs)
	if len(stats) != 1 {
		t.Fatalf("buckets = %d", len(stats))
	}
	r := stats[0]
	if r.StrongCount != 2 || r.EstablishedCount != 1 || r.TentativeCount != 1 {
		t.Errorf("class counts wrong: %+v", r)
	}
}

func mkFedSig(p domain.PatternType, strength, conf float64, now time.Time) domain.Signal {
	return domain.Signal{
		ID: uuid.New(), ScopeID: uuid.New(), Series: uuid.New(),
		Pattern: p, DetectedAt: now,
		Window:   domain.TimeWindow{Start: now.Add(-time.Hour), End: now},
		Strength: strength, Confidence: conf,
	}
}

func mkFedSigClass(p domain.PatternType, cls domain.ConfidenceClass, now time.Time) domain.Signal {
	s := mkFedSig(p, 0.5, 0.5, now)
	s.ConfidenceClass = cls
	return s
}

func decodeFederationResponse(t *testing.T, resp *http.Response) FederationExportResponse {
	t.Helper()
	var out FederationExportResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	return out
}

func containsString(haystack []byte, needle string) bool {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if string(haystack[i:i+len(needle)]) == needle {
			return true
		}
	}
	return false
}

func approxEqualFed(a, b, eps float64) bool {
	d := a - b
	if d < 0 {
		d = -d
	}
	return d <= eps
}
