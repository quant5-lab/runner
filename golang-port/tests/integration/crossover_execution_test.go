package integration

import (
	"encoding/json"
	"os"
	"os/exec"
	"testing"
)

func TestCrossoverExecution(t *testing.T) {
	// Change to golang-port directory for correct template path
	originalDir, _ := os.Getwd()
	os.Chdir("../..")
	defer os.Chdir(originalDir)

	// Build strategy binary
	buildCmd := exec.Command("go", "run", "cmd/pinescript-builder/main.go",
		"-input", "testdata/crossover-builtin-test.pine",
		"-output", "/tmp/test-crossover-exec")

	buildOutput, err := buildCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Build failed: %v\nOutput: %s", err, buildOutput)
	}

	// Compile the generated code
	compileCmd := exec.Command("go", "build",
		"-o", "/tmp/test-crossover-exec",
		"/var/folders/ft/nyw_rm792qb2056vjlkzfj200000gn/T/pine_strategy_temp.go")

	compileOutput, err := compileCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Compile failed: %v\nOutput: %s", err, compileOutput)
	}
	defer os.Remove("/tmp/test-crossover-exec")

	// Execute strategy
	outputFile := "/tmp/crossover-exec-result.json"
	defer os.Remove(outputFile)

	execCmd := exec.Command("/tmp/test-crossover-exec",
		"-symbol", "TEST",
		"-data", "testdata/crossover-bars.json",
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

	// Verify crossover trades occurred
	if len(result.Strategy.OpenTrades) == 0 {
		t.Fatal("Expected crossover trades but got none")
	}

	t.Logf("Crossover trades: %d", len(result.Strategy.OpenTrades))

	// Verify first trade is at crossover point (bar 2: close 104 > open 104, prev close 98 <= prev open 100)
	firstTrade := result.Strategy.OpenTrades[0]
	if firstTrade.EntryBar != 2 {
		t.Errorf("Expected first crossover at bar 2, got bar %d", firstTrade.EntryBar)
	}
	if firstTrade.Direction != "long" {
		t.Errorf("Expected long direction, got %s", firstTrade.Direction)
	}
	if firstTrade.EntryPrice != 104 {
		t.Errorf("Expected entry at 104, got %.2f", firstTrade.EntryPrice)
	}

	// Verify multiple crossover points detected (data has crossovers at bars 2, 4, 6, 8)
	if len(result.Strategy.OpenTrades) < 3 {
		t.Errorf("Expected at least 3 crossover trades, got %d", len(result.Strategy.OpenTrades))
	}

	t.Logf("✓ Crossover execution test passed: %d trades detected", len(result.Strategy.OpenTrades))
}
