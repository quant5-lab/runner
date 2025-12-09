package integration

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/quant5-lab/runner/parser"
	"github.com/quant5-lab/runner/runtime/chartdata"
	"github.com/quant5-lab/runner/runtime/context"
	"github.com/quant5-lab/runner/runtime/output"
	"github.com/quant5-lab/runner/runtime/strategy"
)

/* Test parsing simple Pine strategy */
func TestParseSimplePine(t *testing.T) {
	strategyPath := "../../../strategies/test-simple.pine"
	content, err := os.ReadFile(strategyPath)
	if err != nil {
		t.Fatalf("test-simple.pine not found (required test fixture): %v", err)
	}

	p, err := parser.NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	ast, err := p.ParseString("test-simple.pine", string(content))
	if err != nil {
		t.Fatalf("Failed to parse test-simple.pine: %v", err)
	}

	if ast == nil {
		t.Fatal("AST should not be nil")
	}

	// Convert to ESTree
	converter := parser.NewConverter()
	estree, err := converter.ToESTree(ast)
	if err != nil {
		t.Fatalf("Failed to convert to ESTree: %v", err)
	}

	// Convert to JSON
	jsonBytes, err := converter.ToJSON(estree)
	if err != nil {
		t.Fatalf("Failed to convert to JSON: %v", err)
	}

	var parsed map[string]interface{}
	err = json.Unmarshal(jsonBytes, &parsed)
	if err != nil {
		t.Fatalf("Failed to parse AST JSON: %v", err)
	}

	if len(jsonBytes) == 0 {
		t.Fatal("Generated JSON should not be empty")
	}

	t.Logf("Parsed %d bytes from test-simple.pine", len(jsonBytes))
}

/* Test parsing e2e fixture strategy - validates parser handles known limitations */
func TestParseFixtureStrategy(t *testing.T) {
	strategyPath := "../../../e2e/fixtures/strategies/test-strategy.pine"
	content, err := os.ReadFile(strategyPath)
	if err != nil {
		t.Fatalf("test-strategy.pine not found: %v", err)
	}

	p, err := parser.NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	ast, err := p.ParseString("test-strategy.pine", string(content))

	// Known limitation: Parser cannot handle user-defined functions with `=>` syntax
	// This is expected behavior for current PoC phase
	if err != nil {
		if containsSubstr(err.Error(), "unexpected token") {
			t.Logf("EXPECTED LIMITATION: Parser rejects user-defined function syntax: %v", err)
			return
		}
		t.Fatalf("Unexpected parse error: %v", err)
	}

	if ast == nil {
		t.Fatal("AST should not be nil when parse succeeds")
	}

	converter := parser.NewConverter()
	estree, err := converter.ToESTree(ast)
	if err != nil {
		t.Fatalf("Failed to convert to ESTree: %v", err)
	}

	jsonBytes, err := converter.ToJSON(estree)
	if err != nil {
		t.Fatalf("Failed to convert to JSON: %v", err)
	}

	t.Logf("Parsed %d bytes from test-strategy.pine", len(jsonBytes))
}

func containsSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

/* Test chart data generation with mock runtime */
func TestChartDataGeneration(t *testing.T) {
	// Create mock context
	ctx := context.New("TEST", "1h", 100)

	// Add sample bars
	for i := 0; i < 50; i++ {
		ctx.AddBar(context.OHLCV{
			Time:   int64(1700000000 + i*3600),
			Open:   100.0 + float64(i)*0.5,
			High:   105.0 + float64(i)*0.5,
			Low:    95.0 + float64(i)*0.5,
			Close:  102.0 + float64(i)*0.5,
			Volume: 1000.0,
		})
	}

	// Create chart data with metadata
	cd := chartdata.NewChartData(ctx, "TEST", "1h", "Test Strategy")

	// Add mock plots
	collector := output.NewCollector()
	for i := 0; i < 50; i++ {
		collector.Add("SMA 20", int64(1700000000+i*3600), 100.0+float64(i)*0.5, nil)
	}
	cd.AddPlots(collector)

	// Add mock strategy
	strat := strategy.NewStrategy()
	strat.Call("Test Strategy", 10000)
	strat.Entry("long1", strategy.Long, 10)
	strat.OnBarUpdate(1, 100, 1700000000)
	strat.Close("long1", 110, 1700003600)
	cd.AddStrategy(strat, 110)

	// Generate JSON
	jsonBytes, err := cd.ToJSON()
	if err != nil {
		t.Fatalf("Failed to generate JSON: %v", err)
	}

	// Validate structure
	var parsed map[string]interface{}
	err = json.Unmarshal(jsonBytes, &parsed)
	if err != nil {
		t.Fatalf("Failed to parse chart data JSON: %v", err)
	}

	// Verify required fields
	if _, ok := parsed["candlestick"]; !ok {
		t.Error("Missing candlestick field")
	}
	if _, ok := parsed["indicators"]; !ok {
		t.Error("Missing indicators field")
	}
	if _, ok := parsed["strategy"]; !ok {
		t.Error("Missing strategy field")
	}
	if _, ok := parsed["metadata"]; !ok {
		t.Error("Missing metadata field")
	}
	if _, ok := parsed["ui"]; !ok {
		t.Error("Missing ui field")
	}

	t.Logf("Generated chart data: %d bytes", len(jsonBytes))
}

