// Package insight provides generic types for detected patterns and insights.
package insight

import (
	"time"

	"github.com/google/uuid"
)

// Insight represents a statistically-backed pattern detected by Chronos.
type Insight struct {
	ID        uuid.UUID
	ScopeID   uuid.UUID        // The scope this insight belongs to (coach, team, etc.)
	Type      string           // Pattern type (e.g., "similar_transition", "anomaly_cluster")
	
	// What the insight is about
	SubjectEntity   uuid.UUID
	SubjectTime     time.Time
	
	// Evidence
	SimilarCases    []SimilarCase
	SampleSize      int
	Confidence      float64 // 0.0 - 1.0
	
	// Human-readable
	Title       string
	Summary     string
	Suggestion  string
	
	// Metadata
	GeneratedAt time.Time
	ValidUntil  *time.Time
	DismissedAt *time.Time
	Feedback    *Feedback
}

// SimilarCase is one piece of evidence supporting an insight.
type SimilarCase struct {
	EntityID    uuid.UUID
	Time        time.Time
	Similarity  float64
	OutcomeDiff float64 // Difference in outcome metric vs subject
}

// Feedback captures coach/user reaction to an insight.
type Feedback struct {
	Useful  bool
	Applied bool
	Reason  string
	At      time.Time
}

// IsActive returns true if the insight has not been dismissed and is still valid.
func (i *Insight) IsActive() bool {
	if i.DismissedAt != nil {
		return false
	}
	if i.ValidUntil != nil && time.Now().After(*i.ValidUntil) {
		return false
	}
	return true
}
