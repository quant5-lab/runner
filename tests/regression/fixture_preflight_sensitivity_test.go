package regression

import (
	"os"
	"path/filepath"
	"testing"

	goldenutil "github.com/quant5-lab/runner/tests/golden/testutil"
)

// flatBarPerturbEpsilon is small enough to be below the tick size of any
// wired instrument (SBERP ~233 RUB, BTC ~40 000 USD, AAPL ~27 USD), so
// TA indicators that use price differences are not meaningfully shifted.
const flatBarPerturbEpsilon = 0.001

// shadowFixtureDirForProbe mirrors realFixtureDir via symlinks for every file
// except targetFilename, which is replaced by perturbedContent. This lets
// StrategyRunner set -datadir to the shadow dir so that secondary security()
// fixtures remain accessible during the perturbed run.
func shadowFixtureDirForProbe(t *testing.T, realFixtureDir, targetFilename string, perturbedContent []byte) string {
	t.Helper()
	shadowDir := t.TempDir()

	entries, err := os.ReadDir(realFixtureDir)
	if err != nil {
		t.Fatalf("read fixture dir %s: %v", realFixtureDir, err)
	}
	for _, e := range entries {
		if e.IsDir() || e.Name() == targetFilename {
			continue
		}
		src := filepath.Join(realFixtureDir, e.Name())
		dst := filepath.Join(shadowDir, e.Name())
		if err := os.Symlink(src, dst); err != nil {
			t.Fatalf("symlink %s: %v", e.Name(), err)
		}
	}

	pertPath := filepath.Join(shadowDir, targetFilename)
	if err := os.WriteFile(pertPath, perturbedContent, 0644); err != nil {
		t.Fatalf("write perturbed fixture: %v", err)
	}
	return shadowDir
}

// Open trades are excluded because boundary effects at the fixture end can
// shift one trade between open/closed without any change in strategy logic.
func closedTradeDirections(result *goldenutil.StrategyResult) []string {
	dirs := make([]string, len(result.Trades))
	for i, tr := range result.Trades {
		dirs[i] = tr.Direction
	}
	return dirs
}

func closedTradeDirectionsEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// isFlatBarSensitive proves empirically whether a strategy toggles direction
// based on bar equality. A difference in closed-trade count or direction between
// the real run and the epsilon-perturbed run means the strategy is blocked by
// flat bars in its fixture.
//
// Entry timing shifts from epsilon-scale price changes are intentionally ignored:
// only direction or count changes indicate true direction-toggle sensitivity
// (the class that blocks zigzag on SBERP-1h).
//
// Callers must pre-screen with flatBarIndices; zero flat bars make this probe
// vacuously false and waste two full codegen+build+run cycles.
func isFlatBarSensitive(t *testing.T, root string, tc tvAlignmentCase, fixture preflightFixture) bool {
	t.Helper()

	perturbedJSON, err := fixture.perturbedFixtureJSON(flatBarPerturbEpsilon)
	if err != nil {
		t.Fatalf("%s: build perturbed fixture: %v", tc.Name, err)
	}

	realFixtureDir := filepath.Join(root, "tests", "golden", "fixtures", "data")
	shadowDir := shadowFixtureDirForProbe(t, realFixtureDir, tc.Data, perturbedJSON)

	stratPath := filepath.Join(root, "strategies", tc.Strategy)
	realDataPath := filepath.Join(realFixtureDir, tc.Data)
	perturbedDataPath := filepath.Join(shadowDir, tc.Data)

	runner := goldenutil.NewStrategyRunner(t)
	realResult := runner.Execute(t, stratPath, realDataPath, tc.Symbol, tc.Timeframe)
	pertResult := runner.Execute(t, stratPath, perturbedDataPath, tc.Symbol, tc.Timeframe)

	return !closedTradeDirectionsEqual(
		closedTradeDirections(realResult),
		closedTradeDirections(pertResult),
	)
}

func makeTrade(direction string) goldenutil.Trade {
	return goldenutil.Trade{Direction: direction}
}

