package context

import (
	"testing"
	"time"
)

func TestDecomposeBarTime(t *testing.T) {
	utc := time.UTC
	nyc, _ := time.LoadLocation("America/New_York")
	tokyo, _ := time.LoadLocation("Asia/Tokyo")

	tests := []struct {
		name     string
		ts       int64
		loc      *time.Location
		expected BarCalendar
	}{
		{
			name: "UTC mid-day",
			ts:   time.Date(2024, 3, 15, 14, 30, 45, 0, utc).Unix(),
			loc:  utc,
			expected: BarCalendar{
				Year: 2024, Month: 3, DayOfMonth: 15,
				Hour: 14, Minute: 30, Second: 45,
				DayOfWeek: 6, WeekOfYear: 11,
			},
		},
		{
			name: "midnight boundary",
			ts:   time.Date(2024, 6, 1, 0, 0, 0, 0, utc).Unix(),
			loc:  utc,
			expected: BarCalendar{
				Year: 2024, Month: 6, DayOfMonth: 1,
				Hour: 0, Minute: 0, Second: 0,
				DayOfWeek: 7, WeekOfYear: 22,
			},
		},
		{
			name: "end of day 23:59:59",
			ts:   time.Date(2024, 12, 31, 23, 59, 59, 0, utc).Unix(),
			loc:  utc,
			expected: BarCalendar{
				Year: 2024, Month: 12, DayOfMonth: 31,
				Hour: 23, Minute: 59, Second: 59,
				DayOfWeek: 3, WeekOfYear: 1,
			},
		},
		{
			name: "new year boundary Jan 1",
			ts:   time.Date(2025, 1, 1, 0, 0, 0, 0, utc).Unix(),
			loc:  utc,
			expected: BarCalendar{
				Year: 2025, Month: 1, DayOfMonth: 1,
				Hour: 0, Minute: 0, Second: 0,
				DayOfWeek: 4, WeekOfYear: 1,
			},
		},
		{
			name: "leap year Feb 29",
			ts:   time.Date(2024, 2, 29, 12, 0, 0, 0, utc).Unix(),
			loc:  utc,
			expected: BarCalendar{
				Year: 2024, Month: 2, DayOfMonth: 29,
				Hour: 12, Minute: 0, Second: 0,
				DayOfWeek: 5, WeekOfYear: 9,
			},
		},
		{
			name: "Unix epoch",
			ts:   0,
			loc:  utc,
			expected: BarCalendar{
				Year: 1970, Month: 1, DayOfMonth: 1,
				Hour: 0, Minute: 0, Second: 0,
				DayOfWeek: 5, WeekOfYear: 1,
			},
		},
		{
			name: "NYC timezone shifts date backward",
			ts:   time.Date(2024, 6, 15, 3, 0, 0, 0, utc).Unix(),
			loc:  nyc,
			expected: BarCalendar{
				Year: 2024, Month: 6, DayOfMonth: 14,
				Hour: 23, Minute: 0, Second: 0,
				DayOfWeek: 6, WeekOfYear: 24,
			},
		},
		{
			name: "Tokyo timezone shifts date forward",
			ts:   time.Date(2024, 3, 15, 20, 30, 0, 0, utc).Unix(),
			loc:  tokyo,
			expected: BarCalendar{
				Year: 2024, Month: 3, DayOfMonth: 16,
				Hour: 5, Minute: 30, Second: 0,
				DayOfWeek: 7, WeekOfYear: 11,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cal := DecomposeBarTime(tt.ts, tt.loc)

			if cal.Year != tt.expected.Year {
				t.Errorf("Year = %v, want %v", cal.Year, tt.expected.Year)
			}
			if cal.Month != tt.expected.Month {
				t.Errorf("Month = %v, want %v", cal.Month, tt.expected.Month)
			}
			if cal.DayOfMonth != tt.expected.DayOfMonth {
				t.Errorf("DayOfMonth = %v, want %v", cal.DayOfMonth, tt.expected.DayOfMonth)
			}
			if cal.Hour != tt.expected.Hour {
				t.Errorf("Hour = %v, want %v", cal.Hour, tt.expected.Hour)
			}
			if cal.Minute != tt.expected.Minute {
				t.Errorf("Minute = %v, want %v", cal.Minute, tt.expected.Minute)
			}
			if cal.Second != tt.expected.Second {
				t.Errorf("Second = %v, want %v", cal.Second, tt.expected.Second)
			}
			if cal.DayOfWeek != tt.expected.DayOfWeek {
				t.Errorf("DayOfWeek = %v, want %v", cal.DayOfWeek, tt.expected.DayOfWeek)
			}
			if cal.WeekOfYear != tt.expected.WeekOfYear {
				t.Errorf("WeekOfYear = %v, want %v", cal.WeekOfYear, tt.expected.WeekOfYear)
			}
		})
	}
}

