package security

import (
	"math"
	"testing"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/runtime/context"
)

func TestStreamingBarEvaluator_OHLCVFields(t *testing.T) {
	ctx := createTestContext()
	evaluator := NewStreamingBarEvaluator()

	tests := []struct {
		name     string
		field    string
		barIdx   int
		expected float64
	}{
		{"close_bar0", "close", 0, 102},
		{"close_bar1", "close", 1, 104},
		{"close_bar2", "close", 2, 106},
		{"open_bar0", "open", 0, 100},
		{"open_bar2", "open", 2, 104},
		{"high_bar0", "high", 0, 105},
		{"high_bar1", "high", 1, 107},
		{"low_bar0", "low", 0, 95},
		{"low_bar2", "low", 2, 99},
		{"volume_bar0", "volume", 0, 1000},
		{"volume_bar1", "volume", 1, 1100},
		{"volume_bar2", "volume", 2, 1200},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr := &ast.Identifier{Name: tt.field}

			value, err := evaluator.EvaluateAtBar(expr, ctx, tt.barIdx)
			if err != nil {
				t.Fatalf("EvaluateAtBar failed: %v", err)
			}

			if value != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, value)
			}
		})
	}
}

func TestStreamingBarEvaluator_SMAWarmupAndProgression(t *testing.T) {
	ctx := &context.Context{
		Data: []context.OHLCV{
			{Close: 100},
			{Close: 102},
			{Close: 104},
			{Close: 106},
			{Close: 108},
			{Close: 110},
		},
	}

	evaluator := NewStreamingBarEvaluator()

	callExpr := createTACallExpression("sma", "close", 3.0)

	tests := []struct {
		barIdx   int
		expected float64
		desc     string
	}{
		{0, 0.0, "warmup_bar0"},
		{1, 0.0, "warmup_bar1"},
		{2, 102.0, "first_valid"},
		{3, 104.0, "progression_bar3"},
		{4, 106.0, "progression_bar4"},
		{5, 108.0, "progression_bar5"},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			value, err := evaluator.EvaluateAtBar(callExpr, ctx, tt.barIdx)
			if err != nil {
				t.Fatalf("bar %d: EvaluateAtBar failed: %v", tt.barIdx, err)
			}

			if math.Abs(value-tt.expected) > 0.0001 {
				t.Errorf("bar %d: expected %.4f, got %.4f", tt.barIdx, tt.expected, value)
			}
		})
	}
}

func TestStreamingBarEvaluator_SMAStateReuse(t *testing.T) {
	ctx := &context.Context{
		Data: []context.OHLCV{
			{Close: 100},
			{Close: 102},
			{Close: 104},
			{Close: 106},
		},
	}

	evaluator := NewStreamingBarEvaluator()
	callExpr := createTACallExpression("sma", "close", 3.0)

	value1, _ := evaluator.EvaluateAtBar(callExpr, ctx, 2)
	value2, _ := evaluator.EvaluateAtBar(callExpr, ctx, 2)

	if value1 != value2 {
		t.Errorf("state reuse failed: first call %.4f, second call %.4f", value1, value2)
	}

	value3, _ := evaluator.EvaluateAtBar(callExpr, ctx, 3)
	if value3 <= value1 {
		t.Errorf("progression failed: bar 2 = %.4f, bar 3 = %.4f", value1, value3)
	}
}

func TestStreamingBarEvaluator_EMAWarmupAndConvergence(t *testing.T) {
	ctx := &context.Context{
		Data: []context.OHLCV{
			{Close: 100},
			{Close: 102},
			{Close: 104},
			{Close: 106},
			{Close: 108},
		},
	}

	evaluator := NewStreamingBarEvaluator()
	callExpr := createTACallExpression("ema", "close", 3.0)

	tests := []struct {
		barIdx      int
		minExpected float64
		maxExpected float64
		desc        string
	}{
		{0, 0.0, 0.0, "warmup_bar0"},
		{1, 0.0, 0.0, "warmup_bar1"},
		{2, 101.0, 103.0, "first_valid"},
		{3, 103.0, 105.0, "convergence_bar3"},
		{4, 105.0, 107.0, "convergence_bar4"},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			value, err := evaluator.EvaluateAtBar(callExpr, ctx, tt.barIdx)
			if err != nil {
				t.Fatalf("bar %d: failed: %v", tt.barIdx, err)
			}

			if value < tt.minExpected || value > tt.maxExpected {
				t.Errorf("bar %d: expected [%.2f, %.2f], got %.4f",
					tt.barIdx, tt.minExpected, tt.maxExpected, value)
			}
		})
	}
}