// TestClosedTradeDirections verifies that only closed trades contribute to the
// direction sequence and that open trades are silently excluded at all positions.
func TestClosedTradeDirections(t *testing.T) {
	cases := []struct {
		name       string
		trades     []goldenutil.Trade
		openTrades []goldenutil.Trade
		want       []string
	}{
		{
			name: "empty_result",
			want: []string{},
		},
		{
			name:       "only_open_trades_excluded",
			openTrades: []goldenutil.Trade{makeTrade("long"), makeTrade("short")},
			want:       []string{},
		},
		{
			name:   "only_closed_trades_returned",
			trades: []goldenutil.Trade{makeTrade("long"), makeTrade("short"), makeTrade("long")},
			want:   []string{"long", "short", "long"},
		},
		{
			name:       "open_trades_excluded_closed_preserved",
			trades:     []goldenutil.Trade{makeTrade("long"), makeTrade("short")},
			openTrades: []goldenutil.Trade{makeTrade("long")},
			want:       []string{"long", "short"},
		},
		{
			name:   "order_preserved",
			trades: []goldenutil.Trade{makeTrade("short"), makeTrade("long"), makeTrade("short")},
			want:   []string{"short", "long", "short"},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			result := &goldenutil.StrategyResult{Trades: tc.trades, OpenTrades: tc.openTrades}
			got := closedTradeDirections(result)
			if len(got) != len(tc.want) {
				t.Fatalf("len = %d, want %d: got %v", len(got), len(tc.want), got)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Errorf("[%d] = %q, want %q", i, got[i], tc.want[i])
				}
			}
		})
	}
}

// TestClosedTradeDirectionsEqual covers nil/empty symmetry, length mismatch, and element-level inequality.
func TestClosedTradeDirectionsEqual(t *testing.T) {
	cases := []struct {
		name  string
		a, b  []string
		equal bool
	}{
		{name: "both_nil", a: nil, b: nil, equal: true},
		{name: "both_empty", a: []string{}, b: []string{}, equal: true},
		{name: "nil_and_empty", a: nil, b: []string{}, equal: true},
		{name: "empty_and_nil", a: []string{}, b: nil, equal: true},
		{name: "single_same", a: []string{"long"}, b: []string{"long"}, equal: true},
		{name: "single_different", a: []string{"long"}, b: []string{"short"}, equal: false},
		{name: "a_longer", a: []string{"long", "short"}, b: []string{"long"}, equal: false},
		{name: "b_longer", a: []string{"long"}, b: []string{"long", "short"}, equal: false},
		{name: "multi_same", a: []string{"long", "short", "long"}, b: []string{"long", "short", "long"}, equal: true},
		{name: "first_differs", a: []string{"long", "short"}, b: []string{"short", "short"}, equal: false},
		{name: "middle_differs", a: []string{"long", "long", "short"}, b: []string{"long", "short", "short"}, equal: false},
		{name: "last_differs", a: []string{"long", "short"}, b: []string{"long", "long"}, equal: false},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := closedTradeDirectionsEqual(tc.a, tc.b)
			if got != tc.equal {
				t.Errorf("closedTradeDirectionsEqual(%v, %v) = %v, want %v", tc.a, tc.b, got, tc.equal)
			}
		})
	}
}

// maxFlatBarSensitivityProbeCount is the ceiling on the number of wired alignment
// cases whose primary fixture contains at least one flat O=H=L=C bar. Each such
// case triggers two codegen+build+run cycles (real run and perturbed run) inside
// FlatBarSensitivity. Exceeding this ceiling requires operator approval because
// each additional flat-bar-bearing case adds proportional wall time to the probe
// suite and may push CI past its timeout budget.
const maxFlatBarSensitivityProbeCount = 15

