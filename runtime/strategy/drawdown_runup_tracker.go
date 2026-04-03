package strategy

import "math"

type DrawdownRunupTracker struct {
	peakEquity     float64
	troughEquity   float64
	maxDrawdown    float64
	maxRunup       float64
	maxDrawdownPct float64
	maxRunupPct    float64
	initialized    bool
}

func NewDrawdownRunupTracker() *DrawdownRunupTracker {
	return &DrawdownRunupTracker{}
}

func (t *DrawdownRunupTracker) UpdateWithIntrabar(
	equityClose float64,
	equityAdverse float64,
	equityFavorable float64,
) (maxDrawdown, maxRunup, maxDrawdownPct, maxRunupPct float64) {
	if !t.initialized {
		t.peakEquity = equityClose
		t.troughEquity = equityClose
		t.initialized = true
	}

	if equityFavorable > t.peakEquity {
		t.peakEquity = equityFavorable
	}
	if equityAdverse < t.troughEquity {
		t.troughEquity = equityAdverse
	}

	drawdown := t.peakEquity - equityAdverse
	runup := equityFavorable - t.troughEquity

	if drawdown > t.maxDrawdown {
		t.maxDrawdown = drawdown
		if t.peakEquity != 0 {
			t.maxDrawdownPct = drawdown / t.peakEquity * 100
		}
	}

	if runup > t.maxRunup {
		t.maxRunup = runup
		if t.troughEquity != 0 {
			t.maxRunupPct = runup / math.Abs(t.troughEquity) * 100
		}
	}

	return t.maxDrawdown, t.maxRunup, t.maxDrawdownPct, t.maxRunupPct
}
