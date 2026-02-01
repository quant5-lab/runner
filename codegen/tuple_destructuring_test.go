package codegen

import (
	"fmt"
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/runtime/validation"
)

/* TestTupleDestructuringUnimplementedFunctions verifies graceful degradation
 * when multi-value functions are not yet implemented.
 * Generalized test: covers ANY number of return values, ANY function name.
 */
func TestTupleDestructuringUnimplementedFunctions(t *testing.T) {
	tests := []struct {
		name          string
		varNames      []string
		funcName      string
		expectedCount int
		description   string
	}{
		{
			name:          "two-value function",
			varNames:      []string{"v1", "v2"},
			funcName:      "hypothetical_two_value_indicator",
			expectedCount: 2,
			description:   "unimplemented indicator: returns 2 values",
		},
		{
			name:          "three-value function",
			varNames:      []string{"v1", "v2", "v3"},
			funcName:      "hypothetical_three_value_indicator",
			expectedCount: 3,
			description:   "unimplemented indicator: returns 3 values",
		},
		{
			name:          "five-value function",
			varNames:      []string{"v1", "v2", "v3", "v4", "v5"},
			funcName:      "multi_output_indicator",
			expectedCount: 5,
			description:   "complex indicator: returns 5 values",
		},
		{
			name:          "single-value in tuple",
			varNames:      []string{"result"},
			funcName:      "custom_indicator",
			expectedCount: 1,
			description:   "edge case: tuple with one variable",
		},
		{
			name:          "arbitrary new function",
			varNames:      []string{"a", "b", "c", "d", "e", "f", "g"},
			funcName:      "hypothetical_future_indicator",
			expectedCount: 7,
			description:   "future-proof: handles arbitrary new functions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := newTestGeneratorForTupleTests()

			/* Build ArrayPattern from variable names */
			elements := make([]ast.Identifier, len(tt.varNames))
			for i, name := range tt.varNames {
				elements[i] = ast.Identifier{Name: name}
			}

			declarator := ast.VariableDeclarator{
				ID: &ast.ArrayPattern{Elements: elements},
				Init: &ast.CallExpression{
					Callee: &ast.Identifier{Name: tt.funcName},
					Arguments: []ast.Expression{
						&ast.Identifier{Name: "close"},
						&ast.Literal{Value: 14.0},
					},
				},
			}

			code, err := gen.generateTupleDestructuringDeclaration(declarator)
			if err != nil {
				t.Fatalf("generateTupleDestructuringDeclaration() error: %v", err)
			}

			/* Verify: TODO comment for unimplemented function */
			expectedComment := tt.funcName + "() - TODO: implement"
			if !strings.Contains(code, expectedComment) {
				t.Errorf("Missing TODO comment. Expected %q in:\n%s", expectedComment, code)
			}

			/* Verify: Exactly N placeholder Series.Set(0.0) calls */
			for _, varName := range tt.varNames {
				expectedSet := varName + "Series.Set(0.0)"
				if !strings.Contains(code, expectedSet) {
					t.Errorf("Missing placeholder for %q. Expected %q in:\n%s", varName, expectedSet, code)
				}
			}

			/* Verify: No invalid syntax (no incomplete := statements) */
			invalidPatterns := []string{
				":= // ",
				":= /*",
				", , ",
			}
			for _, invalid := range invalidPatterns {
				if strings.Contains(code, invalid) {
					t.Errorf("Invalid syntax found: %q in:\n%s", invalid, code)
				}
			}

			/* Verify: Count matches expected */
			count := strings.Count(code, "Series.Set(0.0)")
			if count != tt.expectedCount {
				t.Errorf("Expected %d Series.Set(0.0) calls, got %d in:\n%s", tt.expectedCount, count, code)
			}
		})
	}
}

/* TestTupleDestructuringImplementedFunctions verifies normal codegen
 * when functions ARE implemented (should NOT generate placeholders).
 */
