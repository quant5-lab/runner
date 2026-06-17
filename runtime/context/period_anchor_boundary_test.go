package context

import (
	"testing"
	"time"
)

// mustUnixInZone parses "YYYY-MM-DD HH:MM:SS" as a wall-clock time in the
// named IANA timezone and returns the corresponding Unix timestamp in seconds.
// It is the timezone-aware complement to the existing mustUnix helper.
func mustUnixInZone(t *testing.T, s, tz string) int64 {
	t.Helper()
	loc, err := time.LoadLocation(tz)
	if err != nil {
		t.Fatalf("load timezone %q: %v", tz, err)
	}
	ts, err := time.ParseInLocation("2006-01-02 15:04:05", s, loc)
	if err != nil {
		t.Fatalf("parse %q in %s: %v", s, tz, err)
	}
	return ts.Unix()
}

// TestAlignToPeriodWithAnchor_ZeroAnchorEqualsLegacyAcrossAllTimeframes verifies
// that AlignToPeriodWithAnchor with the zero PeriodAnchor returns the same value
// as AlignToPeriod across all timeframe classes (numeric intraday, suffixed
// intraday, daily, weekly, monthly).  This is the backward-compatibility
// contract: adding the anchor parameter must not change any existing behavior
// when the anchor carries no exchange-specific information.
func TestAlignToPeriodWithAnchor_ZeroAnchorEqualsLegacyAcrossAllTimeframes(t *testing.T) {
	a := NewTimeframeBoundaryAligner()
	zero := PeriodAnchor{}

	type row struct {
		name string
		ts   int64
		tf   string
	}
	ts := mustUnix(t, "2025-08-15 09:37:00")
	rows := []row{
		{"numeric_intraday_5m", ts, "5"},
		{"numeric_intraday_15m", ts, "15"},
		{"numeric_intraday_60m", ts, "60"},
		{"numeric_intraday_240m", ts, "240"},
		{"suffixed_1h", ts, "1h"},
		{"suffixed_4h", ts, "4h"},
		{"suffixed_5m", ts, "5m"},
		{"daily_D", mustUnix(t, "2025-08-15 14:30:00"), "1D"},
		{"weekly_W", mustUnix(t, "2025-08-15 14:30:00"), "W"},
		{"monthly_M", mustUnix(t, "2025-08-15 12:00:00"), "M"},
		{"empty_passthrough", ts, ""},
	}

	for _, r := range rows {
		t.Run(r.name, func(t *testing.T) {
			withAnchor := a.AlignToPeriodWithAnchor(r.ts, r.tf, zero)
			legacy := a.AlignToPeriod(r.ts, r.tf)
			if withAnchor != legacy {
				t.Errorf("AlignToPeriodWithAnchor(ts, %q, zero)=%s differs from AlignToPeriod(ts, %q)=%s — backward compat broken",
					r.tf, fmtUnix(withAnchor), r.tf, fmtUnix(legacy))
			}
		})
	}
}

