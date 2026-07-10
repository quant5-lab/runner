package calendar

import (
	"math"
	"time"
)

/*
	RFC 2822 and ISO 8601 layouts ordered from most specific to least.

"2006-01-02:15:04" is the Pine Script v4 timestamp() string form.
*/
var dateFormats = []string{
	time.RFC1123Z,
	time.RFC3339,
	"02 Jan 2006 15:04:05 -0700",
	"2006-01-02T15:04:05",
	"2006-01-02 15:04:05",
	"2006-01-02:15:04",
	"2 Jan 2006",
	"2006-01-02",
}

func Timestamp(year, month, day, hour, minute, second float64, timezone string) float64 {
	if math.IsNaN(year) || math.IsNaN(month) || math.IsNaN(day) {
		return math.NaN()
	}

	h, m, s := 0, 0, 0
	if !math.IsNaN(hour) {
		h = int(hour)
	}
	if !math.IsNaN(minute) {
		m = int(minute)
	}
	if !math.IsNaN(second) {
		s = int(second)
	}

	loc := loadLocation(timezone)
	t := time.Date(int(year), time.Month(int(month)), int(day), h, m, s, 0, loc)
	return float64(t.UnixMilli())
}

func TimestampFromString(dateStr string, timezone string) float64 {
	loc := loadLocation(timezone)

	for _, layout := range dateFormats {
		if t, err := time.ParseInLocation(layout, dateStr, loc); err == nil {
			return float64(t.UnixMilli())
		}
	}

	return math.NaN()
}