func TestTupleDestructuringImplementedFunctions(t *testing.T) {
	tests := []struct {
		name        string
		varNames    []string
		funcName    string
		description string
	}{
		{
			name:        "user-defined function",
			varNames:    []string{"result1", "result2"},
			funcName:    "custom_calc",
			description: "user-defined functions use arrow context",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := newTestGeneratorForTupleTests()

			/* Register as user-defined function */
			gen.variables[tt.funcName] = "function"

			elements := make([]ast.Identifier, len(tt.varNames))
			for i, name := range tt.varNames {
				elements[i] = ast.Identifier{Name: name}
			}

			declarator := ast.VariableDeclarator{
				ID: &ast.ArrayPattern{Elements: elements},
				Init: &ast.CallExpression{
					Callee: &ast.Identifier{Name: tt.funcName},
					Arguments: []ast.Expression{
						&ast.Identifier{Name: "x"},
					},
				},
			}

			code, err := gen.generateTupleDestructuringDeclaration(declarator)
			if err != nil {
				t.Fatalf("generateTupleDestructuringDeclaration() error: %v", err)
			}

			/* Verify: NO placeholder generation for implemented functions */
			if strings.Contains(code, "Series.Set(0.0)") {
				t.Errorf("Should NOT generate placeholders for implemented function:\n%s", code)
			}

			/* Verify: Uses arrow context for user-defined functions */
			if !strings.Contains(code, "arrowCtx_") {
				t.Errorf("Should use arrow context for user-defined function:\n%s", code)
			}
		})
	}
}

/* TestTupleDestructuringEdgeCases covers boundary conditions */
func TestTupleDestructuringEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		setup       func() (*generator, ast.VariableDeclarator)
		wantErr     bool
		description string
	}{
		{
			name: "empty tuple",
			setup: func() (*generator, ast.VariableDeclarator) {
				gen := newTestGeneratorForTupleTests()
				declarator := ast.VariableDeclarator{
					ID: &ast.ArrayPattern{Elements: []ast.Identifier{}},
					Init: &ast.CallExpression{
						Callee: &ast.Identifier{Name: "empty_func"},
					},
				}
				return gen, declarator
			},
			wantErr:     true,
			description: "empty tuple pattern should error",
		},
		{
			name: "non-call init",
			setup: func() (*generator, ast.VariableDeclarator) {
				gen := newTestGeneratorForTupleTests()
				declarator := ast.VariableDeclarator{
					ID:   &ast.ArrayPattern{Elements: []ast.Identifier{{Name: "a"}, {Name: "b"}}},
					Init: &ast.Literal{Value: 42.0},
				}
				return gen, declarator
			},
			wantErr:     true,
			description: "tuple init must be CallExpression",
		},
		{
			name: "non-array-pattern",
			setup: func() (*generator, ast.VariableDeclarator) {
				gen := newTestGeneratorForTupleTests()
				declarator := ast.VariableDeclarator{
					ID: &ast.Identifier{Name: "single"},
					Init: &ast.CallExpression{
						Callee: &ast.Identifier{Name: "func"},
					},
				}
				return gen, declarator
			},
			wantErr:     true,
			description: "single variable not tuple pattern",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen, declarator := tt.setup()

			_, err := gen.generateTupleDestructuringDeclaration(declarator)

			if tt.wantErr && err == nil {
				t.Errorf("Expected error for %s, got nil", tt.description)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error for %s: %v", tt.description, err)
			}
		})
	}
}

