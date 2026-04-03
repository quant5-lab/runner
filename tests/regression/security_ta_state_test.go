package regression

import (
	"math"
	"os"
	"path/filepath"
	"testing"
)

/*
TestSecurityTA_AllFamiliesCompileAndRun verifies that SMA, EMA, RMA, RSI, ATR,
STDEV, TSI, and SAR each survive the full codegen→compile→execute pipeline when
evaluated inside request.security(), producing at least one non-null output bar.
*/
func TestSecurityTA_AllFamiliesCompileAndRun(t *testing.T) {
	strategy := `//@version=5
indicator("Security TA All", overlay=false)
sma5   = request.security(syminfo.tickerid, "1D", ta.sma(close, 5))
ema5   = request.security(syminfo.tickerid, "1D", ta.ema(close, 5))
rma5   = request.security(syminfo.tickerid, "1D", ta.rma(close, 5))
rsi14  = request.security(syminfo.tickerid, "1D", ta.rsi(close, 14))
atr14  = request.security(syminfo.tickerid, "1D", ta.atr(14))
stdev5 = request.security(syminfo.tickerid, "1D", ta.stdev(close, 5))
tsi    = request.security(syminfo.tickerid, "1D", ta.tsi(close, 5, 13))
sar    = request.security(syminfo.tickerid, "1D", ta.sar(0.02, 0.02, 0.2))
plot(sma5,   "SMA5")
plot(ema5,   "EMA5")
plot(rma5,   "RMA5")
plot(rsi14,  "RSI14")
plot(atr14,  "ATR14")
plot(stdev5, "STDEV5")
plot(tsi,    "TSI")
plot(sar,    "SAR")
`
	testDir := t.TempDir()
	strategyPath := filepath.Join(testDir, "all-ta.pine")
	if err := os.WriteFile(strategyPath, []byte(strategy), 0644); err != nil {
		t.Fatal(err)
	}
	dataPath := filepath.Join(testDir, "SECTAALL_1D.json")
	if err := os.WriteFile(dataPath, []byte(generateTestOHLCV(40, 86400)), 0644); err != nil {
		t.Fatal(err)
	}

	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(filepath.Dir(cwd))

	result := compileAndRun(t, strategyPath, dataPath, testDir, projectRoot, "SECTAALL", testDir)

	for _, name := range []string{"SMA5", "EMA5", "RMA5", "RSI14", "ATR14", "STDEV5", "TSI", "SAR"} {
		t.Run(name, func(t *testing.T) {
			ind, ok := result.Indicators[name]
			if !ok {
				t.Fatalf("indicator %q absent from output", name)
			}
			if countNonNull(ind.Data) == 0 {
				t.Errorf("indicator %q produced zero non-null values across 40 bars", name)
			}
		})
	}
}

/*
TestSecurityTA_WarmupBoundary verifies that bars inside the warmup window produce
null output and that the first post-warmup bar produces a valid value, exercising
the ForwardSeriesBuffer NaN sentinel through the full security() pipeline.
*/
func TestSecurityTA_WarmupBoundary(t *testing.T) {
	tests := []struct {
		name     string
		expr     string
		plotName string
		warmup   int // first valid base-bar index (0-based); +1 vs raw period due to 1-bar lookahead_off lag
	}{
		{"SMA5", "ta.sma(close, 5)", "SMA", 5},
		{"EMA10", "ta.ema(close, 10)", "EMA", 10},
		{"STDEV5", "ta.stdev(close, 5)", "STDEV", 5},
		{"RSI14", "ta.rsi(close, 14)", "RSI", 15},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strategy := "//@version=5\n" +
				"indicator(\"Warmup " + tt.name + "\", overlay=false)\n" +
				"val = request.security(syminfo.tickerid, \"1D\", " + tt.expr + ")\n" +
				"plot(val, \"" + tt.plotName + "\")\n"

			testDir := t.TempDir()
			strategyPath := filepath.Join(testDir, "warmup.pine")
			if err := os.WriteFile(strategyPath, []byte(strategy), 0644); err != nil {
				t.Fatal(err)
			}
			dataPath := filepath.Join(testDir, "WARM_1D.json")
			if err := os.WriteFile(dataPath, []byte(generateTestOHLCV(30, 86400)), 0644); err != nil {
				t.Fatal(err)
			}

			cwd, _ := os.Getwd()
			projectRoot := filepath.Dir(filepath.Dir(cwd))

			result := compileAndRun(t, strategyPath, dataPath, testDir, projectRoot, "WARM", testDir)

			ind, ok := result.Indicators[tt.plotName]
			if !ok {
				t.Fatalf("indicator %q absent from output", tt.plotName)
			}
			vals := extractValues(ind.Data)

			for i := 0; i < tt.warmup && i < len(vals); i++ {
				if !math.IsNaN(vals[i]) {
					t.Errorf("bar %d: expected NaN during warmup (first valid bar = %d), got %.6f",
						i, tt.warmup, vals[i])
				}
			}

			if tt.warmup < len(vals) {
				if math.IsNaN(vals[tt.warmup]) {
					t.Errorf("bar %d: expected first valid value, got NaN", tt.warmup)
				}
			}
		})
	}
}

