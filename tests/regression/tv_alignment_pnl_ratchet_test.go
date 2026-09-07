package regression

import (
	"path/filepath"
	"testing"

	tvref "github.com/quant5-lab/runner/tests/regression/tv_reference"
)

// pnlRelativeTolerance is the maximum acceptable relative deviation between
// runner net PnL and TV-reported net PnL per trade
// (Trade.Profit = gross - entry_commission - exit_commission).
//
// Sources of residual error:
//   - TV CSV Net PnL column has 2 decimal places: rounding ≤ 0.005 per trade.
//   - FP accumulation in percent-of-equity sizing across N trades.
const pnlRelativeTolerance = 0.01

// exactAlignmentCases returns cases where entry alignment is exact (both discrepancy
// counts are zero) and per-trade PnL comparison is valid — cases with a
// SkipPnLRatchetReason are excluded because their equityRatio is inapplicable.
func exactAlignmentCases() []tvAlignmentCase {
	var out []tvAlignmentCase
	for _, tc := range tvAlignmentCases() {
		if tc.Discrepancy.RunnerOnly == 0 && tc.Discrepancy.TVOnly == 0 && tc.SkipPnLRatchetReason == "" {
			out = append(out, tc)
		}
	}
	return out
}

func TestPerTradePnL_TVAlignment(t *testing.T) {
	root := projectRootFromCwd()

	for _, tc := range exactAlignmentCases() {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			fixturePath := filepath.Join(root, "tests", "golden", "fixtures", "data", tc.Data)
			fixtureStart, _, err := tvref.FixtureWindow(fixturePath)
			if err != nil {
				t.Fatalf("fixture window: %v", err)
			}

			csvPath := filepath.Join(root, "tests", "regression", "tv_reference", "fixtures", tc.CSV)
			allTVTrades, err := tvref.LoadTrades(csvPath, tc.Timezone)
			if err != nil {
				t.Fatalf("load reference series: %v", err)
			}

			_, fixtureEnd, err := tvref.FixtureWindow(fixturePath)
			if err != nil {
				t.Fatalf("fixture window end: %v", err)
			}
			tvTrades := tvref.FilterByEntryWindow(allTVTrades, fixtureStart, fixtureEnd)
			runnerTrades := loadLiveRunnerClosedTrades(t, root, tc)

			equityRatio := tvref.TVEquityAtWindowStart(allTVTrades, fixtureStart, tc.InitialCapital) / tc.InitialCapital
			matched, pnlMismatch := tvref.MatchNetPnL(runnerTrades, tvTrades, equityRatio, tc.Tolerance.Time, tc.Tolerance.Price, pnlRelativeTolerance)

			if matched == 0 {
				t.Fatalf("%s: no matched trades for PnL comparison", tc.Name)
			}
			if pnlMismatch > tc.PnLDiscrepancy {
				t.Errorf("%s: %d/%d trades exceed %.0f%% net PnL tolerance vs TV reference (expected ≤%d known discrepancies)",
					tc.Name, pnlMismatch, matched, pnlRelativeTolerance*100, tc.PnLDiscrepancy)
			}
		})
	}
}
