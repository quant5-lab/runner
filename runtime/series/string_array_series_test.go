package series

import "testing"

func TestStringArraySeries_InitialState(t *testing.T) {
	s := NewStringArraySeries(10)
	if s.Position() != 0 {
		t.Errorf("initial position: expected 0, got %d", s.Position())
	}
	if s.GetCurrent() != nil {
		t.Errorf("initial current: expected nil, got %v", s.GetCurrent())
	}
}

func TestStringArraySeries_SetAndGetCurrent(t *testing.T) {
	s := NewStringArraySeries(10)
	s.Set([]string{"hello", "world"})
	current := s.GetCurrent()
	if len(current) != 2 || current[0] != "hello" || current[1] != "world" {
		t.Errorf("current value mismatch: got %v", current)
	}
}

func TestStringArraySeries_NextAdvancesCursor(t *testing.T) {
	s := NewStringArraySeries(10)
	s.Set([]string{"a"})
	s.Next()
	s.Set([]string{"b"})

	if s.Position() != 1 {
		t.Errorf("position after Next: expected 1, got %d", s.Position())
	}

	current := s.GetCurrent()
	if len(current) != 1 || current[0] != "b" {
		t.Errorf("current after Next: expected [b], got %v", current)
	}
}

func TestStringArraySeries_GetHistorical(t *testing.T) {
	s := NewStringArraySeries(5)
	s.Set([]string{"bar0"})
	s.Next()
	s.Set([]string{"bar1"})
	s.Next()
	s.Set([]string{"bar2"})

	if val := s.Get(0); len(val) != 1 || val[0] != "bar2" {
		t.Errorf("Get(0): expected [bar2], got %v", val)
	}
	if val := s.Get(1); len(val) != 1 || val[0] != "bar1" {
		t.Errorf("Get(1): expected [bar1], got %v", val)
	}
	if val := s.Get(2); len(val) != 1 || val[0] != "bar0" {
		t.Errorf("Get(2): expected [bar0], got %v", val)
	}
}

func TestStringArraySeries_CircularWrap(t *testing.T) {
	s := NewStringArraySeries(3)
	for i := 0; i < 5; i++ {
		s.Set([]string{string(rune('a' + i))})
		s.Next()
	}

	pos := s.Position()
	if pos != 2 {
		t.Errorf("position after 5 Next() with capacity=3: expected 2, got %d", pos)
	}
}
