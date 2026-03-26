package arrayops

import (
	"testing"

	"github.com/quant5-lab/runner/runtime/series"
)

func TestMutator_Push(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{1.0, 2.0})

	mut := NewMutator()
	mut.Push(s, 3.0)

	result := s.GetCurrent()
	if len(result) != 3 {
		t.Fatalf("Push: expected length 3, got %d", len(result))
	}
	if result[2] != 3.0 {
		t.Errorf("Push: expected result[2] = 3.0, got %f", result[2])
	}
}

func TestMutator_Push_EmptyArray(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set(nil)

	mut := NewMutator()
	mut.Push(s, 5.0)

	result := s.GetCurrent()
	if len(result) != 1 {
		t.Fatalf("Push to nil: expected length 1, got %d", len(result))
	}
	if result[0] != 5.0 {
		t.Errorf("Push to nil: expected result[0] = 5.0, got %f", result[0])
	}
}

func TestMutator_Pop(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{1.0, 2.0, 3.0})

	mut := NewMutator()
	popped := mut.Pop(s)

	if popped != 3.0 {
		t.Errorf("Pop: expected popped value 3.0, got %f", popped)
	}

	result := s.GetCurrent()
	if len(result) != 2 {
		t.Fatalf("Pop: expected length 2, got %d", len(result))
	}
}

func TestMutator_Shift(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{10.0, 20.0, 30.0})

	mut := NewMutator()
	shifted := mut.Shift(s)

	if shifted != 10.0 {
		t.Errorf("Shift: expected shifted value 10.0, got %f", shifted)
	}

	result := s.GetCurrent()
	if len(result) != 2 {
		t.Fatalf("Shift: expected length 2, got %d", len(result))
	}
	if result[0] != 20.0 {
		t.Errorf("Shift: expected result[0] = 20.0, got %f", result[0])
	}
}

func TestMutator_Unshift(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{2.0, 3.0})

	mut := NewMutator()
	mut.Unshift(s, 1.0)

	result := s.GetCurrent()
	if len(result) != 3 {
		t.Fatalf("Unshift: expected length 3, got %d", len(result))
	}
	if result[0] != 1.0 {
		t.Errorf("Unshift: expected result[0] = 1.0, got %f", result[0])
	}
	if result[1] != 2.0 {
		t.Errorf("Unshift: expected result[1] = 2.0, got %f", result[1])
	}
}

func TestMutator_SetElement(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{10.0, 20.0, 30.0})

	mut := NewMutator()
	mut.SetElement(s, 1, 99.0)

	result := s.GetCurrent()
	if result[1] != 99.0 {
		t.Errorf("SetElement: expected result[1] = 99.0, got %f", result[1])
	}
	if result[0] != 10.0 || result[2] != 30.0 {
		t.Errorf("SetElement: expected other elements unchanged")
	}
}

func TestMutator_Insert(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{1.0, 3.0})

	mut := NewMutator()
	mut.Insert(s, 1, 2.0)

	result := s.GetCurrent()
	if len(result) != 3 {
		t.Fatalf("Insert: expected length 3, got %d", len(result))
	}
	if result[0] != 1.0 || result[1] != 2.0 || result[2] != 3.0 {
		t.Errorf("Insert: expected [1, 2, 3], got %v", result)
	}
}

func TestMutator_Remove(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{10.0, 20.0, 30.0})

	mut := NewMutator()
	removed := mut.Remove(s, 1)

	if removed != 20.0 {
		t.Errorf("Remove: expected removed value 20.0, got %f", removed)
	}

	result := s.GetCurrent()
	if len(result) != 2 {
		t.Fatalf("Remove: expected length 2, got %d", len(result))
	}
	if result[0] != 10.0 || result[1] != 30.0 {
		t.Errorf("Remove: expected [10, 30], got %v", result)
	}
}

func TestMutator_Clear(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{1.0, 2.0, 3.0})

	mut := NewMutator()
	mut.Clear(s)

	result := s.GetCurrent()
	if len(result) != 0 {
		t.Errorf("Clear: expected empty array, got length %d", len(result))
	}
}