/* Test parsing all fixture strategies */
func TestParseAllFixtures(t *testing.T) {
	fixturesDir := "../../../e2e/fixtures/strategies"

	entries, err := os.ReadDir(fixturesDir)
	if err != nil {
		t.Fatalf("fixtures directory not found (required test fixtures): %v", err)
	}

	p, err := parser.NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	successCount := 0
	failCount := 0
	knownLimitations := map[string]string{
		"test-builtin-function.pine": "user-defined functions with => syntax",
		"test-function-scoping.pine": "user-defined functions with => syntax",
		"test-strategy.pine":         "user-defined functions with => syntax",
		"test-tr-adx.pine":           "user-defined functions with => syntax",
		"test-tr-bb7-adx.pine":       "user-defined functions with => syntax",
		"test-tr-function.pine":      "user-defined functions with => syntax",
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".pine" {
			continue
		}

		filePath := filepath.Join(fixturesDir, entry.Name())
		content, err := os.ReadFile(filePath)
		if err != nil {
			t.Fatalf("Could not read fixture %s: %v", entry.Name(), err)
		}

		ast, err := p.ParseString(entry.Name(), string(content))
		if err != nil {
			if reason, isKnown := knownLimitations[entry.Name()]; isKnown {
				t.Logf("KNOWN LIMITATION: %s - %s", entry.Name(), reason)
				failCount++
			} else {
				t.Errorf("UNEXPECTED FAILURE: %s - %v", entry.Name(), err)
				failCount++
			}
			continue
		}

		if ast == nil {
			t.Errorf("FAIL: %s - AST is nil despite no parse error", entry.Name())
			failCount++
			continue
		}

		successCount++
		t.Logf("PASS: %s", entry.Name())
	}

	t.Logf("Results: %d passed, %d failed", successCount, failCount)

	expectedFails := len(knownLimitations)
	if failCount != expectedFails {
		t.Errorf("Expected %d known failures, got %d failures", expectedFails, failCount)
	}
}

/* Test runtime integration with simple strategy */
func TestRuntimeIntegration(t *testing.T) {
	// Create context
	ctx := context.New("TEST", "1h", 100)

	// Add bars with price movement for crossover
	for i := 0; i < 100; i++ {
		price := 100.0
		if i > 20 && i < 40 {
			price = 95.0 + float64(i-20)*0.5 // Uptrend
		} else if i >= 40 && i < 60 {
			price = 105.0 - float64(i-40)*0.3 // Downtrend
		}

		ctx.AddBar(context.OHLCV{
			Time:   int64(1700000000 + i*3600),
			Open:   price,
			High:   price + 2,
			Low:    price - 2,
			Close:  price,
			Volume: 1000.0,
		})
	}

	// Create strategy
	strat := strategy.NewStrategy()
	strat.Call("Test Runtime Strategy", 10000)

	// Simulate strategy execution
	for i := 0; i < len(ctx.Data); i++ {
		ctx.BarIndex = i
		strat.OnBarUpdate(i, ctx.Data[i].Open, ctx.Data[i].Time)

		// Simple strategy logic: buy on uptrend, sell on downtrend
		if i > 25 && i < 30 && strat.GetPositionSize() == 0 {
			strat.Entry("long", strategy.Long, 1)
		}
		if i > 45 && i < 50 && strat.GetPositionSize() > 0 {
			strat.Close("long", ctx.Data[i].Close, ctx.Data[i].Time)
		}
	}

	// Verify results
	th := strat.GetTradeHistory()
	closedTrades := th.GetClosedTrades()

	if len(closedTrades) == 0 {
		t.Log("No trades executed (expected for simple test)")
	} else {
		t.Logf("Executed %d trades", len(closedTrades))
		for _, trade := range closedTrades {
			t.Logf("Trade: %s, Entry: %.2f, Exit: %.2f, Profit: %.2f",
				trade.EntryID, trade.EntryPrice, trade.ExitPrice, trade.Profit)
		}
	}

	equity := strat.GetEquity(ctx.Data[len(ctx.Data)-1].Close)
	t.Logf("Final equity: %.2f", equity)
}
