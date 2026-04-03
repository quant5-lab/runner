package series

import (
	"math"
	"testing"
)

// TestNewArraySeries verifies initial state after construction.
func TestNewArraySeries(t *testing.T) {
	s := NewArraySeries(10)

	if s.Capacity() != 10 {
		t.Errorf("Capacity() = %d, want 10", s.Capacity())
	}
	if s.Position() != 0 {
		t.Errorf("Position() = %d, want 0", s.Position())
	}
	if s.Length() != 1 {
		t.Errorf("Length() = %d, want 1", s.Length())
	}
}

func TestNewArraySeriesInvalidCapacity(t *testing.T) {
	for _, cap := range []int{0, -1, -100} {
		cap := cap
		t.Run("zero_or_negative", func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("expected panic for capacity %d", cap)
				}
			}()
			NewArraySeries(cap)
		})
	}
}

// TestArraySeriesSetGet verifies basic Set/Get across three bars.
func TestArraySeriesSetGet(t *testing.T) {
	s := NewArraySeries(5)

	// Bar 0
	s.Set([]float64{1.0, 2.0})
	if got := s.Get(0); got == nil || got[0] != 1.0 || got[1] != 2.0 {
		t.Errorf("Bar 0 Get(0) = %v, want [1 2]", got)
	}

	// Bar 1
	s.Next()
	s.Set([]float64{3.0, 4.0})
	if got := s.Get(0); got == nil || got[0] != 3.0 {
		t.Errorf("Bar 1 Get(0) = %v, want [3 4]", got)
	}
	if got := s.Get(1); got == nil || got[0] != 1.0 {
		t.Errorf("Bar 1 Get(1) = %v, want [1 2]", got)
	}

	// Bar 2
	s.Next()
	s.Set([]float64{5.0})
	if got := s.Get(2); got == nil || got[0] != 1.0 {
		t.Errorf("Bar 2 Get(2) = %v, want [1 2]", got)
	}
}

// TestArraySeriesBeforeFirstBar verifies nil is returned when the requested
// offset falls before bar 0 — the ArraySeries equivalent of NaN in Series.
func TestArraySeriesBeforeFirstBar(t *testing.T) {
	s := NewArraySeries(10)

	// Get(1) on bar 0: target index -1, before first bar
	if got := s.Get(1); got != nil {
		t.Errorf("Get(1) on bar 0 = %v, want nil", got)
	}
	// Large offset also returns nil
	if got := s.Get(5); got != nil {
		t.Errorf("Get(5) on bar 0 = %v, want nil", got)
	}

	// After advancing one bar, Get(2) is still before bar 0
	s.Set([]float64{1.0})
	s.Next()
	s.Set([]float64{2.0})
	if got := s.Get(2); got != nil {
		t.Errorf("Get(2) on bar 1 = %v, want nil", got)
	}
}

// TestArraySeriesSetCopiesSlice verifies that mutating the original slice after
// Set does not affect the stored historical value.
func TestArraySeriesSetCopiesSlice(t *testing.T) {
	s := NewArraySeries(10)
	vals := []float64{1.0, 2.0}
	s.Set(vals)
	vals[0] = 99.0

	if got := s.GetCurrent(); got[0] != 1.0 {
		t.Errorf("Set should copy slice; got[0]=%v after mutating original", got[0])
	}
}

// TestArraySeriesGetCurrentEquivalentToGet0 verifies the two access paths are identical.
func TestArraySeriesGetCurrentEquivalentToGet0(t *testing.T) {
	s := NewArraySeries(10)
	s.Set([]float64{7.0, 8.0})

	current := s.GetCurrent()
	get0 := s.Get(0)
	if current == nil || get0 == nil || current[0] != get0[0] || current[1] != get0[1] {
		t.Errorf("GetCurrent()=%v != Get(0)=%v", current, get0)
	}
}

