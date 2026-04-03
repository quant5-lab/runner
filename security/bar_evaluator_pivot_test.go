package security

import (
	"math"
	"testing"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/runtime/context"
)

// TestPivotEvaluator_HighAndLowDetection covers warmup, detection, and
// non-detection bars for both pivothigh and pivotlow through the evaluator
// dispatch path. Each kind uses an independent evaluator to verify isolation.
// Dataset: single peak/trough at bar 2 (center); leftBars=rightBars=2 means
// detection fires at bar 4, warmup spans bars 0–3, bars 5–6 are non-peaks.
func TestPivotEvaluator_HighAndLowDetection(t *testing.T) {
	ctx := &context.Context{
		Data: []context.OHLCV{
			{High: 95, Low: 95},
			{High: 95, Low: 95},
			{High: 110, Low: 70},
			{High: 95, Low: 95},
			{High: 95, Low: 95},
			{High: 95, Low: 95},
			{High: 95, Low: 95},
		},
	}

	type barCase struct {
		barIdx  int
		wantNaN bool
		want    float64
		label   string
	}
	tests := []struct {
		name     string
		funcName string
		source   string
		cases    []barCase
	}{
		{
			name: "pivothigh", funcName: "pivothigh", source: "high",
			cases: []barCase{
				{0, true, 0, "warmup_bar0"},
				{1, true, 0, "warmup_bar1"},
				{2, true, 0, "warmup_bar2"},
				{3, true, 0, "warmup_bar3"},
				{4, false, 110, "detection_bar4"},
				{5, true, 0, "non_detection_bar5"},
				{6, true, 0, "non_detection_bar6"},
			},
		},
		{
			name: "pivotlow", funcName: "pivotlow", source: "low",
			cases: []barCase{
				{0, true, 0, "warmup_bar0"},
				{1, true, 0, "warmup_bar1"},
				{2, true, 0, "warmup_bar2"},
				{3, true, 0, "warmup_bar3"},
				{4, false, 70, "detection_bar4"},
				{5, true, 0, "non_detection_bar5"},
				{6, true, 0, "non_detection_bar6"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evaluator := NewStreamingBarEvaluator()
			call := &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ta"},
					Property: &ast.Identifier{Name: tt.funcName},
				},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: tt.source},
					lit(2), lit(2),
				},
			}
			for _, bc := range tt.cases {
				t.Run(bc.label, func(t *testing.T) {
					result, err := evaluator.EvaluateAtBar(call, ctx, bc.barIdx)
					if err != nil {
						t.Fatalf("EvaluateAtBar failed: %v", err)
					}
					if bc.wantNaN {
						if !math.IsNaN(result) {
							t.Errorf("expected NaN, got %.4f", result)
						}
					} else {
						assertFloat64(t, bc.label, result, bc.want, 0)
					}
				})
			}
		})
	}
}

