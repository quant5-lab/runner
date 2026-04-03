package arrayops

import (
	"testing"

	"github.com/quant5-lab/runner/runtime/series"
)

func TestStringAccessor_First_ValidArray(t *testing.T) {
	s := series.NewStringArraySeries(10)
	s.Set([]string{"first", "second", "third"})

	acc := NewStringAccessor()
	result := acc.First(s, 0)

	if result != "first" {
		t.Errorf("First: expected 'first', got '%s'", result)
	}
}

func TestStringAccessor_First_EmptyArray(t *testing.T) {
	s := series.NewStringArraySeries(10)
	s.Set([]string{})

	acc := NewStringAccessor()
	result := acc.First(s, 0)

	if result != "" {
		t.Errorf("First on empty: expected empty string, got '%s'", result)
	}
}

func TestStringAccessor_First_NilArray(t *testing.T) {
	s := series.NewStringArraySeries(10)
	s.Set(nil)

	acc := NewStringAccessor()
	result := acc.First(s, 0)

	if result != "" {
		t.Errorf("First on nil: expected empty string, got '%s'", result)
	}
}

func TestStringAccessor_First_HistoricalAccess(t *testing.T) {
	s := series.NewStringArraySeries(10)
	s.Set([]string{"old", "data"})
	s.Next()
	s.Set([]string{"new", "values"})

	acc := NewStringAccessor()

	current := acc.First(s, 0)
	if current != "new" {
		t.Errorf("First current bar: expected 'new', got '%s'", current)
	}

	previous := acc.First(s, 1)
	if previous != "old" {
		t.Errorf("First previous bar: expected 'old', got '%s'", previous)
	}
}

func TestStringAccessor_Last_ValidArray(t *testing.T) {
	s := series.NewStringArraySeries(10)
	s.Set([]string{"alpha", "beta", "gamma"})

	acc := NewStringAccessor()
	result := acc.Last(s, 0)

	if result != "gamma" {
		t.Errorf("Last: expected 'gamma', got '%s'", result)
	}
}

func TestStringAccessor_Last_SingleElement(t *testing.T) {
	s := series.NewStringArraySeries(10)
	s.Set([]string{"only"})

	acc := NewStringAccessor()
	result := acc.Last(s, 0)

	if result != "only" {
		t.Errorf("Last on single element: expected 'only', got '%s'", result)
	}
}

func TestStringAccessor_Last_EmptyArray(t *testing.T) {
	s := series.NewStringArraySeries(10)
	s.Set([]string{})

	acc := NewStringAccessor()
	result := acc.Last(s, 0)

	if result != "" {
		t.Errorf("Last on empty: expected empty string, got '%s'", result)
	}
}

func TestStringAccessor_Includes_Found(t *testing.T) {
	s := series.NewStringArraySeries(10)
	s.Set([]string{"apple", "banana", "cherry"})

	acc := NewStringAccessor()
	result := acc.Includes(s, 0, "banana")

	if !result {
		t.Error("Includes: expected true for existing element")
	}
}

func TestStringAccessor_Includes_NotFound(t *testing.T) {
	s := series.NewStringArraySeries(10)
	s.Set([]string{"apple", "banana"})

	acc := NewStringAccessor()
	result := acc.Includes(s, 0, "grape")

	if result {
		t.Error("Includes: expected false for missing element")
	}
}

func TestStringAccessor_Includes_EmptyString(t *testing.T) {
	s := series.NewStringArraySeries(10)
	s.Set([]string{"a", "", "b"})

	acc := NewStringAccessor()
	result := acc.Includes(s, 0, "")

	if !result {
		t.Error("Includes: expected true for empty string element")
	}
}

func TestStringAccessor_Includes_EmptyArray(t *testing.T) {
	s := series.NewStringArraySeries(10)
	s.Set([]string{})

	acc := NewStringAccessor()
	if acc.Includes(s, 0, "anything") {
		t.Error("Includes on empty: expected false")
	}
}

