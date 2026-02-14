package codegen

import (
	"strings"
	"testing"
)

/* Edge case: nil receiver should not panic */
func TestArrowBuiltinAccessGenerator_NilReceiver(t *testing.T) {
	var gen *ArrowBuiltinAccessGenerator

	nilSafeMethods := map[string]string{
		"CurrentAccess close":       gen.GenerateCurrentAccess("close"),
		"CurrentAccess tr":          gen.GenerateCurrentAccess("tr"),
		"HistoricalAccess close[1]": gen.GenerateHistoricalAccess("close", 1),
		"HistoricalAccess tr[1]":    gen.GenerateHistoricalAccess("tr", 1),
		"StrategyAccess equity":     gen.GenerateStrategyAccess("equity"),
	}

	for name, output := range nilSafeMethods {
		if output != "" {
			t.Errorf("nil receiver %s should return empty string, got: %q", name, output)
		}
	}
}

/* Historical access with offset=0 still generates bounds-checked IIFE for consistency */
func TestArrowBuiltinAccessGenerator_HistoricalOffset_Zero(t *testing.T) {
	registry := NewBuiltinIdentifierRegistry()
	formulaGen := NewDerivedPriceFormulaGenerator()
	gen := NewArrowBuiltinAccessGenerator(registry, formulaGen)

	tests := []struct {
		name     string
		builtin  string
		contains []string
	}{
		{
			"close[0] uses offset=0 in IIFE",
			"close",
			[]string{"ctx.BarIndex-0", "ctx.Data[ctx.BarIndex-0].Close"},
		},
		{
			"bar_index[0] uses offset=0",
			"bar_index",
			[]string{"ctx.BarIndex-0", "float64(ctx.BarIndex-0)"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := gen.GenerateHistoricalAccess(tt.builtin, 0)
			if code == "" {
				t.Errorf("Expected non-empty code for %s[0]", tt.builtin)
			}
			for _, substr := range tt.contains {
				if !strings.Contains(code, substr) {
					t.Errorf("Expected %s[0] to contain %q, got: %s", tt.builtin, substr, code)
				}
			}
		})
	}
}

/* Historical access should generate bounds-checked IIFEs */
func TestArrowBuiltinAccessGenerator_HistoricalBoundsCheck(t *testing.T) {
	registry := NewBuiltinIdentifierRegistry()
	formulaGen := NewDerivedPriceFormulaGenerator()
	gen := NewArrowBuiltinAccessGenerator(registry, formulaGen)

	tests := []struct {
		name     string
		builtin  string
		offset   int
		contains []string
	}{
		{
			"close[1] has bounds check",
			"close",
			1,
			[]string{"ctx.BarIndex-1 < 0", "math.NaN()", "ctx.Data[ctx.BarIndex-1].Close"},
		},
		{
			"bar_index[5] has bounds check",
			"bar_index",
			5,
			[]string{"ctx.BarIndex-5 < 0", "math.NaN()", "float64(ctx.BarIndex-5)"},
		},
		{
			"hl2[3] has bounds check + series access",
			"hl2",
			3,
			[]string{"ctx.BarIndex-3 < 0", "math.NaN()", "highSeries.Get(ctx.BarIndex-3)", "lowSeries.Get(ctx.BarIndex-3)"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := gen.GenerateHistoricalAccess(tt.builtin, tt.offset)
			if code == "" {
				t.Fatalf("Expected non-empty code for %s[%d]", tt.builtin, tt.offset)
			}
			for _, substr := range tt.contains {
				if !strings.Contains(code, substr) {
					t.Errorf("Expected %s[%d] to contain %q\nGot: %s", tt.builtin, tt.offset, substr, code)
				}
			}
		})
	}
}

/* TR historical access uses TrueRangeArrowIIFE with proper bounds check */
func TestArrowBuiltinAccessGenerator_TrueRangeHistorical(t *testing.T) {
	registry := NewBuiltinIdentifierRegistry()
	formulaGen := NewDerivedPriceFormulaGenerator()
	gen := NewArrowBuiltinAccessGenerator(registry, formulaGen)

	offsets := []int{0, 1, 5, 10}
	for _, offset := range offsets {
		t.Run(string(rune('0'+(offset/10)%10))+string(rune('0'+offset%10)), func(t *testing.T) {
			code := gen.GenerateHistoricalAccess("tr", offset)
			if code == "" {
				t.Fatalf("Expected non-empty code for tr[%d]", offset)
			}

			expectedComponents := []string{
				"barIdx := ctx.BarIndex-int(",
				"barIdx >= 0 && barIdx < len(ctx.Data)",
				"math.NaN()",
				"ctx.Data[barIdx]",
				"math.Max",
			}
			for _, component := range expectedComponents {
				if !strings.Contains(code, component) {
					t.Errorf("Expected tr[%d] to contain %q\nGot: %s", offset, component, code)
				}
			}
		})
	}
}

