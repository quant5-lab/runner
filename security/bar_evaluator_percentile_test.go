package security

import (
	"math"
	"testing"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/runtime/context"
)

func makePercentileCall(funcName, source string, period, pct float64) *ast.CallExpression {
	return &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "ta"},
			Property: &ast.Identifier{Name: funcName},
		},
		Arguments: []ast.Expression{
			&ast.Identifier{Name: source},
			&ast.Literal{Value: period},
			&ast.Literal{Value: pct},
		},
	}
}

func makePercentrankCall(source string, period float64) *ast.CallExpression {
	return &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "ta"},
			Property: &ast.Identifier{Name: "percentrank"},
		},
		Arguments: []ast.Expression{
			&ast.Identifier{Name: source},
			&ast.Literal{Value: period},
		},
	}
}

/* =================== Percentrank =================== */

func TestStreamingBarEvaluator_PercentrankWarmupYieldsNaN(t *testing.T) {
	ctx := &context.Context{
		Data: []context.OHLCV{
			{Close: 10}, {Close: 20}, {Close: 30}, {Close: 40}, {Close: 50},
		},
	}
	evaluator := NewStreamingBarEvaluator()
	call := makePercentrankCall("close", 4)

	for barIdx := 0; barIdx < 3; barIdx++ {
		v, err := evaluator.EvaluateAtBar(call, ctx, barIdx)
		if err != nil {
			t.Fatalf("bar %d: unexpected error: %v", barIdx, err)
		}
		if !math.IsNaN(v) {
			t.Errorf("bar %d: expected NaN (warmup), got %f", barIdx, v)
		}
	}
}

func TestStreamingBarEvaluator_PercentrankAscendingSource(t *testing.T) {
	/* In a monotonically increasing window the current bar is always the maximum.
	   All period-1 prior values are strictly less → rank = (period-1)/period * 100. */
	ctx := &context.Context{
		Data: []context.OHLCV{
			{Close: 1}, {Close: 2}, {Close: 3}, {Close: 4}, {Close: 5}, {Close: 6},
		},
	}
	evaluator := NewStreamingBarEvaluator()
	period := 4
	call := makePercentrankCall("close", float64(period))
	expected := float64(period-1) / float64(period) * 100.0

	for barIdx := period - 1; barIdx < len(ctx.Data); barIdx++ {
		v, err := evaluator.EvaluateAtBar(call, ctx, barIdx)
		if err != nil {
			t.Fatalf("bar %d: unexpected error: %v", barIdx, err)
		}
		if math.Abs(v-expected) > 0.0001 {
			t.Errorf("bar %d: expected %.4f, got %.4f", barIdx, expected, v)
		}
	}
}

func TestStreamingBarEvaluator_PercentrankConstantSourceIsZero(t *testing.T) {
	ctx := &context.Context{
		Data: []context.OHLCV{
			{Close: 5}, {Close: 5}, {Close: 5}, {Close: 5}, {Close: 5},
		},
	}
	evaluator := NewStreamingBarEvaluator()
	call := makePercentrankCall("close", 3)

	for barIdx := 2; barIdx < len(ctx.Data); barIdx++ {
		v, err := evaluator.EvaluateAtBar(call, ctx, barIdx)
		if err != nil {
			t.Fatalf("bar %d: unexpected error: %v", barIdx, err)
		}
		if math.Abs(v-0.0) > 0.0001 {
			t.Errorf("bar %d: constant source must yield 0 rank, got %f", barIdx, v)
		}
	}
}

func TestStreamingBarEvaluator_PercentrankRangeIsZeroToHundred(t *testing.T) {
	ctx := &context.Context{
		Data: []context.OHLCV{
			{Close: 3}, {Close: 1}, {Close: 4}, {Close: 1}, {Close: 5},
			{Close: 9}, {Close: 2}, {Close: 6}, {Close: 5}, {Close: 3},
		},
	}
	evaluator := NewStreamingBarEvaluator()
	call := makePercentrankCall("close", 5)

	for barIdx := 4; barIdx < len(ctx.Data); barIdx++ {
		v, err := evaluator.EvaluateAtBar(call, ctx, barIdx)
		if err != nil {
			t.Fatalf("bar %d: unexpected error: %v", barIdx, err)
		}
		if v < -0.0001 || v > 100.0001 {
			t.Errorf("bar %d: percentrank %f outside [0, 100]", barIdx, v)
		}
	}
}

/* =================== PercentileNearestRank =================== */

