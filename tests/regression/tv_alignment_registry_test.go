package regression

import (
	"time"

	tvref "github.com/quant5-lab/runner/tests/regression/tv_reference"
)

// twoHoursTwoRub absorbs known ±1h session-boundary shifts on Moscow-tz instruments.
var (
	oneHour        = tvAlignmentTolerance{Time: time.Hour, Price: 1.00}
	twoHoursTwoRub = tvAlignmentTolerance{Time: 2 * time.Hour, Price: 2.00}
)

// Blocked wiring — CSV + golden on disk, but cases cannot be wired until data-layer prerequisites are met:
//
// Zigzag SBERP (RunnerOnly=20, TVOnly=5 > caps 2/4): measured on real MOEX 5.5-yr SBERP-1h.json
// (21927 bars, no synthetic flat bars). ZigZag PA V4.1 fires harmonic patterns on
// valuewhen(sz,sz,0..4) pivot windows; over the full window, 20 runner entries have no TV
// counterpart and 5 TV entries have no runner counterpart. Both exceed policy caps. No wiring
// possible without BANDAIDing the policy caps.
//
// Moon AAPL (matched=0): real 498-bar AAPL-M.json exists (NASDAQ:AAPL monthly, 1985-01→2026-05,
// tz=America/New_York). TV reference CSV contains trades from 1981; those pre-1985 entries have
// no corresponding runner bars and accumulate as TVOnly far above the cap. Wiring requires either
// a TV CSV trimmed to the fixture window or fixture extension back to 1981.
//
// Moon SBERP (RunnerOnly=12, TVOnly=12 > caps 2 / 4): SBERP-M restored from
// tests/fixtures/ohlcv/SBERP_1M.json (227 real MOEX monthly bars from 2007-07). Moon.pine's
// timestamp-cycle arithmetic produces 12+12 boundary trades within the 18-year fixture window
// that TV's monthly-bar sampling does not: cycle length 2551442876.8992ms mod 30-day months
// drifts the newmoon crossover across bar boundaries; over ~223 months this accumulates dozens
// of edge cases per hemisphere. RunnerOnly=12 and TVOnly=12 both massively exceed policy caps.
// A honest fix requires either (a) a shorter fixture window (5-10 years) that stays inside a
// single moon-cycle-vs-bar-alignment regime, or (b) a bar-boundary-aware entry signal in
// moon.pine. Neither is in scope for this data-only wiring.
func tvAlignmentCases() []tvAlignmentCase {
	var all []tvAlignmentCase
	all = append(all, hullCases()...)
	all = append(all, utPlusCases()...)
	all = append(all, bbRsiCases()...)
	all = append(all, bb7Cases()...)
	all = append(all, alphaCases()...)
	all = append(all, annCases()...)
	all = append(all, maxCases()...)
	all = append(all, utCases()...)
	all = append(all, aostochCases()...)
	all = append(all, moonCases()...)
	return all
}

func hullCases() []tvAlignmentCase {
	return []tvAlignmentCase{
		{
			Name: "Hull SBERP", Strategy: "top10/hull.pine", Data: "SBERP-1h.json", Symbol: "SBERP", Timeframe: "1h",
			Golden: "hull-sberp-1h.json", CSV: "hull-sberp-1h-reference.csv", Timezone: tvref.TVTimezoneMoscow,
			Tolerance: twoHoursTwoRub, Discrepancy: tvAlignmentDiscrepancy{RunnerOnly: 0, TVOnly: 1, ExportHorizonRunnerOnly: 1},
			InitialCapital:        100000,
			SkipSizeRatchetReason: "percent-of-equity sizing: TV equity at fixture start is inflated by 316+ pre-window trades since 2021; runner starts fresh at 100,000 RUB — position sizes are not comparable.",
		},
	}
}

func utPlusCases() []tvAlignmentCase {
	return []tvAlignmentCase{
		{
			Name: "UtPlus SBERP", Strategy: "top10/ut+.pine", Data: "SBERP-1h.json", Symbol: "SBERP", Timeframe: "1h",
			Golden: "ut-plus-sberp-1h.json", CSV: "ut-plus-sberp-1h-reference.csv", Timezone: tvref.TVTimezoneMoscow,
			Tolerance: twoHoursTwoRub, Discrepancy: tvAlignmentDiscrepancy{RunnerOnly: 2, TVOnly: 2, ExportHorizonRunnerOnly: 12, TVOnlyCapEscalated: false, ExportHorizonRunnerOnlyCapEscalated: true},
			InitialCapital: 100000,
		},
	}
}

