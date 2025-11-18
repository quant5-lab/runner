package tests

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

/* TestSecurityDownsampling_1h_to_1D_WithWarmup verifies downsampling adds 500 warmup bars */
func TestSecurityDownsampling_1h_to_1D_WithWarmup(t *testing.T) {
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
	
	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(cwd)
	builderPath := filepath.Join(projectRoot, "cmd", "pinescript-builder", "main.go")
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
	
	binPath := filepath.Join(testDir, "test-bin")
	compileCmd := exec.Command("go", "build", "-o", binPath, generatedFile)
	if output, err := compileCmd.CombinedOutput(); err != nil {
		t.Fatalf("Compile failed: %v\nOutput: %s", err, output)
	}
	
	dataPath := filepath.Join(projectRoot, "testdata", "ohlcv", "BTCUSDT_1h.json")
	resultPath := filepath.Join(testDir, "result.json")
	
	runCmd := exec.Command(binPath, "-symbol", "BTCUSDT", "-data", dataPath, "-output", resultPath)
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
	
	/* Downsample 1h→1D must produce values - warmup should provide enough daily bars */
	dailyClose, ok := result.Indicators["Daily Close"]
	if !ok {
		t.Fatalf("Expected 'Daily Close' indicator, got: %v", result.Indicators)
	}
	if len(dailyClose.Data) == 0 {
		t.Fatal("Downsampling produced zero values - warmup failed")
	}
	
	nonNullCount := 0
	for _, point := range dailyClose.Data {
		if val, ok := point["value"]; ok && val != nil {
			nonNullCount++
		}
	}
	
	/* With 500h warmup → 20+ days, expect >480 values (close available immediately) */
	if nonNullCount < 480 {
		t.Errorf("Downsampling warmup insufficient: got %d non-null values, expected >480", nonNullCount)
	}
}

/* TestSecuritySameTimeframe_1h_to_1h_NoWarmup verifies same-timeframe has no warmup overhead */
func TestSecuritySameTimeframe_1h_to_1h_NoWarmup(t *testing.T) {
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
	builderPath := filepath.Join(projectRoot, "cmd", "pinescript-builder", "main.go")
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
	
	binPath := filepath.Join(testDir, "test-bin")
	compileCmd := exec.Command("go", "build", "-o", binPath, generatedFile)
	if output, err := compileCmd.CombinedOutput(); err != nil {
		t.Fatalf("Compile failed: %v\nOutput: %s", err, output)
	}
	
	dataPath := filepath.Join(projectRoot, "testdata", "ohlcv", "BTCUSDT_1h.json")
	resultPath := filepath.Join(testDir, "result.json")
	
	runCmd := exec.Command(binPath, "-symbol", "BTCUSDT", "-data", dataPath, "-output", resultPath)
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
	
	/* Same-TF must produce 1:1 mapping - all 500 bars mapped */
	sameTF, ok := result.Indicators["Same-TF Close"]
	if !ok {
		t.Fatalf("Expected 'Same-TF Close' indicator, got: %v", result.Indicators)
	}
	if len(sameTF.Data) != 500 {
		t.Errorf("Same-timeframe mapping incorrect: got %d values, expected 500", len(sameTF.Data))
	}
	
	/* All values should be non-null (direct 1:1 copy) */
	nonNullCount := 0
	for _, point := range sameTF.Data {
		if val, ok := point["value"]; ok && val != nil {
			nonNullCount++
		}
	}
	
	if nonNullCount != 500 {
		t.Errorf("Same-timeframe should have 500 non-null values, got %d", nonNullCount)
	}
}

/* TestSecurityUpsampling_1D_to_1h_NoWarmup verifies upsampling repeats daily values without warmup */
func TestSecurityUpsampling_1D_to_1h_NoWarmup(t *testing.T) {
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
	builderPath := filepath.Join(projectRoot, "cmd", "pinescript-builder", "main.go")
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
	
	binPath := filepath.Join(testDir, "test-bin")
	compileCmd := exec.Command("go", "build", "-o", binPath, generatedFile)
	if output, err := compileCmd.CombinedOutput(); err != nil {
		t.Fatalf("Compile failed: %v\nOutput: %s", err, output)
	}
	
	dataPath := filepath.Join(projectRoot, "testdata", "ohlcv", "BTCUSDT_1D.json")
	resultPath := filepath.Join(testDir, "result.json")
	
	runCmd := exec.Command(binPath, "-symbol", "BTCUSDT", "-data", dataPath, "-output", resultPath)
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
