package codegen

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestBooleanConverter_EnsureBooleanOperand_BooleanOperands(t *testing.T) {
	tests := []struct {
		name          string
		expr          ast.Expression
		generatedCode string
	}{
		{
			name: "comparison expression already bool",
			expr: &ast.BinaryExpression{
				Operator: ">",
				Left:     &ast.Identifier{Name: "close"},
				Right:    &ast.Literal{Value: 100.0},
			},
			generatedCode: "(close.GetCurrent() > 100.00)",
		},
		{
			name: "crossover function already bool",
			expr: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ta"},
					Property: &ast.Identifier{Name: "crossover"},
				},
			},
			generatedCode: "ta.Crossover(...)",
		},
		{
			name: "crossunder function already bool",
			expr: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ta"},
					Property: &ast.Identifier{Name: "crossunder"},
				},
			},
			generatedCode: "ta.Crossunder(...)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			typeSystem := NewTypeInferenceEngine()
			converter := NewBooleanConverter(typeSystem)

			if converter.IsAlreadyBoolean(tt.expr) {
				result := converter.EnsureBooleanOperand(tt.expr, tt.generatedCode)
				if result != tt.generatedCode {
					t.Errorf("expected operand unchanged %q, got %q", tt.generatedCode, result)
				}
			}
		})
	}
}

