package regression

import (
	"math"
	"os"
	"path/filepath"
	"testing"
)

/*
TestArrowLoopCounterArithmetic_FactorialValues verifies that for-loop counters inside
arrow functions resolve as float64 scalars — not Series — when used in arithmetic.

Factorial is computed via a loop where the counter `i` is multiplied with the
accumulator each iteration. If `i` resolves to `iSeries.GetCurrent()`, the generated
Go code fails to compile; if it resolves to a bare `int`, the multiplication with a
float64 accumulator fails to compile. Both are regressions of Fix B.

With correct resolution — float64(i) — factorial(5) = 120 and factorial(3) = 6 on
every bar.
*/
func TestArrowLoopCounterArithmetic_FactorialValues(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	strategy := `//@version=5
indicator("Arrow Loop Counter Arithmetic")
factorial(n) =>
    result = 1.0
    for i = 1 to n
        result := result * i
    result
plot(factorial(5), "Fact5")
plot(factorial(3), "Fact3")
`
	testDir := t.TempDir()
	strategyPath := filepath.Join(testDir, "arrow-loop-counter.pine")
	if err := os.WriteFile(strategyPath, []byte(strategy), 0644); err != nil {
		t.Fatal(err)
	}
	dataPath := filepath.Join(testDir, "ARWLP_1D.json")
	if err := os.WriteFile(dataPath, []byte(generateTestOHLCV(10, 86400)), 0644); err != nil {
		t.Fatal(err)
	}

	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(filepath.Dir(cwd))

	result := compileAndRun(t, strategyPath, dataPath, testDir, projectRoot, "ARWLP", testDir)

	for _, tc := range []struct {
		name     string
		expected float64
	}{
		{"Fact5", 120.0},
		{"Fact3", 6.0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ind, ok := result.Indicators[tc.name]
			if !ok {
				t.Fatalf("indicator %q absent from output", tc.name)
			}
			vals := extractValues(ind.Data)
			for i, v := range vals {
				if math.IsNaN(v) {
					t.Errorf("bar %d: %s = NaN, want %.0f", i, tc.name, tc.expected)
				} else if v != tc.expected {
					t.Errorf("bar %d: %s = %.0f, want %.0f", i, tc.name, v, tc.expected)
				}
			}
		})
	}
}

/*
TestArrowLoopCounterArithmetic_NestedLoops verifies that independent loop counters
in nested arrow-function for-loops each receive correct float64 resolution.

The outer counter `r` and inner counter `c` are summed into an accumulator.
If either resolves as a Series, codegen fails. The expected total for a 3x3 traversal
is (1+2+3)*3 + (1+2+3)*3 = 36 via sum of r+c across all cells.
*/
func TestArrowLoopCounterArithmetic_NestedLoops(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	strategy := `//@version=5
indicator("Arrow Nested Loop Counters")
nestedSum(rows, cols) =>
    total = 0.0
    for r = 1 to rows
        for c = 1 to cols
            total := total + r + c
    total
plot(nestedSum(3, 3), "NestedSum")
`
	testDir := t.TempDir()
	strategyPath := filepath.Join(testDir, "arrow-nested-loops.pine")
	if err := os.WriteFile(strategyPath, []byte(strategy), 0644); err != nil {
		t.Fatal(err)
	}
	dataPath := filepath.Join(testDir, "ARWNST_1D.json")
	if err := os.WriteFile(dataPath, []byte(generateTestOHLCV(10, 86400)), 0644); err != nil {
		t.Fatal(err)
	}

	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(filepath.Dir(cwd))

	result := compileAndRun(t, strategyPath, dataPath, testDir, projectRoot, "ARWNST", testDir)

	ind, ok := result.Indicators["NestedSum"]
	if !ok {
		t.Fatal("indicator 'NestedSum' absent from output")
	}

	// nestedSum(3,3): sum of (r+c) for r in [1..3], c in [1..3]
	// = (1+1)+(1+2)+(1+3) + (2+1)+(2+2)+(2+3) + (3+1)+(3+2)+(3+3) = 36
	const expected = 36.0
	vals := extractValues(ind.Data)
	for i, v := range vals {
		if math.IsNaN(v) {
			t.Errorf("bar %d: NestedSum = NaN, want %.0f", i, expected)
		} else if v != expected {
			t.Errorf("bar %d: NestedSum = %.0f, want %.0f", i, v, expected)
		}
	}
}

