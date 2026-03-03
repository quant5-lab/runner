package security

import (
	"math"
	"testing"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/runtime/context"
)

func createSWMACallExpression(source string) *ast.CallExpression {
	return &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "ta"},
			Property: &ast.Identifier{Name: "swma"},
		},
		Arguments: []ast.Expression{
			&ast.Identifier{Name: source},
		},
	}
}

func createTACallExpression3(funcName, source string, p1, p2 float64) *ast.CallExpression {
	return &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "ta"},
			Property: &ast.Identifier{Name: funcName},
		},
		Arguments: []ast.Expression{
			&ast.Identifier{Name: source},
			&ast.Literal{Value: p1},
			&ast.Literal{Value: p2},
		},
	}
}

func TestStreamingBarEvaluator_SWMAWarmupAndResult(t *testing.T) {
	ctx := &context.Context{
		Data: []context.OHLCV{
			{Close: 10},
			{Close: 20},
			{Close: 30},
			{Close: 40},
			{Close: 50},
		},
	}

	evaluator := NewStreamingBarEvaluator()
	callExpr := createSWMACallExpression("close")

	for barIdx := 0; barIdx < 3; barIdx++ {
		value, err := evaluator.EvaluateAtBar(callExpr, ctx, barIdx)
		if err != nil {
			t.Fatalf("bar %d: EvaluateAtBar failed: %v", barIdx, err)
		}
		if !math.IsNaN(value) {
			t.Errorf("bar %d: expected NaN (warmup), got %f", barIdx, value)
		}
	}

	/* bar 3: 10*(1/6) + 20*(2/6) + 30*(2/6) + 40*(1/6) = 25.0 */
	expected := 10.0/6.0 + 20.0*2.0/6.0 + 30.0*2.0/6.0 + 40.0/6.0
	value, err := evaluator.EvaluateAtBar(callExpr, ctx, 3)
	if err != nil {
		t.Fatalf("bar 3: EvaluateAtBar failed: %v", err)
	}
	if math.Abs(value-expected) > 0.0001 {
		t.Errorf("bar 3: expected %.4f, got %.4f", expected, value)
	}
}

func TestStreamingBarEvaluator_CCIFlatSourceZeroOutput(t *testing.T) {
	ctx := &context.Context{
		Data: []context.OHLCV{
			{Close: 100},
			{Close: 100},
			{Close: 100},
			{Close: 100},
			{Close: 100},
		},
	}

	evaluator := NewStreamingBarEvaluator()
	callExpr := createTACallExpression("cci", "close", 5.0)

	value, err := evaluator.EvaluateAtBar(callExpr, ctx, 4)
	if err != nil {
		t.Fatalf("EvaluateAtBar failed: %v", err)
	}

	if math.Abs(value-0.0) > 0.0001 {
		t.Errorf("CCI with flat source = %f, want 0.0", value)
	}
}

func TestStreamingBarEvaluator_CCIWarmupPeriod(t *testing.T) {
	ctx := &context.Context{
		Data: []context.OHLCV{
			{Close: 10},
			{Close: 12},
			{Close: 14},
			{Close: 16},
			{Close: 18},
		},
	}

	evaluator := NewStreamingBarEvaluator()
	callExpr := createTACallExpression("cci", "close", 5.0)

	for barIdx := 0; barIdx < 4; barIdx++ {
		value, err := evaluator.EvaluateAtBar(callExpr, ctx, barIdx)
		if err != nil {
			t.Fatalf("bar %d: EvaluateAtBar failed: %v", barIdx, err)
		}
		if !math.IsNaN(value) {
			t.Errorf("bar %d: expected NaN (warmup), got %f", barIdx, value)
		}
	}
}

func TestStreamingBarEvaluator_BBWFlatSourceZeroOutput(t *testing.T) {
	ctx := &context.Context{
		Data: []context.OHLCV{
			{Close: 50},
			{Close: 50},
			{Close: 50},
			{Close: 50},
			{Close: 50},
		},
	}

	evaluator := NewStreamingBarEvaluator()
	callExpr := createTACallExpression("bbw", "close", 5.0)

	value, err := evaluator.EvaluateAtBar(callExpr, ctx, 4)
	if err != nil {
		t.Fatalf("EvaluateAtBar failed: %v", err)
	}

	if math.Abs(value-0.0) > 0.0001 {
		t.Errorf("BBW with flat source = %f, want 0.0", value)
	}
}

func TestStreamingBarEvaluator_BBWNonNegative(t *testing.T) {
	ctx := &context.Context{
		Data: []context.OHLCV{
			{Close: 10},
			{Close: 20},
			{Close: 15},
			{Close: 25},
			{Close: 12},
		},
	}

	evaluator := NewStreamingBarEvaluator()
	callExpr := createTACallExpression("bbw", "close", 5.0)

	value, err := evaluator.EvaluateAtBar(callExpr, ctx, 4)
	if err != nil {
		t.Fatalf("EvaluateAtBar failed: %v", err)
	}

	if value < 0 {
		t.Errorf("BBW must be non-negative, got %f", value)
	}
}