// TestAlignToPeriodWithAnchor_IntradaySessionTilesFromSessionOpen verifies that
// every bar within a session-anchored intraday window maps to the same boundary
// and that the boundary equals sessionOpen + k*period for the correct k.
// This covers multiple timezones, session-open offsets, and period lengths.
func TestAlignToPeriodWithAnchor_IntradaySessionTilesFromSessionOpen(t *testing.T) {
	a := NewTimeframeBoundaryAligner()

	type windowCase struct {
		name   string
		anchor PeriodAnchor
		tf     string
		// boundary and all bars expressed as wall-clock times in the anchor timezone
		boundaryLocal string
		barsLocal     []string
	}

	cases := []windowCase{
		// MOEX 07:00 MSK session, 4h period — four session windows
		{
			"moex_4h_first_window",
			PeriodAnchor{Timezone: "Europe/Moscow", SessionOpenMinute: 7 * 60},
			"240",
			"2025-08-15 07:00:00",
			[]string{"2025-08-15 07:00:00", "2025-08-15 09:37:00", "2025-08-15 10:59:00"},
		},
		{
			"moex_4h_second_window",
			PeriodAnchor{Timezone: "Europe/Moscow", SessionOpenMinute: 7 * 60},
			"240",
			"2025-08-15 11:00:00",
			[]string{"2025-08-15 11:00:00", "2025-08-15 13:30:00", "2025-08-15 14:59:00"},
		},
		{
			"moex_4h_third_window",
			PeriodAnchor{Timezone: "Europe/Moscow", SessionOpenMinute: 7 * 60},
			"240",
			"2025-08-15 15:00:00",
			[]string{"2025-08-15 15:00:00", "2025-08-15 17:45:00", "2025-08-15 18:59:00"},
		},
		{
			"moex_4h_fourth_window",
			PeriodAnchor{Timezone: "Europe/Moscow", SessionOpenMinute: 7 * 60},
			"240",
			"2025-08-15 19:00:00",
			[]string{"2025-08-15 19:00:00", "2025-08-15 22:59:00"},
		},
		// Hypothetical session at 09:30 local, 1h period — two consecutive windows
		{
			"session_09h30_1h_first_window",
			PeriodAnchor{Timezone: "Europe/Moscow", SessionOpenMinute: 9*60 + 30},
			"60",
			"2025-08-15 09:30:00",
			[]string{"2025-08-15 09:30:00", "2025-08-15 09:59:00", "2025-08-15 10:29:00"},
		},
		{
			"session_09h30_1h_second_window",
			PeriodAnchor{Timezone: "Europe/Moscow", SessionOpenMinute: 9*60 + 30},
			"60",
			"2025-08-15 10:30:00",
			[]string{"2025-08-15 10:30:00", "2025-08-15 11:00:00", "2025-08-15 11:29:00"},
		},
		// Midnight UTC session (zero-offset anchor) — degenerates to UTC arithmetic
		{
			"midnight_utc_4h",
			PeriodAnchor{Timezone: "UTC", SessionOpenMinute: 0},
			"240",
			"2025-08-15 08:00:00",
			[]string{"2025-08-15 08:00:00", "2025-08-15 09:37:00", "2025-08-15 11:59:00"},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			wantBoundary := mustUnixInZone(t, c.boundaryLocal, c.anchor.Timezone)
			for _, barLocal := range c.barsLocal {
				bar := mustUnixInZone(t, barLocal, c.anchor.Timezone)
				got := a.AlignToPeriodWithAnchor(bar, c.tf, c.anchor)
				if got != wantBoundary {
					t.Errorf("bar %s: got %s, want %s (boundary of %s)",
						barLocal, fmtUnix(got), fmtUnix(wantBoundary), c.boundaryLocal)
				}
				if got > bar {
					t.Errorf("bar %s: boundary %s is after the bar (alignment must not advance time)",
						barLocal, fmtUnix(got))
				}
			}
		})
	}
}

// TestAlignToPeriodWithAnchor_CalendarBoundariesUseLocalMidnight verifies that
// daily, weekly, and monthly boundaries are computed at midnight in the
// anchor's timezone rather than at UTC midnight.  This is the correct Pine
// semantics for calendar-scale secondaries on non-UTC exchanges.
func TestAlignToPeriodWithAnchor_CalendarBoundariesUseLocalMidnight(t *testing.T) {
	a := NewTimeframeBoundaryAligner()

	type calendarCase struct {
		name      string
		anchor    PeriodAnchor
		tf        string
		barLocal  string
		wantLocal string
	}

	cases := []calendarCase{
		// Daily — midnight in MSK (= 21:00 UTC prev day)
		{"daily_msk_evening_bar", PeriodAnchor{Timezone: "Europe/Moscow"}, "1D",
			"2025-08-15 22:30:00", "2025-08-15 00:00:00"},
		{"daily_msk_early_bar", PeriodAnchor{Timezone: "Europe/Moscow"}, "1D",
			"2025-08-15 01:00:00", "2025-08-15 00:00:00"},
		{"daily_msk_on_midnight", PeriodAnchor{Timezone: "Europe/Moscow"}, "1D",
			"2025-08-15 00:00:00", "2025-08-15 00:00:00"},

		// Daily — midnight in ET (= 04:00 or 05:00 UTC, DST-aware)
		{"daily_et_market_close", PeriodAnchor{Timezone: "America/New_York"}, "1D",
			"2025-08-15 16:00:00", "2025-08-15 00:00:00"},
		{"daily_et_overnight", PeriodAnchor{Timezone: "America/New_York"}, "1D",
			"2025-08-15 23:30:00", "2025-08-15 00:00:00"},

		// Weekly — Monday midnight in MSK
		{"weekly_msk_friday", PeriodAnchor{Timezone: "Europe/Moscow"}, "W",
			"2025-08-15 14:30:00", "2025-08-11 00:00:00"}, // Aug 15 is Friday; Monday = Aug 11
		{"weekly_msk_monday_on_boundary", PeriodAnchor{Timezone: "Europe/Moscow"}, "W",
			"2025-08-11 00:00:00", "2025-08-11 00:00:00"},

		// Monthly — 1st of month midnight in MSK
		{"monthly_msk_mid_month", PeriodAnchor{Timezone: "Europe/Moscow"}, "M",
			"2025-08-15 12:00:00", "2025-08-01 00:00:00"},
		{"monthly_msk_last_day", PeriodAnchor{Timezone: "Europe/Moscow"}, "M",
			"2025-08-31 23:59:00", "2025-08-01 00:00:00"},

		// Monthly — ET (DST boundary test: June 30 23:00 ET is July 1 03:00 UTC)
		{"monthly_et_june_end", PeriodAnchor{Timezone: "America/New_York"}, "M",
			"2025-06-30 23:00:00", "2025-06-01 00:00:00"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			bar := mustUnixInZone(t, c.barLocal, c.anchor.Timezone)
			want := mustUnixInZone(t, c.wantLocal, c.anchor.Timezone)
			got := a.AlignToPeriodWithAnchor(bar, c.tf, c.anchor)
			if got != want {
				t.Errorf("bar %s (%s) / tf=%s: got %s, want %s (%s)",
					c.barLocal, c.anchor.Timezone, c.tf, fmtUnix(got), fmtUnix(want), c.wantLocal)
			}
			if got > bar {
				t.Errorf("bar %s: boundary %s is after the bar", c.barLocal, fmtUnix(got))
			}
		})
	}
}

