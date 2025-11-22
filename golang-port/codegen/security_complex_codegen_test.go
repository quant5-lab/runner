package codegen

import (
	"strings"
	"testing"

	"github.com/borisquantlab/pinescript-go/ast"
)

/* TestSecurityBinaryExpression tests code generation for binary operations in security() context
 * This validates the hybrid architecture: inline TA leaves + runtime composition
 */
func TestSecurityBinaryExpression(t *testing.T) {
	tests := []struct {
		name           string
		expression     ast.Expression
		expectedCode   []string // Must contain all these substrings
		unexpectedCode []string // Must NOT contain these substrings
	}{
		{
			name: "SMA + EMA addition",
			expression: &ast.BinaryExpression{
				Operator: "+",
				Left: &ast.CallExpression{
					Callee: &ast.MemberExpression{
						Object:   &ast.Identifier{Name: "ta"},
						Property: &ast.Identifier{Name: "sma"},
					},
					Arguments: []ast.Expression{
						&ast.Identifier{Name: "close"},
						&ast.Literal{Value: float64(20)},
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
			expectedCode: []string{
				"Inline SMA(20)",           // Left operand inlined
				"Inline EMA(10)",           // Right operand inlined
				"origCtx := ctx",           // Context switching
				"ctx = secCtx",             // Security context assignment
				"ctx.BarIndex = secBarIdx", // Bar index set
				"ctx = origCtx",            // Context restored
				"Series.GetCurrent() +",    // Binary operation composition
			},
			unexpectedCode: []string{
				"cache.GetExpression", // Should NOT use old expression cache
				"[]float64",           // Should NOT allocate arrays
			},
		},
		{
			name: "SMA * constant multiplication",
			expression: &ast.BinaryExpression{
				Operator: "*",
				Left: &ast.CallExpression{
					Callee: &ast.MemberExpression{
						Object:   &ast.Identifier{Name: "ta"},
						Property: &ast.Identifier{Name: "sma"},
					},
					Arguments: []ast.Expression{
						&ast.Identifier{Name: "close"},
						&ast.Literal{Value: float64(20)},
					},
				},
				Right: &ast.Literal{Value: float64(2.0)},
			},
			expectedCode: []string{
				"Inline SMA(20)",
				"origCtx := ctx",
				"secTmp_test_val_leftSeries := series.NewSeries(1000)",                               // Temp series for left operand
				"secTmp_test_val_rightSeries := series.NewSeries(1000)",                              // Temp series for right operand
				"secTmp_test_val_rightSeries.Set(2.00)",                                              // Literal value
				"secTmp_test_val_leftSeries.GetCurrent() * secTmp_test_val_rightSeries.GetCurrent()", // Composition
			},
		},
		{
			name: "Identifier subtraction (high - low)",
			expression: &ast.BinaryExpression{
				Operator: "-",
				Left:     &ast.Identifier{Name: "high"},
				Right:    &ast.Identifier{Name: "low"},
			},
			expectedCode: []string{
				"ctx.Data[ctx.BarIndex].High", // Direct field access in security context
				"ctx.Data[ctx.BarIndex].Low",
				"secTmp_test_val_leftSeries := series.NewSeries(1000)", // Temp series for composition
				"secTmp_test_val_rightSeries := series.NewSeries(1000)",
				"secTmp_test_val_leftSeries.GetCurrent() - secTmp_test_val_rightSeries.GetCurrent()",
			},
			unexpectedCode: []string{
				"Inline SMA", // Should NOT inline for simple identifiers
			},
		},
		{
			name: "Division (close / open) for returns",
			expression: &ast.BinaryExpression{
				Operator: "/",
				Left:     &ast.Identifier{Name: "close"},
				Right:    &ast.Identifier{Name: "open"},
			},
			expectedCode: []string{
				"ctx.Data[ctx.BarIndex].Close",
				"ctx.Data[ctx.BarIndex].Open",
				"secTmp_test_val_leftSeries := series.NewSeries(1000)",
				"secTmp_test_val_rightSeries := series.NewSeries(1000)",
				"secTmp_test_val_leftSeries.GetCurrent() / secTmp_test_val_rightSeries.GetCurrent()",
			},
		},
		{
			name: "Nested binary: (SMA - EMA) / SMA",
			expression: &ast.BinaryExpression{
				Operator: "/",
				Left: &ast.BinaryExpression{
					Operator: "-",
					Left: &ast.CallExpression{
						Callee: &ast.MemberExpression{
							Object:   &ast.Identifier{Name: "ta"},
							Property: &ast.Identifier{Name: "sma"},
						},
						Arguments: []ast.Expression{
							&ast.Identifier{Name: "close"},
							&ast.Literal{Value: float64(20)},
						},
					},
					Right: &ast.CallExpression{
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
				Right: &ast.CallExpression{
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
			expectedCode: []string{
				"Inline SMA(20)",
				"Inline EMA(20)",
				"secTmp_test_val_leftSeries := series.NewSeries(1000)",       // Outer left operand
				"secTmp_test_val_rightSeries := series.NewSeries(1000)",      // Outer right operand
				"secTmp_test_val_left_leftSeries := series.NewSeries(1000)",  // Nested: (SMA - EMA) left
				"secTmp_test_val_left_rightSeries := series.NewSeries(1000)", // Nested: (SMA - EMA) right
			},
		},
		{
			name: "STDEV * multiplier (BB deviation pattern)",
			expression: &ast.BinaryExpression{
				Operator: "*",
				Left: &ast.CallExpression{
					Callee: &ast.MemberExpression{
						Object:   &ast.Identifier{Name: "ta"},
						Property: &ast.Identifier{Name: "stdev"},
					},
					Arguments: []ast.Expression{
						&ast.Identifier{Name: "close"},
						&ast.Literal{Value: float64(20)},
					},
				},
				Right: &ast.Literal{Value: float64(2.0)},
			},
			expectedCode: []string{
				"Inline STDEV(20)",
				"math.Sqrt(variance)",
				"secTmp_test_val_leftSeries := series.NewSeries(1000)",  // Temp series for STDEV
				"secTmp_test_val_rightSeries := series.NewSeries(1000)", // Temp series for multiplier
				"secTmp_test_val_leftSeries.GetCurrent() * secTmp_test_val_rightSeries.GetCurrent()",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			/* Build minimal program with security() call containing expression */
			program := &ast.Program{
				Body: []ast.Node{
					&ast.ExpressionStatement{
						Expression: &ast.CallExpression{
							Callee: &ast.Identifier{Name: "strategy"},
							Arguments: []ast.Expression{
								&ast.Literal{Value: "Test"},
							},
						},
					},
					&ast.VariableDeclaration{
						Declarations: []ast.VariableDeclarator{
							{
								ID: ast.Identifier{Name: "test_val"},
								Init: &ast.CallExpression{
									Callee: &ast.Identifier{Name: "security"},
									Arguments: []ast.Expression{
										&ast.Literal{Value: "BTCUSD"},
										&ast.Literal{Value: "1D"},
										tt.expression,
									},
								},
							},
						},
					},
				},
			}

			/* Generate code */
			generated, err := GenerateStrategyCodeFromAST(program)
			if err != nil {
				t.Fatalf("Code generation failed: %v", err)
			}
			code := generated.FunctionBody

			/* Verify expected patterns present */
			for _, expected := range tt.expectedCode {
				if !strings.Contains(code, expected) {
					t.Errorf("Expected code to contain %q\nGenerated code:\n%s", expected, code)
				}
			}

			/* Verify unexpected patterns absent */
			for _, unexpected := range tt.unexpectedCode {
				if strings.Contains(code, unexpected) {
					t.Errorf("Expected code NOT to contain %q\nGenerated code:\n%s", unexpected, code)
				}
			}

			/* Verify no placeholder/error markers */
			if strings.Contains(code, "TODO") {
				t.Errorf("Generated code contains TODO markers:\n%s", code)
			}
			if strings.Contains(code, "math.NaN() //") {
				t.Errorf("Generated code contains error NaN markers:\n%s", code)
			}
		})
	}
}

/* TestSecurityConditionalExpression tests ternary expressions in security() context */
func TestSecurityConditionalExpression(t *testing.T) {
	/* Ternary: close > open ? close : open */
	expression := &ast.ConditionalExpression{
		Test: &ast.BinaryExpression{
			Operator: ">",
			Left:     &ast.Identifier{Name: "close"},
			Right:    &ast.Identifier{Name: "open"},
		},
		Consequent: &ast.Identifier{Name: "close"},
		Alternate:  &ast.Identifier{Name: "open"},
	}

	program := &ast.Program{
		Body: []ast.Node{
			&ast.ExpressionStatement{
				Expression: &ast.CallExpression{
					Callee: &ast.Identifier{Name: "strategy"},
					Arguments: []ast.Expression{
						&ast.Literal{Value: "Test"},
					},
				},
			},
			&ast.VariableDeclaration{
				Declarations: []ast.VariableDeclarator{
					{
						ID: ast.Identifier{Name: "test_val"},
						Init: &ast.CallExpression{
							Callee: &ast.Identifier{Name: "security"},
							Arguments: []ast.Expression{
								&ast.Literal{Value: "BTCUSD"},
								&ast.Literal{Value: "1D"},
								expression,
							},
						},
					},
				},
			},
		},
	}

	generated, err := GenerateStrategyCodeFromAST(program)
	if err != nil {
		t.Fatalf("Code generation failed: %v", err)
	}
	code := generated.FunctionBody

	/* Verify conditional code generation */
	expectedPatterns := []string{
		"origCtx := ctx",
		"ctx = secCtx",
		"if",                       // Conditional present
		"} else",                   // Both branches present
		"closeSeries.GetCurrent()", // Uses existing series (not inline identifiers in conditionals yet)
		"openSeries.GetCurrent()",
	}

	for _, pattern := range expectedPatterns {
		if !strings.Contains(code, pattern) {
			t.Errorf("Expected code to contain %q\nGenerated code:\n%s", pattern, code)
		}
	}
}

/* TestSecurityATRGeneration validates ATR inline implementation edge cases */
func TestSecurityATRGeneration(t *testing.T) {
	expression := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "ta"},
			Property: &ast.Identifier{Name: "atr"},
		},
		Arguments: []ast.Expression{
			&ast.Literal{Value: float64(14)},
		},
	}

	program := &ast.Program{
		Body: []ast.Node{
			&ast.ExpressionStatement{
				Expression: &ast.CallExpression{
					Callee: &ast.Identifier{Name: "strategy"},
					Arguments: []ast.Expression{
						&ast.Literal{Value: "Test"},
					},
				},
			},
			&ast.VariableDeclaration{
				Declarations: []ast.VariableDeclarator{
					{
						ID: ast.Identifier{Name: "atr_val"},
						Init: &ast.CallExpression{
							Callee: &ast.Identifier{Name: "security"},
							Arguments: []ast.Expression{
								&ast.Literal{Value: "BTCUSD"},
								&ast.Literal{Value: "1D"},
								expression,
							},
						},
					},
				},
			},
		},
	}

	generated, err := GenerateStrategyCodeFromAST(program)
	if err != nil {
		t.Fatalf("Code generation failed: %v", err)
	}
	code := generated.FunctionBody

	/* Verify ATR-specific patterns */
	expectedPatterns := []string{
		"Inline ATR(14)",
		"ctx.Data[ctx.BarIndex].High",
		"ctx.Data[ctx.BarIndex].Low",
		"ctx.Data[ctx.BarIndex-1].Close",       // Previous close for TR
		"tr := math.Max(hl, math.Max(hc, lc))", // True Range calculation
		"alpha := 1.0 / 14",                    // RMA smoothing
		"prevATR :=",                           // RMA uses previous value
	}

	for _, pattern := range expectedPatterns {
		if !strings.Contains(code, pattern) {
			t.Errorf("Expected ATR code to contain %q\nGenerated code:\n%s", pattern, code)
		}
	}

	/* Verify warmup handling */
	if !strings.Contains(code, "if ctx.BarIndex < 1") {
		t.Error("Expected warmup check for first bar (need previous close)")
	}
	if !strings.Contains(code, "if ctx.BarIndex < 14") {
		t.Error("Expected warmup check for ATR period")
	}
}

/* TestSecuritySTDEVGeneration validates STDEV inline implementation */
func TestSecuritySTDEVGeneration(t *testing.T) {
	expression := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "ta"},
			Property: &ast.Identifier{Name: "stdev"},
		},
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "close"},
			&ast.Literal{Value: float64(20)},
		},
	}

	program := &ast.Program{
		Body: []ast.Node{
			&ast.ExpressionStatement{
				Expression: &ast.CallExpression{
					Callee: &ast.Identifier{Name: "strategy"},
					Arguments: []ast.Expression{
						&ast.Literal{Value: "Test"},
					},
				},
			},
			&ast.VariableDeclaration{
				Declarations: []ast.VariableDeclarator{
					{
						ID: ast.Identifier{Name: "stdev_val"},
						Init: &ast.CallExpression{
							Callee: &ast.Identifier{Name: "security"},
							Arguments: []ast.Expression{
								&ast.Literal{Value: "BTCUSD"},
								&ast.Literal{Value: "1D"},
								expression,
							},
						},
					},
				},
			},
		},
	}

	generated, err := GenerateStrategyCodeFromAST(program)
	if err != nil {
		t.Fatalf("Code generation failed: %v", err)
	}
	code := generated.FunctionBody

	/* Verify STDEV algorithm steps */
	expectedPatterns := []string{
		"Inline STDEV(20)",
		"sum := 0.0",                        // Mean calculation
		"mean := sum / 20.0",                // Mean result
		"variance := 0.0",                   // Variance calculation
		"diff := closeSeries.Get(j) - mean", // Uses closeSeries.Get() with relative offset
		"variance += diff * diff",           // Squared deviation
		"math.Sqrt(variance)",               // Final STDEV
	}

	for _, pattern := range expectedPatterns {
		if !strings.Contains(code, pattern) {
			t.Errorf("Expected STDEV code to contain %q\nGenerated code:\n%s", pattern, code)
		}
	}
}

