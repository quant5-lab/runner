package codegen

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestConditionalArgumentAnalyzer_FindInExpression(t *testing.T) {
	tests := []struct {
		name          string
		expression    ast.Expression
		expectedCount int
		description   string
	}{
		{
			name: "simple_conditional_in_call",
			expression: &ast.CallExpression{
				NodeType: ast.TypeCallExpression,
				Callee:   &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "ta.sma"},
				Arguments: []ast.Expression{
					&ast.ConditionalExpression{
						NodeType: ast.TypeConditionalExpression,
						Test: &ast.BinaryExpression{
							NodeType: ast.TypeBinaryExpression,
							Left:     &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "close"},
							Operator: ">",
							Right:    &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "open"},
						},
						Consequent: &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "high"},
						Alternate:  &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "low"},
					},
					&ast.Literal{NodeType: ast.TypeLiteral, Value: 14.0, Raw: "14"},
				},
			},
			expectedCount: 1,
			description:   "ta.sma(close > open ? high : low, 14)",
		},
		{
			name: "multiple_conditionals_in_call",
			expression: &ast.CallExpression{
				NodeType: ast.TypeCallExpression,
				Callee:   &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "myFunc"},
				Arguments: []ast.Expression{
					&ast.ConditionalExpression{
						NodeType:   ast.TypeConditionalExpression,
						Test:       &ast.Literal{NodeType: ast.TypeLiteral, Value: true, Raw: "true"},
						Consequent: &ast.Literal{NodeType: ast.TypeLiteral, Value: 1.0, Raw: "1"},
						Alternate:  &ast.Literal{NodeType: ast.TypeLiteral, Value: 0.0, Raw: "0"},
					},
					&ast.ConditionalExpression{
						NodeType:   ast.TypeConditionalExpression,
						Test:       &ast.Literal{NodeType: ast.TypeLiteral, Value: false, Raw: "false"},
						Consequent: &ast.Literal{NodeType: ast.TypeLiteral, Value: 2.0, Raw: "2"},
						Alternate:  &ast.Literal{NodeType: ast.TypeLiteral, Value: 3.0, Raw: "3"},
					},
				},
			},
			expectedCount: 2,
			description:   "myFunc(true ? 1 : 0, false ? 2 : 3)",
		},
		{
			name: "nested_call_with_conditional",
			expression: &ast.CallExpression{
				NodeType: ast.TypeCallExpression,
				Callee:   &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "ta.ema"},
				Arguments: []ast.Expression{
					&ast.CallExpression{
						NodeType: ast.TypeCallExpression,
						Callee:   &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "ta.sma"},
						Arguments: []ast.Expression{
							&ast.ConditionalExpression{
								NodeType:   ast.TypeConditionalExpression,
								Test:       &ast.Literal{NodeType: ast.TypeLiteral, Value: true, Raw: "true"},
								Consequent: &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "close"},
								Alternate:  &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "open"},
							},
							&ast.Literal{NodeType: ast.TypeLiteral, Value: 10.0, Raw: "10"},
						},
					},
					&ast.Literal{NodeType: ast.TypeLiteral, Value: 14.0, Raw: "14"},
				},
			},
			expectedCount: 1,
			description:   "ta.ema(ta.sma(true ? close : open, 10), 14)",
		},
		{
			name: "no_conditionals",
			expression: &ast.CallExpression{
				NodeType: ast.TypeCallExpression,
				Callee:   &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "ta.sma"},
				Arguments: []ast.Expression{
					&ast.Identifier{NodeType: ast.TypeIdentifier, Name: "close"},
					&ast.Literal{NodeType: ast.TypeLiteral, Value: 14.0, Raw: "14"},
				},
			},
			expectedCount: 0,
			description:   "ta.sma(close, 14) - no conditionals",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasher := &ExpressionHasher{}
			analyzer := NewConditionalArgumentAnalyzer(hasher)

			results := analyzer.FindInExpression(tt.expression)

			if len(results) != tt.expectedCount {
				t.Errorf("%s: expected %d conditionals, got %d",
					tt.description, tt.expectedCount, len(results))
			}

			for i, result := range results {
				if result.Conditional == nil {
					t.Errorf("Result %d: Conditional is nil", i)
				}
				if result.ParentCall == nil {
					t.Errorf("Result %d: ParentCall is nil", i)
				}
				if result.ContentHash == "" {
					t.Errorf("Result %d: ContentHash is empty", i)
				}
				if result.ArgIndex < 0 {
					t.Errorf("Result %d: ArgIndex is negative: %d", i, result.ArgIndex)
				}
			}
		})
	}
}

func TestConditionalArgumentAnalyzer_HashDeduplication(t *testing.T) {
	hasher := &ExpressionHasher{}
	analyzer := NewConditionalArgumentAnalyzer(hasher)

	cond1 := &ast.ConditionalExpression{
		NodeType: ast.TypeConditionalExpression,
		Test: &ast.BinaryExpression{
			NodeType: ast.TypeBinaryExpression,
			Left:     &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "close"},
			Operator: ">",
			Right:    &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "open"},
		},
		Consequent: &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "high"},
		Alternate:  &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "low"},
	}

	cond2 := &ast.ConditionalExpression{
		NodeType: ast.TypeConditionalExpression,
		Test: &ast.BinaryExpression{
			NodeType: ast.TypeBinaryExpression,
			Left:     &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "close"},
			Operator: ">",
			Right:    &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "open"},
		},
		Consequent: &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "high"},
		Alternate:  &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "low"},
	}

	hash1 := analyzer.computeHash(cond1)
	hash2 := analyzer.computeHash(cond2)

	if hash1 != hash2 {
		t.Errorf("Identical conditionals produced different hashes: %s vs %s", hash1, hash2)
	}

	cond3 := &ast.ConditionalExpression{
		NodeType: ast.TypeConditionalExpression,
		Test: &ast.BinaryExpression{
			NodeType: ast.TypeBinaryExpression,
			Left:     &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "close"},
			Operator: "<",
			Right:    &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "open"},
		},
		Consequent: &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "high"},
		Alternate:  &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "low"},
	}

	hash3 := analyzer.computeHash(cond3)

	if hash1 == hash3 {
		t.Errorf("Different conditionals produced same hash: %s", hash1)
	}
}
