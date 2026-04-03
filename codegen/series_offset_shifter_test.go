package codegen

import "testing"

func TestSeriesOffsetShifter_OHLCVBarFields(t *testing.T) {
	shifter := SeriesOffsetShifter{}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "Close", input: "bar.Close", expected: "ctx.Data[i-1].Close"},
		{name: "Open", input: "bar.Open", expected: "ctx.Data[i-1].Open"},
		{name: "High", input: "bar.High", expected: "ctx.Data[i-1].High"},
		{name: "Low", input: "bar.Low", expected: "ctx.Data[i-1].Low"},
		{name: "Volume", input: "bar.Volume", expected: "ctx.Data[i-1].Volume"},
		{name: "unknown bar field passes through", input: "bar.HLC3", expected: "bar.HLC3"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shifter.ShiftToPrevBar(tt.input)
			if got != tt.expected {
				t.Errorf("ShiftToPrevBar(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestSeriesOffsetShifter_GetCurrentNormalization(t *testing.T) {
	shifter := SeriesOffsetShifter{}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "short series name",
			input:    "rsiSeries.GetCurrent()",
			expected: "rsiSeries.Get(1)",
		},
		{
			name:     "long series name",
			input:    "xATRTrailingStopSeries.GetCurrent()",
			expected: "xATRTrailingStopSeries.Get(1)",
		},
		{
			name:     "series name with digits",
			input:    "sma20Series.GetCurrent()",
			expected: "sma20Series.Get(1)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shifter.ShiftToPrevBar(tt.input)
			if got != tt.expected {
				t.Errorf("ShiftToPrevBar(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestSeriesOffsetShifter_SeriesGetOffsetIncrement(t *testing.T) {
	shifter := SeriesOffsetShifter{}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "offset 0 increments to 1",
			input:    "sma20Series.Get(0)",
			expected: "sma20Series.Get(1)",
		},
		{
			name:     "offset 1 increments to 2",
			input:    "alphaTrendSeries.Get(1)",
			expected: "alphaTrendSeries.Get(2)",
		},
		{
			name:     "offset 2 increments to 3",
			input:    "alphaTrendSeries.Get(2)",
			expected: "alphaTrendSeries.Get(3)",
		},
		{
			name:     "offset 9 increments to 10 across single-to-double digit boundary",
			input:    "emaSeries.Get(9)",
			expected: "emaSeries.Get(10)",
		},
		{
			name:     "large offset increments by exactly 1",
			input:    "emaSeries.Get(99)",
			expected: "emaSeries.Get(100)",
		},
		{
			name:     "series name is preserved unchanged",
			input:    "xATRTrailingStopSeries.Get(2)",
			expected: "xATRTrailingStopSeries.Get(3)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shifter.ShiftToPrevBar(tt.input)
			if got != tt.expected {
				t.Errorf("ShiftToPrevBar(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestSeriesOffsetShifter_HistoricalOHLCVOffsetIncrement(t *testing.T) {
	shifter := SeriesOffsetShifter{}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "close[2] IIFE offset 2 increments to 3",
			input:    "func() float64 { if i-2 >= 0 { return ctx.Data[i-2].Close }; return math.NaN() }()",
			expected: "func() float64 { if i-3 >= 0 { return ctx.Data[i-3].Close }; return math.NaN() }()",
		},
		{
			name:     "high[1] IIFE offset 1 increments to 2",
			input:    "func() float64 { if i-1 >= 0 { return ctx.Data[i-1].High }; return math.NaN() }()",
			expected: "func() float64 { if i-2 >= 0 { return ctx.Data[i-2].High }; return math.NaN() }()",
		},
		{
			name:     "low[9] IIFE offset 9 increments to 10 across single-to-double digit boundary",
			input:    "func() float64 { if i-9 >= 0 { return ctx.Data[i-9].Low }; return math.NaN() }()",
			expected: "func() float64 { if i-10 >= 0 { return ctx.Data[i-10].Low }; return math.NaN() }()",
		},
		{
			name:     "volume[5] IIFE preserves field name",
			input:    "func() float64 { if i-5 >= 0 { return ctx.Data[i-5].Volume }; return math.NaN() }()",
			expected: "func() float64 { if i-6 >= 0 { return ctx.Data[i-6].Volume }; return math.NaN() }()",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shifter.ShiftToPrevBar(tt.input)
			if got != tt.expected {
				t.Errorf("ShiftToPrevBar(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestSeriesOffsetShifter_PassthroughExpressions(t *testing.T) {
	shifter := SeriesOffsetShifter{}

	inputs := []struct {
		name  string
		input string
	}{
		{name: "integer literal", input: "30"},
		{name: "float literal", input: "30.0"},
		{name: "negative float", input: "-1.0"},
		{name: "bare identifier", input: "sma20"},
		{name: "empty string", input: ""},
	}

	for _, tt := range inputs {
		t.Run(tt.name, func(t *testing.T) {
			got := shifter.ShiftToPrevBar(tt.input)
			if got != tt.input {
				t.Errorf("ShiftToPrevBar(%q) = %q, want passthrough %q", tt.input, got, tt.input)
			}
		})
	}
}
