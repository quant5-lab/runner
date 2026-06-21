package regression

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	tvref "github.com/quant5-lab/runner/tests/regression/tv_reference"
)

// 50× clears all real corporate actions (splits ≤10:1, denomination quirks) while
// remaining well below the minimum synthetic-placeholder mismatch observed in practice
// (moon BTCUSDT fixture≈100 vs real≈40 000 → 400×).
const priceScaleIncompatibilityFactor = 50.0

// medianFloat64 returns the statistical median of vals; returns 0 for nil or empty
// input and never modifies the input slice.
func medianFloat64(vals []float64) float64 {
	if len(vals) == 0 {
		return 0
	}
	sorted := make([]float64, len(vals))
	copy(sorted, vals)
	sort.Float64s(sorted)
	n := len(sorted)
	if n%2 == 1 {
		return sorted[n/2]
	}
	return (sorted[n/2-1] + sorted[n/2]) / 2
}

func fixtureMedianClose(bars []preflightBar) float64 {
	closes := make([]float64, len(bars))
	for i, b := range bars {
		closes[i] = b.Close
	}
	return medianFloat64(closes)
}

// referenceInWindowPrices returns entry prices of TV reference trades whose entry
// falls inside the fixture's own time window. Returns nil (no error) when no trades
// are in-window — callers treat an empty slice as "cannot judge scale".
func referenceInWindowPrices(fixturePath, csvPath string, tz tvref.TVTimezone) ([]float64, error) {
	start, end, err := tvref.FixtureWindow(fixturePath)
	if err != nil {
		return nil, fmt.Errorf("fixture window %s: %w", fixturePath, err)
	}
	all, err := tvref.LoadTrades(csvPath, tz)
	if err != nil {
		return nil, fmt.Errorf("load reference CSV %s: %w", csvPath, err)
	}
	inWindow := tvref.FilterByEntryWindow(all, start, end)
	prices := make([]float64, len(inWindow))
	for i, t := range inWindow {
		prices[i] = t.EntryPrice
	}
	return prices, nil
}

// priceScaleFactor returns max(a/b, b/a). Returns 1 when both values are equal
// (including both-zero). Returns math.Inf(1) when exactly one is zero.
func priceScaleFactor(fixtureMedian, referenceMedian float64) float64 {
	if fixtureMedian == referenceMedian {
		return 1
	}
	if fixtureMedian == 0 || referenceMedian == 0 {
		return math.Inf(1)
	}
	r := fixtureMedian / referenceMedian
	if r < 1 {
		r = 1 / r
	}
	return r
}

// priceScaleConsistent reports whether fixture median close and reference median
// entry price are within priceScaleIncompatibilityFactor of each other.
// Returns (true, 0, nil) when no in-window trades exist — an empty reference
// window cannot characterise the expected price scale.
func priceScaleConsistent(fixturePath, csvPath string, tz tvref.TVTimezone) (consistent bool, factor float64, err error) {
	refPrices, err := referenceInWindowPrices(fixturePath, csvPath, tz)
	if err != nil {
		return false, 0, err
	}
	if len(refPrices) == 0 {
		return true, 0, nil
	}
	f, err := loadPreflightFixture(fixturePath)
	if err != nil {
		return false, 0, err
	}
	factor = priceScaleFactor(fixtureMedianClose(f.Bars), medianFloat64(refPrices))
	return factor < priceScaleIncompatibilityFactor, factor, nil
}

// CrossSymbolContentDistinct is the sole verdict for multi-symbol families;
// this predicate gates PriceScaleConsistent off those cases to honour the
// one-verdict-per-family invariant.
func isCrossSymbolCase(tc tvAlignmentCase, allCases []tvAlignmentCase) bool {
	for _, f := range groupCrossSymbolFamilies(allCases) {
		for _, c := range f.Cases {
			if c.Name == tc.Name {
				return true
			}
		}
	}
	return false
}

