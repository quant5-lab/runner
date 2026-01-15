package regression

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

/* TestSecurity_NonOverlappingRanges_Regression validates Bug #2 fix */
func TestSecurity_NonOverlappingRanges_Regression(t *testing.T) {
	testDir := t.TempDir()

	strategy := `//@version=6
strategy("Bug #2 Non-Overlapping Test", overlay=true)

dailyOpen = request.security(syminfo.tickerid, "1D", open, lookahead=barmerge.lookahead_off)

plot(dailyOpen, "Daily Open", color=color.blue)
`
	strategyPath := filepath.Join(testDir, "test_strategy.pine")
	if err := os.WriteFile(strategyPath, []byte(strategy), 0644); err != nil {
		t.Fatal(err)
	}

	hourlyData := generateTestOHLCVWithStartDate(891, 3600, 1720396800)
	hourlyPath := filepath.Join(testDir, "AAPL_1h.json")
	if err := os.WriteFile(hourlyPath, []byte(hourlyData), 0644); err != nil {
		t.Fatal(err)
	}

	dailyData := generateTestOHLCVWithStartDate(100, 86400, 1720396800)
	dailyPath := filepath.Join(testDir, "AAPL_1D.json")
	if err := os.WriteFile(dailyPath, []byte(dailyData), 0644); err != nil {
		t.Fatal(err)
	}

	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(filepath.Dir(cwd))

	result := compileAndRun(t, strategyPath, hourlyPath, testDir, projectRoot, "AAPL", testDir)

	dailyOpen, ok := result.Indicators["Daily Open"]
	if !ok {
		t.Fatalf("Expected 'Daily Open' indicator, got: %v", getIndicatorNames(result.Indicators))
	}

	openCount := countNonNull(dailyOpen.Data)

	if openCount < 850 {
		t.Errorf("Bug #2 Regression: Daily Open has only %d non-null values, expected >850", openCount)
	}
}

/* TestSecurity_FirstBarLookahead_Regression validates Bug #1 fix */
func TestSecurity_FirstBarLookahead_Regression(t *testing.T) {
	testDir := t.TempDir()

	strategy := `//@version=6
strategy("Bug #1 First Bar Test", overlay=true)

dailyOpen = request.security(syminfo.tickerid, "1D", open, lookahead=barmerge.lookahead_off)

plot(dailyOpen, "Daily Open", color=color.green)
`
	strategyPath := filepath.Join(testDir, "test_strategy.pine")
	if err := os.WriteFile(strategyPath, []byte(strategy), 0644); err != nil {
		t.Fatal(err)
	}

	hourlyData := generateTestOHLCV(240, 3600)
	hourlyPath := filepath.Join(testDir, "FIRSTBAR_1h.json")
	if err := os.WriteFile(hourlyPath, []byte(hourlyData), 0644); err != nil {
		t.Fatal(err)
	}

	dailyData := generateTestOHLCV(10, 86400)
	dailyPath := filepath.Join(testDir, "FIRSTBAR_1D.json")
	if err := os.WriteFile(dailyPath, []byte(dailyData), 0644); err != nil {
		t.Fatal(err)
	}

	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(filepath.Dir(cwd))

	result := compileAndRun(t, strategyPath, hourlyPath, testDir, projectRoot, "FIRSTBAR", testDir)

	dailyOpen, ok := result.Indicators["Daily Open"]
	if !ok {
		t.Fatalf("Expected 'Daily Open' indicator")
	}

	if len(dailyOpen.Data) == 0 {
		t.Fatal("No data in Daily Open indicator")
	}

	if len(dailyOpen.Data) > 0 {
		if _, ok := getFloatValue(dailyOpen.Data[0]); !ok {
			t.Errorf("Bug #1 Regression: First bar is null")
		}
	}
}

/* TestSecurity_Upscaling_Complete tests weekly data requested from daily base */
func TestSecurity_Upscaling_Complete(t *testing.T) {
	testDir := t.TempDir()

	strategy := `//@version=6
strategy("Upscaling Test", overlay=true)

weeklyHigh = request.security(syminfo.tickerid, "1W", high, lookahead=barmerge.lookahead_off)

plot(weeklyHigh, "Weekly High", color=color.orange)
`
	strategyPath := filepath.Join(testDir, "test_strategy.pine")
	if err := os.WriteFile(strategyPath, []byte(strategy), 0644); err != nil {
		t.Fatal(err)
	}

	dailyData := generateTestOHLCV(50, 86400)
	dailyPath := filepath.Join(testDir, "UPTEST_1D.json")
	if err := os.WriteFile(dailyPath, []byte(dailyData), 0644); err != nil {
		t.Fatal(err)
	}

	weeklyData := generateTestOHLCV(10, 604800)
	weeklyPath := filepath.Join(testDir, "UPTEST_1W.json")
	if err := os.WriteFile(weeklyPath, []byte(weeklyData), 0644); err != nil {
		t.Fatal(err)
	}

	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(filepath.Dir(cwd))

	result := compileAndRun(t, strategyPath, dailyPath, testDir, projectRoot, "UPTEST", testDir)

	weeklyHigh, ok := result.Indicators["Weekly High"]
	if !ok {
		t.Fatalf("Expected 'Weekly High' indicator")
	}

	nonNullCount := countNonNull(weeklyHigh.Data)

	if nonNullCount < 45 {
		t.Errorf("Upscaling: Expected ~50 non-null values, got %d", nonNullCount)
	}
}

