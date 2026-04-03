package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestTernaryCodegenIntegration(t *testing.T) {
	// Test: signal = close > close_avg ? 1 : 0
	program := &ast.Program{
		Body: []ast.Node{
			&ast.VariableDeclaration{
				Declarations: []ast.VariableDeclarator{
					{
						ID: &ast.Identifier{Name: "signal"},
						Init: &ast.ConditionalExpression{
							Test: &ast.BinaryExpression{
								Operator: ">",
								Left: &ast.MemberExpression{
									Object:   &ast.Identifier{Name: "close"},
									Property: &ast.Literal{Value: float64(0)},
									Computed: true,
								},
								Right: &ast.Identifier{Name: "close_avg"},
							},
							Consequent: &ast.Literal{
								Value: float64(1),
							},
							Alternate: &ast.Literal{
								Value: float64(0),
							},
						},
					},
				},
			},
		},
	}

	gen := newTestGenerator()

	code, err := gen.generateProgram(program)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Verify generated code structure (ForwardSeriesBuffer paradigm)
	if !strings.Contains(code, "var signalSeries *series.Series") {
		t.Errorf("Missing signal Series declaration: got %s", code)
	}

	if !strings.Contains(code, "if (bar.Close > close_avgSeries.GetCurrent()) { return 1") {
		t.Errorf("Missing ternary true branch: got %s", code)
	}

	if !strings.Contains(code, "} else { return 0") {
		t.Errorf("Missing ternary false branch: got %s", code)
	}

	if !strings.Contains(code, "signalSeries.Set(func() float64") {
		t.Errorf("Missing Series.Set with inline function: got %s", code)
	}
}

func TestTernaryWithArithmetic(t *testing.T) {
	// Test: volume_signal = volume > volume_avg * 1.5 ? 1 : 0
	program := &ast.Program{
		Body: []ast.Node{
			&ast.VariableDeclaration{
				Declarations: []ast.VariableDeclarator{
					{
						ID: &ast.Identifier{Name: "volume_signal"},
						Init: &ast.ConditionalExpression{
							Test: &ast.BinaryExpression{
								Operator: ">",
								Left: &ast.MemberExpression{
									Object:   &ast.Identifier{Name: "volume"},
									Property: &ast.Literal{Value: float64(0)},
									Computed: true,
								},
								Right: &ast.BinaryExpression{
									Operator: "*",
									Left:     &ast.Identifier{Name: "volume_avg"},
									Right:    &ast.Literal{Value: float64(1.5)},
								},
							},
							Consequent: &ast.Literal{
								Value: float64(1),
							},
							Alternate: &ast.Literal{
								Value: float64(0),
							},
						},
					},
				},
			},
		},
	}

	gen := newTestGenerator()

	code, err := gen.generateProgram(program)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Verify arithmetic in condition (ForwardSeriesBuffer paradigm)
	if !strings.Contains(code, "volume_avgSeries.GetCurrent() * 1.5") {
		t.Errorf("Missing arithmetic in ternary condition: got %s", code)
	}

	if !strings.Contains(code, "bar.Volume > (volume_avgSeries.GetCurrent() * 1.5)") {
		t.Errorf("Missing complete condition with arithmetic: got %s", code)
	}
}

