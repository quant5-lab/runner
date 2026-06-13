package testutil

import "math"

const priceRelEps = 1e-9

const FinancialToleranceSafetyMultiple = 5.0
const FinancialRelEps = MeasuredSizeResidualBound * FinancialToleranceSafetyMultiple
const financialToleranceSafetyMultiple = FinancialToleranceSafetyMultiple
const financialRelEps = FinancialRelEps
const financialDenominatorFloor = 1.0
const plotAbsEps = 1e-6

type numericTolerancePolicy struct {
	priceRelativeEpsilon     float64
	financialRelativeEpsilon float64
}

var goldenComparisonTolerance = numericTolerancePolicy{
	priceRelativeEpsilon:     priceRelEps,
	financialRelativeEpsilon: financialRelEps,
}

func floatWithin(a, b, tolerance float64) bool {
	if math.IsNaN(a) && math.IsNaN(b) {
		return true
	}
	if math.IsNaN(a) || math.IsNaN(b) {
		return false
	}
	if a == b {
		return true
	}
	return math.Abs(a-b) <= tolerance
}

func matchPrice(a, b float64) bool {
	return goldenComparisonTolerance.matchPrice(a, b)
}

func matchFinancial(a, b float64) bool {
	return goldenComparisonTolerance.matchFinancial(a, b)
}

func matchPlot(a, b float64) bool {
	return floatWithin(a, b, plotAbsEps)
}

func matchRelative(a, b, eps float64) bool {
	return numericTolerancePolicy{}.matchRelative(a, b, eps)
}

func (p numericTolerancePolicy) matchPrice(a, b float64) bool {
	return p.matchRelative(a, b, p.priceRelativeEpsilon)
}

func (p numericTolerancePolicy) matchFinancial(a, b float64) bool {
	return p.matchRelativeWithFloor(a, b, p.financialRelativeEpsilon, financialDenominatorFloor)
}

func (p numericTolerancePolicy) matchRelativeWithFloor(a, b, eps, floor float64) bool {
	if math.IsNaN(a) && math.IsNaN(b) {
		return true
	}
	if math.IsNaN(a) || math.IsNaN(b) {
		return false
	}
	if a == b {
		return true
	}
	scale := math.Max(math.Abs(a), math.Abs(b))
	if scale < floor {
		scale = floor
	}
	return math.Abs(a-b)/scale <= eps
}

func (p numericTolerancePolicy) matchRelative(a, b, eps float64) bool {
	if math.IsNaN(a) && math.IsNaN(b) {
		return true
	}
	if math.IsNaN(a) || math.IsNaN(b) {
		return false
	}
	if a == b {
		return true
	}
	scale := math.Max(math.Abs(a), math.Abs(b))
	if scale == 0 {
		return true
	}
	return math.Abs(a-b)/scale <= eps
}
