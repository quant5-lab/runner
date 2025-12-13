package request

import (
	"math"
	"testing"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/runtime/context"
)

func TestPivotEvaluator_EndToEnd_DirectCall(t *testing.T) {
	evaluator := NewPivotEvaluator()

	secCtx := &context.Context{
		Data: []context.OHLCV{
			{High: 100, Low: 90},
			{High: 110, Low: 95},
			{High: 115, Low: 98},
			{High: 120, Low: 100},
			{High: 115, Low: 97},
			{High: 110, Low: 94},
			{High: 105, Low: 91},
		},
	}

	expr := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "ta"},
			Property: &ast.Identifier{Name: "pivothigh"},
		},
		Arguments: []ast.Expression{
			&ast.Literal{Value: float64(2)},
			&ast.Literal{Value: float64(2)},
		},
	}

	value, ok := evaluator.TryEvaluate(expr, secCtx, 3)

	if !ok {
		t.Fatal("Expected pivot evaluation to succeed")
	}

	_ = value
}

func TestPivotEvaluator_EndToEnd_WithOffset(t *testing.T) {
	evaluator := NewPivotEvaluator()

	secCtx := &context.Context{
		Data: []context.OHLCV{
			{High: 100, Low: 90},
			{High: 110, Low: 95},
			{High: 105, Low: 92},
			{High: 115, Low: 98},
			{High: 108, Low: 94},
			{High: 112, Low: 96},
			{High: 102, Low: 91},
		},
	}

	expr := &ast.MemberExpression{
		Object: &ast.CallExpression{
			Callee: &ast.Identifier{Name: "pivotlow"},
			Arguments: []ast.Expression{
				&ast.Literal{Value: float64(2)},
				&ast.Literal{Value: float64(2)},
			},
		},
		Property: &ast.Literal{Value: float64(1)},
		Computed: true,
	}

	value, ok := evaluator.TryEvaluate(expr, secCtx, 5)

	if !ok {
		t.Fatal("Expected pivot evaluation with offset to succeed")
	}

	_ = value
}

func TestPivotEvaluator_NonPivotExpression(t *testing.T) {
	evaluator := NewPivotEvaluator()

	secCtx := &context.Context{
		Data: []context.OHLCV{
			{High: 100, Low: 90},
		},
	}

	expr := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "ta"},
			Property: &ast.Identifier{Name: "sma"},
		},
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "close"},
			&ast.Literal{Value: float64(20)},
		},
	}

	_, ok := evaluator.TryEvaluate(expr, secCtx, 0)

	if ok {
		t.Error("Expected non-pivot expression to return false")
	}
}

func TestPivotEvaluator_CacheEffectiveness(t *testing.T) {
	evaluator := NewPivotEvaluator()

	secCtx := &context.Context{
		Data: []context.OHLCV{
			{High: 100, Low: 90},
			{High: 110, Low: 95},
			{High: 105, Low: 92},
			{High: 115, Low: 98},
			{High: 108, Low: 94},
		},
	}

	expr := &ast.CallExpression{
		Callee: &ast.Identifier{Name: "pivothigh"},
		Arguments: []ast.Expression{
			&ast.Literal{Value: float64(1)},
			&ast.Literal{Value: float64(1)},
		},
	}

	val1, ok1 := evaluator.TryEvaluate(expr, secCtx, 2)
	if !ok1 {
		t.Fatal("First evaluation failed")
	}

	val2, ok2 := evaluator.TryEvaluate(expr, secCtx, 3)
	if !ok2 {
		t.Fatal("Second evaluation failed")
	}

	_ = val1
	_ = val2
}

func TestPivotEvaluator_OffsetBoundaries(t *testing.T) {
	evaluator := NewPivotEvaluator()

	secCtx := &context.Context{
		Data: []context.OHLCV{
			{High: 100, Low: 90},
			{High: 110, Low: 95},
			{High: 105, Low: 92},
			{High: 115, Low: 98},
			{High: 108, Low: 94},
		},
	}

	tests := []struct {
		name        string
		offset      int
		targetIndex int
	}{
		{"offset_0", 0, 2},
		{"offset_1", 1, 3},
		{"offset_2", 2, 4},
		{"offset_negative", -1, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr := &ast.MemberExpression{
				Object: &ast.CallExpression{
					Callee: &ast.Identifier{Name: "pivothigh"},
					Arguments: []ast.Expression{
						&ast.Literal{Value: float64(1)},
						&ast.Literal{Value: float64(1)},
					},
				},
				Property: &ast.Literal{Value: float64(tt.offset)},
				Computed: true,
			}

			_, ok := evaluator.TryEvaluate(expr, secCtx, tt.targetIndex)

			if !ok {
				t.Errorf("Evaluation failed for offset=%d, targetIndex=%d", tt.offset, tt.targetIndex)
			}
		})
	}
}

