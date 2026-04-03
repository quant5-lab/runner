package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestBooleanConverter_ConditionalExpression(t *testing.T) {
	tests := []struct {
		name           string
		expr           ast.Expression
		generatedCode  string
		expectedResult string
		description    string
	}{
		{
			name: "float_returning_ternary_in_if_context",
			expr: &ast.ConditionalExpression{
				Test:       &ast.Identifier{Name: "x"},
				Consequent: &ast.Literal{Value: 1.0},
				Alternate:  &ast.Literal{Value: 0.0},
			},
			generatedCode:  "func() float64 { if x { return 1.0 } else { return 0.0 } }()",
			expectedResult: "value.IsTrue(func() float64 { if x { return 1.0 } else { return 0.0 } }())",
			description:    "Numeric ternary needs IsTrue wrapper in boolean context",
		},
		{
			name: "boolean_comparison_ternary_in_if_context",
			expr: &ast.ConditionalExpression{
				Test: &ast.Identifier{Name: "condition"},
				Consequent: &ast.BinaryExpression{
					Left:     &ast.Identifier{Name: "a"},
					Operator: ">",
					Right:    &ast.Identifier{Name: "b"},
				},
				Alternate: &ast.BinaryExpression{
					Left:     &ast.Identifier{Name: "c"},
					Operator: "<",
					Right:    &ast.Identifier{Name: "d"},
				},
			},
			generatedCode:  "func() float64 { if condition { return func() float64 { if (a > b) { return 1.0 } else { return 0.0 } }() } else { return func() float64 { if (c < d) { return 1.0 } else { return 0.0 } }() } }()",
			expectedResult: "value.IsTrue(func() float64 { if condition { return func() float64 { if (a > b) { return 1.0 } else { return 0.0 } }() } else { return func() float64 { if (c < d) { return 1.0 } else { return 0.0 } }() } }())",
			description:    "Nested comparison ternary still needs wrapper",
		},
		{
			name: "ternary_iife_pattern",
			expr: &ast.ConditionalExpression{
				Test:       &ast.Identifier{Name: "x"},
				Consequent: &ast.CallExpression{Callee: &ast.Identifier{Name: "f"}},
				Alternate:  &ast.CallExpression{Callee: &ast.Identifier{Name: "g"}},
			},
			generatedCode:  "func() float64 { if x { return f() } else { return g() } }()",
			expectedResult: "value.IsTrue(func() float64 { if x { return f() } else { return g() } }())",
			description:    "Function call ternary wrapped for safety",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			typeSystem := NewTypeInferenceEngine()
			converter := NewBooleanConverter(typeSystem)

			result := converter.ConvertBoolSeriesForIfStatement(tt.expr, tt.generatedCode)

			if result != tt.expectedResult {
				t.Errorf("%s\nGot:  %s\nWant: %s", tt.description, result, tt.expectedResult)
			}
		})
	}
}

func TestBooleanConverter_ConditionalPreservesExistingBooleans(t *testing.T) {
	tests := []struct {
		name          string
		expr          ast.Expression
		generatedCode string
		shouldWrap    bool
	}{
		{
			name: "logical_expression_not_wrapped",
			expr: &ast.LogicalExpression{
				Left:     &ast.Identifier{Name: "a"},
				Operator: "&&",
				Right:    &ast.Identifier{Name: "b"},
			},
			generatedCode: "(a && b)",
			shouldWrap:    false,
		},
		{
			name: "unary_not_not_wrapped",
			expr: &ast.UnaryExpression{
				Operator: "not",
				Argument: &ast.Identifier{Name: "x"},
			},
			generatedCode: "!x",
			shouldWrap:    false,
		},
		{
			name: "conditional_expression_wrapped",
			expr: &ast.ConditionalExpression{
				Test:       &ast.Identifier{Name: "x"},
				Consequent: &ast.Literal{Value: 1.0},
				Alternate:  &ast.Literal{Value: 0.0},
			},
			generatedCode: "func() float64 { if x { return 1.0 } else { return 0.0 } }()",
			shouldWrap:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			typeSystem := NewTypeInferenceEngine()
			converter := NewBooleanConverter(typeSystem)

			result := converter.ConvertBoolSeriesForIfStatement(tt.expr, tt.generatedCode)

			hasWrapper := strings.Contains(result, "value.IsTrue(")

			if hasWrapper != tt.shouldWrap {
				t.Errorf("shouldWrap=%v but hasWrapper=%v for: %s", tt.shouldWrap, hasWrapper, result)
			}
		})
	}
}

func TestGenerateVariableInit_ConditionalBooleanIntegration(t *testing.T) {
	tests := []struct {
		name                 string
		varName              string
		initExpr             ast.Expression
		verifyIsTrueWrapping bool
	}{
		{
			name:    "nested_ternary_boolean_test",
			varName: "signal",
			initExpr: &ast.ConditionalExpression{
				Test: &ast.ConditionalExpression{
					Test:       &ast.Identifier{Name: "x"},
					Consequent: &ast.Literal{Value: 1.0},
					Alternate:  &ast.Literal{Value: 0.0},
				},
				Consequent: &ast.Literal{Value: 10.0},
				Alternate:  &ast.Literal{Value: 20.0},
			},
			verifyIsTrueWrapping: true,
		},
		{
			name:    "ternary_with_comparison_branches",
			varName: "result",
			initExpr: &ast.ConditionalExpression{
				Test: &ast.ConditionalExpression{
					Test: &ast.Identifier{Name: "mode"},
					Consequent: &ast.BinaryExpression{
						Left:     &ast.Identifier{Name: "a"},
						Operator: ">",
						Right:    &ast.Identifier{Name: "b"},
					},
					Alternate: &ast.BinaryExpression{
						Left:     &ast.Identifier{Name: "c"},
						Operator: "<",
						Right:    &ast.Identifier{Name: "d"},
					},
				},
				Consequent: &ast.Literal{Value: 1.0},
				Alternate:  &ast.Literal{Value: 0.0},
			},
			verifyIsTrueWrapping: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := createTestGenerator()
			gen.variables[tt.varName] = "float64"

			code, err := gen.generateVariableInit(tt.varName, tt.initExpr)
			if err != nil {
				t.Fatalf("generateVariableInit error: %v", err)
			}

			hasIsTrue := strings.Contains(code, "value.IsTrue(")

			if tt.verifyIsTrueWrapping && !hasIsTrue {
				t.Error("Expected value.IsTrue() wrapping for nested conditional test")
			}
		})
	}
}
