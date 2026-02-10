package integration

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/tests/util"
)

/* break exits loop early — sum 1..10 stops at i>5, yielding 1+2+3+4+5 = 15 */
func TestBreakEarlyExit(t *testing.T) {
	pineScript := `//@version=5
indicator("Break Early Exit", overlay=false)

sum = 0.0
for i = 1 to 10
	if i > 5
		break
	sum := sum + i

plot(sum, "Break Sum")
`

	exec := util.NewPineExecutor(t)
	output := exec.ExecuteScript(t, "break-early-exit", pineScript)

	vals := exec.ExtractPlotValues(t, output, "Break Sum")
	if len(vals) < 1 {
		t.Fatal("Expected at least 1 data point")
	}

	expected := 15.0
	if vals[0] != expected {
		t.Errorf("sum = %f, want %f", vals[0], expected)
	}

	for i := 1; i < len(vals); i++ {
		if vals[i] != expected {
			t.Errorf("sum[%d] = %f, want %f (constant across bars)", i, vals[i], expected)
			break
		}
	}
}

/* continue skips one iteration — sum 1..10 skipping i==5 yields 50 */
func TestContinueSkipIteration(t *testing.T) {
	pineScript := `//@version=5
indicator("Continue Skip", overlay=false)

sum = 0.0
for i = 1 to 10
	if i == 5
		continue
	sum := sum + i

plot(sum, "Continue Sum")
`

	exec := util.NewPineExecutor(t)
	output := exec.ExecuteScript(t, "continue-skip", pineScript)

	vals := exec.ExtractPlotValues(t, output, "Continue Sum")
	if len(vals) < 1 {
		t.Fatal("Expected at least 1 data point")
	}

	expected := 50.0
	if vals[0] != expected {
		t.Errorf("sum = %f, want %f", vals[0], expected)
	}
}

/* break in nested loop only exits inner loop */
func TestBreakNestedLoop(t *testing.T) {
	pineScript := `//@version=5
indicator("Break Nested", overlay=false)

total = 0.0
for i = 1 to 3
	for j = 1 to 10
		if j > 2
			break
		total := total + 1

plot(total, "Nested Break Total")
`

	exec := util.NewPineExecutor(t)
	output := exec.ExecuteScript(t, "break-nested", pineScript)

	vals := exec.ExtractPlotValues(t, output, "Nested Break Total")
	if len(vals) < 1 {
		t.Fatal("Expected at least 1 data point")
	}

	expected := 6.0
	if vals[0] != expected {
		t.Errorf("total = %f, want %f (3 outer × 2 inner)", vals[0], expected)
	}
}

/* continue in nested loop only skips inner iteration */
func TestContinueNestedLoop(t *testing.T) {
	pineScript := `//@version=5
indicator("Continue Nested", overlay=false)

sum = 0.0
for i = 1 to 3
	for j = 1 to 4
		if j == 2
			continue
		sum := sum + j

plot(sum, "Nested Continue Sum")
`

	exec := util.NewPineExecutor(t)
	output := exec.ExecuteScript(t, "continue-nested", pineScript)

	vals := exec.ExtractPlotValues(t, output, "Nested Continue Sum")
	if len(vals) < 1 {
		t.Fatal("Expected at least 1 data point")
	}

	/* Each outer: j=1+3+4=8, three outer iterations: 24 */
	expected := 24.0
	if vals[0] != expected {
		t.Errorf("sum = %f, want %f", vals[0], expected)
	}
}

/* break with descending loop and step */
func TestBreakDescendingStep(t *testing.T) {
	pineScript := `//@version=5
indicator("Break Descending Step", overlay=false)

sum = 0.0
for i = 20 to 0 by -3
	if i < 10
		break
	sum := sum + i

plot(sum, "Descending Break Sum")
`

	exec := util.NewPineExecutor(t)
	output := exec.ExecuteScript(t, "break-descending-step", pineScript)

	vals := exec.ExtractPlotValues(t, output, "Descending Break Sum")
	if len(vals) < 1 {
		t.Fatal("Expected at least 1 data point")
	}

	/* i=20,17,14,11 (break at i=8<10): sum = 20+17+14+11 = 62 */
	expected := 62.0
	if vals[0] != expected {
		t.Errorf("sum = %f, want %f", vals[0], expected)
	}
}

/* continue with multiple skip conditions */
func TestContinueMultipleConditions(t *testing.T) {
	pineScript := `//@version=5
indicator("Continue Multi", overlay=false)

sum = 0.0
for i = 1 to 10
	if i == 3
		continue
	if i == 7
		continue
	sum := sum + i

plot(sum, "Multi Continue Sum")
`

	exec := util.NewPineExecutor(t)
	output := exec.ExecuteScript(t, "continue-multi", pineScript)

	vals := exec.ExtractPlotValues(t, output, "Multi Continue Sum")
	if len(vals) < 1 {
		t.Fatal("Expected at least 1 data point")
	}

	/* 55 - 3 - 7 = 45 */
	expected := 45.0
	if vals[0] != expected {
		t.Errorf("sum = %f, want %f", vals[0], expected)
	}
}

/* break and continue codegen produces valid Go with correct loop counter resolution */
func TestBreakContinueCodegen(t *testing.T) {
	pineScript := `//@version=5
indicator("Break Continue Codegen", overlay=false)

a = 0.0
for i = 1 to 10
	if i > 5
		break
	a := a + i

b = 0.0
for j = 1 to 10
	if j == 3
		continue
	b := b + j

plot(a, "A")
plot(b, "B")
`

	exec := util.NewPineExecutor(t)
	code, _ := exec.GenerateCode(t, "break-continue-codegen", pineScript)

	if !strings.Contains(code, "break") {
		t.Fatal("Generated code missing break statement")
	}
	if !strings.Contains(code, "continue") {
		t.Fatal("Generated code missing continue statement")
	}

	/* Loop counter in condition resolves to float64(i), not iSeries.GetCurrent() */
	if strings.Contains(code, "iSeries.GetCurrent()") {
		t.Fatal("Loop counter 'i' should resolve to float64(i), not iSeries.GetCurrent()")
	}
	if !strings.Contains(code, "float64(i)") {
		t.Fatal("Loop counter 'i' should appear as float64(i) in conditions")
	}

	/* Step increment in for-statement post-expression, not body */
	if !strings.Contains(code, "; i += _step {") {
		t.Fatal("Step increment should be in for-statement post-expression")
	}

	err := exec.CompileCode(t, code)
	if err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}
}
