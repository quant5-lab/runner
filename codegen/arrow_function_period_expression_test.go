package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

/* TestArrowFunctionTACall_PeriodExpressionExtraction verifies period parameter extraction */
func TestArrowFunctionTACall_PeriodExpressionExtraction(t *testing.T) {
	tests := []struct {
		name           string
		call           *ast.CallExpression
		arrowParams    []string
		expectError    bool
		expectedType   string
		expectedValue  int
		expectedGoExpr string // Expected Go code expression
		description    string
	}{
		{
			name: "Literal integer period",
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "ta.rma"},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
					&ast.Literal{Value: 14.0},
				},
			},
			arrowParams:    []string{},
			expectError:    false,
			expectedType:   "constant",
			expectedValue:  14,
			expectedGoExpr: "14",
			description:    "Literal periods should create ConstantPeriod",
		},
		{
			name: "Arrow parameter period",
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "ta.rma"},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "src"},
					&ast.Identifier{Name: "len"},
				},
			},
			arrowParams:    []string{"src", "len"},
			expectError:    false,
			expectedType:   "runtime",
			expectedValue:  -1,
			expectedGoExpr: "len",
			description:    "Arrow parameters should create RuntimePeriod with variable name",
		},
		{
			name: "Float literal period",
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "ta.ema"},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
					&ast.Literal{Value: 20.0},
				},
			},
			arrowParams:    []string{},
			expectError:    false,
			expectedType:   "constant",
			expectedValue:  20,
			expectedGoExpr: "20",
			description:    "Float literals should be converted to integer constants",
		},
		{
			name: "Integer literal period",
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "ta.sma"},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "high"},
					&ast.Literal{Value: int(50)},
				},
			},
			arrowParams:    []string{},
			expectError:    false,
			expectedType:   "constant",
			expectedValue:  50,
			expectedGoExpr: "50",
			description:    "Integer literals should create ConstantPeriod",
		},
		{
			name: "Non-arrow-param identifier with period variable",
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "ta.rma"},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
					&ast.Identifier{Name: "myPeriod"},
				},
			},
			arrowParams:    []string{"myPeriod"},
			expectError:    false,
			expectedType:   "runtime",
			expectedValue:  -1,
			expectedGoExpr: "myPeriod",
			description:    "Identifier matching arrow param should create RuntimePeriod",
		},
		{
			name: "Unknown identifier - should error",
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "ta.rma"},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
					&ast.Identifier{Name: "unknownVar"},
				},
			},
			arrowParams: []string{"len"},
			expectError: true,
			description: "Unknown identifiers (not arrow parameters) should error during extraction",
		},
		{
			name: "Binary expression period",
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "ta.wma"},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
					&ast.BinaryExpression{
						Left:     &ast.Identifier{Name: "len"},
						Operator: "/",
						Right:    &ast.Literal{Value: 2.0},
					},
				},
			},
			arrowParams:   []string{"len"},
			expectError:   false,
			expectedType:  "computed",
			expectedValue: -1,
			description:   "Binary expressions should create ComputedPeriod with rendered Go code",
		},
		{
			name: "Nested call expression period",
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "ta.wma"},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
					&ast.CallExpression{
						Callee: &ast.Identifier{Name: "math.round"},
						Arguments: []ast.Expression{
							&ast.CallExpression{
								Callee:    &ast.Identifier{Name: "math.sqrt"},
								Arguments: []ast.Expression{&ast.Identifier{Name: "len"}},
							},
						},
					},
				},
			},
			arrowParams:   []string{"len"},
			expectError:   false,
			expectedType:  "computed",
			expectedValue: -1,
			description:   "Nested call expressions should create ComputedPeriod",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newTestGenerator()
			for _, param := range tt.arrowParams {
				g.variables[param] = "float"
			}

			gen := newTestArrowTAGenerator(g)
			funcName := tt.call.Callee.(*ast.Identifier).Name

			_, period, err := gen.extractTAArguments(funcName, tt.call)

			if tt.expectError {
				if err == nil {
					t.Errorf("%s: expected error, got nil", tt.description)
				}
				return
			}

			if err != nil {
				t.Fatalf("%s: unexpected error: %v", tt.description, err)
			}

			if period == nil {
				t.Fatalf("%s: period is nil", tt.description)
			}

			isConstant := period.IsConstant()
			expectedConstant := (tt.expectedType == "constant")
			if isConstant != expectedConstant {
				t.Errorf("%s: IsConstant() = %v, want %v", tt.description, isConstant, expectedConstant)
			}

			actualValue := period.AsInt()
			if actualValue != tt.expectedValue {
				t.Errorf("%s: AsInt() = %d, want %d", tt.description, actualValue, tt.expectedValue)
			}

			actualExpr := period.AsGoExpr()
			if tt.expectedGoExpr != "" {
				if actualExpr != tt.expectedGoExpr {
					t.Errorf("%s: AsGoExpr() = %q, want %q", tt.description, actualExpr, tt.expectedGoExpr)
				}
			}

			if !isConstant {
				if actualExpr == "" {
					t.Errorf("%s: non-constant period has empty GoExpr", tt.description)
				}

				intCast := period.AsIntCast()
				expectedIntCast := "int(" + actualExpr + ")"
				if intCast != expectedIntCast {
					t.Errorf("%s: AsIntCast() = %q, want %q", tt.description, intCast, expectedIntCast)
				}

				floatCast := period.AsFloat64Cast()
				expectedFloatCast := "float64(" + actualExpr + ")"
				if floatCast != expectedFloatCast {
					t.Errorf("%s: AsFloat64Cast() = %q, want %q", tt.description, floatCast, expectedFloatCast)
				}

				if tt.expectedType == "computed" {
					seriesPart := period.AsSeriesNamePart()
					if seriesPart != "computed" {
						t.Errorf("%s: AsSeriesNamePart() = %q, want %q", tt.description, seriesPart, "computed")
					}
				}
			}
		})
	}
}