func TestPivotEvaluator_ArgumentValidation(t *testing.T) {
	ctx := &context.Context{Data: []context.OHLCV{{High: 100, Low: 90}}}
	evaluator := NewStreamingBarEvaluator()

	nonLit := &ast.Identifier{Name: "somevar"}

	tests := []struct {
		name string
		call *ast.CallExpression
	}{
		{
			name: "zero_arguments",
			call: &ast.CallExpression{
				Callee:    &ast.MemberExpression{Object: &ast.Identifier{Name: "ta"}, Property: &ast.Identifier{Name: "pivothigh"}},
				Arguments: []ast.Expression{},
			},
		},
		{
			name: "one_argument_only_source",
			call: &ast.CallExpression{
				Callee:    &ast.MemberExpression{Object: &ast.Identifier{Name: "ta"}, Property: &ast.Identifier{Name: "pivothigh"}},
				Arguments: []ast.Expression{&ast.Identifier{Name: "high"}},
			},
		},
		{
			name: "three_arg_non_numeric_left_bars",
			call: &ast.CallExpression{
				Callee:    &ast.MemberExpression{Object: &ast.Identifier{Name: "ta"}, Property: &ast.Identifier{Name: "pivothigh"}},
				Arguments: []ast.Expression{&ast.Identifier{Name: "high"}, nonLit, lit(2)},
			},
		},
		{
			name: "three_arg_non_numeric_right_bars",
			call: &ast.CallExpression{
				Callee:    &ast.MemberExpression{Object: &ast.Identifier{Name: "ta"}, Property: &ast.Identifier{Name: "pivothigh"}},
				Arguments: []ast.Expression{&ast.Identifier{Name: "high"}, lit(2), nonLit},
			},
		},
		{
			name: "two_arg_non_numeric_left_bars",
			call: &ast.CallExpression{
				Callee:    &ast.MemberExpression{Object: &ast.Identifier{Name: "ta"}, Property: &ast.Identifier{Name: "pivothigh"}},
				Arguments: []ast.Expression{nonLit, lit(2)},
			},
		},
		{
			name: "two_arg_non_numeric_right_bars",
			call: &ast.CallExpression{
				Callee:    &ast.MemberExpression{Object: &ast.Identifier{Name: "ta"}, Property: &ast.Identifier{Name: "pivothigh"}},
				Arguments: []ast.Expression{lit(2), nonLit},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := evaluator.EvaluateAtBar(tt.call, ctx, 0)
			if err == nil {
				t.Errorf("%s: expected error, got nil", tt.name)
			}
		})
	}
}

func TestPivotEvaluator_HistoricalSubscriptAccess(t *testing.T) {
	ctx := &context.Context{
		Data: []context.OHLCV{
			{High: 100}, {High: 105}, {High: 110},
			{High: 108}, {High: 103}, {High: 102},
			{High: 104}, {High: 107}, {High: 106}, {High: 101},
		},
	}
	evaluator := NewStreamingBarEvaluator()

	makePivotHighCall := func() *ast.CallExpression {
		return &ast.CallExpression{
			Callee: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "ta"},
				Property: &ast.Identifier{Name: "pivothigh"},
			},
			Arguments: []ast.Expression{
				&ast.Identifier{Name: "high"},
				&ast.Literal{Value: float64(2)},
				&ast.Literal{Value: float64(2)},
			},
		}
	}

	tests := []struct {
		name      string
		offset    float64
		barIdx    int
		wantValue float64
		wantErr   bool
	}{
		{"offset_1_at_bar5", 1, 5, 110, false},
		{"offset_2_at_bar6", 2, 6, 110, false},
		{"offset_beyond_history", 10, 5, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			memberExpr := &ast.MemberExpression{
				Object:   makePivotHighCall(),
				Property: &ast.Literal{Value: tt.offset},
			}
			result, err := evaluator.EvaluateAtBar(memberExpr, ctx, tt.barIdx)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil (result=%.2f)", result)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			assertFloat64(t, tt.name, result, tt.wantValue, 0)
		})
	}
}

func TestPivotEvaluator_TwoArgDefaultSource(t *testing.T) {
	ctx := &context.Context{
		Data: []context.OHLCV{
			{High: 100, Low: 90},
			{High: 105, Low: 85},
			{High: 112, Low: 78},
			{High: 105, Low: 85},
			{High: 100, Low: 90},
		},
	}
	evaluator := NewStreamingBarEvaluator()

	tests := []struct {
		name      string
		funcName  string
		wantValue float64
	}{
		{"pivothigh_defaults_to_high", "pivothigh", 112},
		{"pivotlow_defaults_to_low", "pivotlow", 78},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			call := &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ta"},
					Property: &ast.Identifier{Name: tt.funcName},
				},
				Arguments: []ast.Expression{
					&ast.Literal{Value: float64(2)},
					&ast.Literal{Value: float64(2)},
				},
			}
			result, err := evaluator.EvaluateAtBar(call, ctx, 4)
			if err != nil {
				t.Fatalf("EvaluateAtBar failed: %v", err)
			}
			assertFloat64(t, tt.name, result, tt.wantValue, 0)
		})
	}
}

