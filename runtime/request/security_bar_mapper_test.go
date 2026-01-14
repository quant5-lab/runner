package request

import (
	"testing"
	"time"

	"github.com/quant5-lab/runner/runtime/context"
)

func TestSecurityBarMapper_BuildMapping(t *testing.T) {
	tests := []struct {
		name               string
		dailyBars          []context.OHLCV
		hourlyBars         []context.OHLCV
		expectedRangeCount int
		description        string
	}{
		{
			name:               "empty bars",
			dailyBars:          []context.OHLCV{},
			hourlyBars:         []context.OHLCV{},
			expectedRangeCount: 0,
			description:        "should handle empty input gracefully",
		},
		{
			name: "single day with multiple hourly bars",
			dailyBars: []context.OHLCV{
				{Time: parseTime("2025-01-01 14:30:00"), Close: 100},
			},
			hourlyBars: []context.OHLCV{
				{Time: parseTime("2025-01-01 14:30:00"), Close: 100},
				{Time: parseTime("2025-01-01 15:30:00"), Close: 101},
				{Time: parseTime("2025-01-01 16:30:00"), Close: 102},
				{Time: parseTime("2025-01-01 17:30:00"), Close: 103},
				{Time: parseTime("2025-01-01 18:30:00"), Close: 104},
				{Time: parseTime("2025-01-01 19:30:00"), Close: 105},
				{Time: parseTime("2025-01-01 20:00:00"), Close: 106},
			},
			expectedRangeCount: 1,
			description:        "should group all hourly bars under single daily bar",
		},
		{
			name: "three days with varying hourly bars",
			dailyBars: []context.OHLCV{
				{Time: parseTime("2025-01-01 14:30:00"), Close: 100},
				{Time: parseTime("2025-01-02 14:30:00"), Close: 110},
				{Time: parseTime("2025-01-03 14:30:00"), Close: 120},
			},
			hourlyBars: []context.OHLCV{
				{Time: parseTime("2025-01-01 14:30:00"), Close: 100},
				{Time: parseTime("2025-01-01 15:30:00"), Close: 101},
				{Time: parseTime("2025-01-01 16:30:00"), Close: 102},
				{Time: parseTime("2025-01-02 14:30:00"), Close: 110},
				{Time: parseTime("2025-01-02 15:30:00"), Close: 111},
				{Time: parseTime("2025-01-03 14:30:00"), Close: 120},
			},
			expectedRangeCount: 3,
			description:        "should create separate ranges for each distinct date",
		},
		{
			name: "single daily bar with single hourly bar",
			dailyBars: []context.OHLCV{
				{Time: parseTime("2025-01-01 14:30:00"), Close: 100},
			},
			hourlyBars: []context.OHLCV{
				{Time: parseTime("2025-01-01 14:30:00"), Close: 100},
			},
			expectedRangeCount: 1,
			description:        "should handle minimal data with single bar per timeframe",
		},
		{
			name: "daily bars with gaps in dates",
			dailyBars: []context.OHLCV{
				{Time: parseTime("2025-01-01 14:30:00"), Close: 100},
				{Time: parseTime("2025-01-05 14:30:00"), Close: 110},
			},
			hourlyBars: []context.OHLCV{
				{Time: parseTime("2025-01-01 14:30:00"), Close: 100},
				{Time: parseTime("2025-01-01 15:30:00"), Close: 101},
				{Time: parseTime("2025-01-05 14:30:00"), Close: 110},
				{Time: parseTime("2025-01-05 15:30:00"), Close: 111},
			},
			expectedRangeCount: 2,
			description:        "should handle non-consecutive dates correctly",
		},
		{
			name: "hourly bars at midnight crossing date boundary",
			dailyBars: []context.OHLCV{
				{Time: parseTime("2025-01-01 14:30:00"), Close: 100},
				{Time: parseTime("2025-01-02 14:30:00"), Close: 110},
			},
			hourlyBars: []context.OHLCV{
				{Time: parseTime("2025-01-01 23:00:00"), Close: 100},
				{Time: parseTime("2025-01-02 00:00:00"), Close: 101},
				{Time: parseTime("2025-01-02 01:00:00"), Close: 102},
			},
			expectedRangeCount: 2,
			description:        "should correctly assign bars to date based on UTC date extraction",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mapper := NewSecurityBarMapper()
			mapper.BuildMapping(tt.dailyBars, tt.hourlyBars)

			if len(mapper.ranges) != tt.expectedRangeCount {
				t.Errorf("%s: expected %d ranges, got %d", tt.description, tt.expectedRangeCount, len(mapper.ranges))
			}
		})
	}
}

