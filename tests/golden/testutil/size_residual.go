package testutil

import "math"

// MeasuredSizeResidualBound is the worst-case relative deviation between stored
// golden position sizes and live runner output, exhaustively measured across all
// 96 strategy goldens by TestSizeResidual_AllGoldens.
//
// Survey results (exhaustive — all goldens with trades):
//
//	SBERP whole-share goldens:  0        (qty-step floor eliminates JSON round-trip drift)
//	AAPL  whole-share goldens:  0
//	BTCUSDT keltner-squeeze:    0        (regenerated with current code — no drift)
//	BTCUSDT macd-crossover:     9.99e-4  (worst case; golden stamped at intermediate VM-7
//	                                      code state; accumulated equity drift over 413 trades;
//	                                      regenerating the golden would reset this to ~0)
//
// This bound stays below the derived financial tolerance and remains strictly positive.
// When the macd-btcusdt golden is regenerated with the current runner, re-run
// the exhaustive survey and lower this constant accordingly.
const MeasuredSizeResidualBound = 9.995e-4

// MaxGoldenTradeCountDrift is the maximum allowed fractional divergence between
// a golden's closed-trade count and the live runner's closed-trade count for the
// same (strategy, fixture) pair:
//
//	|golden_count - live_count| / max(golden_count, live_count)
//
// Calibrated from worst observed boundary variance across all registered goldens
// (≤ 2 boundary trades on a 62-trade series ≈ 3.2%), with a 4× safety multiple.
// A fixture-window expansion that multiplies the trade series — e.g. the
// 5499-bar → 21927-bar swap that inflated 103 trades to 422 (drift ≈ 0.76) —
// exceeds this bound and forces explicit golden regeneration before the change
// can be merged.
const MaxGoldenTradeCountDrift = 0.15

// TradeCountFidelity returns the fractional divergence between the closed-trade
// counts of two strategy results:
//
//	|a_count - b_count| / max(a_count, b_count)
//
// Returns 0 when both are empty or equal. Returns 1 when one is empty and the
// other is not. Use alongside SizeResidual: SizeResidual stays near zero even
// when a golden is stale because it compares only the overlapping prefix of
// trades; TradeCountFidelity surfaces the surplus that SizeResidual ignores.
func TradeCountFidelity(a, b *StrategyResult) float64 {
	ac := len(a.Trades)
	bc := len(b.Trades)
	if ac == bc {
		return 0
	}
	larger := ac
	if bc > larger {
		larger = bc
	}
	if larger == 0 {
		return 0
	}
	smaller := ac
	if bc < smaller {
		smaller = bc
	}
	return float64(larger-smaller) / float64(larger)
}

// SizeResidual returns the maximum relative deviation in position sizes across
// all matched trades between two strategy results.
// Surplus trades beyond the shorter list are skipped.
func SizeResidual(a, b *StrategyResult) float64 {
	max := sizeResidualOverSlice(a.Trades, b.Trades)
	if open := sizeResidualOverSlice(a.OpenTrades, b.OpenTrades); open > max {
		max = open
	}
	return max
}

func sizeResidualOverSlice(a, b []Trade) float64 {
	var max float64
	limit := len(a)
	if len(b) < limit {
		limit = len(b)
	}
	for i := 0; i < limit; i++ {
		if d := sizeRelDev(a[i].Size, b[i].Size); d > max {
			max = d
		}
	}
	return max
}

func sizeRelDev(a, b float64) float64 {
	if a == b {
		return 0
	}
	scale := math.Max(math.Abs(a), math.Abs(b))
	if scale == 0 {
		return 0
	}
	return math.Abs(a-b) / scale
}
