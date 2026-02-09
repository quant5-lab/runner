package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/codegen/series_naming"
)

/* TestSumIIFEGenerator_ConstantPeriod verifies sum with compile-time period */
func TestSumIIFEGenerator_ConstantPeriod(t *testing.T) {
	gen := &SumIIFEGenerator{}
	accessor := NewArrowFunctionParameterAccessor("src")
	period := NewConstantPeriod(10)

	code := gen.Generate(accessor, period, "test")

	mustContain := []string{
		"func() float64",
		"ctx.BarIndex < 9",
		"for j := 0; j < 10; j++",
		"sum += srcSeries.Get(j)",
		"return sum",
	}
	for _, pattern := range mustContain {
		if !strings.Contains(code, pattern) {
			t.Errorf("Missing pattern %q in:\n%s", pattern, code)
		}
	}

	/* Must NOT divide — sum is not SMA */
	if strings.Contains(code, "sum /") || strings.Contains(code, "sum/") {
		t.Errorf("Sum IIFE must NOT divide, but found division in:\n%s", code)
	}
}

/* TestSumIIFEGenerator_RuntimePeriod verifies sum with runtime parameter period */
func TestSumIIFEGenerator_RuntimePeriod(t *testing.T) {
	gen := &SumIIFEGenerator{}
	accessor := NewArrowFunctionParameterAccessor("src")
	period := NewRuntimePeriod("len")

	code := gen.Generate(accessor, period, "test")

	mustContain := []string{
		"func() float64",
		"int(len)-1",
		"for j := 0; j < int(len); j++",
		"srcSeries.Get(j)",
		"return sum",
	}
	for _, pattern := range mustContain {
		if !strings.Contains(code, pattern) {
			t.Errorf("Missing pattern %q in:\n%s", pattern, code)
		}
	}
}

/* TestSWMAIIFEGenerator_FixedPeriod4 verifies SWMA always uses period=4 */
func TestSWMAIIFEGenerator_FixedPeriod4(t *testing.T) {
	gen := &SWMAIIFEGenerator{}
	accessor := NewArrowFunctionParameterAccessor("src")
	anyPeriod := NewConstantPeriod(99)

	code := gen.Generate(accessor, anyPeriod, "test")

	/* Period=4 warmup: ctx.BarIndex < 3 */
	if !strings.Contains(code, "ctx.BarIndex < 3") {
		t.Errorf("SWMA must use fixed period=4 warmup (ctx.BarIndex < 3), got:\n%s", code)
	}

	/* Must have all 4 weight terms */
	mustContain := []string{
		"srcSeries.Get(3)",
		"srcSeries.Get(2)",
		"srcSeries.Get(1)",
		"srcSeries.Get(0)",
		"1.0/6.0",
		"2.0/6.0",
	}
	for _, pattern := range mustContain {
		if !strings.Contains(code, pattern) {
			t.Errorf("Missing SWMA weight pattern %q in:\n%s", pattern, code)
		}
	}

	/* Must NOT have a loop — SWMA is unrolled 4-element computation */
	if strings.Contains(code, "for j") {
		t.Errorf("SWMA should not use a loop:\n%s", code)
	}
}

/* TestATRIIFEGenerator_StatefulRMA verifies ATR generates RMA-based stateful computation */
func TestATRIIFEGenerator_StatefulRMA(t *testing.T) {
	gen := &ATRIIFEGenerator{namingStrategy: series_naming.NewStatefulIndicatorNamer()}
	accessor := NewArrowFunctionParameterAccessor("ignored")
	period := NewConstantPeriod(14)

	code := gen.Generate(accessor, period, "test")

	mustContain := []string{
		"func() float64",
		"arrowCtx.GetOrCreateSeries(",
		"ctx.Data[idx].High",
		"ctx.Data[idx].Low",
		"ctx.Data[idx-1].Close",
		"math.Max",
		"math.Abs",
		"alpha",
	}
	for _, pattern := range mustContain {
		if !strings.Contains(code, pattern) {
			t.Errorf("Missing ATR pattern %q in:\n%s", pattern, code)
		}
	}

	/* ATR must NOT use srcSeries — it uses OHLC from ctx.Data */
	if strings.Contains(code, "srcSeries") || strings.Contains(code, "ignoredSeries") {
		t.Errorf("ATR must use ctx.Data OHLC, not source parameter:\n%s", code)
	}
}

