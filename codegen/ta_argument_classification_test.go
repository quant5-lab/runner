package codegen

import "testing"

// allClassifications lists every TAArgumentClassification value.
// Updating this slice when adding a new classification is the only maintenance
// required to keep every property test below exhaustive.
var allClassifications = []struct {
	value TAArgumentClassification
	name  string
}{
	{TAArgSeriesRequired, "SeriesRequired"},
	{TAArgSeriesOptional, "SeriesOptional"},
	{TAArgScalarInt, "ScalarInt"},
	{TAArgScalarFloat, "ScalarFloat"},
	{TAArgScalarBool, "ScalarBool"},
	{TAArgImplicitOHLC, "ImplicitOHLC"},
}

// TestTAArgumentClassification_IsSeries verifies which classifications represent
// a series source (carries a stream of historical values).
func TestTAArgumentClassification_IsSeries(t *testing.T) {
	expected := map[TAArgumentClassification]bool{
		TAArgSeriesRequired: true,
		TAArgSeriesOptional: true,
		TAArgScalarInt:      false,
		TAArgScalarFloat:    false,
		TAArgScalarBool:     false,
		TAArgImplicitOHLC:   false,
	}

	for _, c := range allClassifications {
		t.Run(c.name, func(t *testing.T) {
			got := c.value.IsSeries()
			if got != expected[c.value] {
				t.Errorf("IsSeries() = %v, want %v", got, expected[c.value])
			}
		})
	}
}

// TestTAArgumentClassification_IsScalar verifies which classifications represent
// a compile-time scalar constant (int, float, or bool literal).
func TestTAArgumentClassification_IsScalar(t *testing.T) {
	expected := map[TAArgumentClassification]bool{
		TAArgSeriesRequired: false,
		TAArgSeriesOptional: false,
		TAArgScalarInt:      true,
		TAArgScalarFloat:    true,
		TAArgScalarBool:     true,
		TAArgImplicitOHLC:   false,
	}

	for _, c := range allClassifications {
		t.Run(c.name, func(t *testing.T) {
			got := c.value.IsScalar()
			if got != expected[c.value] {
				t.Errorf("IsScalar() = %v, want %v", got, expected[c.value])
			}
		})
	}
}

// TestTAArgumentClassification_RequiresHistoricalAccess verifies which
// classifications require the code generator to emit historical (offset) access.
func TestTAArgumentClassification_RequiresHistoricalAccess(t *testing.T) {
	expected := map[TAArgumentClassification]bool{
		TAArgSeriesRequired: true,
		TAArgSeriesOptional: true,
		TAArgScalarInt:      false,
		TAArgScalarFloat:    false,
		TAArgScalarBool:     false,
		TAArgImplicitOHLC:   true,
	}

	for _, c := range allClassifications {
		t.Run(c.name, func(t *testing.T) {
			got := c.value.RequiresHistoricalAccess()
			if got != expected[c.value] {
				t.Errorf("RequiresHistoricalAccess() = %v, want %v", got, expected[c.value])
			}
		})
	}
}

// TestTAArgumentClassification_SeriesScalarMutuallyExclusive verifies the
// invariant: IsSeries() and IsScalar() are never both true for the same
// classification.  TAArgImplicitOHLC is neither series nor scalar.
func TestTAArgumentClassification_SeriesScalarMutuallyExclusive(t *testing.T) {
	for _, c := range allClassifications {
		t.Run(c.name, func(t *testing.T) {
			if c.value.IsSeries() && c.value.IsScalar() {
				t.Errorf("%s: IsSeries and IsScalar are both true — classifications must be mutually exclusive", c.name)
			}
		})
	}
}
