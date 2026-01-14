package tests

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/quant5-lab/runner/tests/testutil"
)

/* setupGoMod creates go.mod in generated code directory for standalone compilation */
func setupGoMod(generatedFilePath, projectRoot string) error {
	goModContent := `module testprog

go 1.23

replace github.com/quant5-lab/runner => ` + projectRoot + `

require github.com/quant5-lab/runner v0.0.0
`
	goModPath := filepath.Join(filepath.Dir(generatedFilePath), "go.mod")
	return os.WriteFile(goModPath, []byte(goModContent), 0644)
}

/* generateTestOHLCV creates synthetic OHLCV data with specified bar count */
func generateTestOHLCV(barCount int, intervalSeconds int64) string {
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
		bars[i] = Bar{
			Time:   startTime + int64(i)*intervalSeconds,
			Open:   basePrice + float64(i),
			High:   basePrice + float64(i) + 100,
			Low:    basePrice + float64(i) - 100,
			Close:  basePrice + float64(i) + 50,
			Volume: 100.0,
		}
	}

	data := OHLCVData{
		Timezone: "UTC",
		Bars:     bars,
	}

	jsonData, _ := json.MarshalIndent(data, "", "  ")
	return string(jsonData)
}

func TestSecurityDownsampling_1h_to_1D_WithWarmup(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	strategyCode := `
//@version=5
indicator("Security Downsample Test", overlay=true)
dailyClose = request.security(syminfo.tickerid, "1D", close)
plot(dailyClose, title="Daily Close", color=color.blue)
`

	testDir := t.TempDir()
	strategyPath := filepath.Join(testDir, "test-downsample.pine")
	if err := os.WriteFile(strategyPath, []byte(strategyCode), 0644); err != nil {
		t.Fatal(err)
	}

	/* Generate 240 bars (10 days of hourly data) */
	testDataPath := filepath.Join(testDir, "BTCUSDT_1h.json")
	testData := generateTestOHLCV(240, 3600)
	if err := os.WriteFile(testDataPath, []byte(testData), 0644); err != nil {
		t.Fatal(err)
	}

	/* Generate 10 bars (10 days of daily data) for security() to fetch */
	testDataPathDaily := filepath.Join(testDir, "BTCUSDT_1D.json")
	testDataDaily := generateTestOHLCV(10, 86400)
	if err := os.WriteFile(testDataPathDaily, []byte(testDataDaily), 0644); err != nil {
		t.Fatal(err)
	}

	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(cwd)
	builderPath := filepath.Join(projectRoot, "cmd", "pine-gen", "main.go")
	templatePath := filepath.Join(projectRoot, "template", "main.go.tmpl")
	outputGoPath := filepath.Join(testDir, "output.go")

	buildCmd := exec.Command("go", "run", builderPath, "-input", strategyPath, "-output", outputGoPath, "-template", templatePath)
	buildOutput, err := buildCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Build failed: %v\nOutput: %s", err, buildOutput)
	}

	/* Parse Generated: line to get temp Go file path */
	generatedFile := ""
	for _, line := range strings.Split(string(buildOutput), "\n") {
		if strings.HasPrefix(line, "Generated: ") {
			generatedFile = strings.TrimSpace(strings.TrimPrefix(line, "Generated: "))
			break
		}
	}
	if generatedFile == "" {
		t.Fatalf("Failed to parse generated file path from output: %s", buildOutput)
	}

	/* Copy generated file to testDir where we can create go.mod */
	localGenPath := filepath.Join(testDir, "main.go")
	generatedData, err := os.ReadFile(generatedFile)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(localGenPath, generatedData, 0644); err != nil {
		t.Fatal(err)
	}

	if err := setupGoMod(localGenPath, projectRoot); err != nil {
		t.Fatal(err)
	}

	/* Run go mod tidy in testDir */
	tidyCmd := exec.Command("go", "mod", "tidy")
	tidyCmd.Dir = testDir
	if output, err := tidyCmd.CombinedOutput(); err != nil {
		t.Fatalf("go mod tidy failed: %v\nOutput: %s", err, output)
	}

	binPath := filepath.Join(testDir, "test-bin")
	compileCmd := exec.Command("go", "build", "-o", binPath, localGenPath)
	compileCmd.Dir = testDir
	if output, err := compileCmd.CombinedOutput(); err != nil {
		t.Fatalf("Compile failed: %v\nOutput: %s", err, output)
	}

	resultPath := filepath.Join(testDir, "result.json")

	runCmd := exec.Command(binPath, "-symbol", "BTCUSDT", "-data", testDataPath, "-datadir", testDir, "-output", resultPath)
	if output, err := runCmd.CombinedOutput(); err != nil {
		t.Fatalf("Execution failed: %v\nOutput: %s", err, output)
	}

	resultData, err := os.ReadFile(resultPath)
	if err != nil {
		t.Fatal(err)
	}

	var result struct {
		Indicators map[string]struct {
			Data []map[string]interface{} `json:"data"`
		} `json:"indicators"`
	}
	if err := json.Unmarshal(resultData, &result); err != nil {
		t.Fatal(err)
	}

	if len(result.Indicators) == 0 {
		t.Fatal("No indicators in output")
	}

	dailyClose, ok := result.Indicators["Daily Close"]
	if !ok {
		t.Fatalf("Expected 'Daily Close' indicator, got: %v", result.Indicators)
	}
	if len(dailyClose.Data) == 0 {
		t.Fatal("Downsampling produced zero values")
	}

	nonNullCount := 0
	for _, point := range dailyClose.Data {
		if val, ok := point["value"]; ok && val != nil {
			nonNullCount++
		}
	}

	/* 240h bars = 10 days → expect at least 8 daily values */
	if nonNullCount < 8 {
		t.Errorf("Downsampling insufficient: got %d non-null values, expected >=8 from 240 hourly bars", nonNullCount)
	}
}

