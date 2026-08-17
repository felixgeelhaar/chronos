package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestPerceptionID_StableAndDiscriminating(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
	scope, series, partner := uuid.New(), uuid.New(), uuid.New()
	base := Signal{
		ScopeID: scope,
		Series:  series,
		Pattern: PatternTypeTrend,
		Window:  TimeWindow{Start: now.Add(-time.Hour), End: now},
	}
	a := PerceptionID(base)
	b := PerceptionID(base)
	if a != b {
		t.Fatalf("same perception produced different ids: %s vs %s", a, b)
	}
	if a == uuid.Nil {
		t.Fatal("perception id is nil")
	}

	grown := base
	grown.Window.End = now.Add(time.Minute)
	if PerceptionID(grown) == a {
		t.Error("growing window.End must mint a new id")
	}

	otherSeries := base
	otherSeries.Series = uuid.New()
	if PerceptionID(otherSeries) == a {
		t.Error("different series must mint a new id")
	}

	pair := Signal{
		ScopeID: scope,
		Series:  series,
		Pattern: PatternTypeCorrelation,
		Window:  TimeWindow{Start: now.Add(-time.Hour), End: now},
		Evidence: []Evidence{{
			Series: partner,
			Kind:   "pair_correlation",
		}},
	}
	otherPartner := pair
	otherPartner.Evidence = []Evidence{{Series: uuid.New(), Kind: "pair_correlation"}}
	if PerceptionID(pair) == PerceptionID(otherPartner) {
		t.Error("correlation pairs that share a window must not collide")
	}
}
