package codegen

import (
	"strings"
	"testing"
)

/*
TestArrowIfStatementCodegen verifies all aspects of arrow-aware if-statement
code generation: dual-storage invariants, branch structure, variable scope,
and condition expression handling.

Scope boundary: this file covers only the behavior of if-statements in non-terminal
UDF body positions.  If-statements used as the final return expression are covered
by arrow_if_statement_return_test.go.  If-statements inside loop bodies are covered
by arrow_function_for_loop_codegen_test.go and arrow_universal_series_test.go.
*/
func TestArrowIfStatementCodegen(t *testing.T) {
	tests := []struct {
		name             string
		pine             string
		mustContainAll   []string
		forbiddenPattern []string
		description      string
	}{
		// ── dual-storage invariant ───────────────────────────────────────────────
		{
			name: "pre-declared variable reassigned in sequential if-chain gets dual storage",
			pine: `//@version=5
indicator("T")
pick(src, mode) =>
    out = 0.0
    if mode == "A"
        out := src + 1.0
        out
    if mode == "B"
        out := src - 1.0
        out
    out
plot(pick(close, "A"))`,
			mustContainAll: []string{
				"out := float64(0)",
				"outSeries.Set(out)",
				"out = ",
			},
			forbiddenPattern: []string{
				"out := (out +", // reassignment must use = not :=
				"out := (out -",
			},
			description: "each branch that reassigns a pre-declared variable emits scalar update + Series.Set",
		},
		{
			name: "variable reassigned in if-branch retains correct value at return site",
			pine: `//@version=5
indicator("T")
applyOffset(src, positive) =>
    result = src
    if positive
        result := src + 10.0
        result
    if na(positive)
        result := 0.0
        result
    result
plot(applyOffset(close, 1))`,
			mustContainAll: []string{
				"result := ",
				"result = ",
				"resultSeries.Set(result)",
				"return result",
			},
			forbiddenPattern: []string{
				"return resultSeries.GetCurrent()", // no loop → scalar return, not series
			},
			description: "scalar is authoritative at return site when variable not loop-modified",
		},

		// ── branch structure ─────────────────────────────────────────────────────
		{
			name: "if-only branch emits no else block",
			pine: `//@version=5
indicator("T")
gate(src, threshold) =>
    out = src
    if threshold > 0
        out := src * 2.0
        out
    out
plot(gate(close, 1))`,
			mustContainAll: []string{
				"if (threshold > 0) {",
				"out = ",
				"outSeries.Set(out)",
			},
			forbiddenPattern: []string{
				"} else {",
				"} else if ",
			},
			description: "if without else clause closes with single brace — no else block emitted",
		},
		{
			name: "if-else emits two arms with dual storage in each",
			pine: `//@version=5
indicator("T")
signed(src) =>
    sign = 0.0
    if src > 0
        sign := 1.0
        sign
    else
        sign := -1.0
        sign
    sign
plot(signed(close))`,
			mustContainAll: []string{
				"} else {",
				"sign = float64(1)",
				"sign = float64(-1)",
				"signSeries.Set(sign)",
			},
			forbiddenPattern: []string{
				"} else if ",
			},
			description: "if-else emits } else { with scalar and series update in both arms",
		},
		{
			name: "if-else-if-else chain emits correct Go branch keywords",
			pine: `//@version=5
indicator("T")
tier(x) =>
    level = 0.0
    if x > 100
        level := 3.0
        level
    else if x > 50
        level := 2.0
        level
    else if x > 10
        level := 1.0
        level
    else
        level := 0.0
        level
    level
plot(tier(close))`,
			mustContainAll: []string{
				"} else if ",
				"} else {",
				"level = float64(3)",
				"level = float64(2)",
				"level = float64(1)",
				"level = float64(0)",
				"levelSeries.Set(level)",
			},
			forbiddenPattern: nil,
			description:      "else-if chain uses } else if { keyword and dual-storage in every arm",
		},

		// ── variable scope in branches ───────────────────────────────────────────
		{
			name: "variable declared inside a branch gets series allocation and dual storage",
			pine: `//@version=5
indicator("T")
process(src) =>
    result = 0.0
    if src > 0
        local = src * 2.0
        result := local
        result
    result
plot(process(close))`,
			mustContainAll: []string{
				`localSeries := arrowCtx.GetOrCreateSeries("local")`,
				"local := ",
				"localSeries.Set(local)",
				"result = local",
				"resultSeries.Set(result)",
			},
			forbiddenPattern: []string{
				"local = ",
			},
			description: "variable first declared inside a branch receives series registration and dual-storage",
		},
		{
			name: "two independent variables reassigned in separate if-branches get independent dual storage",
			pine: `//@version=5
indicator("T")
tally(src) =>
    ups = 0.0
    downs = 0.0
    if src > 0
        ups := ups + 1.0
        ups
    if src < 0
        downs := downs + 1.0
        downs
    ups - downs
plot(tally(close))`,
			mustContainAll: []string{
				"upsSeries := arrowCtx.GetOrCreateSeries(\"ups\")",
				"downsSeries := arrowCtx.GetOrCreateSeries(\"downs\")",
				"ups = (ups + 1)",
				"upsSeries.Set(ups)",
				"downs = (downs + 1)",
				"downsSeries.Set(downs)",
			},
			forbiddenPattern: []string{
				"ups := (ups + ",
				"downs := (downs +",
			},
			description: "two variables each modified in their own if-branch each receive independent dual-storage",
		},
		{
			name: "variable reassigned in else branch gets dual storage",
			pine: `//@version=5
indicator("T")
toggle(src, cond) =>
    val = src
    if cond > 0
        val := src * 2.0
        val
    else
        val := src * 0.5
        val
    val
plot(toggle(close, 1))`,
			mustContainAll: []string{
				"val = (src * 2)",
				"valSeries.Set(val)",
				"} else {",
				"val = (src * 0.5)",
			},
			forbiddenPattern: []string{
				"val := (src * 2)",
				"val := (src * 0.5)",
			},
			description: "else-branch reassignment receives same dual-storage as if-branch",
		},

		// ── nested if ────────────────────────────────────────────────────────────
		{
			name: "if nested inside outer if body routes through arrow-aware path",
			pine: `//@version=5
indicator("T")
clamp(x, lo, hi) =>
    out = x
    if x < lo
        out := lo
        out
    if x > hi
        if out > hi
            out := hi
            out
    out
plot(clamp(close, 10, 100))`,
			mustContainAll: []string{
				"if (x < lo) {",
				"if (x > hi) {",
				"if (out > hi) {",
				"out = lo",
				"outSeries.Set(out)",
				"out = hi",
			},
			forbiddenPattern: []string{
				"out := lo",
				"out := hi",
			},
			description: "inner if inside outer if body also routes through arrow-aware dual-storage path",
		},

		// ── condition expression ─────────────────────────────────────────────────
		{
			name: "comparison condition generates correct Go expression",
			pine: `//@version=5
indicator("T")
threshold(src, level) =>
    above = 0.0
    if src > level
        above := 1.0
        above
    above
plot(threshold(close, 50))`,
			mustContainAll: []string{
				"if (src > level) {",
			},
			forbiddenPattern: nil,
			description:      "comparison expression in if condition emits intact Go comparison without transformation",
		},
		{
			name: "string equality condition generates correct Go expression",
			pine: `//@version=5
indicator("T")
dispatch(src, mode) =>
    out = src
    if mode == "FAST"
        out := src * 2.0
        out
    if mode == "SLOW"
        out := src * 0.5
        out
    out
plot(dispatch(close, "FAST"))`,
			mustContainAll: []string{
				`if (mode == "FAST") {`,
				`if (mode == "SLOW") {`,
			},
			forbiddenPattern: nil,
			description:      "string equality in condition is passed through without modification",
		},

		// ── pure-value expression skip semantics ─────────────────────────────────
		{
			name: "bare identifier expression at branch tail is not emitted as statement",
			pine: `//@version=5
indicator("T")
wrapper(src) =>
    out = src
    if src > 0
        out := src + 1.0
        out
    out
plot(wrapper(close))`,
			mustContainAll: []string{
				"out = (src + 1)",
				"outSeries.Set(out)",
			},
			forbiddenPattern: []string{
				"out\noutSeries",
			},
			description: "bare identifier at branch tail is a Pine implicit-return convention and must be silently dropped",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			goCode, err := compilePineScript(tt.pine)
			if err != nil {
				t.Fatalf("compilePineScript: %v", err)
			}

			for _, pattern := range tt.mustContainAll {
				if !strings.Contains(goCode, pattern) {
					t.Errorf("Missing required pattern: %q\nDescription: %s\nGenerated:\n%s",
						pattern, tt.description, goCode)
				}
			}

			for _, forbidden := range tt.forbiddenPattern {
				if strings.Contains(goCode, forbidden) {
					t.Errorf("Found forbidden pattern: %q\nDescription: %s\nGenerated:\n%s",
						forbidden, tt.description, goCode)
				}
			}
		})
	}
}
