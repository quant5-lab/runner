package integration

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/tests/util"
)

/* Validates control-flow-as-expression (for, while, if) produce correct runtime values */
func TestControlFlowAsExpression(t *testing.T) {
	cases := []struct {
		name     string
		pine     string
		plot     string
		expected float64
	}{
		/* --- for-loop as expression --- */
		{
			name: "ForLoopBareIdentifier",
			pine: `//@version=5
indicator("Test", overlay=false)
x = for i = 1 to 5
    i
plot(x, "Result")`,
			plot:     "Result",
			expected: 5.0,
		},
		{
			name: "ForLoopBinaryExpression",
			pine: `//@version=5
indicator("Test", overlay=false)
x = for i = 1 to 5
    i * 2
plot(x, "Result")`,
			plot:     "Result",
			expected: 10.0,
		},

		/* --- while-loop as expression --- */
		{
			name: "WhileLoopBareIdentifier",
			pine: `//@version=5
indicator("Test", overlay=false)
n = 5
x = while n > 0
    n := n - 1
    n
plot(x, "Result")`,
			plot:     "Result",
			expected: 0.0,
		},
		{
			name: "WhileLoopBinaryExpression",
			pine: `//@version=5
indicator("Test", overlay=false)
n = 3
x = while n > 0
    n := n - 1
    n + 100
plot(x, "Result")`,
			plot:     "Result",
			expected: 100.0,
		},
		{
			name: "WhileLoopAccumulation",
			pine: `//@version=5
indicator("Test", overlay=false)
sum = 0.0
i = 1
x = while i <= 10
    sum := sum + i
    i := i + 1
    sum
plot(x, "Result")`,
			plot:     "Result",
			expected: 55.0,
		},

		/* --- if as expression (consequent path only; else-branch codegen is a known limitation) --- */
		{
			name: "IfExpressionLiteral",
			pine: `//@version=5
indicator("Test", overlay=false)
x = if true
    42.0
plot(x, "Result")`,
			plot:     "Result",
			expected: 42.0,
		},
		{
			name: "IfExpressionSeriesVariable",
			pine: `//@version=5
indicator("Test", overlay=false)
val = 7.0
x = if true
    val
plot(x, "Result")`,
			plot:     "Result",
			expected: 7.0,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			exec := util.NewPineExecutor(t)
			output := exec.ExecuteScript(t, "cfexpr-"+strings.ToLower(tc.name), tc.pine)

			vals := exec.ExtractPlotValues(t, output, tc.plot)
			if len(vals) < 1 {
				t.Fatal("Expected at least 1 data point")
			}
			if vals[0] != tc.expected {
				t.Errorf("got %f, want %f", vals[0], tc.expected)
			}
		})
	}
}
