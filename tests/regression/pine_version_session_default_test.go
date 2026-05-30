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
  - v4 with no ":DAYS" suffix defaults to "23456" (Mon-Fri only).
  - v5 with no ":DAYS" suffix defaults to "1234567" (all 7 days).

Generalised, symbol-agnostic: exercises the pure session API; no strategy
codegen required so this stays fast and deterministic.
*/
func TestPineVersionSessionDefault_Ratchet(t *testing.T) {
	saturday := time.Date(2025, time.August, 30, 12, 0, 0, 0, time.UTC)
	sunday := time.Date(2025, time.August, 31, 12, 0, 0, 0, time.UTC)
	monday := time.Date(2025, time.September, 1, 12, 0, 0, 0, time.UTC)

	cases := []struct {
		name        string
		pineVersion int
		sessionStr  string
		when        time.Time
		want        bool
	}{
		{"v4_default_rejects_saturday", 4, "0950-1645", saturday, false},
		{"v4_default_rejects_sunday", 4, "0950-1645", sunday, false},
		{"v4_default_accepts_monday", 4, "0950-1645", monday, true},

		{"v5_default_accepts_saturday", 5, "0950-1645", saturday, true},
		{"v5_default_accepts_sunday", 5, "0950-1645", sunday, true},
		{"v5_default_accepts_monday", 5, "0950-1645", monday, true},

		// Explicit suffix overrides the per-version default.
		{"v4_explicit_all_days_accepts_saturday", 4, "0950-1645:1234567", saturday, true},
		{"v5_explicit_weekdays_only_rejects_saturday", 5, "0950-1645:23456", saturday, false},
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
