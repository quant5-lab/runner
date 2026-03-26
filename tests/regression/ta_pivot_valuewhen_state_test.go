package regression

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"testing"
)

// generateWaveOHLCV produces OHLCV bars with a period-10 triangular-wave pattern:
// High peaks and Low troughs at bars 5, 15, 25, … at values 50500 / 49500;
// all other fields constant at 50000.
func generateWaveOHLCV(bars int, intervalSec int64) string {
	type Bar struct {
		Time   int64   `json:"time"`
		Open   float64 `json:"open"`
		High   float64 `json:"high"`
		Low    float64 `json:"low"`
		Close  float64 `json:"close"`
		Volume float64 `json:"volume"`
	}
	type OHLCVData struct {
		Timezone string `json:"timezone"`
		Bars     []Bar  `json:"bars"`
	}

	const (
		wavePeriod = 10
		halfPeriod = wavePeriod / 2
		amplitude  = 500.0
		basePrice  = 50000.0
	)

	startTime := int64(1640000000)
	bs := make([]Bar, bars)
	for i := 0; i < bars; i++ {
		phase := i % wavePeriod
		var level float64
		if phase <= halfPeriod {
			level = float64(phase) * (amplitude / float64(halfPeriod))
		} else {
			level = float64(wavePeriod-phase) * (amplitude / float64(halfPeriod))
		}
		bs[i] = Bar{
			Time:   startTime + int64(i)*intervalSec,
			Open:   basePrice,
			High:   basePrice + level,
			Low:    basePrice - level,
			Close:  basePrice,
			Volume: 100.0,
		}
	}

	data := OHLCVData{Timezone: "UTC", Bars: bs}
	jsonData, _ := json.MarshalIndent(data, "", "  ")
	return string(jsonData)
}

// TestValuewhen_InSecurity_NaNConditionProducesAllNull verifies that na as the
// condition to ta.valuewhen produces all-null output end-to-end.
func TestValuewhen_InSecurity_NaNConditionProducesAllNull(t *testing.T) {
	strategy := `//@version=5
indicator("Valuewhen NaN Cond", overlay=false)
v = request.security(syminfo.tickerid, "1D", ta.valuewhen(na, close, 0))
plot(v, "V")
`
	testDir := t.TempDir()
	strategyPath := filepath.Join(testDir, "vw-nan.pine")
	if err := os.WriteFile(strategyPath, []byte(strategy), 0644); err != nil {
		t.Fatal(err)
	}
	dataPath := filepath.Join(testDir, "VWNAN_1D.json")
	if err := os.WriteFile(dataPath, []byte(generateTestOHLCV(20, 86400)), 0644); err != nil {
		t.Fatal(err)
	}

	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(filepath.Dir(cwd))

	result := compileAndRun(t, strategyPath, dataPath, testDir, projectRoot, "VWNAN", testDir)

	ind, ok := result.Indicators["V"]
	if !ok {
		t.Fatal("indicator 'V' absent from output")
	}
	if countNonNull(ind.Data) != 0 {
		t.Errorf("na condition: expected 0 non-null bars, got %d", countNonNull(ind.Data))
	}
}

// TestValuewhen_InSecurity_CarryForward verifies that ta.valuewhen inside
// request.security() carries its last matched value forward on non-matching bars
// and updates on each new match.
func TestValuewhen_InSecurity_CarryForward(t *testing.T) {
	strategy := `//@version=5
indicator("Valuewhen Carry", overlay=false)
v = request.security(syminfo.tickerid, "1D", ta.valuewhen(close % 10.0 == 5.0, close, 0))
plot(v, "V")
`
	testDir := t.TempDir()
	strategyPath := filepath.Join(testDir, "vw-carry.pine")
	if err := os.WriteFile(strategyPath, []byte(strategy), 0644); err != nil {
		t.Fatal(err)
	}
	dataPath := filepath.Join(testDir, "VWCARRY_1D.json")
	if err := os.WriteFile(dataPath, []byte(generateTestOHLCV(30, 86400)), 0644); err != nil {
		t.Fatal(err)
	}

	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(filepath.Dir(cwd))

	result := compileAndRun(t, strategyPath, dataPath, testDir, projectRoot, "VWCARRY", testDir)

	ind, ok := result.Indicators["V"]
	if !ok {
		t.Fatal("indicator 'V' absent from output")
	}
	vals := extractValues(ind.Data)

	for i := 0; i <= 5 && i < len(vals); i++ {
		if !math.IsNaN(vals[i]) {
			t.Errorf("bar %d: expected NaN before first match, got %.4f", i, vals[i])
		}
	}

	if len(vals) > 6 {
		if math.IsNaN(vals[6]) {
			t.Errorf("bar 6: expected 50055 at first match, got NaN")
		} else if math.Abs(vals[6]-50055.0) > 1e-6 {
			t.Errorf("bar 6: expected 50055, got %.4f", vals[6])
		}
	}

	for i := 7; i <= 15 && i < len(vals); i++ {
		if math.IsNaN(vals[i]) || math.Abs(vals[i]-50055.0) > 1e-6 {
			t.Errorf("bar %d: expected 50055 (carry-forward), got %.4f", i, vals[i])
		}
	}

	if len(vals) > 16 {
		if math.IsNaN(vals[16]) {
			t.Errorf("bar 16: expected 50065 at second match, got NaN")
		} else if math.Abs(vals[16]-50065.0) > 1e-6 {
			t.Errorf("bar 16: expected 50065, got %.4f", vals[16])
		}
	}

	for i := 17; i <= 25 && i < len(vals); i++ {
		if math.IsNaN(vals[i]) || math.Abs(vals[i]-50065.0) > 1e-6 {
			t.Errorf("bar %d: expected 50065 (carry-forward), got %.4f", i, vals[i])
		}
	}

	if len(vals) > 26 {
		if math.IsNaN(vals[26]) {
			t.Errorf("bar 26: expected 50075 at third match, got NaN")
		} else if math.Abs(vals[26]-50075.0) > 1e-6 {
			t.Errorf("bar 26: expected 50075, got %.4f", vals[26])
		}
	}
}

