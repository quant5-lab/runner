//go:build integration

package integration

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/tests/util"
)

/*
Validates variables declared inside if-block bodies receive Series declarations at function scope.

	Regression: first-pass codegen previously skipped IfStatement bodies, leaving if-block variable
	declarations unregistered — causing Go compilation failures (undefined xSeries).
*/
func TestIfBlockVariablePromotion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		pine string
	}{
		{
			name: "single variable in if-body",
			pine: `//@version=5
strategy("Test")
if close > open
    x = close + 1`,
		},
		{
			name: "variables in both if and else",
			pine: `//@version=5
strategy("Test")
if close > open
    a = close
else
    b = open`,
		},
		{
			name: "if-else-if chain",
			pine: `//@version=5
strategy("Test")
if close > 110
    a = close
else if close > 100
    b = open
else
    c = high`,
		},
		{
			name: "if-block variable used in expression",
			pine: `//@version=5
strategy("Test")
if close > open
    delta = close - open
    pct = delta / open * 100`,
		},
		{
			name: "multiple declarations in same if-body",
			pine: `//@version=5
strategy("Test")
if close > open
    m = close * 2
    n = open / 2`,
		},
		{
			name: "if-block with arithmetic expression",
			pine: `//@version=5
strategy("Test")
if close > open
    delta = (close - open) / open * 100`,
		},
	}

	exec := util.NewPineExecutor(t)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			code, _ := exec.GenerateCode(t, "if-block-promo", tc.pine)

			if !strings.Contains(code, "Series") {
				t.Fatal("generated code missing Series declarations")
			}

			if err := exec.CompileCode(t, code); err != nil {
				t.Fatalf("compilation failed (variable not promoted): %v", err)
			}
		})
	}
}
