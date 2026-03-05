package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/runtime/validation"
)

/*
Integration tests for crossover codegen with AST construction.

These tests validate internal helper functions and codegen structure using programmatically constructed ASTs.
Test hierarchy:
- crossover_inline_handler_test.go: Unit tests for handler behavior (preferred for new tests)
- generator_crossover_test.go: Integration tests for helper functions and codegen structure
- crossover_arbitrary_pinescript_test.go: E2E compilation tests for full pipeline

Note: extractSeriesExpression is tested here as an internal API; this couples to implementation
details. Prefer crossover_inline_handler_test.go for new behavioral tests.
Offset shifting (convertSeriesAccessToPrev) is unit-tested in series_offset_shifter_test.go.
*/

func TestExtractSeriesExpression(t *testing.T) {
	gen := &generator{
		imports:        make(map[string]bool),
		variables:      make(map[string]string),
		strategyConfig: NewStrategyConfig(),
		taRegistry:     NewTAFunctionRegistry(),
		builtinHandler: NewBuiltinIdentifierHandler(),
	}

	tests := []struct {
		name     string
		expr     ast.Expression
		expected string
	}{
		{
			name: "close built-in series",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "close"},
				Property: &ast.Literal{Value: 0},
				Computed: true,
			},
			expected: "bar.Close",
		},
		{
			name: "open built-in series",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "open"},
				Property: &ast.Literal{Value: 0},
				Computed: true,
			},
			expected: "bar.Open",
		},
		{
			name:     "user variable identifier",
			expr:     &ast.Identifier{Name: "sma20"},
			expected: "sma20Series.GetCurrent()",
		},
		{
			name: "user variable with subscript",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "sma20"},
				Property: &ast.Literal{Value: 0},
			},
			expected: "sma20Series.Get(0)",
		},
		{
			name:     "float literal",
			expr:     &ast.Literal{Value: 100.50},
			expected: "100.5",
		},
		{
			name: "arithmetic expression",
			expr: &ast.BinaryExpression{
				Operator: "*",
				Left:     &ast.Identifier{Name: "sma20"},
				Right:    &ast.Literal{Value: 1.02},
			},
			expected: "(sma20Series.GetCurrent() * 1.02)",
		},
		{
			name: "complex arithmetic",
			expr: &ast.BinaryExpression{
				Operator: "+",
				Left: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "close"},
					Property: &ast.Literal{Value: 0},
					Computed: true,
				},
				Right: &ast.BinaryExpression{
					Operator: "*",
					Left:     &ast.Identifier{Name: "sma20"},
					Right:    &ast.Literal{Value: 0.05},
				},
			},
			expected: "(bar.Close + sma20Series.GetCurrent() * 0.05)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := gen.extractSeriesExpression(tt.expr)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestCrossoverCodegenIntegration(t *testing.T) {
	call := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "ta"},
			Property: &ast.Identifier{Name: "crossover"},
		},
		Arguments: []ast.Expression{
			&ast.MemberExpression{
				Object:   &ast.Identifier{Name: "close"},
				Property: &ast.Literal{Value: 0},
				Computed: true,
			},
			&ast.MemberExpression{
				Object:   &ast.Identifier{Name: "sma20"},
				Property: &ast.Literal{Value: 0},
				Computed: true,
			},
		},
	}

	gen := &generator{
		imports:        make(map[string]bool),
		variables:      make(map[string]string),
		strategyConfig: NewStrategyConfig(),
		taRegistry:     NewTAFunctionRegistry(),
		builtinHandler: NewBuiltinIdentifierHandler(),
	}

	code, err := gen.generateVariableFromCall("longCross", call)
	if err != nil {
		t.Fatalf("generateVariableFromCall failed: %v", err)
	}

	t.Logf("Generated code:\n%s", code)

	if !strings.Contains(code, "longCrossSeries.Set(0.0)") {
		t.Error("Missing initial Series.Set(0.0) assignment")
	}
	if !strings.Contains(code, "if i > 0") {
		t.Error("Missing warmup check")
	}
	if !strings.Contains(code, "ctx.Data[i-1].Close") {
		t.Error("Missing previous close access")
	}
	if !strings.Contains(code, "bar.Close > sma20Series.Get(0)") {
		t.Error("Missing crossover condition (current)")
	}
	if !strings.Contains(code, "&&") {
		t.Error("Missing AND operator")
	}
	if !strings.Contains(code, "<=") {
		t.Error("Missing previous comparison operator")
	}
	if !strings.Contains(code, "longCrossSeries.Set(func() float64") {
		t.Error("Missing Series.Set with bool→float64 conversion")
	}
}

