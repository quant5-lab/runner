package request

import (
	"testing"

	"github.com/quant5-lab/runner/runtime/context"
)

// TestSecurityBarMapper_NonOverlappingDateRanges tests the fix for when Daily and Hourly data
// have different start dates (e.g., Daily starts Aug 15, Hourly starts Jul 8)
func TestSecurityBarMapper_NonOverlappingDateRanges(t *testing.T) {
	tests := []struct {
		name               string
		higherTFBars       []context.OHLCV
		lowerTFBars        []context.OHLCV
		expectedRangeCount int
		firstRangeStart    int // Expected StartHourlyIndex of first range
		description        string
	}{
		{
			name: "hourly data starts before daily data",
			higherTFBars: []context.OHLCV{
				{Time: parseTime("2025-08-15 14:30:00"), Close: 100}, // Daily starts Aug 15
				{Time: parseTime("2025-08-18 14:30:00"), Close: 110}, // Next daily bar
			},
			lowerTFBars: []context.OHLCV{
				// Hourly starts Jul 8 (38 days before Daily)
				{Time: parseTime("2025-07-08 14:30:00"), Close: 50},
				{Time: parseTime("2025-07-08 15:30:00"), Close: 51},
				// ... many hourly bars ...
				{Time: parseTime("2025-08-15 14:30:00"), Close: 100}, // First overlap at index 2
				{Time: parseTime("2025-08-15 15:30:00"), Close: 101},
				{Time: parseTime("2025-08-18 14:30:00"), Close: 110},
			},
			expectedRangeCount: 2,
			firstRangeStart:    2, // Should skip to hourly index 2 (first Aug 15 bar)
			description:        "should skip hourly bars before first daily bar",
		},
		{
			name: "daily data starts before hourly data",
			higherTFBars: []context.OHLCV{
				{Time: parseTime("2025-01-01 14:30:00"), Close: 100}, // Daily starts Jan 1
				{Time: parseTime("2025-01-02 14:30:00"), Close: 110},
				{Time: parseTime("2025-01-05 14:30:00"), Close: 120}, // Daily at Jan 5
			},
			lowerTFBars: []context.OHLCV{
				// Hourly starts Jan 5 (after first 2 Daily bars)
				{Time: parseTime("2025-01-05 14:30:00"), Close: 120},
				{Time: parseTime("2025-01-05 15:30:00"), Close: 121},
			},
			expectedRangeCount: 1, // Only Jan 5 has hourly data
			firstRangeStart:    0,
			description:        "should only build ranges where hourly data exists",
		},
		{
			name: "exact date alignment - no skipping needed",
			higherTFBars: []context.OHLCV{
				{Time: parseTime("2025-01-01 14:30:00"), Close: 100},
				{Time: parseTime("2025-01-02 14:30:00"), Close: 110},
			},
			lowerTFBars: []context.OHLCV{
				{Time: parseTime("2025-01-01 14:30:00"), Close: 100},
				{Time: parseTime("2025-01-01 15:30:00"), Close: 101},
				{Time: parseTime("2025-01-02 14:30:00"), Close: 110},
			},
			expectedRangeCount: 2,
			firstRangeStart:    0, // No skipping needed
			description:        "should work normally when dates align",
		},
		{
			name: "partial overlap - middle section",
			higherTFBars: []context.OHLCV{
				{Time: parseTime("2025-01-01 14:30:00"), Close: 100},
				{Time: parseTime("2025-01-02 14:30:00"), Close: 110},
				{Time: parseTime("2025-01-03 14:30:00"), Close: 120},
				{Time: parseTime("2025-01-04 14:30:00"), Close: 130},
			},
			lowerTFBars: []context.OHLCV{
				// Hourly only for Jan 2-3 (middle of Daily range)
				{Time: parseTime("2025-01-02 14:30:00"), Close: 110},
				{Time: parseTime("2025-01-02 15:30:00"), Close: 111},
				{Time: parseTime("2025-01-03 14:30:00"), Close: 120},
			},
			expectedRangeCount: 2, // Only Jan 2 and Jan 3
			firstRangeStart:    0,
			description:        "should handle partial overlap in middle of range",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mapper := NewSecurityBarMapper()
			mapper.BuildMappingWithDateFilter(tt.higherTFBars, tt.lowerTFBars, DateRange{}, "UTC")

			if len(mapper.ranges) != tt.expectedRangeCount {
				t.Errorf("%s: expected %d ranges, got %d", tt.description, tt.expectedRangeCount, len(mapper.ranges))
			}

			if len(mapper.ranges) > 0 && mapper.ranges[0].StartHourlyIndex != tt.firstRangeStart {
				t.Errorf("%s: expected first range StartHourlyIndex=%d, got %d",
					tt.description, tt.firstRangeStart, mapper.ranges[0].StartHourlyIndex)
			}
		})
	}
}

