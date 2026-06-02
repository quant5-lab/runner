package session

import (
	"fmt"
	"time"
)

// weekdayMask is an accept-set of days-of-week, indexed by time.Weekday (0=Sunday…6=Saturday).
type weekdayMask [7]bool

func (m weekdayMask) allows(w time.Weekday) bool { return m[w] }

// Pine convention: '1'=Sunday, '2'=Monday … '7'=Saturday (not Go's 0-based Weekday).
// Duplicate digits are collapsed (set semantics).
func parseDaysSuffix(suffix string) (weekdayMask, error) {
	var m weekdayMask
	if suffix == "" {
		return m, fmt.Errorf("days suffix cannot be empty")
	}
	for _, ch := range suffix {
		if ch < '1' || ch > '7' {
			return m, fmt.Errorf("invalid days suffix %q: each char must be 1-7", suffix)
		}
		m[ch-'1'] = true
	}
	return m, nil
}

// TV-documented v4→v5 breaking change:
// "The default days are: 1234567, which is different in Pine Script v5 than in
// earlier versions where 23456 (weekdays) is used."
func defaultMaskForVersion(pineVersion int) weekdayMask {
	if pineVersion < pineVersion5 {
		return weekdayOnlyMask()
	}
	return allDaysMask()
}

func weekdayOnlyMask() weekdayMask {
	var m weekdayMask
	m[time.Monday] = true
	m[time.Tuesday] = true
	m[time.Wednesday] = true
	m[time.Thursday] = true
	m[time.Friday] = true
	return m
}

func allDaysMask() weekdayMask {
	var m weekdayMask
	for i := range m {
		m[i] = true
	}
	return m
}
