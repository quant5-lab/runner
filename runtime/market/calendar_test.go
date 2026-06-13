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

func TestRegularSessionCalendar_NonIntradayTimeframesPassThroughWindow(t *testing.T) {
	window := NewSessionWindow(ClockTimeAt(7, 0), ClockTimeAt(23, 50))
	calWithWindow := NewRegularSessionCalendarWithWindow(
		NewWeekdaySet(time.Saturday, time.Sunday), nil, nil, window,
	)
	calMOEXStyle := NewRegularSessionCalendarWithWindow(
		WeekdaySet{}, nil, nil, window,
	)
	calNoWindow := NewRegularWeekdayCalendar(
		NewWeekdaySet(time.Saturday, time.Sunday), nil, nil,
	)

	preSession := context.OHLCV{Time: unixInMoscow(t, "2025-08-15 06:00")}
	inSession := context.OHLCV{Time: unixInMoscow(t, "2025-08-15 13:00")}
	satPreSession := context.OHLCV{Time: unixInMoscow(t, "2025-08-16 06:00")}
	midnight := context.OHLCV{Time: unixInMoscow(t, "2025-08-15 00:00")}

	cases := []struct {
		name      string
		calendar  RegularSessionCalendar
		bar       context.OHLCV
		timeframe string
		want      bool
	}{
		{name: "intraday-1h/before window start rejected", calendar: calWithWindow, bar: preSession, timeframe: "1h", want: false},
		{name: "intraday-1h/within window accepted", calendar: calWithWindow, bar: inSession, timeframe: "1h", want: true},
		{name: "intraday-30/before start rejected", calendar: calWithWindow, bar: preSession, timeframe: "30", want: false},
		{name: "intraday-15/within window accepted", calendar: calWithWindow, bar: inSession, timeframe: "15", want: true},

		// calWithWindow has Sat/Sun closed; preSession/midnight are Friday → weekday open.
		{name: "daily-1D/pre-session weekday accepted", calendar: calWithWindow, bar: preSession, timeframe: "1D", want: true},
		{name: "daily-1D/midnight weekday accepted", calendar: calWithWindow, bar: midnight, timeframe: "1D", want: true},
		{name: "daily-D/pre-session accepted", calendar: calMOEXStyle, bar: preSession, timeframe: "D", want: true},

		{name: "weekly-1W/pre-session accepted", calendar: calMOEXStyle, bar: preSession, timeframe: "1W", want: true},
		{name: "weekly-1W/midnight accepted", calendar: calMOEXStyle, bar: midnight, timeframe: "1W", want: true},
		{name: "monthly-1M/pre-session accepted", calendar: calMOEXStyle, bar: preSession, timeframe: "1M", want: true},
		{name: "monthly-1mo/midnight accepted", calendar: calMOEXStyle, bar: midnight, timeframe: "1mo", want: true},

		// Weekday rejection is NOT timeframe-gated — closed weekdays apply to all timeframes.
		{name: "daily/closed-weekday still rejected", calendar: calWithWindow, bar: satPreSession, timeframe: "1D", want: false},
		{name: "weekly/closed-weekday still rejected", calendar: calWithWindow, bar: satPreSession, timeframe: "1W", want: false},
		{name: "intraday/closed-weekday rejected", calendar: calWithWindow, bar: satPreSession, timeframe: "1h", want: false},

		{name: "no-window/intraday pre-session accepted", calendar: calNoWindow, bar: preSession, timeframe: "1h", want: true},
		{name: "no-window/daily midnight accepted", calendar: calNoWindow, bar: midnight, timeframe: "1D", want: true},
		{name: "no-window/closed-weekday rejected regardless", calendar: calNoWindow, bar: satPreSession, timeframe: "1D", want: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.calendar.Accepts(tc.bar, tc.timeframe, "Europe/Moscow"); got != tc.want {
				t.Fatalf("Accepts(timeframe=%q) = %v, want %v", tc.timeframe, got, tc.want)
			}
		})
	}
}

