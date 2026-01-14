package codegen

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
)

/* Tests literal value handling across all types */
func TestBooleanConverter_LiteralHandling(t *testing.T) {
	typeSystem := NewTypeInferenceEngine()
	converter := NewBooleanConverter(typeSystem)

	tests := []struct {
		name        string
		expr        ast.Expression
		code        string
		expected    string
		description string
	}{
		{
			name:        "bool literal true",
			expr:        &ast.Literal{Value: true},
			code:        "true",
			expected:    "true",
			description: "Bool literals are already boolean - no conversion",
		},
		{
			name:        "bool literal false",
			expr:        &ast.Literal{Value: false},
			code:        "false",
			expected:    "false",
			description: "Bool literals are already boolean - no conversion",
		},
		{
			name:        "numeric literal integer",
			expr:        &ast.Literal{Value: 42},
			code:        "42",
			expected:    "42",
			description: "Numeric literals not wrapped - Go handles implicit conversion",
		},
		{
			name:        "numeric literal float",
			expr:        &ast.Literal{Value: 3.14},
			code:        "3.14",
			expected:    "3.14",
			description: "Float literals not wrapped - would create type error",
		},
		{
			name:        "numeric literal zero",
			expr:        &ast.Literal{Value: 0},
			code:        "0",
			expected:    "0",
			description: "Zero literal not wrapped - explicit false value",
		},
		{
			name:        "string literal",
			expr:        &ast.Literal{Value: "BINANCE"},
			code:        `"BINANCE"`,
			expected:    `"BINANCE"`,
			description: "String literals pass through unchanged",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := converter.EnsureBooleanOperand(tt.expr, tt.code)
			if result != tt.expected {
				t.Errorf("%s\nexpected: %q\ngot:      %q", tt.description, tt.expected, result)
			}
		})
	}
}

/* Tests unary expression handling with various operators and operand types */
func TestBooleanConverter_UnaryExpressions(t *testing.T) {
	typeSystem := NewTypeInferenceEngine()
	converter := NewBooleanConverter(typeSystem)

	tests := []struct {
		name        string
		expr        ast.Expression
		code        string
		expected    string
		description string
	}{
		{
			name: "unary minus with numeric literal",
			expr: &ast.UnaryExpression{
				Operator: "-",
				Argument: &ast.Literal{Value: 50.0},
				Prefix:   true,
			},
			code:        "-50.00",
			expected:    "-50.00",
			description: "Unary minus on literal not wrapped - arithmetic expression",
		},
		{
			name: "unary plus with numeric literal",
			expr: &ast.UnaryExpression{
				Operator: "+",
				Argument: &ast.Literal{Value: 100},
				Prefix:   true,
			},
			code:        "+100",
			expected:    "+100",
			description: "Unary plus on literal not wrapped",
		},
		{
			name: "logical not with identifier",
			expr: &ast.UnaryExpression{
				Operator: "not",
				Argument: &ast.Identifier{Name: "signal"},
				Prefix:   true,
			},
			code:        "!signal",
			expected:    "!signal",
			description: "Logical not produces boolean - no wrapping needed in ConvertBoolSeriesForIfStatement",
		},
		{
			name: "unary minus with identifier",
			expr: &ast.UnaryExpression{
				Operator: "-",
				Argument: &ast.Identifier{Name: "value"},
				Prefix:   true,
			},
			code:        "-value",
			expected:    "(value.IsTrue(-value))",
			description: "Unary minus on identifier may need conversion in operand context",
		},
		{
			name: "unary minus with Series access",
			expr: &ast.UnaryExpression{
				Operator: "-",
				Argument: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "price"},
					Property: &ast.Identifier{Name: "GetCurrent"},
				},
			},
			code:        "-priceSeries.GetCurrent()",
			expected:    "(value.IsTrue(-priceSeries.GetCurrent()))",
			description: "Unary minus on Series needs wrapping",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := converter.EnsureBooleanOperand(tt.expr, tt.code)
			if result != tt.expected {
				t.Errorf("%s\nexpected: %q\ngot:      %q", tt.description, tt.expected, result)
			}
		})
	}
}

