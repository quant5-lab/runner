package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

/* TestInlineExpressionScanner_Integration_PlotWithHoistedTA verifies end-to-end flow */
func TestInlineExpressionScanner_Integration_PlotWithHoistedTA(t *testing.T) {
	program := &ast.Program{
		Body: []ast.Node{
			&ast.ExpressionStatement{
				Expression: &ast.CallExpression{
					Callee: &ast.Identifier{Name: "plot"},
					Arguments: []ast.Expression{
						&ast.CallExpression{
							Callee: &ast.MemberExpression{
								Object:   &ast.Identifier{Name: "ta"},
								Property: &ast.Identifier{Name: "sma"},
							},
							Arguments: []ast.Expression{
								&ast.Identifier{Name: "close"},
								&ast.Literal{Value: float64(20)},
							},
						},
					},
				},
			},
		},
	}

	gen := newTestGenerator()
	scanner := NewInlineExpressionScanner(gen)

	// Step 1: Scanner finds hoistable calls
	hoistable := scanner.ScanProgram(program)
	if len(hoistable) != 1 {
		t.Fatalf("Scanner: expected 1 hoistable call, got %d", len(hoistable))
	}

	// Step 2: TempVariableManager registers the call
	for _, callInfo := range hoistable {
		gen.tempVarMgr.GetOrCreate(callInfo)
	}

	// Step 3: Verify variable was registered
	varName := gen.tempVarMgr.GetVarNameForCall(hoistable[0].Call)
	if varName == "" {
		t.Fatal("TempVariableManager: variable not registered")
	}

	if !strings.HasPrefix(varName, "ta_sma_20_") {
		t.Errorf("TempVariableManager: expected varName prefix 'ta_sma_20_', got %q", varName)
	}

	// Step 4: Generate declarations
	declCode := gen.tempVarMgr.GenerateDeclarations()

	if !strings.Contains(declCode, varName+"Series") {
		t.Errorf("GenerateDeclarations: expected Series variable declaration, got: %s", declCode)
	}

	// Step 5: Generate calculations
	calcCode, err := gen.tempVarMgr.GenerateCalculations()
	if err != nil {
		t.Fatalf("GenerateCalculations: %v", err)
	}

	if !strings.Contains(calcCode, varName+"Series.Set(") {
		t.Errorf("GenerateCalculations: expected Series.Set() call, got: %s", calcCode)
	}
}

/* TestInlineExpressionScanner_Integration_BinaryExpressionHoisting tests multiple TA hoisting */
func TestInlineExpressionScanner_Integration_BinaryExpressionHoisting(t *testing.T) {
	program := &ast.Program{
		Body: []ast.Node{
			&ast.ExpressionStatement{
				Expression: &ast.CallExpression{
					Callee: &ast.Identifier{Name: "plot"},
					Arguments: []ast.Expression{
						&ast.BinaryExpression{
							Operator: "+",
							Left: &ast.CallExpression{
								Callee: &ast.MemberExpression{
									Object:   &ast.Identifier{Name: "ta"},
									Property: &ast.Identifier{Name: "sma"},
								},
								Arguments: []ast.Expression{
									&ast.Identifier{Name: "close"},
									&ast.Literal{Value: float64(10)},
								},
							},
							Right: &ast.CallExpression{
								Callee: &ast.MemberExpression{
									Object:   &ast.Identifier{Name: "ta"},
									Property: &ast.Identifier{Name: "ema"},
								},
								Arguments: []ast.Expression{
									&ast.Identifier{Name: "close"},
									&ast.Literal{Value: float64(10)},
								},
							},
						},
					},
				},
			},
		},
	}

	gen := newTestGenerator()
	scanner := NewInlineExpressionScanner(gen)

	hoistable := scanner.ScanProgram(program)
	if len(hoistable) != 2 {
		t.Fatalf("Scanner: expected 2 hoistable calls, got %d", len(hoistable))
	}

	// Register both calls
	for _, callInfo := range hoistable {
		gen.tempVarMgr.GetOrCreate(callInfo)
	}

	// Verify both registered
	varName1 := gen.tempVarMgr.GetVarNameForCall(hoistable[0].Call)
	varName2 := gen.tempVarMgr.GetVarNameForCall(hoistable[1].Call)

	if varName1 == "" || varName2 == "" {
		t.Fatal("TempVariableManager: not all variables registered")
	}

	if varName1 == varName2 {
		t.Error("TempVariableManager: different TA functions got same variable name")
	}

	// Verify declarations contain both
	declCode := gen.tempVarMgr.GenerateDeclarations()

	if !strings.Contains(declCode, "ta_sma_10_") {
		t.Error("GenerateDeclarations: missing ta.sma declaration")
	}

	if !strings.Contains(declCode, "ta_ema_10_") {
		t.Error("GenerateDeclarations: missing ta.ema declaration")
	}
}

