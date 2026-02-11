package integration

import (
	"math"
	"testing"

	"github.com/quant5-lab/runner/tests/util"
)

func TestUDFCompoundArguments(t *testing.T) {
	t.Parallel()
	baseTime := int64(1700000000)
	testData := []map[string]interface{}{
		{"time": baseTime, "open": 100.0, "high": 105.0, "low": 95.0, "close": 102.0, "volume": 1000.0},
		{"time": baseTime + 3600, "open": 102.0, "high": 108.0, "low": 99.0, "close": 106.0, "volume": 1100.0},
		{"time": baseTime + 7200, "open": 106.0, "high": 110.0, "low": 103.0, "close": 104.0, "volume": 1200.0},
		{"time": baseTime + 10800, "open": 104.0, "high": 107.0, "low": 100.0, "close": 101.0, "volume": 1300.0},
		{"time": baseTime + 14400, "open": 101.0, "high": 106.0, "low": 98.0, "close": 105.0, "volume": 1400.0},
		{"time": baseTime + 18000, "open": 105.0, "high": 112.0, "low": 103.0, "close": 110.0, "volume": 1500.0},
		{"time": baseTime + 21600, "open": 110.0, "high": 115.0, "low": 108.0, "close": 113.0, "volume": 1600.0},
		{"time": baseTime + 25200, "open": 113.0, "high": 116.0, "low": 109.0, "close": 111.0, "volume": 1700.0},
	}

	tests := []struct {
		name     string
		pine     string
		plotName string
		validate func(t *testing.T, values []float64)
	}{
		{
			name: "unary negation as UDF argument",
			pine: `//@version=5
indicator("Unary Neg Arg")
negate(v) => -v
result = negate(-close)
plot(result, "Result")
`,
			plotName: "Result",
			validate: func(t *testing.T, values []float64) {
				if values[0] != 102.0 {
					t.Errorf("negate(-close) should equal close, got %f want 102.0", values[0])
				}
			},
		},
		{
			name: "binary with negated coefficient as UDF argument",
			pine: `//@version=5
indicator("Neg Coeff Arg")
scale(v) => v
result = scale(close * -1.0)
plot(result, "Result")
`,
			plotName: "Result",
			validate: func(t *testing.T, values []float64) {
				if values[0] != -102.0 {
					t.Errorf("scale(close * -1.0) should equal -close, got %f want -102.0", values[0])
				}
			},
		},
		{
			name: "conditional ternary as UDF argument",
			pine: `//@version=5
indicator("Ternary Arg")
passthrough(v) => v
result = passthrough(close > open ? 1.0 : 0.0)
plot(result, "Result")
`,
			plotName: "Result",
			validate: func(t *testing.T, values []float64) {
				/* bar 0: close=102 > open=100 → 1.0; bar 3: close=101 < open=104 → 0.0 */
				if values[0] != 1.0 {
					t.Errorf("bar 0: close > open, expected 1.0 got %f", values[0])
				}
				if values[3] != 0.0 {
					t.Errorf("bar 3: close < open, expected 0.0 got %f", values[3])
				}
			},
		},
		{
			name: "logical and as UDF argument",
			pine: `//@version=5
indicator("Logical Arg")
identity(v) => v
result = identity(close > 100.0 and close < 106.0)
plot(result, "Result")
`,
			plotName: "Result",
			validate: func(t *testing.T, values []float64) {
				/* bar 0: 102 > 100 and 102 < 106 → 1.0; bar 1: 106 > 100 and 106 < 106 → 0.0 */
				if values[0] != 1.0 {
					t.Errorf("bar 0: 100 < 102 < 106, expected 1.0 got %f", values[0])
				}
				if values[1] != 0.0 {
					t.Errorf("bar 1: 106 not < 106, expected 0.0 got %f", values[1])
				}
			},
		},
		{
			name: "nested unary and binary in UDF argument",
			pine: `//@version=5
indicator("Nested Arg")
passthrough(v) => v
result = passthrough(-(close - open))
plot(result, "Result")
`,
			plotName: "Result",
			validate: func(t *testing.T, values []float64) {
				/* bar 0: -(102 - 100) = -2.0 */
				if values[0] != -2.0 {
					t.Errorf("-(close - open) bar 0: got %f want -2.0", values[0])
				}
			},
		},
	}

	exec := util.NewPineExecutor(t)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := exec.ExecuteScriptWithCustomData(t, "udf-compound-args", tt.pine, testData)
			values := exec.ExtractPlotValues(t, result, tt.plotName)

			if len(values) < len(testData) {
				t.Fatalf("expected %d values, got %d", len(testData), len(values))
			}

			for _, v := range values {
				if math.IsInf(v, 0) {
					t.Fatal("output contains Inf — runtime error in generated code")
				}
			}

			tt.validate(t, values)
		})
	}
}
