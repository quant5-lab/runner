package chartdata

import (
	"encoding/json"
	"testing"

	"github.com/quant5-lab/runner/runtime/clock"
	"github.com/quant5-lab/runner/runtime/context"
	"github.com/quant5-lab/runner/runtime/output"
	"github.com/quant5-lab/runner/runtime/strategy"
)

func TestNewChartData(t *testing.T) {
	ctx := context.New("TEST", "1h", 10)
	now := clock.Now().Unix()

	for i := 0; i < 5; i++ {
		ctx.AddBar(context.OHLCV{
			Time:   now + int64(i*3600),
			Open:   100.0 + float64(i),
			High:   105.0 + float64(i),
			Low:    95.0 + float64(i),
			Close:  102.0 + float64(i),
			Volume: 1000.0,
		})
	}

	cd := NewChartData(ctx, "TEST", "1h", "Test Strategy")

	if len(cd.Candlestick) != 5 {
		t.Errorf("Expected 5 candlesticks, got %d", len(cd.Candlestick))
	}
	if cd.Metadata.Timestamp == "" {
		t.Error("Timestamp should not be empty")
	}
	if cd.Indicators == nil {
		t.Error("Indicators map should be initialized")
	}
	if cd.Metadata.Symbol != "TEST" {
		t.Errorf("Expected symbol TEST, got %s", cd.Metadata.Symbol)
	}
	if cd.Metadata.Title != "Test Strategy - TEST" {
		t.Errorf("Expected title 'Test Strategy - TEST', got '%s'", cd.Metadata.Title)
	}
}

func TestAddPlots(t *testing.T) {
	ctx := context.New("TEST", "1h", 10)
	cd := NewChartData(ctx, "TEST", "1h", "")

	collector := output.NewCollector()
	now := clock.Now().Unix()

	collector.Add("SMA 20", now, 100.0, nil)
	collector.Add("SMA 20", now+3600, 102.0, nil)
	collector.Add("RSI", now, 50.0, map[string]interface{}{"pane": "indicator"})

	cd.AddPlots(collector)

	if len(cd.Indicators) != 2 {
		t.Errorf("Expected 2 indicator series, got %d", len(cd.Indicators))
	}

	smaSeries, ok := cd.Indicators["SMA 20"]
	if !ok {
		t.Fatal("SMA 20 series not found")
	}
	if len(smaSeries.Data) != 2 {
		t.Errorf("Expected 2 SMA points, got %d", len(smaSeries.Data))
	}
	if smaSeries.Title != "SMA 20" {
		t.Errorf("Expected title 'SMA 20', got '%s'", smaSeries.Title)
	}
}

func TestAddStrategy(t *testing.T) {
	ctx := context.New("TEST", "1h", 10)
	cd := NewChartData(ctx, "TEST", "1h", "Test Strategy")

	strat := strategy.NewStrategy()
	strat.Call("Test Strategy", 10000)

	// Place and execute trade
	strat.Entry("long1", strategy.Long, 10, "")
	strat.OnBarUpdate(1, 100, 1000)
	strat.Close("long1", 110, 2000, "")

	cd.AddStrategy(strat, 110)

	if cd.Strategy == nil {
		t.Fatal("Strategy data should be set")
	}
	if len(cd.Strategy.Trades) != 1 {
		t.Errorf("Expected 1 closed trade, got %d", len(cd.Strategy.Trades))
	}
	if cd.Strategy.NetProfit != 100 {
		t.Errorf("Expected net profit 100, got %.2f", cd.Strategy.NetProfit)
	}
	if cd.Strategy.Equity != 10100 {
		t.Errorf("Expected equity 10100, got %.2f", cd.Strategy.Equity)
	}
}