// writeTempPricedTradeCSV writes a single-trade TV CSV at entryPrice.
// Exit is pinned to 2099-01-01 so it never falls inside any test fixture window.
func writeTempPricedTradeCSV(t *testing.T, dir, name, entryTime string, entryPrice float64) string {
	t.Helper()
	return writeTempMultiPricedTradeCSV(t, dir, name, entryTime, []float64{entryPrice})
}

func writeTempMultiPricedTradeCSV(t *testing.T, dir, name, entryTime string, prices []float64) string {
	t.Helper()
	var sb strings.Builder
	sb.WriteString("Trade number,Type,Date and time,Price USD\n")
	for i, p := range prices {
		fmt.Fprintf(&sb, "%d,Entry long,%s,%.2f\n", i+1, entryTime, p)
		fmt.Fprintf(&sb, "%d,Exit long,2099-01-01 00:00,%.2f\n", i+1, p)
	}
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(sb.String()), 0644); err != nil {
		t.Fatalf("write multi-priced trade CSV %s: %v", name, err)
	}
	return path
}

// writeTempMultiBarFixture writes a fixture spanning 2024-01-01 00:00 → 2024-01-02 00:00 UTC
// with len(closes) bars; test time constants across this file rely on that fixed window.
func writeTempMultiBarFixture(t *testing.T, dir, name string, closes []float64) string {
	t.Helper()
	const (
		ts1 int64 = 1704067200 // 2024-01-01 00:00 UTC
		ts2 int64 = 1704153600 // 2024-01-02 00:00 UTC
	)
	n := len(closes)
	bars := make([]preflightBar, n)
	for i, c := range closes {
		var ts int64
		switch {
		case n == 1:
			ts = ts1
		default:
			ts = ts1 + int64(i)*(ts2-ts1)/int64(n-1)
		}
		bars[i] = preflightBar{
			Time: ts, Open: c, High: c * 1.01, Low: c * 0.99, Close: c, Volume: 1000,
		}
	}
	writePreflightFixtureToDisk(t, dir, name, bars)
	return filepath.Join(dir, name)
}

// TestMedianFloat64 covers nil/empty, odd/even counts, sorted and unsorted input,
// signed values, fractional values, and the non-mutation invariant.
func TestMedianFloat64(t *testing.T) {
	cases := []struct {
		name string
		vals []float64
		want float64
	}{
		{name: "nil_returns_zero", vals: nil, want: 0},
		{name: "empty_returns_zero", vals: []float64{}, want: 0},
		{name: "single_element", vals: []float64{42}, want: 42},
		{name: "two_elements_returns_average", vals: []float64{10, 20}, want: 15},
		{name: "odd_count_sorted_returns_middle", vals: []float64{1, 3, 5}, want: 3},
		{name: "odd_count_unsorted_returns_middle", vals: []float64{5, 1, 3}, want: 3},
		{name: "even_count_sorted_returns_average_of_middles", vals: []float64{2, 4, 6, 8}, want: 5},
		{name: "even_count_unsorted_returns_average_of_middles", vals: []float64{8, 2, 6, 4}, want: 5},
		{name: "all_same_value_returns_that_value", vals: []float64{7, 7, 7}, want: 7},
		{name: "all_negative_returns_middle", vals: []float64{-4, -2, -6}, want: -4},
		{name: "mixed_sign_symmetric_returns_zero", vals: []float64{-1, 0, 1}, want: 0},
		{name: "large_price_scale", vals: []float64{40000, 41000, 39000}, want: 40000},
		{name: "fractional_values", vals: []float64{1.5, 2.5, 3.5}, want: 2.5},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := medianFloat64(tc.vals)
			if got != tc.want {
				t.Errorf("medianFloat64 = %v, want %v", got, tc.want)
			}
		})
	}

	t.Run("does_not_mutate_input", func(t *testing.T) {
		orig := []float64{5, 1, 3}
		snapshot := []float64{5, 1, 3}
		medianFloat64(orig)
		for i := range snapshot {
			if orig[i] != snapshot[i] {
				t.Errorf("input slice mutated at index %d: got %v want %v", i, orig[i], snapshot[i])
			}
		}
	})
}

