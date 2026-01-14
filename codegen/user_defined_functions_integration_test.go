package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/parser"
)

/* TestUserDefinedFunctions_BasicGeneration validates arrow functions are generated as Go functions */
func TestUserDefinedFunctions_BasicGeneration(t *testing.T) {
	tests := []struct {
		name           string
		source         string
		expectedFunc   string
		expectedParams []string
		mustContain    []string
		mustNotContain []string
	}{
		{
			name: "single parameter function",
			source: `
//@version=4
strategy("Test")
simple(x) =>
    x + 1
result = simple(5)
`,
			expectedFunc:   "func simple(arrowCtx *context.ArrowContext, x float64)",
			expectedParams: []string{"x float64"},
			mustContain:    []string{"func simple", "arrowCtx *context.ArrowContext"},
			mustNotContain: []string{"simpleSeries", "var simple"},
		},
		{
			name: "multiple parameter function",
			source: `
//@version=4
strategy("Test")
calc(a, b, c) =>
    a + b * c
result = calc(1, 2, 3)
`,
			expectedFunc:   "func calc(arrowCtx *context.ArrowContext, a float64, b float64, c float64)",
			expectedParams: []string{"a float64", "b float64", "c float64"},
			mustContain:    []string{"func calc", "a float64", "b float64", "c float64"},
			mustNotContain: []string{"calcSeries", "var calc"},
		},
		{
			name: "zero parameter function",
			source: `
//@version=4
strategy("Test")
constant() =>
    42
result = constant()
`,
			expectedFunc:   "func constant(arrowCtx *context.ArrowContext) float64",
			mustContain:    []string{"func constant", "arrowCtx *context.ArrowContext"},
			mustNotContain: []string{"constantSeries", "var constant"},
		},
		{
			name: "function with tuple return",
			source: `
//@version=4
strategy("Test")
minmax(a, b) =>
    [min(a, b), max(a, b)]
[lower, upper] = minmax(10, 20)
`,
			expectedFunc:   "func minmax(arrowCtx *context.ArrowContext, a float64, b float64) (float64, float64)",
			mustContain:    []string{"func minmax", "(float64, float64)"},
			mustNotContain: []string{"minmaxSeries"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fullCode, err := compilePineScript(tt.source)
			if err != nil {
				t.Fatalf("compilePineScript failed: %v", err)
			}

			// Verify function is generated
			if !strings.Contains(fullCode, tt.expectedFunc) {
				t.Errorf("Expected function signature not found: %q\nGenerated code:\n%s",
					tt.expectedFunc, fullCode)
			}

			// Verify required patterns
			for _, pattern := range tt.mustContain {
				if !strings.Contains(fullCode, pattern) {
					t.Errorf("Missing required pattern: %q", pattern)
				}
			}

			// Verify forbidden patterns
			for _, pattern := range tt.mustNotContain {
				if strings.Contains(fullCode, pattern) {
					t.Errorf("Found forbidden pattern: %q", pattern)
				}
			}
		})
	}
}

/* TestUserDefinedFunctions_NotTreatedAsSeries ensures functions don't create Series variables */
func TestUserDefinedFunctions_NotTreatedAsSeries(t *testing.T) {
	source := `
//@version=4
strategy("Test")
helper(x) =>
    x * 2
result = helper(10)
`

	fullCode, err := compilePineScript(source)
	if err != nil {
		t.Fatalf("compilePineScript failed: %v", err)
	}

	// Function should NOT create Series variable
	forbiddenPatterns := []string{
		"var helperSeries *series.Series",
		"helperSeries = series.NewSeries",
		"helperSeries.Set(",
		"helperSeries.Next()",
		"_ = helperSeries",
	}

	for _, pattern := range forbiddenPatterns {
		if strings.Contains(fullCode, pattern) {
			t.Errorf("Function incorrectly treated as Series variable: found %q", pattern)
		}
	}

	// Function SHOULD be in generated code
	if !strings.Contains(fullCode, "func helper") {
		t.Error("Function not generated")
	}
}