func TestToJSON(t *testing.T) {
	ctx := context.New("TEST", "1h", 10)
	now := clock.Now().Unix()
	ctx.AddBar(context.OHLCV{
		Time: now, Open: 100, High: 105, Low: 95, Close: 102, Volume: 1000,
	})

	cd := NewChartData(ctx, "TEST", "1h", "")

	collector := output.NewCollector()
	collector.Add("SMA", now, 100.0, nil)
	cd.AddPlots(collector)

	jsonBytes, err := cd.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON() failed: %v", err)
	}

	// Validate JSON structure
	var parsed map[string]interface{}
	err = json.Unmarshal(jsonBytes, &parsed)
	if err != nil {
		t.Fatalf("JSON unmarshal failed: %v", err)
	}

	if _, ok := parsed["candlestick"]; !ok {
		t.Error("JSON should have 'candlestick' field")
	}
	if _, ok := parsed["indicators"]; !ok {
		t.Error("JSON should have 'indicators' field")
	}
	if _, ok := parsed["metadata"]; !ok {
		t.Error("JSON should have 'metadata' field")
	}
	if _, ok := parsed["ui"]; !ok {
		t.Error("JSON should have 'ui' field")
	}
}

func TestStrategyDataStructure(t *testing.T) {
	ctx := context.New("TEST", "1h", 10)
	cd := NewChartData(ctx, "TEST", "1h", "Test Strategy")

	strat := strategy.NewStrategy()
	strat.Call("Test Strategy", 10000)

	// Open trade
	strat.Entry("long1", strategy.Long, 5, "")
	strat.OnBarUpdate(1, 100, 1000)

	// Close trade
	strat.Close("long1", 110, 2000, "")

	// Another open trade
	strat.Entry("long2", strategy.Long, 3, "")
	strat.OnBarUpdate(2, 110, 3000)

	cd.AddStrategy(strat, 115)

	if cd.Strategy == nil {
		t.Fatal("Strategy should be set")
	}
	if len(cd.Strategy.Trades) != 1 {
		t.Errorf("Expected 1 closed trade, got %d", len(cd.Strategy.Trades))
	}
	if len(cd.Strategy.OpenTrades) != 1 {
		t.Errorf("Expected 1 open trade, got %d", len(cd.Strategy.OpenTrades))
	}

	// Check closed trade structure
	trade := cd.Strategy.Trades[0]
	if trade.EntryID != "long1" {
		t.Errorf("Expected EntryID 'long1', got '%s'", trade.EntryID)
	}
	if trade.Profit != 50 {
		t.Errorf("Expected profit 50, got %.2f", trade.Profit)
	}

	// Check open trade structure
	openTrade := cd.Strategy.OpenTrades[0]
	if openTrade.EntryID != "long2" {
		t.Errorf("Expected EntryID 'long2', got '%s'", openTrade.EntryID)
	}
}

/* TestTradeCommentSerialization verifies JSON serialization with comments */
func TestTradeCommentSerialization(t *testing.T) {
	ctx := context.New("TEST", "1h", 10)
	cd := NewChartData(ctx, "TEST", "1h", "Test Strategy")

	strat := strategy.NewStrategy()
	strat.Call("Test Strategy", 10000)

	/* Trade with both entry and exit comments */
	strat.Entry("long1", strategy.Long, 10, "Buy on breakout")
	strat.OnBarUpdate(1, 100, 1000)
	strat.Close("long1", 110, 2000, "Take profit")

	/* Trade with entry comment only */
	strat.Entry("long2", strategy.Long, 5, "Second entry")
	strat.OnBarUpdate(2, 110, 3000)
	strat.Close("long2", 115, 4000, "")

	cd.AddStrategy(strat, 115)

	jsonBytes, err := cd.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON() failed: %v", err)
	}

	var parsed map[string]interface{}
	err = json.Unmarshal(jsonBytes, &parsed)
	if err != nil {
		t.Fatalf("JSON unmarshal failed: %v", err)
	}

	strategyData := parsed["strategy"].(map[string]interface{})
	trades := strategyData["trades"].([]interface{})

	if len(trades) != 2 {
		t.Fatalf("Expected 2 trades, got %d", len(trades))
	}

	/* Verify first trade has both comments */
	trade1 := trades[0].(map[string]interface{})
	if entryComment, ok := trade1["entryComment"]; ok {
		if entryComment != "Buy on breakout" {
			t.Errorf("Expected 'Buy on breakout', got %v", entryComment)
		}
	} else {
		t.Error("Trade 1 should have entryComment field")
	}
	if exitComment, ok := trade1["exitComment"]; ok {
		if exitComment != "Take profit" {
			t.Errorf("Expected 'Take profit', got %v", exitComment)
		}
	} else {
		t.Error("Trade 1 should have exitComment field")
	}

	/* Verify second trade has entry comment, exit comment omitted */
	trade2 := trades[1].(map[string]interface{})
	if entryComment, ok := trade2["entryComment"]; ok {
		if entryComment != "Second entry" {
			t.Errorf("Expected 'Second entry', got %v", entryComment)
		}
	} else {
		t.Error("Trade 2 should have entryComment field")
	}
	/* exitComment should be omitted (omitempty behavior) */
	if _, ok := trade2["exitComment"]; ok {
		t.Error("Trade 2 should not have exitComment field (omitempty)")
	}
}

