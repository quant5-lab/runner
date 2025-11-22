package codegen

import "fmt"

// WarmupChecker determines if enough bars are available for calculation
type WarmupChecker struct {
	period int
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
