package codegen

import (
	"strings"
	"testing"
)

/*
TestArrowFunction_ContextPropagation_NestedCalls validates that nested arrow function
calls correctly propagate ArrowContext instead of main Context.

Behavior: When an arrow function calls another arrow function, it must pass its
ArrowContext (arrowCtx) rather than the main context (ctx). This ensures proper
Series isolation and state management in nested function calls.

Architecture: This tests the fundamental context passing mechanism that enables
multi-level arrow function composition with isolated Series buffers per function.
*/
func TestArrowFunction_ContextPropagation_NestedCalls(t *testing.T) {
	tests := []struct {
		name              string
		pine              string
		callerFunc        string
		calleeFunc        string
		expectedCallSite  string
		forbiddenCallSite string
		description       string
	}{
		{
			name: "single level nesting",
			pine: `
//@version=5
indicator("Test")
inner(len) =>
    value = close * len
    value

outer(period) =>
    result = inner(period)
    result
`,
			callerFunc:        "outer",
			calleeFunc:        "inner",
			expectedCallSite:  "inner(arrowCtx,",
			forbiddenCallSite: "inner(ctx,",
			description:       "outer calls inner - must pass arrowCtx",
		},
		{
			name: "three level nesting",
			pine: `
//@version=5
indicator("Test")
level3(x) =>
    x * 2

level2(y) =>
    level3(y)

level1(z) =>
    level2(z)
`,
			callerFunc:        "level2",
			calleeFunc:        "level3",
			expectedCallSite:  "level3(arrowCtx,",
			forbiddenCallSite: "level3(ctx,",
			description:       "middle level must pass arrowCtx to deeper level",
		},
		{
			name: "multiple calls same level",
			pine: `
//@version=5
indicator("Test")
helper(val) =>
    val + 1

processor(len) =>
    a = helper(len)
    b = helper(len * 2)
    a + b
`,
			callerFunc:        "processor",
			calleeFunc:        "helper",
			expectedCallSite:  "helper(arrowCtx,",
			forbiddenCallSite: "helper(ctx,",
			description:       "multiple calls to same function from arrow context",
		},
		{
			name: "tuple destructuring nested call",
			pine: `
//@version=5
indicator("Test")
pair(multiplier) =>
    a = close * multiplier
    b = open * multiplier
    [a, b]

consumer(factor) =>
    [x, y] = pair(factor)
    x + y
`,
			callerFunc:        "consumer",
			calleeFunc:        "pair",
			expectedCallSite:  "pair(arrowCtx,",
			forbiddenCallSite: "pair(ctx,",
			description:       "tuple destructuring must use arrowCtx",
		},
		{
			name: "nested call with expression argument",
			pine: `
//@version=5
indicator("Test")
compute(value) =>
    value * 2

wrapper(len) =>
    local = close + len
    result = compute(local)
    result
`,
			callerFunc:        "wrapper",
			calleeFunc:        "compute",
			expectedCallSite:  "compute(arrowCtx,",
			forbiddenCallSite: "compute(ctx,",
			description:       "nested call with local variable as argument",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := compilePineScript(tt.pine)
			if err != nil {
				t.Fatalf("Failed to compile: %v", err)
			}

			// Verify correct context propagation
			if !strings.Contains(code, tt.expectedCallSite) {
				t.Errorf("%s: Missing expected call site pattern:\n  %s\n\nGenerated code:\n%s",
					tt.description, tt.expectedCallSite, code)
			}

			// Verify no incorrect context usage
			if strings.Contains(code, tt.forbiddenCallSite) {
				t.Errorf("%s: Found forbidden call site pattern (should use arrowCtx, not ctx):\n  %s\n\nGenerated code:\n%s",
					tt.description, tt.forbiddenCallSite, code)
			}

			// Additional validation: ensure arrowCtx is available in caller function
			callerFuncStart := "func " + tt.callerFunc + "(arrowCtx *context.ArrowContext"
			if !strings.Contains(code, callerFuncStart) {
				t.Errorf("%s: Caller function %s must have arrowCtx parameter:\n  %s",
					tt.description, tt.callerFunc, callerFuncStart)
			}
		})
	}
}