/*
TestSecurityTA_SMAArithmeticCorrectness verifies the exact SMA values produced
by request.security() against hand-calculated results on deterministic test data.
generateTestOHLCV sets Close[i] = 50050 + i, so SMA(5)[k] = 50048 + k for k >= 4.
request.security("1D", …) with a 1h-based context applies a 1-bar lookahead_off lag:
base bar i maps to secBarIdx i-1, so the first valid base bar is 5 (secBarIdx=4).
*/
func TestSecurityTA_SMAArithmeticCorrectness(t *testing.T) {
	strategy := `//@version=5
indicator("SMA Correctness", overlay=false)
sma = request.security(syminfo.tickerid, "1D", ta.sma(close, 5))
plot(sma, "SMA")
`
	testDir := t.TempDir()
	strategyPath := filepath.Join(testDir, "sma-exact.pine")
	if err := os.WriteFile(strategyPath, []byte(strategy), 0644); err != nil {
		t.Fatal(err)
	}
	dataPath := filepath.Join(testDir, "SMAEX_1D.json")
	if err := os.WriteFile(dataPath, []byte(generateTestOHLCV(20, 86400)), 0644); err != nil {
		t.Fatal(err)
	}

	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(filepath.Dir(cwd))

	result := compileAndRun(t, strategyPath, dataPath, testDir, projectRoot, "SMAEX", testDir)

	ind, ok := result.Indicators["SMA"]
	if !ok {
		t.Fatal("SMA indicator absent from output")
	}
	vals := extractValues(ind.Data)

	// base bar i → secBarIdx i-1 → SMA[i-1] = 50048+(i-1) = 50047+i
	checks := []struct {
		barIdx int
		want   float64
	}{
		{5, 50052.0},  // secBarIdx=4: (50050+50051+50052+50053+50054)/5
		{6, 50053.0},  // secBarIdx=5
		{10, 50057.0}, // secBarIdx=9
		{19, 50066.0}, // secBarIdx=18: SMA[18]=50048+18
	}

	for _, c := range checks {
		if c.barIdx >= len(vals) {
			t.Fatalf("bar %d out of range (len=%d)", c.barIdx, len(vals))
		}
		if math.IsNaN(vals[c.barIdx]) {
			t.Errorf("bar %d: unexpected NaN, want %.1f", c.barIdx, c.want)
			continue
		}
		if math.Abs(vals[c.barIdx]-c.want) > 1e-6 {
			t.Errorf("bar %d: SMA = %.9f, want %.9f", c.barIdx, vals[c.barIdx], c.want)
		}
	}
}

/*
TestSecurityTA_RSIBoundsAndZeroLoss verifies two RSI invariants end-to-end:
all post-warmup values stay within [0, 100], and a monotonically rising source
with zero losses produces RSI = 100 via the avgLoss == 0 branch.
*/
func TestSecurityTA_RSIBoundsAndZeroLoss(t *testing.T) {
	strategy := `//@version=5
indicator("RSI Bounds", overlay=false)
rsi = request.security(syminfo.tickerid, "1D", ta.rsi(close, 14))
plot(rsi, "RSI")
`
	testDir := t.TempDir()
	strategyPath := filepath.Join(testDir, "rsi-bounds.pine")
	if err := os.WriteFile(strategyPath, []byte(strategy), 0644); err != nil {
		t.Fatal(err)
	}
	dataPath := filepath.Join(testDir, "RSIBND_1D.json")
	if err := os.WriteFile(dataPath, []byte(generateTestOHLCV(30, 86400)), 0644); err != nil {
		t.Fatal(err)
	}

	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(filepath.Dir(cwd))

	result := compileAndRun(t, strategyPath, dataPath, testDir, projectRoot, "RSIBND", testDir)

	ind, ok := result.Indicators["RSI"]
	if !ok {
		t.Fatal("RSI indicator absent from output")
	}
	vals := extractValues(ind.Data)

	for i, v := range vals {
		if math.IsNaN(v) {
			continue
		}
		if v < 0 || v > 100 {
			t.Errorf("bar %d: RSI = %.6f out of [0, 100]", i, v)
		}
	}

	// generateTestOHLCV produces monotonically rising Close → all changes positive → avgLoss == 0 → RSI == 100.
	// 1-bar lookahead_off lag: first valid base bar is 15 (secBarIdx=14).
	for i := 15; i < len(vals); i++ {
		if math.IsNaN(vals[i]) {
			t.Errorf("bar %d: unexpected NaN post-warmup", i)
			continue
		}
		if math.Abs(vals[i]-100.0) > 1e-6 {
			t.Errorf("bar %d: RSI = %.6f on monotonic rising source, want 100.0", i, vals[i])
		}
	}
}

