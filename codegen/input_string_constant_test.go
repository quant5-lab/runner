package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

/* TestInputString_ConstantRegistration validates string input constants are registered and used as scalars */
func TestInputString_ConstantRegistration(t *testing.T) {
	tests := []struct {
		name         string
		declarations []ast.VariableDeclarator
		wantConsts   []string
		noSeries     []string
	}{
		{
			name: "input.string with simple value",
			declarations: []ast.VariableDeclarator{
				{
					ID: &ast.Identifier{Name: "maType"},
					Init: &ast.CallExpression{
						Callee: &ast.MemberExpression{
							Object:   &ast.Identifier{Name: "input"},
							Property: &ast.Identifier{Name: "string"},
						},
						Arguments: []ast.Expression{&ast.Literal{Value: "EMA"}},
					},
				},
			},
			wantConsts: []string{`const maType = "EMA"`},
			noSeries:   []string{"maTypeSeries"},
		},
		{
			name: "input.string with spaces",
			declarations: []ast.VariableDeclarator{
				{
					ID: &ast.Identifier{Name: "title"},
					Init: &ast.CallExpression{
						Callee: &ast.MemberExpression{
							Object:   &ast.Identifier{Name: "input"},
							Property: &ast.Identifier{Name: "string"},
						},
						Arguments: []ast.Expression{&ast.Literal{Value: "Moving Average"}},
					},
				},
			},
			wantConsts: []string{`const title = "Moving Average"`},
			noSeries:   []string{"titleSeries"},
		},
		{
			name: "input.string with empty value",
			declarations: []ast.VariableDeclarator{
				{
					ID: &ast.Identifier{Name: "empty"},
					Init: &ast.CallExpression{
						Callee: &ast.MemberExpression{
							Object:   &ast.Identifier{Name: "input"},
							Property: &ast.Identifier{Name: "string"},
						},
						Arguments: []ast.Expression{&ast.Literal{Value: ""}},
					},
				},
			},
			wantConsts: []string{`const empty = ""`},
			noSeries:   []string{"emptySeries"},
		},
		{
			name: "input.session with time range",
			declarations: []ast.VariableDeclarator{
				{
					ID: &ast.Identifier{Name: "session"},
					Init: &ast.CallExpression{
						Callee: &ast.MemberExpression{
							Object:   &ast.Identifier{Name: "input"},
							Property: &ast.Identifier{Name: "session"},
						},
						Arguments: []ast.Expression{&ast.Literal{Value: "0950-1345"}},
					},
				},
			},
			wantConsts: []string{`const session = "0950-1345"`},
			noSeries:   []string{"sessionSeries"},
		},
		{
			name: "multiple string constants",
			declarations: []ast.VariableDeclarator{
				{
					ID: &ast.Identifier{Name: "ma1Type"},
					Init: &ast.CallExpression{
						Callee: &ast.MemberExpression{
							Object:   &ast.Identifier{Name: "input"},
							Property: &ast.Identifier{Name: "string"},
						},
						Arguments: []ast.Expression{&ast.Literal{Value: "EMA"}},
					},
				},
				{
					ID: &ast.Identifier{Name: "ma2Type"},
					Init: &ast.CallExpression{
						Callee: &ast.MemberExpression{
							Object:   &ast.Identifier{Name: "input"},
							Property: &ast.Identifier{Name: "string"},
						},
						Arguments: []ast.Expression{&ast.Literal{Value: "SMA"}},
					},
				},
			},
			wantConsts: []string{`const ma1Type = "EMA"`, `const ma2Type = "SMA"`},
			noSeries:   []string{"ma1TypeSeries", "ma2TypeSeries"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body []ast.Node
			for _, decl := range tt.declarations {
				body = append(body, &ast.VariableDeclaration{
					Declarations: []ast.VariableDeclarator{decl},
				})
			}

			program := &ast.Program{Body: body}
			code, err := GenerateStrategyCodeFromAST(program)
			if err != nil {
				t.Fatalf("Failed to generate code: %v", err)
			}

			for _, constDecl := range tt.wantConsts {
				if !strings.Contains(code.FunctionBody, constDecl) {
					t.Errorf("Expected constant declaration %q, got:\n%s", constDecl, code.FunctionBody)
				}
			}

			for _, seriesName := range tt.noSeries {
				if strings.Contains(code.FunctionBody, seriesName) {
					t.Errorf("String constant incorrectly generated Series variable: %s", seriesName)
				}
			}
		})
	}
}