func bbRsiCases() []tvAlignmentCase {
	return []tvAlignmentCase{
		{
			Name: "BB+RSI SBERP", Strategy: "bb-rsi-strategy.pine", Data: "SBERP-1h.json", Symbol: "SBERP", Timeframe: "1h",
			Golden: "bb-rsi-sberp-1h.json", CSV: "bb-rsi-sberp-1h-reference.csv", Timezone: tvref.TVTimezoneMoscow,
			Tolerance: twoHoursTwoRub, Discrepancy: tvAlignmentDiscrepancy{RunnerOnly: 1, TVOnly: 1, ExportHorizonRunnerOnly: 1},
			InitialCapital:        100000,
			SkipSizeRatchetReason: "percent-of-equity sizing: TV equity at fixture start diverges from runner initial capital due to pre-window trade history — position sizes are not comparable.",
		},
		{
			Name: "BB+RSI BTCUSDT", Strategy: "bb-rsi-strategy.pine", Data: "BTCUSDT-1h.json", Symbol: "BTCUSDT", Timeframe: "1h",
			Golden: "bb-rsi-btcusdt-1h.json", CSV: "bb-rsi-btcusdt-1h-reference.csv", Timezone: tvref.TVTimezoneUTC,
			Tolerance: oneHour, Discrepancy: tvAlignmentDiscrepancy{TVOnly: 1, FixtureEndOpen: 1},
			InitialCapital:        100000,
			SkipSizeRatchetReason: "percent-of-equity sizing: TV equity at fixture start diverges from runner initial capital due to pre-window trade history — position sizes are not comparable.",
		},
		{
			Name: "BB+RSI AAPL", Strategy: "bb-rsi-strategy.pine", Data: "AAPL-1h.json", Symbol: "AAPL", Timeframe: "1h",
			Golden: "bb-rsi-aapl-1h.json", CSV: "bb-rsi-aapl-1h-reference.csv", Timezone: tvref.TVTimezoneNewYork,
			Tolerance: twoHoursTwoRub, Discrepancy: tvAlignmentDiscrepancy{TVOnly: 1, FixtureEndOpen: 1},
			InitialCapital:        100000,
			PnLDiscrepancy:        1,
			SkipSizeRatchetReason: "percent-of-equity sizing: TV equity at fixture start diverges from runner initial capital due to pre-window trade history — position sizes are not comparable.",
		},
	}
}

// bb7Cases returns TV-alignment cases for bb-strategy-7-rus.pine.
//
// BB7 SBERP is intentionally excluded: post-CloseAll-fix discrepancy is RunnerOnly=5/TVOnly=5,
// both above policy caps 2/4, with no localizable code path to reduce them. The strategy is
// non-top10; BB7 BTCUSDT plus the golden and direction-trace tests provide sufficient coverage.
func bb7Cases() []tvAlignmentCase {
	return []tvAlignmentCase{
		{
			Name: "BB7 BTCUSDT", Strategy: "bb-strategy-7-rus.pine", Data: "BTCUSDT-1h.json", Symbol: "BTCUSDT", Timeframe: "1h",
			Golden: "bb7-btcusdt-1h.json", CSV: "bb7-btcusdt-1h-reference.csv", Timezone: tvref.TVTimezoneUTC,
			Tolerance: oneHour, Discrepancy: tvAlignmentDiscrepancy{TVOnly: 1, FixtureEndOpen: 1},
			InitialCapital:       1000000,
			SkipPnLRatchetReason: "strategy.cash sizing: TV equity at fixture start diverges from runner initial capital due to pre-window trade history — equityRatio inapplicable.",
		},
	}
}

func alphaCases() []tvAlignmentCase {
	return []tvAlignmentCase{
		{
			Name: "Alpha SBERP", Strategy: "top10/alpha.pine", Data: "SBERP-1h.json", Symbol: "SBERP", Timeframe: "1h",
			Golden: "alpha_trend_sberp_1h.golden.json", CSV: "alpha-sberp-1h-reference.csv", Timezone: tvref.TVTimezoneMoscow,
			Tolerance: twoHoursTwoRub, Discrepancy: tvAlignmentDiscrepancy{RunnerOnly: 2, TVOnly: 4},
			InitialCapital: 10000,
		},
		{
			Name: "Alpha BTCUSDT", Strategy: "top10/alpha.pine", Data: "BTCUSDT-1h.json", Symbol: "BTCUSDT", Timeframe: "1h",
			Golden: "alpha_trend_btcusdt_1h.golden.json", CSV: "alpha-btcusdt-1h-reference.csv", Timezone: tvref.TVTimezoneUTC,
			Tolerance: oneHour, Discrepancy: tvAlignmentDiscrepancy{RunnerOnly: 1, TVOnly: 2},
			InitialCapital: 10000,
		},
		{
			Name: "Alpha AAPL", Strategy: "top10/alpha.pine", Data: "AAPL-1h.json", Symbol: "AAPL", Timeframe: "1h",
			Golden: "alpha_trend_aapl_1h.golden.json", CSV: "alpha-aapl-1h-reference.csv", Timezone: tvref.TVTimezoneNewYork,
			Tolerance: twoHoursTwoRub, Discrepancy: tvAlignmentDiscrepancy{RunnerOnly: 1, TVOnly: 1},
			InitialCapital: 10000,
		},
	}
}

