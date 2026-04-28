// Package client is the public Go SDK for the Chronos HTTP API.
//
// It is decoupled from internal/domain: the wire types in this package
// are independent so internal refactors do not break consumers.
// Functional options keep the constructor stable as configuration
// grows.
//
// Typical use:
//
//	c, _ := client.New("http://chronos.local:7778",
//	    client.WithTimeout(10*time.Second),
//	)
//	signals, err := c.Signals().Scope(scopeID).List(ctx)
package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Client is a thread-safe Chronos HTTP API client. Construct with [New];
// the zero value is unusable.
type Client struct {
	baseURL    *url.URL
	httpClient *http.Client
	token      string
	userAgent  string
	logger     *slog.Logger
}

// Option configures a Client.
type Option func(*Client)

// WithHTTPClient overrides the underlying *http.Client. The client must
// be safe for concurrent use; the SDK does not wrap it.
func WithHTTPClient(h *http.Client) Option { return func(c *Client) { c.httpClient = h } }

// WithTimeout applies the given timeout to the underlying http.Client.
func WithTimeout(d time.Duration) Option {
	return func(c *Client) {
		if c.httpClient == nil {
			c.httpClient = &http.Client{}
		}
		c.httpClient.Timeout = d
	}
}

// WithToken sets a bearer token sent on every request as
// "Authorization: Bearer <token>". Empty tokens are ignored.
func WithToken(tok string) Option { return func(c *Client) { c.token = tok } }

// WithUserAgent overrides the default User-Agent header.
func WithUserAgent(ua string) Option { return func(c *Client) { c.userAgent = ua } }

// WithLogger attaches an slog.Logger; the client emits debug-level
// events for each request. Pass nil (or omit the option) to disable.
func WithLogger(l *slog.Logger) Option { return func(c *Client) { c.logger = l } }

// New constructs a Client targeting baseURL.
func New(baseURL string, opts ...Option) (*Client, error) {
	if baseURL == "" {
		return nil, errors.New("chronos client: base URL required")
	}
	parsed, err := url.Parse(strings.TrimRight(baseURL, "/"))
	if err != nil {
		return nil, fmt.Errorf("chronos client: parse base URL: %w", err)
	}
	c := &Client{
		baseURL:    parsed,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		userAgent:  "chronos-go-client/1",
	}
	for _, opt := range opts {
		opt(c)
	}
	if c.httpClient == nil {
		c.httpClient = &http.Client{Timeout: 30 * time.Second}
	}
	return c, nil
}

// Health calls GET /health and returns nil when the server reports
// healthy.
func (c *Client) Health(ctx context.Context) error {
	var body struct {
		Status string `json:"status"`
	}
	if err := c.do(ctx, http.MethodGet, "/health", nil, &body); err != nil {
		return err
	}
	if body.Status != "healthy" {
		return fmt.Errorf("chronos client: server reported status %q", body.Status)
	}
	return nil
}

// Ingest sends a single observation to /v1/ingest. ID and Timestamp
// are populated server-side if zero.
func (c *Client) Ingest(ctx context.Context, req IngestRequest) (uuid.UUID, error) {
	var body struct {
		Status string    `json:"status"`
		ID     uuid.UUID `json:"id"`
	}
	if err := c.do(ctx, http.MethodPost, "/v1/ingest", req, &body); err != nil {
		return uuid.Nil, err
	}
	return body.ID, nil
}

// Signals returns a fluent builder for the /v1/signals endpoint.
func (c *Client) Signals() *SignalQuery { return &SignalQuery{c: c} }

// do performs an HTTP request, applies headers, and decodes the
// response. Non-2xx responses are surfaced as APIError so callers can
// inspect the status code.
func (c *Client) do(ctx context.Context, method, path string, body any, out any) error {
	u := *c.baseURL
	pathPart, query, _ := strings.Cut(path, "?")
	u.Path = strings.TrimRight(u.Path, "/") + pathPart
	u.RawQuery = query

	var reqBody io.Reader
	if body != nil {
		buf, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("chronos client: marshal: %w", err)
		}
		reqBody = strings.NewReader(string(buf))
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), reqBody)
	if err != nil {
		return fmt.Errorf("chronos client: new request: %w", err)
	}
	if reqBody != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")
	if c.userAgent != "" {
		req.Header.Set("User-Agent", c.userAgent)
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	start := time.Now()
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("chronos client: %s %s: %w", method, path, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if c.logger != nil {
		c.logger.Debug("chronos client request",
			"method", method, "path", path, "status", resp.StatusCode, "duration", time.Since(start))
	}

	if resp.StatusCode >= 400 {
		raw, _ := io.ReadAll(resp.Body)
		return &APIError{Status: resp.StatusCode, Body: strings.TrimSpace(string(raw))}
	}
	if out == nil {
		return nil
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("chronos client: decode response: %w", err)
	}
	return nil
}

// APIError is returned for non-2xx HTTP responses. Callers can switch
// on Status to differentiate "not found" (404) from other failures.
type APIError struct {
	Status int
	Body   string
}

func (e *APIError) Error() string {
	if e.Body == "" {
		return fmt.Sprintf("chronos: HTTP %d", e.Status)
	}
	return fmt.Sprintf("chronos: HTTP %d: %s", e.Status, e.Body)
}
