package chartdata

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/borisquantlab/pinescript-go/runtime/context"
	"github.com/borisquantlab/pinescript-go/runtime/output"
	"github.com/borisquantlab/pinescript-go/runtime/strategy"
)

func TestNewChartData(t *testing.T) {
	ctx := context.New("TEST", "1h", 10)
	now := time.Now().Unix()

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
	now := time.Now().Unix()

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
	strat.Entry("long1", strategy.Long, 10)
	strat.OnBarUpdate(1, 100, 1000)
	strat.Close("long1", 110, 2000)

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
	now := time.Now().Unix()
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
	strat.Entry("long1", strategy.Long, 5)
	strat.OnBarUpdate(1, 100, 1000)

	// Close trade
	strat.Close("long1", 110, 2000)

	// Another open trade
	strat.Entry("long2", strategy.Long, 3)
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
