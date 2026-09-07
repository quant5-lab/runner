package chartdata

import (
	"encoding/json"
	"math"
	"time"

	"github.com/quant5-lab/runner/runtime/clock"
	"github.com/quant5-lab/runner/runtime/context"
	"github.com/quant5-lab/runner/runtime/featuregap"
	"github.com/quant5-lab/runner/runtime/output"
	"github.com/quant5-lab/runner/runtime/strategy"
)

/* Metadata contains chart metadata */
type Metadata struct {
	Symbol    string `json:"symbol"`
	Timeframe string `json:"timeframe"`
	Strategy  string `json:"strategy,omitempty"`
	Title     string `json:"title"`
	Timestamp string `json:"timestamp"`
}

/* StyleConfig contains plot styling */
type StyleConfig struct {
	Color     string `json:"color,omitempty"`
	LineWidth int    `json:"lineWidth,omitempty"`
	PlotStyle string `json:"plotStyle,omitempty"`
	Transp    int    `json:"transp,omitempty"`
}

/* IndicatorSeries represents a plot indicator with metadata */
type IndicatorSeries struct {
	Title  string      `json:"title"`
	Pane   string      `json:"pane,omitempty"`
	Style  StyleConfig `json:"style"`
	Offset int         `json:"offset,omitempty"`
	Data   []PlotPoint `json:"data"`
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
	EntryID      string  `json:"entryId"`
	ExitID       string  `json:"exitId,omitempty"`
	EntryPrice   float64 `json:"entryPrice"`
	EntryBar     int     `json:"entryBar"`
	EntryTime    int64   `json:"entryTime"`
	EntryComment string  `json:"entryComment,omitempty"`
	ExitPrice    float64 `json:"exitPrice"`
	ExitBar      int     `json:"exitBar"`
	ExitTime     int64   `json:"exitTime"`
	ExitComment  string  `json:"exitComment,omitempty"`
	Size         float64 `json:"size"`
	Profit       float64 `json:"profit"`
	Commission   float64 `json:"commission"`
	Direction    string  `json:"direction"`
}

/* OpenTrade represents an open trade in chart data */
type OpenTrade struct {
	EntryID      string  `json:"entryId"`
	EntryPrice   float64 `json:"entryPrice"`
	EntryBar     int     `json:"entryBar"`
	EntryTime    int64   `json:"entryTime"`
	EntryComment string  `json:"entryComment,omitempty"`
	Size         float64 `json:"size"`
	Direction    string  `json:"direction"`
}

/* StrategyData represents strategy execution results */
type StrategyData struct {
	Trades         []Trade     `json:"trades"`
	OpenTrades     []OpenTrade `json:"openTrades"`
	Equity         float64     `json:"equity"`
	NetProfit      float64     `json:"netProfit"`
	InitialCapital float64     `json:"initialCapital"`
}

type CompatibilityData struct {
	Status      string           `json:"status"`
	Diagnostics []featuregap.Hit `json:"diagnostics,omitempty"`
}

type BacktestData struct {
	Status        string   `json:"status"`
	AffectedSinks []string `json:"affectedSinks,omitempty"`
}

/* PlotPoint represents a single plot data point */
type PlotPoint struct {
	Time    int64                  `json:"time"`
	Value   float64                `json:"value"`
	Options map[string]interface{} `json:"options,omitempty"`
}

/* MarshalJSON implements custom JSON marshaling to convert NaN to null */
func (p PlotPoint) MarshalJSON() ([]byte, error) {
	type Alias PlotPoint
	var value interface{}
	if math.IsNaN(p.Value) || math.IsInf(p.Value, 0) {
		value = nil
	} else {
		value = p.Value
	}

	return json.Marshal(&struct {
		Time    int64                  `json:"time"`
		Value   interface{}            `json:"value"`
		Options map[string]interface{} `json:"options,omitempty"`
	}{
		Time:    p.Time,
		Value:   value,
		Options: p.Options,
	})
}

/* PlotSeries represents a plot series (deprecated - use IndicatorSeries) */
type PlotSeries struct {
	Title string      `json:"title"`
	Data  []PlotPoint `json:"data"`
	Pane  string      `json:"pane,omitempty"`
}

/* ChartData represents complete unified chart output */
type ChartData struct {
	Metadata            Metadata                   `json:"metadata"`
	ScriptCompatibility CompatibilityData          `json:"scriptCompatibility"`
	Backtest            BacktestData               `json:"backtest"`
	Candlestick         []context.OHLCV            `json:"candlestick"`
	Indicators          map[string]IndicatorSeries `json:"indicators"`
	Strategy            *StrategyData              `json:"strategy,omitempty"`
	UI                  UIConfig                   `json:"ui"`
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
			Timestamp: clock.Now().Format(time.RFC3339),
		},
		ScriptCompatibility: CompatibilityData{Status: "complete"},
		Backtest:            BacktestData{Status: "complete"},
		Candlestick:         ctx.Data,
		Indicators:          make(map[string]IndicatorSeries),
		UI: UIConfig{
			Panes: map[string]PaneConfig{
				"main": {Height: 400, Fixed: true},
			},
		},
	}
}