func TestStreamingBarEvaluator_PercentileNearestRankWarmupYieldsNaN(t *testing.T) {
	ctx := &context.Context{
		Data: []context.OHLCV{
			{Close: 10}, {Close: 20}, {Close: 30}, {Close: 40},
		},
	}
	evaluator := NewStreamingBarEvaluator()
	call := makePercentileCall("percentile_nearest_rank", "close", 4, 50)

	for barIdx := 0; barIdx < 3; barIdx++ {
		v, err := evaluator.EvaluateAtBar(call, ctx, barIdx)
		if err != nil {
			t.Fatalf("bar %d: unexpected error: %v", barIdx, err)
		}
		if !math.IsNaN(v) {
			t.Errorf("bar %d: expected NaN (warmup), got %f", barIdx, v)
		}
	}
}

func TestStreamingBarEvaluator_PercentileNearestRankSortOrderIndependence(t *testing.T) {
	/* The window [30, 10, 50, 20, 40] unsorted must produce the same percentile
	   result as if the input arrived pre-sorted. Verifies the sort is actually applied. */
	ctx := &context.Context{
		Data: []context.OHLCV{
			{Close: 30}, {Close: 10}, {Close: 50}, {Close: 20}, {Close: 40},
		},
	}
	evaluator := NewStreamingBarEvaluator()
	call := makePercentileCall("percentile_nearest_rank", "close", 5, 50)

	/* period=5, pct=50: idx=ceil(0.5*5)-1=ceil(2.5)-1=3-1=2
	   Sorted window: [10,20,30,40,50] → sorted[2]=30 */
	v, err := evaluator.EvaluateAtBar(call, ctx, 4)
	if err != nil {
		t.Fatalf("EvaluateAtBar error: %v", err)
	}
	if math.Abs(v-30.0) > 0.0001 {
		t.Errorf("pct=50 of [30,10,50,20,40]: expected 30.0, got %f", v)
	}
}

func TestStreamingBarEvaluator_PercentileNearestRankPct100ReturnsMax(t *testing.T) {
	ctx := &context.Context{
		Data: []context.OHLCV{
			{Close: 3}, {Close: 1}, {Close: 4}, {Close: 1}, {Close: 5},
		},
	}
	evaluator := NewStreamingBarEvaluator()
	call := makePercentileCall("percentile_nearest_rank", "close", 5, 100)

	v, err := evaluator.EvaluateAtBar(call, ctx, 4)
	if err != nil {
		t.Fatalf("EvaluateAtBar error: %v", err)
	}
	if math.Abs(v-5.0) > 0.0001 {
		t.Errorf("pct=100 must return window max 5.0, got %f", v)
	}
}

func TestStreamingBarEvaluator_PercentileNearestRankConstantSource(t *testing.T) {
	ctx := &context.Context{
		Data: []context.OHLCV{
			{Close: 7}, {Close: 7}, {Close: 7}, {Close: 7}, {Close: 7},
		},
	}
	evaluator := NewStreamingBarEvaluator()
	call := makePercentileCall("percentile_nearest_rank", "close", 4, 50)

	v, err := evaluator.EvaluateAtBar(call, ctx, 3)
	if err != nil {
		t.Fatalf("EvaluateAtBar error: %v", err)
	}
	if math.Abs(v-7.0) > 0.0001 {
		t.Errorf("constant source must return 7.0, got %f", v)
	}
}

/* =================== PercentileLinearInterpolation =================== */

func TestStreamingBarEvaluator_PercentileLinearInterpolationWarmupYieldsNaN(t *testing.T) {
	ctx := &context.Context{
		Data: []context.OHLCV{
			{Close: 10}, {Close: 20}, {Close: 30}, {Close: 40},
		},
	}
	evaluator := NewStreamingBarEvaluator()
	call := makePercentileCall("percentile_linear_interpolation", "close", 4, 50)

	for barIdx := 0; barIdx < 3; barIdx++ {
		v, err := evaluator.EvaluateAtBar(call, ctx, barIdx)
		if err != nil {
			t.Fatalf("bar %d: unexpected error: %v", barIdx, err)
		}
		if !math.IsNaN(v) {
			t.Errorf("bar %d: expected NaN (warmup), got %f", barIdx, v)
		}
	}
}

