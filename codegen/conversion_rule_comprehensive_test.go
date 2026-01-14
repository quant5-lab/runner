package codegen

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
)

/* Tests type-based conversion rule with all type combinations */
func TestTypeBasedRule_AllTypeCombinations(t *testing.T) {
	typeSystem := NewTypeInferenceEngine()

	// Register variables of all types
	typeSystem.RegisterVariable("bool_var", "bool")
	typeSystem.RegisterVariable("float64_var", "float64")
	typeSystem.RegisterVariable("int_var", "int")
	typeSystem.RegisterVariable("string_var", "string")

	// Register constants of all types
	typeSystem.RegisterConstant("bool_const", true)
	typeSystem.RegisterConstant("float_const", 42.5)
	typeSystem.RegisterConstant("int_const", 100)
	typeSystem.RegisterConstant("string_const", "test")

	rule := NewTypeBasedRule(typeSystem)

	tests := []struct {
		name        string
		expr        ast.Expression
		expected    bool
		description string
	}{
		// Bool types - never convert
		{
			name:        "bool variable",
			expr:        &ast.Identifier{Name: "bool_var"},
			expected:    false,
			description: "Bool variables are already boolean",
		},
		{
			name:        "bool constant",
			expr:        &ast.Identifier{Name: "bool_const"},
			expected:    false,
			description: "Bool constants are already boolean",
		},

		// Numeric types - always convert
		{
			name:        "float64 variable",
			expr:        &ast.Identifier{Name: "float64_var"},
			expected:    true,
			description: "Float64 needs explicit boolean conversion",
		},
		{
			name:        "int variable",
			expr:        &ast.Identifier{Name: "int_var"},
			expected:    true,
			description: "Int needs explicit boolean conversion",
		},
		{
			name:        "float constant conservative",
			expr:        &ast.Identifier{Name: "float_const"},
			expected:    false,
			description: "Constants handled conservatively (may be used as literals)",
		},
		{
			name:        "int constant conservative",
			expr:        &ast.Identifier{Name: "int_const"},
			expected:    false,
			description: "Int constants handled conservatively",
		},

		// String type - depends on context
		{
			name:        "string variable",
			expr:        &ast.Identifier{Name: "string_var"},
			expected:    true,
			description: "String variables need conversion",
		},
		{
			name:        "string constant conservative",
			expr:        &ast.Identifier{Name: "string_const"},
			expected:    false,
			description: "String constants handled conservatively",
		},

		// Unregistered identifiers - conservative no conversion
		{
			name:        "unknown identifier",
			expr:        &ast.Identifier{Name: "unknown"},
			expected:    false,
			description: "Unknown identifiers are not converted (conservative)",
		},
		{
			name:        "different unknown",
			expr:        &ast.Identifier{Name: "undefined_var"},
			expected:    false,
			description: "Unregistered variables pass through",
		},

		// Nil expression
		{
			name:        "nil expression",
			expr:        nil,
			expected:    false,
			description: "Nil expression returns false",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rule.ShouldConvert(tt.expr, "")
			if result != tt.expected {
				t.Errorf("%s\nexpected: %v\ngot:      %v", tt.description, tt.expected, result)
			}
		})
	}
}

/* Tests member expression handling in type-based rule */
func TestTypeBasedRule_MemberExpressions(t *testing.T) {
	typeSystem := NewTypeInferenceEngine()
	typeSystem.RegisterVariable("enabled", "bool")
	typeSystem.RegisterVariable("price", "float64")
	typeSystem.RegisterVariable("count", "int")

	rule := NewTypeBasedRule(typeSystem)

	tests := []struct {
		name        string
		expr        ast.Expression
		expected    bool
		description string
	}{
		{
			name: "bool variable member",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "enabled"},
				Property: &ast.Identifier{Name: "value"},
			},
			expected:    false,
			description: "Member of bool variable not converted",
		},
		{
			name: "float64 variable member",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "price"},
				Property: &ast.Identifier{Name: "value"},
			},
			expected:    true,
			description: "Member of float64 variable needs conversion",
		},
		{
			name: "int variable member",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "count"},
				Property: &ast.Identifier{Name: "value"},
			},
			expected:    true,
			description: "Member of int variable needs conversion",
		},
		{
			name: "unknown variable member",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "unknown"},
				Property: &ast.Identifier{Name: "field"},
			},
			expected:    false,
			description: "Member of unknown variable conservative",
		},
		{
			name: "non-identifier object",
			expr: &ast.MemberExpression{
				Object:   &ast.Literal{Value: 100},
				Property: &ast.Identifier{Name: "toString"},
			},
			expected:    false,
			description: "Member expression with non-identifier object",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rule.ShouldConvert(tt.expr, "")
			if result != tt.expected {
				t.Errorf("%s\nexpected: %v\ngot:      %v", tt.description, tt.expected, result)
			}
		})
	}
}

