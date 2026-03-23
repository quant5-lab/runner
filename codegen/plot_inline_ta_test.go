package codegen

import (
	"fmt"
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

/* TestPlotInlineTA_ATR_PeriodVariants verifies that ta.atr generates a named temp
 * series, advances it with Series.Next(), and emits a collector.Add for each
 * representative period — minimal (1), short (2), standard (14), large (100).
 */
func TestPlotInlineTA_ATR_PeriodVariants(t *testing.T) {
	for _, period := range []float64{1, 2, 14, 100} {
		period := period
		t.Run(fmt.Sprintf("period_%d", int(period)), func(t *testing.T) {
			code := generatePlotExpression(t, TACallPeriodOnly("atr", period))

			NewCodeVerifier(code, t).MustContain(
				"ta_atr_",
				"Series.Get(0)",
				"collector.Add",
				"Series.Next()",
			)
		})
	}
}

/* TestPlotInlineTA_ATR_RMAStructure verifies that ta.atr generates the canonical
 * Pine RMA-of-TR structure across all three phases, with TR using handle_na=true
 * (bar 0 returns high-low, enabling a fully-valid SMA seed window).
 */
func TestPlotInlineTA_ATR_RMAStructure(t *testing.T) {
	code := generatePlotExpression(t, TACallPeriodOnly("atr", 14))

	NewCodeVerifier(code, t).
		MustContain("ctx.BarIndex < 13").
		MustContain("ctx.BarIndex == 13").
		MustContain("initialValue := _sma_accumulator / float64(14)").
		MustContain("previousValue :=").
		MustContain("alpha := 1.0 / float64(14)").
		MustContain("newValue := alpha*currentSource + (1-alpha)*previousValue").
		MustContain("if idx == 0 { return h - l }").
		MustNotContain("highSeries.GetCurrent()").
		MustNotContain("lowSeries.GetCurrent()")
}

/* TestPlotInlineTA_ATR_NoIIFEGeneration verifies that ta.atr is not wrapped in an
 * outer arrow-function IIFE.  The TR computation within it uses inline
 * func() float64 lambdas, but the ATR indicator itself must be top-level stateful code.
 */
func TestPlotInlineTA_ATR_NoIIFEGeneration(t *testing.T) {
	code := generatePlotExpression(t, TACallPeriodOnly("atr", 14))

	NewCodeVerifier(code, t).
		MustNotContain("return func")
}