/* Pine convention: 1=Sunday..7=Saturday (Go Weekday 0-based + 1) */
func TestDecomposeBarTime_PineDayOfWeekMapping(t *testing.T) {
	loc := time.UTC

	tests := []struct {
		date    time.Time
		wantDOW float64
		label   string
	}{
		{time.Date(2024, 3, 10, 0, 0, 0, 0, loc), 1, "Sunday"},
		{time.Date(2024, 3, 11, 0, 0, 0, 0, loc), 2, "Monday"},
		{time.Date(2024, 3, 12, 0, 0, 0, 0, loc), 3, "Tuesday"},
		{time.Date(2024, 3, 13, 0, 0, 0, 0, loc), 4, "Wednesday"},
		{time.Date(2024, 3, 14, 0, 0, 0, 0, loc), 5, "Thursday"},
		{time.Date(2024, 3, 15, 0, 0, 0, 0, loc), 6, "Friday"},
		{time.Date(2024, 3, 16, 0, 0, 0, 0, loc), 7, "Saturday"},
	}

	for _, tt := range tests {
		t.Run(tt.label, func(t *testing.T) {
			cal := DecomposeBarTime(tt.date.Unix(), loc)
			if cal.DayOfWeek != tt.wantDOW {
				t.Errorf("%s: DayOfWeek = %v, want %v", tt.label, cal.DayOfWeek, tt.wantDOW)
			}
		})
	}
}

/* US Eastern 2024-03-10: 02:00 → 03:00 spring forward */
func TestDecomposeBarTime_DST(t *testing.T) {
	nyc, _ := time.LoadLocation("America/New_York")

	beforeDST := time.Date(2024, 3, 10, 6, 30, 0, 0, time.UTC).Unix()
	afterDST := time.Date(2024, 3, 10, 7, 30, 0, 0, time.UTC).Unix()

	calBefore := DecomposeBarTime(beforeDST, nyc)
	calAfter := DecomposeBarTime(afterDST, nyc)

	if calBefore.Hour != 1 {
		t.Errorf("before DST: Hour = %v, want 1", calBefore.Hour)
	}
	if calAfter.Hour != 3 {
		t.Errorf("after DST: Hour = %v, want 3 (skipped 2)", calAfter.Hour)
	}
}

/* 2023-01-01 is ISO week 52 of 2022 (Sunday belongs to previous ISO week) */
func TestDecomposeBarTime_ISOWeekYearBoundary(t *testing.T) {
	loc := time.UTC

	tests := []struct {
		name     string
		date     time.Time
		wantWeek float64
	}{
		{"2024-01-01 week 1", time.Date(2024, 1, 1, 0, 0, 0, 0, loc), 1},
		{"2023-01-01 week 52", time.Date(2023, 1, 1, 0, 0, 0, 0, loc), 52},
		{"2020-12-31 week 53", time.Date(2020, 12, 31, 0, 0, 0, 0, loc), 53},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cal := DecomposeBarTime(tt.date.Unix(), loc)
			if cal.WeekOfYear != tt.wantWeek {
				t.Errorf("WeekOfYear = %v, want %v", cal.WeekOfYear, tt.wantWeek)
			}
		})
	}
}

func TestDecomposeBarTime_AllFieldsFloat64(t *testing.T) {
	loc := time.UTC
	cal := DecomposeBarTime(time.Date(2024, 7, 15, 12, 30, 45, 0, loc).Unix(), loc)

	fields := map[string]float64{
		"Year":       cal.Year,
		"Month":      cal.Month,
		"DayOfMonth": cal.DayOfMonth,
		"Hour":       cal.Hour,
		"Minute":     cal.Minute,
		"Second":     cal.Second,
		"DayOfWeek":  cal.DayOfWeek,
		"WeekOfYear": cal.WeekOfYear,
	}

	for name, val := range fields {
		if val != val { // NaN check
			t.Errorf("%s returned NaN", name)
		}
		if val < 0 {
			t.Errorf("%s returned negative value: %v", name, val)
		}
	}
}

func TestDecomposeBarTime_NegativeTimestamp(t *testing.T) {
	loc := time.UTC
	cal := DecomposeBarTime(-86400, loc)

	if cal.Year != 1969 {
		t.Errorf("Year = %v, want 1969", cal.Year)
	}
	if cal.Month != 12 {
		t.Errorf("Month = %v, want 12", cal.Month)
	}
	if cal.DayOfMonth != 31 {
		t.Errorf("DayOfMonth = %v, want 31", cal.DayOfMonth)
	}
}
