package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

/* TestConstantRegistry_InputConstantsIsolation validates input constants never generate Series artifacts */
func TestConstantRegistry_InputConstantsIsolation(t *testing.T) {
	tests := []struct {
		name         string
		declarations []ast.VariableDeclarator
		wantConsts   []string
		wantVars     []string
		noSeries     []string
	}{
		{
			name: "input.int inferred from plain input() with int literal",
			declarations: []ast.VariableDeclarator{
				{
					ID: &ast.Identifier{Name: "length"},
					Init: &ast.CallExpression{
						Callee:    &ast.Identifier{Name: "input"},
						Arguments: []ast.Expression{&ast.Literal{Value: 20.0}},
					},
				},
			},
			wantConsts: []string{"const length = 20"},
			wantVars:   []string{},
			noSeries:   []string{"lengthSeries"},
		},
		{
			name: "input.float inferred from plain input() with float literal",
			declarations: []ast.VariableDeclarator{
				{
					ID: &ast.Identifier{Name: "factor"},
					Init: &ast.CallExpression{
						Callee:    &ast.Identifier{Name: "input"},
						Arguments: []ast.Expression{&ast.Literal{Value: 1.5}},
					},
				},
			},
			wantConsts: []string{"const factor = 1.50"},
			wantVars:   []string{},
			noSeries:   []string{"factorSeries"},
		},
		{
			name: "multiple input constants with different types",
			declarations: []ast.VariableDeclarator{
				{
					ID: &ast.Identifier{Name: "bblenght"},
					Init: &ast.CallExpression{
						Callee:    &ast.Identifier{Name: "input"},
						Arguments: []ast.Expression{&ast.Literal{Value: 46.0}},
					},
				},
				{
					ID: &ast.Identifier{Name: "bbstdev"},
					Init: &ast.CallExpression{
						Callee:    &ast.Identifier{Name: "input"},
						Arguments: []ast.Expression{&ast.Literal{Value: 0.35}},
					},
				},
			},
			wantConsts: []string{"const bblenght = 46", "const bbstdev = 0.35"},
			wantVars:   []string{},
			noSeries:   []string{"bblenghtSeries", "bbstdevSeries"},
		},
		{
			name: "explicit input.float",
			declarations: []ast.VariableDeclarator{
				{
					ID: &ast.Identifier{Name: "mult"},
					Init: &ast.CallExpression{
						Callee: &ast.MemberExpression{
							Object:   &ast.Identifier{Name: "input"},
							Property: &ast.Identifier{Name: "float"},
						},
						Arguments: []ast.Expression{&ast.Literal{Value: 2.5}},
					},
				},
			},
			wantConsts: []string{"const mult = 2.50"},
			wantVars:   []string{},
			noSeries:   []string{"multSeries"},
		},
		{
			name: "explicit input.int",
			declarations: []ast.VariableDeclarator{
				{
					ID: &ast.Identifier{Name: "period"},
					Init: &ast.CallExpression{
						Callee: &ast.MemberExpression{
							Object:   &ast.Identifier{Name: "input"},
							Property: &ast.Identifier{Name: "int"},
						},
						Arguments: []ast.Expression{&ast.Literal{Value: 14.0}},
					},
				},
			},
			wantConsts: []string{"const period = 14"},
			wantVars:   []string{},
			noSeries:   []string{"periodSeries"},
		},
		{
			name: "explicit input.bool",
			declarations: []ast.VariableDeclarator{
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
			wantConsts: []string{"const enabled = true"},
			wantVars:   []string{},
			noSeries:   []string{"enabledSeries"},
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

			// Verify constant declarations present
			for _, constDecl := range tt.wantConsts {
				if !strings.Contains(code.FunctionBody, constDecl) {
					t.Errorf("Expected constant declaration %q", constDecl)
				}
			}

			// Verify variables have Series (if any expected)
			for _, varName := range tt.wantVars {
				if !strings.Contains(code.FunctionBody, "var "+varName+"Series *series.Series") {
					t.Errorf("Expected variable %q to have Series declaration", varName)
				}
				if !strings.Contains(code.FunctionBody, varName+"Series = series.NewSeries") {
					t.Errorf("Expected variable %q to have Series initialization", varName)
				}
				if !strings.Contains(code.FunctionBody, varName+"Series.Next()") {
					t.Errorf("Expected variable %q to have .Next() call", varName)
				}
			}

			// Verify constants do NOT have Series artifacts
			for _, constName := range tt.noSeries {
				if strings.Contains(code.FunctionBody, "var "+constName+" *series.Series") {
					t.Errorf("Constant should NOT have Series declaration: %q", constName)
				}
				if strings.Contains(code.FunctionBody, constName+" = series.NewSeries") {
					t.Errorf("Constant should NOT have Series initialization: %q", constName)
				}
				if strings.Contains(code.FunctionBody, "_ = "+constName) {
					t.Errorf("Constant should NOT have unused suppression: %q", constName)
				}
				if strings.Contains(code.FunctionBody, constName+".Next()") {
					t.Errorf("Constant should NOT have .Next() call: %q", constName)
				}
			}
		})
	}
}

