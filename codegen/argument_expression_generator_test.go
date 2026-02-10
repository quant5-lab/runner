package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func newTestArgumentExpressionGenerator(funcName string, paramIdx int) *ArgumentExpressionGenerator {
	gen := newTestGenerator()
	return NewArgumentExpressionGenerator(gen, funcName, paramIdx)
}

func TestArgumentExpressionGenerator_ExpressionTypeDispatch(t *testing.T) {
	tests := []struct {
		name     string
		expr     ast.Expression
		expected string
	}{
		{
			name:     "literal float",
			expr:     &ast.Literal{Value: 3.14},
			expected: "3.1",
		},
		{
			name:     "literal int",
			expr:     &ast.Literal{Value: 42},
			expected: "42.0",
		},
		{
			name:     "identifier user variable",
			expr:     &ast.Identifier{Name: "src"},
			expected: "srcSeries.GetCurrent()",
		},
		{
			name: "binary expression",
			expr: &ast.BinaryExpression{
				Operator: "+",
				Left:     &ast.Identifier{Name: "a"},
				Right:    &ast.Literal{Value: 1.0},
			},
			expected: "(aSeries.GetCurrent() + 1.0)",
		},
		{
			name: "unary negation",
			expr: &ast.UnaryExpression{
				Operator: "-",
				Argument: &ast.Literal{Value: 5.0},
				Prefix:   true,
			},
			expected: "-5.0",
		},
		{
			name: "unary not",
			expr: &ast.UnaryExpression{
				Operator: "not",
				Argument: &ast.Identifier{Name: "flag"},
				Prefix:   true,
			},
			expected: "func() float64 { if !(value.IsTrue(flagSeries.GetCurrent())) { return 1.0 } else { return 0.0 } }()",
		},
		{
			name: "logical and",
			expr: &ast.LogicalExpression{
				Operator: "and",
				Left:     &ast.Identifier{Name: "a"},
				Right:    &ast.Identifier{Name: "b"},
			},
			expected: "func() float64 { if ((value.IsTrue(aSeries.GetCurrent())) && (value.IsTrue(bSeries.GetCurrent()))) { return 1.0 } else { return 0.0 } }()",
		},
		{
			name: "logical or",
			expr: &ast.LogicalExpression{
				Operator: "or",
				Left:     &ast.Identifier{Name: "x"},
				Right:    &ast.Identifier{Name: "y"},
			},
			expected: "func() float64 { if ((value.IsTrue(xSeries.GetCurrent())) || (value.IsTrue(ySeries.GetCurrent()))) { return 1.0 } else { return 0.0 } }()",
		},
		{
			name: "conditional ternary",
			expr: &ast.ConditionalExpression{
				Test: &ast.BinaryExpression{
					Operator: ">",
					Left:     &ast.Identifier{Name: "a"},
					Right:    &ast.Literal{Value: 0.0},
				},
				Consequent: &ast.Literal{Value: 1.0},
				Alternate:  &ast.Literal{Value: 0.0},
			},
			expected: "func() float64 { if (aSeries.GetCurrent() > 0.0) { return 1.0 } else { return 0.0 } }()",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			aeg := newTestArgumentExpressionGenerator("myFunc", 0)
			result, err := aeg.Generate(tt.expr)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("got:\n  %s\nwant:\n  %s", result, tt.expected)
			}
		})
	}
}

