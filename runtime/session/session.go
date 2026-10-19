// Package session implements Pine Script session-string parsing and bar-time filtering.
//
// Session strings have the format "HHMM-HHMM[:DAYS]" where the optional DAYS
// suffix is a string of Pine day digits ('1'=Sunday … '7'=Saturday). When no
// suffix is present the per-version default applies:
//
//	Pine v4 → "23456" (Mon-Fri only)
//	Pine v5 → "1234567" (all 7 days)
//
// Reference: https://www.tradingview.com/pine-script-docs/v5/concepts/sessions/
package session

import (
	"fmt"
	"math"
	"strings"
	"time"
)

// Session is a parsed Pine session string: an HHMM time range plus a weekday mask.
type Session struct {
	tr   timeRange
	mask weekdayMask
}

const (
	pineVersion4 = 4
	pineVersion5 = 5
)

// Parse is retained for backward compatibility; new callers must use ParseWithVersion.
func Parse(sessionStr string) (*Session, error) {
	return ParseWithVersion(sessionStr, pineVersion5)
}

func ParseWithVersion(sessionStr string, pineVersion int) (*Session, error) {
	if sessionStr == "" {
		return nil, fmt.Errorf("session string cannot be empty")
	}
	timePart, daysPart, hasDays := strings.Cut(sessionStr, ":")
	tr, err := parseTimeRange(timePart)
	if err != nil {
		return nil, fmt.Errorf("invalid session format %q: %w", sessionStr, err)
	}
	var mask weekdayMask
	if hasDays {
		mask, err = parseDaysSuffix(daysPart)
		if err != nil {
			return nil, err
		}
	} else {
		mask = defaultMaskForVersion(pineVersion)
	}
	return &Session{tr: tr, mask: mask}, nil
}

func (s *Session) IsInSession(timestamp int64, timezone string) bool {
	t := toLocalTime(timestamp, timezone)
	if !s.tr.containsTime(t) {
		return false
	}
	return s.mask.allows(s.tr.sessionDayOf(t))
}

func toLocalTime(timestamp int64, timezone string) time.Time {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		loc = time.UTC
	}
	return time.Unix(timestamp/1000, 0).In(loc)
}

// TimeFunc is retained for backward compatibility; new callers must use TimeFuncWithVersion.
func TimeFunc(timestamp int64, timeframe string, sessionStr string, timezone string) float64 {
	return TimeFuncWithVersion(timestamp, timeframe, sessionStr, timezone, pineVersion5)
}

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
