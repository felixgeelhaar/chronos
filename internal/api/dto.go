package api

import (
	"errors"
	"time"

	"github.com/felixgeelhaar/chronos"
	"github.com/felixgeelhaar/chronos/internal/domain"
	"github.com/google/uuid"
)

// SignalDTO is the wire shape for signals returned by the HTTP API. It
// is decoupled from domain.Signal so internal refactors do not break
// clients. There is no Title/Summary/Suggestion: per the cognitive-
// stack vision, Chronos emits signals — not prose. Downstream
// consumers (Nous) interpret the structured fields.
type SignalDTO struct {
	ID         uuid.UUID          `json:"id"`
	ScopeID    uuid.UUID          `json:"scope_id"`
	Series     uuid.UUID          `json:"series"`
	Pattern    string             `json:"pattern"`
	DetectedAt time.Time          `json:"detected_at"`
	Window     TimeWindowDTO      `json:"window"`
	Strength   float64            `json:"strength"`
	Confidence float64            `json:"confidence"`
	Metrics    map[string]float64 `json:"metrics,omitempty"`
	Evidence   []EvidenceDTO      `json:"evidence,omitempty"`
}

// TimeWindowDTO is the analysis window over which the signal was
// detected.
type TimeWindowDTO struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// EvidenceDTO is one piece of supporting evidence.
type EvidenceDTO struct {
	Series  uuid.UUID          `json:"series"`
	Time    time.Time          `json:"time"`
	Kind    string             `json:"kind"`
	Score   float64            `json:"score"`
	Metrics map[string]float64 `json:"metrics,omitempty"`
}

// ToSignalDTO renders a domain.Signal into its wire form.
func ToSignalDTO(s domain.Signal) SignalDTO {
	dto := SignalDTO{
		ID:         s.ID,
		ScopeID:    s.ScopeID,
		Series:     s.Series,
		Pattern:    string(s.Pattern),
		DetectedAt: s.DetectedAt,
		Window:     TimeWindowDTO{Start: s.Window.Start, End: s.Window.End},
		Strength:   s.Strength,
		Confidence: s.Confidence,
		Metrics:    s.Metrics,
	}
	dto.Evidence = make([]EvidenceDTO, 0, len(s.Evidence))
	for _, e := range s.Evidence {
		dto.Evidence = append(dto.Evidence, EvidenceDTO{
			Series:  e.Series,
			Time:    e.Time,
			Kind:    e.Kind,
			Score:   e.Score,
			Metrics: e.Metrics,
		})
	}
	return dto
}

// IngestRequest is the wire shape for POST /v1/ingest. It mirrors the
// vision's TimeSeriesPoint with multi-feature support.
type IngestRequest struct {
	ID        uuid.UUID         `json:"id,omitempty"`
	EntityID  uuid.UUID         `json:"entity_id"`
	ScopeID   uuid.UUID         `json:"scope_id"`
	Timestamp time.Time         `json:"timestamp"`
	Features  []float64         `json:"features"`
	Labels    []string          `json:"labels,omitempty"`
	Meta      map[string]string `json:"meta,omitempty"`
	// Adapter labels the source for retention policy / count queries. If
	// empty, signals fall back to "http" so streaming clients are
	// distinguishable from pull adapters in the Count column.
	Adapter string `json:"adapter,omitempty"`
}

// toEntityState converts the wire request into a chronos.EntityState
// after applying defaults and validating.
func (req IngestRequest) toEntityState() (chronos.EntityState, error) {
	if req.Timestamp.IsZero() {
		req.Timestamp = time.Now().UTC()
	}
	id := req.ID
	if id == uuid.Nil {
		id = uuid.New()
	}
	state := chronos.EntityState{
		ID:        id,
		EntityID:  req.EntityID,
		ScopeID:   req.ScopeID,
		Timestamp: req.Timestamp,
		Features:  req.Features,
		Labels:    req.Labels,
		Meta:      req.Meta,
	}
	if err := state.Validate(); err != nil {
		return chronos.EntityState{}, errors.New("ingest: " + err.Error())
	}
	return state, nil
}
