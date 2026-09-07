package codegen

import (
	"strings"
	"testing"
)

/* TestUniversalSeriesParadigm validates 100% support for ANY ARBITRARY PineScript pattern
 *
 * This test verifies the Universal ForwardSeriesBuffer paradigm:
 * - ALL variables get Series storage (function-scope, loop-scope, if-scope)
 * - Loop-modified variables use Series.GetCurrent() for reads
 * - Nested loops with complex scope chains work correctly
 */
func TestUniversalSeriesParadigm(t *testing.T) {
	tests := []struct {
		name             string
		pine             string
		mustContainAll   []string
		forbiddenPattern []string
		description      string
	}{
		{
			name: "deeply_nested_loops_three_levels",
			pine: `
//@version=5
indicator("Deep Nesting")
calc(len) =>
    a = 0.0
    for i = 0 to len - 1
        b = 0.0
        for j = 0 to 2
            c = 0.0
            for k = 0 to 1
                c := c + 1
            b := b + c
        a := a + b
    a
plot(calc(3))
`,
			mustContainAll: []string{
				"aSeries := arrowCtx.GetOrCreateSeries(\"a\")",
				"bSeries := arrowCtx.GetOrCreateSeries(\"b\")",
				"cSeries := arrowCtx.GetOrCreateSeries(\"c\")",
				"c = (cSeries.GetCurrent() + 1)",
				"b = (bSeries.GetCurrent() + cSeries.GetCurrent())",
				"a = (aSeries.GetCurrent() + bSeries.GetCurrent())",
			},
			forbiddenPattern: []string{
				"c := (c + 1)", // Should use = not :=
				"b := (b + c)", // Should use = not :=
				"a := (a + b)", // Should use = not :=
			},
			description: "three-level nested loops use Series at all levels",
		},
		{
			name: "loop_with_if_statement_variable",
			pine: `
//@version=5
indicator("Loop If Var")
count(len) =>
    result = 0.0
    for i = 0 to len - 1
        if close[i] > open[i]
            gain = close[i] - open[i]
            result := result + gain
    result
plot(count(10))
`,
			mustContainAll: []string{
				"resultSeries := arrowCtx.GetOrCreateSeries(\"result\")",
				"gainSeries := arrowCtx.GetOrCreateSeries(\"gain\")",
				"result = (resultSeries.GetCurrent() + gain)",
				"resultSeries.Set(result)",
			},
			forbiddenPattern: []string{
				"resultSeries.Set((result + gain))", // Should use Series.GetCurrent()
			},
			description: "variables declared in if-statements inside loops get dual-storage (scalar + Series)",
		},
		{
			name: "multiple_loops_same_variable",
			pine: `
//@version=5
indicator("Multi Loop")
process(len) =>
    sum = 0.0
    for i = 0 to len - 1
        sum := sum + close[i]
    for j = 0 to len - 1
        sum := sum - open[j]
    sum
plot(process(5))
`,
			mustContainAll: []string{
				"sumSeries := arrowCtx.GetOrCreateSeries(\"sum\")",
				"sum = (sumSeries.GetCurrent() + ",
				"sum = (sumSeries.GetCurrent() - ",
			},
			forbiddenPattern: []string{
				"sum := (sum +", // Should use = not :=
				"sum := (sum -", // Should use = not :=
			},
			description: "same variable modified in multiple sequential loops uses Series",
		},
		{
			name: "conditional_declaration_before_loop",
			pine: `
//@version=5
indicator("Cond Decl")
adjust(threshold) =>
    if threshold > 0
        offset = 10.0
    else
        offset = -10.0
    result = 0.0
    for i = 0 to 5
        result := result + offset
    result
plot(adjust(1))
`,
			mustContainAll: []string{
				"offsetSeries := arrowCtx.GetOrCreateSeries(\"offset\")",
				"resultSeries := arrowCtx.GetOrCreateSeries(\"result\")",
				"result = (resultSeries.GetCurrent() + offset)",
			},
			forbiddenPattern: []string{
				"result := (result + offset)", // Should use = not :=
			},
			description: "variables declared in if-statement before loop work correctly",
		},
		{
			name: "loop_local_not_series",
			pine: `
//@version=5
indicator("Loop Local")
calc(len) =>
    total = 0.0
    for i = 0 to len - 1
        temp = close[i] * 2
        total := total + temp
    total
plot(calc(5))
`,
			mustContainAll: []string{
				"totalSeries := arrowCtx.GetOrCreateSeries(\"total\")",
				"tempSeries := arrowCtx.GetOrCreateSeries(\"temp\")",
				"total = (totalSeries.GetCurrent() + temp)",
			},
			forbiddenPattern: []string{
				"temp := (tempSeries.GetCurrent()", // temp is loop-local, uses scalar
				"total := (total",                  // Should use = not :=
			},
			description: "loop-local variable gets Series but uses scalar (not modified)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			goCode, err := compilePineScript(tt.pine)
			if err != nil {
				t.Fatalf("Compilation failed: %v", err)
			}

			for _, pattern := range tt.mustContainAll {
				if !strings.Contains(goCode, pattern) {
					t.Errorf("Missing required pattern: %q\nDescription: %s\nGenerated code:\n%s",
						pattern, tt.description, goCode)
				}
			}

			for _, forbidden := range tt.forbiddenPattern {
				if strings.Contains(goCode, forbidden) {
					t.Errorf("Found forbidden pattern: %q\nDescription: %s\nGenerated code:\n%s",
						forbidden, tt.description, goCode)
				}
			}
		})
	}
}
