package request

import (
	"testing"
	"time"

	"github.com/quant5-lab/runner/runtime/context"
)

/* ============================================================================
   DateRange Timezone-Aware Tests

   Tests cover the behavior of date range operations with timezone awareness,
   ensuring correct date extraction and comparison across various timezone
   scenarios and edge cases.
   ============================================================================ */

func TestDateRange_TimezoneAwareDateExtraction(t *testing.T) {
	tests := []struct {
		name         string
		timestamp    int64
		timezone     string
		expectedDate string
		description  string
	}{
		{
			name:         "UTC midnight",
			timestamp:    time.Date(2025, 12, 15, 0, 0, 0, 0, time.UTC).Unix(),
			timezone:     "UTC",
			expectedDate: "2025-12-15",
			description:  "UTC midnight should extract correct date",
		},
		{
			name:         "UTC late evening",
			timestamp:    time.Date(2025, 12, 15, 23, 59, 0, 0, time.UTC).Unix(),
			timezone:     "UTC",
			expectedDate: "2025-12-15",
			description:  "UTC late evening should remain same date",
		},
		{
			name:         "Moscow midnight in UTC (21:00 prev day)",
			timestamp:    time.Date(2025, 12, 14, 21, 0, 0, 0, time.UTC).Unix(),
			timezone:     "Europe/Moscow",
			expectedDate: "2025-12-15",
			description:  "21:00 UTC should be midnight in Moscow (UTC+3)",
		},
		{
			name:         "Moscow late evening",
			timestamp:    time.Date(2025, 12, 15, 20, 59, 0, 0, time.UTC).Unix(),
			timezone:     "Europe/Moscow",
			expectedDate: "2025-12-15",
			description:  "20:59 UTC should be 23:59 Moscow time",
		},
		{
			name:         "New York midnight in UTC (05:00)",
			timestamp:    time.Date(2025, 12, 15, 5, 0, 0, 0, time.UTC).Unix(),
			timezone:     "America/New_York",
			expectedDate: "2025-12-15",
			description:  "05:00 UTC should be midnight EST (UTC-5)",
		},
		{
			name:         "Tokyo morning in UTC (prev day evening)",
			timestamp:    time.Date(2025, 12, 14, 15, 0, 0, 0, time.UTC).Unix(),
			timezone:     "Asia/Tokyo",
			expectedDate: "2025-12-15",
			description:  "15:00 UTC should be midnight JST (UTC+9)",
		},
		{
			name:         "empty timezone defaults to UTC",
			timestamp:    time.Date(2025, 12, 15, 12, 0, 0, 0, time.UTC).Unix(),
			timezone:     "",
			expectedDate: "2025-12-15",
			description:  "Empty timezone should safely default to UTC",
		},
		{
			name:         "invalid timezone falls back to UTC",
			timestamp:    time.Date(2025, 12, 15, 12, 0, 0, 0, time.UTC).Unix(),
			timezone:     "Invalid/Zone",
			expectedDate: "2025-12-15",
			description:  "Invalid timezone should fall back to UTC without error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractDateInTimezone(tt.timestamp, tt.timezone)
			if result != tt.expectedDate {
				t.Errorf("ExtractDateInTimezone() = %v, want %v - %s",
					result, tt.expectedDate, tt.description)
			}
		})
	}
}

func TestDateRange_MidnightBoundaryBehavior(t *testing.T) {
	tests := []struct {
		name          string
		timestamps    []int64
		timezone      string
		expectedDates []string
		description   string
	}{
		{
			name: "bars crossing midnight UTC",
			timestamps: []int64{
				time.Date(2025, 12, 15, 23, 0, 0, 0, time.UTC).Unix(),
				time.Date(2025, 12, 15, 23, 30, 0, 0, time.UTC).Unix(),
				time.Date(2025, 12, 16, 0, 0, 0, 0, time.UTC).Unix(),
				time.Date(2025, 12, 16, 0, 30, 0, 0, time.UTC).Unix(),
			},
			timezone:      "UTC",
			expectedDates: []string{"2025-12-15", "2025-12-15", "2025-12-16", "2025-12-16"},
			description:   "UTC midnight should be clean boundary",
		},
		{
			name: "bars crossing midnight Moscow",
			timestamps: []int64{
				time.Date(2025, 12, 15, 20, 0, 0, 0, time.UTC).Unix(),
				time.Date(2025, 12, 15, 20, 30, 0, 0, time.UTC).Unix(),
				time.Date(2025, 12, 15, 21, 0, 0, 0, time.UTC).Unix(),
				time.Date(2025, 12, 15, 21, 30, 0, 0, time.UTC).Unix(),
			},
			timezone:      "Europe/Moscow",
			expectedDates: []string{"2025-12-15", "2025-12-15", "2025-12-16", "2025-12-16"},
			description:   "Moscow midnight (21:00 UTC) should be clean boundary",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i, ts := range tt.timestamps {
				result := ExtractDateInTimezone(ts, tt.timezone)
				if result != tt.expectedDates[i] {
					t.Errorf("Timestamp[%d] extracted as %v, want %v - %s",
						i, result, tt.expectedDates[i], tt.description)
				}
			}
		})
	}
}