// TestFixtureMedianClose verifies Close-field selection and median correctness
// across all bar-count configurations including even-count averaging.
func TestFixtureMedianClose(t *testing.T) {
	cases := []struct {
		name string
		bars []preflightBar
		want float64
	}{
		{name: "nil_bars_returns_zero", bars: nil, want: 0},
		{name: "empty_bars_returns_zero", bars: []preflightBar{}, want: 0},
		{
			name: "single_bar_returns_its_close",
			bars: []preflightBar{{Open: 99, High: 101, Low: 98, Close: 100}},
			want: 100,
		},
		{
			name: "close_selected_not_open_high_or_low",
			bars: []preflightBar{{Open: 200, High: 300, Low: 50, Close: 100}},
			want: 100,
		},
		{
			name: "two_bars_returns_average_of_closes",
			bars: []preflightBar{{Close: 100}, {Close: 200}},
			want: 150,
		},
		{
			name: "three_bars_unordered_returns_sorted_middle",
			bars: []preflightBar{{Close: 300}, {Close: 100}, {Close: 200}},
			want: 200,
		},
		{
			name: "large_price_scale",
			bars: []preflightBar{{Close: 39000}, {Close: 40000}, {Close: 41000}},
			want: 40000,
		},
		{
			name: "four_bars_even_count",
			bars: []preflightBar{{Close: 100}, {Close: 200}, {Close: 300}, {Close: 400}},
			want: 250,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := fixtureMedianClose(tc.bars)
			if got != tc.want {
				t.Errorf("fixtureMedianClose = %v, want %v", got, tc.want)
			}
		})
	}
}

// TestPriceScaleFactor covers identity, symmetry, zero-argument edges, integer and
// non-integer ratios. Cases named by computed ratio, not by application threshold —
// this is a pure-math function with no threshold knowledge.
func TestPriceScaleFactor(t *testing.T) {
	cases := []struct {
		name         string
		fixture, ref float64
		want         float64
		wantInf      bool
	}{
		{name: "equal_values_returns_one", fixture: 100, ref: 100, want: 1},
		{name: "both_zero_returns_one", fixture: 0, ref: 0, want: 1},
		{name: "fixture_double_ref", fixture: 200, ref: 100, want: 2},
		{name: "ref_double_fixture_symmetric", fixture: 100, ref: 200, want: 2},
		{name: "fixture_half_ref", fixture: 50, ref: 100, want: 2},
		{name: "large_ratio", fixture: 40000, ref: 100, want: 400},
		{name: "large_ratio_inverse_symmetric", fixture: 100, ref: 40000, want: 400},
		{name: "non_integer_ratio", fixture: 150, ref: 100, want: 1.5},
		{name: "non_integer_ratio_inverse_symmetric", fixture: 100, ref: 150, want: 1.5},
		{name: "ratio_49", fixture: 49, ref: 1, want: 49},
		{name: "ratio_50", fixture: 50, ref: 1, want: 50},
		{name: "ratio_51", fixture: 51, ref: 1, want: 51},
		{name: "fixture_zero_ref_nonzero_returns_inf", fixture: 0, ref: 100, wantInf: true},
		{name: "ref_zero_fixture_nonzero_returns_inf", fixture: 100, ref: 0, wantInf: true},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := priceScaleFactor(tc.fixture, tc.ref)
			if tc.wantInf {
				if !math.IsInf(got, 1) {
					t.Errorf("priceScaleFactor(%v, %v) = %v, want +Inf", tc.fixture, tc.ref, got)
				}
				return
			}
			if got != tc.want {
				t.Errorf("priceScaleFactor(%v, %v) = %v, want %v", tc.fixture, tc.ref, got, tc.want)
			}
		})
	}
}

