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

func TestWeekSchedule_IsUnbounded(t *testing.T) {
	cases := []struct {
		name     string
		schedule WeekSchedule
		want     bool
	}{
		{name: "zero value", schedule: WeekSchedule{}, want: true},
		{name: "uniform with unbounded window", schedule: UniformWeekSchedule(SessionWindow{}), want: true},
		{name: "uniform bounded weekday", schedule: UniformWeekSchedule(NewSessionWindow(ClockTimeAt(7, 0), ClockTimeAt(23, 50))), want: false},
		{name: "weekday window only", schedule: WeekSchedule{Weekday: NewSessionWindow(ClockTimeAt(7, 0), ClockTimeAt(23, 50))}, want: false},
		{name: "separate weekday and weekend", schedule: WeekSchedule{Weekday: NewSessionWindow(ClockTimeAt(7, 0), ClockTimeAt(23, 50)), Weekend: NewSessionWindow(ClockTimeAt(10, 0), ClockTimeAt(19, 0))}, want: false},
		{name: "unbounded weekday with bounded weekend", schedule: WeekSchedule{Weekend: NewSessionWindow(ClockTimeAt(10, 0), ClockTimeAt(19, 0))}, want: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.schedule.IsUnbounded(); got != tc.want {
				t.Fatalf("IsUnbounded() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestWeekSchedule_Contains(t *testing.T) {
	utcAt := func(datetime string) time.Time {
		ts, _ := time.Parse("2006-01-02 15:04", datetime)
		return ts
	}

	var (
		zero          = WeekSchedule{}
		uniform       = UniformWeekSchedule(NewSessionWindow(ClockTimeAt(7, 0), ClockTimeAt(23, 50)))
		narrowWeekend = WeekSchedule{
			Weekday: NewSessionWindow(ClockTimeAt(7, 0), ClockTimeAt(23, 50)),
			Weekend: NewSessionWindow(ClockTimeAt(10, 0), ClockTimeAt(19, 0)),
		}
		widerWeekend = WeekSchedule{
			Weekday: NewSessionWindow(ClockTimeAt(9, 0), ClockTimeAt(17, 0)),
			Weekend: NewSessionWindow(ClockTimeAt(7, 0), ClockTimeAt(23, 0)),
		}
		unboundedWeekday = WeekSchedule{
			Weekend: NewSessionWindow(ClockTimeAt(10, 0), ClockTimeAt(19, 0)),
		}
	)

	cases := []struct {
		name     string
		schedule WeekSchedule
		instant  time.Time
		want     bool
	}{
		{name: "zero/weekday early", schedule: zero, instant: utcAt("2025-08-15 00:00"), want: true},
		{name: "zero/weekend early", schedule: zero, instant: utcAt("2025-08-16 01:00"), want: true},

		{name: "uniform/monday at start", schedule: uniform, instant: utcAt("2025-08-11 07:00"), want: true},
		{name: "uniform/monday before start", schedule: uniform, instant: utcAt("2025-08-11 06:59"), want: false},
		{name: "uniform/wednesday midday", schedule: uniform, instant: utcAt("2025-08-13 12:00"), want: true},
		{name: "uniform/friday one before end", schedule: uniform, instant: utcAt("2025-08-15 23:49"), want: true},
		{name: "uniform/friday at end", schedule: uniform, instant: utcAt("2025-08-15 23:50"), want: false},
		{name: "uniform/saturday at start", schedule: uniform, instant: utcAt("2025-08-16 07:00"), want: true},
		{name: "uniform/saturday before start", schedule: uniform, instant: utcAt("2025-08-16 06:59"), want: false},
		{name: "uniform/sunday in window", schedule: uniform, instant: utcAt("2025-08-17 13:00"), want: true},
		{name: "uniform/sunday at end", schedule: uniform, instant: utcAt("2025-08-17 23:50"), want: false},

		{name: "narrow-weekend/monday at weekday start", schedule: narrowWeekend, instant: utcAt("2025-08-11 07:00"), want: true},
		{name: "narrow-weekend/tuesday before weekday start", schedule: narrowWeekend, instant: utcAt("2025-08-12 06:59"), want: false},
		{name: "narrow-weekend/thursday midday", schedule: narrowWeekend, instant: utcAt("2025-08-14 13:00"), want: true},
		{name: "narrow-weekend/friday one before weekday end", schedule: narrowWeekend, instant: utcAt("2025-08-15 23:49"), want: true},
		{name: "narrow-weekend/friday at weekday end", schedule: narrowWeekend, instant: utcAt("2025-08-15 23:50"), want: false},
		{name: "narrow-weekend/saturday before weekend start", schedule: narrowWeekend, instant: utcAt("2025-08-16 09:59"), want: false},
		{name: "narrow-weekend/saturday at weekend start", schedule: narrowWeekend, instant: utcAt("2025-08-16 10:00"), want: true},
		{name: "narrow-weekend/saturday one before weekend end", schedule: narrowWeekend, instant: utcAt("2025-08-16 18:59"), want: true},
		{name: "narrow-weekend/saturday at weekend end", schedule: narrowWeekend, instant: utcAt("2025-08-16 19:00"), want: false},
		{name: "narrow-weekend/saturday post-session", schedule: narrowWeekend, instant: utcAt("2025-08-16 22:00"), want: false},
		{name: "narrow-weekend/sunday before weekend start", schedule: narrowWeekend, instant: utcAt("2025-08-17 09:00"), want: false},
		{name: "narrow-weekend/sunday at weekend start", schedule: narrowWeekend, instant: utcAt("2025-08-17 10:00"), want: true},
		{name: "narrow-weekend/sunday one before weekend end", schedule: narrowWeekend, instant: utcAt("2025-08-17 18:59"), want: true},
		{name: "narrow-weekend/sunday at weekend end", schedule: narrowWeekend, instant: utcAt("2025-08-17 19:00"), want: false},

		{name: "wider-weekend/weekday before start", schedule: widerWeekend, instant: utcAt("2025-08-13 08:59"), want: false},
		{name: "wider-weekend/weekday at start", schedule: widerWeekend, instant: utcAt("2025-08-13 09:00"), want: true},
		{name: "wider-weekend/weekday at end", schedule: widerWeekend, instant: utcAt("2025-08-13 17:00"), want: false},
		{name: "wider-weekend/saturday before wider-start", schedule: widerWeekend, instant: utcAt("2025-08-16 06:59"), want: false},
		{name: "wider-weekend/saturday at wider-start", schedule: widerWeekend, instant: utcAt("2025-08-16 07:00"), want: true},
		{name: "wider-weekend/saturday inside weekday-restricted zone", schedule: widerWeekend, instant: utcAt("2025-08-16 08:00"), want: true},
		{name: "wider-weekend/saturday one before wider-end", schedule: widerWeekend, instant: utcAt("2025-08-16 22:59"), want: true},
		{name: "wider-weekend/saturday at wider-end", schedule: widerWeekend, instant: utcAt("2025-08-16 23:00"), want: false},

		{name: "unbounded-weekday/monday early morning passes", schedule: unboundedWeekday, instant: utcAt("2025-08-11 01:00"), want: true},
		{name: "unbounded-weekday/wednesday midnight passes", schedule: unboundedWeekday, instant: utcAt("2025-08-13 00:00"), want: true},
		{name: "unbounded-weekday/friday late evening passes", schedule: unboundedWeekday, instant: utcAt("2025-08-15 23:59"), want: true},
		{name: "unbounded-weekday/saturday before weekend start", schedule: unboundedWeekday, instant: utcAt("2025-08-16 09:59"), want: false},
		{name: "unbounded-weekday/saturday at weekend start", schedule: unboundedWeekday, instant: utcAt("2025-08-16 10:00"), want: true},
		{name: "unbounded-weekday/saturday midday", schedule: unboundedWeekday, instant: utcAt("2025-08-16 14:00"), want: true},
		{name: "unbounded-weekday/saturday one before weekend end", schedule: unboundedWeekday, instant: utcAt("2025-08-16 18:59"), want: true},
		{name: "unbounded-weekday/saturday at weekend end", schedule: unboundedWeekday, instant: utcAt("2025-08-16 19:00"), want: false},
		{name: "unbounded-weekday/sunday before weekend start", schedule: unboundedWeekday, instant: utcAt("2025-08-17 09:00"), want: false},
		{name: "unbounded-weekday/sunday in session", schedule: unboundedWeekday, instant: utcAt("2025-08-17 12:00"), want: true},
		{name: "unbounded-weekday/sunday at weekend end", schedule: unboundedWeekday, instant: utcAt("2025-08-17 19:00"), want: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.schedule.Contains(tc.instant); got != tc.want {
				t.Fatalf("Contains(%s) = %v, want %v", tc.instant.Format("2006-01-02 15:04 Mon"), got, tc.want)
			}
		})
	}
}

func TestWeekSchedule_ContainsFor(t *testing.T) {
	utcAt := func(datetime string) time.Time {
		ts, _ := time.Parse("2006-01-02 15:04", datetime)
		return ts
	}

	var (
		zero    = WeekSchedule{}
		uniform = UniformWeekSchedule(NewSessionWindow(ClockTimeAt(7, 0), ClockTimeAt(23, 50)))
		split   = WeekSchedule{
			Weekday: NewSessionWindow(ClockTimeAt(7, 0), ClockTimeAt(23, 50)),
			Weekend: NewSessionWindow(ClockTimeAt(10, 0), ClockTimeAt(19, 0)),
		}
	)

	cases := []struct {
		name           string
		schedule       WeekSchedule
		instant        time.Time
		treatAsWeekend bool
		want           bool
	}{
		{name: "zero/saturday/as-weekday", schedule: zero, instant: utcAt("2025-08-16 03:00"), treatAsWeekend: false, want: true},
		{name: "zero/saturday/as-weekend", schedule: zero, instant: utcAt("2025-08-16 03:00"), treatAsWeekend: true, want: true},
		{name: "zero/monday/as-weekday", schedule: zero, instant: utcAt("2025-08-11 03:00"), treatAsWeekend: false, want: true},
		{name: "zero/monday/as-weekend", schedule: zero, instant: utcAt("2025-08-11 03:00"), treatAsWeekend: true, want: true},
		{name: "uniform/saturday/as-weekday/in-window", schedule: uniform, instant: utcAt("2025-08-16 12:00"), treatAsWeekend: false, want: true},
		{name: "uniform/saturday/as-weekend/in-window", schedule: uniform, instant: utcAt("2025-08-16 12:00"), treatAsWeekend: true, want: true},
		{name: "uniform/saturday/as-weekday/before-start", schedule: uniform, instant: utcAt("2025-08-16 06:59"), treatAsWeekend: false, want: false},
		{name: "uniform/saturday/as-weekend/before-start", schedule: uniform, instant: utcAt("2025-08-16 06:59"), treatAsWeekend: true, want: false},
		{name: "uniform/monday/as-weekday/in-window", schedule: uniform, instant: utcAt("2025-08-11 12:00"), treatAsWeekend: false, want: true},
		{name: "uniform/monday/as-weekend/in-window", schedule: uniform, instant: utcAt("2025-08-11 12:00"), treatAsWeekend: true, want: true},
		{name: "split/saturday/as-weekday/below-weekday-start", schedule: split, instant: utcAt("2025-08-16 06:59"), treatAsWeekend: false, want: false},
		{name: "split/saturday/as-weekday/at-weekday-start", schedule: split, instant: utcAt("2025-08-16 07:00"), treatAsWeekend: false, want: true},
		{name: "split/saturday/as-weekday/below-weekend-start", schedule: split, instant: utcAt("2025-08-16 09:00"), treatAsWeekend: false, want: true},
		{name: "split/saturday/as-weekday/midday", schedule: split, instant: utcAt("2025-08-16 14:00"), treatAsWeekend: false, want: true},
		{name: "split/saturday/as-weekday/past-weekend-end", schedule: split, instant: utcAt("2025-08-16 21:00"), treatAsWeekend: false, want: true},
		{name: "split/saturday/as-weekday/before-weekday-end", schedule: split, instant: utcAt("2025-08-16 23:49"), treatAsWeekend: false, want: true},
		{name: "split/saturday/as-weekday/at-weekday-end", schedule: split, instant: utcAt("2025-08-16 23:50"), treatAsWeekend: false, want: false},
		{name: "split/sunday/as-weekday/below-weekend-start", schedule: split, instant: utcAt("2025-08-17 09:00"), treatAsWeekend: false, want: true},
		{name: "split/sunday/as-weekday/past-weekend-end", schedule: split, instant: utcAt("2025-08-17 20:00"), treatAsWeekend: false, want: true},
		{name: "split/sunday/as-weekday/at-weekday-end", schedule: split, instant: utcAt("2025-08-17 23:50"), treatAsWeekend: false, want: false},
		{name: "split/monday/as-weekend/below-weekend-start", schedule: split, instant: utcAt("2025-08-11 09:59"), treatAsWeekend: true, want: false},
		{name: "split/monday/as-weekend/at-weekend-start", schedule: split, instant: utcAt("2025-08-11 10:00"), treatAsWeekend: true, want: true},
		{name: "split/monday/as-weekend/midday", schedule: split, instant: utcAt("2025-08-11 14:00"), treatAsWeekend: true, want: true},
		{name: "split/monday/as-weekend/past-weekend-end", schedule: split, instant: utcAt("2025-08-11 21:00"), treatAsWeekend: true, want: false},
		{name: "split/monday/as-weekend/in-weekday-zone-before-weekend-start", schedule: split, instant: utcAt("2025-08-11 07:00"), treatAsWeekend: true, want: false},
		{name: "split/friday/as-weekend/at-weekend-start", schedule: split, instant: utcAt("2025-08-15 10:00"), treatAsWeekend: true, want: true},
		{name: "split/friday/as-weekend/past-weekend-end", schedule: split, instant: utcAt("2025-08-15 23:49"), treatAsWeekend: true, want: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.schedule.ContainsFor(tc.instant, tc.treatAsWeekend); got != tc.want {
				t.Fatalf("ContainsFor(%s, treatAsWeekend=%v) = %v, want %v",
					tc.instant.Format("2006-01-02 15:04 Mon"), tc.treatAsWeekend, got, tc.want)
			}
		})
	}
}

func TestNewDateSessionWindows(t *testing.T) {
	cases := []struct {
		name    string
		values  map[string]string
		date    string
		instant string
		want    bool
		wantErr bool
	}{
		{name: "empty", values: nil, date: "2025-08-16", instant: "2025-08-16 12:00", want: false},
		{name: "valid", values: map[string]string{"2025-08-16": "1200-1500"}, date: "2025-08-16", instant: "2025-08-16 12:00", want: true},
		{name: "outside", values: map[string]string{"2025-08-16": "1200-1500"}, date: "2025-08-16", instant: "2025-08-16 15:00", want: false},
		{name: "missing date", values: map[string]string{"2025-08-16": "1200-1500"}, date: "2025-08-17", instant: "2025-08-17 12:00", want: false},
		{name: "bad date", values: map[string]string{"2025/08/16": "1200-1500"}, wantErr: true},
		{name: "bad window", values: map[string]string{"2025-08-16": "bad"}, wantErr: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			windows, err := NewDateSessionWindows(tc.values)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			window, ok := windows.Window(tc.date)
			if !ok {
				if tc.want {
					t.Fatalf("window %s missing", tc.date)
				}
				return
			}
			instant, err := time.Parse("2006-01-02 15:04", tc.instant)
			if err != nil {
				t.Fatalf("parse instant: %v", err)
			}
			if got := window.Contains(instant); got != tc.want {
				t.Fatalf("Contains(%s) = %v, want %v", tc.instant, got, tc.want)
			}
		})
	}
}
