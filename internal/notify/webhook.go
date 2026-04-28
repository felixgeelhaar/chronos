package notify

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/felixgeelhaar/chronos/internal/api"
	"github.com/felixgeelhaar/chronos/internal/domain"
	"github.com/felixgeelhaar/chronos/internal/observability"
	"github.com/google/uuid"
)

// WebhookConfig describes a webhook fan-out target set. URLs is the
// only required field; an empty slice disables the transport. Secret
// is optional — when empty, no X-Chronos-Signature header is sent and
// receivers cannot authenticate the payload (acceptable for trusted
// network paths but not over the public internet).
type WebhookConfig struct {
	URLs    []string
	Secret  string
	Timeout time.Duration // default 5s when zero
	Retries int           // default 1 when negative
}

// Webhook posts each newly-persisted signal as JSON to a configured
// set of URLs. It is a transport: it does not interpret the signal,
// does not decide whether to deliver, and never blocks the
// persistence path. Delivery is at-most-once per URL with a small
// best-effort retry on 5xx responses; consumers de-duplicate by
// Signal.ID and treat headers as authoritative.
//
// Headers per request:
//
//	X-Chronos-Signature: sha256=<hex hmac-sha256 of body>  // when Secret is set
//	X-Chronos-Delivery:  <uuid v4>                          // unique per send attempt
//	X-Chronos-Event:     signal.detected
//
// The signed payload is the same JSON shape as /v1/signals returns
// (api.SignalDTO), so a consumer can treat webhook bodies and pulled
// list responses identically.
type Webhook struct {
	cfg     WebhookConfig
	client  *http.Client
	metrics *observability.Metrics
	logger  *slog.Logger

	wg sync.WaitGroup // tracks in-flight goroutines for graceful shutdown
}

// NewWebhook constructs a Webhook from configuration. The returned
// notifier is safe for concurrent use; pass it through ports.Notifier.
func NewWebhook(cfg WebhookConfig, metrics *observability.Metrics, logger *slog.Logger) *Webhook {
	if cfg.Timeout <= 0 {
		cfg.Timeout = 5 * time.Second
	}
	if cfg.Retries < 0 {
		cfg.Retries = 0
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &Webhook{
		cfg:     cfg,
		client:  &http.Client{Timeout: cfg.Timeout},
		metrics: metrics,
		logger:  logger,
	}
}

// Notify fans the signal out to every configured URL in parallel.
// Failures are logged and recorded as metrics but never propagate.
// Returns immediately; sends complete in background goroutines.
func (w *Webhook) Notify(ctx context.Context, sig domain.Signal) {
	if len(w.cfg.URLs) == 0 {
		return
	}
	payload, err := json.Marshal(api.ToSignalDTO(sig))
	if err != nil {
		w.logger.Error("webhook: marshal failed", "err", err, "signal_id", sig.ID)
		return
	}
	signature := w.sign(payload)
	delivery := uuid.NewString()

	for _, url := range w.cfg.URLs {
		w.wg.Add(1)
		go func(target string) {
			defer w.wg.Done()
			w.deliver(ctx, target, payload, signature, delivery, sig.ID)
		}(url)
	}
}

// Close blocks until every in-flight delivery finishes (or its
// per-request timeout expires). Useful from cmd/serve's graceful
// shutdown so we do not abandon pending pushes.
func (w *Webhook) Close() error {
	w.wg.Wait()
	return nil
}

// deliver is the per-URL retry loop. It honours ctx cancellation
// between attempts but lets an in-flight HTTP call complete (the
// http.Client carries its own timeout).
func (w *Webhook) deliver(ctx context.Context, url string, payload []byte, signature, delivery string, signalID uuid.UUID) {
	var lastErr error
	for attempt := 0; attempt <= w.cfg.Retries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				w.logger.Warn("webhook: cancelled before retry", "url", url, "delivery", delivery, "signal_id", signalID)
				return
			case <-time.After(time.Second):
			}
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
		if err != nil {
			lastErr = err
			break // request construction errors are not retryable
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Chronos-Event", "signal.detected")
		req.Header.Set("X-Chronos-Delivery", delivery)
		if signature != "" {
			req.Header.Set("X-Chronos-Signature", signature)
		}

		resp, err := w.client.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()

		switch {
		case resp.StatusCode >= 200 && resp.StatusCode < 300:
			w.metrics.ObserveWebhook("success", resp.StatusCode)
			return
		case resp.StatusCode >= 400 && resp.StatusCode < 500:
			// 4xx is the consumer rejecting our payload; retrying
			// won't change the outcome. Record and stop.
			w.metrics.ObserveWebhook("client_error", resp.StatusCode)
			w.logger.Warn("webhook: 4xx response, not retrying",
				"url", url, "status", resp.StatusCode, "delivery", delivery, "signal_id", signalID)
			return
		default:
			lastErr = fmt.Errorf("status %d", resp.StatusCode)
		}
	}
	w.metrics.ObserveWebhook("failure", 0)
	w.logger.Error("webhook: delivery failed after retries",
		"url", url, "delivery", delivery, "signal_id", signalID,
		"attempts", w.cfg.Retries+1, "err", lastErr)
}

// sign returns "sha256=<hex>" for body, or "" when no secret is
// configured.
func (w *Webhook) sign(body []byte) string {
	if w.cfg.Secret == "" {
		return ""
	}
	mac := hmac.New(sha256.New, []byte(w.cfg.Secret))
	mac.Write(body)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}
