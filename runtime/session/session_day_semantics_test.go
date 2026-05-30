package session

import (
	"math"
	"testing"
	"time"
)

/*
TestSession_DaySemantics_PerVersion codifies the v4-vs-v5 default DAYS
contract per the Pine v5 sessions breaking change. The test asserts the
intersection of weekday and HHMM range for both IsInSession and the
version-aware TimeFuncWithVersion entry point.

v4 default DAYS = "23456" (Mon-Fri only).
v5 default DAYS = "1234567" (all 7 days).
Explicit ":DDDDDDD" suffix overrides the per-version default regardless of
Pine version.
*/
func TestSession_DaySemantics_PerVersion(t *testing.T) {
	// Anchor weekdays in 2025 used across the table.
	// 2025-08-30 is a Saturday; 2025-08-31 is a Sunday; 2025-08-29 is a Friday;
	// 2025-09-01 is a Monday. All times constructed in the named tz at HH:MM.
	mkTime := func(t *testing.T, year int, month time.Month, day, hour, minute int, tzName string) time.Time {
		t.Helper()
		loc, err := time.LoadLocation(tzName)
		if err != nil {
			t.Fatalf("load tz %q: %v", tzName, err)
		}
		return time.Date(year, month, day, hour, minute, 0, 0, loc)
	}

	cases := []struct {
		name        string
		pineVersion int
		sessionStr  string
		when        time.Time
		timezone    string
		inSession   bool // expected IsInSession
		nonNaN      bool // expected TimeFuncWithVersion returns non-NaN
	}{
		// v4 no-suffix → default 23456 (Mon-Fri)
		{"v4_no_suffix_mon_11_00", 4, "0950-1645",
			mkTime(t, 2025, time.September, 1, 11, 0, "Europe/Moscow"), "Europe/Moscow", true, true},
		{"v4_no_suffix_sat_11_00", 4, "0950-1645",
			mkTime(t, 2025, time.August, 30, 11, 0, "Europe/Moscow"), "Europe/Moscow", false, false},
		{"v4_no_suffix_sun_11_00", 4, "0950-1645",
			mkTime(t, 2025, time.August, 31, 11, 0, "Europe/Moscow"), "Europe/Moscow", false, false},
		{"v4_no_suffix_sat_11_00_utc", 4, "0950-1645",
			mkTime(t, 2025, time.August, 30, 11, 0, "UTC"), "UTC", false, false},
		// v5 no-suffix → default 1234567 (all 7 days)
		{"v5_no_suffix_sat_11_00", 5, "0950-1645",
			mkTime(t, 2025, time.August, 30, 11, 0, "UTC"), "UTC", true, true},
		{"v5_no_suffix_sun_11_00", 5, "0950-1645",
			mkTime(t, 2025, time.August, 31, 11, 0, "UTC"), "UTC", true, true},
		// Explicit ":DAYS" overrides per-version default
		{"v4_explicit_1234567_sat", 4, "0950-1645:1234567",
			mkTime(t, 2025, time.August, 30, 11, 0, "UTC"), "UTC", true, true},
		{"v5_explicit_23456_sat", 5, "0950-1645:23456",
			mkTime(t, 2025, time.August, 30, 11, 0, "UTC"), "UTC", false, false},
		// Off-hours regardless of weekday
		{"v4_off_hours_mon_08_00", 4, "0950-1645",
			mkTime(t, 2025, time.September, 1, 8, 0, "Europe/Moscow"), "Europe/Moscow", false, false},
		{"v5_off_hours_mon_17_00", 5, "0950-1645",
			mkTime(t, 2025, time.September, 1, 17, 0, "UTC"), "UTC", false, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s, err := ParseWithVersion(tc.sessionStr, tc.pineVersion)
			if err != nil {
				t.Fatalf("ParseWithVersion(%q, v=%d): %v", tc.sessionStr, tc.pineVersion, err)
			}

			gotIn := s.IsInSession(tc.when.UnixMilli(), tc.timezone)
			if gotIn != tc.inSession {
				t.Fatalf("IsInSession(%s @ %s, v=%d, sess=%q) = %v, want %v",
					tc.when.Format(time.RFC3339), tc.timezone, tc.pineVersion, tc.sessionStr,
					gotIn, tc.inSession)
			}

			gotF := TimeFuncWithVersion(tc.when.UnixMilli(), "1h", tc.sessionStr, tc.timezone, tc.pineVersion)
			gotNonNaN := !math.IsNaN(gotF)
			if gotNonNaN != tc.nonNaN {
				t.Fatalf("TimeFuncWithVersion(%s @ %s, v=%d, sess=%q) nonNaN = %v, want %v (raw=%v)",
					tc.when.Format(time.RFC3339), tc.timezone, tc.pineVersion, tc.sessionStr,
					gotNonNaN, tc.nonNaN, gotF)
			}
		})
	}
}

/*
TestSession_ExplicitInvalidDaysSuffix exercises the validation path for the
DAYS suffix: any digit outside 1-7 must produce a parse error.
*/
func TestSession_ExplicitInvalidDaysSuffix(t *testing.T) {
	cases := []string{
		"0950-1645:0",
		"0950-1645:8",
		"0950-1645:12X",
		"0950-1645:",
	}
	for _, in := range cases {
		t.Run(in, func(t *testing.T) {
			if _, err := ParseWithVersion(in, 5); err == nil {
				t.Fatalf("ParseWithVersion(%q) = nil err; want validation error", in)
			}
		})
	}
}

/*
TestParse_LegacyShim_DefaultsToV5 ensures the deprecated single-arg Parse
preserves backward-compatible behaviour for callers that have not yet been
threaded with the Pine version: v5 default = all 7 days.
*/
func TestParse_LegacyShim_DefaultsToV5(t *testing.T) {
	s, err := Parse("0950-1645")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	// Saturday 2025-08-30 12:00 UTC should be IN under v5 default.
	sat := time.Date(2025, time.August, 30, 12, 0, 0, 0, time.UTC).UnixMilli()
	if !s.IsInSession(sat, "UTC") {
		t.Fatalf("legacy Parse should default to v5 (all 7 days); Saturday rejected")
	}
}