func TestSecurityBarMapper_FindDailyBarIndex(t *testing.T) {
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
		{Time: parseTime("2025-01-02 16:30:00"), Close: 112},
		{Time: parseTime("2025-01-03 14:30:00"), Close: 120},
	}

	mapper.BuildMapping(dailyBars, hourlyBars)

	tests := []struct {
		name          string
		hourlyIndex   int
		lookahead     bool
		expectedDaily int
		description   string
	}{
		{
			name:          "first bar of day 1 with lookahead on",
			hourlyIndex:   0,
			lookahead:     true,
			expectedDaily: 0,
			description:   "lookahead=on should return current forming daily bar",
		},
		{
			name:          "first bar of day 1 with lookahead off",
			hourlyIndex:   0,
			lookahead:     false,
			expectedDaily: 0,
			description:   "lookahead=off returns current Daily bar for first range (FIXED)",
		},
		{
			name:          "mid day 1 with lookahead on",
			hourlyIndex:   1,
			lookahead:     true,
			expectedDaily: 0,
			description:   "lookahead=on during day should return current forming bar",
		},
		{
			name:          "mid day 1 with lookahead off",
			hourlyIndex:   1,
			lookahead:     false,
			expectedDaily: 0,
			description:   "lookahead=off returns current Daily bar for first range (FIXED)",
		},
		{
			name:          "last bar of day 1 with lookahead on",
			hourlyIndex:   2,
			lookahead:     true,
			expectedDaily: 0,
			description:   "lookahead=on at last bar should return current forming bar",
		},
		{
			name:          "last bar of day 1 with lookahead off",
			hourlyIndex:   2,
			lookahead:     false,
			expectedDaily: 0,
			description:   "lookahead=off returns current Daily bar for first range (FIXED)",
		},
		{
			name:          "first bar of day 2 with lookahead on",
			hourlyIndex:   3,
			lookahead:     true,
			expectedDaily: 1,
			description:   "lookahead=on at new day start should return new forming bar",
		},
		{
			name:          "first bar of day 2 with lookahead off",
			hourlyIndex:   3,
			lookahead:     false,
			expectedDaily: 0,
			description:   "lookahead=off at new day start should return previous completed bar",
		},
		{
			name:          "mid day 2 with lookahead on",
			hourlyIndex:   4,
			lookahead:     true,
			expectedDaily: 1,
			description:   "lookahead=on mid day 2 should return current forming bar",
		},
		{
			name:          "mid day 2 with lookahead off",
			hourlyIndex:   4,
			lookahead:     false,
			expectedDaily: 0,
			description:   "lookahead=off mid day 2 should return day 1 completed bar",
		},
		{
			name:          "last bar of day 2 with lookahead on",
			hourlyIndex:   5,
			lookahead:     true,
			expectedDaily: 1,
			description:   "lookahead=on at last bar of day 2 should return forming bar",
		},
		{
			name:          "last bar of day 2 with lookahead off",
			hourlyIndex:   5,
			lookahead:     false,
			expectedDaily: 0,
			description:   "lookahead=off at last bar of day 2 should return day 1",
		},
		{
			name:          "first bar of day 3 with lookahead on",
			hourlyIndex:   6,
			lookahead:     true,
			expectedDaily: 2,
			description:   "lookahead=on at day 3 start should return day 3 forming bar",
		},
		{
			name:          "last bar of day 3 with lookahead off",
			hourlyIndex:   6,
			lookahead:     false,
			expectedDaily: 1,
			description:   "lookahead=off at day 3 start should return day 2 completed bar",
		},
		{
			name:          "beyond last hourly bar with lookahead on",
			hourlyIndex:   100,
			lookahead:     true,
			expectedDaily: 2,
			description:   "beyond bounds with lookahead=on should return last daily bar",
		},
		{
			name:          "beyond last hourly bar with lookahead off",
			hourlyIndex:   100,
			lookahead:     false,
			expectedDaily: 2,
			description:   "beyond bounds with lookahead=off should return last daily bar",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapper.FindTargetBarIndexByContainment(tt.hourlyIndex, tt.lookahead)

			if result != tt.expectedDaily {
				t.Errorf("%s: hourlyIndex=%d lookahead=%v: expected daily=%d, got %d",
					tt.description, tt.hourlyIndex, tt.lookahead, tt.expectedDaily, result)
			}
		})
	}
}

