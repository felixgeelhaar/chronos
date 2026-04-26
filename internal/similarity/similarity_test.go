package similarity

import (
	"math"
	"testing"
)

func TestCosine(t *testing.T) {
	tests := []struct {
		name     string
		a        []float64
		b        []float64
		expected float64
	}{
		{
			name:     "identical vectors",
			a:        []float64{1, 2, 3},
			b:        []float64{1, 2, 3},
			expected: 1.0,
		},
		{
			name:     "orthogonal vectors",
			a:        []float64{1, 0, 0},
			b:        []float64{0, 1, 0},
			expected: 0.0,
		},
		{
			name:     "opposite vectors",
			a:        []float64{1, 2, 3},
			b:        []float64{-1, -2, -3},
			expected: -1.0,
		},
		{
			name:     "similar vectors",
			a:        []float64{1, 2, 3},
			b:        []float64{2, 3, 4},
			expected: 0.992583, // approximately
		},
		{
			name:     "different lengths",
			a:        []float64{1, 2},
			b:        []float64{1, 2, 3},
			expected: 0.0,
		},
		{
			name:     "empty vectors",
			a:        []float64{},
			b:        []float64{},
			expected: 0.0,
		},
		{
			name:     "zero vector",
			a:        []float64{0, 0, 0},
			b:        []float64{1, 2, 3},
			expected: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Cosine(tt.a, tt.b)
			diff := math.Abs(got - tt.expected)
			if diff > 0.0001 {
				t.Errorf("Cosine() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestWeightedCosine(t *testing.T) {
	a := []float64{1, 2, 3}
	b := []float64{2, 3, 4}
	weights := []float64{1, 1, 1}

	// With equal weights, should equal regular cosine
	weighted := WeightedCosine(a, b, weights)
	regular := Cosine(a, b)

	if math.Abs(weighted-regular) > 0.0001 {
		t.Errorf("WeightedCosine with equal weights should equal Cosine: got %v, want %v", weighted, regular)
	}

	// With different weights
	weights2 := []float64{2, 1, 0.5}
	weighted2 := WeightedCosine(a, b, weights2)
	if weighted2 == regular {
		t.Error("WeightedCosine with different weights should differ from regular Cosine")
	}
}

func TestEuclidean(t *testing.T) {
	a := []float64{1, 2, 3}
	b := []float64{4, 5, 6}
	
	got := Euclidean(a, b)
	expected := math.Sqrt(27) // sqrt((3)^2 + (3)^2 + (3)^2)
	
	if math.Abs(got-expected) > 0.0001 {
		t.Errorf("Euclidean() = %v, want %v", got, expected)
	}
}
