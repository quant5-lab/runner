package session

import (
	"math"
	"testing"
	"time"
)

// TestSession_DayFilter is the comprehensive day-mask specification for all session
// kinds. It covers the three orthogonal dimensions of the filter algorithm:
//
//   - Session kind: intraday, overnight (start > end), full-cycle overnight (start == end),
//     and 24-hour ("0000-2359").
//   - Mask source: pre-v5 versions (v1-v4) default to Mon-Fri; v5+ default to all 7
//     days; future/unrecognised versions (>5) fall forward to v5 semantics;
//     explicit ":DAYS" suffix always overrides the version default.
//   - Day assignment rule: intraday bars use their own calendar weekday; overnight bars
//     are assigned to the session's END day (pre-midnight half → tomorrow, post-midnight
//     half → today).
//
// Anchor week (UTC unless tz noted):
//
//	2025-08-30 Sat, 2025-08-31 Sun, 2025-09-01 Mon … 2025-09-05 Fri, 2025-09-06 Sat, 2025-09-07 Sun
func TestSession_DayFilter(t *testing.T) {
	at := func(t *testing.T, y int, mo time.Month, d, h, m int, tz string) time.Time {
		t.Helper()
		loc, err := time.LoadLocation(tz)
		if err != nil {
			t.Fatalf("LoadLocation(%q): %v", tz, err)
		}
		return time.Date(y, mo, d, h, m, 0, 0, loc)
	}

	cases := []struct {
		name        string
		pineVersion int
		sessionStr  string
		when        time.Time
		timezone    string
		wantIn      bool
	}{
		// ── Intraday sessions: bar's own calendar weekday is the session day ─────────
		// v4 default (Mon-Fri): all five weekdays accepted when in-hours.
		{"intraday_v4_monday_in_hours", 4, "0950-1645", at(t, 2025, time.September, 1, 11, 0, "UTC"), "UTC", true},
		{"intraday_v4_tuesday_in_hours", 4, "0950-1645", at(t, 2025, time.September, 2, 11, 0, "UTC"), "UTC", true},
		{"intraday_v4_wednesday_in_hours", 4, "0950-1645", at(t, 2025, time.September, 3, 11, 0, "UTC"), "UTC", true},
		{"intraday_v4_thursday_in_hours", 4, "0950-1645", at(t, 2025, time.September, 4, 11, 0, "UTC"), "UTC", true},
		{"intraday_v4_friday_in_hours", 4, "0950-1645", at(t, 2025, time.September, 5, 11, 0, "UTC"), "UTC", true},
		// v4 default: Saturday and Sunday rejected.
		{"intraday_v4_saturday_in_hours", 4, "0950-1645", at(t, 2025, time.August, 30, 11, 0, "UTC"), "UTC", false},
		{"intraday_v4_sunday_in_hours", 4, "0950-1645", at(t, 2025, time.August, 31, 11, 0, "UTC"), "UTC", false},
		// v4 default: off-hours weekday is rejected by the time filter regardless of mask.
		{"intraday_v4_monday_off_hours", 4, "0950-1645", at(t, 2025, time.September, 1, 8, 0, "UTC"), "UTC", false},

		// v5 default (all 7 days): Saturday and Sunday accepted when in-hours.
		{"intraday_v5_saturday_in_hours", 5, "0950-1645", at(t, 2025, time.August, 30, 11, 0, "UTC"), "UTC", true},
		{"intraday_v5_sunday_in_hours", 5, "0950-1645", at(t, 2025, time.August, 31, 11, 0, "UTC"), "UTC", true},
		// v5 default: off-hours rejected regardless of day.
		{"intraday_v5_monday_off_hours", 5, "0950-1645", at(t, 2025, time.September, 1, 17, 0, "UTC"), "UTC", false},

		// Pre-v5 versions (v1, v2, v3) use Mon-Fri default, matching the TV docs:
		// "earlier versions where 23456 (weekdays) is used."
		// Both rejection (weekends) and acceptance (weekdays) are probed so the mask
		// cannot be an all-reject stub that satisfies only the rejection side.
		{"intraday_v1_rejects_saturday", 1, "0950-1645", at(t, 2025, time.August, 30, 11, 0, "UTC"), "UTC", false},
		{"intraday_v1_accepts_wednesday", 1, "0950-1645", at(t, 2025, time.September, 3, 11, 0, "UTC"), "UTC", true},
		{"intraday_v2_rejects_saturday", 2, "0950-1645", at(t, 2025, time.August, 30, 11, 0, "UTC"), "UTC", false},
		{"intraday_v2_accepts_friday", 2, "0950-1645", at(t, 2025, time.September, 5, 11, 0, "UTC"), "UTC", true},
		{"intraday_v3_rejects_sunday", 3, "0950-1645", at(t, 2025, time.August, 31, 11, 0, "UTC"), "UTC", false},
		{"intraday_v3_accepts_monday", 3, "0950-1645", at(t, 2025, time.September, 1, 11, 0, "UTC"), "UTC", true},
		// Future/unrecognised versions (>5) fall forward to all-7-days (v5 semantics).
		{"intraday_v6_accepts_saturday", 6, "0950-1645", at(t, 2025, time.August, 30, 11, 0, "UTC"), "UTC", true},

		// Explicit ":DAYS" suffix overrides the per-version default in both directions.
		{"intraday_v4_explicit_all_days_saturday", 4, "0950-1645:1234567", at(t, 2025, time.August, 30, 11, 0, "UTC"), "UTC", true},
		{"intraday_v5_explicit_weekdays_saturday", 5, "0950-1645:23456", at(t, 2025, time.August, 30, 11, 0, "UTC"), "UTC", false},

		// Timezone: HHMM is interpreted in exchange local time, not UTC.
		{"intraday_v4_moscow_saturday_in_hours", 4, "0950-1645", at(t, 2025, time.August, 30, 11, 0, "Europe/Moscow"), "Europe/Moscow", false},
		{"intraday_v5_moscow_saturday_in_hours", 5, "0950-1645", at(t, 2025, time.August, 30, 11, 0, "Europe/Moscow"), "Europe/Moscow", true},

		// ── Overnight sessions: bar is assigned to its END day ───────────────────────
		// v4 default (Mon-Fri), "1800-0600":
		// Pre-midnight half (cur ≥ 18:00): session ends tomorrow → (weekday+1) checked.
		// All five Mon-Fri end-days: the evening before each weekday must be IN.
		{"overnight_v4_sun_19_ends_mon", 4, "1800-0600", at(t, 2025, time.August, 31, 19, 0, "UTC"), "UTC", true},
		{"overnight_v4_mon_19_ends_tue", 4, "1800-0600", at(t, 2025, time.September, 1, 19, 0, "UTC"), "UTC", true},
		{"overnight_v4_tue_19_ends_wed", 4, "1800-0600", at(t, 2025, time.September, 2, 19, 0, "UTC"), "UTC", true},
		{"overnight_v4_wed_19_ends_thu", 4, "1800-0600", at(t, 2025, time.September, 3, 19, 0, "UTC"), "UTC", true},
		{"overnight_v4_thu_19_ends_fri", 4, "1800-0600", at(t, 2025, time.September, 4, 19, 0, "UTC"), "UTC", true},
		// Evenings that end on a weekend day → rejected.
		{"overnight_v4_fri_19_ends_sat", 4, "1800-0600", at(t, 2025, time.September, 5, 19, 0, "UTC"), "UTC", false},
		{"overnight_v4_sat_19_ends_sun", 4, "1800-0600", at(t, 2025, time.September, 6, 19, 0, "UTC"), "UTC", false},

		// Post-midnight half (cur < 18:00): session ends today → bar's own weekday checked.
		{"overnight_v4_mon_02_bar_is_mon", 4, "1800-0600", at(t, 2025, time.September, 1, 2, 0, "UTC"), "UTC", true},
		{"overnight_v4_sat_02_bar_is_sat", 4, "1800-0600", at(t, 2025, time.September, 6, 2, 0, "UTC"), "UTC", false},

		// v3 default (Mon-Fri, same as v4): end-day rule applies identically.
		{"overnight_v3_sun_19_ends_mon", 3, "1800-0600", at(t, 2025, time.August, 31, 19, 0, "UTC"), "UTC", true},
		{"overnight_v3_fri_19_ends_sat", 3, "1800-0600", at(t, 2025, time.September, 5, 19, 0, "UTC"), "UTC", false},

		// v5 default (all 7 days): every in-range bar is accepted.
		{"overnight_v5_fri_19_all_days", 5, "1800-0600", at(t, 2025, time.September, 5, 19, 0, "UTC"), "UTC", true},
		{"overnight_v5_sat_02_all_days", 5, "1800-0600", at(t, 2025, time.September, 6, 2, 0, "UTC"), "UTC", true},

		// Explicit suffix: end-day rule applies regardless of suffix source.
		{"overnight_explicit_23456_sun_19_ends_mon", 5, "1800-0600:23456", at(t, 2025, time.August, 31, 19, 0, "UTC"), "UTC", true},
		{"overnight_explicit_23456_fri_19_ends_sat", 5, "1800-0600:23456", at(t, 2025, time.September, 5, 19, 0, "UTC"), "UTC", false},
		// ":1" = Sunday only.
		// Saturday 19:00 pre-midnight → ends Sunday → Sunday in ":1" → accepted.
		// Sunday 02:00 post-midnight → bar's day = Sunday → Sunday in ":1" → accepted.
		// Sunday 19:00 pre-midnight → ends Monday → Monday NOT in ":1" → rejected.
		{"overnight_explicit_sun_only_sat_19_ends_sun", 4, "1800-0600:1", at(t, 2025, time.September, 6, 19, 0, "UTC"), "UTC", true},
		{"overnight_explicit_sun_only_sun_02_bar_is_sun", 4, "1800-0600:1", at(t, 2025, time.September, 7, 2, 0, "UTC"), "UTC", true},
		{"overnight_explicit_sun_only_sun_19_ends_mon", 4, "1800-0600:1", at(t, 2025, time.August, 31, 19, 0, "UTC"), "UTC", false},

		// Time range gap (between end and start): rejected before reaching day-mask check.
		{"overnight_v4_off_hours_gap", 4, "1800-0600", at(t, 2025, time.September, 1, 14, 0, "UTC"), "UTC", false},

		// Timezone: end-day is determined after timestamp is converted to exchange local time.
		{"overnight_v4_moscow_sun_22_ends_mon", 4, "1800-0600", at(t, 2025, time.August, 31, 22, 0, "Europe/Moscow"), "Europe/Moscow", true},

		// ── Full-cycle overnight (start == end, e.g. "1700-1700") ────────────────────
		// Every clock time falls within the time range; only the day mask gates bars.
		// v4 (Mon-Fri): end-day rule still applies (Sunday 19:00 → ends Monday → IN).
		{"full_cycle_v4_sun_19_ends_mon", 4, "1700-1700", at(t, 2025, time.August, 31, 19, 0, "UTC"), "UTC", true},
		{"full_cycle_v4_mon_02_bar_is_mon", 4, "1700-1700", at(t, 2025, time.September, 1, 2, 0, "UTC"), "UTC", true},
		{"full_cycle_v4_fri_19_ends_sat", 4, "1700-1700", at(t, 2025, time.September, 5, 19, 0, "UTC"), "UTC", false},
		{"full_cycle_v4_sat_02_bar_is_sat", 4, "1700-1700", at(t, 2025, time.September, 6, 2, 0, "UTC"), "UTC", false},
		// v5 (all 7 days): any bar on any day and any time → IN.
		{"full_cycle_v5_friday_19_all_days", 5, "1700-1700", at(t, 2025, time.September, 5, 19, 0, "UTC"), "UTC", true},
		{"full_cycle_v5_saturday_midnight_all_days", 5, "1700-1700", at(t, 2025, time.September, 6, 0, 0, "UTC"), "UTC", true},

		// ── 24-hour sessions ("0000-2359"): is24Hour fast-path, day mask still applies ─
		{"24h_v4_monday_midday", 4, "0000-2359", at(t, 2025, time.September, 1, 12, 0, "UTC"), "UTC", true},
		{"24h_v4_saturday_midday", 4, "0000-2359", at(t, 2025, time.August, 30, 12, 0, "UTC"), "UTC", false},
		{"24h_v3_saturday_midday", 3, "0000-2359", at(t, 2025, time.August, 30, 12, 0, "UTC"), "UTC", false},
		{"24h_v5_saturday_midday", 5, "0000-2359", at(t, 2025, time.August, 30, 12, 0, "UTC"), "UTC", true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s, err := ParseWithVersion(tc.sessionStr, tc.pineVersion)
			if err != nil {
				t.Fatalf("ParseWithVersion(%q, v=%d): %v", tc.sessionStr, tc.pineVersion, err)
			}

			gotIn := s.IsInSession(tc.when.UnixMilli(), tc.timezone)
			if gotIn != tc.wantIn {
				t.Errorf("IsInSession(%s tz=%s, v=%d, sess=%q) = %v, want %v",
					tc.when.Format(time.RFC3339), tc.timezone, tc.pineVersion, tc.sessionStr,
					gotIn, tc.wantIn)
			}

			gotF := TimeFuncWithVersion(tc.when.UnixMilli(), "1h", tc.sessionStr, tc.timezone, tc.pineVersion)
			gotNonNaN := !math.IsNaN(gotF)
			if gotNonNaN != tc.wantIn {
				t.Errorf("TimeFuncWithVersion(%s tz=%s, v=%d, sess=%q) nonNaN=%v, want %v",
					tc.when.Format(time.RFC3339), tc.timezone, tc.pineVersion, tc.sessionStr,
					gotNonNaN, tc.wantIn)
			}
		})
	}
}

// TestSession_ExplicitInvalidDaysSuffix exercises the validation path for the DAYS
// suffix: any digit outside 1-7 must produce a parse error.
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

// TestParse_LegacyShim_DefaultsToV5 ensures the deprecated single-arg Parse preserves
// backward-compatible behaviour: v5 default = all 7 days.
func TestParse_LegacyShim_DefaultsToV5(t *testing.T) {
	s, err := Parse("0950-1645")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	sat := time.Date(2025, time.August, 30, 12, 0, 0, 0, time.UTC).UnixMilli()
	if !s.IsInSession(sat, "UTC") {
		t.Fatal("Parse() must default to v5 (all 7 days); Saturday was rejected")
	}
}
