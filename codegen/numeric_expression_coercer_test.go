package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

/* TestNumericExpressionCoercer_LiteralCoercion validates all literal types */
func TestNumericExpressionCoercer_LiteralCoercion(t *testing.T) {
	coercer := NewNumericExpressionCoercer(newTestBoolConverter())

	tests := []struct {
		name     string
		expr     ast.Expression
		code     string
		expected string
	}{
		{
			name:     "true literal becomes 1.0",
			expr:     &ast.Literal{Value: true},
			code:     "true",
			expected: "1.0",
		},
		{
			name:     "false literal becomes 0.0",
			expr:     &ast.Literal{Value: false},
			code:     "false",
			expected: "0.0",
		},
		{
			name:     "float literal passes through",
			expr:     &ast.Literal{Value: 42.0},
			code:     "42.0",
			expected: "42.0",
		},
		{
			name:     "integer literal passes through",
			expr:     &ast.Literal{Value: 7},
			code:     "7",
			expected: "7",
		},
		{
			name:     "string literal passes through",
			expr:     &ast.Literal{Value: "hello"},
			code:     `"hello"`,
			expected: `"hello"`,
		},
		{
			name:     "nil literal passes through",
			expr:     &ast.Literal{Value: nil},
			code:     "math.NaN()",
			expected: "math.NaN()",
		},
		{
			name:     "zero float passes through",
			expr:     &ast.Literal{Value: 0.0},
			code:     "0.0",
			expected: "0.0",
		},
		{
			name:     "negative float passes through",
			expr:     &ast.Literal{Value: -3.14},
			code:     "-3.14",
			expected: "-3.14",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := coercer.CoerceToFloat64(tt.expr, tt.code)
			if result != tt.expected {
				t.Errorf("CoerceToFloat64() = %q, want %q", result, tt.expected)
			}
		})
	}
}

/* TestNumericExpressionCoercer_BooleanExpressionCoercion validates IIFE wrapping for all boolean expression types */
func TestNumericExpressionCoercer_BooleanExpressionCoercion(t *testing.T) {
	coercer := NewNumericExpressionCoercer(newTestBoolConverter())

	tests := []struct {
		name     string
		expr     ast.Expression
		code     string
		expected string
	}{
		{
			name:     "comparison greater-than",
			expr:     &ast.BinaryExpression{Operator: ">", Left: &ast.Identifier{Name: "a"}, Right: &ast.Identifier{Name: "b"}},
			code:     "(a > b)",
			expected: "func() float64 { if (a > b) { return 1.0 } else { return 0.0 } }()",
		},
		{
			name:     "comparison less-than",
			expr:     &ast.BinaryExpression{Operator: "<", Left: &ast.Identifier{Name: "x"}, Right: &ast.Literal{Value: 10.0}},
			code:     "(x < 10.0)",
			expected: "func() float64 { if (x < 10.0) { return 1.0 } else { return 0.0 } }()",
		},
		{
			name:     "comparison equality",
			expr:     &ast.BinaryExpression{Operator: "==", Left: &ast.Identifier{Name: "a"}, Right: &ast.Literal{Value: 0.0}},
			code:     "(a == 0.0)",
			expected: "func() float64 { if (a == 0.0) { return 1.0 } else { return 0.0 } }()",
		},
		{
			name:     "comparison not-equal",
			expr:     &ast.BinaryExpression{Operator: "!=", Left: &ast.Identifier{Name: "a"}, Right: &ast.Literal{Value: 0.0}},
			code:     "(a != 0.0)",
			expected: "func() float64 { if (a != 0.0) { return 1.0 } else { return 0.0 } }()",
		},
		{
			name:     "comparison gte",
			expr:     &ast.BinaryExpression{Operator: ">=", Left: &ast.Identifier{Name: "a"}, Right: &ast.Identifier{Name: "b"}},
			code:     "(a >= b)",
			expected: "func() float64 { if (a >= b) { return 1.0 } else { return 0.0 } }()",
		},
		{
			name:     "comparison lte",
			expr:     &ast.BinaryExpression{Operator: "<=", Left: &ast.Identifier{Name: "a"}, Right: &ast.Identifier{Name: "b"}},
			code:     "(a <= b)",
			expected: "func() float64 { if (a <= b) { return 1.0 } else { return 0.0 } }()",
		},
		{
			name:     "logical and",
			expr:     &ast.LogicalExpression{Operator: "and", Left: &ast.Identifier{Name: "x"}, Right: &ast.Identifier{Name: "y"}},
			code:     "(x && y)",
			expected: "func() float64 { if (x && y) { return 1.0 } else { return 0.0 } }()",
		},
		{
			name:     "logical or",
			expr:     &ast.LogicalExpression{Operator: "or", Left: &ast.Identifier{Name: "a"}, Right: &ast.Identifier{Name: "b"}},
			code:     "(a || b)",
			expected: "func() float64 { if (a || b) { return 1.0 } else { return 0.0 } }()",
		},
		{
			name:     "unary not",
			expr:     &ast.UnaryExpression{Operator: "not", Argument: &ast.Identifier{Name: "x"}},
			code:     "!x",
			expected: "func() float64 { if !x { return 1.0 } else { return 0.0 } }()",
		},
		{
			name:     "unary bang",
			expr:     &ast.UnaryExpression{Operator: "!", Argument: &ast.Identifier{Name: "flag"}},
			code:     "!flag",
			expected: "func() float64 { if !flag { return 1.0 } else { return 0.0 } }()",
		},
		{
			name: "boolean function call (na)",
			expr: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "na"},
			},
			code:     "math.IsNaN(x)",
			expected: "func() float64 { if math.IsNaN(x) { return 1.0 } else { return 0.0 } }()",
		},
		{
			name: "boolean function call (ta.crossover)",
			expr: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ta"},
					Property: &ast.Identifier{Name: "crossover"},
				},
			},
			code:     "ta.Crossover(fast, slow)",
			expected: "func() float64 { if ta.Crossover(fast, slow) { return 1.0 } else { return 0.0 } }()",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := coercer.CoerceToFloat64(tt.expr, tt.code)
			if result != tt.expected {
				t.Errorf("CoerceToFloat64() = %q, want %q", result, tt.expected)
			}
		})
	}
}

