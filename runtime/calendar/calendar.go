package calendar

import (
	"math"
	"time"
)

// Calendar extraction for Pine Script date/time builtins.
// All public functions accept timestamps as Unix milliseconds (Pine convention).

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

func toTime(timestampMs float64, timezone string) (time.Time, bool) {
	if math.IsNaN(timestampMs) {
		return time.Time{}, false
	}
	return time.UnixMilli(int64(timestampMs)).In(loadLocation(timezone)), true
}

func Year(timestampMs float64, timezone string) float64 {
	if t, ok := toTime(timestampMs, timezone); ok {
		return float64(t.Year())
	}
	return math.NaN()
}

func Month(timestampMs float64, timezone string) float64 {
	if t, ok := toTime(timestampMs, timezone); ok {
		return float64(t.Month())
	}
	return math.NaN()
}

// Pine convention: 1=Sunday, 2=Monday...7=Saturday
func DayOfWeek(timestampMs float64, timezone string) float64 {
	if t, ok := toTime(timestampMs, timezone); ok {
		return float64(t.Weekday() + 1)
	}
	return math.NaN()
}

func DayOfMonth(timestampMs float64, timezone string) float64 {
	if t, ok := toTime(timestampMs, timezone); ok {
		return float64(t.Day())
	}
	return math.NaN()
}

func Hour(timestampMs float64, timezone string) float64 {
	if t, ok := toTime(timestampMs, timezone); ok {
		return float64(t.Hour())
	}
	return math.NaN()
}

func Minute(timestampMs float64, timezone string) float64 {
	if t, ok := toTime(timestampMs, timezone); ok {
		return float64(t.Minute())
	}
	return math.NaN()
}

func Second(timestampMs float64, timezone string) float64 {
	if t, ok := toTime(timestampMs, timezone); ok {
		return float64(t.Second())
	}
	return math.NaN()
}

func WeekOfYear(timestampMs float64, timezone string) float64 {
	if t, ok := toTime(timestampMs, timezone); ok {
		_, week := t.ISOWeek()
		return float64(week)
	}
	return math.NaN()
}
