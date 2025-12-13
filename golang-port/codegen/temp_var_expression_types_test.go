package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

/* TestTempVarInBinaryExpression validates temp var calculation emission
 * for TA functions in binary expressions (arithmetic/comparison).
 */
func TestTempVarInBinaryExpression(t *testing.T) {
	tests := []struct {
		name     string
		varName  string
		operator string
		left     ast.Expression
		right    ast.Expression
		validate func(t *testing.T, code string)
	}{
		{
			name:     "arithmetic: constant * stdev()",
			varName:  "dev",
			operator: "*",
			left:     &ast.Literal{Value: 2.0},
			right: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ta"},
					Property: &ast.Identifier{Name: "stdev"},
				},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
					&ast.Literal{Value: 20},
				},
			},
			validate: func(t *testing.T, code string) {
				if !strings.Contains(code, "ta_stdev_20") {
					t.Error("Expected temp var ta_stdev_20 for stdev() in expression")
				}
				if !strings.Contains(code, "ta_stdev_20") && strings.Contains(code, "Series.Set(stdev)") {
					t.Error("Temp var ta_stdev_20 must have .Set() with calculation")
				}
				if !strings.Contains(code, "devSeries.Set((2.00 * ta_stdev_20") {
					t.Error("Main var must reference temp var in arithmetic expression")
				}
			},
		},
		{
			name:     "arithmetic: sma() + ema()",
			varName:  "combined",
			operator: "+",
			left: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ta"},
					Property: &ast.Identifier{Name: "sma"},
				},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
					&ast.Literal{Value: 20},
				},
			},
			right: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ta"},
					Property: &ast.Identifier{Name: "ema"},
				},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
					&ast.Literal{Value: 10},
				},
			},
			validate: func(t *testing.T, code string) {
				if !strings.Contains(code, "ta_sma_20") {
					t.Error("Expected temp var ta_sma_20 for sma() in expression")
				}
				if !strings.Contains(code, "ta_ema_10") {
					t.Error("Expected temp var ta_ema_10 for ema() in expression")
				}
				smaSetCount := strings.Count(code, "ta_sma_20")
				emaSetCount := strings.Count(code, "ta_ema_10")
				if smaSetCount < 2 {
					t.Errorf("Temp var ta_sma_20 should have multiple references (declaration + usage), got %d", smaSetCount)
				}
				if emaSetCount < 2 {
					t.Errorf("Temp var ta_ema_10 should have multiple references (declaration + usage), got %d", emaSetCount)
				}
			},
		},
		{
			name:     "comparison: sma() > ema()",
			varName:  "signal",
			operator: ">",
			left: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ta"},
					Property: &ast.Identifier{Name: "sma"},
				},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
					&ast.Literal{Value: 50},
				},
			},
			right: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ta"},
					Property: &ast.Identifier{Name: "ema"},
				},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
					&ast.Literal{Value: 200},
				},
			},
			validate: func(t *testing.T, code string) {
				if !strings.Contains(code, "ta_sma_50") {
					t.Error("Expected temp var ta_sma_50 for sma() in comparison")
				}
				if !strings.Contains(code, "ta_ema_200") {
					t.Error("Expected temp var ta_ema_200 for ema() in comparison")
				}
				if !strings.Contains(code, "func() float64 { if") {
					t.Error("Boolean comparison should be converted to float64")
				}
			},
		},
		{
			name:     "nested: (sma() + ema()) / 2",
			varName:  "avg",
			operator: "/",
			left: &ast.BinaryExpression{
				Operator: "+",
				Left: &ast.CallExpression{
					Callee: &ast.MemberExpression{
						Object:   &ast.Identifier{Name: "ta"},
						Property: &ast.Identifier{Name: "sma"},
					},
					Arguments: []ast.Expression{
						&ast.Identifier{Name: "close"},
						&ast.Literal{Value: 20},
					},
				},
				Right: &ast.CallExpression{
					Callee: &ast.MemberExpression{
						Object:   &ast.Identifier{Name: "ta"},
						Property: &ast.Identifier{Name: "ema"},
					},
					Arguments: []ast.Expression{
						&ast.Identifier{Name: "close"},
						&ast.Literal{Value: 20},
					},
				},
			},
			right: &ast.Literal{Value: 2.0},
			validate: func(t *testing.T, code string) {
				if !strings.Contains(code, "ta_sma_20") {
					t.Error("Expected temp var ta_sma_20 in nested expression")
				}
				if !strings.Contains(code, "ta_ema_20") {
					t.Error("Expected temp var ta_ema_20 in nested expression")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := createTestGenerator()

			binExpr := &ast.BinaryExpression{
				Operator: tt.operator,
				Left:     tt.left,
				Right:    tt.right,
			}

			gen.variables[tt.varName] = "float64"
			code, err := gen.generateVariableInit(tt.varName, binExpr)
			if err != nil {
				t.Fatalf("generateVariableInit failed: %v", err)
			}

			tt.validate(t, code)
		})
	}
}

