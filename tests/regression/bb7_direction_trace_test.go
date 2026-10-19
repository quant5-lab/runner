package regression

import (
	"math"
	"path/filepath"
	"testing"
)

// bb7SBERPDisputedBars are SBERP 1h bar indices that fall on Sundays in
// Europe/Moscow time, verified by TestBB7_DisputedBars_AreWeekendDays.
var bb7SBERPDisputedBars = []int{1798, 3327, 3766}

// bb7BTCUSDTDisputedBars are BTCUSDT 1h bar indices that fall on weekends in UTC:
// two Saturdays (2181, 4701) and one Sunday (4557), verified by TestBB7_DisputedBars_AreWeekendDays.
var bb7BTCUSDTDisputedBars = []int{2181, 4557, 4701}

type directionProbe struct {
	HourlyBarIdx int
	HourlyDate   string
	HourlyClose  float64
	DailyBarIdx  int
	DailyDate    string
	SMA20        float64
	SMA50        float64
	SmaBullish   bool
	EntryType    string
}

// TestBB7_SBERP_DirectionTrace asserts that at each weekend bar the daily
// SMA direction is internally consistent: SMA20 < SMA50 → entry_type=Short.
func TestBB7_SBERP_DirectionTrace(t *testing.T) {
	root := projectRootFromCwd()
	dataDir := filepath.Join(root, "tests", "golden", "fixtures", "data")

	hourly := loadOHLCVBars(t, filepath.Join(dataDir, "SBERP-1h.json"))
	daily := loadOHLCVBars(t, filepath.Join(dataDir, "SBERP_1D.json"))

	for _, barIdx := range bb7SBERPDisputedBars {
		probe := buildDirectionProbe(t, hourly, daily, barIdx)
		assertProbeConsistency(t, probe)
	}
}

// TestBB7_BTCUSDT_DirectionTrace asserts that at each weekend BTCUSDT bar the
// daily SMA direction is internally consistent: SMA20 < SMA50 → entry_type=Short.
func TestBB7_BTCUSDT_DirectionTrace(t *testing.T) {
	root := projectRootFromCwd()
	dataDir := filepath.Join(root, "tests", "golden", "fixtures", "data")

	hourly := loadOHLCVBars(t, filepath.Join(dataDir, "BTCUSDT-1h.json"))
	daily := loadOHLCVBars(t, filepath.Join(dataDir, "BTCUSDT_1D.json"))

	for _, barIdx := range bb7BTCUSDTDisputedBars {
		probe := buildDirectionProbe(t, hourly, daily, barIdx)
		assertProbeConsistency(t, probe)
	}
}

func buildDirectionProbe(t *testing.T, hourly, daily []ohlcvFixtureBar, barIdx int) directionProbe {
	t.Helper()
	if barIdx >= len(hourly) {
		t.Fatalf("bar index %d out of range (fixture has %d hourly bars)", barIdx, len(hourly))
	}

	h := hourly[barIdx]
	hTime := barTime(h)

	dailyIdx := findLastDailyBarBefore(daily, hTime)
	if dailyIdx < 0 {
		t.Fatalf("no daily bar found for hourly bar %d (%s)", barIdx, hTime.Format("2006-01-02"))
	}

	sma20 := computeSMAFromBars(daily, dailyIdx, 20)
	sma50 := computeSMAFromBars(daily, dailyIdx, 50)
	bullish, entryType := smaBullishDirection(sma20, sma50)

	sma20Display := sma20
	if math.IsNaN(sma20Display) {
		sma20Display = 0
	}
	sma50Display := sma50
	if math.IsNaN(sma50Display) {
		sma50Display = 0
	}

	return directionProbe{
		HourlyBarIdx: barIdx,
		HourlyDate:   hTime.Format("2006-01-02 15:04 MST"),
		HourlyClose:  h.Close,
		DailyBarIdx:  dailyIdx,
		DailyDate:    barTime(daily[dailyIdx]).Format("2006-01-02"),
		SMA20:        sma20Display,
		SMA50:        sma50Display,
		SmaBullish:   bullish,
		EntryType:    entryType,
	}
}

func assertProbeConsistency(t *testing.T, p directionProbe) {
	t.Helper()
	if p.SmaBullish && p.EntryType != "strategy.Long" {
		t.Errorf(
			"bar %d %s close=%.2f daily=%d %s sma20=%.4f sma50=%.4f: sma_bullish=true entry_type=%q",
			p.HourlyBarIdx, p.HourlyDate, p.HourlyClose, p.DailyBarIdx, p.DailyDate, p.SMA20, p.SMA50, p.EntryType,
		)
	}
	if !p.SmaBullish && p.EntryType != "strategy.Short" {
		t.Errorf(
			"bar %d %s close=%.2f daily=%d %s sma20=%.4f sma50=%.4f: sma_bullish=false entry_type=%q",
			p.HourlyBarIdx, p.HourlyDate, p.HourlyClose, p.DailyBarIdx, p.DailyDate, p.SMA20, p.SMA50, p.EntryType,
		)
	}
}