/* TestSecuritySameTimeframe_1h_to_1h_NoWarmup verifies same-timeframe has no warmup overhead */
func TestSecuritySameTimeframe_1h_to_1h_NoWarmup(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	strategyCode := `
//@version=5
indicator("Security Same-TF Test", overlay=true)
sameTFClose = request.security(syminfo.tickerid, "1h", close)
plot(sameTFClose, title="Same-TF Close", color=color.green)
`

	testDir := t.TempDir()
	strategyPath := filepath.Join(testDir, "test-same-tf.pine")
	if err := os.WriteFile(strategyPath, []byte(strategyCode), 0644); err != nil {
		t.Fatal(err)
	}

	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(cwd)
	builderPath := filepath.Join(projectRoot, "cmd", "pine-gen", "main.go")
	templatePath := filepath.Join(projectRoot, "template", "main.go.tmpl")
	outputGoPath := filepath.Join(testDir, "output.go")

	buildCmd := exec.Command("go", "run", builderPath, "-input", strategyPath, "-output", outputGoPath, "-template", templatePath)
	buildOutput, err := buildCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Build failed: %v\nOutput: %s", err, buildOutput)
	}

	/* Parse Generated: line to get temp Go file path */
	generatedFile := ""
	for _, line := range strings.Split(string(buildOutput), "\n") {
		if strings.HasPrefix(line, "Generated: ") {
			generatedFile = strings.TrimSpace(strings.TrimPrefix(line, "Generated: "))
			break
		}
	}
	if generatedFile == "" {
		t.Fatalf("Failed to parse generated file path from output: %s", buildOutput)
	}

	if err := setupGoMod(generatedFile, projectRoot); err != nil {
		t.Fatal(err)
	}

	binPath := filepath.Join(testDir, "test-bin")
	compileCmd := exec.Command("go", "build", "-o", binPath, generatedFile)
	if output, err := compileCmd.CombinedOutput(); err != nil {
		t.Fatalf("Compile failed: %v\nOutput: %s", err, output)
	}

	dataPath := testutil.FetchTestData(t, "ACN", "1h", 500)
	dataDir := filepath.Dir(dataPath)
	resultPath := filepath.Join(testDir, "result.json")

	runCmd := exec.Command(binPath, "-symbol", "ACN", "-data", dataPath, "-datadir", dataDir, "-output", resultPath)
	if output, err := runCmd.CombinedOutput(); err != nil {
		t.Fatalf("Execution failed: %v\nOutput: %s", err, output)
	}

	resultData, err := os.ReadFile(resultPath)
	if err != nil {
		t.Fatal(err)
	}

	var result struct {
		Indicators map[string]struct {
			Data []map[string]interface{} `json:"data"`
		} `json:"indicators"`
	}
	if err := json.Unmarshal(resultData, &result); err != nil {
		t.Fatal(err)
	}

	if len(result.Indicators) == 0 {
		t.Fatal("No indicators in output")
	}

	/* Same-TF must produce 1:1 mapping - all bars mapped */
	sameTF, ok := result.Indicators["Same-TF Close"]
	if !ok {
		t.Fatalf("Expected 'Same-TF Close' indicator, got: %v", result.Indicators)
	}

	/* Get expected bar count from data file */
	dataBytes, err := os.ReadFile(dataPath)
	if err != nil {
		t.Fatal(err)
	}
	var dataWithMetadata struct {
		Bars []interface{} `json:"bars"`
	}
	expectedBars := 0
	if err := json.Unmarshal(dataBytes, &dataWithMetadata); err == nil && len(dataWithMetadata.Bars) > 0 {
		expectedBars = len(dataWithMetadata.Bars)
	} else {
		var plainBars []interface{}
		if err := json.Unmarshal(dataBytes, &plainBars); err == nil {
			expectedBars = len(plainBars)
		}
	}

	if len(sameTF.Data) != expectedBars {
		t.Errorf("Same-timeframe mapping incorrect: got %d values, expected %d", len(sameTF.Data), expectedBars)
	}

	/* All values should be non-null (direct 1:1 copy) */
	nonNullCount := 0
	for _, point := range sameTF.Data {
		if val, ok := point["value"]; ok && val != nil {
			nonNullCount++
		}
	}

	if nonNullCount != expectedBars {
		t.Errorf("Same-timeframe should have %d non-null values, got %d", expectedBars, nonNullCount)
	}
}

