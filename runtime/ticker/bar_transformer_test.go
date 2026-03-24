package ticker

import (
	"math"
	"testing"

	"github.com/quant5-lab/runner/runtime/context"
)

func TestIdentityTransformer(t *testing.T) {
	transformer := &IdentityTransformer{}

	bars := []context.OHLCV{
		{Open: 100.0, High: 110.0, Low: 95.0, Close: 105.0},
		{Open: 105.0, High: 115.0, Low: 100.0, Close: 110.0},
	}

	result := transformer.Transform(bars)

	if len(result.Bars) != len(bars) {
		t.Fatalf("Expected %d bars, got %d", len(bars), len(result.Bars))
	}
	for i := range bars {
		if result.Bars[i] != bars[i] {
			t.Errorf("Bar %d changed: %+v != %+v", i, result.Bars[i], bars[i])
		}
	}
	if len(result.MainToSynthetic) != len(bars) {
		t.Errorf("Mapping length = %d, want %d", len(result.MainToSynthetic), len(bars))
	}
	for i := range bars {
		if result.MainToSynthetic[i] != i {
			t.Errorf("Mapping[%d] = %d, want %d", i, result.MainToSynthetic[i], i)
		}
	}
	if transformer.Type() != "" {
		t.Errorf("IdentityTransformer.Type() = %q, want empty", transformer.Type())
	}
}

func TestHeikinAshiTransformer_EmptyBars(t *testing.T) {
	result := (&HeikinAshiTransformer{}).Transform([]context.OHLCV{})
	if len(result.Bars) != 0 {
		t.Errorf("Expected empty result, got %d bars", len(result.Bars))
	}
}

// TestHeikinAshiTransformer_HACloseFormula verifies haClose = (O+H+L+C)/4 for each bar.
func TestHeikinAshiTransformer_HACloseFormula(t *testing.T) {
	bars := []context.OHLCV{
		{Open: 100.0, High: 110.0, Low: 95.0, Close: 105.0, Volume: 1000.0, Time: 1000},
		{Open: 105.0, High: 120.0, Low: 100.0, Close: 115.0, Volume: 2000.0, Time: 2000},
	}
	result := (&HeikinAshiTransformer{}).Transform(bars)

	for i, bar := range bars {
		expected := (bar.Open + bar.High + bar.Low + bar.Close) / 4.0
		if math.Abs(result.Bars[i].Close-expected) > 1e-9 {
			t.Errorf("bar %d: haClose=%v, want (O+H+L+C)/4=%v", i, result.Bars[i].Close, expected)
		}
	}
}

// TestHeikinAshiTransformer_HAOpenChain verifies the haOpen recurrence:
// haOpen[0] = (O+C)/2; haOpen[i] = (prevHaOpen+prevHaClose)/2.
func TestHeikinAshiTransformer_HAOpenChain(t *testing.T) {
	bars := []context.OHLCV{
		{Open: 100.0, High: 110.0, Low: 95.0, Close: 105.0},
		{Open: 105.0, High: 115.0, Low: 100.0, Close: 110.0},
		{Open: 110.0, High: 120.0, Low: 105.0, Close: 115.0},
	}
	result := (&HeikinAshiTransformer{}).Transform(bars)

	expectedOpen0 := (bars[0].Open + bars[0].Close) / 2.0
	if math.Abs(result.Bars[0].Open-expectedOpen0) > 1e-9 {
		t.Errorf("bar 0: haOpen=%v, want (O+C)/2=%v", result.Bars[0].Open, expectedOpen0)
	}
	for i := 1; i < len(result.Bars); i++ {
		expected := (result.Bars[i-1].Open + result.Bars[i-1].Close) / 2.0
		if math.Abs(result.Bars[i].Open-expected) > 1e-9 {
			t.Errorf("bar %d: haOpen=%v, want (prevHaO+prevHaC)/2=%v", i, result.Bars[i].Open, expected)
		}
	}
}

