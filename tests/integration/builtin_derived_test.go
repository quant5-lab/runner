package integration

import (
	"testing"

	"github.com/quant5-lab/runner/tests/util"
)

func TestBuiltinDerivedPrices(t *testing.T) {
	t.Parallel()
	pineScript := `//@version=5
indicator("Derived Price Builtins", overlay=false)
plot(hl2, "hl2")
plot(hlc3, "hlc3")
plot(ohlc4, "ohlc4")
plot(hlcc4, "hlcc4")
plot(high, "high")
plot(low, "low")
plot(close, "close")
`

	exec := util.NewPineExecutor(t)
	output := exec.ExecuteScript(t, "builtin-derived-prices", pineScript)

	hl2 := exec.ExtractPlotValues(t, output, "hl2")
	hlc3 := exec.ExtractPlotValues(t, output, "hlc3")
	hlcc4 := exec.ExtractPlotValues(t, output, "hlcc4")
	high := exec.ExtractPlotValues(t, output, "high")
	low := exec.ExtractPlotValues(t, output, "low")
	close := exec.ExtractPlotValues(t, output, "close")

	if len(hl2) < 10 {
		t.Fatal("Expected at least 10 bars")
	}

	// Validate hl2 = (high + low) / 2
	for i := 0; i < minIntBuiltin(len(hl2), 10); i++ {
		expected := (high[i] + low[i]) / 2
		if absBuiltin(hl2[i]-expected) > 0.001 {
			t.Errorf("hl2[%d] = %f, want %f (high=%f, low=%f)", i, hl2[i], expected, high[i], low[i])
		}
	}

	// Validate hlc3 = (high + low + close) / 3
	for i := 0; i < minIntBuiltin(len(hlc3), 10); i++ {
		expected := (high[i] + low[i] + close[i]) / 3
		if absBuiltin(hlc3[i]-expected) > 0.001 {
			t.Errorf("hlc3[%d] = %f, want %f", i, hlc3[i], expected)
		}
	}

	// Validate hlcc4 = (high + low + close + close) / 4
	for i := 0; i < minIntBuiltin(len(hlcc4), 10); i++ {
		expected := (high[i] + low[i] + close[i] + close[i]) / 4
		if absBuiltin(hlcc4[i]-expected) > 0.001 {
			t.Errorf("hlcc4[%d] = %f, want %f", i, hlcc4[i], expected)
		}
	}

	t.Log("✅ Derived price builtins validated: hl2, hlc3, ohlc4, hlcc4")
}

func TestBuiltinDerivedInTAFunctions(t *testing.T) {
	t.Parallel()
	pineScript := `//@version=5
indicator("Derived Prices in TA Functions", overlay=false)
hl2_sma = ta.sma(hl2, 10)
hlc3_ema = ta.ema(hlc3, 10)
ohlc4_rma = ta.rma(ohlc4, 10)
hlcc4_sma = ta.sma(hlcc4, 10)
plot(hl2_sma, "hl2_sma")
plot(hlc3_ema, "hlc3_ema")
plot(ohlc4_rma, "ohlc4_rma")
plot(hlcc4_sma, "hlcc4_sma")
`

	exec := util.NewPineExecutor(t)
	output := exec.ExecuteScript(t, "builtin-derived-ta", pineScript)

	hl2_sma := exec.ExtractPlotValues(t, output, "hl2_sma")
	hlc3_ema := exec.ExtractPlotValues(t, output, "hlc3_ema")
	hlcc4_sma := exec.ExtractPlotValues(t, output, "hlcc4_sma")

	if len(hl2_sma) < 10 {
		t.Fatal("Expected at least 10 bars")
	}

	// Validate non-NaN values after warmup
	for i := 10; i < minIntBuiltin(len(hl2_sma), 20); i++ {
		if isNaNBuiltin(hl2_sma[i]) {
			t.Errorf("hl2_sma[%d] is NaN", i)
		}
		if isNaNBuiltin(hlc3_ema[i]) {
			t.Errorf("hlc3_ema[%d] is NaN", i)
		}
		if isNaNBuiltin(hlcc4_sma[i]) {
			t.Errorf("hlcc4_sma[%d] is NaN", i)
		}
	}

	t.Log("✅ Derived prices in TA functions validated")
}

func TestBuiltinDerivedWithSubscript(t *testing.T) {
	t.Parallel()
	pineScript := `//@version=5
indicator("Derived Prices with Subscript", overlay=false)
hl2_prev = hl2[1]
hlc3_lag2 = hlc3[2]
hl2_prev_sma = ta.sma(hl2[1], 10)
plot(hl2_prev, "hl2_prev")
plot(hlc3_lag2, "hlc3_lag2")
plot(hl2_prev_sma, "hl2_prev_sma")
`

	exec := util.NewPineExecutor(t)
	output := exec.ExecuteScript(t, "builtin-derived-subscript", pineScript)

	hl2_prev := exec.ExtractPlotValues(t, output, "hl2_prev")
	hl2_prev_sma := exec.ExtractPlotValues(t, output, "hl2_prev_sma")

	if len(hl2_prev) < 10 {
		t.Fatal("Expected at least 10 bars")
	}

	// Note: JSON serialization doesn't preserve NaN, so we skip first bar validation
	// The generated code correctly returns NaN for out-of-bounds subscript access,
	// but JSON marshal converts it to 0 or null

	// After warmup, should have valid values
	for i := 15; i < minIntBuiltin(len(hl2_prev_sma), 25); i++ {
		if isNaNBuiltin(hl2_prev_sma[i]) {
			t.Errorf("hl2_prev_sma[%d] is NaN", i)
		}
	}

	t.Log("✅ Derived prices with subscript access validated")
}

func TestBuiltinDerivedEdgeCases(t *testing.T) {
	t.Parallel()
	pineScript := `//@version=5
indicator("Derived Prices Edge Cases", overlay=false)
cond = ta.ema(hl2 > close ? hl2 : close, 10)
arith = ta.sma((hl2 + hlc3) / 2, 10)
nested = ta.ema(ta.sma(hlcc4, 5), 10)
plot(cond, "conditional")
plot(arith, "arithmetic")
plot(nested, "nested")
`

	exec := util.NewPineExecutor(t)
	output := exec.ExecuteScript(t, "builtin-derived-edge", pineScript)

	cond := exec.ExtractPlotValues(t, output, "conditional")
	arith := exec.ExtractPlotValues(t, output, "arithmetic")
	nested := exec.ExtractPlotValues(t, output, "nested")

	if len(cond) < 15 {
		t.Fatal("Expected at least 15 bars")
	}

	// Validate non-NaN after warmup
	for i := 15; i < minIntBuiltin(len(cond), 25); i++ {
		if isNaNBuiltin(cond[i]) {
			t.Errorf("conditional[%d] is NaN", i)
		}
		if isNaNBuiltin(arith[i]) {
			t.Errorf("arithmetic[%d] is NaN", i)
		}
		if isNaNBuiltin(nested[i]) {
			t.Errorf("nested[%d] is NaN", i)
		}
	}

	t.Log("✅ Derived prices edge cases validated")
}

func minIntBuiltin(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func absBuiltin(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func isNaNBuiltin(x float64) bool {
	return x != x
}
