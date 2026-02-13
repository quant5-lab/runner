//go:build integration

package integration

import (
	"math"
	"testing"

	"github.com/quant5-lab/runner/tests/util"
)

/*
Validates unary negation in variable assignments produces correct runtime values through full pipeline.

	Regression: codegen previously routed UnaryExpression through generateExpression, embedding
	statement-level newlines into expressions — causing Go compilation failures.
*/
func TestUnaryNegationAssignment(t *testing.T) {
	t.Parallel()

	baseTime := int64(1700000000)
	bars := []map[string]interface{}{
		{"time": baseTime, "open": 100.0, "high": 105.0, "low": 95.0, "close": 102.0, "volume": 1000.0},
		{"time": baseTime + 3600, "open": 102.0, "high": 108.0, "low": 99.0, "close": 106.0, "volume": 1100.0},
		{"time": baseTime + 7200, "open": 106.0, "high": 110.0, "low": 103.0, "close": 104.0, "volume": 1200.0},
		{"time": baseTime + 10800, "open": 104.0, "high": 107.0, "low": 100.0, "close": 101.0, "volume": 1300.0},
	}

	tests := []struct {
		name     string
		pine     string
		plot     string
		expected []float64
	}{
		{
			name: "negated literal",
			pine: `//@version=5
indicator("Test")
x = -1.5
plot(x, "Result")`,
			plot:     "Result",
			expected: []float64{-1.5, -1.5, -1.5, -1.5},
		},
		{
			name: "negated series identifier",
			pine: `//@version=5
indicator("Test")
x = -close
plot(x, "Result")`,
			plot:     "Result",
			expected: []float64{-102.0, -106.0, -104.0, -101.0},
		},
		{
			name: "negated grouped binary",
			pine: `//@version=5
indicator("Test")
x = -(close - open)
plot(x, "Result")`,
			plot:     "Result",
			expected: []float64{-2.0, -4.0, 2.0, 3.0},
		},
		{
			name: "negation within arithmetic",
			pine: `//@version=5
indicator("Test")
x = 10.0 + (-close)
plot(x, "Result")`,
			plot:     "Result",
			expected: []float64{-92.0, -96.0, -94.0, -91.0},
		},
	}

	exec := util.NewPineExecutor(t)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			output := exec.ExecuteScriptWithCustomData(t, "unary-neg-assign", tc.pine, bars)
			values := exec.ExtractPlotValues(t, output, tc.plot)

			if len(values) < len(tc.expected) {
				t.Fatalf("got %d values, need %d", len(values), len(tc.expected))
			}

			for i, want := range tc.expected {
				if math.Abs(values[i]-want) > 1e-9 {
					t.Errorf("bar %d: got %f, want %f", i, values[i], want)
				}
			}
		})
	}
}
