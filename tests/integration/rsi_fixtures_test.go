package integration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/quant5-lab/runner/tests/util"
)

// TestRSIFixtures validates all RSI test fixtures compile and execute
func TestRSIFixtures(t *testing.T) {
	fixturesDir := "../fixtures/integration"

	entries, err := os.ReadDir(fixturesDir)
	if err != nil {
		t.Fatalf("fixtures directory not found: %v", err)
	}

	exec := util.NewPineExecutor(t)
	successCount := 0
	failCount := 0

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".pine" {
			continue
		}

		// Only test RSI fixtures
		name := entry.Name()
		if len(name) < 9 || name[:9] != "test-rsi-" {
			continue
		}

		t.Run(name, func(t *testing.T) {
			filePath := filepath.Join(fixturesDir, name)
			content, err := os.ReadFile(filePath)
			if err != nil {
				t.Fatalf("Could not read fixture %s: %v", name, err)
				return
			}

			output := exec.ExecuteScript(t, name[:len(name)-5], string(content))

			if output == nil {
				t.Errorf("Execution produced no output for %s", name)
				failCount++
				return
			}

			successCount++
			t.Logf("✅ %s compiled and executed", name)
		})
	}

	t.Logf("RSI Fixtures: %d passed, %d failed", successCount, failCount)
}

// TestRSIBasicFixture validates basic RSI period variations
func TestRSIBasicFixture(t *testing.T) {
	content, err := os.ReadFile("../fixtures/integration/test-rsi-basic.pine")
	if err != nil {
		t.Fatalf("test-rsi-basic.pine not found: %v", err)
	}

	exec := util.NewPineExecutor(t)
	output := exec.ExecuteScript(t, "test-rsi-basic", string(content))

	rsi14 := exec.ExtractPlotValues(t, output, "RSI 14")
	_ = exec.ExtractPlotValues(t, output, "RSI 9")
	_ = exec.ExtractPlotValues(t, output, "RSI 50")
	rsi1 := exec.ExtractPlotValues(t, output, "RSI 1")

	if len(rsi14) < 20 {
		t.Fatal("Expected at least 20 bars for RSI 14")
	}

	for i, val := range rsi14 {
		if val != val {
			continue
		}
		if val < 0 || val > 100 {
			t.Errorf("RSI 14[%d] = %f, should be 0-100", i, val)
		}
	}

	if len(rsi1) == 0 {
		t.Error("RSI 1 should have values from bar 0")
	}

	t.Logf("✅ RSI basic validated: %d bars", len(rsi14))
}

// TestRSISourcesFixture validates RSI on different data sources
func TestRSISourcesFixture(t *testing.T) {
	content, err := os.ReadFile("../fixtures/integration/test-rsi-sources.pine")
	if err != nil {
		t.Fatalf("test-rsi-sources.pine not found: %v", err)
	}

	exec := util.NewPineExecutor(t)
	output := exec.ExecuteScript(t, "test-rsi-sources", string(content))

	rsiClose := exec.ExtractPlotValues(t, output, "RSI Close")
	rsiOpen := exec.ExtractPlotValues(t, output, "RSI Open")
	rsiHigh := exec.ExtractPlotValues(t, output, "RSI High")
	rsiLow := exec.ExtractPlotValues(t, output, "RSI Low")
	rsiVolume := exec.ExtractPlotValues(t, output, "RSI Volume")

	if len(rsiClose) < 20 {
		t.Fatal("Expected at least 20 bars for RSI sources")
	}

	if len(rsiOpen) == 0 || len(rsiHigh) == 0 || len(rsiLow) == 0 || len(rsiVolume) == 0 {
		t.Error("All RSI sources should produce values")
	}

	t.Logf("✅ RSI sources validated: %d bars", len(rsiClose))
}

// TestRSIMultipleFixture validates multiple RSI instances compile correctly
func TestRSIMultipleFixture(t *testing.T) {
	content, err := os.ReadFile("../fixtures/integration/test-rsi-multiple.pine")
	if err != nil {
		t.Fatalf("test-rsi-multiple.pine not found: %v", err)
	}

	exec := util.NewPineExecutor(t)
	output := exec.ExecuteScript(t, "test-rsi-multiple", string(content))

	rsiFast := exec.ExtractPlotValues(t, output, "RSI 5")
	rsiStandard := exec.ExtractPlotValues(t, output, "RSI 14")
	rsiSlow := exec.ExtractPlotValues(t, output, "RSI 28")
	rsiVerySlow := exec.ExtractPlotValues(t, output, "RSI 50")

	if len(rsiFast) < 20 {
		t.Fatal("Expected at least 20 bars for multiple RSI test")
	}

	if len(rsiStandard) == 0 || len(rsiSlow) == 0 || len(rsiVerySlow) == 0 {
		t.Error("All RSI instances should produce values")
	}

	t.Logf("✅ Multiple RSI instances validated: %d bars", len(rsiFast))
}