/* Tests rule composition and interaction */
func TestConversionRules_InteractionPatterns(t *testing.T) {
	comparisonMatcher := NewComparisonPattern()
	seriesMatcher := NewSeriesAccessPattern()

	typeSystem := NewTypeInferenceEngine()
	typeSystem.RegisterVariable("enabled", "bool")
	typeSystem.RegisterVariable("signal", "float64")

	skipRule := NewSkipComparisonRule(comparisonMatcher)
	seriesRule := NewConvertSeriesAccessRule(seriesMatcher)
	typeRule := NewTypeBasedRule(typeSystem)

	tests := []struct {
		name          string
		code          string
		expr          ast.Expression
		skipResult    bool
		seriesResult  bool
		typeResult    bool
		finalDecision string
		description   string
	}{
		{
			name:          "comparison blocks all",
			code:          "signal > 100",
			expr:          &ast.Identifier{Name: "signal"},
			skipResult:    false,
			seriesResult:  false,
			typeResult:    true,
			finalDecision: "skip - comparison present",
			description:   "Comparison operator blocks conversion",
		},
		{
			name:          "Series triggers conversion",
			code:          "signalSeries.GetCurrent()",
			expr:          &ast.Identifier{Name: "signal"},
			skipResult:    true,
			seriesResult:  true,
			typeResult:    true,
			finalDecision: "convert - Series pattern",
			description:   "Series access pattern triggers conversion",
		},
		{
			name:          "historical Series Get(N)",
			code:          "signalSeries.Get(2)",
			expr:          &ast.Identifier{Name: "signal"},
			skipResult:    true,
			seriesResult:  true,
			typeResult:    true,
			finalDecision: "convert - Series Get(N) pattern",
			description:   "Historical access triggers conversion",
		},
		{
			name:          "type rule for float64",
			code:          "signal",
			expr:          &ast.Identifier{Name: "signal"},
			skipResult:    true,
			seriesResult:  false,
			typeResult:    true,
			finalDecision: "convert - float64 type",
			description:   "Float64 variable triggers type-based conversion",
		},
		{
			name:          "bool variable no conversion",
			code:          "enabled",
			expr:          &ast.Identifier{Name: "enabled"},
			skipResult:    true,
			seriesResult:  false,
			typeResult:    false,
			finalDecision: "no conversion - already bool",
			description:   "Bool variable passes through all rules",
		},
		{
			name:          "unknown identifier conservative",
			code:          "unknown",
			expr:          &ast.Identifier{Name: "unknown"},
			skipResult:    true,
			seriesResult:  false,
			typeResult:    false,
			finalDecision: "no conversion - unknown",
			description:   "Unknown identifier has no conversion",
		},
		{
			name:          "Series with comparison",
			code:          "priceSeries.GetCurrent() > 100",
			expr:          &ast.BinaryExpression{Operator: ">"},
			skipResult:    false,
			seriesResult:  true,
			typeResult:    false,
			finalDecision: "skip - comparison present",
			description:   "Comparison takes precedence over Series",
		},
		{
			name:          "multiple Series no comparison",
			code:          "aSeries.GetCurrent() + bSeries.Get(1)",
			expr:          &ast.Identifier{Name: "result"},
			skipResult:    true,
			seriesResult:  true,
			typeResult:    false,
			finalDecision: "convert - Series pattern",
			description:   "Arithmetic with Series needs conversion",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			skipResult := skipRule.ShouldConvert(tt.expr, tt.code)
			seriesResult := seriesRule.ShouldConvert(tt.expr, tt.code)
			typeResult := typeRule.ShouldConvert(tt.expr, tt.code)

			if skipResult != tt.skipResult {
				t.Errorf("%s\nskip rule: expected %v, got %v",
					tt.description, tt.skipResult, skipResult)
			}
			if seriesResult != tt.seriesResult {
				t.Errorf("%s\nseries rule: expected %v, got %v",
					tt.description, tt.seriesResult, seriesResult)
			}
			if typeResult != tt.typeResult {
				t.Errorf("%s\ntype rule: expected %v, got %v",
					tt.description, tt.typeResult, typeResult)
			}
		})
	}
}