// TestValuewhen_InSecurity_OccurrenceSemantics verifies that occurrence=0 returns
// the most recent match and occurrence=1 the previous one, staying NaN until
// a second match has occurred.
func TestValuewhen_InSecurity_OccurrenceSemantics(t *testing.T) {
	strategy := `//@version=5
indicator("Valuewhen Occurrence", overlay=false)
occ0 = request.security(syminfo.tickerid, "1D", ta.valuewhen(close % 10.0 == 5.0, close, 0))
occ1 = request.security(syminfo.tickerid, "1D", ta.valuewhen(close % 10.0 == 5.0, close, 1))
plot(occ0, "Occ0")
plot(occ1, "Occ1")
`
	testDir := t.TempDir()
	strategyPath := filepath.Join(testDir, "vw-occ.pine")
	if err := os.WriteFile(strategyPath, []byte(strategy), 0644); err != nil {
		t.Fatal(err)
	}
	dataPath := filepath.Join(testDir, "VWOCC_1D.json")
	if err := os.WriteFile(dataPath, []byte(generateTestOHLCV(30, 86400)), 0644); err != nil {
		t.Fatal(err)
	}

	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(filepath.Dir(cwd))

	result := compileAndRun(t, strategyPath, dataPath, testDir, projectRoot, "VWOCC", testDir)

	occ0Ind, ok := result.Indicators["Occ0"]
	if !ok {
		t.Fatal("indicator 'Occ0' absent from output")
	}
	occ1Ind, ok := result.Indicators["Occ1"]
	if !ok {
		t.Fatal("indicator 'Occ1' absent from output")
	}

	occ0 := extractValues(occ0Ind.Data)
	occ1 := extractValues(occ1Ind.Data)

	t.Run("occ0_first_match_at_bar6", func(t *testing.T) {
		for i := 0; i <= 5 && i < len(occ0); i++ {
			if !math.IsNaN(occ0[i]) {
				t.Errorf("bar %d: expected NaN, got %.4f", i, occ0[i])
			}
		}
		if len(occ0) > 6 && (math.IsNaN(occ0[6]) || math.Abs(occ0[6]-50055.0) > 1e-6) {
			t.Errorf("bar 6: expected 50055, got %.4f", occ0[6])
		}
	})

	t.Run("occ1_null_until_second_match", func(t *testing.T) {
		for i := 0; i <= 15 && i < len(occ1); i++ {
			if !math.IsNaN(occ1[i]) {
				t.Errorf("bar %d: expected NaN (only one match so far), got %.4f", i, occ1[i])
			}
		}
	})

	t.Run("occ1_returns_previous_match_value", func(t *testing.T) {
		if len(occ1) > 16 && (math.IsNaN(occ1[16]) || math.Abs(occ1[16]-50055.0) > 1e-6) {
			t.Errorf("bar 16: occ=1 expected 50055 (first match value), got %.4f", occ1[16])
		}
		for i := 17; i <= 25 && i < len(occ1); i++ {
			if math.IsNaN(occ1[i]) || math.Abs(occ1[i]-50055.0) > 1e-6 {
				t.Errorf("bar %d: occ=1 expected 50055 (carry-forward), got %.4f", i, occ1[i])
			}
		}
		if len(occ1) > 26 && (math.IsNaN(occ1[26]) || math.Abs(occ1[26]-50065.0) > 1e-6) {
			t.Errorf("bar 26: occ=1 expected 50065 (previous match updated), got %.4f", occ1[26])
		}
	})
}

