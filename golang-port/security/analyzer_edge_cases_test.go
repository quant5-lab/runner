package security

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestAnalyzeAST_EdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		program   *ast.Program
		expected  int
		wantPanic bool
	}{
		{
			name:      "nil_program",
			program:   nil,
			expected:  0,
			wantPanic: true,
		},
		{
			name:     "empty_program",
			program:  &ast.Program{Body: []ast.Node{}},
			expected: 0,
		},
		{
			name:     "nil_body",
			program:  &ast.Program{Body: nil},
			expected: 0,
		},
		{
			name: "non_security_calls_only",
			program: &ast.Program{
				Body: []ast.Node{
					&ast.VariableDeclaration{
						Declarations: []ast.VariableDeclarator{
							{
								ID: ast.Identifier{Name: "x"},
								Init: &ast.CallExpression{
									Callee: &ast.MemberExpression{
										Object:   &ast.Identifier{Name: "ta"},
										Property: &ast.Identifier{Name: "sma"},
									},
									Arguments: []ast.Expression{
										&ast.Identifier{Name: "close"},
										&ast.Literal{Value: 20},
									},
								},
							},
						},
					},
				},
			},
			expected: 0,
		},
		{
			name: "nested_non_security",
			program: &ast.Program{
				Body: []ast.Node{
					&ast.VariableDeclaration{
						Declarations: []ast.VariableDeclarator{
							{
								ID: ast.Identifier{Name: "x"},
								Init: &ast.BinaryExpression{
									Operator: "+",
									Left: &ast.CallExpression{
										Callee: &ast.MemberExpression{
											Object:   &ast.Identifier{Name: "math"},
											Property: &ast.Identifier{Name: "max"},
										},
									},
									Right: &ast.Literal{Value: 10.0},
								},
							},
						},
					},
				},
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result []SecurityCall
			var panicked bool

			// Catch panics
			defer func() {
				if r := recover(); r != nil {
					panicked = true
					if !tt.wantPanic {
						t.Errorf("unexpected panic: %v", r)
					}
				}
			}()

			result = AnalyzeAST(tt.program)

			if tt.wantPanic && !panicked {
				t.Error("expected panic but got none")
			}

			if !panicked && len(result) != tt.expected {
				t.Errorf("expected %d calls, got %d", tt.expected, len(result))
			}
		})
	}
}

func TestExtractMaxPeriod_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		expr     ast.Expression
		expected int
	}{
		{
			name:     "nil_expression",
			expr:     nil,
			expected: 0,
		},
		{
			name:     "literal_no_period",
			expr:     &ast.Literal{Value: 42.0},
			expected: 0,
		},
		{
			name:     "identifier_no_period",
			expr:     &ast.Identifier{Name: "close"},
			expected: 0,
		},
		{
			name: "ta_call_missing_args",
			expr: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ta"},
					Property: &ast.Identifier{Name: "sma"},
				},
				Arguments: []ast.Expression{},
			},
			expected: 0,
		},
		{
			name: "ta_call_one_arg",
			expr: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ta"},
					Property: &ast.Identifier{Name: "sma"},
				},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
				},
			},
			expected: 0,
		},
		{
			name: "ta_call_non_numeric_period",
			expr: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ta"},
					Property: &ast.Identifier{Name: "sma"},
				},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
					&ast.Identifier{Name: "period_var"},
				},
			},
			expected: 0,
		},
		{
			name: "ta_call_zero_period",
			expr: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ta"},
					Property: &ast.Identifier{Name: "sma"},
				},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
					&ast.Literal{Value: 0},
				},
			},
			expected: 0,
		},
		{
			name: "ta_call_negative_period",
			expr: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ta"},
					Property: &ast.Identifier{Name: "sma"},
				},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
					&ast.Literal{Value: -10},
				},
			},
			expected: 0,
		},
		{
			name: "ta_call_fractional_period",
			expr: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ta"},
					Property: &ast.Identifier{Name: "sma"},
				},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
					&ast.Literal{Value: 20.5},
				},
			},
			expected: 20,
		},
		{
			name: "non_ta_call",
			expr: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "math"},
					Property: &ast.Identifier{Name: "max"},
				},
				Arguments: []ast.Expression{
					&ast.Literal{Value: 10.0},
					&ast.Literal{Value: 20.0},
				},
			},
			expected: 0,
		},
		{
			name: "valid_ta_call_with_period",
			expr: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ta"},
					Property: &ast.Identifier{Name: "sma"},
				},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
					&ast.Literal{Value: 50.0},
				},
			},
			expected: 50,
		},
		{
			name: "binary_without_ta_calls",
			expr: &ast.BinaryExpression{
				Operator: "+",
				Left:     &ast.Literal{Value: 10.0},
				Right:    &ast.Literal{Value: 20.0},
			},
			expected: 0,
		},
		{
			name: "conditional_without_ta_calls",
			expr: &ast.ConditionalExpression{
				Test:       &ast.BinaryExpression{Operator: ">"},
				Consequent: &ast.Literal{Value: 1.0},
				Alternate:  &ast.Literal{Value: 0.0},
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("unexpected panic: %v", r)
				}
			}()

			result := ExtractMaxPeriod(tt.expr)
			if result != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestContainsFunction_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		pattern  string
		expected bool
	}{
		{
			name:     "empty_text",
			text:     "",
			pattern:  "sma",
			expected: false,
		},
		{
			name:     "empty_pattern",
			text:     "ta.sma(close, 20)",
			pattern:  "",
			expected: true,
		},
		{
			name:     "both_empty",
			text:     "",
			pattern:  "",
			expected: true,
		},
		{
			name:     "pattern_longer_than_text",
			text:     "sma",
			pattern:  "sma_very_long",
			expected: false,
		},
		{
			name:     "exact_match",
			text:     "sma",
			pattern:  "sma",
			expected: true,
		},
		{
			name:     "substring_match",
			text:     "ta.sma(close, 20)",
			pattern:  "sma",
			expected: true,
		},
		{
			name:     "case_sensitive",
			text:     "ta.SMA(close, 20)",
			pattern:  "sma",
			expected: false,
		},
		{
			name:     "multiple_occurrences",
			text:     "sma + sma + sma",
			pattern:  "sma",
			expected: true,
		},
		{
			name:     "special_characters",
			text:     "ta.sma(close[1], 20)",
			pattern:  "sma",
			expected: true,
		},
		{
			name:     "unicode_text",
			text:     "币安.sma(close, 20)",
			pattern:  "sma",
			expected: true,
		},
		{
			name:     "unicode_pattern",
			text:     "function_币安(x)",
			pattern:  "币安",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := contains(tt.text, tt.pattern)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}
