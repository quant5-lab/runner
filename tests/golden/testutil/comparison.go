package testutil

import (
	"fmt"
	"math"
	"sort"
)

const (
	priceTolerance = 0.01
	plotTolerance  = 1e-6
)

func floatWithin(a, b, tolerance float64) bool {
	if math.IsNaN(a) && math.IsNaN(b) {
		return true
	}
	if math.IsNaN(a) || math.IsNaN(b) {
		return false
	}
	if a == b {
		return true
	}
	return math.Abs(a-b) <= tolerance
}

func CompareResults(expected, actual *StrategyResult) error {
	if err := compareTradeSlice("trades", expected.Trades, actual.Trades); err != nil {
		return err
	}
	if err := compareTradeSlice("openTrades", expected.OpenTrades, actual.OpenTrades); err != nil {
		return err
	}
	if expected.TotalTrades != actual.TotalTrades {
		return fmt.Errorf("totalTrades: expected %d, got %d", expected.TotalTrades, actual.TotalTrades)
	}
	if !floatWithin(expected.Equity, actual.Equity, priceTolerance) {
		return fmt.Errorf("equity: expected %.2f, got %.2f", expected.Equity, actual.Equity)
	}
	if !floatWithin(expected.NetProfit, actual.NetProfit, priceTolerance) {
		return fmt.Errorf("netProfit: expected %.2f, got %.2f", expected.NetProfit, actual.NetProfit)
	}
	return comparePlots(expected.Plots, actual.Plots)
}

func ResultsEqual(a, b *StrategyResult) bool {
	return CompareResults(a, b) == nil
}

func compareTradeSlice(field string, expected, actual []Trade) error {
	if len(expected) != len(actual) {
		return fmt.Errorf("%s count: expected %d, got %d", field, len(expected), len(actual))
	}
	for i := range expected {
		if err := compareTrade(field, i, &expected[i], &actual[i]); err != nil {
			return err
		}
	}
	return nil
}

func compareTrade(field string, index int, expected, actual *Trade) error {
	if expected.EntryID != actual.EntryID {
		return fmt.Errorf("%s[%d].entryId: expected %q, got %q", field, index, expected.EntryID, actual.EntryID)
	}
	if expected.EntryBar != actual.EntryBar {
		return fmt.Errorf("%s[%d].entryBar: expected %d, got %d", field, index, expected.EntryBar, actual.EntryBar)
	}
	if expected.EntryTime != actual.EntryTime {
		return fmt.Errorf("%s[%d].entryTime: expected %d, got %d", field, index, expected.EntryTime, actual.EntryTime)
	}
	if !floatWithin(expected.EntryPrice, actual.EntryPrice, priceTolerance) {
		return fmt.Errorf("%s[%d].entryPrice: expected %.2f, got %.2f", field, index, expected.EntryPrice, actual.EntryPrice)
	}
	if expected.EntryComment != actual.EntryComment {
		return fmt.Errorf("%s[%d].entryComment: expected %q, got %q", field, index, expected.EntryComment, actual.EntryComment)
	}
	if expected.ExitBar != actual.ExitBar {
		return fmt.Errorf("%s[%d].exitBar: expected %d, got %d", field, index, expected.ExitBar, actual.ExitBar)
	}
	if expected.ExitTime != actual.ExitTime {
		return fmt.Errorf("%s[%d].exitTime: expected %d, got %d", field, index, expected.ExitTime, actual.ExitTime)
	}
	if !floatWithin(expected.ExitPrice, actual.ExitPrice, priceTolerance) {
		return fmt.Errorf("%s[%d].exitPrice: expected %.2f, got %.2f", field, index, expected.ExitPrice, actual.ExitPrice)
	}
	if expected.ExitComment != actual.ExitComment {
		return fmt.Errorf("%s[%d].exitComment: expected %q, got %q", field, index, expected.ExitComment, actual.ExitComment)
	}
	if !floatWithin(expected.Size, actual.Size, priceTolerance) {
		return fmt.Errorf("%s[%d].size: expected %.4f, got %.4f", field, index, expected.Size, actual.Size)
	}
	if !floatWithin(expected.Profit, actual.Profit, priceTolerance) {
		return fmt.Errorf("%s[%d].profit: expected %.2f, got %.2f", field, index, expected.Profit, actual.Profit)
	}
	if expected.Direction != actual.Direction {
		return fmt.Errorf("%s[%d].direction: expected %q, got %q", field, index, expected.Direction, actual.Direction)
	}
	return nil
}

func comparePlots(expected, actual map[string][]float64) error {
	if len(expected) != len(actual) {
		return fmt.Errorf("plots: expected %d series, got %d", len(expected), len(actual))
	}
	keys := make([]string, 0, len(expected))
	for k := range expected {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, name := range keys {
		actValues, ok := actual[name]
		if !ok {
			return fmt.Errorf("plots: series %q missing in actual", name)
		}
		expValues := expected[name]
		if len(expValues) != len(actValues) {
			return fmt.Errorf("plots[%q]: expected %d values, got %d", name, len(expValues), len(actValues))
		}
		for i, ev := range expValues {
			if !floatWithin(ev, actValues[i], plotTolerance) {
				return fmt.Errorf("plots[%q][%d]: expected %.8f, got %.8f", name, i, ev, actValues[i])
			}
		}
	}
	return nil
}
