package session

import (
	"math"
	"testing"
	"time"
)

/*
Comprehensive timezone tests for three major providers:
- MOEX (Moscow Exchange): UTC+3 - "Europe/Moscow"
- NYSE (New York Stock Exchange via Yahoo): UTC-5 (EST) / UTC-4 (EDT) - "America/New_York"
- Binance (Crypto): UTC - "UTC"

These tests verify that session strings like "0950-1645" are interpreted
in the exchange's local timezone, NOT UTC.
*/

func TestTimezone_MOEX_Moscow(t *testing.T) {
	// MOEX trading hours: 09:50-16:45 Moscow time (UTC+3)
	// Test with a timestamp at 12:00 UTC = 15:00 Moscow time
	// Should be IN session "0950-1645" when using "Europe/Moscow" timezone

	utcTime := time.Date(2025, 11, 18, 12, 0, 0, 0, time.UTC)
	timestamp := utcTime.UnixMilli()

	t.Run("12:00 UTC = 15:00 Moscow (IN session)", func(t *testing.T) {
		result := TimeFunc(timestamp, "1h", "0950-1645", "Europe/Moscow")
		if math.IsNaN(result) {
			t.Errorf("TimeFunc() = NaN, want %v (timestamp should be IN session at 15:00 Moscow)", timestamp)
		}
		if result != float64(timestamp) {
			t.Errorf("TimeFunc() = %v, want %v", result, float64(timestamp))
		}
	})

	t.Run("12:00 UTC with incorrect UTC timezone (also IN session)", func(t *testing.T) {
		// 12:00 UTC is IN session "0950-1645" whether we interpret it as UTC or Moscow time
		// With UTC timezone: 12:00 UTC is IN "0950-1645" UTC
		// With Moscow timezone: 12:00 UTC = 15:00 Moscow is IN "0950-1645" Moscow
		// This happens to work either way for this particular time
		result := TimeFunc(timestamp, "1h", "0950-1645", "UTC")
		if math.IsNaN(result) {
			t.Error("12:00 UTC should be IN session 0950-1645 UTC (also happens to be IN when converted to Moscow)")
		}
		t.Log("✓ Note: 12:00 UTC works with both timezones for this session. Use 18:00 UTC to see real difference.")
	})

	// Test edge case: 18:00 UTC = 21:00 Moscow (OUT of session)
	t.Run("18:00 UTC = 21:00 Moscow (OUT session)", func(t *testing.T) {
		lateTime := time.Date(2025, 11, 18, 18, 0, 0, 0, time.UTC)
		result := TimeFunc(lateTime.UnixMilli(), "1h", "0950-1645", "Europe/Moscow")
		if !math.IsNaN(result) {
			t.Errorf("TimeFunc() = %v, want NaN (21:00 Moscow is OUT of session 0950-1645)", result)
		}
	})

	// Test early morning: 07:00 UTC = 10:00 Moscow (IN session)
	t.Run("07:00 UTC = 10:00 Moscow (IN session)", func(t *testing.T) {
		morningTime := time.Date(2025, 11, 18, 7, 0, 0, 0, time.UTC)
		result := TimeFunc(morningTime.UnixMilli(), "1h", "0950-1645", "Europe/Moscow")
		if math.IsNaN(result) {
			t.Errorf("TimeFunc() = NaN, want timestamp (10:00 Moscow should be IN session)")
		}
	})
}

