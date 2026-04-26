package vector

import (
	"math"
	"testing"

	"github.com/google/uuid"
)

func TestNormalise(t *testing.T) {
	tests := []struct {
		name     string
		input    []float64
		expected []float64
	}{
		{
			name:     "simple values",
			input:    []float64{1, 2, 3, 4, 5},
			expected: []float64{-1.4142, -0.7071, 0, 0.7071, 1.4142},
		},
		{
			name:     "single value",
			input:    []float64{5},
			expected: []float64{0},
		},
		{
			name:     "empty",
			input:    []float64{},
			expected: nil,
		},
		{
			name:     "all same",
			input:    []float64{3, 3, 3},
			expected: []float64{0, 0, 0},
		},
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

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		state   EntityState
		wantErr bool
	}{
		{
			name: "valid",
			state: EntityState{
				EntityID: uuid.MustParse("22222222-2222-2222-2222-222222222222"),
				Features: []float64{1, 2, 3},
			},
			wantErr: false,
		},
		{
			name: "missing entity ID",
			state: EntityState{
				Features: []float64{1, 2, 3},
			},
			wantErr: true,
		},
		{
			name: "no features",
			state: EntityState{
				EntityID: uuid.MustParse("22222222-2222-2222-2222-222222222222"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.state.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
