package testutil

type Bar struct {
	Time   int64   `json:"time"`
	Open   float64 `json:"open"`
	High   float64 `json:"high"`
	Low    float64 `json:"low"`
	Close  float64 `json:"close"`
	Volume float64 `json:"volume"`
}

type MarketData struct {
	Symbol    string `json:"symbol"`
	Timeframe string `json:"timeframe"`
	Period    string `json:"period"`
	Bars      []Bar  `json:"bars"`
}

type Trade struct {
	EntryID      string  `json:"entryId"`
	EntryBar     int     `json:"entryBar"`
	EntryTime    int64   `json:"entryTime"`
	EntryPrice   float64 `json:"entryPrice"`
	EntryComment string  `json:"entryComment"`
	ExitBar      int     `json:"exitBar"`
	ExitTime     int64   `json:"exitTime"`
	ExitPrice    float64 `json:"exitPrice"`
	ExitComment  string  `json:"exitComment"`
	Size         float64 `json:"size"`
	Profit       float64 `json:"profit"`
	Direction    string  `json:"direction"`
}

type StrategyResult struct {
	Trades         []Trade              `json:"trades"`
	OpenTrades     []Trade              `json:"openTrades"`
	Equity         float64              `json:"equity"`
	NetProfit      float64              `json:"netProfit"`
	TotalTrades    int                  `json:"totalTrades"`
	InitialCapital float64              `json:"initialCapital"`
	Plots          map[string][]float64 `json:"plots"`
	// MarkClose must equal the mark the strategy binary used for equity so the
	// open-position assertion is exact rather than tautological.
	MarkClose float64 `json:"-"`
	// FromRunner forces the open-position equity assertion even when MarkClose is zero,
	// preventing a zero mark from silently bypassing verification on real runner results.
	FromRunner bool                 `json:"-"`
	Indicators map[string][]float64 `json:"-"`
}

type ChartOutput struct {
	Strategy    *StrategyResult            `json:"strategy"`
	Plots       []PlotSeries               `json:"plots"`
	Indicators  map[string]IndicatorSeries `json:"indicators"`
	Candlestick []Bar                      `json:"candlestick"`
}

type PlotSeries struct {
	Title  string    `json:"title"`
	Values []float64 `json:"values"`
}

// Runner binary outputs these under the "indicators" JSON key, separate from "plots".
type IndicatorSeries struct {
	Title string         `json:"title"`
	Data  []IndicatorBar `json:"data"`
}

// Value is a pointer so that JSON null (Pine na) unmarshals as nil rather than 0.
type IndicatorBar struct {
	Time  int64    `json:"time"`
	Value *float64 `json:"value"`
}

type GoldenFile struct {
	Version        string          `json:"version"`
	Strategy       string          `json:"strategy"`
	DataSource     string          `json:"dataSource"`
	GeneratedAt    string          `json:"generatedAt"`
	StrategyResult *StrategyResult `json:"result"`
}

type TestAsset struct {
	Symbol    string
	Exchange  string
	Timeframe string
	DataFile  string
}