/* TestUserDefinedFunctions_SecurityInjectionPreservation validates functions survive security() processing */
func TestUserDefinedFunctions_SecurityInjectionPreservation(t *testing.T) {
	source := `
//@version=4
strategy("Test")
custom(x) =>
    x + 1
daily_close = security(syminfo.tickerid, "D", close)
result = custom(daily_close)
`

	p, err := parser.NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	script, err := p.ParseBytes("test.pine", []byte(source))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	converter := parser.NewConverter()
	program, err := converter.ToESTree(script)
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	codeBeforeInjection, err := GenerateStrategyCodeFromAST(program)
	if err != nil {
		t.Fatalf("GenerateStrategyCodeFromAST failed: %v", err)
	}

	// Simulate security injection (this is what happens in the build pipeline)
	codeAfterInjection, err := InjectSecurityCode(codeBeforeInjection, program)
	if err != nil {
		t.Fatalf("InjectSecurityCode failed: %v", err)
	}

	// UserDefinedFunctions MUST be preserved through security injection
	if len(codeAfterInjection.UserDefinedFunctions) == 0 {
		t.Error("UserDefinedFunctions lost during security injection")
	}

	if !strings.Contains(codeAfterInjection.UserDefinedFunctions, "func custom") {
		t.Errorf("Function 'custom' lost during security injection\nUserDefinedFunctions:\n%s",
			codeAfterInjection.UserDefinedFunctions)
	}

	// Verify security prefetch code is also present
	if !strings.Contains(codeAfterInjection.FunctionBody, "request.security") {
		t.Error("Security prefetch code not injected")
	}
}

/* TestUserDefinedFunctions_NestedCalls validates functions calling other user-defined functions */
func TestUserDefinedFunctions_NestedCalls(t *testing.T) {
	source := `
//@version=4
strategy("Test")
inner(x) =>
    x * 2
outer(y) =>
    inner(y) + 1
result = outer(5)
`

	fullCode, err := compilePineScript(source)
	if err != nil {
		t.Fatalf("compilePineScript failed: %v", err)
	}

	// Both functions should be generated
	if !strings.Contains(fullCode, "func inner") {
		t.Error("Inner function not generated")
	}
	if !strings.Contains(fullCode, "func outer") {
		t.Error("Outer function not generated")
	}

	// Neither should create Series
	forbiddenPatterns := []string{"innerSeries", "outerSeries"}
	for _, pattern := range forbiddenPatterns {
		if strings.Contains(fullCode, pattern) {
			t.Errorf("Found forbidden Series pattern: %q", pattern)
		}
	}
}

/* TestUserDefinedFunctions_ComplexBody validates functions with multiple statements and local variables */
func TestUserDefinedFunctions_ComplexBody(t *testing.T) {
	source := `
//@version=4
strategy("Test")
dirmov(len) =>
    up = change(high)
    down = -change(low)
    truerange = rma(tr, len)
    plus = fixnan(100 * rma(up > down and up > 0 ? up : 0, len) / truerange)
    minus = fixnan(100 * rma(down > up and down > 0 ? down : 0, len) / truerange)
    [plus, minus]
[p, m] = dirmov(14)
`

	fullCode, err := compilePineScript(source)
	if err != nil {
		t.Fatalf("compilePineScript failed: %v", err)
	}

	// Function should be generated with proper signature
	if !strings.Contains(fullCode, "func dirmov") {
		t.Error("dirmov function not generated")
	}

	// Should have tuple return type
	if !strings.Contains(fullCode, "(float64, float64)") {
		t.Error("Tuple return type not found")
	}

	// Should contain local variable declarations for arrow function body
	requiredPatterns := []string{
		"upSeries",
		"downSeries",
		"truerangeSeries",
		"plusSeries",
		"minusSeries",
	}

	for _, pattern := range requiredPatterns {
		if !strings.Contains(fullCode, pattern) {
			t.Errorf("Missing local variable pattern in function body: %q", pattern)
		}
	}

	// Main function body should NOT have dirmovSeries
	if strings.Contains(fullCode, "dirmovSeries") {
		t.Error("Function incorrectly treated as Series in main body")
	}
}

