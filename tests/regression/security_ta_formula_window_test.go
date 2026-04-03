package regression

import (
	"math"
	"os"
	"path/filepath"
	"testing"
)

/*
TestSecurityTA_FormulaFunctions_AllCompileAndRun verifies that ta.change, ta.mom,
ta.roc, ta.crossover, ta.crossunder, ta.cross, ta.falling, ta.rising, ta.barssince,
ta.cum, math.max, math.min, and math.abs each survive the full
codegen→compile→execute pipeline when evaluated inside request.security(),
producing at least one non-null output bar.
*/
func TestSecurityTA_FormulaFunctions_AllCompileAndRun(t *testing.T) {
	strategy := `//@version=5
indicator("Security Formula Functions", overlay=false)
chg      = request.security(syminfo.tickerid, "1D", ta.change(close))
mom2     = request.security(syminfo.tickerid, "1D", ta.mom(close, 2))
roc2     = request.security(syminfo.tickerid, "1D", ta.roc(close, 2))
xover    = request.security(syminfo.tickerid, "1D", ta.crossover(close, open))
xunder   = request.security(syminfo.tickerid, "1D", ta.crossunder(close, open))
cross_   = request.security(syminfo.tickerid, "1D", ta.cross(close, open))
falling_ = request.security(syminfo.tickerid, "1D", ta.falling(close, 3))
rising_  = request.security(syminfo.tickerid, "1D", ta.rising(close, 3))
bsince   = request.security(syminfo.tickerid, "1D", ta.barssince(ta.change(close) > 0))
cum_     = request.security(syminfo.tickerid, "1D", ta.cum(close))
maxco    = request.security(syminfo.tickerid, "1D", math.max(close, open))
minco    = request.security(syminfo.tickerid, "1D", math.min(close, open))
absdif   = request.security(syminfo.tickerid, "1D", math.abs(close - open))
plot(chg,      "CHG")
plot(mom2,     "MOM")
plot(roc2,     "ROC")
plot(xover,    "XOVER")
plot(xunder,   "XUNDER")
plot(cross_,   "CROSS")
plot(falling_, "FALLING")
plot(rising_,  "RISING")
plot(bsince,   "BSINCE")
plot(cum_,     "CUM")
plot(maxco,    "MAXCO")
plot(minco,    "MINCO")
plot(absdif,   "ABSDIF")
`
	testDir := t.TempDir()
	strategyPath := filepath.Join(testDir, "formula-functions.pine")
	if err := os.WriteFile(strategyPath, []byte(strategy), 0644); err != nil {
		t.Fatal(err)
	}
	dataPath := filepath.Join(testDir, "FORMFN_1D.json")
	if err := os.WriteFile(dataPath, []byte(generateTestOHLCV(40, 86400)), 0644); err != nil {
		t.Fatal(err)
	}

	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(filepath.Dir(cwd))

	result := compileAndRun(t, strategyPath, dataPath, testDir, projectRoot, "FORMFN", testDir)

	for _, name := range []string{
		"CHG", "MOM", "ROC", "XOVER", "XUNDER", "CROSS",
		"FALLING", "RISING", "BSINCE", "CUM", "MAXCO", "MINCO", "ABSDIF",
	} {
		t.Run(name, func(t *testing.T) {
			ind, ok := result.Indicators[name]
			if !ok {
				t.Fatalf("indicator %q absent from output", name)
			}
			if countNonNull(ind.Data) == 0 {
				t.Errorf("indicator %q produced zero non-null values across 40 bars", name)
			}
		})
	}

	/* generateTestOHLCV: close - open = (50000+i+50) - (50000+i) = 50 on every bar.
	   math.abs(50) = 50 is a constant, non-NaN invariant across all non-null bars. */
	t.Run("ABSDIF_arithmetic", func(t *testing.T) {
		ind, ok := result.Indicators["ABSDIF"]
		if !ok {
			t.Fatal("ABSDIF indicator absent from output")
		}
		for i, bar := range ind.Data {
			if v, ok := getFloatValue(bar); ok {
				if math.Abs(v-50.0) > 1e-6 {
					t.Errorf("bar %d: math.abs(close-open) = %.9f, want 50.0", i, v)
				}
			}
		}
	})
}

