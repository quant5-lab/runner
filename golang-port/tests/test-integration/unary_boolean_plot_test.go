package integration

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestUnaryBooleanInPlot(t *testing.T) {
	originalDir, _ := os.Getwd()
	os.Chdir("../..")
	defer os.Chdir(originalDir)

	tmpDir := t.TempDir()
	tempBinary := filepath.Join(tmpDir, "unary-bool-test")

	// Use pre-existing test fixture
	fixtureFile := "testdata/fixtures/unary-boolean-plot.pine"

	// Build strategy
	buildCmd := exec.Command("go", "run", "cmd/pine-gen/main.go",
		"-input", fixtureFile,
		"-output", tempBinary)

	buildOutput, err := buildCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Build failed: %v\nOutput: %s", err, buildOutput)
	}

	tempGoFile := ParseGeneratedFilePath(t, buildOutput)

	// Compile - this will fail if boolean type mismatches exist
	compileCmd := exec.Command("go", "build",
		"-o", tempBinary,
		tempGoFile)

	compileOutput, err := compileCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Compile failed with type errors: %v\nOutput: %s\n\nThis indicates boolean conversion issues with unary expressions", err, compileOutput)
	}

	// Create test data with values crossing thresholds
	testData := []map[string]interface{}{}
	baseTime := int64(1700000000)
	prices := []float64{95, 98, 105, 112, 108, 102, 115, 120, 98, 95, 110, 118}

	for i, price := range prices {
		testData = append(testData, map[string]interface{}{
			"time":   baseTime + int64(i*3600),
			"open":   price - 1.0,
			"high":   price + 2.0,
			"low":    price - 2.0,
			"close":  price,
			"volume": 1000.0,
		})
	}

	dataFile := filepath.Join(tmpDir, "unary-bool-bars.json")
	dataJSON, _ := json.Marshal(testData)
	err = os.WriteFile(dataFile, dataJSON, 0644)
	if err != nil {
		t.Fatalf("Write data failed: %v", err)
	}

	// Execute strategy
	outputFile := filepath.Join(tmpDir, "unary-bool-result.json")

	execCmd := exec.Command(tempBinary,
		"-symbol", "TEST",
		"-timeframe", "1h",
		"-data", dataFile,
		"-output", outputFile)

	execOutput, err := execCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Execution failed: %v\nOutput: %s", err, execOutput)
	}

	// Verify output exists and contains plots
	outputData, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output: %v", err)
	}

	var result map[string]interface{}
	err = json.Unmarshal(outputData, &result)
	if err != nil {
		t.Fatalf("Failed to parse output JSON: %v", err)
	}

	// Verify indicators map exists (Pine v5 output structure uses map, not array)
	indicators, ok := result["indicators"].(map[string]interface{})
	if !ok {
		t.Fatal("Output missing indicators map")
	}

	// Count indicators with our test titles
	expectedTitles := []string{
		"Buy Active",
		"Sell Active",
		"Has Signal",
	}

	foundTitles := make(map[string]bool)
	for title := range indicators {
		for _, expected := range expectedTitles {
			if title == expected {
				foundTitles[title] = true
			}
		}
	}

	if len(foundTitles) != len(expectedTitles) {
		t.Errorf("Expected %d unary boolean plots, found %d", len(expectedTitles), len(foundTitles))
		t.Logf("Found titles: %v", foundTitles)
	}

	// Verify no runtime errors (strategy executed to completion)
	_, ok = result["candlestick"].([]interface{})
	if !ok {
		t.Fatal("Strategy did not execute properly - no candlestick data in output")
	}

	t.Logf("✓ Unary boolean plot test passed: indicators generated %v", foundTitles)
}

func TestUnaryBooleanInConditional(t *testing.T) {
	originalDir, _ := os.Getwd()
	os.Chdir("../..")
	defer os.Chdir(originalDir)

	tmpDir := t.TempDir()
	tempBinary := filepath.Join(tmpDir, "unary-cond-test")

	// Use pre-existing test fixture
	fixtureFile := "testdata/fixtures/unary-boolean-conditional.pine"

	// Build
	buildCmd := exec.Command("go", "run", "cmd/pine-gen/main.go",
		"-input", fixtureFile,
		"-output", tempBinary)

	buildOutput, err := buildCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Build failed: %v\nOutput: %s", err, buildOutput)
	}

	tempGoFile := ParseGeneratedFilePath(t, buildOutput)

	// Compile
	compileCmd := exec.Command("go", "build",
		"-o", tempBinary,
		tempGoFile)

	compileOutput, err := compileCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Compile failed: %v\nOutput: %s", err, compileOutput)
	}

	// Create test data
	testData := []map[string]interface{}{}
	baseTime := int64(1700000000)
	prices := []float64{100, 102, 98, 105, 103, 101, 107, 110}

	for i, price := range prices {
		testData = append(testData, map[string]interface{}{
			"time":   baseTime + int64(i*3600),
			"open":   price - 1.0,
			"high":   price + 2.0,
			"low":    price - 2.0,
			"close":  price,
			"volume": 1000.0,
		})
	}

	dataFile := filepath.Join(tmpDir, "unary-cond-bars.json")
	dataJSON, _ := json.Marshal(testData)
	err = os.WriteFile(dataFile, dataJSON, 0644)
	if err != nil {
		t.Fatalf("Write data failed: %v", err)
	}

	// Execute
	outputFile := filepath.Join(tmpDir, "unary-cond-result.json")

	execCmd := exec.Command(tempBinary,
		"-symbol", "TEST",
		"-timeframe", "1h",
		"-data", dataFile,
		"-output", outputFile)

	execOutput, err := execCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Execution failed: %v\nOutput: %s", err, execOutput)
	}

	// Verify execution completed
	outputData, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output: %v", err)
	}

	var result map[string]interface{}
	err = json.Unmarshal(outputData, &result)
	if err != nil {
		t.Fatalf("Failed to parse output JSON: %v", err)
	}

	// Verify execution completed (check candlestick data exists)
	_, ok := result["candlestick"].([]interface{})
	if !ok {
		t.Fatal("Strategy did not execute - unary boolean conditionals may have caused runtime errors")
	}

	t.Logf("✓ Unary boolean conditional test passed: candlestick data generated, no runtime errors")
}
