package observability

import (
	"bytes"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestMetrics_RenderShape(t *testing.T) {
	m := New()
	m.ObserveSignal("recurrence")
	m.ObserveSignal("recurrence")
	m.ObserveSignal("trend")
	m.ObserveObservations("ascend", 100)
	m.ObserveObservations("memory", 5)
	m.ObserveHTTP("GET", 200, "/v1/signals", 12*time.Millisecond)
	m.ObserveHTTP("GET", 200, "/v1/signals", 8*time.Millisecond)
	m.ObserveHTTP("GET", 404, "/v1/signals/abc-123", 3*time.Millisecond)

	var buf bytes.Buffer
	if err := m.Render(&buf); err != nil {
		t.Fatalf("Render: %v", err)
	}
	out := buf.String()

	wants := []string{
		`# HELP chronos_signals_emitted_total`,
		`# TYPE chronos_signals_emitted_total counter`,
		`chronos_signals_emitted_total{pattern="recurrence"} 2`,
		`chronos_signals_emitted_total{pattern="trend"} 1`,
		`chronos_observations_total{adapter="ascend"} 100`,
		`chronos_observations_total{adapter="memory"} 5`,
		`chronos_http_requests_total{method="GET",status="200"} 2`,
		`chronos_http_requests_total{method="GET",status="404"} 1`,
		// Per-id paths must be folded to :id so cardinality is bounded.
		`path="/v1/signals/:id"`,
	}
	for _, w := range wants {
		if !strings.Contains(out, w) {
			t.Errorf("Render output missing %q\n--- output ---\n%s", w, out)
		}
	}
}

func TestMetrics_ConcurrentObservations(t *testing.T) {
	m := New()
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				m.ObserveSignal("recurrence")
				m.ObserveObservations("memory", 1)
				m.ObserveHTTP("GET", 200, "/v1/signals", time.Millisecond)
			}
		}()
	}
	wg.Wait()

	var buf bytes.Buffer
	if err := m.Render(&buf); err != nil {
		t.Fatalf("Render: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, `chronos_signals_emitted_total{pattern="recurrence"} 5000`) {
		t.Errorf("expected 5000 recurrence signals, got: %s", out)
	}
	if !strings.Contains(out, `chronos_observations_total{adapter="memory"} 5000`) {
		t.Errorf("expected 5000 observations, got: %s", out)
	}
}

func TestNormalisePath_OnlySignalsByID(t *testing.T) {
	tests := map[string]string{
		"/health":              "/health",
		"/v1/signals":          "/v1/signals",
		"/v1/signals/abc-123":  "/v1/signals/:id",
		"/v1/signals/":         "/v1/signals/",
		"/v1/ingest":           "/v1/ingest",
	}
	for in, want := range tests {
		if got := normalisePath(in); got != want {
			t.Errorf("normalisePath(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestEscapeValue_HandlesQuotesAndBackslashes(t *testing.T) {
	got := escapeValue(`abc"def\ghi` + "\n")
	want := `"abc\"def\\ghi\n"`
	if got != want {
		t.Errorf("escapeValue() = %s, want %s", got, want)
	}
}

func TestNilMetrics_NoOps(t *testing.T) {
	var m *Metrics // nil pointer
	// Must not panic.
	m.ObserveSignal("x")
	m.ObserveObservations("y", 1)
	m.ObserveHTTP("GET", 200, "/", time.Millisecond)
}
