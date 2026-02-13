package codegen

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestInlineExpressionScanner_EmptyProgram(t *testing.T) {
	program := &ast.Program{Body: []ast.Node{}}
	gen := newTestGenerator()
	scanner := NewInlineExpressionScanner(gen)

	hoistable := scanner.ScanProgram(program)

	if len(hoistable) != 0 {
		t.Errorf("Expected 0 hoistable calls in empty program, got %d", len(hoistable))
	}
}

func TestInlineExpressionScanner_PlotWithTASMA(t *testing.T) {
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
								&ast.Literal{Value: float64(14)},
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

	if len(hoistable) != 1 {
		t.Fatalf("Expected 1 hoistable call, got %d", len(hoistable))
	}

	if hoistable[0].FuncName != "ta.sma" {
		t.Errorf("Expected ta.sma, got %s", hoistable[0].FuncName)
	}
}

func TestInlineExpressionScanner_BinaryWithMultipleTA(t *testing.T) {
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
									&ast.Literal{Value: float64(14)},
								},
							},
							Right: &ast.CallExpression{
								Callee: &ast.MemberExpression{
									Object:   &ast.Identifier{Name: "ta"},
									Property: &ast.Identifier{Name: "ema"},
								},
								Arguments: []ast.Expression{
									&ast.Identifier{Name: "close"},
									&ast.Literal{Value: float64(14)},
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
		t.Fatalf("Expected 2 hoistable calls (ta.sma + ta.ema), got %d", len(hoistable))
	}

	funcNames := make(map[string]bool)
	for _, call := range hoistable {
		funcNames[call.FuncName] = true
	}

	if !funcNames["ta.sma"] {
		t.Error("Expected ta.sma to be hoistable")
	}
	if !funcNames["ta.ema"] {
		t.Error("Expected ta.ema to be hoistable")
	}
}

func TestInlineExpressionScanner_NestedTAInValueFunction(t *testing.T) {
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
										Property: &ast.Identifier{Name: "sma"},
									},
									Arguments: []ast.Expression{
										&ast.Identifier{Name: "close"},
										&ast.Literal{Value: float64(14)},
									},
								},
								&ast.Literal{Value: float64(0)},
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

	if len(hoistable) != 1 {
		t.Fatalf("Expected 1 hoistable call (ta.sma inside nz), got %d", len(hoistable))
	}

	if hoistable[0].FuncName != "ta.sma" {
		t.Errorf("Expected ta.sma, got %s", hoistable[0].FuncName)
	}
}

func TestInlineExpressionScanner_NoDuplicateRegistration(t *testing.T) {
	smaCall := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "ta"},
			Property: &ast.Identifier{Name: "sma"},
		},
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "close"},
			&ast.Literal{Value: float64(14)},
		},
	}

	program := &ast.Program{
		Body: []ast.Node{
			&ast.ExpressionStatement{
				Expression: &ast.CallExpression{
					Callee: &ast.Identifier{Name: "plot"},
					Arguments: []ast.Expression{
						&ast.BinaryExpression{
							Operator: "+",
							Left:     smaCall,
							Right:    smaCall,
						},
					},
				},
			},
		},
	}

	gen := newTestGenerator()
	scanner := NewInlineExpressionScanner(gen)

	hoistable := scanner.ScanProgram(program)

	if len(hoistable) != 1 {
		t.Errorf("Expected 1 hoistable call (same sma call used twice), got %d", len(hoistable))
	}
}