/* TestATRIIFEGenerator_RuntimePeriod verifies ATR with runtime period */
func TestATRIIFEGenerator_RuntimePeriod(t *testing.T) {
	gen := &ATRIIFEGenerator{namingStrategy: series_naming.NewStatefulIndicatorNamer()}
	accessor := NewArrowFunctionParameterAccessor("ignored")
	period := NewRuntimePeriod("len")

	code := gen.Generate(accessor, period, "test")

	if !strings.Contains(code, "int(len)") {
		t.Errorf("ATR with runtime period must reference int(len):\n%s", code)
	}
	if !strings.Contains(code, "arrowCtx.GetOrCreateSeries(") {
		t.Errorf("ATR must use arrowCtx series for state:\n%s", code)
	}
}

/* TestATRIIFEGenerator_Bar0EdgeCase verifies TR handles bar 0 (no previous close) */
func TestATRIIFEGenerator_Bar0EdgeCase(t *testing.T) {
	gen := &ATRIIFEGenerator{namingStrategy: series_naming.NewStatefulIndicatorNamer()}
	accessor := NewArrowFunctionParameterAccessor("ignored")
	period := NewConstantPeriod(14)

	code := gen.Generate(accessor, period, "test")

	/* TR at bar 0 uses only high-low */
	if !strings.Contains(code, "if idx == 0 { return h - l }") {
		t.Errorf("ATR must handle bar 0 TR (high-low only):\n%s", code)
	}
}

/* TestSumIIFEGenerator_NoDivision confirms sum never divides (distinguishes from SMA) */
func TestSumIIFEGenerator_NoDivision(t *testing.T) {
	gen := &SumIIFEGenerator{}
	smaGen := &SMAIIFEGenerator{}
	accessor := NewArrowFunctionParameterAccessor("src")
	period := NewConstantPeriod(5)

	sumCode := gen.Generate(accessor, period, "test")
	smaCode := smaGen.Generate(accessor, period, "test")

	/* SMA divides, sum does not */
	if !strings.Contains(smaCode, "sum / float64(5)") {
		t.Errorf("SMA should divide by period")
	}
	if strings.Contains(sumCode, "sum / ") || strings.Contains(sumCode, "return sum /") {
		t.Errorf("Sum must NOT divide:\n%s", sumCode)
	}
}

/* TestInlineTAIIFERegistry_NewRegistrations confirms new functions are registered */
func TestInlineTAIIFERegistry_NewRegistrations(t *testing.T) {
	registry := NewInlineTAIIFERegistry()

	newFunctions := []string{
		"sum", "ta.sum", "math.sum",
		"swma", "ta.swma",
		"atr", "ta.atr",
	}

	for _, funcName := range newFunctions {
		if !registry.IsSupported(funcName) {
			t.Errorf("Function %q should be registered in InlineTAIIFERegistry", funcName)
		}
	}
}

/* TestTrueRangeAccessGenerator_LoopAccess verifies TR accessor generates inline computation */
func TestTrueRangeAccessGenerator_LoopAccess(t *testing.T) {
	gen := NewTrueRangeAccessGenerator()

	loopCode := gen.GenerateLoopValueAccess("j")
	if !strings.Contains(loopCode, "ctx.BarIndex - j") {
		t.Errorf("TR loop access must use ctx.BarIndex - j offset:\n%s", loopCode)
	}
	if !strings.Contains(loopCode, "idx == 0") {
		t.Errorf("TR must handle bar 0 edge case:\n%s", loopCode)
	}

	currentCode := gen.GenerateCurrentValueAccess()
	if !strings.Contains(currentCode, "ctx.BarIndex - 0") {
		t.Errorf("TR current access must use offset 0:\n%s", currentCode)
	}

	if gen.GetBaseOffset() != 0 {
		t.Errorf("TR base offset should be 0, got %d", gen.GetBaseOffset())
	}
}

/* TestArrowCtxSeriesAccessor_PreambleAndAccess verifies series-backed accessor pattern */
func TestArrowCtxSeriesAccessor_PreambleAndAccess(t *testing.T) {
	accessor := NewArrowCtxSeriesAccessor("_ta_src_abc", "smaIIFE()")

	preamble := accessor.GetPreamble()
	if !strings.Contains(preamble, "arrowCtx.GetOrCreateSeries(") {
		t.Errorf("Preamble must create arrowCtx series:\n%s", preamble)
	}
	if !strings.Contains(preamble, ".Set(smaIIFE())") {
		t.Errorf("Preamble must store computed value:\n%s", preamble)
	}

	loopAccess := accessor.GenerateLoopValueAccess("j")
	if loopAccess != "_ta_src_abcSeries.Get(j)" {
		t.Errorf("Loop access should read from series, got: %s", loopAccess)
	}

	currentAccess := accessor.GenerateCurrentValueAccess()
	if currentAccess != "_ta_src_abcSeries.Get(0)" {
		t.Errorf("Current access should use Get(0), got: %s", currentAccess)
	}
}
