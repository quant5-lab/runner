package regression

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"testing"
)

func stochV4Script(k int) string {
	return fmt.Sprintf(`//@version=4
study("Stoch Raw K=%d", overlay=false)
plot(stoch(close, high, low, %d), "K")
`, k, k)
}

func smoothedStochV4Script(k, d int) string {
	return fmt.Sprintf(`//@version=4
study("Smoothed Stoch K=%d D=%d", overlay=false)
plot(sma(stoch(close, high, low, %d), %d), "SK")
`, k, d, k, d)
}

func TestStoch_WarmupIsKMinus1(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cases := []struct {
		name string
		k    int
	}{
		{"K5", 5},
		{"K14", 14},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			vals := runIndicator(t, stochV4Script(tc.k), "K", tc.k+5)

			for bar := 0; bar < tc.k-1 && bar < len(vals); bar++ {
				if !math.IsNaN(vals[bar]) {
					t.Errorf("bar %d: expected NaN (warm-up, K=%d), got %.6f",
						bar, tc.k, vals[bar])
				}
			}
			firstValid := tc.k - 1
			if firstValid >= len(vals) {
				t.Fatalf("not enough bars to reach first valid bar %d", firstValid)
			}
			if math.IsNaN(vals[firstValid]) {
				t.Errorf("bar %d (first valid, K=%d): expected a numeric value, got NaN",
					firstValid, tc.k)
			}
		})
	}
}

func TestStoch_OutputRangeZeroToHundred(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	const k = 5
	const barCount = 30
	vals := runIndicator(t, stochV4Script(k), "K", barCount)

	for bar, v := range vals {
		if math.IsNaN(v) {
			continue // warmup — OK
		}
		if v < 0 || v > 100 {
			t.Errorf("bar %d: K=%.6f outside [0, 100]", bar, v)
		}
	}
}

func TestStoch_ArithmeticAtFirstValidBar(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	const period = 5
	lowestLow := 49900.0 + 0              // low[0]
	highestHigh := 50100.0 + (period - 1) // high[period-1]
	closeAtSeed := 50050.0 + (period - 1) // close[period-1]
	expected := 100.0 * (closeAtSeed - lowestLow) / (highestHigh - lowestLow)

	firstValid := period - 1
	vals := runIndicator(t, stochV4Script(period), "K", firstValid+3)

	assertBarValue(t, "stoch K", vals, firstValid, expected)
}

func TestStoch_SmoothedKWarmup(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cases := []struct {
		name string
		k, d int
	}{
		{"K5_D3", 5, 3},   // first valid bar = 5+3-2 = 6
		{"K14_D3", 14, 3}, // first valid bar = 14+3-2 = 15
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			firstValid := tc.k + tc.d - 2
			vals := runIndicator(t, smoothedStochV4Script(tc.k, tc.d), "SK", firstValid+5)

			for bar := 0; bar < firstValid && bar < len(vals); bar++ {
				if !math.IsNaN(vals[bar]) {
					t.Errorf("bar %d: expected NaN (warm-up, K=%d D=%d), got %.6f",
						bar, tc.k, tc.d, vals[bar])
				}
			}
			if firstValid >= len(vals) {
				t.Fatalf("not enough bars to reach first valid bar %d", firstValid)
			}
			if math.IsNaN(vals[firstValid]) {
				t.Errorf("bar %d (K+D-2=%d, K=%d D=%d): expected numeric value, got NaN",
					firstValid, firstValid, tc.k, tc.d)
			}
		})
	}
}

func TestAO_WarmupIsSlowPeriodMinus1(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cases := []struct {
		name       string
		fast, slow int
	}{
		{"fast5_slow34", 5, 34},
		{"fast5_slow10", 5, 10},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			pine := fmt.Sprintf(`//@version=5
indicator("AO fast=%d slow=%d", overlay=false)
ao = ta.sma(hl2, %d) - ta.sma(hl2, %d)
plot(ao, "AO")
`, tc.fast, tc.slow, tc.fast, tc.slow)

			firstValid := tc.slow - 1
			vals := runIndicator(t, pine, "AO", firstValid+5)

			for bar := 0; bar < firstValid && bar < len(vals); bar++ {
				if !math.IsNaN(vals[bar]) {
					t.Errorf("bar %d: expected NaN (AO warm-up, fast=%d slow=%d), got %.6f",
						bar, tc.fast, tc.slow, vals[bar])
				}
			}
			if firstValid >= len(vals) {
				t.Fatalf("not enough bars to reach first valid bar %d", firstValid)
			}
			if math.IsNaN(vals[firstValid]) {
				t.Errorf("bar %d (first valid AO, fast=%d slow=%d): expected numeric value, got NaN",
					firstValid, tc.fast, tc.slow)
			}
		})
	}
}

