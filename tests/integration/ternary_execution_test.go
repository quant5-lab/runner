package integration

import (
	"testing"

	"github.com/quant5-lab/runner/tests/util"
)

func TestTernaryExecution(t *testing.T) {
	pineScript := `//@version=5
indicator("Ternary Test", overlay=false)

close_avg = ta.sma(close, 20)
signal = close > close_avg ? 1 : 0

plot(signal, "signal", color=color.blue)
`

	testData := []map[string]interface{}{
		{"time": 1700000000, "open": 100.0, "high": 105.0, "low": 95.0, "close": 110.0, "volume": 1000.0},
		{"time": 1700003600, "open": 110.0, "high": 115.0, "low": 105.0, "close": 112.0, "volume": 1100.0},
		{"time": 1700007200, "open": 112.0, "high": 117.0, "low": 107.0, "close": 114.0, "volume": 1200.0},
		{"time": 1700010800, "open": 114.0, "high": 119.0, "low": 109.0, "close": 116.0, "volume": 1300.0},
		{"time": 1700014400, "open": 116.0, "high": 121.0, "low": 111.0, "close": 118.0, "volume": 1400.0},
		{"time": 1700018000, "open": 118.0, "high": 123.0, "low": 113.0, "close": 120.0, "volume": 1500.0},
		{"time": 1700021600, "open": 120.0, "high": 125.0, "low": 115.0, "close": 122.0, "volume": 1600.0},
		{"time": 1700025200, "open": 122.0, "high": 127.0, "low": 117.0, "close": 124.0, "volume": 1700.0},
		{"time": 1700028800, "open": 124.0, "high": 129.0, "low": 119.0, "close": 126.0, "volume": 1800.0},
		{"time": 1700032400, "open": 126.0, "high": 131.0, "low": 121.0, "close": 128.0, "volume": 1900.0},
		{"time": 1700036000, "open": 128.0, "high": 133.0, "low": 123.0, "close": 130.0, "volume": 2000.0},
		{"time": 1700039600, "open": 130.0, "high": 135.0, "low": 125.0, "close": 132.0, "volume": 2100.0},
		{"time": 1700043200, "open": 132.0, "high": 137.0, "low": 127.0, "close": 134.0, "volume": 2200.0},
		{"time": 1700046800, "open": 134.0, "high": 139.0, "low": 129.0, "close": 136.0, "volume": 2300.0},
		{"time": 1700050400, "open": 136.0, "high": 141.0, "low": 131.0, "close": 138.0, "volume": 2400.0},
		{"time": 1700054000, "open": 138.0, "high": 143.0, "low": 133.0, "close": 140.0, "volume": 2500.0},
		{"time": 1700057600, "open": 140.0, "high": 145.0, "low": 135.0, "close": 142.0, "volume": 2600.0},
		{"time": 1700061200, "open": 142.0, "high": 147.0, "low": 137.0, "close": 144.0, "volume": 2700.0},
		{"time": 1700064800, "open": 144.0, "high": 149.0, "low": 139.0, "close": 146.0, "volume": 2800.0},
		{"time": 1700068400, "open": 146.0, "high": 151.0, "low": 141.0, "close": 148.0, "volume": 2900.0},
		{"time": 1700072000, "open": 148.0, "high": 153.0, "low": 143.0, "close": 100.0, "volume": 3000.0},
		{"time": 1700075600, "open": 100.0, "high": 105.0, "low": 95.0, "close": 102.0, "volume": 3100.0},
		{"time": 1700079200, "open": 102.0, "high": 107.0, "low": 97.0, "close": 104.0, "volume": 3200.0},
		{"time": 1700082800, "open": 104.0, "high": 109.0, "low": 99.0, "close": 106.0, "volume": 3300.0},
	}

	exec := util.NewPineExecutor(t)
	result := exec.ExecuteScriptWithCustomData(t, "ternary-test", pineScript, testData)

	signalValues := exec.ExtractPlotValues(t, result, "signal")

	if len(signalValues) < 24 {
		t.Fatalf("Expected at least 24 signal values, got %d", len(signalValues))
	}

	if signalValues[20] != 0.0 {
		t.Errorf("Bar 20: expected signal=0 (close below SMA), got %v", signalValues[20])
	}

	if signalValues[19] != 1.0 {
		t.Errorf("Bar 19: expected signal=1 (close above SMA), got %v", signalValues[19])
	}
}
