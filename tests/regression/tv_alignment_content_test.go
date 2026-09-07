package regression

import (
	"path/filepath"
	"testing"

	tvref "github.com/quant5-lab/runner/tests/regression/tv_reference"
)

// TestTVAlignmentCases_InWindowTVTradeCountNonZero verifies that each registered
// TV reference CSV contains at least one trade whose entry falls within the
// corresponding fixture window.
//
// A registry entry with zero in-window TV trades causes TestGoldenRunnerTVAlignment
// and TestLiveRunnerTVAlignment to fail with an opaque "no TV trades in fixture
// window" error. This test surfaces the problem at registration time — before the
// expensive strategy compilation and execution in the runner tests.
func TestTVAlignmentCases_InWindowTVTradeCountNonZero(t *testing.T) {
	root := projectRootFromCwd()
	for _, tc := range tvAlignmentCases() {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			tv := loadTVTradesInFixtureWindow(t, root, tc)
			if len(tv) == 0 {
				t.Errorf(
					"CSV %q has no trades in the fixture window defined by %q — "+
						"either the CSV is empty/malformed or the fixture window excludes all reference trades; "+
						"alignment tests will fail at runtime with an opaque error",
					tc.CSV, tc.Data,
				)
			}
		})
	}
}

// TestTVAlignmentCases_GoldenHasClosedTrades verifies that each registered golden
// snapshot contains at least one closed trade.
//
// A golden with zero closed trades causes TestGoldenRunnerTVAlignment to fail with
// "runner contains zero trades". Catching this at golden-load time makes the root
// cause immediately obvious: either the golden was captured from a broken strategy
// run, or the strategy never closes a position within the fixture window.
func TestTVAlignmentCases_GoldenHasClosedTrades(t *testing.T) {
	root := projectRootFromCwd()
	for _, tc := range tvAlignmentCases() {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			trades := loadGoldenRunnerTrades(t, root, tc)
			closedCount := 0
			for _, tr := range trades {
				if !tr.IsOpen {
					closedCount++
				}
			}
			if closedCount == 0 {
				t.Errorf(
					"golden %q contains no closed trades — "+
						"regenerate the golden after fixing the strategy, "+
						"or verify the strategy closes at least one position within the fixture window",
					tc.Golden,
				)
			}
		})
	}
}

// TestTVAlignmentCases_SomeCasesHaveTVSizeData guards against HasSizeData
// becoming silently dead code: at least one registered case must carry TV size
// data, otherwise the size-emission ratchet path is unreachable in all
// alignment tests.
func TestTVAlignmentCases_SomeCasesHaveTVSizeData(t *testing.T) {
	root := projectRootFromCwd()
	withSize := 0
	for _, tc := range tvAlignmentCases() {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			tv := loadTVTradesInFixtureWindow(t, root, tc)
			if tvref.HasSizeData(tv) {
				withSize++
			}
		})
	}
	if withSize == 0 {
		t.Error("no TV alignment case has size data in its reference CSV — " +
			"HasSizeData is unreachable; update the test or verify CSV content")
	}
}

// TestTVAlignmentCases_FixtureEndOpenBoundedByGoldenOpenTrades enforces that
// FixtureEndOpen never exceeds the golden snapshot's actual open-trade count.
// An attribution with no matching open runner position is unreachable by
// definition and would silently absorb genuine TV divergence.
func TestTVAlignmentCases_FixtureEndOpenBoundedByGoldenOpenTrades(t *testing.T) {
	root := projectRootFromCwd()
	for _, tc := range tvAlignmentCases() {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			if tc.Discrepancy.FixtureEndOpen == 0 {
				return
			}
			trades := loadGoldenRunnerTrades(t, root, tc)
			openCount := 0
			for _, tr := range trades {
				if tr.IsOpen {
					openCount++
				}
			}
			if tc.Discrepancy.FixtureEndOpen > openCount {
				t.Errorf(
					"FixtureEndOpen=%d exceeds golden open-trade count=%d — "+
						"declared exemptions cannot exceed actual open runner positions",
					tc.Discrepancy.FixtureEndOpen, openCount,
				)
			}
		})
	}
}

