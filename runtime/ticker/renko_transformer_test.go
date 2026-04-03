package ticker

import (
	"testing"

	"github.com/quant5-lab/runner/runtime/context"
)

func TestRenkoTransformer_NilInput(t *testing.T) {
	result := NewRenkoTransformer(RenkoStyleATR, 10).Transform(nil)
	if len(result.Bars) != 0 || len(result.MainToSynthetic) != 0 {
		t.Errorf("nil input: expected empty result, got %d bars and %d mapping entries",
			len(result.Bars), len(result.MainToSynthetic))
	}
}

// TestRenkoTransformer_BrickCount verifies how many bricks form for various price sequences.
func TestRenkoTransformer_BrickCount(t *testing.T) {
	tests := []struct {
		name           string
		prices         []float64
		boxSize        float64
		expectedBricks int
	}{
		{
			name:           "price stays below box threshold",
			prices:         []float64{100, 105, 108},
			boxSize:        20,
			expectedBricks: 0,
		},
		{
			name:           "single up brick at exact box boundary",
			prices:         []float64{100, 110},
			boxSize:        10,
			expectedBricks: 1,
		},
		{
			name:           "single down brick",
			prices:         []float64{100, 90},
			boxSize:        10,
			expectedBricks: 1,
		},
		{
			name:           "large up move forms multiple bricks in one bar",
			prices:         []float64{100, 140},
			boxSize:        10,
			expectedBricks: 4,
		},
		{
			name:           "direction reversal (up bricks then down bricks)",
			prices:         []float64{100, 120, 100},
			boxSize:        10,
			expectedBricks: 4,
		},
		{
			name:           "single bar never forms a brick (sets the level only)",
			prices:         []float64{200},
			boxSize:        10,
			expectedBricks: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bars := pricesToBars(tt.prices)
			result := NewRenkoTransformer(RenkoStyleATR, tt.boxSize).Transform(bars)
			if len(result.Bars) != tt.expectedBricks {
				t.Errorf("brick count = %d, want %d", len(result.Bars), tt.expectedBricks)
			}
		})
	}
}

// TestRenkoTransformer_UpBrickOHLC verifies O=lastLevel, C=lastLevel+boxSize,
// H=max(C,barHigh), L=min(O,barLow) for an up brick.
func TestRenkoTransformer_UpBrickOHLC(t *testing.T) {
	bars := []context.OHLCV{
		{Close: 100, High: 100, Low: 100},
		{Close: 115, High: 120, Low: 108},
	}
	result := NewRenkoTransformer(RenkoStyleATR, 10).Transform(bars)

	if len(result.Bars) != 1 {
		t.Fatalf("expected 1 up brick, got %d", len(result.Bars))
	}
	b := result.Bars[0]
	if b.Open != 100 {
		t.Errorf("up brick Open=%v, want 100 (lastLevel)", b.Open)
	}
	if b.Close != 110 {
		t.Errorf("up brick Close=%v, want 110 (lastLevel+boxSize)", b.Close)
	}
	if b.High != 120 {
		t.Errorf("up brick High=%v, want 120 (max(Close=110, barHigh=120))", b.High)
	}
	if b.Low != 100 {
		t.Errorf("up brick Low=%v, want 100 (min(Open=100, barLow=108))", b.Low)
	}
}

// TestRenkoTransformer_DownBrickOHLC verifies O=lastLevel, C=lastLevel−boxSize,
// H=max(O,barHigh), L=min(C,barLow) for a down brick.
func TestRenkoTransformer_DownBrickOHLC(t *testing.T) {
	bars := []context.OHLCV{
		{Close: 100, High: 100, Low: 100},
		{Close: 88, High: 95, Low: 82},
	}
	result := NewRenkoTransformer(RenkoStyleATR, 10).Transform(bars)

	if len(result.Bars) != 1 {
		t.Fatalf("expected 1 down brick, got %d", len(result.Bars))
	}
	b := result.Bars[0]
	if b.Open != 100 {
		t.Errorf("down brick Open=%v, want 100 (lastLevel)", b.Open)
	}
	if b.Close != 90 {
		t.Errorf("down brick Close=%v, want 90 (lastLevel−boxSize)", b.Close)
	}
	if b.High != 100 {
		t.Errorf("down brick High=%v, want 100 (max(Open=100, barHigh=95))", b.High)
	}
	if b.Low != 82 {
		t.Errorf("down brick Low=%v, want 82 (min(Close=90, barLow=82))", b.Low)
	}
}

