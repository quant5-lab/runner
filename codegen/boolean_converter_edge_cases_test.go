package codegen

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestBooleanConverter_ComprehensiveEdgeCases(t *testing.T) {
	typeSystem := NewTypeInferenceEngine()
	converter := NewBooleanConverter(typeSystem)

	tests := []struct {
		name          string
		expr          ast.Expression
		generatedCode string
		want          string
	}{
		{
			name:          "NaN literal",
			expr:          &ast.Literal{Value: "NaN"},
			generatedCode: "math.NaN()",
			want:          "math.NaN()", // Literals are skipped
		},
		{
			name:          "Zero literal",
			expr:          &ast.Literal{Value: 0.0},
			generatedCode: "0.0",
			want:          "0.0", // Literals are skipped
		},
		{
			name:          "Float variable access",
			expr:          &ast.Identifier{Name: "myFloat"},
			generatedCode: "myFloatSeries.GetCurrent()",
			want:          "value.IsTrue(myFloatSeries.GetCurrent())",
		},
		{
			name: "Logical Expression (AND)",
			expr: &ast.LogicalExpression{
				Operator: "and",
				Left:     &ast.Identifier{Name: "a"},
				Right:    &ast.Identifier{Name: "b"},
			},
			generatedCode: "a && b",
			want:          "a && b", // Should not be wrapped
		},
		{
			name: "Comparison Expression (>)",
			expr: &ast.BinaryExpression{
				Operator: ">",
				Left:     &ast.Identifier{Name: "a"},
				Right:    &ast.Identifier{Name: "b"},
			},
			generatedCode: "a > b",
			want:          "a > b", // Should not be wrapped
		},
		{
			name: "Function call (na)",
			expr: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "na"},
			},
			generatedCode: "math.IsNaN(x)",
			want:          "math.IsNaN(x)", // Should not be wrapped
		},
		{
			name: "Function call (unknown/float)",
			expr: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "sma"},
			},
			generatedCode: "ta.Sma(x, 10)",
			want:          "value.IsTrue(ta.Sma(x, 10))", // Should be wrapped
		},
		{
			name: "Unary NOT expression",
			expr: &ast.UnaryExpression{
				Operator: "not",
				Argument: &ast.Identifier{Name: "x"},
			},
			generatedCode: "!x",
			want:          "!x", // Should not be wrapped
		},
		{
			name: "Unary ! expression",
			expr: &ast.UnaryExpression{
				Operator: "!",
				Argument: &ast.Identifier{Name: "x"},
			},
			generatedCode: "!x",
			want:          "!x", // Should not be wrapped
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := converter.ConvertBoolSeriesForIfStatement(tt.expr, tt.generatedCode)
			if got != tt.want {
				t.Errorf("ConvertBoolSeriesForIfStatement() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBooleanConverter_EnsureBooleanOperand_EdgeCases(t *testing.T) {
	typeSystem := NewTypeInferenceEngine()
	converter := NewBooleanConverter(typeSystem)

	tests := []struct {
		name          string
		expr          ast.Expression
		generatedCode string
		want          string
	}{
		{
			name:          "Float variable",
			expr:          &ast.Identifier{Name: "x"},
			generatedCode: "xSeries.GetCurrent()",
			want:          "(value.IsTrue(xSeries.GetCurrent()))",
		},
		{
			name:          "Comparison",
			expr:          &ast.BinaryExpression{Operator: ">"},
			generatedCode: "a > b",
			want:          "a > b",
		},
		{
			name:          "NaN literal",
			expr:          &ast.Literal{Value: "NaN"},
			generatedCode: "math.NaN()",
			want:          "math.NaN()", // Should be skipped
		},
		{
			name:          "Zero literal",
			expr:          &ast.Literal{Value: 0.0},
			generatedCode: "0.0",
			want:          "0.0", // Should be skipped
		},
		{
			name:          "Function call (na)",
			expr:          &ast.CallExpression{Callee: &ast.Identifier{Name: "na"}},
			generatedCode: "math.IsNaN(x)",
			want:          "math.IsNaN(x)",
		},
		{
			name:          "Function call (unknown)",
			expr:          &ast.CallExpression{Callee: &ast.Identifier{Name: "foo"}},
			generatedCode: "foo(x)",
			want:          "(value.IsTrue(foo(x)))",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := converter.EnsureBooleanOperand(tt.expr, tt.generatedCode)
			if got != tt.want {
				t.Errorf("EnsureBooleanOperand() = %v, want %v", got, tt.want)
			}
		})
	}
}
