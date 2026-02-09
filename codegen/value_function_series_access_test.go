package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

/* Validates inline value functions (nz, na, fixnan) in extractSeriesExpression */
func TestValueFunctionsInSeriesExpressions(t *testing.T) {
	gen := &generator{
		variables:      make(map[string]string),
		varInits:       make(map[string]ast.Expression),
		constants:      make(map[string]interface{}),
		valueHandler:   NewValueHandler(),
		tempVarMgr:     NewTempVariableManager(&generator{}),
		mathHandler:    NewMathHandler(),
		builtinHandler: NewBuiltinIdentifierHandler(),
	}
	gen.tempVarMgr = NewTempVariableManager(gen)

	tests := []struct {
		name     string
		expr     ast.Expression
		expected string
		desc     string
	}{
		{
			name: "nz with series subscript",
			expr: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "nz"},
				Arguments: []ast.Expression{
					&ast.MemberExpression{
						Object:   &ast.Identifier{Name: "value"},
						Property: &ast.Literal{Value: 1.0},
						Computed: true,
					},
				},
			},
			expected: "value.Nz(valueSeries.Get(1), 0)",
			desc:     "nz(value[1]) generates value.Nz() with Series.Get()",
		},
		{
			name: "nz with series subscript and replacement",
			expr: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "nz"},
				Arguments: []ast.Expression{
					&ast.MemberExpression{
						Object:   &ast.Identifier{Name: "count"},
						Property: &ast.Literal{Value: 2.0},
						Computed: true,
					},
					&ast.Literal{Value: -1.0},
				},
			},
			expected: "value.Nz(countSeries.Get(2), -1)",
			desc:     "nz(count[2], -1) generates value.Nz() with custom replacement",
		},
		{
			name: "na with series subscript",
			expr: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "na"},
				Arguments: []ast.Expression{
					&ast.MemberExpression{
						Object:   &ast.Identifier{Name: "signal"},
						Property: &ast.Literal{Value: 0.0},
						Computed: true,
					},
				},
			},
			expected: "math.IsNaN(signalSeries.Get(0))",
			desc:     "na(signal[0]) generates math.IsNaN() with Series.Get()",
		},
		{
			name: "nz with current value",
			expr: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "nz"},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "price"},
				},
			},
			expected: "value.Nz(priceSeries.GetCurrent(), 0)",
			desc:     "nz(price) generates value.Nz() with GetCurrent()",
		},
		{
			name: "nz with builtin series",
			expr: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "nz"},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
				},
			},
			expected: "value.Nz(bar.Close, 0)",
			desc:     "nz(close) handles builtin series",
		},
		{
			name: "nz with literal",
			expr: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "nz"},
				Arguments: []ast.Expression{
					&ast.Literal{Value: 100.0},
					&ast.Literal{Value: 0.0},
				},
			},
			expected: "value.Nz(100, 0)",
			desc:     "nz(100, 0) handles literal arguments",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := gen.extractSeriesExpression(tt.expr)
			if result != tt.expected {
				t.Errorf("%s\nexpected: %s\ngot:      %s", tt.desc, tt.expected, result)
			}
		})
	}
}

/* Validates value functions vs temp Series variables in extractSeriesExpression */
func TestValueFunctionsVsTempVariables(t *testing.T) {
	gen := &generator{
		variables:    make(map[string]string),
		varInits:     make(map[string]ast.Expression),
		constants:    make(map[string]interface{}),
		valueHandler: NewValueHandler(),
		mathHandler:  NewMathHandler(),
	}
	gen.tempVarMgr = NewTempVariableManager(gen)

	gen.variables["ta_sma_20"] = "float64"

	tests := []struct {
		name        string
		expr        ast.Expression
		shouldBeNz  bool
		shouldBeSMA bool
		description string
	}{
		{
			name: "nz function not temp var",
			expr: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "nz"},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "x"},
				},
			},
			shouldBeNz:  true,
			shouldBeSMA: false,
			description: "nz() generates value.Nz(), not nzSeries",
		},
		{
			name: "ta.sma is temp var",
			expr: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ta"},
					Property: &ast.Identifier{Name: "sma"},
				},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
					&ast.Literal{Value: 20.0},
				},
			},
			shouldBeNz:  false,
			shouldBeSMA: true,
			description: "ta.sma() references temp Series variable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := gen.extractSeriesExpression(tt.expr)

			if tt.shouldBeNz {
				if !strings.Contains(result, "value.Nz") {
					t.Errorf("%s: expected value.Nz(), got: %s", tt.description, result)
				}
				if strings.Contains(result, "Series") && strings.Contains(result, "nzSeries") {
					t.Errorf("%s: should NOT reference nzSeries: %s", tt.description, result)
				}
			}

			if tt.shouldBeSMA {
				/* SMA without registered temp var falls through to default naming */
				if strings.Contains(result, "value.") {
					t.Errorf("%s: should NOT be value function: %s", tt.description, result)
				}
			}
		})
	}
}

