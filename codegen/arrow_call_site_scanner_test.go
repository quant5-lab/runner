package codegen

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestArrowCallSiteScanner_EmptyProgram(t *testing.T) {
	scanner := NewArrowCallSiteScanner(map[string]string{})
	program := &ast.Program{Body: []ast.Node{}}

	sites := scanner.ScanForArrowFunctionCalls(program)

	if len(sites) != 0 {
		t.Errorf("Expected no call sites from empty program, got %d", len(sites))
	}
}

func TestArrowCallSiteScanner_NoVariableDeclarations(t *testing.T) {
	scanner := NewArrowCallSiteScanner(map[string]string{})
	program := &ast.Program{
		Body: []ast.Node{
			&ast.ExpressionStatement{},
		},
	}

	sites := scanner.ScanForArrowFunctionCalls(program)

	if len(sites) != 0 {
		t.Errorf("Expected no call sites from non-variable statements, got %d", len(sites))
	}
}

func TestArrowCallSiteScanner_SingleArrowFunctionCall(t *testing.T) {
	variables := map[string]string{
		"adx": "function",
	}
	scanner := NewArrowCallSiteScanner(variables)

	program := &ast.Program{
		Body: []ast.Node{
			&ast.VariableDeclaration{
				Declarations: []ast.VariableDeclarator{
					{
						ID: &ast.Identifier{Name: "myAdx"},
						Init: &ast.CallExpression{
							Callee: &ast.Identifier{Name: "adx"},
						},
					},
				},
			},
		},
	}

	sites := scanner.ScanForArrowFunctionCalls(program)

	if len(sites) != 1 {
		t.Fatalf("Expected 1 call site, got %d", len(sites))
	}

	if sites[0].FunctionName != "adx" {
		t.Errorf("Expected function name 'adx', got %q", sites[0].FunctionName)
	}
	if sites[0].CallIndex != 1 {
		t.Errorf("Expected call index 1, got %d", sites[0].CallIndex)
	}
	if sites[0].ContextVar != "arrowCtx_adx_1" {
		t.Errorf("Expected context var 'arrowCtx_adx_1', got %q", sites[0].ContextVar)
	}
}

func TestArrowCallSiteScanner_MultipleCallsSameFunction(t *testing.T) {
	variables := map[string]string{
		"rma": "function",
	}
	scanner := NewArrowCallSiteScanner(variables)

	program := &ast.Program{
		Body: []ast.Node{
			&ast.VariableDeclaration{
				Declarations: []ast.VariableDeclarator{
					{
						ID:   &ast.Identifier{Name: "rma1"},
						Init: &ast.CallExpression{Callee: &ast.Identifier{Name: "rma"}},
					},
				},
			},
			&ast.VariableDeclaration{
				Declarations: []ast.VariableDeclarator{
					{
						ID:   &ast.Identifier{Name: "rma2"},
						Init: &ast.CallExpression{Callee: &ast.Identifier{Name: "rma"}},
					},
				},
			},
			&ast.VariableDeclaration{
				Declarations: []ast.VariableDeclarator{
					{
						ID:   &ast.Identifier{Name: "rma3"},
						Init: &ast.CallExpression{Callee: &ast.Identifier{Name: "rma"}},
					},
				},
			},
		},
	}

	sites := scanner.ScanForArrowFunctionCalls(program)

	if len(sites) != 3 {
		t.Fatalf("Expected 3 call sites, got %d", len(sites))
	}

	expectedContextVars := []string{"arrowCtx_rma_1", "arrowCtx_rma_2", "arrowCtx_rma_3"}
	for i, site := range sites {
		if site.FunctionName != "rma" {
			t.Errorf("Site %d: expected function 'rma', got %q", i, site.FunctionName)
		}
		if site.CallIndex != i+1 {
			t.Errorf("Site %d: expected call index %d, got %d", i, i+1, site.CallIndex)
		}
		if site.ContextVar != expectedContextVars[i] {
			t.Errorf("Site %d: expected context var %q, got %q", i, expectedContextVars[i], site.ContextVar)
		}
	}
}