// TestTVAlignmentCases_ExportHorizonRunnerOnlyBoundedByPostHorizonGoldenTrades enforces that
// ExportHorizonRunnerOnly never exceeds the count of closed runner trades in the golden snapshot
// that entered after the TV export horizon. A declared exemption with no matching post-horizon
// runner trade is unreachable and would silently absorb genuine TV divergence.
func TestTVAlignmentCases_ExportHorizonRunnerOnlyBoundedByPostHorizonGoldenTrades(t *testing.T) {
	root := projectRootFromCwd()
	for _, tc := range tvAlignmentCases() {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			if tc.Discrepancy.ExportHorizonRunnerOnly == 0 {
				return
			}
			tv := loadTVTradesInFixtureWindow(t, root, tc)
			horizon := tvref.TVExportHorizon(tv)
			if horizon.IsZero() {
				t.Fatalf("export horizon is zero — TV CSV has no trades or all are before fixture start")
			}
			trades := loadGoldenRunnerTrades(t, root, tc)
			_, beyond := tvref.SplitRunnerAtHorizon(trades, horizon)
			postHorizonClosed := tvref.ClosedCount(beyond)
			if tc.Discrepancy.ExportHorizonRunnerOnly > postHorizonClosed {
				t.Errorf(
					"ExportHorizonRunnerOnly=%d exceeds golden closed-trade count after horizon %s: got %d — "+
						"declared exemptions cannot exceed actual post-horizon closed runner positions",
					tc.Discrepancy.ExportHorizonRunnerOnly, horizon.UTC().Format("2006-01-02 15:04 UTC"), postHorizonClosed,
				)
			}
		})
	}
}

// TestTVAlignmentCases_FixtureStartWarmupBoundedByEarlyTVTrades enforces that
// FixtureStartWarmup never exceeds the count of TV reference trades whose entry
// is at or before the fixture's first bar. The runner processes from the first
// fixture bar onward, so no more trades than this can be lost to warmup.
func TestTVAlignmentCases_FixtureStartWarmupBoundedByEarlyTVTrades(t *testing.T) {
	root := projectRootFromCwd()
	for _, tc := range tvAlignmentCases() {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			if tc.Discrepancy.FixtureStartWarmup == 0 {
				return
			}
			fixturePath := filepath.Join(root, "tests", "golden", "fixtures", "data", tc.Data)
			fixtureStart, _, err := tvref.FixtureWindow(fixturePath)
			if err != nil {
				t.Fatalf("fixture window: %v", err)
			}
			csvPath := filepath.Join(root, "tests", "regression", "tv_reference", "fixtures", tc.CSV)
			allTVTrades, err := tvref.LoadTrades(csvPath, tc.Timezone)
			if err != nil {
				t.Fatalf("load reference trades: %v", err)
			}
			earlyCount := 0
			for _, tv := range allTVTrades {
				if !tv.EntryUTC.After(fixtureStart) {
					earlyCount++
				}
			}
			if tc.Discrepancy.FixtureStartWarmup > earlyCount {
				t.Errorf(
					"FixtureStartWarmup=%d exceeds count of TV trades at or before fixture start %s: got %d",
					tc.Discrepancy.FixtureStartWarmup, fixtureStart.Format("2006-01-02"), earlyCount,
				)
			}
		})
	}
}

func TestTVAlignmentCases_ExportHorizonRunnerOnly_ImpliesProvenTruncation(t *testing.T) {
	root := projectRootFromCwd()
	for _, tc := range tvAlignmentCases() {
		if tc.Discrepancy.ExportHorizonRunnerOnly == 0 {
			continue
		}
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			fixturePath := filepath.Join(root, "tests", "golden", "fixtures", "data", tc.Data)
			_, fixtureEnd, err := tvref.FixtureWindow(fixturePath)
			if err != nil {
				t.Fatalf("fixture window: %v", err)
			}
			tv := loadTVTradesInFixtureWindow(t, root, tc)
			horizon := tvref.TVExportHorizon(tv)
			if horizon.IsZero() {
				t.Fatalf("TV export horizon is zero — in-window CSV trades required to establish truncation")
			}
			if !horizon.Before(fixtureEnd) {
				t.Errorf(
					"ExportHorizonRunnerOnly=%d is declared but the TV export horizon %s is not before the fixture end %s — "+
						"runner trades after the horizon are only exempt when the reference is provably truncated; "+
						"if TV genuinely stopped trading, remove ExportHorizonRunnerOnly",
					tc.Discrepancy.ExportHorizonRunnerOnly,
					horizon.UTC().Format("2006-01-02 15:04 UTC"),
					fixtureEnd.UTC().Format("2006-01-02 15:04 UTC"),
				)
			}
		})
	}
}
