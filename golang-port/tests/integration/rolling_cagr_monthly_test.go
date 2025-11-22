package integration

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestRollingCAGR_MonthlyTimeframe(t *testing.T) {
	// Test that rolling-cagr.pine works with monthly data
	// Verifies timeframe.ismonthly detection produces non-zero CAGR values

	// Test runs from golang-port/tests/integration
	strategy := "../../../strategies/rolling-cagr.pine"
	dataFile := "../../testdata/ohlcv/SPY_1M.json"

	// Check if files exist
	if _, err := os.Stat(strategy); os.IsNotExist(err) {
		t.Skip("rolling-cagr.pine not found, skipping test")
	}
	if _, err := os.Stat(dataFile); os.IsNotExist(err) {
		t.Skip("SPY_1M.json not found, skipping test")
	}

	// Read data to check bar count
	data, err := os.ReadFile(dataFile)
	if err != nil {
		t.Fatalf("Failed to read data file: %v", err)
	}

	var bars []map[string]interface{}
	if err := json.Unmarshal(data, &bars); err != nil {
		t.Fatalf("Failed to parse data: %v", err)
	}

	barCount := len(bars)
	t.Logf("Testing with %d monthly bars", barCount)

	// Generate strategy code (must run from golang-port to find templates)
	tempBinary := filepath.Join(t.TempDir(), "rolling-cagr-test")
	absStrategy, _ := filepath.Abs(strategy)

	genCmd := exec.Command("go", "run", "./cmd/pine-gen",
		"-input", absStrategy,
		"-output", tempBinary)
	genCmd.Dir = "../../" // Run from golang-port directory
	genOutput, err := genCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to generate strategy: %v\nOutput: %s", err, genOutput)
	}

	t.Log(string(genOutput))

	// pine-gen always generates to $TMPDIR/pine_strategy_temp.go
	tempSource := filepath.Join(os.TempDir(), "pine_strategy_temp.go")

	// Compile generated code
	absDataFile, _ := filepath.Abs(dataFile)
	buildCmd := exec.Command("go", "build", "-o", tempBinary, tempSource)
	buildCmd.Dir = "../../" // Build from golang-port to access runtime packages
	buildOutput, err := buildCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build strategy: %v\nOutput: %s", err, buildOutput)
	}

	// Run strategy
	outputFile := filepath.Join(t.TempDir(), "output.json")
	runCmd := exec.Command(tempBinary,
		"-symbol", "SPY",
		"-timeframe", "M",
		"-data", absDataFile,
		"-output", outputFile)
	runOutput, err := runCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to run strategy: %v\nOutput: %s", err, runOutput)
	}

	t.Log(string(runOutput))

	// Verify output
	resultData, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output: %v", err)
	}

	var result struct {
		Indicators map[string]struct {
			Title string `json:"title"`
			Data  []struct {
				Time  int64    `json:"time"`
				Value *float64 `json:"value"`
			} `json:"data"`
		} `json:"indicators"`
	}

	if err := json.Unmarshal(resultData, &result); err != nil {
		t.Fatalf("Failed to parse result: %v", err)
	}

	// Check CAGR A indicator exists
	cagrIndicator, exists := result.Indicators["CAGR A"]
	if !exists {
		t.Fatal("CAGR A indicator not found in output")
	}

	if len(cagrIndicator.Data) == 0 {
		t.Fatal("CAGR A has no data points")
	}

	// Count valid (non-null, non-zero) values
	validCount := 0
	nullCount := 0
	zeroCount := 0

	for _, point := range cagrIndicator.Data {
		if point.Value == nil {
			nullCount++
		} else if *point.Value == 0 {
			zeroCount++
		} else {
			validCount++
		}
	}

	t.Logf("CAGR values: %d total, %d valid, %d null, %d zero",
		len(cagrIndicator.Data), validCount, nullCount, zeroCount)

	// For 5-year CAGR on monthly data:
	// - Need 60 months (5 years * 12 months)
	// - SPY has 121 bars
	// - Expected: 121 - 60 = 61 valid values
	expectedValid := barCount - 60

	if validCount == 0 {
		t.Fatal("All CAGR values are zero or null - timeframe.ismonthly likely not working")
	}

	if validCount < expectedValid-10 {
		t.Errorf("Expected ~%d valid values, got %d (tolerance: -10)", expectedValid, validCount)
	}

	// Check that some values are within reasonable CAGR range (e.g., -50% to +100%)
	reasonableCount := 0
	for _, point := range cagrIndicator.Data {
		if point.Value != nil && *point.Value != 0 {
			val := *point.Value
			if val >= -50 && val <= 100 {
				reasonableCount++
			}
		}
	}

	if reasonableCount == 0 {
		t.Error("No reasonable CAGR values found (expected range: -50% to +100%)")
	}

	t.Logf("✓ Rolling CAGR monthly test passed: %d/%d values in reasonable range",
		reasonableCount, validCount)
}