func TestAO_ArithmeticWithMonotonicData(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	const fast, slow = 5, 10
	// hl2[i]=50000+i → AO is constant (slow-1)/2-(fast-1)/2 from bar slow-1 onward
	expectedAO := float64(slow-1)/2.0 - float64(fast-1)/2.0

	pine := fmt.Sprintf(`//@version=5
indicator("AO Arithmetic", overlay=false)
ao = ta.sma(hl2, %d) - ta.sma(hl2, %d)
plot(ao, "AO")
`, fast, slow)

	const firstValid = slow - 1
	vals := runIndicator(t, pine, "AO", firstValid+5)

	for bar := firstValid; bar < len(vals); bar++ {
		if math.IsNaN(vals[bar]) {
			t.Errorf("bar %d: expected %.4f, got NaN", bar, expectedAO)
			continue
		}
		if math.Abs(vals[bar]-expectedAO) > 1e-9 {
			t.Errorf("bar %d: expected %.4f (constant AO for uniform increment), got %.9f",
				bar, expectedAO, vals[bar])
		}
	}
}

func TestRMA_DirectCloseSource(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cases := []struct {
		name   string
		period int
	}{
		{"period3", 3},
		{"period5", 5},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			pine := fmt.Sprintf(`//@version=5
indicator("RMA Close N=%d", overlay=false)
plot(ta.rma(close, %d), "RMA")
`, tc.period, tc.period)

			n := tc.period
			firstValid := n - 1 // SMA seed bar

			seed := 50050.0 + float64(n-1)/2.0
			alpha := 1.0 / float64(n)
			closeAtN := 50050.0 + float64(n)
			firstWilder := alpha*closeAtN + (1-alpha)*seed

			vals := runIndicator(t, pine, "RMA", n+5)

			for bar := 0; bar < firstValid && bar < len(vals); bar++ {
				if !math.IsNaN(vals[bar]) {
					t.Errorf("bar %d: expected NaN (warm-up, period=%d), got %.6f",
						bar, n, vals[bar])
				}
			}
			if firstValid >= len(vals) {
				t.Fatalf("not enough bars to reach seed bar %d", firstValid)
			}
			assertBarValue(t, "RMA seed", vals, firstValid, seed)

			if n >= len(vals) {
				t.Fatalf("not enough bars to reach Wilder bar %d", n)
			}
			assertBarValue(t, "RMA first Wilder", vals, n, firstWilder)
		})
	}
}

func generateConstantOHLCV(barCount int, open, high, low, closePrice float64, intervalSec int64) string {
	type Bar struct {
		Time   int64   `json:"time"`
		Open   float64 `json:"open"`
		High   float64 `json:"high"`
		Low    float64 `json:"low"`
		Close  float64 `json:"close"`
		Volume float64 `json:"volume"`
	}
	type Envelope struct {
		Timezone string `json:"timezone"`
		Bars     []Bar  `json:"bars"`
	}
	const startTime int64 = 1640000000
	bars := make([]Bar, barCount)
	for i := range bars {
		bars[i] = Bar{
			Time:   startTime + int64(i)*intervalSec,
			Open:   open,
			High:   high,
			Low:    low,
			Close:  closePrice,
			Volume: 100,
		}
	}
	b, _ := json.Marshal(Envelope{Timezone: "UTC", Bars: bars})
	return string(b)
}