/* TestSecurityContextIsolation verifies context switching safety */
func TestSecurityContextIsolation(t *testing.T) {
	/* Multiple security() calls with different timeframes and complex expressions */
	program := &ast.Program{
		Body: []ast.Node{
			&ast.ExpressionStatement{
				Expression: &ast.CallExpression{
					Callee: &ast.Identifier{Name: "strategy"},
					Arguments: []ast.Expression{
						&ast.Literal{Value: "Test"},
					},
				},
			},
			&ast.VariableDeclaration{
				Declarations: []ast.VariableDeclarator{
					{
						ID: ast.Identifier{Name: "daily"},
						Init: &ast.CallExpression{
							Callee: &ast.Identifier{Name: "security"},
							Arguments: []ast.Expression{
								&ast.Literal{Value: "BTCUSD"},
								&ast.Literal{Value: "1D"},
								&ast.BinaryExpression{
									Operator: "+",
									Left:     &ast.Identifier{Name: "close"},
									Right:    &ast.Identifier{Name: "open"},
								},
							},
						},
					},
				},
			},
			&ast.VariableDeclaration{
				Declarations: []ast.VariableDeclarator{
					{
						ID: ast.Identifier{Name: "weekly"},
						Init: &ast.CallExpression{
							Callee: &ast.Identifier{Name: "security"},
							Arguments: []ast.Expression{
								&ast.Literal{Value: "BTCUSD"},
								&ast.Literal{Value: "1W"},
								&ast.BinaryExpression{
									Operator: "*",
									Left:     &ast.Identifier{Name: "high"},
									Right:    &ast.Literal{Value: float64(2.0)},
								},
							},
						},
					},
				},
			},
		},
	}

	generated, err := GenerateStrategyCodeFromAST(program)
	if err != nil {
		t.Fatalf("Code generation failed: %v", err)
	}
	code := generated.FunctionBody

	/* Count context switches - should have 2 (one per security call) */
	origCtxCount := strings.Count(code, "origCtx := ctx")
	if origCtxCount != 2 {
		t.Errorf("Expected 2 context switches, found %d", origCtxCount)
	}

	/* Verify context restoration after each call */
	restoreCount := strings.Count(code, "ctx = origCtx")
	if restoreCount != 2 {
		t.Errorf("Expected 2 context restorations, found %d", restoreCount)
	}

	/* Verify no variable collisions */
	if strings.Contains(code, "secTimeframeSeconds :=") {
		t.Error("Found := declaration for secTimeframeSeconds (should use = for reuse)")
	}

	/* Verify both BinaryExpressions generated temp series */
	if !strings.Contains(code, "secTmp_dailySeries := series.NewSeries(1000)") {
		t.Error("Expected temp series for daily security call")
	}
	if !strings.Contains(code, "secTmp_weeklySeries := series.NewSeries(1000)") {
		t.Error("Expected temp series for weekly security call")
	}
}
