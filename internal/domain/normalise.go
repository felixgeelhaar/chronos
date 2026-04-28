package domain

import "math"

// Normalise applies z-score normalisation to a feature slice and returns a
// new slice. It returns nil for an empty input and a zero slice when the
// input has no variance.
func Normalise(features []float64) []float64 {
	if len(features) == 0 {
		return nil
	}

	mean := 0.0
	for _, v := range features {
		mean += v
	}
	mean /= float64(len(features))

	variance := 0.0
	for _, v := range features {
		d := v - mean
		variance += d * d
	}
	stddev := math.Sqrt(variance / float64(len(features)))

	out := make([]float64, len(features))
	if stddev == 0 {
		return out
	}
	for i, v := range features {
		out[i] = (v - mean) / stddev
	}
	return out
}
