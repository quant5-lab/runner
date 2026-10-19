//go:build integration

package integration

import (
	"encoding/json"
	"testing"

	"github.com/quant5-lab/runner/tests/util"
)

func TestCrossoverExecution(t *testing.T) {
	t.Parallel()
	pineScript := `//@version=5
strategy("Simple Crossover", overlay=true, pyramiding=1)

openCrossover = ta.crossover(close, open)
if openCrossover
    strategy.entry("long", strategy.long)
`

	testData := []map[string]interface{}{
		{"time": 1704067200, "open": 100.0, "high": 102.0, "low": 98.0, "close": 99.0, "volume": 1000.0},
		{"time": 1704070800, "open": 100.0, "high": 101.0, "low": 97.0, "close": 98.0, "volume": 1000.0},
		{"time": 1704074400, "open": 100.0, "high": 103.0, "low": 96.0, "close": 97.0, "volume": 1000.0},
		{"time": 1704078000, "open": 100.0, "high": 102.0, "low": 95.0, "close": 96.0, "volume": 1000.0},
		{"time": 1704081600, "open": 100.0, "high": 101.0, "low": 94.0, "close": 95.0, "volume": 1000.0},
		{"time": 1704085200, "open": 100.0, "high": 105.0, "low": 99.0, "close": 101.0, "volume": 1500.0},
		{"time": 1704088800, "open": 100.0, "high": 106.0, "low": 100.0, "close": 102.0, "volume": 1200.0},
		{"time": 1704092400, "open": 100.0, "high": 107.0, "low": 101.0, "close": 103.0, "volume": 1100.0},
		{"time": 1704096000, "open": 100.0, "high": 108.0, "low": 102.0, "close": 104.0, "volume": 1300.0},
		{"time": 1704099600, "open": 100.0, "high": 109.0, "low": 103.0, "close": 105.0, "volume": 1400.0},
		{"time": 1704103200, "open": 100.0, "high": 102.0, "low": 97.0, "close": 98.0, "volume": 1000.0},
		{"time": 1704106800, "open": 100.0, "high": 101.0, "low": 96.0, "close": 97.0, "volume": 1000.0},
		{"time": 1704110400, "open": 100.0, "high": 100.0, "low": 95.0, "close": 96.0, "volume": 1000.0},
		{"time": 1704114000, "open": 100.0, "high": 99.0, "low": 94.0, "close": 95.0, "volume": 1000.0},
		{"time": 1704117600, "open": 100.0, "high": 98.0, "low": 93.0, "close": 94.0, "volume": 1000.0},
		{"time": 1704121200, "open": 100.0, "high": 110.0, "low": 99.0, "close": 106.0, "volume": 1600.0},
		{"time": 1704124800, "open": 100.0, "high": 111.0, "low": 105.0, "close": 107.0, "volume": 1200.0},
		{"time": 1704128400, "open": 100.0, "high": 112.0, "low": 106.0, "close": 108.0, "volume": 1100.0},
		{"time": 1704132000, "open": 100.0, "high": 113.0, "low": 107.0, "close": 109.0, "volume": 1300.0},
		{"time": 1704135600, "open": 100.0, "high": 114.0, "low": 108.0, "close": 110.0, "volume": 1400.0},
	}

	exec := util.NewPineExecutor(t)
	rawOutput := exec.ExecuteScriptWithCustomDataRaw(t, "crossover-test", pineScript, testData)

	var result struct {
		Strategy struct {
			OpenTrades []struct {
				EntryBar   int     `json:"entryBar"`
				EntryPrice float64 `json:"entryPrice"`
				Direction  string  `json:"direction"`
			} `json:"openTrades"`
		} `json:"strategy"`
	}

	if err := json.Unmarshal(rawOutput, &result); err != nil {
		t.Fatalf("Parse result: %v", err)
	}

	// pyramiding=1 + no exit logic in this script → only the first crossover entry
	// lands; the second crossover (bar 16) is blocked because the bar-6 trade is
	// still open. This matches TV semantics: pyramiding=N caps simultaneous
	// same-direction entries at N.
	if len(result.Strategy.OpenTrades) != 1 {
		t.Fatalf("Expected 1 crossover trade (pyramiding=1 blocks second entry), got %d",
			len(result.Strategy.OpenTrades))
	}

	trade := result.Strategy.OpenTrades[0]
	if trade.EntryBar != 6 {
		t.Errorf("Trade 0: expected bar 6, got %d", trade.EntryBar)
	}
	if trade.EntryPrice <= 0 {
		t.Errorf("Trade 0: invalid price %.2f", trade.EntryPrice)
	}
	if trade.Direction != "long" {
		t.Errorf("Trade 0: expected 'long', got %q", trade.Direction)
	}

	t.Logf("Crossover test passed: 1 trade at bar 6 (pyramiding=1)")
}
