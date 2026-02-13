//go:build integration

package integration

import (
	"testing"

	"github.com/quant5-lab/runner/tests/util"
)

func TestForLoopBasic(t *testing.T) {
	t.Parallel()
	pineScript := `//@version=5
indicator("For Loop Basic", overlay=false)

sum = 0.0
for i = 1 to 10
	sum := sum + i

plot(sum, "Sum 1-10")
`

	exec := util.NewPineExecutor(t)
	output := exec.ExecuteScript(t, "for-loop-basic", pineScript)

	sumVals := exec.ExtractPlotValues(t, output, "Sum 1-10")

	if len(sumVals) < 1 {
		t.Fatal("Expected at least 1 data point")
	}

	expected := 55.0
	if sumVals[0] != expected {
		t.Errorf("sum = %f, want %f", sumVals[0], expected)
	}

	for i := 1; i < len(sumVals); i++ {
		if sumVals[i] != expected {
			t.Errorf("sum[%d] = %f, want %f (should be constant)", i, sumVals[i], expected)
		}
	}

	t.Logf("✅ For loop basic sum validated: %f across %d bars", expected, len(sumVals))
}

func TestForLoopDescending(t *testing.T) {
	t.Parallel()
	pineScript := `//@version=5
indicator("For Loop Descending", overlay=false)

product = 1.0
for i = 5 to 1 by -1
	product := product * i

plot(product, "Product 5-1")
`

	exec := util.NewPineExecutor(t)
	output := exec.ExecuteScript(t, "for-loop-descending", pineScript)

	productVals := exec.ExtractPlotValues(t, output, "Product 5-1")

	if len(productVals) < 1 {
		t.Fatal("Expected at least 1 data point")
	}

	expected := 120.0
	if productVals[0] != expected {
		t.Errorf("product = %f, want %f (5! = 120)", productVals[0], expected)
	}

	t.Logf("✅ For loop descending validated: %f", expected)
}

func TestForLoopWithStep(t *testing.T) {
	t.Parallel()
	pineScript := `//@version=5
indicator("For Loop Step", overlay=false)

evenSum = 0.0
for i = 0 to 20 by 2
	evenSum := evenSum + i

oddSum = 0.0
for j = 1 to 19 by 2
	oddSum := oddSum + j

plot(evenSum, "Even Sum")
plot(oddSum, "Odd Sum")
`

	exec := util.NewPineExecutor(t)
	output := exec.ExecuteScript(t, "for-loop-step", pineScript)

	evenVals := exec.ExtractPlotValues(t, output, "Even Sum")
	oddVals := exec.ExtractPlotValues(t, output, "Odd Sum")

	if len(evenVals) < 1 || len(oddVals) < 1 {
		t.Fatal("Expected at least 1 data point for each plot")
	}

	expectedEven := 110.0
	if evenVals[0] != expectedEven {
		t.Errorf("evenSum = %f, want %f", evenVals[0], expectedEven)
	}

	expectedOdd := 100.0
	if oddVals[0] != expectedOdd {
		t.Errorf("oddSum = %f, want %f", oddVals[0], expectedOdd)
	}

	t.Logf("✅ For loop step validated: even=%f, odd=%f", expectedEven, expectedOdd)
}

func TestForLoopNested(t *testing.T) {
	t.Parallel()
	pineScript := `//@version=5
indicator("For Loop Nested", overlay=false)

result = 0.0
for i = 1 to 3
	for j = 1 to 4
		result := result + (i * j)

plot(result, "Nested Sum")
`

	exec := util.NewPineExecutor(t)
	output := exec.ExecuteScript(t, "for-loop-nested", pineScript)

	resultVals := exec.ExtractPlotValues(t, output, "Nested Sum")

	if len(resultVals) < 1 {
		t.Fatal("Expected at least 1 data point")
	}

	expected := 60.0
	if resultVals[0] != expected {
		t.Errorf("nested result = %f, want %f", resultVals[0], expected)
	}

	t.Logf("✅ For loop nested validated: %f", expected)
}

func TestForLoopWithBarData(t *testing.T) {
	t.Parallel()
	pineScript := `//@version=5
indicator("For Loop Bar Data", overlay=false)

highCount = 0.0
for i = 1 to 5
	if close > open
		highCount := highCount + 1

plot(highCount, "High Count")
`

	exec := util.NewPineExecutor(t)
	output := exec.ExecuteScript(t, "for-loop-bar-data", pineScript)

	countVals := exec.ExtractPlotValues(t, output, "High Count")

	if len(countVals) < 10 {
		t.Fatal("Expected at least 10 bars")
	}

	for i := 0; i < len(countVals); i++ {
		if countVals[i] < 0 || countVals[i] > 5 {
			t.Errorf("highCount[%d] = %f, want range [0, 5]", i, countVals[i])
		}
	}

	t.Logf("✅ For loop with bar data validated across %d bars", len(countVals))
}

func TestForLoopSingleIteration(t *testing.T) {
	t.Parallel()
	pineScript := `//@version=5
indicator("For Loop Single", overlay=false)

val = 0.0
for i = 5 to 5
	val := i * 2

plot(val, "Single Value")
`

	exec := util.NewPineExecutor(t)
	output := exec.ExecuteScript(t, "for-loop-single", pineScript)

	vals := exec.ExtractPlotValues(t, output, "Single Value")

	if len(vals) < 1 {
		t.Fatal("Expected at least 1 data point")
	}

	expected := 10.0
	if vals[0] != expected {
		t.Errorf("val = %f, want %f", vals[0], expected)
	}

	t.Logf("✅ For loop single iteration validated: %f", expected)
}