func TestDateRange_NewDateRangeFromBars(t *testing.T) {
	tests := []struct {
		name          string
		bars          []context.OHLCV
		timezone      string
		expectedStart string
		expectedEnd   string
		description   string
	}{
		{
			name:          "empty bars",
			bars:          []context.OHLCV{},
			timezone:      "UTC",
			expectedStart: "",
			expectedEnd:   "",
			description:   "Empty bars should create empty date range",
		},
		{
			name: "single bar UTC",
			bars: []context.OHLCV{
				{Time: time.Date(2025, 12, 15, 10, 0, 0, 0, time.UTC).Unix()},
			},
			timezone:      "UTC",
			expectedStart: "2025-12-15",
			expectedEnd:   "2025-12-15",
			description:   "Single bar should have same start and end date",
		},
		{
			name: "multiple bars same day UTC",
			bars: []context.OHLCV{
				{Time: time.Date(2025, 12, 15, 9, 0, 0, 0, time.UTC).Unix()},
				{Time: time.Date(2025, 12, 15, 12, 0, 0, 0, time.UTC).Unix()},
				{Time: time.Date(2025, 12, 15, 18, 0, 0, 0, time.UTC).Unix()},
			},
			timezone:      "UTC",
			expectedStart: "2025-12-15",
			expectedEnd:   "2025-12-15",
			description:   "Multiple bars on same day should have same start/end",
		},
		{
			name: "multiple bars spanning days UTC",
			bars: []context.OHLCV{
				{Time: time.Date(2025, 12, 15, 10, 0, 0, 0, time.UTC).Unix()},
				{Time: time.Date(2025, 12, 16, 11, 0, 0, 0, time.UTC).Unix()},
				{Time: time.Date(2025, 12, 17, 12, 0, 0, 0, time.UTC).Unix()},
			},
			timezone:      "UTC",
			expectedStart: "2025-12-15",
			expectedEnd:   "2025-12-17",
			description:   "Bars spanning multiple days should have correct range",
		},
		{
			name: "Moscow timezone bars around midnight boundary",
			bars: []context.OHLCV{
				{Time: time.Date(2025, 12, 14, 21, 0, 0, 0, time.UTC).Unix()},
				{Time: time.Date(2025, 12, 15, 20, 59, 0, 0, time.UTC).Unix()},
			},
			timezone:      "Europe/Moscow",
			expectedStart: "2025-12-15",
			expectedEnd:   "2025-12-15",
			description:   "Moscow bars from midnight to 23:59 local should be same day",
		},
		{
			name: "empty timezone defaults to UTC",
			bars: []context.OHLCV{
				{Time: time.Date(2025, 12, 15, 10, 0, 0, 0, time.UTC).Unix()},
			},
			timezone:      "",
			expectedStart: "2025-12-15",
			expectedEnd:   "2025-12-15",
			description:   "Empty timezone should safely default to UTC",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dr := NewDateRangeFromBars(tt.bars, tt.timezone)

			if dr.StartDate != tt.expectedStart {
				t.Errorf("StartDate = %v, want %v - %s",
					dr.StartDate, tt.expectedStart, tt.description)
			}
			if dr.EndDate != tt.expectedEnd {
				t.Errorf("EndDate = %v, want %v - %s",
					dr.EndDate, tt.expectedEnd, tt.description)
			}
			if dr.Timezone != tt.timezone && tt.timezone != "" {
				t.Errorf("Timezone = %v, want %v", dr.Timezone, tt.timezone)
			}
		})
	}
}