// TestPivotEvaluator_CacheIsolationByWindow verifies that two ta.pivothigh calls
// differing only in window parameters maintain independent PivotStateManager
// entries — both leftBars and rightBars are incorporated into the cache key.
func TestPivotEvaluator_CacheIsolationByWindow(t *testing.T) {
	ctx := &context.Context{
		Data: []context.OHLCV{
			{High: 100}, {High: 110}, {High: 115},
			{High: 112}, {High: 108},
			{High: 106}, {High: 104},
		},
	}
	evaluator := NewStreamingBarEvaluator()

	makeCall := func(leftBars, rightBars int) *ast.CallExpression {
		return &ast.CallExpression{
			Callee: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "ta"},
				Property: &ast.Identifier{Name: "pivothigh"},
			},
			Arguments: []ast.Expression{
				&ast.Identifier{Name: "high"},
				&ast.Literal{Value: float64(leftBars)},
				&ast.Literal{Value: float64(rightBars)},
			},
		}
	}

	if _, err := evaluator.EvaluateAtBar(makeCall(1, 1), ctx, 3); err != nil {
		t.Fatalf("call(1,1): %v", err)
	}
	if _, err := evaluator.EvaluateAtBar(makeCall(2, 2), ctx, 4); err != nil {
		t.Fatalf("call(2,2): %v", err)
	}

	if got := len(evaluator.pivotStateCache); got != 2 {
		t.Errorf("expected 2 independent cache entries, got %d", got)
	}
}

// TestPivotEvaluator_CacheIsolationBySourceExpression verifies that two ta.pivothigh
// calls with the same window but different source expressions produce independent
// PivotStateManager entries — the source expression is part of the cache key.
func TestPivotEvaluator_CacheIsolationBySourceExpression(t *testing.T) {
	ctx := &context.Context{
		Data: []context.OHLCV{
			{High: 100, Close: 101},
			{High: 110, Close: 111},
			{High: 115, Close: 116},
			{High: 112, Close: 113},
			{High: 108, Close: 109},
		},
	}
	evaluator := NewStreamingBarEvaluator()

	makeCall := func(source string) *ast.CallExpression {
		return &ast.CallExpression{
			Callee: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "ta"},
				Property: &ast.Identifier{Name: "pivothigh"},
			},
			Arguments: []ast.Expression{
				&ast.Identifier{Name: source},
				lit(2), lit(2),
			},
		}
	}

	if _, err := evaluator.EvaluateAtBar(makeCall("high"), ctx, 4); err != nil {
		t.Fatalf("source=high: %v", err)
	}
	if _, err := evaluator.EvaluateAtBar(makeCall("close"), ctx, 4); err != nil {
		t.Fatalf("source=close: %v", err)
	}

	if got := len(evaluator.pivotStateCache); got != 2 {
		t.Errorf("expected 2 independent cache entries (one per source), got %d", got)
	}
}

// TestPivotEvaluator_CacheIsolationByKind verifies that ta.pivothigh and ta.pivotlow
// with identical source and window produce independent PivotStateManager entries —
// the function name (kind) is part of the cache key.
func TestPivotEvaluator_CacheIsolationByKind(t *testing.T) {
	ctx := &context.Context{
		Data: []context.OHLCV{
			{Close: 100}, {Close: 105}, {Close: 110},
			{Close: 107}, {Close: 103},
		},
	}
	evaluator := NewStreamingBarEvaluator()

	makeCall := func(funcName string) *ast.CallExpression {
		return &ast.CallExpression{
			Callee: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "ta"},
				Property: &ast.Identifier{Name: funcName},
			},
			Arguments: []ast.Expression{
				&ast.Identifier{Name: "close"},
				lit(2), lit(2),
			},
		}
	}

	if _, err := evaluator.EvaluateAtBar(makeCall("pivothigh"), ctx, 4); err != nil {
		t.Fatalf("pivothigh: %v", err)
	}
	if _, err := evaluator.EvaluateAtBar(makeCall("pivotlow"), ctx, 4); err != nil {
		t.Fatalf("pivotlow: %v", err)
	}

	if got := len(evaluator.pivotStateCache); got != 2 {
		t.Errorf("expected 2 independent cache entries (one per kind), got %d", got)
	}
}