/* Calendar builtins use SeriesLookupIIFE with NaN fallback */
func TestArrowBuiltinAccessGenerator_CalendarBuiltinsSeriesLookup(t *testing.T) {
	registry := NewBuiltinIdentifierRegistry()
	formulaGen := NewDerivedPriceFormulaGenerator()
	gen := NewArrowBuiltinAccessGenerator(registry, formulaGen)

	calendarBuiltins := registry.CalendarBuiltinNames()

	for _, builtin := range calendarBuiltins {
		t.Run(builtin, func(t *testing.T) {
			code := gen.GenerateCurrentAccess(builtin)
			if code == "" {
				t.Fatalf("Expected non-empty code for %s", builtin)
			}

			expectedComponents := []string{
				"ctx.LookupSeries",
				"s.GetCurrent()",
				"math.NaN()",
				builtin + "Series",
			}
			for _, component := range expectedComponents {
				if !strings.Contains(code, component) {
					t.Errorf("Expected %s to contain %q\nGot: %s", builtin, component, code)
				}
			}
		})
	}
}

/* Derived prices in arrow scope use ctx.Data[ctx.BarIndex] for all OHLCV fields */
func TestArrowBuiltinAccessGenerator_DerivedPrices(t *testing.T) {
	registry := NewBuiltinIdentifierRegistry()
	formulaGen := NewDerivedPriceFormulaGenerator()
	gen := NewArrowBuiltinAccessGenerator(registry, formulaGen)

	tests := []struct {
		name            string
		derivedPrice    string
		expectedFields  []string
		arithmeticCheck string
	}{
		{
			"hl2 uses High and Low",
			"hl2",
			[]string{"ctx.Data[ctx.BarIndex].High", "ctx.Data[ctx.BarIndex].Low"},
			"/",
		},
		{
			"hlc3 uses High, Low, Close",
			"hlc3",
			[]string{"ctx.Data[ctx.BarIndex].High", "ctx.Data[ctx.BarIndex].Low", "ctx.Data[ctx.BarIndex].Close"},
			"/",
		},
		{
			"ohlc4 uses all four",
			"ohlc4",
			[]string{"ctx.Data[ctx.BarIndex].Open", "ctx.Data[ctx.BarIndex].High", "ctx.Data[ctx.BarIndex].Low", "ctx.Data[ctx.BarIndex].Close"},
			"/",
		},
		{
			"hlcc4 uses High, Low, Close twice",
			"hlcc4",
			[]string{"ctx.Data[ctx.BarIndex].High", "ctx.Data[ctx.BarIndex].Low", "ctx.Data[ctx.BarIndex].Close"},
			"/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := gen.GenerateCurrentAccess(tt.derivedPrice)
			if code == "" {
				t.Fatalf("Expected non-empty code for %s", tt.derivedPrice)
			}
			for _, field := range tt.expectedFields {
				if !strings.Contains(code, field) {
					t.Errorf("Expected %s to contain %q\nGot: %s", tt.derivedPrice, field, code)
				}
			}
			if !strings.Contains(code, tt.arithmeticCheck) {
				t.Errorf("Expected %s to contain arithmetic operator %q\nGot: %s", tt.derivedPrice, tt.arithmeticCheck, code)
			}
		})
	}
}

/* Strategy access generates SeriesLookupIIFE for all properties except position_entry_name */
func TestArrowBuiltinAccessGenerator_StrategyAccess(t *testing.T) {
	registry := NewBuiltinIdentifierRegistry()
	formulaGen := NewDerivedPriceFormulaGenerator()
	gen := NewArrowBuiltinAccessGenerator(registry, formulaGen)

	tests := []struct {
		name       string
		property   string
		expectIIFE bool
		contains   []string
	}{
		{
			"position_avg_price uses series lookup",
			"position_avg_price",
			true,
			[]string{"ctx.LookupSeries", StrategyPositionAvgPriceSeriesName, "math.NaN()"},
		},
		{
			"position_size uses series lookup",
			"position_size",
			true,
			[]string{"ctx.LookupSeries", StrategyPositionSizeSeriesName, "math.NaN()"},
		},
		{
			"equity uses series lookup",
			"equity",
			true,
			[]string{"ctx.LookupSeries", StrategyEquitySeriesName, "math.NaN()"},
		},
		{
			"netprofit uses series lookup",
			"netprofit",
			true,
			[]string{"ctx.LookupSeries", StrategyNetProfitSeriesName, "math.NaN()"},
		},
		{
			"closedtrades uses series lookup",
			"closedtrades",
			true,
			[]string{"ctx.LookupSeries", StrategyClosedTradesSeriesName, "math.NaN()"},
		},
		{
			"position_entry_name returns empty string",
			"position_entry_name",
			false,
			[]string{`""`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := gen.GenerateStrategyAccess(tt.property)
			if code == "" && tt.expectIIFE {
				t.Fatalf("Expected non-empty code for strategy.%s", tt.property)
			}
			for _, substr := range tt.contains {
				if !strings.Contains(code, substr) {
					t.Errorf("Expected strategy.%s to contain %q\nGot: %s", tt.property, substr, code)
				}
			}
		})
	}
}