func TestStoch_BoundaryExtremes(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	const k = 3

	cases := []struct {
		name      string
		ohlcClose float64
		wantK     float64
	}{
		{"close_equals_high_K100", 110, 100},
		{"close_equals_low_K0", 100, 0},
	}

	pine := fmt.Sprintf(`//@version=4
study("Stoch Boundary K=%d", overlay=false)
plot(stoch(close, high, low, %d), "K")
`, k, k)

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			testDir := t.TempDir()
			stratPath := filepath.Join(testDir, "boundary.pine")
			if err := os.WriteFile(stratPath, []byte(pine), 0644); err != nil {
				t.Fatal(err)
			}
			dataPath := filepath.Join(testDir, "DATA_1D.json")
			data := generateConstantOHLCV(k+5, 105, 110, 100, tc.ohlcClose, 86400)
			if err := os.WriteFile(dataPath, []byte(data), 0644); err != nil {
				t.Fatal(err)
			}
			cwd, _ := os.Getwd()
			projectRoot := filepath.Dir(filepath.Dir(cwd))
			result := compileAndRun(t, stratPath, dataPath, testDir, projectRoot, "DATA", testDir)

			ind, ok := result.Indicators["K"]
			if !ok {
				t.Fatal("K indicator absent from output")
			}
			vals := extractValues(ind.Data)
			firstValid := k - 1
			for bar := firstValid; bar < len(vals); bar++ {
				if math.IsNaN(vals[bar]) {
					t.Errorf("bar %d: expected %.1f, got NaN", bar, tc.wantK)
					continue
				}
				if math.Abs(vals[bar]-tc.wantK) > 1e-9 {
					t.Errorf("bar %d: K=%.9f, want %.1f (close=%.0f high=110 low=100)",
						bar, vals[bar], tc.wantK, tc.ohlcClose)
				}
			}
		})
	}
}

func TestStoch_V4SyntaxSmoothedWarmup(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	const k, d = 14, 3
	firstValid := k + d - 2

	testDir := t.TempDir()
	strategyPath := filepath.Join(testDir, "stoch-v4.pine")
	if err := os.WriteFile(strategyPath, []byte(smoothedStochV4Script(k, d)), 0644); err != nil {
		t.Fatal(err)
	}
	dataPath := filepath.Join(testDir, "DATA_1D.json")
	if err := os.WriteFile(dataPath, []byte(generateTestOHLCV(firstValid+5, 86400)), 0644); err != nil {
		t.Fatal(err)
	}
	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(filepath.Dir(cwd))
	result := compileAndRun(t, strategyPath, dataPath, testDir, projectRoot, "DATA", testDir)

	ind, ok := result.Indicators["SK"]
	if !ok {
		t.Fatal("SK indicator absent from output")
	}
	vals := extractValues(ind.Data)

	for bar := 0; bar < firstValid && bar < len(vals); bar++ {
		if !math.IsNaN(vals[bar]) {
			t.Errorf("bar %d: expected NaN (warm-up, K=%d D=%d), got %.6f",
				bar, k, d, vals[bar])
		}
	}
	if firstValid >= len(vals) {
		t.Fatalf("not enough bars to reach first valid bar %d", firstValid)
	}
	if math.IsNaN(vals[firstValid]) {
		t.Errorf("bar %d (K+D-2=%d): expected numeric value, got NaN", firstValid, firstValid)
	}
}

func generateCustomCloseOHLCV(closes []float64) string {
	type Bar struct {
		Time   int64   `json:"time"`
		Open   float64 `json:"open"`
		High   float64 `json:"high"`
		Low    float64 `json:"low"`
		Close  float64 `json:"close"`
		Volume float64 `json:"volume"`
	}
	type Envelope struct {
		Timezone string `json:"timezone"`
		Bars     []Bar  `json:"bars"`
	}
	const start int64 = 1640000000
	bars := make([]Bar, len(closes))
	for i, c := range closes {
		bars[i] = Bar{
			Time:   start + int64(i)*3600,
			Open:   c,
			High:   c + 1,
			Low:    c - 1,
			Close:  c,
			Volume: 100,
		}
	}
	b, _ := json.Marshal(Envelope{Timezone: "UTC", Bars: bars})
	return string(b)
}

func rsiScript(length int) string {
	return fmt.Sprintf(`//@version=4
study("RSI length=%d", overlay=false)
plot(rsi(close, %d), "RSI")
`, length, length)
}