/* TestTupleDestructuringCompilability verifies generated code compiles */
func TestTupleDestructuringCompilability(t *testing.T) {
	tests := []struct {
		name        string
		varNames    []string
		funcName    string
		description string
	}{
		{
			name:        "unimplemented three value indicator",
			varNames:    []string{"result1", "result2", "result3"},
			funcName:    "hypothetical_three_return_indicator",
			description: "unimplemented indicator with 3 return values",
		},
		{
			name:        "large tuple",
			varNames:    []string{"a", "b", "c", "d", "e", "f", "g", "h"},
			funcName:    "complex_indicator",
			description: "stress test: 8 return values",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			/* Create full program with tuple destructuring */
			elements := make([]ast.Identifier, len(tt.varNames))
			for i, name := range tt.varNames {
				elements[i] = ast.Identifier{Name: name}
			}

			program := &ast.Program{
				Body: []ast.Node{
					&ast.VariableDeclaration{
						Declarations: []ast.VariableDeclarator{
							{
								ID: &ast.ArrayPattern{Elements: elements},
								Init: &ast.CallExpression{
									Callee: &ast.Identifier{Name: tt.funcName},
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

			gen := newTestGeneratorForTupleTests()
			code, err := gen.generateProgram(program)
			if err != nil {
				t.Fatalf("generateProgram() error: %v", err)
			}

			/* Verify: All variables declared as Series */
			for _, varName := range tt.varNames {
				expectedDecl := "var " + varName + "Series *series.Series"
				if !strings.Contains(code, expectedDecl) {
					t.Errorf("Missing Series declaration for %q in:\n%s", varName, code)
				}
			}

			/* Verify: All variables initialized */
			for _, varName := range tt.varNames {
				expectedInit := varName + "Series = series.NewSeries(len(ctx.Data))"
				if !strings.Contains(code, expectedInit) {
					t.Errorf("Missing Series initialization for %q in:\n%s", varName, code)
				}
			}

			/* Verify: Placeholder generation in bar loop */
			for _, varName := range tt.varNames {
				expectedSet := varName + "Series.Set(0.0)"
				if !strings.Contains(code, expectedSet) {
					t.Errorf("Missing placeholder for %q in bar loop:\n%s", varName, code)
				}
			}

			/* Verify: No syntax errors in generated code */
			if strings.Contains(code, ":= // ") || strings.Contains(code, ", := ") {
				t.Errorf("Invalid Go syntax in generated code:\n%s", code)
			}
		})
	}
}

/* TestTupleDestructuringAST_SourceOfTruth verifies AST-driven placeholder count */
func TestTupleDestructuringAST_SourceOfTruth(t *testing.T) {
	/* Key architectural principle: AST tells us return count, not a registry */
	tests := []struct {
		name         string
		elementCount int
		description  string
	}{
		{"single", 1, "single element tuple"},
		{"two", 2, "two elements"},
		{"three", 3, "three elements"},
		{"five", 5, "five elements"},
		{"ten", 10, "ten elements (stress test)"},
		{"twenty", 20, "twenty elements (extreme stress test)"},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			gen := newTestGeneratorForTupleTests()

			/* Create tuple with N elements */
			elements := make([]ast.Identifier, tt.elementCount)
			varNames := make([]string, tt.elementCount)
			for i := 0; i < tt.elementCount; i++ {
				varNames[i] = fmt.Sprintf("var%d", i)
				elements[i] = ast.Identifier{Name: varNames[i]}
			}

			declarator := ast.VariableDeclarator{
				ID: &ast.ArrayPattern{Elements: elements},
				Init: &ast.CallExpression{
					Callee:    &ast.Identifier{Name: "never_seen_before_func"},
					Arguments: []ast.Expression{},
				},
			}

			code, err := gen.generateTupleDestructuringDeclaration(declarator)
			if err != nil {
				t.Fatalf("generateTupleDestructuringDeclaration() error: %v", err)
			}

			/* Verify: Exactly N placeholders generated */
			count := strings.Count(code, "Series.Set(0.0)")
			if count != tt.elementCount {
				t.Errorf("Expected %d placeholders for %d-element tuple, got %d",
					tt.elementCount, tt.elementCount, count)
			}

			/* Verify: All variable names present */
			for _, varName := range varNames {
				if !strings.Contains(code, varName+"Series.Set(0.0)") {
					t.Errorf("Missing placeholder for %q", varName)
				}
			}
		})
	}
}

/* Helper: Creates test generator with minimal dependencies */
func newTestGeneratorForTupleTests() *generator {
	gen := &generator{
		imports:          make(map[string]bool),
		variables:        make(map[string]string),
		varInits:         make(map[string]ast.Expression),
		constants:        make(map[string]interface{}),
		reassignedVars:   make(map[string]bool),
		strategyConfig:   NewStrategyConfig(),
		taRegistry:       NewTAFunctionRegistry(),
		typeSystem:       NewTypeInferenceEngine(),
		boolConverter:    NewBooleanConverter(NewTypeInferenceEngine()),
		constantRegistry: NewConstantRegistry(),
		constEvaluator:   validation.NewWarmupAnalyzer(),
		builtinHandler:   NewBuiltinIdentifierHandler(),
	}
	gen.callRouter = NewCallExpressionRouter()
	gen.tempVarMgr = NewTempVariableManager(gen)
	gen.exprAnalyzer = NewExpressionAnalyzer(gen)
	gen.arrowContextLifecycle = NewArrowContextLifecycleManager()
	gen.returnValueStorage = NewReturnValueSeriesStorageHandler("\t")
	gen.barFieldRegistry = NewBarFieldSeriesRegistry()
	gen.plotExprHandler = NewPlotExpressionHandler(gen)
	gen.literalFormatter = NewLiteralFormatter()
	gen.runtimeOnlyFilter = NewRuntimeOnlyFunctionFilter()
	gen.plotCollector = NewPlotCollector()
	gen.mathHandler = NewMathHandler()
	gen.tupleIndicatorHandler = NewTupleIndicatorHandler()
	gen.directionExtractor = NewDefaultDirectionExtractor()
	return gen
}
