package regression

import (
	"math"
	"os"
	"path/filepath"
	"testing"
)

// TestFixnan_InSecurity_ForwardFillsPivotValue verifies that fixnan(ta.pivothigh)
// inside request.security() carries the pivot value forward on every NaN bar
// between detections, and that the carry begins at the first detection bar.
func TestFixnan_InSecurity_ForwardFillsPivotValue(t *testing.T) {
	strategy := `//@version=5
indicator("Fixnan Pivot Fill", overlay=false)
phf = request.security(syminfo.tickerid, "1D", fixnan(ta.pivothigh(high, 2, 2)))
plot(phf, "PHF")
`
	testDir := t.TempDir()
	strategyPath := filepath.Join(testDir, "fixnan-pivot-fill.pine")
	if err := os.WriteFile(strategyPath, []byte(strategy), 0644); err != nil {
		t.Fatal(err)
	}
	dataPath := filepath.Join(testDir, "FIXPIVFILL_1D.json")
	if err := os.WriteFile(dataPath, []byte(generateWaveOHLCV(30, 86400)), 0644); err != nil {
		t.Fatal(err)
	}

	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(filepath.Dir(cwd))

	result := compileAndRun(t, strategyPath, dataPath, testDir, projectRoot, "FIXPIVFILL", testDir)

	ind, ok := result.Indicators["PHF"]
	if !ok {
		t.Fatal("indicator 'PHF' absent from output")
	}
	vals := extractValues(ind.Data)

	// Wave peak at sec bar 5 (High=50500), pivothigh(2,2) detects at sec bar 7.
	// 1-bar security lag: base bar 8 receives the detection. fixnan carries 50500
	// from base bar 8 onward; the next identical peak at sec bar 17 → base bar 18
	// does not alter the carried value.
	const firstFillBar = 8
	const expectedValue = 50500.0

	t.Run("null_before_first_detection", func(t *testing.T) {
		for i := 0; i < firstFillBar && i < len(vals); i++ {
			if !math.IsNaN(vals[i]) {
				t.Errorf("bar %d: expected NaN before first detection, got %.4f", i, vals[i])
			}
		}
	})

	t.Run("carry_forward_from_first_detection", func(t *testing.T) {
		for i := firstFillBar; i < len(vals); i++ {
			if math.IsNaN(vals[i]) {
				t.Errorf("bar %d: expected %.0f (carry-forward), got NaN", i, expectedValue)
			} else if math.Abs(vals[i]-expectedValue) > 1e-6 {
				t.Errorf("bar %d: expected %.0f, got %.4f", i, expectedValue, vals[i])
			}
		}
	})
}

// TestFixnan_InSecurity_TAFunctionSource verifies that fixnan wrapping a TA
// function (ta.sma) inside request.security() forward-fills NaN produced during
// the TA warmup period, then passes valid values through unchanged.
func TestFixnan_InSecurity_TAFunctionSource(t *testing.T) {
	strategy := `//@version=5
indicator("Fixnan SMA Fill", overlay=false)
smaFixed = request.security(syminfo.tickerid, "1D", fixnan(ta.sma(close, 3)))
plot(smaFixed, "SMAFixed")
`
	testDir := t.TempDir()
	strategyPath := filepath.Join(testDir, "fixnan-sma-fill.pine")
	if err := os.WriteFile(strategyPath, []byte(strategy), 0644); err != nil {
		t.Fatal(err)
	}
	dataPath := filepath.Join(testDir, "FIXSMAFILL_1D.json")
	if err := os.WriteFile(dataPath, []byte(generateTestOHLCV(30, 86400)), 0644); err != nil {
		t.Fatal(err)
	}

	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(filepath.Dir(cwd))

	result := compileAndRun(t, strategyPath, dataPath, testDir, projectRoot, "FIXSMAFILL", testDir)

	ind, ok := result.Indicators["SMAFixed"]
	if !ok {
		t.Fatal("indicator 'SMAFixed' absent from output")
	}
	vals := extractValues(ind.Data)

	// generateTestOHLCV: close[i] = 50050+i.
	// SMA(3) at sec bar i = 50049+i (valid for i≥2, NaN for i<2).
	// 1-bar security lag: base bar k (k≥1) maps to sec bar k-1.
	// base bar 0 → sec 0: NaN; base bar 1 → sec 0: NaN; base bar 2 → sec 1: NaN.
	// base bar 3 → sec 2: SMA = 50051; base bar k (k≥3) → 50048+k.
	const warmupBars = 3

	t.Run("null_during_sma_warmup", func(t *testing.T) {
		for i := 0; i < warmupBars && i < len(vals); i++ {
			if !math.IsNaN(vals[i]) {
				t.Errorf("bar %d: expected NaN during SMA warmup, got %.4f", i, vals[i])
			}
		}
	})

	t.Run("valid_after_warmup", func(t *testing.T) {
		for k := warmupBars; k < len(vals); k++ {
			expected := 50048.0 + float64(k)
			if math.IsNaN(vals[k]) {
				t.Errorf("bar %d: expected %.1f, got NaN", k, expected)
			} else if math.Abs(vals[k]-expected) > 1e-4 {
				t.Errorf("bar %d: expected %.1f, got %.4f", k, expected, vals[k])
			}
		}
	})
}