// TestPivot_InSecurity_WarmupBoundaryAndDetection verifies that ta.pivothigh and
// ta.pivotlow inside request.security() produce NaN during warmup, emit the
// extremum value at each detection bar, and return NaN on non-extremum bars.
func TestPivot_InSecurity_WarmupBoundaryAndDetection(t *testing.T) {
	strategy := `//@version=5
indicator("Pivot Warmup Detection", overlay=false)
ph = request.security(syminfo.tickerid, "1D", ta.pivothigh(high, 2, 2))
pl = request.security(syminfo.tickerid, "1D", ta.pivotlow(low, 2, 2))
plot(ph, "PH")
plot(pl, "PL")
`
	testDir := t.TempDir()
	strategyPath := filepath.Join(testDir, "pivot-detect.pine")
	if err := os.WriteFile(strategyPath, []byte(strategy), 0644); err != nil {
		t.Fatal(err)
	}
	dataPath := filepath.Join(testDir, "PIVDET_1D.json")
	if err := os.WriteFile(dataPath, []byte(generateWaveOHLCV(30, 86400)), 0644); err != nil {
		t.Fatal(err)
	}

	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(filepath.Dir(cwd))

	result := compileAndRun(t, strategyPath, dataPath, testDir, projectRoot, "PIVDET", testDir)

	cases := []struct {
		plotName   string
		wantValue  float64
		detections []int
	}{
		{"PH", 50500.0, []int{8, 18, 28}},
		{"PL", 49500.0, []int{8, 18, 28}},
	}

	for _, tc := range cases {
		t.Run(tc.plotName, func(t *testing.T) {
			ind, ok := result.Indicators[tc.plotName]
			if !ok {
				t.Fatalf("indicator %q absent from output", tc.plotName)
			}
			vals := extractValues(ind.Data)

			detectionSet := make(map[int]bool)
			for _, d := range tc.detections {
				detectionSet[d] = true
			}

			for i, v := range vals {
				if detectionSet[i] {
					if math.IsNaN(v) {
						t.Errorf("bar %d: expected %.0f (detection), got NaN", i, tc.wantValue)
					} else if math.Abs(v-tc.wantValue) > 1e-6 {
						t.Errorf("bar %d: expected %.0f, got %.4f", i, tc.wantValue, v)
					}
				} else {
					if !math.IsNaN(v) {
						t.Errorf("bar %d: expected NaN (non-detection), got %.4f", i, v)
					}
				}
			}
		})
	}
}

func TestPivothigh_InSecurity_ArbitraryExpressionSource(t *testing.T) {
	strategy := `//@version=5
indicator("Pivot Expr Source", overlay=false)
phSpread = request.security(syminfo.tickerid, "1D", ta.pivothigh(high - low, 2, 2))
phHigh   = request.security(syminfo.tickerid, "1D", ta.pivothigh(high, 2, 2))
plot(phSpread, "PHSpread")
plot(phHigh,   "PHHigh")
`
	testDir := t.TempDir()
	strategyPath := filepath.Join(testDir, "pivot-expr-src.pine")
	if err := os.WriteFile(strategyPath, []byte(strategy), 0644); err != nil {
		t.Fatal(err)
	}
	dataPath := filepath.Join(testDir, "PIVEXPRSRC_1D.json")
	if err := os.WriteFile(dataPath, []byte(generateWaveOHLCV(30, 86400)), 0644); err != nil {
		t.Fatal(err)
	}

	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(filepath.Dir(cwd))

	result := compileAndRun(t, strategyPath, dataPath, testDir, projectRoot, "PIVEXPRSRC", testDir)

	spreadInd, ok := result.Indicators["PHSpread"]
	if !ok {
		t.Fatal("indicator 'PHSpread' absent from output")
	}
	highInd, ok := result.Indicators["PHHigh"]
	if !ok {
		t.Fatal("indicator 'PHHigh' absent from output")
	}

	spread := extractValues(spreadInd.Data)
	high := extractValues(highInd.Data)

	if len(spread) != len(high) {
		t.Fatalf("length mismatch: PHSpread=%d, PHHigh=%d", len(spread), len(high))
	}

	// Detection base bars: peak at sec bar 5, detected at sec bar 7, base bar 8 (1-bar lag).
	// PHSpread = 1000 (high-low at peak), PHHigh = 50500 at those bars.
	detectionBars := map[int]struct{}{8: {}, 18: {}, 28: {}}

	for i := range spread {
		spreadNaN := math.IsNaN(spread[i])
		highNaN := math.IsNaN(high[i])
		if spreadNaN != highNaN {
			t.Errorf("bar %d: NaN mismatch — PHSpread isNaN=%v, PHHigh isNaN=%v", i, spreadNaN, highNaN)
			continue
		}
		if _, isDetection := detectionBars[i]; isDetection {
			if spreadNaN {
				t.Errorf("bar %d: expected PHSpread=1000 (detection), got NaN", i)
			} else if math.Abs(spread[i]-1000.0) > 1e-6 {
				t.Errorf("bar %d: PHSpread = %.4f, want 1000.0", i, spread[i])
			}
		}
	}
}

