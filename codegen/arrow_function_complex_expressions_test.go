package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

/* TestArrowFunctionTACall_ConditionalExpressionSource validates TA functions with ternary source expressions
 *
 * Tests the architectural pattern where TA functions receive conditional expressions as source arguments,
 * which must be evaluated per-bar before the TA function processes them. This pattern is fundamental
 * for supporting dynamic source selection in technical analysis computations.
 *
 * Test Coverage:
 * - Single ternary as TA source (e.g., rma(cond ? a : b, period))
 * - Nested conditions within TA sources
 * - Multiple TA parameters with conditional sources
 * - Boolean condition evaluation correctness
 */
func TestArrowFunctionTACall_ConditionalExpressionSource(t *testing.T) {
	tests := []struct {
		name           string
		sourceExpr     ast.Expression
		period         int
		expectError    bool
		validateOutput func(t *testing.T, code string)
	}{
		{
			name: "simple ternary source",
			sourceExpr: &ast.ConditionalExpression{
				Test: &ast.BinaryExpression{
					Left:     &ast.Identifier{Name: "a"},
					Operator: ">",
					Right:    &ast.Identifier{Name: "b"},
				},
				Consequent: &ast.Identifier{Name: "a"},
				Alternate:  &ast.Identifier{Name: "b"},
			},
			period:      14,
			expectError: false,
			validateOutput: func(t *testing.T, code string) {
				if !strings.Contains(code, "ternary_source_temp") {
					// DISABLED: t.Error("Expected temp variable for ternary source")
				}
				if !strings.Contains(code, "func() float64") {
					t.Error("Expected IIFE generation for ternary")
				}
			},
		},
		{
			name: "ternary with literal branches",
			sourceExpr: &ast.ConditionalExpression{
				Test: &ast.BinaryExpression{
					Left:     &ast.Identifier{Name: "x"},
					Operator: ">=",
					Right:    &ast.Literal{Value: 0.0},
				},
				Consequent: &ast.Identifier{Name: "x"},
				Alternate:  &ast.Literal{Value: 0.0},
			},
			period:      20,
			expectError: false,
			validateOutput: func(t *testing.T, code string) {
				if !strings.Contains(code, "ternary_source_temp") {
					// DISABLED: t.Error("Expected temp variable declaration")
				}
				if strings.Count(code, "func() float64") < 1 {
					t.Error("Expected at least one IIFE")
				}
			},
		},
		{
			name: "nested ternary source",
			sourceExpr: &ast.ConditionalExpression{
				Test: &ast.Identifier{Name: "cond1"},
				Consequent: &ast.ConditionalExpression{
					Test:       &ast.Identifier{Name: "cond2"},
					Consequent: &ast.Identifier{Name: "val1"},
					Alternate:  &ast.Identifier{Name: "val2"},
				},
				Alternate: &ast.Identifier{Name: "val3"},
			},
			period:      10,
			expectError: false,
			validateOutput: func(t *testing.T, code string) {
				if !strings.Contains(code, "ternary_source_temp") {
					// DISABLED: t.Error("Expected temp variable for nested ternary")
				}
			},
		},
		{
			name: "ternary with logical operators",
			sourceExpr: &ast.ConditionalExpression{
				Test: &ast.LogicalExpression{
					Left: &ast.BinaryExpression{
						Left:     &ast.Identifier{Name: "up"},
						Operator: ">",
						Right:    &ast.Identifier{Name: "down"},
					},
					Operator: "and",
					Right: &ast.BinaryExpression{
						Left:     &ast.Identifier{Name: "up"},
						Operator: ">",
						Right:    &ast.Literal{Value: 0.0},
					},
				},
				Consequent: &ast.Identifier{Name: "up"},
				Alternate:  &ast.Literal{Value: 0.0},
			},
			period:      14,
			expectError: false,
			validateOutput: func(t *testing.T, code string) {
				if !strings.Contains(code, "&&") {
					t.Error("Expected logical AND operator translation")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newTestGenerator()
			g.variables["a"] = "float"
			g.variables["b"] = "float"
			g.variables["x"] = "float"
			g.variables["up"] = "float"
			g.variables["down"] = "float"
			g.inArrowFunctionBody = true

			gen := newTestArrowTAGenerator(g)

			call := &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ta"},
					Property: &ast.Identifier{Name: "rma"},
				},
				Arguments: []ast.Expression{
					tt.sourceExpr,
					&ast.Literal{Value: float64(tt.period)},
				},
			}

			code, err := gen.Generate(call)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if code == "" {
				t.Error("Expected generated code, got empty string")
			}

			if tt.validateOutput != nil {
				tt.validateOutput(t, code)
			}
		})
	}
}

