package detect

import (
	"testing"

	"github.com/felixgeelhaar/chronos/internal/config"
	"github.com/felixgeelhaar/chronos/internal/domain"
)

func cfgWithThresholds(established, strong float64) *config.Config {
	return &config.Config{
		ConfidenceClassEstablished: established,
		ConfidenceClassStrong:      strong,
	}
}

// TestClassifyConfidence_Buckets pins the three buckets at the
// boundaries the issue called for: tentative at the floor,
// established at 2×, strong at 5×.
func TestClassifyConfidence_Buckets(t *testing.T) {
	t.Parallel()
	cfg := cfgWithThresholds(2.0, 5.0)
	cases := []struct {
		n, min int
		want   domain.ConfidenceClass
	}{
		{4, 4, domain.ConfidenceClassTentative},    // ratio 1.0
		{7, 4, domain.ConfidenceClassTentative},    // ratio 1.75
		{8, 4, domain.ConfidenceClassEstablished},  // ratio 2.0
		{12, 4, domain.ConfidenceClassEstablished}, // ratio 3.0
		{20, 4, domain.ConfidenceClassStrong},      // ratio 5.0
		{100, 4, domain.ConfidenceClassStrong},     // ratio 25
	}
	for _, c := range cases {
		got := ClassifyConfidence(c.n, c.min, cfg)
		if got != c.want {
			t.Errorf("ClassifyConfidence(%d, %d) = %q, want %q", c.n, c.min, got, c.want)
		}
	}
}

// TestClassifyConfidence_ZeroInputsReturnEmpty proves the fail-soft
// contract: detectors that don't supply meaningful n / minPoints
// (or pass nil cfg) get an empty class so the downstream "absent ==
// no claim about strength" semantics works.
func TestClassifyConfidence_ZeroInputsReturnEmpty(t *testing.T) {
	t.Parallel()
	cfg := cfgWithThresholds(2.0, 5.0)
	for _, c := range []struct{ n, min int }{
		{0, 5}, {5, 0}, {-1, 5}, {5, -1}, {0, 0},
	} {
		if got := ClassifyConfidence(c.n, c.min, cfg); got != "" {
			t.Errorf("ClassifyConfidence(%d, %d) = %q, want empty", c.n, c.min, got)
		}
	}
	if got := ClassifyConfidence(10, 5, nil); got != "" {
		t.Errorf("nil cfg should yield empty, got %q", got)
	}
}

// TestClassifyConfidence_ConfigurableThresholds proves the
// CHRONOS_CONFIDENCE_* knobs flow through. An operator who wants a
// less generous "strong" bar (e.g. 10×) gets exactly that bar.
func TestClassifyConfidence_ConfigurableThresholds(t *testing.T) {
	t.Parallel()
	cfg := cfgWithThresholds(3.0, 10.0)
	// n=20, min=4 → ratio 5.0. With strong=10 this should NOT be strong.
	if got := ClassifyConfidence(20, 4, cfg); got != domain.ConfidenceClassEstablished {
		t.Errorf("ratio 5 with strong=10 should be established, got %q", got)
	}
	// ratio 10 → strong.
	if got := ClassifyConfidence(40, 4, cfg); got != domain.ConfidenceClassStrong {
		t.Errorf("ratio 10 with strong=10 should be strong, got %q", got)
	}
}
