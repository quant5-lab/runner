package regression

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"testing"
)

/*
TestSecurityChartType_Identity_UserVarResolvesViaBarIndexMapper verifies the
ModeIdentity path: a same-TF heikinashi security() call with a user-defined
variable as the TA source argument produces correct numeric values once the TA
warmup is satisfied.

generateTestOHLCV sets Close[i] = 50050 + i, so myVar[i] = 2*(50050+i).
SMA(5) warmup completes at secBarIdx=4: expected value = 100104.
*/
func TestSecurityChartType_Identity_UserVarResolvesViaBarIndexMapper(t *testing.T) {
	const barCount = 20

	strategy := `//@version=5
indicator("HA UserVar Identity", overlay=false)
myVar = close * 2.0
haResult = request.security(ticker.heikinashi(syminfo.tickerid), timeframe.period, ta.sma(myVar, 5))
plot(haResult, "HAResult")
`
	testDir := t.TempDir()
	strategyPath := filepath.Join(testDir, "ha-identity.pine")
	if err := os.WriteFile(strategyPath, []byte(strategy), 0644); err != nil {
		t.Fatal(err)
	}
	dataPath := filepath.Join(testDir, "IDENHA_1h.json")
	if err := os.WriteFile(dataPath, []byte(generateTestOHLCV(barCount, 3600)), 0644); err != nil {
		t.Fatal(err)
	}

	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(filepath.Dir(cwd))

	result := compileAndRun(t, strategyPath, dataPath, testDir, projectRoot, "IDENHA", testDir)

	ind, ok := result.Indicators["HAResult"]
	if !ok {
		t.Fatalf("indicator 'HAResult' absent — chart type identity security not plotted")
	}
	vals := extractValues(ind.Data)

	if len(vals) != barCount {
		t.Fatalf("got %d bars, want %d", len(vals), barCount)
	}

	/* Bars 0..3: SMA warmup — fewer than 5 data points available. */
	for i := 0; i < 4; i++ {
		if !math.IsNaN(vals[i]) {
			t.Errorf("bar %d: expected NaN during SMA warmup, got %.6f", i, vals[i])
		}
	}

	/* Bar 4: first valid SMA output. */
	if math.IsNaN(vals[4]) {
		t.Fatal("bar 4: got NaN — ModeIdentity bar index mapping not populated")
	}
	const smaBar4 = 100104.0
	if math.Abs(vals[4]-smaBar4) > 1e-6 {
		t.Errorf("bar 4: SMA(myVar,5) = %.9f, want %.9f", vals[4], smaBar4)
	}

	/* Bar 5: SMA window slides by one position. */
	if math.IsNaN(vals[5]) {
		t.Errorf("bar 5: unexpected NaN post-warmup")
	}
	const smaBar5 = 100106.0
	if math.Abs(vals[5]-smaBar5) > 1e-6 {
		t.Errorf("bar 5: SMA(myVar,5) = %.9f, want %.9f", vals[5], smaBar5)
	}

	/* Bars 4..N-1: SMA window fully populated — all must be non-NaN. */
	for i := 4; i < len(vals); i++ {
		if math.IsNaN(vals[i]) {
			t.Errorf("bar %d: unexpected NaN post-warmup", i)
		}
	}
}

