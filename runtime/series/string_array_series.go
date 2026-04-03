package series

import "fmt"

// StringArraySeries is a ForwardSeriesBuffer for []string values.
//
// Mirrors ArraySeries semantics exactly:
//   - Forward-only writes via Set
//   - Immutable historical values via Get(offset)
//   - Cursor advances once per bar via Next
//   - Get returns nil when the requested offset falls before bar 0
type StringArraySeries struct {
	buffer   [][]string
	cursor   int
	capacity int
}

func NewStringArraySeries(capacity int) *StringArraySeries {
	if capacity <= 0 {
		panic(fmt.Sprintf("StringArraySeries: capacity must be positive, got %d", capacity))
	}
	return &StringArraySeries{
		buffer:   make([][]string, capacity),
		cursor:   0,
		capacity: capacity,
	}
}

func (s *StringArraySeries) Position() int {
	return s.cursor
}

func (s *StringArraySeries) Next() {
	s.cursor = (s.cursor + 1) % s.capacity
}

func (s *StringArraySeries) Set(value []string) {
	s.buffer[s.cursor] = value
}

func (s *StringArraySeries) GetCurrent() []string {
	return s.buffer[s.cursor]
}

func (s *StringArraySeries) Get(offset int) []string {
	if offset < 0 || offset >= s.capacity {
		return nil
	}
	index := (s.cursor - offset + s.capacity) % s.capacity
	return s.buffer[index]
}

func (s *StringArraySeries) Elem(offset, index int) string {
	sl := s.Get(offset)
	if sl == nil || index < 0 || index >= len(sl) {
		return ""
	}
	return sl[index]
}