/* TestSecurity_SameTimeframe_Complete tests requesting same timeframe */
func TestSecurity_SameTimeframe_Complete(t *testing.T) {
	testDir := t.TempDir()

	strategy := `//@version=6
strategy("Same Timeframe Test", overlay=true)

sameClose = request.security(syminfo.tickerid, "1D", close, lookahead=barmerge.lookahead_off)

plot(sameClose, "Same TF Close", color=color.purple)
`
	strategyPath := filepath.Join(testDir, "test_strategy.pine")
	if err := os.WriteFile(strategyPath, []byte(strategy), 0644); err != nil {
		t.Fatal(err)
	}

	dailyData := generateTestOHLCV(100, 86400)
	dailyPath := filepath.Join(testDir, "SAMETEST_1D.json")
	if err := os.WriteFile(dailyPath, []byte(dailyData), 0644); err != nil {
		t.Fatal(err)
	}

	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(filepath.Dir(cwd))

	result := compileAndRun(t, strategyPath, dailyPath, testDir, projectRoot, "SAMETEST", testDir)

	sameClose, ok := result.Indicators["Same TF Close"]
	if !ok {
		t.Fatalf("Expected 'Same TF Close' indicator")
	}

	nonNullCount := countNonNull(sameClose.Data)

	if nonNullCount < 95 {
		t.Errorf("Same-timeframe: Expected ~100 non-null values, got %d", nonNullCount)
	}

	if len(sameClose.Data) > 0 {
		if firstVal, ok := getFloatValue(sameClose.Data[0]); ok {
			expectedFirst := 50050.0
			if firstVal != expectedFirst {
				t.Errorf("First bar value mismatch: got %.2f, expected %.2f", firstVal, expectedFirst)
			}
		}
	}
}

/* TestSecurity_Downscaling_WithValidation tests requesting daily data from hourly base */
func TestSecurity_Downscaling_WithValidation(t *testing.T) {
	testDir := t.TempDir()

	strategy := `//@version=6
strategy("Downscaling Test", overlay=true)

dailySMA = request.security(syminfo.tickerid, "1D", ta.sma(close, 5), lookahead=barmerge.lookahead_off)
dailyOpen = request.security(syminfo.tickerid, "1D", open, lookahead=barmerge.lookahead_off)

plot(dailySMA, "Daily SMA5", color=color.blue)
plot(dailyOpen, "Daily Open", color=color.green)
`
	strategyPath := filepath.Join(testDir, "test_strategy.pine")
	if err := os.WriteFile(strategyPath, []byte(strategy), 0644); err != nil {
		t.Fatal(err)
	}

	hourlyData := generateTestOHLCV(240, 3600)
	hourlyPath := filepath.Join(testDir, "DOWNTEST_1h.json")
	if err := os.WriteFile(hourlyPath, []byte(hourlyData), 0644); err != nil {
		t.Fatal(err)
	}

	dailyData := generateTestOHLCV(10, 86400)
	dailyPath := filepath.Join(testDir, "DOWNTEST_1D.json")
	if err := os.WriteFile(dailyPath, []byte(dailyData), 0644); err != nil {
		t.Fatal(err)
	}

	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(filepath.Dir(cwd))

	result := compileAndRun(t, strategyPath, hourlyPath, testDir, projectRoot, "DOWNTEST", testDir)

	dailySMA, ok := result.Indicators["Daily SMA5"]
	if !ok {
		t.Fatalf("Expected 'Daily SMA5' indicator")
	}

	dailyOpen, ok := result.Indicators["Daily Open"]
	if !ok {
		t.Fatalf("Expected 'Daily Open' indicator")
	}

	smaCount := countNonNull(dailySMA.Data)
	openCount := countNonNull(dailyOpen.Data)

	if smaCount < 100 {
		t.Errorf("Downscaling Daily SMA5: only %d non-null values, expected >100", smaCount)
	}

	if openCount < 235 {
		t.Errorf("Downscaling Daily Open: only %d non-null values, expected ~240", openCount)
	}
}