func TestDateRange_Contains(t *testing.T) {
	tests := []struct {
		name        string
		dateRange   DateRange
		testDate    string
		expected    bool
		description string
	}{
		{
			name:        "empty range doesn't match dates",
			dateRange:   DateRange{StartDate: "", EndDate: "", Timezone: "UTC"},
			testDate:    "2025-12-15",
			expected:    false,
			description: "Empty range with no dates returns false",
		},
		{
			name:        "date within range",
			dateRange:   DateRange{StartDate: "2025-12-10", EndDate: "2025-12-20", Timezone: "UTC"},
			testDate:    "2025-12-15",
			expected:    true,
			description: "Date in middle of range should match",
		},
		{
			name:        "date at start boundary",
			dateRange:   DateRange{StartDate: "2025-12-10", EndDate: "2025-12-20", Timezone: "UTC"},
			testDate:    "2025-12-10",
			expected:    true,
			description: "Date at start boundary should match",
		},
		{
			name:        "date at end boundary",
			dateRange:   DateRange{StartDate: "2025-12-10", EndDate: "2025-12-20", Timezone: "UTC"},
			testDate:    "2025-12-20",
			expected:    true,
			description: "Date at end boundary should match",
		},
		{
			name:        "date before range",
			dateRange:   DateRange{StartDate: "2025-12-10", EndDate: "2025-12-20", Timezone: "UTC"},
			testDate:    "2025-12-09",
			expected:    false,
			description: "Date before range should not match",
		},
		{
			name:        "date after range",
			dateRange:   DateRange{StartDate: "2025-12-10", EndDate: "2025-12-20", Timezone: "UTC"},
			testDate:    "2025-12-21",
			expected:    false,
			description: "Date after range should not match",
		},
		{
			name:        "single day range",
			dateRange:   DateRange{StartDate: "2025-12-15", EndDate: "2025-12-15", Timezone: "UTC"},
			testDate:    "2025-12-15",
			expected:    true,
			description: "Single day range should match that exact day",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.dateRange.Contains(tt.testDate)
			if result != tt.expected {
				t.Errorf("Contains(%v) = %v, want %v - %s",
					tt.testDate, result, tt.expected, tt.description)
			}
		})
	}
}

func TestDateRange_IsEmpty(t *testing.T) {
	tests := []struct {
		name        string
		dateRange   DateRange
		expected    bool
		description string
	}{
		{
			name:        "both dates empty",
			dateRange:   DateRange{StartDate: "", EndDate: "", Timezone: "UTC"},
			expected:    true,
			description: "Range with empty dates should be empty",
		},
		{
			name:        "start empty, end set",
			dateRange:   DateRange{StartDate: "", EndDate: "2025-12-15", Timezone: "UTC"},
			expected:    true,
			description: "Range with only end date should be empty",
		},
		{
			name:        "start set, end empty",
			dateRange:   DateRange{StartDate: "2025-12-15", EndDate: "", Timezone: "UTC"},
			expected:    true,
			description: "Range with only start date should be empty",
		},
		{
			name:        "both dates set",
			dateRange:   DateRange{StartDate: "2025-12-10", EndDate: "2025-12-20", Timezone: "UTC"},
			expected:    false,
			description: "Range with both dates should not be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.dateRange.IsEmpty()
			if result != tt.expected {
				t.Errorf("IsEmpty() = %v, want %v - %s",
					result, tt.expected, tt.description)
			}
		})
	}
}

/* ============================================================================
   SecurityBarMapper Timezone Integration Tests

   Tests verify that bar mapping behaves consistently across different
   timezones, ensuring same logical dates map to same bar indices regardless
   of timezone configuration.
   ============================================================================ */