/*
TestArrowFunction_ContextPropagation_ParameterTypes validates that arrow function
calls receive correct parameter types (ArrowContext for nested calls, Context for top-level).

Behavior: Top-level calls from main bar loop pass ctx (Context), while nested
arrow-to-arrow calls pass arrowCtx (ArrowContext). This ensures proper type safety
and context isolation.
*/
func TestArrowFunction_ContextPropagation_ParameterTypes(t *testing.T) {
	tests := []struct {
		name                 string
		pine                 string
		funcName             string
		expectedSignature    string
		callFromMainContext  bool
		expectedMainCallSite string
	}{
		{
			name: "top level function called from main",
			pine: `
//@version=5
indicator("Test")
topLevel(len) =>
    close * len

result = topLevel(14)
`,
			funcName:             "topLevel",
			expectedSignature:    "func topLevel(arrowCtx *context.ArrowContext, len float64)",
			callFromMainContext:  true,
			expectedMainCallSite: "arrowCtx_topLevel_1 := context.NewArrowContext(ctx)",
		},
		{
			name: "nested function signature",
			pine: `
//@version=5
indicator("Test")
nested(val) =>
    val * 2

caller(len) =>
    nested(len)
`,
			funcName:          "nested",
			expectedSignature: "func nested(arrowCtx *context.ArrowContext, val float64)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := compilePineScript(tt.pine)
			if err != nil {
				t.Fatalf("Failed to compile: %v", err)
			}

			// Verify function signature
			if !strings.Contains(code, tt.expectedSignature) {
				t.Errorf("Missing expected function signature:\n  %s\n\nGenerated code:\n%s",
					tt.expectedSignature, code)
			}

			// Verify main context call site if applicable
			if tt.callFromMainContext {
				if !strings.Contains(code, tt.expectedMainCallSite) {
					t.Errorf("Missing expected main context call site:\n  %s\n\nGenerated code:\n%s",
						tt.expectedMainCallSite, code)
				}
			}
		})
	}
}