// TestPivotEvaluator_NaNSourceBlocksDuringWarmup verifies that when a NaN-producing
// source (SMA during warmup) yields NaN at neighbor positions, those NaN values
// block the pivot — matching PineScript semantics where "center > na" is false.
func TestPivotEvaluator_NaNSourceBlocksDuringWarmup(t *testing.T) {
	ctx := makeCtxClose(10, 8, 9, 3, 2)
	evaluator := NewStreamingBarEvaluator()

	smaExpr := makeTACall2Expr("sma", &ast.Identifier{Name: "close"}, lit(3))
	call := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "ta"},
			Property: &ast.Identifier{Name: "pivothigh"},
		},
		Arguments: []ast.Expression{smaExpr, lit(2), lit(2)},
	}

	result, err := evaluator.EvaluateAtBar(call, ctx, 4)
	if err != nil {
		t.Fatalf("EvaluateAtBar failed: %v", err)
	}
	if !math.IsNaN(result) {
		t.Errorf("expected NaN (SMA warmup NaN blocks), got %.6f", result)
	}
}

// TestPivotEvaluator_NaNSourceDetectsAfterWarmup verifies that once an SMA source
// has warmed up and all neighbor positions hold valid values, pivot detection
// proceeds normally — NaN blocking only applies where the source actually yields NaN.
func TestPivotEvaluator_NaNSourceDetectsAfterWarmup(t *testing.T) {
	ctx := makeCtxClose(50, 50, 50, 50, 90, 45, 40)
	evaluator := NewStreamingBarEvaluator()

	smaExpr := makeTACall2Expr("sma", &ast.Identifier{Name: "close"}, lit(3))
	call := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "ta"},
			Property: &ast.Identifier{Name: "pivothigh"},
		},
		Arguments: []ast.Expression{smaExpr, lit(2), lit(2)},
	}

	result, err := evaluator.EvaluateAtBar(call, ctx, 6)
	if err != nil {
		t.Fatalf("EvaluateAtBar failed: %v", err)
	}
	assertFloat64(t, "post_warmup_sma_pivot", result, 190.0/3.0, 1e-10)
}

// TestPivotEvaluator_BinaryExpressionSource verifies that an arbitrary
// BinaryExpression (not just an identifier) is accepted as the pivot source.
// This exercises the ast.Expression generalization of extractPivotArguments
// all the way through evaluator dispatch → PivotStateManager → makeExtractor.
// Dataset: high-low spread peaks at bar 2 (25.0) and is uniform elsewhere (5.0);
// with leftBars=rightBars=2 the spread pivot is detected at bar 4.
func TestPivotEvaluator_BinaryExpressionSource(t *testing.T) {
	ctx := &context.Context{
		Data: []context.OHLCV{
			{High: 95, Low: 90},
			{High: 95, Low: 90},
			{High: 115, Low: 90},
			{High: 95, Low: 90},
			{High: 95, Low: 90},
		},
	}
	evaluator := NewStreamingBarEvaluator()

	spreadExpr := &ast.BinaryExpression{
		Operator: "-",
		Left:     &ast.Identifier{Name: "high"},
		Right:    &ast.Identifier{Name: "low"},
	}
	call := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "ta"},
			Property: &ast.Identifier{Name: "pivothigh"},
		},
		Arguments: []ast.Expression{spreadExpr, lit(2), lit(2)},
	}

	tests := []struct {
		barIdx  int
		wantNaN bool
		want    float64
		label   string
	}{
		{0, true, 0, "warmup_bar0"},
		{1, true, 0, "warmup_bar1"},
		{2, true, 0, "warmup_bar2"},
		{3, true, 0, "warmup_bar3"},
		{4, false, 25.0, "detection_bar4"},
	}

	for _, bc := range tests {
		t.Run(bc.label, func(t *testing.T) {
			result, err := evaluator.EvaluateAtBar(call, ctx, bc.barIdx)
			if err != nil {
				t.Fatalf("EvaluateAtBar failed: %v", err)
			}
			if bc.wantNaN {
				if !math.IsNaN(result) {
					t.Errorf("expected NaN, got %.4f", result)
				}
			} else {
				assertFloat64(t, bc.label, result, bc.want, 0)
			}
		})
	}
}
