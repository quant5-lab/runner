//go:build integration

package integration

import (
	"testing"

	"github.com/quant5-lab/runner/tests/util"
)

func TestArrayOperationsE2E(t *testing.T) {
	t.Parallel()

	pineScript := `//@version=5
strategy("Array Operations Test", overlay=false)

prices = array.new_float(5, 100.0)
volumes = array.from(100, 200, 300, 400, 500)

array.push(prices, close)
lastRemoved = array.pop(prices)
firstRemoved = array.shift(prices)
array.unshift(prices, open)
array.set(prices, 0, high)
array.insert(prices, 1, low)
array.remove(prices, 2)
array.clear(prices)
array.fill(prices, close, 0, 3)
array.reverse(prices)
array.sort(prices, order.ascending)

firstPrice = array.first(prices)
lastPrice = array.last(prices)
hasValue = array.includes(prices, close)
indexPos = array.indexof(prices, close)
lastIndexPos = array.lastindexof(prices, close)

sumVal = array.sum(prices)
avgVal = array.avg(prices)
minVal = array.min(prices)
maxVal = array.max(prices)
medianVal = array.median(prices)
modeVal = array.mode(prices)
stdevVal = array.stdev(prices)
varianceVal = array.variance(prices)
rangeVal = array.range(prices)
percentile50 = array.percentile_linear_interpolation(prices, 50)
percentile75 = array.percentile_nearest_rank(prices, 75)
rankVal = array.percentrank(prices, close)
covVal = array.covariance(prices, volumes)

sizeVal = array.size(prices)
plot(avgVal, "Average")
plot(stdevVal, "StdDev")
plot(sizeVal, "Size")

longCondition = close > avgVal + stdevVal
shortCondition = close < avgVal - stdevVal

if longCondition
    strategy.entry("Long", strategy.long)
if shortCondition
    strategy.entry("Short", strategy.short)
`

	exec := util.NewPineExecutor(t)
	goCode, _ := exec.GenerateCode(t, "test-array-ops", pineScript)

	requiredCode := []string{
		"pricesArraySeries",
		"volumesArraySeries",
		"arrayops.NewMutator().Push",
		"arrayops.NewMutator().Pop",
		"arrayops.NewMutator().Shift",
		"arrayops.NewMutator().Unshift",
		"arrayops.NewMutator().SetElement",
		"arrayops.NewMutator().Insert",
		"arrayops.NewMutator().Remove",
		"arrayops.NewMutator().Clear",
		"arrayops.NewMutator().Fill",
		"arrayops.NewTransformer().Reverse",
		"arrayops.NewTransformer().Sort",
		"arrayops.NewAccessor().First",
		"arrayops.NewAccessor().Last",
		"arrayops.NewAccessor().Includes",
		"arrayops.NewAccessor().IndexOf",
		"arrayops.NewAccessor().LastIndexOf",
		"arrayops.NewStatistics().Sum",
		"arrayops.NewStatistics().Avg",
		"arrayops.NewStatistics().Min",
		"arrayops.NewStatistics().Max",
		"arrayops.NewStatistics().Median",
		"arrayops.NewStatistics().Mode",
		"arrayops.NewStatistics().Stdev",
		"arrayops.NewStatistics().Variance",
		"arrayops.NewStatistics().Range",
		"arrayops.NewStatistics().Percentile",
		"arrayops.NewStatistics().PercentRank",
		"arrayops.NewStatistics().Covariance",
		`"order.ascending"`,
	}

	for _, code := range requiredCode {
		if !contains(goCode, code) {
			t.Errorf("Generated code missing: %s", code)
		}
	}

	output := exec.ExecuteScript(t, "test-array-ops", pineScript)
	avgVals := exec.ExtractPlotValues(t, output, "Average")
	stdevVals := exec.ExtractPlotValues(t, output, "StdDev")
	sizeVals := exec.ExtractPlotValues(t, output, "Size")

	if len(avgVals) < 10 {
		t.Fatalf("Expected at least 10 bars, got %d", len(avgVals))
	}

	for i, val := range sizeVals {
		if val != 0 {
			t.Errorf("Bar %d: size = %f, want 0 (clear makes empty, fill doesn't expand)", i, val)
			break
		}
	}

	for i, val := range avgVals {
		if val < 0 {
			t.Errorf("Bar %d: avg = %f, should be non-negative", i, val)
			break
		}
	}

	for i, val := range stdevVals {
		if val < 0 {
			t.Errorf("Bar %d: stdev = %f, should be non-negative", i, val)
			break
		}
	}
}

