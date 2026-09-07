package regression

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"testing"
)

// syntheticTRValue is the constant ta.tr output for every bar ≥ 1 produced by
// generateTestOHLCV: max(H-L=200, |H-prevClose|=51, |L-prevClose|=149) = 200.0.
const syntheticTRValue = 200.0

/*
TestTrDirect_Bar0NaN verifies that plot(ta.tr) produces NaN on bar 0 and
syntheticTRValue on bar 1.

ta.tr with handle_na=false (the default) has no previous close on bar 0, so the
true range formula is undefined — Pine returns na. Bar 1 onward has a prior close,
so the full max(H-L, |H-pC|, |L-pC|) formula applies.
*/
func TestTrDirect_Bar0NaN(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	assertWarmUp(t,
		"//@version=5\nindicator(\"TR Direct\")\nplot(ta.tr, \"TR\")\n",
		"TR", 1, syntheticTRValue,
	)
}

/*
TestTrSource_SmaWarmUp verifies that ta.sma(ta.tr, period) produces NaN for bars
0 through period-1 inclusive and exactly syntheticTRValue at bar period.

ta.tr default (handle_na=false) returns na on bar 0 because there is no previous
close. That na propagates through the SMA window, so the first fully valid window
spans bars 1..period, making bar period (0-indexed) the first valid output bar.
*/
func TestTrSource_SmaWarmUp(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	for _, tc := range []struct {
		name   string
		period int
	}{
		{"period_1", 1},
		{"period_3", 3},
		{"period_5", 5},
		{"period_14", 14},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assertWarmUp(t, buildTrSmaPineScript(tc.period), "SMA", tc.period, syntheticTRValue)
		})
	}
}

/*
TestTrSource_SmaValidityBeyondSeed verifies that ta.sma(ta.tr, period) sustains
valid output for multiple consecutive bars after the first valid bar at bar period.

Each bar's SMA window [i-period+1..i] for i ≥ period contains only bars ≥ 1 where
ta.tr = syntheticTRValue. On synthetic data where all valid ta.tr values are constant,
every post-seed bar equals syntheticTRValue exactly.
*/
func TestTrSource_SmaValidityBeyondSeed(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	for _, tc := range []struct {
		name   string
		period int
	}{
		{"period_1", 1},
		{"period_3", 3},
		{"period_5", 5},
		{"period_14", 14},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assertValidityBeyondSeed(t, buildTrSmaPineScript(tc.period), "SMA", tc.period)
		})
	}
}

/*
TestTrSource_EmaWarmUp verifies that ta.ema(ta.tr, period) observes the same
warm-up boundary as ta.sma(ta.tr, period): NaN for bars 0..period-1, valid from
bar period. EMA seeding with SMA propagates the bar-0 na through the initial
window identically to SMA.
*/
func TestTrSource_EmaWarmUp(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	for _, tc := range []struct {
		name   string
		period int
	}{
		{"period_1", 1},
		{"period_3", 3},
		{"period_5", 5},
		{"period_14", 14},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assertWarmUp(t, buildTrEmaPineScript(tc.period), "EMA", tc.period, syntheticTRValue)
		})
	}
}

/*
TestTrSource_EmaValidityBeyondSeed verifies that ta.ema(ta.tr, period) sustains
valid output for multiple consecutive bars after recovery at bar period.

When the SMA seed phase encounters a NaN window (due to bar-0 na from ta.tr), the
seed itself becomes NaN. On the next bar EMA restarts from the current source
(Pine: na(ema[1]) ? src), so recovery value = ta.tr at bar period = syntheticTRValue.
The recursive formula alpha*src + (1-alpha)*prev remains valid for all subsequent bars.
*/
func TestTrSource_EmaValidityBeyondSeed(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	for _, tc := range []struct {
		name   string
		period int
	}{
		{"period_1", 1},
		{"period_3", 3},
		{"period_5", 5},
		{"period_14", 14},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assertValidityBeyondSeed(t, buildTrEmaPineScript(tc.period), "EMA", tc.period)
		})
	}
}

/*
TestTrSource_RmaWarmUp verifies that ta.rma(ta.tr, period) is NaN for bars
0..period-1 and recovers to syntheticTRValue at bar period.

Pine's ta.rma pseudocode:

	sum := na(sum[1]) ? ta.sma(src, length) : alpha * src + (1 - alpha) * nz(sum[1])

When sum[1] is NaN, Pine re-seeds with ta.sma(src, length) — not the recursive
formula. On bar period-1 the SMA window spans [0..period-1] which contains bar 0
where ta.tr = NaN (no previous close), so the seed is NaN. On bar period the
SMA window slides to [1..period] — all valid ta.tr bars — so re-seeding succeeds
and the output becomes syntheticTRValue. This is the same warm-up boundary as
ta.sma(ta.tr, period) and ta.ema(ta.tr, period).
*/
func TestTrSource_RmaWarmUp(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	for _, tc := range []struct {
		name   string
		period int
	}{
		{"period_1", 1},
		{"period_3", 3},
		{"period_5", 5},
		{"period_14", 14},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assertWarmUp(t, buildTrRmaPineScript(tc.period), "RMA", tc.period, syntheticTRValue)
		})
	}
}

