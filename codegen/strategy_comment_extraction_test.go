package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

/* TestStrategyEntryCommentExtraction verifies comment parameter extraction for strategy.entry() */
func TestStrategyEntryCommentExtraction(t *testing.T) {
	/* Simulate: strategy.entry("Long", strategy.long, 1, comment="Buy signal") */
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
						&ast.Literal{Value: "Long"},
						&ast.MemberExpression{
							Object:   &ast.Identifier{Name: "strategy"},
							Property: &ast.Identifier{Name: "long"},
						},
						&ast.Literal{Value: 1.0},
						&ast.ObjectExpression{
							NodeType: ast.TypeObjectExpression,
							Properties: []ast.Property{
								{
									NodeType: ast.TypeProperty,
									Key:      &ast.Identifier{Name: "comment"},
									Value:    &ast.Literal{Value: "Buy signal"},
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

	/* Verify comment parameter passed to Entry */
	if !strings.Contains(code.FunctionBody, `strat.Entry("Long", strategy.Long, 1, "Buy signal")`) {
		t.Errorf("Expected comment parameter in Entry call, got:\n%s", code.FunctionBody)
	}
}

/* TestStrategyEntryCommentWithVariable verifies variable-based comment extraction */
func TestStrategyEntryCommentWithVariable(t *testing.T) {
	/* Simulate: strategy.entry("Long", strategy.long, 1, comment=signal_msg) */
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
						&ast.Literal{Value: "Long"},
						&ast.MemberExpression{
							Object:   &ast.Identifier{Name: "strategy"},
							Property: &ast.Identifier{Name: "long"},
						},
						&ast.Literal{Value: 1.0},
						&ast.ObjectExpression{
							NodeType: ast.TypeObjectExpression,
							Properties: []ast.Property{
								{
									NodeType: ast.TypeProperty,
									Key:      &ast.Identifier{Name: "comment"},
									Value:    &ast.Identifier{Name: "signal_msg"},
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

	/* Verify variable reference in generated code */
	if !strings.Contains(code.FunctionBody, "signal_msgSeries.GetCurrent()") {
		t.Errorf("Expected series access for variable comment, got:\n%s", code.FunctionBody)
	}
}

/* TestStrategyCloseCommentExtraction verifies comment parameter extraction for strategy.close() */
func TestStrategyCloseCommentExtraction(t *testing.T) {
	/* Simulate: strategy.close("Long", comment="Exit signal") */
	program := &ast.Program{
		NodeType: ast.TypeProgram,
		Body: []ast.Node{
			&ast.ExpressionStatement{
				Expression: &ast.CallExpression{
					Callee: &ast.MemberExpression{
						Object:   &ast.Identifier{Name: "strategy"},
						Property: &ast.Identifier{Name: "close"},
					},
					Arguments: []ast.Expression{
						&ast.Literal{Value: "Long"},
						&ast.ObjectExpression{
							NodeType: ast.TypeObjectExpression,
							Properties: []ast.Property{
								{
									NodeType: ast.TypeProperty,
									Key:      &ast.Identifier{Name: "comment"},
									Value:    &ast.Literal{Value: "Exit signal"},
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

	/* Verify comment parameter passed to Close */
	if !strings.Contains(code.FunctionBody, `strat.Close("Long", bar.Close, bar.Time, "Exit signal")`) {
		t.Errorf("Expected comment parameter in Close call, got:\n%s", code.FunctionBody)
	}
}

/* TestStrategyExitCommentExtraction verifies comment parameter extraction for strategy.exit() */
func TestStrategyExitCommentExtraction(t *testing.T) {
	/* Simulate: strategy.exit("Exit", "Long", stop=95, limit=110, comment="Stop/Limit") */
	program := &ast.Program{
		NodeType: ast.TypeProgram,
		Body: []ast.Node{
			&ast.ExpressionStatement{
				Expression: &ast.CallExpression{
					Callee: &ast.MemberExpression{
						Object:   &ast.Identifier{Name: "strategy"},
						Property: &ast.Identifier{Name: "exit"},
					},
					Arguments: []ast.Expression{
						&ast.Literal{Value: "Exit"},
						&ast.Literal{Value: "Long"},
						&ast.ObjectExpression{
							NodeType: ast.TypeObjectExpression,
							Properties: []ast.Property{
								{
									NodeType: ast.TypeProperty,
									Key:      &ast.Identifier{Name: "stop"},
									Value:    &ast.Literal{Value: 95.0},
								},
								{
									NodeType: ast.TypeProperty,
									Key:      &ast.Identifier{Name: "limit"},
									Value:    &ast.Literal{Value: 110.0},
								},
								{
									NodeType: ast.TypeProperty,
									Key:      &ast.Identifier{Name: "comment"},
									Value:    &ast.Literal{Value: "Stop/Limit"},
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

	/* Verify comment parameter passed to ExitWithLevels */
	if !strings.Contains(code.FunctionBody, `"Stop/Limit"`) {
		t.Errorf("Expected comment parameter in ExitWithLevels call, got:\n%s", code.FunctionBody)
	}
}

/* TestStrategyCommentMissing verifies behavior when comment parameter omitted */
func TestStrategyCommentMissing(t *testing.T) {
	/* Simulate: strategy.entry("Long", strategy.long, 1) - no comment */
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
						&ast.Literal{Value: "Long"},
						&ast.MemberExpression{
							Object:   &ast.Identifier{Name: "strategy"},
							Property: &ast.Identifier{Name: "long"},
						},
						&ast.Literal{Value: 1.0},
					},
				},
			},
		},
	}

	code, err := GenerateStrategyCodeFromAST(program)
	if err != nil {
		t.Fatalf("GenerateStrategyCodeFromAST failed: %v", err)
	}

	/* Verify empty string default for missing comment */
	if !strings.Contains(code.FunctionBody, `strat.Entry("Long", strategy.Long, 1, "")`) {
		t.Errorf("Expected empty string default for missing comment, got:\n%s", code.FunctionBody)
	}
}

/* TestStrategyCloseAllComment verifies comment extraction for strategy.close_all() */
func TestStrategyCloseAllComment(t *testing.T) {
	/* Simulate: strategy.close_all(comment="Close all") */
	program := &ast.Program{
		NodeType: ast.TypeProgram,
		Body: []ast.Node{
			&ast.ExpressionStatement{
				Expression: &ast.CallExpression{
					Callee: &ast.MemberExpression{
						Object:   &ast.Identifier{Name: "strategy"},
						Property: &ast.Identifier{Name: "close_all"},
					},
					Arguments: []ast.Expression{
						&ast.ObjectExpression{
							NodeType: ast.TypeObjectExpression,
							Properties: []ast.Property{
								{
									NodeType: ast.TypeProperty,
									Key:      &ast.Identifier{Name: "comment"},
									Value:    &ast.Literal{Value: "Close all"},
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

	/* Verify comment parameter passed to CloseAll */
	if !strings.Contains(code.FunctionBody, `strat.CloseAll(bar.Close, bar.Time, "Close all")`) {
		t.Errorf("Expected comment parameter in CloseAll call, got:\n%s", code.FunctionBody)
	}
}

/* TestStrategyCommentArgumentExtractor verifies ArgumentExtractor pattern for comment */
func TestStrategyCommentArgumentExtractor(t *testing.T) {
	g := &generator{}
	extractor := &ArgumentExtractor{generator: g}

	/* Simulate named args with comment */
	args := []ast.Expression{
		&ast.Literal{Value: "Entry1"},
		&ast.Literal{Value: "Long1"},
		&ast.ObjectExpression{
			NodeType: ast.TypeObjectExpression,
			Properties: []ast.Property{
				{
					NodeType: ast.TypeProperty,
					Key:      &ast.Identifier{Name: "comment"},
					Value:    &ast.Literal{Value: "Test comment"},
				},
			},
		},
	}

	remainingArgs := args[2:]
	commentCode := extractor.ExtractCommentArgument(remainingArgs, "comment", 0, `""`)
	if commentCode != `"Test comment"` {
		t.Errorf("Expected '\"Test comment\"', got %q", commentCode)
	}
}

/* TestStrategyCommentEmptyString verifies empty string handling */
func TestStrategyCommentEmptyString(t *testing.T) {
	/* Simulate: strategy.entry("Long", strategy.long, 1, comment="") */
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
						&ast.Literal{Value: "Long"},
						&ast.MemberExpression{
							Object:   &ast.Identifier{Name: "strategy"},
							Property: &ast.Identifier{Name: "long"},
						},
						&ast.Literal{Value: 1.0},
						&ast.ObjectExpression{
							NodeType: ast.TypeObjectExpression,
							Properties: []ast.Property{
								{
									NodeType: ast.TypeProperty,
									Key:      &ast.Identifier{Name: "comment"},
									Value:    &ast.Literal{Value: ""},
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

	/* Verify explicit empty string preserved */
	if !strings.Contains(code.FunctionBody, `strat.Entry("Long", strategy.Long, 1, "")`) {
		t.Errorf("Expected empty string in Entry call, got:\n%s", code.FunctionBody)
	}
}