func TestArrayConstructors(t *testing.T) {
	t.Parallel()

	pineScript := `//@version=5
strategy("Test", overlay=false)
arr1 = array.new_float()
arr2 = array.new_float(10)
arr3 = array.new_float(5, 99.5)
arr4 = array.from(1, 2, 3)
s1 = array.size(arr1)
s4 = array.size(arr4)
plot(s1, "Size1")
plot(s4, "Size4")
`

	exec := util.NewPineExecutor(t)
	goCode, _ := exec.GenerateCode(t, "test-constructors", pineScript)

	expected := []string{
		"arr1ArraySeries.Set([]float64{})",
		"arr2ArraySeries.Set(make([]float64, int(10)))",
		"arr3ArraySeries.Set(arrayops.NewArrayWithValue(int(5), 99.5))",
		"arr4ArraySeries.Set([]float64{1, 2, 3})",
	}

	for _, exp := range expected {
		if !contains(goCode, exp) {
			t.Errorf("Expected: %s", exp)
		}
	}

	output := exec.ExecuteScript(t, "test-constructors", pineScript)
	size1 := exec.ExtractPlotValues(t, output, "Size1")
	size4 := exec.ExtractPlotValues(t, output, "Size4")

	if len(size1) > 0 && size1[0] != 0 {
		t.Errorf("arr1 size = %f, want 0 (empty array)", size1[0])
	}

	if len(size4) > 0 && size4[0] != 3 {
		t.Errorf("arr4 size = %f, want 3 (from 1,2,3)", size4[0])
	}
}

func TestArrayMutators(t *testing.T) {
	t.Parallel()

	pineScript := `//@version=5
strategy("Test", overlay=false)
prices = array.new_float()
array.push(prices, close)
v1 = array.pop(prices)
v2 = array.shift(prices)
array.unshift(prices, open)
array.set(prices, 0, high)
array.insert(prices, 1, low)
array.remove(prices, 2)
array.clear(prices)
array.fill(prices, 100, 0, 5)
array.reverse(prices)
array.sort(prices, order.descending)
sz = array.size(prices)
plot(sz, "Size")
`

	exec := util.NewPineExecutor(t)
	goCode, _ := exec.GenerateCode(t, "test-mutators", pineScript)

	expected := []string{
		"arrayops.NewMutator().Push(pricesArraySeries, ctx.Data[ctx.BarIndex].Close)",
		"arrayops.NewMutator().Pop(pricesArraySeries)",
		"arrayops.NewMutator().Shift(pricesArraySeries)",
		"arrayops.NewMutator().Unshift(pricesArraySeries, ctx.Data[ctx.BarIndex].Open)",
		"arrayops.NewMutator().SetElement(pricesArraySeries, int(0), ctx.Data[ctx.BarIndex].High)",
		"arrayops.NewMutator().Insert(pricesArraySeries, int(1), ctx.Data[ctx.BarIndex].Low)",
		"arrayops.NewMutator().Remove(pricesArraySeries, int(2))",
		"arrayops.NewMutator().Clear(pricesArraySeries)",
		"arrayops.NewMutator().Fill(pricesArraySeries, 100, int(0), int(5))",
		"arrayops.NewTransformer().Reverse(pricesArraySeries)",
		`arrayops.NewTransformer().Sort(pricesArraySeries, "order.descending")`,
	}

	for _, exp := range expected {
		if !contains(goCode, exp) {
			t.Errorf("Expected: %s", exp)
		}
	}

	output := exec.ExecuteScript(t, "test-mutators", pineScript)
	sizeVals := exec.ExtractPlotValues(t, output, "Size")

	if len(sizeVals) > 0 && sizeVals[0] != 0 {
		t.Errorf("size = %f, want 0 (clear makes empty, fill doesn't expand)", sizeVals[0])
	}
}