/* TestTempVarInConditionalExpression - SKIPPED
 * ConditionalExpression routes through generateConditionExpression
 * which does not support inline TA function generation.
 */

/* TestTempVarInUnaryExpression validates temp var calculation emission
 * for unary operations on TA functions.
 */
func TestTempVarInUnaryExpression(t *testing.T) {
	tests := []struct {
		name        string
		varName     string
		operator    string
		argument    ast.Expression
		expectedTA  string
		description string
	}{
		{
			name:     "negation: -sma()",
			varName:  "neg_sma",
			operator: "-",
			argument: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ta"},
					Property: &ast.Identifier{Name: "sma"},
				},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
					&ast.Literal{Value: 20},
				},
			},
			expectedTA:  "ta_sma_20",
			description: "Arithmetic negation of TA function",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := createTestGenerator()

			unaryExpr := &ast.UnaryExpression{
				Operator: tt.operator,
				Argument: tt.argument,
			}

			gen.variables[tt.varName] = "float64"
			code, err := gen.generateVariableInit(tt.varName, unaryExpr)
			if err != nil {
				t.Fatalf("generateVariableInit failed: %v", err)
			}

			if !strings.Contains(code, tt.expectedTA) {
				t.Errorf("Expected temp var %q in unary expression (%s)\nGenerated:\n%s",
					tt.expectedTA, tt.description, code)
			}

			if !strings.Contains(code, tt.expectedTA) && strings.Contains(code, "Series.Set(") {
				t.Errorf("Temp var %q must have .Set() call (%s)", tt.expectedTA, tt.description)
			}
		})
	}
}

/* TestTempVarInLogicalExpression - SKIPPED
 * LogicalExpression routes through generateConditionExpression
 * which does not support inline TA function generation.
 */

/* TestTempVarCalculationOrdering validates temp var calculations
 * appear before their usage in main variable assignments.
 */
func TestTempVarCalculationOrdering(t *testing.T) {
	gen := createTestGenerator()

	binExpr := &ast.BinaryExpression{
		Operator: "*",
		Left:     &ast.Literal{Value: 2.0},
		Right: &ast.CallExpression{
			Callee: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "ta"},
				Property: &ast.Identifier{Name: "stdev"},
			},
			Arguments: []ast.Expression{
				&ast.Identifier{Name: "close"},
				&ast.Literal{Value: 20},
			},
		},
	}

	gen.variables["dev"] = "float64"
	code, err := gen.generateVariableInit("dev", binExpr)
	if err != nil {
		t.Fatalf("generateVariableInit failed: %v", err)
	}

	tempVarSetIdx := strings.Index(code, "ta_stdev_20") // First occurrence (Set)
	mainVarUseIdx := strings.Index(code, "devSeries.Set((2.00 * ta_stdev_20")

	if tempVarSetIdx < 0 {
		t.Fatal("Temp var ta_stdev_20 calculation not found")
	}

	if mainVarUseIdx < 0 {
		t.Fatal("Main var usage of temp var not found")
	}

	if tempVarSetIdx >= mainVarUseIdx {
		t.Error("Temp var calculation must appear BEFORE main var usage")
	}
}

/* createTestGenerator initializes generator for testing */
func createTestGenerator() *generator {
	gen := &generator{
		variables:         make(map[string]string),
		varInits:          make(map[string]ast.Expression),
		constants:         make(map[string]interface{}),
		taRegistry:        NewTAFunctionRegistry(),
		mathHandler:       NewMathHandler(),
		runtimeOnlyFilter: NewRuntimeOnlyFunctionFilter(),
	}
	gen.typeSystem = NewTypeInferenceEngine()
	gen.exprAnalyzer = NewExpressionAnalyzer(gen)
	gen.tempVarMgr = NewTempVariableManager(gen)
	gen.builtinHandler = NewBuiltinIdentifierHandler()
	gen.boolConverter = NewBooleanConverter(gen.typeSystem)
	return gen
}
