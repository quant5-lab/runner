package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/runtime/validation"
)

func TestGenerateStringExpression_ColorConstants(t *testing.T) {
	tests := []struct {
		name     string
		expr     ast.Expression
		expected string
	}{
		{
			name: "color.red constant",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "color"},
				Property: &ast.Identifier{Name: "red"},
			},
			expected: `"#FF0000"`,
		},
		{
			name: "color.lime constant",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "color"},
				Property: &ast.Identifier{Name: "lime"},
			},
			expected: `"#00FF00"`,
		},
		{
			name: "color.blue constant",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "color"},
				Property: &ast.Identifier{Name: "blue"},
			},
			expected: `"#0000FF"`,
		},
		{
			name: "color.maroon constant",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "color"},
				Property: &ast.Identifier{Name: "maroon"},
			},
			expected: `"#800000"`,
		},
		{
			name: "color.fuchsia constant",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "color"},
				Property: &ast.Identifier{Name: "fuchsia"},
			},
			expected: `"#FF00FF"`,
		},
		{
			name: "color.aqua constant",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "color"},
				Property: &ast.Identifier{Name: "aqua"},
			},
			expected: `"#00FFFF"`,
		},
		{
			name: "color.navy constant",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "color"},
				Property: &ast.Identifier{Name: "navy"},
			},
			expected: `"#000080"`,
		},
		{
			name: "color.olive constant",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "color"},
				Property: &ast.Identifier{Name: "olive"},
			},
			expected: `"#808000"`,
		},
		{
			name: "color.silver constant",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "color"},
				Property: &ast.Identifier{Name: "silver"},
			},
			expected: `"#C0C0C0"`,
		},
		{
			name: "color.white constant",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "color"},
				Property: &ast.Identifier{Name: "white"},
			},
			expected: `"#FFFFFF"`,
		},
		{
			name: "strategy.long constant",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "strategy"},
				Property: &ast.Identifier{Name: "long"},
			},
			expected: "strategy.Long",
		},
		{
			name: "strategy.short constant",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "strategy"},
				Property: &ast.Identifier{Name: "short"},
			},
			expected: "strategy.Short",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := createStringExpressionTestGenerator()
			result, err := gen.generateStringExpression(tt.expr)

			if err != nil {
				t.Fatalf("generateStringExpression() error = %v", err)
			}

			if result != tt.expected {
				t.Errorf("generateStringExpression() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestGenerateStringExpression_SimpleConditional(t *testing.T) {
	tests := []struct {
		name            string
		expr            *ast.ConditionalExpression
		expectContains  []string
		expectNotExists []string
	}{
		{
			name: "simple conditional with color branches",
			expr: &ast.ConditionalExpression{
				Test: &ast.BinaryExpression{
					Operator: ">",
					Left:     &ast.Identifier{Name: "close"},
					Right:    &ast.Identifier{Name: "open"},
				},
				Consequent: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "color"},
					Property: &ast.Identifier{Name: "lime"},
				},
				Alternate: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "color"},
					Property: &ast.Identifier{Name: "red"},
				},
			},
			expectContains: []string{
				"func() string {",
				`"#00FF00"`,
				`"#FF0000"`,
			},
		},
		{
			name: "conditional with strategy constants",
			expr: &ast.ConditionalExpression{
				Test: &ast.Identifier{Name: "enabled"},
				Consequent: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "strategy"},
					Property: &ast.Identifier{Name: "long"},
				},
				Alternate: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "strategy"},
					Property: &ast.Identifier{Name: "short"},
				},
			},
			expectContains: []string{
				"func() string {",
				"strategy.Long",
				"strategy.Short",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := createStringExpressionTestGenerator()
			result, err := gen.generateStringExpression(tt.expr)

			if err != nil {
				t.Fatalf("generateStringExpression() error = %v", err)
			}

			for _, expected := range tt.expectContains {
				if !strings.Contains(result, expected) {
					t.Errorf("generateStringExpression() result missing %q\nGot: %s", expected, result)
				}
			}

			for _, notExpected := range tt.expectNotExists {
				if strings.Contains(result, notExpected) {
					t.Errorf("generateStringExpression() result should not contain %q\nGot: %s", notExpected, result)
				}
			}
		})
	}
}