func TestStringAccessor_Includes_NilArray(t *testing.T) {
	s := series.NewStringArraySeries(10)
	s.Set(nil)

	acc := NewStringAccessor()
	result := acc.Includes(s, 0, "anything")

	if result {
		t.Error("Includes on nil: expected false")
	}
}

func TestStringAccessor_IndexOf_Found(t *testing.T) {
	s := series.NewStringArraySeries(10)
	s.Set([]string{"zero", "one", "two"})

	acc := NewStringAccessor()
	result := acc.IndexOf(s, 0, "one", 0)

	if result != 1 {
		t.Errorf("IndexOf: expected 1, got %d", result)
	}
}

func TestStringAccessor_IndexOf_NotFound(t *testing.T) {
	s := series.NewStringArraySeries(10)
	s.Set([]string{"a", "b", "c"})

	acc := NewStringAccessor()
	result := acc.IndexOf(s, 0, "z", 0)

	if result != -1 {
		t.Errorf("IndexOf not found: expected -1, got %d", result)
	}
}

func TestStringAccessor_IndexOf_WithStartIndex(t *testing.T) {
	s := series.NewStringArraySeries(10)
	s.Set([]string{"a", "b", "c", "b", "d"})

	acc := NewStringAccessor()
	result := acc.IndexOf(s, 0, "b", 2)

	if result != 3 {
		t.Errorf("IndexOf with startIndex: expected 3, got %d", result)
	}
}

func TestStringAccessor_IndexOf_NegativeStartIndex(t *testing.T) {
	s := series.NewStringArraySeries(10)
	s.Set([]string{"a", "b", "c"})

	acc := NewStringAccessor()
	result := acc.IndexOf(s, 0, "b", -5)

	if result != 1 {
		t.Errorf("IndexOf negative start: expected 1, got %d", result)
	}
}

func TestStringAccessor_IndexOf_EmptyArray(t *testing.T) {
	s := series.NewStringArraySeries(10)
	s.Set([]string{})

	acc := NewStringAccessor()
	result := acc.IndexOf(s, 0, "anything", 0)

	if result != -1 {
		t.Errorf("IndexOf on empty: expected -1, got %d", result)
	}
}

func TestStringAccessor_LastIndexOf_Found(t *testing.T) {
	s := series.NewStringArraySeries(10)
	s.Set([]string{"apple", "banana", "apple", "cherry"})

	acc := NewStringAccessor()
	result := acc.LastIndexOf(s, 0, "apple", 3)

	if result != 2 {
		t.Errorf("LastIndexOf: expected 2, got %d", result)
	}
}

func TestStringAccessor_LastIndexOf_WithStartIndex(t *testing.T) {
	s := series.NewStringArraySeries(10)
	s.Set([]string{"a", "b", "c", "b", "d"})

	acc := NewStringAccessor()
	result := acc.LastIndexOf(s, 0, "b", 2)

	if result != 1 {
		t.Errorf("LastIndexOf with startIndex: expected 1, got %d", result)
	}
}

func TestStringAccessor_LastIndexOf_StartIndexOutOfBounds(t *testing.T) {
	s := series.NewStringArraySeries(10)
	s.Set([]string{"a", "b", "c"})

	acc := NewStringAccessor()
	result := acc.LastIndexOf(s, 0, "b", 100)

	if result != 1 {
		t.Errorf("LastIndexOf large startIndex: expected 1, got %d", result)
	}
}

func TestStringAccessor_LastIndexOf_NotFound(t *testing.T) {
	s := series.NewStringArraySeries(10)
	s.Set([]string{"a", "b", "c"})

	acc := NewStringAccessor()
	result := acc.LastIndexOf(s, 0, "z", 2)

	if result != -1 {
		t.Errorf("LastIndexOf not found: expected -1, got %d", result)
	}
}

func TestStringAccessor_LastIndexOf_EmptyArray(t *testing.T) {
	s := series.NewStringArraySeries(10)
	s.Set([]string{})

	acc := NewStringAccessor()
	result := acc.LastIndexOf(s, 0, "a", 0)

	if result != -1 {
		t.Errorf("LastIndexOf on empty: expected -1, got %d", result)
	}
}