/* TestOpenTradeCommentSerialization verifies OpenTrade JSON serialization */
func TestOpenTradeCommentSerialization(t *testing.T) {
	ctx := context.New("TEST", "1h", 10)
	cd := NewChartData(ctx, "TEST", "1h", "Test Strategy")

	strat := strategy.NewStrategy()
	strat.Call("Test Strategy", 10000)

	/* Open trade with entry comment */
	strat.Entry("long1", strategy.Long, 10, "Trend entry")
	strat.OnBarUpdate(1, 100, 1000)

	/* Open trade without entry comment */
	strat.Entry("long2", strategy.Long, 5, "")
	strat.OnBarUpdate(2, 105, 2000)

	cd.AddStrategy(strat, 108)

	jsonBytes, err := cd.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON() failed: %v", err)
	}

	var parsed map[string]interface{}
	err = json.Unmarshal(jsonBytes, &parsed)
	if err != nil {
		t.Fatalf("JSON unmarshal failed: %v", err)
	}

	strategyData := parsed["strategy"].(map[string]interface{})
	openTrades := strategyData["openTrades"].([]interface{})

	if len(openTrades) != 2 {
		t.Fatalf("Expected 2 open trades, got %d", len(openTrades))
	}

	/* Verify first open trade has entry comment */
	openTrade1 := openTrades[0].(map[string]interface{})
	if entryComment, ok := openTrade1["entryComment"]; ok {
		if entryComment != "Trend entry" {
			t.Errorf("Expected 'Trend entry', got %v", entryComment)
		}
	} else {
		t.Error("Open trade 1 should have entryComment field")
	}

	/* Verify second open trade omits empty comment */
	openTrade2 := openTrades[1].(map[string]interface{})
	if _, ok := openTrade2["entryComment"]; ok {
		t.Error("Open trade 2 should not have entryComment field (omitempty)")
	}
}

/* TestTradeCommentOmitEmpty verifies omitempty behavior for empty comments */
func TestTradeCommentOmitEmpty(t *testing.T) {
	ctx := context.New("TEST", "1h", 10)
	cd := NewChartData(ctx, "TEST", "1h", "Test Strategy")

	strat := strategy.NewStrategy()
	strat.Call("Test Strategy", 10000)

	/* Trade with no comments (empty strings) */
	strat.Entry("long1", strategy.Long, 10, "")
	strat.OnBarUpdate(1, 100, 1000)
	strat.Close("long1", 110, 2000, "")

	cd.AddStrategy(strat, 110)

	jsonBytes, err := cd.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON() failed: %v", err)
	}

	var parsed map[string]interface{}
	err = json.Unmarshal(jsonBytes, &parsed)
	if err != nil {
		t.Fatalf("JSON unmarshal failed: %v", err)
	}

	strategyData := parsed["strategy"].(map[string]interface{})
	trades := strategyData["trades"].([]interface{})

	if len(trades) != 1 {
		t.Fatalf("Expected 1 trade, got %d", len(trades))
	}

	trade := trades[0].(map[string]interface{})

	/* Both comment fields should be omitted due to omitempty */
	if _, ok := trade["entryComment"]; ok {
		t.Error("Trade should not have entryComment field (omitempty)")
	}
	if _, ok := trade["exitComment"]; ok {
		t.Error("Trade should not have exitComment field (omitempty)")
	}
}