/*
TestTrSource_RmaValidityBeyondSeed verifies that ta.rma(ta.tr, period) sustains
valid output for multiple consecutive bars after the SMA re-seed at bar period.

Once the RMA recovers via re-seeding, subsequent bars use the recursive formula
alpha*src + (1-alpha)*prev, which remains valid. On synthetic data where ta.tr
is constant syntheticTRValue for all bars ≥ 1, every post-seed bar equals
syntheticTRValue exactly. With period=1 alpha=1.0 so result = src; the property
still holds.
*/
func TestTrSource_RmaValidityBeyondSeed(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	for _, tc := range []struct {
		name   string
		period int
	}{
		{"period_1", 1},
		{"period_3", 3},
		{"period_5", 5},
		{"period_14", 14},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assertValidityBeyondSeed(t, buildTrRmaPineScript(tc.period), "RMA", tc.period)
		})
	}
}

/*
TestTrSource_AtrUnaffected verifies that ta.atr(period) first produces a valid
result at bar period-1, one bar earlier than ta.sma(ta.tr, period).

ta.atr internally uses ta.tr(true) (handle_na=true), which returns high-low on
bar 0. The seeding SMA therefore covers a fully valid window [0..period-1], so
the first valid output appears at bar period-1 rather than bar period.
*/
func TestTrSource_AtrUnaffected(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	for _, tc := range []struct {
		name   string
		period int
	}{
		{"period_1", 1},
		{"period_3", 3},
		{"period_5", 5},
		{"period_14", 14},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assertWarmUp(t, buildTrAtrPineScript(tc.period), "ATR", tc.period-1, syntheticTRValue)
		})
	}
}

/*
TestTrSource_WarmUpBoundaryContrast runs ta.sma(ta.tr, N) and ta.atr(N) through
the same pipeline pass and asserts the single-bar offset between their warm-up
boundaries — the direct observable consequence of the handle_na split.
*/
func TestTrSource_WarmUpBoundaryContrast(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	const period = 5
	result := runStrategy(t, buildTrSmaAtrContrastPineScript(period), period+5)

	smaVals := extractValues(requireIndicator(t, result, "SMA"))
	atrVals := extractValues(requireIndicator(t, result, "ATR"))

	if period-1 < len(smaVals) && !math.IsNaN(smaVals[period-1]) {
		t.Errorf("sma bar %d: expected NaN (bar-0 na still in window), got %.6f", period-1, smaVals[period-1])
	}
	assertBarValue(t, "sma", smaVals, period, syntheticTRValue)
	assertBarValue(t, "atr", atrVals, period-1, syntheticTRValue)
}

/*
TestTrBareIdentifier_SmaWarmUp verifies that ta.sma(tr, period) — using the bare
'tr' identifier — produces the same NaN warm-up boundary as ta.sma(ta.tr, period).

Both the member expression 'ta.tr' and the bare identifier 'tr' resolve to
BuiltinTrueRangeAccessor via isTrBuiltin detection in TAArgumentExtractor.
NeedsNaNCheck=true is set for both paths, so the SMA window must wait for a fully
valid window starting at bar 1.
*/
func TestTrBareIdentifier_SmaWarmUp(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	for _, tc := range []struct {
		name   string
		period int
	}{
		{"period_1", 1},
		{"period_3", 3},
		{"period_5", 5},
		{"period_14", 14},
	} {
		t.Run(tc.name, func(t *testing.T) {
			script := fmt.Sprintf("//@version=5\nindicator(\"TR Bare SMA Test\")\nplot(ta.sma(tr, %d), \"SMA\")\n", tc.period)
			assertWarmUp(t, script, "SMA", tc.period, syntheticTRValue)
		})
	}
}

/*
TestTrBareIdentifier_EmaWarmUp verifies that ta.ema(tr, period) produces the same
NaN warm-up boundary as ta.ema(ta.tr, period).

The bare identifier 'tr' must be recognised by isTrBuiltin and routed to
BuiltinTrueRangeAccessor identically to the member-expression form.
*/
func TestTrBareIdentifier_EmaWarmUp(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	for _, tc := range []struct {
		name   string
		period int
	}{
		{"period_1", 1},
		{"period_3", 3},
		{"period_5", 5},
		{"period_14", 14},
	} {
		t.Run(tc.name, func(t *testing.T) {
			script := fmt.Sprintf("//@version=5\nindicator(\"TR Bare EMA Test\")\nplot(ta.ema(tr, %d), \"EMA\")\n", tc.period)
			assertWarmUp(t, script, "EMA", tc.period, syntheticTRValue)
		})
	}
}

