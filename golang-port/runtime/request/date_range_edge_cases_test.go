package request

import (
	"testing"
	"time"

	"github.com/quant5-lab/runner/runtime/context"
)

/* ============================================================================
   DateRange Edge Cases

   Comprehensive edge case tests for timezone-aware date range operations.
   Tests cover boundary conditions, extreme values, and unusual inputs.
   ============================================================================ */

func TestDateRange_MidnightTransitionEdgeCases(t *testing.T) {
	tests := []struct {
		name         string
		timestamp    int64
		timezone     string
		expectedDate string
		description  string
	}{
		{
			name:         "UTC midnight exactly",
			timestamp:    time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC).Unix(),
			timezone:     "UTC",
			expectedDate: "2025-01-01",
			description:  "Exact midnight should extract correct date",
		},
		{
			name:         "one nanosecond before UTC midnight",
			timestamp:    time.Date(2024, 12, 31, 23, 59, 59, 999999999, time.UTC).Unix(),
			timezone:     "UTC",
			expectedDate: "2024-12-31",
			description:  "One ns before midnight should remain previous day",
		},
		{
			name:         "Moscow midnight in UTC (21:00:00 exact)",
			timestamp:    time.Date(2025, 1, 1, 21, 0, 0, 0, time.UTC).Unix(),
			timezone:     "Europe/Moscow",
			expectedDate: "2025-01-02",
			description:  "Exact Moscow midnight boundary",
		},
		{
			name:         "one second before Moscow midnight",
			timestamp:    time.Date(2025, 1, 1, 20, 59, 59, 0, time.UTC).Unix(),
			timezone:     "Europe/Moscow",
			expectedDate: "2025-01-01",
			description:  "One second before Moscow midnight should remain previous day",
		},
		{
			name:         "New York midnight EST (05:00 UTC)",
			timestamp:    time.Date(2025, 1, 2, 5, 0, 0, 0, time.UTC).Unix(),
			timezone:     "America/New_York",
			expectedDate: "2025-01-02",
			description:  "EST midnight boundary",
		},
		{
			name:         "Tokyo midnight JST (15:00 prev day UTC)",
			timestamp:    time.Date(2025, 1, 1, 15, 0, 0, 0, time.UTC).Unix(),
			timezone:     "Asia/Tokyo",
			expectedDate: "2025-01-02",
			description:  "JST midnight boundary (UTC+9)",
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

func TestDateRange_YearBoundaryTransitions(t *testing.T) {
	tests := []struct {
		name         string
		timestamp    int64
		timezone     string
		expectedDate string
		description  string
	}{
		{
			name:         "New Year UTC midnight",
			timestamp:    time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC).Unix(),
			timezone:     "UTC",
			expectedDate: "2025-01-01",
			description:  "New Year should extract correctly",
		},
		{
			name:         "Last second of year UTC",
			timestamp:    time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC).Unix(),
			timezone:     "UTC",
			expectedDate: "2024-12-31",
			description:  "Year end should remain in old year",
		},
		{
			name:         "New Year Moscow time (21:00 UTC Dec 31)",
			timestamp:    time.Date(2024, 12, 31, 21, 0, 0, 0, time.UTC).Unix(),
			timezone:     "Europe/Moscow",
			expectedDate: "2025-01-01",
			description:  "Moscow New Year happens 3 hours before UTC",
		},
		{
			name:         "New Year Tokyo time (15:00 UTC Dec 31)",
			timestamp:    time.Date(2024, 12, 31, 15, 0, 0, 0, time.UTC).Unix(),
			timezone:     "Asia/Tokyo",
			expectedDate: "2025-01-01",
			description:  "Tokyo New Year happens 9 hours before UTC",
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

func TestDateRange_InvalidTimezoneHandling(t *testing.T) {
	tests := []struct {
		name           string
		timezone       string
		shouldNotPanic bool
		description    string
	}{
		{
			name:           "completely invalid timezone",
			timezone:       "Invalid/Nonexistent",
			shouldNotPanic: true,
			description:    "Should fallback to UTC without panic",
		},
		{
			name:           "empty string timezone",
			timezone:       "",
			shouldNotPanic: true,
			description:    "Empty timezone should default to UTC",
		},
		{
			name:           "whitespace timezone",
			timezone:       "   ",
			shouldNotPanic: true,
			description:    "Whitespace timezone should be handled gracefully",
		},
		{
			name:           "special characters",
			timezone:       "@@##$$%%",
			shouldNotPanic: true,
			description:    "Special characters should not cause panic",
		},
		{
			name:           "very long string",
			timezone:       "ThisIsAVeryLongStringThatExceedsNormalTimezoneLengthsAndShouldStillBeHandledGracefully",
			shouldNotPanic: true,
			description:    "Long invalid timezone should not crash",
		},
		{
			name:           "null-like string",
			timezone:       "null",
			shouldNotPanic: true,
			description:    "String 'null' should be handled as invalid timezone",
		},
	}

	timestamp := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC).Unix()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					if tt.shouldNotPanic {
						t.Errorf("ExtractDateInTimezone panicked with %v - %s", r, tt.description)
					}
				}
			}()

			result := ExtractDateInTimezone(timestamp, tt.timezone)
			if result == "" {
				t.Errorf("ExtractDateInTimezone returned empty string - %s", tt.description)
			}
		})
	}
}

