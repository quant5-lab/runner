package integration

import (
	"testing"

	"github.com/quant5-lab/runner/tests/util"
)

func TestUnaryBooleanInPlot(t *testing.T) {
	pineScript := `//@version=5
strategy("Unary Boolean Plot", overlay=false)

buy_signal = close > 110.0 ? 1.0 : na
sell_signal = close < 100.0 ? 1.0 : na

plot(not na(buy_signal) ? 1 : 0, title="Buy Active", color=color.green)
plot(not na(sell_signal) ? 1 : 0, title="Sell Active", color=color.red)

has_signal = not na(buy_signal)
plot(has_signal ? 1 : 0, title="Has Signal", color=color.blue)
`

	baseTime := int64(1700000000)
	prices := []float64{95, 98, 105, 112, 108, 102, 115, 120, 98, 95, 110, 118}

	testData := []map[string]interface{}{}
	for i, price := range prices {
		testData = append(testData, map[string]interface{}{
			"time":   baseTime + int64(i*3600),
			"open":   price - 1.0,
			"high":   price + 2.0,
			"low":    price - 2.0,
			"close":  price,
			"volume": 1000.0,
		})
	}

	exec := util.NewPineExecutor(t)
	result := exec.ExecuteScriptWithCustomData(t, "unary-bool-test", pineScript, testData)

	expectedTitles := []string{"Buy Active", "Sell Active", "Has Signal"}
	foundTitles := make(map[string]bool)

	for _, plot := range result.Plots {
		for _, expected := range expectedTitles {
			if plot.Title == expected {
				foundTitles[plot.Title] = true
			}
		}
	}

	if len(foundTitles) != len(expectedTitles) {
		t.Errorf("Expected %d unary boolean plots, found %d", len(expectedTitles), len(foundTitles))
		t.Logf("Found titles: %v", foundTitles)
	}

	t.Log("✓ Strategy executed successfully with unary boolean plots")
}

func TestUnaryBooleanInConditional(t *testing.T) {
	pineScript := `//@version=5
strategy("Unary Conditional Test", overlay=true)

sma5 = ta.sma(close, 5)

buy_sig = close > sma5 ? close : na
sell_sig = close < sma5 ? close : na

if not na(buy_sig)
    strategy.entry("long", strategy.long)
    
if not na(sell_sig)
    strategy.close("long")

plot(close, title="Close")
`

	baseTime := int64(1700000000)
	prices := []float64{100, 102, 98, 105, 103, 101, 107, 110}

	testData := []map[string]interface{}{}
	for i, price := range prices {
		testData = append(testData, map[string]interface{}{
			"time":   baseTime + int64(i*3600),
			"open":   price - 1.0,
			"high":   price + 2.0,
			"low":    price - 2.0,
			"close":  price,
			"volume": 1000.0,
		})
	}

	exec := util.NewPineExecutor(t)
	result := exec.ExecuteScriptWithCustomData(t, "unary-cond-test", pineScript, testData)

	if len(result.Plots) == 0 {
		t.Fatal("Strategy did not execute - unary boolean conditionals may have caused runtime errors")
	}

	t.Log("✓ Unary boolean conditional test passed: no runtime errors")
}