func TestStreamingBarEvaluator_RMASmoothing(t *testing.T) {
	ctx := &context.Context{
		Data: []context.OHLCV{
			{Close: 100},
			{Close: 110},
			{Close: 105},
			{Close: 115},
			{Close: 108},
		},
	}

	evaluator := NewStreamingBarEvaluator()
	callExpr := createTACallExpression("rma", "close", 3.0)

	value4, err := evaluator.EvaluateAtBar(callExpr, ctx, 4)
	if err != nil {
		t.Fatalf("RMA evaluation failed: %v", err)
	}

	if value4 < 105.0 || value4 > 112.0 {
		t.Errorf("RMA bar 4: expected smoothed value in [105, 112], got %.4f", value4)
	}
}

func TestStreamingBarEvaluator_RSICalculation(t *testing.T) {
	ctx := &context.Context{
		Data: []context.OHLCV{
			{Close: 100},
			{Close: 102},
			{Close: 101},
			{Close: 103},
			{Close: 102},
			{Close: 104},
			{Close: 106},
		},
	}

	evaluator := NewStreamingBarEvaluator()
	callExpr := createTACallExpression("rsi", "close", 3.0)

	value6, err := evaluator.EvaluateAtBar(callExpr, ctx, 6)
	if err != nil {
		t.Fatalf("RSI evaluation failed: %v", err)
	}

	if value6 < 0.0 || value6 > 100.0 {
		t.Errorf("RSI bar 6: expected [0, 100], got %.4f", value6)
	}
}

func TestStreamingBarEvaluator_MultipleTAFunctions(t *testing.T) {
	ctx := &context.Context{
		Data: []context.OHLCV{
			{Close: 100},
			{Close: 102},
			{Close: 104},
			{Close: 106},
		},
	}

	evaluator := NewStreamingBarEvaluator()

	smaExpr := createTACallExpression("sma", "close", 3.0)
	emaExpr := createTACallExpression("ema", "close", 3.0)

	smaValue, err := evaluator.EvaluateAtBar(smaExpr, ctx, 3)
	if err != nil {
		t.Fatalf("SMA evaluation failed: %v", err)
	}

	emaValue, err := evaluator.EvaluateAtBar(emaExpr, ctx, 3)
	if err != nil {
		t.Fatalf("EMA evaluation failed: %v", err)
	}

	if smaValue == 0.0 || emaValue == 0.0 {
		t.Error("multiple TA functions should produce non-zero values")
	}

	if math.Abs(smaValue-emaValue) > 10.0 {
		t.Errorf("SMA and EMA diverged too much: SMA=%.2f, EMA=%.2f", smaValue, emaValue)
	}
}

func TestStreamingBarEvaluator_DifferentSourceFields(t *testing.T) {
	ctx := &context.Context{
		Data: []context.OHLCV{
			{Open: 100, Close: 102, High: 105, Low: 98},
			{Open: 102, Close: 104, High: 107, Low: 100},
			{Open: 104, Close: 106, High: 109, Low: 102},
			{Open: 106, Close: 108, High: 111, Low: 104},
		},
	}

	evaluator := NewStreamingBarEvaluator()

	sources := []string{"close", "open", "high", "low"}
	for _, source := range sources {
		t.Run("sma_"+source, func(t *testing.T) {
			callExpr := createTACallExpression("sma", source, 3.0)
			value, err := evaluator.EvaluateAtBar(callExpr, ctx, 3)
			if err != nil {
				t.Fatalf("SMA(%s) failed: %v", source, err)
			}
			if value == 0.0 {
				t.Errorf("SMA(%s) should not be zero at bar 3", source)
			}
		})
	}
}

func TestStreamingBarEvaluator_UnknownIdentifier(t *testing.T) {
	ctx := createTestContext()
	evaluator := NewStreamingBarEvaluator()

	expr := &ast.Identifier{Name: "unknown"}

	_, err := evaluator.EvaluateAtBar(expr, ctx, 0)
	if err == nil {
		t.Fatal("expected error for unknown identifier")
	}

	assertSecurityErrorType(t, err, "UnknownIdentifier")
}

func TestStreamingBarEvaluator_BarIndexOutOfRange(t *testing.T) {
	ctx := createTestContext()
	evaluator := NewStreamingBarEvaluator()

	tests := []struct {
		name   string
		barIdx int
	}{
		{"negative_index", -1},
		{"beyond_length", 99},
		{"exact_length", len(ctx.Data)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr := &ast.Identifier{Name: "close"}
			_, err := evaluator.EvaluateAtBar(expr, ctx, tt.barIdx)

			if err == nil {
				t.Fatalf("expected error for bar index %d", tt.barIdx)
			}

			assertSecurityErrorType(t, err, "BarIndexOutOfRange")
		})
	}
}

