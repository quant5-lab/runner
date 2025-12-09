package security

import (
	"math"
	"testing"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/runtime/context"
)

func TestStreamingBarEvaluator_BinaryExpression_Arithmetic(t *testing.T) {
	ctx := createTestContextBinary([]float64{100, 105, 110, 115, 120})
	evaluator := NewStreamingBarEvaluator()

	tests := []struct {
		name     string
		expr     *ast.BinaryExpression
		barIdx   int
		expected float64
	}{
		{
			name: "close + 5",
			expr: &ast.BinaryExpression{
				Operator: "+",
				Left:     &ast.Identifier{Name: "close"},
				Right:    &ast.Literal{Value: 5.0},
			},
			barIdx:   2,
			expected: 115.0,
		},
		{
			name: "close - 10",
			expr: &ast.BinaryExpression{
				Operator: "-",
				Left:     &ast.Identifier{Name: "close"},
				Right:    &ast.Literal{Value: 10.0},
			},
			barIdx:   3,
			expected: 105.0,
		},
		{
			name: "close * 2",
			expr: &ast.BinaryExpression{
				Operator: "*",
				Left:     &ast.Identifier{Name: "close"},
				Right:    &ast.Literal{Value: 2.0},
			},
			barIdx:   1,
			expected: 210.0,
		},
		{
			name: "close / 2",
			expr: &ast.BinaryExpression{
				Operator: "/",
				Left:     &ast.Identifier{Name: "close"},
				Right:    &ast.Literal{Value: 2.0},
			},
			barIdx:   4,
			expected: 60.0,
		},
		{
			name: "close % 7",
			expr: &ast.BinaryExpression{
				Operator: "%",
				Left:     &ast.Identifier{Name: "close"},
				Right:    &ast.Literal{Value: 7.0},
			},
			barIdx:   2,
			expected: 5.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, err := evaluator.EvaluateAtBar(tt.expr, ctx, tt.barIdx)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if math.Abs(value-tt.expected) > 1e-10 {
				t.Errorf("expected %.2f, got %.2f", tt.expected, value)
			}
		})
	}
}

func TestStreamingBarEvaluator_BinaryExpression_Comparison(t *testing.T) {
	ctx := createTestContextBinary([]float64{100, 105, 110, 115, 120})
	evaluator := NewStreamingBarEvaluator()

	tests := []struct {
		name     string
		expr     *ast.BinaryExpression
		barIdx   int
		expected float64
	}{
		{
			name: "close > 105",
			expr: &ast.BinaryExpression{
				Operator: ">",
				Left:     &ast.Identifier{Name: "close"},
				Right:    &ast.Literal{Value: 105.0},
			},
			barIdx:   2,
			expected: 1.0,
		},
		{
			name: "close < 105",
			expr: &ast.BinaryExpression{
				Operator: "<",
				Left:     &ast.Identifier{Name: "close"},
				Right:    &ast.Literal{Value: 105.0},
			},
			barIdx:   2,
			expected: 0.0,
		},
		{
			name: "close >= 110",
			expr: &ast.BinaryExpression{
				Operator: ">=",
				Left:     &ast.Identifier{Name: "close"},
				Right:    &ast.Literal{Value: 110.0},
			},
			barIdx:   2,
			expected: 1.0,
		},
		{
			name: "close <= 110",
			expr: &ast.BinaryExpression{
				Operator: "<=",
				Left:     &ast.Identifier{Name: "close"},
				Right:    &ast.Literal{Value: 110.0},
			},
			barIdx:   2,
			expected: 1.0,
		},
		{
			name: "close == 115",
			expr: &ast.BinaryExpression{
				Operator: "==",
				Left:     &ast.Identifier{Name: "close"},
				Right:    &ast.Literal{Value: 115.0},
			},
			barIdx:   3,
			expected: 1.0,
		},
		{
			name: "close != 115",
			expr: &ast.BinaryExpression{
				Operator: "!=",
				Left:     &ast.Identifier{Name: "close"},
				Right:    &ast.Literal{Value: 115.0},
			},
			barIdx:   2,
			expected: 1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, err := evaluator.EvaluateAtBar(tt.expr, ctx, tt.barIdx)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if math.Abs(value-tt.expected) > 1e-10 {
				t.Errorf("expected %.2f, got %.2f", tt.expected, value)
			}
		})
	}
}

func TestStreamingBarEvaluator_BinaryExpression_Nested(t *testing.T) {
	ctx := createTestContextBinary([]float64{100, 105, 110, 115, 120})
	evaluator := NewStreamingBarEvaluator()

	expr := &ast.BinaryExpression{
		Operator: "+",
		Left: &ast.BinaryExpression{
			Operator: "*",
			Left:     &ast.Identifier{Name: "close"},
			Right:    &ast.Literal{Value: 2.0},
		},
		Right: &ast.Literal{Value: 10.0},
	}

	value, err := evaluator.EvaluateAtBar(expr, ctx, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := (110.0 * 2) + 10
	if math.Abs(value-expected) > 1e-10 {
		t.Errorf("expected %.2f, got %.2f", expected, value)
	}
}

func TestStreamingBarEvaluator_BinaryExpression_WithTAFunction(t *testing.T) {
	ctx := createTestContextBinary([]float64{100, 102, 104, 106, 108, 110, 112})
	evaluator := NewStreamingBarEvaluator()

	smaCall := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "ta"},
			Property: &ast.Identifier{Name: "sma"},
		},
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "close"},
			&ast.Literal{Value: 3.0},
		},
	}

	expr := &ast.BinaryExpression{
		Operator: ">",
		Left:     smaCall,
		Right:    &ast.Literal{Value: 105.0},
	}

	value, err := evaluator.EvaluateAtBar(expr, ctx, 4)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if value != 1.0 {
		t.Errorf("expected sma(close,3) > 105 to be true at bar 4, got %.2f", value)
	}
}

func createTestContextBinary(closePrices []float64) *context.Context {
	data := make([]context.OHLCV, len(closePrices))
	for i, price := range closePrices {
		data[i] = context.OHLCV{
			Time:   int64(i * 86400),
			Open:   price,
			High:   price + 5,
			Low:    price - 5,
			Close:  price,
			Volume: 1000,
		}
	}

	return &context.Context{
		Data: data,
	}
}