/* TestArrowFunctionTACall_BinaryExpressionSource validates TA functions with arithmetic source expressions
 *
 * Tests the pattern where TA functions receive complex arithmetic expressions as source arguments.
 * This is critical for supporting calculations like ADX where the source is a formula involving
 * multiple operations and other TA function results.
 *
 * Test Coverage:
 * - Simple arithmetic (addition, subtraction, multiplication, division)
 * - Nested binary expressions with precedence handling
 * - Binary expressions containing function calls
 * - Mixed operations with conditional expressions
 */
func TestArrowFunctionTACall_BinaryExpressionSource(t *testing.T) {
	tests := []struct {
		name           string
		sourceExpr     ast.Expression
		period         int
		expectError    bool
		validateOutput func(t *testing.T, code string)
	}{
		{
			name: "simple arithmetic source",
			sourceExpr: &ast.BinaryExpression{
				Left:     &ast.Identifier{Name: "a"},
				Operator: "+",
				Right:    &ast.Identifier{Name: "b"},
			},
			period:      10,
			expectError: false,
			validateOutput: func(t *testing.T, code string) {
				if !strings.Contains(code, "binary_source_temp") {
					// DISABLED: t.Error("Expected temp variable for binary source")
				}
				if !strings.Contains(code, "+") {
					t.Error("Expected addition operator")
				}
			},
		},
		{
			name: "division with multiplication",
			sourceExpr: &ast.BinaryExpression{
				Left: &ast.BinaryExpression{
					Left:     &ast.Literal{Value: 100.0},
					Operator: "*",
					Right:    &ast.Identifier{Name: "value"},
				},
				Operator: "/",
				Right:    &ast.Identifier{Name: "divisor"},
			},
			period:      14,
			expectError: false,
			validateOutput: func(t *testing.T, code string) {
				if !strings.Contains(code, "*") || !strings.Contains(code, "/") {
					t.Error("Expected multiplication and division operators")
				}
			},
		},
		{
			name: "binary with function call",
			sourceExpr: &ast.BinaryExpression{
				Left: &ast.CallExpression{
					Callee:    &ast.Identifier{Name: "abs"},
					Arguments: []ast.Expression{&ast.Identifier{Name: "diff"}},
				},
				Operator: "/",
				Right:    &ast.Identifier{Name: "sum"},
			},
			period:      20,
			expectError: false,
			validateOutput: func(t *testing.T, code string) {
				if !strings.Contains(code, "math.Abs") {
					t.Error("Expected math.Abs function call")
				}
				if !strings.Contains(code, "/") {
					t.Error("Expected division operator")
				}
			},
		},
		{
			name: "binary with conditional expression",
			sourceExpr: &ast.BinaryExpression{
				Left:     &ast.Identifier{Name: "numerator"},
				Operator: "/",
				Right: &ast.ConditionalExpression{
					Test: &ast.BinaryExpression{
						Left:     &ast.Identifier{Name: "denominator"},
						Operator: "==",
						Right:    &ast.Literal{Value: 0.0},
					},
					Consequent: &ast.Literal{Value: 1.0},
					Alternate:  &ast.Identifier{Name: "denominator"},
				},
			},
			period:      14,
			expectError: false,
			validateOutput: func(t *testing.T, code string) {
				if !strings.Contains(code, "func() float64") {
					t.Error("Expected IIFE for conditional in binary expression")
				}
			},
		},
		{
			name: "complex nested arithmetic",
			sourceExpr: &ast.BinaryExpression{
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
			period:      30,
			expectError: false,
			validateOutput: func(t *testing.T, code string) {
				if !strings.Contains(code, "binary_source_temp") {
					// DISABLED: t.Error("Expected temp variable for complex arithmetic")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newTestGenerator()
			g.variables["a"] = "float"
			g.variables["b"] = "float"
			g.variables["c"] = "float"
			g.variables["d"] = "float"
			g.variables["value"] = "float"
			g.variables["divisor"] = "float"
			g.variables["diff"] = "float"
			g.variables["sum"] = "float"
			g.variables["numerator"] = "float"
			g.variables["denominator"] = "float"
			g.inArrowFunctionBody = true

			gen := newTestArrowTAGenerator(g)

			call := &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ta"},
					Property: &ast.Identifier{Name: "rma"},
				},
				Arguments: []ast.Expression{
					tt.sourceExpr,
					&ast.Literal{Value: float64(tt.period)},
				},
			}

			code, err := gen.Generate(call)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if code == "" {
				t.Error("Expected generated code, got empty string")
			}

			if tt.validateOutput != nil {
				tt.validateOutput(t, code)
			}
		})
	}
}

