package security

import (
	"math"
	"testing"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/runtime/context"
)

func TestValuewhenStateManager_CatchUpLoop(t *testing.T) {
	data := []context.OHLCV{
		{Close: 100.0, High: 105.0},
		{Close: 103.0, High: 108.0},
		{Close: 101.0, High: 106.0},
		{Close: 104.0, High: 109.0},
		{Close: 105.0, High: 110.0},
		{Close: 106.0, High: 111.0},
	}

	ctx := &context.Context{Data: data}
	evaluator := NewStreamingBarEvaluator()

	conditionExpr := &ast.BinaryExpression{
		Operator: ">",
		Left:     &ast.Identifier{Name: "close"},
		Right:    &ast.Literal{Value: 102.0},
	}
	sourceExpr := &ast.Identifier{Name: "high"}

	mgr := NewValuewhenStateManager("test_catchup", 0, conditionExpr, sourceExpr, len(data), evaluator)

	result0, err := mgr.ComputeAtBar(ctx, nil, 3)
	if err != nil {
		t.Fatalf("ComputeAtBar(3) failed: %v", err)
	}
	if result0 != 109.0 {
		t.Errorf("bar 3: expected 109.0, got %.2f", result0)
	}

	result1, err := mgr.ComputeAtBar(ctx, nil, 5)
	if err != nil {
		t.Fatalf("ComputeAtBar(5) failed: %v", err)
	}
	if result1 != 111.0 {
		t.Errorf("bar 5: expected 111.0, got %.2f", result1)
	}

	result2, err := mgr.ComputeAtBar(ctx, nil, 4)
	if err != nil {
		t.Fatalf("ComputeAtBar(4) failed: %v", err)
	}
	if result2 != 110.0 {
		t.Errorf("bar 4: expected 110.0, got %.2f", result2)
	}
}

func TestValuewhenStateManager_OccurrenceTracking(t *testing.T) {
	data := []context.OHLCV{
		{Close: 100.0, High: 105.0},
		{Close: 103.0, High: 108.0},
		{Close: 101.0, High: 106.0},
		{Close: 104.0, High: 109.0},
		{Close: 105.0, High: 110.0},
	}

	ctx := &context.Context{Data: data}
	evaluator := NewStreamingBarEvaluator()

	conditionExpr := &ast.BinaryExpression{
		Operator: ">",
		Left:     &ast.Identifier{Name: "close"},
		Right:    &ast.Literal{Value: 102.0},
	}
	sourceExpr := &ast.Identifier{Name: "high"}

	tests := []struct {
		occurrence int
		barIdx     int
		expected   float64
		desc       string
	}{
		{0, 4, 110.0, "occurrence=0 at bar 4"},
		{1, 4, 109.0, "occurrence=1 at bar 4"},
		{2, 4, 108.0, "occurrence=2 at bar 4"},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			mgr := NewValuewhenStateManager("test_occ", tt.occurrence, conditionExpr, sourceExpr, len(data), evaluator)
			result, err := mgr.ComputeAtBar(ctx, nil, tt.barIdx)
			if err != nil {
				t.Fatalf("%s: ComputeAtBar failed: %v", tt.desc, err)
			}
			if result != tt.expected {
				t.Errorf("%s: expected %.2f, got %.2f", tt.desc, tt.expected, result)
			}
		})
	}
}

func TestValuewhenStateManager_NoMatches(t *testing.T) {
	data := []context.OHLCV{
		{Close: 100.0, High: 105.0},
		{Close: 101.0, High: 106.0},
		{Close: 102.0, High: 107.0},
	}

	ctx := &context.Context{Data: data}
	evaluator := NewStreamingBarEvaluator()

	conditionExpr := &ast.BinaryExpression{
		Operator: ">",
		Left:     &ast.Identifier{Name: "close"},
		Right:    &ast.Literal{Value: 200.0},
	}
	sourceExpr := &ast.Identifier{Name: "high"}

	mgr := NewValuewhenStateManager("test_nomatch", 0, conditionExpr, sourceExpr, len(data), evaluator)

	result, err := mgr.ComputeAtBar(ctx, nil, 2)
	if err != nil {
		t.Fatalf("ComputeAtBar failed: %v", err)
	}

	if !math.IsNaN(result) {
		t.Errorf("expected NaN when no matches exist, got %.2f", result)
	}
}