// TestArraySeriesSetNil verifies nil round-trips through Set/Get without panic.
func TestArraySeriesSetNil(t *testing.T) {
	s := NewArraySeries(10)
	s.Set(nil)

	if got := s.GetCurrent(); got != nil {
		t.Errorf("GetCurrent after Set(nil) = %v, want nil", got)
	}
	if got := s.Get(0); got != nil {
		t.Errorf("Get(0) after Set(nil) = %v, want nil", got)
	}
}

// TestArraySeriesNaNElementsPreserved verifies NaN float64 values inside stored
// slices survive the Set/Get round-trip without corruption.
func TestArraySeriesNaNElementsPreserved(t *testing.T) {
	s := NewArraySeries(10)
	vals := []float64{math.NaN(), 1.0, math.NaN()}
	s.Set(vals)
	got := s.GetCurrent()

	if !math.IsNaN(got[0]) || got[1] != 1.0 || !math.IsNaN(got[2]) {
		t.Errorf("NaN elements not preserved: %v", got)
	}
}

// TestArraySeriesForwardOnlyIteration validates that all historical offsets are
// accurate at every bar as the cursor advances through the full capacity.
func TestArraySeriesForwardOnlyIteration(t *testing.T) {
	values := [][]float64{{100.0}, {110.0}, {120.0}, {130.0}, {140.0}}
	s := NewArraySeries(len(values))

	for i, val := range values {
		s.Set(val)

		if got := s.Get(0); got == nil || got[0] != val[0] {
			t.Errorf("Bar %d: Get(0) = %v, want %v", i, got, val)
		}
		for offset := 1; offset <= i; offset++ {
			expected := values[i-offset][0]
			if got := s.Get(offset); got == nil || got[0] != expected {
				t.Errorf("Bar %d offset %d: Get = %v, want [%v]", i, offset, got, expected)
			}
		}

		if i < len(values)-1 {
			s.Next()
		}
	}
}

// TestArraySeriesImmutability verifies that historical values are never altered
// when subsequent bars are written to the buffer.
func TestArraySeriesImmutability(t *testing.T) {
	s := NewArraySeries(10)
	s.Set([]float64{100.0})
	s.Next()
	s.Set([]float64{110.0})
	s.Next()
	s.Set([]float64{120.0})

	if got := s.Get(2); got == nil || got[0] != 100.0 {
		t.Errorf("Bar 0 mutated: Get(2) = %v, want [100]", got)
	}
	if got := s.Get(1); got == nil || got[0] != 110.0 {
		t.Errorf("Bar 1 mutated: Get(1) = %v, want [110]", got)
	}
}

// TestArraySeriesExceedCapacityPanics verifies Next panics when the cursor
// would advance beyond the allocated capacity.
func TestArraySeriesExceedCapacityPanics(t *testing.T) {
	s := NewArraySeries(3)
	s.Set([]float64{1.0})
	s.Next()
	s.Set([]float64{2.0})
	s.Next()
	s.Set([]float64{3.0})

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic when advancing beyond capacity")
		}
	}()
	s.Next()
}

// TestArraySeriesNegativeOffsetPanics verifies Get panics for negative offsets,
// preventing illegal future access.
func TestArraySeriesNegativeOffsetPanics(t *testing.T) {
	s := NewArraySeries(10)
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for negative offset")
		}
	}()
	s.Get(-1)
}

// TestArraySeriesReset verifies that Reset repositions the cursor and allows
// overwriting from that position.
func TestArraySeriesReset(t *testing.T) {
	s := NewArraySeries(10)
	for i := 0; i < 5; i++ {
		s.Set([]float64{float64(i * 10)})
		if i < 4 {
			s.Next()
		}
	}

	s.Reset(2)
	if s.Position() != 2 {
		t.Errorf("after Reset(2) Position = %d, want 2", s.Position())
	}
	if got := s.GetCurrent()[0]; got != 20.0 {
		t.Errorf("after Reset(2) current value = %v, want 20.0", got)
	}

	s.Set([]float64{999.0})
	if got := s.Get(0)[0]; got != 999.0 {
		t.Errorf("after Reset and Set, Get(0) = %v, want 999.0", got)
	}
}