func TestInlineExpressionScanner_IgnoresNonHoistable(t *testing.T) {
	program := &ast.Program{
		Body: []ast.Node{
			&ast.ExpressionStatement{
				Expression: &ast.CallExpression{
					Callee: &ast.Identifier{Name: "plot"},
					Arguments: []ast.Expression{
						&ast.CallExpression{
							Callee: &ast.Identifier{Name: "nz"},
							Arguments: []ast.Expression{
								&ast.Identifier{Name: "close"},
								&ast.Literal{Value: float64(0)},
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

	if len(hoistable) != 0 {
		t.Errorf("Expected 0 hoistable calls (nz without TA args), got %d", len(hoistable))
	}
}

func TestInlineExpressionScanner_TernaryWithTABranches(t *testing.T) {
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
								Callee: &ast.MemberExpression{
									Object:   &ast.Identifier{Name: "ta"},
									Property: &ast.Identifier{Name: "sma"},
								},
								Arguments: []ast.Expression{
									&ast.Identifier{Name: "close"},
									&ast.Literal{Value: float64(20)},
								},
							},
							Alternate: &ast.CallExpression{
								Callee: &ast.MemberExpression{
									Object:   &ast.Identifier{Name: "ta"},
									Property: &ast.Identifier{Name: "ema"},
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
		},
	}

	gen := newTestGenerator()
	scanner := NewInlineExpressionScanner(gen)

	hoistable := scanner.ScanProgram(program)

	if len(hoistable) != 2 {
		t.Fatalf("Expected 2 hoistable calls (ta.sma in consequent, ta.ema in alternate), got %d", len(hoistable))
	}

	funcNames := make(map[string]bool)
	for _, call := range hoistable {
		funcNames[call.FuncName] = true
	}

	if !funcNames["ta.sma"] {
		t.Error("Expected ta.sma in ternary consequent to be hoistable")
	}
	if !funcNames["ta.ema"] {
		t.Error("Expected ta.ema in ternary alternate to be hoistable")
	}
}

func TestInlineExpressionScanner_UnaryWithTA(t *testing.T) {
	program := &ast.Program{
		Body: []ast.Node{
			&ast.ExpressionStatement{
				Expression: &ast.CallExpression{
					Callee: &ast.Identifier{Name: "plot"},
					Arguments: []ast.Expression{
						&ast.UnaryExpression{
							Operator: "-",
							Argument: &ast.CallExpression{
								Callee: &ast.MemberExpression{
									Object:   &ast.Identifier{Name: "ta"},
									Property: &ast.Identifier{Name: "rsi"},
								},
								Arguments: []ast.Expression{
									&ast.Identifier{Name: "close"},
									&ast.Literal{Value: float64(14)},
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

	if len(hoistable) != 1 {
		t.Fatalf("Expected 1 hoistable call (ta.rsi in unary), got %d", len(hoistable))
	}

	if hoistable[0].FuncName != "ta.rsi" {
		t.Errorf("Expected ta.rsi, got %s", hoistable[0].FuncName)
	}
}

func TestInlineExpressionScanner_MultiLayerNesting(t *testing.T) {
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
									Callee: &ast.Identifier{Name: "fixnan"},
									Arguments: []ast.Expression{
										&ast.CallExpression{
											Callee: &ast.MemberExpression{
												Object:   &ast.Identifier{Name: "ta"},
												Property: &ast.Identifier{Name: "sma"},
											},
											Arguments: []ast.Expression{
												&ast.Identifier{Name: "close"},
												&ast.Literal{Value: float64(50)},
											},
										},
									},
								},
								&ast.Literal{Value: float64(0)},
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
		t.Fatalf("Expected 2 hoistable calls (ta.sma + fixnan), got %d", len(hoistable))
	}

	/* Bottom-up ordering: ta.sma hoisted before fixnan (dependency first) */
	if hoistable[0].FuncName != "ta.sma" {
		t.Errorf("Expected hoistable[0] = ta.sma, got %s", hoistable[0].FuncName)
	}
	if hoistable[1].FuncName != "fixnan" {
		t.Errorf("Expected hoistable[1] = fixnan, got %s", hoistable[1].FuncName)
	}
}

func TestInlineExpressionScanner_TAInIfStatement(t *testing.T) {
	program := &ast.Program{
		Body: []ast.Node{
			&ast.IfStatement{
				Test: &ast.BinaryExpression{
					Operator: ">",
					Left:     &ast.Identifier{Name: "volume"},
					Right:    &ast.Literal{Value: float64(1000)},
				},
				Consequent: []ast.Node{
					&ast.ExpressionStatement{
						Expression: &ast.CallExpression{
							Callee: &ast.Identifier{Name: "plot"},
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

	if len(hoistable) != 1 {
		t.Fatalf("Expected 1 hoistable call (ta.rsi in if body), got %d", len(hoistable))
	}

	if hoistable[0].FuncName != "ta.rsi" {
		t.Errorf("Expected ta.rsi, got %s", hoistable[0].FuncName)
	}
}

/* TestInlineExpressionScanner_TAInLoopBodies validates TA call detection inside all loop container types */
func TestInlineExpressionScanner_TAInLoopBodies(t *testing.T) {
	taCallInPlot := func(taObj, taMethod, source string, period float64) *ast.ExpressionStatement {
		return &ast.ExpressionStatement{
			Expression: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "plot"},
				Arguments: []ast.Expression{
					&ast.CallExpression{
						Callee: &ast.MemberExpression{
							Object:   &ast.Identifier{Name: taObj},
							Property: &ast.Identifier{Name: taMethod},
						},
						Arguments: []ast.Expression{
							&ast.Identifier{Name: source},
							&ast.Literal{Value: period},
						},
					},
				},
			},
		}
	}

	tests := []struct {
		name         string
		body         []ast.Node
		expectedFunc string
	}{
		{
			name: "ForStatement body",
			body: []ast.Node{
				&ast.ForStatement{
					Counter: "i",
					From:    &ast.Literal{Value: float64(0)},
					To:      &ast.Literal{Value: float64(10)},
					Body:    []ast.Node{taCallInPlot("ta", "ema", "high", 9)},
				},
			},
			expectedFunc: "ta.ema",
		},
		{
			name: "ForInStatement body",
			body: []ast.Node{
				&ast.ForInStatement{
					ElementVar: "val",
					Collection: &ast.Identifier{Name: "arr"},
					Body:       []ast.Node{taCallInPlot("ta", "sma", "close", 14)},
				},
			},
			expectedFunc: "ta.sma",
		},
		{
			name: "nested for-in inside for",
			body: []ast.Node{
				&ast.ForStatement{
					Counter: "i",
					From:    &ast.Literal{Value: float64(0)},
					To:      &ast.Literal{Value: float64(5)},
					Body: []ast.Node{
						&ast.ForInStatement{
							ElementVar: "v",
							Collection: &ast.Identifier{Name: "data"},
							Body:       []ast.Node{taCallInPlot("ta", "rsi", "close", 14)},
						},
					},
				},
			},
			expectedFunc: "ta.rsi",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := newTestGenerator()
			scanner := NewInlineExpressionScanner(gen)
			program := &ast.Program{Body: tt.body}

			hoistable := scanner.ScanProgram(program)

			if len(hoistable) != 1 {
				t.Fatalf("Expected 1 hoistable call, got %d", len(hoistable))
			}
			if hoistable[0].FuncName != tt.expectedFunc {
				t.Errorf("Expected %s, got %s", tt.expectedFunc, hoistable[0].FuncName)
			}
		})
	}
}

func TestInlineExpressionScanner_ArrayExpressionWithTA(t *testing.T) {
	t.Skip("ArrayExpression not supported in current AST - tuple destructuring uses ArrayPattern instead")
}

func TestInlineExpressionScanner_IgnoresUserDefinedFunctions(t *testing.T) {
	program := &ast.Program{
		Body: []ast.Node{
			&ast.ExpressionStatement{
				Expression: &ast.CallExpression{
					Callee: &ast.Identifier{Name: "plot"},
					Arguments: []ast.Expression{
						&ast.CallExpression{
							Callee: &ast.Identifier{Name: "myCustomFunc"},
							Arguments: []ast.Expression{
								&ast.Identifier{Name: "close"},
								&ast.Literal{Value: float64(14)},
							},
						},
					},
				},
			},
		},
	}

	gen := newTestGenerator()
	gen.variables["myCustomFunc"] = "function"
	scanner := NewInlineExpressionScanner(gen)

	hoistable := scanner.ScanProgram(program)

	if len(hoistable) != 0 {
		t.Errorf("Expected 0 hoistable calls (user-defined functions not hoistable), got %d", len(hoistable))
	}
}

func TestInlineExpressionScanner_HoistsTADevFunction(t *testing.T) {
	program := &ast.Program{
		Body: []ast.Node{
			&ast.IfStatement{
				Test: &ast.CallExpression{
					Callee: &ast.MemberExpression{
						Object:   &ast.Identifier{Name: "ta"},
						Property: &ast.Identifier{Name: "dev"},
					},
					Arguments: []ast.Expression{
						&ast.Identifier{Name: "close"},
						&ast.Literal{Value: float64(20)},
					},
				},
				Consequent: []ast.Node{},
			},
		},
	}

	gen := newTestGenerator()
	scanner := NewInlineExpressionScanner(gen)

	hoistable := scanner.ScanProgram(program)

	if len(hoistable) != 1 {
		t.Errorf("Expected 1 hoistable call (ta.dev has TAFunctionRegistry handler), got %d", len(hoistable))
	}
}

func TestInlineExpressionScanner_MultiplePlotStatements(t *testing.T) {
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
								&ast.Literal{Value: float64(50)},
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
							Callee: &ast.MemberExpression{
								Object:   &ast.Identifier{Name: "ta"},
								Property: &ast.Identifier{Name: "ema"},
							},
							Arguments: []ast.Expression{
								&ast.Identifier{Name: "close"},
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

	hoistable := scanner.ScanProgram(program)

	if len(hoistable) != 2 {
		t.Fatalf("Expected 2 hoistable calls (independent plot statements), got %d", len(hoistable))
	}

	funcNames := make(map[string]bool)
	for _, call := range hoistable {
		funcNames[call.FuncName] = true
	}

	if !funcNames["ta.sma"] {
		t.Error("Expected ta.sma from first plot to be hoistable")
	}
	if !funcNames["ta.ema"] {
		t.Error("Expected ta.ema from second plot to be hoistable")
	}
}

/* Validates VariableDeclaration with TA in initializer */
func TestInlineExpressionScanner_VariableDeclarationWithTA(t *testing.T) {
	program := &ast.Program{
		Body: []ast.Node{
			&ast.VariableDeclaration{
				Kind: "var",
				Declarations: []ast.VariableDeclarator{
					{
						ID: &ast.Identifier{Name: "myVar"},
						Init: &ast.BinaryExpression{
							Operator: "+",
							Left: &ast.CallExpression{
								Callee: &ast.MemberExpression{
									Object:   &ast.Identifier{Name: "ta"},
									Property: &ast.Identifier{Name: "sma"},
								},
								Arguments: []ast.Expression{
									&ast.Identifier{Name: "close"},
									&ast.Literal{Value: float64(14)},
								},
							},
							Right: &ast.Literal{Value: float64(10)},
						},
					},
				},
			},
		},
	}

	gen := newTestGenerator()
	scanner := NewInlineExpressionScanner(gen)

	hoistable := scanner.ScanProgram(program)

	if len(hoistable) != 1 {
		t.Fatalf("Expected 1 hoistable call (ta.sma nested in binary), got %d", len(hoistable))
	}

	if hoistable[0].FuncName != "ta.sma" {
		t.Errorf("Expected ta.sma, got %s", hoistable[0].FuncName)
	}
}

/* Validates VariableDeclaration with direct TA call NOT hoisted */
func TestInlineExpressionScanner_VariableDeclarationDirectTA(t *testing.T) {
	program := &ast.Program{
		Body: []ast.Node{
			&ast.VariableDeclaration{
				Kind: "var",
				Declarations: []ast.VariableDeclarator{
					{
						ID: &ast.Identifier{Name: "sma14"},
						Init: &ast.CallExpression{
							Callee: &ast.MemberExpression{
								Object:   &ast.Identifier{Name: "ta"},
								Property: &ast.Identifier{Name: "sma"},
							},
							Arguments: []ast.Expression{
								&ast.Identifier{Name: "close"},
								&ast.Literal{Value: float64(14)},
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

	if len(hoistable) != 0 {
		t.Errorf("Expected 0 hoistable calls (direct TA init), got %d", len(hoistable))
	}
}

/* Validates VariableDeclaration with runtime-only function NOT hoisted */
func TestInlineExpressionScanner_VariableDeclarationRuntimeOnly(t *testing.T) {
	program := &ast.Program{
		Body: []ast.Node{
			&ast.VariableDeclaration{
				Kind: "var",
				Declarations: []ast.VariableDeclarator{
					{
						ID: &ast.Identifier{Name: "myVar"},
						Init: &ast.BinaryExpression{
							Operator: "+",
							Left: &ast.CallExpression{
								Callee:    &ast.Identifier{Name: "fixnan"},
								Arguments: []ast.Expression{&ast.Identifier{Name: "close"}},
							},
							Right: &ast.Literal{Value: float64(10)},
						},
					},
				},
			},
		},
	}

	gen := newTestGenerator()
	scanner := NewInlineExpressionScanner(gen)

	hoistable := scanner.ScanProgram(program)

	if len(hoistable) != 0 {
		t.Errorf("Expected 0 hoistable calls (runtime-only fixnan), got %d", len(hoistable))
	}
}

/* Validates VariableDeclaration with math.abs containing TA */
func TestInlineExpressionScanner_VariableDeclarationMathWithTA(t *testing.T) {
	program := &ast.Program{
		Body: []ast.Node{
			&ast.VariableDeclaration{
				Kind: "var",
				Declarations: []ast.VariableDeclarator{
					{
						ID: &ast.Identifier{Name: "absEMA"},
						Init: &ast.CallExpression{
							Callee: &ast.MemberExpression{
								Object:   &ast.Identifier{Name: "math"},
								Property: &ast.Identifier{Name: "abs"},
							},
							Arguments: []ast.Expression{
								&ast.CallExpression{
									Callee: &ast.MemberExpression{
										Object:   &ast.Identifier{Name: "ta"},
										Property: &ast.Identifier{Name: "ema"},
									},
									Arguments: []ast.Expression{
										&ast.Identifier{Name: "close"},
										&ast.Literal{Value: float64(9)},
									},
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

	if len(hoistable) != 1 {
		t.Fatalf("Expected 1 hoistable call (ta.ema), got %d", len(hoistable))
	}

	if hoistable[0].FuncName != "ta.ema" {
		t.Errorf("Expected ta.ema, got %s", hoistable[0].FuncName)
	}
}

/* Validates VariableDeclaration with multiple TA calls */
func TestInlineExpressionScanner_VariableDeclarationMultipleTA(t *testing.T) {
	program := &ast.Program{
		Body: []ast.Node{
			&ast.VariableDeclaration{
				Kind: "var",
				Declarations: []ast.VariableDeclarator{
					{
						ID: &ast.Identifier{Name: "combo"},
						Init: &ast.BinaryExpression{
							Operator: "+",
							Left: &ast.CallExpression{
								Callee: &ast.MemberExpression{
									Object:   &ast.Identifier{Name: "ta"},
									Property: &ast.Identifier{Name: "sma"},
								},
								Arguments: []ast.Expression{
									&ast.Identifier{Name: "close"},
									&ast.Literal{Value: float64(14)},
								},
							},
							Right: &ast.CallExpression{
								Callee: &ast.MemberExpression{
									Object:   &ast.Identifier{Name: "ta"},
									Property: &ast.Identifier{Name: "ema"},
								},
								Arguments: []ast.Expression{
									&ast.Identifier{Name: "close"},
									&ast.Literal{Value: float64(9)},
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
		t.Fatalf("Expected 2 hoistable calls (ta.sma + ta.ema), got %d", len(hoistable))
	}

	funcNames := make(map[string]bool)
	for _, call := range hoistable {
		funcNames[call.FuncName] = true
	}

	if !funcNames["ta.sma"] {
		t.Error("Expected ta.sma to be hoistable")
	}
	if !funcNames["ta.ema"] {
		t.Error("Expected ta.ema to be hoistable")
	}
}

/* Validates mixed VariableDeclaration and ExpressionStatement */
func TestInlineExpressionScanner_MixedDeclarationAndExpression(t *testing.T) {
	program := &ast.Program{
		Body: []ast.Node{
			&ast.VariableDeclaration{
				Kind: "var",
				Declarations: []ast.VariableDeclarator{
					{
						ID: &ast.Identifier{Name: "sma14"},
						Init: &ast.BinaryExpression{
							Operator: "*",
							Left: &ast.CallExpression{
								Callee: &ast.MemberExpression{
									Object:   &ast.Identifier{Name: "ta"},
									Property: &ast.Identifier{Name: "sma"},
								},
								Arguments: []ast.Expression{
									&ast.Identifier{Name: "close"},
									&ast.Literal{Value: float64(14)},
								},
							},
							Right: &ast.Literal{Value: float64(2)},
						},
					},
				},
			},
			&ast.ExpressionStatement{
				Expression: &ast.CallExpression{
					Callee: &ast.Identifier{Name: "plot"},
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
					},
				},
			},
		},
	}

	gen := newTestGenerator()
	scanner := NewInlineExpressionScanner(gen)

	hoistable := scanner.ScanProgram(program)

	if len(hoistable) != 2 {
		t.Fatalf("Expected 2 hoistable calls (ta.sma from var + ta.rsi from plot), got %d", len(hoistable))
	}

	funcNames := make(map[string]bool)
	for _, call := range hoistable {
		funcNames[call.FuncName] = true
	}

	if !funcNames["ta.sma"] {
		t.Error("Expected ta.sma from variable declaration to be hoistable")
	}
	if !funcNames["ta.rsi"] {
		t.Error("Expected ta.rsi from plot expression to be hoistable")
	}
}

/* Validates IfStatement conditional with TA call registration */
func TestInlineExpressionScanner_IfStatementConditionalRegistration(t *testing.T) {
	program := &ast.Program{
		Body: []ast.Node{
			&ast.IfStatement{
				Test: &ast.ConditionalExpression{
					Test: &ast.BinaryExpression{
						Operator: ">",
						Left:     &ast.Identifier{Name: "close"},
						Right:    &ast.Identifier{Name: "open"},
					},
					Consequent: &ast.Literal{Value: float64(100)},
					Alternate:  &ast.Literal{Value: float64(50)},
				},
				Consequent: []ast.Node{
					&ast.ExpressionStatement{
						Expression: &ast.CallExpression{
							Callee:    &ast.Identifier{Name: "plot"},
							Arguments: []ast.Expression{&ast.Identifier{Name: "close"}},
						},
					},
				},
			},
		},
	}

	gen := newTestGenerator()
	scanner := NewInlineExpressionScanner(gen)

	hoistable := scanner.ScanProgram(program)

	if len(hoistable) != 0 {
		t.Errorf("Expected 0 hoistable calls (no TA in conditional), got %d", len(hoistable))
	}
}

/* Validates VariableDeclaration with ternary containing TA */
func TestInlineExpressionScanner_VariableDeclarationWithTernaryTA(t *testing.T) {
	program := &ast.Program{
		Body: []ast.Node{
			&ast.VariableDeclaration{
				Kind: "var",
				Declarations: []ast.VariableDeclarator{
					{
						ID: &ast.Identifier{Name: "adaptiveMA"},
						Init: &ast.ConditionalExpression{
							Test: &ast.BinaryExpression{
								Operator: ">",
								Left:     &ast.Identifier{Name: "close"},
								Right:    &ast.Identifier{Name: "open"},
							},
							Consequent: &ast.CallExpression{
								Callee: &ast.MemberExpression{
									Object:   &ast.Identifier{Name: "ta"},
									Property: &ast.Identifier{Name: "sma"},
								},
								Arguments: []ast.Expression{
									&ast.Identifier{Name: "close"},
									&ast.Literal{Value: float64(10)},
								},
							},
							Alternate: &ast.CallExpression{
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
		t.Fatalf("Expected 2 hoistable calls (ta.sma + ta.ema in ternary), got %d", len(hoistable))
	}

	funcNames := make(map[string]bool)
	for _, call := range hoistable {
		funcNames[call.FuncName] = true
	}

	if !funcNames["ta.sma"] {
		t.Error("Expected ta.sma in ternary consequent to be hoistable")
	}
	if !funcNames["ta.ema"] {
		t.Error("Expected ta.ema in ternary alternate to be hoistable")
	}
}

/* Validates StmtIndex tracks top-level statement index through all container types */
func TestInlineExpressionScanner_StatementIndexing(t *testing.T) {
	makeTACall := func(fnName string) *ast.CallExpression {
		parts := splitDot(fnName)
		return &ast.CallExpression{
			Callee: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: parts[0]},
				Property: &ast.Identifier{Name: parts[1]},
			},
			Arguments: []ast.Expression{
				&ast.Identifier{Name: "close"},
				&ast.Literal{Value: float64(14)},
			},
		}
	}

	wrapInPlot := func(inner ast.Expression) *ast.ExpressionStatement {
		return &ast.ExpressionStatement{
			Expression: &ast.CallExpression{
				Callee:    &ast.Identifier{Name: "plot"},
				Arguments: []ast.Expression{inner},
			},
		}
	}

	spacer := func() *ast.VariableDeclaration {
		return &ast.VariableDeclaration{
			Declarations: []ast.VariableDeclarator{
				{ID: &ast.Identifier{Name: "spacer"}, Init: &ast.Literal{Value: 1}},
			},
		}
	}

	tests := []struct {
		name            string
		body            []ast.Node
		expectedCount   int
		expectedIndices []int // one per hoistable, in order
	}{
		{
			name:            "single expression at index 0",
			body:            []ast.Node{wrapInPlot(makeTACall("ta.sma"))},
			expectedCount:   1,
			expectedIndices: []int{0},
		},
		{
			name: "preceded by non-TA statements",
			body: []ast.Node{
				spacer(), spacer(), wrapInPlot(makeTACall("ta.sma")),
			},
			expectedCount:   1,
			expectedIndices: []int{2},
		},
		{
			name: "multiple TA in separate statements get sequential indices",
			body: []ast.Node{
				spacer(),
				wrapInPlot(makeTACall("ta.sma")),
				wrapInPlot(makeTACall("ta.ema")),
			},
			expectedCount:   2,
			expectedIndices: []int{1, 2},
		},
		{
			name: "TA nested in if-consequent inherits enclosing index",
			body: []ast.Node{
				spacer(),
				&ast.IfStatement{
					Test:       &ast.Literal{Value: true},
					Consequent: []ast.Node{wrapInPlot(makeTACall("ta.sma"))},
				},
			},
			expectedCount:   1,
			expectedIndices: []int{1},
		},
		{
			name: "TA nested in if-alternate inherits enclosing index",
			body: []ast.Node{
				spacer(), spacer(),
				&ast.IfStatement{
					Test:       &ast.Literal{Value: true},
					Consequent: []ast.Node{},
					Alternate:  []ast.Node{wrapInPlot(makeTACall("ta.sma"))},
				},
			},
			expectedCount:   1,
			expectedIndices: []int{2},
		},
		{
			name: "TA in if-condition gets the if-statement index",
			body: []ast.Node{
				spacer(),
				&ast.IfStatement{
					Test: &ast.BinaryExpression{
						Left: makeTACall("ta.sma"), Operator: ">",
						Right: &ast.Literal{Value: float64(0)},
					},
					Consequent: []ast.Node{},
				},
			},
			expectedCount:   1,
			expectedIndices: []int{1},
		},
		{
			name: "TA nested in for-body inherits enclosing index",
			body: []ast.Node{
				spacer(), spacer(),
				&ast.ForStatement{
					Counter: "i",
					From:    &ast.Literal{Value: float64(0)},
					To:      &ast.Literal{Value: float64(10)},
					Body:    []ast.Node{wrapInPlot(makeTACall("ta.sma"))},
				},
			},
			expectedCount:   1,
			expectedIndices: []int{2},
		},
		{
			name: "TA nested in for-in body inherits enclosing index",
			body: []ast.Node{
				&ast.ForInStatement{
					ElementVar: "val",
					Collection: &ast.Identifier{Name: "arr"},
					Body:       []ast.Node{wrapInPlot(makeTACall("ta.sma"))},
				},
			},
			expectedCount:   1,
			expectedIndices: []int{0},
		},
		{
			name: "TA nested in while-body inherits enclosing index",
			body: []ast.Node{
				spacer(),
				&ast.WhileStatement{
					Condition: &ast.Literal{Value: true},
					Body:      []ast.Node{wrapInPlot(makeTACall("ta.sma"))},
				},
			},
			expectedCount:   1,
			expectedIndices: []int{1},
		},
		{
			name: "TA in while-condition gets the while-statement index",
			body: []ast.Node{
				spacer(), spacer(),
				&ast.WhileStatement{
					Condition: &ast.BinaryExpression{
						Left: makeTACall("ta.sma"), Operator: ">",
						Right: &ast.Literal{Value: float64(0)},
					},
					Body: []ast.Node{},
				},
			},
			expectedCount:   1,
			expectedIndices: []int{2},
		},
		{
			name: "deeply nested: for inside if still inherits top-level index",
			body: []ast.Node{
				spacer(),
				&ast.IfStatement{
					Test: &ast.Literal{Value: true},
					Consequent: []ast.Node{
						&ast.ForStatement{
							Counter: "j",
							From:    &ast.Literal{Value: float64(0)},
							To:      &ast.Literal{Value: float64(5)},
							Body:    []ast.Node{wrapInPlot(makeTACall("ta.sma"))},
						},
					},
				},
			},
			expectedCount:   1,
			expectedIndices: []int{1},
		},
		{
			name: "multiple TA in same statement share same index",
			body: []ast.Node{
				spacer(),
				&ast.ExpressionStatement{
					Expression: &ast.CallExpression{
						Callee: &ast.Identifier{Name: "plot"},
						Arguments: []ast.Expression{
							&ast.BinaryExpression{
								Operator: "+",
								Left:     makeTACall("ta.sma"),
								Right:    makeTACall("ta.ema"),
							},
						},
					},
				},
			},
			expectedCount:   2,
			expectedIndices: []int{1, 1},
		},
		{
			name: "variable declaration init carries enclosing statement index",
			body: []ast.Node{
				spacer(), spacer(),
				&ast.VariableDeclaration{
					Kind: "var",
					Declarations: []ast.VariableDeclarator{
						{
							ID: &ast.Identifier{Name: "z"},
							Init: &ast.BinaryExpression{
								Operator: "+",
								Left:     makeTACall("ta.sma"),
								Right:    &ast.Literal{Value: float64(1)},
							},
						},
					},
				},
			},
			expectedCount:   1,
			expectedIndices: []int{2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := newTestGenerator()
			scanner := NewInlineExpressionScanner(gen)
			program := &ast.Program{Body: tt.body}

			hoistable := scanner.ScanProgram(program)

			if len(hoistable) != tt.expectedCount {
				t.Fatalf("got %d hoistable calls, want %d", len(hoistable), tt.expectedCount)
			}
			for i, want := range tt.expectedIndices {
				if hoistable[i].StmtIndex != want {
					t.Errorf("hoistable[%d].StmtIndex = %d, want %d (func=%s)",
						i, hoistable[i].StmtIndex, want, hoistable[i].FuncName)
				}
			}
		})
	}
}

// splitDot splits "a.b" into [2]string{"a","b"}.
func splitDot(s string) [2]string {
	for i, c := range s {
		if c == '.' {
			return [2]string{s[:i], s[i+1:]}
		}
	}
	return [2]string{s, ""}
}
