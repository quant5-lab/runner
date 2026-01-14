package context

import "testing"

func TestTimestampBarMatcher_MatchBarForTimestamp(t *testing.T) {
	matcher := NewTimestampBarMatcher()

	secCtx := &Context{
		Data: []OHLCV{
			{Time: 0, Open: 100.0},
			{Time: 86400, Open: 101.0},
			{Time: 172800, Open: 102.0},
			{Time: 259200, Open: 103.0},
		},
	}

	tests := []struct {
		name        string
		timestamp   int64
		expected    int
		description string
	}{
		{
			name:        "timestamp in first period",
			timestamp:   50000,
			expected:    0,
			description: "should return first bar when timestamp falls within it",
		},
		{
			name:        "timestamp at second period start",
			timestamp:   86400,
			expected:    1,
			description: "should return bar when timestamp matches period start",
		},
		{
			name:        "timestamp in third period",
			timestamp:   200000,
			expected:    2,
			description: "should return third bar when timestamp falls within it",
		},
		{
			name:        "timestamp beyond last bar",
			timestamp:   500000,
			expected:    3,
			description: "should return last bar when beyond all data (no future peeking)",
		},
		{
			name:        "timestamp before first bar",
			timestamp:   -1000,
			expected:    -1,
			description: "should return -1 when no bar exists for timestamp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matcher.MatchBarForTimestamp(secCtx, tt.timestamp)
			if result != tt.expected {
				t.Errorf("%s: MatchBarForTimestamp(%d) = %d, expected %d",
					tt.description, tt.timestamp, result, tt.expected)
			}
		})
	}
}

func TestTimestampBarMatcher_MatchBarWithLookahead(t *testing.T) {
	matcher := NewTimestampBarMatcher()

	secCtx := &Context{
		Data: []OHLCV{
			{Time: 0, Open: 100.0},
			{Time: 86400, Open: 101.0},
			{Time: 172800, Open: 102.0},
		},
	}

	tests := []struct {
		name        string
		timestamp   int64
		expected    int
		description string
	}{
		{
			name:        "lookahead returns current bar",
			timestamp:   50000,
			expected:    0,
			description: "lookahead=on means current bar, not next bar",
		},
		{
			name:        "lookahead at boundary",
			timestamp:   86400,
			expected:    1,
			description: "at boundary, lookahead still returns current bar",
		},
		{
			name:        "lookahead beyond last bar",
			timestamp:   300000,
			expected:    2,
			description: "beyond last bar, lookahead returns last bar",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matcher.MatchBarWithLookahead(secCtx, tt.timestamp)
			if result != tt.expected {
				t.Errorf("%s: MatchBarWithLookahead(%d) = %d, expected %d",
					tt.description, tt.timestamp, result, tt.expected)
			}
		})
	}
}

func TestTimestampBarMatcher_EmptyContext(t *testing.T) {
	matcher := NewTimestampBarMatcher()

	emptyCtx := &Context{Data: []OHLCV{}}

	t.Run("empty context standard match", func(t *testing.T) {
		result := matcher.MatchBarForTimestamp(emptyCtx, 100000)
		if result != -1 {
			t.Errorf("empty context should return -1, got %d", result)
		}
	})

	t.Run("empty context lookahead match", func(t *testing.T) {
		result := matcher.MatchBarWithLookahead(emptyCtx, 100000)
		if result != -1 {
			t.Errorf("empty context with lookahead should return -1, got %d", result)
		}
	})
}

func TestTimestampBarMatcher_RealWorldScenario_DailyValues(t *testing.T) {
	matcher := NewTimestampBarMatcher()

	// Real scenario: Daily bars for Dec 16-18, 2024
	dec16 := int64(1734307200)
	dec17 := int64(1734393600)
	dec18 := int64(1734480000)

	dailyCtx := &Context{
		Data: []OHLCV{
			{Time: dec16, Open: 87000.00},
			{Time: dec17, Open: 87863.43},
			{Time: dec18, Open: 88500.00},
		},
	}

	tests := []struct {
		name         string
		hourlyTime   int64
		expectedBar  int
		expectedOpen float64
		description  string
	}{
		{
			name:         "Dec 17 morning",
			hourlyTime:   dec17 + 10*3600, // Dec 17 10:00
			expectedBar:  1,
			expectedOpen: 87863.43,
			description:  "hourly bars during Dec 17 should match Dec 17 daily bar",
		},
		{
			name:         "Dec 17 boundary",
			hourlyTime:   dec17,
			expectedBar:  1,
			expectedOpen: 87863.43,
			description:  "at daily boundary should match that day",
		},
		{
			name:         "Dec 16 afternoon",
			hourlyTime:   dec16 + 15*3600, // Dec 16 15:00
			expectedBar:  0,
			expectedOpen: 87000.00,
			description:  "Dec 16 hourly bars should match Dec 16 daily bar",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			barIdx := matcher.MatchBarForTimestamp(dailyCtx, tt.hourlyTime)

			if barIdx != tt.expectedBar {
				t.Errorf("%s: expected bar %d, got %d",
					tt.description, tt.expectedBar, barIdx)
			}

			if barIdx >= 0 && barIdx < len(dailyCtx.Data) {
				actualOpen := dailyCtx.Data[barIdx].Open
				if actualOpen != tt.expectedOpen {
					t.Errorf("%s: expected open %.2f, got %.2f",
						tt.description, tt.expectedOpen, actualOpen)
				}
			}
		})
	}
}

func TestTimestampBarMatcher_LookaheadSemantics(t *testing.T) {
	matcher := NewTimestampBarMatcher()

	// Real dates: Dec 16-18, 2024
	dec16 := int64(1734307200) // Dec 16, 2024 00:00 UTC
	dec17 := int64(1734393600) // Dec 17, 2024 00:00 UTC
	dec18 := int64(1734480000) // Dec 18, 2024 00:00 UTC

	ctx := &Context{
		Data: []OHLCV{
			{Time: dec16},
			{Time: dec17},
			{Time: dec18},
		},
	}

	t.Run("lookahead matches by calendar date", func(t *testing.T) {
		// Timestamp during Dec 17 (10 hours after midnight)
		dec17At10AM := dec17 + 10*3600

		standardIdx := matcher.MatchBarForTimestamp(ctx, dec17At10AM)
		lookaheadIdx := matcher.MatchBarWithLookahead(ctx, dec17At10AM)

		// Standard: finds containing bar (Dec 17)
		if standardIdx != 1 {
			t.Errorf("standard match should return bar 1 (Dec 17), got %d", standardIdx)
		}

		// Lookahead: matches by calendar date (Dec 17)
		if lookaheadIdx != 1 {
			t.Errorf("lookahead should match Dec 17 bar (index 1), got %d", lookaheadIdx)
		}
	})
}
