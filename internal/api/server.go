// Package api provides the HTTP REST API for Chronos.
package api

import (
	"encoding/json"
	"net/http"

	"github.com/felixgeelhaar/chronos/internal/config"
	"github.com/felixgeelhaar/chronos/internal/store/sqlite"
	"github.com/felixgeelhaar/chronos/pkg/insight"
	"github.com/google/uuid"
)

// Server wraps the HTTP handlers.
type Server struct {
	store *sqlite.Store
	cfg   *config.Config
}

// NewServer creates an HTTP server.
func NewServer(store *sqlite.Store, cfg *config.Config) *Server {
	return &Server{store: store, cfg: cfg}
}

// RegisterRoutes sets up all HTTP routes.
func (s *Server) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/v1/insights", s.handleInsights)
	mux.HandleFunc("/v1/insights/", s.handleInsightDetail)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{
		"status": "healthy",
		"service": "chronos",
	})
}

func (s *Server) handleInsights(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	scopeIDStr := r.URL.Query().Get("scope_id")
	if scopeIDStr == "" {
		http.Error(w, "scope_id required", http.StatusBadRequest)
		return
	}

	scopeID, err := uuid.Parse(scopeIDStr)
	if err != nil {
		http.Error(w, "invalid scope_id", http.StatusBadRequest)
		return
	}

	insights, err := s.store.LoadInsights(r.Context(), scopeID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"insights": insights,
		"count":    len(insights),
	})
}

func (s *Server) handleInsightDetail(w http.ResponseWriter, r *http.Request) {
	// Extract insight ID from /v1/insights/{id}
	idStr := r.URL.Path[len("/v1/insights/"):]
	insightID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid insight id", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodPost:
		s.handleDismissInsight(w, r, insightID)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleDismissInsight(w http.ResponseWriter, r *http.Request, insightID uuid.UUID) {
	var req struct {
		DismissedBy string `json:"dismissed_by"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	dismissedBy := uuid.Nil
	if req.DismissedBy != "" {
		var err error
		dismissedBy, err = uuid.Parse(req.DismissedBy)
		if err != nil {
			http.Error(w, "invalid dismissed_by", http.StatusBadRequest)
			return
		}
	}

	if err := s.store.DismissInsight(r.Context(), insightID, dismissedBy); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "dismissed"})
}

func respondJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
