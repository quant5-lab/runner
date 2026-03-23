package regression

import (
	"math"
	"os"
	"path/filepath"
	"testing"
)

/* TestCompositeIndicatorInline_InternalSeriesAdvanced verifies that composite
 * indicators used as sub-expressions (hoisted to temp vars) correctly advance
 * all internal state series each bar. When internal series are not advanced,
 * Get(j) reads position 0 for all historical offsets, returning NaN and causing
 * the indicator to output zero or NaN for every bar after warmup.
 *
 * This is the ForwardSeriesBuffer invariant: every series written on each bar
 * must also call Next() at the bar boundary.
 */
func TestCompositeIndicatorInline_InternalSeriesAdvanced(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cases := []struct {
		name     string
		pine     string
		plotName string
		wantMin  float64
		wantMax  float64
		minBars  int
	}{
		{
			name: "mfi_in_conditional",
			pine: `//@version=5
indicator("MFI Inline", overlay=false)
signal = ta.mfi(hlc3, 14) >= 50 ? 1.0 : 0.0
plot(ta.mfi(hlc3, 14), "MFI")
`,
			plotName: "MFI",
			wantMin:  0.0,
			wantMax:  100.0,
			minBars:  5,
		},
		{
			name: "rsi_in_conditional",
			pine: `//@version=5
indicator("RSI Inline", overlay=false)
signal = ta.rsi(close, 14) >= 50 ? 1.0 : 0.0
plot(ta.rsi(close, 14), "RSI")
`,
			plotName: "RSI",
			wantMin:  0.0,
			wantMax:  100.0,
			minBars:  5,
		},
		{
			name: "mfi_rsi_ternary_selector",
			pine: `//@version=5
indicator("MFI RSI Selector", overlay=false)
val = true ? ta.rsi(close, 14) : ta.mfi(hlc3, 14)
plot(val, "Selected")
`,
			plotName: "Selected",
			wantMin:  0.0,
			wantMax:  100.0,
			minBars:  5,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			testDir := t.TempDir()
			strategyPath := filepath.Join(testDir, tc.name+".pine")
			if err := os.WriteFile(strategyPath, []byte(tc.pine), 0644); err != nil {
				t.Fatal(err)
			}
			dataPath := filepath.Join(testDir, tc.name+"_data.json")
			if err := os.WriteFile(dataPath, []byte(generateTestOHLCV(60, 3600)), 0644); err != nil {
				t.Fatal(err)
			}

			cwd, _ := os.Getwd()
			projectRoot := filepath.Dir(filepath.Dir(cwd))

			result := compileAndRun(t, strategyPath, dataPath, testDir, projectRoot, "TEST", testDir)

			ind, ok := result.Indicators[tc.plotName]
			if !ok {
				t.Fatalf("indicator %q absent from output", tc.plotName)
			}

			vals := extractValues(ind.Data)
			validCount := 0
			for _, v := range vals {
				if math.IsNaN(v) {
					continue
				}
				validCount++
				if v < tc.wantMin || v > tc.wantMax {
					t.Errorf("value %.4f outside expected [%.1f, %.1f]", v, tc.wantMin, tc.wantMax)
				}
			}

			if validCount < tc.minBars {
				t.Errorf("only %d valid (non-NaN) bars, want at least %d — internal series likely not advancing", validCount, tc.minBars)
			}
		})
	}
}

/* TestCompositeIndicatorInline_NaNFreeAfterWarmup verifies that after the warmup
 * period, composite indicators used inline produce no NaN values. All-zero output
 * from bar 0 to end is the failure mode when internal series are not advanced.
 */
func TestCompositeIndicatorInline_NaNFreeAfterWarmup(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	pine := `//@version=5
indicator("MFI After Warmup", overlay=false)
mfiVal = ta.mfi(hlc3, 5)
plot(mfiVal, "MFI5")
`
	testDir := t.TempDir()
	strategyPath := filepath.Join(testDir, "mfi-warmup.pine")
	if err := os.WriteFile(strategyPath, []byte(pine), 0644); err != nil {
		t.Fatal(err)
	}
	dataPath := filepath.Join(testDir, "mfi-warmup-data.json")
	if err := os.WriteFile(dataPath, []byte(generateTestOHLCV(40, 86400)), 0644); err != nil {
		t.Fatal(err)
	}

	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(filepath.Dir(cwd))

	result := compileAndRun(t, strategyPath, dataPath, testDir, projectRoot, "TEST", testDir)

	ind, ok := result.Indicators["MFI5"]
	if !ok {
		t.Fatal("MFI5 indicator absent from output")
	}

	vals := extractValues(ind.Data)
	warmup := 5

	nanAfterWarmup := 0
	zeroAfterWarmup := 0
	for i := warmup; i < len(vals); i++ {
		if math.IsNaN(vals[i]) {
			nanAfterWarmup++
		}
		if vals[i] == 0.0 {
			zeroAfterWarmup++
		}
	}

	postWarmupBars := len(vals) - warmup
	if nanAfterWarmup > 0 {
		t.Errorf("%d NaN values after warmup — internal series not advancing", nanAfterWarmup)
	}
	if zeroAfterWarmup == postWarmupBars {
		t.Errorf("all %d post-warmup bars are zero — internal series not advancing", postWarmupBars)
	}
}