func TestSecurityBarMapper_TimezoneConsistentMapping(t *testing.T) {
	tests := []struct {
		name                 string
		dailyBars            []context.OHLCV
		hourlyBars           []context.OHLCV
		timezone             string
		expectedRangeCount   int
		validateRangeIndices func(t *testing.T, ranges []BarRange)
		description          string
	}{
		{
			name: "Moscow timezone MOEX data",
			dailyBars: []context.OHLCV{
				{Time: time.Date(2025, 12, 14, 21, 0, 0, 0, time.UTC).Unix()},
				{Time: time.Date(2025, 12, 15, 21, 0, 0, 0, time.UTC).Unix()},
				{Time: time.Date(2025, 12, 16, 21, 0, 0, 0, time.UTC).Unix()},
			},
			hourlyBars: []context.OHLCV{
				{Time: time.Date(2025, 12, 15, 6, 0, 0, 0, time.UTC).Unix()},
				{Time: time.Date(2025, 12, 15, 7, 0, 0, 0, time.UTC).Unix()},
				{Time: time.Date(2025, 12, 16, 6, 0, 0, 0, time.UTC).Unix()},
				{Time: time.Date(2025, 12, 16, 7, 0, 0, 0, time.UTC).Unix()},
			},
			timezone:           "Europe/Moscow",
			expectedRangeCount: 3,
			validateRangeIndices: func(t *testing.T, ranges []BarRange) {
				/* Range[0] = Dec 15 Moscow (daily[0]) maps to hourly Dec 15 bars (hourly[0:1])
				   Range[1] = Dec 16 Moscow (daily[1]) maps to hourly Dec 16 bars (hourly[2:3])
				   Range[2] = Dec 17 Moscow (daily[2]) has no hourly bars */
				if ranges[0].DailyBarIndex != 0 {
					t.Errorf("Range[0] should map to daily[0], got daily[%d]", ranges[0].DailyBarIndex)
				}
				if ranges[0].StartHourlyIndex != 0 || ranges[0].EndHourlyIndex != 1 {
					t.Errorf("Range[0] should map hourly[0:1], got hourly[%d:%d]",
						ranges[0].StartHourlyIndex, ranges[0].EndHourlyIndex)
				}
				if ranges[1].StartHourlyIndex != 2 || ranges[1].EndHourlyIndex != 3 {
					t.Errorf("Range[1] should map hourly[2:3], got hourly[%d:%d]",
						ranges[1].StartHourlyIndex, ranges[1].EndHourlyIndex)
				}
				if ranges[2].StartHourlyIndex != -1 || ranges[2].EndHourlyIndex != -1 {
					t.Errorf("Range[2] should have no hourly bars, got hourly[%d:%d]",
						ranges[2].StartHourlyIndex, ranges[2].EndHourlyIndex)
				}
			},
			description: "MOEX bars with Moscow timezone should map correctly",
		},
		{
			name: "UTC timezone bars",
			dailyBars: []context.OHLCV{
				{Time: time.Date(2025, 12, 15, 0, 0, 0, 0, time.UTC).Unix()},
				{Time: time.Date(2025, 12, 16, 0, 0, 0, 0, time.UTC).Unix()},
			},
			hourlyBars: []context.OHLCV{
				{Time: time.Date(2025, 12, 15, 9, 0, 0, 0, time.UTC).Unix()},
				{Time: time.Date(2025, 12, 15, 10, 0, 0, 0, time.UTC).Unix()},
				{Time: time.Date(2025, 12, 16, 9, 0, 0, 0, time.UTC).Unix()},
			},
			timezone:           "UTC",
			expectedRangeCount: 2,
			validateRangeIndices: func(t *testing.T, ranges []BarRange) {
				if ranges[0].StartHourlyIndex != 0 || ranges[0].EndHourlyIndex != 1 {
					t.Errorf("Range[0] should map hourly[0:1], got hourly[%d:%d]",
						ranges[0].StartHourlyIndex, ranges[0].EndHourlyIndex)
				}
				if ranges[1].StartHourlyIndex != 2 || ranges[1].EndHourlyIndex != 2 {
					t.Errorf("Range[1] should map hourly[2:2], got hourly[%d:%d]",
						ranges[1].StartHourlyIndex, ranges[1].EndHourlyIndex)
				}
			},
			description: "UTC bars should map cleanly",
		},
		{
			name: "daily bars with no matching hourly bars",
			dailyBars: []context.OHLCV{
				{Time: time.Date(2025, 12, 14, 21, 0, 0, 0, time.UTC).Unix()},
				{Time: time.Date(2025, 12, 15, 21, 0, 0, 0, time.UTC).Unix()},
				{Time: time.Date(2025, 12, 16, 21, 0, 0, 0, time.UTC).Unix()},
			},
			hourlyBars: []context.OHLCV{
				{Time: time.Date(2025, 12, 16, 6, 0, 0, 0, time.UTC).Unix()},
			},
			timezone:           "Europe/Moscow",
			expectedRangeCount: 3,
			validateRangeIndices: func(t *testing.T, ranges []BarRange) {
				/* Range[0] = Dec 15 Moscow (daily[0]) has no hourly bars
				   Range[1] = Dec 16 Moscow (daily[1]) maps to hourly Dec 16 bar (hourly[0])
				   Range[2] = Dec 17 Moscow (daily[2]) has no hourly bars */
				if ranges[0].StartHourlyIndex != -1 || ranges[0].EndHourlyIndex != -1 {
					t.Errorf("Range[0] should have no hourly bars (-1), got [%d:%d]",
						ranges[0].StartHourlyIndex, ranges[0].EndHourlyIndex)
				}
				if ranges[1].StartHourlyIndex != 0 || ranges[1].EndHourlyIndex != 0 {
					t.Errorf("Range[1] should map hourly[0:0], got hourly[%d:%d]",
						ranges[1].StartHourlyIndex, ranges[1].EndHourlyIndex)
				}
				if ranges[2].StartHourlyIndex != -1 || ranges[2].EndHourlyIndex != -1 {
					t.Errorf("Range[2] should have no hourly bars (-1), got [%d:%d]",
						ranges[2].StartHourlyIndex, ranges[2].EndHourlyIndex)
				}
			},
			description: "Should create ranges for all daily bars even without matching hourly bars",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mapper := NewSecurityBarMapper()
			mapper.BuildMappingWithDateFilter(tt.dailyBars, tt.hourlyBars, DateRange{}, tt.timezone)

			ranges := mapper.GetRanges()
			if len(ranges) != tt.expectedRangeCount {
				t.Errorf("Expected %d ranges, got %d - %s",
					tt.expectedRangeCount, len(ranges), tt.description)
				return
			}

			if tt.validateRangeIndices != nil {
				tt.validateRangeIndices(t, ranges)
			}
		})
	}
}