/* Validates value functions in arithmetic and logical expressions */
func TestValueFunctionsInBinaryExpressions(t *testing.T) {
	gen := &generator{
		variables:    make(map[string]string),
		varInits:     make(map[string]ast.Expression),
		constants:    make(map[string]interface{}),
		valueHandler: NewValueHandler(),
		mathHandler:  NewMathHandler(),
	}
	gen.tempVarMgr = NewTempVariableManager(gen)

	tests := []struct {
		name     string
		expr     *ast.BinaryExpression
		mustHave string
		mustNot  string
	}{
		{
			name: "nz in addition",
			expr: &ast.BinaryExpression{
				Operator: "+",
				Left: &ast.CallExpression{
					Callee: &ast.Identifier{Name: "nz"},
					Arguments: []ast.Expression{
						&ast.MemberExpression{
							Object:   &ast.Identifier{Name: "a"},
							Property: &ast.Literal{Value: 1.0},
							Computed: true,
						},
					},
				},
				Right: &ast.Literal{Value: 10.0},
			},
			mustHave: "value.Nz(aSeries.Get(1), 0)",
			mustNot:  "nzSeries",
		},
		{
			name: "na in comparison",
			expr: &ast.BinaryExpression{
				Operator: "==",
				Left: &ast.CallExpression{
					Callee: &ast.Identifier{Name: "na"},
					Arguments: []ast.Expression{
						&ast.Identifier{Name: "x"},
					},
				},
				Right: &ast.Literal{Value: 1.0},
			},
			mustHave: "math.IsNaN(xSeries.GetCurrent())",
			mustNot:  "naSeries",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := gen.extractSeriesExpression(tt.expr)

			if !strings.Contains(result, tt.mustHave) {
				t.Errorf("expected to contain: %s\ngot: %s", tt.mustHave, result)
			}
			if strings.Contains(result, tt.mustNot) {
				t.Errorf("should NOT contain: %s\ngot: %s", tt.mustNot, result)
			}
		})
	}
}

/* Edge cases for value function handling */
func TestValueFunctionsEdgeCases(t *testing.T) {
	gen := &generator{
		variables:    make(map[string]string),
		varInits:     make(map[string]ast.Expression),
		constants:    make(map[string]interface{}),
		valueHandler: NewValueHandler(),
		mathHandler:  NewMathHandler(),
	}
	gen.tempVarMgr = NewTempVariableManager(gen)

	tests := []struct {
		name     string
		expr     ast.Expression
		mustHave []string
		mustNot  []string
	}{
		{
			name: "nz with zero replacement",
			expr: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "nz"},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "val"},
					&ast.Literal{Value: 0.0},
				},
			},
			mustHave: []string{"value.Nz", "0"},
			mustNot:  []string{"nzSeries"},
		},
		{
			name: "na with no arguments",
			expr: &ast.CallExpression{
				Callee:    &ast.Identifier{Name: "na"},
				Arguments: []ast.Expression{},
			},
			mustHave: []string{"true"},
			mustNot:  []string{"naSeries", "math.IsNaN"},
		},
		{
			name: "nz with negative replacement",
			expr: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "nz"},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "delta"},
					&ast.Literal{Value: -999.0},
				},
			},
			mustHave: []string{"value.Nz", "-999"},
			mustNot:  []string{"nzSeries"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := gen.extractSeriesExpression(tt.expr)

			for _, must := range tt.mustHave {
				if !strings.Contains(result, must) {
					t.Errorf("expected to contain: %s\ngot: %s", must, result)
				}
			}
			for _, mustNot := range tt.mustNot {
				if strings.Contains(result, mustNot) {
					t.Errorf("should NOT contain: %s\ngot: %s", mustNot, result)
				}
			}
		})
	}
}