/*
TestSecurityChartType_Transformed_UserVarResolvesViaBarIndexMapper verifies the
ModeTransformed path: a same-TF renko security() call with a user-defined
variable as the TA source argument produces correct numeric values once the TA
warmup is satisfied and enough renko bricks have formed.

generateTestOHLCV sets Close[i] = 50050 + i. With a Traditional renko box of 5,
one brick forms every 5 main bars. SMA(5) over myVar = close*2 is first valid at
secBarIdx=4 (main bar 25): expected value = 100130.
*/
func TestSecurityChartType_Transformed_UserVarResolvesViaBarIndexMapper(t *testing.T) {
	const barCount = 40

	strategy := `//@version=5
indicator("Renko UserVar Transformed", overlay=false)
myVar = close * 2.0
renkoResult = request.security(ticker.renko(syminfo.tickerid, "Traditional", 5), timeframe.period, ta.sma(myVar, 5))
plot(renkoResult, "RenkoResult")
`
	testDir := t.TempDir()
	strategyPath := filepath.Join(testDir, "renko-transformed.pine")
	if err := os.WriteFile(strategyPath, []byte(strategy), 0644); err != nil {
		t.Fatal(err)
	}
	dataPath := filepath.Join(testDir, "RENKOTEST_1h.json")
	if err := os.WriteFile(dataPath, []byte(generateTestOHLCV(barCount, 3600)), 0644); err != nil {
		t.Fatal(err)
	}

	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(filepath.Dir(cwd))

	result := compileAndRun(t, strategyPath, dataPath, testDir, projectRoot, "RENKOTEST", testDir)

	ind, ok := result.Indicators["RenkoResult"]
	if !ok {
		t.Fatalf("indicator 'RenkoResult' absent — chart type transformed security not plotted")
	}
	vals := extractValues(ind.Data)

	if len(vals) != barCount {
		t.Fatalf("got %d bars, want %d", len(vals), barCount)
	}

	/* Bars 0..4: no renko bricks formed yet. */
	for i := 0; i < 5; i++ {
		if !math.IsNaN(vals[i]) {
			t.Errorf("bar %d: expected NaN before first renko brick, got %.6f", i, vals[i])
		}
	}

	/* Bars 5..24: bricks forming but SMA still in warmup. */
	for i := 5; i < 25; i++ {
		if !math.IsNaN(vals[i]) {
			t.Errorf("bar %d: expected NaN during SMA warmup, got %.6f", i, vals[i])
		}
	}

	/* Bar 25: first valid output — SMA warmup complete at secBarIdx=4. */
	if math.IsNaN(vals[25]) {
		t.Fatal("bar 25: got NaN — ModeTransformed reverse bar index mapping not populated")
	}
	const smaBar25 = 100130.0
	if math.Abs(vals[25]-smaBar25) > 1e-6 {
		t.Errorf("bar 25: SMA(myVar,5) = %.9f, want %.9f", vals[25], smaBar25)
	}

	/* Bar 30: SMA window advances by one brick. */
	if math.IsNaN(vals[30]) {
		t.Errorf("bar 30: unexpected NaN post-warmup")
	}
	const smaBar30 = 100140.0
	if math.Abs(vals[30]-smaBar30) > 1e-6 {
		t.Errorf("bar 30: SMA(myVar,5) = %.9f, want %.9f", vals[30], smaBar30)
	}

	/* Bars 25..35: SMA window fully populated — all must be non-NaN. */
	for i := 25; i <= 35; i++ {
		if math.IsNaN(vals[i]) {
			t.Errorf("bar %d: unexpected NaN post-warmup", i)
		}
	}
}

// generateOscillatingTestOHLCV creates OHLCV bars whose close price alternates
// between basePrice+amplitude (even bars) and basePrice-amplitude (odd bars).
// This produces the directional reversals required by kagi and pointfigure
// transformers when the reversal threshold is smaller than amplitude.
func generateOscillatingTestOHLCV(barCount int, intervalSeconds int64, amplitude float64) string {
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

	startTime := int64(1640000000)
	bars := make([]Bar, barCount)
	basePrice := 50000.0

	for i := 0; i < barCount; i++ {
		var closePrice float64
		if i%2 == 0 {
			closePrice = basePrice + amplitude
		} else {
			closePrice = basePrice - amplitude
		}
		bars[i] = Bar{
			Time:   startTime + int64(i)*intervalSeconds,
			Open:   closePrice,
			High:   closePrice + 10,
			Low:    closePrice - 10,
			Close:  closePrice,
			Volume: 100.0,
		}
	}

	data := OHLCVData{Timezone: "UTC", Bars: bars}
	jsonData, _ := json.MarshalIndent(data, "", "  ")
	return string(jsonData)
}