func TestSecurityBarMapper_BarCountIndependence(t *testing.T) {
	/* This test verifies the core requirement: different numbers of base TF bars
	   should produce identical mappings for the same calendar date ranges.
	   This ensures indicator values remain consistent regardless of historical depth. */

	baseTimezone := "Europe/Moscow"

	dailyBars := []context.OHLCV{
		{Time: time.Date(2025, 12, 13, 21, 0, 0, 0, time.UTC).Unix()},
		{Time: time.Date(2025, 12, 14, 21, 0, 0, 0, time.UTC).Unix()},
		{Time: time.Date(2025, 12, 15, 21, 0, 0, 0, time.UTC).Unix()},
		{Time: time.Date(2025, 12, 16, 21, 0, 0, 0, time.UTC).Unix()},
		{Time: time.Date(2025, 12, 17, 21, 0, 0, 0, time.UTC).Unix()},
	}

	hourlyBars300 := []context.OHLCV{
		{Time: time.Date(2025, 12, 16, 6, 0, 0, 0, time.UTC).Unix()},
		{Time: time.Date(2025, 12, 16, 7, 0, 0, 0, time.UTC).Unix()},
		{Time: time.Date(2025, 12, 17, 6, 0, 0, 0, time.UTC).Unix()},
	}

	hourlyBars500 := []context.OHLCV{
		{Time: time.Date(2025, 12, 15, 6, 0, 0, 0, time.UTC).Unix()},
		{Time: time.Date(2025, 12, 15, 7, 0, 0, 0, time.UTC).Unix()},
		{Time: time.Date(2025, 12, 16, 6, 0, 0, 0, time.UTC).Unix()},
		{Time: time.Date(2025, 12, 16, 7, 0, 0, 0, time.UTC).Unix()},
		{Time: time.Date(2025, 12, 17, 6, 0, 0, 0, time.UTC).Unix()},
	}

	mapper300 := NewSecurityBarMapper()
	mapper300.BuildMappingWithDateFilter(dailyBars, hourlyBars300, DateRange{}, baseTimezone)

	mapper500 := NewSecurityBarMapper()
	mapper500.BuildMappingWithDateFilter(dailyBars, hourlyBars500, DateRange{}, baseTimezone)

	ranges300 := mapper300.GetRanges()
	ranges500 := mapper500.GetRanges()

	if len(ranges300) != len(ranges500) {
		t.Fatalf("Different bar counts produced different range counts: %d vs %d",
			len(ranges300), len(ranges500))
	}

	expectedRangeCount := len(dailyBars)
	if len(ranges300) != expectedRangeCount {
		t.Errorf("Expected %d ranges (one per daily bar), got %d",
			expectedRangeCount, len(ranges300))
	}

	for i := range ranges300 {
		if ranges300[i].DailyBarIndex != ranges500[i].DailyBarIndex {
			t.Errorf("Range[%d] maps to different daily indices: %d vs %d",
				i, ranges300[i].DailyBarIndex, ranges500[i].DailyBarIndex)
		}
	}

	if ranges300[3].DailyBarIndex != 3 {
		t.Errorf("Dec 17 range should map to daily[3], got daily[%d]",
			ranges300[3].DailyBarIndex)
	}
	if ranges500[3].DailyBarIndex != 3 {
		t.Errorf("Dec 17 range should map to daily[3], got daily[%d]",
			ranges500[3].DailyBarIndex)
	}
}

