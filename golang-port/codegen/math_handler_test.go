package codegen

import (
	"strings"
	"testing"

	"github.com/borisquantlab/pinescript-go/ast"
)

func TestMathHandler_GenerateMathPow(t *testing.T) {
	mh := NewMathHandler()
	g := &generator{
		variables: make(map[string]string),
		constants: make(map[string]interface{}),
	}

	tests := []struct {
		name     string
		args     []ast.Expression
		expected string
		wantErr  bool
	}{
		{
			name: "literal arguments",
			args: []ast.Expression{
				&ast.Literal{Value: 2.0},
				&ast.Literal{Value: 3.0},
			},
			expected: "math.Pow(2.00, 3.00)",
		},
		{
			name: "identifier arguments",
			args: []ast.Expression{
				&ast.Identifier{Name: "base"},
				&ast.Identifier{Name: "exp"},
			},
			expected: "math.Pow(baseSeries.GetCurrent(), expSeries.GetCurrent())",
		},
		{
			name: "mixed arguments",
			args: []ast.Expression{
				&ast.MemberExpression{
					Object:   &ast.Identifier{Name: "vf"},
					Property: &ast.Literal{Value: float64(0)},
					Computed: true,
				},
				&ast.Literal{Value: -1.0},
			},
			expected: "math.Pow(vfSeries.Get(0), -1.00)",
		},
		{
			name:    "wrong number of args",
			args:    []ast.Expression{&ast.Literal{Value: 2.0}},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := mh.GenerateMathCall("math.pow", tt.args, g)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestMathHandler_GenerateUnaryMath(t *testing.T) {
	mh := NewMathHandler()
	g := &generator{
		variables: make(map[string]string),
		constants: make(map[string]interface{}),
	}

	tests := []struct {
		name     string
		funcName string
		args     []ast.Expression
		expected string
	}{
		{
			name:     "math.abs with literal",
			funcName: "math.abs",
			args: []ast.Expression{
				&ast.Literal{Value: -5.0},
			},
			expected: "math.Abs(-5.00)",
		},
		{
			name:     "math.sqrt with identifier",
			funcName: "math.sqrt",
			args: []ast.Expression{
				&ast.Identifier{Name: "value"},
			},
			expected: "math.Sqrt(valueSeries.GetCurrent())",
		},
		{
			name:     "math.floor with expression",
			funcName: "math.floor",
			args: []ast.Expression{
				&ast.BinaryExpression{
					Operator: "*",
					Left:     &ast.Literal{Value: 1.5},
					Right:    &ast.Literal{Value: 10.0},
				},
			},
			expected: "math.Floor((1.50 * 10.00))",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := mh.GenerateMathCall(tt.funcName, tt.args, g)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestMathHandler_GenerateBinaryMath(t *testing.T) {
	mh := NewMathHandler()
	g := &generator{
		variables: make(map[string]string),
		constants: make(map[string]interface{}),
	}

	tests := []struct {
		name     string
		funcName string
		args     []ast.Expression
		expected string
	}{
		{
			name:     "math.max with literals",
			funcName: "math.max",
			args: []ast.Expression{
				&ast.Literal{Value: 5.0},
				&ast.Literal{Value: 10.0},
			},
			expected: "math.Max(5.00, 10.00)",
		},
		{
			name:     "math.min with identifiers",
			funcName: "math.min",
			args: []ast.Expression{
				&ast.Identifier{Name: "a"},
				&ast.Identifier{Name: "b"},
			},
			expected: "math.Min(aSeries.GetCurrent(), bSeries.GetCurrent())",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := mh.GenerateMathCall(tt.funcName, tt.args, g)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestMathHandler_WithInputConstants(t *testing.T) {
	// Test that math.pow works with input constants
	mh := NewMathHandler()
	g := &generator{
		variables: make(map[string]string),
		constants: map[string]interface{}{
			"yA": "input.float",
		},
	}

	args := []ast.Expression{
		&ast.MemberExpression{
			Object:   &ast.Identifier{Name: "yA"},
			Property: &ast.Literal{Value: float64(0)},
			Computed: true,
		},
		&ast.UnaryExpression{
			Operator: "-",
			Argument: &ast.Literal{Value: 1.0},
		},
	}

	result, err := mh.GenerateMathCall("math.pow", args, g)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should use constant name directly, not Series access
	if !strings.Contains(result, "yA") {
		t.Errorf("expected result to contain 'yA' constant, got %q", result)
	}
	if strings.Contains(result, "yASeries") {
		t.Errorf("result should not use Series access for constant, got %q", result)
	}
}

func TestMathHandler_UnsupportedFunction(t *testing.T) {
	mh := NewMathHandler()
	g := &generator{
		variables: make(map[string]string),
		constants: make(map[string]interface{}),
	}

	args := []ast.Expression{
		&ast.Literal{Value: 1.0},
	}

	_, err := mh.GenerateMathCall("math.unsupported", args, g)
	if err == nil {
		t.Error("expected error for unsupported function, got nil")
	}
}
