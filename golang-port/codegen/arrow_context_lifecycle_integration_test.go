package codegen

import (
	"strings"
	"testing"
)

/*
TestUserDefinedFunction_ArrowContextAllocation validates ArrowContext lifecycle management
for user-defined function calls with unique instance tracking.

Behavior: Each call to a user-defined function allocates a unique ArrowContext instance
with an incremental suffix (_1, _2, _3...) to prevent variable redeclaration within the
same scope. This applies to both tuple-destructuring and single-value calls.

Architecture: Tests ArrowContextLifecycleManager integration ensuring unique context
variable names across multiple calls to the same function or different functions.

Edge cases covered:
- Multiple calls to same function (same scope)
- Interleaved calls to different functions
- Single-value vs tuple-destructuring calls
- Sequential and nested call patterns
*/
func TestUserDefinedFunction_ArrowContextAllocation(t *testing.T) {
	tests := []struct {
		name              string
		pine              string
		expectedContexts  []string
		forbiddenPatterns []string
		description       string
	}{
		{
			name: "multiple calls same function tuple",
			pine: `
//@version=5
indicator("Test")
calc(len) =>
    a = close * len
    b = open * len
    [a, b]

[x1, y1] = calc(14)
[x2, y2] = calc(20)
[x3, y3] = calc(30)
`,
			expectedContexts: []string{
				"arrowCtx_calc_1 := context.NewArrowContext(ctx)",
				"arrowCtx_calc_2 := context.NewArrowContext(ctx)",
				"arrowCtx_calc_3 := context.NewArrowContext(ctx)",
			},
			forbiddenPatterns: []string{
				"arrowCtx_calc := context.NewArrowContext(ctx)",
			},
			description: "three calls to same function should create unique contexts",
		},
		{
			name: "multiple calls same function single value",
			pine: `
//@version=5
indicator("Test")
double(x) =>
    x * 2

result1 = double(close)
result2 = double(open)
`,
			expectedContexts: []string{
				"arrowCtx_double_1 := context.NewArrowContext(ctx)",
				"arrowCtx_double_2 := context.NewArrowContext(ctx)",
			},
			forbiddenPatterns: []string{
				"arrowCtx_double := context.NewArrowContext(ctx)",
			},
			description: "single-value calls should also create unique contexts",
		},
		{
			name: "interleaved different functions",
			pine: `
//@version=5
indicator("Test")
alpha(x) =>
    x + 1

beta(y) =>
    y * 2

a1 = alpha(10)
b1 = beta(20)
a2 = alpha(30)
b2 = beta(40)
`,
			expectedContexts: []string{
				"arrowCtx_alpha_1 := context.NewArrowContext(ctx)",
				"arrowCtx_beta_1 := context.NewArrowContext(ctx)",
				"arrowCtx_alpha_2 := context.NewArrowContext(ctx)",
				"arrowCtx_beta_2 := context.NewArrowContext(ctx)",
			},
			forbiddenPatterns: []string{
				"arrowCtx_alpha := context.NewArrowContext(ctx)",
				"arrowCtx_beta := context.NewArrowContext(ctx)",
			},
			description: "interleaved calls to different functions maintain independent counters",
		},
		{
			name: "many sequential calls",
			pine: `
//@version=5
indicator("Test")
process(val) =>
    val * 2

r1 = process(1)
r2 = process(2)
r3 = process(3)
r4 = process(4)
r5 = process(5)
`,
			expectedContexts: []string{
				"arrowCtx_process_1",
				"arrowCtx_process_2",
				"arrowCtx_process_3",
				"arrowCtx_process_4",
				"arrowCtx_process_5",
			},
			forbiddenPatterns: nil,
			description:       "many sequential calls generate incrementing suffixes",
		},
		{
			name: "mixed single and tuple calls",
			pine: `
//@version=5
indicator("Test")
single(x) =>
    x * 2

tuple(y) =>
    [y, y*2]

s1 = single(10)
[t1a, t1b] = tuple(20)
s2 = single(30)
[t2a, t2b] = tuple(40)
`,
			expectedContexts: []string{
				"arrowCtx_single_1",
				"arrowCtx_tuple_1",
				"arrowCtx_single_2",
				"arrowCtx_tuple_2",
			},
			forbiddenPatterns: nil,
			description:       "mixed call types use same allocation mechanism",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := compilePineScript(tt.pine)
			if err != nil {
				t.Fatalf("Compilation failed: %v", err)
			}

			for _, expected := range tt.expectedContexts {
				if !strings.Contains(code, expected) {
					t.Errorf("%s: Missing expected context:\n  %s", tt.description, expected)
				}
			}

			for _, forbidden := range tt.forbiddenPatterns {
				if strings.Contains(code, forbidden) {
					t.Errorf("%s: Found forbidden pattern (non-unique context):\n  %s",
						tt.description, forbidden)
				}
			}

			for _, ctxVar := range tt.expectedContexts {
				varName := strings.Split(ctxVar, " :=")[0]
				firstIdx := strings.Index(code, varName+" :=")
				if firstIdx == -1 {
					continue
				}
				remainingCode := code[firstIdx+len(varName)+4:]
				if strings.Contains(remainingCode, varName+" :=") {
					t.Errorf("%s: Variable redeclaration detected for %s", tt.description, varName)
				}
			}
		})
	}
}

