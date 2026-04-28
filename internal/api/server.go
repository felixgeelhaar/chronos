// Package api provides Chronos's HTTP REST API. It is a thin transport
// layer over the persistence ports: parse the request, call a port,
// render the response. Per the cognitive-stack vision Chronos surfaces
// signals; it does not interpret or render prose. The wire shape is
// just the structured signal — Title/Summary/Suggestion are not part
// of this API.
package api

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/felixgeelhaar/chronos"
	"github.com/felixgeelhaar/chronos/internal/domain"
	"github.com/felixgeelhaar/chronos/internal/observability"
	"github.com/felixgeelhaar/chronos/internal/ports"
	"github.com/google/uuid"
)

// Server holds dependencies for HTTP handlers.
type Server struct {
	states  ports.EntityStateRepository
	signals ports.SignalRepository
	metrics *observability.Metrics
	logger  *slog.Logger
}

// NewServer wires an HTTP server. Pass slog.Default() for the logger
// if the caller has no preference. The states repository is required
// to support the streaming /v1/ingest endpoint; pass nil to disable
// it (the route still mounts but returns 501). metrics may be nil,
// in which case /metrics returns an empty document and observations
// in handlers are no-ops.
func NewServer(states ports.EntityStateRepository, signals ports.SignalRepository, metrics *observability.Metrics, logger *slog.Logger) *Server {
	if logger == nil {
		logger = slog.Default()
	}
	return &Server{states: states, signals: signals, metrics: metrics, logger: logger}
}

// RegisterRoutes wires API routes onto a ServeMux. Routes are stable
// wire contracts; bump /v1 to introduce incompatible changes.
func (s *Server) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/metrics", s.handleMetrics)
	mux.HandleFunc("/v1/ingest", s.handleIngest)
	mux.HandleFunc("/v1/signals", s.handleSignals)
	mux.HandleFunc("/v1/signals/", s.handleSignalDetail)
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{"status": "healthy", "service": "chronos"})
}

// handleMetrics renders the registry in Prometheus exposition
// format. When no registry is wired, the endpoint still returns 200
// with an empty body so scrape configs do not flap on misconfigured
// deployments.
func (s *Server) handleMetrics(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	if s.metrics == nil {
		return
	}
	if err := s.metrics.Render(w); err != nil {
		s.logger.Error("metrics render failed", "err", err)
	}
}

// handleIngest accepts a single TimeSeriesPoint / EntityState and
// persists it through the EntityStateRepository. Detection is not
// triggered automatically by ingest — the compute pipeline handles
// detection at job time.
func (s *Server) handleIngest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if s.states == nil {
		http.Error(w, "ingest disabled", http.StatusNotImplemented)
		return
	}
	var body IngestRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	state, err := body.toEntityState()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	adapter := body.Adapter
	if adapter == "" {
		adapter = "http"
	}
	if err := s.states.Ingest(r.Context(), adapter, state); err != nil {
		s.logger.Error("ingest failed", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	s.metrics.ObserveObservations(adapter, 1)
	respondJSON(w, http.StatusAccepted, map[string]string{"status": "accepted", "id": state.ID.String()})
}

// handleSignals lists signals matching a filter from the query string.
// All filters are optional except scope_id, which is required.
func (s *Server) handleSignals(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	filter, err := parseSignalFilter(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	signals, err := s.signals.List(r.Context(), filter)
	if err != nil {
		s.logger.Error("list signals failed", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	dtos := make([]SignalDTO, 0, len(signals))
	for _, sig := range signals {
		dtos = append(dtos, ToSignalDTO(sig))
	}
	respondJSON(w, http.StatusOK, map[string]any{"signals": dtos, "count": len(dtos)})
}

func (s *Server) handleSignalDetail(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/v1/signals/")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid signal id", http.StatusBadRequest)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	sig, err := s.signals.Get(r.Context(), id)
	if errors.Is(err, domain.ErrSignalNotFound) {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if err != nil {
		s.logger.Error("get signal failed", "id", id, "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	respondJSON(w, http.StatusOK, ToSignalDTO(sig))
}

func parseSignalFilter(r *http.Request) (ports.SignalFilter, error) {
	q := r.URL.Query()
	scopeIDStr := q.Get("scope_id")
	if scopeIDStr == "" {
		return ports.SignalFilter{}, errors.New("scope_id required")
	}
	scopeID, err := uuid.Parse(scopeIDStr)
	if err != nil {
		return ports.SignalFilter{}, errors.New("invalid scope_id")
	}
	f := ports.SignalFilter{ScopeID: scopeID}

	if v := q.Get("series"); v != "" {
		series, err := uuid.Parse(v)
		if err != nil {
			return ports.SignalFilter{}, errors.New("invalid series")
		}
		f.Series = &series
	}
	if v := q.Get("pattern"); v != "" {
		pat := domain.PatternType(v)
		f.Pattern = &pat
	}
	if v := q.Get("since"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return ports.SignalFilter{}, errors.New("invalid since")
		}
		f.Since = &t
	}
	if v := q.Get("until"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return ports.SignalFilter{}, errors.New("invalid until")
		}
		f.Until = &t
	}
	if v := q.Get("min_confidence"); v != "" {
		conf, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return ports.SignalFilter{}, errors.New("invalid min_confidence")
		}
		f.MinConfidence = &conf
	}
	if v := q.Get("limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 0 {
			return ports.SignalFilter{}, errors.New("invalid limit")
		}
		f.Limit = n
	}
	return f, nil
}

// _ unused-import suppression — keeps chronos imported in case future
// helpers need it. (chronos.EntityState references live in dto.go.)
var _ = chronos.EntityState{}

func respondJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