func TestPivothigh_InSecurity_HistoricalSubscript(t *testing.T) {
	strategy := `//@version=5
indicator("Pivot Subscript", overlay=false)
ph     = request.security(syminfo.tickerid, "1D", ta.pivothigh(high, 2, 2))
phPrev = request.security(syminfo.tickerid, "1D", ta.pivothigh(high, 2, 2)[1])
plot(ph,     "PH")
plot(phPrev, "PHPrev")
`
	testDir := t.TempDir()
	strategyPath := filepath.Join(testDir, "pivot-subscript.pine")
	if err := os.WriteFile(strategyPath, []byte(strategy), 0644); err != nil {
		t.Fatal(err)
	}
	dataPath := filepath.Join(testDir, "PIVSUB_1D.json")
	if err := os.WriteFile(dataPath, []byte(generateWaveOHLCV(30, 86400)), 0644); err != nil {
		t.Fatal(err)
	}

	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(filepath.Dir(cwd))

	result := compileAndRun(t, strategyPath, dataPath, testDir, projectRoot, "PIVSUB", testDir)

	phInd, ok := result.Indicators["PH"]
	if !ok {
		t.Fatal("indicator 'PH' absent from output")
	}
	phPrevInd, ok := result.Indicators["PHPrev"]
	if !ok {
		t.Fatal("indicator 'PHPrev' absent from output")
	}

	ph := extractValues(phInd.Data)
	phPrev := extractValues(phPrevInd.Data)

	for i := 1; i < len(ph) && i < len(phPrev); i++ {
		if math.IsNaN(ph[i-1]) {
			if !math.IsNaN(phPrev[i]) {
				t.Errorf("bar %d: phPrev should be NaN (ph[%d]=NaN), got %.4f", i, i-1, phPrev[i])
			}
			continue
		}
		if math.IsNaN(phPrev[i]) {
			t.Errorf("bar %d: phPrev should be %.4f (ph[%d]), got NaN", i, ph[i-1], i-1)
			continue
		}
		if math.Abs(ph[i-1]-phPrev[i]) > 1e-6 {
			t.Errorf("bar %d: phPrev = %.4f, want ph[%d] = %.4f", i, phPrev[i], i-1, ph[i-1])
		}
	}

	detected := false
	for _, v := range phPrev {
		if !math.IsNaN(v) {
			detected = true
			break
		}
	}
	if !detected {
		t.Error("phPrev produced no non-NaN values — historical subscript not working")
	}
}

func TestPivotValuewhen_InSecurity_Composition(t *testing.T) {
	strategy := `//@version=5
indicator("Pivot Valuewhen Composition", overlay=false)
lastPH = request.security(syminfo.tickerid, "1D", ta.valuewhen(ta.pivothigh(high, 2, 2) > 0, ta.pivothigh(high, 2, 2), 0))
plot(lastPH, "LastPH")
`
	testDir := t.TempDir()
	strategyPath := filepath.Join(testDir, "pivot-vw-comp.pine")
	if err := os.WriteFile(strategyPath, []byte(strategy), 0644); err != nil {
		t.Fatal(err)
	}
	dataPath := filepath.Join(testDir, "PIVVWCOMP_1D.json")
	if err := os.WriteFile(dataPath, []byte(generateWaveOHLCV(30, 86400)), 0644); err != nil {
		t.Fatal(err)
	}

	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(filepath.Dir(cwd))

	result := compileAndRun(t, strategyPath, dataPath, testDir, projectRoot, "PIVVWCOMP", testDir)

	ind, ok := result.Indicators["LastPH"]
	if !ok {
		t.Fatal("indicator 'LastPH' absent from output")
	}
	vals := extractValues(ind.Data)

	// First detection at base bar 8 (sec bar 7, center bar 5 = wave peak).
	const firstDetectionBar = 8
	const expectedValue = 50500.0

	t.Run("null_before_first_detection", func(t *testing.T) {
		for i := 0; i < firstDetectionBar && i < len(vals); i++ {
			if !math.IsNaN(vals[i]) {
				t.Errorf("bar %d: expected NaN before first detection, got %.4f", i, vals[i])
			}
		}
	})

	t.Run("carry_forward_after_first_detection", func(t *testing.T) {
		for i := firstDetectionBar; i < len(vals); i++ {
			if math.IsNaN(vals[i]) {
				t.Errorf("bar %d: expected %.0f (carry-forward), got NaN", i, expectedValue)
			} else if math.Abs(vals[i]-expectedValue) > 1e-6 {
				t.Errorf("bar %d: expected %.0f, got %.4f", i, expectedValue, vals[i])
			}
		}
	})
}

func TestValuewhen_InSecurity_ExpressionCondition(t *testing.T) {
	strategy := `//@version=5
indicator("Valuewhen Expr Cond", overlay=false)
v = request.security(syminfo.tickerid, "1D", ta.valuewhen(high > low, close, 0))
plot(v, "V")
`
	testDir := t.TempDir()
	strategyPath := filepath.Join(testDir, "vw-expr-cond.pine")
	if err := os.WriteFile(strategyPath, []byte(strategy), 0644); err != nil {
		t.Fatal(err)
	}
	dataPath := filepath.Join(testDir, "VWEXPRCOND_1D.json")
	if err := os.WriteFile(dataPath, []byte(generateWaveOHLCV(30, 86400)), 0644); err != nil {
		t.Fatal(err)
	}

	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(filepath.Dir(cwd))

	result := compileAndRun(t, strategyPath, dataPath, testDir, projectRoot, "VWEXPRCOND", testDir)

	ind, ok := result.Indicators["V"]
	if !ok {
		t.Fatal("indicator 'V' absent from output")
	}
	vals := extractValues(ind.Data)

	// Sec bar 0: level=0, high=low=50000, cond=false → NaN.
	// Sec bar 1: level=100, high>low=true, close=50000 → first match.
	// 1-bar lag: first match at sec bar 1 → base bar 2.
	t.Run("null_before_first_match", func(t *testing.T) {
		for i := 0; i <= 1 && i < len(vals); i++ {
			if !math.IsNaN(vals[i]) {
				t.Errorf("bar %d: expected NaN before first match, got %.4f", i, vals[i])
			}
		}
	})

	t.Run("value_from_first_match_onwards", func(t *testing.T) {
		for i := 2; i < len(vals); i++ {
			if math.IsNaN(vals[i]) {
				t.Errorf("bar %d: expected 50000 (carry-forward), got NaN", i)
			} else if math.Abs(vals[i]-50000.0) > 1e-6 {
				t.Errorf("bar %d: expected 50000.0, got %.4f", i, vals[i])
			}
		}
	})
}