func TestDateRange_NewFromBarsEmptyAndNilCases(t *testing.T) {
	tests := []struct {
		name        string
		bars        []context.OHLCV
		timezone    string
		expectEmpty bool
		description string
	}{
		{
			name:        "nil bars",
			bars:        nil,
			timezone:    "UTC",
			expectEmpty: true,
			description: "Nil bars should create empty range",
		},
		{
			name:        "empty slice",
			bars:        []context.OHLCV{},
			timezone:    "UTC",
			expectEmpty: true,
			description: "Empty slice should create empty range",
		},
		{
			name:        "zero-capacity slice",
			bars:        make([]context.OHLCV, 0, 0),
			timezone:    "UTC",
			expectEmpty: true,
			description: "Zero-capacity slice should create empty range",
		},
		{
			name: "single bar with epoch zero",
			bars: []context.OHLCV{
				{Time: 0, Close: 100.0},
			},
			timezone:    "UTC",
			expectEmpty: false,
			description: "Bar with epoch zero should still create range",
		},
		{
			name: "bars with same timestamp",
			bars: []context.OHLCV{
				{Time: time.Date(2025, 6, 15, 10, 0, 0, 0, time.UTC).Unix()},
				{Time: time.Date(2025, 6, 15, 10, 0, 0, 0, time.UTC).Unix()},
				{Time: time.Date(2025, 6, 15, 10, 0, 0, 0, time.UTC).Unix()},
			},
			timezone:    "UTC",
			expectEmpty: false,
			description: "Multiple bars with same timestamp should create valid range",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dr := NewDateRangeFromBars(tt.bars, tt.timezone)

			if tt.expectEmpty {
				if !dr.IsEmpty() {
					t.Errorf("Expected empty range, got StartDate=%v EndDate=%v - %s",
						dr.StartDate, dr.EndDate, tt.description)
				}
			} else {
				if dr.IsEmpty() {
					t.Errorf("Expected non-empty range, got empty - %s", tt.description)
				}
			}
		})
	}
}

func TestDateRange_ExtremeDateValues(t *testing.T) {
	tests := []struct {
		name           string
		timestamp      int64
		timezone       string
		shouldNotPanic bool
		description    string
	}{
		{
			name:           "Unix epoch zero",
			timestamp:      0,
			timezone:       "UTC",
			shouldNotPanic: true,
			description:    "Epoch zero should extract 1970-01-01",
		},
		{
			name:           "negative timestamp (before epoch)",
			timestamp:      -86400, // One day before epoch
			timezone:       "UTC",
			shouldNotPanic: true,
			description:    "Negative timestamp should be handled",
		},
		{
			name:           "far future timestamp (year 2100)",
			timestamp:      time.Date(2100, 12, 31, 23, 59, 59, 0, time.UTC).Unix(),
			timezone:       "UTC",
			shouldNotPanic: true,
			description:    "Far future dates should work",
		},
		{
			name:           "very large timestamp",
			timestamp:      253402300799, // Max 32-bit Unix time (2038 problem related)
			timezone:       "UTC",
			shouldNotPanic: true,
			description:    "Large timestamps should not crash",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					if tt.shouldNotPanic {
						t.Errorf("ExtractDateInTimezone panicked with %v - %s", r, tt.description)
					}
				}
			}()

			result := ExtractDateInTimezone(tt.timestamp, tt.timezone)
			if result == "" {
				t.Errorf("ExtractDateInTimezone returned empty string for timestamp %d - %s",
					tt.timestamp, tt.description)
			}
		})
	}
}

