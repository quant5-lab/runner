package testutil

import (
	"fmt"
	"math"
	"testing"
)

type Asserter struct {
	t        *testing.T
	expected *StrategyResult
	actual   *StrategyResult
}

func NewAsserter(t *testing.T, expected, actual *StrategyResult) *Asserter {
	return &Asserter{
		t:        t,
		expected: expected,
		actual:   actual,
	}
}

func (a *Asserter) AssertTradeCount() *Asserter {
	a.t.Helper()

	if len(a.expected.Trades) != len(a.actual.Trades) {
		a.t.Errorf("Trade count mismatch: expected %d, got %d",
			len(a.expected.Trades), len(a.actual.Trades))
	}

	return a
}

func (a *Asserter) AssertTrades() *Asserter {
	a.t.Helper()

	minLen := len(a.expected.Trades)
	if len(a.actual.Trades) < minLen {
		minLen = len(a.actual.Trades)
	}

	for i := 0; i < minLen; i++ {
		a.assertTrade(i, &a.expected.Trades[i], &a.actual.Trades[i])
	}

	return a
}

func (a *Asserter) AssertSampleTrades(indices []int) *Asserter {
	a.t.Helper()

	for _, i := range indices {
		if i >= len(a.expected.Trades) || i >= len(a.actual.Trades) {
			a.t.Errorf("Trade index %d out of bounds", i)
			continue
		}
		a.assertTrade(i, &a.expected.Trades[i], &a.actual.Trades[i])
	}

	return a
}

func (a *Asserter) assertTrade(index int, expected, actual *Trade) {
	a.t.Helper()

	if expected.EntryBar != actual.EntryBar {
		a.t.Errorf("Trade[%d]: entryBar mismatch: expected %d, got %d",
			index, expected.EntryBar, actual.EntryBar)
	}

	if expected.ExitBar != actual.ExitBar {
		a.t.Errorf("Trade[%d]: exitBar mismatch: expected %d, got %d",
			index, expected.ExitBar, actual.ExitBar)
	}

	if !matchPrice(expected.EntryPrice, actual.EntryPrice) {
		a.t.Errorf("Trade[%d]: entryPrice mismatch: expected %.2f, got %.2f",
			index, expected.EntryPrice, actual.EntryPrice)
	}

	if !matchPrice(expected.ExitPrice, actual.ExitPrice) {
		a.t.Errorf("Trade[%d]: exitPrice mismatch: expected %.2f, got %.2f",
			index, expected.ExitPrice, actual.ExitPrice)
	}

	if !matchFinancial(expected.Profit, actual.Profit) {
		a.t.Errorf("Trade[%d]: profit mismatch: expected %.2f, got %.2f",
			index, expected.Profit, actual.Profit)
	}

	if expected.Direction != actual.Direction {
		a.t.Errorf("Trade[%d]: direction mismatch: expected %s, got %s",
			index, expected.Direction, actual.Direction)
	}
}

func (a *Asserter) AssertPlot(plotName string, sampleBars []int, tolerance float64) *Asserter {
	a.t.Helper()

	expectedValues, okExp := a.expected.Plots[plotName]
	actualValues, okAct := a.actual.Plots[plotName]

	if !okExp {
		a.t.Errorf("Plot %q not found in expected results", plotName)
		return a
	}

	if !okAct {
		a.t.Errorf("Plot %q not found in actual results", plotName)
		return a
	}

	for _, bar := range sampleBars {
		if bar >= len(expectedValues) || bar >= len(actualValues) {
			a.t.Errorf("Plot %q: bar index %d out of bounds", plotName, bar)
			continue
		}

		exp := expectedValues[bar]
		act := actualValues[bar]

		if math.IsNaN(exp) && math.IsNaN(act) {
			continue
		}

		if math.IsNaN(exp) != math.IsNaN(act) {
			a.t.Errorf("Plot %q[%d]: NaN mismatch: expected %v, got %v",
				plotName, bar, exp, act)
			continue
		}

		if !a.floatClose(exp, act, tolerance) {
			a.t.Errorf("Plot %q[%d]: value mismatch: expected %.4f, got %.4f (tolerance=%.4f)",
				plotName, bar, exp, act, tolerance)
		}
	}

	return a
}

func (a *Asserter) AssertNetProfit(tolerance float64) *Asserter {
	a.t.Helper()

	if !a.floatClose(a.expected.NetProfit, a.actual.NetProfit, tolerance) {
		a.t.Errorf("NetProfit mismatch: expected %.2f, got %.2f",
			a.expected.NetProfit, a.actual.NetProfit)
	}

	return a
}

func (a *Asserter) AssertEquity(tolerance float64) *Asserter {
	a.t.Helper()

	if !a.floatClose(a.expected.Equity, a.actual.Equity, tolerance) {
		a.t.Errorf("Equity mismatch: expected %.2f, got %.2f",
			a.expected.Equity, a.actual.Equity)
	}

	return a
}

func (a *Asserter) floatClose(expected, actual, tolerance float64) bool {
	return floatWithin(expected, actual, tolerance)
}

func AssertTradesMatch(t *testing.T, expected, actual []Trade) {
	t.Helper()

	if len(expected) != len(actual) {
		t.Fatalf("Trade count: expected %d, got %d", len(expected), len(actual))
	}

	for i := range expected {
		exp := &expected[i]
		act := &actual[i]

		if exp.EntryBar != act.EntryBar {
			t.Errorf("Trade[%d].entryBar: expected %d, got %d", i, exp.EntryBar, act.EntryBar)
		}

		if exp.ExitBar != act.ExitBar {
			t.Errorf("Trade[%d].exitBar: expected %d, got %d", i, exp.ExitBar, act.ExitBar)
		}

		if !matchFinancial(exp.Profit, act.Profit) {
			t.Errorf("Trade[%d].profit: expected %.2f, got %.2f", i, exp.Profit, act.Profit)
		}
	}
}

func AssertPlotSmokePoints(t *testing.T, plotName string, expectedValues, actualValues []float64, sampleIndices []int) {
	t.Helper()

	for _, idx := range sampleIndices {
		if idx >= len(expectedValues) || idx >= len(actualValues) {
			t.Errorf("Plot %q: index %d out of bounds", plotName, idx)
			continue
		}

		exp := expectedValues[idx]
		act := actualValues[idx]

		if !matchPlot(exp, act) {
			t.Errorf("Plot %q[%d]: expected %.8f, got %.8f", plotName, idx, exp, act)
		}
	}
}

func PrintTradeSummary(t *testing.T, result *StrategyResult) {
	t.Helper()

	t.Logf("=== Trade Summary ===")
	t.Logf("Total trades: %d", len(result.Trades))
	t.Logf("Open trades: %d", len(result.OpenTrades))
	t.Logf("Net profit: %.2f", result.NetProfit)
	t.Logf("Final equity: %.2f", result.Equity)

	if len(result.Trades) > 0 {
		t.Logf("First trade: bar %d -> %d, profit %.2f",
			result.Trades[0].EntryBar,
			result.Trades[0].ExitBar,
			result.Trades[0].Profit)

		lastIdx := len(result.Trades) - 1
		t.Logf("Last trade: bar %d -> %d, profit %.2f",
			result.Trades[lastIdx].EntryBar,
			result.Trades[lastIdx].ExitBar,
			result.Trades[lastIdx].Profit)
	}
}

func DescribeResult(result *StrategyResult) string {
	return fmt.Sprintf("trades=%d open=%d profit=%.2f equity=%.2f",
		len(result.Trades),
		len(result.OpenTrades),
		result.NetProfit,
		result.Equity)
}
