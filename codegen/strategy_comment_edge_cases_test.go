package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

/* TestStrategyCommentBinaryExpression verifies BinaryExpression handling (unsupported) */
func TestStrategyCommentBinaryExpression(t *testing.T) {
	/* Simulate: strategy.entry("Trade", dir, comment="Price: " + str.tostring(close)) */
	program := &ast.Program{
		NodeType: ast.TypeProgram,
		Body: []ast.Node{
			&ast.ExpressionStatement{
				Expression: &ast.CallExpression{
					Callee: &ast.MemberExpression{
						Object:   &ast.Identifier{Name: "strategy"},
						Property: &ast.Identifier{Name: "entry"},
					},
					Arguments: []ast.Expression{
						&ast.Literal{Value: "Trade"},
						&ast.Identifier{Name: "dir"},
						&ast.ObjectExpression{
							NodeType: ast.TypeObjectExpression,
							Properties: []ast.Property{
								{
									NodeType: ast.TypeProperty,
									Key:      &ast.Identifier{Name: "comment"},
									Value: &ast.BinaryExpression{
										Operator: "+",
										Left:     &ast.Literal{Value: "Price: "},
										Right:    &ast.Identifier{Name: "close"},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	code, err := GenerateStrategyCodeFromAST(program)
	if err != nil {
		t.Fatalf("GenerateStrategyCodeFromAST failed: %v", err)
	}

	/* Verify unsupported expression defaults to empty string */
	if !strings.Contains(code.FunctionBody, `strat.Entry("Trade", strategy.Long, 1, "")`) {
		t.Errorf("Expected empty string for unsupported BinaryExpression, got:\n%s", code.FunctionBody)
	}
}

/* TestStrategyCommentUnaryExpression verifies UnaryExpression handling (unsupported) */
func TestStrategyCommentUnaryExpression(t *testing.T) {
	/* Simulate: strategy.entry("Trade", dir, comment=-1) - numeric UnaryExpression */
	program := &ast.Program{
		NodeType: ast.TypeProgram,
		Body: []ast.Node{
			&ast.ExpressionStatement{
				Expression: &ast.CallExpression{
					Callee: &ast.MemberExpression{
						Object:   &ast.Identifier{Name: "strategy"},
						Property: &ast.Identifier{Name: "entry"},
					},
					Arguments: []ast.Expression{
						&ast.Literal{Value: "Trade"},
						&ast.Identifier{Name: "dir"},
						&ast.ObjectExpression{
							NodeType: ast.TypeObjectExpression,
							Properties: []ast.Property{
								{
									NodeType: ast.TypeProperty,
									Key:      &ast.Identifier{Name: "comment"},
									Value: &ast.UnaryExpression{
										Operator: "-",
										Argument: &ast.Literal{Value: float64(1)},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	code, err := GenerateStrategyCodeFromAST(program)
	if err != nil {
		t.Fatalf("GenerateStrategyCodeFromAST failed: %v", err)
	}

	/* Verify unsupported expression defaults to empty string */
	if !strings.Contains(code.FunctionBody, `strat.Entry("Trade", strategy.Long, 1, "")`) {
		t.Errorf("Expected empty string for unsupported UnaryExpression, got:\n%s", code.FunctionBody)
	}
}

/* TestStrategyCommentNumericLiteral verifies numeric literal handling (unsupported) */
func TestStrategyCommentNumericLiteral(t *testing.T) {
	/* Simulate: strategy.entry("Trade", dir, comment=42) - numeric literal instead of string */
	program := &ast.Program{
		NodeType: ast.TypeProgram,
		Body: []ast.Node{
			&ast.ExpressionStatement{
				Expression: &ast.CallExpression{
					Callee: &ast.MemberExpression{
						Object:   &ast.Identifier{Name: "strategy"},
						Property: &ast.Identifier{Name: "entry"},
					},
					Arguments: []ast.Expression{
						&ast.Literal{Value: "Trade"},
						&ast.Identifier{Name: "dir"},
						&ast.ObjectExpression{
							NodeType: ast.TypeObjectExpression,
							Properties: []ast.Property{
								{
									NodeType: ast.TypeProperty,
									Key:      &ast.Identifier{Name: "comment"},
									Value:    &ast.Literal{Value: float64(42)},
								},
							},
						},
					},
				},
			},
		},
	}

	code, err := GenerateStrategyCodeFromAST(program)
	if err != nil {
		t.Fatalf("GenerateStrategyCodeFromAST failed: %v", err)
	}

	/* Verify numeric literal defaults to empty string */
	if !strings.Contains(code.FunctionBody, `strat.Entry("Trade", strategy.Long, 1, "")`) {
		t.Errorf("Expected empty string for numeric literal comment, got:\n%s", code.FunctionBody)
	}
}

/* TestStrategyCommentCallExpression verifies CallExpression handling (unsupported) */
func TestStrategyCommentCallExpression(t *testing.T) {
	/* Simulate: strategy.entry("Trade", dir, comment=str.tostring(close)) */
	program := &ast.Program{
		NodeType: ast.TypeProgram,
		Body: []ast.Node{
			&ast.ExpressionStatement{
				Expression: &ast.CallExpression{
					Callee: &ast.MemberExpression{
						Object:   &ast.Identifier{Name: "strategy"},
						Property: &ast.Identifier{Name: "entry"},
					},
					Arguments: []ast.Expression{
						&ast.Literal{Value: "Trade"},
						&ast.Identifier{Name: "dir"},
						&ast.ObjectExpression{
							NodeType: ast.TypeObjectExpression,
							Properties: []ast.Property{
								{
									NodeType: ast.TypeProperty,
									Key:      &ast.Identifier{Name: "comment"},
									Value: &ast.CallExpression{
										Callee: &ast.MemberExpression{
											Object:   &ast.Identifier{Name: "str"},
											Property: &ast.Identifier{Name: "tostring"},
										},
										Arguments: []ast.Expression{
											&ast.Identifier{Name: "close"},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	code, err := GenerateStrategyCodeFromAST(program)
	if err != nil {
		t.Fatalf("GenerateStrategyCodeFromAST failed: %v", err)
	}

	/* Verify unsupported CallExpression defaults to empty string */
	if !strings.Contains(code.FunctionBody, `strat.Entry("Trade", strategy.Long, 1, "")`) {
		t.Errorf("Expected empty string for unsupported CallExpression, got:\n%s", code.FunctionBody)
	}
}

/* TestStrategyCommentSpecialCharacters verifies special character escaping */
func TestStrategyCommentSpecialCharacters(t *testing.T) {
	/* Simulate: strategy.entry("Trade", dir, comment="Quote: \"buy\"") */
	program := &ast.Program{
		NodeType: ast.TypeProgram,
		Body: []ast.Node{
			&ast.ExpressionStatement{
				Expression: &ast.CallExpression{
					Callee: &ast.MemberExpression{
						Object:   &ast.Identifier{Name: "strategy"},
						Property: &ast.Identifier{Name: "entry"},
					},
					Arguments: []ast.Expression{
						&ast.Literal{Value: "Trade"},
						&ast.Identifier{Name: "dir"},
						&ast.ObjectExpression{
							NodeType: ast.TypeObjectExpression,
							Properties: []ast.Property{
								{
									NodeType: ast.TypeProperty,
									Key:      &ast.Identifier{Name: "comment"},
									Value:    &ast.Literal{Value: `Quote: "buy"`},
								},
							},
						},
					},
				},
			},
		},
	}

	code, err := GenerateStrategyCodeFromAST(program)
	if err != nil {
		t.Fatalf("GenerateStrategyCodeFromAST failed: %v", err)
	}

	/* Verify proper escaping of quotes */
	if !strings.Contains(code.FunctionBody, `\"`) {
		t.Errorf("Expected escaped quotes in comment string, got:\n%s", code.FunctionBody)
	}
}

/* TestStrategyCommentNewlineCharacters verifies newline handling */
func TestStrategyCommentNewlineCharacters(t *testing.T) {
	/* Simulate: strategy.entry("Trade", dir, comment="Line1\nLine2") */
	program := &ast.Program{
		NodeType: ast.TypeProgram,
		Body: []ast.Node{
			&ast.ExpressionStatement{
				Expression: &ast.CallExpression{
					Callee: &ast.MemberExpression{
						Object:   &ast.Identifier{Name: "strategy"},
						Property: &ast.Identifier{Name: "entry"},
					},
					Arguments: []ast.Expression{
						&ast.Literal{Value: "Trade"},
						&ast.Identifier{Name: "dir"},
						&ast.ObjectExpression{
							NodeType: ast.TypeObjectExpression,
							Properties: []ast.Property{
								{
									NodeType: ast.TypeProperty,
									Key:      &ast.Identifier{Name: "comment"},
									Value:    &ast.Literal{Value: "Line1\nLine2"},
								},
							},
						},
					},
				},
			},
		},
	}

	code, err := GenerateStrategyCodeFromAST(program)
	if err != nil {
		t.Fatalf("GenerateStrategyCodeFromAST failed: %v", err)
	}

	/* Verify newline preserved or escaped */
	if !strings.Contains(code.FunctionBody, "Line1") && !strings.Contains(code.FunctionBody, "Line2") {
		t.Errorf("Expected comment with newline content, got:\n%s", code.FunctionBody)
	}
}
