package codegen

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestBuiltinUsageDetector_DetectsInInitialDeclaration(t *testing.T) {
	detector := NewBuiltinUsageDetector([]string{"dayofweek"})

	program := &ast.Program{
		Body: []ast.Node{
			&ast.VariableDeclaration{
				Kind: "let",
				Declarations: []ast.VariableDeclarator{
					{
						ID: &ast.Identifier{Name: "x"},
						Init: &ast.BinaryExpression{
							Operator: "==",
							Left:     &ast.Identifier{Name: "dayofweek"},
							Right:    &ast.Identifier{Name: "1"},
						},
					},
				},
			},
		},
	}

	found := detector.Detect(program)
	if !found["dayofweek"] {
		t.Error("dayofweek not detected in := initial declaration RHS")
	}
}

func TestBuiltinUsageDetector_DetectsInReassignment(t *testing.T) {
	detector := NewBuiltinUsageDetector([]string{"dayofweek"})

	// Pine := reassignment parses as VariableDeclaration with Kind="var"
	program := &ast.Program{
		Body: []ast.Node{
			&ast.VariableDeclaration{
				Kind: "var",
				Declarations: []ast.VariableDeclarator{
					{
						ID: &ast.Identifier{Name: "tradesLastWeek"},
						Init: &ast.BinaryExpression{
							Operator: "==",
							Left:     &ast.Identifier{Name: "dayofweek"},
							Right: &ast.MemberExpression{
								Object:   &ast.Identifier{Name: "dayofweek"},
								Property: &ast.Identifier{Name: "monday"},
							},
						},
					},
				},
			},
		},
	}

	found := detector.Detect(program)
	if !found["dayofweek"] {
		t.Error("dayofweek not detected in := reassignment (VariableDeclaration Kind=var) RHS")
	}
}

func TestBuiltinUsageDetector_DetectsInNestedExpressions(t *testing.T) {
	detector := NewBuiltinUsageDetector([]string{"close", "open", "high", "low"})

	program := &ast.Program{
		Body: []ast.Node{
			&ast.ExpressionStatement{
				Expression: &ast.ConditionalExpression{
					Test: &ast.BinaryExpression{
						Operator: ">",
						Left:     &ast.Identifier{Name: "close"},
						Right:    &ast.Identifier{Name: "open"},
					},
					Consequent: &ast.Identifier{Name: "high"},
					Alternate:  &ast.Identifier{Name: "low"},
				},
			},
		},
	}

	found := detector.Detect(program)
	for _, name := range []string{"close", "open", "high", "low"} {
		if !found[name] {
			t.Errorf("%q not detected in nested conditional expression", name)
		}
	}
}

func TestBuiltinUsageDetector_DetectsInIfStatementBranches(t *testing.T) {
	detector := NewBuiltinUsageDetector([]string{"barstate", "volume"})

	program := &ast.Program{
		Body: []ast.Node{
			&ast.IfStatement{
				Test: &ast.Identifier{Name: "barstate"},
				Consequent: []ast.Node{
					&ast.ExpressionStatement{
						Expression: &ast.Identifier{Name: "volume"},
					},
				},
			},
		},
	}

	found := detector.Detect(program)
	if !found["barstate"] {
		t.Error("barstate not detected in if test")
	}
	if !found["volume"] {
		t.Error("volume not detected in if consequent")
	}
}

func TestBuiltinUsageDetector_MemberExpressionKeys(t *testing.T) {
	detector := NewBuiltinUsageDetectorWithMembers(
		nil,
		[]string{"session.isfirstbar", "dayofweek.monday"},
	)

	program := &ast.Program{
		Body: []ast.Node{
			&ast.ExpressionStatement{
				Expression: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "session"},
					Property: &ast.Identifier{Name: "isfirstbar"},
				},
			},
		},
	}

	found := detector.Detect(program)
	if !found["session.isfirstbar"] {
		t.Error("session.isfirstbar member key not detected")
	}
	if found["dayofweek.monday"] {
		t.Error("dayofweek.monday should not be detected (not in program)")
	}
}

