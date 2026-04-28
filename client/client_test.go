package client_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/felixgeelhaar/chronos/client"
	"github.com/google/uuid"
)

func TestClient_Health(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/health" {
			t.Errorf("path = %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
	}))
	defer srv.Close()

	c, err := client.New(srv.URL)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err := c.Health(context.Background()); err != nil {
		t.Fatalf("Health: %v", err)
	}
}

func TestClient_ListSignals_Filters(t *testing.T) {
	scope := uuid.New()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("scope_id") != scope.String() {
			t.Errorf("scope_id = %q", q.Get("scope_id"))
		}
		if q.Get("pattern") != client.PatternTypeRecurrence {
			t.Errorf("pattern = %q", q.Get("pattern"))
		}
		if q.Get("min_confidence") != "0.7" {
			t.Errorf("min_confidence = %q", q.Get("min_confidence"))
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"count":   1,
			"signals": []client.Signal{{ID: uuid.New(), ScopeID: scope, Pattern: client.PatternTypeRecurrence, Confidence: 0.9, DetectedAt: time.Now()}},
		})
	}))
	defer srv.Close()

	c, _ := client.New(srv.URL)
	got, err := c.Signals().Scope(scope).Pattern(client.PatternTypeRecurrence).MinConfidence(0.7).List(context.Background())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("len = %d", len(got))
	}
	if got[0].Pattern != client.PatternTypeRecurrence {
		t.Errorf("pattern = %s", got[0].Pattern)
	}
}

func TestClient_Ingest(t *testing.T) {
	id := uuid.New()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/ingest" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(map[string]any{"status": "accepted", "id": id})
	}))
	defer srv.Close()

	c, _ := client.New(srv.URL)
	gotID, err := c.Ingest(context.Background(), client.IngestRequest{
		EntityID:  uuid.New(),
		ScopeID:   uuid.New(),
		Features:  []float64{1, 2, 3, 5},
		Adapter:   "test",
		Timestamp: time.Now(),
	})
	if err != nil {
		t.Fatalf("Ingest: %v", err)
	}
	if gotID != id {
		t.Errorf("id = %s, want %s", gotID, id)
	}
}

func TestClient_AuthAndError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer s3cret" {
			t.Errorf("Authorization = %q", got)
		}
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer srv.Close()

	c, _ := client.New(srv.URL, client.WithToken("s3cret"))
	_, err := c.Signals().Scope(uuid.New()).List(context.Background())
	var apiErr *client.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error = %v, want APIError", err)
	}
	if apiErr.Status != http.StatusInternalServerError {
		t.Errorf("status = %d", apiErr.Status)
	}
	if !strings.Contains(apiErr.Body, "boom") {
		t.Errorf("body = %q", apiErr.Body)
	}
}

func TestSignalQuery_RequiresScope(t *testing.T) {
	c, _ := client.New("http://127.0.0.1:1")
	if _, err := c.Signals().List(context.Background()); err == nil {
		t.Errorf("expected error for missing scope")
	}
}