func TestTernaryWithLogicalOperators(t *testing.T) {
	// Test: signal = close > open and volume > 1000 ? 1 : 0
	program := &ast.Program{
		Body: []ast.Node{
			&ast.VariableDeclaration{
				Declarations: []ast.VariableDeclarator{
					{
						ID: &ast.Identifier{Name: "signal"},
						Init: &ast.ConditionalExpression{
							Test: &ast.LogicalExpression{
								Operator: "and",
								Left: &ast.BinaryExpression{
									Operator: ">",
									Left: &ast.MemberExpression{
										Object:   &ast.Identifier{Name: "close"},
										Property: &ast.Literal{Value: float64(0)},
										Computed: true,
									},
									Right: &ast.MemberExpression{
										Object:   &ast.Identifier{Name: "open"},
										Property: &ast.Literal{Value: float64(0)},
										Computed: true,
									},
								},
								Right: &ast.BinaryExpression{
									Operator: ">",
									Left: &ast.MemberExpression{
										Object:   &ast.Identifier{Name: "volume"},
										Property: &ast.Literal{Value: float64(0)},
										Computed: true,
									},
									Right: &ast.Literal{Value: float64(1000)},
								},
							},
							Consequent: &ast.Literal{
								Value: float64(1),
							},
							Alternate: &ast.Literal{
								Value: float64(0),
							},
						},
					},
				},
			},
		},
	}

	gen := newTestGenerator()

	code, err := gen.generateProgram(program)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Verify logical operator in condition
	if !strings.Contains(code, "&&") {
		t.Errorf("Missing && operator: got %s", code)
	}

	if !strings.Contains(code, "bar.Close > bar.Open") {
		t.Errorf("Missing close > open comparison: got %s", code)
	}

	if !strings.Contains(code, "bar.Volume > 1000") {
		t.Errorf("Missing volume > 1000 comparison: got %s", code)
	}
}

