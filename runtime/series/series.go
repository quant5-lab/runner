package series

import (
	"fmt"
	"math"
)

// Series is a forward-only buffer for Pine Script series variables.
// Enforces immutability of historical values and prevents future writes.
// Optimized for per-bar forward calculations without array mutations.
//
// preHistoryDefault controls what Get returns when the requested offset
// predates bar 0 — NaN for numeric types (Pine na semantics) and 0.0
// for bool types (Pine guarantees bool is never na).
type Series struct {
	buffer            []float64
	cursor            int
	capacity          int
	initialized       bool
	preHistoryDefault float64
}

// NewSeries creates a numeric series whose pre-history reads return NaN,
// matching Pine's na semantics for int/float/color/string series.
func NewSeries(capacity int) *Series {
	return newSeries(capacity, math.NaN())
}

// NewBoolSeries creates a bool series whose pre-history reads return 0.0
// (false), matching Pine's guarantee that bool values are never na.
func NewBoolSeries(capacity int) *Series {
	return newSeries(capacity, 0.0)
}

func newSeries(capacity int, preHistoryDefault float64) *Series {
	if capacity <= 0 {
		panic(fmt.Sprintf("Series: capacity must be positive, got %d", capacity))
	}

	return &Series{
		buffer:            make([]float64, capacity),
		cursor:            0,
		capacity:          capacity,
		initialized:       false,
		preHistoryDefault: preHistoryDefault,
	}
}

func (s *Series) Set(value float64) {
	if !s.initialized && s.cursor == 0 {
		s.initialized = true
	}

	if s.cursor >= s.capacity {
		panic(fmt.Sprintf("Series: cursor %d exceeds capacity %d", s.cursor, s.capacity))
	}

	s.buffer[s.cursor] = value
}

func (s *Series) Get(offset int) float64 {
	if offset < 0 {
		panic(fmt.Sprintf("Series: negative offset %d not allowed (prevents future access)", offset))
	}

	targetIndex := s.cursor - offset

	if targetIndex < 0 {
		return s.preHistoryDefault
	}

	return s.buffer[targetIndex]
}

func (s *Series) GetCurrent() float64 {
	if s.cursor >= s.capacity {
		panic(fmt.Sprintf("Series: cursor %d exceeds capacity %d", s.cursor, s.capacity))
	}
	return s.buffer[s.cursor]
}

func (s *Series) Next() {
	if s.cursor >= s.capacity-1 {
		panic(fmt.Sprintf("Series: cannot advance beyond capacity %d", s.capacity))
	}
	s.cursor++
}

func (s *Series) Position() int {
	return s.cursor
}

func (s *Series) Capacity() int {
	return s.capacity
}

func (s *Series) Reset(position int) {
	if position < 0 || position >= s.capacity {
		panic(fmt.Sprintf("Series: invalid reset position %d, capacity is %d", position, s.capacity))
	}
	s.cursor = position
}

func (s *Series) Length() int {
	return s.cursor + 1
}