func TestArrayAccessors(t *testing.T) {
	t.Parallel()

	pineScript := `//@version=5
strategy("Test", overlay=false)
prices = array.from(10, 20, 30)
f = array.first(prices)
l = array.last(prices)
has = array.includes(prices, 20)
idx = array.indexof(prices, 30)
lidx = array.lastindexof(prices, 10)
plot(f, "First")
plot(l, "Last")
`

	exec := util.NewPineExecutor(t)
	goCode, _ := exec.GenerateCode(t, "test-accessors", pineScript)

	expected := []string{
		"arrayops.NewAccessor().First(pricesArraySeries, 0)",
		"arrayops.NewAccessor().Last(pricesArraySeries, 0)",
		"arrayops.NewAccessor().Includes(pricesArraySeries, 0, 20)",
		"arrayops.NewAccessor().IndexOf(pricesArraySeries, 0, 30, 0)",
		"arrayops.NewAccessor().LastIndexOf(pricesArraySeries, 0, 10, -1)",
	}

	for _, exp := range expected {
		if !contains(goCode, exp) {
			t.Errorf("Expected: %s", exp)
		}
	}

	output := exec.ExecuteScript(t, "test-accessors", pineScript)
	firstVals := exec.ExtractPlotValues(t, output, "First")
	lastVals := exec.ExtractPlotValues(t, output, "Last")

	if len(firstVals) > 0 && firstVals[0] != 10 {
		t.Errorf("first = %f, want 10", firstVals[0])
	}

	if len(lastVals) > 0 && lastVals[0] != 30 {
		t.Errorf("last = %f, want 30", lastVals[0])
	}
}

func TestArrayStatistics(t *testing.T) {
	t.Parallel()

	pineScript := `//@version=5
strategy("Test", overlay=false)
prices = array.from(1, 2, 3, 4, 5)
s = array.sum(prices)
a = array.avg(prices)
mn = array.min(prices)
mx = array.max(prices)
med = array.median(prices)
mod = array.mode(prices)
std = array.stdev(prices)
vr = array.variance(prices)
rng = array.range(prices)
p50 = array.percentile_linear_interpolation(prices, 50)
p75 = array.percentile_nearest_rank(prices, 75)
rank = array.percentrank(prices, 3)
plot(a, "Avg")
plot(s, "Sum")
`

	exec := util.NewPineExecutor(t)
	goCode, _ := exec.GenerateCode(t, "test-stats", pineScript)

	expected := []string{
		"arrayops.NewStatistics().Sum(pricesArraySeries, 0)",
		"arrayops.NewStatistics().Avg(pricesArraySeries, 0)",
		"arrayops.NewStatistics().Min(pricesArraySeries, 0)",
		"arrayops.NewStatistics().Max(pricesArraySeries, 0)",
		"arrayops.NewStatistics().Median(pricesArraySeries, 0)",
		"arrayops.NewStatistics().Mode(pricesArraySeries, 0)",
		"arrayops.NewStatistics().Stdev(pricesArraySeries, 0, false)",
		"arrayops.NewStatistics().Variance(pricesArraySeries, 0, false)",
		"arrayops.NewStatistics().Range(pricesArraySeries, 0)",
		`arrayops.NewStatistics().Percentile(pricesArraySeries, 0, 50, "linear")`,
		`arrayops.NewStatistics().Percentile(pricesArraySeries, 0, 75, "nearest_rank")`,
		"arrayops.NewStatistics().PercentRank(pricesArraySeries, 0, 3)",
	}

	for _, exp := range expected {
		if !contains(goCode, exp) {
			t.Errorf("Expected: %s", exp)
		}
	}

	output := exec.ExecuteScript(t, "test-stats", pineScript)
	avgVals := exec.ExtractPlotValues(t, output, "Avg")
	sumVals := exec.ExtractPlotValues(t, output, "Sum")

	if len(avgVals) > 0 && avgVals[0] != 3 {
		t.Errorf("avg([1,2,3,4,5]) = %f, want 3", avgVals[0])
	}

	if len(sumVals) > 0 && sumVals[0] != 15 {
		t.Errorf("sum([1,2,3,4,5]) = %f, want 15", sumVals[0])
	}
}

func TestArrayCovariance(t *testing.T) {
	t.Parallel()

	pineScript := `//@version=5
strategy("Test", overlay=false)
x = array.from(1, 2, 3)
y = array.from(4, 5, 6)
cov = array.covariance(x, y)
plot(cov, "Cov")
`

	exec := util.NewPineExecutor(t)
	goCode, _ := exec.GenerateCode(t, "test-cov", pineScript)

	if !contains(goCode, "arrayops.NewStatistics().Covariance(xArraySeries, 0, yArraySeries, 0, false)") {
		t.Error("Expected covariance call")
	}

	output := exec.ExecuteScript(t, "test-cov", pineScript)
	covVals := exec.ExtractPlotValues(t, output, "Cov")

	if len(covVals) > 0 && covVals[0] != 1 {
		t.Errorf("cov([1,2,3], [4,5,6]) = %f, want 1 (sample covariance)", covVals[0])
	}
}

func contains(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
