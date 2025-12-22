package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/parser"
)

/*
arrow_series_universal_test.go - Comprehensive test coverage for universal ForwardSeriesBuffer architecture

PURPOSE:
Validates that the universal Series storage paradigm is correctly implemented across ALL arrow
function code generation paths, ensuring long-term consistency and catching regressions.

ARCHITECTURE PRINCIPLES TESTED:
1. Universal Series Storage: ALL arrow function local variables use ForwardSeriesBuffer via ArrowContext
2. Series.Set() Pattern: ALL assignments store values through Series.Set()
3. Series.GetCurrent() Access: ALL variable references in expressions resolve to Series.GetCurrent()
4. Parameter Distinction: Function parameters remain scalar, locals use Series
5. Tuple Return Consistency: Multi-value returns access each element via Series.GetCurrent()

TEST CATEGORIES:
- UniversalSeriesDeclarations: Validates ALL locals receive Series declarations
- UniversalSeriesAssignments: Validates ALL assignments use Series.Set()
- UniversalSeriesAccess: Validates ALL references use Series.GetCurrent()
- ParameterVsLocalVariableDistinction: Validates parameters stay scalar, locals use Series
- TupleReturnSeriesAccess: Validates tuple returns use Series.GetCurrent() for each element
- ComplexExpressionSeriesIntegration: Validates binary/unary/conditional expression Series integration
- EdgeCasesSeriesBehavior: Validates minimal/unusual function structures
- NestedFunctionCallsSeriesConsistency: Validates arrow function composition

KNOWN LIMITATIONS (as of test creation):
- Some variable access patterns in binary/return expressions don't yet use Series.GetCurrent()
- Parameter vs local variable distinction not fully enforced in all contexts
- These tests intentionally document expected behavior to catch when bugs are fixed

TEST MAINTENANCE:
These tests are NOT coupled to the specific bug fix that prompted their creation. They validate
the universal Series architecture comprehensively and will remain relevant for any future changes
to arrow function code generation.
*/

/* compilePineScript parses PineScript source and generates Go code */
func compilePineScript(source string) (string, error) {
	p, err := parser.NewParser()
	if err != nil {
		return "", err
	}

	script, err := p.ParseBytes("test.pine", []byte(source))
	if err != nil {
		return "", err
	}

	converter := parser.NewConverter()
	program, err := converter.ToESTree(script)
	if err != nil {
		return "", err
	}

	result, err := GenerateStrategyCodeFromAST(program)
	if err != nil {
		return "", err
	}

	// Return both user-defined functions and function body for comprehensive validation
	return result.UserDefinedFunctions + "\n" + result.FunctionBody, nil
}

