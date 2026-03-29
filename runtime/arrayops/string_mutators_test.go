package arrayops

import (
	"testing"

	"github.com/quant5-lab/runner/runtime/series"
)

func TestStringMutator_Push(t *testing.T) {
	s := series.NewStringArraySeries(10)
	s.Set([]string{"a", "b"})

	mut := NewStringMutator()
	mut.Push(s, "c")

	result := s.GetCurrent()
	if len(result) != 3 {
		t.Fatalf("Push: expected length 3, got %d", len(result))
	}
	if result[2] != "c" {
		t.Errorf("Push: expected result[2] = 'c', got '%s'", result[2])
	}
}

func TestStringMutator_Push_EmptyArray(t *testing.T) {
	s := series.NewStringArraySeries(10)
	s.Set(nil)

	mut := NewStringMutator()
	mut.Push(s, "first")

	result := s.GetCurrent()
	if len(result) != 1 {
		t.Fatalf("Push to nil: expected length 1, got %d", len(result))
	}
	if result[0] != "first" {
		t.Errorf("Push to nil: expected result[0] = 'first', got '%s'", result[0])
	}
}

func TestStringMutator_Pop(t *testing.T) {
	s := series.NewStringArraySeries(10)
	s.Set([]string{"a", "b", "c"})

	mut := NewStringMutator()
	popped := mut.Pop(s)

	if popped != "c" {
		t.Errorf("Pop: expected popped value 'c', got '%s'", popped)
	}

	result := s.GetCurrent()
	if len(result) != 2 {
		t.Fatalf("Pop: expected length 2, got %d", len(result))
	}
}

func TestStringMutator_Pop_EmptyArray(t *testing.T) {
	s := series.NewStringArraySeries(10)
	s.Set([]string{})

	mut := NewStringMutator()
	popped := mut.Pop(s)

	if popped != "" {
		t.Errorf("Pop empty: expected empty string, got '%s'", popped)
	}

	result := s.GetCurrent()
	if result != nil {
		t.Errorf("Pop empty: expected nil, got %v", result)
	}
}

func TestStringMutator_Shift(t *testing.T) {
	s := series.NewStringArraySeries(10)
	s.Set([]string{"first", "second", "third"})

	mut := NewStringMutator()
	shifted := mut.Shift(s)

	if shifted != "first" {
		t.Errorf("Shift: expected shifted value 'first', got '%s'", shifted)
	}

	result := s.GetCurrent()
	if len(result) != 2 {
		t.Fatalf("Shift: expected length 2, got %d", len(result))
	}
	if result[0] != "second" {
		t.Errorf("Shift: expected result[0] = 'second', got '%s'", result[0])
	}
}

func TestStringMutator_Shift_EmptyArray(t *testing.T) {
	s := series.NewStringArraySeries(10)
	s.Set([]string{})

	mut := NewStringMutator()
	shifted := mut.Shift(s)

	if shifted != "" {
		t.Errorf("Shift empty: expected empty string, got '%s'", shifted)
	}

	result := s.GetCurrent()
	if result != nil {
		t.Errorf("Shift empty: expected nil, got %v", result)
	}
}

func TestStringMutator_Unshift(t *testing.T) {
	s := series.NewStringArraySeries(10)
	s.Set([]string{"b", "c"})

	mut := NewStringMutator()
	mut.Unshift(s, "a")

	result := s.GetCurrent()
	if len(result) != 3 {
		t.Fatalf("Unshift: expected length 3, got %d", len(result))
	}
	if result[0] != "a" {
		t.Errorf("Unshift: expected result[0] = 'a', got '%s'", result[0])
	}
}

func TestStringMutator_SetElement(t *testing.T) {
	s := series.NewStringArraySeries(10)
	s.Set([]string{"a", "b", "c"})

	mut := NewStringMutator()
	mut.SetElement(s, 1, "X")

	result := s.GetCurrent()
	if result[1] != "X" {
		t.Errorf("SetElement: expected result[1] = 'X', got '%s'", result[1])
	}
	if result[0] != "a" || result[2] != "c" {
		t.Errorf("SetElement: other elements changed: got %v", result)
	}
}

func TestStringMutator_SetElement_OutOfBounds(t *testing.T) {
	s := series.NewStringArraySeries(10)
	s.Set([]string{"a", "b"})

	mut := NewStringMutator()
	mut.SetElement(s, 5, "X")

	result := s.GetCurrent()
	if len(result) != 2 {
		t.Errorf("SetElement out of bounds: array modified, got %v", result)
	}
}

func TestStringMutator_SetElement_NegativeIndex(t *testing.T) {
	s := series.NewStringArraySeries(10)
	s.Set([]string{"a", "b"})

	mut := NewStringMutator()
	mut.SetElement(s, -1, "X")

	result := s.GetCurrent()
	if len(result) != 2 || result[0] != "a" || result[1] != "b" {
		t.Errorf("SetElement negative index: array modified, got %v", result)
	}
}

