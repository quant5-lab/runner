package regression

import (
	"math"
	"os"
	"path/filepath"
	"testing"
)

/*
TestSecurityTA_EMAArithmeticCorrectness verifies the exact EMA values produced
by request.security() against hand-calculated results on deterministic test data.

generateTestOHLCV: close[s] = 50050 + s.
EMA(3): alpha = 2/(3+1) = 0.5

	secBar 0: seed = 50050                          (warmup, NaN output)
	secBar 1: (50050×1 + 50051)/2 = 50050.5        (warmup, NaN output)
	secBar 2: (50050.5×2 + 50052)/3 = 50051.0       ← first valid
	secBar n ≥ 2: EMA[n] = 50049 + n               (closed form on unit-step input with α=0.5)

1-bar lookahead_off lag: base bar k → secBarIdx k−1; first valid base bar = 3.
EMA[base k] = 50048 + k, k ≥ 3.
*/
func TestSecurityTA_EMAArithmeticCorrectness(t *testing.T) {
	strategy := `//@version=5
indicator("EMA Correctness", overlay=false)
ema = request.security(syminfo.tickerid, "1D", ta.ema(close, 3))
plot(ema, "EMA")
`
	testDir := t.TempDir()
	strategyPath := filepath.Join(testDir, "ema-exact.pine")
	if err := os.WriteFile(strategyPath, []byte(strategy), 0644); err != nil {
		t.Fatal(err)
	}
	dataPath := filepath.Join(testDir, "EMAEX_1D.json")
	if err := os.WriteFile(dataPath, []byte(generateTestOHLCV(20, 86400)), 0644); err != nil {
		t.Fatal(err)
	}

	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(filepath.Dir(cwd))

	result := compileAndRun(t, strategyPath, dataPath, testDir, projectRoot, "EMAEX", testDir)

	ind, ok := result.Indicators["EMA"]
	if !ok {
		t.Fatal("EMA indicator absent from output")
	}
	vals := extractValues(ind.Data)

	// base bar k → secBar k-1 → EMA[secBar k-1] = 50049+(k-1) = 50048+k, k ≥ 3.
	checks := []struct {
		barIdx int
		want   float64
	}{
		{3, 50051.0},  // secBar 2: warmup boundary, first valid
		{4, 50052.0},  // secBar 3: first smoothed bar (α=0.5)
		{10, 50058.0}, // secBar 9
		{19, 50067.0}, // secBar 18
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
			t.Errorf("bar %d: EMA(3) = %.9f, want %.9f", c.barIdx, vals[c.barIdx], c.want)
		}
	}

	for i := 0; i < 3 && i < len(vals); i++ {
		if !math.IsNaN(vals[i]) {
			t.Errorf("bar %d: expected NaN during EMA(3) warmup, got %.6f", i, vals[i])
		}
	}
}