func TestArrowCallSiteScanner_MultipleDistinctFunctions(t *testing.T) {
	variables := map[string]string{
		"adx":    "function",
		"dirmov": "function",
		"ema":    "function",
	}
	scanner := NewArrowCallSiteScanner(variables)

	program := &ast.Program{
		Body: []ast.Node{
			&ast.VariableDeclaration{
				Declarations: []ast.VariableDeclarator{
					{
						ID:   &ast.Identifier{Name: "a1"},
						Init: &ast.CallExpression{Callee: &ast.Identifier{Name: "adx"}},
					},
				},
			},
			&ast.VariableDeclaration{
				Declarations: []ast.VariableDeclarator{
					{
						ID:   &ast.Identifier{Name: "d1"},
						Init: &ast.CallExpression{Callee: &ast.Identifier{Name: "dirmov"}},
					},
				},
			},
			&ast.VariableDeclaration{
				Declarations: []ast.VariableDeclarator{
					{
						ID:   &ast.Identifier{Name: "e1"},
						Init: &ast.CallExpression{Callee: &ast.Identifier{Name: "ema"}},
					},
				},
			},
			&ast.VariableDeclaration{
				Declarations: []ast.VariableDeclarator{
					{
						ID:   &ast.Identifier{Name: "a2"},
						Init: &ast.CallExpression{Callee: &ast.Identifier{Name: "adx"}},
					},
				},
			},
		},
	}

	sites := scanner.ScanForArrowFunctionCalls(program)

	if len(sites) != 4 {
		t.Fatalf("Expected 4 call sites, got %d", len(sites))
	}

	expected := []struct {
		funcName   string
		callIndex  int
		contextVar string
	}{
		{"adx", 1, "arrowCtx_adx_1"},
		{"dirmov", 1, "arrowCtx_dirmov_1"},
		{"ema", 1, "arrowCtx_ema_1"},
		{"adx", 2, "arrowCtx_adx_2"},
	}

	for i, site := range sites {
		if site.FunctionName != expected[i].funcName {
			t.Errorf("Site %d: expected function %q, got %q", i, expected[i].funcName, site.FunctionName)
		}
		if site.CallIndex != expected[i].callIndex {
			t.Errorf("Site %d: expected call index %d, got %d", i, expected[i].callIndex, site.CallIndex)
		}
		if site.ContextVar != expected[i].contextVar {
			t.Errorf("Site %d: expected context var %q, got %q", i, expected[i].contextVar, site.ContextVar)
		}
	}
}

func TestArrowCallSiteScanner_IgnoresBuiltinFunctions(t *testing.T) {
	variables := map[string]string{
		"myFunc": "function",
	}
	scanner := NewArrowCallSiteScanner(variables)

	program := &ast.Program{
		Body: []ast.Node{
			&ast.VariableDeclaration{
				Declarations: []ast.VariableDeclarator{
					{
						ID:   &ast.Identifier{Name: "sma1"},
						Init: &ast.CallExpression{Callee: &ast.Identifier{Name: "ta.sma"}},
					},
				},
			},
			&ast.VariableDeclaration{
				Declarations: []ast.VariableDeclarator{
					{
						ID:   &ast.Identifier{Name: "ema1"},
						Init: &ast.CallExpression{Callee: &ast.Identifier{Name: "ta.ema"}},
					},
				},
			},
			&ast.VariableDeclaration{
				Declarations: []ast.VariableDeclarator{
					{
						ID:   &ast.Identifier{Name: "result"},
						Init: &ast.CallExpression{Callee: &ast.Identifier{Name: "myFunc"}},
					},
				},
			},
		},
	}

	sites := scanner.ScanForArrowFunctionCalls(program)

	if len(sites) != 1 {
		t.Fatalf("Expected 1 call site (myFunc only), got %d", len(sites))
	}

	if sites[0].FunctionName != "myFunc" {
		t.Errorf("Expected user-defined 'myFunc', got %q", sites[0].FunctionName)
	}
}