// TestSecurityBarMapper_DownscalingModes tests all three security() modes with deterministic data
func TestSecurityBarMapper_DownscalingModes(t *testing.T) {
	tests := []struct {
		name        string
		mode        MappingMode
		setupMapper func(*SecurityBarMapper)
		testCases   []struct {
			sourceIndex int
			lookahead   bool
			expected    int
			description string
		}
	}{
		{
			name: "Downscaling H→D with BuildMappingWithDateFilter",
			mode: ModeDownscaling,
			setupMapper: func(m *SecurityBarMapper) {
				dailyBars := []context.OHLCV{
					{Time: parseTime("2025-01-01 14:30:00"), Close: 100}, // Day 0
					{Time: parseTime("2025-01-02 14:30:00"), Close: 110}, // Day 1
					{Time: parseTime("2025-01-05 14:30:00"), Close: 120}, // Day 2 (weekend gap)
				}
				hourlyBars := []context.OHLCV{
					// Day 0: hourly indices 0-6
					{Time: parseTime("2025-01-01 14:30:00"), Close: 100},
					{Time: parseTime("2025-01-01 15:30:00"), Close: 101},
					{Time: parseTime("2025-01-01 16:30:00"), Close: 102},
					{Time: parseTime("2025-01-01 17:30:00"), Close: 103},
					{Time: parseTime("2025-01-01 18:30:00"), Close: 104},
					{Time: parseTime("2025-01-01 19:30:00"), Close: 105},
					{Time: parseTime("2025-01-01 20:00:00"), Close: 106},
					// Day 1: hourly indices 7-9
					{Time: parseTime("2025-01-02 14:30:00"), Close: 110},
					{Time: parseTime("2025-01-02 15:30:00"), Close: 111},
					{Time: parseTime("2025-01-02 16:30:00"), Close: 112},
					// Day 2: hourly indices 10-12
					{Time: parseTime("2025-01-05 14:30:00"), Close: 120},
					{Time: parseTime("2025-01-05 15:30:00"), Close: 121},
					{Time: parseTime("2025-01-05 16:30:00"), Close: 122},
				}
				m.BuildMappingWithDateFilter(dailyBars, hourlyBars, DateRange{}, "UTC")
			},
			testCases: []struct {
				sourceIndex int
				lookahead   bool
				expected    int
				description string
			}{
				// First range (Day 0) - Critical edge case for first-bar fix
				{0, true, 0, "First hourly bar, lookahead=true → current Daily bar (Day 0)"},
				{0, false, 0, "First hourly bar, lookahead=false → current Daily bar (Day 0, FIXED)"},
				{1, true, 0, "Second hourly bar, lookahead=true → current Daily bar (Day 0)"},
				{1, false, 0, "Second hourly bar, lookahead=false → current Daily bar (Day 0, FIXED)"},
				{6, true, 0, "Last hourly of Day 0, lookahead=true → current Daily bar"},
				{6, false, 0, "Last hourly of Day 0, lookahead=false → current Daily bar (FIXED)"},

				// Second range (Day 1)
				{7, true, 1, "First hourly of Day 1, lookahead=true → current Daily bar (Day 1)"},
				{7, false, 0, "First hourly of Day 1, lookahead=false → previous Daily bar (Day 0)"},
				{8, true, 1, "Mid hourly of Day 1, lookahead=true → current Daily bar (Day 1)"},
				{8, false, 0, "Mid hourly of Day 1, lookahead=false → previous Daily bar (Day 0)"},

				// Third range (Day 2, after weekend gap)
				{10, true, 2, "First hourly of Day 2, lookahead=true → current Daily bar (Day 2)"},
				{10, false, 1, "First hourly of Day 2, lookahead=false → previous Daily bar (Day 1)"},
				{12, true, 2, "Last hourly of Day 2, lookahead=true → current Daily bar (Day 2)"},
				{12, false, 1, "Last hourly of Day 2, lookahead=false → previous Daily bar (Day 1)"},

				// Out of bounds
				{13, true, 2, "Beyond last hourly, lookahead=true → last Daily bar"},
				{13, false, 2, "Beyond last hourly, lookahead=false → last Daily bar"},
				{-1, true, -1, "Negative index → -1"},
				{-1, false, -1, "Negative index → -1"},
			},
		},
		{
			name: "Upscaling W→D with BuildMappingForUpscaling",
			mode: ModeUpscaling,
			setupMapper: func(m *SecurityBarMapper) {
				dailyBars := []context.OHLCV{
					// Week 0: Daily bars 0-4 (Mon-Fri)
					{Time: parseTime("2025-01-06 00:00:00"), Close: 100}, // Monday
					{Time: parseTime("2025-01-07 00:00:00"), Close: 101}, // Tuesday
					{Time: parseTime("2025-01-08 00:00:00"), Close: 102}, // Wednesday
					{Time: parseTime("2025-01-09 00:00:00"), Close: 103}, // Thursday
					{Time: parseTime("2025-01-10 00:00:00"), Close: 104}, // Friday
					// Week 1: Daily bars 5-9
					{Time: parseTime("2025-01-13 00:00:00"), Close: 110},
					{Time: parseTime("2025-01-14 00:00:00"), Close: 111},
					{Time: parseTime("2025-01-15 00:00:00"), Close: 112},
					{Time: parseTime("2025-01-16 00:00:00"), Close: 113},
					{Time: parseTime("2025-01-17 00:00:00"), Close: 114},
				}
				weeklyBars := []context.OHLCV{
					{Time: parseTime("2025-01-06 00:00:00"), Close: 100}, // Week 0
					{Time: parseTime("2025-01-13 00:00:00"), Close: 110}, // Week 1
				}
				m.BuildMappingForUpscaling(dailyBars, weeklyBars, "UTC")
			},
			testCases: []struct {
				sourceIndex int
				lookahead   bool
				expected    int
				description string
			}{
				// Week 0: Maps to Daily bars 0-4
				{0, true, 4, "Week 0, lookahead=true → end of week (Friday, Daily 4)"},
				{0, false, 0, "Week 0, lookahead=false → start of week (Monday, Daily 0)"},

				// Week 1: Maps to Daily bars 5-9
				{1, true, 9, "Week 1, lookahead=true → end of week (Friday, Daily 9)"},
				{1, false, 5, "Week 1, lookahead=false → start of week (Monday, Daily 5)"},

				// Out of bounds
				{2, true, -1, "Beyond last weekly bar → -1"},
				{-1, false, -1, "Negative index → -1"},
			},
		},
		{
			name: "Same timeframe (special case)",
			mode: ModeDownscaling,
			setupMapper: func(m *SecurityBarMapper) {
				// When security() uses same timeframe, BuildMappingWithDateFilter creates 3 ranges:
				// Range 0: hourly 0 → daily 0
				// Range 1: hourly 1 → daily 1
				// Range 2: hourly 2 → daily 2
				// All bars are on same date, so each gets its own range
				bars := []context.OHLCV{
					{Time: parseTime("2025-01-01 14:30:00"), Close: 100},
					{Time: parseTime("2025-01-01 15:30:00"), Close: 101},
					{Time: parseTime("2025-01-01 16:30:00"), Close: 102},
				}
				m.BuildMappingWithDateFilter(bars, bars, DateRange{}, "UTC")
			},
			testCases: []struct {
				sourceIndex int
				lookahead   bool
				expected    int
				description string
			}{
				// When same TF, all bars on same date creates single range [0-2]→0
				{0, true, 0, "Same TF, index 0, lookahead=true → daily bar 0"},
				{0, false, 0, "Same TF, index 0, lookahead=false → daily bar 0 (FIXED)"},
				{1, true, 0, "Same TF, index 1, lookahead=true → daily bar 0 (all in same day)"},
				{1, false, 0, "Same TF, index 1, lookahead=false → daily bar 0 (previous in same day)"},
				{2, true, 0, "Same TF, index 2, lookahead=true → daily bar 0 (all in same day)"},
				{2, false, 0, "Same TF, index 2, lookahead=false → daily bar 0 (previous in same day)"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mapper := NewSecurityBarMapper()
			tt.setupMapper(mapper)

			if mapper.mode != tt.mode {
				t.Fatalf("Expected mode %d, got %d", tt.mode, mapper.mode)
			}

			for _, tc := range tt.testCases {
				t.Run(tc.description, func(t *testing.T) {
					result := mapper.FindDailyBarIndex(tc.sourceIndex, tc.lookahead)
					if result != tc.expected {
						t.Errorf("%s: sourceIndex=%d lookahead=%v: expected %d, got %d",
							tc.description, tc.sourceIndex, tc.lookahead, tc.expected, result)
					}
				})
			}
		})
	}
}

// TestSecurityBarMapper_TimezoneMarketHours tests date boundary handling across timezones
func TestSecurityBarMapper_TimezoneMarketHours(t *testing.T) {
	tests := []struct {
		name               string
		timezone           string
		dailyBars          []context.OHLCV
		hourlyBars         []context.OHLCV
		expectedRangeCount int
		description        string
	}{
		{
			name:     "UTC timezone - midnight boundary",
			timezone: "UTC",
			dailyBars: []context.OHLCV{
				{Time: parseTime("2025-01-01 14:30:00"), Close: 100},
				{Time: parseTime("2025-01-02 14:30:00"), Close: 110},
			},
			hourlyBars: []context.OHLCV{
				{Time: parseTime("2025-01-01 23:00:00"), Close: 100}, // 11 PM on Jan 1
				{Time: parseTime("2025-01-02 00:00:00"), Close: 101}, // Midnight - belongs to Jan 2
				{Time: parseTime("2025-01-02 01:00:00"), Close: 102},
			},
			expectedRangeCount: 2,
			description:        "midnight UTC should be start of new day",
		},
		{
			name:     "America/New_York timezone - market hours",
			timezone: "America/New_York",
			dailyBars: []context.OHLCV{
				{Time: parseTime("2025-01-01 14:30:00"), Close: 100}, // 9:30 AM EST (market open)
				{Time: parseTime("2025-01-02 14:30:00"), Close: 110},
			},
			hourlyBars: []context.OHLCV{
				{Time: parseTime("2025-01-01 14:30:00"), Close: 100}, // 9:30 AM EST
				{Time: parseTime("2025-01-01 15:30:00"), Close: 101}, // 10:30 AM EST
				{Time: parseTime("2025-01-01 20:00:00"), Close: 102}, // 3:00 PM EST (market close)
				{Time: parseTime("2025-01-02 14:30:00"), Close: 110},
			},
			expectedRangeCount: 2,
			description:        "should handle US market hours correctly",
		},
		{
			name:     "Empty timezone defaults to UTC",
			timezone: "",
			dailyBars: []context.OHLCV{
				{Time: parseTime("2025-01-01 14:30:00"), Close: 100},
			},
			hourlyBars: []context.OHLCV{
				{Time: parseTime("2025-01-01 14:30:00"), Close: 100},
			},
			expectedRangeCount: 1,
			description:        "empty timezone should default to UTC",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mapper := NewSecurityBarMapper()
			mapper.BuildMappingWithDateFilter(tt.dailyBars, tt.hourlyBars, DateRange{}, tt.timezone)

			if len(mapper.ranges) != tt.expectedRangeCount {
				t.Errorf("%s: expected %d ranges, got %d", tt.description, tt.expectedRangeCount, len(mapper.ranges))
			}
		})
	}
}

