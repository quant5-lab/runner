package request

import (
	"testing"

	"github.com/quant5-lab/runner/runtime/context"
)

/*
   Daily bar at date D maps ALL hourly bars with dates in [D, D_next).
   Naturally handles weekend gaps, holiday gaps, and consecutive days.
*/

func TestSecurityBarMapper_TimeRangeExtension_WeekendGap(t *testing.T) {
	mapper := NewSecurityBarMapper()

	/* Friday + Monday daily bars (weekend gap) */
	dailyBars := []context.OHLCV{
		{Time: parseTime("2025-03-07 20:00:00"), Close: 100}, // Friday 20:00 UTC
		{Time: parseTime("2025-03-10 20:00:00"), Close: 101}, // Monday 20:00 UTC
	}

	/* Hourly bars spanning Friday through Monday */
	hourlyBars := []context.OHLCV{
		{Time: parseTime("2025-03-07 09:00:00"), Close: 100}, // Fri 9am
		{Time: parseTime("2025-03-07 15:00:00"), Close: 100}, // Fri 3pm
		{Time: parseTime("2025-03-08 10:00:00"), Close: 100}, // Sat 10am
		{Time: parseTime("2025-03-09 10:00:00"), Close: 100}, // Sun 10am
		{Time: parseTime("2025-03-10 09:00:00"), Close: 101}, // Mon 9am
	}

	mapper.BuildMappingWithDateFilter(dailyBars, hourlyBars, DateRange{}, "UTC")

	/* Verify range structure: Friday extends through weekend */
	if len(mapper.ranges) != 2 {
		t.Fatalf("Expected 2 ranges (Fri + Mon), got %d", len(mapper.ranges))
	}

	/* Range[0]: Friday daily bar should own hourly indices [0:3] (Fri/Sat/Sun) */
	if mapper.ranges[0].DailyBarIndex != 0 {
		t.Errorf("Range[0] should map to daily[0], got %d", mapper.ranges[0].DailyBarIndex)
	}
	if mapper.ranges[0].StartHourlyIndex != 0 {
		t.Errorf("Range[0] should start at hourly[0], got %d", mapper.ranges[0].StartHourlyIndex)
	}
	if mapper.ranges[0].EndHourlyIndex != 3 {
		t.Errorf("Range[0] should end at hourly[3], got %d (Friday should extend through weekend)", mapper.ranges[0].EndHourlyIndex)
	}

	/* Range[1]: Monday daily bar should own hourly index [4:4] */
	if mapper.ranges[1].DailyBarIndex != 1 {
		t.Errorf("Range[1] should map to daily[1], got %d", mapper.ranges[1].DailyBarIndex)
	}
	if mapper.ranges[1].StartHourlyIndex != 4 {
		t.Errorf("Range[1] should start at hourly[4], got %d", mapper.ranges[1].StartHourlyIndex)
	}
}

func TestSecurityBarMapper_TimeRangeExtension_MultiDayGap(t *testing.T) {
	mapper := NewSecurityBarMapper()

	/* Thursday + Tuesday daily bars (3-day gap: Fri/Sat/Sun) */
	dailyBars := []context.OHLCV{
		{Time: parseTime("2025-03-06 20:00:00"), Close: 100}, // Thursday
		{Time: parseTime("2025-03-11 20:00:00"), Close: 101}, // Tuesday (3-day gap)
	}

	hourlyBars := []context.OHLCV{
		{Time: parseTime("2025-03-06 09:00:00"), Close: 100}, // Thu
		{Time: parseTime("2025-03-07 09:00:00"), Close: 100}, // Fri
		{Time: parseTime("2025-03-08 09:00:00"), Close: 100}, // Sat
		{Time: parseTime("2025-03-09 09:00:00"), Close: 100}, // Sun
		{Time: parseTime("2025-03-10 09:00:00"), Close: 100}, // Mon
		{Time: parseTime("2025-03-11 09:00:00"), Close: 101}, // Tue
	}

	mapper.BuildMappingWithDateFilter(dailyBars, hourlyBars, DateRange{}, "UTC")

	if len(mapper.ranges) != 2 {
		t.Fatalf("Expected 2 ranges, got %d", len(mapper.ranges))
	}

	/* Thursday daily bar extends through entire gap to cover Thu-Mon hourly bars */
	if mapper.ranges[0].EndHourlyIndex != 4 {
		t.Errorf("Range[0] (Thursday) should extend to hourly[4] (Monday), got %d", mapper.ranges[0].EndHourlyIndex)
	}

	/* Tuesday daily bar starts at Tuesday hourly bar */
	if mapper.ranges[1].StartHourlyIndex != 5 {
		t.Errorf("Range[1] (Tuesday) should start at hourly[5], got %d", mapper.ranges[1].StartHourlyIndex)
	}
}

