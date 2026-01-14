package request

/*
BarRange maps a bar from one timeframe to a range of bars in another timeframe.

Field naming is legacy (DailyBarIndex, HourlyIndex) but applies to any timeframe pair.
Actual meaning depends on mapping direction:

DOWNSCALING (D→H, target TF < base TF):
  - DailyBarIndex: Target timeframe bar index (e.g., Daily bar #5)
  - StartHourlyIndex: First base TF bar on that day (e.g., Hourly bar #120)
  - EndHourlyIndex: Last base TF bar on that day (e.g., Hourly bar #126)

UPSCALING (M→D, W→D, target TF > base TF):
  - DailyBarIndex: Base timeframe bar index (e.g., Monthly bar #3)
  - StartHourlyIndex: First target TF bar in that period (e.g., Daily bar #90)
  - EndHourlyIndex: Last target TF bar in that period (e.g., Daily bar #110)

Value -1 indicates no bars available for that index.
*/
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
