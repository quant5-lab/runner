package codegen

import (
	"testing"
)

func TestPlotInlineTA_SMA(t *testing.T) {
	code := generatePlotExpression(t, TACall("sma", Ident("close"), 20))

	NewCodeVerifier(code, t).MustContain(
		"collector.Add",
		"ctx.BarIndex < 19",
		"sum += ctx.Data[ctx.BarIndex-j].Close",
	)
}

func TestPlotInlineTA_MathMax(t *testing.T) {
	code := generatePlotExpression(t, MathCall("max", Ident("high"), Ident("low")))

	NewCodeVerifier(code, t).MustContain(
		"collector.Add",
		"math.Max",
	)
}
