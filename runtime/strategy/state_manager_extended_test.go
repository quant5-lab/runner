package strategy

import (
	"math"
	"testing"
)

func TestStateManagerAllAccessorsNonNil(t *testing.T) {
	sm := NewStateManager(100)

	accessors := []struct {
		name   string
		series interface{}
	}{
		{"PositionAvgPriceSeries", sm.PositionAvgPriceSeries()},
		{"PositionSizeSeries", sm.PositionSizeSeries()},
		{"EquitySeries", sm.EquitySeries()},
		{"NetProfitSeries", sm.NetProfitSeries()},
		{"ClosedTradesSeries", sm.ClosedTradesSeries()},
		{"InitialCapitalSeries", sm.InitialCapitalSeries()},
		{"GrossProfitSeries", sm.GrossProfitSeries()},
		{"GrossLossSeries", sm.GrossLossSeries()},
		{"WinTradesSeries", sm.WinTradesSeries()},
		{"LossTradesSeries", sm.LossTradesSeries()},
		{"EvenTradesSeries", sm.EvenTradesSeries()},
		{"OpenProfitSeries", sm.OpenProfitSeries()},
		{"OpenTradesSeries", sm.OpenTradesSeries()},
		{"AvgTradeSeries", sm.AvgTradeSeries()},
		{"AvgWinningTradeSeries", sm.AvgWinningTradeSeries()},
		{"AvgLosingTradeSeries", sm.AvgLosingTradeSeries()},
		{"MaxDrawdownSeries", sm.MaxDrawdownSeries()},
		{"MaxRunupSeries", sm.MaxRunupSeries()},
		{"MaxDrawdownPctSeries", sm.MaxDrawdownPctSeries()},
		{"MaxRunupPctSeries", sm.MaxRunupPctSeries()},
	}

	for _, a := range accessors {
		t.Run(a.name, func(t *testing.T) {
			if a.series == nil {
				t.Errorf("%s returned nil", a.name)
			}
		})
	}
}

func TestStateManagerTradeStatSeries(t *testing.T) {
	tests := []struct {
		name             string
		setup            func(*Strategy)
		price            float64
		wantClosedTrades float64
		wantGrossProfit  float64
		wantGrossLoss    float64
		wantWinTrades    float64
		wantLossTrades   float64
		wantEvenTrades   float64
	}{
		{
			name:             "no_trades",
			setup:            func(s *Strategy) {},
			price:            100,
			wantClosedTrades: 0, wantGrossProfit: 0, wantGrossLoss: 0,
			wantWinTrades: 0, wantLossTrades: 0, wantEvenTrades: 0,
		},
		{
			name: "one_winning_trade",
			setup: func(s *Strategy) {
				s.Entry("L", Long, 10, "")
				s.OnBarUpdate(1, 100.0, 1001)
				s.Close("L", 110.0, 1002, "")
				s.OnBarUpdate(2, 110.0, 1002)
			},
			price:            110,
			wantClosedTrades: 1, wantGrossProfit: 100, wantGrossLoss: 0,
			wantWinTrades: 1, wantLossTrades: 0, wantEvenTrades: 0,
		},
		{
			name: "one_losing_trade",
			setup: func(s *Strategy) {
				s.Entry("L", Long, 10, "")
				s.OnBarUpdate(1, 100.0, 1001)
				s.Close("L", 90.0, 1002, "")
				s.OnBarUpdate(2, 90.0, 1002)
			},
			price:            90,
			wantClosedTrades: 1, wantGrossProfit: 0, wantGrossLoss: -100,
			wantWinTrades: 0, wantLossTrades: 1, wantEvenTrades: 0,
		},
		{
			name: "breakeven_trade",
			setup: func(s *Strategy) {
				s.Entry("L", Long, 10, "")
				s.OnBarUpdate(1, 100.0, 1001)
				s.Close("L", 100.0, 1002, "")
				s.OnBarUpdate(2, 100.0, 1002)
			},
			price:            100,
			wantClosedTrades: 1, wantGrossProfit: 0, wantGrossLoss: 0,
			wantWinTrades: 0, wantLossTrades: 0, wantEvenTrades: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm := NewStateManager(100)
			strat := NewStrategy()
			strat.CallWithPyramiding("Test", 10000, 0)
			tt.setup(strat)
			sm.SampleCurrentBar(strat, tt.price, tt.price+5.0, tt.price-5.0)

			if sm.ClosedTradesSeries().Get(0) != tt.wantClosedTrades {
				t.Errorf("ClosedTrades = %.0f, want %.0f", sm.ClosedTradesSeries().Get(0), tt.wantClosedTrades)
			}
			if sm.GrossProfitSeries().Get(0) != tt.wantGrossProfit {
				t.Errorf("GrossProfit = %.2f, want %.2f", sm.GrossProfitSeries().Get(0), tt.wantGrossProfit)
			}
			if sm.GrossLossSeries().Get(0) != tt.wantGrossLoss {
				t.Errorf("GrossLoss = %.2f, want %.2f", sm.GrossLossSeries().Get(0), tt.wantGrossLoss)
			}
			if sm.WinTradesSeries().Get(0) != tt.wantWinTrades {
				t.Errorf("WinTrades = %.0f, want %.0f", sm.WinTradesSeries().Get(0), tt.wantWinTrades)
			}
			if sm.LossTradesSeries().Get(0) != tt.wantLossTrades {
				t.Errorf("LossTrades = %.0f, want %.0f", sm.LossTradesSeries().Get(0), tt.wantLossTrades)
			}
			if sm.EvenTradesSeries().Get(0) != tt.wantEvenTrades {
				t.Errorf("EvenTrades = %.0f, want %.0f", sm.EvenTradesSeries().Get(0), tt.wantEvenTrades)
			}
		})
	}
}

