package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestLogicalExpressionInSeriesExpressions(t *testing.T) {
	gen := &generator{
		variables:      make(map[string]string),
		varInits:       make(map[string]ast.Expression),
		constants:      make(map[string]interface{}),
		builtinHandler: NewBuiltinIdentifierHandler(),
	}

	gen.variables["newmoon"] = "float"
	gen.variables["fullmoon"] = "float"
	gen.variables["condition"] = "float"
	gen.variables["signal"] = "float"

	tests := []struct {
		name     string
		expr     ast.Expression
		mustHave []string
		mustNot  []string
	}{
		{
			name: "simple or expression",
			expr: &ast.LogicalExpression{
				Operator: "or",
				Left:     &ast.Identifier{Name: "newmoon"},
				Right:    &ast.Identifier{Name: "fullmoon"},
			},
			mustHave: []string{
				"func() float64",
				"value.IsTrue(newmoonSeries.GetCurrent())",
				"||",
				"value.IsTrue(fullmoonSeries.GetCurrent())",
				"return 1.0",
			},
			mustNot: []string{"or"},
		},
		{
			name: "simple and expression",
			expr: &ast.LogicalExpression{
				Operator: "and",
				Left:     &ast.Identifier{Name: "condition"},
				Right:    &ast.Identifier{Name: "signal"},
			},
			mustHave: []string{
				"func() float64",
				"value.IsTrue(conditionSeries.GetCurrent())",
				"&&",
				"value.IsTrue(signalSeries.GetCurrent())",
				"return 1.0",
			},
			mustNot: []string{"and"},
		},
		{
			name: "nested or within and",
			expr: &ast.LogicalExpression{
				Operator: "and",
				Left: &ast.LogicalExpression{
					Operator: "or",
					Left:     &ast.Identifier{Name: "newmoon"},
					Right:    &ast.Identifier{Name: "fullmoon"},
				},
				Right: &ast.Identifier{Name: "condition"},
			},
			mustHave: []string{
				"func() float64",
				"value.IsTrue(newmoonSeries.GetCurrent())",
				"||",
				"value.IsTrue(fullmoonSeries.GetCurrent())",
				"&&",
				"value.IsTrue(conditionSeries.GetCurrent())",
			},
			mustNot: []string{"or", "and"},
		},
		{
			name: "nested and within or",
			expr: &ast.LogicalExpression{
				Operator: "or",
				Left:     &ast.Identifier{Name: "signal"},
				Right: &ast.LogicalExpression{
					Operator: "and",
					Left:     &ast.Identifier{Name: "newmoon"},
					Right:    &ast.Identifier{Name: "fullmoon"},
				},
			},
			mustHave: []string{
				"func() float64",
				"value.IsTrue(signalSeries.GetCurrent())",
				"||",
				"value.IsTrue(newmoonSeries.GetCurrent())",
				"&&",
				"value.IsTrue(fullmoonSeries.GetCurrent())",
			},
			mustNot: []string{"or", "and"},
		},
		{
			name: "logical with literal true",
			expr: &ast.LogicalExpression{
				Operator: "or",
				Left:     &ast.Identifier{Name: "condition"},
				Right:    &ast.Literal{Value: true},
			},
			mustHave: []string{
				"value.IsTrue(conditionSeries.GetCurrent())",
				"||",
				"value.IsTrue(1.0)",
			},
			mustNot: []string{"or"},
		},
		{
			name: "logical with literal false",
			expr: &ast.LogicalExpression{
				Operator: "and",
				Left:     &ast.Literal{Value: false},
				Right:    &ast.Identifier{Name: "signal"},
			},
			mustHave: []string{
				"value.IsTrue(0.0)",
				"&&",
				"value.IsTrue(signalSeries.GetCurrent())",
			},
			mustNot: []string{"and"},
		},
		{
			name: "logical with binary expression",
			expr: &ast.LogicalExpression{
				Operator: "and",
				Left: &ast.BinaryExpression{
					Operator: ">",
					Left:     &ast.Identifier{Name: "newmoon"},
					Right:    &ast.Literal{Value: 0.0},
				},
				Right: &ast.Identifier{Name: "condition"},
			},
			mustHave: []string{
				"value.IsTrue(",
				"newmoonSeries.GetCurrent()",
				">",
				"&&",
				"value.IsTrue(conditionSeries.GetCurrent())",
			},
			mustNot: []string{"and"},
		},
		{
			name: "three-level nested logical",
			expr: &ast.LogicalExpression{
				Operator: "or",
				Left: &ast.LogicalExpression{
					Operator: "and",
					Left:     &ast.Identifier{Name: "newmoon"},
					Right:    &ast.Identifier{Name: "fullmoon"},
				},
				Right: &ast.LogicalExpression{
					Operator: "and",
					Left:     &ast.Identifier{Name: "condition"},
					Right:    &ast.Identifier{Name: "signal"},
				},
			},
			mustHave: []string{
				"func() float64",
				"value.IsTrue(newmoonSeries.GetCurrent())",
				"&&",
				"value.IsTrue(fullmoonSeries.GetCurrent())",
				"||",
				"value.IsTrue(conditionSeries.GetCurrent())",
				"&&",
				"value.IsTrue(signalSeries.GetCurrent())",
			},
			mustNot: []string{"or", "and"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := gen.extractSeriesExpression(tt.expr)

			for _, must := range tt.mustHave {
				if !strings.Contains(result, must) {
					t.Errorf("expected to contain: %s\ngot: %s", must, result)
				}
			}
			for _, mustNot := range tt.mustNot {
				if strings.Contains(result, mustNot) {
					t.Errorf("should NOT contain: %s\ngot: %s", mustNot, result)
				}
			}
		})
	}
}