/* TestInputString_UsedInConditionalExpression validates string constants used as scalars in ternary */
func TestInputString_UsedInConditionalExpression(t *testing.T) {
	tests := []struct {
		name            string
		program         *ast.Program
		wantScalarUsage []string
		noSeriesUsage   []string
	}{
		{
			name: "ternary expression with string constant comparison",
			program: &ast.Program{
				Body: []ast.Node{
					&ast.VariableDeclaration{
						Declarations: []ast.VariableDeclarator{
							{
								ID: &ast.Identifier{Name: "maType"},
								Init: &ast.CallExpression{
									Callee: &ast.MemberExpression{
										Object:   &ast.Identifier{Name: "input"},
										Property: &ast.Identifier{Name: "string"},
									},
									Arguments: []ast.Expression{&ast.Literal{Value: "EMA"}},
								},
							},
						},
					},
					&ast.VariableDeclaration{
						Declarations: []ast.VariableDeclarator{
							{
								ID: &ast.Identifier{Name: "result"},
								Init: &ast.ConditionalExpression{
									Test: &ast.BinaryExpression{
										Operator: "==",
										Left:     &ast.Identifier{Name: "maType"},
										Right:    &ast.Literal{Value: "EMA"},
									},
									Consequent: &ast.Literal{Value: 1.0},
									Alternate:  &ast.Literal{Value: 0.0},
								},
							},
						},
					},
				},
			},
			wantScalarUsage: []string{`maType == "EMA"`},
			noSeriesUsage:   []string{"maTypeSeries.GetCurrent()"},
		},
		{
			name: "multiple string constants in nested ternary",
			program: &ast.Program{
				Body: []ast.Node{
					&ast.VariableDeclaration{
						Declarations: []ast.VariableDeclarator{
							{
								ID: &ast.Identifier{Name: "type1"},
								Init: &ast.CallExpression{
									Callee: &ast.MemberExpression{
										Object:   &ast.Identifier{Name: "input"},
										Property: &ast.Identifier{Name: "string"},
									},
									Arguments: []ast.Expression{&ast.Literal{Value: "A"}},
								},
							},
						},
					},
					&ast.VariableDeclaration{
						Declarations: []ast.VariableDeclarator{
							{
								ID: &ast.Identifier{Name: "type2"},
								Init: &ast.CallExpression{
									Callee: &ast.MemberExpression{
										Object:   &ast.Identifier{Name: "input"},
										Property: &ast.Identifier{Name: "string"},
									},
									Arguments: []ast.Expression{&ast.Literal{Value: "B"}},
								},
							},
						},
					},
					&ast.VariableDeclaration{
						Declarations: []ast.VariableDeclarator{
							{
								ID: &ast.Identifier{Name: "result"},
								Init: &ast.ConditionalExpression{
									Test: &ast.BinaryExpression{
										Operator: "==",
										Left:     &ast.Identifier{Name: "type1"},
										Right:    &ast.Literal{Value: "A"},
									},
									Consequent: &ast.ConditionalExpression{
										Test: &ast.BinaryExpression{
											Operator: "==",
											Left:     &ast.Identifier{Name: "type2"},
											Right:    &ast.Literal{Value: "B"},
										},
										Consequent: &ast.Literal{Value: 1.0},
										Alternate:  &ast.Literal{Value: 2.0},
									},
									Alternate: &ast.Literal{Value: 3.0},
								},
							},
						},
					},
				},
			},
			wantScalarUsage: []string{`type1 == "A"`, `type2 == "B"`},
			noSeriesUsage:   []string{"type1Series.GetCurrent()", "type2Series.GetCurrent()"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := GenerateStrategyCodeFromAST(tt.program)
			if err != nil {
				t.Fatalf("Failed to generate code: %v", err)
			}

			for _, usage := range tt.wantScalarUsage {
				if !strings.Contains(code.FunctionBody, usage) {
					t.Errorf("Expected scalar usage %q in generated code", usage)
				}
			}

			for _, seriesUsage := range tt.noSeriesUsage {
				if strings.Contains(code.FunctionBody, seriesUsage) {
					t.Errorf("String constant incorrectly used as Series: %s", seriesUsage)
				}
			}
		})
	}
}

