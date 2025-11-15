package math

import (
	gomath "math"
)

/* Abs returns absolute value */
func Abs(x float64) float64 {
	return gomath.Abs(x)
}

/* Max returns maximum of two or more values */
func Max(values ...float64) float64 {
	if len(values) == 0 {
		return gomath.NaN()
	}
	max := values[0]
	for _, v := range values[1:] {
		if v > max {
			max = v
		}
	}
	return max
}

/* Min returns minimum of two or more values */
func Min(values ...float64) float64 {
	if len(values) == 0 {
		return gomath.NaN()
	}
	min := values[0]
	for _, v := range values[1:] {
		if v < min {
			min = v
		}
	}
	return min
}

/* Pow returns x raised to power y */
func Pow(x, y float64) float64 {
	return gomath.Pow(x, y)
}

/* Sqrt returns square root */
func Sqrt(x float64) float64 {
	return gomath.Sqrt(x)
}

/* Floor returns largest integer <= x */
func Floor(x float64) float64 {
	return gomath.Floor(x)
}

/* Ceil returns smallest integer >= x */
func Ceil(x float64) float64 {
	return gomath.Ceil(x)
}

/* Round returns nearest integer */
func Round(x float64) float64 {
	return gomath.Round(x)
}

/* Log returns natural logarithm */
func Log(x float64) float64 {
	return gomath.Log(x)
}

/* Exp returns e^x */
func Exp(x float64) float64 {
	return gomath.Exp(x)
}

/* Sum returns sum of slice */
func Sum(values []float64) float64 {
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum
}

/* Avg returns average of values */
func Avg(values ...float64) float64 {
	if len(values) == 0 {
		return gomath.NaN()
	}
	return Sum(values) / float64(len(values))
}