// TestRSIStrategyFixture validates RSI in strategy context
func TestRSIStrategyFixture(t *testing.T) {
	content, err := os.ReadFile("../fixtures/integration/test-rsi-strategy.pine")
	if err != nil {
		t.Fatalf("test-rsi-strategy.pine not found: %v", err)
	}

	exec := util.NewPineExecutor(t)
	output := exec.ExecuteScript(t, "test-rsi-strategy", string(content))

	rsiValue := exec.ExtractPlotValues(t, output, "RSI")

	if len(rsiValue) < 20 {
		t.Fatal("Expected at least 20 bars for RSI strategy")
	}

	for i, val := range rsiValue {
		if val < 0 || val > 100 {
			t.Errorf("RSI[%d] = %f, should be 0-100", i, val)
		}
	}

	t.Logf("✅ RSI strategy validated: %d bars", len(rsiValue))
}

// TestRSIWarmupFixture validates RSI warmup behavior
func TestRSIWarmupFixture(t *testing.T) {
	content, err := os.ReadFile("../fixtures/integration/test-rsi-warmup.pine")
	if err != nil {
		t.Fatalf("test-rsi-warmup.pine not found: %v", err)
	}

	exec := util.NewPineExecutor(t)
	output := exec.ExecuteScript(t, "test-rsi-warmup", string(content))

	rsi14 := exec.ExtractPlotValues(t, output, "RSI 14")
	_ = exec.ExtractPlotValues(t, output, "RSI 5")
	rsi1 := exec.ExtractPlotValues(t, output, "RSI 1")
	_ = exec.ExtractPlotValues(t, output, "RSI 50")

	if len(rsi1) == 0 {
		t.Error("RSI 1 should have no warmup period")
	}

	if len(rsi14) < 10 {
		t.Error("RSI 14 should have values after warmup period")
	}

	t.Logf("✅ RSI warmup validated: %d bars", len(rsi14))
}

// TestRSIExtremeFixture validates RSI extreme behavior
func TestRSIExtremeFixture(t *testing.T) {
	content, err := os.ReadFile("../fixtures/integration/test-rsi-extreme.pine")
	if err != nil {
		t.Fatalf("test-rsi-extreme.pine not found: %v", err)
	}

	exec := util.NewPineExecutor(t)
	output := exec.ExecuteScript(t, "test-rsi-extreme", string(content))

	rsiMin := exec.ExtractPlotValues(t, output, "RSI 2")
	rsiStandard := exec.ExtractPlotValues(t, output, "RSI 14")
	rsiLarge := exec.ExtractPlotValues(t, output, "RSI 200")

	if len(rsiStandard) < 20 {
		t.Fatal("Expected at least 20 bars for RSI extreme test")
	}

	if len(rsiMin) == 0 || len(rsiLarge) == 0 {
		t.Error("All RSI extreme cases should produce values")
	}

	t.Logf("✅ RSI extreme periods validated: %d bars", len(rsiStandard))
}

// TestRSIComplexFixture validates RSI with complex expressions
func TestRSIComplexFixture(t *testing.T) {
	content, err := os.ReadFile("../fixtures/integration/test-rsi-complex.pine")
	if err != nil {
		t.Fatalf("test-rsi-complex.pine not found: %v", err)
	}

	exec := util.NewPineExecutor(t)
	output := exec.ExecuteScript(t, "test-rsi-complex", string(content))

	rsiOfSMA := exec.ExtractPlotValues(t, output, "RSI of SMA")
	rsiOfEMA := exec.ExtractPlotValues(t, output, "RSI of EMA")
	rsiConstPeriod := exec.ExtractPlotValues(t, output, "RSI Const Period")
	avgRSI := exec.ExtractPlotValues(t, output, "Avg RSI")

	if len(rsiOfSMA) < 20 {
		t.Fatal("Expected at least 20 bars for RSI complex test")
	}

	if len(rsiOfEMA) == 0 || len(rsiConstPeriod) == 0 || len(avgRSI) == 0 {
		t.Error("All complex RSI expressions should produce values")
	}

	t.Logf("✅ RSI complex expressions validated: %d bars", len(rsiOfSMA))
}

// TestRSIPeriodsFixture validates RSI with different period types
func TestRSIPeriodsFixture(t *testing.T) {
	content, err := os.ReadFile("../fixtures/integration/test-rsi-periods.pine")
	if err != nil {
		t.Fatalf("test-rsi-periods.pine not found: %v", err)
	}

	exec := util.NewPineExecutor(t)
	output := exec.ExecuteScript(t, "test-rsi-periods", string(content))

	rsiLiteral := exec.ExtractPlotValues(t, output, "RSI Literal 14")
	rsiConst := exec.ExtractPlotValues(t, output, "RSI Const 20")
	rsiExpression := exec.ExtractPlotValues(t, output, "RSI Expression 14")

	if len(rsiLiteral) < 20 {
		t.Fatal("Expected at least 20 bars for RSI periods test")
	}

	if len(rsiConst) == 0 || len(rsiExpression) == 0 {
		t.Error("All RSI period types should produce values")
	}

	t.Logf("✅ RSI period types validated: %d bars", len(rsiLiteral))
}