func TestArrowBuiltinAccessGenerator_DirectBarFieldAccess(t *testing.T) {
	registry := NewBuiltinIdentifierRegistry()
	formulaGen := NewDerivedPriceFormulaGenerator()
	gen := NewArrowBuiltinAccessGenerator(registry, formulaGen)

	tests := []struct {
		name     string
		expected string
	}{
		{"close", "ctx.Data[ctx.BarIndex].Close"},
		{"open", "ctx.Data[ctx.BarIndex].Open"},
		{"high", "ctx.Data[ctx.BarIndex].High"},
		{"low", "ctx.Data[ctx.BarIndex].Low"},
		{"volume", "ctx.Data[ctx.BarIndex].Volume"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := gen.GenerateCurrentAccess(tt.name)
			if code != tt.expected {
				t.Errorf("GenerateCurrentAccess(%s) = %q, want %q", tt.name, code, tt.expected)
			}
		})
	}
}

func TestArrowBuiltinAccessGenerator_TemporalAndIndexAccess(t *testing.T) {
	registry := NewBuiltinIdentifierRegistry()
	formulaGen := NewDerivedPriceFormulaGenerator()
	gen := NewArrowBuiltinAccessGenerator(registry, formulaGen)

	tests := []struct {
		name     string
		expected string
	}{
		{"bar_index", "float64(ctx.BarIndex)"},
		{"time", "float64(ctx.Data[ctx.BarIndex].Time * 1000)"},
		{"last_bar_index", "float64(len(ctx.Data) - 1)"},
		{"last_bar_time", "float64(ctx.Data[len(ctx.Data)-1].Time * 1000)"},
		{"timenow", "float64(ctx.Data[len(ctx.Data)-1].Time * 1000)"},
		{"time_close", SeriesLookupIIFE("time_closeSeries")},
		{"time_tradingday", SeriesLookupIIFE("time_tradingdaySeries")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := gen.GenerateCurrentAccess(tt.name)
			if code != tt.expected {
				t.Errorf("GenerateCurrentAccess(%s) = %q, want %q", tt.name, code, tt.expected)
			}
		})
	}
}

func TestArrowBuiltinAccessGenerator_TrueRangeCurrent(t *testing.T) {
	registry := NewBuiltinIdentifierRegistry()
	formulaGen := NewDerivedPriceFormulaGenerator()
	gen := NewArrowBuiltinAccessGenerator(registry, formulaGen)

	code := gen.GenerateCurrentAccess("tr")
	if code == "" {
		t.Fatal("GenerateCurrentAccess(tr) returned empty")
	}

	expectedComponents := []string{
		"func() float64",
		"ctx.Data[ctx.BarIndex]",
		"curBar",
		"prevClose",
		"math.Max",
		"math.Abs",
		"if ctx.BarIndex < 1",
	}
	for _, component := range expectedComponents {
		if !strings.Contains(code, component) {
			t.Errorf("GenerateCurrentAccess(tr) missing %q\nGot: %s", component, code)
		}
	}
}

/* Edge case: unknown builtins should return empty string */
func TestArrowBuiltinAccessGenerator_UnknownBuiltins(t *testing.T) {
	registry := NewBuiltinIdentifierRegistry()
	formulaGen := NewDerivedPriceFormulaGenerator()
	gen := NewArrowBuiltinAccessGenerator(registry, formulaGen)

	unknowns := []string{"unknown_var", "my_custom_indicator", "foobar"}

	for _, name := range unknowns {
		t.Run(name, func(t *testing.T) {
			currentCode := gen.GenerateCurrentAccess(name)
			historicalCode := gen.GenerateHistoricalAccess(name, 1)

			if currentCode != "" {
				t.Errorf("Expected empty string for unknown builtin %s (current), got: %q", name, currentCode)
			}
			if historicalCode != "" {
				t.Errorf("Expected empty string for unknown builtin %s (historical), got: %q", name, historicalCode)
			}
		})
	}
}