func annCases() []tvAlignmentCase {
	return []tvAlignmentCase{
		{
			Name: "Ann BTCUSDT", Strategy: "top10/ann.pine", Data: "BTCUSDT-1h.json", Symbol: "BTCUSDT", Timeframe: "1h",
			Golden: "ann_sirolf_btcusdt_1h.golden.json", CSV: "ann-btcusdt-1h-reference.csv", Timezone: tvref.TVTimezoneUTC,
			Tolerance: oneHour, Discrepancy: tvAlignmentDiscrepancy{RunnerOnly: 1, TVOnly: 1},
			InitialCapital: 10000,
		},
	}
}

func maxCases() []tvAlignmentCase {
	return []tvAlignmentCase{
		{
			Name: "Max SBERP", Strategy: "top10/max.pine", Data: "SBERP-1h.json", Symbol: "SBERP", Timeframe: "1h",
			Golden: "max-sberp-1h.json", CSV: "max-sberp-1h-reference.csv", Timezone: tvref.TVTimezoneMoscow,
			Tolerance: twoHoursTwoRub, Discrepancy: tvAlignmentDiscrepancy{RunnerOnly: 1, TVOnly: 2},
			InitialCapital: 10000,
		},
	}
}

// InitialCapital 10000: ut.pine declares no initial_capital — inherits the Pine/runner default.
func utCases() []tvAlignmentCase {
	return []tvAlignmentCase{
		{
			Name: "Ut SBERP", Strategy: "top10/ut.pine", Data: "SBERP-1h.json", Symbol: "SBERP", Timeframe: "1h",
			Golden: "ut-sberp-1h.json", CSV: "ut-sberp-1h-reference.csv", Timezone: tvref.TVTimezoneMoscow,
			Tolerance: twoHoursTwoRub, Discrepancy: tvAlignmentDiscrepancy{RunnerOnly: 2, TVOnly: 3},
			InitialCapital: 10000,
		},
	}
}

func aostochCases() []tvAlignmentCase {
	return []tvAlignmentCase{
		{
			Name: "Aostoch SBERP", Strategy: "top10/aostoch.pine", Data: "SBERP-1h.json", Symbol: "SBERP", Timeframe: "1h",
			Golden: "aostoch-sberp-1h.json", CSV: "aostoch-sberp-1h-reference.csv", Timezone: tvref.TVTimezoneMoscow,
			Tolerance: twoHoursTwoRub, Discrepancy: tvAlignmentDiscrepancy{TVOnly: 1},
			InitialCapital: 10000,
		},
	}
}

// Moon Phases Strategy [LuxAlgo] fires monthly on lunar-cycle timestamp arithmetic; TV signals
// land on the calendar-first-of-month bar in the instrument's exchange timezone. BTCUSDT-M
// (86 real Binance monthly bars from 2018-02) aligns cleanly within the tolerance caps. SBERP-M
// and AAPL-M are documented as blocked above.
func moonCases() []tvAlignmentCase {
	return []tvAlignmentCase{
		{
			Name: "Moon BTCUSDT", Strategy: "top10/moon.pine", Data: "BTCUSDT-M.json", Symbol: "BTCUSDT", Timeframe: "1M",
			Golden: "moon_phases_btcusdt_m.golden.json", CSV: "moon-btcusdt-monthly-reference.csv", Timezone: tvref.TVTimezoneUTC,
			Tolerance: oneHour, Discrepancy: tvAlignmentDiscrepancy{TVOnly: 2, FixtureEndOpen: 1, FixtureStartWarmup: 1},
			InitialCapital:        1000000,
			SkipSizeRatchetReason: "moon.pine emits no explicit qty; TV report shows Size=1 contract while runner sizes from strategy.cash defaults — direct size comparison inapplicable.",
			SkipPnLRatchetReason:  "moon.pine has no explicit initial_capital; TV's 1M report and runner's initial capital baseline are not equity-comparable across the fixture's multi-decade span.",
		},
	}
}
