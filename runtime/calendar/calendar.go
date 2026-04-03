package calendar

import (
	"math"
	"time"
)

/* Calendar extraction for Pine Script date/time builtins (timestamps in Unix seconds) */

func loadLocation(timezone string) *time.Location {
	if timezone == "" {
		return time.UTC
	}
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return time.UTC
	}
	return loc
}

func toTime(timestampSec float64, timezone string) (time.Time, bool) {
	if math.IsNaN(timestampSec) {
		return time.Time{}, false
	}
	return time.Unix(int64(timestampSec), 0).In(loadLocation(timezone)), true
}

func Year(timestampSec float64, timezone string) float64 {
	if t, ok := toTime(timestampSec, timezone); ok {
		return float64(t.Year())
	}
	return math.NaN()
}

func Month(timestampSec float64, timezone string) float64 {
	if t, ok := toTime(timestampSec, timezone); ok {
		return float64(t.Month())
	}
	return math.NaN()
}

/* Pine convention: 1=Sunday, 2=Monday...7=Saturday */
func DayOfWeek(timestampSec float64, timezone string) float64 {
	if t, ok := toTime(timestampSec, timezone); ok {
		return float64(t.Weekday() + 1)
	}
	return math.NaN()
}

func DayOfMonth(timestampSec float64, timezone string) float64 {
	if t, ok := toTime(timestampSec, timezone); ok {
		return float64(t.Day())
	}
	return math.NaN()
}

func Hour(timestampSec float64, timezone string) float64 {
	if t, ok := toTime(timestampSec, timezone); ok {
		return float64(t.Hour())
	}
	return math.NaN()
}

func Minute(timestampSec float64, timezone string) float64 {
	if t, ok := toTime(timestampSec, timezone); ok {
		return float64(t.Minute())
	}
	return math.NaN()
}

func Second(timestampSec float64, timezone string) float64 {
	if t, ok := toTime(timestampSec, timezone); ok {
		return float64(t.Second())
	}
	return math.NaN()
}

func WeekOfYear(timestampSec float64, timezone string) float64 {
	if t, ok := toTime(timestampSec, timezone); ok {
		_, week := t.ISOWeek()
		return float64(week)
	}
	return math.NaN()
}