/* TestInputConstants_VariableSeparation tests that constants never leak into variable lifecycle */
func TestInputConstants_VariableSeparation(t *testing.T) {
	tests := []struct {
		name          string
		body          []ast.Node
		wantConst     string
		noConstSeries string
		wantVar       string
		wantVarSeries string
	}{
		{
			name: "input constant used in binary expression",
			body: []ast.Node{
				&ast.VariableDeclaration{
					Declarations: []ast.VariableDeclarator{
						{
							ID: &ast.Identifier{Name: "length"},
							Init: &ast.CallExpression{
								Callee: &ast.MemberExpression{
									Object:   &ast.Identifier{Name: "input"},
									Property: &ast.Identifier{Name: "int"},
								},
								Arguments: []ast.Expression{&ast.Literal{Value: 20.0}},
							},
						},
					},
				},
				&ast.VariableDeclaration{
					Declarations: []ast.VariableDeclarator{
						{
							ID: &ast.Identifier{Name: "result"},
							Init: &ast.BinaryExpression{
								Operator: "*",
								Left: &ast.MemberExpression{
									Object:   &ast.Identifier{Name: "bar"},
									Property: &ast.Identifier{Name: "Close"},
								},
								Right: &ast.Identifier{Name: "length"},
							},
						},
					},
				},
			},
			wantConst:     "const length = 20",
			noConstSeries: "lengthSeries",
			wantVar:       "result",
			wantVarSeries: "resultSeries",
		},
		{
			name: "input constant referenced by multiple variables",
			body: []ast.Node{
				&ast.VariableDeclaration{
					Declarations: []ast.VariableDeclarator{
						{
							ID: &ast.Identifier{Name: "factor"},
							Init: &ast.CallExpression{
								Callee: &ast.MemberExpression{
									Object:   &ast.Identifier{Name: "input"},
									Property: &ast.Identifier{Name: "float"},
								},
								Arguments: []ast.Expression{&ast.Literal{Value: 1.5}},
							},
						},
					},
				},
				&ast.VariableDeclaration{
					Declarations: []ast.VariableDeclarator{
						{
							ID: &ast.Identifier{Name: "value1"},
							Init: &ast.BinaryExpression{
								Operator: "*",
								Left:     &ast.Identifier{Name: "factor"},
								Right:    &ast.Literal{Value: 100.0},
							},
						},
					},
				},
				&ast.VariableDeclaration{
					Declarations: []ast.VariableDeclarator{
						{
							ID: &ast.Identifier{Name: "value2"},
							Init: &ast.BinaryExpression{
								Operator: "/",
								Left:     &ast.Literal{Value: 50.0},
								Right:    &ast.Identifier{Name: "factor"},
							},
						},
					},
				},
			},
			wantConst:     "const factor = 1.50",
			noConstSeries: "factorSeries",
			wantVar:       "value1",
			wantVarSeries: "value2Series",
		},
		{
			name: "input bool constant in ternary condition",
			body: []ast.Node{
				&ast.VariableDeclaration{
					Declarations: []ast.VariableDeclarator{
						{
							ID: &ast.Identifier{Name: "useFilter"},
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
							ID: &ast.Identifier{Name: "signal"},
							Init: &ast.ConditionalExpression{
								Test:       &ast.Identifier{Name: "useFilter"},
								Consequent: &ast.Literal{Value: 1.0},
								Alternate:  &ast.Literal{Value: 0.0},
							},
						},
					},
				},
			},
			wantConst:     "const useFilter = true",
			noConstSeries: "useFilterSeries",
			wantVar:       "signal",
			wantVarSeries: "signalSeries",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			program := &ast.Program{Body: tt.body}
			code, err := GenerateStrategyCodeFromAST(program)
			if err != nil {
				t.Fatalf("Failed to generate code: %v", err)
			}

			// Input constant should exist
			if !strings.Contains(code.FunctionBody, tt.wantConst) {
				t.Errorf("Expected constant declaration %q", tt.wantConst)
			}

			// Input constant should NOT have Series
			if strings.Contains(code.FunctionBody, "var "+tt.noConstSeries+" *series.Series") {
				t.Errorf("Constant should NOT have Series declaration: %q", tt.noConstSeries)
			}
			if strings.Contains(code.FunctionBody, tt.noConstSeries+" = series.NewSeries") {
				t.Errorf("Constant should NOT have Series initialization: %q", tt.noConstSeries)
			}
			if strings.Contains(code.FunctionBody, tt.noConstSeries+".Next()") {
				t.Errorf("Constant should NOT have .Next() call: %q", tt.noConstSeries)
			}
			if strings.Contains(code.FunctionBody, "_ = "+tt.noConstSeries) {
				t.Errorf("Constant should NOT have unused suppression: %q", tt.noConstSeries)
			}

			// Variables should have Series
			if !strings.Contains(code.FunctionBody, "var "+tt.wantVarSeries+" *series.Series") {
				t.Errorf("Variable should have Series declaration: %q", tt.wantVarSeries)
			}
			if !strings.Contains(code.FunctionBody, tt.wantVarSeries+" = series.NewSeries") {
				t.Errorf("Variable should have Series initialization: %q", tt.wantVarSeries)
			}
			if !strings.Contains(code.FunctionBody, tt.wantVarSeries+".Next()") {
				t.Errorf("Variable should have .Next() call: %q", tt.wantVarSeries)
			}
		})
	}
}

