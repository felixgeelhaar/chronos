package domain

import (
	"errors"
	"math"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNormalise(t *testing.T) {
	tests := []struct {
		name     string
		input    []float64
		expected []float64
	}{
		{"simple values", []float64{1, 2, 3, 4, 5}, []float64{-1.4142, -0.7071, 0, 0.7071, 1.4142}},
		{"single value", []float64{5}, []float64{0}},
		{"empty", []float64{}, nil},
		{"all same", []float64{3, 3, 3}, []float64{0, 0, 0}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Normalise(tt.input)
			if len(got) != len(tt.expected) {
				t.Fatalf("Normalise() length = %d, want %d", len(got), len(tt.expected))
			}
			for i := range got {
				if math.Abs(got[i]-tt.expected[i]) > 0.001 {
					t.Errorf("Normalise()[%d] = %v, want %v", i, got[i], tt.expected[i])
				}
			}
		})
	}
}

func TestSignal_Validate(t *testing.T) {
	scope := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	series := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	now := time.Now()

	base := Signal{
		ID:         uuid.New(),
		ScopeID:    scope,
		Series:     series,
		Pattern:    PatternTypeRecurrence,
		DetectedAt: now,
		Window:     TimeWindow{Start: now.Add(-time.Hour), End: now},
		Strength:   0.9,
		Confidence: 0.8,
	}
	if err := base.Validate(); err != nil {
		t.Fatalf("valid signal rejected: %v", err)
	}

	tests := []struct {
		name string
		mut  func(*Signal)
		want error
	}{
		{"missing scope", func(s *Signal) { s.ScopeID = uuid.Nil }, ErrMissingScopeID},
		{"missing series", func(s *Signal) { s.Series = uuid.Nil }, ErrMissingSeriesID},
		{"missing pattern", func(s *Signal) { s.Pattern = "" }, ErrMissingPattern},
		{"strength > 1", func(s *Signal) { s.Strength = 1.2 }, ErrInvalidStrength},
		{"strength < 0", func(s *Signal) { s.Strength = -0.1 }, ErrInvalidStrength},
		{"confidence > 1", func(s *Signal) { s.Confidence = 1.5 }, ErrInvalidConfidence},
		{"confidence < 0", func(s *Signal) { s.Confidence = -0.1 }, ErrInvalidConfidence},
		{"window inverted", func(s *Signal) { s.Window = TimeWindow{Start: now, End: now.Add(-time.Hour)} }, ErrInvalidWindow},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cp := base
			tt.mut(&cp)
			err := cp.Validate()
			if !errors.Is(err, tt.want) {
				t.Fatalf("Validate() = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestTimeWindow_Validate(t *testing.T) {
	now := time.Now()
	if err := (TimeWindow{Start: now, End: now}).Validate(); err != nil {
		t.Errorf("zero-length window rejected: %v", err)
	}
	if err := (TimeWindow{Start: now, End: now.Add(-time.Second)}).Validate(); !errors.Is(err, ErrInvalidWindow) {
		t.Errorf("inverted window not rejected: %v", err)
	}
}