/*
TestSecurityTA_WindowFunctions_AllCompileAndRun verifies that ta.highest, ta.lowest,
ta.sum, ta.range, ta.dev, ta.variance, ta.median, ta.mode, ta.cmo, ta.wpr, ta.mfi,
ta.vwma, ta.linreg, ta.highestbars, and ta.lowestbars each survive the full
codegen→compile→execute pipeline when evaluated inside request.security(),
producing at least one non-null output bar.
*/
func TestSecurityTA_WindowFunctions_AllCompileAndRun(t *testing.T) {
	strategy := `//@version=5
indicator("Security Window Functions", overlay=false)
highest5 = request.security(syminfo.tickerid, "1D", ta.highest(close, 5))
lowest5  = request.security(syminfo.tickerid, "1D", ta.lowest(close, 5))
sum5     = request.security(syminfo.tickerid, "1D", ta.sum(close, 5))
tarange5 = request.security(syminfo.tickerid, "1D", ta.range(close, 5))
dev5     = request.security(syminfo.tickerid, "1D", ta.dev(close, 5))
var5     = request.security(syminfo.tickerid, "1D", ta.variance(close, 5))
med5     = request.security(syminfo.tickerid, "1D", ta.median(close, 5))
mode5    = request.security(syminfo.tickerid, "1D", ta.mode(close, 5))
cmo14    = request.security(syminfo.tickerid, "1D", ta.cmo(close, 14))
wpr14    = request.security(syminfo.tickerid, "1D", ta.wpr(14))
mfi14    = request.security(syminfo.tickerid, "1D", ta.mfi(close, 14))
vwma5    = request.security(syminfo.tickerid, "1D", ta.vwma(close, 5))
linreg5  = request.security(syminfo.tickerid, "1D", ta.linreg(close, 5, 0))
hbars5   = request.security(syminfo.tickerid, "1D", ta.highestbars(close, 5))
lbars5   = request.security(syminfo.tickerid, "1D", ta.lowestbars(close, 5))
plot(highest5, "HIGHEST5")
plot(lowest5,  "LOWEST5")
plot(sum5,     "SUM5")
plot(tarange5, "RANGE5")
plot(dev5,     "DEV5")
plot(var5,     "VAR5")
plot(med5,     "MED5")
plot(mode5,    "MODE5")
plot(cmo14,    "CMO14")
plot(wpr14,    "WPR14")
plot(mfi14,    "MFI14")
plot(vwma5,    "VWMA5")
plot(linreg5,  "LINREG5")
plot(hbars5,   "HBARS5")
plot(lbars5,   "LBARS5")
`
	testDir := t.TempDir()
	strategyPath := filepath.Join(testDir, "window-functions.pine")
	if err := os.WriteFile(strategyPath, []byte(strategy), 0644); err != nil {
		t.Fatal(err)
	}
	dataPath := filepath.Join(testDir, "WINFN_1D.json")
	if err := os.WriteFile(dataPath, []byte(generateTestOHLCV(40, 86400)), 0644); err != nil {
		t.Fatal(err)
	}

	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(filepath.Dir(cwd))

	result := compileAndRun(t, strategyPath, dataPath, testDir, projectRoot, "WINFN", testDir)

	for _, name := range []string{
		"HIGHEST5", "LOWEST5", "SUM5", "RANGE5", "DEV5", "VAR5",
		"MED5", "MODE5", "CMO14", "WPR14", "MFI14", "VWMA5", "LINREG5", "HBARS5", "LBARS5",
	} {
		t.Run(name, func(t *testing.T) {
			ind, ok := result.Indicators[name]
			if !ok {
				t.Fatalf("indicator %q absent from output", name)
			}
			if countNonNull(ind.Data) == 0 {
				t.Errorf("indicator %q produced zero non-null values across 40 bars", name)
			}
		})
	}
}

/*
TestSecurityTA_BarsSince_CounterAndWarmup verifies that ta.barssince evaluated inside
request.security() produces NaN during warmup and resets the counter to zero when the
condition first fires, through the full codegen→compile→execute pipeline.
generateTestOHLCV emits monotonically rising closes, so ta.change(close) = 1.0 on every
post-warmup security bar: the condition ta.change(close) > 0 is always true, making
barssince always 0 after the first valid bar.
*/
func TestSecurityTA_BarsSince_CounterAndWarmup(t *testing.T) {
	strategy := `//@version=5
indicator("BarsSince Counter", overlay=false)
bsince = request.security(syminfo.tickerid, "1D", ta.barssince(ta.change(close) > 0))
plot(bsince, "BSINCE")
`
	testDir := t.TempDir()
	strategyPath := filepath.Join(testDir, "barssince-counter.pine")
	if err := os.WriteFile(strategyPath, []byte(strategy), 0644); err != nil {
		t.Fatal(err)
	}
	dataPath := filepath.Join(testDir, "BSCNT_1D.json")
	if err := os.WriteFile(dataPath, []byte(generateTestOHLCV(40, 86400)), 0644); err != nil {
		t.Fatal(err)
	}

	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(filepath.Dir(cwd))

	result := compileAndRun(t, strategyPath, dataPath, testDir, projectRoot, "BSCNT", testDir)

	ind, ok := result.Indicators["BSINCE"]
	if !ok {
		t.Fatal("BSINCE indicator absent from output")
	}
	vals := extractValues(ind.Data)

	/* 1-bar lookahead_off lag: base bar i maps to secBarIdx i-1.
	   secBar 0: ta.change warmup (NaN) → barssince condition is NaN/false → NaN.
	   secBar 1+: ta.change = 1.0, 1.0 > 0 = true → barssince = 0.
	   base bar 0: no secBar mapping → NaN.
	   base bar 1: secBar 0 → NaN.
	   base bars 2+: secBar 1+ → barssince = 0. */
	for i := 0; i <= 1 && i < len(vals); i++ {
		if !math.IsNaN(vals[i]) {
			t.Errorf("bar %d: expected NaN during warmup, got %.6f", i, vals[i])
		}
	}

	for i := 2; i < len(vals); i++ {
		if math.IsNaN(vals[i]) {
			t.Errorf("bar %d: unexpected NaN (barssince should be 0 with always-true condition post-warmup)", i)
			continue
		}
		if vals[i] != 0.0 {
			t.Errorf("bar %d: barssince = %.1f, want 0 (condition always true, counter resets every bar)", i, vals[i])
		}
	}
}