func TestRSI_WarmupPeriod(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	for _, period := range []int{3, 10, 14} {
		period := period
		t.Run(fmt.Sprintf("period%d", period), func(t *testing.T) {
			vals := runIndicator(t, rsiScript(period), "RSI", period+5)

			for bar := 0; bar < period && bar < len(vals); bar++ {
				if !math.IsNaN(vals[bar]) {
					t.Errorf("bar %d: expected NaN (warm-up, N=%d), got %.6f",
						bar, period, vals[bar])
				}
			}
			if period >= len(vals) {
				t.Fatalf("not enough bars to reach first valid bar %d", period)
			}
			if math.IsNaN(vals[period]) {
				t.Errorf("bar %d (first valid RSI, N=%d): expected numeric, got NaN",
					period, period)
			}
		})
	}
}

func TestRSI_AllGains_EqualsHundred(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	for _, period := range []int{3, 10} {
		period := period
		t.Run(fmt.Sprintf("period%d", period), func(t *testing.T) {
			// generateTestOHLCV produces close[i] = 50050+i, strictly increasing.
			vals := runIndicator(t, rsiScript(period), "RSI", period+10)

			for bar := period; bar < len(vals); bar++ {
				if math.IsNaN(vals[bar]) {
					t.Errorf("bar %d: expected 100 (all-gains), got NaN", bar)
					continue
				}
				if math.Abs(vals[bar]-100) > 1e-9 {
					t.Errorf("bar %d: expected 100.0 (all-gains, N=%d), got %.6f",
						bar, period, vals[bar])
				}
			}
		})
	}
}

func TestRSI_AllLosses_EqualsZero(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	const period = 3
	const barCount = period + 10

	closes := make([]float64, barCount)
	for i := range closes {
		closes[i] = 1000 - float64(i)
	}

	testDir := t.TempDir()
	stratPath := filepath.Join(testDir, "rsi-down.pine")
	if err := os.WriteFile(stratPath, []byte(rsiScript(period)), 0644); err != nil {
		t.Fatal(err)
	}
	dataPath := filepath.Join(testDir, "DATA_1H.json")
	if err := os.WriteFile(dataPath, []byte(generateCustomCloseOHLCV(closes)), 0644); err != nil {
		t.Fatal(err)
	}
	cwd, _ := os.Getwd()
	root := filepath.Dir(filepath.Dir(cwd))
	result := compileAndRun(t, stratPath, dataPath, testDir, root, "DATA", testDir)

	ind, ok := result.Indicators["RSI"]
	if !ok {
		t.Fatal("RSI indicator absent from output")
	}
	vals := extractValues(ind.Data)

	for bar := period; bar < len(vals); bar++ {
		if math.IsNaN(vals[bar]) {
			t.Errorf("bar %d: expected 0 (all-losses), got NaN", bar)
			continue
		}
		if math.Abs(vals[bar]) > 1e-9 {
			t.Errorf("bar %d: expected 0.0 (all-losses, N=%d), got %.6f",
				bar, period, vals[bar])
		}
	}
}

func TestRSI_WilderArithmetic_AlternatingChanges(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	const period = 3
	closes := []float64{100, 102, 101, 103, 102, 104}

	testDir := t.TempDir()
	stratPath := filepath.Join(testDir, "rsi-alt.pine")
	if err := os.WriteFile(stratPath, []byte(rsiScript(period)), 0644); err != nil {
		t.Fatal(err)
	}
	dataPath := filepath.Join(testDir, "DATA_1H.json")
	if err := os.WriteFile(dataPath, []byte(generateCustomCloseOHLCV(closes)), 0644); err != nil {
		t.Fatal(err)
	}
	cwd, _ := os.Getwd()
	root := filepath.Dir(filepath.Dir(cwd))
	result := compileAndRun(t, stratPath, dataPath, testDir, root, "DATA", testDir)

	ind, ok := result.Indicators["RSI"]
	if !ok {
		t.Fatal("RSI indicator absent from output")
	}
	vals := extractValues(ind.Data)

	// Bar 3 first valid: RSI = 500/6.
	expected3 := 500.0 / 6.0
	if 3 >= len(vals) {
		t.Fatal("not enough bars")
	}
	if math.IsNaN(vals[3]) {
		t.Errorf("bar 3: expected %.9f, got NaN", expected3)
	} else if math.Abs(vals[3]-expected3) > 1e-6 {
		t.Errorf("bar 3: expected %.9f (500/6), got %.9f", expected3, vals[3])
	}

	// Bar 4 Wilder step: RSI = 2000/33.
	expected4 := 2000.0 / 33.0
	if 4 >= len(vals) {
		t.Fatal("not enough bars")
	}
	if math.IsNaN(vals[4]) {
		t.Errorf("bar 4: expected %.9f, got NaN", expected4)
	} else if math.Abs(vals[4]-expected4) > 1e-6 {
		t.Errorf("bar 4: expected %.9f (2000/33), got %.9f", expected4, vals[4])
	}
}

