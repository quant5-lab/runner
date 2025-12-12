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
		{"skip historical access", "priceSeries.Get(1)", false},
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
			name:     "convert bool variable",
			expr:     &ast.Identifier{Name: "enabled"},
			expected: true,
		},
		{
			name:     "skip float64 variable",
			expr:     &ast.Identifier{Name: "price"},
			expected: false,
		},
		{
			name:     "skip int variable",
			expr:     &ast.Identifier{Name: "count"},
			expected: false,
		},
		{
			name:     "skip unregistered variable",
			expr:     &ast.Identifier{Name: "unknown"},
			expected: false,
		},
		{
			name:     "skip nil expression",
			expr:     nil,
			expected: false,
		},
		{
			name: "convert bool member expression",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "enabled"},
				Property: &ast.Identifier{Name: "value"},
			},
			expected: true,
		},
		{
			name: "skip float64 member expression",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "price"},
				Property: &ast.Identifier{Name: "value"},
			},
			expected: false,
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
			expectType:       false,
			expectedDecision: "skip conversion",
		},
		{
			name:             "Series without comparison converts",
			code:             "priceSeries.GetCurrent()",
			expr:             &ast.Identifier{Name: "price"},
			expectSkip:       true,
			expectSeries:     true,
			expectType:       false,
			expectedDecision: "convert via Series rule",
		},
		{
			name:             "bool type without Series converts",
			code:             "enabled",
			expr:             &ast.Identifier{Name: "enabled"},
			expectSkip:       true,
			expectSeries:     false,
			expectType:       true,
			expectedDecision: "convert via type rule",
		},
		{
			name:             "neither pattern nor type",
			code:             "bar.Close",
			expr:             &ast.Identifier{Name: "close"},
			expectSkip:       true,
			expectSeries:     false,
			expectType:       false,
			expectedDecision: "no conversion",
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
