package security

import (
	"math"
	"testing"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/runtime/context"
)

func makeKCWCall(source string, period, mult float64, useTrueRange *bool) *ast.CallExpression {
	args := []ast.Expression{
		&ast.Identifier{Name: source},
		&ast.Literal{Value: period},
		&ast.Literal{Value: mult},
	}
	if useTrueRange != nil {
		args = append(args, &ast.Literal{Value: *useTrueRange})
	}
	return &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "ta"},
			Property: &ast.Identifier{Name: "kcw"},
		},
		Arguments: args,
	}
}

func boolPtr(v bool) *bool { return &v }

func makeKCWContext(bars int, startClose, step, highOffset, lowOffset float64) *context.Context {
	data := make([]context.OHLCV, bars)
	for i := 0; i < bars; i++ {
		c := startClose + float64(i)*step
		data[i] = context.OHLCV{
			Open:   c,
			High:   c + highOffset,
			Low:    c - lowOffset,
			Close:  c,
			Volume: 1000,
		}
	}
	return &context.Context{Data: data}
}

/*
TestStreamingBarEvaluator_KCW_WarmupReturnsNaN verifies that KCW returns NaN
for every bar before the period-1 threshold, regardless of useTrueRange mode.
*/
func TestStreamingBarEvaluator_KCW_WarmupReturnsNaN(t *testing.T) {
	const period = 10
	ctx := makeKCWContext(period+5, 100, 1, 2, 1)

	for _, useTrueRange := range []*bool{nil, boolPtr(true), boolPtr(false)} {
		label := "default"
		if useTrueRange != nil {
			if *useTrueRange {
				label = "useTrueRange=true"
			} else {
				label = "useTrueRange=false"
			}
		}

		t.Run(label, func(t *testing.T) {
			evaluator := NewStreamingBarEvaluator()
			call := makeKCWCall("close", period, 1.5, useTrueRange)

			for barIdx := 0; barIdx < period-1; barIdx++ {
				val, err := evaluator.EvaluateAtBar(call, ctx, barIdx)
				if err != nil {
					t.Fatalf("bar %d: unexpected error: %v", barIdx, err)
				}
				if !math.IsNaN(val) {
					t.Errorf("bar %d: expected NaN during warmup, got %f", barIdx, val)
				}
			}
		})
	}
}

/*
TestStreamingBarEvaluator_KCW_ProducesPositiveResultAfterWarmup confirms that
KCW yields a positive, finite value once enough bars have accumulated, for both
useTrueRange modes. It does not assert an exact value — only that the result
is in the valid range (0, ∞) for positive-priced assets.
*/
func TestStreamingBarEvaluator_KCW_ProducesPositiveResultAfterWarmup(t *testing.T) {
	const period = 5
	ctx := makeKCWContext(period+5, 100, 1, 3, 2)

	for _, tt := range []struct {
		label        string
		useTrueRange *bool
	}{
		{"default_mode", nil},
		{"use_true_range", boolPtr(true)},
		{"use_high_low_sma", boolPtr(false)},
	} {
		t.Run(tt.label, func(t *testing.T) {
			evaluator := NewStreamingBarEvaluator()
			call := makeKCWCall("close", period, 1.5, tt.useTrueRange)

			lastBarIdx := len(ctx.Data) - 1
			val, err := evaluator.EvaluateAtBar(call, ctx, lastBarIdx)
			if err != nil {
				t.Fatalf("EvaluateAtBar error: %v", err)
			}
			if math.IsNaN(val) || math.IsInf(val, 0) {
				t.Errorf("expected finite positive result, got %f", val)
			}
			if val <= 0 {
				t.Errorf("KCW must be positive for positive-priced assets, got %f", val)
			}
		})
	}
}

