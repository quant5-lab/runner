package codegen

import "fmt"

func buildHighLowSMA(rangeName string, period PeriodExpression, ctx StatefulIndicatorContext) string {
	ind := NewCodeIndenter()
	ind.IncreaseIndent()

	code := ind.Line(fmt.Sprintf("if ctx.BarIndex < %s {", warmupBarExpr(period)))
	ind.IncreaseIndent()
	code += ind.Line(ctx.GenerateSeriesUpdate(rangeName, "math.NaN()"))
	ind.DecreaseIndent()
	code += ind.Line("} else {")
	ind.IncreaseIndent()
	code += ind.Line("_hlSum := 0.0")
	code += ind.Line(fmt.Sprintf("for _j := 0; _j < %s; _j++ {", periodAsIntExpr(period)))
	ind.IncreaseIndent()
	code += ind.Line("_hlSum += ctx.Data[ctx.BarIndex-_j].High - ctx.Data[ctx.BarIndex-_j].Low")
	ind.DecreaseIndent()
	code += ind.Line("}")
	code += ind.Line(ctx.GenerateSeriesUpdate(rangeName, fmt.Sprintf("_hlSum / float64(%s)", periodAsIntExpr(period))))
	ind.DecreaseIndent()
	code += ind.Line("}")
	ind.DecreaseIndent()

	return code
}

func warmupBarExpr(period PeriodExpression) string {
	if period.IsConstant() {
		return fmt.Sprintf("%d", period.AsInt()-1)
	}
	return fmt.Sprintf("%s-1", period.AsIntCast())
}

func periodAsIntExpr(period PeriodExpression) string {
	if period.IsConstant() {
		return fmt.Sprintf("%d", period.AsInt())
	}
	return period.AsIntCast()
}