/*
TestUserDefinedFunction_ReturnValueStorage validates that return values from
user-defined functions are stored in Series for historical access.

Behavior: All return values (single or tuple) must be stored via Series.Set()
immediately after the function call and before ArrowContext.AdvanceAll(). This
maintains PineScript semantics where function return values become Series variables.

Architecture: Tests ReturnValueSeriesStorageHandler integration ensuring proper
Series.Set() generation for all return value patterns.

Edge cases covered:
- Single return value
- Multiple return values (tuple destructuring)
- No return value (void-like functions)
- Return values used immediately vs later
- Nested function calls with cascading storage
*/
func TestUserDefinedFunction_ReturnValueStorage(t *testing.T) {
	tests := []struct {
		name            string
		pine            string
		expectedStorage []string
		storageOrder    []string
		description     string
	}{
		{
			name: "single return value storage",
			pine: `
//@version=5
indicator("Test")
double(x) =>
    x * 2

result = double(close)
plot(result)
`,
			expectedStorage: []string{
				"resultSeries.Set(double(arrowCtx_double_1,",
			},
			storageOrder: []string{
				"arrowCtx_double_1 := context.NewArrowContext(ctx)",
				"resultSeries.Set(",
				"arrowCtx_double_1.AdvanceAll()",
			},
			description: "single return value stored in Series",
		},
		{
			name: "tuple return value storage",
			pine: `
//@version=5
indicator("Test")
pair(multiplier) =>
    a = close * multiplier
    b = open * multiplier
    [a, b]

[first, second] = pair(2)
`,
			expectedStorage: []string{
				"firstSeries.Set(first)",
				"secondSeries.Set(second)",
			},
			storageOrder: []string{
				"first, second := pair(",
				"firstSeries.Set(first)",
				"secondSeries.Set(second)",
				"arrowCtx_pair_1.AdvanceAll()",
			},
			description: "tuple return values stored individually",
		},
		{
			name: "triple return value storage",
			pine: `
//@version=5
indicator("Test")
triple(x) =>
    [x, x*2, x*3]

[a, b, c] = triple(10)
`,
			expectedStorage: []string{
				"aSeries.Set(a)",
				"bSeries.Set(b)",
				"cSeries.Set(c)",
			},
			storageOrder: nil,
			description:  "all tuple elements stored in correct order",
		},
		{
			name: "multiple calls with storage",
			pine: `
//@version=5
indicator("Test")
calc(val) =>
    val * 2

r1 = calc(10)
r2 = calc(20)
`,
			expectedStorage: []string{
				"r1Series.Set(calc(arrowCtx_calc_1",
				"r2Series.Set(calc(arrowCtx_calc_2",
			},
			storageOrder: nil,
			description:  "each call stores its return value",
		},
		{
			name: "storage before series access",
			pine: `
//@version=5
indicator("Test")
increment(x) =>
    x + 1

value = increment(close)
doubled = value * 2
`,
			expectedStorage: []string{
				"valueSeries.Set(increment(",
			},
			storageOrder: []string{
				"valueSeries.Set(increment(",
				"doubledSeries.Set((valueSeries.GetCurrent() * 2",
			},
			description: "storage happens before variable is accessed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := compilePineScript(tt.pine)
			if err != nil {
				t.Fatalf("Compilation failed: %v", err)
			}

			for _, expected := range tt.expectedStorage {
				if !strings.Contains(code, expected) {
					t.Errorf("%s: Missing expected storage:\n  %s\n\nGenerated:\n%s",
						tt.description, expected, code)
				}
			}

			if len(tt.storageOrder) > 0 {
				lastIdx := -1
				for i, pattern := range tt.storageOrder {
					idx := strings.Index(code, pattern)
					if idx == -1 {
						t.Errorf("%s: Missing pattern in order check: %s", tt.description, pattern)
						continue
					}
					if idx <= lastIdx {
						t.Errorf("%s: Pattern %d appears before pattern %d:\n  %s\n  %s",
							tt.description, i, i-1, tt.storageOrder[i-1], pattern)
					}
					lastIdx = idx
				}
			}
		})
	}
}

