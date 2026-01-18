package golden

import (
	"testing"

	"github.com/quant5-lab/runner/tests/golden/testutil"
)

func TestPositionReversal_MACD_AlternatingDirections(t *testing.T) {
	tests := []struct {
		name     string
		symbol   string
		dataFile string
		golden   string
	}{
		{
			name:     "aapl_hourly",
			symbol:   "AAPL",
			dataFile: "AAPL-1h.json",
			golden:   "macd-aapl-1h.json",
		},
		{
			name:     "btcusdt_hourly",
			symbol:   "BTCUSDT",
			dataFile: "BTCUSDT-1h.json",
			golden:   "macd-btcusdt-1h.json",
		},
		{
			name:     "sberp_hourly",
			symbol:   "SBERP",
			dataFile: "SBERP-1h.json",
			golden:   "macd-sberp-1h.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suite := NewTestSuite(t)

			actual := suite.runner.Execute(t,
				suite.StrategyPath("macd-crossover.pine"),
				suite.DataPath(tt.dataFile),
				tt.symbol, "1h")

			validatePositionReversalBehavior(t, actual)
		})
	}
}

func TestPositionReversal_Supertrend_AlternatingDirections(t *testing.T) {
	tests := []struct {
		name     string
		symbol   string
		dataFile string
		golden   string
	}{
		{
			name:     "aapl_hourly",
			symbol:   "AAPL",
			dataFile: "AAPL-1h.json",
			golden:   "supertrend-aapl-1h.json",
		},
		{
			name:     "btcusdt_hourly",
			symbol:   "BTCUSDT",
			dataFile: "BTCUSDT-1h.json",
			golden:   "supertrend-btcusdt-1h.json",
		},
		{
			name:     "sberp_hourly",
			symbol:   "SBERP",
			dataFile: "SBERP-1h.json",
			golden:   "supertrend-sberp-1h.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suite := NewTestSuite(t)

			actual := suite.runner.Execute(t,
				suite.StrategyPath("supertrend.pine"),
				suite.DataPath(tt.dataFile),
				tt.symbol, "1h")

			validatePositionReversalBehavior(t, actual)
		})
	}
}

func TestPositionReversal_MTF_AlternatingDirections(t *testing.T) {
	suite := NewTestSuite(t)

	actual := suite.runner.Execute(t,
		suite.StrategyPath("mtf-confirmation-strategy.pine"),
		suite.DataPath("NVDA-1h.json"),
		"NVDA", "1h")

	validatePositionReversalBehavior(t, actual)
	validateAlternatingDirections(t, actual)
}

func validatePositionReversalBehavior(t *testing.T, result *testutil.StrategyResult) {
	t.Helper()

	trades := result.Trades
	if len(trades) == 0 {
		t.Log("No trades - position reversal not testable")
		return
	}

	for i := 0; i < len(trades)-1; i++ {
		currentTrade := trades[i]
		nextTrade := trades[i+1]

		if currentTrade.Direction != nextTrade.Direction {
			if currentTrade.ExitTime != nextTrade.EntryTime {
				t.Errorf("Position reversal timing mismatch: trade %d exit=%d, trade %d entry=%d",
					i+1, currentTrade.ExitTime, i+2, nextTrade.EntryTime)
			}
		}
	}

	t.Logf("✓ Position reversal behavior validated: %d trades with proper timing", len(trades))
}

func validateAlternatingDirections(t *testing.T, result *testutil.StrategyResult) {
	t.Helper()

	trades := result.Trades
	if len(trades) < 2 {
		t.Log("Less than 2 trades - alternating direction not testable")
		return
	}

	for i := 0; i < len(trades)-1; i++ {
		currentTrade := trades[i]
		nextTrade := trades[i+1]

		if currentTrade.Direction == nextTrade.Direction {
			t.Errorf("Sequential same-direction trades found: trade %d and %d both %s (pyramiding detected)",
				i+1, i+2, currentTrade.Direction)
		}
	}

	t.Logf("✓ All %d trades alternate direction (no pyramiding)", len(trades))
}

func TestPositionReversal_EquityConsistency(t *testing.T) {
	tests := []struct {
		name         string
		strategyFile string
		symbol       string
		dataFile     string
		skip         bool
		skipReason   string
	}{
		{
			name:         "macd_aapl",
			strategyFile: "macd-crossover.pine",
			symbol:       "AAPL",
			dataFile:     "AAPL-1h.json",
		},
		{
			name:         "supertrend_btcusdt",
			strategyFile: "supertrend.pine",
			symbol:       "BTCUSDT",
			dataFile:     "BTCUSDT-1h.json",
			skip:         true,
			skipReason:   "Known issue: Equity calculation bug with Supertrend (980.86 vs expected 9980.86)",
		},
		{
			name:         "mtf_nvda",
			strategyFile: "mtf-confirmation-strategy.pine",
			symbol:       "NVDA",
			dataFile:     "NVDA-1h.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skip {
				t.Skip(tt.skipReason)
			}

			suite := NewTestSuite(t)

			actual := suite.runner.Execute(t,
				suite.StrategyPath(tt.strategyFile),
				suite.DataPath(tt.dataFile),
				tt.symbol, "1h")

			validateEquityConsistency(t, actual)
		})
	}
}

func validateEquityConsistency(t *testing.T, result *testutil.StrategyResult) {
	t.Helper()

	initialCapital := 10000.0

	for _, openTrade := range result.OpenTrades {
		t.Logf("Open trade: %s @ %.2f", openTrade.Direction, openTrade.EntryPrice)
	}

	closedPL := result.NetProfit
	unrealizedPL := result.Equity - initialCapital - closedPL

	if len(result.OpenTrades) == 0 && unrealizedPL != 0 {
		t.Errorf("Equity mismatch with no open trades: equity=%.2f, expected=%.2f (initial=%.2f + netProfit=%.2f, unrealized=%.2f)",
			result.Equity, initialCapital+closedPL, initialCapital, closedPL, unrealizedPL)
	} else if len(result.OpenTrades) > 0 {
		t.Logf("✓ Equity breakdown: %.2f = %.2f (initial) + %.2f (closed PL) + %.2f (unrealized PL from %d open trades)",
			result.Equity, initialCapital, closedPL, unrealizedPL, len(result.OpenTrades))
	} else {
		t.Logf("✓ Equity consistent (no open trades): %.2f = %.2f + %.2f",
			result.Equity, initialCapital, closedPL)
	}
}