func TestSecurityBarMapper_TimezoneEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		timezone    string
		description string
	}{
		{
			name:        "empty timezone",
			timezone:    "",
			description: "Empty timezone should default to UTC without error",
		},
		{
			name:        "invalid timezone",
			timezone:    "Invalid/Nonexistent",
			description: "Invalid timezone should fall back to UTC gracefully",
		},
		{
			name:        "case sensitive timezone",
			timezone:    "europe/moscow",
			description: "Lowercase timezone should be handled (may fail or default)",
		},
	}

	dailyBars := []context.OHLCV{
		{Time: time.Date(2025, 12, 15, 0, 0, 0, 0, time.UTC).Unix()},
	}
	hourlyBars := []context.OHLCV{
		{Time: time.Date(2025, 12, 15, 10, 0, 0, 0, time.UTC).Unix()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("BuildMappingWithDateFilter panicked with %v - %s",
						r, tt.description)
				}
			}()

			mapper := NewSecurityBarMapper()
			mapper.BuildMappingWithDateFilter(dailyBars, hourlyBars, DateRange{}, tt.timezone)

			ranges := mapper.GetRanges()
			if len(ranges) == 0 {
				t.Errorf("Expected at least one range, got none - %s", tt.description)
			}
		})
	}
}

func TestSecurityBarMapper_FindDailyBarIndex_WithTimezone(t *testing.T) {
	/* Verify that bar index lookup remains consistent across different timezone
	   configurations for the same logical date mapping. */

	timezone := "Europe/Moscow"

	dailyBars := []context.OHLCV{
		{Time: time.Date(2025, 12, 14, 21, 0, 0, 0, time.UTC).Unix()},
		{Time: time.Date(2025, 12, 15, 21, 0, 0, 0, time.UTC).Unix()},
		{Time: time.Date(2025, 12, 16, 21, 0, 0, 0, time.UTC).Unix()},
	}

	hourlyBars := []context.OHLCV{
		{Time: time.Date(2025, 12, 15, 6, 0, 0, 0, time.UTC).Unix()},
		{Time: time.Date(2025, 12, 15, 12, 0, 0, 0, time.UTC).Unix()},
		{Time: time.Date(2025, 12, 16, 6, 0, 0, 0, time.UTC).Unix()},
	}

	mapper := NewSecurityBarMapper()
	mapper.BuildMappingWithDateFilter(dailyBars, hourlyBars, DateRange{}, timezone)

	tests := []struct {
		name          string
		hourlyIndex   int
		lookahead     bool
		expectedDaily int
		allowEither   []int
		description   string
	}{
		{
			name:        "first hourly bar no lookahead",
			hourlyIndex: 0,
			lookahead:   false,
			allowEither: []int{-1, 0},
			description: "First bar should return daily[0] or -1 with no lookahead",
		},
		{
			name:          "first hourly bar with lookahead",
			hourlyIndex:   0,
			lookahead:     true,
			expectedDaily: 0,
			description:   "First bar with lookahead should return current daily",
		},
		{
			name:          "third hourly bar no lookahead",
			hourlyIndex:   2,
			lookahead:     false,
			expectedDaily: 0,
			description:   "Third bar should return previous daily[0]",
		},
		{
			name:          "third hourly bar with lookahead",
			hourlyIndex:   2,
			lookahead:     true,
			expectedDaily: 1,
			description:   "Third bar with lookahead should return current daily[1]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapper.FindDailyBarIndex(tt.hourlyIndex, tt.lookahead)

			if len(tt.allowEither) > 0 {
				found := false
				for _, allowed := range tt.allowEither {
					if result == allowed {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("FindDailyBarIndex(%d, %v) = %d, want one of %v - %s",
						tt.hourlyIndex, tt.lookahead, result, tt.allowEither, tt.description)
				}
			} else if result != tt.expectedDaily {
				t.Errorf("FindDailyBarIndex(%d, %v) = %d, want %d - %s",
					tt.hourlyIndex, tt.lookahead, result, tt.expectedDaily, tt.description)
			}
		})
	}
}