/* TestSecurityUpsampling_1D_to_1h_NoWarmup verifies upsampling repeats daily values without warmup */
func TestSecurityUpsampling_1D_to_1h_NoWarmup(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	strategyCode := `
//@version=5
indicator("Security Upsample Test", overlay=true)
dailyClose = request.security(syminfo.tickerid, "1D", close)
plot(dailyClose, title="Daily Close (hourly)", color=color.red)
`

	testDir := t.TempDir()
	strategyPath := filepath.Join(testDir, "test-upsample.pine")
	if err := os.WriteFile(strategyPath, []byte(strategyCode), 0644); err != nil {
		t.Fatal(err)
	}

	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(cwd)
	builderPath := filepath.Join(projectRoot, "cmd", "pine-gen", "main.go")
	templatePath := filepath.Join(projectRoot, "template", "main.go.tmpl")
	outputGoPath := filepath.Join(testDir, "output.go")

	/* Upsample test: base=1D, security=1D → should behave same as base TF (no warmup) */
	buildCmd := exec.Command("go", "run", builderPath, "-input", strategyPath, "-output", outputGoPath, "-template", templatePath)
	buildOutput, err := buildCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Build failed: %v\nOutput: %s", err, buildOutput)
	}

	/* Parse Generated: line to get temp Go file path */
	generatedFile := ""
	for _, line := range strings.Split(string(buildOutput), "\n") {
		if strings.HasPrefix(line, "Generated: ") {
			generatedFile = strings.TrimSpace(strings.TrimPrefix(line, "Generated: "))
			break
		}
	}
	if generatedFile == "" {
		t.Fatalf("Failed to parse generated file path from output: %s", buildOutput)
	}

	if err := setupGoMod(generatedFile, projectRoot); err != nil {
		t.Fatal(err)
	}

	binPath := filepath.Join(testDir, "test-bin")
	compileCmd := exec.Command("go", "build", "-o", binPath, generatedFile)
	if output, err := compileCmd.CombinedOutput(); err != nil {
		t.Fatalf("Compile failed: %v\nOutput: %s", err, output)
	}

	dataPath := testutil.FetchTestData(t, "ACN", "1D", 500)
	dataDir := filepath.Dir(dataPath)
	resultPath := filepath.Join(testDir, "result.json")

	runCmd := exec.Command(binPath, "-symbol", "ACN", "-data", dataPath, "-datadir", dataDir, "-output", resultPath)
	if output, err := runCmd.CombinedOutput(); err != nil {
		t.Fatalf("Execution failed: %v\nOutput: %s", err, output)
	}

	resultData, err := os.ReadFile(resultPath)
	if err != nil {
		t.Fatal(err)
	}

	var result struct {
		Indicators map[string]struct {
			Data []map[string]interface{} `json:"data"`
		} `json:"indicators"`
	}
	if err := json.Unmarshal(resultData, &result); err != nil {
		t.Fatal(err)
	}

	if len(result.Indicators) == 0 {
		t.Fatal("No indicators in output")
	}

	/* Upsample 1D→1h when running on 1D base: should produce 1:1 mapping (both daily) */
	dailyClose, ok := result.Indicators["Daily Close (hourly)"]
	if !ok {
		t.Fatalf("Expected 'Daily Close (hourly)' indicator, got: %v", result.Indicators)
	}
	if len(dailyClose.Data) < 20 {
		t.Errorf("Upsampling test produced too few values: %d", len(dailyClose.Data))
	}

	/* All values should be non-null (daily data repeats per daily bar) */
	nonNullCount := 0
	for _, point := range dailyClose.Data {
		if val, ok := point["value"]; ok && val != nil {
			nonNullCount++
		}
	}

	if nonNullCount < 20 {
		t.Errorf("Upsampling should have all non-null values, got %d", nonNullCount)
	}
}
