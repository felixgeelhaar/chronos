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

func seedSignal(t *testing.T, ts string, scope uuid.UUID, at time.Time, id uuid.UUID, conf float64) domain.Signal {
	t.Helper()
	_ = ts
	return domain.Signal{
		ID:         id,
		ScopeID:    scope,
		Series:     uuid.New(),
		Pattern:    domain.PatternTypeRecurrence,
		DetectedAt: at,
		Window:     domain.TimeWindow{Start: at.Add(-time.Hour), End: at},
		Strength:   conf,
		Confidence: conf,
	}
}

// TestCursor_RoundTrip is the unit-level check: encode → decode must
// be lossless on both the timestamp and the id.
func TestCursor_RoundTrip(t *testing.T) {
	t.Parallel()
	in := signalCursor{
		DetectedAt: time.Date(2026, 5, 24, 13, 14, 15, 123456789, time.UTC),
		ID:         uuid.New(),
	}
	out, err := decodeSignalCursor(encodeSignalCursor(in))
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !out.DetectedAt.Equal(in.DetectedAt) {
		t.Errorf("DetectedAt drift: got %v, want %v", out.DetectedAt, in.DetectedAt)
	}
	if out.ID != in.ID {
		t.Errorf("ID drift: got %s, want %s", out.ID, in.ID)
	}
}

// TestCursor_Decode_BadInput pins the fail-closed contract — junk in,
// error out, no silent fallback to "no filter applied".
func TestCursor_Decode_BadInput(t *testing.T) {
	t.Parallel()
	for _, raw := range []string{"", "not-base64!", "bm9zZXBhcmF0b3I", "bm90LWEtdGltZXxub3QtYS11dWlk"} {
		if _, err := decodeSignalCursor(raw); err == nil {
			t.Errorf("expected error for input %q", raw)
		}
	}
}

// TestListSignals_SinceCursor_ResumesAfterCursor pins the polling
// contract: a client that received next_cursor=X gets back only
// signals strictly after X on the next poll, never the cursor row
// itself or anything before it.
func TestListSignals_SinceCursor_ResumesAfterCursor(t *testing.T) {
	ts, mem := setupServer(t)
	defer ts.Close()
	scope := uuid.New()
	t0 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	earliest := seedSignal(t, "old", scope, t0, uuid.New(), 0.6)
	middle := seedSignal(t, "mid", scope, t0.Add(time.Hour), uuid.New(), 0.7)
	newest := seedSignal(t, "new", scope, t0.Add(2*time.Hour), uuid.New(), 0.8)
	for _, s := range []domain.Signal{earliest, middle, newest} {
		if err := mem.Signals.Save(context.Background(), s); err != nil {
			t.Fatalf("Save: %v", err)
		}
	}

	// First request: no cursor, expect all three + a next_cursor.
	resp, _ := http.Get(ts.URL + "/v1/signals?scope_id=" + scope.String())
	defer func() { _ = resp.Body.Close() }()
	var first struct {
		Signals    []SignalDTO `json:"signals"`
		NextCursor string      `json:"next_cursor"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&first); err != nil {
		t.Fatalf("decode 1: %v", err)
	}
	if len(first.Signals) != 3 {
		t.Fatalf("first page got %d, want 3", len(first.Signals))
	}
	if first.NextCursor == "" {
		t.Fatal("first response missing next_cursor")
	}
	if first.Signals[0].ID != newest.ID {
		t.Errorf("first row should be newest")
	}

	// Now poll with that cursor. New batch should be empty (no
	// signals after the cursor) and there should be no next_cursor.
	resp2, _ := http.Get(ts.URL + "/v1/signals?scope_id=" + scope.String() + "&since_cursor=" + first.NextCursor)
	defer func() { _ = resp2.Body.Close() }()
	var second struct {
		Signals    []SignalDTO `json:"signals"`
		NextCursor string      `json:"next_cursor"`
	}
	if err := json.NewDecoder(resp2.Body).Decode(&second); err != nil {
		t.Fatalf("decode 2: %v", err)
	}
	if len(second.Signals) != 0 {
		t.Errorf("post-cursor returned %d signals, want 0: %+v", len(second.Signals), second.Signals)
	}
	if second.NextCursor != "" {
		t.Errorf("empty response carried next_cursor %q", second.NextCursor)
	}

	// Insert one fresh signal after the cursor. Polling again must
	// surface only that signal.
	fresh := seedSignal(t, "fresh", scope, t0.Add(3*time.Hour), uuid.New(), 0.9)
	if err := mem.Signals.Save(context.Background(), fresh); err != nil {
		t.Fatalf("Save fresh: %v", err)
	}
	resp3, _ := http.Get(ts.URL + "/v1/signals?scope_id=" + scope.String() + "&since_cursor=" + first.NextCursor)
	defer func() { _ = resp3.Body.Close() }()
	var third struct {
		Signals []SignalDTO `json:"signals"`
	}
	if err := json.NewDecoder(resp3.Body).Decode(&third); err != nil {
		t.Fatalf("decode 3: %v", err)
	}
	if len(third.Signals) != 1 || third.Signals[0].ID != fresh.ID {
		t.Errorf("expected only the fresh signal, got %+v", third.Signals)
	}
}

// TestListSignals_SinceCursor_TieBreaksOnEqualTimestamp pins the
// motivating advantage of cursor over since=<timestamp>: two signals
// at the exact same DetectedAt must not be returned twice on resume.
func TestListSignals_SinceCursor_TieBreaksOnEqualTimestamp(t *testing.T) {
	ts, mem := setupServer(t)
	defer ts.Close()
	scope := uuid.New()
	sameTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	// Generate two ids and pin the lex-smaller one as "earlier-by-tie".
	a, b := uuid.New(), uuid.New()
	earlierID, laterID := a, b
	if a.String() > b.String() {
		earlierID, laterID = b, a
	}
	first := seedSignal(t, "first", scope, sameTime, earlierID, 0.7)
	second := seedSignal(t, "second", scope, sameTime, laterID, 0.7)
	for _, s := range []domain.Signal{first, second} {
		if err := mem.Signals.Save(context.Background(), s); err != nil {
			t.Fatalf("Save: %v", err)
		}
	}

	cursor := encodeSignalCursor(signalCursor{DetectedAt: sameTime, ID: earlierID})
	resp, _ := http.Get(ts.URL + "/v1/signals?scope_id=" + scope.String() + "&since_cursor=" + cursor)
	defer func() { _ = resp.Body.Close() }()
	var got struct {
		Signals []SignalDTO `json:"signals"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(got.Signals) != 1 || got.Signals[0].ID != laterID {
		t.Errorf("expected only the later-by-tie row %s, got %+v", laterID, got.Signals)
	}
}

// TestListSignals_BadCursor returns 400, not 500. A garbage cursor
// must not silently degrade into an unfiltered list — that would
// flood the caller with already-seen rows on a paste error.
func TestListSignals_BadCursor(t *testing.T) {
	ts, _ := setupServer(t)
	defer ts.Close()
	resp, err := http.Get(ts.URL + "/v1/signals?scope_id=" + uuid.New().String() + "&since_cursor=garbage!!!")
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", resp.StatusCode)
	}
}
