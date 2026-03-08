package regression

import (
	"math"
	"os"
	"path/filepath"
	"testing"
)

/*
TestArrowStringParam_BranchSelection verifies that a string-typed formal parameter
is detected via equality comparison against a string literal and emitted as a Go
`string` parameter — not float64 — allowing the conditional branch to select the
correct value at runtime.

Without Fix D, the parameter would be typed float64, causing a compile error when
the string literal operand is generated as a quoted string.
*/
func TestArrowStringParam_BranchSelection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	strategy := `//@version=5
indicator("Arrow String Param Branch")
pick(a, b, which) =>
    which == "a" ? a : b
plot(pick(1.0, 2.0, "a"), "PickA")
plot(pick(1.0, 2.0, "b"), "PickB")
`
	testDir := t.TempDir()
	strategyPath := filepath.Join(testDir, "arrow-string-param.pine")
	if err := os.WriteFile(strategyPath, []byte(strategy), 0644); err != nil {
		t.Fatal(err)
	}
	dataPath := filepath.Join(testDir, "ARWSTR_1D.json")
	if err := os.WriteFile(dataPath, []byte(generateTestOHLCV(10, 86400)), 0644); err != nil {
		t.Fatal(err)
	}

	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(filepath.Dir(cwd))

	result := compileAndRun(t, strategyPath, dataPath, testDir, projectRoot, "ARWSTR", testDir)

	for _, tc := range []struct {
		name     string
		expected float64
	}{
		{"PickA", 1.0},
		{"PickB", 2.0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ind, ok := result.Indicators[tc.name]
			if !ok {
				t.Fatalf("indicator %q absent from output", tc.name)
			}
			vals := extractValues(ind.Data)
			for i, v := range vals {
				if math.IsNaN(v) {
					t.Errorf("bar %d: %s = NaN, want %.1f", i, tc.name, tc.expected)
				} else if v != tc.expected {
					t.Errorf("bar %d: %s = %.1f, want %.1f", i, tc.name, v, tc.expected)
				}
			}
		})
	}
}

/*
TestArrowStringParam_InequalityBranch verifies that `!=` against a string literal
also triggers string parameter detection, matching the symmetric case of Fix D.
*/
func TestArrowStringParam_InequalityBranch(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	strategy := `//@version=5
indicator("Arrow String Param Inequality")
notA(val, which) =>
    which != "a" ? val : 0.0
plot(notA(5.0, "a"), "WhenA")
plot(notA(5.0, "b"), "WhenNotA")
`
	testDir := t.TempDir()
	strategyPath := filepath.Join(testDir, "arrow-string-neq.pine")
	if err := os.WriteFile(strategyPath, []byte(strategy), 0644); err != nil {
		t.Fatal(err)
	}
	dataPath := filepath.Join(testDir, "ARWNEQ_1D.json")
	if err := os.WriteFile(dataPath, []byte(generateTestOHLCV(10, 86400)), 0644); err != nil {
		t.Fatal(err)
	}

	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(filepath.Dir(cwd))

	result := compileAndRun(t, strategyPath, dataPath, testDir, projectRoot, "ARWNEQ", testDir)

	for _, tc := range []struct {
		name     string
		expected float64
	}{
		{"WhenA", 0.0},
		{"WhenNotA", 5.0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ind, ok := result.Indicators[tc.name]
			if !ok {
				t.Fatalf("indicator %q absent from output", tc.name)
			}
			vals := extractValues(ind.Data)
			for i, v := range vals {
				if math.IsNaN(v) {
					t.Errorf("bar %d: %s = NaN, want %.1f", i, tc.name, tc.expected)
				} else if v != tc.expected {
					t.Errorf("bar %d: %s = %.1f, want %.1f", i, tc.name, v, tc.expected)
				}
			}
		})
	}
}

