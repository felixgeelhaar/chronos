package client

import (
	"time"

	"github.com/google/uuid"
)

// Signal is the wire shape of a signal returned by the HTTP API. It
// mirrors the API's JSON contract; new fields the server adds are
// picked up automatically as long as their JSON keys match. This type
// duplicates the internal/api.SignalDTO definition rather than
// importing it: the SDK is part of the public surface, internal/api
// is not.
type Signal struct {
	ID         uuid.UUID          `json:"id"`
	ScopeID    uuid.UUID          `json:"scope_id"`
	Series     uuid.UUID          `json:"series"`
	Pattern    string             `json:"pattern"`
	DetectedAt time.Time          `json:"detected_at"`
	Window     TimeWindow         `json:"window"`
	Strength   float64            `json:"strength"`
	Confidence float64            `json:"confidence"`
	Metrics    map[string]float64 `json:"metrics,omitempty"`
	Evidence   []Evidence         `json:"evidence,omitempty"`
	// Explanation carries detector-side context that lets downstream
	// consumers narrate WHY the signal fired. Omitted when the
	// detector did not surface one.
	Explanation *Explanation `json:"explanation,omitempty"`
	// ConfidenceClass is the qualitative grade (tentative /
	// established / strong). Empty when the detector did not classify.
	ConfidenceClass string `json:"confidence_class,omitempty"`
}

// Explanation is the wire shape of a detector explainability payload.
type Explanation struct {
	FeatureEvolution   []FeatureSample `json:"feature_evolution,omitempty"`
	ComparablePeers    int             `json:"comparable_peers,omitempty"`
	BaselineWindowDays int             `json:"baseline_window_days,omitempty"`
	ThresholdUsed      float64         `json:"threshold_used,omitempty"`
	DetectorVersion    string          `json:"detector_version,omitempty"`
}

// FeatureSample is one observation in an Explanation's feature evolution.
type FeatureSample struct {
	At    time.Time `json:"at"`
	Value float64   `json:"value"`
}

// TimeWindow is the analysis window over which a signal was detected.
type TimeWindow struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// Evidence is one piece of supporting evidence for a signal.
type Evidence struct {
	Series  uuid.UUID          `json:"series"`
	Time    time.Time          `json:"time"`
	Kind    string             `json:"kind"`
	Score   float64            `json:"score"`
	Metrics map[string]float64 `json:"metrics,omitempty"`
}

// IngestRequest is the wire shape for POST /v1/ingest. It mirrors the
// vision's TimeSeriesPoint with Chronos's multi-feature support.
type IngestRequest struct {
	ID        uuid.UUID         `json:"id,omitempty"`
	EntityID  uuid.UUID         `json:"entity_id"`
	ScopeID   uuid.UUID         `json:"scope_id"`
	Timestamp time.Time         `json:"timestamp"`
	Features  []float64         `json:"features"`
	Labels    []string          `json:"labels,omitempty"`
	Meta      map[string]string `json:"meta,omitempty"`
	Adapter   string            `json:"adapter,omitempty"`
}

// PatternType constants mirror the engine's domain.PatternType enum so
// callers can switch on stable string values without importing
// internal packages.
const (
	PatternTypeRecurrence            = "recurrence"
	PatternTypeTrend                 = "trend"
	PatternTypeSpike                 = "spike"
	PatternTypeDrop                  = "drop"
	PatternTypeStall                 = "stall"
	PatternTypeAnomaly               = "anomaly"
	PatternTypeSeasonality           = "seasonality"
	PatternTypeCorrelation           = "correlation"
	PatternTypeChangePoint           = "change_point"
	PatternTypeOutlierCluster        = "outlier_cluster"
	PatternTypeCrossScopeCorrelation = "cross_scope_correlation"
)
