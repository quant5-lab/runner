package security

import (
	"math"
	"testing"

	"github.com/borisquantlab/pinescript-go/ast"
	"github.com/borisquantlab/pinescript-go/runtime/context"
)

func TestEvaluateExpression_Identifier(t *testing.T) {
	/* Create test context with OHLCV data */
	ctx := context.New("TEST", "1h", 3)
	ctx.AddBar(context.OHLCV{Time: 1700000000, Open: 100, High: 105, Low: 95, Close: 102, Volume: 1000})
	ctx.AddBar(context.OHLCV{Time: 1700003600, Open: 102, High: 107, Low: 97, Close: 104, Volume: 1100})
	ctx.AddBar(context.OHLCV{Time: 1700007200, Open: 104, High: 109, Low: 99, Close: 106, Volume: 1200})

	tests := []struct {
		name     string
		field    string
		expected []float64
	}{
		{"close", "close", []float64{102, 104, 106}},
		{"open", "open", []float64{100, 102, 104}},
		{"high", "high", []float64{105, 107, 109}},
		{"low", "low", []float64{95, 97, 99}},
		{"volume", "volume", []float64{1000, 1100, 1200}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr := &ast.Identifier{Name: tt.field}
			values, err := EvaluateExpression(expr, ctx)

			if err != nil {
				t.Fatalf("EvaluateExpression failed: %v", err)
			}

			if len(values) != len(tt.expected) {
				t.Fatalf("Expected %d values, got %d", len(tt.expected), len(values))
			}

			for i, expected := range tt.expected {
				if values[i] != expected {
					t.Errorf("Value[%d]: expected %.2f, got %.2f", i, expected, values[i])
				}
			}
		})
	}
}

func TestEvaluateExpression_TASma(t *testing.T) {
	/* Create context with 5 bars */
	ctx := context.New("TEST", "1D", 5)
	ctx.AddBar(context.OHLCV{Time: 1700000000, Close: 100, Open: 100, High: 100, Low: 100, Volume: 1000})
	ctx.AddBar(context.OHLCV{Time: 1700086400, Close: 102, Open: 102, High: 102, Low: 102, Volume: 1000})
	ctx.AddBar(context.OHLCV{Time: 1700172800, Close: 104, Open: 104, High: 104, Low: 104, Volume: 1000})
	ctx.AddBar(context.OHLCV{Time: 1700259200, Close: 106, Open: 106, High: 106, Low: 106, Volume: 1000})
	ctx.AddBar(context.OHLCV{Time: 1700345600, Close: 108, Open: 108, High: 108, Low: 108, Volume: 1000})

	/* Create ta.sma(close, 3) expression */
	closeExpr := &ast.Identifier{Name: "close"}
	periodExpr := &ast.Literal{Value: float64(3)}

	callExpr := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "ta"},
			Property: &ast.Identifier{Name: "sma"},
		},
		Arguments: []ast.Expression{closeExpr, periodExpr},
	}

	values, err := EvaluateExpression(callExpr, ctx)
	if err != nil {
		t.Fatalf("EvaluateExpression failed: %v", err)
	}

	if len(values) != 5 {
		t.Fatalf("Expected 5 values, got %d", len(values))
	}

	/* First 2 values should be NaN (warmup), then 102, 104, 106 */
	if !math.IsNaN(values[0]) || !math.IsNaN(values[1]) {
		t.Error("Expected first 2 values to be NaN (warmup)")
	}

	expected := []float64{102, 104, 106}
	for i, exp := range expected {
		actual := values[i+2]
		if math.Abs(actual-exp) > 0.01 {
			t.Errorf("SMA[%d]: expected %.2f, got %.2f", i+2, exp, actual)
		}
	}
}

func TestEvaluateExpression_TAEma(t *testing.T) {
	ctx := context.New("TEST", "1h", 4)
	ctx.AddBar(context.OHLCV{Time: 1700000000, Close: 100, Open: 100, High: 100, Low: 100, Volume: 1000})
	ctx.AddBar(context.OHLCV{Time: 1700003600, Close: 110, Open: 110, High: 110, Low: 110, Volume: 1000})
	ctx.AddBar(context.OHLCV{Time: 1700007200, Close: 120, Open: 120, High: 120, Low: 120, Volume: 1000})
	ctx.AddBar(context.OHLCV{Time: 1700010800, Close: 130, Open: 130, High: 130, Low: 130, Volume: 1000})

	callExpr := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "ta"},
			Property: &ast.Identifier{Name: "ema"},
		},
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "close"},
			&ast.Literal{Value: float64(2)},
		},
	}

	values, err := EvaluateExpression(callExpr, ctx)
	if err != nil {
		t.Fatalf("EvaluateExpression failed: %v", err)
	}

	if len(values) != 4 {
		t.Fatalf("Expected 4 values, got %d", len(values))
	}

	/* EMA should have warmup period then calculated values */
	/* Just verify no error and reasonable values */
	for i, v := range values {
		if !math.IsNaN(v) && (v < 90 || v > 140) {
			t.Errorf("EMA[%d]: value %.2f outside reasonable range", i, v)
		}
	}
}

func TestEvaluateExpression_UnsupportedType(t *testing.T) {
	ctx := context.New("TEST", "1h", 1)

	/* BinaryExpression not supported */
	binExpr := &ast.BinaryExpression{
		Operator: "+",
		Left:     &ast.Literal{Value: float64(1)},
		Right:    &ast.Literal{Value: float64(2)},
	}

	_, err := EvaluateExpression(binExpr, ctx)
	if err == nil {
		t.Error("Expected error for unsupported expression type")
	}
}

func TestEvaluateExpression_UnknownIdentifier(t *testing.T) {
	ctx := context.New("TEST", "1h", 1)
	ctx.AddBar(context.OHLCV{Time: 1700000000, Close: 100, Open: 100, High: 100, Low: 100, Volume: 1000})

	expr := &ast.Identifier{Name: "unknown_field"}

	_, err := EvaluateExpression(expr, ctx)
	if err == nil {
		t.Error("Expected error for unknown identifier")
	}
}

func TestEvaluateExpression_TAInsufficientArgs(t *testing.T) {
	ctx := context.New("TEST", "1h", 1)

	/* ta.sma with only 1 argument (needs 2) */
	callExpr := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "ta"},
			Property: &ast.Identifier{Name: "sma"},
		},
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "close"},
		},
	}

	_, err := EvaluateExpression(callExpr, ctx)
	if err == nil {
		t.Error("Expected error for insufficient arguments")
	}
}
