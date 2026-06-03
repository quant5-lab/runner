package regression

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/quant5-lab/runner/runtime/context"
	"github.com/quant5-lab/runner/runtime/market"
)

type referenceSessionChartOutput struct {
	Candlestick []context.OHLCV                      `json:"candlestick"`
	Indicators  map[string]referenceSessionIndicator `json:"indicators"`
}

type referenceSessionIndicator struct {
	Data []referenceSessionPoint `json:"data"`
}

type referenceSessionPoint struct {
	Value *float64 `json:"value"`
}

func TestReferenceSession_RegularExchangeFixtureFiltersPreSessionAuctionBarsAfterNormalization(t *testing.T) {
	root := projectRootFromCwd()
	fixtureBars := loadOHLCVBars(t, filepath.Join(root, "tests", "golden", "fixtures", "data", "SBERP-1h.json"))

	normalized, timezone, referenceSession := market.NormalizeBarsWithReferenceSession("SBERP", "1h", "", "regular", fixtureBarsToRuntimeBars(fixtureBars))

	if timezone != "Europe/Moscow" {
		t.Fatalf("timezone = %q, want Europe/Moscow", timezone)
	}
	if referenceSession != market.ReferenceSessionRegular {
		t.Fatalf("reference session = %q, want regular", referenceSession)
	}
	if len(normalized) >= len(fixtureBars) {
		t.Fatalf("regular exchange normalization removed no bars: before=%d after=%d", len(fixtureBars), len(normalized))
	}
	// MOEX trades 7 days/week; regular session filters by time window (07:00–23:50 MSK) only.
	// Weekend bars within the window are preserved; pre-session auction bars (< 07:00) are removed.
	loc := mustLocation(t, timezone)
	for _, bar := range normalized {
		instant := runtimeBarInstant(bar).In(loc)
		h, m := instant.Hour(), instant.Minute()
		if h*60+m < 7*60 {
			t.Fatalf("pre-session bar survived normalization: %s", instant)
		}
	}
}

func TestReferenceSession_AlwaysOpenFixturePreservesWeekendBars(t *testing.T) {
	root := projectRootFromCwd()
	bars := loadOHLCVBars(t, filepath.Join(root, "tests", "golden", "fixtures", "data", "BTCUSDT-1h.json"))

	normalized, timezone := market.NormalizeBars("BTCUSDT", "1h", "", fixtureBarsToRuntimeBars(bars))

	if timezone != "UTC" {
		t.Fatalf("timezone = %q, want UTC", timezone)
	}
	if len(normalized) != len(bars) {
		t.Fatalf("BTCUSDT bars changed: before=%d after=%d", len(bars), len(normalized))
	}
	if countWeekdayBars(normalized, time.Saturday, timezone) == 0 || countWeekdayBars(normalized, time.Sunday, timezone) == 0 {
		t.Fatal("always-open fixture lost weekend coverage")
	}
}

func TestReferenceSession_GeneratedRunnerPreservesMidDayBarsFromMetadata(t *testing.T) {
	root := projectRootFromCwd()
	tmpDir := t.TempDir()
	bars := moscowRuntimeWeekdayWeekendBars(t)
	dataPath := writeReferenceSessionBars(t, tmpDir, "SBERP_1h.json", "Europe/Moscow", "regular", bars)
	outputPath := runReferenceSessionProbe(t, root, tmpDir, dataPath, "SBERP", "1h")
	output := readReferenceSessionChartOutput(t, outputPath)

	if len(output.Candlestick) != len(bars) {
		t.Fatalf("candlestick bars = %d, want %d (MOEX regular session must keep all mid-day bars including weekends)", len(output.Candlestick), len(bars))
	}
}

func TestReferenceSession_GeneratedRunnerPreservesAlwaysOpenChartBars(t *testing.T) {
	root := projectRootFromCwd()
	tmpDir := t.TempDir()
	bars := utcRuntimeWeekdayWeekendBars()
	dataPath := writeReferenceSessionBars(t, tmpDir, "BTCUSDT_1h.json", "UTC", "", bars)
	outputPath := runReferenceSessionProbe(t, root, tmpDir, dataPath, "BTCUSDT", "1h")
	output := readReferenceSessionChartOutput(t, outputPath)

	if len(output.Candlestick) != len(bars) {
		t.Fatalf("candlestick bars = %d, want %d", len(output.Candlestick), len(bars))
	}
	if countWeekdayBars(output.Candlestick, time.Saturday, "UTC") == 0 || countWeekdayBars(output.Candlestick, time.Sunday, "UTC") == 0 {
		t.Fatal("generated runner lost always-open weekend bars")
	}
}