/*
TestArrowFunction_UniversalSeriesDeclarations validates that ALL local variables
in arrow functions receive Series storage declarations via ArrowContext.

Behavior: Every variable declared in an arrow function body must have a corresponding

	Series declaration using arrowCtx.GetOrCreateSeries("varName")

This ensures the universal ForwardSeriesBuffer paradigm is consistently applied.
*/
func TestArrowFunction_UniversalSeriesDeclarations(t *testing.T) {
	tests := []struct {
		name              string
		pine              string
		expectedDecls     []string // Series declarations that MUST be present
		unexpectedPattern string   // Pattern that should NOT appear (scalar declarations)
	}{
		{
			name: "single local variable",
			pine: `
//@version=5
indicator("Test")
calc(len) =>
    result = close * len
    result
`,
			expectedDecls: []string{
				`resultSeries := arrowCtx.GetOrCreateSeries("result")`,
			},
			unexpectedPattern: `result :=`, // No scalar assignment
		},
		{
			name: "multiple local variables",
			pine: `
//@version=5
indicator("Test")
compute(period) =>
    avg = close + open
    diff = high - low
    ratio = avg / diff
    ratio
`,
			expectedDecls: []string{
				`avgSeries := arrowCtx.GetOrCreateSeries("avg")`,
				`diffSeries := arrowCtx.GetOrCreateSeries("diff")`,
				`ratioSeries := arrowCtx.GetOrCreateSeries("ratio")`,
			},
			unexpectedPattern: `avg :=`, // No scalar assignments
		},
		{
			name: "variables used in expressions",
			pine: `
//@version=5
indicator("Test")
calculate(len) =>
    upper = close + 10
    lower = close - 10
    mid = (upper + lower) / 2
    mid
`,
			expectedDecls: []string{
				`upperSeries := arrowCtx.GetOrCreateSeries("upper")`,
				`lowerSeries := arrowCtx.GetOrCreateSeries("lower")`,
				`midSeries := arrowCtx.GetOrCreateSeries("mid")`,
			},
			unexpectedPattern: `upper :=`,
		},
		{
			name: "variables in conditional expressions",
			pine: `
//@version=5
indicator("Test")
check(threshold) =>
    value = close > open ? high : low
    adjusted = value * threshold
    adjusted
`,
			expectedDecls: []string{
				`valueSeries := arrowCtx.GetOrCreateSeries("value")`,
				`adjustedSeries := arrowCtx.GetOrCreateSeries("adjusted")`,
			},
			unexpectedPattern: `value :=`,
		},
		{
			name: "tuple destructuring variables",
			pine: `
//@version=5
indicator("Test")
helper(len) =>
    a = close * 2
    b = open / 2
    [a, b]

main(period) =>
    [x, y] = helper(period)
    result = x + y
    result
`,
			expectedDecls: []string{
				// Note: x, y may be temp variables in main context, but result must be Series
				`resultSeries := arrowCtx.GetOrCreateSeries("result")`,
				// In helper function, a and b must be Series
				`aSeries := arrowCtx.GetOrCreateSeries("a")`,
				`bSeries := arrowCtx.GetOrCreateSeries("b")`,
			},
			unexpectedPattern: `result :=`, // result must use Series, not scalar
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := compilePineScript(tt.pine)
			if err != nil {
				t.Fatalf("Failed to compile: %v", err)
			}

			// Verify ALL expected Series declarations are present
			for _, expectedDecl := range tt.expectedDecls {
				if !strings.Contains(code, expectedDecl) {
					t.Errorf("Missing expected Series declaration:\n  %s\n\nGenerated code:\n%s",
						expectedDecl, code)
				}
			}

			// Verify scalar assignment pattern does NOT appear
			if tt.unexpectedPattern != "" {
				// Check for scalar assignment in arrow function bodies only
				// (avoid false positives from tuple temp variables)
				if strings.Contains(code, "arrowCtx *context.ArrowContext") {
					arrowFuncStart := strings.Index(code, "arrowCtx *context.ArrowContext")
					arrowFuncEnd := strings.Index(code[arrowFuncStart:], "\n}\n")
					if arrowFuncEnd != -1 {
						arrowFuncBody := code[arrowFuncStart : arrowFuncStart+arrowFuncEnd]
						if strings.Contains(arrowFuncBody, tt.unexpectedPattern) &&
							!strings.Contains(arrowFuncBody, "temp_") { // Allow temp_ variables
							t.Errorf("Found unexpected scalar assignment pattern in arrow function: %s\n\nArrow function body:\n%s",
								tt.unexpectedPattern, arrowFuncBody)
						}
					}
				}
			}
		})
	}
}