/* TestCompositeIndicatorInline_HistoricalConsistency verifies that composite
 * indicator values accessed at offset [1] equal the previous bar's value,
 * confirming the ForwardSeriesBuffer cursor advances correctly.
 */
func TestCompositeIndicatorInline_HistoricalConsistency(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	pine := `//@version=5
indicator("MFI Historical", overlay=false)
mfiVal = ta.mfi(hlc3, 5)
plot(mfiVal,    "MFI")
plot(mfiVal[1], "MFI Prev")
`
	testDir := t.TempDir()
	strategyPath := filepath.Join(testDir, "mfi-historical.pine")
	if err := os.WriteFile(strategyPath, []byte(pine), 0644); err != nil {
		t.Fatal(err)
	}
	dataPath := filepath.Join(testDir, "mfi-historical-data.json")
	if err := os.WriteFile(dataPath, []byte(generateTestOHLCV(40, 86400)), 0644); err != nil {
		t.Fatal(err)
	}

	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(filepath.Dir(cwd))

	result := compileAndRun(t, strategyPath, dataPath, testDir, projectRoot, "TEST", testDir)

	mfi, ok := result.Indicators["MFI"]
	if !ok {
		t.Fatal("MFI indicator absent")
	}
	mfiPrev, ok := result.Indicators["MFI Prev"]
	if !ok {
		t.Fatal("MFI Prev indicator absent")
	}

	mfiVals := extractValues(mfi.Data)
	prevVals := extractValues(mfiPrev.Data)

	checked := 0
	for i := 1; i < len(mfiVals) && i < len(prevVals); i++ {
		curr := mfiVals[i-1]
		prev := prevVals[i]
		if math.IsNaN(curr) || math.IsNaN(prev) {
			continue
		}
		if math.Abs(curr-prev) > 0.001 {
			t.Errorf("bar %d: mfi[1] = %.4f, mfi at bar %d = %.4f — series cursor not advancing", i, prev, i-1, curr)
		}
		checked++
	}

	if checked == 0 {
		t.Error("no bars with valid MFI and MFI[1] to compare — warmup too long or series not computing")
	}
}

/* TestCompositeIndicatorInline_TernaryConditionWithIndicators verifies that a
 * ternary whose condition is itself a ternary producing an indicator comparison
 * correctly hoists both indicators and evaluates them every bar.
 *
 * Pattern: (boolFlag ? rsi >= threshold : mfi >= threshold) ? branchA : branchB
 *
 * Both ta.rsi and ta.mfi must be hoisted and advanced regardless of which branch
 * the outer ternary selects.
 */
func TestCompositeIndicatorInline_TernaryConditionWithIndicators(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cases := []struct {
		name       string
		pine       string
		plotName   string
		warmupBars int
	}{
		{
			name: "indicator_comparison_as_condition",
			pine: `//@version=5
indicator("Indicator Condition", overlay=false)
useRSI = true
val = (useRSI ? ta.rsi(close, 5) : ta.mfi(hlc3, 5)) >= 50 ? 1.0 : -1.0
plot(val, "Signal")
`,
			plotName:   "Signal",
			warmupBars: 5,
		},
		{
			name: "ternary_condition_is_indicator_ternary",
			pine: `//@version=5
indicator("Nested Ternary Condition", overlay=false)
useRSI = false
outerCond = useRSI ? ta.rsi(close, 5) >= 50 : ta.mfi(hlc3, 5) >= 50
val = outerCond ? 1.0 : -1.0
plot(val, "Signal")
`,
			plotName:   "Signal",
			warmupBars: 5,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			testDir := t.TempDir()
			strategyPath := filepath.Join(testDir, tc.name+".pine")
			if err := os.WriteFile(strategyPath, []byte(tc.pine), 0644); err != nil {
				t.Fatal(err)
			}
			dataPath := filepath.Join(testDir, tc.name+"_data.json")
			if err := os.WriteFile(dataPath, []byte(generateTestOHLCV(60, 3600)), 0644); err != nil {
				t.Fatal(err)
			}

			cwd, _ := os.Getwd()
			projectRoot := filepath.Dir(filepath.Dir(cwd))

			result := compileAndRun(t, strategyPath, dataPath, testDir, projectRoot, "TEST", testDir)

			ind, ok := result.Indicators[tc.plotName]
			if !ok {
				t.Fatalf("indicator %q absent from output", tc.plotName)
			}

			vals := extractValues(ind.Data)

			nanAfterWarmup := 0
			zeroAfterWarmup := 0
			for i := tc.warmupBars; i < len(vals); i++ {
				if math.IsNaN(vals[i]) {
					nanAfterWarmup++
				}
				if vals[i] == 0.0 {
					zeroAfterWarmup++
				}
			}

			postWarmup := len(vals) - tc.warmupBars
			if nanAfterWarmup > 0 {
				t.Errorf("%d NaN values after warmup — indicator not hoisted or internal series not advancing", nanAfterWarmup)
			}
			if zeroAfterWarmup == postWarmup {
				t.Errorf("all %d post-warmup bars are zero — condition never evaluates non-zero branch", zeroAfterWarmup)
			}
		})
	}
}