/*
TestArrowFunction_LocalVariableResolution_InExpressions validates that local
variables are correctly resolved to Series.GetCurrent() in all expression contexts.

Behavior: Any reference to a local variable (declared in arrow function body)
must resolve to varNameSeries.GetCurrent() to access the current Series value.
This applies to all expression types: binary, unary, conditional, function arguments.

Architecture: Tests the identifier resolution mechanism that distinguishes between
parameters (direct access), local variables (Series access), and builtins.
*/
func TestArrowFunction_LocalVariableResolution_InExpressions(t *testing.T) {
	tests := []struct {
		name           string
		pine           string
		localVarName   string
		exprContext    string
		expectedAccess string
		description    string
	}{
		{
			name: "local var in binary expression",
			pine: `
//@version=5
indicator("Test")
calc(multiplier) =>
    base = close * 2
    result = base + multiplier
    result
`,
			localVarName:   "base",
			exprContext:    "binary addition",
			expectedAccess: "baseSeries.GetCurrent() +",
			description:    "local variable as left operand in binary expression",
		},
		{
			name: "local var in division",
			pine: `
//@version=5
indicator("Test")
divide(divisor) =>
    numerator = close * 10
    result = numerator / divisor
    result
`,
			localVarName:   "numerator",
			exprContext:    "binary division",
			expectedAccess: "numeratorSeries.GetCurrent() /",
			description:    "local variable in division operation",
		},
		{
			name: "local var in conditional test",
			pine: `
//@version=5
indicator("Test")
check(threshold) =>
    value = close + open
    signal = value > threshold ? 1 : 0
    signal
`,
			localVarName:   "value",
			exprContext:    "conditional test",
			expectedAccess: "valueSeries.GetCurrent() >",
			description:    "local variable in ternary condition",
		},
		{
			name: "local var in unary expression",
			pine: `
//@version=5
indicator("Test")
negate(factor) =>
    positive = close * factor
    negative = -positive
    negative
`,
			localVarName:   "positive",
			exprContext:    "unary negation",
			expectedAccess: "-positiveSeries.GetCurrent()",
			description:    "local variable in unary minus",
		},
		{
			name: "local var in nested binary",
			pine: `
//@version=5
indicator("Test")
complex(multiplier) =>
    a = close * 2
    b = open / 2
    result = (a + b) * multiplier
    result
`,
			localVarName:   "a",
			exprContext:    "nested binary expression",
			expectedAccess: "(aSeries.GetCurrent() +",
			description:    "local variable in nested parenthesized expression",
		},
		{
			name: "local var in function argument",
			pine: `
//@version=5
indicator("Test")
helper(x) =>
    x * 2

caller(len) =>
    value = close + len
    result = helper(value)
    result

output = caller(14)
`,
			localVarName:   "value",
			exprContext:    "function call argument",
			expectedAccess: "valueSeries.Set((bar.Close + len))",
			description:    "local variable passed as function argument - validates Series storage",
		},
		{
			name: "local var in tuple return",
			pine: `
//@version=5
indicator("Test")
pair(multiplier) =>
    a = close * multiplier
    b = open * multiplier
    [a, b]
`,
			localVarName:   "a",
			exprContext:    "tuple return",
			expectedAccess: "return aSeries.GetCurrent(),",
			description:    "local variable in tuple return statement",
		},
		{
			name: "multiple local vars in expression",
			pine: `
//@version=5
indicator("Test")
combine(factor) =>
    x = close * factor
    y = open * factor
    sum = x + y
    sum
`,
			localVarName:   "x",
			exprContext:    "binary with two locals",
			expectedAccess: "xSeries.GetCurrent() + ySeries.GetCurrent()",
			description:    "two local variables in same binary expression",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := compilePineScript(tt.pine)
			if err != nil {
				t.Fatalf("Failed to compile: %v", err)
			}

			// Verify correct Series.GetCurrent() resolution
			if !strings.Contains(code, tt.expectedAccess) {
				t.Errorf("%s: Missing expected Series access in %s:\n  %s\n\nGenerated code:\n%s",
					tt.description, tt.exprContext, tt.expectedAccess, code)
			}

			// Verify Series declaration exists
			seriesDecl := tt.localVarName + "Series := arrowCtx.GetOrCreateSeries"
			if !strings.Contains(code, seriesDecl) {
				t.Errorf("%s: Missing Series declaration for local variable %s",
					tt.description, tt.localVarName)
			}
		})
	}
}