// TestFixnan_InSecurity_TwoExpressionsAreIsolated verifies that two fixnan
// expressions evaluated by the same evaluator inside request.security() maintain
// independent forwardBufferE series and do not share state.
func TestFixnan_InSecurity_TwoExpressionsAreIsolated(t *testing.T) {
	strategy := `//@version=5
indicator("Fixnan Isolation", overlay=false)
phf = request.security(syminfo.tickerid, "1D", fixnan(ta.pivothigh(high, 2, 2)))
plf = request.security(syminfo.tickerid, "1D", fixnan(ta.pivotlow(low, 2, 2)))
plot(phf, "PHF")
plot(plf, "PLF")
`
	testDir := t.TempDir()
	strategyPath := filepath.Join(testDir, "fixnan-isolation.pine")
	if err := os.WriteFile(strategyPath, []byte(strategy), 0644); err != nil {
		t.Fatal(err)
	}
	dataPath := filepath.Join(testDir, "FIXISOLATE_1D.json")
	if err := os.WriteFile(dataPath, []byte(generateWaveOHLCV(30, 86400)), 0644); err != nil {
		t.Fatal(err)
	}

	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(filepath.Dir(cwd))

	result := compileAndRun(t, strategyPath, dataPath, testDir, projectRoot, "FIXISOLATE", testDir)

	phfInd, ok := result.Indicators["PHF"]
	if !ok {
		t.Fatal("indicator 'PHF' absent from output")
	}
	plfInd, ok := result.Indicators["PLF"]
	if !ok {
		t.Fatal("indicator 'PLF' absent from output")
	}

	phf := extractValues(phfInd.Data)
	plf := extractValues(plfInd.Data)

	// Wave peak (High=50500) and trough (Low=49500) both at sec bar 5, detected
	// at sec bar 7. 1-bar security lag: base bar 8 carries each value independently.
	const firstFillBar = 8
	const expectedHigh = 50500.0
	const expectedLow = 49500.0

	t.Run("null_before_first_detection", func(t *testing.T) {
		for i := 0; i < firstFillBar; i++ {
			if i < len(phf) && !math.IsNaN(phf[i]) {
				t.Errorf("PHF bar %d: expected NaN, got %.4f", i, phf[i])
			}
			if i < len(plf) && !math.IsNaN(plf[i]) {
				t.Errorf("PLF bar %d: expected NaN, got %.4f", i, plf[i])
			}
		}
	})

	t.Run("phf_carries_peak_value", func(t *testing.T) {
		for i := firstFillBar; i < len(phf); i++ {
			if math.IsNaN(phf[i]) {
				t.Errorf("PHF bar %d: expected %.0f, got NaN", i, expectedHigh)
			} else if math.Abs(phf[i]-expectedHigh) > 1e-6 {
				t.Errorf("PHF bar %d: expected %.0f, got %.4f", i, expectedHigh, phf[i])
			}
		}
	})

	t.Run("plf_carries_trough_value", func(t *testing.T) {
		for i := firstFillBar; i < len(plf); i++ {
			if math.IsNaN(plf[i]) {
				t.Errorf("PLF bar %d: expected %.0f, got NaN", i, expectedLow)
			} else if math.Abs(plf[i]-expectedLow) > 1e-6 {
				t.Errorf("PLF bar %d: expected %.0f, got %.4f", i, expectedLow, plf[i])
			}
		}
	})

	t.Run("phf_and_plf_differ", func(t *testing.T) {
		if len(phf) > firstFillBar && len(plf) > firstFillBar {
			if math.Abs(phf[firstFillBar]-plf[firstFillBar]) < 1e-6 {
				t.Errorf("PHF and PLF share the same value %.4f — isolation broken", phf[firstFillBar])
			}
		}
	})
}