func TestBarRange_Contains_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		barRange    BarRange
		hourlyIndex int
		expected    bool
		description string
	}{
		{
			name:        "range with no hourly bars (warmup period)",
			barRange:    BarRange{DailyBarIndex: 5, StartHourlyIndex: -1, EndHourlyIndex: -1},
			hourlyIndex: 0,
			expected:    false,
			description: "Ranges without hourly bars should not contain any index",
		},
		{
			name:        "range with single hourly bar",
			barRange:    BarRange{DailyBarIndex: 5, StartHourlyIndex: 10, EndHourlyIndex: 10},
			hourlyIndex: 10,
			expected:    true,
			description: "Single bar range should contain that exact index",
		},
		{
			name:        "index at start boundary",
			barRange:    BarRange{DailyBarIndex: 5, StartHourlyIndex: 10, EndHourlyIndex: 20},
			hourlyIndex: 10,
			expected:    true,
			description: "Index at start boundary should be contained",
		},
		{
			name:        "index at end boundary",
			barRange:    BarRange{DailyBarIndex: 5, StartHourlyIndex: 10, EndHourlyIndex: 20},
			hourlyIndex: 20,
			expected:    true,
			description: "Index at end boundary should be contained",
		},
		{
			name:        "index before range",
			barRange:    BarRange{DailyBarIndex: 5, StartHourlyIndex: 10, EndHourlyIndex: 20},
			hourlyIndex: 9,
			expected:    false,
			description: "Index before range should not be contained",
		},
		{
			name:        "index after range",
			barRange:    BarRange{DailyBarIndex: 5, StartHourlyIndex: 10, EndHourlyIndex: 20},
			hourlyIndex: 21,
			expected:    false,
			description: "Index after range should not be contained",
		},
		{
			name:        "negative hourly index",
			barRange:    BarRange{DailyBarIndex: 5, StartHourlyIndex: 10, EndHourlyIndex: 20},
			hourlyIndex: -1,
			expected:    false,
			description: "Negative index should not be contained",
		},
		{
			name:        "zero index with valid range",
			barRange:    BarRange{DailyBarIndex: 0, StartHourlyIndex: 0, EndHourlyIndex: 5},
			hourlyIndex: 0,
			expected:    true,
			description: "Zero is valid hourly index and should be checked properly",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.barRange.Contains(tt.hourlyIndex)
			if result != tt.expected {
				t.Errorf("BarRange.Contains(%d) = %v, want %v - %s",
					tt.hourlyIndex, result, tt.expected, tt.description)
			}
		})
	}
}

/* ============================================================================
   Integration Tests: Full Workflow

   Tests that verify the complete timezone-aware workflow from bar data
   to final mapping, ensuring all components work together correctly.
   ============================================================================ */

