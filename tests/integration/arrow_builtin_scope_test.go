//go:build integration

package integration

import (
	"math"
	"testing"

	"github.com/quant5-lab/runner/tests/util"
)

/* Bars with distinct OHLCV for unambiguous derived-price validation */
func arrowScopeBars() []map[string]interface{} {
	baseTime := int64(1700000000)
	return []map[string]interface{}{
		{"time": baseTime, "open": 100.0, "high": 110.0, "low": 90.0, "close": 105.0, "volume": 1000.0},
		{"time": baseTime + 3600, "open": 106.0, "high": 112.0, "low": 94.0, "close": 108.0, "volume": 1100.0},
		{"time": baseTime + 7200, "open": 109.0, "high": 118.0, "low": 96.0, "close": 114.0, "volume": 1200.0},
		{"time": baseTime + 10800, "open": 115.0, "high": 120.0, "low": 102.0, "close": 110.0, "volume": 1300.0},
		{"time": baseTime + 14400, "open": 111.0, "high": 116.0, "low": 98.0, "close": 107.0, "volume": 1400.0},
		{"time": baseTime + 18000, "open": 108.0, "high": 122.0, "low": 100.0, "close": 119.0, "volume": 1500.0},
		{"time": baseTime + 21600, "open": 120.0, "high": 125.0, "low": 104.0, "close": 121.0, "volume": 1600.0},
		{"time": baseTime + 25200, "open": 122.0, "high": 130.0, "low": 106.0, "close": 124.0, "volume": 1700.0},
	}
}

func TestArrowScope_DerivedPrices(t *testing.T) {
	t.Parallel()
	bars := arrowScopeBars()

	tests := []struct {
		name     string
		pine     string
		plotName string
		formula  func(o, h, l, c float64) float64
	}{
		{
			name: "ohlc4 in arrow body",
			pine: `//@version=5
indicator("Arrow ohlc4")
getOhlc4() => ohlc4
plot(getOhlc4(), "result")
`,
			plotName: "result",
			formula:  func(o, h, l, c float64) float64 { return (o + h + l + c) / 4 },
		},
		{
			name: "hl2 in arrow body",
			pine: `//@version=5
indicator("Arrow hl2")
getHl2() => hl2
plot(getHl2(), "result")
`,
			plotName: "result",
			formula:  func(_, h, l, _ float64) float64 { return (h + l) / 2 },
		},
		{
			name: "hlc3 in arrow body",
			pine: `//@version=5
indicator("Arrow hlc3")
getHlc3() => hlc3
plot(getHlc3(), "result")
`,
			plotName: "result",
			formula:  func(_, h, l, c float64) float64 { return (h + l + c) / 3 },
		},
		{
			name: "hlcc4 in arrow body",
			pine: `//@version=5
indicator("Arrow hlcc4")
getHlcc4() => hlcc4
plot(getHlcc4(), "result")
`,
			plotName: "result",
			formula:  func(_, h, l, c float64) float64 { return (h + l + c + c) / 4 },
		},
	}

	exec := util.NewPineExecutor(t)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			output := exec.ExecuteScriptWithCustomData(t, "arrow-derived", tt.pine, bars)
			values := exec.ExtractPlotValues(t, output, tt.plotName)

			if len(values) < len(bars) {
				t.Fatalf("expected %d values, got %d", len(bars), len(values))
			}

			for i, bar := range bars {
				expected := tt.formula(bar["open"].(float64), bar["high"].(float64), bar["low"].(float64), bar["close"].(float64))
				if math.Abs(values[i]-expected) > 0.001 {
					t.Errorf("bar[%d]: got %f, want %f", i, values[i], expected)
				}
			}
		})
	}
}

func TestArrowScope_OHLCVBuiltins(t *testing.T) {
	t.Parallel()
	bars := arrowScopeBars()

	pine := `//@version=5
indicator("Arrow OHLCV")
getClose() => close
getHigh() => high
getLow() => low
getOpen() => open
getVolume() => volume
plot(getClose(), "close")
plot(getHigh(), "high")
plot(getLow(), "low")
plot(getOpen(), "open")
plot(getVolume(), "volume")
`
	exec := util.NewPineExecutor(t)
	output := exec.ExecuteScriptWithCustomData(t, "arrow-ohlcv", pine, bars)

	fields := []struct {
		plot string
		key  string
	}{
		{"close", "close"},
		{"high", "high"},
		{"low", "low"},
		{"open", "open"},
		{"volume", "volume"},
	}

	for _, f := range fields {
		t.Run(f.plot, func(t *testing.T) {
			values := exec.ExtractPlotValues(t, output, f.plot)
			if len(values) < len(bars) {
				t.Fatalf("expected %d values, got %d", len(bars), len(values))
			}
			for i, bar := range bars {
				expected := bar[f.key].(float64)
				if values[i] != expected {
					t.Errorf("bar[%d]: got %f, want %f", i, values[i], expected)
				}
			}
		})
	}
}

