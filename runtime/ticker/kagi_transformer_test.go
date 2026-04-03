package ticker

import (
	"testing"
)

// TestKagiTransformer_NoSegmentsWithoutDirectionChange covers the two conditions
// that produce zero output segments: nil input and flat price (undecided direction).
func TestKagiTransformer_NoSegmentsWithoutDirectionChange(t *testing.T) {
	tests := []struct {
		name   string
		prices []float64
	}{
		{"nil_input", nil},
		{"all_same_price", []float64{100, 100, 100}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewKagiTransformer(1.0).Transform(pricesToBars(tt.prices))
			if len(result.Bars) != 0 {
				t.Errorf("expected 0 segments, got %d", len(result.Bars))
			}
		})
	}
}

// TestKagiTransformer_DevelopingSegmentFlush verifies that the ongoing developing
// segment is always emitted as the final bar, for both trend directions.
func TestKagiTransformer_DevelopingSegmentFlush(t *testing.T) {
	tests := []struct {
		name          string
		prices        []float64
		wantSegments  int
		wantLastClose float64
	}{
		{
			name:          "up_trend_flushed",
			prices:        []float64{100, 105, 110},
			wantSegments:  1,
			wantLastClose: 110,
		},
		{
			name:          "down_trend_flushed",
			prices:        []float64{110, 105, 100},
			wantSegments:  1,
			wantLastClose: 100,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewKagiTransformer(2.0).Transform(pricesToBars(tt.prices))
			if len(result.Bars) != tt.wantSegments {
				t.Fatalf("segment count = %d, want %d", len(result.Bars), tt.wantSegments)
			}
			last := result.Bars[len(result.Bars)-1]
			if last.Close != tt.wantLastClose {
				t.Errorf("last segment Close=%v, want %v", last.Close, tt.wantLastClose)
			}
		})
	}
}

// TestKagiTransformer_ReversalCount verifies that N reversals produce N+1 segments
// (N completed + 1 developing segment that is flushed).
func TestKagiTransformer_ReversalCount(t *testing.T) {
	tests := []struct {
		name         string
		prices       []float64
		reversal     float64
		wantSegments int
	}{
		{
			name:         "one_reversal",
			prices:       []float64{100, 110, 104},
			reversal:     5.0,
			wantSegments: 2,
		},
		{
			name:         "two_reversals",
			prices:       []float64{100, 110, 104, 112},
			reversal:     5.0,
			wantSegments: 3,
		},
		{
			// bar[5]=108 is below the up-reversal threshold (106+5=111), so no 4th reversal.
			name:         "three_reversals",
			prices:       []float64{100, 110, 104, 112, 106, 108},
			reversal:     5.0,
			wantSegments: 4,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewKagiTransformer(tt.reversal).Transform(pricesToBars(tt.prices))
			if len(result.Bars) != tt.wantSegments {
				t.Errorf("segment count = %d, want %d", len(result.Bars), tt.wantSegments)
			}
		})
	}
}

// TestKagiTransformer_SegmentOHLC verifies the OHLC construction for completed segments:
//
//	up segment:   O=segStart C=segHigh H=C  L=O
//	down segment: O=segStart C=segLow  H=O  L=C
func TestKagiTransformer_SegmentOHLC(t *testing.T) {
	// up segment: segStart=100, extreme rises to 110; reversal at 104.
	upResult := NewKagiTransformer(5.0).Transform(pricesToBars([]float64{100, 110, 104}))
	if len(upResult.Bars) < 1 {
		t.Fatalf("expected at least 1 segment for up-trend case")
	}
	us := upResult.Bars[0]
	if us.Open != 100 || us.Close != 110 {
		t.Errorf("up segment O=%v C=%v, want O=100 C=110", us.Open, us.Close)
	}
	if us.High != us.Close {
		t.Errorf("up segment High=%v, want Close=%v", us.High, us.Close)
	}
	if us.Low != us.Open {
		t.Errorf("up segment Low=%v, want Open=%v", us.Low, us.Open)
	}

	// down segment: segStart=100, extreme falls to 90; reversal at 96.
	downResult := NewKagiTransformer(5.0).Transform(pricesToBars([]float64{100, 90, 96}))
	if len(downResult.Bars) < 1 {
		t.Fatalf("expected at least 1 segment for down-trend case")
	}
	ds := downResult.Bars[0]
	if ds.Open != 100 || ds.Close != 90 {
		t.Errorf("down segment O=%v C=%v, want O=100 C=90", ds.Open, ds.Close)
	}
	if ds.High != ds.Open {
		t.Errorf("down segment High=%v, want Open=%v", ds.High, ds.Open)
	}
	if ds.Low != ds.Close {
		t.Errorf("down segment Low=%v, want Close=%v", ds.Low, ds.Close)
	}
}

// TestKagiTransformer_MappingPreFirstReversal verifies that source bars preceding the
// first reversal are all mapped to segment 0 after -1 normalization.
func TestKagiTransformer_MappingPreFirstReversal(t *testing.T) {
	// Bars 0,1,2 occur before the reversal triggered at bar 3 (drop from 108 to 103).
	prices := []float64{100, 105, 108, 103}
	result := NewKagiTransformer(5.0).Transform(pricesToBars(prices))

	if len(result.Bars) < 2 {
		t.Fatalf("expected at least 2 segments (1 completed + 1 developing), got %d", len(result.Bars))
	}
	for i := 0; i <= 2; i++ {
		if result.MainToSynthetic[i] != 0 {
			t.Errorf("MainToSynthetic[%d]=%d, want 0 (normalized pre-reversal bar)", i, result.MainToSynthetic[i])
		}
	}
}

// TestKagiTransformer_MappingInvariants verifies structural mapping guarantees
// across multiple price sequences.
func TestKagiTransformer_MappingInvariants(t *testing.T) {
	tests := []struct {
		name     string
		prices   []float64
		reversal float64
	}{
		{"single_up_trend", linspace(100, 3, 15), 2.0},
		{"single_down_trend", linspace(200, -3, 15), 2.0},
		{"multi_reversal", oscillate(100, 8, -6, 20), 5.0},
		{"tight_oscillation", oscillate(100, 2, -2, 20), 1.0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewKagiTransformer(tt.reversal).Transform(pricesToBars(tt.prices))
			assertMappingLength(t, result, len(tt.prices))
			assertMappingNonDecreasing(t, result)
			if len(result.Bars) > 0 {
				assertMappingValidIndices(t, result)
				assertOHLCInvariants(t, result.Bars)
			}
		})
	}
}

func TestKagiTransformer_Type(t *testing.T) {
	if NewKagiTransformer(1.0).Type() != ModifierKagi {
		t.Errorf("Type() = %q, want %q", NewKagiTransformer(1.0).Type(), ModifierKagi)
	}
}
