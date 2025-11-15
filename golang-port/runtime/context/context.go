package context

import "time"

type OHLCV struct {
	Time   int64   `json:"time"`
	Open   float64 `json:"open"`
	High   float64 `json:"high"`
	Low    float64 `json:"low"`
	Close  float64 `json:"close"`
	Volume float64 `json:"volume"`
}

type Context struct {
	Symbol    string
	Timeframe string
	Bars      int
	Data      []OHLCV
	BarIndex  int
}

func New(symbol, timeframe string, bars int) *Context {
	return &Context{
		Symbol:    symbol,
		Timeframe: timeframe,
		Bars:      bars,
		Data:      make([]OHLCV, 0, bars),
		BarIndex:  0,
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
