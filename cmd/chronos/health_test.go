package main

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestRunHealth_Healthy proves the happy path: a server answering
// /health with status "healthy" makes the probe exit successfully.
// This is the assertion the container HEALTHCHECK depends on.
func TestRunHealth_Healthy(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/health" {
			t.Errorf("probed %q, want /health", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"healthy","service":"chronos"}`))
	}))
	defer ts.Close()

	if err := runHealth([]string{"-addr", ts.URL}); err != nil {
		t.Fatalf("runHealth: %v", err)
	}
}

// TestRunHealth_Unhealthy: a 200 that does not say "healthy" must
// still fail. A probe that only checks the status code would report a
// degraded server as up.
func TestRunHealth_Unhealthy(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"degraded","service":"chronos"}`))
	}))
	defer ts.Close()

	err := runHealth([]string{"-addr", ts.URL})
	if err == nil {
		t.Fatal("expected an error for a non-healthy status")
	}
	assertExitCode(t, err, ExitError)
}

// TestRunHealth_Unreachable: nothing listening is the case the
// HEALTHCHECK exists to catch, so it must be a non-zero exit rather
// than a hang.
func TestRunHealth_Unreachable(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	url := ts.URL
	ts.Close() // free the port so the dial is refused

	start := time.Now()
	err := runHealth([]string{"-addr", url, "-timeout", "2s"})
	if err == nil {
		t.Fatal("expected an error probing a closed port")
	}
	if elapsed := time.Since(start); elapsed > 5*time.Second {
		t.Errorf("probe took %s; the timeout flag is not being honoured", elapsed)
	}
	assertExitCode(t, err, ExitError)
}

// TestRunHealth_BadFlag pins that flag misuse is a usage error, not a
// probe failure — the two exit codes mean different things to a
// wrapper script.
func TestRunHealth_BadFlag(t *testing.T) {
	err := runHealth([]string{"-nope"})
	if err == nil {
		t.Fatal("expected a usage error for an unknown flag")
	}
	assertExitCode(t, err, ExitUsage)
}

func assertExitCode(t *testing.T, err error, want ExitCode) {
	t.Helper()
	var ce *ChronosError
	if !errors.As(err, &ce) {
		t.Fatalf("error %v is not a *ChronosError", err)
	}
	if ce.Code != want {
		t.Errorf("exit code = %d, want %d", ce.Code, want)
	}
}
