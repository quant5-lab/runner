package codegen

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestStatementConditionalAnalyzer_PlotStatement(t *testing.T) {
	g := newTestGenerator()
	analyzer := NewStatementConditionalAnalyzer(g)

	program := &ast.Program{
		Body: []ast.Node{
			&ast.ExpressionStatement{
				Expression: &ast.CallExpression{
					Callee: &ast.Identifier{Name: "plot"},
					Arguments: []ast.Expression{
						&ast.ConditionalExpression{
							Test: &ast.BinaryExpression{
								Left:     &ast.Identifier{Name: "close"},
								Operator: ">",
								Right:    &ast.Identifier{Name: "open"},
							},
							Consequent: &ast.Identifier{Name: "high"},
							Alternate:  &ast.Identifier{Name: "low"},
						},
					},
				},
			},
		},
	}

	analyzer.Analyze(program)

	if len(g.tempVarMgr.conditionalVars) == 0 {
		t.Error("Expected conditional in plot statement to be registered")
	}
}

func TestStatementConditionalAnalyzer_StrategyEntry(t *testing.T) {
	g := newTestGenerator()
	analyzer := NewStatementConditionalAnalyzer(g)

	program := &ast.Program{
		Body: []ast.Node{
			&ast.ExpressionStatement{
				Expression: &ast.CallExpression{
					Callee: &ast.MemberExpression{
						Object:   &ast.Identifier{Name: "strategy"},
						Property: &ast.Identifier{Name: "entry"},
					},
					Arguments: []ast.Expression{
						&ast.Literal{Value: "Long"},
						&ast.Identifier{Name: "strategy_long"},
						&ast.ConditionalExpression{
							Test: &ast.BinaryExpression{
								Left:     &ast.Identifier{Name: "close"},
								Operator: ">",
								Right:    &ast.Identifier{Name: "open"},
							},
							Consequent: &ast.Literal{Value: true},
							Alternate:  &ast.Literal{Value: false},
						},
					},
				},
			},
		},
	}

	analyzer.Analyze(program)

	if len(g.tempVarMgr.conditionalVars) == 0 {
		t.Error("Expected conditional in strategy.entry to be registered")
	}
}

func TestStatementConditionalAnalyzer_IfStatement(t *testing.T) {
	g := newTestGenerator()
	analyzer := NewStatementConditionalAnalyzer(g)

	program := &ast.Program{
		Body: []ast.Node{
			&ast.IfStatement{
				Test: &ast.Identifier{Name: "condition"},
				Consequent: []ast.Node{
					&ast.ExpressionStatement{
						Expression: &ast.CallExpression{
							Callee: &ast.Identifier{Name: "plot"},
							Arguments: []ast.Expression{
								&ast.ConditionalExpression{
									Test: &ast.BinaryExpression{
										Left:     &ast.Identifier{Name: "close"},
										Operator: ">",
										Right:    &ast.Identifier{Name: "open"},
									},
									Consequent: &ast.Identifier{Name: "high"},
									Alternate:  &ast.Identifier{Name: "low"},
								},
							},
						},
					},
				},
			},
		},
	}

	analyzer.Analyze(program)

	if len(g.tempVarMgr.conditionalVars) == 0 {
		t.Error("Expected conditional in if-statement block to be registered")
	}
}

func TestStatementConditionalAnalyzer_NestedBlocks(t *testing.T) {
	g := newTestGenerator()
	analyzer := NewStatementConditionalAnalyzer(g)

	program := &ast.Program{
		Body: []ast.Node{
			&ast.IfStatement{
				Test: &ast.Identifier{Name: "condition1"},
				Consequent: []ast.Node{
					&ast.IfStatement{
						Test: &ast.Identifier{Name: "condition2"},
						Consequent: []ast.Node{
							&ast.ExpressionStatement{
								Expression: &ast.CallExpression{
									Callee: &ast.Identifier{Name: "plot"},
									Arguments: []ast.Expression{
										&ast.ConditionalExpression{
											Test: &ast.BinaryExpression{
												Left:     &ast.Identifier{Name: "close"},
												Operator: ">",
												Right:    &ast.Identifier{Name: "open"},
											},
											Consequent: &ast.Identifier{Name: "high"},
											Alternate:  &ast.Identifier{Name: "low"},
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

	analyzer.Analyze(program)

	if len(g.tempVarMgr.conditionalVars) == 0 {
		t.Error("Expected conditional in nested blocks to be registered")
	}
}

func TestStatementConditionalAnalyzer_MultipleConditionals(t *testing.T) {
	g := newTestGenerator()
	analyzer := NewStatementConditionalAnalyzer(g)

	program := &ast.Program{
		Body: []ast.Node{
			&ast.ExpressionStatement{
				Expression: &ast.CallExpression{
					Callee: &ast.Identifier{Name: "plot"},
					Arguments: []ast.Expression{
						&ast.ConditionalExpression{
							Test:       &ast.BinaryExpression{Left: &ast.Identifier{Name: "close"}, Operator: ">", Right: &ast.Identifier{Name: "open"}},
							Consequent: &ast.Identifier{Name: "high"},
							Alternate:  &ast.Identifier{Name: "low"},
						},
					},
				},
			},
			&ast.ExpressionStatement{
				Expression: &ast.CallExpression{
					Callee: &ast.Identifier{Name: "plot"},
					Arguments: []ast.Expression{
						&ast.ConditionalExpression{
							Test:       &ast.BinaryExpression{Left: &ast.Identifier{Name: "high"}, Operator: ">", Right: &ast.Identifier{Name: "low"}},
							Consequent: &ast.Literal{Value: 1.0},
							Alternate:  &ast.Literal{Value: 0.0},
						},
					},
				},
			},
		},
	}

	analyzer.Analyze(program)

	if len(g.tempVarMgr.conditionalVars) != 2 {
		t.Errorf("Expected 2 conditionals registered, got %d", len(g.tempVarMgr.conditionalVars))
	}
}

func TestStatementConditionalAnalyzer_SkipsNonStatementCalls(t *testing.T) {
	g := newTestGenerator()
	analyzer := NewStatementConditionalAnalyzer(g)

	program := &ast.Program{
		Body: []ast.Node{
			&ast.VariableDeclaration{
				Declarations: []ast.VariableDeclarator{
					{
						ID: &ast.Identifier{Name: "ma"},
						Init: &ast.CallExpression{
							Callee: &ast.MemberExpression{
								Object:   &ast.Identifier{Name: "ta"},
								Property: &ast.Identifier{Name: "sma"},
							},
							Arguments: []ast.Expression{
								&ast.ConditionalExpression{
									Test:       &ast.BinaryExpression{Left: &ast.Identifier{Name: "close"}, Operator: ">", Right: &ast.Identifier{Name: "open"}},
									Consequent: &ast.Identifier{Name: "high"},
									Alternate:  &ast.Identifier{Name: "low"},
								},
								&ast.Literal{Value: 14.0},
							},
						},
					},
				},
			},
		},
	}

	analyzer.Analyze(program)

	if len(g.tempVarMgr.conditionalVars) != 0 {
		t.Error("Expected statement analyzer to skip variable declarations")
	}
}
