package integration

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/quant5-lab/runner/tests/util"
)

/* Bars at known UTC timestamps spanning Mon–Sun with varied hours */
func calendarBars() ([]map[string]interface{}, []time.Time) {
	dates := []time.Time{
		time.Date(2024, 3, 11, 9, 30, 0, 0, time.UTC),
		time.Date(2024, 3, 12, 10, 15, 0, 0, time.UTC),
		time.Date(2024, 3, 13, 14, 0, 0, 0, time.UTC),
		time.Date(2024, 3, 14, 22, 45, 0, 0, time.UTC),
		time.Date(2024, 3, 15, 6, 30, 0, 0, time.UTC),
		time.Date(2024, 3, 16, 12, 0, 0, 0, time.UTC),
		time.Date(2024, 3, 17, 18, 0, 0, 0, time.UTC),
		time.Date(2024, 3, 18, 8, 0, 0, 0, time.UTC),
	}

	bars := make([]map[string]interface{}, len(dates))
	for i, dt := range dates {
		bars[i] = map[string]interface{}{
			"time":   dt.Unix(),
			"open":   100.0 + float64(i),
			"high":   105.0 + float64(i),
			"low":    95.0 + float64(i),
			"close":  102.0 + float64(i),
			"volume": 1000.0,
		}
	}
	return bars, dates
}

/* Pine dayofweek: 1=Sunday, 2=Monday ... 7=Saturday (Go Weekday()+1) */
func pineDayOfWeek(t time.Time) float64 {
	return float64(t.Weekday() + 1)
}

func TestCalendarBuiltin_DayOfWeek(t *testing.T) {
	t.Parallel()
	bars, dates := calendarBars()

	script := `//@version=5
indicator("Calendar DOW")
plot(dayofweek, "dow")
`
	exec := util.NewPineExecutor(t)
	output := exec.ExecuteScriptWithCustomData(t, "cal-dow", script, bars)
	vals := exec.ExtractPlotValues(t, output, "dow")

	for i, dt := range dates {
		want := pineDayOfWeek(dt)
		if vals[i] != want {
			t.Errorf("bar[%d] %s: dayofweek=%v, want %v", i, dt.Weekday(), vals[i], want)
		}
	}
}

func TestCalendarBuiltin_HourMinute(t *testing.T) {
	t.Parallel()
	bars, dates := calendarBars()

	script := `//@version=5
indicator("Calendar HM")
plot(hour, "hour")
plot(minute, "minute")
`
	exec := util.NewPineExecutor(t)
	output := exec.ExecuteScriptWithCustomData(t, "cal-hm", script, bars)
	hours := exec.ExtractPlotValues(t, output, "hour")
	minutes := exec.ExtractPlotValues(t, output, "minute")

	for i, dt := range dates {
		if hours[i] != float64(dt.Hour()) {
			t.Errorf("bar[%d]: hour=%v, want %v", i, hours[i], dt.Hour())
		}
		if minutes[i] != float64(dt.Minute()) {
			t.Errorf("bar[%d]: minute=%v, want %v", i, minutes[i], dt.Minute())
		}
	}
}

func TestCalendarBuiltin_YearMonthDay(t *testing.T) {
	t.Parallel()
	bars, dates := calendarBars()

	script := `//@version=5
indicator("Calendar YMD")
plot(year, "year")
plot(month, "month")
plot(dayofmonth, "dom")
`
	exec := util.NewPineExecutor(t)
	output := exec.ExecuteScriptWithCustomData(t, "cal-ymd", script, bars)
	years := exec.ExtractPlotValues(t, output, "year")
	months := exec.ExtractPlotValues(t, output, "month")
	doms := exec.ExtractPlotValues(t, output, "dom")

	for i, dt := range dates {
		if years[i] != float64(dt.Year()) {
			t.Errorf("bar[%d]: year=%v, want %v", i, years[i], dt.Year())
		}
		if months[i] != float64(dt.Month()) {
			t.Errorf("bar[%d]: month=%v, want %v", i, months[i], int(dt.Month()))
		}
		if doms[i] != float64(dt.Day()) {
			t.Errorf("bar[%d]: dayofmonth=%v, want %v", i, doms[i], dt.Day())
		}
	}
}