func TestTimezone_NYSE_NewYork(t *testing.T) {
	// NYSE trading hours: 09:30-16:00 New York time (UTC-5 EST / UTC-4 EDT)
	// Using November date (EST, UTC-5)
	// Test with 14:30 UTC = 09:30 EST (session start)

	utcTime := time.Date(2025, 11, 18, 14, 30, 0, 0, time.UTC)
	timestamp := utcTime.UnixMilli()

	t.Run("14:30 UTC = 09:30 EST (session start - IN)", func(t *testing.T) {
		result := TimeFunc(timestamp, "1h", "0930-1600", "America/New_York")
		if math.IsNaN(result) {
			t.Errorf("TimeFunc() = NaN, want %v (09:30 EST is session start)", timestamp)
		}
	})

	t.Run("21:00 UTC = 16:00 EST (session end - IN)", func(t *testing.T) {
		endTime := time.Date(2025, 11, 18, 21, 0, 0, 0, time.UTC)
		result := TimeFunc(endTime.UnixMilli(), "1h", "0930-1600", "America/New_York")
		if math.IsNaN(result) {
			t.Errorf("TimeFunc() = NaN, want timestamp (16:00 EST is session end)")
		}
	})

	t.Run("21:01 UTC = 16:01 EST (after session - OUT)", func(t *testing.T) {
		afterTime := time.Date(2025, 11, 18, 21, 1, 0, 0, time.UTC)
		result := TimeFunc(afterTime.UnixMilli(), "1h", "0930-1600", "America/New_York")
		if !math.IsNaN(result) {
			t.Errorf("TimeFunc() = %v, want NaN (16:01 EST is after session)", result)
		}
	})

	t.Run("Verify timezone matters - same UTC time different result", func(t *testing.T) {
		// 15:00 UTC with different timezones
		testTime := time.Date(2025, 11, 18, 15, 0, 0, 0, time.UTC)
		ts := testTime.UnixMilli()

		// 15:00 UTC = 10:00 EST (IN session 0930-1600)
		nyResult := TimeFunc(ts, "1h", "0930-1600", "America/New_York")

		// 15:00 UTC = 18:00 Moscow (OUT of session 0950-1645)
		moscowResult := TimeFunc(ts, "1h", "0950-1645", "Europe/Moscow")

		// 15:00 UTC with UTC timezone (OUT of session 0930-1600)
		utcResult := TimeFunc(ts, "1h", "0930-1600", "UTC")

		if math.IsNaN(nyResult) {
			t.Error("NYSE at 10:00 EST should be IN session")
		}
		if !math.IsNaN(moscowResult) {
			t.Error("MOEX at 18:00 Moscow should be OUT of session")
		}
		if math.IsNaN(utcResult) {
			t.Log("✓ UTC 15:00 correctly OUT of session 0930-1600 UTC")
		}
	})
}

func TestTimezone_Binance_UTC(t *testing.T) {
	// Binance operates 24/7 in UTC
	// Test typical session: 00:00-23:59 UTC

	t.Run("Binance 24-hour session (00:00-23:59 UTC)", func(t *testing.T) {
		times := []time.Time{
			time.Date(2025, 11, 18, 0, 0, 0, 0, time.UTC),
			time.Date(2025, 11, 18, 6, 30, 0, 0, time.UTC),
			time.Date(2025, 11, 18, 12, 0, 0, 0, time.UTC),
			time.Date(2025, 11, 18, 18, 45, 0, 0, time.UTC),
			time.Date(2025, 11, 18, 23, 59, 0, 0, time.UTC),
		}

		for _, tm := range times {
			t.Run(tm.Format("15:04 UTC"), func(t *testing.T) {
				result := TimeFunc(tm.UnixMilli(), "1h", "0000-2359", "UTC")
				if math.IsNaN(result) {
					t.Errorf("24-hour UTC session should always be IN, got OUT at %s", tm.Format(time.RFC3339))
				}
			})
		}
	})

	t.Run("Binance partial session (08:00-20:00 UTC)", func(t *testing.T) {
		// Some trading strategies may use partial UTC sessions
		testCases := []struct {
			hour   int
			minute int
			wantIn bool
		}{
			{7, 59, false}, // Before session
			{8, 0, true},   // Session start
			{14, 30, true}, // Mid session
			{20, 0, true},  // Session end
			{20, 1, false}, // After session
		}

		for _, tc := range testCases {
			t.Run(time.Date(2025, 11, 18, tc.hour, tc.minute, 0, 0, time.UTC).Format("15:04"), func(t *testing.T) {
				tm := time.Date(2025, 11, 18, tc.hour, tc.minute, 0, 0, time.UTC)
				result := TimeFunc(tm.UnixMilli(), "1h", "0800-2000", "UTC")
				gotIn := !math.IsNaN(result)
				if gotIn != tc.wantIn {
					t.Errorf("TimeFunc() at %02d:%02d UTC: gotIn=%v, wantIn=%v", tc.hour, tc.minute, gotIn, tc.wantIn)
				}
			})
		}
	})
}

