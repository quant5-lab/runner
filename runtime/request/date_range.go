package request

import (
	"time"

	"github.com/quant5-lab/runner/runtime/context"
)

type DateRange struct {
	StartDate string
	EndDate   string
	Timezone  string
}

func NewDateRangeFromBars(bars []context.OHLCV, timezone string) DateRange {
	if len(bars) == 0 {
		return DateRange{Timezone: timezone}
	}

	firstTimestamp := bars[0].Time
	lastTimestamp := bars[len(bars)-1].Time

	return DateRange{
		StartDate: formatAsDateInTimezone(firstTimestamp, timezone),
		EndDate:   formatAsDateInTimezone(lastTimestamp, timezone),
		Timezone:  timezone,
	}
}

func (dr DateRange) Contains(date string) bool {
	return date >= dr.StartDate && date <= dr.EndDate
}

func (dr DateRange) IsEmpty() bool {
	return dr.StartDate == "" || dr.EndDate == ""
}

func formatAsDateInTimezone(timestamp int64, timezone string) string {
	if timezone == "" {
		timezone = "UTC"
	}

	location, err := time.LoadLocation(timezone)
	if err != nil {
		location = time.UTC
	}

	return time.Unix(timestamp, 0).In(location).Format("2006-01-02")
}

func ExtractDateInTimezone(timestamp int64, timezone string) string {
	return formatAsDateInTimezone(timestamp, timezone)
}
