package api

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestBearerAuth_DisabledWhenEmpty(t *testing.T) {
	called := false
	h := BearerAuth("")(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodGet, "/v1/signals", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if !called {
		t.Fatal("handler was gated despite empty token")
	}
}

func TestBearerAuth_AllowsHealthWithoutToken(t *testing.T) {
	called := false
	h := BearerAuth("s3cret")(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if !called {
		t.Fatal("/health was gated by auth — should be public")
	}
}

func TestBearerAuth_RequiresMatchingToken(t *testing.T) {
	h := BearerAuth("s3cret")(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	tests := []struct {
		name   string
		header string
		want   int
	}{
		{"missing", "", http.StatusUnauthorized},
		{"wrong scheme", "Basic abc", http.StatusUnauthorized},
		{"wrong token", "Bearer wrong", http.StatusUnauthorized},
		{"correct token", "Bearer s3cret", http.StatusOK},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/v1/signals", nil)
			if tt.header != "" {
				req.Header.Set("Authorization", tt.header)
			}
			rr := httptest.NewRecorder()
			h.ServeHTTP(rr, req)
			if rr.Code != tt.want {
				t.Errorf("status = %d, want %d", rr.Code, tt.want)
			}
		})
	}
}

func TestLogging_RecordsStatusAndPath(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))

	h := Logging(logger, nil)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	}))

	req := httptest.NewRequest(http.MethodGet, "/v1/signals", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusTeapot {
		t.Errorf("status = %d", rr.Code)
	}
	out := buf.String()
	if !strings.Contains(out, "status=418") {
		t.Errorf("log missing status: %s", out)
	}
	if !strings.Contains(out, "path=/v1/signals") {
		t.Errorf("log missing path: %s", out)
	}
}

func TestRecover_TurnsPanicInto500(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelError}))

	h := Recover(logger)(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		panic("kaboom")
	}))

	req := httptest.NewRequest(http.MethodGet, "/anything", nil)
	rr := httptest.NewRecorder()

	// The middleware must not propagate the panic.
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", rr.Code)
	}
	if !strings.Contains(buf.String(), "kaboom") {
		t.Errorf("log missing panic value: %s", buf.String())
	}
}

func TestChain_OutermostFirst(t *testing.T) {
	var order []string
	a := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "a-before")
			next.ServeHTTP(w, r)
			order = append(order, "a-after")
		})
	}
	b := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "b-before")
			next.ServeHTTP(w, r)
			order = append(order, "b-after")
		})
	}
	final := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		order = append(order, "handler")
		w.WriteHeader(http.StatusOK)
	})

	h := Chain(final, a, b)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(httptest.NewRecorder(), req)

	want := []string{"a-before", "b-before", "handler", "b-after", "a-after"}
	if len(order) != len(want) {
		t.Fatalf("len = %d want %d (got %v)", len(order), len(want), order)
	}
	for i := range want {
		if order[i] != want[i] {
			t.Errorf("order[%d] = %q, want %q", i, order[i], want[i])
		}
	}
}
