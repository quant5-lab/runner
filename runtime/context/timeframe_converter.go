package context

import "fmt"

/* Pine Script spec: "calculations use 2628003 as the number of seconds in one month (365/12 days)" */
const secondsPerMonth int64 = 2628003

/* Pine Script spec: "All values above 31,622,400 (366 days) return 12M" */
const maxFromSecondsThreshold int64 = 31_622_400

type TimeframeConverter struct{}

func NewTimeframeConverter() *TimeframeConverter {
	return &TimeframeConverter{}
}

func (c *TimeframeConverter) ToSeconds(timeframe string) int64 {
	if len(timeframe) == 0 {
		return 0
	}

	// Must be tested before the single-char and suffixed paths so the last digit is
	// never mistaken for a unit letter.
	if c.isAllDigits(timeframe) {
		return c.parseMinuteCount(timeframe) * 60
	}

	if len(timeframe) == 1 {
		return c.unitToSeconds(timeframe[0])
	}

	return c.extractNumericPart(timeframe) * c.unitToSeconds(timeframe[len(timeframe)-1])
}

func (c *TimeframeConverter) isAllDigits(s string) bool {
	for _, ch := range s {
		if ch < '0' || ch > '9' {
			return false
		}
	}
	return len(s) > 0
}

// Returns 1 for a zero or overflowed value to uphold Pine's minimum resolution of 1 minute.
func (c *TimeframeConverter) parseMinuteCount(s string) int64 {
	var result int64
	for _, ch := range s {
		result = result*10 + int64(ch-'0')
	}
	if result <= 0 {
		return 1
	}
	return result
}

// Only called for suffixed tokens; the all-digit path is handled in ToSeconds first.
func (c *TimeframeConverter) extractNumericPart(timeframe string) int64 {
	numStr := timeframe[:len(timeframe)-1]

	if numStr == "" {
		return 1
	}

	var result int64
	for _, char := range numStr {
		if char >= '0' && char <= '9' {
			result = result*10 + int64(char-'0')
		}
	}

	if result == 0 {
		return 1
	}

	return result
}

func (c *TimeframeConverter) FromSeconds(seconds int64) string {
	if seconds <= 0 {
		return ""
	}

	if seconds > maxFromSecondsThreshold {
		return "12M"
	}

	type unit struct {
		suffix string
		secs   int64
	}
	units := []unit{
		{"M", secondsPerMonth},
		{"W", 604800},
		{"D", 86400},
		{"h", 3600},
		{"m", 60},
		{"s", 1},
	}

	for _, u := range units {
		if seconds >= u.secs && seconds%u.secs == 0 {
			multiplier := seconds / u.secs
			if multiplier == 1 {
				return "1" + u.suffix
			}
			return fmt.Sprintf("%d%s", multiplier, u.suffix)
		}
	}

	return "1s"
}

func (c *TimeframeConverter) unitToSeconds(unit byte) int64 {
	unitMap := map[byte]int64{
		's': 1,
		'm': 60,
		'h': 3600,
		'D': 86400,
		'd': 86400,
		'W': 604800,
		'w': 604800,
		'M': secondsPerMonth,
	}

	if seconds, exists := unitMap[unit]; exists {
		return seconds
	}

	return 0
}