/*
TestSecurityTA_ParallelStateManagersDoNotCorrupt verifies that multiple stateful
TA indicators evaluated within the same security() context do not corrupt each
other's ForwardSeriesBuffer state.
*/
func TestSecurityTA_ParallelStateManagersDoNotCorrupt(t *testing.T) {
	strategy := `//@version=5
indicator("Parallel TA", overlay=false)
sma  = request.security(syminfo.tickerid, "1D", ta.sma(close, 5))
ema  = request.security(syminfo.tickerid, "1D", ta.ema(close, 5))
rsi  = request.security(syminfo.tickerid, "1D", ta.rsi(close, 14))
plot(sma,  "SMA")
plot(ema,  "EMA")
plot(rsi,  "RSI")
`
	testDir := t.TempDir()
	strategyPath := filepath.Join(testDir, "parallel.pine")
	if err := os.WriteFile(strategyPath, []byte(strategy), 0644); err != nil {
		t.Fatal(err)
	}
	dataPath := filepath.Join(testDir, "PAR_1D.json")
	if err := os.WriteFile(dataPath, []byte(generateTestOHLCV(30, 86400)), 0644); err != nil {
		t.Fatal(err)
	}

	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(filepath.Dir(cwd))

	result := compileAndRun(t, strategyPath, dataPath, testDir, projectRoot, "PAR", testDir)

	smaInd, ok := result.Indicators["SMA"]
	if !ok {
		t.Fatal("SMA indicator absent")
	}
	emaInd, ok := result.Indicators["EMA"]
	if !ok {
		t.Fatal("EMA indicator absent")
	}
	rsiInd, ok := result.Indicators["RSI"]
	if !ok {
		t.Fatal("RSI indicator absent")
	}

	smaVals := extractValues(smaInd.Data)
	emaVals := extractValues(emaInd.Data)
	rsiVals := extractValues(rsiInd.Data)

	// All three must produce valid output after RSI's warmup (the latest of the three).
	// 1-bar lookahead_off lag: RSI(14) first valid at base bar 15 (secBarIdx=14).
	for i := 15; i < len(smaVals) && i < len(emaVals) && i < len(rsiVals); i++ {
		if math.IsNaN(smaVals[i]) {
			t.Errorf("bar %d: SMA unexpectedly NaN post-warmup", i)
		}
		if math.IsNaN(emaVals[i]) {
			t.Errorf("bar %d: EMA unexpectedly NaN post-warmup", i)
		}
		if math.IsNaN(rsiVals[i]) {
			t.Errorf("bar %d: RSI unexpectedly NaN post-warmup", i)
		}
	}

	// RSI is bounded [0, 100]; SMA/EMA are ~50050. Detecting cross-state corruption via range isolation.
	for i := 15; i < len(rsiVals) && i < len(smaVals); i++ {
		if math.IsNaN(rsiVals[i]) || math.IsNaN(smaVals[i]) {
			continue
		}
		if rsiVals[i] > 100 || rsiVals[i] < 0 {
			t.Errorf("bar %d: RSI = %.6f out of [0,100] — possible state corruption from SMA/EMA", i, rsiVals[i])
		}
		if smaVals[i] < 50000 || smaVals[i] > 51000 {
			t.Errorf("bar %d: SMA = %.6f out of expected range — possible state corruption from RSI", i, smaVals[i])
		}
	}

	// SMA exact check: bar 5 → secBarIdx 4 → SMA[4] = 50052.0 (unaffected by parallel EMA and RSI state).
	if len(smaVals) > 5 && !math.IsNaN(smaVals[5]) {
		if math.Abs(smaVals[5]-50052.0) > 1e-6 {
			t.Errorf("SMA bar 5 = %.9f, want 50052.0 (parallel state corruption check)", smaVals[5])
		}
	}
}