/* Tests skip comparison rule with various comparison operators */
func TestSkipComparisonRule_AllOperators(t *testing.T) {
	comparisonMatcher := NewComparisonPattern()
	rule := NewSkipComparisonRule(comparisonMatcher)

	tests := []struct {
		name        string
		code        string
		expected    bool
		description string
	}{
		// Should skip (comparison present) - returns false
		{
			name:        "greater than",
			code:        "x > 10",
			expected:    false,
			description: "Skip conversion when > present",
		},
		{
			name:        "less than",
			code:        "x < 10",
			expected:    false,
			description: "Skip conversion when < present",
		},
		{
			name:        "equal",
			code:        "x == 10",
			expected:    false,
			description: "Skip conversion when == present",
		},
		{
			name:        "not equal",
			code:        "x != 10",
			expected:    false,
			description: "Skip conversion when != present",
		},
		{
			name:        "greater or equal",
			code:        "x >= 10",
			expected:    false,
			description: "Skip conversion when >= present",
		},
		{
			name:        "less or equal",
			code:        "x <= 10",
			expected:    false,
			description: "Skip conversion when <= present",
		},

		// Should not skip (no comparison) - returns true
		{
			name:        "Series access only",
			code:        "priceSeries.GetCurrent()",
			expected:    true,
			description: "Don't skip when no comparison",
		},
		{
			name:        "identifier only",
			code:        "signal",
			expected:    true,
			description: "Don't skip plain identifier",
		},
		{
			name:        "arithmetic expression",
			code:        "x + 10",
			expected:    true,
			description: "Don't skip arithmetic",
		},
		{
			name:        "function call",
			code:        "ta.Sma(close, 20)",
			expected:    true,
			description: "Don't skip function call",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rule.ShouldConvert(nil, tt.code)
			if result != tt.expected {
				t.Errorf("%s\ncode=%q\nexpected: %v\ngot:      %v",
					tt.description, tt.code, tt.expected, result)
			}
		})
	}
}

/* Tests Series access rule with various patterns */
func TestConvertSeriesAccessRule_AllPatterns(t *testing.T) {
	seriesMatcher := NewSeriesAccessPattern()
	rule := NewConvertSeriesAccessRule(seriesMatcher)

	tests := []struct {
		name        string
		code        string
		expected    bool
		description string
	}{
		// Should convert (Series pattern present)
		{
			name:        "GetCurrent method",
			code:        "priceSeries.GetCurrent()",
			expected:    true,
			description: "Convert Series.GetCurrent() access",
		},
		{
			name:        "Get(1) historical",
			code:        "priceSeries.Get(1)",
			expected:    true,
			description: "Convert Series.Get(1) historical access",
		},
		{
			name:        "Get(N) deep history",
			code:        "signalSeries.Get(5)",
			expected:    true,
			description: "Convert Series.Get(N) deep historical",
		},
		{
			name:        "nested in function",
			code:        "ta.Ema(closeSeries.GetCurrent(), 10)",
			expected:    true,
			description: "Convert nested Series access",
		},
		{
			name:        "multiple Series",
			code:        "aSeries.GetCurrent() + bSeries.Get(1)",
			expected:    true,
			description: "Convert when multiple Series present",
		},

		// Should not convert (no Series pattern)
		{
			name:        "plain identifier",
			code:        "price",
			expected:    false,
			description: "Don't convert plain identifier",
		},
		{
			name:        "numeric literal",
			code:        "100",
			expected:    false,
			description: "Don't convert literal",
		},
		{
			name:        "comparison",
			code:        "x > 100",
			expected:    false,
			description: "Don't convert comparison without Series",
		},
		{
			name:        "empty string",
			code:        "",
			expected:    false,
			description: "Don't convert empty code",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rule.ShouldConvert(nil, tt.code)
			if result != tt.expected {
				t.Errorf("%s\ncode=%q\nexpected: %v\ngot:      %v",
					tt.description, tt.code, tt.expected, result)
			}
		})
	}
}