func TestPivotEvaluator_OutOfBoundsAccess(t *testing.T) {
	evaluator := NewPivotEvaluator()

	secCtx := &context.Context{
		Data: []context.OHLCV{
			{High: 100, Low: 90},
			{High: 110, Low: 95},
		},
	}

	expr := &ast.MemberExpression{
		Object: &ast.CallExpression{
			Callee: &ast.Identifier{Name: "pivothigh"},
			Arguments: []ast.Expression{
				&ast.Literal{Value: float64(1)},
				&ast.Literal{Value: float64(1)},
			},
		},
		Property: &ast.Literal{Value: float64(10)},
		Computed: true,
	}

	value, ok := evaluator.TryEvaluate(expr, secCtx, 1)

	if !ok {
		t.Fatal("Expected evaluation to succeed even with out-of-bounds offset")
	}

	if !math.IsNaN(value) {
		t.Errorf("Expected NaN for out-of-bounds access, got %f", value)
	}
}

func TestPivotEvaluator_EmptyContext(t *testing.T) {
	evaluator := NewPivotEvaluator()

	secCtx := &context.Context{
		Data: []context.OHLCV{},
	}

	expr := &ast.CallExpression{
		Callee: &ast.Identifier{Name: "pivothigh"},
		Arguments: []ast.Expression{
			&ast.Literal{Value: float64(1)},
			&ast.Literal{Value: float64(1)},
		},
	}

	value, ok := evaluator.TryEvaluate(expr, secCtx, 0)

	if !ok {
		t.Fatal("Expected evaluation to succeed for empty context")
	}

	if !math.IsNaN(value) {
		t.Errorf("Expected NaN for empty context, got %f", value)
	}
}

func TestPivotEvaluator_ClearCache(t *testing.T) {
	evaluator := NewPivotEvaluator()

	secCtx := &context.Context{
		Data: []context.OHLCV{
			{High: 100, Low: 90},
			{High: 110, Low: 95},
			{High: 105, Low: 92},
		},
	}

	expr := &ast.CallExpression{
		Callee: &ast.Identifier{Name: "pivothigh"},
		Arguments: []ast.Expression{
			&ast.Literal{Value: float64(1)},
			&ast.Literal{Value: float64(1)},
		},
	}

	// Evaluate to populate cache
	_, ok := evaluator.TryEvaluate(expr, secCtx, 1)
	if !ok {
		t.Fatal("Initial evaluation failed")
	}

	// Clear cache
	evaluator.ClearCache()

	// Evaluate again - should recompute
	_, ok = evaluator.TryEvaluate(expr, secCtx, 1)
	if !ok {
		t.Fatal("Post-clear evaluation failed")
	}
}

func TestPivotEvaluator_MultipleTypes(t *testing.T) {
	evaluator := NewPivotEvaluator()

	secCtx := &context.Context{
		Data: []context.OHLCV{
			{High: 100, Low: 90},
			{High: 110, Low: 95},
			{High: 105, Low: 92},
			{High: 115, Low: 98},
			{High: 108, Low: 94},
		},
	}

	exprHigh := &ast.CallExpression{
		Callee: &ast.Identifier{Name: "pivothigh"},
		Arguments: []ast.Expression{
			&ast.Literal{Value: float64(1)},
			&ast.Literal{Value: float64(1)},
		},
	}

	exprLow := &ast.CallExpression{
		Callee: &ast.Identifier{Name: "pivotlow"},
		Arguments: []ast.Expression{
			&ast.Literal{Value: float64(1)},
			&ast.Literal{Value: float64(1)},
		},
	}

	valHigh, okHigh := evaluator.TryEvaluate(exprHigh, secCtx, 2)
	valLow, okLow := evaluator.TryEvaluate(exprLow, secCtx, 2)

	if !okHigh {
		t.Error("pivothigh evaluation failed")
	}

	if !okLow {
		t.Error("pivotlow evaluation failed")
	}

	if !math.IsNaN(valHigh) && !math.IsNaN(valLow) && valHigh == valLow {
		t.Error("pivothigh and pivotlow should produce different values")
	}
}

func TestPivotEvaluator_AsymmetricBars(t *testing.T) {
	evaluator := NewPivotEvaluator()

	secCtx := &context.Context{
		Data: []context.OHLCV{
			{High: 100, Low: 90},
			{High: 105, Low: 92},
			{High: 110, Low: 95},
			{High: 115, Low: 98},
			{High: 108, Low: 94},
			{High: 112, Low: 96},
			{High: 106, Low: 93},
		},
	}

	expr := &ast.CallExpression{
		Callee: &ast.Identifier{Name: "pivothigh"},
		Arguments: []ast.Expression{
			&ast.Literal{Value: float64(1)},
			&ast.Literal{Value: float64(3)},
		},
	}

	value, ok := evaluator.TryEvaluate(expr, secCtx, 4)

	if !ok {
		t.Fatal("Evaluation with asymmetric bars failed")
	}

	_ = value
}

func TestPivotEvaluator_ZeroBarsParameters(t *testing.T) {
	evaluator := NewPivotEvaluator()

	secCtx := &context.Context{
		Data: []context.OHLCV{
			{High: 100, Low: 90},
			{High: 110, Low: 95},
			{High: 105, Low: 92},
		},
	}

	expr := &ast.CallExpression{
		Callee: &ast.Identifier{Name: "pivothigh"},
		Arguments: []ast.Expression{
			&ast.Literal{Value: float64(0)},
			&ast.Literal{Value: float64(0)},
		},
	}

	_, ok := evaluator.TryEvaluate(expr, secCtx, 1)

	if !ok {
		t.Fatal("Evaluation with zero bars failed")
	}
}
