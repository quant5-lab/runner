package regression

import (
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
	RunnerOnly int
	TVOnly     int
}

func exactTVAlignment() tvAlignmentDiscrepancy {
	return tvAlignmentDiscrepancy{}
}

func tvAlignmentCases() []tvAlignmentCase {
	oneHour := tvAlignmentTolerance{Time: time.Hour, Price: 1.00}
	twoHoursTwoRub := tvAlignmentTolerance{Time: 2 * time.Hour, Price: 2.00}
	return []tvAlignmentCase{
		{
			Name: "Hull SBERP", Strategy: "top10/hull.pine", Data: "SBERP-1h.json", Symbol: "SBERP", Timeframe: "1h",
			Golden: "hull-sberp-1h.json", CSV: "hull-sberp-1h-reference.csv", Timezone: tvref.TVTimezoneMoscow,
			Tolerance: twoHoursTwoRub, Discrepancy: tvAlignmentDiscrepancy{RunnerOnly: 1, TVOnly: 3},
			InitialCapital:        100000,
			SkipSizeRatchetReason: "percent-of-equity sizing: TV equity at fixture start is inflated by 316+ pre-window trades since 2021; runner starts fresh at 100,000 RUB — position sizes are not comparable.",
		},
		{
			Name: "UtPlus SBERP", Strategy: "top10/ut+.pine", Data: "SBERP-1h.json", Symbol: "SBERP", Timeframe: "1h",
			Golden: "ut-plus-sberp-1h.json", CSV: "ut-plus-sberp-1h-reference.csv", Timezone: tvref.TVTimezoneMoscow,
			Tolerance: tvAlignmentTolerance{Time: 2 * time.Hour, Price: 0.10}, Discrepancy: tvAlignmentDiscrepancy{RunnerOnly: 2, TVOnly: 4},
			InitialCapital: 100000,
		},
		{
			Name: "BB+RSI SBERP", Strategy: "bb-rsi-strategy.pine", Data: "SBERP-1h.json", Symbol: "SBERP", Timeframe: "1h",
			Golden: "bb-rsi-sberp-1h.json", CSV: "bb-rsi-sberp-1h-reference.csv", Timezone: tvref.TVTimezoneMoscow,
			Tolerance: twoHoursTwoRub, Discrepancy: tvAlignmentDiscrepancy{RunnerOnly: 1, TVOnly: 1},
			InitialCapital:        100000,
			SkipSizeRatchetReason: "percent-of-equity sizing: TV equity at fixture start diverges from runner initial capital due to pre-window trade history — position sizes are not comparable.",
		},
		{
			Name: "BB+RSI BTCUSDT", Strategy: "bb-rsi-strategy.pine", Data: "BTCUSDT-1h.json", Symbol: "BTCUSDT", Timeframe: "1h",
			Golden: "bb-rsi-btcusdt-1h.json", CSV: "bb-rsi-btcusdt-1h-reference.csv", Timezone: tvref.TVTimezoneUTC,
			Tolerance: oneHour, Discrepancy: exactTVAlignment(),
			InitialCapital:        100000,
			SkipSizeRatchetReason: "percent-of-equity sizing: TV equity at fixture start diverges from runner initial capital due to pre-window trade history — position sizes are not comparable.",
		},
		{
			Name: "BB+RSI AAPL", Strategy: "bb-rsi-strategy.pine", Data: "AAPL-1h.json", Symbol: "AAPL", Timeframe: "1h",
			Golden: "bb-rsi-aapl-1h.json", CSV: "bb-rsi-aapl-1h-reference.csv", Timezone: tvref.TVTimezoneNewYork,
			Tolerance: twoHoursTwoRub, Discrepancy: exactTVAlignment(),
			InitialCapital:        100000,
			PnLDiscrepancy:        1, // trade at 2025-10-09 enters 1h before TV (cold-start: TV had open SHORT from before fixture)
			SkipSizeRatchetReason: "percent-of-equity sizing: TV equity at fixture start diverges from runner initial capital due to pre-window trade history — position sizes are not comparable.",
		},
		{
			Name: "BB7 SBERP", Strategy: "bb-strategy-7-rus.pine", Data: "SBERP-1h.json", Symbol: "SBERP", Timeframe: "1h",
			Golden: "bb7-sberp-1h.json", CSV: "bb7-sberp-1h-reference.csv", Timezone: tvref.TVTimezoneMoscow,
			Tolerance: oneHour, Discrepancy: exactTVAlignment(),
			InitialCapital: 1000000,
			SkipPnLRatchetReason: "strategy.cash sizing: TV equity at fixture start (1,453,603 RUB) diverges from " +
				"runner initial capital (1,000,000 RUB) due to pre-window trade history — equityRatio inapplicable. " +
				"Exit-timing shifts (trailing-stop + close_all execution-order difference) further invalidate per-trade PnL comparison.",
		},
		{
			Name: "BB7 BTCUSDT", Strategy: "bb-strategy-7-rus.pine", Data: "BTCUSDT-1h.json", Symbol: "BTCUSDT", Timeframe: "1h",
			Golden: "bb7-btcusdt-1h.json", CSV: "bb7-btcusdt-1h-reference.csv", Timezone: tvref.TVTimezoneUTC,
			Tolerance: oneHour, Discrepancy: exactTVAlignment(),
			InitialCapital:       1000000,
			SkipPnLRatchetReason: "strategy.cash sizing: same pre-window equity divergence as BB7 SBERP — equityRatio inapplicable.",
		},
		// supertrend-sberp-1h-reference.csv is NOT a ratchet candidate: the CSV is exported from the
		// Alorse Supertrend variant (2 trades in fixture window) while runner uses supertrend.pine
		// (139 trades). This is a strategy-identity mismatch, not a runner defect — same boundary
		// as the accepted [~] supertrend-btcusdt item in TODO.md.
	}
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

func assertTVAlignment(t *testing.T, runner []tvref.RunnerTrade, tv []tvref.TVTrade, tc tvAlignmentCase) {
	t.Helper()
	if len(runner) == 0 {
		t.Fatalf("%s: runner contains zero trades", tc.Name)
	}
	if len(tv) == 0 {
		t.Fatalf("%s: no TV trades in fixture window", tc.Name)
	}

	matched, runnerOnly, tvOnly := tvref.MatchExact(runner, tv, tc.Tolerance.Time, tc.Tolerance.Price)
	if runnerOnly != tc.Discrepancy.RunnerOnly || tvOnly != tc.Discrepancy.TVOnly {
		t.Errorf("%s TV alignment discrepancy changed: matched=%d runner=%d tv=%d runner-only=%d want %d, tv-only=%d want %d",
			tc.Name, matched, len(runner), len(tv), runnerOnly, tc.Discrepancy.RunnerOnly, tvOnly, tc.Discrepancy.TVOnly)
	}

	assertSizeAlignment(t, runner, tv, tc)
}

func assertSizeAlignment(t *testing.T, runner []tvref.RunnerTrade, tv []tvref.TVTrade, tc tvAlignmentCase) {
	t.Helper()
	if tc.SkipSizeRatchetReason != "" {
		return
	}
	if !tvref.HasSizeData(tv) {
		return
	}
	_, sizeMismatch, _ := tvref.MatchSize(runner, tv, tc.Tolerance.Time, tc.Tolerance.Price, sizeRelativeTolerance)
	if sizeMismatch > 0 {
		t.Errorf("%s: %d matched trade(s) have size mismatch exceeding %.0f%% tolerance vs TV reference",
			tc.Name, sizeMismatch, sizeRelativeTolerance*100)
	}
}