/* TestArrowFunctionTACall_PeriodExpressionInGeneratedCode verifies end-to-end code patterns */
func TestArrowFunctionTACall_PeriodExpressionInGeneratedCode(t *testing.T) {
	tests := []struct {
		name           string
		call           *ast.CallExpression
		arrowParams    []string
		mustContain    []string
		mustNotContain []string
		description    string
	}{
		{
			name: "Constant period RMA uses literals",
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "ta.rma"},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
					&ast.Literal{Value: 14.0},
				},
			},
			arrowParams: []string{},
			mustContain: []string{
				"for j := 0; j < 14",
				"alpha := 1.0 / float64(14)",
				"_rma_14_",
			},
			mustNotContain: []string{
				"int(14)",
				"for j := 0; j < 20",
			},
			description: "Constant periods should generate optimized literal code",
		},
		{
			name: "Runtime period RMA uses variable with casts",
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "ta.rma"},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
					&ast.Identifier{Name: "len"},
				},
			},
			arrowParams: []string{"len"},
			mustContain: []string{
				"int(len)",
				"float64(len)",
				"_rma_runtime_",
			},
			mustNotContain: []string{
				"for j := 0; j < 20",
				"alpha := 1.0 / float64(20)",
				"_rma_20_",
			},
			description: "Runtime periods must use parameter variable, never hardcoded fallback",
		},
		{
			name: "Constant EMA alpha formula",
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "ta.ema"},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
					&ast.Literal{Value: 20.0},
				},
			},
			arrowParams: []string{},
			mustContain: []string{
				"2.0 / float64(20+1)",
			},
			mustNotContain: []string{
				"(float64(20)+1)",
			},
			description: "EMA alpha calculation optimizes for constants",
		},
		{
			name: "Runtime EMA uses correct alpha formula",
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "ta.ema"},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
					&ast.Identifier{Name: "len"},
				},
			},
			arrowParams: []string{"len"},
			mustContain: []string{
				"(float64(len)+1)",
			},
			mustNotContain: []string{
				"(float64(20)+1)",
			},
			description: "EMA alpha calculation must use runtime parameter",
		},
		{
			name: "Computed period WMA with binary expression",
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "ta.wma"},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
					&ast.BinaryExpression{
						Left:     &ast.Identifier{Name: "len"},
						Operator: "/",
						Right:    &ast.Literal{Value: 2.0},
					},
				},
			},
			arrowParams: []string{"len"},
			mustContain: []string{
				"lenSeries.GetCurrent() / 2",
			},
			mustNotContain: []string{
				"for j := 0; j < 20",
				"_wma_20_",
			},
			description: "Computed binary period must use rendered expression, not hardcoded fallback",
		},
		{
			name: "Computed period SMA with call expression",
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "ta.sma"},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
					&ast.CallExpression{
						Callee: &ast.Identifier{Name: "math.round"},
						Arguments: []ast.Expression{
							&ast.Identifier{Name: "len"},
						},
					},
				},
			},
			arrowParams: []string{"len"},
			mustContain: []string{
				"math.Round(lenSeries.GetCurrent())",
			},
			mustNotContain: []string{
				"for j := 0; j < 20",
				"_sma_20_",
			},
			description: "Computed call period must use rendered expression, not hardcoded fallback",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newTestGenerator()
			for _, param := range tt.arrowParams {
				g.variables[param] = "float"
			}

			gen := newTestArrowTAGenerator(g)

			code, err := gen.Generate(tt.call)
			if err != nil {
				t.Fatalf("%s: code generation error: %v", tt.description, err)
			}

			for _, pattern := range tt.mustContain {
				if !strings.Contains(code, pattern) {
					t.Errorf("%s: generated code missing required pattern %q", tt.description, pattern)
					t.Logf("Generated code:\n%s", code)
				}
			}

			for _, pattern := range tt.mustNotContain {
				if strings.Contains(code, pattern) {
					t.Errorf("%s: generated code contains prohibited pattern %q", tt.description, pattern)
					t.Logf("Generated code:\n%s", code)
				}
			}
		})
	}
}

