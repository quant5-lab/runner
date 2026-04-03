package arrayops

import (
	"testing"

	"github.com/quant5-lab/runner/runtime/series"
)

func TestStringFormatters_Join(t *testing.T) {
	s := series.NewStringArraySeries(10)
	s.Set([]string{"apple", "banana", "cherry"})

	f := NewStringFormatters()
	joined := f.Join(s, 0, ", ")

	expected := "apple, banana, cherry"
	if joined != expected {
		t.Errorf("Join: expected '%s', got '%s'", expected, joined)
	}
}

func TestStringFormatters_Join_SingleElement(t *testing.T) {
	s := series.NewStringArraySeries(10)
	s.Set([]string{"only"})

	f := NewStringFormatters()
	joined := f.Join(s, 0, ", ")

	if joined != "only" {
		t.Errorf("Join single: expected 'only', got '%s'", joined)
	}
}

func TestStringFormatters_Join_Empty(t *testing.T) {
	s := series.NewStringArraySeries(10)
	s.Set([]string{})

	f := NewStringFormatters()
	joined := f.Join(s, 0, ", ")

	if joined != "" {
		t.Errorf("Join empty: expected '', got '%s'", joined)
	}
}

func TestStringFormatters_Join_EmptySeparator(t *testing.T) {
	s := series.NewStringArraySeries(10)
	s.Set([]string{"a", "b", "c"})

	f := NewStringFormatters()
	joined := f.Join(s, 0, "")

	if joined != "abc" {
		t.Errorf("Join no separator: expected 'abc', got '%s'", joined)
	}
}

func TestStringFormatters_Join_WithEmptyStrings(t *testing.T) {
	s := series.NewStringArraySeries(10)
	s.Set([]string{"a", "", "c"})

	f := NewStringFormatters()
	joined := f.Join(s, 0, "-")

	expected := "a--c"
	if joined != expected {
		t.Errorf("Join with empty strings: expected '%s', got '%s'", expected, joined)
	}
}

func TestStringFormatters_Join_HistoricalAccess(t *testing.T) {
	s := series.NewStringArraySeries(10)
	s.Set([]string{"old1", "old2"})
	s.Next()
	s.Set([]string{"new1", "new2"})

	f := NewStringFormatters()

	current := f.Join(s, 0, "|")
	if current != "new1|new2" {
		t.Errorf("Join current: expected 'new1|new2', got '%s'", current)
	}

	previous := f.Join(s, 1, "|")
	if previous != "old1|old2" {
		t.Errorf("Join historical: expected 'old1|old2', got '%s'", previous)
	}
}
