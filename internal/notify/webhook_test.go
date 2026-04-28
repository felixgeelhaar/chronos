package notify

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/felixgeelhaar/chronos/internal/observability"
	"github.com/google/uuid"
)

// receivedRequest captures what the test's HTTP server saw, in a form
// that's safe to inspect from the test goroutine.
type receivedRequest struct {
	body      []byte
	signature string
	delivery  string
	event     string
}

// recorder is a thread-safe collector for received requests. The
// httptest handler runs on its own goroutine, so a plain slice would
// race with the test goroutine reading the count.
type recorder struct {
	mu   sync.Mutex
	reqs []receivedRequest
}

func (r *recorder) add(req receivedRequest) {
	r.mu.Lock()
	r.reqs = append(r.reqs, req)
	r.mu.Unlock()
}

func (r *recorder) snapshot() []receivedRequest {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]receivedRequest, len(r.reqs))
	copy(out, r.reqs)
	return out
}

func (r *recorder) count() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.reqs)
}

// recordingServer returns a handler that stashes each incoming request
// into rec, plus a handle to set its next response status (atomic so
// the test goroutine can twiddle it between requests).
func recordingServer(t *testing.T, status *atomic.Int32, rec *recorder) *httptest.Server {
	t.Helper()
	if status.Load() == 0 {
		status.Store(http.StatusAccepted)
	}
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = r.Body.Close()
		rec.add(receivedRequest{
			body:      body,
			signature: r.Header.Get("X-Chronos-Signature"),
			delivery:  r.Header.Get("X-Chronos-Delivery"),
			event:     r.Header.Get("X-Chronos-Event"),
		})
		w.WriteHeader(int(status.Load()))
	}))
}

func TestWebhook_DeliversWithSignedPayload(t *testing.T) {
	rec := &recorder{}
	var status atomic.Int32
	srv := recordingServer(t, &status, rec)
	defer srv.Close()

	wh := NewWebhook(WebhookConfig{
		URLs:    []string{srv.URL},
		Secret:  "topsecret",
		Timeout: time.Second,
	}, observability.New(), nil)

	wh.Notify(context.Background(), sampleSignal())
	_ = wh.Close()

	reqs := rec.snapshot()
	if len(reqs) != 1 {
		t.Fatalf("expected 1 request, got %d", len(reqs))
	}
	got := reqs[0]
	if got.event != "signal.detected" {
		t.Errorf("X-Chronos-Event = %q", got.event)
	}
	if _, err := uuid.Parse(got.delivery); err != nil {
		t.Errorf("X-Chronos-Delivery is not a UUID: %q", got.delivery)
	}
	mac := hmac.New(sha256.New, []byte("topsecret"))
	mac.Write(got.body)
	want := "sha256=" + hex.EncodeToString(mac.Sum(nil))
	if got.signature != want {
		t.Errorf("signature mismatch:\n  got:  %s\n  want: %s", got.signature, want)
	}
	if !strings.Contains(string(got.body), `"pattern":"recurrence"`) {
		t.Errorf("body did not contain expected pattern field: %s", got.body)
	}
}

func TestWebhook_NoSecret_OmitsSignatureHeader(t *testing.T) {
	rec := &recorder{}
	var status atomic.Int32
	srv := recordingServer(t, &status, rec)
	defer srv.Close()

	wh := NewWebhook(WebhookConfig{URLs: []string{srv.URL}}, observability.New(), nil)
	wh.Notify(context.Background(), sampleSignal())
	_ = wh.Close()

	reqs := rec.snapshot()
	if len(reqs) != 1 || reqs[0].signature != "" {
		t.Fatalf("expected empty signature when no secret; got %q (req count %d)", reqs[0].signature, len(reqs))
	}
}

func TestWebhook_FansOutToAllURLs(t *testing.T) {
	recA, recB := &recorder{}, &recorder{}
	var statusA, statusB atomic.Int32
	srvA := recordingServer(t, &statusA, recA)
	srvB := recordingServer(t, &statusB, recB)
	defer srvA.Close()
	defer srvB.Close()

	wh := NewWebhook(WebhookConfig{URLs: []string{srvA.URL, srvB.URL}}, observability.New(), nil)
	wh.Notify(context.Background(), sampleSignal())
	_ = wh.Close()

	if recA.count() != 1 || recB.count() != 1 {
		t.Fatalf("fan-out did not reach both: A=%d B=%d", recA.count(), recB.count())
	}
}

func TestWebhook_RetriesOn5xxAndStopsOnSuccess(t *testing.T) {
	rec := &recorder{}
	var status atomic.Int32
	status.Store(http.StatusInternalServerError)
	srv := recordingServer(t, &status, rec)
	defer srv.Close()

	wh := NewWebhook(WebhookConfig{URLs: []string{srv.URL}, Retries: 2, Timeout: time.Second}, observability.New(), nil)

	done := make(chan struct{})
	go func() {
		wh.Notify(context.Background(), sampleSignal())
		_ = wh.Close()
		close(done)
	}()

	deadline := time.After(3 * time.Second)
	for {
		if rec.count() >= 1 {
			status.Store(http.StatusOK)
			break
		}
		select {
		case <-deadline:
			t.Fatal("no attempts received within 3s")
		case <-time.After(20 * time.Millisecond):
		}
	}
	<-done

	if rec.count() < 2 {
		t.Errorf("expected retries, got %d attempts", rec.count())
	}
}

func TestWebhook_DoesNotRetryOn4xx(t *testing.T) {
	rec := &recorder{}
	var status atomic.Int32
	status.Store(http.StatusBadRequest)
	srv := recordingServer(t, &status, rec)
	defer srv.Close()

	wh := NewWebhook(WebhookConfig{URLs: []string{srv.URL}, Retries: 3, Timeout: time.Second}, observability.New(), nil)
	wh.Notify(context.Background(), sampleSignal())
	_ = wh.Close()

	if rec.count() != 1 {
		t.Errorf("4xx must not retry; got %d attempts", rec.count())
	}
}

func TestWebhook_NilMetricsIsSafe(t *testing.T) {
	rec := &recorder{}
	var status atomic.Int32
	srv := recordingServer(t, &status, rec)
	defer srv.Close()

	wh := NewWebhook(WebhookConfig{URLs: []string{srv.URL}}, nil, nil)
	wh.Notify(context.Background(), sampleSignal())
	_ = wh.Close()
	if rec.count() != 1 {
		t.Errorf("nil metrics must not block delivery: got %d", rec.count())
	}
}

func TestWebhook_EmptyURLsIsNoop(t *testing.T) {
	wh := NewWebhook(WebhookConfig{}, observability.New(), nil)
	// Should not panic, should not block.
	wh.Notify(context.Background(), sampleSignal())
	_ = wh.Close()
}