// TestHeikinAshiTransformer_OHLCInvariants verifies H>=max(O,C) and L<=min(O,C)
// across trend directions and edge-price inputs.
func TestHeikinAshiTransformer_OHLCInvariants(t *testing.T) {
	tests := []struct {
		name string
		bars []context.OHLCV
	}{
		{
			name: "bullish_trend",
			bars: []context.OHLCV{
				{Open: 100, High: 110, Low: 98, Close: 108},
				{Open: 108, High: 120, Low: 105, Close: 117},
				{Open: 117, High: 125, Low: 113, Close: 122},
			},
		},
		{
			name: "bearish_trend",
			bars: []context.OHLCV{
				{Open: 100, High: 105, Low: 90, Close: 92},
				{Open: 92, High: 95, Low: 80, Close: 83},
				{Open: 83, High: 86, Low: 72, Close: 74},
			},
		},
		{
			name: "flat_price",
			bars: []context.OHLCV{
				{Open: 100, High: 100, Low: 100, Close: 100},
				{Open: 100, High: 100, Low: 100, Close: 100},
			},
		},
		{
			name: "reversal",
			bars: []context.OHLCV{
				{Open: 100, High: 115, Low: 98, Close: 112},
				{Open: 112, High: 113, Low: 95, Close: 97},
				{Open: 97, High: 99, Low: 82, Close: 85},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := (&HeikinAshiTransformer{}).Transform(tt.bars)
			assertOHLCInvariants(t, result.Bars)
		})
	}
}

// TestHeikinAshiTransformer_MetadataPassthrough verifies that Volume and Time are
// preserved from each source bar, that bar count equals input count, and that the
// mapping is identity (HeikinAshi is a 1:1 transform).
func TestHeikinAshiTransformer_MetadataPassthrough(t *testing.T) {
	bars := []context.OHLCV{
		{Open: 100, High: 110, Low: 95, Close: 105, Volume: 12345, Time: 1609459200},
		{Open: 105, High: 115, Low: 100, Close: 112, Volume: 67890, Time: 1609545600},
		{Open: 112, High: 120, Low: 108, Close: 118, Volume: 11111, Time: 1609632000},
	}
	result := (&HeikinAshiTransformer{}).Transform(bars)

	if len(result.Bars) != len(bars) {
		t.Fatalf("bar count = %d, want %d", len(result.Bars), len(bars))
	}
	assertMappingLength(t, result, len(bars))
	for i, bar := range bars {
		if result.Bars[i].Volume != bar.Volume {
			t.Errorf("bar %d: Volume=%v, want %v", i, result.Bars[i].Volume, bar.Volume)
		}
		if result.Bars[i].Time != bar.Time {
			t.Errorf("bar %d: Time=%v, want %v", i, result.Bars[i].Time, bar.Time)
		}
		if result.MainToSynthetic[i] != i {
			t.Errorf("MainToSynthetic[%d]=%d, want %d (identity mapping)", i, result.MainToSynthetic[i], i)
		}
	}
}

func TestHeikinAshiTransformer_Type(t *testing.T) {
	if (&HeikinAshiTransformer{}).Type() != ModifierHeikinAshi {
		t.Errorf("Type() = %q, want %q", (&HeikinAshiTransformer{}).Type(), ModifierHeikinAshi)
	}
}

func TestNewTransformer(t *testing.T) {
	tests := []struct {
		modifierType ModifierType
		expectedType string
	}{
		{ModifierHeikinAshi, "*ticker.HeikinAshiTransformer"},
		{ModifierRenko, "*ticker.RenkoTransformer"},
		{ModifierKagi, "*ticker.KagiTransformer"},
		{ModifierLineBreak, "*ticker.LineBreakTransformer"},
		{ModifierPointFig, "*ticker.PointFigureTransformer"},
		{"", "*ticker.IdentityTransformer"},
	}

	for _, tt := range tests {
		t.Run(string(tt.modifierType), func(t *testing.T) {
			transformer := NewTransformer(tt.modifierType)
			if transformer == nil {
				t.Fatal("NewTransformer returned nil")
			}
			actualType := transformerTypeName(transformer)
			if actualType != tt.expectedType {
				t.Errorf("NewTransformer(%q) returned %s, want %s",
					tt.modifierType, actualType, tt.expectedType)
			}
		})
	}
}

func transformerTypeName(v interface{}) string {
	switch v.(type) {
	case *HeikinAshiTransformer:
		return "*ticker.HeikinAshiTransformer"
	case *IdentityTransformer:
		return "*ticker.IdentityTransformer"
	case *RenkoTransformer:
		return "*ticker.RenkoTransformer"
	case *KagiTransformer:
		return "*ticker.KagiTransformer"
	case *LineBreakTransformer:
		return "*ticker.LineBreakTransformer"
	case *PointFigureTransformer:
		return "*ticker.PointFigureTransformer"
	default:
		return "unknown"
	}
}