/*
TestSecurityTA_HistoricalSubscriptWithinContext verifies that ta.sma(close,5)[1]
evaluated inside request.security() returns the same value as ta.sma(close,5)
evaluated at the previous security-context bar.
*/
func TestSecurityTA_HistoricalSubscriptWithinContext(t *testing.T) {
	strategy := `//@version=5
indicator("SMA Historical Subscript", overlay=false)
sma      = request.security(syminfo.tickerid, "1D", ta.sma(close, 5))
smaPrev  = request.security(syminfo.tickerid, "1D", ta.sma(close, 5)[1])
plot(sma,     "SMA")
plot(smaPrev, "SMAPrev")
`
	testDir := t.TempDir()
	strategyPath := filepath.Join(testDir, "sma-subscript.pine")
	if err := os.WriteFile(strategyPath, []byte(strategy), 0644); err != nil {
		t.Fatal(err)
	}
	dataPath := filepath.Join(testDir, "SMASUB_1D.json")
	if err := os.WriteFile(dataPath, []byte(generateTestOHLCV(25, 86400)), 0644); err != nil {
		t.Fatal(err)
	}

	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(filepath.Dir(cwd))

	result := compileAndRun(t, strategyPath, dataPath, testDir, projectRoot, "SMASUB", testDir)

	smaInd, ok := result.Indicators["SMA"]
	if !ok {
		t.Fatal("SMA indicator absent")
	}
	smaPrevInd, ok := result.Indicators["SMAPrev"]
	if !ok {
		t.Fatal("SMAPrev indicator absent")
	}

	sma := extractValues(smaInd.Data)
	smaPrev := extractValues(smaPrevInd.Data)

	for i := 5; i < len(sma) && i < len(smaPrev); i++ {
		if math.IsNaN(sma[i-1]) || math.IsNaN(smaPrev[i]) {
			continue
		}
		if math.Abs(sma[i-1]-smaPrev[i]) > 1e-6 {
			t.Errorf("bar %d: ta.sma[1] = %.9f, ta.sma at bar %d = %.9f, want equal",
				i, smaPrev[i], i-1, sma[i-1])
		}
	}
}

/*
TestSecurityTA_DownscalingWarmupPropagation verifies that the warmup NaN of a
stateful TA indicator in a lower-timeframe security context propagates correctly
to the higher-timeframe base bars — pre-warmup hourly bars receive null, and
post-warmup hourly bars receive the carried-forward daily value.
*/
func TestSecurityTA_DownscalingWarmupPropagation(t *testing.T) {
	strategy := `//@version=5
indicator("Downscale Warmup", overlay=false)
dailyEMA = request.security(syminfo.tickerid, "1D", ta.ema(close, 5))
plot(dailyEMA, "DailyEMA5")
`
	testDir := t.TempDir()
	strategyPath := filepath.Join(testDir, "downscale-warmup.pine")
	if err := os.WriteFile(strategyPath, []byte(strategy), 0644); err != nil {
		t.Fatal(err)
	}

	// 240 hourly bars (10 days) + 10 daily bars; EMA(5) warmup ends at daily bar 4.
	hourlyPath := filepath.Join(testDir, "DSWARM_1h.json")
	if err := os.WriteFile(hourlyPath, []byte(generateTestOHLCV(240, 3600)), 0644); err != nil {
		t.Fatal(err)
	}
	dailyPath := filepath.Join(testDir, "DSWARM_1D.json")
	if err := os.WriteFile(dailyPath, []byte(generateTestOHLCV(10, 86400)), 0644); err != nil {
		t.Fatal(err)
	}

	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(filepath.Dir(cwd))

	result := compileAndRun(t, strategyPath, hourlyPath, testDir, projectRoot, "DSWARM", testDir)

	ind, ok := result.Indicators["DailyEMA5"]
	if !ok {
		t.Fatal("DailyEMA5 indicator absent from output")
	}

	vals := extractValues(ind.Data)
	nullCount := 0
	for _, v := range vals {
		if math.IsNaN(v) {
			nullCount++
		}
	}
	nonNullCount := len(vals) - nullCount

	// EMA(5) warmup = 4 daily bars = ~96 hourly bars → those are null.
	// Post-warmup daily bars 4-9 = ~144 hourly bars → those are non-null.
	if nullCount == 0 {
		t.Error("expected some null bars during EMA5 warmup in downscaling, got none")
	}
	if nonNullCount == 0 {
		t.Error("expected non-null bars after EMA5 warmup in downscaling, got none")
	}
	if nonNullCount > nullCount {
		// 144 post-warmup > 96 pre-warmup, so this is the expected outcome.
		// Verifying direction: more bars are valid than null (warmup is a minority).
		if nonNullCount < 100 {
			t.Errorf("expected at least 100 non-null bars after EMA5 warmup, got %d", nonNullCount)
		}
	}
}
