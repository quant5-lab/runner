package regression

import (
	"math"
	"os"
	"path/filepath"
	"testing"
)

/*
extractValues converts IndicatorData rows into a float64 slice, mapping absent

	values to NaN so index alignment with the bar series is preserved.
*/
func extractValues(data []map[string]interface{}) []float64 {
	vals := make([]float64, len(data))
	for i, bar := range data {
		if v, ok := getFloatValue(bar); ok {
			vals[i] = v
		} else {
			vals[i] = math.NaN()
		}
	}
	return vals
}

/*
TestVolumeIndicators_AllCompileAndRun verifies all 8 ta.* volume variables survive

	the full codegen→compile→execute pipeline and produce output on standard OHLCV data.
*/
func TestVolumeIndicators_AllCompileAndRun(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	strategy := `//@version=5
indicator("Volume Indicators All", overlay=false)
plot(ta.obv,     "OBV")
plot(ta.accdist, "ACCDIST")
plot(ta.pvt,     "PVT")
plot(ta.iii,     "III")
plot(ta.wvad,    "WVAD")
plot(ta.nvi,     "NVI")
plot(ta.pvi,     "PVI")
plot(ta.wad,     "WAD")
`
	testDir := t.TempDir()
	strategyPath := filepath.Join(testDir, "volume-all.pine")
	if err := os.WriteFile(strategyPath, []byte(strategy), 0644); err != nil {
		t.Fatal(err)
	}
	dataPath := filepath.Join(testDir, "VOLALL_1D.json")
	if err := os.WriteFile(dataPath, []byte(generateTestOHLCV(20, 86400)), 0644); err != nil {
		t.Fatal(err)
	}

	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(filepath.Dir(cwd))

	result := compileAndRun(t, strategyPath, dataPath, testDir, projectRoot, "VOLALL", testDir)

	for _, name := range []string{"OBV", "ACCDIST", "PVT", "III", "WVAD", "NVI", "PVI", "WAD"} {
		t.Run(name, func(t *testing.T) {
			ind, ok := result.Indicators[name]
			if !ok {
				t.Fatalf("indicator %q absent from output", name)
			}
			if countNonNull(ind.Data) == 0 {
				t.Errorf("indicator %q produced zero non-null values", name)
			}
		})
	}
}

/*
TestOBVIndicator_MonotonicPriceAccumulation verifies OBV grows on every bar of a

	monotonically rising price series — the defining domain invariant of OBV.
*/
func TestOBVIndicator_MonotonicPriceAccumulation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	strategy := `//@version=5
indicator("OBV Monotonic", overlay=false)
plot(ta.obv, "OBV")
`
	testDir := t.TempDir()
	strategyPath := filepath.Join(testDir, "obv-monotonic.pine")
	if err := os.WriteFile(strategyPath, []byte(strategy), 0644); err != nil {
		t.Fatal(err)
	}
	dataPath := filepath.Join(testDir, "OBVMON_1D.json")
	if err := os.WriteFile(dataPath, []byte(generateTestOHLCV(20, 86400)), 0644); err != nil {
		t.Fatal(err)
	}

	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(filepath.Dir(cwd))

	result := compileAndRun(t, strategyPath, dataPath, testDir, projectRoot, "OBVMON", testDir)

	ind, ok := result.Indicators["OBV"]
	if !ok {
		t.Fatal("OBV indicator not found in output")
	}

	vals := extractValues(ind.Data)

	for i := 1; i < len(vals); i++ {
		if math.IsNaN(vals[i]) || math.IsNaN(vals[i-1]) {
			continue
		}
		if vals[i] <= vals[i-1] {
			t.Errorf("bar %d: OBV did not increase on rising close (%.2f <= %.2f)", i, vals[i], vals[i-1])
		}
	}

	last := vals[len(vals)-1]
	if math.IsNaN(last) || last <= 0 {
		t.Errorf("last bar OBV = %v, want > 0 on monotonic rising series", last)
	}
}