func TestStateManagerAvgTradeSeries(t *testing.T) {
	tests := []struct {
		name             string
		setup            func(*Strategy)
		price            float64
		wantAvgTrade     float64
		wantAvgWinTrade  float64
		wantAvgLossTrade float64
	}{
		{
			name:             "no_closed_trades",
			setup:            func(s *Strategy) {},
			price:            100,
			wantAvgTrade:     0,
			wantAvgWinTrade:  0,
			wantAvgLossTrade: 0,
		},
		{
			name: "wins_and_losses",
			setup: func(s *Strategy) {
				for i, profit := range []float64{100, -50, 200} {
					id := "t" + string(rune('A'+i))
					entry := 100.0
					exit := entry + profit/10.0
					s.Entry(id, Long, 10, "")
					s.OnBarUpdate(i*2+1, entry, int64(1000+i*2))
					s.Close(id, exit, int64(1001+i*2), "")
					s.OnBarUpdate(i*2+2, exit, int64(1001+i*2))
				}
			},
			price:            100,
			wantAvgTrade:     250.0 / 3,
			wantAvgWinTrade:  150,
			wantAvgLossTrade: -50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm := NewStateManager(100)
			strat := NewStrategy()
			strat.CallWithPyramiding("Test", 10000, 5)
			tt.setup(strat)
			sm.SampleCurrentBar(strat, tt.price, tt.price+5.0, tt.price-5.0)

			if sm.AvgTradeSeries().Get(0) != tt.wantAvgTrade {
				t.Errorf("AvgTrade = %.6f, want %.6f", sm.AvgTradeSeries().Get(0), tt.wantAvgTrade)
			}
			if sm.AvgWinningTradeSeries().Get(0) != tt.wantAvgWinTrade {
				t.Errorf("AvgWinningTrade = %.6f, want %.6f", sm.AvgWinningTradeSeries().Get(0), tt.wantAvgWinTrade)
			}
			if sm.AvgLosingTradeSeries().Get(0) != tt.wantAvgLossTrade {
				t.Errorf("AvgLosingTrade = %.6f, want %.6f", sm.AvgLosingTradeSeries().Get(0), tt.wantAvgLossTrade)
			}
		})
	}
}

