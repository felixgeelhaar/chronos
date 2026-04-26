// Package similarity provides generic similarity computation for feature vectors.
package similarity

import "math"

// Cosine computes the cosine similarity between two vectors.
// Returns a value in [-1, 1]. For pattern detection, values near 1.0 indicate high similarity.
func Cosine(a, b []float64) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}

	dot := 0.0
	normA := 0.0
	normB := 0.0

	for i := range a {
		dot += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}

// WeightedCosine computes cosine similarity with per-dimension weights.
func WeightedCosine(a, b, weights []float64) float64 {
	if len(a) != len(b) || len(a) != len(weights) || len(a) == 0 {
		return 0
	}

	dot := 0.0
	normA := 0.0
	normB := 0.0

	for i := range a {
		wa := a[i] * weights[i]
		wb := b[i] * weights[i]
		dot += wa * wb
		normA += wa * wa
		normB += wb * wb
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}

// Euclidean computes the Euclidean distance between two vectors.
func Euclidean(a, b []float64) float64 {
	if len(a) != len(b) {
		return math.Inf(1)
	}

	sum := 0.0
	for i := range a {
		d := a[i] - b[i]
		sum += d * d
	}

	return math.Sqrt(sum)
}
