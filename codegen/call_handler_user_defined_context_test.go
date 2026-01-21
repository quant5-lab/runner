package codegen

import (
	"strings"
	"testing"
)

/*
TestUserDefinedFunctionCalls_ContextAllocationConsistency validates ArrowContext allocation across statement types.

Integration test between ArrowCallSiteScanner (hoisting) and UserDefinedFunctionHandler (call generation).
Ensures every call site receives the correct arrowCtx_{func}_{N} variable.

Coverage: VariableDeclaration, IfStatement, BinaryExpression, ConditionalExpression, nested calls.
*/
func TestUserDefinedFunctionCalls_ContextAllocationConsistency(t *testing.T) {
	tests := []struct {
		name           string
		pine           string
		mustContain    []string
		mustNotContain []string
		description    string
	}{
		{
			name: "call in if condition",
			pine: `
//@version=5
strategy("Test")
myFunc(x) =>
    x * 2
if myFunc(10) > 15
    strategy.entry("Long", strategy.long)
`,
			mustContain: []string{
				"arrowCtx_myFunc_1 := context.NewArrowContext(ctx)",
				"if (myFunc(arrowCtx_myFunc_1, 10.0) > 15)",
			},
			mustNotContain: []string{
				"myFunc(ctx,",
			},
			description: "if condition call uses hoisted arrowCtx, not raw ctx",
		},
		{
			name: "multiple calls same function",
			pine: `
//@version=5
indicator("Test")
calc(x) =>
    x * 2
a = calc(5)
b = calc(10)
if calc(15) > 20
    calc(25)
`,
			mustContain: []string{
				"arrowCtx_calc_1 := context.NewArrowContext(ctx)",
				"arrowCtx_calc_2 := context.NewArrowContext(ctx)",
				"arrowCtx_calc_3 := context.NewArrowContext(ctx)",
				"arrowCtx_calc_4 := context.NewArrowContext(ctx)",
				"aSeries.Set(calc(arrowCtx_calc_1, 5.0))",
				"bSeries.Set(calc(arrowCtx_calc_2, 10.0))",
				"if (calc(arrowCtx_calc_3, 15.0) > 20)",
				"calc(arrowCtx_calc_4, 25.0)",
			},
			mustNotContain: []string{
				"calc(ctx,",
			},
			description: "each call gets unique incrementing arrowCtx",
		},
		{
			name: "nested function calls",
			pine: `
//@version=5
indicator("Test")
inner(x) =>
    x + 1
outer(y) =>
    inner(y) * 2
result = outer(10)
`,
			mustContain: []string{
				"arrowCtx_outer_1 := context.NewArrowContext(ctx)",
				"resultSeries.Set(outer(arrowCtx_outer_1, 10.0))",
				"inner(arrowCtx,",
			},
			mustNotContain: []string{
				"outer(ctx,",
			},
			description: "nested calls maintain independent context allocation",
		},
		{
			name: "calls in binary expression",
			pine: `
//@version=5
indicator("Test")
left(x) =>
    x * 2
right(y) =>
    y + 3
result = left(5) + right(10)
`,
			mustContain: []string{
				"arrowCtx_left_1 := context.NewArrowContext(ctx)",
				"arrowCtx_right_1 := context.NewArrowContext(ctx)",
				"leftSeries.GetCurrent()",
				"rightSeries.GetCurrent()",
			},
			mustNotContain: []string{
				"left(ctx,",
				"right(ctx,",
			},
			description: "binary expression operands use correct contexts",
		},
		{
			name: "calls in conditional expression",
			pine: `
//@version=5
indicator("Test")
trueBranch(x) =>
    x * 2
falseBranch(y) =>
    y + 5
result = close > open ? trueBranch(10) : falseBranch(20)
`,
			mustContain: []string{
				"arrowCtx_trueBranch_1 := context.NewArrowContext(ctx)",
				"arrowCtx_falseBranch_1 := context.NewArrowContext(ctx)",
				"trueBranch(arrowCtx_trueBranch_1, 10.0)",
				"falseBranch(arrowCtx_falseBranch_1, 20.0)",
			},
			mustNotContain: []string{
				"trueBranch(ctx,",
				"falseBranch(ctx,",
			},
			description: "ternary branches use hoisted contexts",
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
`,
			mustContain: []string{
				"arrowCtx_alpha_1 := context.NewArrowContext(ctx)",
				"arrowCtx_beta_1 := context.NewArrowContext(ctx)",
				"arrowCtx_alpha_2 := context.NewArrowContext(ctx)",
				"a1Series.Set(alpha(arrowCtx_alpha_1, 10.0))",
				"b1Series.Set(beta(arrowCtx_beta_1, 20.0))",
				"a2Series.Set(alpha(arrowCtx_alpha_2, 30.0))",
			},
			mustNotContain: []string{
				"alpha(ctx,",
				"beta(ctx,",
			},
			description: "independent function counters maintained correctly",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := compilePineScript(tt.pine)
			if err != nil {
				t.Fatalf("Compilation failed: %v", err)
			}

			for _, pattern := range tt.mustContain {
				if !strings.Contains(code, pattern) {
					t.Errorf("Missing expected pattern: %q\nDescription: %s\nGenerated code:\n%s",
						pattern, tt.description, code)
				}
			}

			for _, pattern := range tt.mustNotContain {
				if strings.Contains(code, pattern) {
					t.Errorf("Found forbidden pattern: %q\nDescription: %s\nGenerated code:\n%s",
						pattern, tt.description, code)
				}
			}
		})
	}
}

/*
TestUserDefinedFunctionCalls_ArgumentTypeHandling validates argument expression handling.

Coverage: scalar literals, series identifiers, nested expressions, multiple arguments.
*/
func TestUserDefinedFunctionCalls_ArgumentTypeHandling(t *testing.T) {
	tests := []struct {
		name        string
		pine        string
		mustContain []string
		description string
	}{
		{
			name: "scalar literal arguments",
			pine: `
//@version=5
indicator("Test")
multi(a, b, c) =>
    a + b + c
result = multi(10.5, 20, 30.0)
`,
			mustContain: []string{
				"multi(arrowCtx_multi_1, 10.5, 20.0, 30.0)",
			},
			description: "scalar literals passed directly",
		},
		{
			name: "series identifier arguments",
			pine: `
//@version=5
indicator("Test")
combine(src1, src2) =>
    src1 + src2
result = combine(close, open)
`,
			mustContain: []string{
				"combine(arrowCtx_combine_1, closeSeries.Get(0), openSeries.Get(0))",
			},
			description: "series identifiers resolve to series variables",
		},
		{
			name: "expression arguments",
			pine: `
//@version=5
indicator("Test")
calc(x) =>
    x * 2
result = calc(close * 1.5)
`,
			mustContain: []string{
				"calc(arrowCtx_calc_1, (closeSeries.Get(0) * 1.5))",
			},
			description: "expressions evaluated before passing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := compilePineScript(tt.pine)
			if err != nil {
				t.Fatalf("Compilation failed: %v", err)
			}

			for _, pattern := range tt.mustContain {
				if !strings.Contains(code, pattern) {
					t.Errorf("Missing expected pattern: %q\nDescription: %s\nGenerated code:\n%s",
						pattern, tt.description, code)
				}
			}
		})
	}
}

/*
TestUserDefinedFunctionCalls_AdvancementConsistency validates AdvanceAll() generation for hoisted contexts.
*/
func TestUserDefinedFunctionCalls_AdvancementConsistency(t *testing.T) {
	pine := `
//@version=5
indicator("Test")
calc(x) =>
    x * 2
a = calc(10)
b = calc(20)
c = calc(30)
`

	code, err := compilePineScript(pine)
	if err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}

	expectedAdvancements := []string{
		"arrowCtx_calc_1.AdvanceAll()",
		"arrowCtx_calc_2.AdvanceAll()",
		"arrowCtx_calc_3.AdvanceAll()",
	}

	for _, advancement := range expectedAdvancements {
		if !strings.Contains(code, advancement) {
			t.Errorf("Missing context advancement: %q\nGenerated code:\n%s", advancement, code)
		}
	}
}
