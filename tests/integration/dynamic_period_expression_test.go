//go:build integration

package integration

import (
	"math"
	"testing"

	"github.com/quant5-lab/runner/tests/util"
)

/* Runtime-dynamic period: bar_index-dependent ternary prevents constant folding */
func TestDynamicPeriod_AllSupportedFunctions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		pine string
	}{
		{
			name: "ta.sma",
			pine: `//@version=5
indicator("Test")
period = bar_index > 50 ? 20 : 10
plot(ta.sma(close, period), title="Result")`,
		},
		{
			name: "ta.ema",
			pine: `//@version=5
indicator("Test")
period = bar_index > 50 ? 20 : 10
plot(ta.ema(close, period), title="Result")`,
		},
		{
			name: "ta.stdev",
			pine: `//@version=5
indicator("Test")
period = bar_index > 50 ? 20 : 10
plot(ta.stdev(close, period), title="Result")`,
		},
		{
			name: "ta.highest",
			pine: `//@version=5
indicator("Test")
period = bar_index > 50 ? 20 : 10
plot(ta.highest(close, period), title="Result")`,
		},
		{
			name: "ta.lowest",
			pine: `//@version=5
indicator("Test")
period = bar_index > 50 ? 20 : 10
plot(ta.lowest(close, period), title="Result")`,
		},
		{
			name: "ta.rsi",
			pine: `//@version=5
indicator("Test")
period = bar_index > 50 ? 20 : 10
plot(ta.rsi(close, period), title="Result")`,
		},
		{
			name: "ta.atr",
			pine: `//@version=5
indicator("Test")
period = bar_index > 50 ? 20 : 10
plot(ta.atr(period), title="Result")`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			exec := util.NewPineExecutor(t)
			output := exec.ExecuteScript(t, "dyn-"+tt.name, tt.pine)

			values := exec.ExtractPlotValues(t, output, "Result")
			if len(values) == 0 {
				t.Fatal("Expected plot output")
			}
			requireSomeFinite(t, values, tt.name)
		})
	}
}

func TestDynamicPeriod_VariablePosition(t *testing.T) {
	t.Parallel()

	pine := `//@version=5
indicator("Test")
period = bar_index > 50 ? 20 : 10
smaVal = ta.sma(close, period)
plot(smaVal, title="Result")`

	exec := util.NewPineExecutor(t)
	output := exec.ExecuteScript(t, "dyn-var-pos", pine)

	values := exec.ExtractPlotValues(t, output, "Result")
	if len(values) == 0 {
		t.Fatal("Expected plot output")
	}
	requireSomeFinite(t, values, "variable-position SMA")
}

func TestDynamicPeriod_ScopeIsolation(t *testing.T) {
	t.Parallel()

	pine := `//@version=5
indicator("Test")
period = bar_index > 50 ? 20 : 10
smaVal = ta.sma(close, period)
rsiVal = ta.rsi(close, period)
atrVal = ta.atr(period)
plot(smaVal + rsiVal + atrVal, title="Result")`

	exec := util.NewPineExecutor(t)
	output := exec.ExecuteScript(t, "dyn-scope-iso", pine)

	values := exec.ExtractPlotValues(t, output, "Result")
	if len(values) == 0 {
		t.Fatal("Expected plot output")
	}
	requireSomeFinite(t, values, "scope-isolated multi-call")
}

func TestDynamicPeriod_RSI_ValueBounds(t *testing.T) {
	t.Parallel()

	pine := `//@version=5
indicator("Test")
period = bar_index > 50 ? 20 : 10
plot(ta.rsi(close, period), title="RSI")`

	exec := util.NewPineExecutor(t)
	output := exec.ExecuteScript(t, "dyn-rsi-bounds", pine)

	values := exec.ExtractPlotValues(t, output, "RSI")
	requireSomeFinite(t, values, "RSI")
	requireBoundedAfterWarmup(t, values, 25, 0.0, 100.0, "RSI")
}

func TestDynamicPeriod_ATR_PositiveValues(t *testing.T) {
	t.Parallel()

	pine := `//@version=5
indicator("Test")
period = bar_index > 50 ? 20 : 10
plot(ta.atr(period), title="ATR")`

	exec := util.NewPineExecutor(t)
	output := exec.ExecuteScript(t, "dyn-atr-positive", pine)

	values := exec.ExtractPlotValues(t, output, "ATR")
	requireSomeFinite(t, values, "ATR")
	requirePositiveAfterWarmup(t, values, 25, "ATR")
}

func TestDynamicPeriod_StrategyContext(t *testing.T) {
	t.Parallel()

	pine := `//@version=5
strategy("Test", overlay=true)
period = bar_index > 50 ? 20 : 10
rsiVal = ta.rsi(close, period)
atrVal = ta.atr(period)
if rsiVal < 30
    strategy.entry("long", strategy.long)
if rsiVal > 70
    strategy.close("long")
plot(rsiVal, title="RSI")
plot(atrVal, title="ATR")`

	exec := util.NewPineExecutor(t)
	output := exec.ExecuteScript(t, "dyn-strategy-ctx", pine)

	rsi := exec.ExtractPlotValues(t, output, "RSI")
	atr := exec.ExtractPlotValues(t, output, "ATR")
	requireSomeFinite(t, rsi, "strategy RSI")
	requireSomeFinite(t, atr, "strategy ATR")
	requireBoundedAfterWarmup(t, rsi, 25, 0.0, 100.0, "strategy RSI")
	requirePositiveAfterWarmup(t, atr, 25, "strategy ATR")
}

func requireSomeFinite(t *testing.T, values []float64, label string) {
	t.Helper()
	for _, v := range values {
		if !math.IsNaN(v) && !math.IsInf(v, 0) && v != 0 {
			return
		}
	}
	t.Errorf("%s: no finite non-zero values in %d data points", label, len(values))
}

func requireBoundedAfterWarmup(t *testing.T, values []float64, warmup int, min, max float64, label string) {
	t.Helper()
	if len(values) <= warmup {
		t.Fatalf("%s: insufficient bars (%d) for warmup %d", label, len(values), warmup)
	}
	for i := warmup; i < len(values); i++ {
		if math.IsNaN(values[i]) {
			continue
		}
		if values[i] < min || values[i] > max {
			t.Errorf("%s[%d] = %f, want [%f, %f]", label, i, values[i], min, max)
		}
	}
}

func requirePositiveAfterWarmup(t *testing.T, values []float64, warmup int, label string) {
	t.Helper()
	if len(values) <= warmup {
		t.Fatalf("%s: insufficient bars (%d) for warmup %d", label, len(values), warmup)
	}
	for i := warmup; i < len(values); i++ {
		if math.IsNaN(values[i]) {
			continue
		}
		if values[i] <= 0 {
			t.Errorf("%s[%d] = %f, want > 0", label, i, values[i])
		}
	}
}
