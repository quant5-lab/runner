package series

import "fmt"

// Series is a forward-only buffer for Pine Script series variables
// Enforces immutability of historical values and prevents future writes
// Optimized for per-bar forward calculations without array mutations
type Series struct {
	buffer      []float64
	cursor      int
	capacity    int
	initialized bool
}

// NewSeries creates a new series buffer with given capacity
func NewSeries(capacity int) *Series {
	if capacity <= 0 {
		panic(fmt.Sprintf("Series: capacity must be positive, got %d", capacity))
	}

	return &Series{
		buffer:      make([]float64, capacity),
		cursor:      0,
		capacity:    capacity,
		initialized: false,
	}
}

// Set writes value at current cursor position
// Only current bar [0] can be written - historical values are immutable
func (s *Series) Set(value float64) {
	if !s.initialized && s.cursor == 0 {
		s.initialized = true
	}

	if s.cursor >= s.capacity {
		panic(fmt.Sprintf("Series: cursor %d exceeds capacity %d", s.cursor, s.capacity))
	}

	s.buffer[s.cursor] = value
}

// Get retrieves value at specified offset from current cursor
// offset=0 returns current bar, offset=1 returns previous bar, etc.
func (s *Series) Get(offset int) float64 {
	if offset < 0 {
		panic(fmt.Sprintf("Series: negative offset %d not allowed (prevents future access)", offset))
	}

	targetIndex := s.cursor - offset

	if targetIndex < 0 {
		// Warmup period - return 0.0 (Pine Script uses na, we use 0.0)
		return 0.0
	}

	return s.buffer[targetIndex]
}

// GetCurrent returns value at current cursor (equivalent to Get(0))
func (s *Series) GetCurrent() float64 {
	if s.cursor >= s.capacity {
		panic(fmt.Sprintf("Series: cursor %d exceeds capacity %d", s.cursor, s.capacity))
	}
	return s.buffer[s.cursor]
}

// Next advances cursor to next bar (forward-only iteration)
func (s *Series) Next() {
	if s.cursor >= s.capacity-1 {
		panic(fmt.Sprintf("Series: cannot advance beyond capacity %d", s.capacity))
	}
	s.cursor++
}

// Position returns current cursor position
func (s *Series) Position() int {
	return s.cursor
}

// Capacity returns buffer capacity
func (s *Series) Capacity() int {
	return s.capacity
}

// Reset moves cursor to specified position (for recalculation)
func (s *Series) Reset(position int) {
	if position < 0 || position >= s.capacity {
		panic(fmt.Sprintf("Series: invalid reset position %d, capacity is %d", position, s.capacity))
	}
	s.cursor = position
}

// Length returns number of bars processed (cursor + 1)
func (s *Series) Length() int {
	return s.cursor + 1
}