func TestStreamingBarEvaluator_InsufficientArguments(t *testing.T) {
	ctx := createTestContext()
	evaluator := NewStreamingBarEvaluator()

	tests := []struct {
		name     string
		funcName string
		argCount int
	}{
		{"sma_no_args", "sma", 0},
		{"sma_one_arg", "sma", 1},
		{"ema_one_arg", "ema", 1},
		{"rma_no_args", "rma", 0},
		{"rsi_one_arg", "rsi", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := make([]ast.Expression, tt.argCount)
			for i := 0; i < tt.argCount; i++ {
				args[i] = &ast.Identifier{Name: "close"}
			}

			callExpr := &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ta"},
					Property: &ast.Identifier{Name: tt.funcName},
				},
				Arguments: args,
			}

			_, err := evaluator.EvaluateAtBar(callExpr, ctx, 0)
			if err == nil {
				t.Fatal("expected error for insufficient arguments")
			}

			assertSecurityErrorType(t, err, "InsufficientArguments")
		})
	}
}

func TestStreamingBarEvaluator_UnsupportedExpression(t *testing.T) {
	ctx := createTestContext()
	evaluator := NewStreamingBarEvaluator()

	unsupportedExpr := &ast.MemberExpression{
		Object:   &ast.Identifier{Name: "syminfo"},
		Property: &ast.Identifier{Name: "tickerid"},
	}

	_, err := evaluator.EvaluateAtBar(unsupportedExpr, ctx, 0)
	if err == nil {
		t.Fatal("expected error for unsupported expression")
	}

	assertSecurityErrorType(t, err, "UnsupportedExpression")
}

func TestStreamingBarEvaluator_UnsupportedFunction(t *testing.T) {
	ctx := createTestContext()
	evaluator := NewStreamingBarEvaluator()

	callExpr := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "ta"},
			Property: &ast.Identifier{Name: "unknown"},
		},
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "close"},
			&ast.Literal{Value: 14.0},
		},
	}

	_, err := evaluator.EvaluateAtBar(callExpr, ctx, 0)
	if err == nil {
		t.Fatal("expected error for unsupported function")
	}

	assertSecurityErrorType(t, err, "UnsupportedFunction")
}

func TestStreamingBarEvaluator_EmptyContext(t *testing.T) {
	emptyCtx := &context.Context{
		Data: []context.OHLCV{},
	}

	evaluator := NewStreamingBarEvaluator()
	expr := &ast.Identifier{Name: "close"}

	_, err := evaluator.EvaluateAtBar(expr, emptyCtx, 0)
	if err == nil {
		t.Fatal("expected error for empty context")
	}

	assertSecurityErrorType(t, err, "BarIndexOutOfRange")
}

func TestStreamingBarEvaluator_SingleBarContext(t *testing.T) {
	singleBarCtx := &context.Context{
		Data: []context.OHLCV{
			{Close: 100},
		},
	}

	evaluator := NewStreamingBarEvaluator()

	t.Run("ohlcv_access", func(t *testing.T) {
		expr := &ast.Identifier{Name: "close"}
		value, err := evaluator.EvaluateAtBar(expr, singleBarCtx, 0)
		if err != nil {
			t.Fatalf("failed: %v", err)
		}
		if value != 100.0 {
			t.Errorf("expected 100.0, got %.2f", value)
		}
	})

	t.Run("sma_warmup", func(t *testing.T) {
		callExpr := createTACallExpression("sma", "close", 3.0)
		value, err := evaluator.EvaluateAtBar(callExpr, singleBarCtx, 0)
		if err != nil {
			t.Fatalf("failed: %v", err)
		}
		if !math.IsNaN(value) {
			t.Errorf("expected warmup NaN, got %.2f", value)
		}
	})
}

func createTACallExpression(funcName, source string, period float64) *ast.CallExpression {
	return &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "ta"},
			Property: &ast.Identifier{Name: funcName},
		},
		Arguments: []ast.Expression{
			&ast.Identifier{Name: source},
			&ast.Literal{Value: period},
		},
	}
}

func assertSecurityErrorType(t *testing.T, err error, expectedType string) {
	t.Helper()

	secErr, ok := err.(*SecurityError)
	if !ok {
		t.Fatalf("expected SecurityError, got %T", err)
	}

	if secErr.Type != expectedType {
		t.Errorf("expected %s error, got %s", expectedType, secErr.Type)
	}
}

func createTestContext() *context.Context {
	return &context.Context{
		Data: []context.OHLCV{
			{Open: 100, High: 105, Low: 95, Close: 102, Volume: 1000},
			{Open: 102, High: 107, Low: 97, Close: 104, Volume: 1100},
			{Open: 104, High: 109, Low: 99, Close: 106, Volume: 1200},
		},
	}
}