func TestReferenceSession_GeneratedRunnerHonorsExplicitBarUniverse(t *testing.T) {
	root := projectRootFromCwd()
	tmpDir := t.TempDir()
	bars := utcRuntimeWeekdayWeekendBars()
	dataPath := writeReferenceSessionFixture(t, tmpDir, "CUSTOM_1h.json", market.SourceMetadata{
		Timezone:           "UTC",
		ReferenceSession:   "regular",
		IncludedTimestamps: []int64{bars[1].Time, bars[3].Time},
	}, bars)
	outputPath := runReferenceSessionProbe(t, root, tmpDir, dataPath, "CUSTOM", "1h")
	output := readReferenceSessionChartOutput(t, outputPath)

	assertRuntimeCloseSequence(t, output.Candlestick, []float64{2, 4})
}

func TestReferenceSession_GeneratedRunnerNormalizesSecurityBarsFromMetadata(t *testing.T) {
	root := projectRootFromCwd()
	tmpDir := t.TempDir()
	mainBars := utcRuntimeLongWeekendBars()
	dailyBars := utcRuntimeLongWeekendBars()
	writeReferenceSessionBars(t, tmpDir, "SBERP_1h.json", "Europe/Moscow", "", mainBars)
	writeReferenceSessionFixture(t, tmpDir, "SBERP_1D.json", market.SourceMetadata{
		Exchange:         "MOEX",
		Timezone:         "Europe/Moscow",
		ReferenceSession: "regular",
	}, dailyBars)

	outputPath := runSecurityReferenceSessionProbe(t, root, tmpDir, filepath.Join(tmpDir, "SBERP_1h.json"), "SBERP", "1h")
	output := readReferenceSessionChartOutput(t, outputPath)

	assertIndicatorValues(t, output, "Security close", []float64{1, 1, 2, 3, 4})
}

func TestReferenceSession_GeneratedRunnerPreservesAlwaysOpenSecurityBars(t *testing.T) {
	root := projectRootFromCwd()
	tmpDir := t.TempDir()
	mainBars := utcRuntimeLongWeekendBars()
	dailyBars := utcRuntimeLongWeekendBars()
	writeReferenceSessionBars(t, tmpDir, "BTCUSDT_1h.json", "UTC", "", mainBars)
	writeReferenceSessionBars(t, tmpDir, "BTCUSDT_1D.json", "UTC", "", dailyBars)

	outputPath := runSecurityReferenceSessionProbe(t, root, tmpDir, filepath.Join(tmpDir, "BTCUSDT_1h.json"), "BTCUSDT", "1h")
	output := readReferenceSessionChartOutput(t, outputPath)

	assertIndicatorValues(t, output, "Security close", []float64{1, 1, 2, 3, 4})
}

func TestReferenceSession_GeneratedRunnerInheritsRegularSessionForUnannotatedSecurityBars(t *testing.T) {
	root := projectRootFromCwd()
	tmpDir := t.TempDir()
	mainBars := utcRuntimeLongWeekendBars()
	dailyBars := utcRuntimeLongWeekendBars()
	writeReferenceSessionFixture(t, tmpDir, "SBERP_1h.json", market.SourceMetadata{
		Exchange:         "MOEX",
		Timezone:         "Europe/Moscow",
		ReferenceSession: "regular",
	}, mainBars)
	writeReferenceSessionFixture(t, tmpDir, "SBERP_1D.json", market.SourceMetadata{}, dailyBars)

	outputPath := runSecurityReferenceSessionProbe(t, root, tmpDir, filepath.Join(tmpDir, "SBERP_1h.json"), "SBERP", "1h")
	output := readReferenceSessionChartOutput(t, outputPath)

	assertRuntimeCloseSequence(t, output.Candlestick, []float64{1, 2, 3, 4, 5})
	assertIndicatorValues(t, output, "Security close", []float64{1, 1, 2, 3, 4})
}

