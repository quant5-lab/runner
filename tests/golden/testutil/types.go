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
	Trades      []Trade              `json:"trades"`
	OpenTrades  []Trade              `json:"openTrades"`
	Equity      float64              `json:"equity"`
	NetProfit   float64              `json:"netProfit"`
	TotalTrades int                  `json:"totalTrades"`
	Plots       map[string][]float64 `json:"plots"`
}

type ChartOutput struct {
	Strategy *StrategyResult `json:"strategy"`
	Plots    []PlotSeries    `json:"plots"`
}

type PlotSeries struct {
	Title  string    `json:"title"`
	Values []float64 `json:"values"`
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
