package integration

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/tests/util"
)

/* TestSecurityArrowFunction_OHLCV validates security() OHLCV field access inside arrow functions */
func TestSecurityArrowFunction_OHLCV(t *testing.T) {
	t.Parallel()
	fields := []struct {
		name  string
		field string
	}{
		{"close", "close"},
		{"open", "open"},
		{"high", "high"},
		{"low", "low"},
		{"volume", "volume"},
	}

	exec := util.NewPineExecutor(t)
	for _, tc := range fields {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			script := `//@version=5
indicator("Arrow OHLCV ` + tc.name + `", overlay=true)

getDaily() =>
    request.security(syminfo.tickerid, "1D", ` + tc.field + `)

result = getDaily()
plot(result, "Daily ` + tc.name + `")
`
			generatedCode, _ := exec.GenerateCode(t, "arrow_ohlcv_"+tc.name, script)

			if !strings.Contains(generatedCode, "arrowCtx.SecurityContexts") {
				t.Error("Expected ArrowContext security bridge access")
			}

			if !strings.Contains(generatedCode, "secCtx.Data[secBarIdx]") {
				t.Error("Expected direct OHLCV field access on security context data")
			}

			if err := exec.CompileCode(t, generatedCode); err != nil {
				t.Fatalf("Compilation failed for field %q: %v", tc.field, err)
			}
		})
	}
}

/* TestSecurityArrowFunction_TACall validates security() with TA indicators inside arrow functions */
func TestSecurityArrowFunction_TACall(t *testing.T) {
	t.Parallel()
	patterns := []struct {
		name   string
		script string
	}{
		{
			name: "SMA",
			script: `//@version=5
indicator("Arrow TA SMA", overlay=true)

getDailySMA(len) =>
    request.security(syminfo.tickerid, "1D", ta.sma(close, len))

result = getDailySMA(20)
plot(result, "Daily SMA")
`,
		},
		{
			name: "EMA",
			script: `//@version=5
indicator("Arrow TA EMA", overlay=true)

getDailyEMA(len) =>
    request.security(syminfo.tickerid, "1D", ta.ema(close, len))

result = getDailyEMA(14)
plot(result, "Daily EMA")
`,
		},
	}

	exec := util.NewPineExecutor(t)
	for _, tc := range patterns {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			generatedCode, _ := exec.GenerateCode(t, "arrow_ta_"+tc.name, tc.script)

			if !strings.Contains(generatedCode, "arrowCtx.SecurityContexts") {
				t.Error("Expected ArrowContext security bridge access")
			}

			if !strings.Contains(generatedCode, "EvaluateAtBar") {
				t.Error("Expected streaming evaluation for TA call")
			}

			if !strings.Contains(generatedCode, "ast.CallExpression") {
				t.Error("Expected serialized AST call expression")
			}

			if err := exec.CompileCode(t, generatedCode); err != nil {
				t.Fatalf("Compilation failed for %s: %v", tc.name, err)
			}
		})
	}
}

/* TestSecurityArrowFunction_V4Syntax validates v4 security() (no request. prefix) in arrow functions */
func TestSecurityArrowFunction_V4Syntax(t *testing.T) {
	t.Parallel()
	script := `//@version=4
strategy("Arrow v4 Security", overlay=true)

getDailyClose() =>
    security(syminfo.tickerid, "D", close)

getDailySMA() =>
    security(syminfo.tickerid, "1D", sma(close, 20))

dc = getDailyClose()
dsma = getDailySMA()

longCond = close > dsma
if longCond
    strategy.entry("Long", strategy.long)
if close < dsma
    strategy.close("Long")

plot(dc, "Daily Close")
plot(dsma, "Daily SMA")
`
	exec := util.NewPineExecutor(t)
	generatedCode, _ := exec.GenerateCode(t, "arrow_v4_security", script)

	if !strings.Contains(generatedCode, "arrowCtx.SecurityContexts") {
		t.Error("Expected ArrowContext security bridge in v4 code")
	}

	if err := exec.CompileCode(t, generatedCode); err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}
}

