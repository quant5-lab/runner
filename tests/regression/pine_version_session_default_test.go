package regression

import (
	"math"
	"testing"
	"time"

	"github.com/quant5-lab/runner/runtime/session"
)

/*
TestPineVersionSessionDefault_Ratchet codifies the v4-vs-v5 default DAYS
contract at the package boundary so the regression suite catches any
regression that re-enables weekend bars for v4 strategies (or accidentally
restricts v5 strategies to Mon-Fri).

Per the Pine v4→v5 sessions breaking change documented at
https://www.tradingview.com/pine-script-docs/v5/concepts/sessions/ :
  - All versions prior to v5 (v1-v4) with no ":DAYS" suffix default to "23456" (Mon-Fri only).
  - v5 and later with no ":DAYS" suffix default to "1234567" (all 7 days).
  - An explicit ":DAYS" suffix always overrides the version default.

Generalised, symbol-agnostic: exercises the pure session API; no strategy
codegen required so this stays fast and deterministic.
*/
func TestPineVersionSessionDefault_Ratchet(t *testing.T) {
	// Anchor week at 12:00 UTC — covers all 7 weekdays across two calendar weeks.
	friday := time.Date(2025, time.August, 29, 12, 0, 0, 0, time.UTC)
	saturday := time.Date(2025, time.August, 30, 12, 0, 0, 0, time.UTC)
	sunday := time.Date(2025, time.August, 31, 12, 0, 0, 0, time.UTC)
	monday := time.Date(2025, time.September, 1, 12, 0, 0, 0, time.UTC)
	wednesday := time.Date(2025, time.September, 3, 12, 0, 0, 0, time.UTC)

	cases := []struct {
		name        string
		pineVersion int
		sessionStr  string
		when        time.Time
		timezone    string
		want        bool
	}{
		// v4 no-suffix → default 23456 (Mon-Fri only).
		{"v4_default_rejects_saturday", 4, "0950-1645", saturday, "UTC", false},
		{"v4_default_rejects_sunday", 4, "0950-1645", sunday, "UTC", false},
		{"v4_default_accepts_monday", 4, "0950-1645", monday, "UTC", true},
		{"v4_default_accepts_friday", 4, "0950-1645", friday, "UTC", true},

		// v5 no-suffix → default 1234567 (all 7 days).
		{"v5_default_accepts_saturday", 5, "0950-1645", saturday, "UTC", true},
		{"v5_default_accepts_sunday", 5, "0950-1645", sunday, "UTC", true},
		{"v5_default_accepts_monday", 5, "0950-1645", monday, "UTC", true},

		// Explicit suffix overrides the per-version default regardless of version.
		{"v4_explicit_all_days_accepts_saturday", 4, "0950-1645:1234567", saturday, "UTC", true},
		{"v5_explicit_weekdays_only_rejects_saturday", 5, "0950-1645:23456", saturday, "UTC", false},

		{"v4_moscow_rejects_saturday", 4, "0950-1645", saturday, "Europe/Moscow", false},
		{"v5_moscow_accepts_saturday", 5, "0950-1645", saturday, "Europe/Moscow", true},

		// Pre-v5 versions (v1-v3) use Mon-Fri default, same as v4.
		// Both rejection and acceptance are probed so the mask cannot be an all-reject stub.
		{"v1_rejects_saturday", 1, "0950-1645", saturday, "UTC", false},
		{"v1_accepts_monday", 1, "0950-1645", monday, "UTC", true},
		{"v2_rejects_sunday", 2, "0950-1645", sunday, "UTC", false},
		{"v2_accepts_wednesday", 2, "0950-1645", wednesday, "UTC", true},
		{"v3_rejects_saturday", 3, "0950-1645", saturday, "UTC", false},
		{"v3_accepts_friday", 3, "0950-1645", friday, "UTC", true},
		// Future/unrecognised versions (>5) fall forward to all-7-days.
		{"v6_accepts_saturday", 6, "0950-1645", saturday, "UTC", true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := session.TimeFuncWithVersion(tc.when.UnixMilli(), "1h", tc.sessionStr, tc.timezone, tc.pineVersion)
			got := !math.IsNaN(f)
			if got != tc.want {
				t.Fatalf("TimeFuncWithVersion(v=%d, sess=%q, tz=%q, %s) inSession=%v want %v",
					tc.pineVersion, tc.sessionStr, tc.timezone, tc.when.Format(time.RFC3339), got, tc.want)
			}
		})
	}
}