func TestCalendarBuiltin_DayOfWeekConstantsMatch(t *testing.T) {
	t.Parallel()
	bars, dates := calendarBars()

	script := `//@version=5
indicator("DOW Constants")
isMonday = dayofweek == dayofweek.monday ? 1.0 : 0.0
isFriday = dayofweek == dayofweek.friday ? 1.0 : 0.0
isSunday = dayofweek == dayofweek.sunday ? 1.0 : 0.0
plot(isMonday, "mon")
plot(isFriday, "fri")
plot(isSunday, "sun")
`
	exec := util.NewPineExecutor(t)
	output := exec.ExecuteScriptWithCustomData(t, "cal-dow-const", script, bars)
	mon := exec.ExtractPlotValues(t, output, "mon")
	fri := exec.ExtractPlotValues(t, output, "fri")
	sun := exec.ExtractPlotValues(t, output, "sun")

	for i, dt := range dates {
		wantMon := boolToFloat(dt.Weekday() == time.Monday)
		wantFri := boolToFloat(dt.Weekday() == time.Friday)
		wantSun := boolToFloat(dt.Weekday() == time.Sunday)

		if mon[i] != wantMon {
			t.Errorf("bar[%d] %s: isMonday=%v, want %v", i, dt.Weekday(), mon[i], wantMon)
		}
		if fri[i] != wantFri {
			t.Errorf("bar[%d] %s: isFriday=%v, want %v", i, dt.Weekday(), fri[i], wantFri)
		}
		if sun[i] != wantSun {
			t.Errorf("bar[%d] %s: isSunday=%v, want %v", i, dt.Weekday(), sun[i], wantSun)
		}
	}
}

func TestCalendarBuiltin_LastBarIndex(t *testing.T) {
	t.Parallel()
	bars, _ := calendarBars()

	script := `//@version=5
indicator("Last Bar Index")
plot(last_bar_index, "lbi")
plot(bar_index, "bi")
`
	exec := util.NewPineExecutor(t)
	output := exec.ExecuteScriptWithCustomData(t, "cal-lbi", script, bars)
	lbi := exec.ExtractPlotValues(t, output, "lbi")
	bi := exec.ExtractPlotValues(t, output, "bi")

	expected := float64(len(bars) - 1)
	for i := range bars {
		if lbi[i] != expected {
			t.Errorf("bar[%d]: last_bar_index=%v, want %v", i, lbi[i], expected)
		}
		if bi[i] != float64(i) {
			t.Errorf("bar[%d]: bar_index=%v, want %v", i, bi[i], i)
		}
	}
}

func TestCalendarBuiltin_DayOfWeekStrategyTrades(t *testing.T) {
	t.Parallel()

	dates := make([]time.Time, 21)
	for i := range dates {
		dates[i] = time.Date(2024, 3, 4+i, 14, 0, 0, 0, time.UTC)
	}

	bars := make([]map[string]interface{}, len(dates))
	for i, dt := range dates {
		bars[i] = map[string]interface{}{
			"time":   dt.Unix(),
			"open":   100.0 + float64(i%5),
			"high":   105.0 + float64(i%5),
			"low":    95.0 + float64(i%5),
			"close":  102.0 + float64(i%5),
			"volume": 1000.0,
		}
	}

	script := `//@version=5
strategy("DOW Strategy", overlay=true)
if dayofweek == dayofweek.wednesday
    strategy.entry("Long", strategy.long)
if dayofweek == dayofweek.friday
    strategy.close("Long")
`
	exec := util.NewPineExecutor(t)
	raw := exec.ExecuteScriptWithCustomDataRaw(t, "cal-dow-strat", script, bars)

	var result struct {
		Strategy struct {
			Trades []struct {
				EntryBar int `json:"entryBar"`
				ExitBar  int `json:"exitBar"`
			} `json:"trades"`
		} `json:"strategy"`
	}
	if err := json.Unmarshal(raw, &result); err != nil {
		t.Fatalf("parse output: %v", err)
	}
	if len(result.Strategy.Trades) == 0 {
		t.Fatal("dayofweek-driven strategy produced no trades")
	}
}

func TestCalendarBuiltin_HistoricalSubscript(t *testing.T) {
	t.Parallel()
	bars, dates := calendarBars()

	script := `//@version=5
indicator("DOW Historical")
plot(dayofweek, "dow0")
plot(dayofweek[1], "dow1")
`
	exec := util.NewPineExecutor(t)
	output := exec.ExecuteScriptWithCustomData(t, "cal-dow-hist", script, bars)
	dow0 := exec.ExtractPlotValues(t, output, "dow0")
	dow1 := exec.ExtractPlotValues(t, output, "dow1")

	for i := 1; i < len(dates); i++ {
		wantCurrent := pineDayOfWeek(dates[i])
		wantPrev := pineDayOfWeek(dates[i-1])

		if dow0[i] != wantCurrent {
			t.Errorf("bar[%d]: dayofweek[0]=%v, want %v", i, dow0[i], wantCurrent)
		}
		if dow1[i] != wantPrev {
			t.Errorf("bar[%d]: dayofweek[1]=%v, want %v (previous bar)", i, dow1[i], wantPrev)
		}
	}
}

func boolToFloat(b bool) float64 {
	if b {
		return 1.0
	}
	return 0.0
}