/*
TestSecurityChartType_Range_PassthroughViaModeIdentity verifies the ModeIdentity
path for ticker.range: the IdentityTransformer preserves bars 1:1 so TA over a
user-defined variable produces the same output as a plain same-TF security call.

generateTestOHLCV sets Close[i] = 50050+i, so myVar[i] = 2*(50050+i).
SMA(5) warmup completes at bar 4: expected value = 100104.
*/
func TestSecurityChartType_Range_PassthroughViaModeIdentity(t *testing.T) {
	const barCount = 20

	strategy := `//@version=5
indicator("Range Passthrough", overlay=false)
myVar = close * 2.0
rangeResult = request.security(ticker.range(syminfo.tickerid), timeframe.period, ta.sma(myVar, 5))
plot(rangeResult, "RangeResult")
`
	testDir := t.TempDir()
	strategyPath := filepath.Join(testDir, "range-identity.pine")
	if err := os.WriteFile(strategyPath, []byte(strategy), 0644); err != nil {
		t.Fatal(err)
	}
	dataPath := filepath.Join(testDir, "RANGETEST_1h.json")
	if err := os.WriteFile(dataPath, []byte(generateTestOHLCV(barCount, 3600)), 0644); err != nil {
		t.Fatal(err)
	}

	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(filepath.Dir(cwd))

	result := compileAndRun(t, strategyPath, dataPath, testDir, projectRoot, "RANGETEST", testDir)

	ind, ok := result.Indicators["RangeResult"]
	if !ok {
		t.Fatalf("indicator 'RangeResult' absent — ticker.range security not plotted")
	}
	vals := extractValues(ind.Data)

	if len(vals) != barCount {
		t.Fatalf("got %d bars, want %d", len(vals), barCount)
	}

	/* Bars 0..3: SMA warmup — fewer than 5 data points available. */
	for i := 0; i < 4; i++ {
		if !math.IsNaN(vals[i]) {
			t.Errorf("bar %d: expected NaN during SMA warmup, got %.6f", i, vals[i])
		}
	}

	/* Bar 4: first valid SMA output. */
	if math.IsNaN(vals[4]) {
		t.Fatal("bar 4: got NaN — ModeIdentity range passthrough not populated")
	}
	const smaBar4 = 100104.0
	if math.Abs(vals[4]-smaBar4) > 1e-6 {
		t.Errorf("bar 4: SMA(myVar,5) = %.9f, want %.9f", vals[4], smaBar4)
	}

	/* Bar 5: SMA window slides by one position. */
	const smaBar5 = 100106.0
	if math.IsNaN(vals[5]) || math.Abs(vals[5]-smaBar5) > 1e-6 {
		t.Errorf("bar 5: SMA(myVar,5) = %.9f, want %.9f", vals[5], smaBar5)
	}

	/* Bars 4..N-1: SMA window fully populated — all must be non-NaN. */
	for i := 4; i < len(vals); i++ {
		if math.IsNaN(vals[i]) {
			t.Errorf("bar %d: unexpected NaN post-warmup", i)
		}
	}
}

/*
TestSecurityChartType_LineBreak_MonotonicViaModeTransformed verifies the
ModeTransformed path for ticker.linebreak: with monotonically increasing prices
every source bar exceeds the highest of the last 3 lines, forming a new line on
every bar (1:1 synthetic mapping). SMA(5) over myVar therefore produces the same
first-valid value as the identity case.

generateTestOHLCV sets Close[i] = 50050+i. Each linebreak bar's close equals the
source bar's close, so myVar[i] = 2*(50050+i). SMA(5) first valid at bar 4: 100104.
*/
func TestSecurityChartType_LineBreak_MonotonicViaModeTransformed(t *testing.T) {
	const barCount = 20

	strategy := `//@version=5
indicator("LineBreak Transformed", overlay=false)
myVar = close * 2.0
lbResult = request.security(ticker.linebreak(syminfo.tickerid, 3), timeframe.period, ta.sma(myVar, 5))
plot(lbResult, "LBResult")
`
	testDir := t.TempDir()
	strategyPath := filepath.Join(testDir, "linebreak-transformed.pine")
	if err := os.WriteFile(strategyPath, []byte(strategy), 0644); err != nil {
		t.Fatal(err)
	}
	dataPath := filepath.Join(testDir, "LBTEST_1h.json")
	if err := os.WriteFile(dataPath, []byte(generateTestOHLCV(barCount, 3600)), 0644); err != nil {
		t.Fatal(err)
	}

	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(filepath.Dir(cwd))

	result := compileAndRun(t, strategyPath, dataPath, testDir, projectRoot, "LBTEST", testDir)

	ind, ok := result.Indicators["LBResult"]
	if !ok {
		t.Fatalf("indicator 'LBResult' absent — ticker.linebreak security not plotted")
	}
	vals := extractValues(ind.Data)

	if len(vals) != barCount {
		t.Fatalf("got %d bars, want %d", len(vals), barCount)
	}

	/* Bars 0..3: SMA warmup — fewer than 5 linebreak bars visible. */
	for i := 0; i < 4; i++ {
		if !math.IsNaN(vals[i]) {
			t.Errorf("bar %d: expected NaN during SMA warmup, got %.6f", i, vals[i])
		}
	}

	/* Bar 4: first valid SMA output — 1:1 linebreak mapping confirmed. */
	if math.IsNaN(vals[4]) {
		t.Fatal("bar 4: got NaN — ModeTransformed linebreak 1:1 mapping not populated")
	}
	const smaBar4 = 100104.0
	if math.Abs(vals[4]-smaBar4) > 1e-6 {
		t.Errorf("bar 4: SMA(myVar,5) = %.9f, want %.9f", vals[4], smaBar4)
	}

	/* Bar 5: SMA window slides by one position. */
	const smaBar5 = 100106.0
	if math.IsNaN(vals[5]) || math.Abs(vals[5]-smaBar5) > 1e-6 {
		t.Errorf("bar 5: SMA(myVar,5) = %.9f, want %.9f", vals[5], smaBar5)
	}

	/* Bars 4..N-1: SMA window fully populated — all must be non-NaN. */
	for i := 4; i < len(vals); i++ {
		if math.IsNaN(vals[i]) {
			t.Errorf("bar %d: unexpected NaN post-warmup", i)
		}
	}
}

