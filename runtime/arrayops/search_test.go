package arrayops

import (
	"testing"

	"github.com/quant5-lab/runner/runtime/series"
)

func TestSearch_BinarySearch_Found(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{1.0, 3.0, 5.0, 7.0, 9.0})

	search := NewSearch()
	result := search.BinarySearch(s, 0, 5.0)

	if result != 2 {
		t.Errorf("BinarySearch: expected 2, got %d", result)
	}
}

func TestSearch_BinarySearch_NotFound(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{1.0, 3.0, 5.0, 7.0, 9.0})

	search := NewSearch()
	result := search.BinarySearch(s, 0, 4.0)

	if result != -1 {
		t.Errorf("BinarySearch not found: expected -1, got %d", result)
	}
}

func TestSearch_BinarySearch_EmptyArray(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{})

	search := NewSearch()
	result := search.BinarySearch(s, 0, 5.0)

	if result != -1 {
		t.Errorf("BinarySearch empty: expected -1, got %d", result)
	}
}

func TestSearch_BinarySearch_SingleElement(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{5.0})

	search := NewSearch()
	result := search.BinarySearch(s, 0, 5.0)

	if result != 0 {
		t.Errorf("BinarySearch single element: expected 0, got %d", result)
	}
}

func TestSearch_BinarySearchLeftmost_MultipleDuplicates(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{1.0, 3.0, 5.0, 5.0, 5.0, 7.0, 9.0})

	search := NewSearch()
	result := search.BinarySearchLeftmost(s, 0, 5.0)

	if result != 2 {
		t.Errorf("BinarySearchLeftmost: expected 2 (leftmost), got %d", result)
	}
}

func TestSearch_BinarySearchLeftmost_NotFound(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{1.0, 3.0, 7.0, 9.0})

	search := NewSearch()
	result := search.BinarySearchLeftmost(s, 0, 5.0)

	if result != -1 {
		t.Errorf("BinarySearchLeftmost not found: expected -1, got %d", result)
	}
}

func TestSearch_BinarySearchRightmost_MultipleDuplicates(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{1.0, 3.0, 5.0, 5.0, 5.0, 7.0, 9.0})

	search := NewSearch()
	result := search.BinarySearchRightmost(s, 0, 5.0)

	if result != 4 {
		t.Errorf("BinarySearchRightmost: expected 4 (rightmost), got %d", result)
	}
}

func TestSearch_BinarySearchRightmost_NotFound(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{1.0, 3.0, 7.0, 9.0})

	search := NewSearch()
	result := search.BinarySearchRightmost(s, 0, 5.0)

	if result != -1 {
		t.Errorf("BinarySearchRightmost not found: expected -1, got %d", result)
	}
}

func TestSearch_BinarySearch_HistoricalAccess(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{1.0, 2.0, 3.0})
	s.Next()
	s.Set([]float64{10.0, 20.0, 30.0, 40.0})

	search := NewSearch()

	current := search.BinarySearch(s, 0, 30.0)
	if current != 2 {
		t.Errorf("BinarySearch current bar: expected 2, got %d", current)
	}

	previous := search.BinarySearch(s, 1, 2.0)
	if previous != 1 {
		t.Errorf("BinarySearch previous bar: expected 1, got %d", previous)
	}
}

func TestSearch_BinarySearchLeftmost_SingleOccurrence(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{1.0, 3.0, 5.0, 7.0, 9.0})

	search := NewSearch()
	result := search.BinarySearchLeftmost(s, 0, 5.0)

	if result != 2 {
		t.Errorf("BinarySearchLeftmost single: expected 2, got %d", result)
	}
}

func TestSearch_BinarySearchRightmost_SingleOccurrence(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{1.0, 3.0, 5.0, 7.0, 9.0})

	search := NewSearch()
	result := search.BinarySearchRightmost(s, 0, 5.0)

	if result != 2 {
		t.Errorf("BinarySearchRightmost single: expected 2, got %d", result)
	}
}

func TestSearch_BinarySearch_NilArray(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set(nil)

	search := NewSearch()
	result := search.BinarySearch(s, 0, 5.0)

	if result != -1 {
		t.Errorf("BinarySearch nil: expected -1, got %d", result)
	}
}

func TestSearch_BinarySearch_BoundaryValues(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{1.0, 3.0, 5.0, 7.0, 9.0})

	search := NewSearch()

	first := search.BinarySearch(s, 0, 1.0)
	if first != 0 {
		t.Errorf("BinarySearch first element: expected 0, got %d", first)
	}

	last := search.BinarySearch(s, 0, 9.0)
	if last != 4 {
		t.Errorf("BinarySearch last element: expected 4, got %d", last)
	}
}

