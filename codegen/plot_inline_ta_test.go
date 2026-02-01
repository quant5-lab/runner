package codegen

import (
	"testing"
)

func TestPlotInlineTA_SMA(t *testing.T) {
	code := generatePlotExpression(t, TACall("sma", Ident("close"), 20))

	NewCodeVerifier(code, t).MustContain(
		"collector.Add",
		"ctx.BarIndex < 19",
		"sum += closeSeries.Get(j)",
	)
}

func TestPlotInlineTA_MathMax(t *testing.T) {
	code := generatePlotExpression(t, MathCall("max", Ident("high"), Ident("low")))

	NewCodeVerifier(code, t).MustContain(
		"collector.Add",
		"math.Max",
	)
}

func TestPlotInlineTA_ATR_BasicPeriod(t *testing.T) {
	code := generatePlotExpression(t, TACallPeriodOnly("atr", 14))

	NewCodeVerifier(code, t).MustContain(
		"ta_atr_",
		"Series.Get(0)",
		"collector.Add",
	)
}

func TestPlotInlineTA_ATR_ShortPeriod(t *testing.T) {
	code := generatePlotExpression(t, TACallPeriodOnly("atr", 2))

	NewCodeVerifier(code, t).MustContain(
		"ta_atr_",
		"Series.Get(0)",
		"collector.Add",
	)
}

func TestPlotInlineTA_ATR_MinimalPeriod(t *testing.T) {
	code := generatePlotExpression(t, TACallPeriodOnly("atr", 1))

	NewCodeVerifier(code, t).MustContain(
		"ta_atr_",
		"Series.Get(0)",
		"collector.Add",
	)
}

func TestPlotInlineTA_ATR_LargePeriod(t *testing.T) {
	code := generatePlotExpression(t, TACallPeriodOnly("atr", 100))

	NewCodeVerifier(code, t).MustContain(
		"ta_atr_",
		"Series.Get(0)",
		"collector.Add",
	)
}

func TestPlotInlineTA_ATR_GeneratesTempVariable(t *testing.T) {
	code := generatePlotExpression(t, TACallPeriodOnly("atr", 14))

	NewCodeVerifier(code, t).
		MustContain("ta_atr_").
		MustContain("Series.Next()")
}

func TestPlotInlineTA_ATR_NoIIFEGeneration(t *testing.T) {
	code := generatePlotExpression(t, TACallPeriodOnly("atr", 14))

	NewCodeVerifier(code, t).
		MustNotContain("func()").
		MustNotContain("return func")
}
