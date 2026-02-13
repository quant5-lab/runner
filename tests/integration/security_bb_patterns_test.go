//go:build integration

package integration

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/tests/util"
)

/* TestSecurityBBRealWorldPatterns tests actual security() patterns from production BB strategies */
func TestSecurityBBRealWorldPatterns(t *testing.T) {
	t.Parallel()
	patterns := []struct {
		name        string
		script      string
		description string
	}{
		{
			name: "SMA_Daily_v4",
			script: `//@version=4
strategy("BB SMA Test", overlay=true)
sma_1d_20 = security(syminfo.tickerid, 'D', sma(close, 20))
plot(sma_1d_20, "SMA20 1D")
`,
			description: "Simple Moving Average on daily timeframe (BB7 pattern)",
		},
		{
			name: "SMA_Daily_v5",
			script: `//@version=5
indicator("BB SMA Test", overlay=true)
sma_1d_20 = request.security(syminfo.tickerid, '1D', ta.sma(close, 20))
plot(sma_1d_20, "SMA20 1D")
`,
			description: "Simple Moving Average on daily timeframe (v5 syntax)",
		},
		{
			name: "Multiple_SMA_Daily",
			script: `//@version=4
strategy("BB Multiple SMA", overlay=true)
sma_1d_20 = security(syminfo.tickerid, 'D', sma(close, 20))
sma_1d_50 = security(syminfo.tickerid, 'D', sma(close, 50))
sma_1d_200 = security(syminfo.tickerid, 'D', sma(close, 200))
plot(sma_1d_20, "SMA20")
plot(sma_1d_50, "SMA50")
plot(sma_1d_200, "SMA200")
`,
			description: "Multiple SMA calculations (BB7/8/9 pattern)",
		},
		{
			name: "Open_Daily_Lookahead",
			script: `//@version=4
strategy("BB Open Test", overlay=true)
open_1d = security(syminfo.tickerid, "D", open, lookahead=barmerge.lookahead_on)
plot(open_1d, "Open 1D")
`,
			description: "Daily open with lookahead (BB7 pattern)",
		},
		{
			name: "BB_Basis_SMA",
			script: `//@version=4
strategy("BB Basis Test", overlay=true)
bb_1d_basis = security(syminfo.tickerid, "1D", sma(close, 46))
plot(bb_1d_basis, "BB Basis")
`,
			description: "Bollinger Band basis calculation (BB8 pattern)",
		},
		{
			name: "Close_Simple",
			script: `//@version=5
indicator("Close Test", overlay=true)
close_1d = request.security(syminfo.tickerid, "1D", close)
plot(close_1d, "Close 1D")
`,
			description: "Simple close value from daily timeframe",
		},
		{
			name: "EMA_Daily",
			script: `//@version=5
indicator("EMA Test", overlay=true)
ema_1d_10 = request.security(syminfo.tickerid, "1D", ta.ema(close, 10))
plot(ema_1d_10, "EMA10 1D")
`,
			description: "Exponential Moving Average on daily timeframe",
		},
	}

	exec := util.NewPineExecutor(t)
	for _, tc := range patterns {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			generatedCode, _ := exec.GenerateCode(t, tc.name, tc.script)
			if err := exec.CompileCode(t, generatedCode); err != nil {
				t.Fatalf("'%s' failed: %s - %v", tc.name, tc.description, err)
			}
			t.Logf("'%s' - %s", tc.name, tc.description)
		})
	}

	t.Logf("All %d BB strategy patterns compiled successfully", len(patterns))
}