/*
TestSecurityChartType_Kagi_OscillatingViaModeTransformed verifies the
ModeTransformed path for ticker.kagi with alternating high/low prices.

generateOscillatingTestOHLCV produces Close[i]=50200 (even) / 49800 (odd).
With reversal=50, a reversal triggers on every bar from i=2 onward: each
two-bar cycle produces one completed kagi segment. mapping[i]=max(0,i-2).

myVar = close*2 uses the main chart close at each segment's representative main
bar (first main bar mapping to that secBarIdx): bars 0,3,4,5,6 give closes
50200,49800,50200,49800,50200 → myVar 100400,99600,100400,99600,100400.
SMA(5) first valid at secBarIdx=4 (main bar 6): expected 100080.
*/
func TestSecurityChartType_Kagi_OscillatingViaModeTransformed(t *testing.T) {
	const barCount = 40

	strategy := `//@version=5
indicator("Kagi Transformed", overlay=false)
myVar = close * 2.0
kagiResult = request.security(ticker.kagi(syminfo.tickerid, 50), timeframe.period, ta.sma(myVar, 5))
plot(kagiResult, "KagiResult")
`
	testDir := t.TempDir()
	strategyPath := filepath.Join(testDir, "kagi-transformed.pine")
	if err := os.WriteFile(strategyPath, []byte(strategy), 0644); err != nil {
		t.Fatal(err)
	}
	dataPath := filepath.Join(testDir, "KAGITEST_1h.json")
	if err := os.WriteFile(dataPath, []byte(generateOscillatingTestOHLCV(barCount, 3600, 200)), 0644); err != nil {
		t.Fatal(err)
	}

	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(filepath.Dir(cwd))

	result := compileAndRun(t, strategyPath, dataPath, testDir, projectRoot, "KAGITEST", testDir)

	ind, ok := result.Indicators["KagiResult"]
	if !ok {
		t.Fatalf("indicator 'KagiResult' absent — ticker.kagi security not plotted")
	}
	vals := extractValues(ind.Data)

	if len(vals) != barCount {
		t.Fatalf("got %d bars, want %d", len(vals), barCount)
	}

	/* Bars 0..5: SMA warmup — fewer than 5 kagi segments formed. */
	for i := 0; i < 6; i++ {
		if !math.IsNaN(vals[i]) {
			t.Errorf("bar %d: expected NaN during kagi SMA warmup, got %.6f", i, vals[i])
		}
	}

	/* Bar 6: fifth kagi segment complete — first valid SMA(5) output. */
	if math.IsNaN(vals[6]) {
		t.Fatal("bar 6: got NaN — ModeTransformed kagi reverse bar index mapping not populated")
	}
	const smaBar6 = 100080.0
	if math.Abs(vals[6]-smaBar6) > 1e-6 {
		t.Errorf("bar 6: SMA(myVar,5) = %.9f, want %.9f", vals[6], smaBar6)
	}

	/* Bar 7: SMA window advances by one segment. */
	if math.IsNaN(vals[7]) {
		t.Errorf("bar 7: unexpected NaN post-warmup")
	}
	const smaBar7 = 99920.0
	if math.Abs(vals[7]-smaBar7) > 1e-6 {
		t.Errorf("bar 7: SMA(myVar,5) = %.9f, want %.9f", vals[7], smaBar7)
	}

	/* Bars 6..barCount-1: SMA window fully populated — all must be non-NaN. */
	for i := 6; i < len(vals); i++ {
		if math.IsNaN(vals[i]) {
			t.Errorf("bar %d: unexpected NaN post-warmup", i)
		}
	}
}