func TestSearch_BinarySearch_TwoElements(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{3.0, 7.0})

	search := NewSearch()

	found := search.BinarySearch(s, 0, 7.0)
	if found != 1 {
		t.Errorf("BinarySearch two elements: expected 1, got %d", found)
	}

	notFound := search.BinarySearch(s, 0, 5.0)
	if notFound != -1 {
		t.Errorf("BinarySearch two elements not found: expected -1, got %d", notFound)
	}
}

func TestSearch_BinarySearchLeftmost_AllDuplicates(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{5.0, 5.0, 5.0, 5.0})

	search := NewSearch()
	result := search.BinarySearchLeftmost(s, 0, 5.0)

	if result != 0 {
		t.Errorf("BinarySearchLeftmost all duplicates: expected 0, got %d", result)
	}
}

func TestSearch_BinarySearchRightmost_AllDuplicates(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{5.0, 5.0, 5.0, 5.0})

	search := NewSearch()
	result := search.BinarySearchRightmost(s, 0, 5.0)

	if result != 3 {
		t.Errorf("BinarySearchRightmost all duplicates: expected 3, got %d", result)
	}
}

func TestSearch_BinarySearchLeftmost_DuplicatesAtStart(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{1.0, 1.0, 1.0, 3.0, 5.0})

	search := NewSearch()
	result := search.BinarySearchLeftmost(s, 0, 1.0)

	if result != 0 {
		t.Errorf("BinarySearchLeftmost duplicates at start: expected 0, got %d", result)
	}
}

func TestSearch_BinarySearchRightmost_DuplicatesAtEnd(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{1.0, 3.0, 9.0, 9.0, 9.0})

	search := NewSearch()
	result := search.BinarySearchRightmost(s, 0, 9.0)

	if result != 4 {
		t.Errorf("BinarySearchRightmost duplicates at end: expected 4, got %d", result)
	}
}

func TestSearch_BinarySearchLeftmost_EmptyArray(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{})

	search := NewSearch()
	result := search.BinarySearchLeftmost(s, 0, 5.0)

	if result != -1 {
		t.Errorf("BinarySearchLeftmost empty: expected -1, got %d", result)
	}
}

func TestSearch_BinarySearchRightmost_EmptyArray(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{})

	search := NewSearch()
	result := search.BinarySearchRightmost(s, 0, 5.0)

	if result != -1 {
		t.Errorf("BinarySearchRightmost empty: expected -1, got %d", result)
	}
}

func TestSearch_BinarySearchLeftmost_NilArray(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set(nil)

	search := NewSearch()
	result := search.BinarySearchLeftmost(s, 0, 5.0)

	if result != -1 {
		t.Errorf("BinarySearchLeftmost nil: expected -1, got %d", result)
	}
}

func TestSearch_BinarySearchRightmost_NilArray(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set(nil)

	search := NewSearch()
	result := search.BinarySearchRightmost(s, 0, 5.0)

	if result != -1 {
		t.Errorf("BinarySearchRightmost nil: expected -1, got %d", result)
	}
}

func TestSearch_BinarySearchLeftmost_HistoricalAccess(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{1.0, 1.0, 3.0})
	s.Next()
	s.Set([]float64{5.0, 5.0, 5.0, 7.0})

	search := NewSearch()

	current := search.BinarySearchLeftmost(s, 0, 5.0)
	if current != 0 {
		t.Errorf("BinarySearchLeftmost current bar: expected 0, got %d", current)
	}

	previous := search.BinarySearchLeftmost(s, 1, 1.0)
	if previous != 0 {
		t.Errorf("BinarySearchLeftmost previous bar: expected 0, got %d", previous)
	}
}

func TestSearch_BinarySearchRightmost_HistoricalAccess(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{1.0, 1.0, 3.0})
	s.Next()
	s.Set([]float64{5.0, 5.0, 5.0, 7.0})

	search := NewSearch()

	current := search.BinarySearchRightmost(s, 0, 5.0)
	if current != 2 {
		t.Errorf("BinarySearchRightmost current bar: expected 2, got %d", current)
	}

	previous := search.BinarySearchRightmost(s, 1, 1.0)
	if previous != 1 {
		t.Errorf("BinarySearchRightmost previous bar: expected 1, got %d", previous)
	}
}
