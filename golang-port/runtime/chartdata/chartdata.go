package chartdata

import (
	"encoding/json"
	"time"

	"github.com/borisquantlab/pinescript-go/runtime/context"
	"github.com/borisquantlab/pinescript-go/runtime/output"
	"github.com/borisquantlab/pinescript-go/runtime/strategy"
)

/* Metadata contains chart metadata */
type Metadata struct {
	Symbol     string `json:"symbol"`
	Timeframe  string `json:"timeframe"`
	Strategy   string `json:"strategy,omitempty"`
	Title      string `json:"title"`
	Timestamp  string `json:"timestamp"`
}

/* StyleConfig contains plot styling */
type StyleConfig struct {
	Color     string `json:"color,omitempty"`
	LineWidth int    `json:"lineWidth,omitempty"`
}

/* IndicatorSeries represents a plot indicator with metadata */
type IndicatorSeries struct {
	Title string       `json:"title"`
	Pane  string       `json:"pane,omitempty"`
	Style StyleConfig  `json:"style"`
	Data  []PlotPoint  `json:"data"`
}

/* PaneConfig contains pane layout configuration */
type PaneConfig struct {
	Height int  `json:"height"`
	Fixed  bool `json:"fixed,omitempty"`
}

/* UIConfig contains UI hints for visualization */
type UIConfig struct {
	Panes map[string]PaneConfig `json:"panes"`
}

/* Trade represents a closed trade in chart data */
type Trade struct {
	EntryID    string  `json:"entryId"`
	EntryPrice float64 `json:"entryPrice"`
	EntryBar   int     `json:"entryBar"`
	EntryTime  int64   `json:"entryTime"`
	ExitPrice  float64 `json:"exitPrice"`
	ExitBar    int     `json:"exitBar"`
	ExitTime   int64   `json:"exitTime"`
	Size       float64 `json:"size"`
	Profit     float64 `json:"profit"`
	Direction  string  `json:"direction"`
}

/* OpenTrade represents an open trade in chart data */
type OpenTrade struct {
	EntryID    string  `json:"entryId"`
	EntryPrice float64 `json:"entryPrice"`
	EntryBar   int     `json:"entryBar"`
	EntryTime  int64   `json:"entryTime"`
	Size       float64 `json:"size"`
	Direction  string  `json:"direction"`
}

/* StrategyData represents strategy execution results */
type StrategyData struct {
	Trades     []Trade     `json:"trades"`
	OpenTrades []OpenTrade `json:"openTrades"`
	Equity     float64     `json:"equity"`
	NetProfit  float64     `json:"netProfit"`
}

/* PlotPoint represents a single plot data point */
type PlotPoint struct {
	Time    int64                  `json:"time"`
	Value   float64                `json:"value"`
	Options map[string]interface{} `json:"options,omitempty"`
}

/* PlotSeries represents a plot series (deprecated - use IndicatorSeries) */
type PlotSeries struct {
	Title string      `json:"title"`
	Data  []PlotPoint `json:"data"`
	Pane  string      `json:"pane,omitempty"`
}

/* ChartData represents complete unified chart output */
type ChartData struct {
	Metadata    Metadata                   `json:"metadata"`
	Candlestick []context.OHLCV            `json:"candlestick"`
	Indicators  map[string]IndicatorSeries `json:"indicators"`
	Strategy    *StrategyData              `json:"strategy,omitempty"`
	UI          UIConfig                   `json:"ui"`
}

/* NewChartData creates a new chart data structure */
func NewChartData(ctx *context.Context, symbol, timeframe, strategyName string) *ChartData {
	title := symbol
	if strategyName != "" {
		title = strategyName + " - " + symbol
	}
	
	return &ChartData{
		Metadata: Metadata{
			Symbol:    symbol,
			Timeframe: timeframe,
			Strategy:  strategyName,
			Title:     title,
			Timestamp: time.Now().Format(time.RFC3339),
		},
		Candlestick: ctx.Data,
		Indicators:  make(map[string]IndicatorSeries),
		UI: UIConfig{
			Panes: map[string]PaneConfig{
				"main": {Height: 400, Fixed: true},
				"indicator": {Height: 200, Fixed: false},
			},
		},
	}
}

/* AddPlots adds plot data to chart as indicators */
func (cd *ChartData) AddPlots(collector *output.Collector) {
	series := collector.GetSeries()
	colors := []string{"#2196F3", "#4CAF50", "#FF9800", "#F44336", "#9C27B0", "#00BCD4"}
	
	for i, s := range series {
		plotPoints := make([]PlotPoint, len(s.Data))
		for j, p := range s.Data {
			plotPoints[j] = PlotPoint{
				Time:    p.Time,
				Value:   p.Value,
				Options: p.Options,
			}
		}
		
		/* Backend emits raw data without presentation concerns */
		color := colors[i%len(colors)]
		
		cd.Indicators[s.Title] = IndicatorSeries{
			Title: s.Title,
			Pane:  "", /* Presentation layer assigns pane based on range analysis */
			Style: StyleConfig{
				Color:     color,
				LineWidth: 2,
			},
			Data: plotPoints,
		}
	}
}

/* AddStrategy adds strategy data to chart */
func (cd *ChartData) AddStrategy(strat *strategy.Strategy, currentPrice float64) {
	th := strat.GetTradeHistory()
	closedTrades := th.GetClosedTrades()
	openTrades := th.GetOpenTrades()

	trades := make([]Trade, len(closedTrades))
	for i, t := range closedTrades {
		trades[i] = Trade{
			EntryID:    t.EntryID,
			EntryPrice: t.EntryPrice,
			EntryBar:   t.EntryBar,
			EntryTime:  t.EntryTime,
			ExitPrice:  t.ExitPrice,
			ExitBar:    t.ExitBar,
			ExitTime:   t.ExitTime,
			Size:       t.Size,
			Profit:     t.Profit,
			Direction:  t.Direction,
		}
	}

	openTradesData := make([]OpenTrade, len(openTrades))
	for i, t := range openTrades {
		openTradesData[i] = OpenTrade{
			EntryID:    t.EntryID,
			EntryPrice: t.EntryPrice,
			EntryBar:   t.EntryBar,
			EntryTime:  t.EntryTime,
			Size:       t.Size,
			Direction:  t.Direction,
		}
	}

	cd.Strategy = &StrategyData{
		Trades:     trades,
		OpenTrades: openTradesData,
		Equity:     strat.GetEquity(currentPrice),
		NetProfit:  strat.GetNetProfit(),
	}
}

/* ToJSON converts chart data to JSON bytes */
func (cd *ChartData) ToJSON() ([]byte, error) {
	return json.MarshalIndent(cd, "", "  ")
}
