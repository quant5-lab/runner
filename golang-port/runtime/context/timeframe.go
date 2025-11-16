package context

import (
	"math"
)

/* FindBarIndexByTimestamp finds security bar index matching primary context timestamp
 * Handles downsampling (1h→1D), upsampling (1D→1h), same timeframe (1h→1h)
 * Returns -1 if no matching bar found (timestamp before security data starts)
 */
func FindBarIndexByTimestamp(secCtx *Context, targetTimestamp int64) int {
	if len(secCtx.Data) == 0 {
		return -1
	}

	/* Binary search for closest bar <= targetTimestamp */
	left, right := 0, len(secCtx.Data)-1
	result := -1

	for left <= right {
		mid := (left + right) / 2
		barTime := secCtx.Data[mid].Time

		if barTime <= targetTimestamp {
			result = mid
			left = mid + 1
		} else {
			right = mid - 1
		}
	}

	return result
}

/* GetSecurityValue retrieves value from security context at matched bar index
 * Returns NaN if bar not found or index out of range
 */
func GetSecurityValue(secCtx *Context, targetTimestamp int64, getValue func(*Context, int) float64) float64 {
	barIdx := FindBarIndexByTimestamp(secCtx, targetTimestamp)
	if barIdx < 0 || barIdx >= len(secCtx.Data) {
		return math.NaN()
	}

	/* Temporarily set BarIndex for offset-based accessors (close[1], etc) */
	originalIdx := secCtx.BarIndex
	secCtx.BarIndex = barIdx
	value := getValue(secCtx, barIdx)
	secCtx.BarIndex = originalIdx

	return value
}

/* TimeframeToSeconds converts Pine timeframe string to seconds
 * Examples: "1h" → 3600, "1D" → 86400, "5m" → 300
 */
func TimeframeToSeconds(tf string) int64 {
	if len(tf) < 2 {
		return 0
	}

	multiplier := int64(1)
	unit := tf[len(tf)-1]

	/* Extract numeric multiplier */
	numStr := tf[:len(tf)-1]
	if numStr != "" {
		var num int64
		for _, c := range numStr {
			if c >= '0' && c <= '9' {
				num = num*10 + int64(c-'0')
			}
		}
		if num > 0 {
			multiplier = num
		}
	}

	/* Convert unit to seconds */
	switch unit {
	case 'm', 'M':
		return multiplier * 60
	case 'h', 'H':
		return multiplier * 3600
	case 'D', 'd':
		return multiplier * 86400
	case 'W', 'w':
		return multiplier * 604800
	default:
		return 0
	}
}

/* AlignTimestampToTimeframe rounds timestamp down to timeframe boundary
 * Example: 2024-01-01 14:30:00 aligned to 1D → 2024-01-01 00:00:00
 */
func AlignTimestampToTimeframe(timestamp int64, timeframeSeconds int64) int64 {
	if timeframeSeconds <= 0 {
		return timestamp
	}
	return (timestamp / timeframeSeconds) * timeframeSeconds
}

/* GetAlignedTimestamp returns timestamp aligned to security timeframe
 * Used for upsampling: repeat daily value across all hourly bars of that day
 */
func GetAlignedTimestamp(ctx *Context, secTimeframe string) int64 {
	if ctx.BarIndex < 0 || ctx.BarIndex >= len(ctx.Data) {
		return 0
	}

	currentTimestamp := ctx.Data[ctx.BarIndex].Time
	tfSeconds := TimeframeToSeconds(secTimeframe)
	return AlignTimestampToTimeframe(currentTimestamp, tfSeconds)
}