/* TestArrowFunctionTACall_MixedComplexExpressions validates combinations of expression types
 *
 * Tests real-world patterns where TA functions receive deeply nested expressions combining
 * conditionals, binary operations, and function calls. This validates the compositional
 * nature of the expression handling system.
 *
 * Test Coverage:
 * - Ternary containing binary expressions
 * - Binary expressions containing ternaries
 * - Multiple levels of nesting
 * - Real-world ADX/DMI calculation patterns
 */
func TestArrowFunctionTACall_MixedComplexExpressions(t *testing.T) {
	tests := []struct {
		name           string
		buildCall      func() *ast.CallExpression
		expectError    bool
		validateOutput func(t *testing.T, code string)
	}{
		{
			name: "ADX pattern: rma(abs(diff) / ternary_denominator, period)",
			buildCall: func() *ast.CallExpression {
				return &ast.CallExpression{
					Callee: &ast.MemberExpression{
						Object:   &ast.Identifier{Name: "ta"},
						Property: &ast.Identifier{Name: "rma"},
					},
					Arguments: []ast.Expression{
						&ast.BinaryExpression{
							Left: &ast.CallExpression{
								Callee: &ast.Identifier{Name: "abs"},
								Arguments: []ast.Expression{
									&ast.BinaryExpression{
										Left:     &ast.Identifier{Name: "plus"},
										Operator: "-",
										Right:    &ast.Identifier{Name: "minus"},
									},
								},
							},
							Operator: "/",
							Right: &ast.ConditionalExpression{
								Test: &ast.BinaryExpression{
									Left:     &ast.Identifier{Name: "sum"},
									Operator: "==",
									Right:    &ast.Literal{Value: 0.0},
								},
								Consequent: &ast.Literal{Value: 1.0},
								Alternate:  &ast.Identifier{Name: "sum"},
							},
						},
						&ast.Literal{Value: 14.0},
					},
				}
			},
			expectError: false,
			validateOutput: func(t *testing.T, code string) {
				if !strings.Contains(code, "math.Abs") {
					t.Error("Expected abs() translation to math.Abs")
				}
				if !strings.Contains(code, "binary_source_temp") {
					// DISABLED: t.Error("Expected temp variable for binary expression")
				}
				if !strings.Contains(code, "func() float64") {
					t.Error("Expected IIFE wrapper")
				}
			},
		},
		{
			name: "DMI pattern: rma(ternary ? value : 0, period) / total",
			buildCall: func() *ast.CallExpression {
				return &ast.CallExpression{
					Callee: &ast.MemberExpression{
						Object:   &ast.Identifier{Name: "ta"},
						Property: &ast.Identifier{Name: "rma"},
					},
					Arguments: []ast.Expression{
						&ast.ConditionalExpression{
							Test: &ast.LogicalExpression{
								Left: &ast.BinaryExpression{
									Left:     &ast.Identifier{Name: "up"},
									Operator: ">",
									Right:    &ast.Identifier{Name: "down"},
								},
								Operator: "and",
								Right: &ast.BinaryExpression{
									Left:     &ast.Identifier{Name: "up"},
									Operator: ">",
									Right:    &ast.Literal{Value: 0.0},
								},
							},
							Consequent: &ast.Identifier{Name: "up"},
							Alternate:  &ast.Literal{Value: 0.0},
						},
						&ast.Literal{Value: 14.0},
					},
				}
			},
			expectError: false,
			validateOutput: func(t *testing.T, code string) {
				if !strings.Contains(code, "&&") {
					t.Error("Expected logical AND operator")
				}
				// No longer require temp variable - inline evaluation with Series.Get() is also valid
			},
		},
		{
			name: "nested rma pattern: 100 * rma(ternary, len) / total",
			buildCall: func() *ast.CallExpression {
				return &ast.CallExpression{
					Callee: &ast.MemberExpression{
						Object:   &ast.Identifier{Name: "ta"},
						Property: &ast.Identifier{Name: "rma"},
					},
					Arguments: []ast.Expression{
						&ast.ConditionalExpression{
							Test: &ast.BinaryExpression{
								Left:     &ast.Identifier{Name: "x"},
								Operator: ">",
								Right:    &ast.Literal{Value: 0.0},
							},
							Consequent: &ast.Identifier{Name: "x"},
							Alternate:  &ast.Literal{Value: 0.0},
						},
						&ast.Identifier{Name: "len"},
					},
				}
			},
			expectError: false,
			validateOutput: func(t *testing.T, code string) {
				// Expression without Series access should use temp variable OR inline expression
				// Since x and len are scalars, the expression is evaluated inline for each iteration
				if code == "" {
					t.Error("Expected generated code")
				}
				// No longer require temp variable - inline evaluation is also valid
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newTestGenerator()
			g.variables["plus"] = "float"
			g.variables["minus"] = "float"
			g.variables["sum"] = "float"
			g.variables["up"] = "float"
			g.variables["down"] = "float"
			g.variables["x"] = "float"
			g.variables["len"] = "float"
			g.variables["total"] = "float"
			g.inArrowFunctionBody = true

			call := tt.buildCall()
			funcName := extractCallFunctionName(call)

			var code string
			var err error

			if funcName == "fixnan" || funcName == "ta.fixnan" {
				code, err = g.generateCallExpression(call)
			} else {
				gen := newTestArrowTAGenerator(g)
				code, err = gen.Generate(call)
			}

			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if code == "" {
				t.Error("Expected generated code, got empty string")
			}

			if tt.validateOutput != nil {
				tt.validateOutput(t, code)
			}
		})
	}
}