func makeFixtureWithFlatBarFraction(t *testing.T, dir, name string, totalBars, flatBars int) {
	t.Helper()
	bars := make([]preflightBar, totalBars)
	for i := range bars {
		bars[i] = preflightBar{Time: int64(1000 + i*60), Open: 100, High: 102, Low: 98, Close: 101, Volume: 500}
	}
	for i := 0; i < flatBars && i < totalBars; i++ {
		bars[i].High = bars[i].Open
		bars[i].Low = bars[i].Open
		bars[i].Close = bars[i].Open
	}
	writePreflightFixtureToDisk(t, dir, name, bars)
}

// Zero-bar fixtures never exceed the threshold regardless of flatCount.
func flatFractionExceedsThreshold(flatCount, totalBars int) bool {
	if totalBars == 0 {
		return false
	}
	return float64(flatCount)/float64(totalBars) > maxPrimaryFlatBarFraction
}

// Cases whose fixture cannot be loaded are excluded from the count; they are
// reported separately by TestFixtureQualityPreflight/MinimumBarCount and do not
// trigger probes.
func countFlatBarBearingCases(cases []tvAlignmentCase, fixtureDir string) int {
	n := 0
	for _, tc := range cases {
		f, err := loadPreflightFixture(filepath.Join(fixtureDir, tc.Data))
		if err != nil {
			continue
		}
		flatIdx := flatBarIndices(f.Bars)
		if flatFractionExceedsThreshold(len(flatIdx), len(f.Bars)) {
			n++
		}
	}
	return n
}

// Callers must invoke this before launching the parallel probe subtests so that
// a budget breach is reported immediately rather than after they complete.
func assertFlatBarSensitivityProbeBudget(t *testing.T, cases []tvAlignmentCase, fixtureDir string) {
	t.Helper()
	n := countFlatBarBearingCases(cases, fixtureDir)
	if n > maxFlatBarSensitivityProbeCount {
		t.Errorf(
			"flat-bar probe would execute %d codegen+build cycles (cap %d × 2 = %d)"+
				" — wire alignment cases with zero-flat-bar fixtures, or raise"+
				" maxFlatBarSensitivityProbeCount with operator approval",
			n*2, maxFlatBarSensitivityProbeCount, maxFlatBarSensitivityProbeCount*2,
		)
	}
}

// TestCountFlatBarBearingCases verifies the count across nil/empty input, all-clean, all-flat, mixed,
// unreadable-fixture, and below-threshold-flat-fraction cases.
func TestCountFlatBarBearingCases(t *testing.T) {
	dir := t.TempDir()

	flatBar := preflightBar{Time: 1000, Open: 100, High: 100, Low: 100, Close: 100, Volume: 500}
	normalBar := preflightBar{Time: 2000, Open: 100, High: 102, Low: 98, Close: 101, Volume: 500}

	writePreflightFixtureToDisk(t, dir, "flat.json", []preflightBar{flatBar})
	writePreflightFixtureToDisk(t, dir, "normal.json", []preflightBar{normalBar})
	writePreflightFixtureToDisk(t, dir, "mixed.json", []preflightBar{normalBar, flatBar})
	makeFixtureWithFlatBarFraction(t, dir, "below_threshold.json", 25, 1)
	makeFixtureWithFlatBarFraction(t, dir, "at_threshold.json", 100, 5)
	makeFixtureWithFlatBarFraction(t, dir, "above_threshold.json", 100, 6)

	caseFlat := makeCase("s.pine", "A", "flat.json")
	caseNormal := makeCase("s.pine", "B", "normal.json")
	caseMixed := makeCase("s.pine", "C", "mixed.json")
	caseMissing := makeCase("s.pine", "D", "nonexistent.json")
	caseBelowThreshold := makeCase("s.pine", "E", "below_threshold.json")
	caseAtThreshold := makeCase("s.pine", "F", "at_threshold.json")
	caseAboveThreshold := makeCase("s.pine", "G", "above_threshold.json")

	cases := []struct {
		name  string
		input []tvAlignmentCase
		want  int
	}{
		{name: "nil_input_returns_zero", input: nil, want: 0},
		{name: "empty_input_returns_zero", input: []tvAlignmentCase{}, want: 0},
		{name: "all_flat_free_returns_zero", input: []tvAlignmentCase{caseNormal}, want: 0},
		{name: "single_flat_bearing_returns_one", input: []tvAlignmentCase{caseFlat}, want: 1},
		{name: "mixed_fixture_counts_as_bearing", input: []tvAlignmentCase{caseMixed}, want: 1},
		{name: "two_flat_bearing_returns_two", input: []tvAlignmentCase{caseFlat, caseMixed}, want: 2},
		{name: "mixed_cases_returns_correct_count", input: []tvAlignmentCase{caseFlat, caseNormal, caseMixed}, want: 2},
		{name: "load_error_excluded_not_fatal", input: []tvAlignmentCase{caseFlat, caseMissing, caseNormal}, want: 1},
		// Two alignment cases reference the same flat-bar fixture file;
		// counting is per-case, not per distinct file.
		{name: "same_fixture_two_cases_both_counted", input: []tvAlignmentCase{caseFlat, caseFlat}, want: 2},
		{name: "all_unreadable_fixtures_returns_zero", input: []tvAlignmentCase{caseMissing, caseMissing}, want: 0},
		{name: "below_threshold_flat_fraction_not_counted", input: []tvAlignmentCase{caseBelowThreshold}, want: 0},
		{name: "at_threshold_fraction_not_counted", input: []tvAlignmentCase{caseAtThreshold}, want: 0},
		{name: "just_above_threshold_counted", input: []tvAlignmentCase{caseAboveThreshold}, want: 1},
		{name: "threshold_boundary_mix", input: []tvAlignmentCase{caseBelowThreshold, caseAtThreshold, caseAboveThreshold}, want: 1},
		{name: "above_threshold_with_below_mix_counts_only_above", input: []tvAlignmentCase{caseFlat, caseBelowThreshold}, want: 1},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := countFlatBarBearingCases(tc.input, dir)
			if got != tc.want {
				t.Errorf("countFlatBarBearingCases = %d, want %d", got, tc.want)
			}
		})
	}
}

