package codegen

import "fmt"

/* BuiltinTrueRangeAccessor generates inline tr calculations for TA loop iterations */
type BuiltinTrueRangeAccessor struct{}

func NewBuiltinTrueRangeAccessor() *BuiltinTrueRangeAccessor {
	return &BuiltinTrueRangeAccessor{}
}

/* GenerateLoopValueAccess generates tr calculation at loop offset */
func (a *BuiltinTrueRangeAccessor) GenerateLoopValueAccess(loopVar string) string {
	return fmt.Sprintf(
		"func() float64 { "+
			"barIdx := ctx.BarIndex-%s; "+
			"if barIdx < 1 { return ctx.Data[barIdx].High - ctx.Data[barIdx].Low }; "+
			"prevClose := ctx.Data[barIdx-1].Close; "+
			"currentBar := ctx.Data[barIdx]; "+
			"return math.Max(currentBar.High - currentBar.Low, math.Max(math.Abs(currentBar.High - prevClose), math.Abs(currentBar.Low - prevClose))) "+
			"}()",
		loopVar,
	)
}

/* GenerateInitialValueAccess generates tr calculation for windowed TA initialization */
func (a *BuiltinTrueRangeAccessor) GenerateInitialValueAccess(period int) string {
	return fmt.Sprintf(
		"func() float64 { "+
			"barIdx := ctx.BarIndex-%d; "+
			"if barIdx < 1 { return ctx.Data[barIdx].High - ctx.Data[barIdx].Low }; "+
			"prevClose := ctx.Data[barIdx-1].Close; "+
			"currentBar := ctx.Data[barIdx]; "+
			"return math.Max(currentBar.High - currentBar.Low, math.Max(math.Abs(currentBar.High - prevClose), math.Abs(currentBar.Low - prevClose))) "+
			"}()",
		period-1,
	)
}

/* GenerateCurrentValueAccess generates tr calculation for the current bar */
func (a *BuiltinTrueRangeAccessor) GenerateCurrentValueAccess() string {
	return "func() float64 { " +
		"if ctx.BarIndex < 1 { return ctx.Data[ctx.BarIndex].High - ctx.Data[ctx.BarIndex].Low }; " +
		"prevClose := ctx.Data[ctx.BarIndex-1].Close; " +
		"currentBar := ctx.Data[ctx.BarIndex]; " +
		"return math.Max(currentBar.High - currentBar.Low, math.Max(math.Abs(currentBar.High - prevClose), math.Abs(currentBar.Low - prevClose))) " +
		"}()"
}

/*
GetPreamble returns empty string - tr calculation is self-contained.
*/
func (a *BuiltinTrueRangeAccessor) GetPreamble() string {
	return ""
}