/* TestNumericExpressionCoercer_NonBooleanPassthrough validates non-boolean expressions pass unmodified */
func TestNumericExpressionCoercer_NonBooleanPassthrough(t *testing.T) {
	coercer := NewNumericExpressionCoercer(newTestBoolConverter())

	tests := []struct {
		name     string
		expr     ast.Expression
		code     string
		expected string
	}{
		{
			name:     "arithmetic addition",
			expr:     &ast.BinaryExpression{Operator: "+", Left: &ast.Identifier{Name: "a"}, Right: &ast.Identifier{Name: "b"}},
			code:     "(a + b)",
			expected: "(a + b)",
		},
		{
			name:     "arithmetic multiplication",
			expr:     &ast.BinaryExpression{Operator: "*", Left: &ast.Identifier{Name: "x"}, Right: &ast.Literal{Value: 2.0}},
			code:     "(x * 2.0)",
			expected: "(x * 2.0)",
		},
		{
			name:     "arithmetic modulo",
			expr:     &ast.BinaryExpression{Operator: "%", Left: &ast.Identifier{Name: "a"}, Right: &ast.Literal{Value: 3.0}},
			code:     "math.Mod(a, 3.0)",
			expected: "math.Mod(a, 3.0)",
		},
		{
			name:     "identifier",
			expr:     &ast.Identifier{Name: "x"},
			code:     "x",
			expected: "x",
		},
		{
			name:     "non-boolean function call",
			expr:     &ast.CallExpression{Callee: &ast.Identifier{Name: "math.sqrt"}},
			code:     "math.Sqrt(x)",
			expected: "math.Sqrt(x)",
		},
		{
			name:     "non-boolean TA function call",
			expr:     &ast.CallExpression{Callee: &ast.Identifier{Name: "ta.sma"}},
			code:     "ta.Sma(close, 14)",
			expected: "ta.Sma(close, 14)",
		},
		{
			name: "member expression",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "bar"},
				Property: &ast.Identifier{Name: "Close"},
			},
			code:     "bar.Close",
			expected: "bar.Close",
		},
		{
			name: "conditional expression (not boolean itself)",
			expr: &ast.ConditionalExpression{
				Test:       &ast.BinaryExpression{Operator: ">", Left: &ast.Identifier{Name: "a"}, Right: &ast.Identifier{Name: "b"}},
				Consequent: &ast.Identifier{Name: "a"},
				Alternate:  &ast.Identifier{Name: "b"},
			},
			code:     "func() float64 { if (a > b) { return a } else { return b } }()",
			expected: "func() float64 { if (a > b) { return a } else { return b } }()",
		},
		{
			name:     "unary negation (not boolean)",
			expr:     &ast.UnaryExpression{Operator: "-", Argument: &ast.Identifier{Name: "val"}},
			code:     "-val",
			expected: "-val",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := coercer.CoerceToFloat64(tt.expr, tt.code)
			if result != tt.expected {
				t.Errorf("CoerceToFloat64() = %q, want %q", result, tt.expected)
			}
		})
	}
}

/* TestNumericExpressionCoercer_IIFEStructure validates the generated IIFE follows Go syntax */
func TestNumericExpressionCoercer_IIFEStructure(t *testing.T) {
	coercer := NewNumericExpressionCoercer(newTestBoolConverter())

	expr := &ast.BinaryExpression{Operator: ">", Left: &ast.Identifier{Name: "a"}, Right: &ast.Identifier{Name: "b"}}
	result := coercer.CoerceToFloat64(expr, "(a > b)")

	checks := []struct {
		name     string
		contains string
	}{
		{"self-invoking function", "func() float64 {"},
		{"if condition", "if (a > b)"},
		{"true branch returns 1.0", "return 1.0"},
		{"false branch returns 0.0", "return 0.0"},
		{"immediate invocation", "}()"},
	}

	for _, check := range checks {
		if !strings.Contains(result, check.contains) {
			t.Errorf("%s: IIFE missing %q\ngot: %s", check.name, check.contains, result)
		}
	}
}

/* TestNumericExpressionCoercer_BoolLiteralPriority verifies bool literal path takes precedence over IsAlreadyBoolean */
func TestNumericExpressionCoercer_BoolLiteralPriority(t *testing.T) {
	coercer := NewNumericExpressionCoercer(newTestBoolConverter())

	/* Bool literal is both a literal AND recognized by IsAlreadyBoolean — must return "1.0" not IIFE */
	result := coercer.CoerceToFloat64(&ast.Literal{Value: true}, "true")
	if result != "1.0" {
		t.Errorf("expected direct 1.0 for bool literal, got %q", result)
	}

	result = coercer.CoerceToFloat64(&ast.Literal{Value: false}, "false")
	if result != "0.0" {
		t.Errorf("expected direct 0.0 for bool literal, got %q", result)
	}
}

func newTestBoolConverter() *BooleanConverter {
	typeSystem := NewTypeInferenceEngine()
	return NewBooleanConverter(typeSystem)
}