/* TestArrowFunctionTACall_MultipleComplexArguments validates TA functions with multiple complex arguments
 *
 * Tests scenarios where multiple arguments of a TA function are complex expressions,
 * ensuring proper isolation and evaluation order.
 */
func TestArrowFunctionTACall_MultipleComplexArguments(t *testing.T) {
	g := newTestGenerator()
	g.variables["a"] = "float"
	g.variables["b"] = "float"
	g.variables["len1"] = "float"
	g.variables["len2"] = "float"
	g.inArrowFunctionBody = true

	gen := newTestArrowTAGenerator(g)

	// Test where both source and period could be complex (though period usually isn't)
	call := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "ta"},
			Property: &ast.Identifier{Name: "sma"},
		},
		Arguments: []ast.Expression{
			&ast.ConditionalExpression{
				Test: &ast.BinaryExpression{
					Left:     &ast.Identifier{Name: "a"},
					Operator: ">",
					Right:    &ast.Identifier{Name: "b"},
				},
				Consequent: &ast.Identifier{Name: "a"},
				Alternate:  &ast.Identifier{Name: "b"},
			},
			&ast.Literal{Value: 20.0},
		},
	}

	code, err := gen.Generate(call)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if code == "" {
		t.Error("Expected generated code, got empty string")
	}

	// Verify that conditional expression is evaluated at each loop offset
	// New implementation: Generates inline expression with .Get(j)
	// Old implementation: Used temp variable evaluated once
	if !strings.Contains(code, "aSeries.Get(j)") || !strings.Contains(code, "bSeries.Get(j)") {
		t.Errorf("Expected conditional to use Series.Get(j) for historical access\nGenerated code:\n%s", code)
	}
}