func TestStateManagerInitialCapitalConstant(t *testing.T) {
	sm := NewStateManager(50)
	strat := NewStrategy()
	strat.CallWithPyramiding("Test", 25000, 0)

	for i := 0; i < 10; i++ {
		price := 100.0 + float64(i)
		sm.SampleCurrentBar(strat, price, price+5.0, price-5.0)
		ic := sm.InitialCapitalSeries().Get(0)
		if ic != 25000 {
			t.Errorf("Bar %d: InitialCapital = %.2f, want 25000", i, ic)
		}
		sm.AdvanceCursors()
	}
}

func TestStateManagerMaxDrawdownNonDecreasing(t *testing.T) {
	sm := NewStateManager(100)
	strat := NewStrategy()
	strat.CallWithPyramiding("Test", 10000, 0)

	strat.Entry("L", Long, 10, "")
	strat.OnBarUpdate(1, 100.0, 1001)

	prices := []float64{100, 105, 100, 90, 95, 110, 85}
	var prevMaxDrawdown float64

	for i, price := range prices {
		sm.SampleCurrentBar(strat, price, price+5.0, price-5.0)
		md := sm.MaxDrawdownSeries().Get(0)

		if md < 0 {
			t.Errorf("Bar %d price=%.0f: MaxDrawdown=%.2f must be >= 0", i, price, md)
		}
		if md < prevMaxDrawdown {
			t.Errorf("Bar %d price=%.0f: MaxDrawdown=%.2f decreased from %.2f", i, price, md, prevMaxDrawdown)
		}
		prevMaxDrawdown = md
		sm.AdvanceCursors()
	}
}

func TestStateManagerMaxRunupNonDecreasing(t *testing.T) {
	sm := NewStateManager(100)
	strat := NewStrategy()
	strat.CallWithPyramiding("Test", 10000, 0)

	strat.Entry("L", Long, 10, "")
	strat.OnBarUpdate(1, 100.0, 1001)

	prices := []float64{100, 90, 95, 110, 100, 120, 115}
	var prevMaxRunup float64

	for i, price := range prices {
		sm.SampleCurrentBar(strat, price, price+5.0, price-5.0)
		mr := sm.MaxRunupSeries().Get(0)

		if mr < 0 {
			t.Errorf("Bar %d price=%.0f: MaxRunup=%.2f must be >= 0", i, price, mr)
		}
		if mr < prevMaxRunup {
			t.Errorf("Bar %d price=%.0f: MaxRunup=%.2f decreased from %.2f", i, price, mr, prevMaxRunup)
		}
		prevMaxRunup = mr
		sm.AdvanceCursors()
	}
}

func TestStateManagerDrawdownPeakTracking(t *testing.T) {
	sm := NewStateManager(100)
	strat := NewStrategy()
	strat.CallWithPyramiding("Test", 10000, 0)

	strat.Entry("L", Long, 10, "")
	strat.OnBarUpdate(1, 100.0, 1001)

	sm.SampleCurrentBar(strat, 105.0, 110.0, 100.0)
	firstDD := sm.MaxDrawdownSeries().Get(0)
	expectedFirstDD := (10000.0 + (110.0-100.0)*10) - (10000.0 + (100.0-100.0)*10)
	if firstDD != expectedFirstDD {
		t.Errorf("First bar intrabar drawdown: got %.2f, want %.2f", firstDD, expectedFirstDD)
	}
	sm.AdvanceCursors()

	sm.SampleCurrentBar(strat, 100.0, 105.0, 95.0)
	peakEquity := 10000.0 + (110.0-100.0)*10
	adverseEquity := 10000.0 + (95.0-100.0)*10
	expectedDD := peakEquity - adverseEquity
	if sm.MaxDrawdownSeries().Get(0) != expectedDD {
		t.Errorf("Drawdown from peak: got %.2f, want %.2f", sm.MaxDrawdownSeries().Get(0), expectedDD)
	}
	sm.AdvanceCursors()

	sm.SampleCurrentBar(strat, 90.0, 95.0, 85.0)
	adverseEquity = 10000.0 + (85.0-100.0)*10
	expectedDD = peakEquity - adverseEquity
	if sm.MaxDrawdownSeries().Get(0) != expectedDD {
		t.Errorf("Deeper drawdown: got %.2f, want %.2f", sm.MaxDrawdownSeries().Get(0), expectedDD)
	}
	sm.AdvanceCursors()

	sm.SampleCurrentBar(strat, 110.0, 115.0, 105.0)
	if sm.MaxDrawdownSeries().Get(0) != expectedDD {
		t.Errorf("MaxDrawdown must persist after recovery: got %.2f, want %.2f", sm.MaxDrawdownSeries().Get(0), expectedDD)
	}
}