/*
TestStreamingBarEvaluator_KCW_ModesProduceDifferentResultsOnGappingData shows that
useTrueRange=true and useTrueRange=false yield DIFFERENT values when price gaps
are present (prevClose far from next bar's H/L). This confirms the two code paths
diverge as intended.
*/
func TestStreamingBarEvaluator_KCW_ModesProduceDifferentResultsOnGappingData(t *testing.T) {
	// Create bars with significant overnight gaps: prevClose << next bar's Low
	bars := []context.OHLCV{
		{Open: 100, High: 102, Low: 98, Close: 100},
		{Open: 115, High: 120, Low: 112, Close: 118}, // gap up: prevClose=100, Low=112 → TR >> H-L
		{Open: 116, High: 121, Low: 113, Close: 119},
		{Open: 117, High: 122, Low: 114, Close: 120},
		{Open: 118, High: 123, Low: 115, Close: 121},
		{Open: 119, High: 124, Low: 116, Close: 122},
		{Open: 120, High: 125, Low: 117, Close: 123},
	}
	ctx := &context.Context{Data: bars}

	const period = 3
	const mult = 1.5
	lastBar := len(bars) - 1

	evalATR := NewStreamingBarEvaluator()
	callATR := makeKCWCall("close", period, mult, boolPtr(true))
	valATR, err := evalATR.EvaluateAtBar(callATR, ctx, lastBar)
	if err != nil {
		t.Fatalf("ATR mode error: %v", err)
	}

	evalHL := NewStreamingBarEvaluator()
	callHL := makeKCWCall("close", period, mult, boolPtr(false))
	valHL, err := evalHL.EvaluateAtBar(callHL, ctx, lastBar)
	if err != nil {
		t.Fatalf("H-L mode error: %v", err)
	}

	if math.IsNaN(valATR) || math.IsNaN(valHL) {
		t.Fatalf("neither mode should return NaN after warmup: ATR=%f HL=%f", valATR, valHL)
	}
	// On gapping data, ATR captures gap size, so KCW(ATR) > KCW(H-L)
	if valATR <= valHL {
		t.Errorf("on gapping data, KCW(useTrueRange=true)=%.4f should exceed KCW(useTrueRange=false)=%.4f", valATR, valHL)
	}
}

/*
TestStreamingBarEvaluator_KCW_ModesAgreeOnGaplessData verifies that when there
are no overnight gaps (TR = H-L exactly on every bar), both range-measurement
modes produce the same result, since ATR ≈ SMA(H-L) when gaps are absent.
*/
func TestStreamingBarEvaluator_KCW_ModesAgreeOnGaplessData(t *testing.T) {
	// Gapless bars: prevClose == next bar's open == midpoint of H+L, no gaps.
	// Use constant H-L spread so RMA converges to the same value as SMA.
	bars := make([]context.OHLCV, 30)
	for i := range bars {
		c := 100.0 + float64(i)
		bars[i] = context.OHLCV{High: c + 5, Low: c - 5, Close: c, Open: c}
	}
	ctx := &context.Context{Data: bars}

	const period = 5
	const mult = 1.5
	lastBar := len(bars) - 1

	evalATR := NewStreamingBarEvaluator()
	valATR, _ := evalATR.EvaluateAtBar(makeKCWCall("close", period, mult, boolPtr(true)), ctx, lastBar)

	evalHL := NewStreamingBarEvaluator()
	valHL, _ := evalHL.EvaluateAtBar(makeKCWCall("close", period, mult, boolPtr(false)), ctx, lastBar)

	if math.IsNaN(valATR) || math.IsNaN(valHL) {
		t.Fatalf("both modes should produce valid results: ATR=%f HL=%f", valATR, valHL)
	}
	// After sufficient warmup, RMA converges toward SMA on constant-spread data.
	// Allow 5% tolerance for RMA convergence lag.
	if math.Abs(valATR-valHL)/valHL > 0.05 {
		t.Errorf("on gapless constant-spread data, ATR=%.4f and H-L SMA=%.4f should be within 5%%", valATR, valHL)
	}
}

