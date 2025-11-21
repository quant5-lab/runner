package integration

import (
	"encoding/json"
	"os"
	"os/exec"
	"testing"
)

func TestSeriesStrategyExecution(t *testing.T) {
	// Change to golang-port directory for correct template path
	originalDir, _ := os.Getwd()
	os.Chdir("../..")
	defer os.Chdir(originalDir)

	// Build strategy binary from Series test strategy
	buildCmd := exec.Command("go", "run", "cmd/pine-gen/main.go",
		"-input", "testdata/strategy-sma-crossover-series.pine",
		"-output", "/tmp/test-series-strategy")

	buildOutput, err := buildCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Build failed: %v\nOutput: %s", err, buildOutput)
	}

	// Compile the generated code
	compileCmd := exec.Command("go", "build",
		"-o", "/tmp/test-series-strategy",
		"/var/folders/ft/nyw_rm792qb2056vjlkzfj200000gn/T/pine_strategy_temp.go")

	compileOutput, err := compileCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Compile failed: %v\nOutput: %s", err, compileOutput)
	}
	defer os.Remove("/tmp/test-series-strategy")

	// Create test data with clear SMA crossover pattern
	testData := createSMACrossoverTestData()
	dataFile := "/tmp/series-test-data.json"
	data, _ := json.Marshal(testData)
	os.WriteFile(dataFile, data, 0644)
	defer os.Remove(dataFile)

	// Execute strategy
	outputFile := "/tmp/series-strategy-result.json"
	defer os.Remove(outputFile)

	execCmd := exec.Command("/tmp/test-series-strategy",
		"-symbol", "TEST",
		"-data", dataFile,
		"-output", outputFile)

	execOutput, err := execCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Execution failed: %v\nOutput: %s", err, execOutput)
	}

	// Verify output
	resultData, err := os.ReadFile(outputFile)
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
		Indicators map[string]struct {
			Title string `json:"title"`
			Data  []struct {
				Time  int64   `json:"time"`
				Value float64 `json:"value"`
			} `json:"data"`
		} `json:"indicators"`
	}

	err = json.Unmarshal(resultData, &result)
	if err != nil {
		t.Fatalf("Failed to parse result: %v", err)
	}

	t.Logf("Strategy execution completed")
	t.Logf("Open trades: %d", len(result.Strategy.OpenTrades))
	t.Logf("Closed trades: %d", len(result.Strategy.Trades))

	// Verify trades were executed at crossover points
	if len(result.Strategy.OpenTrades) == 0 {
		t.Error("Expected trades at crossover points but got none")
	}

	// Verify that we have long trades (crossover signals)
	longTrades := 0
	shortTrades := 0
	for _, trade := range result.Strategy.OpenTrades {
		if trade.Direction == "long" {
			longTrades++
		} else if trade.Direction == "short" {
			shortTrades++
		}
	}
	t.Logf("Long trades: %d, Short trades: %d", longTrades, shortTrades)

	if longTrades == 0 {
		t.Error("Expected at least one long trade from crossover")
	}

	t.Log("Series strategy execution test passed")
}

func createSMACrossoverTestData() []map[string]interface{} {
	// Create data with clear SMA20 crossing above SMA50
	// Need at least 50 bars for SMA50 warmup, plus crossover pattern
	bars := []map[string]interface{}{}

	baseTime := int64(1700000000) // Unix timestamp

	// First 50 bars: downtrend (close below previous, SMA20 < SMA50)
	for i := 0; i < 50; i++ {
		close := 100.0 - float64(i)*0.5 // Decreasing from 100 to 75
		bars = append(bars, map[string]interface{}{
			"time":   baseTime + int64(i)*3600,
			"open":   close + 1,
			"high":   close + 2,
			"low":    close - 1,
			"close":  close,
			"volume": 1000.0,
		})
	}

	// Next 30 bars: uptrend (close above previous, SMA20 crosses above SMA50)
	for i := 0; i < 30; i++ {
		close := 75.0 + float64(i)*1.0 // Increasing from 75 to 105
		bars = append(bars, map[string]interface{}{
			"time":   baseTime + int64(50+i)*3600,
			"open":   close - 1,
			"high":   close + 2,
			"low":    close - 2,
			"close":  close,
			"volume": 1000.0,
		})
	}

	return bars
}
