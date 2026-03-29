package arrayops

import (
	"testing"

	"github.com/quant5-lab/runner/runtime/series"
)

func TestStringTransformers_Reverse(t *testing.T) {
	s := series.NewStringArraySeries(10)
	s.Set([]string{"a", "b", "c"})

	tr := NewStringTransformers()
	tr.Reverse(s)

	result := s.GetCurrent()
	if len(result) != 3 || result[0] != "c" || result[1] != "b" || result[2] != "a" {
		t.Errorf("Reverse: expected [c b a], got %v", result)
	}
}

func TestStringTransformers_Reverse_SingleElement(t *testing.T) {
	s := series.NewStringArraySeries(10)
	s.Set([]string{"only"})

	tr := NewStringTransformers()
	tr.Reverse(s)

	result := s.GetCurrent()
	if len(result) != 1 || result[0] != "only" {
		t.Errorf("Reverse single: expected [only], got %v", result)
	}
}

func TestStringTransformers_Reverse_EmptyArray(t *testing.T) {
	s := series.NewStringArraySeries(10)
	s.Set([]string{})

	tr := NewStringTransformers()
	tr.Reverse(s)

	result := s.GetCurrent()
	if len(result) != 0 {
		t.Errorf("Reverse empty: expected [], got %v", result)
	}
}

func TestStringTransformers_Concat(t *testing.T) {
	s1 := series.NewStringArraySeries(10)
	s1.Set([]string{"a", "b"})
	s2 := series.NewStringArraySeries(10)
	s2.Set([]string{"c", "d"})

	tr := NewStringTransformers()
	result := tr.Concat(s1, 0, s2, 0)

	if len(result) != 4 || result[0] != "a" || result[3] != "d" {
		t.Errorf("Concat: expected [a b c d], got %v", result)
	}
}

func TestStringTransformers_Concat_EmptyArrays(t *testing.T) {
	s1 := series.NewStringArraySeries(10)
	s1.Set([]string{})
	s2 := series.NewStringArraySeries(10)
	s2.Set([]string{})

	tr := NewStringTransformers()
	result := tr.Concat(s1, 0, s2, 0)

	if len(result) != 0 {
		t.Errorf("Concat empty: expected [], got %v", result)
	}
}

func TestStringTransformers_Slice(t *testing.T) {
	s := series.NewStringArraySeries(10)
	s.Set([]string{"a", "b", "c", "d"})

	tr := NewStringTransformers()
	result := tr.Slice(s, 0, 1, 3)

	if len(result) != 2 || result[0] != "b" || result[1] != "c" {
		t.Errorf("Slice: expected [b c], got %v", result)
	}
}

func TestStringTransformers_Slice_EmptyRange(t *testing.T) {
	s := series.NewStringArraySeries(10)
	s.Set([]string{"a", "b", "c"})

	tr := NewStringTransformers()
	result := tr.Slice(s, 0, 2, 2)

	if len(result) != 0 {
		t.Errorf("Slice empty range: expected [], got %v", result)
	}
}

func TestStringTransformers_Slice_NegativeBounds(t *testing.T) {
	s := series.NewStringArraySeries(10)
	s.Set([]string{"a", "b", "c"})

	tr := NewStringTransformers()
	result := tr.Slice(s, 0, -1, 5)

	if len(result) != 3 {
		t.Errorf("Slice negative bounds: expected clamp to [0:3], got %v", result)
	}
}

func TestStringTransformers_Copy(t *testing.T) {
	s := series.NewStringArraySeries(10)
	s.Set([]string{"a", "b"})

	tr := NewStringTransformers()
	copied := tr.Copy(s, 0)

	if len(copied) != 2 || copied[0] != "a" || copied[1] != "b" {
		t.Errorf("Copy: expected [a b], got %v", copied)
	}

	copied[0] = "X"
	original := s.GetCurrent()
	if original[0] != "a" {
		t.Error("Copy: mutation leaked to original")
	}
}

func TestStringTransformers_Copy_Nil(t *testing.T) {
	s := series.NewStringArraySeries(10)
	s.Set(nil)

	tr := NewStringTransformers()
	copied := tr.Copy(s, 0)

	if copied != nil {
		t.Errorf("Copy nil: expected nil, got %v", copied)
	}
}