func TestConditionalExpressionOperatorPrecedence(t *testing.T) {
	tests := []struct {
		name       string
		initDecls  []ast.VariableDeclarator
		testExpr   ast.Expression
		expectCode []string
	}{
		{
			name: "arithmetic: multiplication with subtraction",
			initDecls: []ast.VariableDeclarator{
				{
					ID:   &ast.Identifier{Name: "factor"},
					Init: &ast.Literal{Value: 0.02},
				},
			},
			testExpr: &ast.ConditionalExpression{
				Test: &ast.Identifier{Name: "condition"},
				Consequent: &ast.BinaryExpression{
					Left:     &ast.Identifier{Name: "value"},
					Operator: "*",
					Right: &ast.BinaryExpression{
						Left:     &ast.Literal{Value: 1.0},
						Operator: "-",
						Right:    &ast.Identifier{Name: "factor"},
					},
				},
				Alternate: &ast.Identifier{Name: "fallback"},
			},
			expectCode: []string{
				"(1 - factorSeries.GetCurrent())",
			},
		},
		{
			name: "arithmetic: division with addition",
			testExpr: &ast.ConditionalExpression{
				Test: &ast.BinaryExpression{
					Left:     &ast.Identifier{Name: "x"},
					Operator: ">",
					Right:    &ast.Literal{Value: 0.0},
				},
				Consequent: &ast.Identifier{Name: "result"},
				Alternate: &ast.BinaryExpression{
					Left:     &ast.Identifier{Name: "numerator"},
					Operator: "/",
					Right: &ast.BinaryExpression{
						Left:     &ast.Identifier{Name: "denominator"},
						Operator: "+",
						Right:    &ast.Literal{Value: 1},
					},
				},
			},
			expectCode: []string{
				"(denominatorSeries.GetCurrent() + 1)",
			},
		},
		{
			name: "comparison: nested arithmetic",
			testExpr: &ast.ConditionalExpression{
				Test: &ast.BinaryExpression{
					Left: &ast.BinaryExpression{
						Left:     &ast.Identifier{Name: "a"},
						Operator: "+",
						Right:    &ast.Identifier{Name: "b"},
					},
					Operator: ">",
					Right: &ast.BinaryExpression{
						Left:     &ast.Identifier{Name: "c"},
						Operator: "*",
						Right:    &ast.Literal{Value: 2.0},
					},
				},
				Consequent: &ast.Literal{Value: 1.0},
				Alternate:  &ast.Literal{Value: 0.0},
			},
			expectCode: []string{
				"((aSeries.GetCurrent() + bSeries.GetCurrent()) > (cSeries.GetCurrent() * 2))",
			},
		},
		{
			name: "logical: and with comparisons",
			testExpr: &ast.ConditionalExpression{
				Test: &ast.BinaryExpression{
					Left: &ast.BinaryExpression{
						Left:     &ast.Identifier{Name: "price"},
						Operator: ">",
						Right:    &ast.Literal{Value: 100.0},
					},
					Operator: "and",
					Right: &ast.BinaryExpression{
						Left:     &ast.Identifier{Name: "volume"},
						Operator: ">",
						Right:    &ast.Literal{Value: 1000.0},
					},
				},
				Consequent: &ast.Identifier{Name: "signal_on"},
				Alternate:  &ast.Identifier{Name: "signal_off"},
			},
			expectCode: []string{
				"(priceSeries.GetCurrent() > 100)",
				"&&",
				"bar.Volume > 1000",
			},
		},
		{
			name: "logical: or with comparisons",
			testExpr: &ast.ConditionalExpression{
				Test: &ast.BinaryExpression{
					Left: &ast.BinaryExpression{
						Left:     &ast.Identifier{Name: "stop_loss"},
						Operator: "<=",
						Right:    &ast.Identifier{Name: "price"},
					},
					Operator: "or",
					Right: &ast.BinaryExpression{
						Left:     &ast.Identifier{Name: "take_profit"},
						Operator: ">=",
						Right:    &ast.Identifier{Name: "price"},
					},
				},
				Consequent: &ast.Identifier{Name: "close_pos"},
				Alternate:  &ast.Identifier{Name: "hold_pos"},
			},
			expectCode: []string{
				"(stop_lossSeries.GetCurrent() <= priceSeries.GetCurrent())",
				"||",
				"(take_profitSeries.GetCurrent() >= priceSeries.GetCurrent())",
			},
		},
		{
			name: "modulo: remainder with comparison",
			testExpr: &ast.ConditionalExpression{
				Test: &ast.BinaryExpression{
					Left: &ast.BinaryExpression{
						Left:     &ast.Identifier{Name: "bar_index"},
						Operator: "%",
						Right:    &ast.Literal{Value: 5.0},
					},
					Operator: "==",
					Right:    &ast.Literal{Value: 0.0},
				},
				Consequent: &ast.Identifier{Name: "execute"},
				Alternate:  &ast.Identifier{Name: "skip"},
			},
			expectCode: []string{
				"(float64(i % 5) == 0)",
			},
		},
		{
			name: "nested: multi-level expressions",
			testExpr: &ast.ConditionalExpression{
				Test: &ast.Identifier{Name: "flag"},
				Consequent: &ast.BinaryExpression{
					Left: &ast.BinaryExpression{
						Left:     &ast.Identifier{Name: "a"},
						Operator: "+",
						Right:    &ast.Identifier{Name: "b"},
					},
					Operator: "*",
					Right: &ast.BinaryExpression{
						Left:     &ast.Identifier{Name: "c"},
						Operator: "-",
						Right:    &ast.Identifier{Name: "d"},
					},
				},
				Alternate: &ast.Literal{Value: 0.0},
			},
			expectCode: []string{
				"(aSeries.GetCurrent() + bSeries.GetCurrent())",
				"(cSeries.GetCurrent() - dSeries.GetCurrent())",
			},
		},
		{
			name: "mixed: nested arithmetic in division",
			testExpr: &ast.ConditionalExpression{
				Test: &ast.BinaryExpression{
					Left:     &ast.Identifier{Name: "high"},
					Operator: "!=",
					Right:    &ast.Identifier{Name: "low"},
				},
				Consequent: &ast.BinaryExpression{
					Left: &ast.BinaryExpression{
						Left:     &ast.Identifier{Name: "high"},
						Operator: "-",
						Right:    &ast.Identifier{Name: "low"},
					},
					Operator: "/",
					Right:    &ast.Identifier{Name: "close"},
				},
				Alternate: &ast.Literal{Value: 0.0},
			},
			expectCode: []string{
				"((bar.High - bar.Low) / bar.Close)",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := []ast.Node{}
			for _, decl := range tt.initDecls {
				body = append(body, &ast.VariableDeclaration{
					Declarations: []ast.VariableDeclarator{decl},
				})
			}
			body = append(body, &ast.VariableDeclaration{
				Declarations: []ast.VariableDeclarator{
					{
						ID:   &ast.Identifier{Name: "test_result"},
						Init: tt.testExpr,
					},
				},
			})

			program := &ast.Program{Body: body}
			gen := newTestGenerator()

			code, err := gen.generateProgram(program)
			if err != nil {
				t.Fatalf("Generate failed: %v", err)
			}

			for _, expectStr := range tt.expectCode {
				if !strings.Contains(code, expectStr) {
					t.Errorf("Expected pattern %q not found in generated code:\n%s", expectStr, code)
				}
			}
		})
	}
}

