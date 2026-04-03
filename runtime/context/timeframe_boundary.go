package context

import "time"

type TimeframeBoundaryAligner struct{}

func NewTimeframeBoundaryAligner() *TimeframeBoundaryAligner {
	return &TimeframeBoundaryAligner{}
}

func (a *TimeframeBoundaryAligner) AlignToPeriod(timestampSec int64, timeframe string) int64 {
	multiplier, unit := parseTimeframeComponents(timeframe)
	if multiplier <= 0 {
		return timestampSec
	}

	switch unit {
	case 'M':
		return a.alignToMonthBoundary(timestampSec, multiplier)
	case 'W', 'w':
		return a.alignToWeekBoundary(timestampSec, multiplier)
	default:
		return a.alignToFixedDuration(timestampSec, multiplier, unit)
	}
}

func (a *TimeframeBoundaryAligner) alignToFixedDuration(timestampSec int64, multiplier int, unit byte) int64 {
	unitSec, ok := fixedDurationUnitSeconds[unit]
	if !ok {
		return timestampSec
	}
	periodSec := int64(multiplier) * unitSec
	return (timestampSec / periodSec) * periodSec
}

func (a *TimeframeBoundaryAligner) alignToWeekBoundary(timestampSec int64, multiplier int) int64 {
	t := time.Unix(timestampSec, 0).UTC()
	daysSinceMonday := (int(t.Weekday()) + 6) % 7
	monday := time.Date(t.Year(), t.Month(), t.Day()-daysSinceMonday, 0, 0, 0, 0, time.UTC)

	if multiplier <= 1 {
		return monday.Unix()
	}

	mondaySec := monday.Unix()
	weekPeriodSec := int64(multiplier) * 604800
	return epochReferenceMonday + ((mondaySec-epochReferenceMonday)/weekPeriodSec)*weekPeriodSec
}

func (a *TimeframeBoundaryAligner) alignToMonthBoundary(timestampSec int64, multiplier int) int64 {
	t := time.Unix(timestampSec, 0).UTC()
	if multiplier <= 1 {
		return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC).Unix()
	}

	monthIndex := int(t.Month()) - 1
	alignedMonth := (monthIndex/multiplier)*multiplier + 1
	return time.Date(t.Year(), time.Month(alignedMonth), 1, 0, 0, 0, 0, time.UTC).Unix()
}

/* Jan 5, 1970 00:00:00 UTC — first Monday after Unix epoch */
const epochReferenceMonday int64 = 345600

var fixedDurationUnitSeconds = map[byte]int64{
	's': 1,
	'm': 60,
	'h': 3600,
	'D': 86400,
	'd': 86400,
}

func parseTimeframeComponents(timeframe string) (int, byte) {
	if len(timeframe) == 0 {
		return 0, 0
	}

	last := timeframe[len(timeframe)-1]
	if last >= '0' && last <= '9' {
		return parsePositiveInt(timeframe), 'm'
	}

	if len(timeframe) == 1 {
		return 1, last
	}

	return parsePositiveInt(timeframe[:len(timeframe)-1]), last
}

func parsePositiveInt(s string) int {
	var result int
	for _, ch := range s {
		if ch >= '0' && ch <= '9' {
			result = result*10 + int(ch-'0')
		}
	}
	if result <= 0 {
		return 1
	}
	return result
}
