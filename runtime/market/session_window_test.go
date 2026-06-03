package market

import (
	"testing"
	"time"
)

func TestParseClockTime(t *testing.T) {
	cases := []struct {
		input   string
		hour    int
		minute  int
		wantErr bool
	}{
		{input: "0000", hour: 0, minute: 0},
		{input: "0700", hour: 7, minute: 0},
		{input: "0930", hour: 9, minute: 30},
		{input: "1200", hour: 12, minute: 0},
		{input: "2350", hour: 23, minute: 50},
		{input: "2359", hour: 23, minute: 59},
		{input: "", wantErr: true},
		{input: "07", wantErr: true},
		{input: "070", wantErr: true},
		{input: "07000", wantErr: true},
		{input: "2400", wantErr: true},
		{input: "0060", wantErr: true},
		{input: "ab00", wantErr: true},
		{input: "00zz", wantErr: true},
		{input: "-100", wantErr: true},
	}

	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			ct, err := ParseClockTime(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error for %q, got nil", tc.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if ct.Hour != tc.hour || ct.Minute != tc.minute {
				t.Fatalf("got {%d,%d}, want {%d,%d}", ct.Hour, ct.Minute, tc.hour, tc.minute)
			}
		})
	}
}

func TestParseSessionWindow(t *testing.T) {
	cases := []struct {
		input   string
		start   ClockTime
		end     ClockTime
		wantErr bool
	}{
		{input: "0700-2350", start: ClockTimeAt(7, 0), end: ClockTimeAt(23, 50)},
		{input: "0930-1630", start: ClockTimeAt(9, 30), end: ClockTimeAt(16, 30)},
		{input: "2200-0600", start: ClockTimeAt(22, 0), end: ClockTimeAt(6, 0)},
		{input: "0000-2359", start: ClockTimeAt(0, 0), end: ClockTimeAt(23, 59)},
		{input: "0000-0000", start: ClockTimeAt(0, 0), end: ClockTimeAt(0, 0)},
		{input: "0700-0700", start: ClockTimeAt(7, 0), end: ClockTimeAt(7, 0)},
		{input: "", wantErr: true},
		{input: "0700", wantErr: true},
		{input: "0700-", wantErr: true},
		{input: "-2350", wantErr: true},
		{input: "0700+2350", wantErr: true},
		{input: "07002350", wantErr: true},
		{input: "2400-2350", wantErr: true},
		{input: "0700-2360", wantErr: true},
	}

	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			w, err := ParseSessionWindow(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error for %q, got nil", tc.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if w.Start != tc.start || w.End != tc.end {
				t.Fatalf("got {%v,%v}, want {%v,%v}", w.Start, w.End, tc.start, tc.end)
			}
		})
	}
}

