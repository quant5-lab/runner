package tv_reference

import (
	"testing"
	"time"
)

// TestTVExportHorizon covers all input shapes: empty, single, multiple with the
// maximum at the start / middle / end, all-equal timestamps, and a trade with
// non-zero ExitUTC to prove only EntryUTC is inspected.
func TestTVExportHorizon(t *testing.T) {
	t0 := parseUTC("2025-01-10 09:00")
	tLatest := parseUTC("2025-06-15 12:00")
	t2 := parseUTC("2025-03-20 08:00")

	cases := []struct {
		name   string
		trades []TVTrade
		want   time.Time // zero value signals "expect zero time"
	}{
		{
			name:   "empty_slice_returns_zero_time",
			trades: nil,
			want:   time.Time{},
		},
		{
			name:   "single_trade_returns_its_entry_time",
			trades: []TVTrade{{EntryUTC: t0}},
			want:   t0,
		},
		{
			name:   "latest_entry_time_in_middle_of_slice",
			trades: []TVTrade{{EntryUTC: t0}, {EntryUTC: tLatest}, {EntryUTC: t2}},
			want:   tLatest,
		},
		{
			name:   "latest_entry_time_at_start_of_slice",
			trades: []TVTrade{{EntryUTC: tLatest}, {EntryUTC: t2}, {EntryUTC: t0}},
			want:   tLatest,
		},
		{
			name:   "latest_entry_time_at_end_of_slice",
			trades: []TVTrade{{EntryUTC: t0}, {EntryUTC: t2}, {EntryUTC: tLatest}},
			want:   tLatest,
		},
		{
			name:   "all_equal_timestamps_returns_that_time",
			trades: []TVTrade{{EntryUTC: t0}, {EntryUTC: t0}, {EntryUTC: t0}},
			want:   t0,
		},
		{
			name:   "exit_utc_and_other_fields_do_not_affect_result",
			trades: []TVTrade{{EntryUTC: t0, ExitUTC: tLatest, EntryPrice: 999}},
			want:   t0,
		},
		{
			name:   "all_zero_entry_times_returns_zero_time",
			trades: []TVTrade{{EntryUTC: time.Time{}}, {EntryUTC: time.Time{}}},
			want:   time.Time{},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := TVExportHorizon(tc.trades)
			if tc.want.IsZero() {
				if !got.IsZero() {
					t.Errorf("got %v, want zero time", got)
				}
				return
			}
			if !got.Equal(tc.want) {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}

// TestSplitRunnerAtHorizon covers all partition shapes — empty input, all-within,
// all-beyond, exact-at-boundary, nanosecond-past-boundary, and mixed — as well as
// the invariant that the open/closed flag is irrelevant to which partition a trade
// lands in (partitioning is purely by entry time).
func TestSplitRunnerAtHorizon(t *testing.T) {
	h := parseUTC("2025-04-01 00:00")

	r := func(datetime string, price float64, open bool) RunnerTrade {
		return RunnerTrade{EntryUTC: parseUTC(datetime), EntryPrice: price, IsOpen: open}
	}

	cases := []struct {
		name       string
		runner     []RunnerTrade
		horizon    time.Time
		wantWithin []float64
		wantBeyond []float64
	}{
		{
			name:       "empty_input_both_partitions_empty",
			runner:     nil,
			horizon:    h,
			wantWithin: nil,
			wantBeyond: nil,
		},
		{
			name:       "zero_horizon_all_trades_go_to_within",
			runner:     []RunnerTrade{r("2025-01-01 09:00", 1, false), r("2025-06-01 10:00", 2, false)},
			horizon:    time.Time{},
			wantWithin: []float64{1, 2},
			wantBeyond: nil,
		},
		{
			name:       "all_before_horizon_none_beyond",
			runner:     []RunnerTrade{r("2025-01-01 09:00", 1, false), r("2025-03-31 23:59", 2, false)},
			horizon:    h,
			wantWithin: []float64{1, 2},
			wantBeyond: nil,
		},
		{
			name:       "all_after_horizon_none_within",
			runner:     []RunnerTrade{r("2025-04-01 00:01", 1, false), r("2025-06-01 10:00", 2, false)},
			horizon:    h,
			wantWithin: nil,
			wantBeyond: []float64{1, 2},
		},
		{
			name:       "trade_exactly_at_horizon_belongs_in_within",
			runner:     []RunnerTrade{r("2025-04-01 00:00", 42, false)},
			horizon:    h,
			wantWithin: []float64{42},
			wantBeyond: nil,
		},
		{
			name:       "trade_one_nanosecond_after_horizon_belongs_in_beyond",
			runner:     []RunnerTrade{{EntryUTC: h.Add(time.Nanosecond), EntryPrice: 42}},
			horizon:    h,
			wantWithin: nil,
			wantBeyond: []float64{42},
		},
		{
			name: "mixed_partition_preserves_relative_input_order",
			runner: []RunnerTrade{
				r("2025-01-01 09:00", 10, false),
				r("2025-04-01 00:00", 20, false),
				r("2025-04-02 08:00", 30, false),
				r("2025-06-01 10:00", 40, false),
			},
			horizon:    h,
			wantWithin: []float64{10, 20},
			wantBeyond: []float64{30, 40},
		},
		{
			name: "open_and_closed_trades_partitioned_by_entry_time_not_by_open_flag",
			runner: []RunnerTrade{
				r("2025-01-01 09:00", 1, true),
				r("2025-01-02 10:00", 2, false),
				r("2025-05-01 08:00", 3, true),
				r("2025-06-01 10:00", 4, false),
			},
			horizon:    h,
			wantWithin: []float64{1, 2},
			wantBeyond: []float64{3, 4},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			within, beyond := SplitRunnerAtHorizon(tc.runner, tc.horizon)

			assertRunnerSliceByPrice(t, "within", within, tc.wantWithin)
			assertRunnerSliceByPrice(t, "beyond", beyond, tc.wantBeyond)

			if total := len(within) + len(beyond); total != len(tc.runner) {
				t.Errorf("partition completeness: within(%d)+beyond(%d)=%d != runner(%d)",
					len(within), len(beyond), total, len(tc.runner))
			}
		})
	}
}

// TestSplitRunnerAtHorizon_FieldsPreserved verifies that all RunnerTrade fields
// (EntryPrice, Direction, Size, NetPnL, IsOpen, EntryUTC) survive the split
// unchanged in both partitions. The split must not mutate or discard any field
// on any trade regardless of which side of the horizon it lands on.
func TestSplitRunnerAtHorizon_FieldsPreserved(t *testing.T) {
	horizon := parseUTC("2025-06-01 00:00")

	assertAllFields := func(t *testing.T, label string, got, src RunnerTrade) {
		t.Helper()
		if !got.EntryUTC.Equal(src.EntryUTC) {
			t.Errorf("%s EntryUTC: got %v, want %v", label, got.EntryUTC, src.EntryUTC)
		}
		if got.EntryPrice != src.EntryPrice {
			t.Errorf("%s EntryPrice: got %v, want %v", label, got.EntryPrice, src.EntryPrice)
		}
		if got.Direction != src.Direction {
			t.Errorf("%s Direction: got %q, want %q", label, got.Direction, src.Direction)
		}
		if got.Size != src.Size {
			t.Errorf("%s Size: got %v, want %v", label, got.Size, src.Size)
		}
		if got.NetPnL != src.NetPnL {
			t.Errorf("%s NetPnL: got %v, want %v", label, got.NetPnL, src.NetPnL)
		}
		if got.IsOpen != src.IsOpen {
			t.Errorf("%s IsOpen: got %v, want %v", label, got.IsOpen, src.IsOpen)
		}
	}

	t.Run("within_partition", func(t *testing.T) {
		src := RunnerTrade{
			EntryUTC:   parseUTC("2025-01-15 09:30"),
			EntryPrice: 123.45,
			Direction:  "short",
			Size:       7.5,
			NetPnL:     -42.0,
			IsOpen:     true,
		}
		within, _ := SplitRunnerAtHorizon([]RunnerTrade{src}, horizon)
		if len(within) != 1 {
			t.Fatalf("expected 1 trade in within, got %d", len(within))
		}
		assertAllFields(t, "within[0]", within[0], src)
	})

	t.Run("beyond_partition", func(t *testing.T) {
		src := RunnerTrade{
			EntryUTC:   parseUTC("2025-09-20 14:00"),
			EntryPrice: 987.65,
			Direction:  "long",
			Size:       3.0,
			NetPnL:     88.0,
			IsOpen:     false,
		}
		_, beyond := SplitRunnerAtHorizon([]RunnerTrade{src}, horizon)
		if len(beyond) != 1 {
			t.Fatalf("expected 1 trade in beyond, got %d", len(beyond))
		}
		assertAllFields(t, "beyond[0]", beyond[0], src)
	})
}

// TestClosedCount covers all input shapes: empty, single-closed, single-open,
// all-closed, all-open, and two mixed arrangements to detect position-sensitivity bugs.
func TestClosedCount(t *testing.T) {
	opn := func() RunnerTrade { return RunnerTrade{IsOpen: true} }
	cls := func() RunnerTrade { return RunnerTrade{IsOpen: false} }

	cases := []struct {
		name  string
		input []RunnerTrade
		want  int
	}{
		{"empty_slice", nil, 0},
		{"single_closed", []RunnerTrade{cls()}, 1},
		{"single_open", []RunnerTrade{opn()}, 0},
		{"all_closed", []RunnerTrade{cls(), cls(), cls()}, 3},
		{"all_open", []RunnerTrade{opn(), opn()}, 0},
		{"closed_then_open", []RunnerTrade{cls(), cls(), opn(), cls()}, 3},
		{"open_then_closed_alternating", []RunnerTrade{opn(), cls(), opn(), cls(), opn()}, 2},
		{"open_only_at_start", []RunnerTrade{opn(), cls(), cls()}, 2},
		{"open_only_at_end", []RunnerTrade{cls(), cls(), opn()}, 2},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if got := ClosedCount(tc.input); got != tc.want {
				t.Errorf("got %d, want %d", got, tc.want)
			}
		})
	}
}

