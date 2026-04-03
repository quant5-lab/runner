package ticker

import (
	"testing"

	"github.com/quant5-lab/runner/runtime/context"
)

func TestLineBreakTransformer_NilInput(t *testing.T) {
	result := NewLineBreakTransformer(3).Transform(nil)
	if len(result.Bars) != 0 || len(result.MainToSynthetic) != 0 {
		t.Errorf("nil input: expected empty result, got %d bars and %d mapping entries",
			len(result.Bars), len(result.MainToSynthetic))
	}
}

// TestLineBreakTransformer_FirstBarAlwaysSeeds verifies that bar 0 unconditionally
// seeds the first line and that the mapping entry for it is 0.
func TestLineBreakTransformer_FirstBarAlwaysSeeds(t *testing.T) {
	tests := []struct {
		name string
		bar  context.OHLCV
	}{
		{"normal", context.OHLCV{Open: 98, Close: 100}},
		{"zero_price", context.OHLCV{Open: 0, Close: 0}},
		{"flat_ohlc", context.OHLCV{Open: 50, Close: 50}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewLineBreakTransformer(3).Transform([]context.OHLCV{tt.bar})
			if len(result.Bars) != 1 {
				t.Fatalf("expected 1 seed line, got %d", len(result.Bars))
			}
			if result.Bars[0].Close != tt.bar.Close {
				t.Errorf("seed line Close=%v, want %v", result.Bars[0].Close, tt.bar.Close)
			}
			if result.MainToSynthetic[0] != 0 {
				t.Errorf("MainToSynthetic[0]=%d, want 0", result.MainToSynthetic[0])
			}
		})
	}
}

// TestLineBreakTransformer_LineOHLC verifies the OHLC layout for up and down lines:
//
//	up line:   O=prevClose C=barClose H=C L=O
//	down line: O=prevClose C=barClose H=O L=C
func TestLineBreakTransformer_LineOHLC(t *testing.T) {
	// up line: seed at 100, then bar closes at 110.
	upBars := []context.OHLCV{{Close: 100}, {Close: 110}}
	upResult := NewLineBreakTransformer(3).Transform(upBars)
	if len(upResult.Bars) != 2 {
		t.Fatalf("expected 2 lines (seed + up), got %d", len(upResult.Bars))
	}
	ul := upResult.Bars[1]
	if ul.Open != 100 || ul.Close != 110 {
		t.Errorf("up line O=%v C=%v, want O=100 C=110", ul.Open, ul.Close)
	}
	if ul.High != ul.Close {
		t.Errorf("up line High=%v, want Close=%v", ul.High, ul.Close)
	}
	if ul.Low != ul.Open {
		t.Errorf("up line Low=%v, want Open=%v", ul.Low, ul.Open)
	}

	// down line: seed at 100, then bar closes at 90.
	downBars := []context.OHLCV{{Close: 100}, {Close: 90}}
	downResult := NewLineBreakTransformer(3).Transform(downBars)
	if len(downResult.Bars) != 2 {
		t.Fatalf("expected 2 lines (seed + down), got %d", len(downResult.Bars))
	}
	dl := downResult.Bars[1]
	if dl.Open != 100 || dl.Close != 90 {
		t.Errorf("down line O=%v C=%v, want O=100 C=90", dl.Open, dl.Close)
	}
	if dl.High != dl.Open {
		t.Errorf("down line High=%v, want Open=%v", dl.High, dl.Open)
	}
	if dl.Low != dl.Close {
		t.Errorf("down line Low=%v, want Close=%v", dl.Low, dl.Close)
	}
}

// TestLineBreakTransformer_NoLineWithinRange verifies that a price equal to or
// inside [low, high] of the last N lines does not produce a new line.
func TestLineBreakTransformer_NoLineWithinRange(t *testing.T) {
	// After seed(100) and up-line(110), the range is [100,110].
	// Prices 101, 105, 109 are strictly inside and must not form new lines.
	bars := []context.OHLCV{{Close: 100}, {Close: 110}, {Close: 105}, {Close: 101}, {Close: 109}}
	result := NewLineBreakTransformer(3).Transform(bars)

	if len(result.Bars) != 2 {
		t.Errorf("expected 2 lines (seed+up), got %d (price within range should be silent)", len(result.Bars))
	}
	for i := 2; i <= 4; i++ {
		if result.MainToSynthetic[i] != 1 {
			t.Errorf("MainToSynthetic[%d]=%d, want 1 (last formed line index)", i, result.MainToSynthetic[i])
		}
	}
}

// TestLineBreakTransformer_NLineBreakWindow verifies the N-line window capping:
// only the last N line closes are used to determine the break threshold.
func TestLineBreakTransformer_NLineBreakWindow(t *testing.T) {
	tests := []struct {
		name      string
		n         int
		prices    []float64
		wantLines int
	}{
		{
			// N=1: every directional move above or below the single most-recent close forms a line.
			name:      "n1_every_direction_change",
			n:         1,
			prices:    []float64{100, 110, 105, 115, 108},
			wantLines: 5,
		},
		{
			// N=3: window=[100,110,120]; down-break requires price < 100.
			// Price=105 is inside the window → no new line.
			name:      "n3_window_caps_break_threshold",
			n:         3,
			prices:    []float64{100, 110, 120, 105},
			wantLines: 3,
		},
		{
			// N=3: after a 4th line the window slides to [110,120,130].
			// Price=115 remains inside the slid window → no new line.
			name:      "n3_window_slides_after_fourth_line",
			n:         3,
			prices:    []float64{100, 110, 120, 130, 115},
			wantLines: 4,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewLineBreakTransformer(tt.n).Transform(pricesToBars(tt.prices))
			if len(result.Bars) != tt.wantLines {
				t.Errorf("line count = %d, want %d", len(result.Bars), tt.wantLines)
			}
		})
	}
}

// TestLineBreakTransformer_InvalidNFallsBackToDefault verifies that N <= 0 is
// rejected and the default is used instead.
func TestLineBreakTransformer_InvalidNFallsBackToDefault(t *testing.T) {
	for _, n := range []int{0, -1, -100} {
		tr := NewLineBreakTransformer(n)
		if tr.numberOfLines != LineBreakDefaultLines {
			t.Errorf("N=%d → numberOfLines=%d, want default %d", n, tr.numberOfLines, LineBreakDefaultLines)
		}
	}
}

// TestLineBreakTransformer_MappingInvariants verifies structural mapping guarantees.
func TestLineBreakTransformer_MappingInvariants(t *testing.T) {
	tests := []struct {
		name   string
		prices []float64
		n      int
	}{
		{"n3_uptrend", linspace(100, 15, 10), 3},
		{"n3_downtrend", linspace(200, -15, 10), 3},
		{"n1_alternating", oscillate(100, 20, -15, 15), 1},
		{"n2_multi_break", oscillate(100, 25, -10, 20), 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewLineBreakTransformer(tt.n).Transform(pricesToBars(tt.prices))
			assertMappingLength(t, result, len(tt.prices))
			assertMappingNonDecreasing(t, result)
			assertMappingValidIndices(t, result)
			assertOHLCInvariants(t, result.Bars)
		})
	}
}

func TestLineBreakTransformer_Type(t *testing.T) {
	if NewLineBreakTransformer(3).Type() != ModifierLineBreak {
		t.Errorf("Type() = %q, want %q", NewLineBreakTransformer(3).Type(), ModifierLineBreak)
	}
}