func TestBuiltinUsageDetector_IfStatementAsRvalue(t *testing.T) {
	makeProgram := func(initExpr ast.Expression) *ast.Program {
		return &ast.Program{
			Body: []ast.Node{
				&ast.VariableDeclaration{
					Kind: "let",
					Declarations: []ast.VariableDeclarator{
						{ID: &ast.Identifier{Name: "x"}, Init: initExpr},
					},
				},
			},
		}
	}

	tests := []struct {
		name     string
		targets  []string
		program  *ast.Program
		wantKeys []string
		wantMiss []string
	}{
		{
			name:    "builtin in if-rvalue test expression",
			targets: []string{"dayofweek"},
			program: makeProgram(&ast.IfStatement{
				Test: &ast.BinaryExpression{
					Operator: "==",
					Left:     &ast.Identifier{Name: "dayofweek"},
					Right:    &ast.Literal{Value: 1.0},
				},
				Consequent: []ast.Node{
					&ast.ExpressionStatement{Expression: &ast.Literal{Value: 1.0}},
				},
			}),
			wantKeys: []string{"dayofweek"},
		},
		{
			name:    "builtin in if-rvalue consequent",
			targets: []string{"volume"},
			program: makeProgram(&ast.IfStatement{
				Test: &ast.Literal{Value: true},
				Consequent: []ast.Node{
					&ast.ExpressionStatement{Expression: &ast.Identifier{Name: "volume"}},
				},
			}),
			wantKeys: []string{"volume"},
		},
		{
			name:    "builtin in if-rvalue alternate",
			targets: []string{"close"},
			program: makeProgram(&ast.IfStatement{
				Test: &ast.Literal{Value: true},
				Consequent: []ast.Node{
					&ast.ExpressionStatement{Expression: &ast.Literal{Value: 0.0}},
				},
				Alternate: []ast.Node{
					&ast.ExpressionStatement{Expression: &ast.Identifier{Name: "close"}},
				},
			}),
			wantKeys: []string{"close"},
		},
		{
			name:    "if-rvalue with no alternate — test and consequent scanned",
			targets: []string{"open", "high"},
			program: makeProgram(&ast.IfStatement{
				Test: &ast.Identifier{Name: "open"},
				Consequent: []ast.Node{
					&ast.ExpressionStatement{Expression: &ast.Identifier{Name: "high"}},
				},
				// no Alternate
			}),
			wantKeys: []string{"open", "high"},
		},
		{
			name:    "builtins distributed across test, consequent, and alternate",
			targets: []string{"dayofweek", "close", "volume"},
			program: makeProgram(&ast.IfStatement{
				Test: &ast.Identifier{Name: "dayofweek"},
				Consequent: []ast.Node{
					&ast.ExpressionStatement{Expression: &ast.Identifier{Name: "close"}},
				},
				Alternate: []ast.Node{
					&ast.ExpressionStatement{Expression: &ast.Identifier{Name: "volume"}},
				},
			}),
			wantKeys: []string{"dayofweek", "close", "volume"},
		},
		{
			name:    "nested if-rvalue in alternate (else-if chain)",
			targets: []string{"open", "close", "high"},
			program: makeProgram(&ast.IfStatement{
				Test: &ast.Identifier{Name: "open"},
				Consequent: []ast.Node{
					&ast.ExpressionStatement{Expression: &ast.Literal{Value: 1.0}},
				},
				Alternate: []ast.Node{
					&ast.IfStatement{
						Test: &ast.Identifier{Name: "close"},
						Consequent: []ast.Node{
							&ast.ExpressionStatement{Expression: &ast.Identifier{Name: "high"}},
						},
					},
				},
			}),
			wantKeys: []string{"open", "close", "high"},
		},
		{
			name:    "if-rvalue inside for-loop body",
			targets: []string{"volume"},
			program: &ast.Program{
				Body: []ast.Node{
					&ast.ForStatement{
						Counter: "i",
						From:    &ast.Literal{Value: 0.0},
						To:      &ast.Literal{Value: 10.0},
						Body: []ast.Node{
							&ast.VariableDeclaration{
								Kind: "let",
								Declarations: []ast.VariableDeclarator{
									{
										ID: &ast.Identifier{Name: "v"},
										Init: &ast.IfStatement{
											Test: &ast.Literal{Value: true},
											Consequent: []ast.Node{
												&ast.ExpressionStatement{Expression: &ast.Identifier{Name: "volume"}},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			wantKeys: []string{"volume"},
		},
		{
			name:    "if-rvalue inside arrow function body",
			targets: []string{"close"},
			program: &ast.Program{
				Body: []ast.Node{
					&ast.VariableDeclaration{
						Kind: "let",
						Declarations: []ast.VariableDeclarator{
							{
								ID: &ast.Identifier{Name: "fn"},
								Init: &ast.ArrowFunctionExpression{
									Params: []ast.Identifier{{Name: "x"}},
									Body: []ast.Node{
										&ast.VariableDeclaration{
											Kind: "let",
											Declarations: []ast.VariableDeclarator{
												{
													ID: &ast.Identifier{Name: "r"},
													Init: &ast.IfStatement{
														Test: &ast.Literal{Value: true},
														Consequent: []ast.Node{
															&ast.ExpressionStatement{Expression: &ast.Identifier{Name: "close"}},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			wantKeys: []string{"close"},
		},
		{
			name:    "non-targeted identifier in if-rvalue is not reported",
			targets: []string{"volume"},
			program: makeProgram(&ast.IfStatement{
				Test: &ast.Identifier{Name: "close"},
				Consequent: []ast.Node{
					&ast.ExpressionStatement{Expression: &ast.Identifier{Name: "open"}},
				},
			}),
			wantMiss: []string{"volume"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detector := NewBuiltinUsageDetector(tt.targets)
			found := detector.Detect(tt.program)
			for _, key := range tt.wantKeys {
				if !found[key] {
					t.Errorf("%q not detected", key)
				}
			}
			for _, key := range tt.wantMiss {
				if found[key] {
					t.Errorf("%q should not be detected", key)
				}
			}
		})
	}
}

func TestBuiltinUsageDetector_IfRvalueIntegration(t *testing.T) {
	tests := []struct {
		name       string
		script     string
		wantSeries []string
		wantAbsent []string
	}{
		{
			name: "dayofweek in if-rvalue triggers series declaration",
			script: `
x = 0
x := if (dayofweek == dayofweek.monday) and (dayofweek != dayofweek[1])
    close
else
    x[1]
`,
			wantSeries: []string{"dayofweekSeries"},
		},
		{
			name: "volume in if-rvalue consequent triggers series declaration",
			script: `
y = 0.0
y := if barstate.isconfirmed
    volume
else
    y[1]
`,
			wantSeries: []string{"volumeSeries"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := compilePineScript(tt.script)
			if err != nil {
				t.Fatalf("compile failed: %v", err)
			}
			for _, s := range tt.wantSeries {
				if !contains(code, s) {
					t.Errorf("%q not declared in generated code", s)
				}
			}
			for _, s := range tt.wantAbsent {
				if contains(code, s) {
					t.Errorf("%q should not appear in generated code", s)
				}
			}
		})
	}
}