/* TestUserDefinedFunctions_MultipleDeclarations validates multiple functions in same script */
func TestUserDefinedFunctions_MultipleDeclarations(t *testing.T) {
	source := `
//@version=4
strategy("Test")
func1(x) =>
    x + 1
func2(y) =>
    y * 2
func3(z) =>
    z - 1
result = func1(10) + func2(20) + func3(30)
`

	fullCode, err := compilePineScript(source)
	if err != nil {
		t.Fatalf("compilePineScript failed: %v", err)
	}

	// All three functions should be generated
	requiredFunctions := []string{"func func1", "func func2", "func func3"}
	for _, funcSig := range requiredFunctions {
		if !strings.Contains(fullCode, funcSig) {
			t.Errorf("Function not found: %q", funcSig)
		}
	}

	// None should create Series for the FUNCTION DECLARATION itself
	// Note: Function call RESULTS will create Series (e.g., "result = func1(10)" → resultSeries)
	// But func1, func2, func3 should NOT be declared as Series variables
	forbiddenPatterns := []string{
		"var func1Series *series.Series",
		"func1Series = series.NewSeries",
		"var func2Series *series.Series",
		"func2Series = series.NewSeries",
		"var func3Series *series.Series",
		"func3Series = series.NewSeries",
	}
	for _, pattern := range forbiddenPatterns {
		if strings.Contains(fullCode, pattern) {
			t.Errorf("Found forbidden Series pattern: %q", pattern)
		}
	}

	// Verify all functions are called in the body
	for _, funcCall := range []string{"func1(", "func2(", "func3("} {
		if !strings.Contains(fullCode, funcCall) {
			t.Errorf("Function call not found in body: %q", funcCall)
		}
	}
}

/* TestUserDefinedFunctions_MixedWithVariables ensures functions and regular variables coexist */
func TestUserDefinedFunctions_MixedWithVariables(t *testing.T) {
	source := `
//@version=4
strategy("Test")
myVar = 10
myFunc(x) =>
    x * 2
myVar2 = 20
result = myFunc(myVar) + myVar2
`

	fullCode, err := compilePineScript(source)
	if err != nil {
		t.Fatalf("compilePineScript failed: %v", err)
	}

	// Function should be generated
	if !strings.Contains(fullCode, "func myFunc") {
		t.Error("myFunc not found")
	}

	// Variables should create Series
	requiredSeries := []string{"myVarSeries", "myVar2Series", "resultSeries"}
	for _, seriesVar := range requiredSeries {
		if !strings.Contains(fullCode, seriesVar) {
			t.Errorf("Variable Series not found: %q", seriesVar)
		}
	}

	// Function should NOT be declared as Series variable
	forbiddenPatterns := []string{
		"var myFuncSeries *series.Series",
		"myFuncSeries = series.NewSeries",
	}
	for _, pattern := range forbiddenPatterns {
		if strings.Contains(fullCode, pattern) {
			t.Errorf("Function incorrectly created Series variable: %q", pattern)
		}
	}
}

/* TestUserDefinedFunctions_VariableRegistryIsolation ensures function variables don't pollute main scope */
func TestUserDefinedFunctions_VariableRegistryIsolation(t *testing.T) {
	source := `
//@version=4
strategy("Test")
myFunc(len) =>
    temp = len * 2
    result = temp + 1
    result
value = myFunc(10)
`

	fullCode, err := compilePineScript(source)
	if err != nil {
		t.Fatalf("compilePineScript failed: %v", err)
	}

	// Function-local variables (temp, result) should be in generated code
	if !strings.Contains(fullCode, "tempSeries") {
		t.Error("Function-local 'temp' variable not found")
	}

	// Main scope variable 'value' SHOULD be in generated code
	if !strings.Contains(fullCode, "valueSeries") {
		t.Error("Main scope variable 'value' not found")
	}
}