/* TestSecurityStdevWorkaround tests BB strategy pattern with stdev */
func TestSecurityStdevWorkaround(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name   string
		script string
		status string
	}{
		{
			name: "Stdev_Simple_Works",
			script: `//@version=4
strategy("Stdev Works", overlay=true)
dev_1d = security(syminfo.tickerid, "1D", stdev(close, 20))
plot(dev_1d, "Stdev")
`,
			status: "WORKS - simple stdev call",
		},
		{
			name: "Stdev_PreMultiplied_Works",
			script: `//@version=4
strategy("Stdev Workaround", overlay=true)
// Workaround: calculate with multiplier outside security()
bbstdev = 0.35
dev_1d = security(syminfo.tickerid, "1D", stdev(close, 20))
bb_dev = bbstdev * dev_1d
plot(bb_dev, "BB Dev")
`,
			status: "WORKS - multiplication outside security()",
		},
	}

	exec := util.NewPineExecutor(t)
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			generatedCode, _ := exec.GenerateCode(t, tc.name, tc.script)
			if err := exec.CompileCode(t, generatedCode); err != nil {
				t.Fatalf("Test failed: %s - %v", tc.status, err)
			}
			t.Logf("%s: %s", tc.name, tc.status)
		})
	}
}

/* TestSecurityLongTermStability tests patterns for regression safety */
func TestSecurityLongTermStability(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name        string
		script      string
		criticalFor string
	}{
		{
			name: "SMA_Warmup_Handling",
			script: `//@version=5
indicator("SMA Warmup", overlay=true)
// With 20-period SMA, first 19 bars should be NaN
sma20_1d = request.security(syminfo.tickerid, "1D", ta.sma(close, 20))
plot(sma20_1d, "SMA20")
`,
			criticalFor: "NaN handling with insufficient warmup period",
		},
		{
			name: "Multiple_Timeframes",
			script: `//@version=4
strategy("Multi TF", overlay=true)
close_1d = security(syminfo.tickerid, "1D", close)
close_1w = security(syminfo.tickerid, "1W", close)
plot(close_1d, "Daily")
plot(close_1w, "Weekly")
`,
			criticalFor: "Multiple security() calls with different timeframes",
		},
		{
			name: "Mixed_v4_v5_Syntax",
			script: `//@version=4
strategy("Mixed Syntax", overlay=true)
// v4 syntax
sma_1d = security(syminfo.tickerid, "1D", sma(close, 20))
plot(sma_1d, "SMA")
`,
			criticalFor: "Pine v4 to v5 migration compatibility",
		},
	}

	exec := util.NewPineExecutor(t)
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			generatedCode, _ := exec.GenerateCode(t, tc.name, tc.script)
			if err := exec.CompileCode(t, generatedCode); err != nil {
				t.Fatalf("REGRESSION: %s failed - critical for: %s - %v", tc.name, tc.criticalFor, err)
			}
			t.Logf("Stability check passed: %s", tc.criticalFor)
		})
	}
}

/* TestSecurityInlineTA_Validation validates inline TA code generation */
func TestSecurityInlineTA_Validation(t *testing.T) {
	t.Parallel()
	pineScript := `//@version=5
indicator("Inline TA Check", overlay=true)
sma20_1d = request.security(syminfo.tickerid, "1D", ta.sma(close, 20))
ema10_1d = request.security(syminfo.tickerid, "1D", ta.ema(close, 10))
plot(sma20_1d, "SMA")
plot(ema10_1d, "EMA")
`

	exec := util.NewPineExecutor(t)
	generatedCode, _ := exec.GenerateCode(t, "inline_ta_check", pineScript)

	if !strings.Contains(generatedCode, "ta.sma") {
		t.Error("Expected inline SMA generation (not runtime lookup)")
	}

	if !strings.Contains(generatedCode, "ta.ema") {
		t.Error("Expected inline EMA generation (not runtime lookup)")
	}

	if !strings.Contains(generatedCode, "secBarEvaluator") {
		t.Error("Expected StreamingBarEvaluator for security() expressions")
	}

	if !strings.Contains(generatedCode, "EvaluateAtBar") {
		t.Error("Expected EvaluateAtBar() call for streaming evaluation")
	}

	if !strings.Contains(generatedCode, "math.NaN()") {
		t.Error("Expected NaN handling for insufficient warmup")
	}

	t.Log("Inline TA code generation validated")
}
