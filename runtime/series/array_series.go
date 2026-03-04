package series

import (
	"fmt"
	"math"
)

// ArraySeries is a ForwardSeriesBuffer for []float64 values.
//
// Mirrors Series semantics exactly:
//   - Forward-only writes via Set
//   - Immutable historical values via Get(offset)
//   - Cursor advances once per bar via Next
//   - Get returns nil (not NaN) when the requested offset falls before bar 0
//
// Nil return from Get signals "no value exists at that bar" — callers check for nil
// the same way callers check for math.NaN() on Series.
type ArraySeries struct {
	buffer   [][]float64
	cursor   int
	capacity int
}

// NewArraySeries creates an ArraySeries with the given capacity.
func NewArraySeries(capacity int) *ArraySeries {
	if capacity <= 0 {
		panic(fmt.Sprintf("ArraySeries: capacity must be positive, got %d", capacity))
	}
	return &ArraySeries{
		buffer:   make([][]float64, capacity),
		cursor:   0,
		capacity: capacity,
	}
}

// Set writes a copy of value at the current cursor position.
// Copies the slice to ensure historical immutability.
func (s *ArraySeries) Set(value []float64) {
	if s.cursor >= s.capacity {
		panic(fmt.Sprintf("ArraySeries: cursor %d exceeds capacity %d", s.cursor, s.capacity))
	}
	if value == nil {
		s.buffer[s.cursor] = nil
		return
	}
	snapshot := make([]float64, len(value))
	copy(snapshot, value)
	s.buffer[s.cursor] = snapshot
}

// Get retrieves the value at the given historical offset from the current bar.
// offset=0 returns the current bar's value; offset=1 the previous bar's value.
// Returns nil when the requested offset falls before bar 0.
func (s *ArraySeries) Get(offset int) []float64 {
	if offset < 0 {
		panic(fmt.Sprintf("ArraySeries: negative offset %d not allowed (prevents future access)", offset))
	}
	targetIndex := s.cursor - offset
	if targetIndex < 0 {
		return nil
	}
	return s.buffer[targetIndex]
}

// GetCurrent returns the value at the current cursor (equivalent to Get(0)).
func (s *ArraySeries) GetCurrent() []float64 {
	if s.cursor >= s.capacity {
		panic(fmt.Sprintf("ArraySeries: cursor %d exceeds capacity %d", s.cursor, s.capacity))
	}
	return s.buffer[s.cursor]
}

// Next advances the cursor to the next bar position.
func (s *ArraySeries) Next() {
	if s.cursor >= s.capacity-1 {
		panic(fmt.Sprintf("ArraySeries: cannot advance beyond capacity %d", s.capacity))
	}
	s.cursor++
}

// Position returns the current cursor position (0-based bar index).
func (s *ArraySeries) Position() int {
	return s.cursor
}

// Capacity returns the total buffer capacity.
func (s *ArraySeries) Capacity() int {
	return s.capacity
}

// Reset moves the cursor to the specified position for replay or recalculation.
func (s *ArraySeries) Reset(position int) {
	if position < 0 || position >= s.capacity {
		panic(fmt.Sprintf("ArraySeries: invalid reset position %d, capacity is %d", position, s.capacity))
	}
	s.cursor = position
}

// Length returns the number of bars processed so far (cursor + 1).
func (s *ArraySeries) Length() int {
	return s.cursor + 1
}

// Elem returns the element at the given bar offset and element index.
// Returns math.NaN() when the slice at that offset is nil or index is out of range.
func (s *ArraySeries) Elem(offset, index int) float64 {
	sl := s.Get(offset)
	if sl == nil || index < 0 || index >= len(sl) {
		return math.NaN()
	}
	return sl[index]
}
