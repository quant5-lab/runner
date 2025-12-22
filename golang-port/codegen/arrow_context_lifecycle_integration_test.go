package codegen

import (
	"strings"
	"testing"
)

/* Tests unique ArrowContext allocation preventing variable redeclaration */
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

/* Tests Series.Set() generation for return values maintaining PineScript historical access semantics */
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

/* Tests complete ArrowContext lifecycle: create → call → store → advance */
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

/* Tests complex patterns: deep nesting, large tuples, parameterless functions, production patterns */
func TestUserDefinedFunction_ComplexCallPatterns(t *testing.T) {
	tests := []struct {
		name                 string
		pine                 string
		expectedContexts     []string
		expectedStorageStmts []string
		forbiddenPatterns    []string
		description          string
	}{
		{
			name: "large tuple return (5 values)",
			pine: `
//@version=5
indicator("Test")
quintuple() =>
    [close, open, high, low, volume]

[a, b, c, d, e] = quintuple()
`,
			expectedContexts: []string{
				"arrowCtx_quintuple_1 := context.NewArrowContext(ctx)",
			},
			expectedStorageStmts: []string{
				"aSeries.Set(a)",
				"bSeries.Set(b)",
				"cSeries.Set(c)",
				"dSeries.Set(d)",
				"eSeries.Set(e)",
			},
			forbiddenPatterns: nil,
			description:       "large tuple return creates storage for all values",
		},
		{
			name: "function with no parameters",
			pine: `
//@version=5
indicator("Test")
constant() =>
    42.0

value = constant()
`,
			expectedContexts: []string{
				"arrowCtx_constant_1 := context.NewArrowContext(ctx)",
			},
			expectedStorageStmts: []string{
				"valueSeries.Set(",
			},
			forbiddenPatterns: nil,
			description:       "parameterless function allocates context",
		},
		{
			name: "function called multiple times in sequence",
			pine: `
//@version=5
indicator("Test")
increment(x) =>
    x + 1

a = increment(10)
b = increment(20)
c = increment(30)
d = increment(40)
e = increment(50)
`,
			expectedContexts: []string{
				"arrowCtx_increment_1",
				"arrowCtx_increment_2",
				"arrowCtx_increment_3",
				"arrowCtx_increment_4",
				"arrowCtx_increment_5",
			},
			expectedStorageStmts: []string{
				"aSeries.Set(",
				"bSeries.Set(",
				"cSeries.Set(",
				"dSeries.Set(",
				"eSeries.Set(",
			},
			forbiddenPatterns: []string{
				"arrowCtx_increment := context.NewArrowContext",
			},
			description: "many sequential calls create unique contexts",
		},
		{
			name: "mixed parameter types",
			pine: `
//@version=5
indicator("Test")
mixer(scalar, src) =>
    src * scalar

result = mixer(2.0, close)
`,
			expectedContexts: []string{
				"arrowCtx_mixer_1 := context.NewArrowContext(ctx)",
			},
			expectedStorageStmts: []string{
				"resultSeries.Set(",
			},
			forbiddenPatterns: nil,
			description:       "mixed scalar and series parameters work correctly",
		},
		{
			name: "interleaved multiple functions",
			pine: `
//@version=5
indicator("Test")
double(x) =>
    x * 2

triple(x) =>
    x * 3

a1 = double(10)
t1 = triple(10)
a2 = double(20)
t2 = triple(20)
a3 = double(30)
`,
			expectedContexts: []string{
				"arrowCtx_double_1",
				"arrowCtx_triple_1",
				"arrowCtx_double_2",
				"arrowCtx_triple_2",
				"arrowCtx_double_3",
			},
			expectedStorageStmts: []string{
				"a1Series.Set(",
				"t1Series.Set(",
				"a2Series.Set(",
				"t2Series.Set(",
				"a3Series.Set(",
			},
			forbiddenPatterns: nil,
			description:       "interleaved calls maintain independent counters",
		},
		{
			name: "tuple and single value mixed",
			pine: `
//@version=5
indicator("Test")
pair() =>
    [close, open]

single() =>
    high

[a, b] = pair()
c = single()
[d, e] = pair()
f = single()
`,
			expectedContexts: []string{
				"arrowCtx_pair_1",
				"arrowCtx_single_1",
				"arrowCtx_pair_2",
				"arrowCtx_single_2",
			},
			expectedStorageStmts: []string{
				"aSeries.Set(a)",
				"bSeries.Set(b)",
				"cSeries.Set(",
				"dSeries.Set(d)",
				"eSeries.Set(e)",
				"fSeries.Set(",
			},
			forbiddenPatterns: nil,
			description:       "mixed tuple and single value calls handled correctly",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := compilePineScript(tt.pine)
			if err != nil {
				t.Fatalf("Compilation failed: %v", err)
			}

			for _, expectedCtx := range tt.expectedContexts {
				if !strings.Contains(code, expectedCtx) {
					t.Errorf("%s: Missing expected context:\n  %s", tt.description, expectedCtx)
				}
			}

			for _, expectedStorage := range tt.expectedStorageStmts {
				if !strings.Contains(code, expectedStorage) {
					t.Errorf("%s: Missing expected storage statement:\n  %s", tt.description, expectedStorage)
				}
			}

			for _, forbidden := range tt.forbiddenPatterns {
				if strings.Contains(code, forbidden) {
					t.Errorf("%s: Found forbidden pattern:\n  %s", tt.description, forbidden)
				}
			}
		})
	}
}