func TestArrowCallSiteScanner_IgnoresNonFunctionVariables(t *testing.T) {
	variables := map[string]string{
		"myFunc":     "function",
		"someNumber": "float",
		"someString": "string",
	}
	scanner := NewArrowCallSiteScanner(variables)

	program := &ast.Program{
		Body: []ast.Node{
			&ast.VariableDeclaration{
				Declarations: []ast.VariableDeclarator{
					{
						ID:   &ast.Identifier{Name: "n"},
						Init: &ast.CallExpression{Callee: &ast.Identifier{Name: "someNumber"}},
					},
				},
			},
			&ast.VariableDeclaration{
				Declarations: []ast.VariableDeclarator{
					{
						ID:   &ast.Identifier{Name: "f"},
						Init: &ast.CallExpression{Callee: &ast.Identifier{Name: "myFunc"}},
					},
				},
			},
		},
	}

	sites := scanner.ScanForArrowFunctionCalls(program)

	if len(sites) != 1 {
		t.Fatalf("Expected 1 call site (myFunc only), got %d", len(sites))
	}

	if sites[0].FunctionName != "myFunc" {
		t.Errorf("Expected 'myFunc', got %q", sites[0].FunctionName)
	}
}

func TestArrowCallSiteScanner_MultipleDeclaratorsInSingleStatement(t *testing.T) {
	variables := map[string]string{
		"calc": "function",
	}
	scanner := NewArrowCallSiteScanner(variables)

	program := &ast.Program{
		Body: []ast.Node{
			&ast.VariableDeclaration{
				Declarations: []ast.VariableDeclarator{
					{
						ID:   &ast.Identifier{Name: "c1"},
						Init: &ast.CallExpression{Callee: &ast.Identifier{Name: "calc"}},
					},
					{
						ID:   &ast.Identifier{Name: "c2"},
						Init: &ast.CallExpression{Callee: &ast.Identifier{Name: "calc"}},
					},
					{
						ID:   &ast.Identifier{Name: "c3"},
						Init: &ast.CallExpression{Callee: &ast.Identifier{Name: "calc"}},
					},
				},
			},
		},
	}

	sites := scanner.ScanForArrowFunctionCalls(program)

	if len(sites) != 3 {
		t.Fatalf("Expected 3 call sites from multiple declarators, got %d", len(sites))
	}

	for i := 0; i < 3; i++ {
		if sites[i].CallIndex != i+1 {
			t.Errorf("Site %d: expected call index %d, got %d", i, i+1, sites[i].CallIndex)
		}
	}
}

func TestArrowCallSiteScanner_NilInitExpression(t *testing.T) {
	variables := map[string]string{
		"func1": "function",
	}
	scanner := NewArrowCallSiteScanner(variables)

	program := &ast.Program{
		Body: []ast.Node{
			&ast.VariableDeclaration{
				Declarations: []ast.VariableDeclarator{
					{
						ID:   &ast.Identifier{Name: "uninitialized"},
						Init: nil,
					},
				},
			},
			&ast.VariableDeclaration{
				Declarations: []ast.VariableDeclarator{
					{
						ID:   &ast.Identifier{Name: "f1"},
						Init: &ast.CallExpression{Callee: &ast.Identifier{Name: "func1"}},
					},
				},
			},
		},
	}

	sites := scanner.ScanForArrowFunctionCalls(program)

	if len(sites) != 1 {
		t.Fatalf("Expected 1 call site (skipping nil init), got %d", len(sites))
	}

	if sites[0].FunctionName != "func1" {
		t.Errorf("Expected 'func1', got %q", sites[0].FunctionName)
	}
}