/*
TestTrBareIdentifier_RmaWarmUp verifies that ta.rma(tr, period) produces the same
NaN warm-up boundary as ta.rma(ta.tr, period).

The bare identifier 'tr' must be recognised by isTrBuiltin and routed to
BuiltinTrueRangeAccessor identically to the member-expression form.
*/
func TestTrBareIdentifier_RmaWarmUp(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	for _, tc := range []struct {
		name   string
		period int
	}{
		{"period_1", 1},
		{"period_3", 3},
		{"period_5", 5},
		{"period_14", 14},
	} {
		t.Run(tc.name, func(t *testing.T) {
			script := fmt.Sprintf("//@version=5\nindicator(\"TR Bare RMA Test\")\nplot(ta.rma(tr, %d), \"RMA\")\n", tc.period)
			assertWarmUp(t, script, "RMA", tc.period, syntheticTRValue)
		})
	}
}

func runStrategy(t *testing.T, strategy string, barCount int) TestResult {
	t.Helper()

	testDir := t.TempDir()
	strategyPath := filepath.Join(testDir, "strategy.pine")
	if err := os.WriteFile(strategyPath, []byte(strategy), 0644); err != nil {
		t.Fatal(err)
	}
	dataPath := filepath.Join(testDir, "DATA_1D.json")
	if err := os.WriteFile(dataPath, []byte(generateTestOHLCV(barCount, 86400)), 0644); err != nil {
		t.Fatal(err)
	}

	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(filepath.Dir(cwd))
	return compileAndRun(t, strategyPath, dataPath, testDir, projectRoot, "DATA", testDir)
}

func runIndicator(t *testing.T, strategy, indicatorName string, barCount int) []float64 {
	t.Helper()
	return extractValues(requireIndicator(t, runStrategy(t, strategy, barCount), indicatorName))
}

// assertWarmUp verifies bars 0..firstValidBar-1 produce NaN and bar firstValidBar
// produces expectedValue.
func assertWarmUp(t *testing.T, strategy, indicatorName string, firstValidBar int, expectedValue float64) {
	t.Helper()
	vals := runIndicator(t, strategy, indicatorName, firstValidBar+5)

	for bar := 0; bar < firstValidBar && bar < len(vals); bar++ {
		if !math.IsNaN(vals[bar]) {
			t.Errorf("bar %d: expected NaN (warm-up), got %.6f (firstValidBar=%d)", bar, vals[bar], firstValidBar)
		}
	}
	assertBarValue(t, indicatorName, vals, firstValidBar, expectedValue)
}

// assertValidityBeyondSeed verifies bars 0..period-1 produce NaN and the four bars
// starting at period all equal syntheticTRValue.
func assertValidityBeyondSeed(t *testing.T, strategy, indicatorName string, period int) {
	t.Helper()
	vals := runIndicator(t, strategy, indicatorName, period+5)

	for bar := 0; bar < period && bar < len(vals); bar++ {
		if !math.IsNaN(vals[bar]) {
			t.Errorf("bar %d: expected NaN (warm-up), got %.6f", bar, vals[bar])
		}
	}
	for bar := period; bar < period+4 && bar < len(vals); bar++ {
		assertBarValue(t, indicatorName, vals, bar, syntheticTRValue)
	}
}

// requireIndicator retrieves named indicator data or fails the test immediately.
func requireIndicator(t *testing.T, result TestResult, name string) []map[string]interface{} {
	t.Helper()
	ind, ok := result.Indicators[name]
	if !ok {
		t.Fatalf("indicator %q absent from output", name)
	}
	return ind.Data
}

// assertBarValue fails if the value at barIndex is NaN or differs from expected by more than 1e-9.
func assertBarValue(t *testing.T, label string, vals []float64, barIndex int, expected float64) {
	t.Helper()
	if barIndex >= len(vals) {
		t.Errorf("%s bar %d: out of range (len=%d)", label, barIndex, len(vals))
		return
	}
	if math.IsNaN(vals[barIndex]) {
		t.Errorf("%s bar %d: expected %.1f, got NaN", label, barIndex, expected)
		return
	}
	if math.Abs(vals[barIndex]-expected) > 1e-9 {
		t.Errorf("%s bar %d: expected %.1f, got %.6f", label, barIndex, expected, vals[barIndex])
	}
}

func buildTrSmaPineScript(period int) string {
	return fmt.Sprintf("//@version=5\nindicator(\"TR SMA Test\")\nplot(ta.sma(ta.tr, %d), \"SMA\")\n", period)
}

func buildTrEmaPineScript(period int) string {
	return fmt.Sprintf("//@version=5\nindicator(\"TR EMA Test\")\nplot(ta.ema(ta.tr, %d), \"EMA\")\n", period)
}

func buildTrRmaPineScript(period int) string {
	return fmt.Sprintf("//@version=5\nindicator(\"TR RMA Test\")\nplot(ta.rma(ta.tr, %d), \"RMA\")\n", period)
}

func buildTrAtrPineScript(period int) string {
	return fmt.Sprintf("//@version=5\nindicator(\"ATR Test\")\nplot(ta.atr(%d), \"ATR\")\n", period)
}

func buildTrSmaAtrContrastPineScript(period int) string {
	return fmt.Sprintf("//@version=5\nindicator(\"TR Contrast\")\nplot(ta.sma(ta.tr, %d), \"SMA\")\nplot(ta.atr(%d), \"ATR\")\n", period, period)
}
