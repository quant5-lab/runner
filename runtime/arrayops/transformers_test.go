package arrayops

import (
	"testing"

	"github.com/quant5-lab/runner/runtime/series"
)

func TestTransformer_Copy_ValidArray(t *testing.T) {
	s := series.NewArraySeries(10)
	original := []float64{1.0, 2.0, 3.0}
	s.Set(original)

	tf := NewTransformer()
	copied := tf.Copy(s, 0)

	if len(copied) != len(original) {
		t.Fatalf("Copy: expected length %d, got %d", len(original), len(copied))
	}

	for i := range original {
		if copied[i] != original[i] {
			t.Errorf("Copy: index %d: expected %f, got %f", i, original[i], copied[i])
		}
	}

	copied[0] = 999.0
	if s.GetCurrent()[0] == 999.0 {
		t.Error("Copy: mutation of copy affected original (not a deep copy)")
	}
}

func TestTransformer_Copy_EmptyArray(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{})

	tf := NewTransformer()
	copied := tf.Copy(s, 0)

	if copied == nil {
		t.Fatal("Copy of empty: expected non-nil slice")
	}
	if len(copied) != 0 {
		t.Errorf("Copy of empty: expected length 0, got %d", len(copied))
	}
}

func TestTransformer_Copy_NilArray(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set(nil)

	tf := NewTransformer()
	copied := tf.Copy(s, 0)

	if copied != nil {
		t.Error("Copy of nil: expected nil, got non-nil")
	}
}

func TestTransformer_Copy_HistoricalBar(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{1.0, 2.0})
	s.Next()
	s.Set([]float64{10.0, 20.0})

	tf := NewTransformer()
	previousCopy := tf.Copy(s, 1)

	if len(previousCopy) != 2 || previousCopy[0] != 1.0 || previousCopy[1] != 2.0 {
		t.Errorf("Copy historical: expected [1, 2], got %v", previousCopy)
	}
}

func TestTransformer_Slice_ValidRange(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{10.0, 20.0, 30.0, 40.0, 50.0})

	tf := NewTransformer()
	sliced := tf.Slice(s, 0, 1, 4)

	expected := []float64{20.0, 30.0, 40.0}
	if len(sliced) != len(expected) {
		t.Fatalf("Slice: expected length %d, got %d", len(expected), len(sliced))
	}
	for i := range expected {
		if sliced[i] != expected[i] {
			t.Errorf("Slice: index %d: expected %f, got %f", i, expected[i], sliced[i])
		}
	}
}

func TestTransformer_Slice_FromStart(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{1.0, 2.0, 3.0, 4.0})

	tf := NewTransformer()
	sliced := tf.Slice(s, 0, 0, 2)

	if len(sliced) != 2 || sliced[0] != 1.0 || sliced[1] != 2.0 {
		t.Errorf("Slice from start: expected [1, 2], got %v", sliced)
	}
}

func TestTransformer_Slice_ToEnd(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{1.0, 2.0, 3.0, 4.0})

	tf := NewTransformer()
	sliced := tf.Slice(s, 0, 2, -1)

	if len(sliced) != 2 || sliced[0] != 3.0 || sliced[1] != 4.0 {
		t.Errorf("Slice to end: expected [3, 4], got %v", sliced)
	}
}

func TestTransformer_Slice_NegativeIndexNormalized(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{1.0, 2.0, 3.0})

	tf := NewTransformer()
	sliced := tf.Slice(s, 0, -10, 2)

	if len(sliced) != 2 || sliced[0] != 1.0 || sliced[1] != 2.0 {
		t.Errorf("Slice negative from: expected [1, 2], got %v", sliced)
	}
}

func TestTransformer_Slice_EmptyRange(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{1.0, 2.0, 3.0})

	tf := NewTransformer()
	sliced := tf.Slice(s, 0, 2, 1)

	if len(sliced) != 0 {
		t.Errorf("Slice empty range: expected empty, got %v", sliced)
	}
}

func TestTransformer_Slice_EmptyArray(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{})

	tf := NewTransformer()
	sliced := tf.Slice(s, 0, 0, 5)

	if sliced != nil {
		t.Errorf("Slice of empty: expected nil, got %v", sliced)
	}
}

func TestTransformer_Reverse_ValidArray(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{1.0, 2.0, 3.0, 4.0})

	tf := NewTransformer()
	tf.Reverse(s)

	result := s.GetCurrent()
	expected := []float64{4.0, 3.0, 2.0, 1.0}

	for i := range expected {
		if result[i] != expected[i] {
			t.Errorf("Reverse: index %d: expected %f, got %f", i, expected[i], result[i])
		}
	}
}

func TestTransformer_Reverse_SingleElement(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{42.0})

	tf := NewTransformer()
	tf.Reverse(s)

	result := s.GetCurrent()
	if len(result) != 1 || result[0] != 42.0 {
		t.Errorf("Reverse single: expected [42], got %v", result)
	}
}

