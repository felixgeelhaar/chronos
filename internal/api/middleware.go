package api

import (
	"crypto/subtle"
	"log/slog"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/felixgeelhaar/chronos/internal/observability"
)

// Middleware wraps an http.Handler with cross-cutting behaviour. It is
// the standard "decorator" signature used by net/http composition.
type Middleware func(http.Handler) http.Handler

// Chain composes middlewares into a single handler. The first
// middleware in the slice is the outermost — i.e. it sees the request
// first and the response last.
func Chain(h http.Handler, mws ...Middleware) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}

// Logging emits a structured slog event for every HTTP request once
// the response has been written, and (when metrics is non-nil) records
// the observation for /metrics scraping. Fields on the slog event:
// method, path, status, duration, remote.
func Logging(logger *slog.Logger, metrics *observability.Metrics) Middleware {
	if logger == nil {
		logger = slog.Default()
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(rec, r)
			dur := time.Since(start)
			logger.Info("http request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", rec.status,
				"duration", dur,
				"remote", r.RemoteAddr,
			)
			metrics.ObserveHTTP(r.Method, rec.status, r.URL.Path, dur)
		})
	}
}

// BearerAuth requires every request to carry an "Authorization: Bearer
// <token>" header matching the configured token. If token is empty,
// authentication is disabled and the middleware is a no-op — that
// matches the documented "no auth out of the box" default. The
// /health route is always public so liveness probes never need
// credentials.
//
// Token comparison uses subtle.ConstantTimeCompare to avoid leaking
// the prefix length via timing.
func BearerAuth(token string) Middleware {
	if token == "" {
		return func(next http.Handler) http.Handler { return next }
	}
	want := []byte(token)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/health" {
				next.ServeHTTP(w, r)
				return
			}
			header := r.Header.Get("Authorization")
			const prefix = "Bearer "
			if !strings.HasPrefix(header, prefix) {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			got := []byte(header[len(prefix):])
			if subtle.ConstantTimeCompare(got, want) != 1 {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// Recover converts panics in downstream handlers into 500 responses
// and logs the stack trace. Without it, a single buggy handler can
// crash the whole HTTP server because net/http re-throws the panic.
func Recover(logger *slog.Logger) Middleware {
	if logger == nil {
		logger = slog.Default()
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					logger.Error("http panic recovered",
						"err", rec,
						"path", r.URL.Path,
						"stack", string(debug.Stack()),
					)
					http.Error(w, "internal error", http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// statusRecorder lets the logging middleware see the response status
// without buffering the body. WriteHeader stores the code; subsequent
// calls go straight to the wrapped writer.
type statusRecorder struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

// WriteHeader captures the status the handler chose, then forwards
// the call. It is part of the http.ResponseWriter interface and so
// must be exported even though statusRecorder itself is unexported.
func (s *statusRecorder) WriteHeader(code int) {
	if !s.wroteHeader {
		s.status = code
		s.wroteHeader = true
	}
	s.ResponseWriter.WriteHeader(code)
}

// Write forwards the body and notes that the header has been
// implicitly written if the handler did not call WriteHeader first.
func (s *statusRecorder) Write(p []byte) (int, error) {
	if !s.wroteHeader {
		s.wroteHeader = true
	}
	return s.ResponseWriter.Write(p)
}