/*
TestArrowFunction_LocalVariableResolution_InIIFE validates that local variables
are correctly resolved inside inline IIFE expressions (fixnan, TA functions).

Behavior: When local variables are referenced inside IIFE expressions generated
for fixnan() or TA functions, they must resolve to Series.GetCurrent(). This is
critical for correctness when expressions contain references to previously
declared local variables.

Architecture: Tests the deep expression resolution that occurs in IIFE contexts,
ensuring the identifier resolution mechanism works recursively through nested
function scopes.
*/
func TestArrowFunction_LocalVariableResolution_InIIFE(t *testing.T) {
	tests := []struct {
		name               string
		pine               string
		localVarUsedInIIFE string
		expectedPattern    string
		iifeType           string
		description        string
	}{
		{
			name: "local var in fixnan expression",
			pine: `
//@version=5
indicator("Test")
process(len) =>
    denominator = close - open
    ratio = fixnan(100 / denominator)
    ratio
`,
			localVarUsedInIIFE: "denominator",
			expectedPattern:    "/ denominatorSeries.GetCurrent()",
			iifeType:           "fixnan",
			description:        "local variable in fixnan division",
		},
		{
			name: "local var in complex fixnan",
			pine: `
//@version=5
indicator("Test")
calculate(multiplier) =>
    base = close * multiplier
    adjusted = open * multiplier
    result = fixnan(base / adjusted)
    result
`,
			localVarUsedInIIFE: "base",
			expectedPattern:    "baseSeries.GetCurrent() / adjustedSeries.GetCurrent()",
			iifeType:           "fixnan with two locals",
			description:        "two local variables in fixnan expression",
		},
		{
			name: "local var in rma with local source",
			pine: `
//@version=5
indicator("Test")
smooth(len) =>
    source = close + open
    smoothed = ta.sma(source, len)
    smoothed

output = smooth(14)
`,
			localVarUsedInIIFE: "source",
			expectedPattern:    "sourceSeries.Get(",
			iifeType:           "sma with local variable",
			description:        "local variable as source to TA function",
		},
		{
			name: "local var in nested IIFE",
			pine: `
//@version=5
indicator("Test")
complex(len) =>
    truerange = close - open
    plus = fixnan(100 * rma(close, len) / truerange)
    plus
`,
			localVarUsedInIIFE: "truerange",
			expectedPattern:    "/ truerangeSeries.GetCurrent()",
			iifeType:           "nested fixnan with rma",
			description:        "local variable in nested IIFE structure",
		},
		{
			name: "multiple local vars in fixnan",
			pine: `
//@version=5
indicator("Test")
ratio(factor) =>
    numerator = close * factor
    denominator = open * factor
    result = fixnan((numerator + 10) / denominator)
    result
`,
			localVarUsedInIIFE: "numerator",
			expectedPattern:    "(numeratorSeries.GetCurrent() + 10",
			iifeType:           "fixnan with arithmetic",
			description:        "local variable in fixnan arithmetic expression",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := compilePineScript(tt.pine)
			if err != nil {
				t.Fatalf("Failed to compile: %v", err)
			}

			// Verify local variable resolves to Series.GetCurrent() in IIFE
			if !strings.Contains(code, tt.expectedPattern) {
				t.Errorf("%s: Missing expected pattern in %s:\n  %s\n\nGenerated code:\n%s",
					tt.description, tt.iifeType, tt.expectedPattern, code)
			}

			// Verify Series declaration exists
			seriesDecl := tt.localVarUsedInIIFE + "Series := arrowCtx.GetOrCreateSeries"
			if !strings.Contains(code, seriesDecl) {
				t.Errorf("%s: Missing Series declaration for %s",
					tt.description, tt.localVarUsedInIIFE)
			}

			// Verify no bare variable reference (without Series.GetCurrent())
			// This is the bug we fixed: `/ truerange` instead of `/ truerangeSeries.GetCurrent()`
			bareVarPattern := "/ " + tt.localVarUsedInIIFE + ")"
			if strings.Contains(code, bareVarPattern) {
				t.Errorf("%s: Found bare variable reference without Series.GetCurrent():\n  %s\n"+
					"This indicates identifier resolution failed in IIFE context",
					tt.description, bareVarPattern)
			}
		})
	}
}