// generateClosePatternOHLCV produces OHLCV bars where Close follows the given
// absolute values. Open equals Close; High and Low are offset by a fixed spread
// to provide a non-degenerate OHLCV structure without interfering with close-based tests.
func generateClosePatternOHLCV(closes []float64, intervalSec int64) string {
	type Bar struct {
		Time   int64   `json:"time"`
		Open   float64 `json:"open"`
		High   float64 `json:"high"`
		Low    float64 `json:"low"`
		Close  float64 `json:"close"`
		Volume float64 `json:"volume"`
	}
	type OHLCVData struct {
		Timezone string `json:"timezone"`
		Bars     []Bar  `json:"bars"`
	}

	const spread = 50.0
	startTime := int64(1640000000)
	bars := make([]Bar, len(closes))
	for i, c := range closes {
		bars[i] = Bar{
			Time:   startTime + int64(i)*intervalSec,
			Open:   c,
			High:   c + spread,
			Low:    c - spread,
			Close:  c,
			Volume: 100.0,
		}
	}
	data := OHLCVData{Timezone: "UTC", Bars: bars}
	jsonData, _ := json.MarshalIndent(data, "", "  ")
	return string(jsonData)
}

// TestPivot_InSecurity_NaNNeighborBlocking verifies that when a SMA source
// produces NaN during warmup, those NaN-valued neighbor positions block pivot
// detection for both high and low kinds — matching PineScript where any
// comparison with na evaluates to false.
func TestPivot_InSecurity_NaNNeighborBlocking(t *testing.T) {
	// SMA(3) bars 0,1 = NaN; any center at bar 2 has NaN left neighbors → all output blocked.
	closes := []float64{50000, 49800, 50300, 49900, 49800, 49700, 49600, 49500, 49400, 49300}

	cases := []struct {
		name     string
		strategy string
		plotName string
	}{
		{
			name:     "pivothigh",
			plotName: "PH",
			strategy: `//@version=5
indicator("Pivot NaN Blocking High", overlay=false)
ph = request.security(syminfo.tickerid, "1D", ta.pivothigh(ta.sma(close, 3), 2, 2))
plot(ph, "PH")
`,
		},
		{
			name:     "pivotlow",
			plotName: "PL",
			strategy: `//@version=5
indicator("Pivot NaN Blocking Low", overlay=false)
pl = request.security(syminfo.tickerid, "1D", ta.pivotlow(ta.sma(close, 3), 2, 2))
plot(pl, "PL")
`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			testDir := t.TempDir()
			strategyPath := filepath.Join(testDir, "pivot-nan-blk.pine")
			if err := os.WriteFile(strategyPath, []byte(tc.strategy), 0644); err != nil {
				t.Fatal(err)
			}
			dataPath := filepath.Join(testDir, "PIVNANBLK_1D.json")
			if err := os.WriteFile(dataPath, []byte(generateClosePatternOHLCV(closes, 86400)), 0644); err != nil {
				t.Fatal(err)
			}

			cwd, _ := os.Getwd()
			projectRoot := filepath.Dir(filepath.Dir(cwd))

			result := compileAndRun(t, strategyPath, dataPath, testDir, projectRoot, "PIVNANBLK", testDir)

			ind, ok := result.Indicators[tc.plotName]
			if !ok {
				t.Fatalf("indicator %q absent from output", tc.plotName)
			}
			for i, v := range extractValues(ind.Data) {
				if !math.IsNaN(v) {
					t.Errorf("bar %d: expected NaN (NaN neighbor blocks pivot), got %.4f", i, v)
				}
			}
		})
	}
}

