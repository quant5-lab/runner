package codegen

import (
	"fmt"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

/* Test suite for extended AST traversal in ArrowCallSiteScanner */

func TestArrowCallSiteScanner_IfStatementCondition(t *testing.T) {
	variables := map[string]string{
		"simpleCount": "function",
	}
	scanner := NewArrowCallSiteScanner(variables)

	program := &ast.Program{
		Body: []ast.Node{
			&ast.IfStatement{
				Test: &ast.BinaryExpression{
					Operator: "==",
					Left:     &ast.CallExpression{Callee: &ast.Identifier{Name: "simpleCount"}},
					Right:    &ast.Identifier{Name: "lookback"},
				},
				Consequent: []ast.Node{},
			},
		},
	}

	sites := scanner.ScanForArrowFunctionCalls(program)

	if len(sites) != 1 {
		t.Fatalf("[SCANNER_IF_COND] Expected 1 call site in if condition, got %d", len(sites))
	}

	if sites[0].FunctionName != "simpleCount" {
		t.Errorf("[SCANNER_IF_COND] Expected 'simpleCount', got %q", sites[0].FunctionName)
	}
	if sites[0].ContextVar != "arrowCtx_simpleCount_1" {
		t.Errorf("[SCANNER_IF_COND] Expected 'arrowCtx_simpleCount_1', got %q", sites[0].ContextVar)
	}
}

func TestArrowCallSiteScanner_IfStatementConsequent(t *testing.T) {
	variables := map[string]string{
		"myCalc": "function",
	}
	scanner := NewArrowCallSiteScanner(variables)

	program := &ast.Program{
		Body: []ast.Node{
			&ast.IfStatement{
				Test: &ast.Identifier{Name: "condition"},
				Consequent: []ast.Node{
					&ast.VariableDeclaration{
						Declarations: []ast.VariableDeclarator{
							{
								ID:   &ast.Identifier{Name: "result"},
								Init: &ast.CallExpression{Callee: &ast.Identifier{Name: "myCalc"}},
							},
						},
					},
				},
			},
		},
	}

	sites := scanner.ScanForArrowFunctionCalls(program)

	if len(sites) != 1 {
		t.Fatalf("[SCANNER_IF_CONSEQ] Expected 1 call site in consequent, got %d", len(sites))
	}
	if sites[0].FunctionName != "myCalc" {
		t.Errorf("[SCANNER_IF_CONSEQ] Expected 'myCalc', got %q", sites[0].FunctionName)
	}
}

func TestArrowCallSiteScanner_IfStatementAlternate(t *testing.T) {
	variables := map[string]string{
		"altFunc": "function",
	}
	scanner := NewArrowCallSiteScanner(variables)

	program := &ast.Program{
		Body: []ast.Node{
			&ast.IfStatement{
				Test:       &ast.Identifier{Name: "condition"},
				Consequent: []ast.Node{},
				Alternate: []ast.Node{
					&ast.VariableDeclaration{
						Declarations: []ast.VariableDeclarator{
							{
								ID:   &ast.Identifier{Name: "result"},
								Init: &ast.CallExpression{Callee: &ast.Identifier{Name: "altFunc"}},
							},
						},
					},
				},
			},
		},
	}

	sites := scanner.ScanForArrowFunctionCalls(program)

	if len(sites) != 1 {
		t.Fatalf("[SCANNER_IF_ALT] Expected 1 call site in alternate, got %d", len(sites))
	}
	if sites[0].FunctionName != "altFunc" {
		t.Errorf("[SCANNER_IF_ALT] Expected 'altFunc', got %q", sites[0].FunctionName)
	}
}

func TestArrowCallSiteScanner_ExpressionStatement(t *testing.T) {
	variables := map[string]string{
		"standalone": "function",
	}
	scanner := NewArrowCallSiteScanner(variables)

	program := &ast.Program{
		Body: []ast.Node{
			&ast.ExpressionStatement{
				Expression: &ast.CallExpression{
					Callee: &ast.Identifier{Name: "standalone"},
				},
			},
		},
	}

	sites := scanner.ScanForArrowFunctionCalls(program)

	if len(sites) != 1 {
		t.Fatalf("[SCANNER_EXPR_STMT] Expected 1 call site in expression statement, got %d", len(sites))
	}
	if sites[0].FunctionName != "standalone" {
		t.Errorf("[SCANNER_EXPR_STMT] Expected 'standalone', got %q", sites[0].FunctionName)
	}
}

func TestArrowCallSiteScanner_NestedCallInArguments(t *testing.T) {
	variables := map[string]string{
		"outer": "function",
		"inner": "function",
	}
	scanner := NewArrowCallSiteScanner(variables)

	program := &ast.Program{
		Body: []ast.Node{
			&ast.VariableDeclaration{
				Declarations: []ast.VariableDeclarator{
					{
						ID: &ast.Identifier{Name: "result"},
						Init: &ast.CallExpression{
							Callee: &ast.Identifier{Name: "outer"},
							Arguments: []ast.Expression{
								&ast.CallExpression{
									Callee: &ast.Identifier{Name: "inner"},
								},
							},
						},
					},
				},
			},
		},
	}

	sites := scanner.ScanForArrowFunctionCalls(program)

	if len(sites) != 2 {
		t.Fatalf("[SCANNER_NESTED_ARGS] Expected 2 call sites (outer + inner), got %d", len(sites))
	}

	/* Ensure both functions detected */
	funcNames := map[string]bool{}
	for _, site := range sites {
		funcNames[site.FunctionName] = true
	}

	if !funcNames["outer"] {
		t.Errorf("[SCANNER_NESTED_ARGS] Missing 'outer' function")
	}
	if !funcNames["inner"] {
		t.Errorf("[SCANNER_NESTED_ARGS] Missing 'inner' function")
	}
}

func TestArrowCallSiteScanner_BinaryExpressionBothSides(t *testing.T) {
	variables := map[string]string{
		"leftFunc":  "function",
		"rightFunc": "function",
	}
	scanner := NewArrowCallSiteScanner(variables)

	program := &ast.Program{
		Body: []ast.Node{
			&ast.VariableDeclaration{
				Declarations: []ast.VariableDeclarator{
					{
						ID: &ast.Identifier{Name: "result"},
						Init: &ast.BinaryExpression{
							Operator: "+",
							Left:     &ast.CallExpression{Callee: &ast.Identifier{Name: "leftFunc"}},
							Right:    &ast.CallExpression{Callee: &ast.Identifier{Name: "rightFunc"}},
						},
					},
				},
			},
		},
	}

	sites := scanner.ScanForArrowFunctionCalls(program)

	if len(sites) != 2 {
		t.Fatalf("[SCANNER_BINARY] Expected 2 call sites in binary expression, got %d", len(sites))
	}

	funcNames := map[string]bool{}
	for _, site := range sites {
		funcNames[site.FunctionName] = true
	}

	if !funcNames["leftFunc"] || !funcNames["rightFunc"] {
		t.Errorf("[SCANNER_BINARY] Missing expected functions. Found: %v", funcNames)
	}
}

func TestArrowCallSiteScanner_ConditionalExpression(t *testing.T) {
	variables := map[string]string{
		"testFunc":  "function",
		"trueFunc":  "function",
		"falseFunc": "function",
	}
	scanner := NewArrowCallSiteScanner(variables)

	program := &ast.Program{
		Body: []ast.Node{
			&ast.VariableDeclaration{
				Declarations: []ast.VariableDeclarator{
					{
						ID: &ast.Identifier{Name: "result"},
						Init: &ast.ConditionalExpression{
							Test:       &ast.CallExpression{Callee: &ast.Identifier{Name: "testFunc"}},
							Consequent: &ast.CallExpression{Callee: &ast.Identifier{Name: "trueFunc"}},
							Alternate:  &ast.CallExpression{Callee: &ast.Identifier{Name: "falseFunc"}},
						},
					},
				},
			},
		},
	}

	sites := scanner.ScanForArrowFunctionCalls(program)

	if len(sites) != 3 {
		t.Fatalf("[SCANNER_TERNARY] Expected 3 call sites in ternary, got %d", len(sites))
	}

	funcNames := map[string]bool{}
	for _, site := range sites {
		funcNames[site.FunctionName] = true
	}

	if !funcNames["testFunc"] || !funcNames["trueFunc"] || !funcNames["falseFunc"] {
		t.Errorf("[SCANNER_TERNARY] Missing functions. Found: %v", funcNames)
	}
}

func TestArrowCallSiteScanner_ForLoopBody(t *testing.T) {
	variables := map[string]string{
		"loopFunc": "function",
	}
	scanner := NewArrowCallSiteScanner(variables)

	program := &ast.Program{
		Body: []ast.Node{
			&ast.ForStatement{
				Counter: "i",
				From:    &ast.Literal{Value: 0.0},
				To:      &ast.Literal{Value: 10.0},
				Body: []ast.Node{
					&ast.VariableDeclaration{
						Declarations: []ast.VariableDeclarator{
							{
								ID:   &ast.Identifier{Name: "result"},
								Init: &ast.CallExpression{Callee: &ast.Identifier{Name: "loopFunc"}},
							},
						},
					},
				},
			},
		},
	}

	sites := scanner.ScanForArrowFunctionCalls(program)

	if len(sites) != 1 {
		t.Fatalf("[SCANNER_FOR_BODY] Expected 1 call site in for loop, got %d", len(sites))
	}
	if sites[0].FunctionName != "loopFunc" {
		t.Errorf("[SCANNER_FOR_BODY] Expected 'loopFunc', got %q", sites[0].FunctionName)
	}
}

func TestArrowCallSiteScanner_ArrowFunctionBody(t *testing.T) {
	variables := map[string]string{
		"innerFunc": "function",
	}
	scanner := NewArrowCallSiteScanner(variables)

	program := &ast.Program{
		Body: []ast.Node{
			&ast.VariableDeclaration{
				Declarations: []ast.VariableDeclarator{
					{
						ID: &ast.Identifier{Name: "arrowDef"},
						Init: &ast.ArrowFunctionExpression{
							Params: []ast.Identifier{{Name: "x"}},
							Body: []ast.Node{
								&ast.VariableDeclaration{
									Declarations: []ast.VariableDeclarator{
										{
											ID:   &ast.Identifier{Name: "inner"},
											Init: &ast.CallExpression{Callee: &ast.Identifier{Name: "innerFunc"}},
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

	sites := scanner.ScanForArrowFunctionCalls(program)

	if len(sites) != 1 {
		t.Fatalf("[SCANNER_ARROW_BODY] Expected 1 call site in arrow function body, got %d", len(sites))
	}
	if sites[0].FunctionName != "innerFunc" {
		t.Errorf("[SCANNER_ARROW_BODY] Expected 'innerFunc', got %q", sites[0].FunctionName)
	}
}

func TestArrowCallSiteScanner_ComplexNestedStructure(t *testing.T) {
	variables := map[string]string{
		"func1": "function",
		"func2": "function",
		"func3": "function",
		"func4": "function",
	}
	scanner := NewArrowCallSiteScanner(variables)

	program := &ast.Program{
		Body: []ast.Node{
			&ast.IfStatement{
				Test: &ast.CallExpression{Callee: &ast.Identifier{Name: "func1"}},
				Consequent: []ast.Node{
					&ast.ForStatement{
						Counter: "i",
						From:    &ast.Literal{Value: 0.0},
						To:      &ast.Literal{Value: 5.0},
						Body: []ast.Node{
							&ast.VariableDeclaration{
								Declarations: []ast.VariableDeclarator{
									{
										ID: &ast.Identifier{Name: "nested"},
										Init: &ast.BinaryExpression{
											Operator: "+",
											Left:     &ast.CallExpression{Callee: &ast.Identifier{Name: "func2"}},
											Right:    &ast.CallExpression{Callee: &ast.Identifier{Name: "func3"}},
										},
									},
								},
							},
						},
					},
				},
				Alternate: []ast.Node{
					&ast.ExpressionStatement{
						Expression: &ast.CallExpression{Callee: &ast.Identifier{Name: "func4"}},
					},
				},
			},
		},
	}

	sites := scanner.ScanForArrowFunctionCalls(program)

	if len(sites) != 4 {
		t.Fatalf("[SCANNER_COMPLEX] Expected 4 call sites in nested structure, got %d", len(sites))
	}

	funcNames := map[string]bool{}
	for _, site := range sites {
		funcNames[site.FunctionName] = true
	}

	expectedFuncs := []string{"func1", "func2", "func3", "func4"}
	for _, name := range expectedFuncs {
		if !funcNames[name] {
			t.Errorf("[SCANNER_COMPLEX] Missing function %q", name)
		}
	}
}

func TestArrowCallSiteScanner_LogicalExpression(t *testing.T) {
	variables := map[string]string{
		"check1": "function",
		"check2": "function",
	}
	scanner := NewArrowCallSiteScanner(variables)

	program := &ast.Program{
		Body: []ast.Node{
			&ast.VariableDeclaration{
				Declarations: []ast.VariableDeclarator{
					{
						ID: &ast.Identifier{Name: "result"},
						Init: &ast.LogicalExpression{
							Operator: "and",
							Left:     &ast.CallExpression{Callee: &ast.Identifier{Name: "check1"}},
							Right:    &ast.CallExpression{Callee: &ast.Identifier{Name: "check2"}},
						},
					},
				},
			},
		},
	}

	sites := scanner.ScanForArrowFunctionCalls(program)

	if len(sites) != 2 {
		t.Fatalf("[SCANNER_LOGICAL] Expected 2 call sites in logical expression, got %d", len(sites))
	}

	funcNames := map[string]bool{}
	for _, site := range sites {
		funcNames[site.FunctionName] = true
	}

	if !funcNames["check1"] || !funcNames["check2"] {
		t.Errorf("[SCANNER_LOGICAL] Missing functions. Found: %v", funcNames)
	}
}

func TestArrowCallSiteScanner_UnaryExpression(t *testing.T) {
	variables := map[string]string{
		"getValue": "function",
	}
	scanner := NewArrowCallSiteScanner(variables)

	program := &ast.Program{
		Body: []ast.Node{
			&ast.VariableDeclaration{
				Declarations: []ast.VariableDeclarator{
					{
						ID: &ast.Identifier{Name: "result"},
						Init: &ast.UnaryExpression{
							Operator: "-",
							Argument: &ast.CallExpression{Callee: &ast.Identifier{Name: "getValue"}},
						},
					},
				},
			},
		},
	}

	sites := scanner.ScanForArrowFunctionCalls(program)

	if len(sites) != 1 {
		t.Fatalf("[SCANNER_UNARY] Expected 1 call site in unary expression, got %d", len(sites))
	}
	if sites[0].FunctionName != "getValue" {
		t.Errorf("[SCANNER_UNARY] Expected 'getValue', got %q", sites[0].FunctionName)
	}
}

func TestArrowCallSiteScanner_MemberExpressionComputedProperty(t *testing.T) {
	variables := map[string]string{
		"getIndex": "function",
	}
	scanner := NewArrowCallSiteScanner(variables)

	program := &ast.Program{
		Body: []ast.Node{
			&ast.VariableDeclaration{
				Declarations: []ast.VariableDeclarator{
					{
						ID: &ast.Identifier{Name: "result"},
						Init: &ast.MemberExpression{
							Object:   &ast.Identifier{Name: "array"},
							Property: &ast.CallExpression{Callee: &ast.Identifier{Name: "getIndex"}},
							Computed: true,
						},
					},
				},
			},
		},
	}

	sites := scanner.ScanForArrowFunctionCalls(program)

	if len(sites) != 1 {
		t.Fatalf("[SCANNER_MEMBER_PROP] Expected 1 call site in member property, got %d", len(sites))
	}
	if sites[0].FunctionName != "getIndex" {
		t.Errorf("[SCANNER_MEMBER_PROP] Expected 'getIndex', got %q", sites[0].FunctionName)
	}
}

func TestArrowCallSiteScanner_ObjectExpressionProperties(t *testing.T) {
	variables := map[string]string{
		"calcStop":  "function",
		"calcLimit": "function",
	}
	scanner := NewArrowCallSiteScanner(variables)

	program := &ast.Program{
		Body: []ast.Node{
			&ast.VariableDeclaration{
				Declarations: []ast.VariableDeclarator{
					{
						ID: &ast.Identifier{Name: "config"},
						Init: &ast.ObjectExpression{
							Properties: []ast.Property{
								{
									Key:   &ast.Identifier{Name: "stop"},
									Value: &ast.CallExpression{Callee: &ast.Identifier{Name: "calcStop"}},
								},
								{
									Key:   &ast.Identifier{Name: "limit"},
									Value: &ast.CallExpression{Callee: &ast.Identifier{Name: "calcLimit"}},
								},
							},
						},
					},
				},
			},
		},
	}

	sites := scanner.ScanForArrowFunctionCalls(program)

	if len(sites) != 2 {
		t.Fatalf("[SCANNER_OBJECT_PROPS] Expected 2 call sites in object properties, got %d", len(sites))
	}

	funcNames := map[string]bool{}
	for _, site := range sites {
		funcNames[site.FunctionName] = true
	}

	if !funcNames["calcStop"] || !funcNames["calcLimit"] {
		t.Errorf("[SCANNER_OBJECT_PROPS] Missing functions. Found: %v", funcNames)
	}
}

func TestArrowCallSiteScanner_ArrayLiteralElements(t *testing.T) {
	variables := map[string]string{
		"func1": "function",
		"func2": "function",
		"func3": "function",
	}
	scanner := NewArrowCallSiteScanner(variables)

	program := &ast.Program{
		Body: []ast.Node{
			&ast.VariableDeclaration{
				Declarations: []ast.VariableDeclarator{
					{
						ID: &ast.Identifier{Name: "arr"},
						Init: &ast.Literal{
							Value: []ast.Expression{
								&ast.CallExpression{Callee: &ast.Identifier{Name: "func1"}},
								&ast.CallExpression{Callee: &ast.Identifier{Name: "func2"}},
								&ast.CallExpression{Callee: &ast.Identifier{Name: "func3"}},
							},
						},
					},
				},
			},
		},
	}

	sites := scanner.ScanForArrowFunctionCalls(program)

	if len(sites) != 3 {
		t.Fatalf("[SCANNER_ARRAY_LITERAL] Expected 3 call sites in array literal, got %d", len(sites))
	}

	funcNames := map[string]bool{}
	for _, site := range sites {
		funcNames[site.FunctionName] = true
	}

	if !funcNames["func1"] || !funcNames["func2"] || !funcNames["func3"] {
		t.Errorf("[SCANNER_ARRAY_LITERAL] Missing functions. Found: %v", funcNames)
	}
}

func TestArrowCallSiteScanner_CompleteAST_Coverage(t *testing.T) {
	variables := map[string]string{
		"f1":  "function",
		"f2":  "function",
		"f3":  "function",
		"f4":  "function",
		"f5":  "function",
		"f6":  "function",
		"f7":  "function",
		"f8":  "function",
		"f9":  "function",
		"f10": "function",
	}
	scanner := NewArrowCallSiteScanner(variables)

	program := &ast.Program{
		Body: []ast.Node{
			/* IfStatement with call in condition */
			&ast.IfStatement{
				Test: &ast.CallExpression{Callee: &ast.Identifier{Name: "f1"}},
				Consequent: []ast.Node{
					/* ExpressionStatement with call */
					&ast.ExpressionStatement{
						Expression: &ast.CallExpression{Callee: &ast.Identifier{Name: "f2"}},
					},
					/* VariableDeclaration with nested call in binary expression */
					&ast.VariableDeclaration{
						Declarations: []ast.VariableDeclarator{
							{
								ID: &ast.Identifier{Name: "x"},
								Init: &ast.BinaryExpression{
									Operator: "+",
									Left:     &ast.CallExpression{Callee: &ast.Identifier{Name: "f3"}},
									Right: &ast.ConditionalExpression{
										Test:       &ast.CallExpression{Callee: &ast.Identifier{Name: "f4"}},
										Consequent: &ast.CallExpression{Callee: &ast.Identifier{Name: "f5"}},
										Alternate:  &ast.CallExpression{Callee: &ast.Identifier{Name: "f6"}},
									},
								},
							},
						},
					},
				},
				Alternate: []ast.Node{
					/* ForStatement with call in body */
					&ast.ForStatement{
						Counter: "i",
						From:    &ast.Literal{Value: 0.0},
						To:      &ast.Literal{Value: 10.0},
						Body: []ast.Node{
							/* Object literal with call in property */
							&ast.VariableDeclaration{
								Declarations: []ast.VariableDeclarator{
									{
										ID: &ast.Identifier{Name: "obj"},
										Init: &ast.ObjectExpression{
											Properties: []ast.Property{
												{
													Key:   &ast.Identifier{Name: "val"},
													Value: &ast.CallExpression{Callee: &ast.Identifier{Name: "f7"}},
												},
											},
										},
									},
								},
							},
							/* Array literal with call */
							&ast.VariableDeclaration{
								Declarations: []ast.VariableDeclarator{
									{
										ID: &ast.Identifier{Name: "arr"},
										Init: &ast.Literal{
											Value: []ast.Expression{
												&ast.CallExpression{Callee: &ast.Identifier{Name: "f8"}},
											},
										},
									},
								},
							},
							/* Member expression computed property */
							&ast.VariableDeclaration{
								Declarations: []ast.VariableDeclarator{
									{
										ID: &ast.Identifier{Name: "elem"},
										Init: &ast.MemberExpression{
											Object:   &ast.Identifier{Name: "data"},
											Property: &ast.CallExpression{Callee: &ast.Identifier{Name: "f9"}},
											Computed: true,
										},
									},
								},
							},
						},
					},
				},
			},
			/* Arrow function with call in body */
			&ast.VariableDeclaration{
				Declarations: []ast.VariableDeclarator{
					{
						ID: &ast.Identifier{Name: "arrowDef"},
						Init: &ast.ArrowFunctionExpression{
							Params: []ast.Identifier{{Name: "x"}},
							Body: []ast.Node{
								&ast.ExpressionStatement{
									Expression: &ast.CallExpression{Callee: &ast.Identifier{Name: "f10"}},
								},
							},
						},
					},
				},
			},
		},
	}

	sites := scanner.ScanForArrowFunctionCalls(program)

	if len(sites) != 10 {
		t.Fatalf("[SCANNER_COMPLETE_COVERAGE] Expected 10 call sites covering all AST node types, got %d", len(sites))
	}

	/* Verify all 10 functions detected */
	funcNames := map[string]bool{}
	for _, site := range sites {
		funcNames[site.FunctionName] = true
	}

	for i := 1; i <= 10; i++ {
		fname := fmt.Sprintf("f%d", i)
		if !funcNames[fname] {
			t.Errorf("[SCANNER_COMPLETE_COVERAGE] Missing function %q", fname)
		}
	}
}