/*
TestArrowFunction_UniversalSeriesAssignments validates that ALL variable assignments
use Series.Set() instead of scalar assignment.

Behavior: Every assignment to a local variable must use the pattern:

	varNameSeries.Set(expression)

This ensures values are stored in ForwardSeriesBuffer for historical access.
*/
func TestArrowFunction_UniversalSeriesAssignments(t *testing.T) {
	tests := []struct {
		name                string
		pine                string
		expectedAssignments []string
	}{
		{
			name: "simple assignment",
			pine: `
//@version=5
indicator("Test")
calc(multiplier) =>
    result = close * multiplier
    result
`,
			expectedAssignments: []string{
				`resultSeries.Set(`,
			},
		},
		{
			name: "multiple assignments",
			pine: `
//@version=5
indicator("Test")
compute(len) =>
    a = close + open
    b = high - low
    c = a / b
    c
`,
			expectedAssignments: []string{
				`aSeries.Set(`,
				`bSeries.Set(`,
				`cSeries.Set(`,
			},
		},
		{
			name: "assignment with complex expression",
			pine: `
//@version=5
indicator("Test")
calc(period) =>
    avg = (close + open + high + low) / 4
    avg
`,
			expectedAssignments: []string{
				`avgSeries.Set((`,
			},
		},
		{
			name: "assignment with conditional",
			pine: `
//@version=5
indicator("Test")
check(threshold) =>
    value = close > threshold ? high : low
    value
`,
			expectedAssignments: []string{
				`valueSeries.Set(func() float64`,
			},
		},
		{
			name: "unary expression assignment",
			pine: `
//@version=5
indicator("Test")
negate(val) =>
    negative = -val
    negative
`,
			expectedAssignments: []string{
				`negativeSeries.Set(-`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := compilePineScript(tt.pine)
			if err != nil {
				t.Fatalf("Failed to compile: %v", err)
			}

			for _, expectedAssignment := range tt.expectedAssignments {
				if !strings.Contains(code, expectedAssignment) {
					t.Errorf("Missing expected Series assignment pattern:\n  %s\n\nGenerated code:\n%s",
						expectedAssignment, code)
				}
			}
		})
	}
}

/*
TestArrowFunction_UniversalSeriesAccess validates that ALL variable references
resolve to Series.GetCurrent() access pattern.

Behavior: When a local variable is referenced in an expression, it must access

	the current value via: varNameSeries.GetCurrent()

This ensures consistent Series access across all expression contexts.
*/
func TestArrowFunction_UniversalSeriesAccess(t *testing.T) {
	tests := []struct {
		name           string
		pine           string
		expectedAccess []string
		contextDesc    string
	}{
		{
			name: "variable in binary expression",
			pine: `
//@version=5
indicator("Test")
calc(multiplier) =>
    base = close * 2
    result = base + multiplier
    result
`,
			expectedAccess: []string{
				`baseSeries.GetCurrent()`,
			},
			contextDesc: "binary expression with local variable",
		},
		{
			name: "variable in conditional test",
			pine: `
//@version=5
indicator("Test")
check(threshold) =>
    value = close + open
    result = value > threshold ? 1 : 0
    result
`,
			expectedAccess: []string{
				`valueSeries.GetCurrent() >`,
			},
			contextDesc: "conditional expression test",
		},
		{
			name: "variable in return expression",
			pine: `
//@version=5
indicator("Test")
compute(len) =>
    result = close * len
    result
`,
			expectedAccess: []string{
				`return resultSeries.GetCurrent()`,
			},
			contextDesc: "return statement",
		},
		{
			name: "multiple variable references",
			pine: `
//@version=5
indicator("Test")
combine(multiplier) =>
    a = close * 2
    b = open / 2
    result = a + b
    result
`,
			expectedAccess: []string{
				`aSeries.GetCurrent()`,
				`bSeries.GetCurrent()`,
			},
			contextDesc: "multiple variables in single expression",
		},
		{
			name: "variable in nested expression",
			pine: `
//@version=5
indicator("Test")
nested(threshold) =>
    base = close + open
    adjusted = (base * 2) / threshold
    adjusted
`,
			expectedAccess: []string{
				`(baseSeries.GetCurrent() * 2)`,
			},
			contextDesc: "nested parenthesized expression",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := compilePineScript(tt.pine)
			if err != nil {
				t.Fatalf("Failed to compile: %v", err)
			}

			for _, expectedAccess := range tt.expectedAccess {
				if !strings.Contains(code, expectedAccess) {
					t.Errorf("Missing expected Series access pattern in %s:\n  %s\n\nGenerated code:\n%s",
						tt.contextDesc, expectedAccess, code)
				}
			}
		})
	}
}

