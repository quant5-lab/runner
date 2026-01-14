package context

import "testing"

func TestIsMonthlyTimeframe(t *testing.T) {
	tests := []struct {
		name     string
		tf       string
		expected bool
	}{
		{"M format", "M", true},
		{"1M format", "1M", true},
		{"1mo format", "1mo", true},
		{"D format", "D", false},
		{"1D format", "1D", false},
		{"W format", "W", false},
		{"1h format", "1h", false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsMonthlyTimeframe(tt.tf)
			if got != tt.expected {
				t.Errorf("IsMonthlyTimeframe(%q) = %v, want %v", tt.tf, got, tt.expected)
			}
		})
	}
}

func TestIsDailyTimeframe(t *testing.T) {
	tests := []struct {
		name     string
		tf       string
		expected bool
	}{
		{"D format", "D", true},
		{"1D format", "1D", true},
		{"1d format", "1d", true},
		{"M format", "M", false},
		{"W format", "W", false},
		{"1h format", "1h", false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsDailyTimeframe(tt.tf)
			if got != tt.expected {
				t.Errorf("IsDailyTimeframe(%q) = %v, want %v", tt.tf, got, tt.expected)
			}
		})
	}
}

func TestIsWeeklyTimeframe(t *testing.T) {
	tests := []struct {
		name     string
		tf       string
		expected bool
	}{
		{"W format", "W", true},
		{"1W format", "1W", true},
		{"1w format", "1w", true},
		{"1wk format", "1wk", true},
		{"D format", "D", false},
		{"M format", "M", false},
		{"1h format", "1h", false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsWeeklyTimeframe(tt.tf)
			if got != tt.expected {
				t.Errorf("IsWeeklyTimeframe(%q) = %v, want %v", tt.tf, got, tt.expected)
			}
		})
	}
}

func TestIsIntradayTimeframe(t *testing.T) {
	tests := []struct {
		name     string
		tf       string
		expected bool
	}{
		{"1m format", "1m", true},
		{"5m format", "5m", true},
		{"1h format", "1h", true},
		{"4h format", "4h", true},
		{"D format", "D", false},
		{"1D format", "1D", false},
		{"W format", "W", false},
		{"M format", "M", false},
		{"empty string", "", true}, // Not monthly/daily/weekly = intraday
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsIntradayTimeframe(tt.tf)
			if got != tt.expected {
				t.Errorf("IsIntradayTimeframe(%q) = %v, want %v", tt.tf, got, tt.expected)
			}
		})
	}
}

func TestContextTimeframeFlags(t *testing.T) {
	tests := []struct {
		name           string
		timeframe      string
		expectMonthly  bool
		expectDaily    bool
		expectWeekly   bool
		expectIntraday bool
	}{
		{
			name:           "Monthly M",
			timeframe:      "M",
			expectMonthly:  true,
			expectDaily:    false,
			expectWeekly:   false,
			expectIntraday: false,
		},
		{
			name:           "Monthly 1mo",
			timeframe:      "1mo",
			expectMonthly:  true,
			expectDaily:    false,
			expectWeekly:   false,
			expectIntraday: false,
		},
		{
			name:           "Daily D",
			timeframe:      "D",
			expectMonthly:  false,
			expectDaily:    true,
			expectWeekly:   false,
			expectIntraday: false,
		},
		{
			name:           "Daily 1d",
			timeframe:      "1d",
			expectMonthly:  false,
			expectDaily:    true,
			expectWeekly:   false,
			expectIntraday: false,
		},
		{
			name:           "Weekly W",
			timeframe:      "W",
			expectMonthly:  false,
			expectDaily:    false,
			expectWeekly:   true,
			expectIntraday: false,
		},
		{
			name:           "Hourly 1h",
			timeframe:      "1h",
			expectMonthly:  false,
			expectDaily:    false,
			expectWeekly:   false,
			expectIntraday: true,
		},
		{
			name:           "Minute 5m",
			timeframe:      "5m",
			expectMonthly:  false,
			expectDaily:    false,
			expectWeekly:   false,
			expectIntraday: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := New("TEST", tt.timeframe, 100)

			if ctx.IsMonthly != tt.expectMonthly {
				t.Errorf("IsMonthly = %v, want %v", ctx.IsMonthly, tt.expectMonthly)
			}
			if ctx.IsDaily != tt.expectDaily {
				t.Errorf("IsDaily = %v, want %v", ctx.IsDaily, tt.expectDaily)
			}
			if ctx.IsWeekly != tt.expectWeekly {
				t.Errorf("IsWeekly = %v, want %v", ctx.IsWeekly, tt.expectWeekly)
			}
			if ctx.IsIntraday != tt.expectIntraday {
				t.Errorf("IsIntraday = %v, want %v", ctx.IsIntraday, tt.expectIntraday)
			}
		})
	}
}