func TestArrowCallSiteScanner_NonCallExpressionInit(t *testing.T) {
	variables := map[string]string{
		"arrowFunc": "function",
	}
	scanner := NewArrowCallSiteScanner(variables)

	program := &ast.Program{
		Body: []ast.Node{
			&ast.VariableDeclaration{
				Declarations: []ast.VariableDeclarator{
					{
						ID:   &ast.Identifier{Name: "literal"},
						Init: &ast.Literal{Value: 42.0},
					},
				},
			},
			&ast.VariableDeclaration{
				Declarations: []ast.VariableDeclarator{
					{
						ID:   &ast.Identifier{Name: "binary"},
						Init: &ast.BinaryExpression{Operator: "+"},
					},
				},
			},
			&ast.VariableDeclaration{
				Declarations: []ast.VariableDeclarator{
					{
						ID:   &ast.Identifier{Name: "result"},
						Init: &ast.CallExpression{Callee: &ast.Identifier{Name: "arrowFunc"}},
					},
				},
			},
		},
	}

	sites := scanner.ScanForArrowFunctionCalls(program)

	if len(sites) != 1 {
		t.Fatalf("Expected 1 call site (skipping non-call expressions), got %d", len(sites))
	}

	if sites[0].FunctionName != "arrowFunc" {
		t.Errorf("Expected 'arrowFunc', got %q", sites[0].FunctionName)
	}
}

func TestArrowCallSiteScanner_MemberExpressionCallee(t *testing.T) {
	variables := map[string]string{}
	scanner := NewArrowCallSiteScanner(variables)

	program := &ast.Program{
		Body: []ast.Node{
			&ast.VariableDeclaration{
				Declarations: []ast.VariableDeclarator{
					{
						ID: &ast.Identifier{Name: "result"},
						Init: &ast.CallExpression{
							Callee: &ast.MemberExpression{
								Object:   &ast.Identifier{Name: "ta"},
								Property: &ast.Identifier{Name: "sma"},
							},
						},
					},
				},
			},
		},
	}

	sites := scanner.ScanForArrowFunctionCalls(program)

	if len(sites) != 0 {
		t.Errorf("Expected 0 call sites (member expressions are built-ins), got %d", len(sites))
	}
}

func TestArrowCallSiteScanner_OrderPreservation(t *testing.T) {
	variables := map[string]string{
		"first":  "function",
		"second": "function",
		"third":  "function",
	}
	scanner := NewArrowCallSiteScanner(variables)

	program := &ast.Program{
		Body: []ast.Node{
			&ast.VariableDeclaration{
				Declarations: []ast.VariableDeclarator{
					{
						ID:   &ast.Identifier{Name: "a"},
						Init: &ast.CallExpression{Callee: &ast.Identifier{Name: "first"}},
					},
				},
			},
			&ast.VariableDeclaration{
				Declarations: []ast.VariableDeclarator{
					{
						ID:   &ast.Identifier{Name: "b"},
						Init: &ast.CallExpression{Callee: &ast.Identifier{Name: "second"}},
					},
				},
			},
			&ast.VariableDeclaration{
				Declarations: []ast.VariableDeclarator{
					{
						ID:   &ast.Identifier{Name: "c"},
						Init: &ast.CallExpression{Callee: &ast.Identifier{Name: "third"}},
					},
				},
			},
		},
	}

	sites := scanner.ScanForArrowFunctionCalls(program)

	if len(sites) != 3 {
		t.Fatalf("Expected 3 call sites, got %d", len(sites))
	}

	expectedOrder := []string{"first", "second", "third"}
	for i, site := range sites {
		if site.FunctionName != expectedOrder[i] {
			t.Errorf("Order violation at position %d: expected %q, got %q", i, expectedOrder[i], site.FunctionName)
		}
	}
}

func TestArrowCallSiteScanner_LargeProgramStressTest(t *testing.T) {
	variables := map[string]string{}
	for i := 0; i < 50; i++ {
		variables["func"+string(rune('A'+i%26))] = "function"
	}
	scanner := NewArrowCallSiteScanner(variables)

	program := &ast.Program{Body: []ast.Node{}}
	for i := 0; i < 100; i++ {
		funcName := "func" + string(rune('A'+i%26))
		program.Body = append(program.Body, &ast.VariableDeclaration{
			Declarations: []ast.VariableDeclarator{
				{
					ID:   &ast.Identifier{Name: "var" + string(rune('0'+i%10))},
					Init: &ast.CallExpression{Callee: &ast.Identifier{Name: funcName}},
				},
			},
		})
	}

	sites := scanner.ScanForArrowFunctionCalls(program)

	if len(sites) != 100 {
		t.Errorf("Stress test: expected 100 call sites, got %d", len(sites))
	}
}