/* TestSecurityArrowFunction_MultipleCallSites validates multiple arrow functions with security() */
func TestSecurityArrowFunction_MultipleCallSites(t *testing.T) {
	t.Parallel()
	script := `//@version=5
indicator("Arrow Multi Security", overlay=true)

getDailyClose() =>
    request.security(syminfo.tickerid, "1D", close)

getDailyHigh() =>
    request.security(syminfo.tickerid, "1D", high)

getDailyLow() =>
    request.security(syminfo.tickerid, "1D", low)

getDailySMA(len) =>
    request.security(syminfo.tickerid, "1D", ta.sma(close, len))

dc = getDailyClose()
dh = getDailyHigh()
dl = getDailyLow()
dsma = getDailySMA(20)

plot(dc, "Daily Close")
plot(dh, "Daily High")
plot(dl, "Daily Low")
plot(dsma, "Daily SMA")
`
	exec := util.NewPineExecutor(t)
	generatedCode, _ := exec.GenerateCode(t, "arrow_multi_security", script)

	fnCount := strings.Count(generatedCode, "arrowCtx.SecurityContexts")
	if fnCount < 4 {
		t.Errorf("Expected at least 4 ArrowContext security bridge accesses, got %d", fnCount)
	}

	if err := exec.CompileCode(t, generatedCode); err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}
}

/* TestSecurityArrowFunction_MixedWithMainBody validates arrow security coexists with main-body security */
func TestSecurityArrowFunction_MixedWithMainBody(t *testing.T) {
	t.Parallel()
	script := `//@version=5
indicator("Arrow Mixed Security", overlay=true)

getDailySMA(len) =>
    request.security(syminfo.tickerid, "1D", ta.sma(close, len))

mainBodySecurity = request.security(syminfo.tickerid, "1D", close)
arrowSecurity = getDailySMA(20)

plot(mainBodySecurity, "Main Body")
plot(arrowSecurity, "Arrow Function")
`
	exec := util.NewPineExecutor(t)
	generatedCode, _ := exec.GenerateCode(t, "arrow_mixed_security", script)

	if !strings.Contains(generatedCode, "arrowCtx.SecurityContexts") {
		t.Error("Expected ArrowContext bridge for arrow function security")
	}

	if err := exec.CompileCode(t, generatedCode); err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}
}

/* TestSecurityArrowFunction_NaNGuards validates fallback to NaN for missing security data */
func TestSecurityArrowFunction_NaNGuards(t *testing.T) {
	t.Parallel()
	script := `//@version=5
indicator("Arrow NaN Guards", overlay=true)

getDailyClose() =>
    request.security(syminfo.tickerid, "1D", close)

result = getDailyClose()
plot(result, "Daily Close")
`
	exec := util.NewPineExecutor(t)
	generatedCode, _ := exec.GenerateCode(t, "arrow_nan_guards", script)

	nanCount := strings.Count(generatedCode, "math.NaN()")
	if nanCount < 3 {
		t.Errorf("Expected at least 3 NaN guard returns (context missing, mapper missing, bar negative), got %d", nanCount)
	}

	if err := exec.CompileCode(t, generatedCode); err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}
}

/* TestSecurityArrowFunction_SecurityBridge validates ArrowContext receives security bridge wiring */
func TestSecurityArrowFunction_SecurityBridge(t *testing.T) {
	t.Parallel()
	script := `//@version=5
indicator("Arrow Bridge", overlay=true)

getDailyClose() =>
    request.security(syminfo.tickerid, "1D", close)

result = getDailyClose()
plot(result, "Daily Close")
`
	exec := util.NewPineExecutor(t)
	generatedCode, _ := exec.GenerateCode(t, "arrow_bridge", script)

	if !strings.Contains(generatedCode, "SecurityContexts") || !strings.Contains(generatedCode, "SetBarMapper") {
		t.Error("Expected SecurityContexts assignment and SetBarMapper calls in hoisted bridge code")
	}

	if !strings.Contains(generatedCode, "FindDailyBarIndex") {
		t.Error("Expected FindDailyBarIndex call for bar mapping")
	}

	if err := exec.CompileCode(t, generatedCode); err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}
}
