package integration

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

/* generateDeterministicCrossoverData creates synthetic OHLC bars with guaranteed crossover patterns */
func generateDeterministicCrossoverData(filepath string) error {
	// Generate deterministic bars that create crossover signals
	// Pattern: close starts below open, crosses above twice
	baseTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	bars := []map[string]interface{}{
		// Bar 0-4: close < open (no crossover)
		{"time": baseTime.Unix(), "open": 100.0, "high": 102.0, "low": 98.0, "close": 99.0, "volume": 1000.0},
		{"time": baseTime.Add(1 * time.Hour).Unix(), "open": 100.0, "high": 101.0, "low": 97.0, "close": 98.0, "volume": 1000.0},
		{"time": baseTime.Add(2 * time.Hour).Unix(), "open": 100.0, "high": 103.0, "low": 96.0, "close": 97.0, "volume": 1000.0},
		{"time": baseTime.Add(3 * time.Hour).Unix(), "open": 100.0, "high": 102.0, "low": 95.0, "close": 96.0, "volume": 1000.0},
		{"time": baseTime.Add(4 * time.Hour).Unix(), "open": 100.0, "high": 101.0, "low": 94.0, "close": 95.0, "volume": 1000.0},

		// Bar 5: CROSSOVER #1 - close crosses above open (95 → 101)
		{"time": baseTime.Add(5 * time.Hour).Unix(), "open": 100.0, "high": 105.0, "low": 99.0, "close": 101.0, "volume": 1500.0},

		// Bar 6-9: close remains above open
		{"time": baseTime.Add(6 * time.Hour).Unix(), "open": 100.0, "high": 106.0, "low": 100.0, "close": 102.0, "volume": 1200.0},
		{"time": baseTime.Add(7 * time.Hour).Unix(), "open": 100.0, "high": 107.0, "low": 101.0, "close": 103.0, "volume": 1100.0},
		{"time": baseTime.Add(8 * time.Hour).Unix(), "open": 100.0, "high": 108.0, "low": 102.0, "close": 104.0, "volume": 1300.0},
		{"time": baseTime.Add(9 * time.Hour).Unix(), "open": 100.0, "high": 109.0, "low": 103.0, "close": 105.0, "volume": 1400.0},

		// Bar 10-14: close drops below open again
		{"time": baseTime.Add(10 * time.Hour).Unix(), "open": 100.0, "high": 102.0, "low": 97.0, "close": 98.0, "volume": 1000.0},
		{"time": baseTime.Add(11 * time.Hour).Unix(), "open": 100.0, "high": 101.0, "low": 96.0, "close": 97.0, "volume": 1000.0},
		{"time": baseTime.Add(12 * time.Hour).Unix(), "open": 100.0, "high": 100.0, "low": 95.0, "close": 96.0, "volume": 1000.0},
		{"time": baseTime.Add(13 * time.Hour).Unix(), "open": 100.0, "high": 99.0, "low": 94.0, "close": 95.0, "volume": 1000.0},
		{"time": baseTime.Add(14 * time.Hour).Unix(), "open": 100.0, "high": 98.0, "low": 93.0, "close": 94.0, "volume": 1000.0},

		// Bar 15: CROSSOVER #2 - close crosses above open again (94 → 106)
		{"time": baseTime.Add(15 * time.Hour).Unix(), "open": 100.0, "high": 110.0, "low": 99.0, "close": 106.0, "volume": 1600.0},

		// Bar 16-19: close remains above open
		{"time": baseTime.Add(16 * time.Hour).Unix(), "open": 100.0, "high": 111.0, "low": 105.0, "close": 107.0, "volume": 1200.0},
		{"time": baseTime.Add(17 * time.Hour).Unix(), "open": 100.0, "high": 112.0, "low": 106.0, "close": 108.0, "volume": 1100.0},
		{"time": baseTime.Add(18 * time.Hour).Unix(), "open": 100.0, "high": 113.0, "low": 107.0, "close": 109.0, "volume": 1300.0},
		{"time": baseTime.Add(19 * time.Hour).Unix(), "open": 100.0, "high": 114.0, "low": 108.0, "close": 110.0, "volume": 1400.0},
	}

	data, err := json.MarshalIndent(bars, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath, data, 0644)
}

func TestCrossoverExecution(t *testing.T) {
	// Change to golang-port directory for correct template path
	originalDir, _ := os.Getwd()
	os.Chdir("../..")
	defer os.Chdir(originalDir)

	tmpDir := t.TempDir()
	tempBinary := filepath.Join(tmpDir, "test-crossover-exec")
	outputFile := filepath.Join(tmpDir, "crossover-exec-result.json")
	testDataFile := filepath.Join(tmpDir, "crossover-test-data.json")
	tempGoFile := filepath.Join(os.TempDir(), "pine_strategy_temp.go")

	// Generate deterministic test data
	if err := generateDeterministicCrossoverData(testDataFile); err != nil {
		t.Fatalf("Failed to generate test data: %v", err)
	}

	// Build strategy binary
	buildCmd := exec.Command("go", "run", "cmd/pine-gen/main.go",
		"-input", "testdata/fixtures/crossover-builtin-test.pine",
		"-output", tempBinary)

	buildOutput, err := buildCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Build failed: %v\nOutput: %s", err, buildOutput)
	}

	compileCmd := exec.Command("go", "build",
		"-o", tempBinary,
		tempGoFile)

	compileOutput, err := compileCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Compile failed: %v\nOutput: %s", err, compileOutput)
	}

	// Execute strategy with generated test data
	execCmd := exec.Command(tempBinary,
		"-symbol", "TEST",
		"-data", testDataFile,
		"-output", outputFile)

	execOutput, err := execCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Execution failed: %v\nOutput: %s", err, execOutput)
	}

	// Verify output
	data, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output: %v", err)
	}

	var result struct {
		Strategy struct {
			Trades     []interface{} `json:"trades"`
			OpenTrades []struct {
				EntryID    string  `json:"entryId"`
				EntryPrice float64 `json:"entryPrice"`
				EntryBar   int     `json:"entryBar"`
				Direction  string  `json:"direction"`
			} `json:"openTrades"`
			Equity    float64 `json:"equity"`
			NetProfit float64 `json:"netProfit"`
		} `json:"strategy"`
	}

	err = json.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("Failed to parse result: %v", err)
	}

	// Verify exactly 2 crossover trades occurred (deterministic test data has 2 crossovers)
	if len(result.Strategy.OpenTrades) != 2 {
		t.Fatalf("Expected exactly 2 crossover trades (bars 5 and 15), got %d", len(result.Strategy.OpenTrades))
	}

	t.Logf("✓ Crossover trades detected: %d", len(result.Strategy.OpenTrades))

	/* Verify all trades have valid data */
	// Crossovers occur at bars 5 and 15, but entries execute on NEXT bar (6 and 16)
	expectedBars := []int{6, 16}
	for i, trade := range result.Strategy.OpenTrades {
		if trade.EntryBar != expectedBars[i] {
			t.Errorf("Trade %d: expected entry bar %d, got %d", i, expectedBars[i], trade.EntryBar)
		}
		if trade.EntryPrice <= 0 {
			t.Errorf("Trade %d: invalid entry price %.2f", i, trade.EntryPrice)
		}
		if trade.Direction != "long" {
			t.Errorf("Trade %d: expected direction 'long', got %q", i, trade.Direction)
		}
		t.Logf("  Trade %d: bar=%d, price=%.2f, direction=%s", i, trade.EntryBar, trade.EntryPrice, trade.Direction)
	}

	t.Logf("✓ Crossover execution test passed with deterministic data")
}