/*
TestStreamingBarEvaluator_KCW_DefaultEqualsUseTrueRangeTrue verifies that
omitting the 4th argument produces the same numerical result as explicitly
passing useTrueRange=true, confirming the default is ATR mode.
*/
func TestStreamingBarEvaluator_KCW_DefaultEqualsUseTrueRangeTrue(t *testing.T) {
	ctx := makeKCWContext(20, 100, 1, 3, 2)
	const period = 5
	const mult = 1.5
	lastBar := len(ctx.Data) - 1

	evalDefault := NewStreamingBarEvaluator()
	valDefault, err := evalDefault.EvaluateAtBar(makeKCWCall("close", period, mult, nil), ctx, lastBar)
	if err != nil {
		t.Fatalf("default: error = %v", err)
	}

	evalExplicit := NewStreamingBarEvaluator()
	valExplicit, err := evalExplicit.EvaluateAtBar(makeKCWCall("close", period, mult, boolPtr(true)), ctx, lastBar)
	if err != nil {
		t.Fatalf("explicit true: error = %v", err)
	}

	if math.Abs(valDefault-valExplicit) > 1e-9 {
		t.Errorf("default (%.6f) != explicit useTrueRange=true (%.6f): default must equal ATR mode", valDefault, valExplicit)
	}
}

/*
TestStreamingBarEvaluator_ComputeHighLowSMA_Arithmetic verifies the exact
numerical output of the SMA(high-low) helper at a concrete bar.

Given constant H-L spread = 10 and period = 3, the SMA must equal 10.
Given a linearly increasing H-L spread, the SMA must equal the mean of
the last `period` H-L values.
*/
func TestStreamingBarEvaluator_ComputeHighLowSMA_Arithmetic(t *testing.T) {
	tests := []struct {
		name     string
		bars     []context.OHLCV
		barIdx   int
		period   int
		expected float64
	}{
		{
			name: "constant_spread_period_3",
			bars: []context.OHLCV{
				{High: 110, Low: 100},
				{High: 120, Low: 110},
				{High: 130, Low: 120},
			},
			barIdx:   2,
			period:   3,
			expected: 10.0,
		},
		{
			name: "varying_spread_mean_of_last_3",
			// H-L: 8, 10, 12  → mean = 10
			bars: []context.OHLCV{
				{High: 108, Low: 100},
				{High: 120, Low: 110},
				{High: 132, Low: 120},
			},
			barIdx:   2,
			period:   3,
			expected: 10.0,
		},
		{
			name: "period_2_uses_last_2_bars_only",
			// H-L: 8, 10, 12  → period=2 → mean of last 2 = (10+12)/2 = 11
			bars: []context.OHLCV{
				{High: 108, Low: 100},
				{High: 120, Low: 110},
				{High: 132, Low: 120},
			},
			barIdx:   2,
			period:   2,
			expected: 11.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewStreamingBarEvaluator()
			ctx := &context.Context{Data: tt.bars}
			got, err := e.computeHighLowSMAAtBar(ctx, tt.barIdx, tt.period)
			if err != nil {
				t.Fatalf("computeHighLowSMAAtBar error: %v", err)
			}
			if math.Abs(got-tt.expected) > 1e-9 {
				t.Errorf("SMA(H-L) = %.6f, want %.6f", got, tt.expected)
			}
		})
	}
}

/*
TestStreamingBarEvaluator_ComputeHighLowSMA_WarmupReturnsNaN confirms that
SMA(H-L) returns NaN until barIdx >= period-1, matching the warmup semantics
of all other ta.* indicators.
*/
func TestStreamingBarEvaluator_ComputeHighLowSMA_WarmupReturnsNaN(t *testing.T) {
	bars := make([]context.OHLCV, 10)
	for i := range bars {
		bars[i] = context.OHLCV{High: 110, Low: 100}
	}
	ctx := &context.Context{Data: bars}
	e := NewStreamingBarEvaluator()

	const period = 5
	for barIdx := 0; barIdx < period-1; barIdx++ {
		val, err := e.computeHighLowSMAAtBar(ctx, barIdx, period)
		if err != nil {
			t.Fatalf("bar %d: unexpected error: %v", barIdx, err)
		}
		if !math.IsNaN(val) {
			t.Errorf("bar %d: expected NaN (warmup), got %f", barIdx, val)
		}
	}

	// First valid bar
	val, err := e.computeHighLowSMAAtBar(ctx, period-1, period)
	if err != nil {
		t.Fatalf("first valid bar: unexpected error: %v", err)
	}
	if math.IsNaN(val) {
		t.Errorf("bar %d: expected finite value, got NaN", period-1)
	}
}