// TestArraySeriesResetInvalidPositionPanics verifies that out-of-range reset
// positions cause a panic.
func TestArraySeriesResetInvalidPositionPanics(t *testing.T) {
	tests := []struct {
		name     string
		position int
	}{
		{"negative", -1},
		{"at_capacity", 5},
		{"beyond_capacity", 10},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			s := NewArraySeries(5)
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("expected panic for Reset(%d)", tt.position)
				}
			}()
			s.Reset(tt.position)
		})
	}
}

// TestArraySeriesLength verifies Length increments by one for each Next call.
func TestArraySeriesLength(t *testing.T) {
	s := NewArraySeries(10)

	if s.Length() != 1 {
		t.Errorf("initial Length = %d, want 1", s.Length())
	}
	s.Set([]float64{1.0})
	s.Next()
	if s.Length() != 2 {
		t.Errorf("after 1st Next Length = %d, want 2", s.Length())
	}
	s.Set([]float64{2.0})
	s.Next()
	if s.Length() != 3 {
		t.Errorf("after 2nd Next Length = %d, want 3", s.Length())
	}
}

// TestArraySeriesPosition verifies Position tracks the cursor index.
func TestArraySeriesPosition(t *testing.T) {
	s := NewArraySeries(10)

	if s.Position() != 0 {
		t.Errorf("initial Position = %d, want 0", s.Position())
	}
	s.Set([]float64{1.0})
	s.Next()
	if s.Position() != 1 {
		t.Errorf("after 1st Next Position = %d, want 1", s.Position())
	}
	s.Set([]float64{2.0})
	s.Next()
	if s.Position() != 2 {
		t.Errorf("after 2nd Next Position = %d, want 2", s.Position())
	}
}

// TestArraySeriesElem verifies nil-safe element access across bar offsets.
func TestArraySeriesElem(t *testing.T) {
	s := NewArraySeries(10)
	s.Set([]float64{1.0, 2.0, 3.0})
	s.Next()
	s.Set([]float64{4.0, 5.0})

	// current bar
	if got := s.Elem(0, 0); got != 4.0 {
		t.Errorf("Elem(0,0) = %v, want 4.0", got)
	}
	if got := s.Elem(0, 1); got != 5.0 {
		t.Errorf("Elem(0,1) = %v, want 5.0", got)
	}
	// previous bar
	if got := s.Elem(1, 2); got != 3.0 {
		t.Errorf("Elem(1,2) = %v, want 3.0", got)
	}
	// offset before bar 0 returns NaN
	if got := s.Elem(2, 0); !math.IsNaN(got) {
		t.Errorf("Elem(2,0) before bar 0 = %v, want NaN", got)
	}
	// out-of-range element index returns NaN
	if got := s.Elem(0, 5); !math.IsNaN(got) {
		t.Errorf("Elem(0,5) out-of-range = %v, want NaN", got)
	}
	// nil slice returns NaN
	s2 := NewArraySeries(5)
	if got := s2.Elem(0, 0); !math.IsNaN(got) {
		t.Errorf("Elem(0,0) on nil-slot = %v, want NaN", got)
	}
}

func BenchmarkArraySeriesSequentialAccess(b *testing.B) {
	s := NewArraySeries(10000)
	val := []float64{1.0, 2.0, 3.0}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		idx := i % 10000
		if idx == 0 && i > 0 {
			s.Reset(0)
		}
		s.Set(val)
		_ = s.Get(0)
		if idx < 9999 {
			s.Next()
		}
	}
}

func BenchmarkArraySeriesHistoricalAccess(b *testing.B) {
	s := NewArraySeries(1000)
	for i := 0; i < 1000; i++ {
		s.Set([]float64{float64(i)})
		if i < 999 {
			s.Next()
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = s.Get(0)
		_ = s.Get(1)
		_ = s.Get(10)
		_ = s.Get(50)
		_ = s.Get(100)
	}
}