/*
TestArrowSeriesParameter_WrappedSMA verifies that a parameter passed to a TA function
inside an arrow function receives *series.Series typing and produces numerically
correct output.

If the parameter were typed float64 (pre-fix), the TA function would receive a scalar
instead of the series and produce wrong or NaN output. The wrapped mySma must match
the direct ta.sma for all bars where ta.sma is defined.
*/
func TestArrowSeriesParameter_WrappedSMA(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	strategy := `//@version=5
indicator("Arrow Series Parameter SMA")
mySma(src, len) =>
    ta.sma(src, len)
plot(mySma(close, 3), "MySma3")
plot(ta.sma(close, 3), "DirectSma3")
`
	testDir := t.TempDir()
	strategyPath := filepath.Join(testDir, "arrow-series-param.pine")
	if err := os.WriteFile(strategyPath, []byte(strategy), 0644); err != nil {
		t.Fatal(err)
	}
	dataPath := filepath.Join(testDir, "ARWSMA_1D.json")
	if err := os.WriteFile(dataPath, []byte(generateTestOHLCV(20, 86400)), 0644); err != nil {
		t.Fatal(err)
	}

	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(filepath.Dir(cwd))

	result := compileAndRun(t, strategyPath, dataPath, testDir, projectRoot, "ARWSMA", testDir)

	myInd, ok := result.Indicators["MySma3"]
	if !ok {
		t.Fatal("indicator 'MySma3' absent from output")
	}
	dirInd, ok := result.Indicators["DirectSma3"]
	if !ok {
		t.Fatal("indicator 'DirectSma3' absent from output")
	}

	myVals := extractValues(myInd.Data)
	dirVals := extractValues(dirInd.Data)

	if len(myVals) != len(dirVals) {
		t.Fatalf("length mismatch: MySma3=%d DirectSma3=%d", len(myVals), len(dirVals))
	}

	for i := range myVals {
		myNaN := math.IsNaN(myVals[i])
		dirNaN := math.IsNaN(dirVals[i])
		if myNaN != dirNaN {
			t.Errorf("bar %d: NaN mismatch — MySma3=%v DirectSma3=%v", i, myVals[i], dirVals[i])
			continue
		}
		if !myNaN && math.Abs(myVals[i]-dirVals[i]) > 1e-9 {
			t.Errorf("bar %d: MySma3=%.6f != DirectSma3=%.6f", i, myVals[i], dirVals[i])
		}
	}
}

/*
TestArrowSeriesParameter_MultipleSeriesParams verifies that two independent series
parameters in the same arrow function are each resolved correctly through the full
codegen→compile→run pipeline.

The function computes (sma(a, len) - sma(b, len)) which is zero when a and b are
identical series, confirming both parameters carry their respective series data.
*/
func TestArrowSeriesParameter_MultipleSeriesParams(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	strategy := `//@version=5
indicator("Arrow Multiple Series Params")
smaDiff(a, b, len) =>
    ta.sma(a, len) - ta.sma(b, len)
plot(smaDiff(close, close, 3), "ZeroDiff")
plot(smaDiff(high, low, 3),    "HLDiff")
`
	testDir := t.TempDir()
	strategyPath := filepath.Join(testDir, "arrow-multi-series.pine")
	if err := os.WriteFile(strategyPath, []byte(strategy), 0644); err != nil {
		t.Fatal(err)
	}
	dataPath := filepath.Join(testDir, "ARWMS_1D.json")
	if err := os.WriteFile(dataPath, []byte(generateTestOHLCV(20, 86400)), 0644); err != nil {
		t.Fatal(err)
	}

	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(filepath.Dir(cwd))

	result := compileAndRun(t, strategyPath, dataPath, testDir, projectRoot, "ARWMS", testDir)

	zeroInd, ok := result.Indicators["ZeroDiff"]
	if !ok {
		t.Fatal("indicator 'ZeroDiff' absent from output")
	}
	hlInd, ok := result.Indicators["HLDiff"]
	if !ok {
		t.Fatal("indicator 'HLDiff' absent from output")
	}

	zeroVals := extractValues(zeroInd.Data)
	hlVals := extractValues(hlInd.Data)

	nonNaN := 0
	for i, v := range zeroVals {
		if math.IsNaN(v) {
			continue
		}
		nonNaN++
		if math.Abs(v) > 1e-9 {
			t.Errorf("bar %d: ZeroDiff(close, close) = %.9f, want 0", i, v)
		}
	}
	if nonNaN == 0 {
		t.Error("ZeroDiff produced no non-NaN values")
	}

	nonNaN = 0
	for i, v := range hlVals {
		if math.IsNaN(v) {
			continue
		}
		nonNaN++
		// high - low per bar = 200 in test data; SMA of constant = 200
		if math.Abs(v-200.0) > 1e-9 {
			t.Errorf("bar %d: HLDiff = %.6f, want 200.0", i, v)
		}
	}
	if nonNaN == 0 {
		t.Error("HLDiff produced no non-NaN values")
	}
}

