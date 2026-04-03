package arrayops

import (
	"testing"

	"github.com/quant5-lab/runner/runtime/series"
)

func TestPredicates_Every_AllTrue(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{1.0, 2.0, 3.0, 4.0})

	pred := NewPredicates()
	result := pred.Every(s, 0)

	if !result {
		t.Error("Every: expected true for all non-zero values")
	}
}

func TestPredicates_Every_OneFalse(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{1.0, 0.0, 3.0, 4.0})

	pred := NewPredicates()
	result := pred.Every(s, 0)

	if result {
		t.Error("Every: expected false when one element is zero")
	}
}

func TestPredicates_Every_AllFalse(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{0.0, 0.0, 0.0})

	pred := NewPredicates()
	result := pred.Every(s, 0)

	if result {
		t.Error("Every: expected false for all zero values")
	}
}

func TestPredicates_Every_EmptyArray(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{})

	pred := NewPredicates()
	result := pred.Every(s, 0)

	if !result {
		t.Error("Every: expected true for empty array (vacuous truth)")
	}
}

func TestPredicates_Every_NilArray(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set(nil)

	pred := NewPredicates()
	result := pred.Every(s, 0)

	if !result {
		t.Error("Every: expected true for nil array (vacuous truth)")
	}
}

func TestPredicates_Some_HasTrue(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{0.0, 0.0, 1.0, 0.0})

	pred := NewPredicates()
	result := pred.Some(s, 0)

	if !result {
		t.Error("Some: expected true when at least one non-zero value")
	}
}

func TestPredicates_Some_AllFalse(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{0.0, 0.0, 0.0})

	pred := NewPredicates()
	result := pred.Some(s, 0)

	if result {
		t.Error("Some: expected false for all zero values")
	}
}

func TestPredicates_Some_AllTrue(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{1.0, 2.0, 3.0})

	pred := NewPredicates()
	result := pred.Some(s, 0)

	if !result {
		t.Error("Some: expected true when all values are non-zero")
	}
}

func TestPredicates_Some_EmptyArray(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{})

	pred := NewPredicates()
	result := pred.Some(s, 0)

	if result {
		t.Error("Some: expected false for empty array")
	}
}

func TestPredicates_Some_NilArray(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set(nil)

	pred := NewPredicates()
	result := pred.Some(s, 0)

	if result {
		t.Error("Some: expected false for nil array")
	}
}

func TestPredicates_Every_HistoricalAccess(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{0.0, 1.0})
	s.Next()
	s.Set([]float64{1.0, 2.0, 3.0})

	pred := NewPredicates()

	current := pred.Every(s, 0)
	if !current {
		t.Error("Every current bar: expected true")
	}

	previous := pred.Every(s, 1)
	if previous {
		t.Error("Every previous bar: expected false")
	}
}

func TestPredicates_Some_HistoricalAccess(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{0.0, 0.0})
	s.Next()
	s.Set([]float64{1.0, 2.0})

	pred := NewPredicates()

	current := pred.Some(s, 0)
	if !current {
		t.Error("Some current bar: expected true")
	}

	previous := pred.Some(s, 1)
	if previous {
		t.Error("Some previous bar: expected false")
	}
}

func TestPredicates_Every_SingleTrue(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{1.0})

	pred := NewPredicates()
	result := pred.Every(s, 0)

	if !result {
		t.Error("Every single true: expected true")
	}
}

func TestPredicates_Every_SingleFalse(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{0.0})

	pred := NewPredicates()
	result := pred.Every(s, 0)

	if result {
		t.Error("Every single false: expected false")
	}
}

func TestPredicates_Some_SingleTrue(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{1.0})

	pred := NewPredicates()
	result := pred.Some(s, 0)

	if !result {
		t.Error("Some single true: expected true")
	}
}

func TestPredicates_Some_SingleFalse(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{0.0})

	pred := NewPredicates()
	result := pred.Some(s, 0)

	if result {
		t.Error("Some single false: expected false")
	}
}

func TestPredicates_Every_MixedValues(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{-1.0, 0.5, 2.0, 100.0})

	pred := NewPredicates()
	result := pred.Every(s, 0)

	if !result {
		t.Error("Every mixed non-zero: expected true")
	}
}

func TestPredicates_Every_FirstElementZero(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{0.0, 1.0, 2.0})

	pred := NewPredicates()
	result := pred.Every(s, 0)

	if result {
		t.Error("Every first zero: expected false")
	}
}

func TestPredicates_Every_LastElementZero(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{1.0, 2.0, 0.0})

	pred := NewPredicates()
	result := pred.Every(s, 0)

	if result {
		t.Error("Every last zero: expected false")
	}
}

func TestPredicates_Some_FirstElementTrue(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{1.0, 0.0, 0.0})

	pred := NewPredicates()
	result := pred.Some(s, 0)

	if !result {
		t.Error("Some first true: expected true")
	}
}

func TestPredicates_Some_LastElementTrue(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{0.0, 0.0, 1.0})

	pred := NewPredicates()
	result := pred.Some(s, 0)

	if !result {
		t.Error("Some last true: expected true")
	}
}

func TestPredicates_Some_NegativeValues(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{0.0, 0.0, -5.0})

	pred := NewPredicates()
	result := pred.Some(s, 0)

	if !result {
		t.Error("Some negative: expected true (non-zero)")
	}
}

func TestPredicates_Every_NegativeValues(t *testing.T) {
	s := series.NewArraySeries(10)
	s.Set([]float64{-1.0, -2.0, -3.0})

	pred := NewPredicates()
	result := pred.Every(s, 0)

	if !result {
		t.Error("Every negative: expected true (all non-zero)")
	}
}
