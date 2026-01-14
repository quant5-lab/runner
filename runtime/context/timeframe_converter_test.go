package context

import "testing"

func TestTimeframeConverter_ToSeconds(t *testing.T) {
	converter := NewTimeframeConverter()

	tests := []struct {
		name        string
		timeframe   string
		expected    int64
		description string
	}{
		{
			name:        "second resolution",
			timeframe:   "1s",
			expected:    1,
			description: "1 second should convert to 1",
		},
		{
			name:        "minute resolution",
			timeframe:   "5m",
			expected:    300,
			description: "5 minutes should convert to 300 seconds",
		},
		{
			name:        "hour resolution",
			timeframe:   "1h",
			expected:    3600,
			description: "1 hour should convert to 3600 seconds",
		},
		{
			name:        "multi hour",
			timeframe:   "4h",
			expected:    14400,
			description: "4 hours should convert to 14400 seconds",
		},
		{
			name:        "daily uppercase",
			timeframe:   "1D",
			expected:    86400,
			description: "1 day (uppercase) should convert to 86400 seconds",
		},
		{
			name:        "daily lowercase",
			timeframe:   "1d",
			expected:    86400,
			description: "1 day (lowercase) should convert to 86400 seconds",
		},
		{
			name:        "weekly uppercase",
			timeframe:   "1W",
			expected:    604800,
			description: "1 week (uppercase) should convert to 604800 seconds",
		},
		{
			name:        "weekly lowercase",
			timeframe:   "1w",
			expected:    604800,
			description: "1 week (lowercase) should convert to 604800 seconds",
		},
		{
			name:        "monthly",
			timeframe:   "1M",
			expected:    2592000,
			description: "1 month should convert to 2592000 seconds (30 days)",
		},
		{
			name:        "single char daily",
			timeframe:   "D",
			expected:    86400,
			description: "single char D should convert to 86400 (PineScript shorthand)",
		},
		{
			name:        "single char weekly",
			timeframe:   "W",
			expected:    604800,
			description: "single char W should convert to 604800 (PineScript shorthand)",
		},
		{
			name:        "single char monthly",
			timeframe:   "M",
			expected:    2592000,
			description: "single char M should convert to 2592000 (PineScript shorthand)",
		},
		{
			name:        "empty string",
			timeframe:   "",
			expected:    0,
			description: "empty string should return 0",
		},
		{
			name:        "invalid unit",
			timeframe:   "5x",
			expected:    0,
			description: "invalid unit should return 0",
		},
		{
			name:        "large multiplier",
			timeframe:   "240h",
			expected:    864000,
			description: "240 hours should convert correctly",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := converter.ToSeconds(tt.timeframe)
			if result != tt.expected {
				t.Errorf("%s: ToSeconds(%q) = %d, expected %d",
					tt.description, tt.timeframe, result, tt.expected)
			}
		})
	}
}

func TestTimeframeConverter_EdgeCases(t *testing.T) {
	converter := NewTimeframeConverter()

	t.Run("zero multiplier defaults to 1", func(t *testing.T) {
		result := converter.ToSeconds("0m")
		expected := int64(60) // Should default to 1m
		if result != expected {
			t.Errorf("zero multiplier should default to 1: got %d, expected %d", result, expected)
		}
	})

	t.Run("no number defaults to 1", func(t *testing.T) {
		result := converter.ToSeconds("h")
		expected := int64(3600) // Should be 1h
		if result != expected {
			t.Errorf("no number should default to 1h: got %d, expected %d", result, expected)
		}
	})

	t.Run("case sensitivity for units", func(t *testing.T) {
		upperD := converter.ToSeconds("1D")
		lowerD := converter.ToSeconds("1d")
		if upperD != lowerD || upperD != 86400 {
			t.Errorf("uppercase and lowercase D should be equivalent: %d vs %d", upperD, lowerD)
		}
	})
}
