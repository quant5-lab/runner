package regression

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"

	goldenutil "github.com/quant5-lab/runner/tests/golden/testutil"
	tvref "github.com/quant5-lab/runner/tests/regression/tv_reference"
)

type tvAlignmentCase struct {
	Name                  string
	Strategy              string
	Data                  string
	Symbol                string
	Timeframe             string
	Golden                string
	CSV                   string
	Timezone              tvref.TVTimezone
	Tolerance             tvAlignmentTolerance
	Discrepancy           tvAlignmentDiscrepancy
	InitialCapital        float64
	PnLDiscrepancy        int    // known PnL mismatches due to cold-start timing offset, not commission bugs
	SkipPnLRatchetReason  string // non-empty: excluded from TestPerTradePnL_TVAlignment; must state why equityRatio is inapplicable
	SkipSizeRatchetReason string // non-empty: excluded from size assertion in assertTVAlignment; must state why runner and TV sizes are not comparable
}

type tvAlignmentTolerance struct {
	Time  time.Duration
	Price float64
}

type tvAlignmentDiscrepancy struct {
	RunnerOnly                          int
	TVOnly                              int
	FixtureEndOpen                      int  // TV trades that runner holds as fixture-end open positions; exempt from the XOR guard
	FixtureStartWarmup                  int  // TV trades before the fixture warmup boundary that runner cannot reproduce; exempt from the XOR guard
	ExportHorizonRunnerOnly             int  // beyond TV export coverage; not counted in RunnerOnly cap
	ExportHorizonRunnerOnlyCapEscalated bool // ExportHorizonRunnerOnly exceeds maxRunnerOnly; set only with operator approval; counted by TestTVAlignmentCases_ExportHorizonCapEscalationRatchet
	RunnerOnlyCapEscalated              bool // RunnerOnly exceeds maxRunnerOnly; set only with operator approval; counted by TestTVAlignmentCases_CapEscalationRatchet
	TVOnlyCapEscalated                  bool // TVOnly exceeds maxTVOnly; set only with operator approval; counted by TestTVAlignmentCases_CapEscalationRatchet
}

func exactTVAlignment() tvAlignmentDiscrepancy {
	return tvAlignmentDiscrepancy{}
}

func tvAlignmentCaseByName(t *testing.T, name string) tvAlignmentCase {
	t.Helper()
	for _, tc := range tvAlignmentCases() {
		if tc.Name == name {
			return tc
		}
	}
	t.Fatalf("TV alignment case %q not found", name)
	return tvAlignmentCase{}
}

func loadTVTradesInFixtureWindow(t *testing.T, root string, tc tvAlignmentCase) []tvref.TVTrade {
	t.Helper()
	fixturePath := filepath.Join(root, "tests", "golden", "fixtures", "data", tc.Data)
	fixtureStart, fixtureEnd, err := tvref.FixtureWindow(fixturePath)
	if err != nil {
		t.Fatalf("fixture window: %v", err)
	}
	tvTrades, err := tvref.LoadTrades(filepath.Join(root, "tests", "regression", "tv_reference", "fixtures", tc.CSV), tc.Timezone)
	if err != nil {
		t.Fatalf("load reference series: %v", err)
	}
	return tvref.FilterByEntryWindow(tvTrades, fixtureStart, fixtureEnd)
}

func loadGoldenRunnerTrades(t *testing.T, root string, tc tvAlignmentCase) []tvref.RunnerTrade {
	t.Helper()
	trades, err := tvref.LoadRunnerTrades(filepath.Join(root, "tests", "golden", "fixtures", "expected", tc.Golden))
	if err != nil {
		t.Fatalf("load golden: %v", err)
	}
	return trades
}

func loadLiveRunnerTrades(t *testing.T, root string, tc tvAlignmentCase) []tvref.RunnerTrade {
	t.Helper()
	result := goldenutil.NewStrategyRunner(t).Execute(
		t,
		filepath.Join(root, "strategies", tc.Strategy),
		filepath.Join(root, "tests", "golden", "fixtures", "data", tc.Data),
		tc.Symbol,
		tc.Timeframe,
	)
	return runnerTradesFromResult(result)
}

func loadLiveRunnerClosedTrades(t *testing.T, root string, tc tvAlignmentCase) []tvref.RunnerTrade {
	t.Helper()
	result := goldenutil.NewStrategyRunner(t).Execute(
		t,
		filepath.Join(root, "strategies", tc.Strategy),
		filepath.Join(root, "tests", "golden", "fixtures", "data", tc.Data),
		tc.Symbol,
		tc.Timeframe,
	)
	return runnerClosedTradesFromResult(result)
}

func runnerTradesFromResult(result *goldenutil.StrategyResult) []tvref.RunnerTrade {
	out := make([]tvref.RunnerTrade, 0, len(result.Trades)+len(result.OpenTrades))
	for _, tr := range result.Trades {
		out = append(out, tvref.RunnerTrade{
			EntryUTC:   time.Unix(tr.EntryTime, 0).UTC(),
			EntryPrice: tr.EntryPrice,
			Direction:  tr.Direction,
			Size:       tr.Size,
			NetPnL:     tr.Profit,
		})
	}
	for _, tr := range result.OpenTrades {
		out = append(out, tvref.RunnerTrade{
			EntryUTC:   time.Unix(tr.EntryTime, 0).UTC(),
			EntryPrice: tr.EntryPrice,
			Direction:  tr.Direction,
			Size:       tr.Size,
			IsOpen:     true,
		})
	}
	return out
}

