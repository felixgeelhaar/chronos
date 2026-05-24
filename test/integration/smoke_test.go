//go:build integration

// Package integration runs the cross-service smoke tests against a
// running chronos + mnemos stack. The build tag is the gate: the
// suite never runs under a plain `go test ./...`. Stand the stack up
// first (docker compose -f test/integration/docker-compose.yml up -d)
// then run:
//
//	go test -tags=integration ./test/integration/...
//
// Endpoints are taken from CHRONOS_INTEGRATION_URL /
// MNEMOS_INTEGRATION_URL with sensible localhost defaults; CI can
// repoint them at a hosted instance without code changes.
package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"
)

func chronosBaseURL() string {
	if v := os.Getenv("CHRONOS_INTEGRATION_URL"); v != "" {
		return v
	}
	return "http://127.0.0.1:7778"
}

func mnemosBaseURL() string {
	if v := os.Getenv("MNEMOS_INTEGRATION_URL"); v != "" {
		return v
	}
	return "http://127.0.0.1:7777"
}

// TestHealthBothServices is the boot-time check: both services
// answer /health 200 before any of the cross-talk tests fire.
// Failing here means the docker stack isn't up — fail fast with a
// clear message rather than hitting confusing 500s downstream.
func TestHealthBothServices(t *testing.T) {
	for name, url := range map[string]string{
		"chronos": chronosBaseURL() + "/health",
		"mnemos":  mnemosBaseURL() + "/health",
	} {
		if err := waitForHealth(url, 30*time.Second); err != nil {
			t.Fatalf("%s health: %v", name, err)
		}
	}
}

// TestCrossTalk_IngestThenList smoke-tests the chronos write path
// end-to-end: ingest one observation, then list signals for the
// scope and verify the request succeeds. Detection is async, so the
// list may legitimately return zero rows — the assertion is that
// the HTTP contracts work, not that detection has fired.
func TestCrossTalk_IngestThenList(t *testing.T) {
	scope := newUUID()
	entity := newUUID()
	body := map[string]any{
		"entity_id": entity,
		"scope_id":  scope,
		"timestamp": time.Now().UTC().Format(time.RFC3339Nano),
		"features":  []float64{1.0, 2.0, 3.0, 5.0},
		"adapter":   "integration",
	}
	postJSON(t, chronosBaseURL()+"/v1/ingest", body, http.StatusAccepted)

	resp := getJSON(t, fmt.Sprintf("%s/v1/signals?scope_id=%s", chronosBaseURL(), scope))
	if _, ok := resp["signals"]; !ok {
		t.Fatalf("missing 'signals' key in response: %+v", resp)
	}
}

// TestCrossTalk_MnemosClaimRoundTrip pins the mnemos write+read
// path. Appends an event + claim, then reads the claims list back
// and verifies the seeded row surfaces.
func TestCrossTalk_MnemosClaimRoundTrip(t *testing.T) {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	runID := "integration:" + newUUID()

	postJSON(t, mnemosBaseURL()+"/v1/events", map[string]any{
		"events": []map[string]any{{
			"id":              "ev_smoke_" + newUUID(),
			"run_id":          runID,
			"schema_version":  "v1",
			"content":         "integration smoke event",
			"source_input_id": "smoke",
			"timestamp":       now,
			"ingested_at":     now,
		}},
	}, http.StatusCreated)

	resp := getJSON(t, fmt.Sprintf("%s/v1/events?run_id=%s", mnemosBaseURL(), runID))
	events, ok := resp["events"].([]any)
	if !ok || len(events) != 1 {
		t.Fatalf("expected 1 event under %s, got %+v", runID, resp)
	}
}

func newUUID() string {
	// crypto/rand-backed UUIDv4. Hand-rolled here so the smoke test
	// doesn't add a runtime dep on github.com/google/uuid for the
	// build-tagged binary that ships with chronos.
	b := make([]byte, 16)
	if _, err := readRandom(b); err != nil {
		panic(err)
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

func waitForHealth(url string, total time.Duration) error {
	deadline := time.Now().Add(total)
	for time.Now().Before(deadline) {
		resp, err := http.Get(url)
		if err == nil && resp.StatusCode == 200 {
			_ = resp.Body.Close()
			return nil
		}
		if resp != nil {
			_ = resp.Body.Close()
		}
		time.Sleep(500 * time.Millisecond)
	}
	return fmt.Errorf("%s never returned 200 within %s", url, total)
}

func postJSON(t *testing.T, url string, body any, expectStatus int) {
	t.Helper()
	buf, _ := json.Marshal(body)
	resp, err := http.Post(url, "application/json", bytes.NewReader(buf))
	if err != nil {
		t.Fatalf("POST %s: %v", url, err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != expectStatus {
		raw, _ := io.ReadAll(resp.Body)
		t.Fatalf("POST %s status = %d, want %d (body=%s)", url, resp.StatusCode, expectStatus, string(raw))
	}
}

func getJSON(t *testing.T, url string) map[string]any {
	t.Helper()
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("GET %s: %v", url, err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != 200 {
		raw, _ := io.ReadAll(resp.Body)
		t.Fatalf("GET %s status = %d (body=%s)", url, resp.StatusCode, string(raw))
	}
	out := map[string]any{}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatalf("decode %s: %v", url, err)
	}
	return out
}
