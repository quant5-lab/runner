//go:build integration

package integration

import (
	"math"
	"strings"
	"testing"

	"github.com/quant5-lab/runner/tests/util"
)

func TestExpressionPositionDispatch_VariableTernaryTA(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		pine        string
		mustContain []string
	}{
		{
			name: "SMA vs EMA ternary",
			pine: `//@version=5
indicator("Test")
mode = input.int(1)
ma = mode == 1 ? ta.sma(close, 10) : ta.ema(close, 10)
plot(ma, title="MA")`,
			mustContain: []string{
				"maSeries.Set",
				".GetCurrent()",
				"series.NewSeries",
			},
		},
		{
			name: "nested ternary three branches",
			pine: `//@version=5
indicator("Test")
mode = input.int(1)
ma = mode == 1 ? ta.sma(close, 10) : mode == 2 ? ta.ema(close, 10) : ta.rma(close, 10)
plot(ma, title="MA")`,
			mustContain: []string{
				"maSeries.Set",
				"func() float64 {",
			},
		},
		{
			name: "TA with arithmetic in branches",
			pine: `//@version=5
indicator("Test")
multiplier = input.float(2.0)
mode = input.bool(true)
result = mode ? ta.sma(close, 10) * multiplier : ta.ema(close, 10) / multiplier
plot(result, title="Result")`,
			mustContain: []string{
				"resultSeries.Set",
				".GetCurrent()",
			},
		},
		{
			name: "TA mixed with literal in branch",
			pine: `//@version=5
indicator("Test")
useMA = input.bool(true)
val = useMA ? ta.sma(close, 10) : 0.0
plot(val, title="Val")`,
			mustContain: []string{
				"valSeries.Set",
				"func() float64 {",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			exec := util.NewPineExecutor(t)
			code, _ := exec.GenerateCode(t, "expr-pos-"+tt.name, tt.pine)

			for _, pattern := range tt.mustContain {
				if !strings.Contains(code, pattern) {
					t.Errorf("Missing pattern: %q", pattern)
				}
			}

			if err := exec.CompileCode(t, code); err != nil {
				t.Fatalf("Compile failed: %v", err)
			}
		})
	}
}

func TestExpressionPositionDispatch_BinaryVariableTA(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		pine string
	}{
		{
			name: "SMA plus EMA",
			pine: `//@version=5
indicator("Test")
combo = ta.sma(close, 10) + ta.ema(close, 10)
plot(combo, title="Combo")`,
		},
		{
			name: "SMA minus literal",
			pine: `//@version=5
indicator("Test")
offset_ma = ta.sma(close, 10) - 5.0
plot(offset_ma, title="Offset")`,
		},
		{
			name: "multiple TA in arithmetic chain",
			pine: `//@version=5
indicator("Test")
spread = ta.sma(close, 10) - ta.ema(close, 10)
normalized = spread / ta.atr(14)
plot(normalized, title="Normalized")`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			exec := util.NewPineExecutor(t)
			code, _ := exec.GenerateCode(t, "bin-ta-"+tt.name, tt.pine)

			if err := exec.CompileCode(t, code); err != nil {
				t.Fatalf("Compile failed: %v", err)
			}
		})
	}
}

/* ta.dev is in TAFunctionRegistry but not InlineTAIIFERegistry */
func TestExpressionPositionDispatch_DevInTernary(t *testing.T) {
	t.Parallel()
	pine := `//@version=5
indicator("Test")
len = 10
vol_high = ta.dev(close, len) > 0.5 ? 1.0 : 0.0
plot(vol_high, title="VolHigh")
`
	exec := util.NewPineExecutor(t)
	code, _ := exec.GenerateCode(t, "dev-ternary", pine)

	if !strings.Contains(code, "vol_highSeries.Set") {
		t.Error("Missing vol_highSeries.Set")
	}
	if !strings.Contains(code, "devSum") {
		t.Error("Missing dev inline calculation")
	}

	if err := exec.CompileCode(t, code); err != nil {
		t.Fatalf("Compile failed: %v", err)
	}
}