// TestReferenceInWindowPrices verifies window-boundary inclusivity, directional
// exclusions, multi-trade price aggregation, entry-vs-exit price selection, and
// error propagation from both the fixture and CSV paths.
func TestReferenceInWindowPrices(t *testing.T) {
	const (
		startBoundary = "2024-01-01 00:00" // equals ts1 inside writeTempMultiBarFixture
		endBoundary   = "2024-01-02 00:00" // equals ts2 inside writeTempMultiBarFixture
		inWindowTime  = "2024-01-01 12:00"
		beforeWindow  = "2023-12-31 12:00"
		afterWindow   = "2024-01-03 00:00"
	)

	t.Run("empty_csv_returns_empty_prices", func(t *testing.T) {
		dir := t.TempDir()
		fix := writeTempMultiBarFixture(t, dir, "fix.json", []float64{100, 100})
		csv := writeTempTradeCSV(t, dir, "trades.csv", []barcountTestTrade{})
		prices, err := referenceInWindowPrices(fix, csv, tvref.TVTimezoneUTC)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(prices) != 0 {
			t.Errorf("empty CSV produced %d prices, want 0", len(prices))
		}
	})

	t.Run("trade_before_window_excluded", func(t *testing.T) {
		dir := t.TempDir()
		fix := writeTempMultiBarFixture(t, dir, "fix.json", []float64{100, 100})
		csv := writeTempPricedTradeCSV(t, dir, "trades.csv", beforeWindow, 250.0)
		prices, err := referenceInWindowPrices(fix, csv, tvref.TVTimezoneUTC)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(prices) != 0 {
			t.Errorf("pre-window trade must not appear in prices; got %d prices", len(prices))
		}
	})

	t.Run("trade_after_window_excluded", func(t *testing.T) {
		dir := t.TempDir()
		fix := writeTempMultiBarFixture(t, dir, "fix.json", []float64{100, 100})
		csv := writeTempPricedTradeCSV(t, dir, "trades.csv", afterWindow, 250.0)
		prices, err := referenceInWindowPrices(fix, csv, tvref.TVTimezoneUTC)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(prices) != 0 {
			t.Errorf("post-window trade must not appear in prices; got %d prices", len(prices))
		}
	})

	t.Run("start_boundary_included", func(t *testing.T) {
		dir := t.TempDir()
		fix := writeTempMultiBarFixture(t, dir, "fix.json", []float64{100, 100})
		csv := writeTempPricedTradeCSV(t, dir, "trades.csv", startBoundary, 250.0)
		prices, err := referenceInWindowPrices(fix, csv, tvref.TVTimezoneUTC)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(prices) != 1 {
			t.Errorf("trade exactly on start boundary must be included; got %d prices", len(prices))
		}
	})

	t.Run("end_boundary_included", func(t *testing.T) {
		dir := t.TempDir()
		fix := writeTempMultiBarFixture(t, dir, "fix.json", []float64{100, 100})
		csv := writeTempPricedTradeCSV(t, dir, "trades.csv", endBoundary, 250.0)
		prices, err := referenceInWindowPrices(fix, csv, tvref.TVTimezoneUTC)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(prices) != 1 {
			t.Errorf("trade exactly on end boundary must be included; got %d prices", len(prices))
		}
	})

	t.Run("single_in_window_trade_price_returned", func(t *testing.T) {
		dir := t.TempDir()
		fix := writeTempMultiBarFixture(t, dir, "fix.json", []float64{100, 100})
		csv := writeTempPricedTradeCSV(t, dir, "trades.csv", inWindowTime, 250.0)
		prices, err := referenceInWindowPrices(fix, csv, tvref.TVTimezoneUTC)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(prices) != 1 || prices[0] != 250.0 {
			t.Errorf("prices = %v, want [250]", prices)
		}
	})

	t.Run("multiple_in_window_trades_all_prices_returned_in_order", func(t *testing.T) {
		dir := t.TempDir()
		fix := writeTempMultiBarFixture(t, dir, "fix.json", []float64{100, 100})
		csv := writeTempMultiPricedTradeCSV(t, dir, "trades.csv", inWindowTime, []float64{300.0, 400.0})
		prices, err := referenceInWindowPrices(fix, csv, tvref.TVTimezoneUTC)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(prices) != 2 || prices[0] != 300.0 || prices[1] != 400.0 {
			t.Errorf("prices = %v, want [300 400]", prices)
		}
	})

	t.Run("entry_price_used_not_exit_price", func(t *testing.T) {
		dir := t.TempDir()
		fix := writeTempMultiBarFixture(t, dir, "fix.json", []float64{100, 100})
		content := "Trade number,Type,Date and time,Price USD\n" +
			"1,Entry long," + inWindowTime + ",300.00\n" +
			"1,Exit long,2099-01-01 00:00,900.00\n"
		csv := filepath.Join(dir, "trades.csv")
		if err := os.WriteFile(csv, []byte(content), 0644); err != nil {
			t.Fatalf("write CSV: %v", err)
		}
		prices, err := referenceInWindowPrices(fix, csv, tvref.TVTimezoneUTC)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(prices) != 1 || prices[0] != 300.0 {
			t.Errorf("prices = %v, want [300] (entry price, not exit 900)", prices)
		}
	})

	t.Run("missing_fixture_returns_error", func(t *testing.T) {
		dir := t.TempDir()
		csv := writeTempPricedTradeCSV(t, dir, "trades.csv", inWindowTime, 250.0)
		_, err := referenceInWindowPrices(filepath.Join(dir, "nonexistent.json"), csv, tvref.TVTimezoneUTC)
		if err == nil {
			t.Error("expected error for missing fixture, got nil")
		}
	})

	t.Run("missing_csv_returns_error", func(t *testing.T) {
		dir := t.TempDir()
		fix := writeTempMultiBarFixture(t, dir, "fix.json", []float64{100, 100})
		_, err := referenceInWindowPrices(fix, filepath.Join(dir, "nonexistent.csv"), tvref.TVTimezoneUTC)
		if err == nil {
			t.Error("expected error for missing CSV, got nil")
		}
	})
}

