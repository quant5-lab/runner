package security

import "github.com/quant5-lab/runner/runtime/series"

// forwardBuffer accumulates bar values lazily via a caller-supplied function,
// advancing only as far as needed and keeping the full history accessible
// via O(1) cursor offset — the canonical ForwardSeriesBuffer pattern for
// self-contained security TA state managers.
type forwardBuffer struct {
	buf      *series.Series
	computed int
}

func newForwardBuffer(capacity int) forwardBuffer {
	return forwardBuffer{buf: series.NewSeries(max(capacity, 1))}
}

func (b *forwardBuffer) growsFor(contextSize int) bool {
	return contextSize > b.buf.Capacity()
}

func (b *forwardBuffer) reallocate(capacity int) {
	b.buf = series.NewSeries(capacity)
	b.computed = 0
}

// advanceTo drives the catch-up loop for all bars not yet accumulated.
// The explicit currentBar parameter prevents closure-over-mutable-state bugs.
func (b *forwardBuffer) advanceTo(barIdx int, computeBar func(currentBar int) float64) {
	for b.computed <= barIdx {
		if b.computed > 0 {
			b.buf.Next()
		}
		b.buf.Set(computeBar(b.computed))
		b.computed++
	}
}

func (b *forwardBuffer) at(barIdx int) float64 {
	return b.buf.Get(b.buf.Position() - barIdx)
}