// TestAlignToPeriodWithAnchor_IdempotencyAndMonotonicityWithAnchor verifies two
// universal properties of the alignment function under any anchor:
//
//  1. Idempotency: aligning an already-aligned boundary returns itself.
//  2. Monotonicity: the boundary never exceeds the input timestamp.
func TestAlignToPeriodWithAnchor_IdempotencyAndMonotonicityWithAnchor(t *testing.T) {
	a := NewTimeframeBoundaryAligner()

	anchors := []PeriodAnchor{
		{},
		{Timezone: "Europe/Moscow", SessionOpenMinute: 7 * 60},
		{Timezone: "America/New_York", SessionOpenMinute: 0},
	}
	tfs := []string{"5", "60", "240", "1h", "4h", "1D", "W", "M"}
	timestamps := []int64{
		mustUnix(t, "2025-08-15 09:37:00"),
		mustUnix(t, "2024-02-29 12:00:00"), // leap day
		mustUnix(t, "2023-12-31 23:59:59"), // year boundary
		mustUnix(t, "2025-01-01 00:00:00"), // year boundary exact
		0,
	}

	for _, anchor := range anchors {
		for _, tf := range tfs {
			for _, ts := range timestamps {
				first := a.AlignToPeriodWithAnchor(ts, tf, anchor)
				second := a.AlignToPeriodWithAnchor(first, tf, anchor)
				if first != second {
					t.Errorf("anchor{%s,open=%d} tf=%s ts=%s: not idempotent (%s → %s)",
						anchor.Timezone, anchor.SessionOpenMinute, tf,
						fmtUnix(ts), fmtUnix(first), fmtUnix(second))
				}
				if first > ts {
					t.Errorf("anchor{%s,open=%d} tf=%s ts=%s: boundary %s exceeds input",
						anchor.Timezone, anchor.SessionOpenMinute, tf,
						fmtUnix(ts), fmtUnix(first))
				}
			}
		}
	}
}

// TestAlignToPeriodWithAnchor_GateContractWithSessionAnchor verifies the
// change(time(tf)) gate: the boundary value must change exactly when entering
// a new period and remain constant within a period.  The first bar produces no
// change (mirrors Pine's change() returning NaN on bar 0).
func TestAlignToPeriodWithAnchor_GateContractWithSessionAnchor(t *testing.T) {
	a := NewTimeframeBoundaryAligner()

	type sequence struct {
		name       string
		anchor     PeriodAnchor
		tf         string
		barsLocal  []string
		wantChange []bool
	}

	cases := []sequence{
		{
			// MOEX 4h: windows at 07,11,15,19 MSK
			name:   "moex_4h_session_windows",
			anchor: PeriodAnchor{Timezone: "Europe/Moscow", SessionOpenMinute: 7 * 60},
			tf:     "240",
			barsLocal: []string{
				"2025-08-15 07:00:00", // bar 0: no prior
				"2025-08-15 08:00:00", // same [07:00,11:00) window
				"2025-08-15 10:59:00", // same [07:00,11:00) window
				"2025-08-15 11:00:00", // enters [11:00,15:00)
				"2025-08-15 13:30:00", // same [11:00,15:00)
				"2025-08-15 15:00:00", // enters [15:00,19:00)
				"2025-08-15 19:00:00", // enters [19:00,23:00)
			},
			wantChange: []bool{false, false, false, true, false, true, true},
		},
		{
			// Daily boundary in MSK: transition at MSK midnight = 21:00 UTC
			name:   "msk_daily_midnight_transition",
			anchor: PeriodAnchor{Timezone: "Europe/Moscow"},
			tf:     "1D",
			barsLocal: []string{
				"2025-08-14 07:00:00", // bar 0: Aug 14
				"2025-08-14 20:00:00", // same Aug 14 MSK day
				"2025-08-15 07:00:00", // enters Aug 15 MSK day
				"2025-08-15 22:00:00", // same Aug 15 MSK day
			},
			wantChange: []bool{false, false, true, false},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if len(c.barsLocal) != len(c.wantChange) {
				t.Fatalf("barsLocal and wantChange must have equal length")
			}
			prev := int64(-1)
			for i, barLocal := range c.barsLocal {
				bar := mustUnixInZone(t, barLocal, c.anchor.Timezone)
				cur := a.AlignToPeriodWithAnchor(bar, c.tf, c.anchor)
				changed := prev >= 0 && cur != prev
				if changed != c.wantChange[i] {
					t.Errorf("bar[%d] %s: change=%v, want %v (boundary=%s, prev=%s)",
						i, barLocal, changed, c.wantChange[i], fmtUnix(cur), fmtUnix(prev))
				}
				prev = cur
			}
		})
	}
}