// TestPriceScaleConsistent exercises the full pipeline: scale-compatible fixtures
// pass, incompatible fixtures fail, median determines the outcome (not first/last
// values), threshold boundary is exact, trivial-pass on empty reference, and errors
// propagate from both the fixture and CSV paths.
func TestPriceScaleConsistent(t *testing.T) {
	const (
		inWindowTime = "2024-01-01 12:00"
		beforeWindow = "2023-12-31 12:00"
	)

	t.Run("same_scale_passes_and_factor_is_non_zero", func(t *testing.T) {
		dir := t.TempDir()
		fix := writeTempMultiBarFixture(t, dir, "fix.json", []float64{250, 260})
		csv := writeTempPricedTradeCSV(t, dir, "trades.csv", inWindowTime, 255.0)
		ok, factor, err := priceScaleConsistent(fix, csv, tvref.TVTimezoneUTC)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !ok {
			t.Error("same-scale fixture incorrectly flagged as inconsistent")
		}
		if factor == 0 {
			t.Error("factor must be the computed ratio (non-zero) when in-window trades exist")
		}
	})

	t.Run("orders_of_magnitude_mismatch_fails", func(t *testing.T) {
		dir := t.TempDir()
		fix := writeTempMultiBarFixture(t, dir, "fix.json", []float64{100, 100})
		csv := writeTempPricedTradeCSV(t, dir, "trades.csv", inWindowTime, 40000.0)
		ok, factor, err := priceScaleConsistent(fix, csv, tvref.TVTimezoneUTC)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ok {
			t.Error("400× mismatch must be flagged as inconsistent")
		}
		if factor < priceScaleIncompatibilityFactor {
			t.Errorf("factor = %v, expected ≥ %v", factor, priceScaleIncompatibilityFactor)
		}
	})

	t.Run("inverse_large_mismatch_fails_symmetry", func(t *testing.T) {
		dir := t.TempDir()
		fix := writeTempMultiBarFixture(t, dir, "fix.json", []float64{40000, 40000})
		csv := writeTempPricedTradeCSV(t, dir, "trades.csv", inWindowTime, 100.0)
		ok, _, err := priceScaleConsistent(fix, csv, tvref.TVTimezoneUTC)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ok {
			t.Error("400× inverse mismatch must also fail (scale factor is symmetric)")
		}
	})

	t.Run("just_below_threshold_passes", func(t *testing.T) {
		dir := t.TempDir()
		fix := writeTempMultiBarFixture(t, dir, "fix.json", []float64{49, 49})
		csv := writeTempPricedTradeCSV(t, dir, "trades.csv", inWindowTime, 1.0)
		ok, _, err := priceScaleConsistent(fix, csv, tvref.TVTimezoneUTC)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !ok {
			t.Errorf("factor 49× must pass (threshold is exclusive: factor < %.0f)", priceScaleIncompatibilityFactor)
		}
	})

	t.Run("exactly_at_threshold_fails", func(t *testing.T) {
		dir := t.TempDir()
		fix := writeTempMultiBarFixture(t, dir, "fix.json", []float64{50, 50})
		csv := writeTempPricedTradeCSV(t, dir, "trades.csv", inWindowTime, 1.0)
		ok, factor, err := priceScaleConsistent(fix, csv, tvref.TVTimezoneUTC)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ok {
			t.Errorf("factor %.0f× exactly at threshold must fail (comparison is strict <, not ≤)", factor)
		}
	})

	t.Run("median_not_first_bar_close_determines_consistency", func(t *testing.T) {
		dir := t.TempDir()
		fix := writeTempMultiBarFixture(t, dir, "fix.json", []float64{10000, 100, 100})
		csv := writeTempPricedTradeCSV(t, dir, "trades.csv", inWindowTime, 100.0)
		ok, _, err := priceScaleConsistent(fix, csv, tvref.TVTimezoneUTC)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !ok {
			t.Error("median fixture close (100) matches reference (100); outlier first bar (10000) must not cause rejection")
		}
	})

	t.Run("median_not_first_reference_price_determines_consistency", func(t *testing.T) {
		dir := t.TempDir()
		fix := writeTempMultiBarFixture(t, dir, "fix.json", []float64{100, 100})
		csv := writeTempMultiPricedTradeCSV(t, dir, "trades.csv", inWindowTime, []float64{10000, 100, 100})
		ok, _, err := priceScaleConsistent(fix, csv, tvref.TVTimezoneUTC)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !ok {
			t.Error("median reference price (100) matches fixture median (100); outlier first trade (10000) must not cause rejection")
		}
	})

	t.Run("single_bar_fixture_compared_correctly", func(t *testing.T) {
		dir := t.TempDir()
		// single-bar fixture: start == end == 2024-01-01 00:00 UTC; the trade at
		// that timestamp falls on both boundaries and must be included.
		fix := writeTempMultiBarFixture(t, dir, "fix.json", []float64{200})
		csv := writeTempPricedTradeCSV(t, dir, "trades.csv", "2024-01-01 00:00", 200.0)
		ok, _, err := priceScaleConsistent(fix, csv, tvref.TVTimezoneUTC)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !ok {
			t.Error("single-bar fixture at same price as reference must be consistent")
		}
	})

	t.Run("empty_reference_csv_trivially_passes_with_zero_factor", func(t *testing.T) {
		dir := t.TempDir()
		fix := writeTempMultiBarFixture(t, dir, "fix.json", []float64{100, 100})
		csv := writeTempTradeCSV(t, dir, "trades.csv", []barcountTestTrade{})
		ok, factor, err := priceScaleConsistent(fix, csv, tvref.TVTimezoneUTC)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !ok {
			t.Error("empty reference must pass (no in-window trades to compare against)")
		}
		if factor != 0 {
			t.Errorf("factor must be 0 for trivial-pass (no reference), got %v", factor)
		}
	})

	t.Run("all_reference_trades_outside_window_trivially_passes", func(t *testing.T) {
		dir := t.TempDir()
		fix := writeTempMultiBarFixture(t, dir, "fix.json", []float64{100, 100})
		csv := writeTempPricedTradeCSV(t, dir, "trades.csv", beforeWindow, 40000.0)
		ok, _, err := priceScaleConsistent(fix, csv, tvref.TVTimezoneUTC)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !ok {
			t.Error("out-of-window trades must not trigger price-scale rejection")
		}
	})

	t.Run("missing_fixture_returns_error", func(t *testing.T) {
		dir := t.TempDir()
		csv := writeTempPricedTradeCSV(t, dir, "trades.csv", inWindowTime, 250.0)
		_, _, err := priceScaleConsistent(filepath.Join(dir, "nonexistent.json"), csv, tvref.TVTimezoneUTC)
		if err == nil {
			t.Error("expected error for missing fixture, got nil")
		}
	})

	t.Run("missing_csv_returns_error", func(t *testing.T) {
		dir := t.TempDir()
		fix := writeTempMultiBarFixture(t, dir, "fix.json", []float64{100, 100})
		_, _, err := priceScaleConsistent(fix, filepath.Join(dir, "nonexistent.csv"), tvref.TVTimezoneUTC)
		if err == nil {
			t.Error("expected error for missing CSV, got nil")
		}
	})
}