// TestSecurityBarMapper_ExtremeCases tests pathological edge cases
func TestSecurityBarMapper_ExtremeCases(t *testing.T) {
	tests := []struct {
		name        string
		setupMapper func(*SecurityBarMapper)
		testIndex   int
		lookahead   bool
		expected    int
		description string
	}{
		{
			name: "Empty ranges - should return -1",
			setupMapper: func(m *SecurityBarMapper) {
				// Don't build any mapping
				m.mode = ModeDownscaling
				m.ranges = []BarRange{}
			},
			testIndex:   0,
			lookahead:   true,
			expected:    -1,
			description: "empty ranges should always return -1",
		},
		{
			name: "Single range, single bar",
			setupMapper: func(m *SecurityBarMapper) {
				dailyBars := []context.OHLCV{
					{Time: parseTime("2025-01-01 14:30:00"), Close: 100},
				}
				hourlyBars := []context.OHLCV{
					{Time: parseTime("2025-01-01 14:30:00"), Close: 100},
				}
				m.BuildMappingWithDateFilter(dailyBars, hourlyBars, DateRange{}, "UTC")
			},
			testIndex:   0,
			lookahead:   false,
			expected:    0,
			description: "single bar should return itself (FIXED: was returning -1)",
		},
		{
			name: "Large index beyond all ranges",
			setupMapper: func(m *SecurityBarMapper) {
				dailyBars := []context.OHLCV{
					{Time: parseTime("2025-01-01 14:30:00"), Close: 100},
				}
				hourlyBars := []context.OHLCV{
					{Time: parseTime("2025-01-01 14:30:00"), Close: 100},
				}
				m.BuildMappingWithDateFilter(dailyBars, hourlyBars, DateRange{}, "UTC")
			},
			testIndex:   999999,
			lookahead:   true,
			expected:    0,
			description: "index beyond all ranges should return last Daily bar",
		},
		{
			name: "Very dense hourly data - 24 bars per day",
			setupMapper: func(m *SecurityBarMapper) {
				dailyBars := []context.OHLCV{
					{Time: parseTime("2025-01-01 00:00:00"), Close: 100},
				}
				// Create 24 hourly bars for one day
				hourlyBars := make([]context.OHLCV, 24)
				for i := 0; i < 24; i++ {
					hourlyBars[i] = context.OHLCV{
						Time:  parseTime("2025-01-01 00:00:00") + int64(i*3600),
						Close: 100 + float64(i),
					}
				}
				m.BuildMappingWithDateFilter(dailyBars, hourlyBars, DateRange{}, "UTC")
			},
			testIndex:   0,
			lookahead:   false,
			expected:    0,
			description: "dense hourly data should still work correctly (FIXED)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mapper := NewSecurityBarMapper()
			tt.setupMapper(mapper)

			result := mapper.FindDailyBarIndex(tt.testIndex, tt.lookahead)
			if result != tt.expected {
				t.Errorf("%s: expected %d, got %d", tt.description, tt.expected, result)
			}
		})
	}
}