func TestSecurityBarMapper_GapScenarios(t *testing.T) {
	tests := []struct {
		name          string
		dailyBars     []context.OHLCV
		hourlyBars    []context.OHLCV
		hourlyIndex   int
		lookahead     bool
		expectedDaily int
		description   string
	}{
		{
			name: "missing daily bars - weekend gap",
			dailyBars: []context.OHLCV{
				{Time: parseTime("2025-01-02 14:30:00"), Close: 100},
				{Time: parseTime("2025-01-05 14:30:00"), Close: 110},
			},
			hourlyBars: []context.OHLCV{
				{Time: parseTime("2025-01-02 14:30:00"), Close: 100},
				{Time: parseTime("2025-01-02 15:30:00"), Close: 101},
				{Time: parseTime("2025-01-05 14:30:00"), Close: 110},
			},
			hourlyIndex:   2,
			lookahead:     true,
			expectedDaily: 1,
			description:   "should handle date gaps and map to correct daily bar",
		},
		{
			name: "missing daily bars - lookahead off at gap boundary",
			dailyBars: []context.OHLCV{
				{Time: parseTime("2025-01-02 14:30:00"), Close: 100},
				{Time: parseTime("2025-01-05 14:30:00"), Close: 110},
			},
			hourlyBars: []context.OHLCV{
				{Time: parseTime("2025-01-02 14:30:00"), Close: 100},
				{Time: parseTime("2025-01-02 15:30:00"), Close: 101},
				{Time: parseTime("2025-01-05 14:30:00"), Close: 110},
			},
			hourlyIndex:   2,
			lookahead:     false,
			expectedDaily: 0,
			description:   "lookahead=off at gap boundary should return previous completed bar",
		},
		{
			name: "sparse hourly data - single bar per day",
			dailyBars: []context.OHLCV{
				{Time: parseTime("2025-01-01 14:30:00"), Close: 100},
				{Time: parseTime("2025-01-02 14:30:00"), Close: 110},
				{Time: parseTime("2025-01-03 14:30:00"), Close: 120},
			},
			hourlyBars: []context.OHLCV{
				{Time: parseTime("2025-01-01 14:30:00"), Close: 100},
				{Time: parseTime("2025-01-02 14:30:00"), Close: 110},
				{Time: parseTime("2025-01-03 14:30:00"), Close: 120},
			},
			hourlyIndex:   1,
			lookahead:     false,
			expectedDaily: 0,
			description:   "should handle sparse hourly data with single bar per day",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mapper := NewSecurityBarMapper()
			mapper.BuildMapping(tt.dailyBars, tt.hourlyBars)
			result := mapper.FindTargetBarIndexByContainment(tt.hourlyIndex, tt.lookahead)

			if result != tt.expectedDaily {
				t.Errorf("%s: expected %d, got %d", tt.description, tt.expectedDaily, result)
			}
		})
	}
}

func TestBarRange_Predicates(t *testing.T) {
	tests := []struct {
		name                string
		rangeStart          int
		rangeEnd            int
		hourlyIndex         int
		expectContains      bool
		expectIsBeforeRange bool
		expectIsAfterRange  bool
		description         string
	}{
		{
			name:                "index before range",
			rangeStart:          10,
			rangeEnd:            20,
			hourlyIndex:         5,
			expectContains:      false,
			expectIsBeforeRange: true,
			expectIsAfterRange:  false,
			description:         "index before range should return false for Contains, true for IsBeforeRange",
		},
		{
			name:                "index at range start boundary",
			rangeStart:          10,
			rangeEnd:            20,
			hourlyIndex:         10,
			expectContains:      true,
			expectIsBeforeRange: false,
			expectIsAfterRange:  false,
			description:         "index at start boundary should be contained",
		},
		{
			name:                "index within range",
			rangeStart:          10,
			rangeEnd:            20,
			hourlyIndex:         15,
			expectContains:      true,
			expectIsBeforeRange: false,
			expectIsAfterRange:  false,
			description:         "index within range should be contained",
		},
		{
			name:                "index at range end boundary",
			rangeStart:          10,
			rangeEnd:            20,
			hourlyIndex:         20,
			expectContains:      true,
			expectIsBeforeRange: false,
			expectIsAfterRange:  false,
			description:         "index at end boundary should be contained",
		},
		{
			name:                "index after range",
			rangeStart:          10,
			rangeEnd:            20,
			hourlyIndex:         25,
			expectContains:      false,
			expectIsBeforeRange: false,
			expectIsAfterRange:  true,
			description:         "index after range should return false for Contains, true for IsAfterRange",
		},
		{
			name:                "index exactly one before start",
			rangeStart:          10,
			rangeEnd:            20,
			hourlyIndex:         9,
			expectContains:      false,
			expectIsBeforeRange: true,
			expectIsAfterRange:  false,
			description:         "index at start-1 should be before range",
		},
		{
			name:                "index exactly one after end",
			rangeStart:          10,
			rangeEnd:            20,
			hourlyIndex:         21,
			expectContains:      false,
			expectIsBeforeRange: false,
			expectIsAfterRange:  true,
			description:         "index at end+1 should be after range",
		},
		{
			name:                "single element range - at index",
			rangeStart:          10,
			rangeEnd:            10,
			hourlyIndex:         10,
			expectContains:      true,
			expectIsBeforeRange: false,
			expectIsAfterRange:  false,
			description:         "single element range should contain exact index",
		},
		{
			name:                "single element range - before",
			rangeStart:          10,
			rangeEnd:            10,
			hourlyIndex:         9,
			expectContains:      false,
			expectIsBeforeRange: true,
			expectIsAfterRange:  false,
			description:         "index before single element range should be detected",
		},
		{
			name:                "single element range - after",
			rangeStart:          10,
			rangeEnd:            10,
			hourlyIndex:         11,
			expectContains:      false,
			expectIsBeforeRange: false,
			expectIsAfterRange:  true,
			description:         "index after single element range should be detected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewBarRange(0, tt.rangeStart, tt.rangeEnd)

			contains := r.Contains(tt.hourlyIndex)
			if contains != tt.expectContains {
				t.Errorf("%s: Contains(%d) = %v, expected %v",
					tt.description, tt.hourlyIndex, contains, tt.expectContains)
			}

			isBefore := r.IsBeforeRange(tt.hourlyIndex)
			if isBefore != tt.expectIsBeforeRange {
				t.Errorf("%s: IsBeforeRange(%d) = %v, expected %v",
					tt.description, tt.hourlyIndex, isBefore, tt.expectIsBeforeRange)
			}

			isAfter := r.IsAfterRange(tt.hourlyIndex)
			if isAfter != tt.expectIsAfterRange {
				t.Errorf("%s: IsAfterRange(%d) = %v, expected %v",
					tt.description, tt.hourlyIndex, isAfter, tt.expectIsAfterRange)
			}
		})
	}
}

