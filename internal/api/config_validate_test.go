package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
)

func findDetector(reports []DetectorReport, name string) *DetectorReport {
	for i := range reports {
		if reports[i].Name == name {
			return &reports[i]
		}
	}
	return nil
}

// TestConfigValidate_DefaultsAllEnabled is the happy-path: an empty
// override set yields all ten detectors enabled. If a future config
// change accidentally disables a detector at defaults, this test
// fails loudly.
func TestConfigValidate_DefaultsAllEnabled(t *testing.T) {
	ts, _ := setupServer(t)
	defer ts.Close()
	body, _ := json.Marshal(ConfigValidateRequest{})
	resp, err := http.Post(ts.URL+"/v1/config/validate", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("POST: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	var got ConfigValidateResponse
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	wantDetectors := []string{
		"anomaly", "change_point", "correlation", "cross_scope_correlation",
		"drop", "outlier_cluster", "seasonality", "spike", "stall", "trend",
	}
	if len(got.Detectors) != len(wantDetectors) {
		t.Fatalf("got %d detectors, want %d", len(got.Detectors), len(wantDetectors))
	}
	for _, name := range wantDetectors {
		r := findDetector(got.Detectors, name)
		if r == nil {
			t.Errorf("detector %q missing", name)
			continue
		}
		if !r.Enabled {
			t.Errorf("detector %q disabled at defaults: %s", name, r.Reason)
		}
	}
}

// TestConfigValidate_DisablesCrossScopeOnHighMinPoints pins the
// motivating example from the issue: an operator who sets
// CROSS_SCOPE_MIN_POINTS=999999 (the current Thor workaround) should
// see cross_scope_correlation reported as disabled with the offending
// value cited.
func TestConfigValidate_DisablesCrossScopeOnHighMinPoints(t *testing.T) {
	ts, _ := setupServer(t)
	defer ts.Close()
	body, _ := json.Marshal(ConfigValidateRequest{
		Env: map[string]string{"CHRONOS_CROSS_SCOPE_MIN_POINTS": "0"},
	})
	resp, _ := http.Post(ts.URL+"/v1/config/validate", "application/json", bytes.NewReader(body))
	defer func() { _ = resp.Body.Close() }()
	var got ConfigValidateResponse
	_ = json.NewDecoder(resp.Body).Decode(&got)
	r := findDetector(got.Detectors, "cross_scope_correlation")
	if r == nil {
		t.Fatal("cross_scope_correlation report missing")
	}
	if r.Enabled {
		t.Errorf("cross_scope_correlation should be disabled when min_points < 2")
	}
	if r.Reason == "" {
		t.Errorf("missing reason for disabled detector")
	}
}

// TestConfigValidate_WarnsOnCrossScopeWithoutAnonymize proves the
// detector is reported enabled BUT a warning surfaces when anonymize
// is off — the kind of nuance an operator needs before flipping the
// detector on in a multi-tenant deployment.
func TestConfigValidate_WarnsOnCrossScopeWithoutAnonymize(t *testing.T) {
	ts, _ := setupServer(t)
	defer ts.Close()
	body, _ := json.Marshal(ConfigValidateRequest{
		Env: map[string]string{
			"CHRONOS_CROSS_SCOPE_MIN_POINTS": "5",
			"CHRONOS_ANONYMIZE_CROSS_SCOPE":  "false",
		},
	})
	resp, _ := http.Post(ts.URL+"/v1/config/validate", "application/json", bytes.NewReader(body))
	defer func() { _ = resp.Body.Close() }()
	var got ConfigValidateResponse
	_ = json.NewDecoder(resp.Body).Decode(&got)
	r := findDetector(got.Detectors, "cross_scope_correlation")
	if r == nil || !r.Enabled {
		t.Fatalf("cross_scope_correlation should be enabled: %+v", r)
	}
	if len(r.Warnings) == 0 {
		t.Errorf("expected anonymize warning when running multi-tenant without it, got %+v", r)
	}
}

// TestConfigValidate_TrendDisabledOnZeroSlope guards against the
// quieter footgun: a zero slope threshold silently disables trend
// detection. The endpoint must surface this clearly.
func TestConfigValidate_TrendDisabledOnZeroSlope(t *testing.T) {
	ts, _ := setupServer(t)
	defer ts.Close()
	body, _ := json.Marshal(ConfigValidateRequest{
		Env: map[string]string{"CHRONOS_TREND_MIN_SLOPE": "0"},
	})
	resp, _ := http.Post(ts.URL+"/v1/config/validate", "application/json", bytes.NewReader(body))
	defer func() { _ = resp.Body.Close() }()
	var got ConfigValidateResponse
	_ = json.NewDecoder(resp.Body).Decode(&got)
	r := findDetector(got.Detectors, "trend")
	if r == nil {
		t.Fatal("trend report missing")
	}
	if r.Enabled {
		t.Errorf("trend should be disabled at min_slope=0, got %+v", r)
	}
}

// TestConfigValidate_RestoresEnvAfterRequest pins the load-bearing
// safety property: the handler must NOT pollute the live process env
// with caller-supplied overrides. Otherwise validate calls would
// silently reconfigure the running server.
func TestConfigValidate_RestoresEnvAfterRequest(t *testing.T) {
	const key = "CHRONOS_TREND_MIN_POINTS"
	original, originalSet := lookupEnvSafe(t, key)
	t.Cleanup(func() {
		restoreEnv(t, key, original, originalSet)
	})

	ts, _ := setupServer(t)
	defer ts.Close()
	body, _ := json.Marshal(ConfigValidateRequest{
		Env: map[string]string{key: "777"},
	})
	resp, _ := http.Post(ts.URL+"/v1/config/validate", "application/json", bytes.NewReader(body))
	defer func() { _ = resp.Body.Close() }()

	after, afterSet := lookupEnvSafe(t, key)
	if afterSet != originalSet || after != original {
		t.Errorf("env leak: %s before=%q(set=%v) after=%q(set=%v)",
			key, original, originalSet, after, afterSet)
	}
}

// TestConfigValidate_RejectsBadJSON returns 400, not 500, for
// malformed bodies — input validation surfaces.
func TestConfigValidate_RejectsBadJSON(t *testing.T) {
	ts, _ := setupServer(t)
	defer ts.Close()
	resp, err := http.Post(ts.URL+"/v1/config/validate", "application/json", bytes.NewReader([]byte("{not json")))
	if err != nil {
		t.Fatalf("POST: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", resp.StatusCode)
	}
}

// TestConfigValidate_RejectsNonPOST guards the HTTP method contract.
func TestConfigValidate_RejectsNonPOST(t *testing.T) {
	ts, _ := setupServer(t)
	defer ts.Close()
	resp, err := http.Get(ts.URL + "/v1/config/validate")
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want 405", resp.StatusCode)
	}
}
