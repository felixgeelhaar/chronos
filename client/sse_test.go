package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
)

// streamingHandler is a minimal SSE producer for the tests. It writes
// a heartbeat comment, optional preamble events, then each provided
// signal as an `event: signal` frame. The handler holds the
// connection open until ctx is cancelled or the channel is closed.
func streamingHandler(t *testing.T, frames []string) http.HandlerFunc {
	t.Helper()
	return func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			t.Fatal("ResponseWriter does not support flushing")
			return
		}
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, ": connected\n\n")
		flusher.Flush()
		for _, f := range frames {
			_, _ = fmt.Fprint(w, f)
			flusher.Flush()
		}
		// Hold open until the client disconnects so we observe the
		// channel-close behaviour on ctx cancellation.
		<-r.Context().Done()
	}
}

func TestStream_DeliversFramesAndClosesOnCancel(t *testing.T) {
	scope := uuid.New()
	sigA := Signal{ID: uuid.New(), ScopeID: scope, Series: uuid.New(), Pattern: "recurrence", Strength: 0.9, Confidence: 0.8}
	sigB := Signal{ID: uuid.New(), ScopeID: scope, Series: uuid.New(), Pattern: "trend", Strength: 0.7, Confidence: 0.6}

	mustEvent := func(s Signal) string {
		body, _ := json.Marshal(s)
		return fmt.Sprintf("event: signal\ndata: %s\n\n", body)
	}
	frames := []string{mustEvent(sigA), mustEvent(sigB)}

	mux := http.NewServeMux()
	mux.HandleFunc("/v1/signals/stream", streamingHandler(t, frames))
	ts := httptest.NewServer(mux)
	defer ts.Close()

	c, err := New(ts.URL)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch, err := c.Signals().Scope(scope).Stream(ctx)
	if err != nil {
		t.Fatalf("Stream: %v", err)
	}

	got := readN(t, ch, 2, 2*time.Second)
	if len(got) != 2 {
		t.Fatalf("expected 2 signals, got %d", len(got))
	}
	if got[0].ID != sigA.ID || got[1].ID != sigB.ID {
		t.Errorf("ordering mismatch:\n  got[0]=%s want=%s\n  got[1]=%s want=%s",
			got[0].ID, sigA.ID, got[1].ID, sigB.ID)
	}

	// Cancelling ctx must close the channel within a reasonable window.
	cancel()
	select {
	case _, open := <-ch:
		if open {
			t.Fatal("expected channel close after ctx cancel")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("channel did not close after ctx cancel")
	}
}

func TestStream_RejectsWhenScopeMissing(t *testing.T) {
	c, _ := New("http://example.invalid")
	if _, err := c.Signals().Stream(context.Background()); err == nil {
		t.Fatal("expected error when Scope is unset")
	}
}

func TestStream_SurfacesNon2xxAsAPIError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/signals/stream", func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "streaming disabled", http.StatusNotImplemented)
	})
	ts := httptest.NewServer(mux)
	defer ts.Close()

	c, _ := New(ts.URL)
	_, err := c.Signals().Scope(uuid.New()).Stream(context.Background())
	if err == nil {
		t.Fatal("expected error on 501")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}
	if apiErr.Status != http.StatusNotImplemented {
		t.Errorf("status = %d", apiErr.Status)
	}
}

func TestStream_IgnoresHeartbeatsAndUnknownEvents(t *testing.T) {
	scope := uuid.New()
	wanted := Signal{ID: uuid.New(), ScopeID: scope, Series: uuid.New(), Pattern: "spike", Strength: 1, Confidence: 1}
	body, _ := json.Marshal(wanted)
	frames := []string{
		": keep-alive\n\n",          // SSE comment
		"event: ping\ndata: {}\n\n", // unrelated event type
		fmt.Sprintf("event: signal\ndata: %s\n\n", body),
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/signals/stream", streamingHandler(t, frames))
	ts := httptest.NewServer(mux)
	defer ts.Close()

	c, _ := New(ts.URL)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ch, err := c.Signals().Scope(scope).Stream(ctx)
	if err != nil {
		t.Fatalf("Stream: %v", err)
	}

	got := readN(t, ch, 1, 2*time.Second)
	if len(got) != 1 || got[0].ID != wanted.ID {
		t.Fatalf("expected only the signal event, got %+v", got)
	}
}

func TestStream_HandlesMultiLineDataField(t *testing.T) {
	scope := uuid.New()
	wanted := Signal{ID: uuid.New(), ScopeID: scope, Series: uuid.New(), Pattern: "stall", Strength: 0.5, Confidence: 0.5}

	// Split the JSON object across two `data:` lines at a structural
	// boundary (comma between top-level fields) so the SSE-mandated
	// newline reassembly still produces parseable JSON. SSE
	// concatenates consecutive data lines with "\n", which is
	// whitespace inside a JSON object — valid.
	idHalf := fmt.Sprintf(`{"id":%q,"scope_id":%q,"series":%q`, wanted.ID, wanted.ScopeID, wanted.Series)
	rest := fmt.Sprintf(`"pattern":"stall","strength":0.5,"confidence":0.5}`)
	frame := "event: signal\ndata: " + idHalf + ",\ndata: " + rest + "\n\n"

	mux := http.NewServeMux()
	mux.HandleFunc("/v1/signals/stream", streamingHandler(t, []string{frame}))
	ts := httptest.NewServer(mux)
	defer ts.Close()

	c, _ := New(ts.URL)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ch, err := c.Signals().Scope(scope).Stream(ctx)
	if err != nil {
		t.Fatalf("Stream: %v", err)
	}

	got := readN(t, ch, 1, 2*time.Second)
	if len(got) != 1 || got[0].ID != wanted.ID {
		t.Fatalf("multi-line data was not reassembled: got %+v", got)
	}
}

// readN waits for up to timeout to receive count signals, returning
// however many it got within the deadline.
func readN(t *testing.T, ch <-chan Signal, count int, timeout time.Duration) []Signal {
	t.Helper()
	out := make([]Signal, 0, count)
	deadline := time.After(timeout)
	for len(out) < count {
		select {
		case s, ok := <-ch:
			if !ok {
				return out
			}
			out = append(out, s)
		case <-deadline:
			return out
		}
	}
	return out
}
