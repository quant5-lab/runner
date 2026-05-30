package session

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

/*
Session represents a time range filter for trading hours plus an optional
weekday mask. Pine format: "HHMM-HHMM[:DAYS]".

Pine documentation references (sessions concept):
https://www.tradingview.com/pine-script-docs/v5/concepts/sessions/

DAYS digit convention per Pine:

	1 = Sunday, 2 = Monday, 3 = Tuesday, 4 = Wednesday,
	5 = Thursday, 6 = Friday, 7 = Saturday.

Per-version defaults when the ":DAYS" suffix is absent:

	v4 → "23456" (Mon-Fri only)
	v5 → "1234567" (all 7 days)

Design Philosophy (SOLID):
  - Single Responsibility: session parsing and time/weekday filtering.
  - Open/Closed: new fields (days mask) added without breaking IsInSession.
  - Interface Segregation: minimal public API (ParseWithVersion + IsInSession + TimeFuncWithVersion).
  - Dependency Inversion: standard library time only.
*/
type Session struct {
	startHour   int
	startMinute int
	endHour     int
	endMinute   int
	is24Hour    bool    // Optimization: 0000-2359 sessions
	days        [7]bool // index = int(time.Weekday): Sunday=0 .. Saturday=6
}

const (
	pineVersion4 = 4
	pineVersion5 = 5
)

/*
defaultDaysForVersion returns the implicit DAYS mask used when no ":DDDDDDD"
suffix is supplied. v4 defaults to Mon-Fri; v5 defaults to all 7 days. For
any other or unknown version we apply v5 (all 7 days) as a forward-compat
fallback. Callers may log a warning via the returned ok=false signal.
*/
func defaultDaysForVersion(pineVersion int) (days [7]bool, ok bool) {
	switch pineVersion {
	case pineVersion4:
		// "23456" → Mon-Fri
		days[time.Monday] = true
		days[time.Tuesday] = true
		days[time.Wednesday] = true
		days[time.Thursday] = true
		days[time.Friday] = true
		return days, true
	case pineVersion5:
		for i := range days {
			days[i] = true
		}
		return days, true
	default:
		// Forward-compat: assume v5 semantics.
		for i := range days {
			days[i] = true
		}
		return days, false
	}
}

/*
parseDaysSuffix decodes a string like "23456" into a [7]bool weekday mask.
Each character must be a digit '1'..'7'. Duplicates are tolerated (collapse
to set semantics). Returns an error for any illegal character.

Pine digit d → time.Weekday(d-1):

	'1' → Sunday(0), '2' → Monday(1), ..., '7' → Saturday(6).
*/
func parseDaysSuffix(suffix string) ([7]bool, error) {
	var days [7]bool
	if suffix == "" {
		return days, fmt.Errorf("days suffix cannot be empty")
	}
	for _, ch := range suffix {
		if ch < '1' || ch > '7' {
			return days, fmt.Errorf("invalid days suffix %q: each char must be 1-7", suffix)
		}
		days[int(ch-'1')] = true
	}
	return days, nil
}

/*
Parse is the deprecated single-argument entry point retained for backwards
compatibility. It defers to ParseWithVersion with pineVersion=5, matching
the v5 default of "all 7 days" so legacy callers see no behavioural change
from the pre-day-mask implementation.

Deprecated: use ParseWithVersion to thread the script's Pine version. New
code MUST call ParseWithVersion so that v4 strategies receive the correct
Mon-Fri default.
*/
func Parse(sessionStr string) (*Session, error) {
	return ParseWithVersion(sessionStr, pineVersion5)
}

/*
ParseWithVersion creates a Session from "HHMM-HHMM[:DAYS]" applying the
per-version default for DAYS when the suffix is absent.

Inputs accepted:

	"0950-1645"          → per-version default for the day mask
	"0950-1645:23456"    → explicit Mon-Fri mask (overrides default)
	"0000-2359:1234567"  → 24-hour all-week session

Validation: time portion must be "HHMM-HHMM" with valid hours/minutes;
DAYS suffix (if present) must consist of digits 1-7 only.
*/
func ParseWithVersion(sessionStr string, pineVersion int) (*Session, error) {
	if sessionStr == "" {
		return nil, fmt.Errorf("session string cannot be empty")
	}

	timePart := sessionStr
	daysPart := ""
	hasColon := false
	if idx := strings.Index(sessionStr, ":"); idx >= 0 {
		timePart = sessionStr[:idx]
		daysPart = sessionStr[idx+1:]
		hasColon = true
	}

	parts := strings.Split(timePart, "-")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid session format: %q (expected HHMM-HHMM[:DAYS])", sessionStr)
	}

	startTime := parts[0]
	endTime := parts[1]

	if len(startTime) != 4 || len(endTime) != 4 {
		return nil, fmt.Errorf("invalid session format: %q (times must be 4 digits)", sessionStr)
	}

	startHour, err := strconv.Atoi(startTime[:2])
	if err != nil || startHour < 0 || startHour > 23 {
		return nil, fmt.Errorf("invalid start hour: %q", startTime[:2])
	}

	startMinute, err := strconv.Atoi(startTime[2:4])
	if err != nil || startMinute < 0 || startMinute > 59 {
		return nil, fmt.Errorf("invalid start minute: %q", startTime[2:4])
	}

	endHour, err := strconv.Atoi(endTime[:2])
	if err != nil || endHour < 0 || endHour > 23 {
		return nil, fmt.Errorf("invalid end hour: %q", endTime[:2])
	}

	endMinute, err := strconv.Atoi(endTime[2:4])
	if err != nil || endMinute < 0 || endMinute > 59 {
		return nil, fmt.Errorf("invalid end minute: %q", endTime[2:4])
	}

	var days [7]bool
	if hasColon {
		days, err = parseDaysSuffix(daysPart)
		if err != nil {
			return nil, err
		}
	} else {
		days, _ = defaultDaysForVersion(pineVersion)
	}

	s := &Session{
		startHour:   startHour,
		startMinute: startMinute,
		endHour:     endHour,
		endMinute:   endMinute,
		is24Hour:    startHour == 0 && startMinute == 0 && endHour == 23 && endMinute == 59,
		days:        days,
	}

	return s, nil
}