/*
TestSecurityTA_RMAArithmeticCorrectness verifies the exact RMA values at the
warmup boundary and first Wilder-smoothed bar through the full pipeline.

generateTestOHLCV: close[s] = 50050 + s.
RMA(3): α = 1/3

	secBar 2 (warmup boundary): identical to SMA running avg = 50051.0
	secBar 3 (first Wilder smooth): (1/3)×50053 + (2/3)×50051.0 = 50051 + 2/3

1-bar lookahead_off lag: first valid base bar = 3.
*/
func TestSecurityTA_RMAArithmeticCorrectness(t *testing.T) {
	strategy := `//@version=5
indicator("RMA Correctness", overlay=false)
rma = request.security(syminfo.tickerid, "1D", ta.rma(close, 3))
plot(rma, "RMA")
`
	testDir := t.TempDir()
	strategyPath := filepath.Join(testDir, "rma-exact.pine")
	if err := os.WriteFile(strategyPath, []byte(strategy), 0644); err != nil {
		t.Fatal(err)
	}
	dataPath := filepath.Join(testDir, "RMAEX_1D.json")
	if err := os.WriteFile(dataPath, []byte(generateTestOHLCV(15, 86400)), 0644); err != nil {
		t.Fatal(err)
	}

	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(filepath.Dir(cwd))

	result := compileAndRun(t, strategyPath, dataPath, testDir, projectRoot, "RMAEX", testDir)

	ind, ok := result.Indicators["RMA"]
	if !ok {
		t.Fatal("RMA indicator absent from output")
	}
	vals := extractValues(ind.Data)

	checks := []struct {
		barIdx int
		want   float64
	}{
		{3, 50051.0},           // secBar 2: warmup boundary — running avg = SMA = 50051.0
		{4, 50051.0 + 2.0/3.0}, // secBar 3: (1/3)×50053 + (2/3)×50051.0
	}
	for _, c := range checks {
		if c.barIdx >= len(vals) {
			t.Fatalf("bar %d out of range (len=%d)", c.barIdx, len(vals))
		}
		if math.IsNaN(vals[c.barIdx]) {
			t.Errorf("bar %d: unexpected NaN, want %.9f", c.barIdx, c.want)
			continue
		}
		if math.Abs(vals[c.barIdx]-c.want) > 1e-9 {
			t.Errorf("bar %d: RMA(3) = %.12f, want %.12f", c.barIdx, vals[c.barIdx], c.want)
		}
	}

	for i := 0; i < 3 && i < len(vals); i++ {
		if !math.IsNaN(vals[i]) {
			t.Errorf("bar %d: expected NaN during RMA(3) warmup, got %.6f", i, vals[i])
		}
	}
}

/*
TestSecurityTA_CUMRunningTotal verifies exact cumulative sum values through
request.security(), validating the prev()-based accumulation path of CUMStateManager.

generateTestOHLCV: close[s] = 50050 + s.
cum(close)[secBar s] = Σ_{i=0}^{s}(50050+i) = (s+1)×50050 + s×(s+1)/2

1-bar lookahead_off lag: base bar k → secBarIdx k−1.
cum[base k] = k×50050 + (k−1)×k/2, k ≥ 1.
*/
func TestSecurityTA_CUMRunningTotal(t *testing.T) {
	strategy := `//@version=5
indicator("CUM Running Total", overlay=false)
cum_ = request.security(syminfo.tickerid, "1D", ta.cum(close))
plot(cum_, "CUM")
`
	testDir := t.TempDir()
	strategyPath := filepath.Join(testDir, "cum-total.pine")
	if err := os.WriteFile(strategyPath, []byte(strategy), 0644); err != nil {
		t.Fatal(err)
	}
	dataPath := filepath.Join(testDir, "CUMTOT_1D.json")
	if err := os.WriteFile(dataPath, []byte(generateTestOHLCV(15, 86400)), 0644); err != nil {
		t.Fatal(err)
	}

	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(filepath.Dir(cwd))

	result := compileAndRun(t, strategyPath, dataPath, testDir, projectRoot, "CUMTOT", testDir)

	ind, ok := result.Indicators["CUM"]
	if !ok {
		t.Fatal("CUM indicator absent from output")
	}
	vals := extractValues(ind.Data)

	// cum[base k] = k×50050 + (k-1)×k/2; all values are exact integers.
	for k := 1; k < len(vals); k++ {
		if math.IsNaN(vals[k]) {
			t.Errorf("bar %d: unexpected NaN — CUM has no warmup period", k)
			continue
		}
		want := float64(k)*50050.0 + float64(k-1)*float64(k)/2.0
		if math.Abs(vals[k]-want) > 1e-6 {
			t.Errorf("bar %d: cum(close) = %.6f, want %.6f", k, vals[k], want)
		}
	}
}