func runnerClosedTradesFromResult(result *goldenutil.StrategyResult) []tvref.RunnerTrade {
	out := make([]tvref.RunnerTrade, 0, len(result.Trades))
	for _, tr := range result.Trades {
		out = append(out, tvref.RunnerTrade{
			EntryUTC:   time.Unix(tr.EntryTime, 0).UTC(),
			EntryPrice: tr.EntryPrice,
			Direction:  tr.Direction,
			Size:       tr.Size,
			NetPnL:     tr.Profit,
		})
	}
	return out
}

// sizeRelativeTolerance covers floating-point rounding at the lot-quantization step;
// observed residuals for strategy.cash strategies are < 0.01%.
const sizeRelativeTolerance = 0.02

func tvOnlyCapViolation(name string, d tvAlignmentDiscrepancy) string {
	return capEscalationViolation(name, "TVOnly", d.TVOnly, maxTVOnly, d.TVOnlyCapEscalated, "TVOnlyCapEscalated")
}

func runnerOnlyCapViolation(name string, d tvAlignmentDiscrepancy) string {
	return capEscalationViolation(name, "RunnerOnly", d.RunnerOnly, maxRunnerOnly, d.RunnerOnlyCapEscalated, "RunnerOnlyCapEscalated")
}

func exportHorizonRunnerOnlyCapViolation(name string, d tvAlignmentDiscrepancy) string {
	return capEscalationViolation(name, "ExportHorizonRunnerOnly", d.ExportHorizonRunnerOnly, maxRunnerOnly, d.ExportHorizonRunnerOnlyCapEscalated, "ExportHorizonRunnerOnlyCapEscalated")
}

func capEscalationViolation(name, boundary string, actual, cap int, escalated bool, flag string) string {
	if actual > cap && !escalated {
		return fmt.Sprintf(
			"%s: %s boundary %d exceeds policy cap %d; set %s:true for operator-approved escalations",
			name, boundary, actual, cap, flag,
		)
	}
	if escalated && actual <= cap {
		return fmt.Sprintf(
			"%s: %s is set but %s=%d does not exceed cap %d — remove %s",
			name, flag, boundary, actual, cap, flag,
		)
	}
	return ""
}

func assertTVAlignment(t *testing.T, runner []tvref.RunnerTrade, tv []tvref.TVTrade, tc tvAlignmentCase) {
	t.Helper()
	if len(runner) == 0 {
		t.Fatalf("%s: runner contains zero trades", tc.Name)
	}
	if len(tv) == 0 {
		t.Fatalf("%s: no TV trades in fixture window", tc.Name)
	}

	horizon := tvref.TVExportHorizon(tv)
	runnerWithin, runnerBeyond := tvref.SplitRunnerAtHorizon(runner, horizon)
	exportHorizonRunnerOnly := tvref.ClosedCount(runnerBeyond)

	matched, runnerOnly, tvOnly := tvref.MatchExact(runnerWithin, tv, tc.Tolerance.Time, tc.Tolerance.Price)
	if runnerOnly != tc.Discrepancy.RunnerOnly || tvOnly != tc.Discrepancy.TVOnly {
		t.Errorf("%s TV alignment discrepancy changed: matched=%d runner=%d tv=%d runner-only=%d want %d, tv-only=%d want %d",
			tc.Name, matched, len(runnerWithin), len(tv), runnerOnly, tc.Discrepancy.RunnerOnly, tvOnly, tc.Discrepancy.TVOnly)
	}
	if exportHorizonRunnerOnly != tc.Discrepancy.ExportHorizonRunnerOnly {
		t.Errorf("%s: export-horizon runner-only count changed: got %d want %d (closed runner trades entered after TV export horizon %s)",
			tc.Name, exportHorizonRunnerOnly, tc.Discrepancy.ExportHorizonRunnerOnly, horizon.UTC().Format("2006-01-02 15:04 UTC"))
	}

	assertSizeAlignment(t, runnerWithin, tv, tc)
}

func assertSizeAlignment(t *testing.T, runner []tvref.RunnerTrade, tv []tvref.TVTrade, tc tvAlignmentCase) {
	t.Helper()
	if !tvref.HasSizeData(tv) {
		return
	}
	if n := tvref.ZeroSizeRunnerMatches(runner, tv, tc.Tolerance.Time, tc.Tolerance.Price); n > 0 {
		t.Errorf("%s: %d matched runner trade(s) carry Size=0 — runner is not emitting position quantities; named-argument qty resolution suspected",
			tc.Name, n)
	}
	if tc.SkipSizeRatchetReason != "" {
		return
	}
	_, sizeMismatch, _ := tvref.MatchSize(runner, tv, tc.Tolerance.Time, tc.Tolerance.Price, sizeRelativeTolerance)
	if sizeMismatch > 0 {
		t.Errorf("%s: %d matched trade(s) have size mismatch exceeding %.0f%% tolerance vs TV reference",
			tc.Name, sizeMismatch, sizeRelativeTolerance*100)
	}
}