/*
TestNVIPVI_SeedAt1000_ConstantVolume verifies NVI and PVI are seeded at 1000 and

	remain there when volume never changes — neither indicator's trigger fires.
*/
func TestNVIPVI_SeedAt1000_ConstantVolume(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	strategy := `//@version=5
indicator("NVI PVI Seed", overlay=false)
plot(ta.nvi, "NVI")
plot(ta.pvi, "PVI")
`
	testDir := t.TempDir()
	strategyPath := filepath.Join(testDir, "nvi-pvi-seed.pine")
	if err := os.WriteFile(strategyPath, []byte(strategy), 0644); err != nil {
		t.Fatal(err)
	}
	/* generateTestOHLCV produces volume=100 on every bar, so no volume change occurs */
	dataPath := filepath.Join(testDir, "NVIPVI_1D.json")
	if err := os.WriteFile(dataPath, []byte(generateTestOHLCV(15, 86400)), 0644); err != nil {
		t.Fatal(err)
	}

	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(filepath.Dir(cwd))

	result := compileAndRun(t, strategyPath, dataPath, testDir, projectRoot, "NVIPVI", testDir)

	for _, name := range []string{"NVI", "PVI"} {
		t.Run(name, func(t *testing.T) {
			ind, ok := result.Indicators[name]
			if !ok {
				t.Fatalf("indicator %q not found in output", name)
			}
			for i, v := range extractValues(ind.Data) {
				if math.IsNaN(v) {
					t.Errorf("bar %d: %s is NaN, want 1000", i, name)
					continue
				}
				if math.Abs(v-1000) > 0.001 {
					t.Errorf("bar %d: %s = %.4f, want 1000 (seed held with flat volume)", i, name, v)
				}
			}
		})
	}
}

/*
TestVolumeIndicator_HistoricalAccess verifies ta.obv[1] at bar N equals ta.obv at

	bar N-1, confirming ForwardSeriesBuffer historical indexing is wired correctly.
*/
func TestVolumeIndicator_HistoricalAccess(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	strategy := `//@version=5
indicator("OBV Historical Access", overlay=false)
plot(ta.obv,    "OBV")
plot(ta.obv[1], "OBV Prev")
`
	testDir := t.TempDir()
	strategyPath := filepath.Join(testDir, "obv-historical.pine")
	if err := os.WriteFile(strategyPath, []byte(strategy), 0644); err != nil {
		t.Fatal(err)
	}
	dataPath := filepath.Join(testDir, "OBVHIST_1D.json")
	if err := os.WriteFile(dataPath, []byte(generateTestOHLCV(20, 86400)), 0644); err != nil {
		t.Fatal(err)
	}

	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(filepath.Dir(cwd))

	result := compileAndRun(t, strategyPath, dataPath, testDir, projectRoot, "OBVHIST", testDir)

	obv, ok := result.Indicators["OBV"]
	if !ok {
		t.Fatal("OBV indicator not found")
	}
	obvPrev, ok := result.Indicators["OBV Prev"]
	if !ok {
		t.Fatal("OBV Prev indicator not found")
	}

	vals := extractValues(obv.Data)
	prevVals := extractValues(obvPrev.Data)

	for i := 1; i < len(vals) && i < len(prevVals); i++ {
		if math.IsNaN(vals[i-1]) || math.IsNaN(prevVals[i]) {
			continue
		}
		if math.Abs(vals[i-1]-prevVals[i]) > 0.001 {
			t.Errorf("bar %d: ta.obv[1] = %.4f, ta.obv at bar %d = %.4f, want equal",
				i, prevVals[i], i-1, vals[i-1])
		}
	}
}

/*
TestVolumeIndicator_SecurityEvaluatorPath verifies ta.obv is accessible as a

	request.security() expression, routed through StreamingBarEvaluator.
*/
func TestVolumeIndicator_SecurityEvaluatorPath(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	strategy := `//@version=5
indicator("OBV via Security", overlay=false)
obvSec = request.security(syminfo.tickerid, "1D", ta.obv)
plot(obvSec, "OBV Security")
`
	testDir := t.TempDir()
	strategyPath := filepath.Join(testDir, "obv-security.pine")
	if err := os.WriteFile(strategyPath, []byte(strategy), 0644); err != nil {
		t.Fatal(err)
	}
	dailyData := generateTestOHLCV(20, 86400)
	dataPath := filepath.Join(testDir, "OBVSEC_1D.json")
	if err := os.WriteFile(dataPath, []byte(dailyData), 0644); err != nil {
		t.Fatal(err)
	}

	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(filepath.Dir(cwd))

	result := compileAndRun(t, strategyPath, dataPath, testDir, projectRoot, "OBVSEC", testDir)

	ind, ok := result.Indicators["OBV Security"]
	if !ok {
		t.Fatal("'OBV Security' indicator not found in output")
	}
	if countNonNull(ind.Data) == 0 {
		t.Error("request.security(ta.obv) produced zero non-null values")
	}
}
