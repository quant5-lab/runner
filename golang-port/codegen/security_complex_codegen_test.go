package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestSecurityBinaryExpression(t *testing.T) {
	tests := []struct {
		name       string
		expression ast.Expression
		expect     []string
		reject     []string
	}{
		{
			name:       "SMA + EMA addition",
			expression: BinaryExpr("+", TACall("sma", Ident("close"), 20), TACall("ema", Ident("close"), 10)),
			expect: []string{
				"secBarEvaluator.EvaluateAtBar",
				"&ast.BinaryExpression{Operator: \"+\"",
				"secBarIdx",
			},
			reject: []string{
				"origCtx := ctx",
				"secTmp_test_val_leftSeries",
				"secTmp_test_val_rightSeries",
			},
		},
		{
			name:       "SMA * constant multiplication",
			expression: BinaryExpr("*", TACall("sma", Ident("close"), 20), Lit(2.0)),
			expect: []string{
				"secBarEvaluator.EvaluateAtBar",
				"&ast.BinaryExpression{Operator: \"*\"",
				"&ast.Literal{Value: 2.0}",
			},
			reject: []string{
				"secTmp_test_val_leftSeries",
				"secTmp_test_val_rightSeries",
			},
		},
		{
			name:       "Identifier subtraction (high - low)",
			expression: BinaryExpr("-", Ident("high"), Ident("low")),
			expect: []string{
				"secBarEvaluator.EvaluateAtBar",
				"&ast.BinaryExpression{Operator: \"-\"",
				"&ast.Identifier{Name: \"high\"}",
				"&ast.Identifier{Name: \"low\"}",
			},
			reject: []string{
				"secTmp_test_val_leftSeries",
				"secTmp_test_val_rightSeries",
			},
		},
		{
			name:       "Division (close / open) for returns",
			expression: BinaryExpr("/", Ident("close"), Ident("open")),
			expect: []string{
				"secBarEvaluator.EvaluateAtBar",
				"&ast.BinaryExpression{Operator: \"/\"",
				"&ast.Identifier{Name: \"close\"}",
				"&ast.Identifier{Name: \"open\"}",
			},
			reject: []string{
				"secTmp_test_val_leftSeries",
			},
		},
		{
			name: "Nested binary: (SMA - EMA) / SMA",
			expression: BinaryExpr("/",
				BinaryExpr("-", TACall("sma", Ident("close"), 20), TACall("ema", Ident("close"), 20)),
				TACall("sma", Ident("close"), 20),
			),
			expect: []string{
				"secBarEvaluator.EvaluateAtBar",
				"&ast.BinaryExpression{Operator: \"/\"",
				"Left: &ast.BinaryExpression{Operator: \"-\"",
			},
			reject: []string{
				"secTmp_test_val_leftSeries",
				"secTmp_test_val_left_leftSeries",
			},
		},
		{
			name:       "STDEV * multiplier (BB deviation pattern)",
			expression: BinaryExpr("*", TACall("stdev", Ident("close"), 20), Lit(2.0)),
			expect: []string{
				"secBarEvaluator.EvaluateAtBar",
				"&ast.BinaryExpression{Operator: \"*\"",
				"&ast.CallExpression{Callee: &ast.MemberExpression",
				"&ast.Literal{Value: 2.0}",
			},
			reject: []string{
				"secTmp_test_val_leftSeries",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := generateSecurityExpression(t, "test_val", tt.expression)
			verifier := NewCodeVerifier(code, t).MustContain(tt.expect...).MustNotHavePlaceholders()
			if len(tt.reject) > 0 {
				verifier.MustNotContain(tt.reject...)
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
		"secBarEvaluator.EvaluateAtBar",
		"&ast.ConditionalExpression",
		"Test: &ast.BinaryExpression{Operator: \">\"",
		"secBarIdx",
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
		"ta.stdev(20)",
		"sum := 0.0",                // Mean calculation
		"mean := sum / float64(20)", // Mean result
		"variance := 0.0",           // Variance calculation
		"diff := ctx.Data[ctx.BarIndex-j].Close - mean", // Uses built-in with relative offset
		"variance += diff * diff",                       // Squared deviation
		"math.Sqrt(variance / float64(20))",             // Final STDEV
	}

	for _, pattern := range expectedPatterns {
		if !strings.Contains(code, pattern) {
			t.Errorf("Expected STDEV code to contain %q\nGenerated code:\n%s", pattern, code)
		}
	}
}

func TestSecurityContextIsolation(t *testing.T) {
	code := generateMultiSecurityProgram(t, map[string]ast.Expression{
		"daily":  BinaryExpr("+", Ident("close"), Ident("open")),
		"weekly": BinaryExpr("*", Ident("high"), Lit(2.0)),
	})

	NewCodeVerifier(code, t).
		CountOccurrences("secBarEvaluator.EvaluateAtBar", 2).
		MustNotContain(
			"origCtx := ctx",
			"ctx = origCtx",
			"secTmp_dailySeries",
			"secTmp_weeklySeries",
		).
		MustContain(
			"&ast.BinaryExpression{Operator: \"+\"",
			"&ast.BinaryExpression{Operator: \"*\"",
		)
}