// TestOvernightSessionEndDayAssignment_Ratchet locks the Pine overnight-session
// END-day rule at the regression boundary. An overnight session is assigned to the
// day it ends on — so a v4 "1800-0600" session accepts Sunday-evening bars (end day
// = Monday, in Mon-Fri mask) and rejects Friday-evening bars (end day = Saturday,
// not in Mon-Fri mask). This prevents a regression back to calendar-weekday checking.
func TestOvernightSessionEndDayAssignment_Ratchet(t *testing.T) {
	// 2025-08-31 Sun, 2025-09-01 Mon, 2025-09-04 Thu, 2025-09-05 Fri, 2025-09-06 Sat — all UTC.
	sunEvening := time.Date(2025, time.August, 31, 19, 0, 0, 0, time.UTC)   // Sunday 19:00 pre-midnight
	monMorning := time.Date(2025, time.September, 1, 2, 0, 0, 0, time.UTC)  // Monday 02:00 post-midnight
	thuEvening := time.Date(2025, time.September, 4, 19, 0, 0, 0, time.UTC) // Thursday 19:00
	friEvening := time.Date(2025, time.September, 5, 19, 0, 0, 0, time.UTC) // Friday 19:00 pre-midnight
	satMorning := time.Date(2025, time.September, 6, 2, 0, 0, 0, time.UTC)  // Saturday 02:00 post-midnight
	satEvening := time.Date(2025, time.September, 6, 19, 0, 0, 0, time.UTC) // Saturday 19:00 pre-midnight
	sunMorning := time.Date(2025, time.September, 7, 2, 0, 0, 0, time.UTC)  // Sunday 02:00 post-midnight

	cases := []struct {
		name        string
		pineVersion int
		sessionStr  string
		when        time.Time
		want        bool
	}{
		// v4 default "1800-0600" — Mon-Fri mask applied against END day.
		{"v4_sunday_evening_ends_monday_IN", 4, "1800-0600", sunEvening, true},
		{"v4_monday_morning_post_midnight_IN", 4, "1800-0600", monMorning, true},
		{"v4_thursday_evening_ends_friday_IN", 4, "1800-0600", thuEvening, true},
		{"v4_friday_evening_ends_saturday_OUT", 4, "1800-0600", friEvening, false},
		{"v4_saturday_morning_post_midnight_OUT", 4, "1800-0600", satMorning, false},

		// v5 default "1800-0600" — all 7 days → any in-range bar is IN.
		{"v5_friday_evening_all_days_IN", 5, "1800-0600", friEvening, true},
		{"v5_saturday_morning_all_days_IN", 5, "1800-0600", satMorning, true},

		// Explicit ":1" (Sunday only) on overnight session: Saturday evening → ends Sunday → IN.
		{"explicit_sunday_only_saturday_evening_IN", 4, "1800-0600:1", satEvening, true},
		{"explicit_sunday_only_sunday_morning_IN", 4, "1800-0600:1", sunMorning, true},
		// Sunday evening → ends Monday → Monday NOT in ":1" → OUT.
		{"explicit_sunday_only_sunday_evening_OUT", 4, "1800-0600:1", sunEvening, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := session.TimeFuncWithVersion(tc.when.UnixMilli(), "1h", tc.sessionStr, "UTC", tc.pineVersion)
			got := !math.IsNaN(f)
			if got != tc.want {
				t.Fatalf("TimeFuncWithVersion(v=%d, sess=%q, %s) inSession=%v want %v",
					tc.pineVersion, tc.sessionStr, tc.when.Format(time.RFC3339), got, tc.want)
			}
		})
	}
}
