package arrayops

import (
	"testing"

	"github.com/quant5-lab/runner/runtime/series"
)

func TestAccessor_First_ValidArray(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{10.0, 20.0, 30.0})

	acc := NewAccessor()
	result := acc.First(s, 0)

	if result != 10.0 {
		t.Errorf("First: expected 10.0, got %f", result)
	}
}

func TestAccessor_First_EmptyArray(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{})

	acc := NewAccessor()
	result := acc.First(s, 0)

	if result != 0.0 {
		t.Errorf("First on empty: expected 0.0, got %f", result)
	}
}

func TestAccessor_First_NilArray(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set(nil)

	acc := NewAccessor()
	result := acc.First(s, 0)

	if result != 0.0 {
		t.Errorf("First on nil: expected 0.0, got %f", result)
	}
}

func TestAccessor_First_HistoricalAccess(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{1.0, 2.0})
	s.Next()
	s.Set([]float64{10.0, 20.0})

	acc := NewAccessor()

	current := acc.First(s, 0)
	if current != 10.0 {
		t.Errorf("First current bar: expected 10.0, got %f", current)
	}

	previous := acc.First(s, 1)
	if previous != 1.0 {
		t.Errorf("First previous bar: expected 1.0, got %f", previous)
	}
}

func TestAccessor_Last_ValidArray(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{10.0, 20.0, 30.0})

	acc := NewAccessor()
	result := acc.Last(s, 0)

	if result != 30.0 {
		t.Errorf("Last: expected 30.0, got %f", result)
	}
}

func TestAccessor_Last_SingleElement(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{42.0})

	acc := NewAccessor()
	result := acc.Last(s, 0)

	if result != 42.0 {
		t.Errorf("Last on single element: expected 42.0, got %f", result)
	}
}

func TestAccessor_Last_EmptyArray(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{})

	acc := NewAccessor()
	result := acc.Last(s, 0)

	if result != 0.0 {
		t.Errorf("Last on empty: expected 0.0, got %f", result)
	}
}

func TestAccessor_Includes_Found(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{1.0, 2.0, 3.0, 4.0, 5.0})

	acc := NewAccessor()

	tests := []float64{1.0, 3.0, 5.0}
	for _, val := range tests {
		if !acc.Includes(s, 0, val) {
			t.Errorf("Includes: expected true for %f", val)
		}
	}
}

func TestAccessor_Includes_NotFound(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{1.0, 2.0, 3.0})

	acc := NewAccessor()

	tests := []float64{0.0, 4.0, 10.0, -1.0}
	for _, val := range tests {
		if acc.Includes(s, 0, val) {
			t.Errorf("Includes: expected false for %f", val)
		}
	}
}

func TestAccessor_Includes_EmptyArray(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{})

	acc := NewAccessor()
	if acc.Includes(s, 0, 1.0) {
		t.Error("Includes on empty: expected false")
	}
}

func TestAccessor_Includes_NilArray(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set(nil)

	acc := NewAccessor()
	if acc.Includes(s, 0, 1.0) {
		t.Error("Includes on nil: expected false")
	}
}

func TestAccessor_IndexOf_Found(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{10.0, 20.0, 30.0, 20.0, 40.0})

	acc := NewAccessor()

	if idx := acc.IndexOf(s, 0, 20.0, 0); idx != 1 {
		t.Errorf("IndexOf: expected 1, got %d", idx)
	}
}

func TestAccessor_IndexOf_NotFound(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{10.0, 20.0, 30.0})

	acc := NewAccessor()

	if idx := acc.IndexOf(s, 0, 99.0, 0); idx != -1 {
		t.Errorf("IndexOf not found: expected -1, got %d", idx)
	}
}

func TestAccessor_IndexOf_WithStartIndex(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{10.0, 20.0, 30.0, 20.0, 40.0})

	acc := NewAccessor()

	if idx := acc.IndexOf(s, 0, 20.0, 2); idx != 3 {
		t.Errorf("IndexOf with startIndex: expected 3, got %d", idx)
	}
}

func TestAccessor_IndexOf_NegativeStartIndexNormalized(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{10.0, 20.0, 30.0})

	acc := NewAccessor()

	if idx := acc.IndexOf(s, 0, 10.0, -5); idx != 0 {
		t.Errorf("IndexOf with negative start: expected 0, got %d", idx)
	}
}

func TestAccessor_IndexOf_EmptyArray(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{})

	acc := NewAccessor()

	if idx := acc.IndexOf(s, 0, 1.0, 0); idx != -1 {
		t.Errorf("IndexOf on empty: expected -1, got %d", idx)
	}
}

func TestAccessor_LastIndexOf_Found(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{10.0, 20.0, 30.0, 20.0, 40.0})

	acc := NewAccessor()

	if idx := acc.LastIndexOf(s, 0, 20.0, 4); idx != 3 {
		t.Errorf("LastIndexOf: expected 3, got %d", idx)
	}
}

func TestAccessor_LastIndexOf_NotFound(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{10.0, 20.0, 30.0})

	acc := NewAccessor()

	if idx := acc.LastIndexOf(s, 0, 99.0, 2); idx != -1 {
		t.Errorf("LastIndexOf not found: expected -1, got %d", idx)
	}
}

func TestAccessor_LastIndexOf_WithStartIndex(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{10.0, 20.0, 30.0, 20.0, 40.0})

	acc := NewAccessor()

	if idx := acc.LastIndexOf(s, 0, 20.0, 2); idx != 1 {
		t.Errorf("LastIndexOf with startIndex: expected 1, got %d", idx)
	}
}

func TestAccessor_LastIndexOf_StartIndexOutOfBounds(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{10.0, 20.0, 30.0})

	acc := NewAccessor()

	if idx := acc.LastIndexOf(s, 0, 20.0, 100); idx != 1 {
		t.Errorf("LastIndexOf large startIndex: expected 1, got %d", idx)
	}
}

func TestAccessor_LastIndexOf_EmptyArray(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{})

	acc := NewAccessor()

	if idx := acc.LastIndexOf(s, 0, 1.0, 0); idx != -1 {
		t.Errorf("LastIndexOf on empty: expected -1, got %d", idx)
	}
}