/*
TestSecurityTA_HistoricalSubscript_SelfReferentialManagers verifies that the [1]
historical subscript on EMA and CUM returns the same value as the unsubscripted
indicator at the previous base bar — the ForwardSeriesBuffer look-back contract
for self-referential managers that accumulate via prev().

Both EMA and CUM rely on prev() for their per-bar recurrence; a failure here
indicates a cursor offset error in the forwardBufferE historical indexing.
*/
func TestSecurityTA_HistoricalSubscript_SelfReferentialManagers(t *testing.T) {
	strategy := `//@version=5
indicator("Historical Subscript Self-Referential", overlay=false)
ema     = request.security(syminfo.tickerid, "1D", ta.ema(close, 3))
emaPrev = request.security(syminfo.tickerid, "1D", ta.ema(close, 3)[1])
cum_    = request.security(syminfo.tickerid, "1D", ta.cum(close))
cumPrev = request.security(syminfo.tickerid, "1D", ta.cum(close)[1])
plot(ema,     "EMA")
plot(emaPrev, "EMAPrev")
plot(cum_,    "CUM")
plot(cumPrev, "CUMPrev")
`
	testDir := t.TempDir()
	strategyPath := filepath.Join(testDir, "subscript-selfreferential.pine")
	if err := os.WriteFile(strategyPath, []byte(strategy), 0644); err != nil {
		t.Fatal(err)
	}
	dataPath := filepath.Join(testDir, "SUBSR_1D.json")
	if err := os.WriteFile(dataPath, []byte(generateTestOHLCV(25, 86400)), 0644); err != nil {
		t.Fatal(err)
	}

	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(filepath.Dir(cwd))

	result := compileAndRun(t, strategyPath, dataPath, testDir, projectRoot, "SUBSR", testDir)

	emaInd, ok := result.Indicators["EMA"]
	if !ok {
		t.Fatal("EMA indicator absent")
	}
	emaPrevInd, ok := result.Indicators["EMAPrev"]
	if !ok {
		t.Fatal("EMAPrev indicator absent")
	}
	cumInd, ok := result.Indicators["CUM"]
	if !ok {
		t.Fatal("CUM indicator absent")
	}
	cumPrevInd, ok := result.Indicators["CUMPrev"]
	if !ok {
		t.Fatal("CUMPrev indicator absent")
	}

	ema := extractValues(emaInd.Data)
	emaPrev := extractValues(emaPrevInd.Data)
	cum := extractValues(cumInd.Data)
	cumPrev := extractValues(cumPrevInd.Data)

	// EMA(3) first valid: base bar 3. Both bar k and k-1 are valid from k=4.
	t.Run("EMA", func(t *testing.T) {
		for i := 4; i < len(ema) && i < len(emaPrev); i++ {
			if math.IsNaN(ema[i-1]) || math.IsNaN(emaPrev[i]) {
				continue
			}
			if math.Abs(ema[i-1]-emaPrev[i]) > 1e-6 {
				t.Errorf("bar %d: ta.ema[1] = %.9f, ta.ema at bar %d = %.9f, want equal",
					i, emaPrev[i], i-1, ema[i-1])
			}
		}
	})

	// CUM has no warmup; both bar k and k-1 are valid from k=2.
	t.Run("CUM", func(t *testing.T) {
		for i := 2; i < len(cum) && i < len(cumPrev); i++ {
			if math.IsNaN(cum[i-1]) || math.IsNaN(cumPrev[i]) {
				continue
			}
			if math.Abs(cum[i-1]-cumPrev[i]) > 1e-6 {
				t.Errorf("bar %d: ta.cum[1] = %.6f, ta.cum at bar %d = %.6f, want equal",
					i, cumPrev[i], i-1, cum[i-1])
			}
		}
	})
}

