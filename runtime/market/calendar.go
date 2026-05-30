package market

import (
	"strings"
	"time"

	"github.com/quant5-lab/runner/runtime/context"
)

type Calendar interface {
	Accepts(bar context.OHLCV, timeframe string, timezone string) bool
}

type AlwaysOpenCalendar struct{}

func (AlwaysOpenCalendar) Accepts(context.OHLCV, string, string) bool { return true }

type DateSet map[string]struct{}

func NewDateSet(dates ...string) DateSet {
	set := make(DateSet, len(dates))
	for _, date := range dates {
		date = strings.TrimSpace(date)
		if date != "" {
			set[date] = struct{}{}
		}
	}
	return set
}

func (s DateSet) Contains(date string) bool {
	_, ok := s[date]
	return ok
}

func (s DateSet) Empty() bool { return len(s) == 0 }

type TimestampSet map[int64]struct{}

func NewTimestampSet(values ...int64) TimestampSet {
	set := make(TimestampSet, len(values))
	for _, value := range values {
		set[unixSecond(value)] = struct{}{}
	}
	return set
}

func (s TimestampSet) Contains(value int64) bool {
	_, ok := s[unixSecond(value)]
	return ok
}

func (s TimestampSet) Empty() bool { return len(s) == 0 }

type WeekdaySet map[time.Weekday]struct{}

func NewWeekdaySet(days ...time.Weekday) WeekdaySet {
	set := make(WeekdaySet, len(days))
	for _, day := range days {
		set[day] = struct{}{}
	}
	return set
}

func WeekdaySetFromNames(names ...string) WeekdaySet {
	set := WeekdaySet{}
	for _, name := range names {
		if day, ok := parseWeekday(name); ok {
			set[day] = struct{}{}
		}
	}
	return set
}

func (s WeekdaySet) Contains(day time.Weekday) bool {
	_, ok := s[day]
	return ok
}

func (s WeekdaySet) Empty() bool { return len(s) == 0 }

type ExactTimestampCalendar struct {
	Allowed TimestampSet
}

func (c ExactTimestampCalendar) Accepts(bar context.OHLCV, _ string, _ string) bool {
	return c.Allowed.Contains(bar.Time)
}

type DateWhitelistCalendar struct {
	Allowed DateSet
}

func (c DateWhitelistCalendar) Accepts(bar context.OHLCV, _ string, timezone string) bool {
	return c.Allowed.Contains(localDate(bar, timezone))
}

type RegularSessionCalendar struct {
	ClosedWeekdays WeekdaySet
	SpecialOpen    DateSet
	SpecialClosed  DateSet
}

func (c RegularSessionCalendar) Accepts(bar context.OHLCV, timeframe string, timezone string) bool {
	date := localDate(bar, timezone)
	if c.SpecialClosed.Contains(date) {
		return false
	}
	if c.SpecialOpen.Contains(date) {
		return true
	}
	return !c.ClosedWeekdays.Contains(barInstant(bar).In(locationOrUTC(timezone)).Weekday())
}

func NewRegularWeekdayCalendar(closed WeekdaySet, specialOpen DateSet, specialClosed DateSet) RegularSessionCalendar {
	if closed == nil {
		closed = WeekdaySet{}
	}
	if specialOpen == nil {
		specialOpen = DateSet{}
	}
	if specialClosed == nil {
		specialClosed = DateSet{}
	}
	return RegularSessionCalendar{
		ClosedWeekdays: closed,
		SpecialOpen:    specialOpen,
		SpecialClosed:  specialClosed,
	}
}

func localDate(bar context.OHLCV, timezone string) string {
	return barInstant(bar).In(locationOrUTC(timezone)).Format("2006-01-02")
}

func locationOrUTC(timezone string) *time.Location {
	location, err := time.LoadLocation(timezone)
	if err != nil {
		return time.UTC
	}
	return location
}

func barInstant(bar context.OHLCV) time.Time {
	return time.Unix(unixSecond(bar.Time), 0)
}

func unixSecond(value int64) int64 {
	if value > 10_000_000_000 {
		return value / 1000
	}
	return value
}

func parseWeekday(value string) (time.Weekday, bool) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "0", "sun", "sunday":
		return time.Sunday, true
	case "1", "mon", "monday":
		return time.Monday, true
	case "2", "tue", "tuesday":
		return time.Tuesday, true
	case "3", "wed", "wednesday":
		return time.Wednesday, true
	case "4", "thu", "thursday":
		return time.Thursday, true
	case "5", "fri", "friday":
		return time.Friday, true
	case "6", "sat", "saturday":
		return time.Saturday, true
	default:
		return time.Sunday, false
	}
}
