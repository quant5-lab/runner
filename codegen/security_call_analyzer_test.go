package codegen

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestSecurityCallAnalyzer_SumWithConditional(t *testing.T) {
	g := newTestGenerator()
	analyzer := NewSecurityCallAnalyzer(g)

	program := &ast.Program{
		Body: []ast.Node{
			&ast.VariableDeclaration{
				Declarations: []ast.VariableDeclarator{
					{
						ID: &ast.Identifier{Name: "total"},
						Init: &ast.CallExpression{
							Callee: &ast.Identifier{Name: "sum"},
							Arguments: []ast.Expression{
								&ast.ConditionalExpression{
									Test: &ast.BinaryExpression{
										Left:     &ast.Identifier{Name: "close"},
										Operator: ">",
										Right:    &ast.Identifier{Name: "open"},
									},
									Consequent: &ast.Literal{Value: 1.0},
									Alternate:  &ast.Literal{Value: 0.0},
								},
								&ast.Literal{Value: 10.0},
							},
						},
					},
				},
			},
		},
	}

	analyzer.Analyze(program)

	if len(g.tempVarMgr.varToCallInfo) == 0 {
		t.Error("Expected sum with conditional to register temp var")
	}
	if len(g.tempVarMgr.conditionalVars) == 0 {
		t.Error("Expected conditional in sum to be registered")
	}
}

func TestSecurityCallAnalyzer_NestedMathWithTA(t *testing.T) {
	g := newTestGenerator()
	analyzer := NewSecurityCallAnalyzer(g)

	program := &ast.Program{
		Body: []ast.Node{
			&ast.VariableDeclaration{
				Declarations: []ast.VariableDeclarator{
					{
						ID: &ast.Identifier{Name: "result"},
						Init: &ast.CallExpression{
							Callee: &ast.MemberExpression{
								Object:   &ast.Identifier{Name: "math"},
								Property: &ast.Identifier{Name: "abs"},
							},
							Arguments: []ast.Expression{
								&ast.CallExpression{
									Callee: &ast.MemberExpression{
										Object:   &ast.Identifier{Name: "ta"},
										Property: &ast.Identifier{Name: "sma"},
									},
									Arguments: []ast.Expression{
										&ast.Identifier{Name: "close"},
										&ast.Literal{Value: 20.0},
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
		t.Error("Expected nested TA call in math function to register temp var")
	}
}

func TestSecurityCallAnalyzer_SkipsRuntimeOnly(t *testing.T) {
	g := newTestGenerator()
	analyzer := NewSecurityCallAnalyzer(g)

	program := &ast.Program{
		Body: []ast.Node{
			&ast.VariableDeclaration{
				Declarations: []ast.VariableDeclarator{
					{
						ID: &ast.Identifier{Name: "ph"},
						Init: &ast.CallExpression{
							Callee: &ast.Identifier{Name: "pivothigh"},
							Arguments: []ast.Expression{
								&ast.Literal{Value: 5.0},
								&ast.Literal{Value: 5.0},
							},
						},
					},
				},
			},
		},
	}

	analyzer.Analyze(program)

	if len(g.tempVarMgr.varToCallInfo) != 0 {
		t.Error("Expected pivothigh (runtime-only) to be skipped")
	}
}