/*
TestArrowFunction_ParameterVsLocalVariableDistinction validates that function parameters
remain scalar while local variables use Series.

Behavior: Parameters are passed as scalar float64 values and accessed directly.

	Local variables declared in function body use Series storage.

This ensures efficient parameter passing without unnecessary Series overhead.
*/
func TestArrowFunction_ParameterVsLocalVariableDistinction(t *testing.T) {
	tests := []struct {
		name              string
		pine              string
		paramName         string
		localVarName      string
		expectParamDirect bool // Should parameter be accessed directly?
		expectLocalSeries bool // Should local variable use Series?
	}{
		{
			name: "parameter and local variable",
			pine: `
//@version=5
indicator("Test")
calc(multiplier) =>
    result = close * multiplier
    result
`,
			paramName:         "multiplier",
			localVarName:      "result",
			expectParamDirect: true,
			expectLocalSeries: true,
		},
		{
			name: "multiple parameters",
			pine: `
//@version=5
indicator("Test")
compute(len, offset) =>
    adjusted = close + offset
    scaled = adjusted * len
    scaled
`,
			paramName:         "len",
			localVarName:      "adjusted",
			expectParamDirect: true,
			expectLocalSeries: true,
		},
		{
			name: "parameter used in expression with local variable",
			pine: `
//@version=5
indicator("Test")
calculate(factor) =>
    base = close * 2
    result = base * factor
    result
`,
			paramName:         "factor",
			localVarName:      "base",
			expectParamDirect: true,
			expectLocalSeries: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := compilePineScript(tt.pine)
			if err != nil {
				t.Fatalf("Failed to compile: %v", err)
			}

			if tt.expectParamDirect {
				// Parameter should NOT have Series declaration
				paramSeriesDecl := tt.paramName + "Series := arrowCtx.GetOrCreateSeries"
				if strings.Contains(code, paramSeriesDecl) {
					t.Errorf("Parameter %s should NOT have Series declaration, but found: %s",
						tt.paramName, paramSeriesDecl)
				}

				// Parameter should be used directly in expressions (as scalar)
				// Check that parameter appears without .GetCurrent() suffix
				if strings.Contains(code, tt.paramName+"Series.GetCurrent()") {
					t.Errorf("Parameter %s should be accessed directly, not via Series.GetCurrent()",
						tt.paramName)
				}
			}

			if tt.expectLocalSeries {
				// Local variable MUST have Series declaration
				localSeriesDecl := tt.localVarName + "Series := arrowCtx.GetOrCreateSeries"
				if !strings.Contains(code, localSeriesDecl) {
					t.Errorf("Local variable %s MUST have Series declaration:\n  %s\n\nGenerated code:\n%s",
						tt.localVarName, localSeriesDecl, code)
				}

				// Local variable MUST use Series.Set()
				localSeriesSet := tt.localVarName + "Series.Set("
				if !strings.Contains(code, localSeriesSet) {
					t.Errorf("Local variable %s MUST use Series.Set():\n  %s\n\nGenerated code:\n%s",
						tt.localVarName, localSeriesSet, code)
				}
			}
		})
	}
}

