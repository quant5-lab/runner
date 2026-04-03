package calendar

import (
	"math"
	"testing"
)

func TestTimestamp_BasicBehavior(t *testing.T) {
	tests := []struct {
		name               string
		y, m, d, h, min, s float64
		tz                 string
		want               float64
	}{
		{"date_only", 2024, 1, 15, 0, 0, 0, "UTC", 1705276800},
		{"with_time", 2024, 1, 15, 14, 30, 0, "UTC", 1705329000},
		{"with_seconds", 2024, 1, 15, 14, 30, 45, "UTC", 1705329045},
		{"epoch", 1970, 1, 1, 0, 0, 0, "UTC", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Timestamp(tt.y, tt.m, tt.d, tt.h, tt.min, tt.s, tt.tz); got != tt.want {
				t.Errorf("Timestamp = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTimestamp_TimezoneConversion(t *testing.T) {
	tests := []struct {
		tz   string
		h    float64
		want float64
	}{
		{"UTC", 12, 1705320000},
		{"Europe/Moscow", 15, 1705320000},
		{"America/New_York", 7, 1705320000},
		{"", 12, 1705320000},
	}

	for _, tt := range tests {
		t.Run(tt.tz, func(t *testing.T) {
			if got := Timestamp(2024, 1, 15, tt.h, 0, 0, tt.tz); got != tt.want {
				t.Errorf("Timestamp(%s) = %v, want %v", tt.tz, got, tt.want)
			}
		})
	}
}

func TestTimestamp_NaNHandling(t *testing.T) {
	nan := math.NaN()

	tests := []struct {
		name               string
		y, m, d, h, min, s float64
	}{
		{"NaN_year", nan, 1, 15, 0, 0, 0},
		{"NaN_month", 2024, nan, 15, 0, 0, 0},
		{"NaN_day", 2024, 1, nan, 0, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Timestamp(tt.y, tt.m, tt.d, tt.h, tt.min, tt.s, "UTC"); !math.IsNaN(got) {
				t.Errorf("Timestamp = %v, want NaN", got)
			}
		})
	}
}

func TestTimestamp_NaNTimeComponentsDefaultToZero(t *testing.T) {
	nan := math.NaN()
	want := Timestamp(2024, 1, 15, 0, 0, 0, "UTC")

	tests := []struct {
		name               string
		y, m, d, h, min, s float64
	}{
		{"NaN_hour", 2024, 1, 15, nan, 0, 0},
		{"NaN_minute", 2024, 1, 15, 0, nan, 0},
		{"NaN_second", 2024, 1, 15, 0, 0, nan},
		{"NaN_all_time", 2024, 1, 15, nan, nan, nan},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Timestamp(tt.y, tt.m, tt.d, tt.h, tt.min, tt.s, "UTC"); got != want {
				t.Errorf("Timestamp = %v, want %v", got, want)
			}
		})
	}
}

func TestTimestamp_BoundaryValues(t *testing.T) {
	tests := []struct {
		name               string
		y, m, d, h, min, s float64
		notNaN             bool
	}{
		{"year_1", 1, 1, 1, 0, 0, 0, true},
		{"year_9999", 9999, 12, 31, 23, 59, 59, true},
		{"month_1", 2024, 1, 1, 0, 0, 0, true},
		{"month_12", 2024, 12, 31, 0, 0, 0, true},
		{"day_1", 2024, 1, 1, 0, 0, 0, true},
		{"day_31", 2024, 1, 31, 0, 0, 0, true},
		{"hour_0", 2024, 1, 1, 0, 0, 0, true},
		{"hour_23", 2024, 1, 1, 23, 0, 0, true},
		{"minute_0", 2024, 1, 1, 0, 0, 0, true},
		{"minute_59", 2024, 1, 1, 0, 59, 0, true},
		{"second_0", 2024, 1, 1, 0, 0, 0, true},
		{"second_59", 2024, 1, 1, 0, 0, 59, true},
		{"leap_year", 2020, 2, 29, 0, 0, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Timestamp(tt.y, tt.m, tt.d, tt.h, tt.min, tt.s, "UTC")
			if tt.notNaN && math.IsNaN(got) {
				t.Errorf("Timestamp = NaN, want valid timestamp")
			}
		})
	}
}

func TestTimestampFromString_AllFormats(t *testing.T) {
	tests := []struct {
		format string
		input  string
		want   float64
	}{
		{"ISO8601_date", "2024-01-15", 1705276800},
		{"ISO8601_datetime_T", "2024-01-15T14:30:00", 1705329000},
		{"ISO8601_datetime_space", "2024-01-15 14:30:00", 1705329000},
		{"RFC3339_Z", "2024-01-15T14:30:00Z", 1705329000},
		{"RFC3339_offset", "2024-01-15T14:30:00+00:00", 1705329000},
		{"RFC1123Z", "Mon, 15 Jan 2024 14:30:00 +0000", 1705329000},
		{"RFC2822_no_day", "15 Jan 2024 14:30:00 +0000", 1705329000},
		{"RFC2822_date_only", "20 Feb 2020", 1582156800},
		{"Pine_doc_example_1", "20 Feb 2020", 1582156800},
		{"Pine_doc_example_2", "2011-10-10T14:48:00", 1318258080},
	}

	for _, tt := range tests {
		t.Run(tt.format, func(t *testing.T) {
			if got := TimestampFromString(tt.input, "UTC"); got != tt.want {
				t.Errorf("TimestampFromString(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestTimestampFromString_TimezoneHandling(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		tzParam string
		want    float64
	}{
		{"RFC3339_Z_ignores_param", "2024-01-15T12:00:00Z", "America/New_York", 1705320000},
		{"RFC3339_offset_overrides", "2024-01-15T12:00:00+05:00", "UTC", 1705302000},
		{"ISO8601_uses_param", "2024-01-15 12:00:00", "America/New_York", 1705338000},
		{"date_only_uses_param", "2024-01-15", "Europe/Moscow", 1705266000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TimestampFromString(tt.input, tt.tzParam); got != tt.want {
				t.Errorf("TimestampFromString(%q, %q) = %v, want %v", tt.input, tt.tzParam, got, tt.want)
			}
		})
	}
}

func TestTimestampFromString_EdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"empty", ""},
		{"invalid", "not-a-date"},
		{"partial", "2024-01"},
		{"wrong_format", "01/15/2024"},
		{"comma_separated", "Jan 15, 2024"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TimestampFromString(tt.input, "UTC"); !math.IsNaN(got) {
				t.Errorf("TimestampFromString(%q) = %v, want NaN", tt.input, got)
			}
		})
	}
}

func TestTimestampFromString_BoundaryDates(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  float64
	}{
		{"epoch", "1970-01-01", 0},
		{"leap_day", "2020-02-29", 1582934400},
		{"year_2038", "2038-01-19 03:14:07", 2147483647},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TimestampFromString(tt.input, "UTC"); got != tt.want {
				t.Errorf("TimestampFromString(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