/* ========== HELPER FUNCTIONS ========== */

func generateTestOHLCVWithStartDate(bars int, intervalSec int, startUnix int64) string {
	type Bar struct {
		Time   int64   `json:"time"`
		Open   float64 `json:"open"`
		High   float64 `json:"high"`
		Low    float64 `json:"low"`
		Close  float64 `json:"close"`
		Volume float64 `json:"volume"`
	}

	var data []Bar
	for i := 0; i < bars; i++ {
		timestamp := startUnix + int64(i*intervalSec)
		open := 50000.0 + float64(i*100)
		high := open + 75.0
		low := open - 25.0
		close := open + 50.0
		volume := 1000000.0 + float64(i*1000)

		data = append(data, Bar{
			Time:   timestamp * 1000,
			Open:   open,
			High:   high,
			Low:    low,
			Close:  close,
			Volume: volume,
		})
	}

	jsonData, _ := json.Marshal(data)
	return string(jsonData)
}

func compileAndRun(t *testing.T, strategyPath, dataPath, testDir, projectRoot, symbol, dataDir string) TestResult {
	t.Helper()

	outputPath := filepath.Join(testDir, "strategy.go")
	builderPath := filepath.Join(projectRoot, "cmd", "pine-gen", "main.go")
	templatePath := filepath.Join(projectRoot, "template", "main.go.tmpl")

	compileCmd := exec.Command(
		"go", "run", builderPath,
		"-input", strategyPath,
		"-output", outputPath,
		"-template", templatePath,
	)
	compileOutput, err := compileCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Pine compilation failed: %v\n%s", err, compileOutput)
	}

	generatedFile := outputPath
	lines := []byte(compileOutput)
	for i := 0; i < len(lines); i++ {
		if i+11 < len(lines) && string(lines[i:i+11]) == "Generated: " {
			start := i + 11
			end := start
			for end < len(lines) && lines[end] != '\n' {
				end++
			}
			generatedFile = string(lines[start:end])
			generatedFile = filepath.Clean(generatedFile)
			break
		}
	}

	generatedContent, err := os.ReadFile(generatedFile)
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}
	localGenFile := filepath.Join(testDir, "strategy.go")
	if err := os.WriteFile(localGenFile, generatedContent, 0644); err != nil {
		t.Fatalf("Failed to copy generated file: %v", err)
	}

	if err := setupGoMod(localGenFile, projectRoot); err != nil {
		t.Fatalf("Failed to setup go.mod: %v", err)
	}

	tidyCmd := exec.Command("go", "mod", "tidy")
	tidyCmd.Dir = testDir
	if output, err := tidyCmd.CombinedOutput(); err != nil {
		t.Fatalf("go mod tidy failed: %v\n%s", err, output)
	}

	exePath := filepath.Join(testDir, "strategy")
	buildCmd := exec.Command("go", "build", "-o", exePath, localGenFile)
	buildCmd.Dir = testDir
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Go build failed: %v\n%s", err, output)
	}

	resultPath := filepath.Join(testDir, "result.json")
	runCmd := exec.Command(exePath, "-symbol", symbol, "-data", dataPath, "-datadir", testDir, "-output", resultPath)
	if output, err := runCmd.CombinedOutput(); err != nil {
		t.Fatalf("Strategy execution failed: %v\n%s", err, output)
	}

	outputData, err := os.ReadFile(resultPath)
	if err != nil {
		t.Fatalf("Failed to read result file: %v", err)
	}

	var result TestResult
	if err := json.Unmarshal(outputData, &result); err != nil {
		t.Fatalf("Failed to parse JSON output: %v\nOutput: %s", err, outputData)
	}

	return result
}

type TestResult struct {
	Indicators map[string]IndicatorData `json:"indicators"`
}

type IndicatorData struct {
	Data []map[string]interface{} `json:"data"`
}

func countNonNull(data []map[string]interface{}) int {
	count := 0
	for _, bar := range data {
		if val, ok := bar["value"]; ok && val != nil {
			count++
		}
	}
	return count
}

func getFloatValue(bar map[string]interface{}) (float64, bool) {
	if val, ok := bar["value"]; ok && val != nil {
		if fval, ok := val.(float64); ok {
			return fval, true
		}
	}
	return 0, false
}

func getIndicatorNames(indicators map[string]IndicatorData) []string {
	names := make([]string, 0, len(indicators))
	for name := range indicators {
		names = append(names, name)
	}
	return names
}
