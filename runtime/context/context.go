package context

import (
	"time"

	"github.com/quant5-lab/runner/runtime/series"
)

type OHLCV struct {
	Time   int64   `json:"time"`
	Open   float64 `json:"open"`
	High   float64 `json:"high"`
	Low    float64 `json:"low"`
	Close  float64 `json:"close"`
	Volume float64 `json:"volume"`
}

type Context struct {
	Symbol           string
	Timeframe        string
	Timezone         string // Exchange timezone: "UTC" (Binance), "America/New_York" (NYSE/Yahoo), "Europe/Moscow" (MOEX)
	ReferenceSession string
	Bars             int
	Data             []OHLCV
	BarIndex         int
	IsMonthly        bool
	IsDaily          bool
	IsWeekly         bool
	IsIntraday       bool

	parent           *Context
	variableResolver VariableResolver
	seriesRegistry   SeriesRegistry
}

func New(symbol, timeframe string, bars int) *Context {
	return &Context{
		Symbol:           symbol,
		Timeframe:        timeframe,
		Timezone:         "UTC", // Default to UTC, should be set by provider
		ReferenceSession: "always-open",
		Bars:             bars,
		Data:             make([]OHLCV, 0, bars),
		BarIndex:         0,
		IsMonthly:        IsMonthlyTimeframe(timeframe),
		IsDaily:          IsDailyTimeframe(timeframe),
		IsWeekly:         IsWeeklyTimeframe(timeframe),
		IsIntraday:       IsIntradayTimeframe(timeframe),
	}
}

func (c *Context) AddBar(bar OHLCV) {
	c.Data = append(c.Data, bar)
}

func (c *Context) GetClose(offset int) float64 {
	idx := c.BarIndex - offset
	if idx < 0 || idx >= len(c.Data) {
		return 0
	}
	return c.Data[idx].Close
}

func (c *Context) GetOpen(offset int) float64 {
	idx := c.BarIndex - offset
	if idx < 0 || idx >= len(c.Data) {
		return 0
	}
	return c.Data[idx].Open
}

func (c *Context) GetHigh(offset int) float64 {
	idx := c.BarIndex - offset
	if idx < 0 || idx >= len(c.Data) {
		return 0
	}
	return c.Data[idx].High
}

func (c *Context) GetLow(offset int) float64 {
	idx := c.BarIndex - offset
	if idx < 0 || idx >= len(c.Data) {
		return 0
	}
	return c.Data[idx].Low
}

func (c *Context) GetVolume(offset int) float64 {
	idx := c.BarIndex - offset
	if idx < 0 || idx >= len(c.Data) {
		return 0
	}
	return c.Data[idx].Volume
}

func (c *Context) GetTime(offset int) time.Time {
	idx := c.BarIndex - offset
	if idx < 0 || idx >= len(c.Data) {
		return time.Time{}
	}
	return time.Unix(c.Data[idx].Time, 0)
}

func (c *Context) LastBarIndex() int {
	return len(c.Data) - 1
}

func (c *Context) SetParent(parent *Context, barAligner BarAligner) {
	c.parent = parent

	if c.seriesRegistry == nil {
		c.seriesRegistry = NewMapBasedRegistry()
	}

	var parentResolver VariableResolver
	if parent != nil && parent.variableResolver != nil {
		parentResolver = parent.variableResolver
	}

	c.variableResolver = NewRecursiveResolver(
		c.seriesRegistry,
		barAligner,
		parentResolver,
	)
}

func (c *Context) ResolveVariable(name string) VariableResolutionResult {
	if c.variableResolver == nil {
		return VariableResolutionResult{Found: false}
	}
	return c.variableResolver.Resolve(name, c.BarIndex)
}

func (c *Context) RegisterSeries(name string, series *series.Series) {
	if c.seriesRegistry == nil {
		c.seriesRegistry = NewMapBasedRegistry()
	}
	c.seriesRegistry.Set(name, series)
}

func (c *Context) LookupSeries(name string) (*series.Series, bool) {
	if c.seriesRegistry == nil {
		return nil, false
	}
	return c.seriesRegistry.Get(name)
}

func (c *Context) GetParent() *Context {
	return c.parent
}

/* Timeframe type detection helpers */
func IsMonthlyTimeframe(tf string) bool {
	return tf == "M" || tf == "1M" || tf == "1mo"
}

func IsDailyTimeframe(tf string) bool {
	return tf == "D" || tf == "1D" || tf == "1d"
}

func IsWeeklyTimeframe(tf string) bool {
	return tf == "W" || tf == "1W" || tf == "1w" || tf == "1wk"
}

func IsIntradayTimeframe(tf string) bool {
	return !IsMonthlyTimeframe(tf) && !IsDailyTimeframe(tf) && !IsWeeklyTimeframe(tf)
}