func TestSecurityBarMapper_TimeRangeExtension_LastDailyBarConsumesRemaining(t *testing.T) {
	mapper := NewSecurityBarMapper()

	/* Single daily bar with trailing hourly bars */
	dailyBars := []context.OHLCV{
		{Time: parseTime("2025-03-10 20:00:00"), Close: 100},
	}

	hourlyBars := []context.OHLCV{
		{Time: parseTime("2025-03-10 09:00:00"), Close: 100},
		{Time: parseTime("2025-03-11 09:00:00"), Close: 100},
		{Time: parseTime("2025-03-12 09:00:00"), Close: 100},
	}

	mapper.BuildMappingWithDateFilter(dailyBars, hourlyBars, DateRange{}, "UTC")

	if len(mapper.ranges) != 1 {
		t.Fatalf("Expected 1 range, got %d", len(mapper.ranges))
	}

	/* Last daily bar should consume ALL remaining hourly bars */
	if mapper.ranges[0].EndHourlyIndex != 2 {
		t.Errorf("Last daily bar should consume all hourly bars through index[2], got %d", mapper.ranges[0].EndHourlyIndex)
	}
}

func TestSecurityBarMapper_TimeRangeExtension_MOEXWeekendTradingPattern(t *testing.T) {
	mapper := NewSecurityBarMapper()

	/* MOEX weekend trading: Fri/Sat/Sun/Mon daily bars with hourly trading */
	dailyBars := []context.OHLCV{
		{Time: parseTime("2025-02-14 20:00:00"), Close: 100}, // Fri daily
		{Time: parseTime("2025-02-15 20:00:00"), Close: 101}, // Sat daily
		{Time: parseTime("2025-02-16 20:00:00"), Close: 102}, // Sun daily
		{Time: parseTime("2025-02-17 20:00:00"), Close: 103}, // Mon daily
	}

	hourlyBars := []context.OHLCV{
		{Time: parseTime("2025-02-14 09:00:00"), Close: 100}, // Fri 9am
		{Time: parseTime("2025-02-14 15:00:00"), Close: 100}, // Fri 3pm
		{Time: parseTime("2025-02-15 10:00:00"), Close: 101}, // Sat 10am
		{Time: parseTime("2025-02-15 14:00:00"), Close: 101}, // Sat 2pm
		{Time: parseTime("2025-02-16 10:00:00"), Close: 102}, // Sun 10am
		{Time: parseTime("2025-02-16 14:00:00"), Close: 102}, // Sun 2pm
		{Time: parseTime("2025-02-17 09:00:00"), Close: 103}, // Mon 9am
	}

	mapper.BuildMappingWithDateFilter(dailyBars, hourlyBars, DateRange{}, "UTC")

	if len(mapper.ranges) != 4 {
		t.Fatalf("Expected 4 ranges (Fri/Sat/Sun/Mon), got %d", len(mapper.ranges))
	}

	/* Each daily bar should own its corresponding hourly bars by date */
	/* Friday: hourly[0:1] */
	if mapper.ranges[0].StartHourlyIndex != 0 || mapper.ranges[0].EndHourlyIndex != 1 {
		t.Errorf("Friday range should be [0:1], got [%d:%d]", mapper.ranges[0].StartHourlyIndex, mapper.ranges[0].EndHourlyIndex)
	}

	/* Saturday: hourly[2:3] */
	if mapper.ranges[1].StartHourlyIndex != 2 || mapper.ranges[1].EndHourlyIndex != 3 {
		t.Errorf("Saturday range should be [2:3], got [%d:%d]", mapper.ranges[1].StartHourlyIndex, mapper.ranges[1].EndHourlyIndex)
	}

	/* Sunday: hourly[4:5] */
	if mapper.ranges[2].StartHourlyIndex != 4 || mapper.ranges[2].EndHourlyIndex != 5 {
		t.Errorf("Sunday range should be [4:5], got [%d:%d]", mapper.ranges[2].StartHourlyIndex, mapper.ranges[2].EndHourlyIndex)
	}

	/* Monday: hourly[6:6] */
	if mapper.ranges[3].StartHourlyIndex != 6 || mapper.ranges[3].EndHourlyIndex != 6 {
		t.Errorf("Monday range should be [6:6], got [%d:%d]", mapper.ranges[3].StartHourlyIndex, mapper.ranges[3].EndHourlyIndex)
	}
}
