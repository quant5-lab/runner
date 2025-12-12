package codegen

import "fmt"

// WarmupChecker generates code to handle the warmup period for technical indicators.
//
// Technical indicators require a minimum number of bars (the "period") before they can
// produce valid calculations. During the warmup phase, indicators should output NaN.
//
// For example, a 20-period SMA needs 20 bars of historical data before it can calculate
// the first valid average. Bars 0-18 should return NaN, and calculation starts at bar 19.
//
// Usage:
//
//	checker := NewWarmupChecker(20)
//	indenter := NewCodeIndenter()
//	code := checker.GenerateCheck("sma20", &indenter)
//
// Generated code:
//
//	if ctx.BarIndex < 19 {
//	    sma20Series.Set(math.NaN())
//	} else {
//	    // ... calculation code ...
//	}
//
// Design:
//   - Single Responsibility: Only handles warmup period logic
//   - Reusable: Works with any indicator that needs warmup handling
//   - Testable: Easy to verify warmup boundary conditions
type WarmupChecker struct {
	period int // Minimum bars required for valid calculation
}

func NewWarmupChecker(period int) *WarmupChecker {
	return &WarmupChecker{period: period}
}

func (w *WarmupChecker) GenerateCheck(varName string, indenter *CodeIndenter) string {
	code := indenter.Line(fmt.Sprintf("if ctx.BarIndex < %d-1 {", w.period))
	indenter.IncreaseIndent()
	code += indenter.Line(fmt.Sprintf("%sSeries.Set(math.NaN())", varName))
	indenter.DecreaseIndent()
	code += indenter.Line("} else {")
	return code
}

func (w *WarmupChecker) MinimumBarsRequired() int {
	return w.period
}
