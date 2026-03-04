package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

/* TrHandler: TR = max(H-L, |H-prevClose|, |L-prevClose|)
 * handleNA=false → NaN on bar 0  (ta.tr variable semantics)
 * handleNA=true  → H-L on bar 0  (ATR seed / ta.tr(true) semantics) */
type TrHandler struct{}

func (h *TrHandler) CanHandle(funcName string) bool {
	return funcName == "ta.tr" || funcName == "tr"
}

func (h *TrHandler) GenerateCode(g *generator, varName string, call *ast.CallExpression) (string, error) {
	handleNA := extractHandleNAArg(call)
	expr := generateTRExpression(handleNA)
	return g.ind() + fmt.Sprintf("%sSeries.Set(%s)\n", varName, expr), nil
}

func extractHandleNAArg(call *ast.CallExpression) bool {
	if len(call.Arguments) < 1 {
		return false
	}
	lit, ok := call.Arguments[0].(*ast.Literal)
	if !ok {
		return false
	}
	boolVal, ok := lit.Value.(bool)
	if !ok {
		return false
	}
	return boolVal
}

func generateTRExpression(handleNA bool) string {
	if handleNA {
		return "func() float64 { " +
			"bar := ctx.Data[ctx.BarIndex]; " +
			"if ctx.BarIndex < 1 { return bar.High - bar.Low }; " +
			"prevClose := ctx.Data[ctx.BarIndex-1].Close; " +
			"return math.Max(bar.High - bar.Low, math.Max(math.Abs(bar.High - prevClose), math.Abs(bar.Low - prevClose))) " +
			"}()"
	}
	return "func() float64 { " +
		"if ctx.BarIndex < 1 { return math.NaN() }; " +
		"bar := ctx.Data[ctx.BarIndex]; " +
		"prevClose := ctx.Data[ctx.BarIndex-1].Close; " +
		"return math.Max(bar.High - bar.Low, math.Max(math.Abs(bar.High - prevClose), math.Abs(bar.Low - prevClose))) " +
		"}()"
}