/* TestArrowFunctionTACall_PeriodExpressionEdgeCases tests boundary conditions */
func TestArrowFunctionTACall_PeriodExpressionEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		call        *ast.CallExpression
		arrowParams []string
		validate    func(t *testing.T, period PeriodExpression, code string)
		description string
	}{
		{
			name: "Period value 1 (minimum)",
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "ta.rma"},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
					&ast.Literal{Value: 1.0},
				},
			},
			arrowParams: []string{},
			validate: func(t *testing.T, period PeriodExpression, code string) {
				if !period.IsConstant() {
					t.Error("Period 1 should be constant")
				}
				if period.AsInt() != 1 {
					t.Errorf("Period should be 1, got %d", period.AsInt())
				}
				if !strings.Contains(code, "for j := 0; j < 1") {
					t.Error("Loop should use literal 1")
				}
			},
			description: "Minimum period value should work correctly",
		},
		{
			name: "Very large constant period",
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "ta.sma"},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
					&ast.Literal{Value: 1000.0},
				},
			},
			arrowParams: []string{},
			validate: func(t *testing.T, period PeriodExpression, code string) {
				if period.AsInt() != 1000 {
					t.Errorf("Period should be 1000, got %d", period.AsInt())
				}
				if !strings.Contains(code, "1000") {
					t.Error("Large period should appear in generated code")
				}
			},
			description: "Large periods should be handled without overflow",
		},
		{
			name: "Parameter named 'length' (common alternative to 'len')",
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "ta.rma"},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "src"},
					&ast.Identifier{Name: "length"},
				},
			},
			arrowParams: []string{"src", "length"},
			validate: func(t *testing.T, period PeriodExpression, code string) {
				if period.IsConstant() {
					t.Error("Parameter 'length' should create RuntimePeriod")
				}
				if period.AsGoExpr() != "length" {
					t.Errorf("Variable name should be 'length', got %q", period.AsGoExpr())
				}
				if !strings.Contains(code, "int(length)") {
					t.Error("Generated code should use int(length)")
				}
			},
			description: "Alternative parameter names should work correctly",
		},
		{
			name: "Parameter with underscore naming",
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "ta.ema"},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
					&ast.Identifier{Name: "my_period"},
				},
			},
			arrowParams: []string{"my_period"},
			validate: func(t *testing.T, period PeriodExpression, code string) {
				if period.AsGoExpr() != "my_period" {
					t.Errorf("Variable name should be 'my_period', got %q", period.AsGoExpr())
				}
				if !strings.Contains(code, "float64(my_period)") {
					t.Error("Generated code should use float64(my_period)")
				}
			},
			description: "Underscore in parameter names should be preserved",
		},
		{
			name: "Computed period preserves binary structure",
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "ta.wma"},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
					&ast.BinaryExpression{
						Left:     &ast.Identifier{Name: "n"},
						Operator: "/",
						Right:    &ast.Literal{Value: 2.0},
					},
				},
			},
			arrowParams: []string{"n"},
			validate: func(t *testing.T, period PeriodExpression, code string) {
				if period.IsConstant() {
					t.Error("Binary expression period must not be constant")
				}
				if period.AsInt() != -1 {
					t.Errorf("AsInt() = %d, want -1", period.AsInt())
				}
				if period.AsSeriesNamePart() != "computed" {
					t.Errorf("Series name part = %q, want %q", period.AsSeriesNamePart(), "computed")
				}
			},
			description: "Binary expression period should create ComputedPeriod with correct properties",
		},
		{
			name: "Computed period with nested function calls",
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "ta.wma"},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
					&ast.CallExpression{
						Callee: &ast.Identifier{Name: "math.round"},
						Arguments: []ast.Expression{
							&ast.CallExpression{
								Callee:    &ast.Identifier{Name: "math.sqrt"},
								Arguments: []ast.Expression{&ast.Identifier{Name: "n"}},
							},
						},
					},
				},
			},
			arrowParams: []string{"n"},
			validate: func(t *testing.T, period PeriodExpression, code string) {
				if period.IsConstant() {
					t.Error("Call expression period must not be constant")
				}
				if period.AsSeriesNamePart() != "computed" {
					t.Errorf("Series name part = %q, want %q", period.AsSeriesNamePart(), "computed")
				}
				goExpr := period.AsGoExpr()
				if goExpr == "" {
					t.Error("ComputedPeriod Go expression must not be empty")
				}
			},
			description: "Nested call expression period should create ComputedPeriod",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newTestGenerator()
			for _, param := range tt.arrowParams {
				g.variables[param] = "float"
			}

			gen := newTestArrowTAGenerator(g)
			funcName := tt.call.Callee.(*ast.Identifier).Name

			_, period, err := gen.extractTAArguments(funcName, tt.call)
			if err != nil {
				t.Fatalf("%s: unexpected error: %v", tt.description, err)
			}

			code, err := gen.Generate(tt.call)
			if err != nil {
				t.Fatalf("%s: code generation error: %v", tt.description, err)
			}

			tt.validate(t, period, code)
		})
	}
}

