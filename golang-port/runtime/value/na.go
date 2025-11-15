package value

import "math"

/* NA constant */
var Na = math.NaN()

/* IsNa checks if value is NaN */
func IsNa(v float64) bool {
	return math.IsNaN(v)
}

/* Nz replaces NaN with replacement value (default 0) */
func Nz(value, replacement float64) float64 {
	if math.IsNaN(value) {
		return replacement
	}
	return value
}

/* Fixnan fills NaN values with last valid value, iterating backwards */
func Fixnan(source []float64) []float64 {
	if len(source) == 0 {
		return source
	}

	result := make([]float64, len(source))
	lastValid := math.NaN()

	for i := len(source) - 1; i >= 0; i-- {
		if !math.IsNaN(source[i]) {
			lastValid = source[i]
			result[i] = source[i]
		} else {
			result[i] = lastValid
		}
	}

	return result
}
