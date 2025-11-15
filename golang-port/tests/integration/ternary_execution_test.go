package integration

import (
	"encoding/json"
	"os"
	"os/exec"
	"testing"
)

func TestTernaryExecution(t *testing.T) {
	// Change to golang-port directory for correct template path
	originalDir, _ := os.Getwd()
	os.Chdir("../..")
	defer os.Chdir(originalDir)

	// Build strategy binary
	buildCmd := exec.Command("go", "run", "cmd/pinescript-builder/main.go",
		"-input", "testdata/ternary-test.pine",
		"-output", "/tmp/test-ternary-exec")

	buildOutput, err := buildCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Build failed: %v\nOutput: %s", err, buildOutput)
	}

	// Compile the generated code
	compileCmd := exec.Command("go", "build",
		"-o", "/tmp/test-ternary-exec",
		"/var/folders/ft/nyw_rm792qb2056vjlkzfj200000gn/T/pine_strategy_temp.go")

	compileOutput, err := compileCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Compile failed: %v\nOutput: %s", err, compileOutput)
	}
	defer os.Remove("/tmp/test-ternary-exec")

	// Create test data - alternating close above/below SMA
	testData := []map[string]interface{}{
		{"time": 1700000000, "open": 100.0, "high": 105.0, "low": 95.0, "close": 110.0, "volume": 1000.0},
		{"time": 1700003600, "open": 110.0, "high": 115.0, "low": 105.0, "close": 112.0, "volume": 1100.0},
		{"time": 1700007200, "open": 112.0, "high": 117.0, "low": 107.0, "close": 114.0, "volume": 1200.0},
		{"time": 1700010800, "open": 114.0, "high": 119.0, "low": 109.0, "close": 116.0, "volume": 1300.0},
		{"time": 1700014400, "open": 116.0, "high": 121.0, "low": 111.0, "close": 118.0, "volume": 1400.0},
		{"time": 1700018000, "open": 118.0, "high": 123.0, "low": 113.0, "close": 120.0, "volume": 1500.0},
		{"time": 1700021600, "open": 120.0, "high": 125.0, "low": 115.0, "close": 122.0, "volume": 1600.0},
		{"time": 1700025200, "open": 122.0, "high": 127.0, "low": 117.0, "close": 124.0, "volume": 1700.0},
		{"time": 1700028800, "open": 124.0, "high": 129.0, "low": 119.0, "close": 126.0, "volume": 1800.0},
		{"time": 1700032400, "open": 126.0, "high": 131.0, "low": 121.0, "close": 128.0, "volume": 1900.0},
		{"time": 1700036000, "open": 128.0, "high": 133.0, "low": 123.0, "close": 130.0, "volume": 2000.0},
		{"time": 1700039600, "open": 130.0, "high": 135.0, "low": 125.0, "close": 132.0, "volume": 2100.0},
		{"time": 1700043200, "open": 132.0, "high": 137.0, "low": 127.0, "close": 134.0, "volume": 2200.0},
		{"time": 1700046800, "open": 134.0, "high": 139.0, "low": 129.0, "close": 136.0, "volume": 2300.0},
		{"time": 1700050400, "open": 136.0, "high": 141.0, "low": 131.0, "close": 138.0, "volume": 2400.0},
		{"time": 1700054000, "open": 138.0, "high": 143.0, "low": 133.0, "close": 140.0, "volume": 2500.0},
		{"time": 1700057600, "open": 140.0, "high": 145.0, "low": 135.0, "close": 142.0, "volume": 2600.0},
		{"time": 1700061200, "open": 142.0, "high": 147.0, "low": 137.0, "close": 144.0, "volume": 2700.0},
		{"time": 1700064800, "open": 144.0, "high": 149.0, "low": 139.0, "close": 146.0, "volume": 2800.0},
		{"time": 1700068400, "open": 146.0, "high": 151.0, "low": 141.0, "close": 148.0, "volume": 2900.0},
		{"time": 1700072000, "open": 148.0, "high": 153.0, "low": 143.0, "close": 100.0, "volume": 3000.0},
		{"time": 1700075600, "open": 100.0, "high": 105.0, "low": 95.0, "close": 102.0, "volume": 3100.0},
		{"time": 1700079200, "open": 102.0, "high": 107.0, "low": 97.0, "close": 104.0, "volume": 3200.0},
		{"time": 1700082800, "open": 104.0, "high": 109.0, "low": 99.0, "close": 106.0, "volume": 3300.0},
	}

	dataFile := "/tmp/ternary-test-bars.json"
	defer os.Remove(dataFile)
	dataJSON, _ := json.Marshal(testData)
	err = os.WriteFile(dataFile, dataJSON, 0644)
	if err != nil {
		t.Fatalf("Write data failed: %v", err)
	}

	// Execute strategy
	outputFile := "/tmp/ternary-exec-result.json"
	defer os.Remove(outputFile)

	execCmd := exec.Command("/tmp/test-ternary-exec",
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
		t.Fatalf("Read output failed: %v", err)
	}

	var result map[string]interface{}
	err = json.Unmarshal(resultData, &result)
	if err != nil {
		t.Fatalf("Parse output failed: %v\nOutput: %s", err, resultData)
	}

	// Verify signal values
	plots, ok := result["plots"].(map[string]interface{})
	if !ok {
		t.Fatalf("Missing plots in output")
	}

	signalPlotObj, ok := plots["signal"].(map[string]interface{})
	if !ok {
		t.Fatalf("Missing signal plot object")
	}

	signalPlot, ok := signalPlotObj["data"].([]interface{})
	if !ok {
		t.Fatalf("Missing signal plot data")
	}

	// After first 20 bars (SMA period), check signals
	// Bars 0-19: SMA warming up
	// Bars 20-23: Close below SMA, signal should be 0
	if len(signalPlot) < 24 {
		t.Fatalf("Expected at least 24 signal values, got %d", len(signalPlot))
	}

	// Check bar 20 (first bar after warmup with close=100, below SMA of ~134)
	bar20Signal := signalPlot[20].(map[string]interface{})
	if bar20Signal["value"].(float64) != 0.0 {
		t.Errorf("Bar 20: expected signal=0 (close below SMA), got %v", bar20Signal["value"])
	}

	// Check bar 19 (last bar with close above SMA)
	bar19Signal := signalPlot[19].(map[string]interface{})
	if bar19Signal["value"].(float64) != 1.0 {
		t.Errorf("Bar 19: expected signal=1 (close above SMA), got %v", bar19Signal["value"])
	}
}