func TestStringMutator_Insert(t *testing.T) {
	s := series.NewStringArraySeries(10)
	s.Set([]string{"a", "c"})

	mut := NewStringMutator()
	mut.Insert(s, 1, "b")

	result := s.GetCurrent()
	if len(result) != 3 {
		t.Fatalf("Insert: expected length 3, got %d", len(result))
	}
	if result[0] != "a" || result[1] != "b" || result[2] != "c" {
		t.Errorf("Insert: expected [a b c], got %v", result)
	}
}

func TestStringMutator_Insert_AtEnd(t *testing.T) {
	s := series.NewStringArraySeries(10)
	s.Set([]string{"a", "b"})

	mut := NewStringMutator()
	mut.Insert(s, 2, "c")

	result := s.GetCurrent()
	if len(result) != 3 || result[2] != "c" {
		t.Errorf("Insert at end: expected [a b c], got %v", result)
	}
}

func TestStringMutator_Insert_OutOfBounds(t *testing.T) {
	s := series.NewStringArraySeries(10)
	s.Set([]string{"a", "b"})

	mut := NewStringMutator()
	mut.Insert(s, 5, "X")

	result := s.GetCurrent()
	if len(result) != 2 {
		t.Errorf("Insert out of bounds: array modified, got %v", result)
	}
}

func TestStringMutator_Remove(t *testing.T) {
	s := series.NewStringArraySeries(10)
	s.Set([]string{"a", "b", "c"})

	mut := NewStringMutator()
	removed := mut.Remove(s, 1)

	if removed != "b" {
		t.Errorf("Remove: expected removed value 'b', got '%s'", removed)
	}

	result := s.GetCurrent()
	if len(result) != 2 {
		t.Fatalf("Remove: expected length 2, got %d", len(result))
	}
	if result[0] != "a" || result[1] != "c" {
		t.Errorf("Remove: expected [a c], got %v", result)
	}
}

func TestStringMutator_Remove_OutOfBounds(t *testing.T) {
	s := series.NewStringArraySeries(10)
	s.Set([]string{"a", "b"})

	mut := NewStringMutator()
	removed := mut.Remove(s, 5)

	if removed != "" {
		t.Errorf("Remove out of bounds: expected empty string, got '%s'", removed)
	}

	result := s.GetCurrent()
	if len(result) != 2 {
		t.Errorf("Remove out of bounds: array modified, got %v", result)
	}
}

func TestStringMutator_Clear(t *testing.T) {
	s := series.NewStringArraySeries(10)
	s.Set([]string{"a", "b", "c"})

	mut := NewStringMutator()
	mut.Clear(s)

	result := s.GetCurrent()
	if len(result) != 0 {
		t.Errorf("Clear: expected empty array, got %v", result)
	}
}

func TestStringMutator_Fill(t *testing.T) {
	s := series.NewStringArraySeries(10)
	s.Set([]string{"a", "b", "c", "d"})

	mut := NewStringMutator()
	mut.Fill(s, "X", 1, 3)

	result := s.GetCurrent()
	if result[0] != "a" || result[1] != "X" || result[2] != "X" || result[3] != "d" {
		t.Errorf("Fill: expected [a X X d], got %v", result)
	}
}

func TestStringMutator_Fill_FullRange(t *testing.T) {
	s := series.NewStringArraySeries(10)
	s.Set([]string{"a", "b", "c"})

	mut := NewStringMutator()
	mut.Fill(s, "Z", 0, 3)

	result := s.GetCurrent()
	for i, v := range result {
		if v != "Z" {
			t.Errorf("Fill full range: expected 'Z' at index %d, got '%s'", i, v)
		}
	}
}

func TestStringMutator_Fill_InvalidRange(t *testing.T) {
	s := series.NewStringArraySeries(10)
	s.Set([]string{"a", "b", "c"})

	mut := NewStringMutator()
	mut.Fill(s, "X", 2, 1)

	result := s.GetCurrent()
	if result[0] != "a" || result[1] != "b" || result[2] != "c" {
		t.Errorf("Fill invalid range: array modified, got %v", result)
	}
}

func TestStringMutator_HistoricalImmutability(t *testing.T) {
	s := series.NewStringArraySeries(10)
	s.Set([]string{"a", "b"})
	s.Next()
	s.Set([]string{"c", "d"})

	mut := NewStringMutator()
	mut.Push(s, "e")

	historical := s.Get(1)
	if len(historical) != 2 || historical[0] != "a" || historical[1] != "b" {
		t.Errorf("Historical immutability: mutation leaked to history, got %v", historical)
	}

	current := s.GetCurrent()
	if len(current) != 3 || current[2] != "e" {
		t.Errorf("Historical immutability: current incorrect, got %v", current)
	}
}