/* TestInlineExpressionScanner_Integration_ValueFunctionNesting tests nz(ta.sma()) pattern */
func TestInlineExpressionScanner_Integration_ValueFunctionNesting(t *testing.T) {
	program := &ast.Program{
		Body: []ast.Node{
			&ast.ExpressionStatement{
				Expression: &ast.CallExpression{
					Callee: &ast.Identifier{Name: "plot"},
					Arguments: []ast.Expression{
						&ast.CallExpression{
							Callee: &ast.Identifier{Name: "nz"},
							Arguments: []ast.Expression{
								&ast.CallExpression{
									Callee: &ast.MemberExpression{
										Object:   &ast.Identifier{Name: "ta"},
										Property: &ast.Identifier{Name: "rsi"},
									},
									Arguments: []ast.Expression{
										&ast.Identifier{Name: "close"},
										&ast.Literal{Value: float64(14)},
									},
								},
								&ast.Literal{Value: float64(50)},
							},
						},
					},
				},
			},
		},
	}

	gen := newTestGenerator()
	scanner := NewInlineExpressionScanner(gen)

	// Scanner should find ta.rsi inside nz()
	hoistable := scanner.ScanProgram(program)
	if len(hoistable) != 1 {
		t.Fatalf("Scanner: expected 1 hoistable call (ta.rsi inside nz), got %d", len(hoistable))
	}

	if hoistable[0].FuncName != "ta.rsi" {
		t.Errorf("Scanner: expected ta.rsi, got %s", hoistable[0].FuncName)
	}

	// Register and verify
	gen.tempVarMgr.GetOrCreate(hoistable[0])
	varName := gen.tempVarMgr.GetVarNameForCall(hoistable[0].Call)

	if !strings.HasPrefix(varName, "ta_rsi_14_") {
		t.Errorf("TempVariableManager: expected varName prefix 'ta_rsi_14_', got %q", varName)
	}
}

/* TestInlineExpressionScanner_Integration_DeduplicationAcrossStatements tests same TA call dedup */
func TestInlineExpressionScanner_Integration_DeduplicationAcrossStatements(t *testing.T) {
	smaCall := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "ta"},
			Property: &ast.Identifier{Name: "sma"},
		},
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "close"},
			&ast.Literal{Value: float64(50)},
		},
	}

	program := &ast.Program{
		Body: []ast.Node{
			&ast.ExpressionStatement{
				Expression: &ast.CallExpression{
					Callee:    &ast.Identifier{Name: "plot"},
					Arguments: []ast.Expression{smaCall},
				},
			},
			&ast.ExpressionStatement{
				Expression: &ast.CallExpression{
					Callee:    &ast.Identifier{Name: "plot"},
					Arguments: []ast.Expression{smaCall},
				},
			},
		},
	}

	gen := newTestGenerator()
	scanner := NewInlineExpressionScanner(gen)

	hoistable := scanner.ScanProgram(program)

	// Scanner should deduplicate same call
	if len(hoistable) != 1 {
		t.Fatalf("Scanner: expected 1 hoistable call (deduplicated), got %d", len(hoistable))
	}

	// Register
	gen.tempVarMgr.GetOrCreate(hoistable[0])

	// Should only generate one declaration
	declCode := gen.tempVarMgr.GenerateDeclarations()

	// Count occurrences of "var ta_sma_50_"
	count := strings.Count(declCode, "var ta_sma_50_")
	if count != 1 {
		t.Errorf("GenerateDeclarations: expected 1 declaration for deduplicated call, got %d", count)
	}
}

