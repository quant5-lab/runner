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
	program := &ast.Program{
		NodeType: ast.TypeProgram,
		Body: []ast.Node{
			&ast.VariableDeclaration{
				Declarations: []ast.VariableDeclarator{
					{ID: &ast.Identifier{Name: "signal_msg"}, Init: &ast.Literal{Value: "go"}},
				},
			},
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
	if !strings.Contains(code.FunctionBody, `strat.Entry("Long", strategy.Long, 1, signal_msg)`) {
		t.Errorf("Expected string variable comment, got:\n%s", code.FunctionBody)
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

/*
TestStrategyPositionalCommentResolvesFromDeclaredSlot verifies every strategy
order command reads a positional comment from its own Pine slot and never from a
neighbouring slot that carries a different argument.
*/
func TestStrategyPositionalCommentResolvesFromDeclaredSlot(t *testing.T) {
	handler := &StrategyActionHandler{}

	tests := []struct {
		name           string
		funcName       string
		argumentCount  int
		commentIndex   int
		comment        string
		neighbourIndex int
		neighbour      string
	}{
		{name: "entry reads slot 7", funcName: "entry", argumentCount: 8, commentIndex: 7, comment: "entry-comment", neighbourIndex: 3, neighbour: "entry-limit"},
		{name: "order reads slot 7", funcName: "order", argumentCount: 8, commentIndex: 7, comment: "order-comment", neighbourIndex: 3, neighbour: "order-limit"},
		{name: "close reads slot 1", funcName: "close", argumentCount: 3, commentIndex: 1, comment: "close-comment", neighbourIndex: 2, neighbour: "close-qty"},
		{name: "close_all reads slot 0", funcName: "close_all", argumentCount: 2, commentIndex: 0, comment: "closeall-comment", neighbourIndex: 1, neighbour: "closeall-alert"},
		{name: "exit reads slot 12", funcName: "exit", argumentCount: 13, commentIndex: 12, comment: "exit-comment", neighbourIndex: 11, neighbour: "exit-oca"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			call := positionalStrategyCall(tt.funcName, tt.argumentCount, tt.commentIndex, tt.comment, tt.neighbourIndex, tt.neighbour)

			code, err := handler.GenerateCode(newTestGenerator(), call)
			if err != nil {
				t.Fatalf("GenerateCode() unexpected error: %v", err)
			}
			if !strings.Contains(code, `"`+tt.comment+`"`) {
				t.Fatalf("GenerateCode() = %q, want comment %q from slot %d", code, tt.comment, tt.commentIndex)
			}
			if strings.Contains(code, `"`+tt.neighbour+`"`) {
				t.Fatalf("GenerateCode() = %q, must not read %q from slot %d as comment", code, tt.neighbour, tt.neighbourIndex)
			}
		})
	}
}

// positionalStrategyCall fills every argument with `na` so only the declared
// comment slot and the neighbouring marker carry a distinguishable value.
func positionalStrategyCall(funcName string, argumentCount, commentIndex int, comment string, neighbourIndex int, neighbour string) *ast.CallExpression {
	args := make([]ast.Expression, argumentCount)
	for index := range args {
		args[index] = naIdentifier()
	}
	switch funcName {
	case "entry", "order":
		args[0] = &ast.Literal{Value: "id"}
		args[1] = &ast.MemberExpression{Object: &ast.Identifier{Name: "strategy"}, Property: &ast.Identifier{Name: "long"}}
		args[2] = &ast.Literal{Value: 1.0}
	case "exit":
		args[0] = &ast.Literal{Value: "id"}
		args[1] = &ast.Literal{Value: "from_entry"}
	default:
		args[0] = &ast.Literal{Value: "id"}
	}
	args[neighbourIndex] = &ast.Literal{Value: neighbour}
	args[commentIndex] = &ast.Literal{Value: comment}

	return &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "strategy"},
			Property: &ast.Identifier{Name: funcName},
		},
		Arguments: args,
	}
}

func naIdentifier() ast.Expression {
	return &ast.Identifier{Name: "na"}
}