/*
TestArrowFunction_TupleReturnSeriesAccess validates that tuple return values
access Series.GetCurrent() for each element.

Behavior: When returning multiple values, each returned variable must use

	Series.GetCurrent() to get the final value.

This ensures tuple returns work correctly with universal Series storage.
*/
func TestArrowFunction_TupleReturnSeriesAccess(t *testing.T) {
	tests := []struct {
		name           string
		pine           string
		expectedReturn string
	}{
		{
			name: "two-element tuple",
			pine: `
//@version=5
indicator("Test")
pair(multiplier) =>
    a = close * multiplier
    b = open * multiplier
    [a, b]
`,
			expectedReturn: `return aSeries.GetCurrent(), bSeries.GetCurrent()`,
		},
		{
			name: "three-element tuple",
			pine: `
//@version=5
indicator("Test")
triple(len) =>
    x = close + len
    y = open + len
    z = high + len
    [x, y, z]
`,
			expectedReturn: `return xSeries.GetCurrent(), ySeries.GetCurrent(), zSeries.GetCurrent()`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := compilePineScript(tt.pine)
			if err != nil {
				t.Fatalf("Failed to compile: %v", err)
			}

			if !strings.Contains(code, tt.expectedReturn) {
				t.Errorf("Missing expected tuple return pattern:\n  %s\n\nGenerated code:\n%s",
					tt.expectedReturn, code)
			}
		})
	}
}

/*
TestArrowFunction_ComplexExpressionSeriesIntegration validates that complex expressions
correctly integrate with universal Series storage.

Behavior: Complex expressions (binary, unary, conditional) must correctly resolve

	local variable references to Series.GetCurrent() while maintaining
	expression semantics.

This ensures all expression types work seamlessly with Series-based variables.
*/
func TestArrowFunction_ComplexExpressionSeriesIntegration(t *testing.T) {
	tests := []struct {
		name            string
		pine            string
		exprType        string
		expectedPattern string
	}{
		{
			name: "binary with both locals",
			pine: `
//@version=5
indicator("Test")
combine(factor) =>
    a = close * factor
    b = open * factor
    result = a + b
    result
`,
			exprType:        "binary expression with two local variables",
			expectedPattern: `aSeries.GetCurrent() + bSeries.GetCurrent()`,
		},
		{
			name: "unary with local",
			pine: `
//@version=5
indicator("Test")
negate(multiplier) =>
    value = close * multiplier
    inverted = -value
    inverted
`,
			exprType:        "unary expression with local variable",
			expectedPattern: `-valueSeries.GetCurrent()`,
		},
		{
			name: "conditional with local in test",
			pine: `
//@version=5
indicator("Test")
check(threshold) =>
    value = close + open
    signal = value > threshold ? 1 : 0
    signal
`,
			exprType:        "conditional with local variable in test",
			expectedPattern: `valueSeries.GetCurrent() > threshold`,
		},
		{
			name: "nested binary expressions",
			pine: `
//@version=5
indicator("Test")
calculate(factor) =>
    a = close * 2
    b = open / 2
    c = (a + b) * factor
    c
`,
			exprType:        "nested binary with locals",
			expectedPattern: `(aSeries.GetCurrent() + bSeries.GetCurrent())`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := compilePineScript(tt.pine)
			if err != nil {
				t.Fatalf("Failed to compile: %v", err)
			}

			if !strings.Contains(code, tt.expectedPattern) {
				t.Errorf("Missing expected pattern for %s:\n  %s\n\nGenerated code:\n%s",
					tt.exprType, tt.expectedPattern, code)
			}
		})
	}
}

