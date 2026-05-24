package detect

import (
	"github.com/felixgeelhaar/chronos/internal/config"
	"github.com/felixgeelhaar/chronos/internal/domain"
)

// ClassifyConfidence buckets a signal into tentative / established /
// strong based on how many supporting observations the detector saw
// relative to its MIN_POINTS floor. The thresholds are multipliers on
// MIN_POINTS so the classifier scales naturally across detectors —
// a trend with MinPoints=4 and n=20 is "strong" by the same rule as
// a cross-scope-correlation with MinPoints=5 and n=25.
//
// Returns the empty class (zero ConfidenceClass) when n or minPoints
// is non-positive; the caller decides whether to omit the field or
// leave it empty in the signal. A non-positive minPoints would be a
// detector misconfiguration; failing silent here avoids surfacing
// noise downstream.
func ClassifyConfidence(n, minPoints int, cfg *config.Config) domain.ConfidenceClass {
	if n <= 0 || minPoints <= 0 || cfg == nil {
		return ""
	}
	ratio := float64(n) / float64(minPoints)
	if cfg.ConfidenceClassStrong > 0 && ratio >= cfg.ConfidenceClassStrong {
		return domain.ConfidenceClassStrong
	}
	if cfg.ConfidenceClassEstablished > 0 && ratio >= cfg.ConfidenceClassEstablished {
		return domain.ConfidenceClassEstablished
	}
	return domain.ConfidenceClassTentative
}