func TestLogicalExpressionOperatorNormalization(t *testing.T) {
	gen := &generator{
		variables:      make(map[string]string),
		varInits:       make(map[string]ast.Expression),
		constants:      make(map[string]interface{}),
		builtinHandler: NewBuiltinIdentifierHandler(),
	}

	gen.variables["a"] = "float"
	gen.variables["b"] = "float"

	tests := []struct {
		operator     string
		expectedGoOp string
	}{
		{"or", "||"},
		{"and", "&&"},
	}

	for _, tt := range tests {
		t.Run("operator_"+tt.operator, func(t *testing.T) {
			expr := &ast.LogicalExpression{
				Operator: tt.operator,
				Left:     &ast.Identifier{Name: "a"},
				Right:    &ast.Identifier{Name: "b"},
			}

			result := gen.extractSeriesExpression(expr)

			if !strings.Contains(result, tt.expectedGoOp) {
				t.Errorf("expected Go operator %q in result, got: %s", tt.expectedGoOp, result)
			}
			if strings.Contains(result, tt.operator) {
				t.Errorf("PineScript operator %q should be normalized to %q, got: %s",
					tt.operator, tt.expectedGoOp, result)
			}
		})
	}
}

func TestLogicalExpressionWithSeriesOffset(t *testing.T) {
	gen := &generator{
		variables:      make(map[string]string),
		varInits:       make(map[string]ast.Expression),
		constants:      make(map[string]interface{}),
		builtinHandler: NewBuiltinIdentifierHandler(),
	}

	gen.variables["condition"] = "float"
	gen.variables["signal"] = "float"

	tests := []struct {
		name     string
		expr     ast.Expression
		mustHave []string
	}{
		{
			name: "logical with subscripted series",
			expr: &ast.LogicalExpression{
				Operator: "or",
				Left: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "condition"},
					Property: &ast.Literal{Value: 1.0},
					Computed: true,
				},
				Right: &ast.Identifier{Name: "signal"},
			},
			mustHave: []string{
				"value.IsTrue(conditionSeries.Get(1))",
				"||",
				"value.IsTrue(signalSeries.GetCurrent())",
			},
		},
		{
			name: "logical with both sides subscripted",
			expr: &ast.LogicalExpression{
				Operator: "and",
				Left: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "condition"},
					Property: &ast.Literal{Value: 0.0},
					Computed: true,
				},
				Right: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "signal"},
					Property: &ast.Literal{Value: 1.0},
					Computed: true,
				},
			},
			mustHave: []string{
				"value.IsTrue(conditionSeries.Get(0))",
				"&&",
				"value.IsTrue(signalSeries.Get(1))",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := gen.extractSeriesExpression(tt.expr)

			for _, must := range tt.mustHave {
				if !strings.Contains(result, must) {
					t.Errorf("expected to contain: %s\ngot: %s", must, result)
				}
			}
		})
	}
}