func TestMutator_Fill(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{1.0, 2.0, 3.0, 4.0, 5.0})

	mut := NewMutator()
	mut.Fill(s, 99.0, 1, 3)

	result := s.GetCurrent()
	expected := []float64{1.0, 99.0, 99.0, 4.0, 5.0}
	for i := range expected {
		if result[i] != expected[i] {
			t.Errorf("Fill: expected result[%d] = %f, got %f", i, expected[i], result[i])
		}
	}
}

func TestMutator_SetElement_OutOfBounds(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{1.0, 2.0, 3.0})

	mut := NewMutator()
	mut.SetElement(s, 10, 99.0)

	result := s.GetCurrent()
	if len(result) != 3 {
		t.Error("SetElement out of bounds: should not modify array")
	}
}

func TestMutator_SetElement_NegativeIndex(t *testing.T) {
	s := series.NewArraySeries(10)
	original := []float64{1.0, 2.0, 3.0}
	s.Set(original)

	mut := NewMutator()
	mut.SetElement(s, -1, 99.0)

	result := s.GetCurrent()
	for i := range original {
		if result[i] != original[i] {
			t.Error("SetElement negative index: should not modify array")
		}
	}
}

func TestMutator_Insert_AtEnd(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{1.0, 2.0})

	mut := NewMutator()
	mut.Insert(s, 2, 3.0)

	result := s.GetCurrent()
	if len(result) != 3 || result[2] != 3.0 {
		t.Errorf("Insert at end: expected [1, 2, 3], got %v", result)
	}
}

func TestMutator_Insert_OutOfBounds(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{1.0, 2.0})

	mut := NewMutator()
	mut.Insert(s, 10, 99.0)

	result := s.GetCurrent()
	if len(result) != 2 {
		t.Error("Insert out of bounds: should not modify array")
	}
}

func TestMutator_Remove_OutOfBounds(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{1.0, 2.0})

	mut := NewMutator()
	removed := mut.Remove(s, 5)

	if removed != 0.0 {
		t.Errorf("Remove out of bounds: expected 0.0, got %f", removed)
	}
	result := s.GetCurrent()
	if len(result) != 2 {
		t.Error("Remove out of bounds: should not modify array")
	}
}

func TestMutator_Pop_EmptyArray(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{})

	mut := NewMutator()
	popped := mut.Pop(s)

	if popped != 0.0 {
		t.Errorf("Pop empty: expected 0.0, got %f", popped)
	}
	result := s.GetCurrent()
	if len(result) != 0 {
		t.Error("Pop empty: array should remain nil or empty")
	}
}

func TestMutator_Shift_EmptyArray(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{})

	mut := NewMutator()
	shifted := mut.Shift(s)

	if shifted != 0.0 {
		t.Errorf("Shift empty: expected 0.0, got %f", shifted)
	}
	result := s.GetCurrent()
	if len(result) != 0 {
		t.Error("Shift empty: array should remain nil or empty")
	}
}

func TestMutator_Fill_FullRange(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{1.0, 2.0, 3.0, 4.0})

	mut := NewMutator()
	mut.Fill(s, 0.0, 0, -1)

	result := s.GetCurrent()
	for i := range result {
		if result[i] != 0.0 {
			t.Errorf("Fill full: expected all 0.0, got result[%d] = %f", i, result[i])
		}
	}
}

func TestMutator_Fill_InvalidRange(t *testing.T) {
	s := series.NewArraySeries(10)
	original := []float64{1.0, 2.0, 3.0}
	s.Set(original)

	mut := NewMutator()
	mut.Fill(s, 99.0, 3, 1)

	result := s.GetCurrent()
	for i := range original {
		if result[i] != original[i] {
			t.Error("Fill invalid range: should not modify array")
		}
	}
}

func TestMutator_HistoricalImmutability(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{1.0, 2.0})
	s.Next()
	s.Set([]float64{10.0, 20.0})

	mut := NewMutator()
	mut.Push(s, 30.0)

	current := s.GetCurrent()
	if len(current) != 3 || current[2] != 30.0 {
		t.Fatal("Current bar mutation failed")
	}

	previous := s.Get(1)
	if len(previous) != 2 || previous[0] != 1.0 || previous[1] != 2.0 {
		t.Error("Historical immutability violated: previous bar was modified")
	}
}