func TestAO_NegativeButIncreasing(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	const fast, slow = 5, 34
	const totalBars = slow + 5 + 10
	type Bar struct {
		Time   int64   `json:"time"`
		Open   float64 `json:"open"`
		High   float64 `json:"high"`
		Low    float64 `json:"low"`
		Close  float64 `json:"close"`
		Volume float64 `json:"volume"`
	}
	type Envelope struct {
		Timezone string `json:"timezone"`
		Bars     []Bar  `json:"bars"`
	}
	const baseHl2High = 1000.0
	const dropHl2 = 700.0
	bars := make([]Bar, totalBars)
	for i := range bars {
		var hl2 float64
		if i < slow {
			hl2 = baseHl2High
		} else if i < slow+5 {
			hl2 = dropHl2
		} else {
			hl2 = dropHl2 + float64(i-(slow+5)+1)
		}
		bars[i] = Bar{
			Time:   1640000000 + int64(i)*3600,
			Open:   hl2,
			High:   hl2 + 0.5,
			Low:    hl2 - 0.5,
			Close:  hl2,
			Volume: 100,
		}
	}
	env := Envelope{Timezone: "UTC", Bars: bars}
	data, _ := json.Marshal(env)

	pine := fmt.Sprintf(`//@version=4
study("AO negative but increasing", overlay=false)
ao = (sma(hl2, %d) - sma(hl2, %d)) * 1000
plot(ao, "AO")
`, fast, slow)

	testDir := t.TempDir()
	stratPath := filepath.Join(testDir, "ao_neg.pine")
	if err := os.WriteFile(stratPath, []byte(pine), 0644); err != nil {
		t.Fatal(err)
	}
	dataPath := filepath.Join(testDir, "DATA_1H.json")
	if err := os.WriteFile(dataPath, data, 0644); err != nil {
		t.Fatal(err)
	}
	cwd, _ := os.Getwd()
	root := filepath.Dir(filepath.Dir(cwd))
	result := compileAndRun(t, stratPath, dataPath, testDir, root, "DATA", testDir)

	ind, ok := result.Indicators["AO"]
	if !ok {
		t.Fatal("AO indicator absent from output")
	}
	vals := extractValues(ind.Data)

	firstValid := slow - 1
	// Recovery starts at bar slow+5. Confirm AO is negative but increasing there.
	recoveryStart := slow + 5
	if recoveryStart >= len(vals) {
		t.Fatalf("not enough bars to reach recovery start %d (len=%d)", recoveryStart, len(vals))
	}

	negativeCount := 0
	increasingCount := 0
	for bar := recoveryStart; bar < len(vals) && bar <= recoveryStart+5; bar++ {
		if bar <= firstValid || math.IsNaN(vals[bar]) || math.IsNaN(vals[bar-1]) {
			continue
		}
		if vals[bar] < 0 {
			negativeCount++
		}
		if vals[bar] > vals[bar-1] {
			increasingCount++
		}
	}
	if negativeCount == 0 {
		t.Errorf("AO was never negative during recovery window — test setup invalid (no bearish-but-recovering case)")
	}
	if increasingCount == 0 {
		t.Errorf("AO never increased (ao > ao[1]) during recovery — momentum direction comparison broken")
	}
}