func TestRegularSessionCalendar_SessionWindowContract(t *testing.T) {
	window := NewSessionWindow(ClockTimeAt(7, 0), ClockTimeAt(23, 50))
	noWindow := SessionWindow{}

	cases := []struct {
		name     string
		calendar RegularSessionCalendar
		time     string
		want     bool
	}{
		{name: "unbounded/pre-dawn weekday accepted", calendar: NewRegularWeekdayCalendar(NewWeekdaySet(time.Saturday, time.Sunday), nil, nil), time: "2025-08-15 06:00", want: true},
		{name: "unbounded/midnight weekday accepted", calendar: NewRegularWeekdayCalendar(NewWeekdaySet(time.Saturday, time.Sunday), nil, nil), time: "2025-08-15 00:00", want: true},

		{name: "window/bar before start rejected", calendar: NewRegularSessionCalendarWithWindow(NewWeekdaySet(time.Saturday, time.Sunday), nil, nil, window), time: "2025-08-15 06:00", want: false},
		{name: "window/bar at start accepted", calendar: NewRegularSessionCalendarWithWindow(NewWeekdaySet(time.Saturday, time.Sunday), nil, nil, window), time: "2025-08-15 07:00", want: true},
		{name: "window/midday bar accepted", calendar: NewRegularSessionCalendarWithWindow(NewWeekdaySet(time.Saturday, time.Sunday), nil, nil, window), time: "2025-08-15 13:00", want: true},
		{name: "window/bar at end rejected", calendar: NewRegularSessionCalendarWithWindow(NewWeekdaySet(time.Saturday, time.Sunday), nil, nil, window), time: "2025-08-15 23:50", want: false},

		{name: "window/closed weekday outside window rejected", calendar: NewRegularSessionCalendarWithWindow(NewWeekdaySet(time.Saturday, time.Sunday), nil, nil, window), time: "2025-08-16 13:00", want: false},
		{name: "no-window/closed weekday still rejected", calendar: NewRegularWeekdayCalendar(NewWeekdaySet(time.Saturday, time.Sunday), nil, nil), time: "2025-08-16 13:00", want: false},

		{name: "special-closed/inside window rejected", calendar: NewRegularSessionCalendarWithWindow(NewWeekdaySet(time.Saturday, time.Sunday), nil, NewDateSet("2025-08-15"), window), time: "2025-08-15 13:00", want: false},
		{name: "special-closed/no-window rejected", calendar: NewRegularWeekdayCalendar(NewWeekdaySet(time.Saturday, time.Sunday), nil, NewDateSet("2025-08-15")), time: "2025-08-15 13:00", want: false},

		{name: "special-open/inside window accepted", calendar: NewRegularSessionCalendarWithWindow(NewWeekdaySet(time.Saturday, time.Sunday), NewDateSet("2025-08-16"), nil, window), time: "2025-08-16 13:00", want: true},
		{name: "special-open/before window rejected", calendar: NewRegularSessionCalendarWithWindow(NewWeekdaySet(time.Saturday, time.Sunday), NewDateSet("2025-08-16"), nil, window), time: "2025-08-16 06:00", want: false},
		{name: "special-open/no-window any hour accepted", calendar: NewRegularWeekdayCalendar(NewWeekdaySet(time.Saturday, time.Sunday), NewDateSet("2025-08-16"), nil), time: "2025-08-16 06:00", want: true},

		{name: "closed-beats-open/inside window rejected", calendar: NewRegularSessionCalendarWithWindow(NewWeekdaySet(time.Saturday, time.Sunday), NewDateSet("2025-08-18"), NewDateSet("2025-08-18"), noWindow), time: "2025-08-18 13:00", want: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			bar := context.OHLCV{Time: unixInMoscow(t, tc.time)}
			if got := tc.calendar.Accepts(bar, "1h", "Europe/Moscow"); got != tc.want {
				t.Fatalf("Accepts() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestRegularSessionCalendar_DateSessionWindowPrecedence(t *testing.T) {
	baseSchedule := WeekSchedule{
		Weekday: NewSessionWindow(ClockTimeAt(7, 0), ClockTimeAt(23, 50)),
		Weekend: NewSessionWindow(ClockTimeAt(10, 0), ClockTimeAt(19, 0)),
	}
	dateWindows := DateSessionWindows{
		"2025-08-16": NewSessionWindow(ClockTimeAt(12, 0), ClockTimeAt(15, 0)),
	}
	calendar := NewRegularSessionCalendar(nil, NewDateSet("2025-08-16"), nil, baseSchedule, dateWindows)
	closedCalendar := NewRegularSessionCalendar(nil, NewDateSet("2025-08-16"), NewDateSet("2025-08-16"), baseSchedule, dateWindows)
	weekdayClosedCalendar := NewRegularSessionCalendar(NewWeekdaySet(time.Saturday), nil, nil, baseSchedule, dateWindows)

	cases := []struct {
		name      string
		calendar  RegularSessionCalendar
		timeframe string
		datetime  string
		want      bool
	}{
		{name: "intraday before override start", calendar: calendar, timeframe: "1h", datetime: "2025-08-16 11:59", want: false},
		{name: "intraday at override start", calendar: calendar, timeframe: "1h", datetime: "2025-08-16 12:00", want: true},
		{name: "intraday before override end", calendar: calendar, timeframe: "1h", datetime: "2025-08-16 14:59", want: true},
		{name: "intraday at override end", calendar: calendar, timeframe: "1h", datetime: "2025-08-16 15:00", want: false},
		{name: "adjacent date uses regular weekend schedule", calendar: calendar, timeframe: "1h", datetime: "2025-08-17 10:00", want: true},
		{name: "daily bypasses date window", calendar: calendar, timeframe: "1D", datetime: "2025-08-16 20:00", want: true},
		{name: "special closed wins", calendar: closedCalendar, timeframe: "1h", datetime: "2025-08-16 12:00", want: false},
		{name: "closed weekday wins without special open", calendar: weekdayClosedCalendar, timeframe: "1h", datetime: "2025-08-16 12:00", want: false},
	}

	seen := map[string]bool{}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if seen[tc.name] {
				t.Fatalf("duplicate case name %q", tc.name)
			}
			seen[tc.name] = true
			bar := context.OHLCV{Time: unixInMoscow(t, tc.datetime)}
			if got := tc.calendar.Accepts(bar, tc.timeframe, "Europe/Moscow"); got != tc.want {
				t.Fatalf("Accepts(%s, %s) = %v, want %v", tc.datetime, tc.timeframe, got, tc.want)
			}
		})
	}
}
