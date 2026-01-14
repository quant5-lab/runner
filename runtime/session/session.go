package session

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

/*
Session represents a time range filter for trading hours.
Format: "HHMM-HHMM" (e.g., "0950-1645" = 09:50 to 16:45)

Design Philosophy (SOLID):
- Single Responsibility: Session parsing and time range checking only
- Open/Closed: Extensible for timezone support without modification
- Interface Segregation: Minimal public API (Parse + IsInSession)
- Dependency Inversion: Uses standard library time.Time interface
*/
type Session struct {
	startHour   int
	startMinute int
	endHour     int
	endMinute   int
	is24Hour    bool // Optimization: 0000-2359 sessions
}

/*
Parse creates a Session from "HHMM-HHMM" format string.
Returns error for invalid formats.

Examples:

	"0950-1645" → 09:50 to 16:45 (regular trading hours)
	"0000-2359" → full 24-hour session
	"1800-0600" → overnight session (18:00 to next day 06:00)

Rationale: Parse validates format at creation time (fail-fast principle)
*/
func Parse(sessionStr string) (*Session, error) {
	if sessionStr == "" {
		return nil, fmt.Errorf("session string cannot be empty")
	}

	parts := strings.Split(sessionStr, "-")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid session format: %q (expected HHMM-HHMM)", sessionStr)
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

	s := &Session{
		startHour:   startHour,
		startMinute: startMinute,
		endHour:     endHour,
		endMinute:   endMinute,
		is24Hour:    startHour == 0 && startMinute == 0 && endHour == 23 && endMinute == 59,
	}

	return s, nil
}

/*
IsInSession checks if the given timestamp is within the session time range.
Returns true if within session, false otherwise.

Parameters:

	timestamp: Unix timestamp in MILLISECONDS
	timezone: IANA timezone name (e.g., "UTC", "America/New_York", "Europe/Moscow")

Performance: O(1) time complexity using pre-parsed hour/minute values.
Optimization: 24-hour sessions short-circuit to always return true.

Edge Cases:
- Overnight sessions (18:00-06:00): Handles day boundary crossing
- Exact boundaries: 09:50:00 is IN, 16:45:00 is IN, 16:45:01 is OUT
- 24-hour session (0000-2359): Always returns true (fast path)
- Timezone conversion: Converts timestamp to exchange timezone before comparison
*/
func (s *Session) IsInSession(timestamp int64, timezone string) bool {
	if s.is24Hour {
		return true // Fast path for 24-hour sessions
	}

	// Load the exchange timezone
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		// Fallback to UTC if timezone is invalid
		loc = time.UTC
	}

	// Convert timestamp to exchange timezone
	t := time.Unix(timestamp/1000, 0).In(loc)
	hour := t.Hour()
	minute := t.Minute()
	second := t.Second()

	startMinutes := s.startHour*60 + s.startMinute
	endMinutes := s.endHour*60 + s.endMinute
	currentMinutes := hour*60 + minute

	if startMinutes <= endMinutes {
		// Regular session (same day): 0950-1645
		// Start is INCLUSIVE (09:50:00 is IN)
		// End is INCLUSIVE at exact minute, EXCLUSIVE after first second
		// So: 16:45:00 is IN, 16:45:01+ is OUT
		if currentMinutes < startMinutes {
			return false
		}
		if currentMinutes > endMinutes {
			return false
		}
		// currentMinutes == startMinutes or endMinutes: check seconds
		if currentMinutes == endMinutes && second > 0 {
			return false // 16:45:01+ is OUT
		}
		return true
	}

	// Overnight session (crosses midnight): 1800-0600
	// True if: >= 18:00 OR <= 06:00 (with same second-level precision)
	afterStart := currentMinutes > startMinutes || (currentMinutes == startMinutes)
	beforeEnd := currentMinutes < endMinutes || (currentMinutes == endMinutes && second == 0)
	return afterStart || beforeEnd
}

/*
TimeFunc implements Pine Script's time(timeframe, session, timezone) function.
Returns timestamp if bar is within session, NaN if outside session.

This matches Pine Script semantics where time() with session parameter
acts as a filter: returns valid timestamp during session, NaN otherwise.

The returned timestamp is used with na() to check session state:

	session_open = na(time(timeframe.period, "0950-1645")) ? false : true

Parameters:

	timestamp: Unix timestamp in MILLISECONDS
	timeframe: Timeframe string (currently unused, reserved for future)
	sessionStr: Session string in format "HHMM-HHMM" (e.g., "0950-1645")
	timezone: IANA timezone name (e.g., "UTC", "America/New_York", "Europe/Moscow")
	         Session times are interpreted in this timezone (matches syminfo.timezone behavior)

Performance Consideration:
Pine Script precomputes session bitmasks for O(1) filtering.
Our implementation: O(1) per-bar check using hour/minute comparison.
Result: Equivalent performance for runtime execution.

Timezone Handling:
According to Pine Script documentation, session times are always interpreted
in the exchange timezone (syminfo.timezone), NOT UTC. For example:
  - MOEX: "0950-1645" means 09:50-16:45 Moscow time (UTC+3)
  - NYSE: "0930-1600" means 09:30-16:00 New York time (UTC-5)
  - Binance: "0000-2359" means 00:00-23:59 UTC
*/
func TimeFunc(timestamp int64, timeframe string, sessionStr string, timezone string) float64 {
	if sessionStr == "" {
		return float64(timestamp)
	}

	session, err := Parse(sessionStr)
	if err != nil {
		return math.NaN() // Invalid session = always out of session
	}

	if session.IsInSession(timestamp, timezone) {
		return float64(timestamp)
	}

	return math.NaN()
}