func TestStreamingBarEvaluator_PercentileLinearInterpolationSortOrderIndependence(t *testing.T) {
	/* Unsorted window [1,4,3,2] with pct=50, period=4:
	   rank = 0.5*(4-1) = 1.5 → sorted[1]=2, sorted[2]=3, frac=0.5 → 2+0.5*(3-2) = 2.5 */
	ctx := &context.Context{
		Data: []context.OHLCV{
			{Close: 1}, {Close: 4}, {Close: 3}, {Close: 2},
		},
	}
	evaluator := NewStreamingBarEvaluator()
	call := makePercentileCall("percentile_linear_interpolation", "close", 4, 50)

	v, err := evaluator.EvaluateAtBar(call, ctx, 3)
	if err != nil {
		t.Fatalf("EvaluateAtBar error: %v", err)
	}
	if math.Abs(v-2.5) > 0.0001 {
		t.Errorf("PLI pct=50 of [1,4,3,2]: expected 2.5, got %f", v)
	}
}

func TestStreamingBarEvaluator_PercentileLinearInterpolationPct0ReturnsMin(t *testing.T) {
	ctx := &context.Context{
		Data: []context.OHLCV{
			{Close: 3}, {Close: 1}, {Close: 4}, {Close: 1}, {Close: 5},
		},
	}
	evaluator := NewStreamingBarEvaluator()
	call := makePercentileCall("percentile_linear_interpolation", "close", 5, 0)

	v, err := evaluator.EvaluateAtBar(call, ctx, 4)
	if err != nil {
		t.Fatalf("EvaluateAtBar error: %v", err)
	}
	if math.Abs(v-1.0) > 0.0001 {
		t.Errorf("pct=0 must return window min 1.0, got %f", v)
	}
}

func TestStreamingBarEvaluator_PercentileLinearInterpolationPct100ReturnsMax(t *testing.T) {
	ctx := &context.Context{
		Data: []context.OHLCV{
			{Close: 3}, {Close: 1}, {Close: 4}, {Close: 1}, {Close: 5},
		},
	}
	evaluator := NewStreamingBarEvaluator()
	call := makePercentileCall("percentile_linear_interpolation", "close", 5, 100)

	v, err := evaluator.EvaluateAtBar(call, ctx, 4)
	if err != nil {
		t.Fatalf("EvaluateAtBar error: %v", err)
	}
	if math.Abs(v-5.0) > 0.0001 {
		t.Errorf("pct=100 must return window max 5.0, got %f", v)
	}
}

func TestStreamingBarEvaluator_PercentileLinearInterpolationOrderingInvariant(t *testing.T) {
	/* PLI(pct=25) <= PLI(pct=50) <= PLI(pct=75) at every valid bar. */
	ctx := &context.Context{
		Data: []context.OHLCV{
			{Close: 4}, {Close: 2}, {Close: 7}, {Close: 1}, {Close: 9},
			{Close: 3}, {Close: 8}, {Close: 5}, {Close: 6}, {Close: 10},
		},
	}
	evaluator25 := NewStreamingBarEvaluator()
	evaluator50 := NewStreamingBarEvaluator()
	evaluator75 := NewStreamingBarEvaluator()

	call25 := makePercentileCall("percentile_linear_interpolation", "close", 5, 25)
	call50 := makePercentileCall("percentile_linear_interpolation", "close", 5, 50)
	call75 := makePercentileCall("percentile_linear_interpolation", "close", 5, 75)

	for barIdx := 4; barIdx < len(ctx.Data); barIdx++ {
		p25, _ := evaluator25.EvaluateAtBar(call25, ctx, barIdx)
		p50, _ := evaluator50.EvaluateAtBar(call50, ctx, barIdx)
		p75, _ := evaluator75.EvaluateAtBar(call75, ctx, barIdx)

		if p25 > p50+0.0001 {
			t.Errorf("bar %d: PLI(25)=%f > PLI(50)=%f (ordering violated)", barIdx, p25, p50)
		}
		if p50 > p75+0.0001 {
			t.Errorf("bar %d: PLI(50)=%f > PLI(75)=%f (ordering violated)", barIdx, p50, p75)
		}
	}
}

func TestStreamingBarEvaluator_PercentileLinearInterpolationConstantSource(t *testing.T) {
	ctx := &context.Context{
		Data: []context.OHLCV{
			{Close: 7}, {Close: 7}, {Close: 7}, {Close: 7}, {Close: 7},
		},
	}
	evaluator := NewStreamingBarEvaluator()
	call := makePercentileCall("percentile_linear_interpolation", "close", 4, 75)

	v, err := evaluator.EvaluateAtBar(call, ctx, 3)
	if err != nil {
		t.Fatalf("EvaluateAtBar error: %v", err)
	}
	if math.Abs(v-7.0) > 0.0001 {
		t.Errorf("constant source must return 7.0, got %f", v)
	}
}