func TestTransformer_Reverse_EmptyArray(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{})

	tf := NewTransformer()
	tf.Reverse(s)

	result := s.GetCurrent()
	if len(result) != 0 {
		t.Errorf("Reverse empty: expected empty, got %v", result)
	}
}

func TestTransformer_Sort_Ascending(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{5.0, 2.0, 8.0, 1.0, 9.0})

	tf := NewTransformer()
	tf.Sort(s, "order.ascending")

	result := s.GetCurrent()
	expected := []float64{1.0, 2.0, 5.0, 8.0, 9.0}

	for i := range expected {
		if result[i] != expected[i] {
			t.Errorf("Sort ascending: index %d: expected %f, got %f", i, expected[i], result[i])
		}
	}
}

func TestTransformer_Sort_Descending(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{5.0, 2.0, 8.0, 1.0, 9.0})

	tf := NewTransformer()
	tf.Sort(s, "order.descending")

	result := s.GetCurrent()
	expected := []float64{9.0, 8.0, 5.0, 2.0, 1.0}

	for i := range expected {
		if result[i] != expected[i] {
			t.Errorf("Sort descending: index %d: expected %f, got %f", i, expected[i], result[i])
		}
	}
}

func TestTransformer_Sort_AlreadySorted(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{1.0, 2.0, 3.0})

	tf := NewTransformer()
	tf.Sort(s, "order.ascending")

	result := s.GetCurrent()
	for i, v := range []float64{1.0, 2.0, 3.0} {
		if result[i] != v {
			t.Errorf("Sort already sorted: index %d: expected %f, got %f", i, v, result[i])
		}
	}
}

func TestTransformer_SortIndices_Ascending(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{30.0, 10.0, 20.0})

	tf := NewTransformer()
	indices := tf.SortIndices(s, 0, "order.ascending")

	expected := []float64{1.0, 2.0, 0.0}
	for i := range expected {
		if indices[i] != expected[i] {
			t.Errorf("SortIndices ascending: index %d: expected %f, got %f", i, expected[i], indices[i])
		}
	}
}

func TestTransformer_SortIndices_Descending(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{30.0, 10.0, 20.0})

	tf := NewTransformer()
	indices := tf.SortIndices(s, 0, "order.descending")

	expected := []float64{0.0, 2.0, 1.0}
	for i := range expected {
		if indices[i] != expected[i] {
			t.Errorf("SortIndices descending: index %d: expected %f, got %f", i, expected[i], indices[i])
		}
	}
}

func TestTransformer_SortIndices_EmptyArray(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{})

	tf := NewTransformer()
	indices := tf.SortIndices(s, 0, "order.ascending")

	if len(indices) != 0 {
		t.Errorf("SortIndices empty: expected empty, got %v", indices)
	}
}

func TestTransformer_Concat_BothValid(t *testing.T) {
	s1 := series.NewArraySeries(10)
	s1.Set([]float64{1.0, 2.0})

	s2 := series.NewArraySeries(10)
	s2.Set([]float64{3.0, 4.0})

	tf := NewTransformer()
	result := tf.Concat(s1, 0, s2, 0)

	expected := []float64{1.0, 2.0, 3.0, 4.0}
	for i := range expected {
		if result[i] != expected[i] {
			t.Errorf("Concat: index %d: expected %f, got %f", i, expected[i], result[i])
		}
	}
}

func TestTransformer_Concat_FirstNil(t *testing.T) {
	s1 := series.NewArraySeries(10)
	s1.Set(nil)

	s2 := series.NewArraySeries(10)
	s2.Set([]float64{3.0, 4.0})

	tf := NewTransformer()
	result := tf.Concat(s1, 0, s2, 0)

	if len(result) != 2 || result[0] != 3.0 || result[1] != 4.0 {
		t.Errorf("Concat first nil: expected [3, 4], got %v", result)
	}
}

func TestTransformer_Concat_SecondNil(t *testing.T) {
	s1 := series.NewArraySeries(10)
	s1.Set([]float64{1.0, 2.0})

	s2 := series.NewArraySeries(10)
	s2.Set(nil)

	tf := NewTransformer()
	result := tf.Concat(s1, 0, s2, 0)

	if len(result) != 2 || result[0] != 1.0 || result[1] != 2.0 {
		t.Errorf("Concat second nil: expected [1, 2], got %v", result)
	}
}

func TestTransformer_Concat_BothNil(t *testing.T) {
	s1 := series.NewArraySeries(10)
	s1.Set(nil)

	s2 := series.NewArraySeries(10)
	s2.Set(nil)

	tf := NewTransformer()
	result := tf.Concat(s1, 0, s2, 0)

	if result != nil {
		t.Errorf("Concat both nil: expected nil, got %v", result)
	}
}

func TestTransformer_Concat_BothEmpty(t *testing.T) {
	s1 := series.NewArraySeries(10)
	s1.Set([]float64{})

	s2 := series.NewArraySeries(10)
	s2.Set([]float64{})

	tf := NewTransformer()
	result := tf.Concat(s1, 0, s2, 0)

	if len(result) != 0 {
		t.Errorf("Concat both empty: expected empty, got %v", result)
	}
}
