package market

import (
	"testing"
	"time"

	"github.com/quant5-lab/runner/runtime/context"
)

func unixInMoscow(t *testing.T, value string) int64 {
	t.Helper()
	location, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		t.Fatalf("load Moscow timezone: %v", err)
	}
	parsed, err := time.ParseInLocation("2006-01-02 15:04", value, location)
	if err != nil {
		t.Fatalf("parse time: %v", err)
	}
	return parsed.Unix()
}

func TestRegularSessionCalendar_AcceptsByClosedWeekdaysAndSpecialDates(t *testing.T) {
	calendar := NewRegularWeekdayCalendar(
		NewWeekdaySet(time.Saturday, time.Sunday),
		NewDateSet("2025-08-16"),
		NewDateSet("2025-08-18"),
	)

	cases := []struct {
		name string
		time string
		want bool
	}{
		{name: "weekday accepted", time: "2025-08-15 13:00", want: true},
		{name: "closed weekday rejected", time: "2025-08-17 13:00", want: false},
		{name: "special open overrides closed weekday", time: "2025-08-16 13:00", want: true},
		{name: "special close overrides open weekday", time: "2025-08-18 13:00", want: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			bar := context.OHLCV{Time: unixInMoscow(t, tc.time)}
			if got := calendar.Accepts(bar, "1h", "Europe/Moscow"); got != tc.want {
				t.Fatalf("Accepts() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestRegularSessionCalendar_TimeInputEdges(t *testing.T) {
	calendar := NewRegularWeekdayCalendar(NewWeekdaySet(time.Sunday), nil, nil)

	cases := []struct {
		name     string
		bar      context.OHLCV
		timezone string
		want     bool
	}{
		{
			name:     "invalid timezone falls back to UTC",
			bar:      context.OHLCV{Time: time.Date(2025, 8, 17, 13, 0, 0, 0, time.UTC).Unix()},
			timezone: "not-a-timezone",
			want:     false,
		},
		{
			name:     "unix millisecond timestamp normalizes before weekday check",
			bar:      context.OHLCV{Time: time.Date(2025, 8, 17, 13, 0, 0, 0, time.UTC).UnixMilli()},
			timezone: "UTC",
			want:     false,
		},
		{
			name:     "timezone shifts UTC instant into local weekday",
			bar:      context.OHLCV{Time: time.Date(2025, 8, 17, 22, 30, 0, 0, time.UTC).Unix()},
			timezone: "Europe/Moscow",
			want:     true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := calendar.Accepts(tc.bar, "1h", tc.timezone); got != tc.want {
				t.Fatalf("Accepts() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestCalendarSets_NormalizeInputs(t *testing.T) {
	if !NewDateSet(" 2025-08-16 ", "").Contains("2025-08-16") {
		t.Fatal("DateSet should trim dates and ignore empty entries")
	}

	weekdays := WeekdaySetFromNames(" Monday ", "2", "bad")
	if !weekdays.Contains(time.Monday) || !weekdays.Contains(time.Tuesday) {
		t.Fatalf("weekday names/numbers not parsed: %#v", weekdays)
	}
	if weekdays.Contains(time.Sunday) {
		t.Fatalf("invalid weekday entry should be ignored: %#v", weekdays)
	}
}

func TestExplicitUniverseCalendars(t *testing.T) {
	bars := []context.OHLCV{
		{Time: time.Date(2025, 8, 15, 13, 0, 0, 0, time.UTC).Unix(), Close: 1},
		{Time: time.Date(2025, 8, 16, 13, 0, 0, 0, time.UTC).Unix(), Close: 2},
		{Time: time.Date(2025, 8, 17, 13, 0, 0, 0, time.UTC).UnixMilli(), Close: 3},
	}

	exact := ExactTimestampCalendar{Allowed: NewTimestampSet(bars[1].Time, bars[2].Time)}
	if exact.Accepts(bars[0], "1h", "UTC") || !exact.Accepts(bars[1], "1h", "UTC") || !exact.Accepts(bars[2], "1h", "UTC") {
		t.Fatal("exact timestamp calendar must accept only listed instants")
	}

	dates := DateWhitelistCalendar{Allowed: NewDateSet("2025-08-16")}
	if dates.Accepts(bars[0], "1h", "UTC") || !dates.Accepts(bars[1], "1h", "UTC") {
		t.Fatal("date whitelist calendar must accept only listed local dates")
	}
}