// TestPivot_InSecurity_NaNSourceDetectsAfterWarmup verifies that once a
// NaN-producing source expression has warmed up — so all neighbor positions
// hold valid values — pivot detection proceeds normally for both high and low kinds.
func TestPivot_InSecurity_NaNSourceDetectsAfterWarmup(t *testing.T) {
	cases := []struct {
		name          string
		strategy      string
		plotName      string
		closes        []float64
		expectedValue float64
		detectionBar  int
	}{
		{
			name:     "pivothigh",
			plotName: "PH",
			strategy: `//@version=5
indicator("Pivot NaN Post Warmup High", overlay=false)
ph = request.security(syminfo.tickerid, "1D", ta.pivothigh(ta.sma(close, 3), 2, 2))
plot(ph, "PH")
`,
			closes:        []float64{50000, 50000, 50000, 50000, 50900, 49800, 49700, 49600, 49500, 49400},
			expectedValue: (50000.0 + 50000.0 + 50900.0) / 3.0,
			detectionBar:  7,
		},
		{
			name:     "pivotlow",
			plotName: "PL",
			strategy: `//@version=5
indicator("Pivot NaN Post Warmup Low", overlay=false)
pl = request.security(syminfo.tickerid, "1D", ta.pivotlow(ta.sma(close, 3), 2, 2))
plot(pl, "PL")
`,
			closes:        []float64{50000, 50000, 50000, 50000, 49100, 50200, 50300, 50400, 50500, 50600},
			expectedValue: (50000.0 + 50000.0 + 49100.0) / 3.0,
			detectionBar:  7,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			testDir := t.TempDir()
			strategyPath := filepath.Join(testDir, "pivot-nan-post.pine")
			if err := os.WriteFile(strategyPath, []byte(tc.strategy), 0644); err != nil {
				t.Fatal(err)
			}
			dataPath := filepath.Join(testDir, "PIVNANPOST_1D.json")
			if err := os.WriteFile(dataPath, []byte(generateClosePatternOHLCV(tc.closes, 86400)), 0644); err != nil {
				t.Fatal(err)
			}

			cwd, _ := os.Getwd()
			projectRoot := filepath.Dir(filepath.Dir(cwd))

			result := compileAndRun(t, strategyPath, dataPath, testDir, projectRoot, "PIVNANPOST", testDir)

			ind, ok := result.Indicators[tc.plotName]
			if !ok {
				t.Fatalf("indicator %q absent from output", tc.plotName)
			}
			vals := extractValues(ind.Data)

			detected := false
			for i, v := range vals {
				if i == tc.detectionBar {
					if math.IsNaN(v) {
						t.Errorf("bar %d: expected %.4f (post-warmup detection), got NaN", tc.detectionBar, tc.expectedValue)
					} else if math.Abs(v-tc.expectedValue) > 1e-4 {
						t.Errorf("bar %d: expected %.4f, got %.4f", tc.detectionBar, tc.expectedValue, v)
					} else {
						detected = true
					}
				} else if !math.IsNaN(v) {
					t.Errorf("bar %d: expected NaN (no pivot), got %.4f", i, v)
				}
			}
			if !detected {
				t.Errorf("no post-warmup pivot detected for %s", tc.name)
			}
		})
	}
}

func TestPivot_InSecurity_FixnanSubscriptComposition(t *testing.T) {
	strategy := `//@version=5
indicator("Pivot Fixnan Subscript", overlay=false)
ph = request.security(syminfo.tickerid, "1D", fixnan(ta.pivothigh(high, 2, 2)[1]))
plot(ph, "PH")
`
	testDir := t.TempDir()
	strategyPath := filepath.Join(testDir, "pivot-fixnan-sub.pine")
	if err := os.WriteFile(strategyPath, []byte(strategy), 0644); err != nil {
		t.Fatal(err)
	}
	dataPath := filepath.Join(testDir, "PIVFIXNANSUB_1D.json")
	if err := os.WriteFile(dataPath, []byte(generateWaveOHLCV(30, 86400)), 0644); err != nil {
		t.Fatal(err)
	}

	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(filepath.Dir(cwd))

	result := compileAndRun(t, strategyPath, dataPath, testDir, projectRoot, "PIVFIXNANSUB", testDir)

	ind, ok := result.Indicators["PH"]
	if !ok {
		t.Fatal("indicator 'PH' absent from output")
	}
	vals := extractValues(ind.Data)

	// Wave peak at sec bar 5 (High=50500), detected at sec bar 7, shifted [1] to sec bar 8.
	// Sec bar 8 → base bar 9 (1-bar security lag). fixnan carries 50500 from bar 9 onward.
	const firstFillBar = 9
	const expectedValue = 50500.0

	for i, v := range vals {
		if i < firstFillBar {
			if !math.IsNaN(v) {
				t.Errorf("bar %d: expected NaN before fixnan fill, got %.4f", i, v)
			}
		} else if math.IsNaN(v) || math.Abs(v-expectedValue) > 1e-6 {
			t.Errorf("bar %d: expected %.0f (fixnan carry-forward), got %.4f", i, expectedValue, v)
		}
	}
}

