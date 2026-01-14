package series

import (
	"testing"
)

func TestNewSeries(t *testing.T) {
	s := NewSeries(10)

	if s.Capacity() != 10 {
		t.Errorf("Expected capacity 10, got %d", s.Capacity())
	}

	if s.Position() != 0 {
		t.Errorf("Expected initial position 0, got %d", s.Position())
	}
}

func TestNewSeriesInvalidCapacity(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for zero capacity")
		}
	}()
	NewSeries(0)
}

func TestSeriesSetGet(t *testing.T) {
	s := NewSeries(5)

	// Bar 0
	s.Set(100.0)
	if got := s.Get(0); got != 100.0 {
		t.Errorf("Bar 0: expected 100.0, got %f", got)
	}

	// Bar 1
	s.Next()
	s.Set(110.0)
	if got := s.Get(0); got != 110.0 {
		t.Errorf("Bar 1 current: expected 110.0, got %f", got)
	}
	if got := s.Get(1); got != 100.0 {
		t.Errorf("Bar 1 previous: expected 100.0, got %f", got)
	}

	// Bar 2
	s.Next()
	s.Set(120.0)
	if got := s.Get(0); got != 120.0 {
		t.Errorf("Bar 2 current: expected 120.0, got %f", got)
	}
	if got := s.Get(1); got != 110.0 {
		t.Errorf("Bar 2 [1]: expected 110.0, got %f", got)
	}
	if got := s.Get(2); got != 100.0 {
		t.Errorf("Bar 2 [2]: expected 100.0, got %f", got)
	}
}

func TestSeriesWarmupPeriod(t *testing.T) {
	s := NewSeries(10)

	// Bar 0: no history, Get(1) should return 0.0
	s.Set(100.0)
	if got := s.Get(1); got != 0.0 {
		t.Errorf("Warmup: expected 0.0 for Get(1) on first bar, got %f", got)
	}

	if got := s.Get(5); got != 0.0 {
		t.Errorf("Warmup: expected 0.0 for Get(5) on first bar, got %f", got)
	}
}

func TestSeriesNegativeOffsetPanics(t *testing.T) {
	s := NewSeries(10)
	s.Set(100.0)

	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for negative offset")
		}
	}()
	s.Get(-1)
}

func TestSeriesExceedCapacityPanics(t *testing.T) {
	s := NewSeries(3)
	s.Set(100.0)
	s.Next()
	s.Set(110.0)
	s.Next()
	s.Set(120.0)

	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic when advancing beyond capacity")
		}
	}()
	s.Next() // This should panic
}

func TestSeriesForwardOnlyIteration(t *testing.T) {
	s := NewSeries(5)

	values := []float64{100, 110, 120, 130, 140}

	for i, val := range values {
		s.Set(val)

		// Verify current value
		if got := s.Get(0); got != val {
			t.Errorf("Bar %d: expected current value %f, got %f", i, val, got)
		}

		// Verify all historical values are accessible
		for offset := 1; offset <= i; offset++ {
			expected := values[i-offset]
			if got := s.Get(offset); got != expected {
				t.Errorf("Bar %d offset %d: expected %f, got %f", i, offset, expected, got)
			}
		}

		if i < len(values)-1 {
			s.Next()
		}
	}
}

func TestSeriesImmutability(t *testing.T) {
	s := NewSeries(10)

	// Set value at bar 0
	s.Set(100.0)
	originalValue := s.Get(0)

	// Move to bar 1
	s.Next()
	s.Set(110.0)

	// Verify bar 0 value hasn't changed (accessed via offset)
	if got := s.Get(1); got != originalValue {
		t.Errorf("Historical value mutated: expected %f, got %f", originalValue, got)
	}

	// Move to bar 2
	s.Next()
	s.Set(120.0)

	// Verify bar 0 and bar 1 values are still intact
	if got := s.Get(2); got != 100.0 {
		t.Errorf("Bar 0 value mutated: expected 100.0, got %f", got)
	}
	if got := s.Get(1); got != 110.0 {
		t.Errorf("Bar 1 value mutated: expected 110.0, got %f", got)
	}
}

func TestSeriesPosition(t *testing.T) {
	s := NewSeries(10)

	if s.Position() != 0 {
		t.Errorf("Initial position: expected 0, got %d", s.Position())
	}

	s.Set(100.0)
	s.Next()
	if s.Position() != 1 {
		t.Errorf("After Next: expected 1, got %d", s.Position())
	}

	s.Set(110.0)
	s.Next()
	if s.Position() != 2 {
		t.Errorf("After 2nd Next: expected 2, got %d", s.Position())
	}
}

func TestSeriesReset(t *testing.T) {
	s := NewSeries(10)

	// Fill some bars
	for i := 0; i < 5; i++ {
		s.Set(float64(100 + i*10))
		if i < 4 {
			s.Next()
		}
	}

	if s.Position() != 4 {
		t.Errorf("Before reset: expected position 4, got %d", s.Position())
	}

	// Reset to position 2
	s.Reset(2)
	if s.Position() != 2 {
		t.Errorf("After reset: expected position 2, got %d", s.Position())
	}

	// Can overwrite from this position
	s.Set(999.0)
	if got := s.Get(0); got != 999.0 {
		t.Errorf("After reset and set: expected 999.0, got %f", got)
	}
}

func TestSeriesLength(t *testing.T) {
	s := NewSeries(10)

	if s.Length() != 1 {
		t.Errorf("Initial length: expected 1, got %d", s.Length())
	}

	s.Set(100.0)
	s.Next()
	if s.Length() != 2 {
		t.Errorf("After 1 Next: expected 2, got %d", s.Length())
	}

	s.Set(110.0)
	s.Next()
	if s.Length() != 3 {
		t.Errorf("After 2 Next: expected 3, got %d", s.Length())
	}
}

func BenchmarkSeriesSequentialAccess(b *testing.B) {
	s := NewSeries(10000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		idx := i % 10000
		if idx == 0 && i > 0 {
			s.Reset(0)
		}
		s.Set(float64(idx))
		_ = s.Get(0)
		if idx < 9999 {
			s.Next()
		}
	}
}

func BenchmarkSeriesHistoricalAccess(b *testing.B) {
	s := NewSeries(1000)

	// Populate series
	for i := 0; i < 1000; i++ {
		s.Set(float64(i))
		if i < 999 {
			s.Next()
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Access various historical offsets
		_ = s.Get(0)
		_ = s.Get(1)
		_ = s.Get(10)
		_ = s.Get(50)
		_ = s.Get(100)
	}
}