func TestCrossunderCodegenIntegration(t *testing.T) {
	call := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "ta"},
			Property: &ast.Identifier{Name: "crossunder"},
		},
		Arguments: []ast.Expression{
			&ast.MemberExpression{
				Object:   &ast.Identifier{Name: "close"},
				Property: &ast.Literal{Value: 0},
				Computed: true,
			},
			&ast.Identifier{Name: "sma50"},
		},
	}

	gen := &generator{
		imports:        make(map[string]bool),
		variables:      make(map[string]string),
		strategyConfig: NewStrategyConfig(),
		taRegistry:     NewTAFunctionRegistry(),
		builtinHandler: NewBuiltinIdentifierHandler(),
	}

	code, err := gen.generateVariableFromCall("shortCross", call)
	if err != nil {
		t.Fatalf("generateVariableFromCall failed: %v", err)
	}

	t.Logf("Generated code:\n%s", code)

	if !strings.Contains(code, "shortCrossSeries.Set(0.0)") {
		t.Error("Missing initial Series.Set(0.0) assignment")
	}
	if !strings.Contains(code, "if i > 0") {
		t.Error("Missing warmup check")
	}
	// sma50 is an Identifier (not MemberExpression), so it uses GetCurrent()
	if !strings.Contains(code, "bar.Close < sma50Series.GetCurrent()") && !strings.Contains(code, "bar.Close < sma50Series.Get(0)") {
		t.Error("Missing crossunder condition (current below)")
	}
	if !strings.Contains(code, ">=") {
		t.Error("Missing previous >= operator for crossunder")
	}
	if !strings.Contains(code, "shortCrossSeries.Set(func() float64") {
		t.Error("Missing Series.Set with bool→float64 conversion")
	}
}

func TestCrossoverWithArithmetic(t *testing.T) {
	call := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "ta"},
			Property: &ast.Identifier{Name: "crossover"},
		},
		Arguments: []ast.Expression{
			&ast.MemberExpression{
				Object:   &ast.Identifier{Name: "close"},
				Property: &ast.Literal{Value: 0},
				Computed: true,
			},
			&ast.BinaryExpression{
				Operator: "*",
				Left:     &ast.Identifier{Name: "sma20"},
				Right:    &ast.Literal{Value: 1.02},
			},
		},
	}

	gen := &generator{
		imports:        make(map[string]bool),
		variables:      make(map[string]string),
		strategyConfig: NewStrategyConfig(),
		taRegistry:     NewTAFunctionRegistry(),
		builtinHandler: NewBuiltinIdentifierHandler(),
	}

	code, err := gen.generateVariableFromCall("crossAboveThreshold", call)
	if err != nil {
		t.Fatalf("generateVariableFromCall failed: %v", err)
	}

	t.Logf("Generated code:\n%s", code)

	if !strings.Contains(code, "(sma20Series.GetCurrent() * 1.02)") {
		t.Error("Missing arithmetic expression in crossover")
	}
	if !strings.Contains(code, "bar.Close > (sma20Series.GetCurrent() * 1.02)") {
		t.Error("Missing arithmetic comparison")
	}
}

func TestBooleanTypeTracking(t *testing.T) {
	program := &ast.Program{
		Body: []ast.Node{
			&ast.VariableDeclaration{
				Declarations: []ast.VariableDeclarator{
					{
						ID: &ast.Identifier{Name: "longCross"},
						Init: &ast.CallExpression{
							Callee: &ast.MemberExpression{
								Object:   &ast.Identifier{Name: "ta"},
								Property: &ast.Identifier{Name: "crossover"},
							},
							Arguments: []ast.Expression{
								&ast.Identifier{Name: "close"},
								&ast.Identifier{Name: "sma20"},
							},
						},
					},
				},
			},
			&ast.VariableDeclaration{
				Declarations: []ast.VariableDeclarator{
					{
						ID: &ast.Identifier{Name: "sma50"},
						Init: &ast.CallExpression{
							Callee: &ast.MemberExpression{
								Object:   &ast.Identifier{Name: "ta"},
								Property: &ast.Identifier{Name: "sma"},
							},
							Arguments: []ast.Expression{
								&ast.Identifier{Name: "close"},
								&ast.Literal{Value: 50.0},
							},
						},
					},
				},
			},
		},
	}

	gen := &generator{
		imports:           make(map[string]bool),
		variables:         make(map[string]string),
		varInits:          make(map[string]ast.Expression),
		constants:         make(map[string]interface{}),
		strategyConfig:    NewStrategyConfig(),
		taRegistry:        NewTAFunctionRegistry(),
		typeSystem:        NewTypeInferenceEngine(),
		boolConverter:     NewBooleanConverter(NewTypeInferenceEngine()),
		constantRegistry:  NewConstantRegistry(),
		runtimeOnlyFilter: NewRuntimeOnlyFunctionFilter(),
		constEvaluator:    validation.NewWarmupAnalyzer(),
	}
	gen.tempVarMgr = NewTempVariableManager(gen)
	gen.exprAnalyzer = NewExpressionAnalyzer(gen)
	gen.statementAnalyzer = NewStatementConditionalAnalyzer(gen)
	gen.conditionalArgAnalyzer = NewConditionalArgumentAnalyzer(&ExpressionHasher{})

	code, err := gen.generateProgram(program)
	if err != nil {
		t.Fatalf("generateProgram failed: %v", err)
	}

	if !strings.Contains(code, "var longCrossSeries *series.Series") {
		t.Error("longCross should be declared as *series.Series")
	}
	if !strings.Contains(code, "var sma50Series *series.Series") {
		t.Error("sma50 should be declared as *series.Series")
	}
	if gen.variables["longCross"] != "bool" {
		t.Errorf("longCross should be tracked as bool type, got: %s", gen.variables["longCross"])
	}
	if gen.variables["sma50"] != "float64" {
		t.Errorf("sma50 should be tracked as float64 type, got: %s", gen.variables["sma50"])
	}
}