/* Tests Series access patterns with various contexts */
func TestBooleanConverter_SeriesAccessPatterns(t *testing.T) {
	typeSystem := NewTypeInferenceEngine()
	converter := NewBooleanConverter(typeSystem)

	tests := []struct {
		name        string
		expr        ast.Expression
		code        string
		expectedOp  string
		expectedIf  string
		description string
	}{
		{
			name:        "GetCurrent() method",
			expr:        &ast.Identifier{Name: "signal"},
			code:        "signalSeries.GetCurrent()",
			expectedOp:  "(value.IsTrue(signalSeries.GetCurrent()))",
			expectedIf:  "value.IsTrue(signalSeries.GetCurrent())",
			description: "Current value access needs wrapping",
		},
		{
			name:        "Get(1) historical access",
			expr:        &ast.Identifier{Name: "previous"},
			code:        "previousSeries.Get(1)",
			expectedOp:  "(value.IsTrue(previousSeries.Get(1)))",
			expectedIf:  "value.IsTrue(previousSeries.Get(1))",
			description: "Historical access (1 bar ago) needs wrapping",
		},
		{
			name:        "Get(N) multi-bar historical",
			expr:        &ast.Identifier{Name: "past"},
			code:        "pastSeries.Get(5)",
			expectedOp:  "(value.IsTrue(pastSeries.Get(5)))",
			expectedIf:  "value.IsTrue(pastSeries.Get(5))",
			description: "Deep historical access needs wrapping",
		},
		{
			name:        "nested Series in function call",
			expr:        &ast.Identifier{Name: "sma"},
			code:        "ta.Sma(closeSeries.GetCurrent(), 20)",
			expectedOp:  "(value.IsTrue(ta.Sma(closeSeries.GetCurrent(), 20)))",
			expectedIf:  "value.IsTrue(ta.Sma(closeSeries.GetCurrent(), 20))",
			description: "Function call with Series argument needs wrapping",
		},
		{
			name:        "multiple Series in expression",
			expr:        &ast.Identifier{Name: "combined"},
			code:        "highSeries.GetCurrent() - lowSeries.GetCurrent()",
			expectedOp:  "(value.IsTrue(highSeries.GetCurrent() - lowSeries.GetCurrent()))",
			expectedIf:  "value.IsTrue(highSeries.GetCurrent() - lowSeries.GetCurrent())",
			description: "Arithmetic with multiple Series needs wrapping",
		},
		{
			name:        "Series in comparison stays unchanged",
			expr:        &ast.BinaryExpression{Operator: ">"},
			code:        "priceSeries.GetCurrent() > 100",
			expectedOp:  "priceSeries.GetCurrent() > 100",
			expectedIf:  "priceSeries.GetCurrent() > 100",
			description: "Comparison with Series is already boolean",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resultOp := converter.EnsureBooleanOperand(tt.expr, tt.code)
			if resultOp != tt.expectedOp {
				t.Errorf("%s (EnsureBooleanOperand)\nexpected: %q\ngot:      %q",
					tt.description, tt.expectedOp, resultOp)
			}

			resultIf := converter.ConvertBoolSeriesForIfStatement(tt.expr, tt.code)
			if resultIf != tt.expectedIf {
				t.Errorf("%s (ConvertBoolSeriesForIfStatement)\nexpected: %q\ngot:      %q",
					tt.description, tt.expectedIf, resultIf)
			}
		})
	}
}

