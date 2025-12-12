package strategy

import (
	"math"
	"testing"
)

func TestStateManagerInitialization(t *testing.T) {
	sm := NewStateManager(100)

	if sm.PositionAvgPriceSeries() == nil {
		t.Error("PositionAvgPriceSeries should be initialized")
	}
	if sm.PositionSizeSeries() == nil {
		t.Error("PositionSizeSeries should be initialized")
	}
	if sm.EquitySeries() == nil {
		t.Error("EquitySeries should be initialized")
	}
	if sm.NetProfitSeries() == nil {
		t.Error("NetProfitSeries should be initialized")
	}
	if sm.ClosedTradesSeries() == nil {
		t.Error("ClosedTradesSeries should be initialized")
	}
}

func TestStateManagerSamplesAllFields(t *testing.T) {
	sm := NewStateManager(100)
	strat := NewStrategy()
	strat.Call("Test", 10000)

	sm.SampleCurrentBar(strat, 100.0)

	if !math.IsNaN(sm.PositionAvgPriceSeries().Get(0)) {
		t.Errorf("Expected NaN for position_avg_price with no position, got %.2f", sm.PositionAvgPriceSeries().Get(0))
	}
	if sm.PositionSizeSeries().Get(0) != 0 {
		t.Errorf("Expected 0 for position_size, got %.2f", sm.PositionSizeSeries().Get(0))
	}
	if sm.EquitySeries().Get(0) != 10000 {
		t.Errorf("Expected 10000 for equity, got %.2f", sm.EquitySeries().Get(0))
	}
	if sm.NetProfitSeries().Get(0) != 0 {
		t.Errorf("Expected 0 for net_profit, got %.2f", sm.NetProfitSeries().Get(0))
	}
	if sm.ClosedTradesSeries().Get(0) != 0 {
		t.Errorf("Expected 0 for closed_trades, got %.0f", sm.ClosedTradesSeries().Get(0))
	}
}

func TestStateManagerLongPosition(t *testing.T) {
	sm := NewStateManager(100)
	strat := NewStrategy()
	strat.Call("Test", 10000)

	strat.Entry("Long", Long, 10)
	strat.OnBarUpdate(1, 105.0, 1001)
	sm.SampleCurrentBar(strat, 105.0)

	if sm.PositionAvgPriceSeries().Get(0) != 105.0 {
		t.Errorf("Expected avg price 105.0, got %.2f", sm.PositionAvgPriceSeries().Get(0))
	}
	if sm.PositionSizeSeries().Get(0) != 10.0 {
		t.Errorf("Expected size 10.0, got %.2f", sm.PositionSizeSeries().Get(0))
	}
}

func TestStateManagerShortPosition(t *testing.T) {
	sm := NewStateManager(100)
	strat := NewStrategy()
	strat.Call("Test", 10000)

	strat.Entry("Short", Short, 5)
	strat.OnBarUpdate(1, 100.0, 1001)
	sm.SampleCurrentBar(strat, 100.0)

	if sm.PositionAvgPriceSeries().Get(0) != 100.0 {
		t.Errorf("Expected avg price 100.0, got %.2f", sm.PositionAvgPriceSeries().Get(0))
	}
	if sm.PositionSizeSeries().Get(0) != -5.0 {
		t.Errorf("Expected size -5.0, got %.2f", sm.PositionSizeSeries().Get(0))
	}
}

func TestStateManagerHistoricalAccess(t *testing.T) {
	sm := NewStateManager(100)
	strat := NewStrategy()
	strat.Call("Test", 10000)

	strat.OnBarUpdate(0, 100.0, 1000)
	sm.SampleCurrentBar(strat, 100.0)
	sm.AdvanceCursors()

	strat.Entry("Long", Long, 10)
	strat.OnBarUpdate(1, 105.0, 1001)
	sm.SampleCurrentBar(strat, 105.0)

	if sm.PositionAvgPriceSeries().Get(0) != 105.0 {
		t.Errorf("Expected current avg price 105.0, got %.2f", sm.PositionAvgPriceSeries().Get(0))
	}
	if !math.IsNaN(sm.PositionAvgPriceSeries().Get(1)) {
		t.Errorf("Expected historical avg price [1] to be NaN, got %.2f", sm.PositionAvgPriceSeries().Get(1))
	}
}

func TestStateManagerPositionLifecycle(t *testing.T) {
	sm := NewStateManager(100)
	strat := NewStrategy()
	strat.Call("Test", 10000)

	strat.OnBarUpdate(0, 100.0, 1000)
	sm.SampleCurrentBar(strat, 100.0)
	if !math.IsNaN(sm.PositionAvgPriceSeries().Get(0)) {
		t.Error("Bar 0: Expected NaN when flat")
	}
	sm.AdvanceCursors()

	strat.Entry("Long", Long, 10)
	strat.OnBarUpdate(1, 105.0, 1001)
	sm.SampleCurrentBar(strat, 105.0)
	if sm.PositionSizeSeries().Get(0) != 10.0 {
		t.Error("Bar 1: Expected long position size 10")
	}
	sm.AdvanceCursors()

	strat.Close("Long", 110.0, 1002)
	strat.OnBarUpdate(2, 110.0, 1002)
	sm.SampleCurrentBar(strat, 110.0)
	if !math.IsNaN(sm.PositionAvgPriceSeries().Get(0)) {
		t.Error("Bar 2: Expected NaN when flat after close")
	}
	if sm.ClosedTradesSeries().Get(0) != 1 {
		t.Errorf("Bar 2: Expected 1 closed trade, got %.0f", sm.ClosedTradesSeries().Get(0))
	}
}