func TestDateRange_ContainsEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		dateRange   DateRange
		testDate    string
		expected    bool
		description string
	}{
		{
			name:        "empty date string",
			dateRange:   DateRange{StartDate: "2025-01-01", EndDate: "2025-12-31", Timezone: "UTC"},
			testDate:    "",
			expected:    false,
			description: "Empty date string should not match",
		},
		{
			name:        "malformed date (missing year)",
			dateRange:   DateRange{StartDate: "2025-01-01", EndDate: "2025-12-31", Timezone: "UTC"},
			testDate:    "01-15",
			expected:    false,
			description: "Malformed date should not match",
		},
		{
			name:        "malformed date (wrong separator)",
			dateRange:   DateRange{StartDate: "2025-01-01", EndDate: "2025-12-31", Timezone: "UTC"},
			testDate:    "2025/06/15",
			expected:    false,
			description: "Wrong separator should not match",
		},
		{
			name:        "date with time component",
			dateRange:   DateRange{StartDate: "2025-01-01", EndDate: "2025-12-31", Timezone: "UTC"},
			testDate:    "2025-06-15 12:00:00",
			expected:    true,
			description: "Date with time should match via string comparison (starts with date)",
		},
		{
			name:        "whitespace in date",
			dateRange:   DateRange{StartDate: "2025-01-01", EndDate: "2025-12-31", Timezone: "UTC"},
			testDate:    " 2025-06-15 ",
			expected:    false,
			description: "Whitespace-padded date should not match",
		},
		{
			name:        "reverse range (end before start)",
			dateRange:   DateRange{StartDate: "2025-12-31", EndDate: "2025-01-01", Timezone: "UTC"},
			testDate:    "2025-06-15",
			expected:    false,
			description: "Reverse range should not match middle date",
		},
		{
			name:        "single day range exact match",
			dateRange:   DateRange{StartDate: "2025-06-15", EndDate: "2025-06-15", Timezone: "UTC"},
			testDate:    "2025-06-15",
			expected:    true,
			description: "Single day range should match exact date",
		},
		{
			name:        "single day range no match",
			dateRange:   DateRange{StartDate: "2025-06-15", EndDate: "2025-06-15", Timezone: "UTC"},
			testDate:    "2025-06-14",
			expected:    false,
			description: "Single day range should not match adjacent date",
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

func TestDateRange_CrossTimezoneConsistency(t *testing.T) {
	/* Verify same absolute timestamp extracts consistently across multiple timezone conversions */

	baseTimestamp := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC).Unix()

	timezones := []string{
		"UTC",
		"Europe/Moscow",
		"America/New_York",
		"America/Los_Angeles",
		"Asia/Tokyo",
		"Asia/Shanghai",
		"Australia/Sydney",
		"Europe/London",
		"Pacific/Honolulu",
	}

	results := make(map[string]string)
	for _, tz := range timezones {
		results[tz] = ExtractDateInTimezone(baseTimestamp, tz)
	}

	for tz1, date1 := range results {
		for tz2, date2 := range results {
			if tz1 == tz2 {
				continue
			}
			/* Same timestamp should either extract same date or adjacent dates depending on offset */
			if date1 != date2 {
				year1, month1, day1 := parseDate(date1)
				year2, month2, day2 := parseDate(date2)

				/* Dates can differ by at most 1 day due to timezone offsets */
				if year1 != year2 {
					if !(year1 == year2-1 && month1 == 12 && day1 == 31 && month2 == 1 && day2 == 1) {
						t.Errorf("Timezones %s and %s extracted dates with >1 year difference: %v vs %v",
							tz1, tz2, date1, date2)
					}
				} else if month1 == month2 {
					dayDiff := day1 - day2
					if dayDiff < -1 || dayDiff > 1 {
						t.Errorf("Timezones %s and %s extracted dates with >1 day difference: %v vs %v",
							tz1, tz2, date1, date2)
					}
				}
			}
		}
	}
}

func parseDate(dateStr string) (year, month, day int) {
	/* Simple date parser for testing - expects YYYY-MM-DD format */
	if parsed, err := time.Parse("2006-01-02", dateStr); err == nil {
		return parsed.Year(), int(parsed.Month()), parsed.Day()
	}
	return 0, 0, 0
}

func TestDateRange_SequentialDayMapping(t *testing.T) {
	/* Test that sequential days map correctly with different bar counts.
	   Note: Ranges are only created for daily bars that have corresponding hourly data */

	timezone := "Europe/Moscow"

	dailyBars := []context.OHLCV{
		{Time: time.Date(2025, 12, 13, 21, 0, 0, 0, time.UTC).Unix()}, // Dec 14 Moscow
		{Time: time.Date(2025, 12, 14, 21, 0, 0, 0, time.UTC).Unix()}, // Dec 15 Moscow
		{Time: time.Date(2025, 12, 15, 21, 0, 0, 0, time.UTC).Unix()}, // Dec 16 Moscow
		{Time: time.Date(2025, 12, 16, 21, 0, 0, 0, time.UTC).Unix()}, // Dec 17 Moscow
		{Time: time.Date(2025, 12, 17, 21, 0, 0, 0, time.UTC).Unix()}, // Dec 18 Moscow
	}

	/* Test with different hourly bar configurations */
	hourlyConfigs := []struct {
		name           string
		bars           []context.OHLCV
		expectedRanges int
		firstDailyIdx  int
	}{
		{
			name: "all days have hourly bars",
			bars: []context.OHLCV{
				{Time: time.Date(2025, 12, 14, 6, 0, 0, 0, time.UTC).Unix()}, // Dec 14 Moscow
				{Time: time.Date(2025, 12, 15, 6, 0, 0, 0, time.UTC).Unix()}, // Dec 15 Moscow
				{Time: time.Date(2025, 12, 16, 6, 0, 0, 0, time.UTC).Unix()}, // Dec 16 Moscow
				{Time: time.Date(2025, 12, 17, 6, 0, 0, 0, time.UTC).Unix()}, // Dec 17 Moscow
			},
			expectedRanges: 4, // 4 daily bars have hourly data
			firstDailyIdx:  0, // First range maps to dailyBars[0]
		},
		{
			name: "skip middle days",
			bars: []context.OHLCV{
				{Time: time.Date(2025, 12, 14, 6, 0, 0, 0, time.UTC).Unix()}, // Dec 14 Moscow
				{Time: time.Date(2025, 12, 17, 6, 0, 0, 0, time.UTC).Unix()}, // Dec 17 Moscow
			},
			expectedRanges: 2, // Only 2 daily bars have hourly data
			firstDailyIdx:  0, // First range maps to dailyBars[0]
		},
		{
			name: "only last day",
			bars: []context.OHLCV{
				{Time: time.Date(2025, 12, 17, 6, 0, 0, 0, time.UTC).Unix()}, // Dec 17 Moscow
			},
			expectedRanges: 1, // Only 1 daily bar has hourly data
			firstDailyIdx:  3, // First range maps to dailyBars[3] (Dec 17)
		},
	}

	for _, config := range hourlyConfigs {
		t.Run(config.name, func(t *testing.T) {
			mapper := NewSecurityBarMapper()
			mapper.BuildMappingWithDateFilter(dailyBars, config.bars, DateRange{}, timezone)

			ranges := mapper.GetRanges()

			/* Only creates ranges for daily bars with hourly data */
			if len(ranges) != config.expectedRanges {
				t.Errorf("Expected %d ranges, got %d", config.expectedRanges, len(ranges))
			}

			/* Verify first range maps to correct daily bar */
			if len(ranges) > 0 && ranges[0].DailyBarIndex != config.firstDailyIdx {
				t.Errorf("First range should map to daily[%d], got daily[%d]",
					config.firstDailyIdx, ranges[0].DailyBarIndex)
			}
		})
	}
}
