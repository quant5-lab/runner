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

/* Variable-position: period defined then consumed in separate statement */
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

func requireSomeFinite(t *testing.T, values []float64, label string) {
	t.Helper()
	for _, v := range values {
		if !math.IsNaN(v) && !math.IsInf(v, 0) && v != 0 {
			return
		}
	}
	t.Errorf("%s: no finite non-zero values in %d data points", label, len(values))
}
