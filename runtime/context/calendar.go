package context

import "time"

type BarCalendar struct {
	DayOfWeek  float64
	DayOfMonth float64
	Hour       float64
	Minute     float64
	Month      float64
	Second     float64
	Year       float64
	WeekOfYear float64
}

func DecomposeBarTime(unixSeconds int64, loc *time.Location) BarCalendar {
	t := time.Unix(unixSeconds, 0).In(loc)
	_, isoWeek := t.ISOWeek()

	return BarCalendar{
		DayOfWeek:  float64(t.Weekday() + 1),
		DayOfMonth: float64(t.Day()),
		Hour:       float64(t.Hour()),
		Minute:     float64(t.Minute()),
		Month:      float64(t.Month()),
		Second:     float64(t.Second()),
		Year:       float64(t.Year()),
		WeekOfYear: float64(isoWeek),
	}
}
