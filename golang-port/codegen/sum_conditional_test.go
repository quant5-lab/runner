package codegen

import (
	"fmt"
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/runtime/validation"
)

// TestSumWithConditionalExpression validates sum(ternary ? a : b, period) handling
func TestSumWithConditionalExpression(t *testing.T) {
	tests := []struct {
		name               string
		testExpression     ast.Expression
		consequent         ast.Expression
		alternate          ast.Expression
		period             int
		expectedTernary    string
		expectedSumPattern string
		expectedTempVar    string
	}{
		{
			name: "sum(close > open ? 1 : 0, 10) - literal ternary",
			testExpression: &ast.BinaryExpression{
				NodeType: ast.TypeBinaryExpression,
				Left:     &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "close"},
				Operator: ">",
				Right:    &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "open"},
			},
			consequent:         &ast.Literal{NodeType: ast.TypeLiteral, Value: 1.0, Raw: "1"},
			alternate:          &ast.Literal{NodeType: ast.TypeLiteral, Value: 0.0, Raw: "0"},
			period:             10,
			expectedTernary:    "ternary_",
			expectedSumPattern: "/* Inline sum(10) */",
			expectedTempVar:    "Series.Set(func() float64 { if",
		},
		{
			name: "sum(volume > volume[1] ? high : low, 5) - series ternary",
			testExpression: &ast.BinaryExpression{
				NodeType: ast.TypeBinaryExpression,
				Left:     &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "volume"},
				Operator: ">",
				Right: &ast.MemberExpression{
					NodeType: ast.TypeMemberExpression,
					Object:   &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "volume"},
					Property: &ast.Literal{NodeType: ast.TypeLiteral, Value: 1.0, Raw: "1"},
					Computed: true,
				},
			},
			consequent:         &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "high"},
			alternate:          &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "low"},
			period:             5,
			expectedTernary:    "ternary_",
			expectedSumPattern: "/* Inline sum(5) */",
			expectedTempVar:    "Series.Set(func() float64 { if",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			typeSystem := NewTypeInferenceEngine()
			gen := &generator{
				variables:     make(map[string]string),
				varInits:      make(map[string]ast.Expression),
				constants:     make(map[string]interface{}),
				taRegistry:    NewTAFunctionRegistry(),
				mathHandler:   NewMathHandler(),
				typeSystem:    typeSystem,
				boolConverter: NewBooleanConverter(typeSystem),
			}
			gen.exprAnalyzer = NewExpressionAnalyzer(gen)
			gen.tempVarMgr = NewTempVariableManager(gen)
			gen.constEvaluator = validation.NewWarmupAnalyzer()

			// Setup built-in variables
			gen.variables["close"] = "float64"
			gen.variables["open"] = "float64"
			gen.variables["high"] = "float64"
			gen.variables["low"] = "float64"
			gen.variables["volume"] = "float64"

			// Create sum call with conditional expression
			sumCall := &ast.CallExpression{
				NodeType: ast.TypeCallExpression,
				Callee:   &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "sum"},
				Arguments: []ast.Expression{
					&ast.ConditionalExpression{
						NodeType:   ast.TypeConditionalExpression,
						Test:       tt.testExpression,
						Consequent: tt.consequent,
						Alternate:  tt.alternate,
					},
					&ast.Literal{NodeType: ast.TypeLiteral, Value: float64(tt.period), Raw: fmt.Sprintf("%d", tt.period)},
				},
			}

			gen.variables["result"] = "float64"
			gen.varInits["result"] = sumCall

			code, err := gen.generateVariableInit("result", sumCall)
			if err != nil {
				t.Fatalf("generateVariableInit failed: %v", err)
			}

			// Verify temp var created for ternary expression
			if !strings.Contains(code, tt.expectedTernary) {
				t.Errorf("Expected ternary temp var '%s' to be created\nGenerated:\n%s", tt.expectedTernary, code)
			}

			// Verify sum loop generated
			if !strings.Contains(code, tt.expectedSumPattern) {
				t.Errorf("Expected sum pattern '%s'\nGenerated:\n%s", tt.expectedSumPattern, code)
			}

			// Verify ternary temp var inline IIFE pattern
			if !strings.Contains(code, tt.expectedTempVar) {
				t.Errorf("Expected ternary temp var pattern '%s'\nGenerated:\n%s", tt.expectedTempVar, code)
			}

			// Verify ternary temp var accessor used in sum loop
			if !strings.Contains(code, "Series.Get(") {
				t.Errorf("Expected sum loop to use ternary temp var accessor\nGenerated:\n%s", code)
			}
		})
	}
}

// TestSumWithoutConditionalExpression validates standard sum() behavior unchanged
func TestSumWithoutConditionalExpression(t *testing.T) {
	gen := &generator{
		variables:   make(map[string]string),
		varInits:    make(map[string]ast.Expression),
		constants:   make(map[string]interface{}),
		taRegistry:  NewTAFunctionRegistry(),
		mathHandler: NewMathHandler(),
	}
	gen.exprAnalyzer = NewExpressionAnalyzer(gen)
	gen.tempVarMgr = NewTempVariableManager(gen)
	gen.constEvaluator = validation.NewWarmupAnalyzer()

	gen.variables["close"] = "float64"

	// Standard sum call without ternary
	sumCall := &ast.CallExpression{
		NodeType: ast.TypeCallExpression,
		Callee:   &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "sum"},
		Arguments: []ast.Expression{
			&ast.Identifier{NodeType: ast.TypeIdentifier, Name: "close"},
			&ast.Literal{NodeType: ast.TypeLiteral, Value: 10.0, Raw: "10"},
		},
	}

	gen.variables["result"] = "float64"
	gen.varInits["result"] = sumCall

	code, err := gen.generateVariableInit("result", sumCall)
	if err != nil {
		t.Fatalf("generateVariableInit failed: %v", err)
	}

	// Should NOT create ternary temp var
	if strings.Contains(code, "ternary_") {
		t.Errorf("Should not create ternary temp var for standard sum\nGenerated:\n%s", code)
	}

	// Should use direct data accessor for built-in variable
	if !strings.Contains(code, "ctx.Data[") && !strings.Contains(code, "closeSeries.Get(") {
		t.Errorf("Expected direct data or series accessor for standard sum\nGenerated:\n%s", code)
	}

	// Verify sum loop
	if !strings.Contains(code, "/* Inline sum(10) */") {
		t.Errorf("Expected sum loop with period 10\nGenerated:\n%s", code)
	}
}