/*
TestSecurityTA_BarsSince_IndependentStateManagers verifies that two distinct
ta.barssince expressions evaluated within the same request.security() context
maintain independent ForwardSeriesBuffer counter state — neither condition's fire
history leaks into the other's counter, even when they first fire at different bars.
*/
func TestSecurityTA_BarsSince_IndependentStateManagers(t *testing.T) {
	strategy := `//@version=5
indicator("BarsSince State Isolation", overlay=false)
bs_change = request.security(syminfo.tickerid, "1D", ta.barssince(ta.change(close) > 0))
bs_65     = request.security(syminfo.tickerid, "1D", ta.barssince(close > 50065.0))
plot(bs_change, "BS_CHANGE")
plot(bs_65,     "BS_65")
`
	testDir := t.TempDir()
	strategyPath := filepath.Join(testDir, "barssince-isolation.pine")
	if err := os.WriteFile(strategyPath, []byte(strategy), 0644); err != nil {
		t.Fatal(err)
	}
	dataPath := filepath.Join(testDir, "BSISO_1D.json")
	if err := os.WriteFile(dataPath, []byte(generateTestOHLCV(40, 86400)), 0644); err != nil {
		t.Fatal(err)
	}

	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(filepath.Dir(cwd))

	result := compileAndRun(t, strategyPath, dataPath, testDir, projectRoot, "BSISO", testDir)

	bsChangeInd, ok := result.Indicators["BS_CHANGE"]
	if !ok {
		t.Fatal("BS_CHANGE indicator absent from output")
	}
	bs65Ind, ok := result.Indicators["BS_65"]
	if !ok {
		t.Fatal("BS_65 indicator absent from output")
	}

	changeVals := extractValues(bsChangeInd.Data)
	vals65 := extractValues(bs65Ind.Data)

	/* generateTestOHLCV: Close[i] = 50050 + i.
	   bs_change fires every bar after change warmup → barssince = 0 from base bar 2.
	   bs_65 (close > 50065): close[16] = 50066 > 50065 → first fires at secBar 16 = base bar 17.
	   At base bar 3 (secBar 2): bs_change = 0, bs_65 = NaN (close[2]=50052 ≯ 50065).
	   Shared-state corruption would cause bs_65 to return 0 instead of NaN at bar 3. */
	if len(changeVals) > 3 && len(vals65) > 3 {
		t.Run("early_bar_isolation", func(t *testing.T) {
			if math.IsNaN(changeVals[3]) {
				t.Errorf("bar 3: bs_change expected 0, got NaN")
			} else if changeVals[3] != 0.0 {
				t.Errorf("bar 3: bs_change = %.1f, want 0", changeVals[3])
			}
			if !math.IsNaN(vals65[3]) {
				t.Errorf("bar 3: bs_65 = %.1f, want NaN (condition not yet fired) — possible state leak from bs_change", vals65[3])
			}
		})
	}

	/* At base bar 16 (secBar 15): close[15] = 50065, 50065 ≯ 50065 (strict >) → bs_65 still NaN. */
	if len(vals65) > 16 {
		t.Run("boundary_bar_still_nan", func(t *testing.T) {
			if !math.IsNaN(vals65[16]) {
				t.Errorf("bar 16: bs_65 = %.1f, want NaN (close[15]=50065 not strictly > 50065)", vals65[16])
			}
		})
	}

	/* At base bar 17 (secBar 16): close[16] = 50066 > 50065 → first fire → bs_65 = 0. */
	if len(changeVals) > 17 && len(vals65) > 17 {
		t.Run("first_fire_bar", func(t *testing.T) {
			if math.IsNaN(changeVals[17]) {
				t.Errorf("bar 17: bs_change expected 0, got NaN")
			} else if changeVals[17] != 0.0 {
				t.Errorf("bar 17: bs_change = %.1f, want 0", changeVals[17])
			}
			if math.IsNaN(vals65[17]) {
				t.Errorf("bar 17: bs_65 expected 0 (first fire at secBar 16), got NaN")
			} else if vals65[17] != 0.0 {
				t.Errorf("bar 17: bs_65 = %.1f, want 0", vals65[17])
			}
		})
	}
}
