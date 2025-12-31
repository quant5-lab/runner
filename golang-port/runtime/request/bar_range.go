package request

type BarRange struct {
	DailyBarIndex    int
	StartHourlyIndex int
	EndHourlyIndex   int
}

func NewBarRange(dailyIdx, startHourly, endHourly int) BarRange {
	return BarRange{
		DailyBarIndex:    dailyIdx,
		StartHourlyIndex: startHourly,
		EndHourlyIndex:   endHourly,
	}
}

func (r BarRange) Contains(hourlyIndex int) bool {
	/* Ranges with no hourly bars (-1) cannot contain any index */
	if r.StartHourlyIndex < 0 || r.EndHourlyIndex < 0 {
		return false
	}
	return hourlyIndex >= r.StartHourlyIndex && hourlyIndex <= r.EndHourlyIndex
}

func (r BarRange) IsBeforeRange(hourlyIndex int) bool {
	return hourlyIndex < r.StartHourlyIndex
}

func (r BarRange) IsAfterRange(hourlyIndex int) bool {
	return hourlyIndex > r.EndHourlyIndex
}