func TestSecurityBarMapper_DateBoundaries(t *testing.T) {
	tests := []struct {
		name          string
		dailyBars     []context.OHLCV
		hourlyBars    []context.OHLCV
		hourlyIndex   int
		lookahead     bool
		expectedDaily int
		description   string
	}{
		{
			name: "hourly bar at midnight UTC",
			dailyBars: []context.OHLCV{
				{Time: parseTime("2025-01-01 14:30:00"), Close: 100},
				{Time: parseTime("2025-01-02 14:30:00"), Close: 110},
			},
			hourlyBars: []context.OHLCV{
				{Time: parseTime("2025-01-01 23:00:00"), Close: 100},
				{Time: parseTime("2025-01-02 00:00:00"), Close: 101},
				{Time: parseTime("2025-01-02 01:00:00"), Close: 102},
			},
			hourlyIndex:   1,
			lookahead:     true,
			expectedDaily: 1,
			description:   "bar at exactly midnight should belong to new date",
		},
		{
			name: "hourly bar one second before midnight",
			dailyBars: []context.OHLCV{
				{Time: parseTime("2025-01-01 14:30:00"), Close: 100},
				{Time: parseTime("2025-01-02 14:30:00"), Close: 110},
			},
			hourlyBars: []context.OHLCV{
				{Time: parseTime("2025-01-01 23:59:59"), Close: 100},
				{Time: parseTime("2025-01-02 00:00:00"), Close: 101},
			},
			hourlyIndex:   0,
			lookahead:     true,
			expectedDaily: 0,
			description:   "bar before midnight should belong to previous date",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mapper := NewSecurityBarMapper()
			mapper.BuildMapping(tt.dailyBars, tt.hourlyBars)
			result := mapper.FindTargetBarIndexByContainment(tt.hourlyIndex, tt.lookahead)

			if result != tt.expectedDaily {
				t.Errorf("%s: expected %d, got %d", tt.description, tt.expectedDaily, result)
			}
		})
	}
}

func TestBarRange_Contains(t *testing.T) {
	t.Skip("replaced by TestBarRange_Predicates for comprehensive predicate testing")
	r := NewBarRange(0, 10, 20)

	tests := []struct {
		hourlyIndex int
		expected    bool
	}{
		{5, false},
		{10, true},
		{15, true},
		{20, true},
		{25, false},
	}

	for _, tt := range tests {
		result := r.Contains(tt.hourlyIndex)
		if result != tt.expected {
			t.Errorf("Contains(%d) = %v, expected %v", tt.hourlyIndex, result, tt.expected)
		}
	}
}

func parseTime(layout string) int64 {
	t, _ := timeFromString(layout)
	return t
}

func timeFromString(s string) (int64, error) {
	layout := "2006-01-02 15:04:05"
	t, err := parseUTC(s, layout)
	if err != nil {
		return 0, err
	}
	return t.Unix(), nil
}

func parseUTC(value, layout string) (t time.Time, err error) {
	return time.Parse(layout, value)
}