func TestConditionalExpressionNumericCoercion(t *testing.T) {
	tests := []struct {
		name     string
		program  *ast.Program
		mustHave []string
	}{
		{
			name: "boolean literals coerce to float64",
			program: &ast.Program{
				Body: []ast.Node{
					&ast.VariableDeclaration{
						Declarations: []ast.VariableDeclarator{
							{
								ID: &ast.Identifier{Name: "signal"},
								Init: &ast.ConditionalExpression{
									Test: &ast.BinaryExpression{
										Operator: ">",
										Left:     &ast.Identifier{Name: "close"},
										Right:    &ast.Identifier{Name: "open"},
									},
									Consequent: &ast.Literal{Value: true},
									Alternate:  &ast.Literal{Value: false},
								},
							},
						},
					},
				},
			},
			mustHave: []string{
				"return 1.0",
				"return 0.0",
			},
		},
		{
			name: "nested ternary with boolean literals",
			program: &ast.Program{
				Body: []ast.Node{
					&ast.VariableDeclaration{
						Declarations: []ast.VariableDeclarator{
							{
								ID: &ast.Identifier{Name: "nested"},
								Init: &ast.ConditionalExpression{
									Test: &ast.BinaryExpression{
										Operator: ">",
										Left:     &ast.Identifier{Name: "close"},
										Right:    &ast.Literal{Value: 100.0},
									},
									Consequent: &ast.ConditionalExpression{
										Test: &ast.BinaryExpression{
											Operator: ">",
											Left:     &ast.Identifier{Name: "close"},
											Right:    &ast.Literal{Value: 150.0},
										},
										Consequent: &ast.Literal{Value: true},
										Alternate:  &ast.Literal{Value: false},
									},
									Alternate: &ast.Literal{Value: false},
								},
							},
						},
					},
				},
			},
			mustHave: []string{
				"return 1.0",
				"return 0.0",
				"func() float64",
			},
		},
		{
			name: "boolean ternary in condition position",
			program: &ast.Program{
				Body: []ast.Node{
					&ast.VariableDeclaration{
						Declarations: []ast.VariableDeclarator{
							{
								ID: &ast.Identifier{Name: "result"},
								Init: &ast.ConditionalExpression{
									Test: &ast.BinaryExpression{
										Operator: ">",
										Left: &ast.ConditionalExpression{
											Test: &ast.BinaryExpression{
												Operator: ">",
												Left:     &ast.Identifier{Name: "close"},
												Right:    &ast.Literal{Value: 100.0},
											},
											Consequent: &ast.Literal{Value: true},
											Alternate:  &ast.Literal{Value: false},
										},
										Right: &ast.Literal{Value: 0.5},
									},
									Consequent: &ast.Literal{Value: 10.0},
									Alternate:  &ast.Literal{Value: 20.0},
								},
							},
						},
					},
				},
			},
			mustHave: []string{
				"return 1.0",
				"return 0.0",
				"return 10",
				"return 20",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := newTestGenerator()

			code, err := gen.generateProgram(tt.program)
			if err != nil {
				t.Fatalf("Generate failed: %v", err)
			}

			for _, pattern := range tt.mustHave {
				if !strings.Contains(code, pattern) {
					t.Errorf("Required pattern %q not found in generated code", pattern)
				}
			}
		})
	}
}
