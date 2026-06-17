package context

import (
	"testing"
)

// TestBarOpenTimeAtTimeframe_Delegation verifies that BarOpenTimeAtTimeframe
// produces the correct period-start timestamp for every supported timeframe
// class (intraday numeric, intraday suffixed, daily, weekly, monthly).
func TestBarOpenTimeAtTimeframe_Delegation(t *testing.T) {
	tests := []struct {
		name      string
		inputStr  string
		timeframe string
		wantStr   string
	}{
		// Intraday — numeric Pine tokens (bare minute counts)
		{"60min_numeric", "2025-08-15 09:37:00", "60", "2025-08-15 09:00:00"},
		{"240min_numeric", "2025-08-15 09:37:00", "240", "2025-08-15 08:00:00"},
		{"5min_numeric", "2025-08-15 09:37:00", "5", "2025-08-15 09:35:00"},
		{"15min_numeric", "2025-08-15 09:37:00", "15", "2025-08-15 09:30:00"},
		// Intraday — suffixed tokens
		{"1h_suffixed", "2025-08-15 09:37:00", "1h", "2025-08-15 09:00:00"},
		{"4h_suffixed", "2025-08-15 09:37:00", "4h", "2025-08-15 08:00:00"},
		{"5m_suffixed", "2025-08-15 09:37:00", "5m", "2025-08-15 09:35:00"},
		// Calendar-scale
		{"daily", "2025-08-15 14:30:00", "1D", "2025-08-15 00:00:00"},
		{"weekly_friday_aligns_to_monday", "2025-08-15 14:30:00", "W", "2025-08-11 00:00:00"},
		{"monthly", "2025-08-15 12:00:00", "M", "2025-08-01 00:00:00"},
		// Bar already on a period boundary — must be idempotent
		{"on_4h_boundary", "2025-08-15 08:00:00", "4h", "2025-08-15 08:00:00"},
		{"on_day_boundary", "2025-08-15 00:00:00", "1D", "2025-08-15 00:00:00"},
		// Unknown / empty timeframe — must return input unchanged
		{"empty_tf_passthrough", "2025-08-15 09:37:00", "", "2025-08-15 09:37:00"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := mustUnix(t, tt.inputStr)
			want := mustUnix(t, tt.wantStr)
			got := BarOpenTimeAtTimeframe(in, tt.timeframe)
			if got != want {
				t.Errorf("BarOpenTimeAtTimeframe(%s, %q) = %s, want %s",
					fmtUnix(in), tt.timeframe, fmtUnix(got), fmtUnix(want))
			}
		})
	}
}

// TestBarOpenTimeAtTimeframe_NumericAndSuffixedTokensAreEquivalent verifies
// that bare minute-count tokens and suffixed tokens that name the same period
// produce the same boundary — ensuring the choice of Pine token form is
// irrelevant to evaluation.
func TestBarOpenTimeAtTimeframe_NumericAndSuffixedTokensAreEquivalent(t *testing.T) {
	pairs := [][2]string{
		{"1h", "60"},
		{"4h", "240"},
		{"5m", "5"},
		{"15m", "15"},
		{"30m", "30"},
	}
	bar := mustUnix(t, "2025-08-15 09:37:00")
	for _, pair := range pairs {
		suffixed, numeric := pair[0], pair[1]
		a := BarOpenTimeAtTimeframe(bar, suffixed)
		b := BarOpenTimeAtTimeframe(bar, numeric)
		if a != b {
			t.Errorf("BarOpenTimeAtTimeframe(bar, %q)=%s != BarOpenTimeAtTimeframe(bar, %q)=%s — equivalent tokens must agree",
				suffixed, fmtUnix(a), numeric, fmtUnix(b))
		}
	}
}