// TestSecurityBarMapper_RangeIntegrity validates that ranges maintain internal consistency
func TestSecurityBarMapper_RangeIntegrity(t *testing.T) {
	mapper := NewSecurityBarMapper()

	dailyBars := []context.OHLCV{
		{Time: parseTime("2025-01-01 14:30:00"), Close: 100},
		{Time: parseTime("2025-01-02 14:30:00"), Close: 110},
		{Time: parseTime("2025-01-03 14:30:00"), Close: 120},
	}

	hourlyBars := []context.OHLCV{
		{Time: parseTime("2025-01-01 14:30:00"), Close: 100},
		{Time: parseTime("2025-01-01 15:30:00"), Close: 101},
		{Time: parseTime("2025-01-01 16:30:00"), Close: 102},
		{Time: parseTime("2025-01-02 14:30:00"), Close: 110},
		{Time: parseTime("2025-01-02 15:30:00"), Close: 111},
		{Time: parseTime("2025-01-03 14:30:00"), Close: 120},
		{Time: parseTime("2025-01-03 15:30:00"), Close: 121},
		{Time: parseTime("2025-01-03 16:30:00"), Close: 122},
	}

	mapper.BuildMappingWithDateFilter(dailyBars, hourlyBars, DateRange{}, "UTC")

	// Validate ranges
	if len(mapper.ranges) != 3 {
		t.Fatalf("Expected 3 ranges, got %d", len(mapper.ranges))
	}

	// Check range continuity - ranges should be non-overlapping and sequential
	for i := 1; i < len(mapper.ranges); i++ {
		prevRange := mapper.ranges[i-1]
		currRange := mapper.ranges[i]

		// Current range should start immediately after previous range ends
		if currRange.StartHourlyIndex != prevRange.EndHourlyIndex+1 {
			t.Errorf("Range discontinuity: range[%d].EndHourlyIndex=%d, range[%d].StartHourlyIndex=%d",
				i-1, prevRange.EndHourlyIndex, i, currRange.StartHourlyIndex)
		}

		// Daily bar indices should be sequential
		if currRange.DailyBarIndex != prevRange.DailyBarIndex+1 {
			t.Errorf("Non-sequential DailyBarIndex: range[%d].DailyBarIndex=%d, range[%d].DailyBarIndex=%d",
				i-1, prevRange.DailyBarIndex, i, currRange.DailyBarIndex)
		}
	}

	// Validate first range
	firstRange := mapper.ranges[0]
	if firstRange.DailyBarIndex != 0 {
		t.Errorf("First range DailyBarIndex should be 0, got %d", firstRange.DailyBarIndex)
	}
	if firstRange.StartHourlyIndex != 0 {
		t.Errorf("First range StartHourlyIndex should be 0, got %d", firstRange.StartHourlyIndex)
	}

	// Validate last range
	lastRange := mapper.ranges[len(mapper.ranges)-1]
	if lastRange.EndHourlyIndex != len(hourlyBars)-1 {
		t.Errorf("Last range EndHourlyIndex should be %d, got %d",
			len(hourlyBars)-1, lastRange.EndHourlyIndex)
	}
}