/*
IsInSession checks if the given timestamp is within the session time range
AND inside the session's weekday mask.

Parameters:

	timestamp: Unix timestamp in MILLISECONDS
	timezone:  IANA timezone name (e.g., "UTC", "America/New_York", "Europe/Moscow")

Order of checks:
 1. Day mask (cheap, short-circuits weekends for v4 strategies).
 2. 24-hour fast path.
 3. HHMM range, intra-day or overnight.

Edge Cases:
  - Overnight sessions (18:00-06:00): handles day boundary crossing.
  - Exact boundaries: 09:50:00 IN, 16:45:00 IN, 16:45:01 OUT.
  - 24-hour session (0000-2359): returns true once day mask passes.
  - Timezone conversion: timestamp resolved into exchange timezone before
    weekday and HHMM comparison.
*/
func (s *Session) IsInSession(timestamp int64, timezone string) bool {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		loc = time.UTC
	}

	t := time.Unix(timestamp/1000, 0).In(loc)

	if !s.days[int(t.Weekday())] {
		return false
	}

	if s.is24Hour {
		return true
	}

	hour := t.Hour()
	minute := t.Minute()
	second := t.Second()

	startMinutes := s.startHour*60 + s.startMinute
	endMinutes := s.endHour*60 + s.endMinute
	currentMinutes := hour*60 + minute

	if startMinutes <= endMinutes {
		// Regular session (same day): 0950-1645
		// Start INCLUSIVE (09:50:00 IN); end INCLUSIVE at exact minute,
		// EXCLUSIVE one second past (16:45:01+ OUT).
		if currentMinutes < startMinutes {
			return false
		}
		if currentMinutes > endMinutes {
			return false
		}
		if currentMinutes == endMinutes && second > 0 {
			return false
		}
		return true
	}

	// Overnight session (crosses midnight): 1800-0600
	afterStart := currentMinutes > startMinutes || (currentMinutes == startMinutes)
	beforeEnd := currentMinutes < endMinutes || (currentMinutes == endMinutes && second == 0)
	return afterStart || beforeEnd
}

/*
TimeFunc is the deprecated 4-arg shim that defers to TimeFuncWithVersion
with pineVersion=5. It preserves behaviour for any caller that has not yet
been threaded with the script's Pine version.

Deprecated: use TimeFuncWithVersion. The runner codegen emits the
version-aware form so that v4 strategies pick up the Mon-Fri default.
*/
func TimeFunc(timestamp int64, timeframe string, sessionStr string, timezone string) float64 {
	return TimeFuncWithVersion(timestamp, timeframe, sessionStr, timezone, pineVersion5)
}

/*
TimeFuncWithVersion implements Pine Script's time(timeframe, session) plus
the version-correct default day mask. Returns the bar timestamp (as float64)
when the bar passes both the day mask and the HHMM range; returns NaN
otherwise. Matches Pine semantics where time() with a session parameter is
a filter: caller uses na(time(...)) to detect "outside session".

Parameters:

	timestamp:   Unix timestamp in MILLISECONDS
	timeframe:   reserved for future timeframe-aware behaviour (currently unused)
	sessionStr:  "HHMM-HHMM[:DAYS]" (empty string disables filtering)
	timezone:    IANA timezone name
	pineVersion: 4 or 5 (selects the default DAYS mask when absent)

Timezone Handling:
Session times are interpreted in the exchange timezone (syminfo.timezone)
per Pine spec — MOEX "0950-1645" means 09:50-16:45 Moscow, NYSE "0930-1600"
means 09:30-16:00 New York, Binance "0000-2359" is UTC.
*/
func TimeFuncWithVersion(timestamp int64, timeframe string, sessionStr string, timezone string, pineVersion int) float64 {
	if sessionStr == "" {
		return float64(timestamp)
	}

	s, err := ParseWithVersion(sessionStr, pineVersion)
	if err != nil {
		return math.NaN()
	}

	if s.IsInSession(timestamp, timezone) {
		return float64(timestamp)
	}

	return math.NaN()
}
