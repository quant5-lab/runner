package codegen

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestUDFTempVarAnalyzer_ConditionalInUDF(t *testing.T) {
	g := newTestGenerator()
	g.variables["myFunc"] = "function"
	analyzer := NewUDFTempVarAnalyzer(g)

	program := &ast.Program{
		Body: []ast.Node{
			&ast.VariableDeclaration{
				Declarations: []ast.VariableDeclarator{
					{
						ID: &ast.Identifier{Name: "result"},
						Init: &ast.CallExpression{
							Callee: &ast.Identifier{Name: "myFunc"},
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
		t.Error("Expected conditional in UDF call to be registered")
	}
}

func TestUDFTempVarAnalyzer_TAInUDFArgument(t *testing.T) {
	g := newTestGenerator()
	g.variables["myFunc"] = "function"
	analyzer := NewUDFTempVarAnalyzer(g)

	program := &ast.Program{
		Body: []ast.Node{
			&ast.VariableDeclaration{
				Declarations: []ast.VariableDeclarator{
					{
						ID: &ast.Identifier{Name: "result"},
						Init: &ast.CallExpression{
							Callee: &ast.Identifier{Name: "myFunc"},
							Arguments: []ast.Expression{
								&ast.CallExpression{
									Callee: &ast.MemberExpression{
										Object:   &ast.Identifier{Name: "ta"},
										Property: &ast.Identifier{Name: "sma"},
									},
									Arguments: []ast.Expression{
										&ast.Identifier{Name: "close"},
										&ast.Literal{Value: 14.0},
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

	if len(g.tempVarMgr.varToCallInfo) == 0 {
		t.Error("Expected TA call in UDF argument to register temp var")
	}
}

func TestUDFTempVarAnalyzer_SkipsTopLevel(t *testing.T) {
	g := newTestGenerator()
	g.variables["myFunc"] = "function"
	analyzer := NewUDFTempVarAnalyzer(g)

	program := &ast.Program{
		Body: []ast.Node{
			&ast.VariableDeclaration{
				Declarations: []ast.VariableDeclarator{
					{
						ID: &ast.Identifier{Name: "result"},
						Init: &ast.CallExpression{
							Callee: &ast.Identifier{Name: "myFunc"},
							Arguments: []ast.Expression{
								&ast.CallExpression{
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
			},
		},
	}

	initialCount := len(g.tempVarMgr.varToCallInfo)
	analyzer.Analyze(program)

	if len(g.tempVarMgr.varToCallInfo) == initialCount {
		t.Error("Expected nested TA call to register temp var")
	}
}

func TestUDFTempVarAnalyzer_SkipsNonUDFCalls(t *testing.T) {
	g := newTestGenerator()
	analyzer := NewUDFTempVarAnalyzer(g)

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
								&ast.Identifier{Name: "close"},
								&ast.Literal{Value: 14.0},
							},
						},
					},
				},
			},
		},
	}

	analyzer.Analyze(program)

	if len(g.tempVarMgr.varToCallInfo) != 0 {
		t.Error("Expected non-UDF call to be skipped by UDF analyzer")
	}
}