func TestValuewhenStateManager_OccurrenceBeyondAvailable(t *testing.T) {
	data := []context.OHLCV{
		{Close: 100.0, High: 105.0},
		{Close: 103.0, High: 108.0},
		{Close: 104.0, High: 109.0},
	}

	ctx := &context.Context{Data: data}
	evaluator := NewStreamingBarEvaluator()

	conditionExpr := &ast.BinaryExpression{
		Operator: ">",
		Left:     &ast.Identifier{Name: "close"},
		Right:    &ast.Literal{Value: 102.0},
	}
	sourceExpr := &ast.Identifier{Name: "high"}

	mgr := NewValuewhenStateManager("test_beyond", 10, conditionExpr, sourceExpr, len(data), evaluator)

	result, err := mgr.ComputeAtBar(ctx, nil, 2)
	if err != nil {
		t.Fatalf("ComputeAtBar failed: %v", err)
	}

	if !math.IsNaN(result) {
		t.Errorf("expected NaN when occurrence exceeds match count, got %.2f", result)
	}
}

func TestValuewhenStateManager_SequentialAccumulation(t *testing.T) {
	data := []context.OHLCV{
		{Close: 100.0, High: 105.0},
		{Close: 103.0, High: 108.0},
		{Close: 101.0, High: 106.0},
		{Close: 104.0, High: 109.0},
		{Close: 102.0, High: 107.0},
		{Close: 105.0, High: 110.0},
	}

	ctx := &context.Context{Data: data}
	evaluator := NewStreamingBarEvaluator()

	conditionExpr := &ast.BinaryExpression{
		Operator: ">",
		Left:     &ast.Identifier{Name: "close"},
		Right:    &ast.Literal{Value: 102.0},
	}
	sourceExpr := &ast.Identifier{Name: "high"}

	mgr := NewValuewhenStateManager("test_sequential", 0, conditionExpr, sourceExpr, len(data), evaluator)

	expectedSequence := []float64{
		math.NaN(),
		108.0,
		108.0,
		109.0,
		109.0,
		110.0,
	}

	for barIdx, expected := range expectedSequence {
		result, err := mgr.ComputeAtBar(ctx, nil, barIdx)
		if err != nil {
			t.Fatalf("bar %d: ComputeAtBar failed: %v", barIdx, err)
		}

		if math.IsNaN(expected) {
			if !math.IsNaN(result) {
				t.Errorf("bar %d: expected NaN, got %.2f", barIdx, result)
			}
		} else {
			if result != expected {
				t.Errorf("bar %d: expected %.2f, got %.2f", barIdx, expected, result)
			}
		}
	}
}