/*
TestArrowFunction_EdgeCasesSeriesBehavior validates Series storage behavior
in edge cases and boundary conditions.

Behavior: Universal Series storage must work correctly even in minimal or

	unusual function structures.

This ensures robustness across all valid PineScript patterns.
*/
func TestArrowFunction_EdgeCasesSeriesBehavior(t *testing.T) {
	tests := []struct {
		name        string
		pine        string
		description string
		validate    func(t *testing.T, code string)
	}{
		{
			name: "single variable function",
			pine: `
//@version=5
indicator("Test")
identity(x) =>
    result = x
    result
`,
			description: "function with single local variable",
			validate: func(t *testing.T, code string) {
				if !strings.Contains(code, `resultSeries := arrowCtx.GetOrCreateSeries("result")`) {
					t.Error("Single variable must have Series declaration")
				}
				if !strings.Contains(code, `resultSeries.Set(`) {
					t.Error("Single variable must use Series.Set()")
				}
			},
		},
		{
			name: "immediate return of expression",
			pine: `
//@version=5
indicator("Test")
direct(multiplier) =>
    value = close * multiplier
    value
`,
			description: "variable assigned and immediately returned",
			validate: func(t *testing.T, code string) {
				if !strings.Contains(code, `valueSeries := arrowCtx.GetOrCreateSeries("value")`) {
					t.Error("Immediately returned variable must have Series declaration")
				}
				if !strings.Contains(code, `return valueSeries.GetCurrent()`) {
					t.Error("Return must use Series.GetCurrent()")
				}
			},
		},
		{
			name: "many local variables",
			pine: `
//@version=5
indicator("Test")
many(factor) =>
    a = close * factor
    b = open * factor
    c = high * factor
    d = low * factor
    e = (a + b + c + d) / 4
    e
`,
			description: "function with many local variables",
			validate: func(t *testing.T, code string) {
				varNames := []string{"a", "b", "c", "d", "e"}
				for _, varName := range varNames {
					decl := varName + "Series := arrowCtx.GetOrCreateSeries"
					if !strings.Contains(code, decl) {
						t.Errorf("Variable %s missing Series declaration", varName)
					}
				}
			},
		},
		{
			name: "variable shadowing parameter name style",
			pine: `
//@version=5
indicator("Test")
process(length) =>
    lengthAdjusted = length * 2
    lengthAdjusted
`,
			description: "local variable name similar to parameter",
			validate: func(t *testing.T, code string) {
				// Parameter 'length' should NOT have Series
				if strings.Contains(code, `lengthSeries := arrowCtx.GetOrCreateSeries("length")`) {
					t.Error("Parameter should not have Series declaration")
				}
				// Local 'lengthAdjusted' MUST have Series
				if !strings.Contains(code, `lengthAdjustedSeries := arrowCtx.GetOrCreateSeries("lengthAdjusted")`) {
					t.Error("Local variable must have Series declaration")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := compilePineScript(tt.pine)
			if err != nil {
				t.Fatalf("Failed to compile: %v", err)
			}

			tt.validate(t, code)
		})
	}
}

/*
TestArrowFunction_NestedFunctionCallsSeriesConsistency validates that nested
arrow function calls maintain Series consistency.

Behavior: When one arrow function calls another, the returned values should

	properly integrate with the calling function's Series storage.

This ensures composability of arrow functions with universal Series.
*/
func TestArrowFunction_NestedFunctionCallsSeriesConsistency(t *testing.T) {
	pine := `
//@version=5
indicator("Test")

helper(multiplier) =>
    value = close * multiplier
    value

main(factor) =>
    intermediate = helper(factor)
    result = intermediate * 2
    result
`

	code, err := compilePineScript(pine)
	if err != nil {
		t.Fatalf("Failed to compile: %v", err)
	}

	// Helper function should have Series for its local variable
	if !strings.Contains(code, `valueSeries := arrowCtx.GetOrCreateSeries("value")`) {
		t.Error("Helper function local variable must have Series declaration")
	}
	if !strings.Contains(code, `return valueSeries.GetCurrent()`) {
		t.Error("Helper function must return via Series.GetCurrent()")
	}

	// Main function should store helper's return value in Series
	if !strings.Contains(code, `intermediateSeries := arrowCtx.GetOrCreateSeries("intermediate")`) {
		t.Error("Main function variable storing nested call result must have Series declaration")
	}
	if !strings.Contains(code, `intermediateSeries.Set(`) {
		t.Error("Nested call result must be stored via Series.Set()")
	}
	if !strings.Contains(code, `intermediateSeries.GetCurrent()`) {
		t.Error("Variable from nested call must be accessed via Series.GetCurrent()")
	}
}