/* TestInputString_MixedWithOtherInputTypes validates string constants coexist with numeric/bool constants */
func TestInputString_MixedWithOtherInputTypes(t *testing.T) {
	program := &ast.Program{
		Body: []ast.Node{
			&ast.VariableDeclaration{
				Declarations: []ast.VariableDeclarator{
					{
						ID: &ast.Identifier{Name: "maType"},
						Init: &ast.CallExpression{
							Callee: &ast.MemberExpression{
								Object:   &ast.Identifier{Name: "input"},
								Property: &ast.Identifier{Name: "string"},
							},
							Arguments: []ast.Expression{&ast.Literal{Value: "EMA"}},
						},
					},
				},
			},
			&ast.VariableDeclaration{
				Declarations: []ast.VariableDeclarator{
					{
						ID: &ast.Identifier{Name: "length"},
						Init: &ast.CallExpression{
							Callee: &ast.MemberExpression{
								Object:   &ast.Identifier{Name: "input"},
								Property: &ast.Identifier{Name: "int"},
							},
							Arguments: []ast.Expression{&ast.Literal{Value: 14.0}},
						},
					},
				},
			},
			&ast.VariableDeclaration{
				Declarations: []ast.VariableDeclarator{
					{
						ID: &ast.Identifier{Name: "enabled"},
						Init: &ast.CallExpression{
							Callee: &ast.MemberExpression{
								Object:   &ast.Identifier{Name: "input"},
								Property: &ast.Identifier{Name: "bool"},
							},
							Arguments: []ast.Expression{&ast.Literal{Value: true}},
						},
					},
				},
			},
			&ast.VariableDeclaration{
				Declarations: []ast.VariableDeclarator{
					{
						ID: &ast.Identifier{Name: "result"},
						Init: &ast.ConditionalExpression{
							Test: &ast.BinaryExpression{
								Operator: "==",
								Left:     &ast.Identifier{Name: "maType"},
								Right:    &ast.Literal{Value: "EMA"},
							},
							Consequent: &ast.Identifier{Name: "length"},
							Alternate:  &ast.Literal{Value: 20.0},
						},
					},
				},
			},
		},
	}

	code, err := GenerateStrategyCodeFromAST(program)
	if err != nil {
		t.Fatalf("Failed to generate code: %v", err)
	}

	wantConsts := []string{
		`const maType = "EMA"`,
		"const length = 14",
		"const enabled = true",
	}
	for _, constDecl := range wantConsts {
		if !strings.Contains(code.FunctionBody, constDecl) {
			t.Errorf("Expected constant declaration %q", constDecl)
		}
	}

	noSeries := []string{"maTypeSeries", "lengthSeries", "enabledSeries"}
	for _, seriesName := range noSeries {
		if strings.Contains(code.FunctionBody, seriesName) {
			t.Errorf("Input constant incorrectly generated Series: %s", seriesName)
		}
	}

	if !strings.Contains(code.FunctionBody, `maType == "EMA"`) {
		t.Error("String constant should be used as scalar in comparison")
	}
	if !strings.Contains(code.FunctionBody, "length") && !strings.Contains(code.FunctionBody, "14") {
		t.Error("Int constant should be accessible in ternary consequent")
	}
}

/* TestInputString_NotConfusedWithSeriesVariables validates string constants vs actual series variables */
func TestInputString_NotConfusedWithSeriesVariables(t *testing.T) {
	program := &ast.Program{
		Body: []ast.Node{
			&ast.VariableDeclaration{
				Declarations: []ast.VariableDeclarator{
					{
						ID: &ast.Identifier{Name: "maType"},
						Init: &ast.CallExpression{
							Callee: &ast.MemberExpression{
								Object:   &ast.Identifier{Name: "input"},
								Property: &ast.Identifier{Name: "string"},
							},
							Arguments: []ast.Expression{&ast.Literal{Value: "EMA"}},
						},
					},
				},
			},
			&ast.VariableDeclaration{
				Declarations: []ast.VariableDeclarator{
					{
						ID: &ast.Identifier{Name: "smaValue"},
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

	code, err := GenerateStrategyCodeFromAST(program)
	if err != nil {
		t.Fatalf("Failed to generate code: %v", err)
	}

	if !strings.Contains(code.FunctionBody, `const maType = "EMA"`) {
		t.Error("String constant should be declared as const")
	}

	if !strings.Contains(code.FunctionBody, "smaValueSeries") {
		t.Error("TA function result should create Series variable")
	}

	if strings.Contains(code.FunctionBody, "maTypeSeries") {
		t.Error("String constant should NOT create Series variable")
	}
}
