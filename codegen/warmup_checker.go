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
	period     int
	baseOffset int
	strategy   SeriesAccessStrategy
}

func NewWarmupChecker(period int) *WarmupChecker {
	return &WarmupChecker{
		period:     period,
		baseOffset: 0,
		strategy:   NewTopLevelSeriesAccessStrategy(),
	}
}

func NewWarmupCheckerWithOffset(period int, baseOffset int) *WarmupChecker {
	return &WarmupChecker{
		period:     period,
		baseOffset: baseOffset,
		strategy:   NewTopLevelSeriesAccessStrategy(),
	}
}

/* WithSeriesStrategy configures context-aware series access for warmup check. */
func (w *WarmupChecker) WithSeriesStrategy(strategy SeriesAccessStrategy) *WarmupChecker {
	w.strategy = strategy
	return w
}

func (w *WarmupChecker) GenerateCheck(varName string, indenter *CodeIndenter) string {
	totalWarmup := w.period + w.baseOffset - 1
	code := indenter.Line(fmt.Sprintf("if ctx.BarIndex < %d {", totalWarmup))
	indenter.IncreaseIndent()
	code += indenter.Line(w.strategy.GenerateSet(varName, "math.NaN()"))
	indenter.DecreaseIndent()
	code += indenter.Line("} else {")
	return code
}

func (w *WarmupChecker) MinimumBarsRequired() int {
	return w.period + w.baseOffset
}