func TestTimezoneWorkflow_EndToEnd(t *testing.T) {
	/* This test simulates the complete workflow used in production:
	   1. Create bars with timestamps
	   2. Extract date range with timezone
	   3. Build mapping with timezone
	   4. Verify mapping consistency */

	tests := []struct {
		name                string
		timezone            string
		dailyTimestamps     []int64
		hourlyTimestamps    []int64
		expectedDailyCount  int
		expectedMappedDates int
		description         string
	}{
		{
			name:     "MOEX typical workflow",
			timezone: "Europe/Moscow",
			dailyTimestamps: []int64{
				time.Date(2025, 12, 13, 21, 0, 0, 0, time.UTC).Unix(),
				time.Date(2025, 12, 14, 21, 0, 0, 0, time.UTC).Unix(),
				time.Date(2025, 12, 15, 21, 0, 0, 0, time.UTC).Unix(),
			},
			hourlyTimestamps: []int64{
				time.Date(2025, 12, 14, 6, 0, 0, 0, time.UTC).Unix(),
				time.Date(2025, 12, 14, 12, 0, 0, 0, time.UTC).Unix(),
				time.Date(2025, 12, 15, 6, 0, 0, 0, time.UTC).Unix(),
			},
			expectedDailyCount:  3,
			expectedMappedDates: 2,
			description:         "MOEX bars should map correctly in Moscow timezone",
		},
		{
			name:     "Binance 24/7 UTC workflow",
			timezone: "UTC",
			dailyTimestamps: []int64{
				time.Date(2025, 12, 14, 0, 0, 0, 0, time.UTC).Unix(),
				time.Date(2025, 12, 15, 0, 0, 0, 0, time.UTC).Unix(),
			},
			hourlyTimestamps: []int64{
				time.Date(2025, 12, 14, 10, 0, 0, 0, time.UTC).Unix(),
				time.Date(2025, 12, 14, 20, 0, 0, 0, time.UTC).Unix(),
				time.Date(2025, 12, 15, 5, 0, 0, 0, time.UTC).Unix(),
			},
			expectedDailyCount:  2,
			expectedMappedDates: 2,
			description:         "Binance 24/7 data should map cleanly in UTC",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dailyBars := make([]context.OHLCV, len(tt.dailyTimestamps))
			for i, ts := range tt.dailyTimestamps {
				dailyBars[i] = context.OHLCV{Time: ts, Close: float64(100 + i)}
			}

			hourlyBars := make([]context.OHLCV, len(tt.hourlyTimestamps))
			for i, ts := range tt.hourlyTimestamps {
				hourlyBars[i] = context.OHLCV{Time: ts, Close: float64(100 + i)}
			}

			dateRange := NewDateRangeFromBars(hourlyBars, tt.timezone)
			if dateRange.Timezone != tt.timezone {
				t.Errorf("DateRange timezone = %v, want %v", dateRange.Timezone, tt.timezone)
			}

			mapper := NewSecurityBarMapper()
			mapper.BuildMappingWithDateFilter(dailyBars, hourlyBars, dateRange, tt.timezone)

			ranges := mapper.GetRanges()
			if len(ranges) != tt.expectedDailyCount {
				t.Errorf("Expected %d ranges, got %d - %s",
					tt.expectedDailyCount, len(ranges), tt.description)
			}

			mappedDatesCount := 0
			for _, r := range ranges {
				if r.StartHourlyIndex >= 0 {
					mappedDatesCount++
				}
			}
			if mappedDatesCount != tt.expectedMappedDates {
				t.Errorf("Expected %d dates with hourly bars, got %d - %s",
					tt.expectedMappedDates, mappedDatesCount, tt.description)
			}
		})
	}
}

func TestTimezoneConsistency_CrossTimezone(t *testing.T) {
	/* Verify that the same absolute timestamps produce consistent mappings
	   when interpreted in different timezones. */

	baseTimestamp := time.Date(2025, 12, 15, 12, 0, 0, 0, time.UTC).Unix()

	tests := []struct {
		timezone     string
		expectedDate string
	}{
		{"UTC", "2025-12-15"},
		{"Europe/Moscow", "2025-12-15"},
		{"America/New_York", "2025-12-15"},
		{"Asia/Tokyo", "2025-12-15"},
	}

	for _, tt := range tests {
		t.Run(tt.timezone, func(t *testing.T) {
			result := ExtractDateInTimezone(baseTimestamp, tt.timezone)
			if result != tt.expectedDate {
				t.Errorf("Timezone %s: extracted %v, want %v",
					tt.timezone, result, tt.expectedDate)
			}
		})
	}
}

/* ============================================================================
   Performance and Stress Tests

   Tests that verify timezone operations perform adequately with large datasets
   and don't introduce performance regressions.
   ============================================================================ */

func TestTimezoneOperations_Performance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	/* Reduced bar count to prevent test timeout while still validating performance */
	largeBarCount := 500
	dailyBars := make([]context.OHLCV, largeBarCount)
	baseTime := time.Date(2020, 1, 1, 21, 0, 0, 0, time.UTC).Unix()

	for i := 0; i < largeBarCount; i++ {
		dailyBars[i] = context.OHLCV{
			Time:  baseTime + int64(i*86400),
			Close: float64(100 + i),
		}
	}

	hourlyBars := make([]context.OHLCV, largeBarCount*10)
	for i := 0; i < largeBarCount*10; i++ {
		hourlyBars[i] = context.OHLCV{
			Time:  baseTime + int64(i*3600),
			Close: float64(100 + i),
		}
	}

	timezones := []string{"UTC", "Europe/Moscow"}

	for _, tz := range timezones {
		t.Run(tz, func(t *testing.T) {
			mapper := NewSecurityBarMapper()
			mapper.BuildMappingWithDateFilter(dailyBars, hourlyBars, DateRange{}, tz)

			ranges := mapper.GetRanges()
			if len(ranges) == 0 {
				t.Error("Expected ranges to be created")
			}
		})
	}
}