/*
TestArrowOuterSeriesCapture_CurrentBar verifies that an outer-scope float series
variable referenced inside an arrow function body delivers the same per-bar value
as a direct reference to the same series.

Without Fix E, the outer variable is unresolved and codegen fails to compile.
*/
func TestArrowOuterSeriesCapture_CurrentBar(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	strategy := `//@version=5
indicator("Outer Series Capture Current Bar")
myClose = close
getValue() =>
    myClose
plot(getValue(), "Captured")
plot(close, "Direct")
`
	testDir := t.TempDir()
	strategyPath := filepath.Join(testDir, "arrow-outer-capture.pine")
	if err := os.WriteFile(strategyPath, []byte(strategy), 0644); err != nil {
		t.Fatal(err)
	}
	dataPath := filepath.Join(testDir, "ARWCAP_1D.json")
	if err := os.WriteFile(dataPath, []byte(generateTestOHLCV(20, 86400)), 0644); err != nil {
		t.Fatal(err)
	}

	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(filepath.Dir(cwd))

	result := compileAndRun(t, strategyPath, dataPath, testDir, projectRoot, "ARWCAP", testDir)

	capInd, ok := result.Indicators["Captured"]
	if !ok {
		t.Fatal("indicator 'Captured' absent from output")
	}
	dirInd, ok := result.Indicators["Direct"]
	if !ok {
		t.Fatal("indicator 'Direct' absent from output")
	}

	capVals := extractValues(capInd.Data)
	dirVals := extractValues(dirInd.Data)

	if len(capVals) != len(dirVals) {
		t.Fatalf("length mismatch: Captured=%d Direct=%d", len(capVals), len(dirVals))
	}
	for i := range capVals {
		if math.IsNaN(capVals[i]) != math.IsNaN(dirVals[i]) {
			t.Errorf("bar %d: NaN mismatch — Captured=%v Direct=%v", i, capVals[i], dirVals[i])
			continue
		}
		if !math.IsNaN(capVals[i]) && math.Abs(capVals[i]-dirVals[i]) > 1e-9 {
			t.Errorf("bar %d: Captured=%.4f != Direct=%.4f", i, capVals[i], dirVals[i])
		}
	}
}

/*
TestArrowOuterSeriesCapture_HistoricalSubscript verifies that a historical subscript
on a captured outer-scope series resolves to the same value as the equivalent direct
subscript expression.

This exercises the Get(offset) path rather than GetCurrent() — a distinct code path
in the access resolver that must also honour the captured *series.Series injection.
*/
func TestArrowOuterSeriesCapture_HistoricalSubscript(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	strategy := `//@version=5
indicator("Outer Series Historical Subscript")
myClose = close
prevBar() =>
    myClose[1]
plot(prevBar(), "CapturedPrev")
plot(close[1], "DirectPrev")
`
	testDir := t.TempDir()
	strategyPath := filepath.Join(testDir, "arrow-outer-hist.pine")
	if err := os.WriteFile(strategyPath, []byte(strategy), 0644); err != nil {
		t.Fatal(err)
	}
	dataPath := filepath.Join(testDir, "ARWHST_1D.json")
	if err := os.WriteFile(dataPath, []byte(generateTestOHLCV(20, 86400)), 0644); err != nil {
		t.Fatal(err)
	}

	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(filepath.Dir(cwd))

	result := compileAndRun(t, strategyPath, dataPath, testDir, projectRoot, "ARWHST", testDir)

	capInd, ok := result.Indicators["CapturedPrev"]
	if !ok {
		t.Fatal("indicator 'CapturedPrev' absent from output")
	}
	dirInd, ok := result.Indicators["DirectPrev"]
	if !ok {
		t.Fatal("indicator 'DirectPrev' absent from output")
	}

	capVals := extractValues(capInd.Data)
	dirVals := extractValues(dirInd.Data)

	if len(capVals) != len(dirVals) {
		t.Fatalf("length mismatch: CapturedPrev=%d DirectPrev=%d", len(capVals), len(dirVals))
	}

	nonNaN := 0
	for i := range capVals {
		if math.IsNaN(capVals[i]) != math.IsNaN(dirVals[i]) {
			t.Errorf("bar %d: NaN mismatch — CapturedPrev=%v DirectPrev=%v", i, capVals[i], dirVals[i])
			continue
		}
		if !math.IsNaN(capVals[i]) {
			nonNaN++
			if math.Abs(capVals[i]-dirVals[i]) > 1e-9 {
				t.Errorf("bar %d: CapturedPrev=%.4f != DirectPrev=%.4f", i, capVals[i], dirVals[i])
			}
		}
	}
	if nonNaN == 0 {
		t.Error("CapturedPrev produced no non-NaN values")
	}
}

