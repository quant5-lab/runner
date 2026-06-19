package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

/* TestStrategyCommentTernarySimple verifies ternary expression in comment parameter */
func TestStrategyCommentTernarySimple(t *testing.T) {
	/* Simulate: strategy.entry("Trade", dir, comment=bullish ? "Long signal" : "Short signal") */
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
						&ast.MemberExpression{Object: &ast.Identifier{Name: "strategy"}, Property: &ast.Identifier{Name: "long"}},
						&ast.ObjectExpression{
							NodeType: ast.TypeObjectExpression,
							Properties: []ast.Property{
								{
									NodeType: ast.TypeProperty,
									Key:      &ast.Identifier{Name: "comment"},
									Value: &ast.ConditionalExpression{
										Test:       &ast.Identifier{Name: "bullish"},
										Consequent: &ast.Literal{Value: "Long signal"},
										Alternate:  &ast.Literal{Value: "Short signal"},
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

	/* Verify ternary generates IIFE with string return */
	if !strings.Contains(code.FunctionBody, "func() string {") {
		t.Errorf("Expected IIFE function for ternary comment, got:\n%s", code.FunctionBody)
	}
	if !strings.Contains(code.FunctionBody, `return "Long signal"`) {
		t.Errorf("Expected true branch with Long signal, got:\n%s", code.FunctionBody)
	}
	if !strings.Contains(code.FunctionBody, `return "Short signal"`) {
		t.Errorf("Expected false branch with Short signal, got:\n%s", code.FunctionBody)
	}
	if !strings.Contains(code.FunctionBody, "bullishSeries.GetCurrent()") {
		t.Errorf("Expected condition using Series access, got:\n%s", code.FunctionBody)
	}
}

/* TestStrategyCommentTernaryWithVariables verifies ternary with variable string branches */
func TestStrategyCommentTernaryWithVariables(t *testing.T) {
	/* Simulate: strategy.entry("Trade", dir, comment=condition ? long_msg : short_msg) */
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
						&ast.MemberExpression{Object: &ast.Identifier{Name: "strategy"}, Property: &ast.Identifier{Name: "long"}},
						&ast.ObjectExpression{
							NodeType: ast.TypeObjectExpression,
							Properties: []ast.Property{
								{
									NodeType: ast.TypeProperty,
									Key:      &ast.Identifier{Name: "comment"},
									Value: &ast.ConditionalExpression{
										Test:       &ast.Identifier{Name: "condition"},
										Consequent: &ast.Identifier{Name: "long_msg"},
										Alternate:  &ast.Identifier{Name: "short_msg"},
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

	/* Verify ternary branches use Series.GetCurrent() for variable references */
	if !strings.Contains(code.FunctionBody, "long_msgSeries.GetCurrent()") {
		t.Errorf("Expected long_msg Series access in true branch, got:\n%s", code.FunctionBody)
	}
	if !strings.Contains(code.FunctionBody, "short_msgSeries.GetCurrent()") {
		t.Errorf("Expected short_msg Series access in false branch, got:\n%s", code.FunctionBody)
	}
}

/* TestStrategyCommentTernaryBinaryCondition verifies binary expression in ternary condition */
func TestStrategyCommentTernaryBinaryCondition(t *testing.T) {
	/* Simulate: strategy.entry("Trade", dir, comment=close > sma ? "Above SMA" : "Below SMA") */
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
						&ast.MemberExpression{Object: &ast.Identifier{Name: "strategy"}, Property: &ast.Identifier{Name: "long"}},
						&ast.ObjectExpression{
							NodeType: ast.TypeObjectExpression,
							Properties: []ast.Property{
								{
									NodeType: ast.TypeProperty,
									Key:      &ast.Identifier{Name: "comment"},
									Value: &ast.ConditionalExpression{
										Test: &ast.BinaryExpression{
											Operator: ">",
											Left:     &ast.Identifier{Name: "close"},
											Right:    &ast.Identifier{Name: "sma"},
										},
										Consequent: &ast.Literal{Value: "Above SMA"},
										Alternate:  &ast.Literal{Value: "Below SMA"},
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

	/* Verify binary condition with Series access */
	if !strings.Contains(code.FunctionBody, ">") {
		t.Errorf("Expected comparison operator in condition, got:\n%s", code.FunctionBody)
	}
	if !strings.Contains(code.FunctionBody, "smaSeries.GetCurrent()") {
		t.Errorf("Expected sma Series access in condition, got:\n%s", code.FunctionBody)
	}
	if !strings.Contains(code.FunctionBody, `"Above SMA"`) {
		t.Errorf("Expected Above SMA in true branch, got:\n%s", code.FunctionBody)
	}
	if !strings.Contains(code.FunctionBody, `"Below SMA"`) {
		t.Errorf("Expected Below SMA in false branch, got:\n%s", code.FunctionBody)
	}
}

/* TestStrategyCommentTernaryInExit verifies ternary in strategy.exit comment */
func TestStrategyCommentTernaryInExit(t *testing.T) {
	/* Simulate: strategy.exit("Exit", "Long", stop=stop_loss, limit=take_profit, comment=hit_stop ? "Stop loss" : "Take profit") */
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
									Value:    &ast.Identifier{Name: "stop_loss"},
								},
								{
									NodeType: ast.TypeProperty,
									Key:      &ast.Identifier{Name: "limit"},
									Value:    &ast.Identifier{Name: "take_profit"},
								},
								{
									NodeType: ast.TypeProperty,
									Key:      &ast.Identifier{Name: "comment"},
									Value: &ast.ConditionalExpression{
										Test:       &ast.Identifier{Name: "hit_stop"},
										Consequent: &ast.Literal{Value: "Stop loss"},
										Alternate:  &ast.Literal{Value: "Take profit"},
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

	/* Verify ternary comment in ExitWithLevels call */
	if !strings.Contains(code.FunctionBody, "func() string {") {
		t.Errorf("Expected IIFE for ternary comment in exit, got:\n%s", code.FunctionBody)
	}
	if !strings.Contains(code.FunctionBody, `"Stop loss"`) {
		t.Errorf("Expected Stop loss in exit comment, got:\n%s", code.FunctionBody)
	}
	if !strings.Contains(code.FunctionBody, `"Take profit"`) {
		t.Errorf("Expected Take profit in exit comment, got:\n%s", code.FunctionBody)
	}
}

/* TestStrategyCommentTernaryInCloseAll verifies ternary in strategy.close_all comment */
func TestStrategyCommentTernaryInCloseAll(t *testing.T) {
	/* Simulate: strategy.close_all(comment=end_of_day ? "EOD close" : "Risk limit") */
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
									Value: &ast.ConditionalExpression{
										Test:       &ast.Identifier{Name: "end_of_day"},
										Consequent: &ast.Literal{Value: "EOD close"},
										Alternate:  &ast.Literal{Value: "Risk limit"},
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

	/* Verify ternary comment in CloseAll call */
	if !strings.Contains(code.FunctionBody, "func() string {") {
		t.Errorf("Expected IIFE for ternary comment in close_all, got:\n%s", code.FunctionBody)
	}
	if !strings.Contains(code.FunctionBody, `"EOD close"`) {
		t.Errorf("Expected EOD close in close_all comment, got:\n%s", code.FunctionBody)
	}
	if !strings.Contains(code.FunctionBody, `"Risk limit"`) {
		t.Errorf("Expected Risk limit in close_all comment, got:\n%s", code.FunctionBody)
	}
}

/* TestStrategyCommentTernaryMixedTypes verifies ternary with literal and variable branches */
func TestStrategyCommentTernaryMixedTypes(t *testing.T) {
	/* Simulate: strategy.entry("Trade", strategy.long, comment=use_custom ? custom_msg : "Default signal") */
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
						&ast.MemberExpression{
							Object:   &ast.Identifier{Name: "strategy"},
							Property: &ast.Identifier{Name: "long"},
						},
						&ast.ObjectExpression{
							NodeType: ast.TypeObjectExpression,
							Properties: []ast.Property{
								{
									NodeType: ast.TypeProperty,
									Key:      &ast.Identifier{Name: "comment"},
									Value: &ast.ConditionalExpression{
										Test:       &ast.Identifier{Name: "use_custom"},
										Consequent: &ast.Identifier{Name: "custom_msg"},
										Alternate:  &ast.Literal{Value: "Default signal"},
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

	/* Verify mixed types: variable (Series access) and literal (quoted string) */
	if !strings.Contains(code.FunctionBody, "custom_msgSeries.GetCurrent()") {
		t.Errorf("Expected Series access for variable in true branch, got:\n%s", code.FunctionBody)
	}
	if !strings.Contains(code.FunctionBody, `"Default signal"`) {
		t.Errorf("Expected literal string in false branch, got:\n%s", code.FunctionBody)
	}
}