func TestStreamingBarEvaluator_COGSymmetricWindow(t *testing.T) {
	/* Uniform source: COG = -(1+2+...+period)/period */
	ctx := &context.Context{
		Data: []context.OHLCV{
			{Close: 5},
			{Close: 5},
			{Close: 5},
		},
	}

	evaluator := NewStreamingBarEvaluator()
	callExpr := createTACallExpression("cog", "close", 3.0)

	value, err := evaluator.EvaluateAtBar(callExpr, ctx, 2)
	if err != nil {
		t.Fatalf("EvaluateAtBar failed: %v", err)
	}

	/* -(5*1+5*2+5*3)/(5*3) = -30/15 = -2.0 */
	expected := -(5.0*1 + 5.0*2 + 5.0*3) / (5.0 * 3)
	if math.Abs(value-expected) > 0.0001 {
		t.Errorf("COG uniform source = %f, want %f", value, expected)
	}
}

func TestStreamingBarEvaluator_COGWarmupPeriod(t *testing.T) {
	ctx := &context.Context{
		Data: []context.OHLCV{
			{Close: 10},
			{Close: 20},
			{Close: 30},
			{Close: 40},
		},
	}

	evaluator := NewStreamingBarEvaluator()
	callExpr := createTACallExpression("cog", "close", 3.0)

	value, err := evaluator.EvaluateAtBar(callExpr, ctx, 0)
	if err != nil {
		t.Fatalf("EvaluateAtBar failed: %v", err)
	}
	if !math.IsNaN(value) {
		t.Errorf("bar 0: expected NaN (warmup), got %f", value)
	}

	value, err = evaluator.EvaluateAtBar(callExpr, ctx, 1)
	if err != nil {
		t.Fatalf("EvaluateAtBar failed: %v", err)
	}
	if !math.IsNaN(value) {
		t.Errorf("bar 1: expected NaN (warmup for period=3), got %f", value)
	}
}

func TestStreamingBarEvaluator_TSIWarmupPeriod(t *testing.T) {
	tests := []struct {
		name  string
		short float64
		long  float64
	}{
		{"short1_long1", 1, 1},
		{"short1_long13", 1, 13},
		{"short5_long1", 5, 1},
		{"short5_long13", 5, 13},
		{"short13_long25", 13, 25},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			warmup := int(tt.short) + int(tt.long) - 1
			barCount := warmup + 5
			ctx := &context.Context{Data: make([]context.OHLCV, barCount)}
			for i := range ctx.Data {
				ctx.Data[i] = context.OHLCV{Close: float64(100 + i)}
			}

			evaluator := NewStreamingBarEvaluator()
			callExpr := createTACallExpression3("tsi", "close", tt.short, tt.long)

			for barIdx := 0; barIdx < warmup; barIdx++ {
				value, err := evaluator.EvaluateAtBar(callExpr, ctx, barIdx)
				if err != nil {
					t.Fatalf("bar %d: EvaluateAtBar failed: %v", barIdx, err)
				}
				if !math.IsNaN(value) {
					t.Errorf("bar %d: expected NaN during warmup, got %f", barIdx, value)
				}
			}

			value, err := evaluator.EvaluateAtBar(callExpr, ctx, warmup)
			if err != nil {
				t.Fatalf("bar %d (first post-warmup): EvaluateAtBar failed: %v", warmup, err)
			}
			if math.IsNaN(value) {
				t.Errorf("bar %d (first post-warmup): expected valid value, got NaN", warmup)
			}
		})
	}
}

func TestStreamingBarEvaluator_TSIFlatSourceZeroOutput(t *testing.T) {
	ctx := &context.Context{
		Data: make([]context.OHLCV, 30),
	}
	for i := range ctx.Data {
		ctx.Data[i] = context.OHLCV{Close: 100}
	}

	evaluator := NewStreamingBarEvaluator()
	callExpr := createTACallExpression3("tsi", "close", 5.0, 13.0)

	value, err := evaluator.EvaluateAtBar(callExpr, ctx, 25)
	if err != nil {
		t.Fatalf("EvaluateAtBar failed: %v", err)
	}

	if math.Abs(value-0.0) > 0.0001 {
		t.Errorf("TSI with flat source = %f, want 0.0", value)
	}
}

func TestStreamingBarEvaluator_SWMAInsufficientArguments(t *testing.T) {
	ctx := createTestContext()
	evaluator := NewStreamingBarEvaluator()

	callExpr := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "ta"},
			Property: &ast.Identifier{Name: "swma"},
		},
		Arguments: []ast.Expression{},
	}

	_, err := evaluator.EvaluateAtBar(callExpr, ctx, 0)
	if err == nil {
		t.Fatal("expected error for zero arguments to swma")
	}
}

func TestStreamingBarEvaluator_TSIInsufficientArguments(t *testing.T) {
	ctx := createTestContext()
	evaluator := NewStreamingBarEvaluator()

	tests := []struct {
		name     string
		argCount int
	}{
		{"no_args", 0},
		{"one_arg", 1},
		{"two_args", 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := make([]ast.Expression, tt.argCount)
			for i := range args {
				args[i] = &ast.Identifier{Name: "close"}
			}

			callExpr := &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ta"},
					Property: &ast.Identifier{Name: "tsi"},
				},
				Arguments: args,
			}

			_, err := evaluator.EvaluateAtBar(callExpr, ctx, 0)
			if err == nil {
				t.Fatalf("expected error for %d arguments to tsi", tt.argCount)
			}
		})
	}
}
