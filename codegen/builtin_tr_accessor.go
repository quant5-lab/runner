package codegen

import "fmt"

// BuiltinTrueRangeAccessor generates inline tr calculations when ta.tr is passed as a
// source argument to TA functions (e.g. ta.sma(ta.tr, period), ta.ema(ta.tr, period)).
//
// Semantic contract: mirrors Pine's ta.tr default (handle_na=false).
// Bar 0 has no previous close → returns NaN, identical to Pine's na.
//
// Contrast with TrueRangeAccessGenerator, which is used internally by ta.atr and
// mirrors Pine's ta.tr(true) (handle_na=true) → returns high-low on bar 0.
type BuiltinTrueRangeAccessor struct{}

func NewBuiltinTrueRangeAccessor() *BuiltinTrueRangeAccessor {
	return &BuiltinTrueRangeAccessor{}
}

func (a *BuiltinTrueRangeAccessor) GenerateLoopValueAccess(loopVar string) string {
	return fmt.Sprintf(
		"func() float64 { "+
			"barIdx := ctx.BarIndex-%s; "+
			"if barIdx < 1 { return math.NaN() }; "+
			"prevClose := ctx.Data[barIdx-1].Close; "+
			"currentBar := ctx.Data[barIdx]; "+
			"return math.Max(currentBar.High - currentBar.Low, math.Max(math.Abs(currentBar.High - prevClose), math.Abs(currentBar.Low - prevClose))) "+
			"}()",
		loopVar,
	)
}

func (a *BuiltinTrueRangeAccessor) GenerateInitialValueAccess(period int) string {
	return fmt.Sprintf(
		"func() float64 { "+
			"barIdx := ctx.BarIndex-%d; "+
			"if barIdx < 1 { return math.NaN() }; "+
			"prevClose := ctx.Data[barIdx-1].Close; "+
			"currentBar := ctx.Data[barIdx]; "+
			"return math.Max(currentBar.High - currentBar.Low, math.Max(math.Abs(currentBar.High - prevClose), math.Abs(currentBar.Low - prevClose))) "+
			"}()",
		period-1,
	)
}

func (a *BuiltinTrueRangeAccessor) GenerateCurrentValueAccess() string {
	return "func() float64 { " +
		"if ctx.BarIndex < 1 { return math.NaN() }; " +
		"prevClose := ctx.Data[ctx.BarIndex-1].Close; " +
		"currentBar := ctx.Data[ctx.BarIndex]; " +
		"return math.Max(currentBar.High - currentBar.Low, math.Max(math.Abs(currentBar.High - prevClose), math.Abs(currentBar.Low - prevClose))) " +
		"}()"
}

func (a *BuiltinTrueRangeAccessor) GetPreamble() string { return "" }

func (a *BuiltinTrueRangeAccessor) GetBaseOffset() int { return 0 }
