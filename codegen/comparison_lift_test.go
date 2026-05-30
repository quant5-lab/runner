package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

// TestIsComparisonOperator verifies that the operator classifier correctly
// distinguishes the six Pine comparison operators from all arithmetic and
// logical operators.
func TestIsComparisonOperator(t *testing.T) {
	tests := []struct {
		op   string
		want bool
	}{
		{">", true},
		{"<", true},
		{">=", true},
		{"<=", true},
		{"==", true},
		{"!=", true},

		{"+", false},
		{"-", false},
		{"*", false},
		{"/", false},
		{"%", false},

		// Logical operators (handled by LogicalExpression, not BinaryExpression)
		{"and", false},
		{"or", false},
		{"&&", false},
		{"||", false},

		{"", false},
		{"=", false},
		{"!", false},
	}

	for _, tt := range tests {
		t.Run(tt.op, func(t *testing.T) {
			got := isComparisonOperator(tt.op)
			if got != tt.want {
				t.Errorf("isComparisonOperator(%q) = %v, want %v", tt.op, got, tt.want)
			}
		})
	}
}

// TestLiftComparisonToFloat64 verifies the promotion helper used by
// extractSeriesExpression when a LogicalExpression child is itself a
// BinaryExpression that returns a Go bool.
//
// Comparison children must be wrapped in a float64 IIFE so that
// value.IsTrue(float64) can accept them at the call site.
// All other expression types are returned unchanged.
func TestLiftComparisonToFloat64(t *testing.T) {
	const code = "aSeries.GetCurrent() OP bSeries.GetCurrent()"

	comparisonOps := []string{">", "<", ">=", "<=", "==", "!="}
	arithmeticOps := []string{"+", "-", "*", "/"}

	t.Run("comparison operators produce float64 IIFE", func(t *testing.T) {
		for _, op := range comparisonOps {
			op := op
			t.Run(op, func(t *testing.T) {
				expr := &ast.BinaryExpression{
					Operator: op,
					Left:     &ast.Identifier{Name: "a"},
					Right:    &ast.Identifier{Name: "b"},
				}
				result := liftComparisonToFloat64(expr, code)

				if !strings.HasPrefix(result, "func() float64 {") {
					t.Errorf("operator %q: expected IIFE prefix, got: %s", op, result)
				}
				if !strings.Contains(result, code) {
					t.Errorf("operator %q: original code must appear inside IIFE, got: %s", op, result)
				}
				if !strings.Contains(result, "return 1.0") || !strings.Contains(result, "return 0.0") {
					t.Errorf("operator %q: IIFE must return 1.0 / 0.0, got: %s", op, result)
				}
			})
		}
	})

	t.Run("arithmetic operators pass through unchanged", func(t *testing.T) {
		for _, op := range arithmeticOps {
			op := op
			t.Run(op, func(t *testing.T) {
				expr := &ast.BinaryExpression{
					Operator: op,
					Left:     &ast.Identifier{Name: "a"},
					Right:    &ast.Identifier{Name: "b"},
				}
				result := liftComparisonToFloat64(expr, code)
				if result != code {
					t.Errorf("operator %q: expected passthrough, got: %s", op, result)
				}
			})
		}
	})

	t.Run("non-binary expressions pass through unchanged", func(t *testing.T) {
		nonBinaryExprs := []struct {
			name string
			expr ast.Expression
		}{
			{"identifier", &ast.Identifier{Name: "x"}},
			{"float literal", &ast.Literal{Value: 1.0}},
			{"bool literal", &ast.Literal{Value: true}},
			{"logical expression", &ast.LogicalExpression{Operator: "and", Left: &ast.Identifier{Name: "a"}, Right: &ast.Identifier{Name: "b"}}},
			{"unary expression", &ast.UnaryExpression{Operator: "not", Argument: &ast.Identifier{Name: "x"}}},
		}
		for _, tc := range nonBinaryExprs {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				result := liftComparisonToFloat64(tc.expr, code)
				if result != code {
					t.Errorf("expr type %T: expected passthrough, got: %s", tc.expr, result)
				}
			})
		}
	})
}