// TestFlatFractionExceedsThreshold exercises the threshold gate across the full
// boundary surface: zero-bar fixture, zero flat bars, at-threshold, below, above,
// all-flat, and single-element cases. Tests are independent of any file I/O.
func TestFlatFractionExceedsThreshold(t *testing.T) {
	cases := []struct {
		name       string
		flat       int
		total      int
		wantExceed bool
	}{
		{name: "zero_total_bars_never_exceeds", flat: 0, total: 0, wantExceed: false},
		{name: "nonzero_flat_zero_total_never_exceeds", flat: 1, total: 0, wantExceed: false},
		{name: "zero_flat_bars_never_exceeds", flat: 0, total: 100, wantExceed: false},
		{name: "one_percent_below_threshold", flat: 1, total: 100, wantExceed: false},
		{name: "four_percent_below_threshold", flat: 4, total: 100, wantExceed: false},
		{name: "at_threshold_not_exceeded", flat: 5, total: 100, wantExceed: false},
		{name: "one_step_above_threshold", flat: 6, total: 100, wantExceed: true},
		{name: "fifty_percent_well_above_threshold", flat: 50, total: 100, wantExceed: true},
		{name: "all_flat_bars_exceeds", flat: 100, total: 100, wantExceed: true},
		{name: "single_bar_flat_exceeds", flat: 1, total: 1, wantExceed: true},
		{name: "small_count_below_threshold", flat: 1, total: 25, wantExceed: false},
		{name: "small_count_above_threshold", flat: 2, total: 25, wantExceed: true},
		{name: "large_fixture_at_threshold", flat: 500, total: 10000, wantExceed: false},
		{name: "large_fixture_above_threshold", flat: 501, total: 10000, wantExceed: true},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := flatFractionExceedsThreshold(tc.flat, tc.total)
			if got != tc.wantExceed {
				t.Errorf("flatFractionExceedsThreshold(%d, %d) = %v, want %v (threshold %.0f%%)",
					tc.flat, tc.total, got, tc.wantExceed, maxPrimaryFlatBarFraction*100)
			}
		})
	}
}
