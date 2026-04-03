package ta

import (
	"math"
)

/* Linreg calculates Linear Regression using least squares method */
func Linreg(source []float64, length int, offset int) []float64 {
	if length <= 0 || len(source) == 0 {
		return source
	}

	result := make([]float64, len(source))
	for i := range result {
		result[i] = math.NaN()
	}

	n := float64(length)
	formulaMultiplier := float64(length - 1 - offset)

	for i := length - 1; i < len(source); i++ {
		var sumX, sumY, sumXY, sumX2 float64

		for j := 0; j < length; j++ {
			x := float64(j)
			y := source[i-length+1+j]

			if math.IsNaN(y) {
				sumX = math.NaN()
				break
			}

			sumX += x
			sumY += y
			sumXY += x * y
			sumX2 += x * x
		}

		if math.IsNaN(sumX) {
			continue
		}

		denominator := n*sumX2 - sumX*sumX
		if denominator == 0 {
			result[i] = sumY / n
			continue
		}

		slope := (n*sumXY - sumX*sumY) / denominator
		intercept := (sumY - slope*sumX) / n
		result[i] = intercept + slope*formulaMultiplier
	}

	return result
}

/* calculateLeastSquares computes slope and intercept for linear regression */
func calculateLeastSquares(window []float64) (slope, intercept float64) {
	n := float64(len(window))
	if n == 0 {
		return 0, 0
	}

	var sumX, sumY, sumXY, sumX2 float64

	for i, y := range window {
		x := float64(i)
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}

	denominator := n*sumX2 - sumX*sumX
	if denominator == 0 {
		return 0, sumY / n
	}

	slope = (n*sumXY - sumX*sumY) / denominator
	intercept = (sumY - slope*sumX) / n
	return slope, intercept
}