/* TestArrowFunctionTACall_NoHardcodedFallbacks guards against hardcoded period regressions across all period types */
func TestArrowFunctionTACall_NoHardcodedFallbacks(t *testing.T) {
	tests := []struct {
		name        string
		call        *ast.CallExpression
		arrowParams []string
		periodType  string
		description string
	}{
		{
			name: "RMA with runtime identifier period",
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "ta.rma"},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "src"},
					&ast.Identifier{Name: "len"},
				},
			},
			arrowParams: []string{"src", "len"},
			periodType:  "runtime",
			description: "RMA with runtime period",
		},
		{
			name: "EMA with runtime identifier period",
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "ta.ema"},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
					&ast.Identifier{Name: "period"},
				},
			},
			arrowParams: []string{"period"},
			periodType:  "runtime",
			description: "EMA with runtime period",
		},
		{
			name: "SMA with runtime identifier period",
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "ta.sma"},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "high"},
					&ast.Identifier{Name: "length"},
				},
			},
			arrowParams: []string{"length"},
			periodType:  "runtime",
			description: "SMA with runtime period",
		},
		{
			name: "WMA with computed binary period",
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "ta.wma"},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
					&ast.BinaryExpression{
						Left:     &ast.Identifier{Name: "n"},
						Operator: "/",
						Right:    &ast.Literal{Value: 2.0},
					},
				},
			},
			arrowParams: []string{"n"},
			periodType:  "computed",
			description: "WMA with computed binary period",
		},
		{
			name: "SMA with computed call period",
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "ta.sma"},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
					&ast.CallExpression{
						Callee: &ast.Identifier{Name: "math.round"},
						Arguments: []ast.Expression{
							&ast.Identifier{Name: "n"},
						},
					},
				},
			},
			arrowParams: []string{"n"},
			periodType:  "computed",
			description: "SMA with computed call period",
		},
		{
			name: "EMA with computed multiply period",
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "ta.ema"},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
					&ast.BinaryExpression{
						Left:     &ast.Identifier{Name: "n"},
						Operator: "*",
						Right:    &ast.Literal{Value: 2.0},
					},
				},
			},
			arrowParams: []string{"n"},
			periodType:  "computed",
			description: "EMA with computed multiply period",
		},
	}

	prohibitedPatterns := []string{
		"for j := 0; j < 20",
		"alpha := 1.0 / float64(20)",
		"2.0 / float64(20+1)",
		"_rma_20_",
		"_ema_20_",
		"_sma_20_",
		"_wma_20_",
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newTestGenerator()

			for _, param := range tt.arrowParams {
				g.variables[param] = "float"
			}

			gen := newTestArrowTAGenerator(g)

			code, err := gen.Generate(tt.call)
			if err != nil {
				t.Fatalf("%s: code generation error: %v", tt.description, err)
			}

			for _, prohibited := range prohibitedPatterns {
				if strings.Contains(code, prohibited) {
					t.Errorf("%s: REGRESSION - hardcoded pattern %q found", tt.description, prohibited)
					t.Logf("Generated code:\n%s", code)
				}
			}

			switch tt.periodType {
			case "runtime":
				paramName := tt.arrowParams[len(tt.arrowParams)-1]
				if !strings.Contains(code, "int("+paramName+")") {
					t.Errorf("%s: missing int(%s) cast in generated code", tt.description, paramName)
					t.Logf("Generated code:\n%s", code)
				}
			case "computed":
				/* Stateful IIFEs (EMA/RMA) use series names; stateless IIFEs (SMA/WMA) inline the expression */
				paramName := tt.arrowParams[len(tt.arrowParams)-1]
				dynamicRef := paramName + "Series.GetCurrent()"
				if !strings.Contains(code, dynamicRef) {
					t.Errorf("%s: missing dynamic parameter reference %q in generated code", tt.description, dynamicRef)
					t.Logf("Generated code:\n%s", code)
				}
			}
		})
	}
}