/* Regression tests for production patterns (bb7-dissect-adx, dirmov) ensuring stability */
func TestUserDefinedFunction_RegressionSafety(t *testing.T) {
	tests := []struct {
		name             string
		pine             string
		criticalPatterns []string
		description      string
	}{
		{
			name: "bb7-dissect-adx pattern (multiple ADX calls)",
			pine: `
//@version=5
indicator("BB7 Pattern")
adx_calc(dilength, adxlength) =>
    [ta.sma(close, dilength), ta.sma(open, adxlength), ta.sma(high, dilength)]

[ADX, up, down] = adx_calc(14, 14)
[ADX2, up2, down2] = adx_calc(21, 21)
`,
			criticalPatterns: []string{
				"arrowCtx_adx_calc_1 := context.NewArrowContext(ctx)",
				"arrowCtx_adx_calc_2 := context.NewArrowContext(ctx)",
				"ADXSeries.Set(ADX)",
				"upSeries.Set(up)",
				"downSeries.Set(down)",
				"ADX2Series.Set(ADX2)",
				"up2Series.Set(up2)",
				"down2Series.Set(down2)",
				"arrowCtx_adx_calc_1.AdvanceAll()",
				"arrowCtx_adx_calc_2.AdvanceAll()",
			},
			description: "bb7-dissect-adx pattern must generate unique contexts and complete storage",
		},
		{
			name: "dirmov pattern (nested TA calls)",
			pine: `
//@version=5
indicator("Dirmov Pattern")
dirmov(len) =>
    up = ta.change(high)
    down = -ta.change(low)
    [up, down]

[plus, minus] = dirmov(14)
`,
			criticalPatterns: []string{
				"arrowCtx_dirmov_1 := context.NewArrowContext(ctx)",
				"plusSeries.Set(plus)",
				"minusSeries.Set(minus)",
				"arrowCtx_dirmov_1.AdvanceAll()",
			},
			description: "dirmov pattern with nested TA calls must work correctly",
		},
		{
			name: "multiple functions multiple calls",
			pine: `
//@version=5
indicator("Complex Pattern")
calc_a(x) =>
    ta.sma(close, x)

calc_b(y) =>
    ta.ema(open, y)

r1 = calc_a(10)
r2 = calc_b(20)
r3 = calc_a(30)
r4 = calc_b(40)
`,
			criticalPatterns: []string{
				"arrowCtx_calc_a_1",
				"arrowCtx_calc_b_1",
				"arrowCtx_calc_a_2",
				"arrowCtx_calc_b_2",
				"r1Series.Set(",
				"r2Series.Set(",
				"r3Series.Set(",
				"r4Series.Set(",
			},
			description: "multiple functions with multiple calls each maintain separate counters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := compilePineScript(tt.pine)
			if err != nil {
				t.Fatalf("%s: Compilation failed: %v", tt.description, err)
			}

			for _, pattern := range tt.criticalPatterns {
				if !strings.Contains(code, pattern) {
					t.Errorf("%s: REGRESSION - Missing critical pattern:\n  %s\n\nThis pattern MUST exist for production code to work correctly.",
						tt.description, pattern)
				}
			}

			forbiddenPatterns := []string{
				"arrowCtx_calc_a := context.NewArrowContext",
				"arrowCtx_calc_b := context.NewArrowContext",
				"arrowCtx_adx_calc := context.NewArrowContext",
				"arrowCtx_dirmov := context.NewArrowContext",
			}

			for _, pattern := range forbiddenPatterns {
				if strings.Contains(code, pattern) {
					t.Errorf("%s: REGRESSION - Found non-unique context pattern:\n  %s\n\nThis indicates context variable redeclaration bug has returned.",
						tt.description, pattern)
				}
			}
		})
	}
}
