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
// Zigzag SBERP (RunnerOnly=6 > cap 2): SBERP-1h.json has 333 synthetic flat session-open bars
// (O=H=L=C). Pine's >= and <= are both true at equality, so the zigzag UDF reverses direction on
// every flat bar, emitting a spurious pivot that shifts the valuewhen(sz,sz,0..4) window and fires
// harmonic patterns where TV (real MOEX data, zero flat bars) does not. No code fix is possible
// without BANDAIDing >= / <= semantics globally. Operator must replace SBERP-1h.json with real
// MOEX intraday data.
//
// Moon SBERP/AAPL/BTCUSDT (matched=0): fixture prices (≈100) are incompatible with TV reference
// prices (SBERP≈250 RUB, AAPL≈27 USD, BTC≈40k USD). Moon.pine derives entries from Unix timestamp
// arithmetic alone — price tolerance cannot compensate for mismatched data. Operator must provide
// real monthly fixtures per symbol; no code path forward.
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
	return all
}

func hullCases() []tvAlignmentCase {
	return []tvAlignmentCase{
		{
			Name: "Hull SBERP", Strategy: "top10/hull.pine", Data: "SBERP-1h.json", Symbol: "SBERP", Timeframe: "1h",
			Golden: "hull-sberp-1h.json", CSV: "hull-sberp-1h-reference.csv", Timezone: tvref.TVTimezoneMoscow,
			Tolerance: twoHoursTwoRub, Discrepancy: tvAlignmentDiscrepancy{RunnerOnly: 1, TVOnly: 4},
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
			Tolerance: tvAlignmentTolerance{Time: 2 * time.Hour, Price: 0.10}, Discrepancy: tvAlignmentDiscrepancy{RunnerOnly: 2, TVOnly: 5, TVOnlyCapEscalated: true},
			InitialCapital: 100000,
		},
	}
}

func bbRsiCases() []tvAlignmentCase {
	return []tvAlignmentCase{
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

func bb7Cases() []tvAlignmentCase {
	return []tvAlignmentCase{
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
			Tolerance: oneHour, Discrepancy: tvAlignmentDiscrepancy{TVOnly: 1, FixtureEndOpen: 1},
			InitialCapital:       1000000,
			SkipPnLRatchetReason: "strategy.cash sizing: same pre-window equity divergence as BB7 SBERP — equityRatio inapplicable.",
		},
	}
}

func alphaCases() []tvAlignmentCase {
	return []tvAlignmentCase{
		{
			Name: "Alpha SBERP", Strategy: "top10/alpha.pine", Data: "SBERP-1h.json", Symbol: "SBERP", Timeframe: "1h",
			Golden: "alpha_trend_sberp_1h.golden.json", CSV: "alpha-sberp-1h-reference.csv", Timezone: tvref.TVTimezoneMoscow,
			Tolerance: twoHoursTwoRub, Discrepancy: tvAlignmentDiscrepancy{RunnerOnly: 2, TVOnly: 3},
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
			Tolerance: twoHoursTwoRub, Discrepancy: tvAlignmentDiscrepancy{RunnerOnly: 1, TVOnly: 1},
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
			Tolerance: twoHoursTwoRub, Discrepancy: tvAlignmentDiscrepancy{RunnerOnly: 1, TVOnly: 4},
			InitialCapital: 10000,
		},
	}
}

func aostochCases() []tvAlignmentCase {
	return []tvAlignmentCase{
		{
			Name: "Aostoch SBERP", Strategy: "top10/aostoch.pine", Data: "SBERP-1h.json", Symbol: "SBERP", Timeframe: "1h",
			Golden: "aostoch-sberp-1h.json", CSV: "aostoch-sberp-1h-reference.csv", Timezone: tvref.TVTimezoneMoscow,
			Tolerance: twoHoursTwoRub, Discrepancy: tvAlignmentDiscrepancy{RunnerOnly: 2, TVOnly: 2},
			InitialCapital: 10000,
		},
	}
}
