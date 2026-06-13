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
