//go:build integration

package integration

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/tests/util"
)

/* Full pipeline: Pine parse → AST → codegen → Go compile for tuple security */
func TestSecurityTuple_Compilation(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		script string
	}{
		{
			name: "OHLCV pair",
			script: `//@version=5
indicator("Test", overlay=true)
[c, v] = request.security(syminfo.tickerid, "D", [close, volume])
plot(c)
plot(v)`,
		},
		{
			name: "full OHLCV",
			script: `//@version=5
indicator("Test", overlay=true)
[o, h, l, c, v] = request.security(syminfo.tickerid, "D", [open, high, low, close, volume])
plot(o)
plot(h)
plot(l)
plot(c)
plot(v)`,
		},
		{
			name: "complex TA expressions",
			script: `//@version=5
indicator("Test", overlay=true)
[ema_d, sma_d] = request.security(syminfo.tickerid, "D", [ta.ema(close, 10), ta.sma(close, 20)])
plot(ema_d)
plot(sma_d)`,
		},
		{
			name: "mixed OHLCV and complex",
			script: `//@version=5
indicator("Test", overlay=true)
[c, ema_d, v] = request.security(syminfo.tickerid, "D", [close, ta.ema(close, 10), volume])
plot(c)
plot(ema_d)
plot(v)`,
		},
		{
			name: "arithmetic expressions",
			script: `//@version=5
indicator("Test", overlay=true)
[spread, mid] = request.security(syminfo.tickerid, "D", [high - low, (high + low) / 2])
plot(spread)
plot(mid)`,
		},
		{
			name: "v4 syntax with lookahead",
			script: `//@version=4
strategy("Test", overlay=true)
[c, o] = security(syminfo.tickerid, "D", [close, open], lookahead=barmerge.lookahead_on)
plot(c)
plot(o)`,
		},
		{
			name: "multiple tuple calls same strategy",
			script: `//@version=5
indicator("Test", overlay=true)
[c1, v1] = request.security(syminfo.tickerid, "D", [close, volume])
[ema_d, sma_d] = request.security(syminfo.tickerid, "D", [ta.ema(close, 10), ta.sma(close, 20)])
[h1, l1] = request.security(syminfo.tickerid, "W", [high, low])
plot(c1)
plot(v1)
plot(ema_d)
plot(sma_d)
plot(h1)
plot(l1)`,
		},
		{
			name: "tuple alongside single-var security",
			script: `//@version=5
indicator("Test", overlay=true)
daily_close = request.security(syminfo.tickerid, "D", close)
[ema_d, sma_d] = request.security(syminfo.tickerid, "D", [ta.ema(close, 10), ta.sma(close, 20)])
daily_vol = request.security(syminfo.tickerid, "D", volume)
plot(daily_close)
plot(ema_d)
plot(sma_d)
plot(daily_vol)`,
		},
	}

	exec := util.NewPineExecutor(t)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			generatedCode, _ := exec.GenerateCode(t, tt.name, tt.script)
			if err := exec.CompileCode(t, generatedCode); err != nil {
				t.Fatalf("compilation failed: %v", err)
			}
		})
	}
}

/* Verify generated code structure for key patterns */
func TestSecurityTuple_CodeStructure(t *testing.T) {
	t.Parallel()
	exec := util.NewPineExecutor(t)

	t.Run("OHLCV uses direct field access", func(t *testing.T) {
		t.Parallel()
		script := `//@version=5
indicator("Test", overlay=true)
[c, v] = request.security(syminfo.tickerid, "D", [close, volume])
plot(c)
plot(v)`

		code, _ := exec.GenerateCode(t, "ohlcv_structure", script)

		if !strings.Contains(code, "secCtx.Data[secBarIdx].Close") {
			t.Error("expected direct Close field access for OHLCV")
		}
		if !strings.Contains(code, "secCtx.Data[secBarIdx].Volume") {
			t.Error("expected direct Volume field access for OHLCV")
		}
	})

	t.Run("complex expressions use streaming evaluator", func(t *testing.T) {
		t.Parallel()
		script := `//@version=5
indicator("Test", overlay=true)
[ema_d, sma_d] = request.security(syminfo.tickerid, "D", [ta.ema(close, 10), ta.sma(close, 20)])
plot(ema_d)
plot(sma_d)`

		code, _ := exec.GenerateCode(t, "complex_structure", script)

		if !strings.Contains(code, "secBarEvaluator") {
			t.Error("expected StreamingBarEvaluator for complex expressions")
		}
		if !strings.Contains(code, "EvaluateAtBar") {
			t.Error("expected EvaluateAtBar calls for complex expressions")
		}

		evalCount := strings.Count(code, "secValue, err := secBarEvaluator.EvaluateAtBar(")
		if evalCount != 2 {
			t.Errorf("expected 2 EvaluateAtBar calls (scope-isolated), got %d", evalCount)
		}
	})

	t.Run("NaN fallbacks for all variables", func(t *testing.T) {
		t.Parallel()
		script := `//@version=5
indicator("Test", overlay=true)
[o, h, l] = request.security(syminfo.tickerid, "D", [open, high, low])
plot(o)
plot(h)
plot(l)`

		code, _ := exec.GenerateCode(t, "nan_fallbacks", script)

		for _, v := range []string{"oSeries.Set(math.NaN())", "hSeries.Set(math.NaN())", "lSeries.Set(math.NaN())"} {
			if !strings.Contains(code, v) {
				t.Errorf("expected NaN fallback %q", v)
			}
		}
	})
}