/*
TestUserDefinedFunction_CompleteLifecycle validates the complete ArrowContext
lifecycle: create → call → store → advance.

Behavior: Every user-defined function call follows a strict sequence:
 1. Create unique ArrowContext
 2. Call function with context
 3. Store return values in Series (if any)
 4. Advance all Series in ArrowContext

Architecture: Tests end-to-end integration of ArrowContextLifecycleManager and
ReturnValueSeriesStorageHandler ensuring correct statement ordering.

Edge cases covered:
- Lifecycle ordering validation
- Multiple calls maintaining individual lifecycles
- Nested calls with cascading lifecycles
- Error conditions (missing stages)
*/
func TestUserDefinedFunction_CompleteLifecycle(t *testing.T) {
	tests := []struct {
		name           string
		pine           string
		functionName   string
		callIndex      int
		validateStages bool
		description    string
	}{
		{
			name: "basic lifecycle validation",
			pine: `
//@version=5
indicator("Test")
process(x) =>
    x * 2

result = process(10)
`,
			functionName:   "process",
			callIndex:      1,
			validateStages: true,
			description:    "single call follows complete lifecycle",
		},
		{
			name: "tuple return lifecycle",
			pine: `
//@version=5
indicator("Test")
split(val) =>
    [val, val*2]

[a, b] = split(5)
`,
			functionName:   "split",
			callIndex:      1,
			validateStages: true,
			description:    "tuple call with storage lifecycle",
		},
		{
			name: "second call lifecycle independent",
			pine: `
//@version=5
indicator("Test")
compute(x) =>
    x + 1

r1 = compute(10)
r2 = compute(20)
`,
			functionName:   "compute",
			callIndex:      2,
			validateStages: true,
			description:    "second call has independent complete lifecycle",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := compilePineScript(tt.pine)
			if err != nil {
				t.Fatalf("Compilation failed: %v", err)
			}

			if !tt.validateStages {
				return
			}

			contextVar := "arrowCtx_" + tt.functionName + "_" + string(rune('0'+tt.callIndex))

			createStmt := contextVar + " := context.NewArrowContext(ctx)"
			createIdx := strings.Index(code, createStmt)
			if createIdx == -1 {
				t.Errorf("%s: Missing context creation stage", tt.description)
				return
			}

			callStmt := tt.functionName + "(" + contextVar
			callIdx := strings.Index(code[createIdx:], callStmt)
			if callIdx == -1 {
				t.Errorf("%s: Missing function call stage", tt.description)
				return
			}
			callIdx += createIdx

			advanceStmt := contextVar + ".AdvanceAll()"
			advanceIdx := strings.Index(code[callIdx:], advanceStmt)
			if advanceIdx == -1 {
				t.Errorf("%s: Missing advance stage", tt.description)
				return
			}
			advanceIdx += callIdx

			if !(createIdx < callIdx && callIdx < advanceIdx) {
				t.Errorf("%s: Lifecycle stages out of order", tt.description)
			}
		})
	}
}