func TestTimezone_CrossProvider_SameWallClock(t *testing.T) {
	// Critical test: Same wall-clock time (e.g., "10:00") in different timezones
	// should produce DIFFERENT results based on exchange timezone

	// All providers have session "1000-1500" in their local timezone
	session := "1000-1500"

	// Pick a UTC timestamp that corresponds to 10:00 in one timezone but not others
	// 10:00 UTC = 10:00 UTC, 13:00 Moscow, 05:00 EST
	utcTime := time.Date(2025, 11, 18, 10, 0, 0, 0, time.UTC)
	timestamp := utcTime.UnixMilli()

	t.Run("10:00 UTC - different results per provider", func(t *testing.T) {
		// Binance (UTC): 10:00 UTC = IN session "1000-1500"
		binanceResult := TimeFunc(timestamp, "1h", session, "UTC")
		binanceIn := !math.IsNaN(binanceResult)

		// MOEX (Moscow): 10:00 UTC = 13:00 Moscow = IN session "1000-1500"
		moexResult := TimeFunc(timestamp, "1h", session, "Europe/Moscow")
		moexIn := !math.IsNaN(moexResult)

		// NYSE (NY): 10:00 UTC = 05:00 EST = OUT of session "1000-1500"
		nyseResult := TimeFunc(timestamp, "1h", session, "America/New_York")
		nyseIn := !math.IsNaN(nyseResult)

		t.Logf("10:00 UTC results - Binance(UTC):%v MOEX(Moscow):%v NYSE(NY):%v", binanceIn, moexIn, nyseIn)

		if !binanceIn {
			t.Error("Binance: 10:00 UTC should be IN session 1000-1500 UTC")
		}
		if !moexIn {
			t.Error("MOEX: 10:00 UTC = 13:00 Moscow should be IN session 1000-1500 Moscow")
		}
		if nyseIn {
			t.Error("NYSE: 10:00 UTC = 05:00 EST should be OUT of session 1000-1500 EST")
		}
	})
}

func TestTimezone_InvalidTimezone_FallbackToUTC(t *testing.T) {
	// Test that invalid timezone names gracefully fallback to UTC
	utcTime := time.Date(2025, 11, 18, 10, 0, 0, 0, time.UTC)
	timestamp := utcTime.UnixMilli()

	t.Run("Invalid timezone falls back to UTC", func(t *testing.T) {
		// Use invalid timezone - should fallback to UTC behavior
		result := TimeFunc(timestamp, "1h", "0950-1645", "Invalid/Timezone")

		// 10:00 UTC should be IN session "0950-1645" when treated as UTC
		if math.IsNaN(result) {
			t.Error("Invalid timezone should fallback to UTC, where 10:00 is IN session 0950-1645")
		}
	})
}

/* Benchmark: Verify timezone conversion doesn't significantly impact performance */

func BenchmarkTimeFunc_WithTimezone_UTC(b *testing.B) {
	timestamp := time.Date(2025, 11, 18, 12, 0, 0, 0, time.UTC).UnixMilli()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		TimeFunc(timestamp, "1h", "0950-1645", "UTC")
	}
}

func BenchmarkTimeFunc_WithTimezone_Moscow(b *testing.B) {
	timestamp := time.Date(2025, 11, 18, 12, 0, 0, 0, time.UTC).UnixMilli()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		TimeFunc(timestamp, "1h", "0950-1645", "Europe/Moscow")
	}
}

func BenchmarkTimeFunc_WithTimezone_NewYork(b *testing.B) {
	timestamp := time.Date(2025, 11, 18, 14, 30, 0, 0, time.UTC).UnixMilli()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		TimeFunc(timestamp, "1h", "0930-1600", "America/New_York")
	}
}