func TestStateManagerRunupTroughTracking(t *testing.T) {
	sm := NewStateManager(100)
	strat := NewStrategy()
	strat.CallWithPyramiding("Test", 10000, 0)

	strat.Entry("L", Long, 10, "")
	strat.OnBarUpdate(1, 100.0, 1001)

	sm.SampleCurrentBar(strat, 95.0, 100.0, 90.0)
	firstRU := sm.MaxRunupSeries().Get(0)
	troughEquity := 10000.0 + (90.0-100.0)*10
	favorableEquity := 10000.0 + (100.0-100.0)*10
	expectedFirstRU := favorableEquity - troughEquity
	if firstRU != expectedFirstRU {
		t.Errorf("First bar intrabar runup: got %.2f, want %.2f", firstRU, expectedFirstRU)
	}
	sm.AdvanceCursors()

	sm.SampleCurrentBar(strat, 100.0, 105.0, 95.0)
	favorableEquity = 10000.0 + (105.0-100.0)*10
	expectedRU := favorableEquity - troughEquity
	if sm.MaxRunupSeries().Get(0) != expectedRU {
		t.Errorf("Runup from trough: got %.2f, want %.2f", sm.MaxRunupSeries().Get(0), expectedRU)
	}
	sm.AdvanceCursors()

	sm.SampleCurrentBar(strat, 110.0, 115.0, 105.0)
	favorableEquity = 10000.0 + (115.0-100.0)*10
	expectedRU = favorableEquity - troughEquity
	if sm.MaxRunupSeries().Get(0) != expectedRU {
		t.Errorf("Larger runup: got %.2f, want %.2f", sm.MaxRunupSeries().Get(0), expectedRU)
	}
	sm.AdvanceCursors()

	sm.SampleCurrentBar(strat, 90.0, 95.0, 85.0)
	if sm.MaxRunupSeries().Get(0) != expectedRU {
		t.Errorf("MaxRunup must persist after decline: got %.2f, want %.2f", sm.MaxRunupSeries().Get(0), expectedRU)
	}
}

func TestStateManagerMaxDrawdownPercent(t *testing.T) {
	sm := NewStateManager(100)
	strat := NewStrategy()
	strat.CallWithPyramiding("Test", 10000, 0)

	strat.Entry("L", Long, 10, "")
	strat.OnBarUpdate(1, 100.0, 1001)

	sm.SampleCurrentBar(strat, 105.0, 110.0, 100.0)
	sm.AdvanceCursors()

	sm.SampleCurrentBar(strat, 95.0, 100.0, 90.0)
	peakEquity := 10000.0 + (110.0-100.0)*10
	adverseEquity := 10000.0 + (90.0-100.0)*10
	expectedDrawdown := peakEquity - adverseEquity
	expectedPct := expectedDrawdown / peakEquity * 100.0
	got := sm.MaxDrawdownPctSeries().Get(0)
	if math.Abs(got-expectedPct) > 0.001 {
		t.Errorf("MaxDrawdownPct = %.6f, want %.6f", got, expectedPct)
	}
}