// TestRenkoTransformer_MappingBeforeFirstBrick verifies that mapping entries are -1
// for all source bars before the first brick forms.
func TestRenkoTransformer_MappingBeforeFirstBrick(t *testing.T) {
	// boxSize=20: prices 100→105→108 never reach the 120 threshold.
	bars := pricesToBars([]float64{100, 105, 108})
	result := NewRenkoTransformer(RenkoStyleATR, 20).Transform(bars)

	if len(result.Bars) != 0 {
		t.Fatalf("expected 0 bricks, got %d", len(result.Bars))
	}
	for i, v := range result.MainToSynthetic {
		if v != -1 {
			t.Errorf("MainToSynthetic[%d]=%d, want -1 (no brick formed yet)", i, v)
		}
	}
}

// TestRenkoTransformer_TimeAndVolumeFromTriggeringBar verifies that each brick
// carries the Time and Volume of the source bar that caused it to form.
func TestRenkoTransformer_TimeAndVolumeFromTriggeringBar(t *testing.T) {
	bars := []context.OHLCV{
		{Close: 100, Time: 1000, Volume: 500},
		{Close: 125, Time: 2000, Volume: 9999},
	}
	result := NewRenkoTransformer(RenkoStyleATR, 10).Transform(bars)

	if len(result.Bars) != 2 {
		t.Fatalf("expected 2 bricks, got %d", len(result.Bars))
	}
	for i, b := range result.Bars {
		if b.Time != 2000 {
			t.Errorf("brick %d: Time=%v, want 2000 (triggering bar time)", i, b.Time)
		}
		if b.Volume != 9999 {
			t.Errorf("brick %d: Volume=%v, want 9999 (triggering bar volume)", i, b.Volume)
		}
	}
}

// TestRenkoTransformer_MappingInvariants verifies structural mapping guarantees
// across different price sequences: mapping length equals input, and is non-decreasing.
func TestRenkoTransformer_MappingInvariants(t *testing.T) {
	tests := []struct {
		name    string
		prices  []float64
		boxSize float64
	}{
		{"steady_uptrend", linspace(100, 5, 20), 8},
		{"steady_downtrend", linspace(200, -4, 20), 7},
		{"oscillating", oscillate(100, 12, -8, 20), 9},
		{"flat_then_spike", append(repeat(100, 5), 160), 15},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewRenkoTransformer(RenkoStyleATR, tt.boxSize).Transform(pricesToBars(tt.prices))
			assertMappingLength(t, result, len(tt.prices))
			assertMappingNonDecreasing(t, result)
			assertOHLCInvariants(t, result.Bars)
		})
	}
}

func TestRenkoTransformer_Type(t *testing.T) {
	if NewRenkoTransformer(RenkoStyleATR, 10).Type() != ModifierRenko {
		t.Errorf("Type() = %q, want %q", NewRenkoTransformer(RenkoStyleATR, 10).Type(), ModifierRenko)
	}
}

func pricesToBars(prices []float64) []context.OHLCV {
	bars := make([]context.OHLCV, len(prices))
	for i, p := range prices {
		bars[i] = context.OHLCV{Open: p, High: p + 1, Low: p - 1, Close: p}
	}
	return bars
}

func linspace(start, step float64, n int) []float64 {
	s := make([]float64, n)
	for i := range s {
		s[i] = start + float64(i)*step
	}
	return s
}

func oscillate(start, up, down float64, n int) []float64 {
	s := make([]float64, n)
	s[0] = start
	for i := 1; i < n; i++ {
		if i%2 == 1 {
			s[i] = s[i-1] + up
		} else {
			s[i] = s[i-1] + down
		}
	}
	return s
}

func repeat(v float64, n int) []float64 {
	s := make([]float64, n)
	for i := range s {
		s[i] = v
	}
	return s
}