// TestAlignToPeriodWithAnchor_PreSessionBarMapsToCurrentOrPriorPeriod verifies
// that a bar timestamped before the day's session open is placed in the last
// period of the prior session window, with no panic.  The calendar filter
// already removes pre-session bars in production, so this only guards the
// robustness of the boundary computation itself.
func TestAlignToPeriodWithAnchor_PreSessionBarMapsToCurrentOrPriorPeriod(t *testing.T) {
	a := NewTimeframeBoundaryAligner()

	type edgeCase struct {
		name      string
		anchor    PeriodAnchor
		tf        string
		barLocal  string
		wantLocal string
	}

	cases := []edgeCase{
		{
			// 01:00 MSK falls in the [23:00 Aug 14, 03:00 Aug 15) window
			"moex_4h_01h00_to_prior_23h00",
			PeriodAnchor{Timezone: "Europe/Moscow", SessionOpenMinute: 7 * 60},
			"240",
			"2025-08-15 01:00:00",
			"2025-08-14 23:00:00",
		},
		{
			// 05:00 MSK falls in the [03:00, 07:00) window (period before session open)
			"moex_4h_05h00_to_03h00",
			PeriodAnchor{Timezone: "Europe/Moscow", SessionOpenMinute: 7 * 60},
			"240",
			"2025-08-15 05:00:00",
			"2025-08-15 03:00:00",
		},
		{
			// Exactly at session open — maps to the first period of that session
			"moex_4h_at_session_open",
			PeriodAnchor{Timezone: "Europe/Moscow", SessionOpenMinute: 7 * 60},
			"240",
			"2025-08-15 07:00:00",
			"2025-08-15 07:00:00",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			bar := mustUnixInZone(t, c.barLocal, c.anchor.Timezone)
			want := mustUnixInZone(t, c.wantLocal, c.anchor.Timezone)
			got := a.AlignToPeriodWithAnchor(bar, c.tf, c.anchor)
			if got != want {
				t.Errorf("bar %s: got %s, want %s",
					c.barLocal, fmtUnix(got), fmtUnix(want))
			}
		})
	}
}

// TestAlignTimestampToPeriodWithAnchor_DelegatesCorrectly verifies that the
// package-level convenience function delegates to the aligner method and
// produces correct values for both zero and non-zero anchors.
func TestAlignTimestampToPeriodWithAnchor_DelegatesCorrectly(t *testing.T) {
	ts := mustUnix(t, "2025-08-15 09:37:00")

	t.Run("zero_anchor_matches_legacy", func(t *testing.T) {
		want := AlignTimestampToPeriod(ts, "60")
		got := AlignTimestampToPeriodWithAnchor(ts, "60", PeriodAnchor{})
		if got != want {
			t.Errorf("zero-anchor result %s differs from legacy %s", fmtUnix(got), fmtUnix(want))
		}
	})

	t.Run("non_zero_anchor_reflects_session", func(t *testing.T) {
		anchor := PeriodAnchor{Timezone: "Europe/Moscow", SessionOpenMinute: 7 * 60}
		// bar at 09:37 MSK is in the 07:00 session window
		got := AlignTimestampToPeriodWithAnchor(
			mustUnixInZone(t, "2025-08-15 09:37:00", "Europe/Moscow"),
			"240",
			anchor,
		)
		want := mustUnixInZone(t, "2025-08-15 07:00:00", "Europe/Moscow")
		if got != want {
			t.Errorf("session-anchored 4h at 09:37 MSK: got %s, want 07:00 MSK (%s)",
				fmtUnix(got), fmtUnix(want))
		}
	})
}
