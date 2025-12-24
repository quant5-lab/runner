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
	return hourlyIndex >= r.StartHourlyIndex && hourlyIndex <= r.EndHourlyIndex
}

func (r BarRange) IsBeforeRange(hourlyIndex int) bool {
	return hourlyIndex < r.StartHourlyIndex
}

func (r BarRange) IsAfterRange(hourlyIndex int) bool {
	return hourlyIndex > r.EndHourlyIndex
}