func TestExpressionPositionDispatch_Execution(t *testing.T) {
	t.Parallel()
	pine := `//@version=5
indicator("Execution Test")
sma5 = ta.sma(close, 5)
trend_up = close > sma5
ma = trend_up ? ta.sma(close, 3) : ta.ema(close, 3)
plot(ma, title="MA")
plot(sma5, title="SMA5")
`

	baseTime := int64(1704067200)
	prices := []float64{100, 102, 104, 106, 108, 110, 112, 114, 116, 118}

	testData := make([]map[string]interface{}, len(prices))
	for i, price := range prices {
		testData[i] = map[string]interface{}{
			"time":   baseTime + int64(i*3600),
			"open":   price - 1.0,
			"high":   price + 2.0,
			"low":    price - 2.0,
			"close":  price,
			"volume": 1000.0,
		}
	}

	exec := util.NewPineExecutor(t)
	result := exec.ExecuteScriptWithCustomData(t, "expr-pos-exec", pine, testData)

	sma5Values := exec.ExtractPlotValues(t, result, "SMA5")
	maValues := exec.ExtractPlotValues(t, result, "MA")

	if len(sma5Values) != len(prices) {
		t.Fatalf("Expected %d SMA5 values, got %d", len(prices), len(sma5Values))
	}
	if len(maValues) != len(prices) {
		t.Fatalf("Expected %d MA values, got %d", len(prices), len(maValues))
	}

	/* SMA5 at bar 4 (0-indexed): mean(100,102,104,106,108) = 104.0 */
	expectedSMA5Bar4 := 104.0
	if math.Abs(sma5Values[4]-expectedSMA5Bar4) > 0.01 {
		t.Errorf("SMA5[4]: expected %.2f, got %.2f", expectedSMA5Bar4, sma5Values[4])
	}

	/* SMA5 at bar 9: mean(110,112,114,116,118) = 114.0 */
	expectedSMA5Bar9 := 114.0
	if math.Abs(sma5Values[9]-expectedSMA5Bar9) > 0.01 {
		t.Errorf("SMA5[9]: expected %.2f, got %.2f", expectedSMA5Bar9, sma5Values[9])
	}

	/* All bars after warmup: close > sma5 (monotonically rising), so trend_up=true,
	 * MA should equal SMA3. SMA3 at bar 9: mean(114,116,118) = 116.0 */
	expectedMABar9 := 116.0
	if math.Abs(maValues[9]-expectedMABar9) > 0.01 {
		t.Errorf("MA[9]: expected %.2f (SMA3 since trend_up=true), got %.2f", expectedMABar9, maValues[9])
	}
}

func TestExpressionPositionDispatch_ExecutionBranchSwitch(t *testing.T) {
	t.Parallel()
	pine := `//@version=5
indicator("Branch Switch")
sma10 = ta.sma(close, 10)
above = close > sma10
signal = above ? 1.0 : -1.0
plot(signal, title="Signal")
`

	baseTime := int64(1704067200)
	/* 20 bars: first 10 rising (above SMA), last 10 falling (below SMA) */
	prices := []float64{
		100, 102, 104, 106, 108, 110, 112, 114, 116, 118,
		100, 98, 96, 94, 92, 90, 88, 86, 84, 82,
	}

	testData := make([]map[string]interface{}, len(prices))
	for i, price := range prices {
		testData[i] = map[string]interface{}{
			"time":   baseTime + int64(i*3600),
			"open":   price - 1.0,
			"high":   price + 2.0,
			"low":    price - 2.0,
			"close":  price,
			"volume": 1000.0,
		}
	}

	exec := util.NewPineExecutor(t)
	result := exec.ExecuteScriptWithCustomData(t, "branch-switch", pine, testData)

	signalValues := exec.ExtractPlotValues(t, result, "Signal")

	if len(signalValues) != len(prices) {
		t.Fatalf("Expected %d signal values, got %d", len(prices), len(signalValues))
	}

	/* Bar 9 (last rising bar): close=118, SMA10=mean(100..118)=109.0 → above → 1.0 */
	if signalValues[9] != 1.0 {
		t.Errorf("Bar 9: expected 1.0 (close > SMA10), got %v", signalValues[9])
	}

	/* Late falling bars: close well below SMA10 → -1.0 */
	if signalValues[19] != -1.0 {
		t.Errorf("Bar 19: expected -1.0 (close < SMA10), got %v", signalValues[19])
	}

	/* Verify both branches activated across the dataset */
	hasPositive, hasNegative := false, false
	for _, v := range signalValues {
		if v == 1.0 {
			hasPositive = true
		}
		if v == -1.0 {
			hasNegative = true
		}
	}
	if !hasPositive || !hasNegative {
		t.Error("Expected both positive and negative signals across dataset")
	}
}
