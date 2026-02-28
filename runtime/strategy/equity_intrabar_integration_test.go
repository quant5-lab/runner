package strategy

import (
	"testing"
)

func TestIntrabarEquityExtremes_DrawdownRunupParity(t *testing.T) {
	tests := []struct {
		name                string
		direction           string
		initialCapital      float64
		entryPrice          float64
		size                float64
		bar1Close           float64
		bar1High            float64
		bar1Low             float64
		bar2Close           float64
		bar2High            float64
		bar2Low             float64
		wantDrawdownGreater bool
		wantRunupGreater    bool
	}{
		{
			name:                "long_position_volatile_swing",
			direction:           Long,
			initialCapital:      10000,
			entryPrice:          100,
			size:                10,
			bar1Close:           100,
			bar1High:            105,
			bar1Low:             95,
			bar2Close:           95,
			bar2High:            100,
			bar2Low:             90,
			wantDrawdownGreater: true,
			wantRunupGreater:    false,
		},
		{
			name:                "long_position_uptrend_swing",
			direction:           Long,
			initialCapital:      10000,
			entryPrice:          100,
			size:                10,
			bar1Close:           100,
			bar1High:            105,
			bar1Low:             95,
			bar2Close:           105,
			bar2High:            110,
			bar2Low:             100,
			wantDrawdownGreater: false,
			wantRunupGreater:    true,
		},
		{
			name:                "short_position_volatile_swing",
			direction:           Short,
			initialCapital:      10000,
			entryPrice:          100,
			size:                10,
			bar1Close:           100,
			bar1High:            105,
			bar1Low:             95,
			bar2Close:           105,
			bar2High:            110,
			bar2Low:             100,
			wantDrawdownGreater: true,
			wantRunupGreater:    false,
		},
		{
			name:                "short_position_downtrend_swing",
			direction:           Short,
			initialCapital:      10000,
			entryPrice:          100,
			size:                10,
			bar1Close:           100,
			bar1High:            105,
			bar1Low:             95,
			bar2Close:           95,
			bar2High:            100,
			bar2Low:             90,
			wantDrawdownGreater: false,
			wantRunupGreater:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm := NewStateManager(100)
			strat := NewStrategy()
			strat.CallWithPyramiding("Test", tt.initialCapital, 1)

			strat.Entry("Trade", tt.direction, tt.size, "")
			strat.OnBarUpdate(1, tt.entryPrice, 1001)

			sm.SampleCurrentBar(strat, tt.bar1Close, tt.bar1High, tt.bar1Low)
			sm.AdvanceCursors()

			sm.SampleCurrentBar(strat, tt.bar2Close, tt.bar2High, tt.bar2Low)

			actualDrawdown := sm.MaxDrawdownSeries().Get(0)
			actualRunup := sm.MaxRunupSeries().Get(0)

			closeOnlyDrawdown := 0.0
			closeOnlyRunup := 0.0

			if tt.wantDrawdownGreater && actualDrawdown <= closeOnlyDrawdown {
				t.Errorf("expected intrabar drawdown (%.2f) > close-only (%.2f)",
					actualDrawdown, closeOnlyDrawdown)
			}

			if tt.wantRunupGreater && actualRunup <= closeOnlyRunup {
				t.Errorf("expected intrabar runup (%.2f) > close-only (%.2f)",
					actualRunup, closeOnlyRunup)
			}
		})
	}
}

func TestIntrabarEquityExtremes_PercentageConsistency(t *testing.T) {
	sm := NewStateManager(100)
	strat := NewStrategy()
	strat.CallWithPyramiding("Test", 10000, 1)

	strat.Entry("Long", Long, 10, "")
	strat.OnBarUpdate(1, 100.0, 1001)

	sm.SampleCurrentBar(strat, 100.0, 105.0, 95.0)
	sm.AdvanceCursors()

	sm.SampleCurrentBar(strat, 95.0, 100.0, 90.0)

	drawdown := sm.MaxDrawdownSeries().Get(0)
	drawdownPct := sm.MaxDrawdownPctSeries().Get(0)

	peakEquity := 10000.0 + (105.0-100.0)*10
	expectedPct := drawdown / peakEquity * 100.0

	if drawdownPct != expectedPct {
		t.Errorf("drawdown percent = %.6f, want %.6f (from drawdown %.2f / peak %.2f)",
			drawdownPct, expectedPct, drawdown, peakEquity)
	}
}

func TestIntrabarEquityExtremes_Monotonicity(t *testing.T) {
	sm := NewStateManager(100)
	strat := NewStrategy()
	strat.CallWithPyramiding("Test", 10000, 1)

	strat.Entry("Long", Long, 10, "")
	strat.OnBarUpdate(1, 100.0, 1001)

	prices := []struct {
		close float64
		high  float64
		low   float64
	}{
		{100, 105, 95},
		{95, 100, 90},
		{90, 95, 85},
		{95, 100, 90},
	}

	prevDrawdown := 0.0
	prevRunup := 0.0

	for i, p := range prices {
		sm.SampleCurrentBar(strat, p.close, p.high, p.low)

		currentDrawdown := sm.MaxDrawdownSeries().Get(0)
		currentRunup := sm.MaxRunupSeries().Get(0)

		if currentDrawdown < prevDrawdown {
			t.Errorf("bar %d: drawdown decreased from %.2f to %.2f (must be monotonic non-decreasing)",
				i, prevDrawdown, currentDrawdown)
		}

		if currentRunup < prevRunup {
			t.Errorf("bar %d: runup decreased from %.2f to %.2f (must be monotonic non-decreasing)",
				i, prevRunup, currentRunup)
		}

		prevDrawdown = currentDrawdown
		prevRunup = currentRunup

		if i < len(prices)-1 {
			sm.AdvanceCursors()
		}
	}
}

func TestIntrabarEquityExtremes_FlatPositionInvariance(t *testing.T) {
	sm := NewStateManager(100)
	strat := NewStrategy()
	strat.CallWithPyramiding("Test", 10000, 1)

	prices := []struct {
		close float64
		high  float64
		low   float64
	}{
		{100, 110, 90},
		{105, 115, 95},
		{95, 105, 85},
	}

	for i, p := range prices {
		sm.SampleCurrentBar(strat, p.close, p.high, p.low)

		drawdown := sm.MaxDrawdownSeries().Get(0)
		runup := sm.MaxRunupSeries().Get(0)

		if drawdown != 0 {
			t.Errorf("bar %d: flat position should have zero drawdown, got %.2f", i, drawdown)
		}

		if runup != 0 {
			t.Errorf("bar %d: flat position should have zero runup, got %.2f", i, runup)
		}

		if i < len(prices)-1 {
			sm.AdvanceCursors()
		}
	}
}
