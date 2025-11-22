package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestSubscriptResolver_LiteralIndex(t *testing.T) {
	sr := NewSubscriptResolver()
	g := &generator{
		variables: make(map[string]string),
		constants: make(map[string]interface{}),
	}

	tests := []struct {
		name       string
		seriesName string
		indexExpr  ast.Expression
		expected   string
	}{
		{
			name:       "user series with literal 0",
			seriesName: "sma",
			indexExpr:  &ast.Literal{Value: float64(0)},
			expected:   "smaSeries.Get(0)",
		},
		{
			name:       "user series with literal 5",
			seriesName: "ema",
			indexExpr:  &ast.Literal{Value: float64(5)},
			expected:   "emaSeries.Get(5)",
		},
		{
			name:       "close with literal 0",
			seriesName: "close",
			indexExpr:  &ast.Literal{Value: float64(0)},
			expected:   "bar.Close",
		},
		{
			name:       "close with literal 1",
			seriesName: "close",
			indexExpr:  &ast.Literal{Value: float64(1)},
			expected:   "ctx.Data[i-1].Close",
		},
		{
			name:       "open with literal 5",
			seriesName: "open",
			indexExpr:  &ast.Literal{Value: float64(5)},
			expected:   "ctx.Data[i-5].Open",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sr.ResolveSubscript(tt.seriesName, tt.indexExpr, g)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestSubscriptResolver_VariableIndex(t *testing.T) {
	sr := NewSubscriptResolver()
	g := &generator{
		variables: make(map[string]string),
		constants: make(map[string]interface{}),
	}

	tests := []struct {
		name       string
		seriesName string
		indexExpr  ast.Expression
		expected   string
	}{
		{
			name:       "user series with variable index",
			seriesName: "sma",
			indexExpr:  &ast.Identifier{Name: "offset"},
			expected:   "smaSeries.Get(int(offsetSeries.GetCurrent()))",
		},
		{
			name:       "close with variable index",
			seriesName: "close",
			indexExpr:  &ast.Identifier{Name: "nA"},
			expected:   "func() float64 { idx := i - int(nASeries.GetCurrent()); if idx >= 0 && idx < len(ctx.Data) { return ctx.Data[idx].Close } else { return math.NaN() } }()",
		},
		{
			name:       "high with expression index",
			seriesName: "high",
			indexExpr: &ast.BinaryExpression{
				Operator: "*",
				Left:     &ast.Identifier{Name: "period"},
				Right:    &ast.Literal{Value: 2.0},
			},
			expected: "func() float64 { idx := i - int((periodSeries.GetCurrent() * 2.00)); if idx >= 0 && idx < len(ctx.Data) { return ctx.Data[idx].High } else { return math.NaN() } }()",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sr.ResolveSubscript(tt.seriesName, tt.indexExpr, g)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestSubscriptResolver_InputSourceAlias(t *testing.T) {
	// Test that input.source is correctly aliased to close
	sr := NewSubscriptResolver()
	g := &generator{
		variables: make(map[string]string),
		constants: map[string]interface{}{
			"src": "input.source",
		},
	}

	tests := []struct {
		name      string
		indexExpr ast.Expression
		expected  string
	}{
		{
			name:      "literal 0",
			indexExpr: &ast.Literal{Value: float64(0)},
			expected:  "bar.Close",
		},
		{
			name:      "literal 1",
			indexExpr: &ast.Literal{Value: float64(1)},
			expected:  "ctx.Data[i-1].Close",
		},
		{
			name:      "variable index",
			indexExpr: &ast.Identifier{Name: "nA"},
			expected:  "func() float64 { idx := i - int(nASeries.GetCurrent()); if idx >= 0 && idx < len(ctx.Data) { return ctx.Data[idx].Close } else { return math.NaN() } }()",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sr.ResolveSubscript("src", tt.indexExpr, g)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestSubscriptResolver_BoundsChecking(t *testing.T) {
	// Verify bounds checking is present for variable indices on built-in series
	sr := NewSubscriptResolver()
	g := &generator{
		variables: make(map[string]string),
		constants: make(map[string]interface{}),
	}

	indexExpr := &ast.Identifier{Name: "offset"}
	result := sr.ResolveSubscript("close", indexExpr, g)

	// Should contain bounds check
	if !strings.Contains(result, "idx >= 0 && idx < len(ctx.Data)") {
		t.Errorf("result missing bounds check: %s", result)
	}
	if !strings.Contains(result, "math.NaN()") {
		t.Errorf("result missing NaN fallback: %s", result)
	}
}

func TestSubscriptResolver_AllBuiltinSeries(t *testing.T) {
	sr := NewSubscriptResolver()
	g := &generator{
		variables: make(map[string]string),
		constants: make(map[string]interface{}),
	}

	builtins := []string{"close", "open", "high", "low", "volume"}
	indexExpr := &ast.Identifier{Name: "n"}

	for _, builtin := range builtins {
		t.Run(builtin, func(t *testing.T) {
			result := sr.ResolveSubscript(builtin, indexExpr, g)

			// Should NOT use builtin name + "Series" pattern (e.g., "closeSeries")
			builtinSeries := builtin + "Series"
			if strings.Contains(result, builtinSeries) {
				t.Errorf("builtin %s should not use %s: %s", builtin, builtinSeries, result)
			}
			// Should use ctx.Data access
			if !strings.Contains(result, "ctx.Data") {
				t.Errorf("builtin %s should use ctx.Data: %s", builtin, result)
			}
		})
	}
}