func TestArgumentExpressionGenerator_UnsupportedTypeReturnsError(t *testing.T) {
	aeg := newTestArgumentExpressionGenerator("myFunc", 0)

	_, err := aeg.Generate(&ast.ArrowFunctionExpression{})
	if err == nil {
		t.Fatal("expected error for unsupported expression type")
	}
	if !strings.Contains(err.Error(), "unsupported argument expression type") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestArgumentExpressionGenerator_NestedComposition(t *testing.T) {
	tests := []struct {
		name           string
		expr           ast.Expression
		expectContains []string
	}{
		{
			name: "unary negation inside binary multiplication",
			expr: &ast.BinaryExpression{
				Operator: "*",
				Left:     &ast.Identifier{Name: "x"},
				Right: &ast.UnaryExpression{
					Operator: "-",
					Argument: &ast.Literal{Value: 0.5},
					Prefix:   true,
				},
			},
			expectContains: []string{"xSeries.GetCurrent()", "*", "-0.5"},
		},
		{
			name: "logical inside conditional test",
			expr: &ast.ConditionalExpression{
				Test: &ast.LogicalExpression{
					Operator: "and",
					Left: &ast.BinaryExpression{
						Operator: ">",
						Left:     &ast.Identifier{Name: "a"},
						Right:    &ast.Literal{Value: 0.0},
					},
					Right: &ast.BinaryExpression{
						Operator: "<",
						Left:     &ast.Identifier{Name: "b"},
						Right:    &ast.Literal{Value: 100.0},
					},
				},
				Consequent: &ast.Literal{Value: 1.0},
				Alternate:  &ast.Literal{Value: 0.0},
			},
			expectContains: []string{"if", "&&", "return 1.0", "return 0.0"},
		},
		{
			name: "conditional inside binary addition",
			expr: &ast.BinaryExpression{
				Operator: "+",
				Left:     &ast.Identifier{Name: "base"},
				Right: &ast.ConditionalExpression{
					Test: &ast.BinaryExpression{
						Operator: ">",
						Left:     &ast.Identifier{Name: "x"},
						Right:    &ast.Literal{Value: 0.0},
					},
					Consequent: &ast.Literal{Value: 10.0},
					Alternate:  &ast.Literal{Value: 0.0},
				},
			},
			expectContains: []string{"baseSeries.GetCurrent()", "+", "func() float64"},
		},
		{
			name: "not wrapping logical expression",
			expr: &ast.UnaryExpression{
				Operator: "not",
				Argument: &ast.LogicalExpression{
					Operator: "or",
					Left:     &ast.Identifier{Name: "a"},
					Right:    &ast.Identifier{Name: "b"},
				},
				Prefix: true,
			},
			expectContains: []string{"!", "||"},
		},
		{
			name: "binary inside unary negation",
			expr: &ast.UnaryExpression{
				Operator: "-",
				Argument: &ast.BinaryExpression{
					Operator: "+",
					Left:     &ast.Identifier{Name: "a"},
					Right:    &ast.Identifier{Name: "b"},
				},
				Prefix: true,
			},
			expectContains: []string{"-(aSeries.GetCurrent() + bSeries.GetCurrent())"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			aeg := newTestArgumentExpressionGenerator("myFunc", 0)
			result, err := aeg.Generate(tt.expr)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			for _, sub := range tt.expectContains {
				if !strings.Contains(result, sub) {
					t.Errorf("result %q missing expected substring %q", result, sub)
				}
			}
		})
	}
}

func TestArgumentExpressionGenerator_UDFCompoundArguments(t *testing.T) {
	tests := []struct {
		name        string
		pine        string
		mustContain []string
	}{
		{
			name: "unary negation in UDF argument",
			pine: `
//@version=5
indicator("Test")
negate(v) =>
    -v
result = negate(-close)
`,
			mustContain: []string{
				"negate(arrowCtx_negate_1,",
			},
		},
		{
			name: "binary with negated literal coefficient",
			pine: `
//@version=5
indicator("Test")
scale(v) =>
    v * 100
result = scale(close * -0.5)
`,
			mustContain: []string{
				"scale(arrowCtx_scale_1,",
			},
		},
		{
			name: "conditional ternary as UDF argument",
			pine: `
//@version=5
indicator("Test")
pick(v) =>
    v * 2
result = pick(close > open ? high : low)
`,
			mustContain: []string{
				"pick(arrowCtx_pick_1,",
				"func() float64",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := compilePineScript(tt.pine)
			if err != nil {
				t.Fatalf("compilation failed: %v", err)
			}
			for _, pattern := range tt.mustContain {
				if !strings.Contains(code, pattern) {
					t.Errorf("missing expected pattern: %q\ngenerated code:\n%s", pattern, code)
				}
			}
		})
	}
}

func TestArgumentExpressionGenerator_IfForStatementDispatch(t *testing.T) {
	t.Run("IfStatement produces IIFE returning float64", func(t *testing.T) {
		aeg := newTestArgumentExpressionGenerator("plot", 0)
		ifStmt := &ast.IfStatement{
			Test: &ast.BinaryExpression{
				Operator: ">",
				Left:     &ast.Identifier{Name: "close"},
				Right:    &ast.Literal{Value: 100.0},
			},
			Consequent: []ast.Node{
				&ast.ExpressionStatement{
					Expression: &ast.Literal{Value: 1.0},
				},
			},
			Alternate: []ast.Node{
				&ast.ExpressionStatement{
					Expression: &ast.Literal{Value: 0.0},
				},
			},
		}

		result, err := aeg.Generate(ifStmt)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		for _, sub := range []string{"func() float64", "if"} {
			if !strings.Contains(result, sub) {
				t.Errorf("result %q missing expected substring %q", result, sub)
			}
		}
	})

	t.Run("ForStatement produces IIFE returning float64", func(t *testing.T) {
		aeg := newTestArgumentExpressionGenerator("plot", 0)
		forStmt := &ast.ForStatement{
			Counter: "i",
			From:    &ast.Literal{Value: 0},
			To:      &ast.Literal{Value: 10},
			Body: []ast.Node{
				&ast.ExpressionStatement{
					Expression: &ast.Identifier{Name: "i"},
				},
			},
		}

		result, err := aeg.Generate(forStmt)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		for _, sub := range []string{"func() float64", "for"} {
			if !strings.Contains(result, sub) {
				t.Errorf("result %q missing expected substring %q", result, sub)
			}
		}
	})
}