/* TestInlineExpressionScanner_Integration_NoInterferenceWithArrowFunctions verifies arrow functions unaffected */
func TestInlineExpressionScanner_Integration_NoInterferenceWithArrowFunctions(t *testing.T) {
	program := &ast.Program{
		Body: []ast.Node{
			&ast.VariableDeclaration{
				Declarations: []ast.VariableDeclarator{
					{
						ID: &ast.Identifier{Name: "myFunc"},
						Init: &ast.ArrowFunctionExpression{
							Params: []ast.Identifier{
								{Name: "src"},
								{Name: "length"},
							},
							Body: []ast.Node{
								&ast.ExpressionStatement{
									Expression: &ast.CallExpression{
										Callee: &ast.MemberExpression{
											Object:   &ast.Identifier{Name: "ta"},
											Property: &ast.Identifier{Name: "sma"},
										},
										Arguments: []ast.Expression{
											&ast.Identifier{Name: "src"},
											&ast.Identifier{Name: "length"},
										},
									},
								},
							},
						},
					},
				},
			},
			&ast.ExpressionStatement{
				Expression: &ast.CallExpression{
					Callee: &ast.Identifier{Name: "plot"},
					Arguments: []ast.Expression{
						&ast.CallExpression{
							Callee: &ast.Identifier{Name: "myFunc"},
							Arguments: []ast.Expression{
								&ast.Identifier{Name: "close"},
								&ast.Literal{Value: float64(20)},
							},
						},
					},
				},
			},
		},
	}

	gen := newTestGenerator()
	gen.variables["myFunc"] = "function"
	scanner := NewInlineExpressionScanner(gen)

	hoistable := scanner.ScanProgram(program)

	// Should NOT hoist user-defined function calls
	if len(hoistable) != 0 {
		t.Errorf("Scanner: expected 0 hoistable calls (arrow functions not hoisted), got %d", len(hoistable))
	}
}

/* TestInlineExpressionScanner_Integration_ComplexNesting tests realistic complex expression */
func TestInlineExpressionScanner_Integration_ComplexNesting(t *testing.T) {
	program := &ast.Program{
		Body: []ast.Node{
			&ast.ExpressionStatement{
				Expression: &ast.CallExpression{
					Callee: &ast.Identifier{Name: "plot"},
					Arguments: []ast.Expression{
						&ast.ConditionalExpression{
							Test: &ast.BinaryExpression{
								Operator: ">",
								Left:     &ast.Identifier{Name: "close"},
								Right:    &ast.Identifier{Name: "open"},
							},
							Consequent: &ast.CallExpression{
								Callee: &ast.Identifier{Name: "nz"},
								Arguments: []ast.Expression{
									&ast.CallExpression{
										Callee: &ast.MemberExpression{
											Object:   &ast.Identifier{Name: "ta"},
											Property: &ast.Identifier{Name: "sma"},
										},
										Arguments: []ast.Expression{
											&ast.Identifier{Name: "high"},
											&ast.Literal{Value: float64(20)},
										},
									},
									&ast.Literal{Value: float64(0)},
								},
							},
							Alternate: &ast.BinaryExpression{
								Operator: "*",
								Left: &ast.CallExpression{
									Callee: &ast.MemberExpression{
										Object:   &ast.Identifier{Name: "ta"},
										Property: &ast.Identifier{Name: "ema"},
									},
									Arguments: []ast.Expression{
										&ast.Identifier{Name: "low"},
										&ast.Literal{Value: float64(50)},
									},
								},
								Right: &ast.Literal{Value: float64(2)},
							},
						},
					},
				},
			},
		},
	}

	gen := newTestGenerator()
	scanner := NewInlineExpressionScanner(gen)

	// Should find: ta.sma(high,20) in nz() in consequent, ta.ema(low,50) in alternate
	hoistable := scanner.ScanProgram(program)
	if len(hoistable) != 2 {
		t.Fatalf("Scanner: expected 2 hoistable calls in complex ternary, got %d", len(hoistable))
	}

	funcNames := make(map[string]bool)
	for _, call := range hoistable {
		funcNames[call.FuncName] = true
		gen.tempVarMgr.GetOrCreate(call)
	}

	if !funcNames["ta.sma"] {
		t.Error("Scanner: missed ta.sma in nested ternary consequent")
	}
	if !funcNames["ta.ema"] {
		t.Error("Scanner: missed ta.ema in ternary alternate")
	}

	// Verify both declared
	declCode := gen.tempVarMgr.GenerateDeclarations()

	if !strings.Contains(declCode, "ta_sma_20_") || !strings.Contains(declCode, "ta_ema_50_") {
		t.Error("GenerateDeclarations: missing expected Series declarations for complex expression")
	}
}
