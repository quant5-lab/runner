package context

import "time"

type TimeframeBoundaryAligner struct{}

func NewTimeframeBoundaryAligner() *TimeframeBoundaryAligner {
	return &TimeframeBoundaryAligner{}
}

// AlignToPeriod aligns timestampSec to the UTC period boundary for timeframe.
func (a *TimeframeBoundaryAligner) AlignToPeriod(timestampSec int64, timeframe string) int64 {
	return a.AlignToPeriodWithAnchor(timestampSec, timeframe, PeriodAnchor{})
}

// AlignToPeriodWithAnchor aligns timestampSec to the period boundary for
// timeframe, tiling from anchor.SessionOpenMinute within anchor.Timezone.
//
// Calendar-scale boundaries (D, W, M) use midnight in anchor.Timezone rather
// than UTC midnight.  Intraday boundaries tile from SessionOpenMinute within
// each calendar day in anchor.Timezone.  A zero PeriodAnchor produces identical
// results to UTC-epoch floor division — maintaining backward compatibility.
func (a *TimeframeBoundaryAligner) AlignToPeriodWithAnchor(timestampSec int64, timeframe string, anchor PeriodAnchor) int64 {
	multiplier, unit := parseTimeframeComponents(timeframe)
	if multiplier <= 0 {
		return timestampSec
	}

	loc := loadLocation(anchor.Timezone)

	switch unit {
	case 'M':
		return a.alignToMonthBoundary(timestampSec, multiplier, loc)
	case 'W', 'w':
		return a.alignToWeekBoundary(timestampSec, multiplier, loc)
	case 'D', 'd':
		return a.alignToDayBoundary(timestampSec, multiplier, loc)
	default:
		return a.alignToIntradayPeriod(timestampSec, multiplier, unit, loc, anchor.SessionOpenMinute)
	}
}

// alignToIntradayPeriod: when sessionOpenMinute is zero and loc is UTC, the
// result is identical to UTC-epoch floor division (backward compatibility).
func (a *TimeframeBoundaryAligner) alignToIntradayPeriod(
	timestampSec int64, multiplier int, unit byte, loc *time.Location, sessionOpenMinute int,
) int64 {
	unitSec, ok := fixedDurationUnitSeconds[unit]
	if !ok {
		return timestampSec
	}
	periodMinutes := int(int64(multiplier)*unitSec) / 60
	if periodMinutes <= 0 {
		// Sub-minute period: fall back to epoch-relative floor in seconds.
		periodSec := int64(multiplier) * unitSec
		return (timestampSec / periodSec) * periodSec
	}

	barLocal := time.Unix(timestampSec, 0).In(loc)
	y, m, d := barLocal.Date()
	sessionOpen := time.Date(y, m, d, sessionOpenMinute/60, sessionOpenMinute%60, 0, 0, loc)

	minutesSinceOpen := int(barLocal.Sub(sessionOpen) / time.Minute)
	if minutesSinceOpen < 0 {
		// Recompute from the adjusted sessionOpen rather than doing arithmetic on
		// the truncated value — avoids off-by-one when barLocal has sub-minute seconds.
		sessionOpen = sessionOpen.Add(-24 * time.Hour)
		minutesSinceOpen = int(barLocal.Sub(sessionOpen) / time.Minute)
	}

	periodIndex := minutesSinceOpen / periodMinutes
	return sessionOpen.Add(time.Duration(periodIndex*periodMinutes) * time.Minute).Unix()
}

func (a *TimeframeBoundaryAligner) alignToDayBoundary(timestampSec int64, multiplier int, loc *time.Location) int64 {
	t := time.Unix(timestampSec, 0).In(loc)
	y, m, d := t.Date()
	if multiplier <= 1 {
		return time.Date(y, m, d, 0, 0, 0, 0, loc).Unix()
	}
	midnight := time.Date(y, m, d, 0, 0, 0, 0, loc)
	periodSec := int64(multiplier) * 86400
	return (midnight.Unix() / periodSec) * periodSec
}

func (a *TimeframeBoundaryAligner) alignToWeekBoundary(timestampSec int64, multiplier int, loc *time.Location) int64 {
	t := time.Unix(timestampSec, 0).In(loc)
	daysSinceMonday := (int(t.Weekday()) + 6) % 7
	y, m, d := t.Date()
	monday := time.Date(y, m, d-daysSinceMonday, 0, 0, 0, 0, loc)

	if multiplier <= 1 {
		return monday.Unix()
	}

	mondaySec := monday.Unix()
	weekPeriodSec := int64(multiplier) * 604800
	return epochReferenceMonday + ((mondaySec-epochReferenceMonday)/weekPeriodSec)*weekPeriodSec
}

func (a *TimeframeBoundaryAligner) alignToMonthBoundary(timestampSec int64, multiplier int, loc *time.Location) int64 {
	t := time.Unix(timestampSec, 0).In(loc)
	if multiplier <= 1 {
		return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, loc).Unix()
	}

	monthIndex := int(t.Month()) - 1
	alignedMonth := (monthIndex/multiplier)*multiplier + 1
	return time.Date(t.Year(), time.Month(alignedMonth), 1, 0, 0, 0, 0, loc).Unix()
}

// loadLocation resolves an IANA timezone name to a *time.Location.
// Empty string and "UTC" both return time.UTC.  Unrecognised names fall back
// to time.UTC so period computation is safe even with unexpected zone strings.
func loadLocation(tz string) *time.Location {
	if tz == "" || tz == "UTC" {
		return time.UTC
	}
	loc, err := time.LoadLocation(tz)
	if err != nil {
		return time.UTC
	}
	return loc
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
