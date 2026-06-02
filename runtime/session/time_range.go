package session

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// An overnight session has startMinutes > endMinutes (e.g. "1800-0600").
type timeRange struct {
	startMinutes int
	endMinutes   int
	is24Hour     bool
}

func parseTimeRange(timePart string) (timeRange, error) {
	parts := strings.SplitN(timePart, "-", 2)
	if len(parts) != 2 {
		return timeRange{}, fmt.Errorf("invalid time range %q: expected HHMM-HHMM", timePart)
	}
	start, err := parseHHMM(parts[0])
	if err != nil {
		return timeRange{}, fmt.Errorf("start time: %w", err)
	}
	end, err := parseHHMM(parts[1])
	if err != nil {
		return timeRange{}, fmt.Errorf("end time: %w", err)
	}
	return timeRange{
		startMinutes: start,
		endMinutes:   end,
		is24Hour:     start == 0 && end == 23*60+59,
	}, nil
}

func parseHHMM(s string) (int, error) {
	if len(s) != 4 {
		return 0, fmt.Errorf("time component %q must be exactly 4 digits", s)
	}
	h, err := strconv.Atoi(s[:2])
	if err != nil || h < 0 || h > 23 {
		return 0, fmt.Errorf("invalid hour in %q", s)
	}
	m, err := strconv.Atoi(s[2:])
	if err != nil || m < 0 || m > 59 {
		return 0, fmt.Errorf("invalid minute in %q", s)
	}
	return h*60 + m, nil
}

// isOvernight reports whether the session crosses midnight. Pine treats start==end
// (e.g. "1700-1700") as a full-cycle overnight session covering all 24 hours on the
// specified days; is24Hour is reserved for the canonical "0000-2359" form.
func (r timeRange) isOvernight() bool { return r.startMinutes >= r.endMinutes && !r.is24Hour }

// Boundary semantics: start is inclusive; end is inclusive at the exact minute
// but exclusive one second past (Pine convention — 16:45:01 is outside "0950-1645").
func (r timeRange) containsTime(t time.Time) bool {
	if r.is24Hour {
		return true
	}
	cur := t.Hour()*60 + t.Minute()
	sec := t.Second()
	if !r.isOvernight() {
		if cur < r.startMinutes || cur > r.endMinutes {
			return false
		}
		return !(cur == r.endMinutes && sec > 0)
	}
	// Overnight: [startMinutes, midnight) ∪ [midnight, endMinutes]
	afterStart := cur >= r.startMinutes
	beforeEnd := cur < r.endMinutes || (cur == r.endMinutes && sec == 0)
	return afterStart || beforeEnd
}

// Pine assigns overnight sessions to their END day, not the bar's calendar day.
// TV reference: "The Monday session starts Sunday at 17:00 and ends Monday at 17:00.
// It applies Monday through Friday." — the day mask is checked against the closing day.
func (r timeRange) sessionDayOf(t time.Time) time.Weekday {
	if !r.isOvernight() {
		return t.Weekday()
	}
	if t.Hour()*60+t.Minute() >= r.startMinutes {
		return (t.Weekday() + 1) % 7
	}
	return t.Weekday()
}
