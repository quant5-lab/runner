//go:build integration

package integration

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/tests/util"
)

func TestWhileLoop(t *testing.T) {
	cases := []struct {
		name     string
		pine     string
		plot     string
		expected float64
	}{
		{
			name: "BasicSum",
			pine: `//@version=5
indicator("While Basic Sum", overlay=false)
sum = 0.0
i = 1
while i <= 10
    sum := sum + i
    i := i + 1
plot(sum, "Result")`,
			plot:     "Result",
			expected: 55.0,
		},
		{
			name: "Countdown",
			pine: `//@version=5
indicator("While Countdown", overlay=false)
counter = 10
steps = 0.0
while counter > 0
    counter := counter - 1
    steps := steps + 1
plot(steps, "Result")`,
			plot:     "Result",
			expected: 10.0,
		},
		{
			name: "Nested",
			pine: `//@version=5
indicator("While Nested", overlay=false)
total = 0.0
i = 1
j = 0
while i <= 3
    j := 1
    while j <= 4
        total := total + 1
        j := j + 1
    i := i + 1
plot(total, "Result")`,
			plot:     "Result",
			expected: 12.0, /* 3 outer × 4 inner */
		},
		{
			name: "ConditionFalseInitially",
			pine: `//@version=5
indicator("While Never Enters", overlay=false)
sum = 99.0
i = 20
while i < 10
    sum := sum + i
    i := i + 1
plot(sum, "Result")`,
			plot:     "Result",
			expected: 99.0,
		},
		{
			name: "SingleIteration",
			pine: `//@version=5
indicator("While Single", overlay=false)
val = 0.0
done = 0
while done == 0
    val := 42
    done := 1
plot(val, "Result")`,
			plot:     "Result",
			expected: 42.0,
		},
		{
			name: "Factorial",
			pine: `//@version=5
indicator("While Factorial", overlay=false)
result = 1.0
counter = 1
while counter <= 5
    result := result * counter
    counter := counter + 1
plot(result, "Result")`,
			plot:     "Result",
			expected: 120.0, /* 5! */
		},
		{
			name: "BreakEarlyExit",
			pine: `//@version=5
indicator("While Break", overlay=false)
sum = 0.0
idx = 1
while idx <= 20
    if idx > 10
        break
    sum := sum + idx
    idx := idx + 1
plot(sum, "Result")`,
			plot:     "Result",
			expected: 55.0, /* 1+2+...+10 */
		},
		{
			name: "ContinueSkip",
			pine: `//@version=5
indicator("While Continue", overlay=false)
sum = 0.0
idx = 0
while idx < 10
    idx := idx + 1
    if idx > 5
        continue
    sum := sum + idx
plot(sum, "Result")`,
			plot:     "Result",
			expected: 15.0, /* 1+2+3+4+5 */
		},
		{
			name: "BreakNestedInner",
			pine: `//@version=5
indicator("While Break Nested", overlay=false)
total = 0.0
i = 1
j = 0
while i <= 3
    j := 1
    while j <= 10
        if j > 2
            break
        total := total + 1
        j := j + 1
    i := i + 1
plot(total, "Result")`,
			plot:     "Result",
			expected: 6.0, /* 3 outer × 2 inner (break at j>2) */
		},
		{
			name: "ContinueNestedInner",
			pine: `//@version=5
indicator("While Continue Nested", overlay=false)
sum = 0.0
i = 1
j = 0
while i <= 3
    j := 0
    while j < 4
        j := j + 1
        if j == 2
            continue
        sum := sum + j
    i := i + 1
plot(sum, "Result")`,
			plot:     "Result",
			expected: 24.0, /* 3 × (1+3+4) */
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			exec := util.NewPineExecutor(t)
			output := exec.ExecuteScript(t, "while-"+strings.ToLower(tc.name), tc.pine)

			vals := exec.ExtractPlotValues(t, output, tc.plot)
			if len(vals) < 1 {
				t.Fatal("Expected at least 1 data point")
			}
			if vals[0] != tc.expected {
				t.Errorf("got %f, want %f", vals[0], tc.expected)
			}
			for i := 1; i < len(vals); i++ {
				if vals[i] != tc.expected {
					t.Errorf("bar[%d] = %f, want %f (constant across bars)", i, vals[i], tc.expected)
					break
				}
			}
		})
	}
}

func TestWhileLoopCodegen(t *testing.T) {
	pineScript := `//@version=5
indicator("While Codegen", overlay=false)

a = 0.0
i = 1
while i <= 10
    if i > 5
        break
    a := a + i
    i := i + 1

b = 0.0
j = 0
while j < 10
    j := j + 1
    if j == 3
        continue
    b := b + j

plot(a, "A")
plot(b, "B")
`

	exec := util.NewPineExecutor(t)
	code, _ := exec.GenerateCode(t, "while-codegen", pineScript)

	required := []string{"break", "continue", "__whileGuard"}
	for _, pat := range required {
		if !strings.Contains(code, pat) {
			t.Fatalf("generated code missing %q", pat)
		}
	}

	if err := exec.CompileCode(t, code); err != nil {
		t.Fatalf("compilation failed: %v", err)
	}
}