/* TestInputConstants_SeriesLifecycleEdgeCases tests boundary conditions */
func TestInputConstants_SeriesLifecycleEdgeCases(t *testing.T) {
	tests := []struct {
		name           string
		body           []ast.Node
		mustNotContain []string
		mustContain    []string
		description    string
	}{
		{
			name: "empty input constants should not generate artifacts",
			body: []ast.Node{
				&ast.VariableDeclaration{
					Declarations: []ast.VariableDeclarator{
						{
							ID: &ast.Identifier{Name: "x"},
							Init: &ast.CallExpression{
								Callee: &ast.MemberExpression{
									Object:   &ast.Identifier{Name: "input"},
									Property: &ast.Identifier{Name: "float"},
								},
								Arguments: []ast.Expression{},
							},
						},
					},
				},
			},
			mustNotContain: []string{
				"var xSeries *series.Series",
				"xSeries = series.NewSeries",
				"xSeries.Next()",
				"_ = xSeries",
			},
			mustContain: []string{"const x = 0.00"},
			description: "Default input.float(0.0) should not create Series",
		},
		{
			name: "input constants in first pass only",
			body: []ast.Node{
				&ast.VariableDeclaration{
					Declarations: []ast.VariableDeclarator{
						{
							ID: &ast.Identifier{Name: "a"},
							Init: &ast.CallExpression{
								Callee:    &ast.Identifier{Name: "input"},
								Arguments: []ast.Expression{&ast.Literal{Value: 10.0}},
							},
						},
					},
				},
				&ast.VariableDeclaration{
					Declarations: []ast.VariableDeclarator{
						{
							ID: &ast.Identifier{Name: "b"},
							Init: &ast.CallExpression{
								Callee:    &ast.Identifier{Name: "input"},
								Arguments: []ast.Expression{&ast.Literal{Value: 20.0}},
							},
						},
					},
				},
				&ast.VariableDeclaration{
					Declarations: []ast.VariableDeclarator{
						{
							ID: &ast.Identifier{Name: "c"},
							Init: &ast.CallExpression{
								Callee:    &ast.Identifier{Name: "input"},
								Arguments: []ast.Expression{&ast.Literal{Value: 30.0}},
							},
						},
					},
				},
			},
			mustNotContain: []string{
				"var aSeries",
				"var bSeries",
				"var cSeries",
				"aSeries.Next()",
				"bSeries.Next()",
				"cSeries.Next()",
			},
			mustContain: []string{
				"const a = 10",
				"const b = 20",
				"const c = 30",
			},
			description: "Multiple sequential input constants should not leak into variable lifecycle",
		},
		{
			name: "input constants never in unused suppression block",
			body: []ast.Node{
				&ast.VariableDeclaration{
					Declarations: []ast.VariableDeclarator{
						{
							ID: &ast.Identifier{Name: "unused_input"},
							Init: &ast.CallExpression{
								Callee: &ast.MemberExpression{
									Object:   &ast.Identifier{Name: "input"},
									Property: &ast.Identifier{Name: "int"},
								},
								Arguments: []ast.Expression{&ast.Literal{Value: 5.0}},
							},
						},
					},
				},
				&ast.VariableDeclaration{
					Declarations: []ast.VariableDeclarator{
						{
							ID:   &ast.Identifier{Name: "used_var"},
							Init: &ast.Literal{Value: 100.0},
						},
					},
				},
			},
			mustNotContain: []string{
				"_ = unused_inputSeries",
			},
			mustContain: []string{
				"const unused_input = 5",
				"_ = used_varSeries",
			},
			description: "Unused input constants should not appear in suppression block",
		},
		{
			name: "input constants do not trigger bar field Series registration",
			body: []ast.Node{
				&ast.VariableDeclaration{
					Declarations: []ast.VariableDeclarator{
						{
							ID: &ast.Identifier{Name: "Close"},
							Init: &ast.CallExpression{
								Callee: &ast.MemberExpression{
									Object:   &ast.Identifier{Name: "input"},
									Property: &ast.Identifier{Name: "float"},
								},
								Arguments: []ast.Expression{&ast.Literal{Value: 123.45}},
							},
						},
					},
				},
			},
			mustNotContain: []string{
				"var CloseSeries *series.Series",
			},
			mustContain: []string{
				"const Close = 123.45",
				"var closeSeries *series.Series", // Bar field should still exist
			},
			description: "Input constant shadowing bar field name should not affect bar field Series",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			program := &ast.Program{Body: tt.body}
			code, err := GenerateStrategyCodeFromAST(program)
			if err != nil {
				t.Fatalf("Failed to generate code: %v", err)
			}

			for _, forbidden := range tt.mustNotContain {
				if strings.Contains(code.FunctionBody, forbidden) {
					t.Errorf("%s: Found forbidden pattern %q", tt.description, forbidden)
				}
			}

			for _, required := range tt.mustContain {
				if !strings.Contains(code.FunctionBody, required) {
					t.Errorf("%s: Missing required pattern %q", tt.description, required)
				}
			}
		})
	}
}

