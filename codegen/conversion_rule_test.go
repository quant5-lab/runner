package codegen

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestSkipComparisonRule_ShouldConvert(t *testing.T) {
	comparisonMatcher := NewComparisonPattern()
	rule := NewSkipComparisonRule(comparisonMatcher)

	tests := []struct {
		name     string
		code     string
		expected bool
	}{
		{"skip when has greater than", "price > 100", false},
		{"skip when has less than", "a < b", false},
		{"skip when has equality", "price == 100", false},
		{"skip when has not equal", "x != y", false},
		{"skip when has greater equal", "val >= threshold", false},
		{"skip when has less equal", "val <= max", false},
		{"convert when no comparison", "priceSeries.GetCurrent()", true},
		{"convert when arithmetic only", "price + 100", true},
		{"convert when empty", "", true},
		{"skip when complex comparison", "(a > b) && (c < d)", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if result := rule.ShouldConvert(nil, tt.code); result != tt.expected {
				t.Errorf("code=%q: expected %v, got %v", tt.code, tt.expected, result)
			}
		})
	}
}

func TestConvertSeriesAccessRule_ShouldConvert(t *testing.T) {
	seriesMatcher := NewSeriesAccessPattern()
	rule := NewConvertSeriesAccessRule(seriesMatcher)

	tests := []struct {
		name     string
		code     string
		expected bool
	}{
		{"convert Series GetCurrent", "priceSeries.GetCurrent()", true},
		{"convert nested Series", "ta.sma(closeSeries.GetCurrent(), 20)", true},
		{"skip non-Series identifier", "price", false},
		{"skip literal", "100", false},
		{"skip empty", "", false},
		{"convert multiple Series", "aSeries.GetCurrent() + bSeries.GetCurrent()", true},
		{"convert historical access Series.Get(N)", "priceSeries.Get(1)", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if result := rule.ShouldConvert(nil, tt.code); result != tt.expected {
				t.Errorf("code=%q: expected %v, got %v", tt.code, tt.expected, result)
			}
		})
	}
}

func TestTypeBasedRule_ShouldConvert(t *testing.T) {
	typeSystem := NewTypeInferenceEngine()
	typeSystem.RegisterVariable("enabled", "bool")
	typeSystem.RegisterVariable("price", "float64")
	typeSystem.RegisterVariable("count", "int")

	rule := NewTypeBasedRule(typeSystem)

	tests := []struct {
		name     string
		expr     ast.Expression
		expected bool
	}{
		{
			name:     "skip bool variable (already bool)",
			expr:     &ast.Identifier{Name: "enabled"},
			expected: false,
		},
		{
			name:     "convert float64 variable (needs != 0)",
			expr:     &ast.Identifier{Name: "price"},
			expected: true,
		},
		{
			name:     "convert int variable (needs != 0)",
			expr:     &ast.Identifier{Name: "count"},
			expected: true,
		},
		{
			name:     "skip unregistered variable (conservative)",
			expr:     &ast.Identifier{Name: "unknown"},
			expected: false,
		},
		{
			name:     "skip nil expression",
			expr:     nil,
			expected: false,
		},
		{
			name: "skip bool member expression (already bool)",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "enabled"},
				Property: &ast.Identifier{Name: "value"},
			},
			expected: false,
		},
		{
			name: "convert float64 member expression (needs != 0)",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "price"},
				Property: &ast.Identifier{Name: "value"},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if result := rule.ShouldConvert(tt.expr, ""); result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestConversionRule_Composition(t *testing.T) {
	comparisonMatcher := NewComparisonPattern()
	seriesMatcher := NewSeriesAccessPattern()
	typeSystem := NewTypeInferenceEngine()
	typeSystem.RegisterVariable("enabled", "bool")

	skipRule := NewSkipComparisonRule(comparisonMatcher)
	seriesRule := NewConvertSeriesAccessRule(seriesMatcher)
	typeRule := NewTypeBasedRule(typeSystem)

	tests := []struct {
		name             string
		code             string
		expr             ast.Expression
		expectSkip       bool
		expectSeries     bool
		expectType       bool
		expectedDecision string
	}{
		{
			name:             "comparison blocks all rules",
			code:             "price > 100",
			expr:             &ast.Identifier{Name: "price"},
			expectSkip:       false,
			expectSeries:     false,
			expectType:       false, // unregistered identifier → conservative, don't convert
			expectedDecision: "skip conversion due to comparison",
		},
		{
			name:             "Series without comparison converts",
			code:             "priceSeries.GetCurrent()",
			expr:             &ast.Identifier{Name: "price"},
			expectSkip:       true,
			expectSeries:     true,
			expectType:       false, // unregistered identifier → conservative
			expectedDecision: "convert via Series rule (takes precedence)",
		},
		{
			name:             "bool type without Series skips conversion",
			code:             "enabled",
			expr:             &ast.Identifier{Name: "enabled"},
			expectSkip:       true,
			expectSeries:     false,
			expectType:       false, // bool variable → don't convert
			expectedDecision: "no conversion (already bool)",
		},
		{
			name:             "neither pattern nor type - conservative",
			code:             "bar.Close",
			expr:             &ast.Identifier{Name: "close"},
			expectSkip:       true,
			expectSeries:     false,
			expectType:       false, // unregistered identifier → conservative
			expectedDecision: "no conversion (unregistered, conservative)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			skipResult := skipRule.ShouldConvert(tt.expr, tt.code)
			seriesResult := seriesRule.ShouldConvert(tt.expr, tt.code)
			typeResult := typeRule.ShouldConvert(tt.expr, tt.code)

			if skipResult != tt.expectSkip {
				t.Errorf("skip rule: expected %v, got %v", tt.expectSkip, skipResult)
			}
			if seriesResult != tt.expectSeries {
				t.Errorf("series rule: expected %v, got %v", tt.expectSeries, seriesResult)
			}
			if typeResult != tt.expectType {
				t.Errorf("type rule: expected %v, got %v", tt.expectType, typeResult)
			}
		})
	}
}
