package client

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// openStream issues a long-lived GET against an SSE endpoint and
// returns a channel that emits decoded Signal values until the stream
// closes (server EOF, ctx cancellation, or unrecoverable error).
//
// SSE connections must outlive c.httpClient.Timeout; we clone the
// underlying client and zero its Timeout so a 30-second default
// doesn't terminate the stream. Per-request cancellation flows
// through ctx as usual.
func (c *Client) openStream(ctx context.Context, path string) (<-chan Signal, error) {
	u := *c.baseURL
	pathPart, query, _ := strings.Cut(path, "?")
	u.Path = strings.TrimRight(u.Path, "/") + pathPart
	u.RawQuery = query

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("chronos client: build stream request: %w", err)
	}
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")
	if c.userAgent != "" {
		req.Header.Set("User-Agent", c.userAgent)
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	streamClient := *c.httpClient // shallow copy
	streamClient.Timeout = 0      // disable per-request deadline; ctx still applies

	resp, err := streamClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("chronos client: open stream: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		return nil, &APIError{Status: resp.StatusCode, Body: strings.TrimSpace(string(raw))}
	}

	out := make(chan Signal, 16)
	go c.runStream(ctx, resp, out)
	return out, nil
}

// runStream parses the SSE event stream and forwards `event: signal`
// frames as decoded Signal values. The channel is always closed before
// the goroutine returns so consumers can range over it safely.
func (c *Client) runStream(ctx context.Context, resp *http.Response, out chan<- Signal) {
	defer close(out)
	defer func() { _ = resp.Body.Close() }()

	// Default bufio Scanner token size is 64KB. SignalDTO with full
	// Evidence can comfortably fit, but raise the cap to 1 MiB so a
	// burst of long evidence rows doesn't truncate the stream.
	sc := bufio.NewScanner(resp.Body)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	var event string
	var data strings.Builder

	dispatch := func() {
		defer func() {
			event = ""
			data.Reset()
		}()
		// We only forward signal events. Other event names (future
		// "signal.batch", server "ping", etc.) are skipped so the
		// consumer doesn't have to filter.
		if event != "signal" || data.Len() == 0 {
			return
		}
		var sig Signal
		if err := json.Unmarshal([]byte(data.String()), &sig); err != nil {
			if c.logger != nil {
				c.logger.Debug("chronos client: stream decode failed", "err", err)
			}
			return
		}
		select {
		case out <- sig:
		case <-ctx.Done():
		}
	}

	for sc.Scan() {
		select {
		case <-ctx.Done():
			return
		default:
		}
		line := sc.Text()
		if line == "" {
			dispatch()
			continue
		}
		if strings.HasPrefix(line, ":") {
			// SSE comment (heartbeat). Ignore.
			continue
		}
		name, val, ok := strings.Cut(line, ":")
		if !ok {
			// Field-name-only lines are valid in SSE (treated as empty
			// value) but not meaningful here.
			continue
		}
		val = strings.TrimPrefix(val, " ")
		switch name {
		case "event":
			event = val
		case "data":
			if data.Len() > 0 {
				data.WriteByte('\n')
			}
			data.WriteString(val)
		}
		// id, retry: silently ignored — Chronos doesn't use Last-Event-ID
		// based replay, so they carry no meaning for this client.
	}
	// scanner.Err() is intentionally not surfaced: when ctx is
	// cancelled, the Read returns an error that has already been
	// signalled to the caller via channel close.
}
