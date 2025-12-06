package context

import (
	"testing"
)

func TestContextNew(t *testing.T) {
	ctx := New("SBER", "1h", 100)
	if ctx.Symbol != "SBER" {
		t.Errorf("Symbol = %s, want SBER", ctx.Symbol)
	}
	if ctx.Timeframe != "1h" {
		t.Errorf("Timeframe = %s, want 1h", ctx.Timeframe)
	}
	if ctx.Bars != 100 {
		t.Errorf("Bars = %d, want 100", ctx.Bars)
	}
}

func TestContextAddBar(t *testing.T) {
	ctx := New("SBER", "1h", 10)
	bar := OHLCV{
		Time:   1700000000,
		Open:   100.0,
		High:   105.0,
		Low:    99.0,
		Close:  102.0,
		Volume: 1000,
	}
	ctx.AddBar(bar)

	if len(ctx.Data) != 1 {
		t.Errorf("Data length = %d, want 1", len(ctx.Data))
	}
	if ctx.Data[0].Close != 102.0 {
		t.Errorf("Close = %f, want 102.0", ctx.Data[0].Close)
	}
}

func TestContextGetClose(t *testing.T) {
	ctx := New("SBER", "1h", 10)
	ctx.AddBar(OHLCV{Close: 100.0})
	ctx.AddBar(OHLCV{Close: 101.0})
	ctx.AddBar(OHLCV{Close: 102.0})

	ctx.BarIndex = 2

	if got := ctx.GetClose(0); got != 102.0 {
		t.Errorf("GetClose(0) = %f, want 102.0", got)
	}
	if got := ctx.GetClose(1); got != 101.0 {
		t.Errorf("GetClose(1) = %f, want 101.0", got)
	}
	if got := ctx.GetClose(2); got != 100.0 {
		t.Errorf("GetClose(2) = %f, want 100.0", got)
	}
}

func TestContextGetTime(t *testing.T) {
	ctx := New("SBER", "1h", 10)
	timestamp := int64(1700000000)
	ctx.AddBar(OHLCV{Time: timestamp})
	ctx.BarIndex = 0

	tm := ctx.GetTime(0)
	if tm.Unix() != timestamp {
		t.Errorf("GetTime(0) = %d, want %d", tm.Unix(), timestamp)
	}
}

func TestContextBoundsCheck(t *testing.T) {
	ctx := New("SBER", "1h", 10)
	ctx.AddBar(OHLCV{Close: 100.0})
	ctx.BarIndex = 0

	if got := ctx.GetClose(1); got != 0 {
		t.Errorf("GetClose(1) out of bounds = %f, want 0", got)
	}
	if got := ctx.GetClose(-1); got != 0 {
		t.Errorf("GetClose(-1) negative = %f, want 0", got)
	}
}

func TestLastBarIndex(t *testing.T) {
	ctx := New("SBER", "1h", 10)
	ctx.AddBar(OHLCV{})
	ctx.AddBar(OHLCV{})
	ctx.AddBar(OHLCV{})

	if got := ctx.LastBarIndex(); got != 2 {
		t.Errorf("LastBarIndex() = %d, want 2", got)
	}
}