/*
TestArrowOuterSeriesCapture_MultipleDistinctCaptures verifies that two independent
outer-scope series are each captured separately, injected as distinct *series.Series
parameters, and deliver the correct per-bar values inside the function body.

High - Low = 200.0 on every bar in the synthetic data (constant spread), making the
arithmetic outcome independent of bar index and trivially verifiable.
*/
func TestArrowOuterSeriesCapture_MultipleDistinctCaptures(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	strategy := `//@version=5
indicator("Outer Series Multiple Captures")
myHigh = high
myLow = low
spread() =>
    myHigh - myLow
plot(spread(), "Spread")
`
	testDir := t.TempDir()
	strategyPath := filepath.Join(testDir, "arrow-multi-capture.pine")
	if err := os.WriteFile(strategyPath, []byte(strategy), 0644); err != nil {
		t.Fatal(err)
	}
	dataPath := filepath.Join(testDir, "ARWMCP_1D.json")
	if err := os.WriteFile(dataPath, []byte(generateTestOHLCV(20, 86400)), 0644); err != nil {
		t.Fatal(err)
	}

	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(filepath.Dir(cwd))

	result := compileAndRun(t, strategyPath, dataPath, testDir, projectRoot, "ARWMCP", testDir)

	ind, ok := result.Indicators["Spread"]
	if !ok {
		t.Fatal("indicator 'Spread' absent from output")
	}

	// high[i] = 50100+i, low[i] = 49900+i → spread = 200.0 on every bar
	const expected = 200.0
	vals := extractValues(ind.Data)
	for i, v := range vals {
		if math.IsNaN(v) {
			t.Errorf("bar %d: Spread = NaN, want %.1f", i, expected)
		} else if math.Abs(v-expected) > 1e-9 {
			t.Errorf("bar %d: Spread = %.4f, want %.1f", i, v, expected)
		}
	}
}

/*
TestArrowOuterSeriesCapture_VariableCallSite verifies capture injection through the
generateUserDefinedFunctionCallWithContext code path — exercised when the call result
is stored in a strategy-level variable rather than used inline.

Without Fix F, the second call-site path never consulted arrowCaptureRegistry, so
the captured series argument was absent and compilation failed.
*/
func TestArrowOuterSeriesCapture_VariableCallSite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	strategy := `//@version=5
indicator("Outer Capture Variable Call Site")
base = close
doubled() =>
    base * 2.0
result = doubled()
plot(result, "Doubled")
`
	testDir := t.TempDir()
	strategyPath := filepath.Join(testDir, "arrow-var-callsite.pine")
	if err := os.WriteFile(strategyPath, []byte(strategy), 0644); err != nil {
		t.Fatal(err)
	}
	dataPath := filepath.Join(testDir, "ARWVCS_1D.json")
	if err := os.WriteFile(dataPath, []byte(generateTestOHLCV(20, 86400)), 0644); err != nil {
		t.Fatal(err)
	}

	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(filepath.Dir(cwd))

	result := compileAndRun(t, strategyPath, dataPath, testDir, projectRoot, "ARWVCS", testDir)

	ind, ok := result.Indicators["Doubled"]
	if !ok {
		t.Fatal("indicator 'Doubled' absent from output")
	}

	// close[i] = 50050 + i → doubled = 2 * (50050 + i) = 100100 + 2*i
	vals := extractValues(ind.Data)
	for i, v := range vals {
		expected := 2.0 * (50050.0 + float64(i))
		if math.IsNaN(v) {
			t.Errorf("bar %d: Doubled = NaN, want %.1f", i, expected)
		} else if math.Abs(v-expected) > 1e-9 {
			t.Errorf("bar %d: Doubled = %.4f, want %.4f", i, v, expected)
		}
	}
}
