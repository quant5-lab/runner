package context

type TimeframeConverter struct{}

func NewTimeframeConverter() *TimeframeConverter {
	return &TimeframeConverter{}
}

func (c *TimeframeConverter) ToSeconds(timeframe string) int64 {
	if len(timeframe) == 0 {
		return 0
	}

	if len(timeframe) == 1 {
		return c.unitToSeconds(timeframe[0])
	}

	numericPart := c.extractNumericPart(timeframe)
	unit := timeframe[len(timeframe)-1]

	return numericPart * c.unitToSeconds(unit)
}

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

func (c *TimeframeConverter) unitToSeconds(unit byte) int64 {
	unitMap := map[byte]int64{
		's': 1,
		'm': 60,
		'h': 3600,
		'D': 86400,
		'd': 86400,
		'W': 604800,
		'w': 604800,
		'M': 2592000,
	}

	if seconds, exists := unitMap[unit]; exists {
		return seconds
	}

	return 0
}