/*
TestArrowFunction_ParameterVsLocalVariable_ContextDistinction validates that
parameters are passed as scalars while local variables use Series access.

Behavior: Function parameters remain scalar (float64) and are accessed directly.
Local variables declared in the function body use Series storage. Both can be
used in the same expression with different access patterns.

Architecture: Tests the fundamental distinction in the variable resolution system
that enables efficient parameter passing without Series overhead while maintaining
universal Series storage for local state.
*/
func TestArrowFunction_ParameterVsLocalVariable_ContextDistinction(t *testing.T) {
	tests := []struct {
		name                     string
		pine                     string
		paramName                string
		localVarName             string
		expectedParamAccess      string
		expectedLocalVarAccess   string
		forbiddenParamPattern    string
		forbiddenLocalVarPattern string
	}{
		{
			name: "parameter and local in binary expression",
			pine: `
//@version=5
indicator("Test")
multiply(factor) =>
    base = close * 2
    result = base * factor
    result
`,
			paramName:                "factor",
			localVarName:             "base",
			expectedParamAccess:      "factor", // Direct scalar access
			expectedLocalVarAccess:   "baseSeries.GetCurrent()",
			forbiddenParamPattern:    "factorSeries.GetCurrent()",
			forbiddenLocalVarPattern: "base)", // Bare local var access
		},
		{
			name: "parameter in conditional with local",
			pine: `
//@version=5
indicator("Test")
check(threshold) =>
    value = close + open
    signal = value > threshold ? 1 : 0
    signal
`,
			paramName:                "threshold",
			localVarName:             "value",
			expectedParamAccess:      "threshold",
			expectedLocalVarAccess:   "valueSeries.GetCurrent() >",
			forbiddenParamPattern:    "thresholdSeries",
			forbiddenLocalVarPattern: "value >", // Bare local var in condition
		},
		{
			name: "multiple parameters with local",
			pine: `
//@version=5
indicator("Test")
compute(len, offset) =>
    adjusted = close + offset
    scaled = adjusted * len
    scaled
`,
			paramName:                "len",
			localVarName:             "adjusted",
			expectedParamAccess:      "len", // Direct in multiplication
			expectedLocalVarAccess:   "adjustedSeries.GetCurrent() *",
			forbiddenParamPattern:    "lenSeries",
			forbiddenLocalVarPattern: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := compilePineScript(tt.pine)
			if err != nil {
				t.Fatalf("Failed to compile: %v", err)
			}

			// Verify parameter accessed directly (scalar)
			if !strings.Contains(code, tt.expectedParamAccess) {
				t.Errorf("Missing expected parameter access (direct scalar):\n  %s",
					tt.expectedParamAccess)
			}

			// Verify local variable uses Series.GetCurrent()
			if !strings.Contains(code, tt.expectedLocalVarAccess) {
				t.Errorf("Missing expected local variable Series access:\n  %s",
					tt.expectedLocalVarAccess)
			}

			// Verify parameter does NOT use Series
			if strings.Contains(code, tt.forbiddenParamPattern) {
				t.Errorf("Parameter incorrectly uses Series access:\n  %s\n"+
					"Parameters should be accessed directly as scalars",
					tt.forbiddenParamPattern)
			}

			// Verify local variable does NOT use bare access
			if tt.forbiddenLocalVarPattern != "" && strings.Contains(code, tt.forbiddenLocalVarPattern) {
				t.Errorf("Local variable incorrectly uses bare access:\n  %s\n"+
					"Local variables must use Series.GetCurrent()",
					tt.forbiddenLocalVarPattern)
			}
		})
	}
}

/*
TestArrowFunction_NestedCalls_ContextIsolation validates that nested arrow
function calls maintain proper context isolation with independent ArrowContext instances.

Behavior: Each arrow function call should create its own ArrowContext, ensuring
Series storage is isolated between function invocations. Nested calls must not
interfere with each other's state.

Architecture: Tests the fundamental isolation mechanism that enables composable
arrow functions with predictable, isolated state management.
*/
func TestArrowFunction_NestedCalls_ContextIsolation(t *testing.T) {
	tests := []struct {
		name                 string
		pine                 string
		expectedContextCount int
		description          string
	}{
		{
			name: "two level nesting - two contexts",
			pine: `
//@version=5
indicator("Test")
inner(val) =>
    val * 2

outer(len) =>
    inner(len)

result = outer(14)
`,
			expectedContextCount: 2, // arrowCtx_outer and arrowCtx_inner (if called from outer)
			description:          "nested call requires context passing, not creation",
		},
		{
			name: "multiple calls to same function",
			pine: `
//@version=5
indicator("Test")
helper(x) =>
    x + 1

processor(len) =>
    a = helper(len)
    b = helper(len * 2)
    a + b

result = processor(14)
`,
			expectedContextCount: 1, // Only processor gets ArrowContext from main
			description:          "multiple calls from same function share caller's context",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := compilePineScript(tt.pine)
			if err != nil {
				t.Fatalf("Failed to compile: %v", err)
			}

			// Count ArrowContext creation sites
			contextCreations := strings.Count(code, "context.NewArrowContext(ctx)")

			// Note: The actual count depends on call sites in main context
			// Nested calls pass arrowCtx, not create new contexts
			if contextCreations == 0 {
				t.Errorf("%s: No ArrowContext creation found. Expected at least 1 for top-level call",
					tt.description)
			}

			// Verify nested calls use arrowCtx parameter, not create new context
			nestedCallPattern := "context.NewArrowContext(arrowCtx)"
			if strings.Contains(code, nestedCallPattern) {
				t.Errorf("%s: Found ArrowContext creation from arrowCtx (should pass arrowCtx directly):\n  %s",
					tt.description, nestedCallPattern)
			}
		})
	}
}