/*
TestSecurityChartType_PointFigure_OscillatingViaModeTransformed verifies the
ModeTransformed path for ticker.pointfigure with alternating high/low prices.

With source="close", style="Traditional", boxSize=50, reversal=1 (threshold=50),
the same oscillating data as the kagi test triggers a column reversal every two
bars. Column closes alternate 49800/50200 identically to kagi segments, yielding
the same SMA(5) value at main bar 6: expected 100080.
*/
func TestSecurityChartType_PointFigure_OscillatingViaModeTransformed(t *testing.T) {
	const barCount = 40

	strategy := `//@version=5
indicator("PointFigure Transformed", overlay=false)
myVar = close * 2.0
pfResult = request.security(ticker.pointfigure(syminfo.tickerid, "close", "Traditional", 50, 1), timeframe.period, ta.sma(myVar, 5))
plot(pfResult, "PFResult")
`
	testDir := t.TempDir()
	strategyPath := filepath.Join(testDir, "pf-transformed.pine")
	if err := os.WriteFile(strategyPath, []byte(strategy), 0644); err != nil {
		t.Fatal(err)
	}
	dataPath := filepath.Join(testDir, "PFTEST_1h.json")
	if err := os.WriteFile(dataPath, []byte(generateOscillatingTestOHLCV(barCount, 3600, 200)), 0644); err != nil {
		t.Fatal(err)
	}

	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(filepath.Dir(cwd))

	result := compileAndRun(t, strategyPath, dataPath, testDir, projectRoot, "PFTEST", testDir)

	ind, ok := result.Indicators["PFResult"]
	if !ok {
		t.Fatalf("indicator 'PFResult' absent — ticker.pointfigure security not plotted")
	}
	vals := extractValues(ind.Data)

	if len(vals) != barCount {
		t.Fatalf("got %d bars, want %d", len(vals), barCount)
	}

	/* Bars 0..5: SMA warmup — fewer than 5 P&F columns formed. */
	for i := 0; i < 6; i++ {
		if !math.IsNaN(vals[i]) {
			t.Errorf("bar %d: expected NaN during P&F SMA warmup, got %.6f", i, vals[i])
		}
	}

	/* Bar 6: fifth P&F column complete — first valid SMA(5) output. */
	if math.IsNaN(vals[6]) {
		t.Fatal("bar 6: got NaN — ModeTransformed pointfigure reverse bar index mapping not populated")
	}
	const smaBar6 = 100080.0
	if math.Abs(vals[6]-smaBar6) > 1e-6 {
		t.Errorf("bar 6: SMA(myVar,5) = %.9f, want %.9f", vals[6], smaBar6)
	}

	/* Bar 7: SMA window advances by one column. */
	if math.IsNaN(vals[7]) {
		t.Errorf("bar 7: unexpected NaN post-warmup")
	}
	const smaBar7 = 99920.0
	if math.Abs(vals[7]-smaBar7) > 1e-6 {
		t.Errorf("bar 7: SMA(myVar,5) = %.9f, want %.9f", vals[7], smaBar7)
	}

	/* Bars 6..barCount-1: SMA window fully populated — all must be non-NaN. */
	for i := 6; i < len(vals); i++ {
		if math.IsNaN(vals[i]) {
			t.Errorf("bar %d: unexpected NaN post-warmup", i)
		}
	}
}