/*
TestSecurityTA_SAR_DirectionalInvariant verifies that in a persistent uptrend
ta.sar stays strictly below the low of its own security-context bar across the
full codegen→compile→execute pipeline, validating the scalar side-state closure
(sar, ep, af, isUptrend) captured by the forwardBufferE callback.

generateTestOHLCV: high[s] = 50100+s, low[s] = 49900+s — monotonically rising,
guaranteeing uptrend initialization (high[1] > high[0]) and no reversal.
Both ta.sar and low are requested from the same security context so the 1-bar
lookahead_off lag cancels and the comparison is bar-aligned.
*/
func TestSecurityTA_SAR_DirectionalInvariant(t *testing.T) {
	strategy := `//@version=5
indicator("SAR Directional Invariant", overlay=false)
sarVal = request.security(syminfo.tickerid, "1D", ta.sar(0.02, 0.02, 0.2))
sarLow = request.security(syminfo.tickerid, "1D", low)
plot(sarVal, "SAR")
plot(sarLow, "LOW")
`
	testDir := t.TempDir()
	strategyPath := filepath.Join(testDir, "sar-directional.pine")
	if err := os.WriteFile(strategyPath, []byte(strategy), 0644); err != nil {
		t.Fatal(err)
	}
	dataPath := filepath.Join(testDir, "SARDIR_1D.json")
	if err := os.WriteFile(dataPath, []byte(generateTestOHLCV(20, 86400)), 0644); err != nil {
		t.Fatal(err)
	}

	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(filepath.Dir(cwd))

	result := compileAndRun(t, strategyPath, dataPath, testDir, projectRoot, "SARDIR", testDir)

	sarInd, ok := result.Indicators["SAR"]
	if !ok {
		t.Fatal("SAR indicator absent from output")
	}
	lowInd, ok := result.Indicators["LOW"]
	if !ok {
		t.Fatal("LOW indicator absent from output")
	}

	sarVals := extractValues(sarInd.Data)
	lowVals := extractValues(lowInd.Data)

	nonNull := 0
	for i := 0; i < len(sarVals) && i < len(lowVals); i++ {
		if math.IsNaN(sarVals[i]) {
			continue
		}
		nonNull++
		if sarVals[i] <= 0 {
			t.Errorf("bar %d: SAR = %.6f — must be positive for positive price data", i, sarVals[i])
		}
		if !math.IsNaN(lowVals[i]) && sarVals[i] >= lowVals[i] {
			t.Errorf("bar %d: SAR %.6f ≥ low %.6f — uptrend SAR must stay below lows", i, sarVals[i], lowVals[i])
		}
	}
	if nonNull == 0 {
		t.Error("SAR produced zero non-null values")
	}
}

/*
TestSecurityTA_TSI_Bounds verifies that TSI values produced by request.security()
stay within the theoretical [-100, 100] range and are strictly positive on a
monotonically rising input — validating that the prevSource scalar and four
streaming EMA chain in TSIStateManager accumulate correctly via forwardBufferE.

generateTestOHLCV: close[s] = 50050+s — unit-step increase every bar.
All momentums are +1; ema2Abs = ema2Mom → TSI approaches +100 post-warmup.
*/
func TestSecurityTA_TSI_Bounds(t *testing.T) {
	strategy := `//@version=5
indicator("TSI Bounds", overlay=false)
tsi = request.security(syminfo.tickerid, "1D", ta.tsi(close, 5, 13))
plot(tsi, "TSI")
`
	testDir := t.TempDir()
	strategyPath := filepath.Join(testDir, "tsi-bounds.pine")
	if err := os.WriteFile(strategyPath, []byte(strategy), 0644); err != nil {
		t.Fatal(err)
	}
	dataPath := filepath.Join(testDir, "TSIBND_1D.json")
	if err := os.WriteFile(dataPath, []byte(generateTestOHLCV(30, 86400)), 0644); err != nil {
		t.Fatal(err)
	}

	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(filepath.Dir(cwd))

	result := compileAndRun(t, strategyPath, dataPath, testDir, projectRoot, "TSIBND", testDir)

	ind, ok := result.Indicators["TSI"]
	if !ok {
		t.Fatal("TSI indicator absent from output")
	}
	vals := extractValues(ind.Data)

	// TSI(5,13) warmup = 5+13−1 = 17 secBars → first valid base bar = 18.
	nonNull := 0
	for i, v := range vals {
		if math.IsNaN(v) {
			continue
		}
		nonNull++
		if v < -100 || v > 100 {
			t.Errorf("bar %d: TSI = %.6f out of [-100, 100]", i, v)
		}
		if i >= 18 && v <= 0 {
			t.Errorf("bar %d: TSI = %.6f — must be positive on monotonically rising input", i, v)
		}
	}
	if nonNull == 0 {
		t.Error("TSI produced zero non-null values across 30 bars")
	}
}