func TestReferenceSession_GeneratedRunnerInheritsAlwaysOpenSessionForUnannotatedSecurityBars(t *testing.T) {
	root := projectRootFromCwd()
	tmpDir := t.TempDir()
	mainBars := utcRuntimeLongWeekendBars()
	dailyBars := utcRuntimeLongWeekendBars()
	writeReferenceSessionFixture(t, tmpDir, "SBERP_1h.json", market.SourceMetadata{
		Exchange:         "MOEX",
		Timezone:         "Europe/Moscow",
		ReferenceSession: "always-open",
	}, mainBars)
	writeReferenceSessionFixture(t, tmpDir, "SBERP_1D.json", market.SourceMetadata{}, dailyBars)

	outputPath := runSecurityReferenceSessionProbe(t, root, tmpDir, filepath.Join(tmpDir, "SBERP_1h.json"), "SBERP", "1h")
	output := readReferenceSessionChartOutput(t, outputPath)

	assertRuntimeCloseSequence(t, output.Candlestick, []float64{1, 2, 3, 4, 5})
	assertIndicatorValues(t, output, "Security close", []float64{1, 1, 2, 3, 4})
}

func fixtureBarsToRuntimeBars(bars []ohlcvFixtureBar) []context.OHLCV {
	converted := make([]context.OHLCV, len(bars))
	for i, bar := range bars {
		converted[i] = context.OHLCV{
			Time:   bar.Time,
			Open:   bar.Open,
			High:   bar.High,
			Low:    bar.Low,
			Close:  bar.Close,
			Volume: bar.Volume,
		}
	}
	return converted
}

func countWeekdayBars(bars []context.OHLCV, weekday time.Weekday, timezone string) int {
	location, err := time.LoadLocation(timezone)
	if err != nil {
		location = time.UTC
	}
	count := 0
	for _, bar := range bars {
		if runtimeBarInstant(bar).In(location).Weekday() == weekday {
			count++
		}
	}
	return count
}

func runtimeBarInstant(bar context.OHLCV) time.Time {
	if bar.Time > 10_000_000_000 {
		return time.UnixMilli(bar.Time).UTC()
	}
	return time.Unix(bar.Time, 0).UTC()
}

func mustLocation(t *testing.T, name string) *time.Location {
	t.Helper()
	location, err := time.LoadLocation(name)
	if err != nil {
		t.Fatalf("load location %q: %v", name, err)
	}
	return location
}

func moscowRuntimeWeekdayWeekendBars(t *testing.T) []context.OHLCV {
	t.Helper()
	return []context.OHLCV{
		{Time: unixInRegressionLocation(t, "Europe/Moscow", "2025-08-15 13:00"), Open: 1, High: 1, Low: 1, Close: 1, Volume: 1},
		{Time: unixInRegressionLocation(t, "Europe/Moscow", "2025-08-16 13:00"), Open: 2, High: 2, Low: 2, Close: 2, Volume: 1},
		{Time: unixInRegressionLocation(t, "Europe/Moscow", "2025-08-17 13:00"), Open: 3, High: 3, Low: 3, Close: 3, Volume: 1},
		{Time: unixInRegressionLocation(t, "Europe/Moscow", "2025-08-18 13:00"), Open: 4, High: 4, Low: 4, Close: 4, Volume: 1},
	}
}

func utcRuntimeWeekdayWeekendBars() []context.OHLCV {
	return []context.OHLCV{
		{Time: time.Date(2025, 8, 15, 13, 0, 0, 0, time.UTC).Unix(), Open: 1, High: 1, Low: 1, Close: 1, Volume: 1},
		{Time: time.Date(2025, 8, 16, 13, 0, 0, 0, time.UTC).Unix(), Open: 2, High: 2, Low: 2, Close: 2, Volume: 1},
		{Time: time.Date(2025, 8, 17, 13, 0, 0, 0, time.UTC).Unix(), Open: 3, High: 3, Low: 3, Close: 3, Volume: 1},
		{Time: time.Date(2025, 8, 18, 13, 0, 0, 0, time.UTC).Unix(), Open: 4, High: 4, Low: 4, Close: 4, Volume: 1},
	}
}

func utcRuntimeLongWeekendBars() []context.OHLCV {
	return []context.OHLCV{
		{Time: time.Date(2025, 8, 15, 13, 0, 0, 0, time.UTC).Unix(), Open: 1, High: 1, Low: 1, Close: 1, Volume: 1},
		{Time: time.Date(2025, 8, 16, 13, 0, 0, 0, time.UTC).Unix(), Open: 2, High: 2, Low: 2, Close: 2, Volume: 1},
		{Time: time.Date(2025, 8, 17, 13, 0, 0, 0, time.UTC).Unix(), Open: 3, High: 3, Low: 3, Close: 3, Volume: 1},
		{Time: time.Date(2025, 8, 18, 13, 0, 0, 0, time.UTC).Unix(), Open: 4, High: 4, Low: 4, Close: 4, Volume: 1},
		{Time: time.Date(2025, 8, 19, 13, 0, 0, 0, time.UTC).Unix(), Open: 5, High: 5, Low: 5, Close: 5, Volume: 1},
	}
}