func TestGenerateStringExpression_NestedConditional(t *testing.T) {
	tests := []struct {
		name           string
		expr           *ast.ConditionalExpression
		expectContains []string
	}{
		{
			name: "nested conditional with color branches",
			expr: &ast.ConditionalExpression{
				Test: &ast.Identifier{Name: "cond1"},
				Consequent: &ast.ConditionalExpression{
					Test: &ast.Identifier{Name: "cond2"},
					Consequent: &ast.MemberExpression{
						Object:   &ast.Identifier{Name: "color"},
						Property: &ast.Identifier{Name: "lime"},
					},
					Alternate: &ast.MemberExpression{
						Object:   &ast.Identifier{Name: "color"},
						Property: &ast.Identifier{Name: "maroon"},
					},
				},
				Alternate: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "color"},
					Property: &ast.Identifier{Name: "red"},
				},
			},
			expectContains: []string{
				"func() string {",
				`"#00FF00"`,
				`"#800000"`,
				`"#FF0000"`,
			},
		},
		{
			name: "deeply nested conditional with multiple colors",
			expr: &ast.ConditionalExpression{
				Test: &ast.Identifier{Name: "cond1"},
				Consequent: &ast.ConditionalExpression{
					Test: &ast.Identifier{Name: "cond2"},
					Consequent: &ast.ConditionalExpression{
						Test: &ast.Identifier{Name: "cond3"},
						Consequent: &ast.MemberExpression{
							Object:   &ast.Identifier{Name: "color"},
							Property: &ast.Identifier{Name: "lime"},
						},
						Alternate: &ast.MemberExpression{
							Object:   &ast.Identifier{Name: "color"},
							Property: &ast.Identifier{Name: "blue"},
						},
					},
					Alternate: &ast.MemberExpression{
						Object:   &ast.Identifier{Name: "color"},
						Property: &ast.Identifier{Name: "maroon"},
					},
				},
				Alternate: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "color"},
					Property: &ast.Identifier{Name: "red"},
				},
			},
			expectContains: []string{
				"func() string {",
				`"#00FF00"`,
				`"#0000FF"`,
				`"#800000"`,
				`"#FF0000"`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := createStringExpressionTestGenerator()
			result, err := gen.generateStringExpression(tt.expr)

			if err != nil {
				t.Fatalf("generateStringExpression() error = %v", err)
			}

			for _, expected := range tt.expectContains {
				if !strings.Contains(result, expected) {
					t.Errorf("generateStringExpression() result missing %q\nGot: %s", expected, result)
				}
			}
		})
	}
}

func TestGenerateStringExpression_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		expr        ast.Expression
		expectError bool
	}{
		{
			name:        "unsupported identifier",
			expr:        &ast.Identifier{Name: "unknownVar"},
			expectError: true,
		},
		{
			name:        "unsupported literal",
			expr:        &ast.Literal{Value: 123.45},
			expectError: true,
		},
		{
			name: "unsupported call expression",
			expr: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ta"},
					Property: &ast.Identifier{Name: "sma"},
				},
			},
			expectError: true,
		},
		{
			name: "unsupported ta member expression",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "ta"},
				Property: &ast.Identifier{Name: "sma"},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := createStringExpressionTestGenerator()
			_, err := gen.generateStringExpression(tt.expr)

			if tt.expectError && err == nil {
				t.Error("generateStringExpression() expected error, got nil")
			}

			if !tt.expectError && err != nil {
				t.Errorf("generateStringExpression() unexpected error = %v", err)
			}
		})
	}
}

func createStringExpressionTestGenerator() *generator {
	typeSystem := NewTypeInferenceEngine()
	return &generator{
		imports:        make(map[string]bool),
		variables:      make(map[string]string),
		strategyConfig: NewStrategyConfig(),
		taRegistry:     NewTAFunctionRegistry(),
		constEvaluator: validation.NewWarmupAnalyzer(),
		boolConverter:  NewBooleanConverter(typeSystem),
		typeSystem:     typeSystem,
	}
}
