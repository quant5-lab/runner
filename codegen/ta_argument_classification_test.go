package codegen

import "testing"

func TestTAArgumentClassification_IsSeries(t *testing.T) {
	tests := []struct {
		classification TAArgumentClassification
		expected       bool
	}{
		{TAArgSeriesRequired, true},
		{TAArgSeriesOptional, true},
		{TAArgScalarInt, false},
		{TAArgScalarFloat, false},
		{TAArgImplicitOHLC, false},
	}

	for _, tt := range tests {
		result := tt.classification.IsSeries()
		if result != tt.expected {
			t.Errorf("IsSeries(%v) = %v, want %v", tt.classification, result, tt.expected)
		}
	}
}

func TestTAArgumentClassification_IsScalar(t *testing.T) {
	tests := []struct {
		classification TAArgumentClassification
		expected       bool
	}{
		{TAArgSeriesRequired, false},
		{TAArgSeriesOptional, false},
		{TAArgScalarInt, true},
		{TAArgScalarFloat, true},
		{TAArgImplicitOHLC, false},
	}

	for _, tt := range tests {
		result := tt.classification.IsScalar()
		if result != tt.expected {
			t.Errorf("IsScalar(%v) = %v, want %v", tt.classification, result, tt.expected)
		}
	}
}

func TestTAArgumentClassification_RequiresHistoricalAccess(t *testing.T) {
	tests := []struct {
		classification TAArgumentClassification
		expected       bool
	}{
		{TAArgSeriesRequired, true},
		{TAArgSeriesOptional, true},
		{TAArgScalarInt, false},
		{TAArgScalarFloat, false},
		{TAArgImplicitOHLC, true},
	}

	for _, tt := range tests {
		result := tt.classification.RequiresHistoricalAccess()
		if result != tt.expected {
			t.Errorf("RequiresHistoricalAccess(%v) = %v, want %v", tt.classification, result, tt.expected)
		}
	}
}
