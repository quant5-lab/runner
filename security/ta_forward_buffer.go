package security

import "github.com/quant5-lab/runner/runtime/series"

// bufferCore is the shared cursor-and-series state for all forward-series
// accumulation primitives. Both forwardBuffer and forwardBufferE embed this.
type bufferCore struct {
	buf      *series.Series
	computed int
}

func (c *bufferCore) at(barIdx int) float64 {
	return c.buf.Get(c.buf.Position() - barIdx)
}

// prev returns the last committed bar value. Callers use this inside
// advanceTo callbacks for self-referential recurrences (EMA, RMA, CUM, …).
func (c *bufferCore) prev() float64 {
	return c.buf.Get(1)
}

func (c *bufferCore) growsFor(contextSize int) bool {
	return contextSize > c.buf.Capacity()
}

func (c *bufferCore) reallocate(capacity int) {
	c.buf = series.NewSeries(capacity)
	c.computed = 0
}

// ── forwardBuffer ─────────────────────────────────────────────────────────────
// Used by pivot and valuewhen whose per-bar callbacks cannot fail.

type forwardBuffer struct{ bufferCore }

func newForwardBuffer(capacity int) forwardBuffer {
	return forwardBuffer{bufferCore{buf: series.NewSeries(max(capacity, 1))}}
}

func (b *forwardBuffer) advanceTo(barIdx int, computeBar func(currentBar int) float64) {
	for b.computed <= barIdx {
		if b.computed > 0 {
			b.buf.Next()
		}
		b.buf.Set(computeBar(b.computed))
		b.computed++
	}
}

// ── forwardBufferE ────────────────────────────────────────────────────────────
// Used by evaluator-dependent managers (SMA, EMA, RMA, RSI, STDEV, CUM,
// BarsSince, ATR, SAR, TSI, volume indicators) whose callbacks may fail.

type forwardBufferE struct{ bufferCore }

func newForwardBufferE(capacity int) forwardBufferE {
	return forwardBufferE{bufferCore{buf: series.NewSeries(max(capacity, 1))}}
}

func (b *forwardBufferE) advanceTo(barIdx int, computeBar func(currentBar int) (float64, error)) error {
	for b.computed <= barIdx {
		if b.computed > 0 {
			b.buf.Next()
		}
		val, err := computeBar(b.computed)
		if err != nil {
			return err
		}
		b.buf.Set(val)
		b.computed++
	}
	return nil
}
