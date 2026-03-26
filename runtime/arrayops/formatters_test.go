package arrayops

import (
	"testing"

	"github.com/quant5-lab/runner/runtime/series"
)

func TestFormatters_Join_CommaSeparator(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{1.0, 2.0, 3.0})

	fmt := NewFormatters()
	result := fmt.Join(s, 0, ", ")

	expected := "1, 2, 3"
	if result != expected {
		t.Errorf("Join: expected %q, got %q", expected, result)
	}
}

func TestFormatters_Join_EmptyArray(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{})

	fmt := NewFormatters()
	result := fmt.Join(s, 0, ", ")

	if result != "" {
		t.Errorf("Join empty: expected empty string, got %q", result)
	}
}

func TestFormatters_Join_NilArray(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set(nil)

	fmt := NewFormatters()
	result := fmt.Join(s, 0, ", ")

	if result != "" {
		t.Errorf("Join nil: expected empty string, got %q", result)
	}
}

func TestFormatters_Join_SingleElement(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{42.0})

	fmt := NewFormatters()
	result := fmt.Join(s, 0, ", ")

	expected := "42"
	if result != expected {
		t.Errorf("Join single element: expected %q, got %q", expected, result)
	}
}

func TestFormatters_Join_CustomSeparator(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{10.0, 20.0, 30.0})

	fmt := NewFormatters()
	result := fmt.Join(s, 0, " | ")

	expected := "10 | 20 | 30"
	if result != expected {
		t.Errorf("Join custom separator: expected %q, got %q", expected, result)
	}
}

func TestFormatters_Join_EmptySeparator(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{1.0, 2.0, 3.0})

	fmt := NewFormatters()
	result := fmt.Join(s, 0, "")

	expected := "123"
	if result != expected {
		t.Errorf("Join no separator: expected %q, got %q", expected, result)
	}
}

func TestFormatters_Join_FloatingPointValues(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{1.5, 2.75, 3.125})

	fmt := NewFormatters()
	result := fmt.Join(s, 0, ", ")

	expected := "1.5, 2.75, 3.125"
	if result != expected {
		t.Errorf("Join floats: expected %q, got %q", expected, result)
	}
}

func TestFormatters_Join_HistoricalAccess(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{1.0, 2.0})
	s.Next()
	s.Set([]float64{10.0, 20.0, 30.0})

	fmt := NewFormatters()

	current := fmt.Join(s, 0, ", ")
	expectedCurrent := "10, 20, 30"
	if current != expectedCurrent {
		t.Errorf("Join current bar: expected %q, got %q", expectedCurrent, current)
	}

	previous := fmt.Join(s, 1, ", ")
	expectedPrevious := "1, 2"
	if previous != expectedPrevious {
		t.Errorf("Join previous bar: expected %q, got %q", expectedPrevious, previous)
	}
}

func TestFormatters_Join_NegativeValues(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{-1.0, -2.5, -3.75})

	fmt := NewFormatters()
	result := fmt.Join(s, 0, ", ")

	expected := "-1, -2.5, -3.75"
	if result != expected {
		t.Errorf("Join negatives: expected %q, got %q", expected, result)
	}
}

func TestFormatters_Join_ZeroValues(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{0.0, 0.0, 0.0})

	fmt := NewFormatters()
	result := fmt.Join(s, 0, ", ")

	expected := "0, 0, 0"
	if result != expected {
		t.Errorf("Join zeros: expected %q, got %q", expected, result)
	}
}

func TestFormatters_Join_MixedPositiveNegative(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{-5.0, 0.0, 5.0})

	fmt := NewFormatters()
	result := fmt.Join(s, 0, " ")

	expected := "-5 0 5"
	if result != expected {
		t.Errorf("Join mixed: expected %q, got %q", expected, result)
	}
}

func TestFormatters_Join_LargeNumbers(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{1000000.0, 2000000.0})

	fmt := NewFormatters()
	result := fmt.Join(s, 0, ", ")

	expected := "1000000, 2000000"
	if result != expected {
		t.Errorf("Join large: expected %q, got %q", expected, result)
	}
}

func TestFormatters_Join_SmallDecimals(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{0.001, 0.002, 0.003})

	fmt := NewFormatters()
	result := fmt.Join(s, 0, ", ")

	expected := "0.001, 0.002, 0.003"
	if result != expected {
		t.Errorf("Join decimals: expected %q, got %q", expected, result)
	}
}

func TestFormatters_Join_NewlineSeparator(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{1.0, 2.0, 3.0})

	fmt := NewFormatters()
	result := fmt.Join(s, 0, "\n")

	expected := "1\n2\n3"
	if result != expected {
		t.Errorf("Join newline: expected %q, got %q", expected, result)
	}
}

func TestFormatters_Join_LongSeparator(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{1.0, 2.0})

	fmt := NewFormatters()
	result := fmt.Join(s, 0, " -> ")

	expected := "1 -> 2"
	if result != expected {
		t.Errorf("Join long separator: expected %q, got %q", expected, result)
	}
}
