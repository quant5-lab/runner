package security

import (
	"testing"

	"github.com/quant5-lab/runner/runtime/context"
)

/* TestSARAlgorithm_TrendDetection verifies SAR correctly classifies uptrend vs
 * downtrend based on price movement direction, which determines whether SAR
 * value sits below (uptrend) or above (downtrend) price for stop placement
 */
func TestSARAlgorithm_TrendDetection(t *testing.T) {
	tests := []struct {
		name          string
		data          []context.OHLCV
		barToCheck    int
		expectUptrend bool
	}{
		{
			name: "uptrend",
			data: []context.OHLCV{
				{High: 100.0, Low: 95.0},
				{High: 105.0, Low: 100.0},
				{High: 110.0, Low: 105.0},
				{High: 115.0, Low: 110.0},
				{High: 120.0, Low: 115.0},
			},
			barToCheck:    4,
			expectUptrend: true,
		},
		{
			name: "downtrend",
			data: []context.OHLCV{
				{High: 120.0, Low: 115.0},
				{High: 115.0, Low: 110.0},
				{High: 110.0, Low: 105.0},
				{High: 105.0, Low: 100.0},
				{High: 100.0, Low: 95.0},
			},
			barToCheck:    4,
			expectUptrend: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &context.Context{Data: tt.data}
			mgr := NewSARStateManager("sar_"+tt.name, 0.02, 0.02, 0.20, len(tt.data))

			_, err := mgr.ComputeAtBar(ctx, nil, tt.barToCheck)
			if err != nil {
				t.Fatalf("ComputeAtBar: %v", err)
			}

			if mgr.isUptrend != tt.expectUptrend {
				t.Errorf("trend mismatch: got %v, want %v", mgr.isUptrend, tt.expectUptrend)
			}
		})
	}
}

/* TestSARAlgorithm_AccelerationFactorBounds verifies AF increments on new
 * extreme points but never exceeds maxAF, preventing excessive SAR sensitivity
 * that causes premature stop triggers in trending markets
 */
func TestSARAlgorithm_AccelerationFactorBounds(t *testing.T) {
	data := []context.OHLCV{
		{High: 100.0, Low: 95.0},
		{High: 105.0, Low: 100.0},
		{High: 112.0, Low: 107.0},
		{High: 120.0, Low: 115.0},
		{High: 130.0, Low: 125.0},
	}
	ctx := &context.Context{Data: data}

	start := 0.02
	inc := 0.02
	maxAF := 0.20

	mgr := NewSARStateManager("sar_af", start, inc, maxAF, len(data))

	_, err := mgr.ComputeAtBar(ctx, nil, 1)
	if err != nil {
		t.Fatalf("bar 1: %v", err)
	}
	initialAF := mgr.af

	_, err = mgr.ComputeAtBar(ctx, nil, 4)
	if err != nil {
		t.Fatalf("bar 4: %v", err)
	}

	if mgr.af <= initialAF {
		t.Errorf("AF should increase with new extremes: initial=%.4f, final=%.4f", initialAF, mgr.af)
	}

	if mgr.af > maxAF {
		t.Errorf("AF exceeded max: %.4f > %.4f", mgr.af, maxAF)
	}
}

/* TestSARAlgorithm_ExtremePointTracking verifies EP tracks the highest high in
 * uptrends (or lowest low in downtrends), ensuring SAR acceleration responds to
 * actual trend strength rather than stale extremes
 */
func TestSARAlgorithm_ExtremePointTracking(t *testing.T) {
	data := []context.OHLCV{
		{High: 100.0, Low: 95.0},
		{High: 110.0, Low: 100.0},
		{High: 115.0, Low: 105.0},
		{High: 112.0, Low: 107.0},
		{High: 120.0, Low: 110.0},
	}
	ctx := &context.Context{Data: data}

	mgr := NewSARStateManager("sar_ep", 0.02, 0.02, 0.20, len(data))

	_, err := mgr.ComputeAtBar(ctx, nil, 4)
	if err != nil {
		t.Fatalf("ComputeAtBar: %v", err)
	}

	if mgr.ep != 120.0 {
		t.Errorf("EP should track highest high in uptrend: expected 120.00, got %.2f", mgr.ep)
	}
}

/* TestSARAlgorithm_TrendReversal verifies SAR flips trend classification when
 * price crosses SAR value, triggering trend-following stop reversal and
 * resetting AF to prevent whipsaw in choppy markets
 */
func TestSARAlgorithm_TrendReversal(t *testing.T) {
	data := []context.OHLCV{
		{High: 100.0, Low: 95.0},
		{High: 105.0, Low: 100.0},
		{High: 110.0, Low: 105.0},
		{High: 115.0, Low: 110.0},
		{High: 100.0, Low: 95.0},
		{High: 95.0, Low: 90.0},
	}
	ctx := &context.Context{Data: data}

	mgr := NewSARStateManager("sar_reversal", 0.02, 0.02, 0.20, len(data))

	_, err := mgr.ComputeAtBar(ctx, nil, 3)
	if err != nil {
		t.Fatalf("before reversal: %v", err)
	}
	wasUptrend := mgr.isUptrend

	_, err = mgr.ComputeAtBar(ctx, nil, 5)
	if err != nil {
		t.Fatalf("after reversal: %v", err)
	}

	if wasUptrend == mgr.isUptrend {
		t.Error("expected trend reversal after sharp decline")
	}
}
