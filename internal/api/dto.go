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
	// Explanation carries detector-side context that lets downstream
	// consumers narrate WHY the signal fired without re-deriving the
	// data. Omitted when the detector did not surface one.
	Explanation *ExplanationDTO `json:"explanation,omitempty"`
}

// ExplanationDTO is the wire shape of an Explanation value object.
// All fields are optional; a Signal without an Explanation omits this
// object entirely (pointer in SignalDTO).
type ExplanationDTO struct {
	FeatureEvolution   []FeatureSampleDTO `json:"feature_evolution,omitempty"`
	ComparablePeers    int                `json:"comparable_peers,omitempty"`
	BaselineWindowDays int                `json:"baseline_window_days,omitempty"`
	ThresholdUsed      float64            `json:"threshold_used,omitempty"`
	DetectorVersion    string             `json:"detector_version,omitempty"`
}

// FeatureSampleDTO is one observation in the feature evolution series.
type FeatureSampleDTO struct {
	At    time.Time `json:"at"`
	Value float64   `json:"value"`
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
	if !s.Explanation.IsZero() {
		ex := &ExplanationDTO{
			ComparablePeers:    s.Explanation.ComparablePeers,
			BaselineWindowDays: s.Explanation.BaselineWindowDays,
			ThresholdUsed:      s.Explanation.ThresholdUsed,
			DetectorVersion:    s.Explanation.DetectorVersion,
		}
		for _, fs := range s.Explanation.FeatureEvolution {
			ex.FeatureEvolution = append(ex.FeatureEvolution, FeatureSampleDTO{
				At: fs.At, Value: fs.Value,
			})
		}
		dto.Explanation = ex
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

// IngestBatchRequest is the wire shape for POST /v1/ingest/batch. It
// accepts an array of observations in one HTTP call so integrators
// backfilling history don't pay one round-trip per datapoint.
//
// All observations land in a single repository call; on any one
// observation failing validation the entire batch is rejected
// (all-or-nothing semantics — partial writes would be hard to roll
// back and harder to reason about).
type IngestBatchRequest struct {
	Observations []IngestRequest `json:"observations"`
	// DeferDetection is accepted for API forward-compatibility. Chronos
	// already separates ingest from detection (the compute pipeline
	// runs detection at job time, not on ingest), so the flag is
	// currently a no-op — it is echoed in the response so callers can
	// rely on the field shape without branching.
	DeferDetection bool `json:"defer_detection,omitempty"`
}

// IngestBatchResponse confirms how many observations were persisted
// and whether detection was deferred.
type IngestBatchResponse struct {
	Accepted       int  `json:"accepted"`
	DeferDetection bool `json:"defer_detection"`
}

// MaxIngestBatchSize caps a single /v1/ingest/batch payload. Beyond
// this, callers must chunk — the limit guards memory in the handler
// path and gives operators a predictable upper bound on transaction
// duration.
const MaxIngestBatchSize = 1000

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