func (cd *ChartData) AddCompatibility(diagnostics []featuregap.Hit) {
	if len(diagnostics) == 0 {
		cd.ScriptCompatibility = CompatibilityData{Status: "complete"}
		cd.Backtest = BacktestData{Status: "complete"}
		return
	}

	cd.ScriptCompatibility = CompatibilityData{
		Status:      "degraded",
		Diagnostics: diagnostics,
	}
	if featuregap.HasBacktestCritical(diagnostics) {
		cd.Backtest = BacktestData{
			Status:        "unsupported_dependency",
			AffectedSinks: featuregap.BacktestSinks(diagnostics),
		}
		return
	}
	cd.Backtest = BacktestData{Status: "complete"}
}

func (cd *ChartData) BacktestComplete() bool {
	return cd.Backtest.Status == "complete"
}

/* AddPlots adds plot data to chart as indicators */
func (cd *ChartData) AddPlots(collector *output.Collector) {
	series := collector.GetSeries()
	colors := []string{"#2196F3", "#4CAF50", "#FF9800", "#F44336", "#9C27B0", "#00BCD4"}

	for i, s := range series {
		plotPoints := make([]PlotPoint, len(s.Data))
		offset := 0
		color := ""
		lineWidth := 0
		style := ""
		pane := ""
		transp := 0

		for j, p := range s.Data {
			plotPoints[j] = PlotPoint{
				Time:    p.Time,
				Value:   p.Value,
				Options: p.Options,
			}

			if p.Options != nil {
				if offset == 0 {
					if offsetVal, ok := p.Options["offset"].(int); ok {
						offset = offsetVal
					} else if offsetValFloat, ok := p.Options["offset"].(float64); ok {
						offset = int(offsetValFloat)
					}
				}
				if color == "" {
					if colorVal, ok := p.Options["color"].(string); ok {
						color = colorVal
					}
				}
				if lineWidth == 0 {
					if lwVal, ok := p.Options["linewidth"].(int); ok {
						lineWidth = lwVal
					} else if lwValFloat, ok := p.Options["linewidth"].(float64); ok {
						lineWidth = int(lwValFloat)
					}
				}
				if style == "" {
					if styleVal, ok := p.Options["style"].(string); ok {
						style = styleVal
					}
				}
				if pane == "" {
					if paneVal, ok := p.Options["pane"].(string); ok {
						pane = paneVal
					}
				}
				if transp == 0 {
					if transpVal, ok := p.Options["transp"].(int); ok {
						transp = transpVal
					} else if transpValFloat, ok := p.Options["transp"].(float64); ok {
						transp = int(transpValFloat)
					}
				}
			}
		}

		if color == "" {
			color = colors[i%len(colors)]
		}
		if lineWidth == 0 {
			lineWidth = 2
		}

		cd.Indicators[s.Title] = IndicatorSeries{
			Title:  s.Title,
			Pane:   pane,
			Offset: offset,
			Style: StyleConfig{
				Color:     color,
				LineWidth: lineWidth,
				PlotStyle: style,
				Transp:    transp,
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
			EntryID:      t.EntryID,
			ExitID:       t.ExitID,
			EntryPrice:   t.EntryPrice,
			EntryBar:     t.EntryBar,
			EntryTime:    t.EntryTime,
			EntryComment: t.EntryComment,
			ExitPrice:    t.ExitPrice,
			ExitBar:      t.ExitBar,
			ExitTime:     t.ExitTime,
			ExitComment:  t.ExitComment,
			Size:         t.Size,
			Profit:       t.Profit,
			Commission:   t.Commission,
			Direction:    t.Direction,
		}
	}

	openTradesData := make([]OpenTrade, len(openTrades))
	for i, t := range openTrades {
		openTradesData[i] = OpenTrade{
			EntryID:      t.EntryID,
			EntryPrice:   t.EntryPrice,
			EntryBar:     t.EntryBar,
			EntryTime:    t.EntryTime,
			EntryComment: t.EntryComment,
			Size:         t.Size,
			Direction:    t.Direction,
		}
	}

	cd.Strategy = &StrategyData{
		Trades:         trades,
		OpenTrades:     openTradesData,
		Equity:         strat.GetEquity(currentPrice),
		NetProfit:      strat.GetNetProfit(),
		InitialCapital: strat.GetInitialCapital(),
	}
}

func (cd *ChartData) ToJSON() ([]byte, error) {
	return json.MarshalIndent(cd, "", "  ")
}
