package regression

import (
	"path/filepath"
	"sort"
	"testing"
	"time"

	tvref "github.com/quant5-lab/runner/tests/regression/tv_reference"
)

// TestMoexGapProbe_FullUnmatchedList logs every unmatched runner and TV trade
// for all SBERP alignment cases. Derives tolerance and case parameters from the
// live registry, so the output stays consistent when declarations change.
// Intentionally assertion-free to avoid interfering with the alignment ratchet.
func TestMoexGapProbe_FullUnmatchedList(t *testing.T) {
	root := projectRootFromCwd()
	fixturePath := filepath.Join(root, "tests/golden/fixtures/data/SBERP-1h.json")
	fixtureStart, fixtureEnd, _ := tvref.FixtureWindow(fixturePath)

	for _, tc := range tvAlignmentCases() {
		if tc.Data != "SBERP-1h.json" {
			continue
		}

		runner, _ := tvref.LoadRunnerTrades(filepath.Join(root, "tests/golden/fixtures/expected", tc.Golden))
		tvTrades, _ := tvref.LoadTrades(filepath.Join(root, "tests/regression/tv_reference/fixtures", tc.CSV), tc.Timezone)
		tvInWindow := tvref.FilterByEntryWindow(tvTrades, fixtureStart, fixtureEnd)
		tol := tvref.MatchTolerance{Time: tc.Tolerance.Time, Price: tc.Tolerance.Price}
		pairs := tvref.ClosedRunnerPairs(runner, tvInWindow, tol)

		matchedRunner := map[int]bool{}
		matchedTV := map[int]bool{}
		for _, p := range pairs {
			matchedRunner[p.RunnerIdx] = true
			matchedTV[p.TVIdx] = true
		}

		var roDates, tvDates []time.Time
		for i, r := range runner {
			if !r.IsOpen && !matchedRunner[i] {
				roDates = append(roDates, r.EntryUTC)
			}
		}
		for i, tv := range tvInWindow {
			if !matchedTV[i] {
				tvDates = append(tvDates, tv.EntryUTC)
			}
		}

		sort.Slice(roDates, func(i, j int) bool { return roDates[i].Before(roDates[j]) })
		sort.Slice(tvDates, func(i, j int) bool { return tvDates[i].Before(tvDates[j]) })

		t.Logf("=== %s (tol: %s / %.2f) ===", tc.Name, tc.Tolerance.Time, tc.Tolerance.Price)
		t.Logf("RunnerOnly (%d):", len(roDates))
		for _, d := range roDates {
			t.Logf("  %s UTC", d.Format("2006-01-02 Mon 15:04"))
		}
		t.Logf("TVOnly (%d):", len(tvDates))
		for _, d := range tvDates {
			t.Logf("  %s UTC", d.Format("2006-01-02 Mon 15:04"))
		}
	}
}