/* Tests type-based conversion with registered and unregistered identifiers */
func TestBooleanConverter_TypeBasedConversion(t *testing.T) {
	typeSystem := NewTypeInferenceEngine()
	typeSystem.RegisterVariable("enabled", "bool")
	typeSystem.RegisterVariable("signal", "bool")
	typeSystem.RegisterVariable("price", "float64")
	typeSystem.RegisterVariable("volume", "int")
	typeSystem.RegisterConstant("show_trades", true)
	typeSystem.RegisterConstant("threshold", 50.0)

	converter := NewBooleanConverter(typeSystem)

	tests := []struct {
		name        string
		expr        ast.Expression
		code        string
		expectedIf  string
		description string
	}{
		{
			name:        "bool variable no conversion",
			expr:        &ast.Identifier{Name: "enabled"},
			code:        "enabled",
			expectedIf:  "enabled",
			description: "Bool variables are already boolean type",
		},
		{
			name:        "bool constant no conversion",
			expr:        &ast.Identifier{Name: "show_trades"},
			code:        "show_trades",
			expectedIf:  "show_trades",
			description: "Bool constants from input.bool are already boolean",
		},
		{
			name:        "float64 variable gets conversion",
			expr:        &ast.Identifier{Name: "price"},
			code:        "price",
			expectedIf:  "value.IsTrue(price)",
			description: "Float64 variables need explicit boolean conversion",
		},
		{
			name:        "int variable gets conversion",
			expr:        &ast.Identifier{Name: "volume"},
			code:        "volume",
			expectedIf:  "value.IsTrue(volume)",
			description: "Int variables need explicit boolean conversion",
		},
		{
			name:        "float64 constant conservative",
			expr:        &ast.Identifier{Name: "threshold"},
			code:        "threshold",
			expectedIf:  "threshold",
			description: "Numeric constants handled conservatively",
		},
		{
			name:        "unregistered identifier conservative",
			expr:        &ast.Identifier{Name: "unknown"},
			code:        "unknown",
			expectedIf:  "unknown",
			description: "Unknown identifiers pass through (conservative)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := converter.ConvertBoolSeriesForIfStatement(tt.expr, tt.code)
			if result != tt.expectedIf {
				t.Errorf("%s\nexpected: %q\ngot:      %q", tt.description, tt.expectedIf, result)
			}
		})
	}
}

/* Tests complex nested expressions and edge cases */
func TestBooleanConverter_ComplexExpressions(t *testing.T) {
	typeSystem := NewTypeInferenceEngine()
	typeSystem.RegisterVariable("condition", "bool")
	converter := NewBooleanConverter(typeSystem)

	tests := []struct {
		name        string
		expr        ast.Expression
		code        string
		expected    string
		description string
	}{
		{
			name: "nested ternary with Series",
			expr: &ast.ConditionalExpression{
				Test:       &ast.Identifier{Name: "signal"},
				Consequent: &ast.Literal{Value: 1.0},
				Alternate:  &ast.Literal{Value: 0.0},
			},
			code:        "signalSeries.GetCurrent() != 0 ? 1.0 : 0.0",
			expected:    "(value.IsTrue(signalSeries.GetCurrent() != 0 ? 1.0 : 0.0))",
			description: "Ternary expression result needs wrapping",
		},
		{
			name: "logical AND with Series and comparison",
			expr: &ast.LogicalExpression{
				Operator: "&&",
				Left:     &ast.Identifier{Name: "a"},
				Right:    &ast.BinaryExpression{Operator: ">"},
			},
			code:        "aSeries.GetCurrent() && price > 100",
			expected:    "aSeries.GetCurrent() && price > 100",
			description: "Logical expression is already boolean",
		},
		{
			name: "function call with multiple Series arguments",
			expr: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ta"},
					Property: &ast.Identifier{Name: "crossover"},
				},
			},
			code:        "ta.Crossover(fastSeries.GetCurrent(), slowSeries.GetCurrent())",
			expected:    "ta.Crossover(fastSeries.GetCurrent(), slowSeries.GetCurrent())",
			description: "Boolean function calls not wrapped (already return bool)",
		},
		{
			name:        "arithmetic with Series and literals",
			expr:        &ast.Identifier{Name: "result"},
			code:        "priceSeries.GetCurrent() * 1.5 + 10",
			expected:    "(value.IsTrue(priceSeries.GetCurrent() * 1.5 + 10))",
			description: "Complex arithmetic needs wrapping",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := converter.EnsureBooleanOperand(tt.expr, tt.code)
			if result != tt.expected {
				t.Errorf("%s\nexpected: %q\ngot:      %q", tt.description, tt.expected, result)
			}
		})
	}
}