// TestIsCrossSymbolCase covers all registry configurations: nil/empty, single-symbol,
// two- and three-symbol families, mixed registries, and absent cases.
func TestIsCrossSymbolCase(t *testing.T) {
	multiA := makeCase("multi.pine", "AAPL", "a.json")
	multiB := makeCase("multi.pine", "BTC", "b.json")
	triA := makeCase("tri.pine", "EUR", "e.json")
	triB := makeCase("tri.pine", "GBP", "g.json")
	triC := makeCase("tri.pine", "JPY", "j.json")
	single := makeCase("single.pine", "SBERP", "s.json")
	absent := makeCase("absent.pine", "X", "x.json")

	twoSymbolRegistry := []tvAlignmentCase{multiA, multiB, single}
	threeSymbolRegistry := []tvAlignmentCase{triA, triB, triC}
	mixedRegistry := []tvAlignmentCase{multiA, multiB, triA, triB, triC, single}

	cases := []struct {
		name string
		tc   tvAlignmentCase
		all  []tvAlignmentCase
		want bool
	}{
		{name: "nil_registry_returns_false", tc: single, all: nil, want: false},
		{name: "empty_registry_returns_false", tc: single, all: []tvAlignmentCase{}, want: false},
		{name: "all_single_symbol_strategies_not_cross", tc: single, all: []tvAlignmentCase{single}, want: false},
		{name: "two_symbol_family_first_case_is_cross", tc: multiA, all: twoSymbolRegistry, want: true},
		{name: "two_symbol_family_second_case_is_cross", tc: multiB, all: twoSymbolRegistry, want: true},
		{name: "single_symbol_case_in_mixed_registry_not_cross", tc: single, all: twoSymbolRegistry, want: false},
		{name: "three_symbol_family_first_case_is_cross", tc: triA, all: threeSymbolRegistry, want: true},
		{name: "three_symbol_family_second_case_is_cross", tc: triB, all: threeSymbolRegistry, want: true},
		{name: "three_symbol_family_third_case_is_cross", tc: triC, all: threeSymbolRegistry, want: true},
		{name: "two_symbol_case_in_mixed_registry_is_cross", tc: multiA, all: mixedRegistry, want: true},
		{name: "three_symbol_case_in_mixed_registry_is_cross", tc: triA, all: mixedRegistry, want: true},
		{name: "single_symbol_case_in_mixed_registry_not_cross", tc: single, all: mixedRegistry, want: false},
		{name: "case_absent_from_registry_not_cross", tc: absent, all: twoSymbolRegistry, want: false},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := isCrossSymbolCase(tc.tc, tc.all)
			if got != tc.want {
				t.Errorf("isCrossSymbolCase = %v, want %v", got, tc.want)
			}
		})
	}
}