func TestValuewhenStateManager_NonDecreasingBarIdx(t *testing.T) {
	data := []context.OHLCV{
		{Close: 100.0, High: 105.0},
		{Close: 103.0, High: 108.0},
		{Close: 104.0, High: 109.0},
		{Close: 105.0, High: 110.0},
	}

	ctx := &context.Context{Data: data}
	evaluator := NewStreamingBarEvaluator()

	conditionExpr := &ast.BinaryExpression{
		Operator: ">",
		Left:     &ast.Identifier{Name: "close"},
		Right:    &ast.Literal{Value: 102.0},
	}
	sourceExpr := &ast.Identifier{Name: "high"}

	mgr := NewValuewhenStateManager("test_nondecreasing", 0, conditionExpr, sourceExpr, len(data), evaluator)

	barSequence := []int{0, 0, 0, 1, 1, 2, 2, 2, 3, 3}
	expectedResults := []float64{
		math.NaN(),
		math.NaN(),
		math.NaN(),
		108.0,
		108.0,
		109.0,
		109.0,
		109.0,
		110.0,
		110.0,
	}

	for i, barIdx := range barSequence {
		result, err := mgr.ComputeAtBar(ctx, nil, barIdx)
		if err != nil {
			t.Fatalf("iteration %d (bar %d): ComputeAtBar failed: %v", i, barIdx, err)
		}

		expected := expectedResults[i]
		if math.IsNaN(expected) {
			if !math.IsNaN(result) {
				t.Errorf("iteration %d (bar %d): expected NaN, got %.2f", i, barIdx, result)
			}
		} else {
			if result != expected {
				t.Errorf("iteration %d (bar %d): expected %.2f, got %.2f", i, barIdx, expected, result)
			}
		}
	}
}

func TestValuewhenStateManager_ComplexExpressions(t *testing.T) {
	data := []context.OHLCV{
		{Close: 100.0, High: 105.0, Low: 95.0},
		{Close: 103.0, High: 108.0, Low: 98.0},
		{Close: 104.0, High: 109.0, Low: 99.0},
	}

	ctx := &context.Context{Data: data}
	evaluator := NewStreamingBarEvaluator()

	conditionExpr := &ast.BinaryExpression{
		Operator: ">",
		Left:     &ast.Identifier{Name: "close"},
		Right:    &ast.Literal{Value: 102.0},
	}
	sourceExpr := &ast.BinaryExpression{
		Operator: "+",
		Left:     &ast.Identifier{Name: "high"},
		Right:    &ast.Identifier{Name: "low"},
	}

	mgr := NewValuewhenStateManager("test_complex", 0, conditionExpr, sourceExpr, len(data), evaluator)

	result, err := mgr.ComputeAtBar(ctx, nil, 2)
	if err != nil {
		t.Fatalf("ComputeAtBar failed: %v", err)
	}

	expected := 109.0 + 99.0
	if result != expected {
		t.Errorf("expected %.2f, got %.2f", expected, result)
	}
}

func TestValuewhenStateManager_EarlyBarBehavior(t *testing.T) {
	data := []context.OHLCV{
		{Close: 100.0, High: 105.0},
		{Close: 101.0, High: 106.0},
		{Close: 103.0, High: 108.0},
	}

	ctx := &context.Context{Data: data}
	evaluator := NewStreamingBarEvaluator()

	conditionExpr := &ast.BinaryExpression{
		Operator: ">",
		Left:     &ast.Identifier{Name: "close"},
		Right:    &ast.Literal{Value: 102.0},
	}
	sourceExpr := &ast.Identifier{Name: "high"}

	mgr := NewValuewhenStateManager("test_early", 0, conditionExpr, sourceExpr, len(data), evaluator)

	result0, err := mgr.ComputeAtBar(ctx, nil, 0)
	if err != nil {
		t.Fatalf("bar 0: ComputeAtBar failed: %v", err)
	}
	if !math.IsNaN(result0) {
		t.Errorf("bar 0: expected NaN (no matches yet), got %.2f", result0)
	}

	result1, err := mgr.ComputeAtBar(ctx, nil, 1)
	if err != nil {
		t.Fatalf("bar 1: ComputeAtBar failed: %v", err)
	}
	if !math.IsNaN(result1) {
		t.Errorf("bar 1: expected NaN (no matches yet), got %.2f", result1)
	}

	result2, err := mgr.ComputeAtBar(ctx, nil, 2)
	if err != nil {
		t.Fatalf("bar 2: ComputeAtBar failed: %v", err)
	}
	if result2 != 108.0 {
		t.Errorf("bar 2: expected 108.0 (first match), got %.2f", result2)
	}
}
