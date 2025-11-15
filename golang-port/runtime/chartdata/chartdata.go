package chartdata

import (
	"encoding/json"
	"time"

	"github.com/borisquantlab/pinescript-go/runtime/context"
	"github.com/borisquantlab/pinescript-go/runtime/output"
	"github.com/borisquantlab/pinescript-go/runtime/strategy"
)

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
	Time  int64                  `json:"time"`
	Value float64                `json:"value"`
	Options map[string]interface{} `json:"options,omitempty"`
}

/* PlotSeries represents a plot series */
type PlotSeries struct {
	Title string      `json:"title"`
	Data  []PlotPoint `json:"data"`
	Pane  string      `json:"pane,omitempty"`
}

/* ChartData represents complete chart output */
type ChartData struct {
	Candlestick []context.OHLCV        `json:"candlestick"`
	Plots       map[string]PlotSeries  `json:"plots"`
	Strategy    *StrategyData          `json:"strategy,omitempty"`
	Timestamp   string                 `json:"timestamp"`
}

/* NewChartData creates a new chart data structure */
func NewChartData(ctx *context.Context) *ChartData {
	return &ChartData{
		Candlestick: ctx.Data,
		Plots:       make(map[string]PlotSeries),
		Timestamp:   time.Now().Format(time.RFC3339),
	}
}

/* AddPlots adds plot data to chart */
func (cd *ChartData) AddPlots(collector *output.Collector) {
	series := collector.GetSeries()
	for _, s := range series {
		plotPoints := make([]PlotPoint, len(s.Data))
		for i, p := range s.Data {
			plotPoints[i] = PlotPoint{
				Time:    p.Time,
				Value:   p.Value,
				Options: p.Options,
			}
		}
		cd.Plots[s.Title] = PlotSeries{
			Title: s.Title,
			Data:  plotPoints,
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