/* TestInputConstants_ConstantRegistryConsistency tests registry state integrity */
func TestInputConstants_ConstantRegistryConsistency(t *testing.T) {
	program := &ast.Program{
		Body: []ast.Node{
			&ast.VariableDeclaration{
				Declarations: []ast.VariableDeclarator{
					{
						ID: &ast.Identifier{Name: "period"},
						Init: &ast.CallExpression{
							Callee:    &ast.Identifier{Name: "input"},
							Arguments: []ast.Expression{&ast.Literal{Value: 14.0}},
						},
					},
				},
			},
			&ast.VariableDeclaration{
				Declarations: []ast.VariableDeclarator{
					{
						ID: &ast.Identifier{Name: "sma_val"},
						Init: &ast.CallExpression{
							Callee: &ast.MemberExpression{
								Object:   &ast.Identifier{Name: "ta"},
								Property: &ast.Identifier{Name: "sma"},
							},
							Arguments: []ast.Expression{
								&ast.MemberExpression{
									Object:   &ast.Identifier{Name: "bar"},
									Property: &ast.Identifier{Name: "Close"},
								},
								&ast.Identifier{Name: "period"},
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

	// Period should be constant
	if !strings.Contains(code.FunctionBody, "const period = 14") {
		t.Error("Input constant 'period' should be declared")
	}

	// Period should be usable as constant in ta.sma call (no Series)
	if strings.Contains(code.FunctionBody, "periodSeries") {
		t.Error("Input constant should not create Series: periodSeries")
	}

	// sma_val should have Series
	if !strings.Contains(code.FunctionBody, "var sma_valSeries *series.Series") {
		t.Error("Variable sma_val should have Series")
	}
}