func TestArrowScope_DerivedPriceArithmetic(t *testing.T) {
	t.Parallel()
	bars := arrowScopeBars()

	pine := `//@version=5
indicator("Arrow Derived Arithmetic")
spread() => ohlc4 - hl2
plot(spread(), "spread")
`
	exec := util.NewPineExecutor(t)
	output := exec.ExecuteScriptWithCustomData(t, "arrow-derived-arith", pine, bars)
	values := exec.ExtractPlotValues(t, output, "spread")

	if len(values) < len(bars) {
		t.Fatalf("expected %d values, got %d", len(bars), len(values))
	}

	for i, bar := range bars {
		o, h, l, c := bar["open"].(float64), bar["high"].(float64), bar["low"].(float64), bar["close"].(float64)
		ohlc4 := (o + h + l + c) / 4
		hl2 := (h + l) / 2
		expected := ohlc4 - hl2
		if math.Abs(values[i]-expected) > 0.001 {
			t.Errorf("bar[%d]: got %f, want %f", i, values[i], expected)
		}
	}
}

func TestArrowScope_MultiStatementBody(t *testing.T) {
	t.Parallel()
	bars := arrowScopeBars()

	pine := `//@version=5
indicator("Arrow Multi Statement")
avgPrice(weight) =>
    mid = hl2
    weighted = mid * weight + ohlc4 * (1.0 - weight)
    weighted
plot(avgPrice(0.5), "weighted")
`
	exec := util.NewPineExecutor(t)
	output := exec.ExecuteScriptWithCustomData(t, "arrow-multi-stmt", pine, bars)
	values := exec.ExtractPlotValues(t, output, "weighted")

	if len(values) < len(bars) {
		t.Fatalf("expected %d values, got %d", len(bars), len(values))
	}

	for i, bar := range bars {
		o, h, l, c := bar["open"].(float64), bar["high"].(float64), bar["low"].(float64), bar["close"].(float64)
		hl2 := (h + l) / 2
		ohlc4 := (o + h + l + c) / 4
		expected := hl2*0.5 + ohlc4*0.5
		if math.Abs(values[i]-expected) > 0.001 {
			t.Errorf("bar[%d]: got %f, want %f", i, values[i], expected)
		}
	}
}

func TestArrowScope_TrueRangeAccessPatterns(t *testing.T) {
	t.Parallel()
	bars := arrowScopeBars()

	tests := []struct {
		name       string
		pine       string
		arrowPlot  string
		directPlot string
		minBar     int
	}{
		{
			name: "current bar",
			pine: `//@version=5
indicator("Arrow TR current")
getTr() => tr
plot(getTr(), "arrow")
plot(tr, "direct")
`,
			arrowPlot:  "arrow",
			directPlot: "direct",
			minBar:     1,
		},
		{
			name: "historical offset 1",
			pine: `//@version=5
indicator("Arrow TR hist")
getPrevTr() => tr[1]
plot(getPrevTr(), "arrow")
plot(tr[1], "direct")
`,
			arrowPlot:  "arrow",
			directPlot: "direct",
			minBar:     2,
		},
	}

	exec := util.NewPineExecutor(t)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			output := exec.ExecuteScriptWithCustomData(t, "arrow-tr-"+tt.name, tt.pine, bars)
			arrowVals := exec.ExtractPlotValues(t, output, tt.arrowPlot)
			directVals := exec.ExtractPlotValues(t, output, tt.directPlot)

			if len(arrowVals) < len(bars) {
				t.Fatalf("expected %d values, got %d", len(bars), len(arrowVals))
			}

			for i := tt.minBar; i < len(bars); i++ {
				if math.Abs(arrowVals[i]-directVals[i]) > 0.001 {
					t.Errorf("bar[%d]: arrow=%f, direct=%f", i, arrowVals[i], directVals[i])
				}
			}
		})
	}
}