func TestValuewhen_InSecurity_ArithmeticSourceExpression(t *testing.T) {
	strategy := `//@version=5
indicator("Valuewhen Arithmetic Src", overlay=false)
v = request.security(syminfo.tickerid, "1D", ta.valuewhen(close > open, high + low, 0))
plot(v, "V")
`
	testDir := t.TempDir()
	strategyPath := filepath.Join(testDir, "vw-arith-src.pine")
	if err := os.WriteFile(strategyPath, []byte(strategy), 0644); err != nil {
		t.Fatal(err)
	}
	dataPath := filepath.Join(testDir, "VWARITHSRC_1D.json")
	if err := os.WriteFile(dataPath, []byte(generateTestOHLCV(20, 86400)), 0644); err != nil {
		t.Fatal(err)
	}

	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(filepath.Dir(cwd))

	result := compileAndRun(t, strategyPath, dataPath, testDir, projectRoot, "VWARITHSRC", testDir)

	ind, ok := result.Indicators["V"]
	if !ok {
		t.Fatal("indicator 'V' absent from output")
	}
	vals := extractValues(ind.Data)

	if len(vals) == 0 {
		t.Fatal("no output bars")
	}

	// generateTestOHLCV: open=50000+i, close=50050+i, high=50100+i, low=49900+i.
	// close > open is always true; high+low = 100000+2i for sec bar i.
	// Base bar 0 and bar 1 both map to sec bar 0 (1-bar security lag); value = 100000.
	// Base bar k≥1 maps to sec bar k-1.
	for k := 0; k < len(vals); k++ {
		secBar := k - 1
		if secBar < 0 {
			secBar = 0
		}
		expected := 100000.0 + 2.0*float64(secBar)
		if math.IsNaN(vals[k]) {
			t.Errorf("bar %d: expected %.0f (high+low), got NaN", k, expected)
		} else if math.Abs(vals[k]-expected) > 1e-6 {
			t.Errorf("bar %d: expected %.0f, got %.4f", k, expected, vals[k])
		}
	}
}

// TestPivothigh_InSecurity_TwoArgDefaultSourceMatchesExplicit verifies that the
// 2-arg form ta.pivothigh(l, r) produces output identical to ta.pivothigh(high, l, r),
// and analogously for ta.pivotlow.
// TestValuewhen_InSecurity_HistoricalSubscriptAccess verifies that
// ta.valuewhen(...)[1] inside request.security() returns the previous
// security-bar's valuewhen value, exercising ForwardSeriesBuffer history
// access through the full pipeline — the symmetric counterpart to
// TestPivothigh_InSecurity_HistoricalSubscript.
func TestValuewhen_InSecurity_HistoricalSubscriptAccess(t *testing.T) {
	strategy := `//@version=5
indicator("Valuewhen Subscript", overlay=false)
vw     = request.security(syminfo.tickerid, "1D", ta.valuewhen(close % 10.0 == 5.0, close, 0))
vwPrev = request.security(syminfo.tickerid, "1D", ta.valuewhen(close % 10.0 == 5.0, close, 0)[1])
plot(vw,     "VW")
plot(vwPrev, "VWPrev")
`
	testDir := t.TempDir()
	strategyPath := filepath.Join(testDir, "vw-subscript.pine")
	if err := os.WriteFile(strategyPath, []byte(strategy), 0644); err != nil {
		t.Fatal(err)
	}
	dataPath := filepath.Join(testDir, "VWSUB_1D.json")
	if err := os.WriteFile(dataPath, []byte(generateTestOHLCV(30, 86400)), 0644); err != nil {
		t.Fatal(err)
	}

	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(filepath.Dir(cwd))

	result := compileAndRun(t, strategyPath, dataPath, testDir, projectRoot, "VWSUB", testDir)

	vwInd, ok := result.Indicators["VW"]
	if !ok {
		t.Fatal("indicator 'VW' absent from output")
	}
	vwPrevInd, ok := result.Indicators["VWPrev"]
	if !ok {
		t.Fatal("indicator 'VWPrev' absent from output")
	}

	vw := extractValues(vwInd.Data)
	vwPrev := extractValues(vwPrevInd.Data)

	for i := 1; i < len(vw) && i < len(vwPrev); i++ {
		if math.IsNaN(vw[i-1]) {
			if !math.IsNaN(vwPrev[i]) {
				t.Errorf("bar %d: vwPrev should be NaN (vw[%d]=NaN), got %.4f", i, i-1, vwPrev[i])
			}
			continue
		}
		if math.IsNaN(vwPrev[i]) {
			t.Errorf("bar %d: vwPrev should be %.4f (vw[%d]), got NaN", i, vw[i-1], i-1)
			continue
		}
		if math.Abs(vw[i-1]-vwPrev[i]) > 1e-6 {
			t.Errorf("bar %d: vwPrev = %.4f, want vw[%d] = %.4f", i, vwPrev[i], i-1, vw[i-1])
		}
	}

	detected := false
	for _, v := range vwPrev {
		if !math.IsNaN(v) {
			detected = true
			break
		}
	}
	if !detected {
		t.Error("vwPrev produced no non-NaN values — historical subscript not working")
	}
}

