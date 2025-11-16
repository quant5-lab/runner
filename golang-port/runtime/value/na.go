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

/* Valuewhen returns source value when condition was true N occurrences ago (PineTS compatible) */
func Valuewhen(condition []bool, source []float64, occurrence int) []float64 {
	if len(condition) == 0 || len(source) == 0 || len(condition) != len(source) {
		return make([]float64, len(source))
	}

	result := make([]float64, len(source))
	for i := range result {
		result[i] = math.NaN()
	}

	for i := 0; i < len(condition); i++ {
		// Count how many times condition was true from start up to current bar
		trueCount := 0
		foundIndex := -1

		for j := i; j >= 0; j-- {
			if condition[j] {
				if trueCount == occurrence {
					foundIndex = j
					break
				}
				trueCount++
			}
		}

		// If we found the Nth occurrence, use that source value
		if foundIndex >= 0 {
			result[i] = source[foundIndex]
		}
	}

	return result
}