func TestSessionWindow_IsUnbounded(t *testing.T) {
	cases := []struct {
		name   string
		window SessionWindow
		want   bool
	}{
		{name: "zero value", window: SessionWindow{}, want: true},
		{name: "zero-zero explicit", window: NewSessionWindow(ClockTimeAt(0, 0), ClockTimeAt(0, 0)), want: true},
		{name: "normal window", window: NewSessionWindow(ClockTimeAt(7, 0), ClockTimeAt(23, 50)), want: false},
		{name: "overnight window", window: NewSessionWindow(ClockTimeAt(22, 0), ClockTimeAt(6, 0)), want: false},
		{name: "start only non-zero", window: SessionWindow{Start: ClockTimeAt(7, 0)}, want: false},
		{name: "end only non-zero", window: SessionWindow{End: ClockTimeAt(23, 50)}, want: false},
		{name: "full-day window", window: NewSessionWindow(ClockTimeAt(0, 0), ClockTimeAt(23, 59)), want: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.window.IsUnbounded(); got != tc.want {
				t.Fatalf("IsUnbounded() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestSessionWindow_Contains(t *testing.T) {
	cases := []struct {
		name   string
		window SessionWindow
		h, m   int
		want   bool
	}{
		{name: "unbounded/midnight", window: SessionWindow{}, h: 0, m: 0, want: true},
		{name: "unbounded/pre-dawn", window: SessionWindow{}, h: 6, m: 0, want: true},
		{name: "unbounded/midday", window: SessionWindow{}, h: 12, m: 0, want: true},
		{name: "unbounded/evening", window: SessionWindow{}, h: 23, m: 59, want: true},

		{name: "normal/before start", window: NewSessionWindow(ClockTimeAt(7, 0), ClockTimeAt(23, 50)), h: 6, m: 59, want: false},
		{name: "normal/at start", window: NewSessionWindow(ClockTimeAt(7, 0), ClockTimeAt(23, 50)), h: 7, m: 0, want: true},
		{name: "normal/one after start", window: NewSessionWindow(ClockTimeAt(7, 0), ClockTimeAt(23, 50)), h: 7, m: 1, want: true},
		{name: "normal/midday", window: NewSessionWindow(ClockTimeAt(7, 0), ClockTimeAt(23, 50)), h: 12, m: 0, want: true},
		{name: "normal/one before end", window: NewSessionWindow(ClockTimeAt(7, 0), ClockTimeAt(23, 50)), h: 23, m: 49, want: true},
		{name: "normal/at end", window: NewSessionWindow(ClockTimeAt(7, 0), ClockTimeAt(23, 50)), h: 23, m: 50, want: false},
		{name: "normal/after end", window: NewSessionWindow(ClockTimeAt(7, 0), ClockTimeAt(23, 50)), h: 23, m: 59, want: false},

		{name: "overnight/before start", window: NewSessionWindow(ClockTimeAt(22, 0), ClockTimeAt(6, 0)), h: 21, m: 59, want: false},
		{name: "overnight/at start", window: NewSessionWindow(ClockTimeAt(22, 0), ClockTimeAt(6, 0)), h: 22, m: 0, want: true},
		{name: "overnight/late evening", window: NewSessionWindow(ClockTimeAt(22, 0), ClockTimeAt(6, 0)), h: 23, m: 0, want: true},
		{name: "overnight/midnight", window: NewSessionWindow(ClockTimeAt(22, 0), ClockTimeAt(6, 0)), h: 0, m: 0, want: true},
		{name: "overnight/one before end", window: NewSessionWindow(ClockTimeAt(22, 0), ClockTimeAt(6, 0)), h: 5, m: 59, want: true},
		{name: "overnight/at end", window: NewSessionWindow(ClockTimeAt(22, 0), ClockTimeAt(6, 0)), h: 6, m: 0, want: false},
		{name: "overnight/midday excluded", window: NewSessionWindow(ClockTimeAt(22, 0), ClockTimeAt(6, 0)), h: 12, m: 0, want: false},

		{name: "degenerate-equal/early morning", window: NewSessionWindow(ClockTimeAt(7, 0), ClockTimeAt(7, 0)), h: 0, m: 0, want: true},
		{name: "degenerate-equal/at boundary", window: NewSessionWindow(ClockTimeAt(7, 0), ClockTimeAt(7, 0)), h: 7, m: 0, want: true},
		{name: "degenerate-equal/midday", window: NewSessionWindow(ClockTimeAt(7, 0), ClockTimeAt(7, 0)), h: 12, m: 0, want: true},

		{name: "minute/7:29 before window", window: NewSessionWindow(ClockTimeAt(7, 30), ClockTimeAt(16, 0)), h: 7, m: 29, want: false},
		{name: "minute/7:30 at start", window: NewSessionWindow(ClockTimeAt(7, 30), ClockTimeAt(16, 0)), h: 7, m: 30, want: true},
		{name: "minute/15:59 before end", window: NewSessionWindow(ClockTimeAt(7, 30), ClockTimeAt(16, 0)), h: 15, m: 59, want: true},
		{name: "minute/16:00 at end", window: NewSessionWindow(ClockTimeAt(7, 30), ClockTimeAt(16, 0)), h: 16, m: 0, want: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Use a fixed arbitrary date — the date does not affect hour/minute logic.
			instant := time.Date(2025, 8, 15, tc.h, tc.m, 0, 0, time.UTC)
			if got := tc.window.Contains(instant); got != tc.want {
				t.Fatalf("Contains(%02d:%02d) = %v, want %v", tc.h, tc.m, got, tc.want)
			}
		})
	}
}