// TestPivot_InSecurity_AsymmetricWindow verifies that asymmetric left/right bar
// counts produce detection at the correct base bar. With (leftBars=1, rightBars=3)
// detection occurs 3 bars after the center; with (leftBars=3, rightBars=1) it
// occurs 1 bar after — confirming the delayed-detection formula end-to-end.
func TestPivot_InSecurity_AsymmetricWindow(t *testing.T) {
	strategy := `//@version=5
indicator("Pivot Asymmetric Window", overlay=false)
ph13 = request.security(syminfo.tickerid, "1D", ta.pivothigh(high, 1, 3))
ph31 = request.security(syminfo.tickerid, "1D", ta.pivothigh(high, 3, 1))
plot(ph13, "PH13")
plot(ph31, "PH31")
`
	testDir := t.TempDir()
	strategyPath := filepath.Join(testDir, "pivot-asym.pine")
	if err := os.WriteFile(strategyPath, []byte(strategy), 0644); err != nil {
		t.Fatal(err)
	}
	dataPath := filepath.Join(testDir, "PIVASYM_1D.json")
	if err := os.WriteFile(dataPath, []byte(generateWaveOHLCV(30, 86400)), 0644); err != nil {
		t.Fatal(err)
	}

	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(filepath.Dir(cwd))

	result := compileAndRun(t, strategyPath, dataPath, testDir, projectRoot, "PIVASYM", testDir)

	cases := []struct {
		plotName   string
		detections []int
		wantValue  float64
	}{
		// Wave peak at sec bar 5 (High=50500).
		// (1,3): detected at sec bar 5+3=8 → base bar 9.
		{"PH13", []int{9, 19, 29}, 50500.0},
		// (3,1): detected at sec bar 5+1=6 → base bar 7.
		{"PH31", []int{7, 17, 27}, 50500.0},
	}

	for _, tc := range cases {
		t.Run(tc.plotName, func(t *testing.T) {
			ind, ok := result.Indicators[tc.plotName]
			if !ok {
				t.Fatalf("indicator %q absent from output", tc.plotName)
			}
			vals := extractValues(ind.Data)
			detectionSet := make(map[int]bool, len(tc.detections))
			for _, d := range tc.detections {
				detectionSet[d] = true
			}
			for i, v := range vals {
				if detectionSet[i] {
					if math.IsNaN(v) {
						t.Errorf("bar %d: expected %.0f (detection), got NaN", i, tc.wantValue)
					} else if math.Abs(v-tc.wantValue) > 1e-6 {
						t.Errorf("bar %d: expected %.0f, got %.4f", i, tc.wantValue, v)
					}
				} else if !math.IsNaN(v) {
					t.Errorf("bar %d: expected NaN (non-detection), got %.4f", i, v)
				}
			}
		})
	}
}

func TestPivothigh_InSecurity_TwoArgDefaultSourceMatchesExplicit(t *testing.T) {
	strategy := `//@version=5
indicator("Pivot Default Source", overlay=false)
phExplicit = request.security(syminfo.tickerid, "1D", ta.pivothigh(high, 2, 2))
phDefault  = request.security(syminfo.tickerid, "1D", ta.pivothigh(2, 2))
plExplicit = request.security(syminfo.tickerid, "1D", ta.pivotlow(low, 2, 2))
plDefault  = request.security(syminfo.tickerid, "1D", ta.pivotlow(2, 2))
plot(phExplicit, "PHExplicit")
plot(phDefault,  "PHDefault")
plot(plExplicit, "PLExplicit")
plot(plDefault,  "PLDefault")
`
	testDir := t.TempDir()
	strategyPath := filepath.Join(testDir, "pivot-default-src.pine")
	if err := os.WriteFile(strategyPath, []byte(strategy), 0644); err != nil {
		t.Fatal(err)
	}
	dataPath := filepath.Join(testDir, "PIVDEFSRC_1D.json")
	if err := os.WriteFile(dataPath, []byte(generateWaveOHLCV(30, 86400)), 0644); err != nil {
		t.Fatal(err)
	}

	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(filepath.Dir(cwd))

	result := compileAndRun(t, strategyPath, dataPath, testDir, projectRoot, "PIVDEFSRC", testDir)

	pairs := [][2]string{
		{"PHExplicit", "PHDefault"},
		{"PLExplicit", "PLDefault"},
	}

	for _, pair := range pairs {
		t.Run(pair[0]+"_vs_"+pair[1], func(t *testing.T) {
			explInd, ok := result.Indicators[pair[0]]
			if !ok {
				t.Fatalf("indicator %q absent from output", pair[0])
			}
			defInd, ok := result.Indicators[pair[1]]
			if !ok {
				t.Fatalf("indicator %q absent from output", pair[1])
			}

			expl := extractValues(explInd.Data)
			def := extractValues(defInd.Data)

			if len(expl) != len(def) {
				t.Fatalf("length mismatch: explicit=%d, default=%d", len(expl), len(def))
			}

			for i := range expl {
				explNaN := math.IsNaN(expl[i])
				defNaN := math.IsNaN(def[i])
				if explNaN != defNaN {
					t.Errorf("bar %d: null mismatch — explicit isNaN=%v, default isNaN=%v", i, explNaN, defNaN)
					continue
				}
				if !explNaN && math.Abs(expl[i]-def[i]) > 1e-6 {
					t.Errorf("bar %d: value mismatch — explicit=%.4f, default=%.4f", i, expl[i], def[i])
				}
			}
		})
	}
}