/*
TestV3BarIndexAlias_TopLevel verifies that the Pine v3 identifier `n` resolves to
bar_index at top-level scope, producing a monotonically increasing 0-based bar counter.

Before the alias fix, `n` generated `nSeries.GetCurrent()`, which is an undefined
series, causing a compile error.
*/
func TestV3BarIndexAlias_TopLevel(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	strategy := `//@version=3
study("V3 Bar Index Alias Top Level")
barNum = n
plot(barNum, "BarNum")
`
	testDir := t.TempDir()
	strategyPath := filepath.Join(testDir, "v3-alias-toplevel.pine")
	if err := os.WriteFile(strategyPath, []byte(strategy), 0644); err != nil {
		t.Fatal(err)
	}
	dataPath := filepath.Join(testDir, "V3ALIAS_1D.json")
	if err := os.WriteFile(dataPath, []byte(generateTestOHLCV(10, 86400)), 0644); err != nil {
		t.Fatal(err)
	}

	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(filepath.Dir(cwd))

	result := compileAndRun(t, strategyPath, dataPath, testDir, projectRoot, "V3ALIAS", testDir)

	ind, ok := result.Indicators["BarNum"]
	if !ok {
		t.Fatal("indicator 'BarNum' absent from output")
	}

	vals := extractValues(ind.Data)
	for i, v := range vals {
		if math.IsNaN(v) {
			t.Errorf("bar %d: BarNum = NaN, want %.0f", i, float64(i))
		} else if v != float64(i) {
			t.Errorf("bar %d: BarNum = %.0f, want %d", i, v, i)
		}
	}
}

/*
TestV3BarIndexAlias_InArrowFunction verifies that `n` resolves to bar_index inside
an arrow function body, producing bar_index * multiplier on each bar.
*/
func TestV3BarIndexAlias_InArrowFunction(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	strategy := `//@version=3
study("V3 Bar Index Alias In Arrow")
doubled(mult) =>
    n * mult
plot(doubled(2.0), "Doubled")
`
	testDir := t.TempDir()
	strategyPath := filepath.Join(testDir, "v3-alias-arrow.pine")
	if err := os.WriteFile(strategyPath, []byte(strategy), 0644); err != nil {
		t.Fatal(err)
	}
	dataPath := filepath.Join(testDir, "V3ARROW_1D.json")
	if err := os.WriteFile(dataPath, []byte(generateTestOHLCV(10, 86400)), 0644); err != nil {
		t.Fatal(err)
	}

	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(filepath.Dir(cwd))

	result := compileAndRun(t, strategyPath, dataPath, testDir, projectRoot, "V3ARROW", testDir)

	ind, ok := result.Indicators["Doubled"]
	if !ok {
		t.Fatal("indicator 'Doubled' absent from output")
	}

	vals := extractValues(ind.Data)
	for i, v := range vals {
		expected := float64(i) * 2.0
		if math.IsNaN(v) {
			t.Errorf("bar %d: Doubled = NaN, want %.0f", i, expected)
		} else if v != expected {
			t.Errorf("bar %d: Doubled = %.0f, want %.0f", i, v, expected)
		}
	}
}

/*
TestV3BarIndexAlias_WithLoopCounter verifies that `n` (bar_index alias) and a
for-loop counter coexist correctly inside the same arrow function. The function
accumulates loop indices and offsets by bar_index.

This exercises the intersection of Fix B (loop counter float64 cast) and Fix C
(v3 alias resolution) in a single codegen path.
*/
func TestV3BarIndexAlias_WithLoopCounter(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	strategy := `//@version=3
study("V3 Alias With Loop Counter")
loopPlusBar(len) =>
    total = 0.0
    for i = 1 to len
        total := total + i
    total + n
plot(loopPlusBar(3), "LoopPlusBar")
`
	testDir := t.TempDir()
	strategyPath := filepath.Join(testDir, "v3-alias-loop.pine")
	if err := os.WriteFile(strategyPath, []byte(strategy), 0644); err != nil {
		t.Fatal(err)
	}
	dataPath := filepath.Join(testDir, "V3LOOP_1D.json")
	if err := os.WriteFile(dataPath, []byte(generateTestOHLCV(10, 86400)), 0644); err != nil {
		t.Fatal(err)
	}

	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(filepath.Dir(cwd))

	result := compileAndRun(t, strategyPath, dataPath, testDir, projectRoot, "V3LOOP", testDir)

	ind, ok := result.Indicators["LoopPlusBar"]
	if !ok {
		t.Fatal("indicator 'LoopPlusBar' absent from output")
	}

	// loopPlusBar(3) = (1+2+3) + bar_index = 6 + i
	vals := extractValues(ind.Data)
	for i, v := range vals {
		expected := 6.0 + float64(i)
		if math.IsNaN(v) {
			t.Errorf("bar %d: LoopPlusBar = NaN, want %.0f", i, expected)
		} else if v != expected {
			t.Errorf("bar %d: LoopPlusBar = %.0f, want %.0f", i, v, expected)
		}
	}
}