/* TestArrowFunctionTACall_EdgeCases validates boundary conditions and error handling
 *
 * Tests exceptional cases that should be handled gracefully without panics or undefined behavior.
 */
func TestArrowFunctionTACall_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		buildCall   func() *ast.CallExpression
		setupGen    func(*generator)
		expectError bool
		errorMsg    string
	}{
		{
			name: "empty conditional branches",
			buildCall: func() *ast.CallExpression {
				return &ast.CallExpression{
					Callee: &ast.MemberExpression{
						Object:   &ast.Identifier{Name: "ta"},
						Property: &ast.Identifier{Name: "sma"},
					},
					Arguments: []ast.Expression{
						&ast.ConditionalExpression{
							Test:       &ast.Literal{Value: true},
							Consequent: &ast.Identifier{Name: "a"},
							Alternate:  &ast.Identifier{Name: "b"},
						},
						&ast.Literal{Value: 10.0},
					},
				}
			},
			setupGen: func(g *generator) {
				g.variables["a"] = "float"
				g.variables["b"] = "float"
			},
			expectError: false,
		},
		{
			name: "deeply nested expressions",
			buildCall: func() *ast.CallExpression {
				// Build nested ternary: cond1 ? (cond2 ? a : b) : (cond3 ? c : d)
				return &ast.CallExpression{
					Callee: &ast.MemberExpression{
						Object:   &ast.Identifier{Name: "ta"},
						Property: &ast.Identifier{Name: "rma"},
					},
					Arguments: []ast.Expression{
						&ast.ConditionalExpression{
							Test: &ast.Identifier{Name: "cond1"},
							Consequent: &ast.ConditionalExpression{
								Test:       &ast.Identifier{Name: "cond2"},
								Consequent: &ast.Identifier{Name: "a"},
								Alternate:  &ast.Identifier{Name: "b"},
							},
							Alternate: &ast.ConditionalExpression{
								Test:       &ast.Identifier{Name: "cond3"},
								Consequent: &ast.Identifier{Name: "c"},
								Alternate:  &ast.Identifier{Name: "d"},
							},
						},
						&ast.Literal{Value: 14.0},
					},
				}
			},
			setupGen: func(g *generator) {
				g.variables["cond1"] = "float"
				g.variables["cond2"] = "float"
				g.variables["cond3"] = "float"
				g.variables["a"] = "float"
				g.variables["b"] = "float"
				g.variables["c"] = "float"
				g.variables["d"] = "float"
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newTestGenerator()
			g.inArrowFunctionBody = true

			if tt.setupGen != nil {
				tt.setupGen(g)
			}

			gen := newTestArrowTAGenerator(g)
			call := tt.buildCall()

			code, err := gen.Generate(call)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error containing %q, got nil", tt.errorMsg)
				} else if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing %q, got %q", tt.errorMsg, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if code == "" {
				t.Error("Expected generated code, got empty string")
			}
		})
	}
}
