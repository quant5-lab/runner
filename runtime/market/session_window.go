package market

import (
	"fmt"
	"sort"
	"strconv"
	"time"
)

// Zero value is 00:00 (midnight).
type ClockTime struct {
	Hour   int
	Minute int
}

func ClockTimeAt(hour, minute int) ClockTime {
	return ClockTime{Hour: hour, Minute: minute}
}

func ParseClockTime(s string) (ClockTime, error) {
	if len(s) != 4 {
		return ClockTime{}, fmt.Errorf("clock time must be 4 digits (HHMM), got %q", s)
	}
	h, err := strconv.Atoi(s[:2])
	if err != nil || h < 0 || h > 23 {
		return ClockTime{}, fmt.Errorf("invalid hour in %q: must be 00–23", s)
	}
	m, err := strconv.Atoi(s[2:])
	if err != nil || m < 0 || m > 59 {
		return ClockTime{}, fmt.Errorf("invalid minute in %q: must be 00–59", s)
	}
	return ClockTime{Hour: h, Minute: m}, nil
}

func (c ClockTime) totalMinutes() int {
	return c.Hour*60 + c.Minute
}

// Zero value (both endpoints at 00:00) is the unbounded sentinel — imposes no
// restriction. When Start >= End the window crosses midnight (overnight session).
type SessionWindow struct {
	Start ClockTime
	End   ClockTime
}

func NewSessionWindow(start, end ClockTime) SessionWindow {
	return SessionWindow{Start: start, End: end}
}

func ParseSessionWindow(s string) (SessionWindow, error) {
	if len(s) != 9 || s[4] != '-' {
		return SessionWindow{}, fmt.Errorf("session window must be HHMM-HHMM, got %q", s)
	}
	start, err := ParseClockTime(s[:4])
	if err != nil {
		return SessionWindow{}, fmt.Errorf("invalid start in %q: %w", s, err)
	}
	end, err := ParseClockTime(s[5:])
	if err != nil {
		return SessionWindow{}, fmt.Errorf("invalid end in %q: %w", s, err)
	}
	return NewSessionWindow(start, end), nil
}

// IsUnbounded returns true for the zero-value sentinel (both endpoints 00:00).
func (w SessionWindow) IsUnbounded() bool {
	return w.Start == (ClockTime{}) && w.End == (ClockTime{})
}

// Contains uses half-open interval [Start, End). Unbounded windows always
// return true. Overnight windows (Start >= End) span [Start, 24:00) ∪ [00:00, End).
func (w SessionWindow) Contains(t time.Time) bool {
	if w.IsUnbounded() {
		return true
	}
	barMinutes := t.Hour()*60 + t.Minute()
	start := w.Start.totalMinutes()
	end := w.End.totalMinutes()
	if start < end {
		return barMinutes >= start && barMinutes < end
	}
	return barMinutes >= start || barMinutes < end
}

type WeekSchedule struct {
	Weekday SessionWindow
	Weekend SessionWindow
}

func UniformWeekSchedule(w SessionWindow) WeekSchedule {
	return WeekSchedule{Weekday: w}
}

func (s WeekSchedule) IsUnbounded() bool {
	return s.Weekday.IsUnbounded() && s.effectiveWeekend().IsUnbounded()
}

func (s WeekSchedule) Contains(t time.Time) bool {
	return s.ContainsFor(t, isWeekendDay(t))
}

func (s WeekSchedule) ContainsFor(t time.Time, treatAsWeekend bool) bool {
	if treatAsWeekend {
		return s.effectiveWeekend().Contains(t)
	}
	return s.Weekday.Contains(t)
}

func (s WeekSchedule) effectiveWeekend() SessionWindow {
	if s.Weekend.IsUnbounded() {
		return s.Weekday
	}
	return s.Weekend
}

func isWeekendDay(t time.Time) bool {
	return t.Weekday() == time.Saturday || t.Weekday() == time.Sunday
}

type DateSessionWindows map[string]SessionWindow

func NewDateSessionWindows(values map[string]string) (DateSessionWindows, error) {
	if len(values) == 0 {
		return nil, nil
	}

	dates := make([]string, 0, len(values))
	for date := range values {
		dates = append(dates, date)
	}
	sort.Strings(dates)

	windows := make(DateSessionWindows, len(values))
	for _, date := range dates {
		if _, err := time.Parse("2006-01-02", date); err != nil {
			return nil, fmt.Errorf("date session window %q: date must be YYYY-MM-DD", date)
		}
		window, err := ParseSessionWindow(values[date])
		if err != nil {
			return nil, fmt.Errorf("date session window %s: %w", date, err)
		}
		windows[date] = window
	}
	return windows, nil
}

func (w DateSessionWindows) Empty() bool {
	return len(w) == 0
}

func (w DateSessionWindows) Window(date string) (SessionWindow, bool) {
	window, ok := w[date]
	return window, ok
}