// TestBarOpenTimeAtTimeframe_AllBarsInWindowReturnSameBoundary verifies the
// core invariant: every bar whose timestamp falls within the same period must
// return the identical boundary value, so Pine's time(tf) series is constant
// within a secondary window and changes only at the window start.
func TestBarOpenTimeAtTimeframe_AllBarsInWindowReturnSameBoundary(t *testing.T) {
	tests := []struct {
		name      string
		timeframe string
		boundary  string
		bars      []string
	}{
		{
			name:      "1h_bars_within_4h_window",
			timeframe: "240",
			boundary:  "2025-08-15 08:00:00",
			bars:      []string{"2025-08-15 08:00:00", "2025-08-15 09:00:00", "2025-08-15 10:00:00", "2025-08-15 11:00:00"},
		},
		{
			name:      "minute_bars_within_1h_window",
			timeframe: "60",
			boundary:  "2025-08-15 09:00:00",
			bars:      []string{"2025-08-15 09:00:00", "2025-08-15 09:15:00", "2025-08-15 09:30:00", "2025-08-15 09:45:00"},
		},
		{
			name:      "1h_bars_within_daily_window",
			timeframe: "1D",
			boundary:  "2025-08-15 00:00:00",
			bars:      []string{"2025-08-15 07:00:00", "2025-08-15 12:00:00", "2025-08-15 22:00:00"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			want := mustUnix(t, tt.boundary)
			for _, barStr := range tt.bars {
				bar := mustUnix(t, barStr)
				got := BarOpenTimeAtTimeframe(bar, tt.timeframe)
				if got != want {
					t.Errorf("[%s] bar at %s: BarOpenTimeAtTimeframe(..., %q) = %s, want %s",
						tt.name, barStr, tt.timeframe, fmtUnix(got), fmtUnix(want))
				}
			}
		})
	}
}

// TestBarOpenTimeAtTimeframe_GateContract verifies the semantic used by Pine's
// change(time(tf)) != 0 gate: the boundary value must change exactly when
// entering a new period and stay constant within a period. The first bar in a
// sequence has no "previous" so it never triggers a change (mirrors Pine's
// change() returning NaN on the first bar).
func TestBarOpenTimeAtTimeframe_GateContract(t *testing.T) {
	tests := []struct {
		name       string
		timeframe  string
		bars       []string
		wantChange []bool // true if boundary differs from previous; false for the first bar
	}{
		{
			name:      "4h_window_multi_bar",
			timeframe: "240",
			bars: []string{
				"2025-08-15 08:00:00",
				"2025-08-15 09:00:00",
				"2025-08-15 10:00:00",
				"2025-08-15 11:00:00",
				"2025-08-15 12:00:00",
				"2025-08-15 13:00:00",
				"2025-08-15 16:00:00",
			},
			wantChange: []bool{false, false, false, false, true, false, true},
		},
		{
			name:      "1h_every_bar_is_a_new_window",
			timeframe: "1h",
			bars: []string{
				"2025-08-15 09:00:00",
				"2025-08-15 10:00:00",
				"2025-08-15 11:00:00",
			},
			wantChange: []bool{false, true, true},
		},
		{
			name:      "daily_transition_across_midnight",
			timeframe: "1D",
			bars: []string{
				"2025-08-14 22:00:00",
				"2025-08-14 23:00:00",
				"2025-08-15 07:00:00",
				"2025-08-15 08:00:00",
			},
			wantChange: []bool{false, false, true, false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if len(tt.bars) != len(tt.wantChange) {
				t.Fatalf("bars and wantChange lengths must match: %d vs %d", len(tt.bars), len(tt.wantChange))
			}
			prev := int64(-1)
			for i, barStr := range tt.bars {
				bar := mustUnix(t, barStr)
				cur := BarOpenTimeAtTimeframe(bar, tt.timeframe)
				changed := prev >= 0 && cur != prev
				if changed != tt.wantChange[i] {
					t.Errorf("[%s] bar[%d] at %s: change=%v, want %v (boundary=%s, prev=%s)",
						tt.name, i, barStr, changed, tt.wantChange[i],
						fmtUnix(cur), fmtUnix(prev))
				}
				prev = cur
			}
		})
	}
}
