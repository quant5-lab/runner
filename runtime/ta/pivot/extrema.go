package pivot

import "math"

type ValueExtractor func(index int) float64

type ExtremaChecker interface {
	IsCenterExtremum(centerValue float64, neighbors []float64) bool
}

type MaximumChecker struct{}

func (MaximumChecker) IsCenterExtremum(centerValue float64, neighbors []float64) bool {
	for _, neighbor := range neighbors {
		if math.IsNaN(neighbor) || neighbor >= centerValue {
			return false
		}
	}
	return true
}

type MinimumChecker struct{}

func (MinimumChecker) IsCenterExtremum(centerValue float64, neighbors []float64) bool {
	for _, neighbor := range neighbors {
		if math.IsNaN(neighbor) || neighbor <= centerValue {
			return false
		}
	}
	return true
}