func TestStateManagerOpenPositionSeries(t *testing.T) {
	tests := []struct {
		name           string
		setup          func(*Strategy)
		price          float64
		wantOpenProfit float64
		wantOpenTrades float64
	}{
		{
			name:           "flat",
			setup:          func(s *Strategy) {},
			price:          110,
			wantOpenProfit: 0,
			wantOpenTrades: 0,
		},
		{
			name: "single_long_in_profit",
			setup: func(s *Strategy) {
				s.Entry("L", Long, 10, "")
				s.OnBarUpdate(1, 100.0, 1001)
			},
			price:          110,
			wantOpenProfit: 100,
			wantOpenTrades: 1,
		},
		{
			name: "single_short_in_profit",
			setup: func(s *Strategy) {
				s.Entry("S", Short, 5, "")
				s.OnBarUpdate(1, 100.0, 1001)
			},
			price:          90,
			wantOpenProfit: 50,
			wantOpenTrades: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm := NewStateManager(100)
			strat := NewStrategy()
			strat.CallWithPyramiding("Test", 10000, 5)
			tt.setup(strat)
			sm.SampleCurrentBar(strat, tt.price, tt.price+5.0, tt.price-5.0)

			if sm.OpenProfitSeries().Get(0) != tt.wantOpenProfit {
				t.Errorf("OpenProfit = %.2f, want %.2f", sm.OpenProfitSeries().Get(0), tt.wantOpenProfit)
			}
			if sm.OpenTradesSeries().Get(0) != tt.wantOpenTrades {
				t.Errorf("OpenTrades = %.0f, want %.0f", sm.OpenTradesSeries().Get(0), tt.wantOpenTrades)
			}
		})
	}
}

func TestStateManagerHistoricalAccessExtendedSeries(t *testing.T) {
	sm := NewStateManager(100)
	strat := NewStrategy()
	strat.CallWithPyramiding("Test", 10000, 0)

	strat.OnBarUpdate(0, 100.0, 1000)
	sm.SampleCurrentBar(strat, 100.0, 105.0, 95.0)
	sm.AdvanceCursors()

	strat.Entry("L", Long, 10, "")
	strat.OnBarUpdate(1, 110.0, 1001)
	sm.SampleCurrentBar(strat, 110.0, 115.0, 105.0)

	if sm.GrossProfitSeries().Get(0) != 0 {
		t.Errorf("Bar1[0] GrossProfit = %.2f, want 0", sm.GrossProfitSeries().Get(0))
	}
	if sm.GrossLossSeries().Get(1) != 0 {
		t.Errorf("Bar0[1] GrossLoss = %.2f, want 0", sm.GrossLossSeries().Get(1))
	}
	if sm.InitialCapitalSeries().Get(0) != sm.InitialCapitalSeries().Get(1) {
		t.Errorf("InitialCapital changed across bars")
	}
}

func TestStateManagerAllSeriesAdvanceUniformly(t *testing.T) {
	sm := NewStateManager(50)
	strat := NewStrategy()
	strat.CallWithPyramiding("Test", 10000, 0)

	strat.Entry("L", Long, 10, "")
	strat.OnBarUpdate(1, 100.0, 1001)

	prices := []float64{100, 105, 95, 110, 100}
	for _, price := range prices {
		sm.SampleCurrentBar(strat, price, price+5.0, price-5.0)
		sm.AdvanceCursors()
	}
	sm.SampleCurrentBar(strat, 108.0, 113.0, 103.0)

	equityBar0 := sm.EquitySeries().Get(5)
	if equityBar0 != 10000 {
		t.Errorf("EquitySeries[5] (bar 0) = %.2f, want 10000", equityBar0)
	}

	icBar0 := sm.InitialCapitalSeries().Get(5)
	if icBar0 != 10000 {
		t.Errorf("InitialCapitalSeries[5] (bar 0) = %.2f, want 10000", icBar0)
	}
}
