package calendar

import (
	"math"
	"testing"
)

// All Pine `time` values are Unix milliseconds. Tests use *1000 to convert
// second-anchored UTC epochs to the ms convention the public API expects.

func TestCalendarExtraction_BasicBehavior(t *testing.T) {
	ts := float64(1721046600 * 1000)

	tests := []struct {
		name string
		fn   func(float64, string) float64
		want float64
	}{
		{"Year", Year, 2024},
		{"Month", Month, 7},
		{"DayOfWeek", DayOfWeek, 2},
		{"DayOfMonth", DayOfMonth, 15},
		{"Hour", Hour, 12},
		{"Minute", Minute, 30},
		{"Second", Second, 0},
		{"WeekOfYear", WeekOfYear, 29},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.fn(ts, "UTC"); got != tt.want {
				t.Errorf("%s(%v) = %v, want %v", tt.name, ts, got, tt.want)
			}
		})
	}
}

func TestCalendarExtraction_NaNHandling(t *testing.T) {
	nan := math.NaN()
	funcs := []struct {
		name string
		fn   func(float64, string) float64
	}{
		{"Year", Year},
		{"Month", Month},
		{"DayOfWeek", DayOfWeek},
		{"DayOfMonth", DayOfMonth},
		{"Hour", Hour},
		{"Minute", Minute},
		{"Second", Second},
		{"WeekOfYear", WeekOfYear},
	}

	for _, f := range funcs {
		t.Run(f.name, func(t *testing.T) {
			if got := f.fn(nan, "UTC"); !math.IsNaN(got) {
				t.Errorf("%s(NaN) = %v, want NaN", f.name, got)
			}
		})
	}
}

func TestCalendarExtraction_TimezoneEffects(t *testing.T) {
	ts := float64(1721086200 * 1000)

	tests := []struct {
		tz        string
		day       float64
		hour      float64
		dayOfWeek float64
	}{
		{"UTC", 15, 23, 2},
		{"Asia/Tokyo", 16, 8, 3},
		{"America/New_York", 15, 19, 2},
		{"", 15, 23, 2},
	}

	for _, tt := range tests {
		t.Run(tt.tz, func(t *testing.T) {
			if got := DayOfMonth(ts, tt.tz); got != tt.day {
				t.Errorf("DayOfMonth(%s) = %v, want %v", tt.tz, got, tt.day)
			}
			if got := Hour(ts, tt.tz); got != tt.hour {
				t.Errorf("Hour(%s) = %v, want %v", tt.tz, got, tt.hour)
			}
			if got := DayOfWeek(ts, tt.tz); got != tt.dayOfWeek {
				t.Errorf("DayOfWeek(%s) = %v, want %v", tt.tz, got, tt.dayOfWeek)
			}
		})
	}
}

func TestDayOfWeek_PineConvention(t *testing.T) {
	tests := []struct {
		date string
		ts   float64
		want float64
	}{
		{"2024-07-14 (Sun)", 1720915200 * 1000, 1},
		{"2024-07-15 (Mon)", 1721001600 * 1000, 2},
		{"2024-07-16 (Tue)", 1721088000 * 1000, 3},
		{"2024-07-17 (Wed)", 1721174400 * 1000, 4},
		{"2024-07-18 (Thu)", 1721260800 * 1000, 5},
		{"2024-07-19 (Fri)", 1721347200 * 1000, 6},
		{"2024-07-20 (Sat)", 1721433600 * 1000, 7},
	}

	for _, tt := range tests {
		t.Run(tt.date, func(t *testing.T) {
			if got := DayOfWeek(tt.ts, "UTC"); got != tt.want {
				t.Errorf("DayOfWeek = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMonth_AllMonths(t *testing.T) {
	tests := []struct {
		month string
		ts    float64
		want  float64
	}{
		{"Jan", 1704067200 * 1000, 1},
		{"Feb", 1706745600 * 1000, 2},
		{"Mar", 1709251200 * 1000, 3},
		{"Apr", 1711929600 * 1000, 4},
		{"May", 1714521600 * 1000, 5},
		{"Jun", 1717200000 * 1000, 6},
		{"Jul", 1719792000 * 1000, 7},
		{"Aug", 1722470400 * 1000, 8},
		{"Sep", 1725148800 * 1000, 9},
		{"Oct", 1727740800 * 1000, 10},
		{"Nov", 1730419200 * 1000, 11},
		{"Dec", 1733011200 * 1000, 12},
	}

	for _, tt := range tests {
		t.Run(tt.month, func(t *testing.T) {
			if got := Month(tt.ts, "UTC"); got != tt.want {
				t.Errorf("Month = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCalendarExtraction_BoundaryValues(t *testing.T) {
	tests := []struct {
		name string
		ts   float64
		year float64
		hour float64
		min  float64
		sec  float64
	}{
		{"epoch", 0, 1970, 0, 0, 0},
		{"negative", -86400 * 1000, 1969, 0, 0, 0},
		{"year_2038", 2147483647 * 1000, 2038, 3, 14, 7},
		{"midnight", 1721001600 * 1000, 2024, 0, 0, 0},
		{"end_of_day", 1721087999 * 1000, 2024, 23, 59, 59},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Year(tt.ts, "UTC"); got != tt.year {
				t.Errorf("Year = %v, want %v", got, tt.year)
			}
			if got := Hour(tt.ts, "UTC"); got != tt.hour {
				t.Errorf("Hour = %v, want %v", got, tt.hour)
			}
			if got := Minute(tt.ts, "UTC"); got != tt.min {
				t.Errorf("Minute = %v, want %v", got, tt.min)
			}
			if got := Second(tt.ts, "UTC"); got != tt.sec {
				t.Errorf("Second = %v, want %v", got, tt.sec)
			}
		})
	}
}

func TestWeekOfYear_EdgeCases(t *testing.T) {
	tests := []struct {
		date string
		ts   float64
		want float64
	}{
		{"week_1", 1704067200 * 1000, 1},
		{"year_end_week_1", 1735603200 * 1000, 1},
		{"mid_year", 1721001600 * 1000, 29},
	}

	for _, tt := range tests {
		t.Run(tt.date, func(t *testing.T) {
			if got := WeekOfYear(tt.ts, "UTC"); got != tt.want {
				t.Errorf("WeekOfYear = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestMsRoundtripWithTimestamp verifies the calendar.Timestamp → Year/Month/DayOfMonth/Hour/Minute round-trip.
func TestMsRoundtripWithTimestamp(t *testing.T) {
	ts := Timestamp(2024, 7, 15, 12, 30, 0, "UTC")
	if got := Year(ts, "UTC"); got != 2024 {
		t.Errorf("Year(Timestamp(2024,...)) = %v, want 2024", got)
	}
	if got := Month(ts, "UTC"); got != 7 {
		t.Errorf("Month(Timestamp(...,7,...)) = %v, want 7", got)
	}
	if got := DayOfMonth(ts, "UTC"); got != 15 {
		t.Errorf("DayOfMonth(Timestamp(...,15,...)) = %v, want 15", got)
	}
	if got := Hour(ts, "UTC"); got != 12 {
		t.Errorf("Hour(Timestamp(...,12,...)) = %v, want 12", got)
	}
	if got := Minute(ts, "UTC"); got != 30 {
		t.Errorf("Minute(Timestamp(...,30)) = %v, want 30", got)
	}
}
