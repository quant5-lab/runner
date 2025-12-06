package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestTempVarCreationForMathWithNestedTA(t *testing.T) {
	tests := []struct {
		name          string
		varName       string
		initExpr      ast.Expression
		expectedCode  []string
		unexpectedCode []string
	}{
		{
			name:    "rma with max(change(x), 0) creates temp vars",
			varName: "sr_up",
			initExpr: &ast.CallExpression{
				NodeType: ast.TypeCallExpression,
				Callee: &ast.MemberExpression{
					NodeType: ast.TypeMemberExpression,
					Object:   &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "ta"},
					Property: &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "rma"},
				},
				Arguments: []ast.Expression{
					&ast.CallExpression{
						NodeType: ast.TypeCallExpression,
						Callee:   &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "max"},
						Arguments: []ast.Expression{
							&ast.CallExpression{
								NodeType: ast.TypeCallExpression,
								Callee: &ast.MemberExpression{
									NodeType: ast.TypeMemberExpression,
									Object:   &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "ta"},
									Property: &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "change"},
								},
								Arguments: []ast.Expression{
									&ast.Identifier{NodeType: ast.TypeIdentifier, Name: "src"},
								},
							},
							&ast.Literal{NodeType: ast.TypeLiteral, Value: 0.0, Raw: "0"},
						},
					},
					&ast.Literal{NodeType: ast.TypeLiteral, Value: 9.0, Raw: "9"},
				},
			},
			expectedCode: []string{
				"ta_change",          // Temp var for change()
				"Series.Set(",        // Temp var Series.Set()
				"max_",               // Temp var for max() with hash
				"sr_upSeries.Set(",   // Main variable Series.Set()
				"GetCurrent()",       // Accessor for temp var
			},
			unexpectedCode: []string{
				"func() float64",     // Should not inline change() as IIFE
				"bar.Close - ctx.Data", // Should not inline change calculation
			},
		},
		{
			name:    "rma with -min(change(x), 0) creates temp vars",
			varName: "sr_down",
			initExpr: &ast.CallExpression{
				NodeType: ast.TypeCallExpression,
				Callee: &ast.MemberExpression{
					NodeType: ast.TypeMemberExpression,
					Object:   &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "ta"},
					Property: &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "rma"},
				},
				Arguments: []ast.Expression{
					&ast.UnaryExpression{
						NodeType: ast.TypeUnaryExpression,
						Operator: "-",
						Argument: &ast.CallExpression{
							NodeType: ast.TypeCallExpression,
							Callee:   &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "min"},
							Arguments: []ast.Expression{
								&ast.CallExpression{
									NodeType: ast.TypeCallExpression,
									Callee: &ast.MemberExpression{
										NodeType: ast.TypeMemberExpression,
										Object:   &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "ta"},
										Property: &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "change"},
									},
									Arguments: []ast.Expression{
										&ast.Identifier{NodeType: ast.TypeIdentifier, Name: "src"},
									},
								},
								&ast.Literal{NodeType: ast.TypeLiteral, Value: 0.0, Raw: "0"},
							},
						},
					},
					&ast.Literal{NodeType: ast.TypeLiteral, Value: 9.0, Raw: "9"},
				},
			},
			expectedCode: []string{
				"ta_change",
				"min_",               // Temp var with hash
				"sr_downSeries.Set(",
			},
		},
		{
			name:    "pure math function without TA - no temp var",
			varName: "result",
			initExpr: &ast.CallExpression{
				NodeType: ast.TypeCallExpression,
				Callee:   &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "max"},
				Arguments: []ast.Expression{
					&ast.Identifier{NodeType: ast.TypeIdentifier, Name: "a"},
					&ast.Literal{NodeType: ast.TypeLiteral, Value: 0.0, Raw: "0"},
				},
			},
			expectedCode: []string{
				"resultSeries.Set(",
				"math.Max(",
			},
			unexpectedCode: []string{
				"math_max",          // No temp var for pure math (temp var names have hash)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := &generator{
				variables:      make(map[string]string),
				varInits:       make(map[string]ast.Expression),
				constants:      make(map[string]interface{}),
				taRegistry:     NewTAFunctionRegistry(),
				mathHandler:    NewMathHandler(),
			}
			gen.exprAnalyzer = NewExpressionAnalyzer(gen)
			gen.tempVarMgr = NewTempVariableManager(gen)

			// Add variable to context
			gen.variables[tt.varName] = "float64"
			gen.varInits[tt.varName] = tt.initExpr

			code, err := gen.generateVariableInit(tt.varName, tt.initExpr)
			if err != nil {
				t.Fatalf("generateVariableInit failed: %v", err)
			}

			// Check expected code patterns
			for _, expected := range tt.expectedCode {
				if !strings.Contains(code, expected) {
					t.Errorf("Expected code to contain %q\nGenerated code:\n%s", expected, code)
				}
			}

			// Check unexpected code patterns
			for _, unexpected := range tt.unexpectedCode {
				if strings.Contains(code, unexpected) {
					t.Errorf("Expected code NOT to contain %q\nGenerated code:\n%s", unexpected, code)
				}
			}
		})
	}
}

func TestTempVarRegistrationBeforeUsage(t *testing.T) {
	gen := &generator{
		variables:      make(map[string]string),
		varInits:       make(map[string]ast.Expression),
		constants:      make(map[string]interface{}),
		taRegistry:     NewTAFunctionRegistry(),
		mathHandler:    NewMathHandler(),
	}
	gen.exprAnalyzer = NewExpressionAnalyzer(gen)
	gen.tempVarMgr = NewTempVariableManager(gen)

	changeCall := &ast.CallExpression{
		NodeType: ast.TypeCallExpression,
		Callee: &ast.MemberExpression{
			NodeType: ast.TypeMemberExpression,
			Object:   &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "ta"},
			Property: &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "change"},
		},
		Arguments: []ast.Expression{
			&ast.Identifier{NodeType: ast.TypeIdentifier, Name: "src"},
		},
	}

	maxCall := &ast.CallExpression{
		NodeType: ast.TypeCallExpression,
		Callee:   &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "max"},
		Arguments: []ast.Expression{
			changeCall,
			&ast.Literal{NodeType: ast.TypeLiteral, Value: 0.0, Raw: "0"},
		},
	}

	gen.variables["test_var"] = "float64"
	gen.varInits["test_var"] = maxCall

	// Generate code - should create temp var for change(), but not for max()
	// max() is top-level expression, doesn't need temp var
	code, err := gen.generateVariableInit("test_var", maxCall)
	if err != nil {
		t.Fatalf("generateVariableInit failed: %v", err)
	}

	// Verify temp var created for nested TA function (change)
	if !strings.Contains(code, "ta_change") {
		t.Errorf("Expected ta_change temp var for nested TA call\nGenerated:\n%s", code)
	}

	// Verify max() is inlined directly (no temp var needed for top-level math function)
	if !strings.Contains(code, "test_varSeries.Set(math.Max(") {
		t.Errorf("Expected max() to be inlined directly\nGenerated:\n%s", code)
	}

	// Verify change temp var appears before max usage
	if !strings.Contains(code, "ta_change_") || !strings.Contains(code, "math.Max(ta_change_") {
		t.Errorf("Expected change temp var to be created before max() usage\nGenerated:\n%s", code)
	}
}
