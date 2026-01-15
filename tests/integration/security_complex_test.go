package integration

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/tests/util"
)

/* TestSecurityTACombination tests inline TA combination inside security() */
func TestSecurityTACombination(t *testing.T) {
	pineScript := `//@version=5
indicator("TA Combo Security", overlay=true)
combined = request.security(syminfo.tickerid, "1D", ta.sma(close, 20) + ta.ema(close, 10))
plot(combined, "Combined", color=color.blue)
`

	exec := util.NewPineExecutor(t)
	generatedCode, _ := exec.GenerateCode(t, "ta_combo", pineScript)

	if !strings.Contains(generatedCode, "ta.sma") {
		t.Error("Expected inline SMA generation in security context")
	}

	if !strings.Contains(generatedCode, "ta.ema") {
		t.Error("Expected inline EMA generation in security context")
	}

	if err := exec.CompileCode(t, generatedCode); err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}

	t.Log("TA combination security() compiled successfully")
}

/* TestSecurityArithmeticExpression tests arithmetic expressions inside security() */
func TestSecurityArithmeticExpression(t *testing.T) {
	pineScript := `//@version=5
indicator("Arithmetic Security", overlay=true)
volatility = request.security(syminfo.tickerid, "1D", (high - low) / close * 100)
plot(volatility, "Volatility %", color=color.red)
`

	exec := util.NewPineExecutor(t)
	generatedCode, _ := exec.GenerateCode(t, "arithmetic", pineScript)

	if !strings.Contains(generatedCode, "secBarEvaluator") {
		t.Error("Expected StreamingBarEvaluator for complex arithmetic expression")
	}

	if !strings.Contains(generatedCode, "EvaluateAtBar") {
		t.Error("Expected EvaluateAtBar() call for expression evaluation")
	}

	if !strings.Contains(generatedCode, "BinaryExpression") {
		t.Error("Expected BinaryExpression AST node in serialized expression")
	}

	if err := exec.CompileCode(t, generatedCode); err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}

	t.Log("Arithmetic expression security() compiled successfully")
}

/* TestSecurityBBStrategy7Patterns tests real-world patterns from bb-strategy-7-rus.pine */
func TestSecurityBBStrategy7Patterns(t *testing.T) {
	patterns := []struct {
		name   string
		script string
	}{
		{
			name: "SMA on daily timeframe",
			script: `//@version=4
strategy("BB7-SMA", overlay=true)
sma_1d_20 = security(syminfo.tickerid, 'D', sma(close, 20))
plot(sma_1d_20)`,
		},
		{
			name: "ATR on daily timeframe",
			script: `//@version=4
strategy("BB7-ATR", overlay=true)
atr_1d = security(syminfo.tickerid, "1D", atr(14))
plot(atr_1d)`,
		},
		{
			name: "Open with lookahead",
			script: `//@version=4
strategy("BB7-Open", overlay=true)
open_1d = security(syminfo.tickerid, "D", open, lookahead=barmerge.lookahead_on)
plot(open_1d)`,
		},
	}

	exec := util.NewPineExecutor(t)
	for _, tc := range patterns {
		t.Run(tc.name, func(t *testing.T) {
			generatedCode, _ := exec.GenerateCode(t, tc.name, tc.script)
			if err := exec.CompileCode(t, generatedCode); err != nil {
				t.Fatalf("Pattern '%s' failed: %v", tc.name, err)
			}
			t.Logf("BB7 pattern '%s' compiled successfully", tc.name)
		})
	}
}

/* TestSecurityBBStrategy8Patterns tests real-world patterns from bb-strategy-8-rus.pine */
func TestSecurityBBStrategy8Patterns(t *testing.T) {
	patterns := []struct {
		name   string
		script string
	}{
		{
			name: "BB basis with SMA",
			script: `//@version=4
strategy("BB8-Basis", overlay=true)
bb_1d_basis = security(syminfo.tickerid, "1D", sma(close, 46))
plot(bb_1d_basis)`,
		},
		{
			name: "BB deviation with stdev multiplication",
			script: `//@version=4
strategy("BB8-Dev", overlay=true)
bb_1d_dev = security(syminfo.tickerid, "1D", 0.35 * stdev(close, 46))
plot(bb_1d_dev)`,
		},
	}

	exec := util.NewPineExecutor(t)
	for _, tc := range patterns {
		t.Run(tc.name, func(t *testing.T) {
			generatedCode, _ := exec.GenerateCode(t, tc.name, tc.script)
			if err := exec.CompileCode(t, generatedCode); err != nil {
				t.Fatalf("Pattern '%s' failed: %v", tc.name, err)
			}
			t.Logf("BB8 pattern '%s' compiled successfully", tc.name)
		})
	}
}

/* TestSecurityStability_RegressionSuite comprehensive regression test suite */
func TestSecurityStability_RegressionSuite(t *testing.T) {
	testCases := []struct {
		name        string
		script      string
		description string
	}{
		{
			name: "TA_Combo_Add",
			script: `//@version=5
indicator("Test")
result = request.security(syminfo.tickerid, "1D", ta.sma(close, 20) + ta.ema(close, 10))
plot(result)`,
			description: "SMA + EMA combination",
		},
		{
			name: "TA_Combo_Subtract",
			script: `//@version=5
indicator("Test")
result = request.security(syminfo.tickerid, "1D", ta.sma(close, 20) - ta.ema(close, 10))
plot(result)`,
			description: "SMA - EMA subtraction",
		},
		{
			name: "TA_Combo_Multiply",
			script: `//@version=5
indicator("Test")
result = request.security(syminfo.tickerid, "1D", ta.sma(close, 20) * 1.5)
plot(result)`,
			description: "SMA multiplication by constant",
		},
		{
			name: "Arithmetic_HighLow",
			script: `//@version=5
indicator("Test")
result = request.security(syminfo.tickerid, "1D", (high - low) / close * 100)
plot(result)`,
			description: "High-Low volatility percentage",
		},
		{
			name: "Arithmetic_OHLC",
			script: `//@version=5
indicator("Test")
result = request.security(syminfo.tickerid, "1D", (open + high + low + close) / 4)
plot(result)`,
			description: "OHLC average",
		},
		{
			name: "Stdev_Multiplication",
			script: `//@version=4
strategy("Test", overlay=true)
dev = security(syminfo.tickerid, "1D", 2.0 * stdev(close, 20))
plot(dev)`,
			description: "Stdev with constant multiplication (BB pattern)",
		},
	}

	exec := util.NewPineExecutor(t)
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			generatedCode, _ := exec.GenerateCode(t, tc.name, tc.script)
			if err := exec.CompileCode(t, generatedCode); err != nil {
				t.Fatalf("'%s' failed: %s - %v", tc.name, tc.description, err)
			}
			t.Logf("'%s' - %s", tc.name, tc.description)
		})
	}

	t.Logf("All %d regression test cases passed", len(testCases))
}

/* TestSecurityNaN_Handling ensures NaN values are handled correctly */
func TestSecurityNaN_Handling(t *testing.T) {
	pineScript := `//@version=5
indicator("NaN Test", overlay=true)
sma20 = request.security(syminfo.tickerid, "1D", ta.sma(close, 20))
plot(sma20, "SMA20")`

	exec := util.NewPineExecutor(t)
	generatedCode, _ := exec.GenerateCode(t, "nan_test", pineScript)

	if !strings.Contains(generatedCode, "math.NaN()") {
		t.Error("Expected NaN handling in generated code for insufficient warmup")
	}

	if err := exec.CompileCode(t, generatedCode); err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}

	t.Log("NaN handling compiled successfully")
}