func unixInRegressionLocation(t *testing.T, locationName, value string) int64 {
	t.Helper()
	location := mustLocation(t, locationName)
	parsed, err := time.ParseInLocation("2006-01-02 15:04", value, location)
	if err != nil {
		t.Fatalf("parse %q in %s: %v", value, locationName, err)
	}
	return parsed.Unix()
}

func writeReferenceSessionBars(t *testing.T, dir, filename, timezone, referenceSession string, bars []context.OHLCV) string {
	t.Helper()
	return writeReferenceSessionFixture(t, dir, filename, market.SourceMetadata{Timezone: timezone, ReferenceSession: referenceSession}, bars)
}

func writeReferenceSessionFixture(t *testing.T, dir, filename string, metadata market.SourceMetadata, bars []context.OHLCV) string {
	t.Helper()
	payload := struct {
		market.SourceMetadata
		Bars []context.OHLCV `json:"bars"`
	}{SourceMetadata: metadata, Bars: bars}
	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal fixture: %v", err)
	}
	path := filepath.Join(dir, filename)
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	return path
}

func assertRuntimeCloseSequence(t *testing.T, got []context.OHLCV, want []float64) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("candlestick bars = %d, want %d: %+v", len(got), len(want), got)
	}
	for i := range want {
		if got[i].Close != want[i] {
			t.Fatalf("bar %d close = %.2f, want %.2f", i, got[i].Close, want[i])
		}
	}
}

func runReferenceSessionProbe(t *testing.T, root, tmpDir, dataPath, symbol, timeframe string) string {
	t.Helper()
	const source = `//@version=4
strategy("Reference Session Probe", overlay=true)
plot(close)
`
	return runReferenceSessionStrategy(t, root, tmpDir, dataPath, symbol, timeframe, "reference_session_probe", source)
}

func runSecurityReferenceSessionProbe(t *testing.T, root, tmpDir, dataPath, symbol, timeframe string) string {
	t.Helper()
	const source = `//@version=4
strategy("Reference Session Security Probe", overlay=true)
secClose = security(syminfo.tickerid, "1D", close)
plot(secClose, title="Security close")
`
	return runReferenceSessionStrategy(t, root, tmpDir, dataPath, symbol, timeframe, "reference_session_security_probe", source)
}

func runReferenceSessionStrategy(t *testing.T, root, tmpDir, dataPath, symbol, timeframe, name, source string) string {
	t.Helper()
	built, ok := codegenAndBuild(t, tmpDir, name, source, root)
	if !ok {
		t.Fatal("strategy codegen/build failed")
	}
	outputPath := filepath.Join(tmpDir, "chart-data.json")
	cmd := exec.Command(
		built.BinaryPath,
		"-symbol", symbol,
		"-timeframe", timeframe,
		"-data", dataPath,
		"-datadir", filepath.Dir(dataPath),
		"-output", outputPath,
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("run generated fixture: %v\n%s", err, output)
	}
	return outputPath
}

func assertIndicatorValues(t *testing.T, output referenceSessionChartOutput, name string, want []float64) {
	t.Helper()
	indicator, ok := output.Indicators[name]
	if !ok {
		t.Fatalf("indicator %q missing", name)
	}
	if len(indicator.Data) != len(want) {
		t.Fatalf("indicator %q points = %d, want %d", name, len(indicator.Data), len(want))
	}
	for i := range want {
		if indicator.Data[i].Value == nil {
			t.Fatalf("indicator %q[%d] = nil, want %.2f", name, i, want[i])
		}
		if *indicator.Data[i].Value != want[i] {
			t.Fatalf("indicator %q[%d] = %.2f, want %.2f", name, i, *indicator.Data[i].Value, want[i])
		}
	}
}

func readReferenceSessionChartOutput(t *testing.T, path string) referenceSessionChartOutput {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read chart output: %v", err)
	}
	var output referenceSessionChartOutput
	if err := json.Unmarshal(data, &output); err != nil {
		t.Fatalf("parse chart output: %v", err)
	}
	return output
}
