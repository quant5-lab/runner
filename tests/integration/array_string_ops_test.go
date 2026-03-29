//go:build integration

package integration

import (
	"testing"

	"github.com/quant5-lab/runner/tests/util"
)

func TestArrayStringOperations(t *testing.T) {
	t.Parallel()

	script := `//@version=5
indicator("Array String Operations", overlay=false)

labels = array.new_string(3, "default")
array.push(labels, "long")
array.push(labels, "short")
array.set(labels, 0, "first")

size = array.size(labels)
plot(size, "Size")`

	exec := util.NewPineExecutor(t)
	output := exec.ExecuteScript(t, "array-string-ops", script)
	values := exec.ExtractPlotValues(t, output, "Size")

	if len(values) == 0 {
		t.Fatal("Size indicator has no values")
	}

	lastValue := values[len(values)-1]
	expectedSize := 5.0
	if lastValue != expectedSize {
		t.Errorf("Expected array size %.0f (3 initial + 2 pushes), got %v", expectedSize, lastValue)
	}
}

func TestArrayFloatAllowsNumeric(t *testing.T) {
	t.Parallel()

	script := `//@version=5
indicator("Float Array Numeric Ops", overlay=false)

values = array.new_float(5, 10.0)
array.push(values, 20.0)
array.push(values, 30.0)
total = array.sum(values)
avg = array.avg(values)

plot(total, "Sum")
plot(avg, "Average")`

	exec := util.NewPineExecutor(t)
	output := exec.ExecuteScript(t, "float-numeric", script)

	sumValues := exec.ExtractPlotValues(t, output, "Sum")
	if len(sumValues) == 0 {
		t.Fatal("Sum indicator has no values")
	}

	lastSum := sumValues[len(sumValues)-1]
	expectedSum := 100.0
	if lastSum != expectedSum {
		t.Errorf("Expected sum %.0f (5×10 + 20 + 30), got %v", expectedSum, lastSum)
	}

	avgValues := exec.ExtractPlotValues(t, output, "Average")
	lastAvg := avgValues[len(avgValues)-1]
	expectedAvg := 14.285714285714286
	if lastAvg != expectedAvg {
		t.Errorf("Expected avg %.15f, got %.15f", expectedAvg, lastAvg)
	}
}

func TestArrayMixedTypes(t *testing.T) {
	t.Parallel()

	script := `//@version=5
indicator("Mixed Array Types", overlay=false)

floats = array.new_float(3, 1.5)
strings = array.new_string(2, "label")

array.push(floats, 2.5)
array.push(strings, "tag")

floatSum = array.sum(floats)
stringSize = array.size(strings)

plot(floatSum, "Float Sum")
plot(stringSize, "String Size")`

	exec := util.NewPineExecutor(t)
	output := exec.ExecuteScript(t, "mixed-types", script)

	sumValues := exec.ExtractPlotValues(t, output, "Float Sum")
	lastSum := sumValues[len(sumValues)-1]
	expectedSum := 7.0
	if lastSum != expectedSum {
		t.Errorf("Expected sum %.0f (3×1.5 + 2.5), got %v", expectedSum, lastSum)
	}

	sizeValues := exec.ExtractPlotValues(t, output, "String Size")
	lastSize := sizeValues[len(sizeValues)-1]
	expectedSize := 3.0
	if lastSize != expectedSize {
		t.Errorf("Expected size %.0f (2 initial + 1 push), got %v", expectedSize, lastSize)
	}
}