func TestBooleanConverter_EnsureBooleanOperand_Float64Series(t *testing.T) {
	tests := []struct {
		name          string
		generatedCode string
		expr          ast.Expression
		expected      string
	}{
		{
			name:          "float64 Series identifier wrapped",
			generatedCode: "enabledSeries.GetCurrent()",
			expr:          &ast.Identifier{Name: "enabled"},
			expected:      "(enabledSeries.GetCurrent() != 0)",
		},
		{
			name:          "float64 Series member access wrapped",
			generatedCode: "valueSeries.GetCurrent()",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "value"},
				Property: &ast.Identifier{Name: "prop"},
			},
			expected: "(valueSeries.GetCurrent() != 0)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			typeSystem := NewTypeInferenceEngine()
			converter := NewBooleanConverter(typeSystem)

			result := converter.EnsureBooleanOperand(tt.expr, tt.generatedCode)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestBooleanConverter_IsAlreadyBoolean_ComparisonOperators(t *testing.T) {
	tests := []struct {
		name     string
		expr     *ast.BinaryExpression
		expected bool
	}{
		{
			name:     "greater than is boolean",
			expr:     &ast.BinaryExpression{Operator: ">"},
			expected: true,
		},
		{
			name:     "less than is boolean",
			expr:     &ast.BinaryExpression{Operator: "<"},
			expected: true,
		},
		{
			name:     "greater equal is boolean",
			expr:     &ast.BinaryExpression{Operator: ">="},
			expected: true,
		},
		{
			name:     "less equal is boolean",
			expr:     &ast.BinaryExpression{Operator: "<="},
			expected: true,
		},
		{
			name:     "equal is boolean",
			expr:     &ast.BinaryExpression{Operator: "=="},
			expected: true,
		},
		{
			name:     "not equal is boolean",
			expr:     &ast.BinaryExpression{Operator: "!="},
			expected: true,
		},
		{
			name:     "addition is not boolean",
			expr:     &ast.BinaryExpression{Operator: "+"},
			expected: false,
		},
		{
			name:     "multiplication is not boolean",
			expr:     &ast.BinaryExpression{Operator: "*"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			typeSystem := NewTypeInferenceEngine()
			converter := NewBooleanConverter(typeSystem)

			result := converter.IsAlreadyBoolean(tt.expr)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestBooleanConverter_IsBooleanFunction(t *testing.T) {
	tests := []struct {
		name     string
		expr     *ast.CallExpression
		expected bool
	}{
		{
			name: "crossover is boolean function",
			expr: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ta"},
					Property: &ast.Identifier{Name: "crossover"},
				},
			},
			expected: true,
		},
		{
			name: "crossunder is boolean function",
			expr: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ta"},
					Property: &ast.Identifier{Name: "crossunder"},
				},
			},
			expected: true,
		},
		{
			name: "sma is not boolean function",
			expr: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ta"},
					Property: &ast.Identifier{Name: "sma"},
				},
			},
			expected: false,
		},
		{
			name: "ema is not boolean function",
			expr: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ta"},
					Property: &ast.Identifier{Name: "ema"},
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			typeSystem := NewTypeInferenceEngine()
			converter := NewBooleanConverter(typeSystem)

			result := converter.IsBooleanFunction(tt.expr)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestBooleanConverter_IsFloat64SeriesAccess(t *testing.T) {
	tests := []struct {
		name     string
		operand  string
		expected bool
	}{
		{
			name:     "Series GetCurrent is float64 Series access",
			operand:  "enabledSeries.GetCurrent()",
			expected: true,
		},
		{
			name:     "bool constant is not Series access",
			operand:  "true",
			expected: false,
		},
		{
			name:     "comparison with GetCurrent is still Series access (contains pattern)",
			operand:  "(close.GetCurrent() > 100.00)",
			expected: true, // IsFloat64SeriesAccess checks for .GetCurrent() substring
		},
		{
			name:     "ta function call is not Series access",
			operand:  "ta.Crossover(...)",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			typeSystem := NewTypeInferenceEngine()
			converter := NewBooleanConverter(typeSystem)

			result := converter.IsFloat64SeriesAccess(tt.operand)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestBooleanConverter_ConvertBoolSeriesForIfStatement(t *testing.T) {
	tests := []struct {
		name          string
		generatedCode string
		expr          ast.Expression
		varName       string
		varType       string
		expected      string
	}{
		{
			name:          "bool variable Series converted",
			generatedCode: "enabledSeries.GetCurrent()",
			expr:          &ast.Identifier{Name: "enabled"},
			varName:       "enabled",
			varType:       "bool",
			expected:      "enabledSeries.GetCurrent() != 0",
		},
		{
			name:          "float64 variable not converted",
			generatedCode: "priceSeries.GetCurrent()",
			expr:          &ast.Identifier{Name: "price"},
			varName:       "price",
			varType:       "float64",
			expected:      "priceSeries.GetCurrent()",
		},
		{
			name:          "unregistered variable not converted",
			generatedCode: "unknownSeries.GetCurrent()",
			expr:          &ast.Identifier{Name: "unknown"},
			varName:       "unknown",
			varType:       "",
			expected:      "unknownSeries.GetCurrent()",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			typeSystem := NewTypeInferenceEngine()
			if tt.varType != "" {
				typeSystem.RegisterVariable(tt.varName, tt.varType)
			}
			converter := NewBooleanConverter(typeSystem)

			result := converter.ConvertBoolSeriesForIfStatement(tt.expr, tt.generatedCode)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestBooleanConverter_IsComparisonOperator(t *testing.T) {
	tests := []struct {
		operator string
		expected bool
	}{
		{">", true},
		{"<", true},
		{">=", true},
		{"<=", true},
		{"==", true},
		{"!=", true},
		{"+", false},
		{"-", false},
		{"*", false},
		{"/", false},
		{"%", false},
	}

	for _, tt := range tests {
		t.Run("operator "+tt.operator, func(t *testing.T) {
			typeSystem := NewTypeInferenceEngine()
			converter := NewBooleanConverter(typeSystem)

			result := converter.IsComparisonOperator(tt.operator)
			if result != tt.expected {
				t.Errorf("IsComparisonOperator(%q) expected %v, got %v", tt.operator, tt.expected, result)
			}
		})
	}
}

func TestBooleanConverter_EdgeCases(t *testing.T) {
	t.Run("nil expression not already boolean", func(t *testing.T) {
		typeSystem := NewTypeInferenceEngine()
		converter := NewBooleanConverter(typeSystem)

		result := converter.IsAlreadyBoolean(nil)
		if result {
			t.Error("expected false for nil expression")
		}
	})

	t.Run("non-BinaryExpression and non-CallExpression not already boolean", func(t *testing.T) {
		typeSystem := NewTypeInferenceEngine()
		converter := NewBooleanConverter(typeSystem)

		result := converter.IsAlreadyBoolean(&ast.Identifier{Name: "value"})
		if result {
			t.Error("expected false for Identifier expression")
		}
	})

	t.Run("empty code string handled", func(t *testing.T) {
		typeSystem := NewTypeInferenceEngine()
		converter := NewBooleanConverter(typeSystem)

		expr := &ast.Identifier{Name: "test"}
		result := converter.EnsureBooleanOperand(expr, "")
		expected := ""
		if result != expected {
			t.Errorf("expected %q, got %q", expected, result)
		}
	})

	t.Run("LogicalExpression recognized as boolean", func(t *testing.T) {
		typeSystem := NewTypeInferenceEngine()
		converter := NewBooleanConverter(typeSystem)

		expr := &ast.LogicalExpression{
			Operator: "&&",
			Left:     &ast.Identifier{Name: "cond1"},
			Right:    &ast.Identifier{Name: "cond2"},
		}

		result := converter.IsAlreadyBoolean(expr)
		if !result {
			t.Error("expected true for LogicalExpression")
		}
	})
}

func TestBooleanConverter_Integration_MixedTypes(t *testing.T) {
	typeSystem := NewTypeInferenceEngine()
	typeSystem.RegisterVariable("enabled", "bool")
	typeSystem.RegisterVariable("longSignal", "bool")
	typeSystem.RegisterVariable("price", "float64")
	converter := NewBooleanConverter(typeSystem)

	tests := []struct {
		name          string
		generatedCode string
		expr          ast.Expression
		expected      string
	}{
		{
			name:          "bool variable Series wrapped",
			generatedCode: "enabledSeries.GetCurrent()",
			expr:          &ast.Identifier{Name: "enabled"},
			expected:      "(enabledSeries.GetCurrent() != 0)",
		},
		{
			name:          "comparison already bool not wrapped",
			generatedCode: "(price.GetCurrent() > 100.00)",
			expr: &ast.BinaryExpression{
				Operator: ">",
				Left:     &ast.Identifier{Name: "price"},
				Right:    &ast.Literal{Value: 100.0},
			},
			expected: "(price.GetCurrent() > 100.00)",
		},
		{
			name:          "crossover function not wrapped",
			generatedCode: "ta.Crossover(close, sma)",
			expr: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ta"},
					Property: &ast.Identifier{Name: "crossover"},
				},
			},
			expected: "ta.Crossover(close, sma)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := converter.EnsureBooleanOperand(tt.expr, tt.generatedCode)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}