func TestStateManagerPositionReversal(t *testing.T) {
	sm := NewStateManager(100)
	strat := NewStrategy()
	strat.Call("Test", 10000)

	strat.Entry("Long", Long, 10)
	strat.OnBarUpdate(1, 100.0, 1001)
	sm.SampleCurrentBar(strat, 100.0)
	if sm.PositionSizeSeries().Get(0) != 10.0 {
		t.Error("Expected long position")
	}
	sm.AdvanceCursors()

	strat.Close("Long", 105.0, 1002)
	strat.Entry("Short", Short, 5)
	strat.OnBarUpdate(2, 105.0, 1002)
	strat.OnBarUpdate(3, 105.0, 1003)
	sm.SampleCurrentBar(strat, 105.0)
	if sm.PositionSizeSeries().Get(0) != -5.0 {
		t.Errorf("Expected short position size -5, got %.2f", sm.PositionSizeSeries().Get(0))
	}
}

func TestStateManagerEquityWithUnrealizedPL(t *testing.T) {
	sm := NewStateManager(100)
	strat := NewStrategy()
	strat.Call("Test", 10000)

	strat.Entry("Long", Long, 10)
	strat.OnBarUpdate(1, 100.0, 1001)
	sm.SampleCurrentBar(strat, 100.0)

	initialEquity := sm.EquitySeries().Get(0)
	if initialEquity != 10000 {
		t.Errorf("Expected equity 10000, got %.2f", initialEquity)
	}
	sm.AdvanceCursors()

	sm.SampleCurrentBar(strat, 110.0)
	equityWithProfit := sm.EquitySeries().Get(0)
	expectedEquity := 10000.0 + 100.0
	if equityWithProfit != expectedEquity {
		t.Errorf("Expected equity %.2f with unrealized profit, got %.2f", expectedEquity, equityWithProfit)
	}
}

func TestStateManagerMultipleClosedTrades(t *testing.T) {
	sm := NewStateManager(100)
	strat := NewStrategy()
	strat.Call("Test", 10000)

	barIndex := 0
	for i := 0; i < 3; i++ {
		tradeID := "trade" + string(rune('A'+i))

		strat.Entry(tradeID, Long, 10)
		barIndex++
		strat.OnBarUpdate(barIndex, 100.0, int64(1000+barIndex))

		barIndex++
		strat.Close(tradeID, 105.0, int64(1000+barIndex))
	}

	sm.SampleCurrentBar(strat, 105.0)

	if sm.ClosedTradesSeries().Get(0) != 3 {
		t.Errorf("Expected 3 closed trades, got %.0f", sm.ClosedTradesSeries().Get(0))
	}

	expectedProfit := 3 * 50.0
	if sm.NetProfitSeries().Get(0) != expectedProfit {
		t.Errorf("Expected net profit %.2f, got %.2f", expectedProfit, sm.NetProfitSeries().Get(0))
	}
}

func TestStateManagerNaNPropagation(t *testing.T) {
	sm := NewStateManager(100)
	strat := NewStrategy()
	strat.Call("Test", 10000)

	for i := 0; i < 5; i++ {
		sm.SampleCurrentBar(strat, 100.0)
		if !math.IsNaN(sm.PositionAvgPriceSeries().Get(0)) {
			t.Errorf("Bar %d: Expected NaN when no position", i)
		}
		sm.AdvanceCursors()
	}
}

func TestStateManagerCursorAdvancement(t *testing.T) {
	sm := NewStateManager(10)
	strat := NewStrategy()
	strat.Call("Test", 10000)

	values := []float64{100, 105, 110, 115, 120}

	for i, price := range values {
		if i == 2 {
			strat.Entry("Long", Long, 10)
			strat.OnBarUpdate(i, price, int64(1000+i))
		}
		sm.SampleCurrentBar(strat, price)
		if i < len(values)-1 {
			sm.AdvanceCursors()
		}
	}

	for offset := 0; offset < len(values); offset++ {
		val := sm.PositionAvgPriceSeries().Get(offset)
		if offset <= 2 {
			if val != 110.0 {
				t.Errorf("Offset %d: Expected 110.0, got %.2f", offset, val)
			}
		} else {
			if !math.IsNaN(val) {
				t.Errorf("Offset %d: Expected NaN, got %.2f", offset, val)
			}
		}
	}
}