// TestExportHorizonPipeline exercises the composed behavior of
// TVExportHorizon → SplitRunnerAtHorizon → ClosedCount, which is the
// exact pipeline the alignment ratchet uses to separate phantom RunnerOnly
// trades (beyond TV export coverage) from genuine divergence.
// Each case asserts both the ExportHorizonRunnerOnly count and the true
// RunnerOnly count produced by MatchExact on the within-horizon partition.
func TestExportHorizonPipeline(t *testing.T) {
	timeTol := time.Hour
	priceTol := 1.0

	mkTV := func(datetime string) TVTrade { return makeTVTrade(datetime, 100) }
	mkClosed := func(datetime string) RunnerTrade { return makeRunnerTrade(datetime, 100) }
	mkOpen := func(datetime string) RunnerTrade { return makeOpenRunnerTrade(datetime, 100) }

	cases := []struct {
		name                   string
		tv                     []TVTrade
		runner                 []RunnerTrade
		wantExportHorizonCount int
		wantRunnerOnlyInZone   int
	}{
		{
			name:                   "no_tv_trades_produces_zero_horizon_no_runner_excluded",
			tv:                     nil,
			runner:                 []RunnerTrade{mkClosed("2025-01-01 10:00"), mkClosed("2025-02-01 10:00")},
			wantExportHorizonCount: 0,
			wantRunnerOnlyInZone:   2,
		},
		{
			name:                   "all_runner_within_tv_coverage_no_export_horizon_phantoms",
			tv:                     []TVTrade{mkTV("2025-06-01 10:00")},
			runner:                 []RunnerTrade{mkClosed("2025-01-01 10:00"), mkClosed("2025-03-01 10:00")},
			wantExportHorizonCount: 0,
			wantRunnerOnlyInZone:   2,
		},
		{
			name:                   "closed_runner_trades_beyond_horizon_counted_as_export_horizon_phantoms",
			tv:                     []TVTrade{mkTV("2025-03-01 10:00")},
			runner:                 []RunnerTrade{mkClosed("2025-01-01 10:00"), mkClosed("2025-04-01 10:00"), mkClosed("2025-05-01 10:00")},
			wantExportHorizonCount: 2,
			wantRunnerOnlyInZone:   1,
		},
		{
			name:                   "open_runner_trades_beyond_horizon_not_counted_as_export_horizon_phantoms",
			tv:                     []TVTrade{mkTV("2025-03-01 10:00")},
			runner:                 []RunnerTrade{mkClosed("2025-01-01 10:00"), mkOpen("2025-04-01 10:00"), mkOpen("2025-05-01 10:00")},
			wantExportHorizonCount: 0,
			wantRunnerOnlyInZone:   1,
		},
		{
			name:                   "mixed_open_and_closed_beyond_horizon_only_closed_counted",
			tv:                     []TVTrade{mkTV("2025-03-01 10:00")},
			runner:                 []RunnerTrade{mkClosed("2025-01-01 10:00"), mkClosed("2025-04-01 10:00"), mkOpen("2025-05-01 10:00"), mkClosed("2025-06-01 10:00")},
			wantExportHorizonCount: 2,
			wantRunnerOnlyInZone:   1,
		},
		{
			name:                   "empty_runner_produces_zero_phantoms_and_zero_runner_only",
			tv:                     []TVTrade{mkTV("2025-06-01 10:00")},
			runner:                 nil,
			wantExportHorizonCount: 0,
			wantRunnerOnlyInZone:   0,
		},
		{
			name:                   "runner_entry_at_exact_horizon_boundary_placed_in_within_not_phantom",
			tv:                     []TVTrade{mkTV("2025-03-01 10:00")},
			runner:                 []RunnerTrade{mkClosed("2025-03-01 10:00")},
			wantExportHorizonCount: 0,
			wantRunnerOnlyInZone:   0,
		},
		{
			name:                   "runner_matching_tv_within_horizon_produces_zero_runner_only",
			tv:                     []TVTrade{mkTV("2025-03-01 10:00"), mkTV("2025-06-01 10:00")},
			runner:                 []RunnerTrade{mkClosed("2025-03-01 10:00")},
			wantExportHorizonCount: 0,
			wantRunnerOnlyInZone:   0,
		},
		{
			name:                   "all_closed_runner_beyond_horizon_no_within_partition",
			tv:                     []TVTrade{mkTV("2025-01-15 10:00")},
			runner:                 []RunnerTrade{mkClosed("2025-03-01 10:00"), mkClosed("2025-04-01 10:00"), mkClosed("2025-05-01 10:00")},
			wantExportHorizonCount: 3,
			wantRunnerOnlyInZone:   0,
		},
		{
			name: "large_closed_burst_beyond_horizon_count_scales_linearly",
			tv:   []TVTrade{mkTV("2025-01-15 10:00")},
			runner: func() []RunnerTrade {
				base := parseUTC("2025-02-01 10:00")
				out := make([]RunnerTrade, 12)
				for i := range out {
					out[i] = RunnerTrade{EntryUTC: base.Add(time.Duration(i) * 24 * time.Hour), EntryPrice: 100, Direction: "long"}
				}
				return out
			}(),
			wantExportHorizonCount: 12,
			wantRunnerOnlyInZone:   0,
		},
		{
			name: "within_horizon_runner_only_and_export_horizon_count_are_independent",
			tv:   []TVTrade{mkTV("2025-01-10 10:00"), mkTV("2025-01-15 10:00")},
			runner: func() []RunnerTrade {
				base := parseUTC("2025-02-01 10:00")
				out := make([]RunnerTrade, 14)
				out[0] = mkClosed("2025-01-10 10:00") // matches TV Jan-10 within horizon
				out[1] = mkClosed("2025-01-12 10:00") // within horizon, no compatible TV trade
				for i := 0; i < 12; i++ {
					out[i+2] = RunnerTrade{EntryUTC: base.Add(time.Duration(i) * 24 * time.Hour), EntryPrice: float64(i + 200), Direction: "long"}
				}
				return out
			}(),
			wantExportHorizonCount: 12,
			wantRunnerOnlyInZone:   1,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			horizon := TVExportHorizon(tc.tv)
			runnerWithin, runnerBeyond := SplitRunnerAtHorizon(tc.runner, horizon)
			exportHorizonCount := ClosedCount(runnerBeyond)

			if exportHorizonCount != tc.wantExportHorizonCount {
				t.Errorf("ExportHorizonRunnerOnly: got %d, want %d", exportHorizonCount, tc.wantExportHorizonCount)
			}

			_, runnerOnly, _ := MatchExact(runnerWithin, tc.tv, timeTol, priceTol)
			if runnerOnly != tc.wantRunnerOnlyInZone {
				t.Errorf("RunnerOnly within coverage zone: got %d, want %d", runnerOnly, tc.wantRunnerOnlyInZone)
			}
		})
	}
}

// assertRunnerSliceByPrice checks that the slice contains exactly the given
// prices in the given order. Price is used as a stable per-trade identifier
// in tests that construct trades with distinct prices.
func assertRunnerSliceByPrice(t *testing.T, label string, trades []RunnerTrade, wantPrices []float64) {
	t.Helper()
	if len(trades) != len(wantPrices) {
		t.Errorf("%s: got %d trades, want %d", label, len(trades), len(wantPrices))
		return
	}
	for i, want := range wantPrices {
		if trades[i].EntryPrice != want {
			t.Errorf("%s[%d]: EntryPrice=%.2f, want %.2f", label, i, trades[i].EntryPrice, want)
		}
	}
}
