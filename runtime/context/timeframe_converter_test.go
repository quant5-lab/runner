package context

import "testing"

func TestTimeframeConverter_ToSeconds(t *testing.T) {
	converter := NewTimeframeConverter()

	tests := []struct {
		name      string
		timeframe string
		expected  int64
	}{
		{"second", "1s", 1},
		{"5 minutes", "5m", 300},
		{"1 hour", "1h", 3600},
		{"4 hours", "4h", 14400},
		{"daily uppercase", "1D", 86400},
		{"daily lowercase", "1d", 86400},
		{"weekly uppercase", "1W", 604800},
		{"weekly lowercase", "1w", 604800},
		{"monthly", "1M", 2628003},
		{"2 months", "2M", 5256006},
		{"12 months", "12M", 31536036},
		{"single char D", "D", 86400},
		{"single char W", "W", 604800},
		{"single char M", "M", 2628003},
		{"zero multiplier", "0m", 60},
		{"empty string", "", 0},
		{"invalid unit", "5x", 0},
		{"large multiplier", "240h", 864000},
		{"0 minutes clamps to 1", "0", 60},
		{"1 minute", "1", 60},
		{"3 minutes", "3", 180},
		{"5 minutes numeric", "5", 300},
		{"10 minutes", "10", 600},
		{"15 minutes numeric", "15", 900},
		{"30 minutes numeric", "30", 1800},
		{"45 minutes", "45", 2700},
		{"60 minutes", "60", 3600},
		{"120 minutes", "120", 7200},
		{"240 minutes", "240", 14400},
		{"1440 minutes equals one day", "1440", 86400},
		{"10080 minutes equals one week", "10080", 604800},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := converter.ToSeconds(tt.timeframe)
			if result != tt.expected {
				t.Errorf("ToSeconds(%q) = %d, want %d", tt.timeframe, result, tt.expected)
			}
		})
	}
}

func TestTimeframeConverter_FromSeconds(t *testing.T) {
	converter := NewTimeframeConverter()

	tests := []struct {
		seconds  int64
		expected string
	}{
		{1, "1s"},
		{60, "1m"},
		{300, "5m"},
		{3600, "1h"},
		{14400, "4h"},
		{86400, "1D"},
		{604800, "1W"},
		{2628003, "1M"},
		{120, "2m"},
		{7200, "2h"},
		{172800, "2D"},
		{5256006, "2M"},
		{15768018, "6M"},
		{31536036, "12M"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := converter.FromSeconds(tt.seconds)
			if result != tt.expected {
				t.Errorf("FromSeconds(%d) = %q, want %q", tt.seconds, result, tt.expected)
			}
		})
	}

	t.Run("zero returns empty", func(t *testing.T) {
		result := converter.FromSeconds(0)
		if result != "" {
			t.Errorf("FromSeconds(0) = %q, want empty", result)
		}
	})

	t.Run("negative returns empty", func(t *testing.T) {
		result := converter.FromSeconds(-1)
		if result != "" {
			t.Errorf("FromSeconds(-1) = %q, want empty", result)
		}
	})

	t.Run("above 366 days caps to 12M", func(t *testing.T) {
		result := converter.FromSeconds(31622401)
		if result != "12M" {
			t.Errorf("FromSeconds(31622401) = %q, want %q", result, "12M")
		}
	})

	t.Run("exactly 366 days is 366D", func(t *testing.T) {
		result := converter.FromSeconds(31622400)
		if result != "366D" {
			t.Errorf("FromSeconds(31622400) = %q, want %q", result, "366D")
		}
	})

	t.Run("very large value caps to 12M", func(t *testing.T) {
		result := converter.FromSeconds(100_000_000)
		if result != "12M" {
			t.Errorf("FromSeconds(100000000) = %q, want %q", result, "12M")
		}
	})
}

func TestTimeframeConverter_RoundtripProperties(t *testing.T) {
	converter := NewTimeframeConverter()

	t.Run("canonical forms survive roundtrip", func(t *testing.T) {
		canonicalTimeframes := []string{
			"1s", "1m", "5m", "15m",
			"1h", "4h",
			"1D", "2D",
			"1W",
			"1M", "2M", "6M", "12M",
		}
		for _, tf := range canonicalTimeframes {
			seconds := converter.ToSeconds(tf)
			roundtripped := converter.FromSeconds(seconds)
			if roundtripped != tf {
				t.Errorf("FromSeconds(ToSeconds(%q)) = %q, want %q (via %d seconds)",
					tf, roundtripped, tf, seconds)
			}
		}
	})

	t.Run("lowercase aliases normalize to canonical", func(t *testing.T) {
		aliases := map[string]string{
			"1d": "1D",
			"1w": "1W",
		}
		for alias, canonical := range aliases {
			seconds := converter.ToSeconds(alias)
			result := converter.FromSeconds(seconds)
			if result != canonical {
				t.Errorf("FromSeconds(ToSeconds(%q)) = %q, want canonical %q",
					alias, result, canonical)
			}
		}
	})

	t.Run("numeric tokens round-trip to canonical suffixed form", func(t *testing.T) {
		tokens := map[string]string{
			"1":    "1m",
			"5":    "5m",
			"15":   "15m",
			"30":   "30m",
			"60":   "1h",
			"120":  "2h",
			"240":  "4h",
			"1440": "1D",
		}
		for input, want := range tokens {
			secs := converter.ToSeconds(input)
			got := converter.FromSeconds(secs)
			if got != want {
				t.Errorf("FromSeconds(ToSeconds(%q)) = %q, want %q", input, got, want)
			}
		}
	})
}
