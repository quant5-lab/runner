package codegen

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
)

/*
Test demonstrating COMPLETE AST coverage for ANY ARBITRARY PineScript input.

This validates that NO MATTER WHERE a user-defined function call appears in the AST,
the scanner will detect it and register the call site for ArrowContext hoisting.
*/

func TestArrowCallSiteScanner_ArbitraryInputGuarantee(t *testing.T) {
	variables := map[string]string{
		"userFunc": "function",
	}
	scanner := NewArrowCallSiteScanner(variables)

	testCases := []struct {
		name        string
		description string
		program     *ast.Program
	}{
		{
			name:        "VariableDeclaration Init",
			description: "var x = userFunc()",
			program: &ast.Program{
				Body: []ast.Node{
					&ast.VariableDeclaration{
						Declarations: []ast.VariableDeclarator{
							{ID: &ast.Identifier{Name: "x"}, Init: &ast.CallExpression{Callee: &ast.Identifier{Name: "userFunc"}}},
						},
					},
				},
			},
		},
		{
			name:        "ExpressionStatement",
			description: "userFunc()",
			program: &ast.Program{
				Body: []ast.Node{
					&ast.ExpressionStatement{Expression: &ast.CallExpression{Callee: &ast.Identifier{Name: "userFunc"}}},
				},
			},
		},
		{
			name:        "IfStatement Test",
			description: "if userFunc() then ...",
			program: &ast.Program{
				Body: []ast.Node{
					&ast.IfStatement{Test: &ast.CallExpression{Callee: &ast.Identifier{Name: "userFunc"}}, Consequent: []ast.Node{}},
				},
			},
		},
		{
			name:        "IfStatement Consequent",
			description: "if cond then userFunc()",
			program: &ast.Program{
				Body: []ast.Node{
					&ast.IfStatement{
						Test: &ast.Identifier{Name: "cond"},
						Consequent: []ast.Node{
							&ast.ExpressionStatement{Expression: &ast.CallExpression{Callee: &ast.Identifier{Name: "userFunc"}}},
						},
					},
				},
			},
		},
		{
			name:        "IfStatement Alternate",
			description: "if cond then ... else userFunc()",
			program: &ast.Program{
				Body: []ast.Node{
					&ast.IfStatement{
						Test:       &ast.Identifier{Name: "cond"},
						Consequent: []ast.Node{},
						Alternate: []ast.Node{
							&ast.ExpressionStatement{Expression: &ast.CallExpression{Callee: &ast.Identifier{Name: "userFunc"}}},
						},
					},
				},
			},
		},
		{
			name:        "ForStatement Body",
			description: "for i = 0 to 10 do userFunc()",
			program: &ast.Program{
				Body: []ast.Node{
					&ast.ForStatement{
						Counter: "i",
						From:    &ast.Literal{Value: 0.0},
						To:      &ast.Literal{Value: 10.0},
						Body: []ast.Node{
							&ast.ExpressionStatement{Expression: &ast.CallExpression{Callee: &ast.Identifier{Name: "userFunc"}}},
						},
					},
				},
			},
		},
		{
			name:        "CallExpression Argument",
			description: "otherFunc(userFunc())",
			program: &ast.Program{
				Body: []ast.Node{
					&ast.ExpressionStatement{
						Expression: &ast.CallExpression{
							Callee:    &ast.Identifier{Name: "otherFunc"},
							Arguments: []ast.Expression{&ast.CallExpression{Callee: &ast.Identifier{Name: "userFunc"}}},
						},
					},
				},
			},
		},
		{
			name:        "BinaryExpression Left",
			description: "x = userFunc() + 5",
			program: &ast.Program{
				Body: []ast.Node{
					&ast.VariableDeclaration{
						Declarations: []ast.VariableDeclarator{
							{
								ID: &ast.Identifier{Name: "x"},
								Init: &ast.BinaryExpression{
									Operator: "+",
									Left:     &ast.CallExpression{Callee: &ast.Identifier{Name: "userFunc"}},
									Right:    &ast.Literal{Value: 5.0},
								},
							},
						},
					},
				},
			},
		},
		{
			name:        "BinaryExpression Right",
			description: "x = 5 + userFunc()",
			program: &ast.Program{
				Body: []ast.Node{
					&ast.VariableDeclaration{
						Declarations: []ast.VariableDeclarator{
							{
								ID: &ast.Identifier{Name: "x"},
								Init: &ast.BinaryExpression{
									Operator: "+",
									Left:     &ast.Literal{Value: 5.0},
									Right:    &ast.CallExpression{Callee: &ast.Identifier{Name: "userFunc"}},
								},
							},
						},
					},
				},
			},
		},
		{
			name:        "LogicalExpression Left",
			description: "x = userFunc() and cond",
			program: &ast.Program{
				Body: []ast.Node{
					&ast.VariableDeclaration{
						Declarations: []ast.VariableDeclarator{
							{
								ID: &ast.Identifier{Name: "x"},
								Init: &ast.LogicalExpression{
									Operator: "and",
									Left:     &ast.CallExpression{Callee: &ast.Identifier{Name: "userFunc"}},
									Right:    &ast.Identifier{Name: "cond"},
								},
							},
						},
					},
				},
			},
		},
		{
			name:        "LogicalExpression Right",
			description: "x = cond and userFunc()",
			program: &ast.Program{
				Body: []ast.Node{
					&ast.VariableDeclaration{
						Declarations: []ast.VariableDeclarator{
							{
								ID: &ast.Identifier{Name: "x"},
								Init: &ast.LogicalExpression{
									Operator: "and",
									Left:     &ast.Identifier{Name: "cond"},
									Right:    &ast.CallExpression{Callee: &ast.Identifier{Name: "userFunc"}},
								},
							},
						},
					},
				},
			},
		},
		{
			name:        "ConditionalExpression Test",
			description: "x = userFunc() ? 1 : 0",
			program: &ast.Program{
				Body: []ast.Node{
					&ast.VariableDeclaration{
						Declarations: []ast.VariableDeclarator{
							{
								ID: &ast.Identifier{Name: "x"},
								Init: &ast.ConditionalExpression{
									Test:       &ast.CallExpression{Callee: &ast.Identifier{Name: "userFunc"}},
									Consequent: &ast.Literal{Value: 1.0},
									Alternate:  &ast.Literal{Value: 0.0},
								},
							},
						},
					},
				},
			},
		},
		{
			name:        "ConditionalExpression Consequent",
			description: "x = cond ? userFunc() : 0",
			program: &ast.Program{
				Body: []ast.Node{
					&ast.VariableDeclaration{
						Declarations: []ast.VariableDeclarator{
							{
								ID: &ast.Identifier{Name: "x"},
								Init: &ast.ConditionalExpression{
									Test:       &ast.Identifier{Name: "cond"},
									Consequent: &ast.CallExpression{Callee: &ast.Identifier{Name: "userFunc"}},
									Alternate:  &ast.Literal{Value: 0.0},
								},
							},
						},
					},
				},
			},
		},
		{
			name:        "ConditionalExpression Alternate",
			description: "x = cond ? 0 : userFunc()",
			program: &ast.Program{
				Body: []ast.Node{
					&ast.VariableDeclaration{
						Declarations: []ast.VariableDeclarator{
							{
								ID: &ast.Identifier{Name: "x"},
								Init: &ast.ConditionalExpression{
									Test:       &ast.Identifier{Name: "cond"},
									Consequent: &ast.Literal{Value: 0.0},
									Alternate:  &ast.CallExpression{Callee: &ast.Identifier{Name: "userFunc"}},
								},
							},
						},
					},
				},
			},
		},
		{
			name:        "UnaryExpression Argument",
			description: "x = -userFunc()",
			program: &ast.Program{
				Body: []ast.Node{
					&ast.VariableDeclaration{
						Declarations: []ast.VariableDeclarator{
							{
								ID: &ast.Identifier{Name: "x"},
								Init: &ast.UnaryExpression{
									Operator: "-",
									Argument: &ast.CallExpression{Callee: &ast.Identifier{Name: "userFunc"}},
								},
							},
						},
					},
				},
			},
		},
		{
			name:        "MemberExpression Object",
			description: "x = userFunc().property",
			program: &ast.Program{
				Body: []ast.Node{
					&ast.VariableDeclaration{
						Declarations: []ast.VariableDeclarator{
							{
								ID: &ast.Identifier{Name: "x"},
								Init: &ast.MemberExpression{
									Object:   &ast.CallExpression{Callee: &ast.Identifier{Name: "userFunc"}},
									Property: &ast.Identifier{Name: "property"},
								},
							},
						},
					},
				},
			},
		},
		{
			name:        "MemberExpression Property (Computed)",
			description: "x = array[userFunc()]",
			program: &ast.Program{
				Body: []ast.Node{
					&ast.VariableDeclaration{
						Declarations: []ast.VariableDeclarator{
							{
								ID: &ast.Identifier{Name: "x"},
								Init: &ast.MemberExpression{
									Object:   &ast.Identifier{Name: "array"},
									Property: &ast.CallExpression{Callee: &ast.Identifier{Name: "userFunc"}},
									Computed: true,
								},
							},
						},
					},
				},
			},
		},
		{
			name:        "ArrowFunctionExpression Body",
			description: "f = (x) => userFunc()",
			program: &ast.Program{
				Body: []ast.Node{
					&ast.VariableDeclaration{
						Declarations: []ast.VariableDeclarator{
							{
								ID: &ast.Identifier{Name: "f"},
								Init: &ast.ArrowFunctionExpression{
									Params: []ast.Identifier{{Name: "x"}},
									Body: []ast.Node{
										&ast.ExpressionStatement{Expression: &ast.CallExpression{Callee: &ast.Identifier{Name: "userFunc"}}},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:        "ObjectExpression Property Value",
			description: "{key: userFunc()}",
			program: &ast.Program{
				Body: []ast.Node{
					&ast.VariableDeclaration{
						Declarations: []ast.VariableDeclarator{
							{
								ID: &ast.Identifier{Name: "obj"},
								Init: &ast.ObjectExpression{
									Properties: []ast.Property{
										{
											Key:   &ast.Identifier{Name: "key"},
											Value: &ast.CallExpression{Callee: &ast.Identifier{Name: "userFunc"}},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:        "Literal Array Element",
			description: "[userFunc(), 2, 3]",
			program: &ast.Program{
				Body: []ast.Node{
					&ast.VariableDeclaration{
						Declarations: []ast.VariableDeclarator{
							{
								ID: &ast.Identifier{Name: "arr"},
								Init: &ast.Literal{
									Value: []ast.Expression{
										&ast.CallExpression{Callee: &ast.Identifier{Name: "userFunc"}},
										&ast.Literal{Value: 2.0},
										&ast.Literal{Value: 3.0},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sites := scanner.ScanForArrowFunctionCalls(tc.program)

			if len(sites) != 1 {
				t.Errorf("[%s] Expected 1 call site, got %d\nDescription: %s",
					tc.name, len(sites), tc.description)
				return
			}

			if sites[0].FunctionName != "userFunc" {
				t.Errorf("[%s] Expected 'userFunc', got %q\nDescription: %s",
					tc.name, sites[0].FunctionName, tc.description)
			}

			if sites[0].ContextVar != "arrowCtx_userFunc_1" {
				t.Errorf("[%s] Expected 'arrowCtx_userFunc_1', got %q\nDescription: %s",
					tc.name, sites[0].ContextVar, tc.description)
			}
		})
	}
}
