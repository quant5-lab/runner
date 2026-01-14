package codegen

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestTypeInferenceEngine_InferType_BinaryExpression(t *testing.T) {
	tests := []struct {
		name     string
		expr     *ast.BinaryExpression
		expected string
	}{
		{
			name: "comparison operator returns bool",
			expr: &ast.BinaryExpression{
				Operator: ">",
				Left:     &ast.Identifier{Name: "close"},
				Right:    &ast.Literal{Value: 100.0},
			},
			expected: "bool",
		},
		{
			name: "equality operator returns bool",
			expr: &ast.BinaryExpression{
				Operator: "==",
				Left:     &ast.Identifier{Name: "value"},
				Right:    &ast.Literal{Value: 50.0},
			},
			expected: "bool",
		},
		{
			name: "arithmetic operator returns float64",
			expr: &ast.BinaryExpression{
				Operator: "+",
				Left:     &ast.Identifier{Name: "close"},
				Right:    &ast.Literal{Value: 10.0},
			},
			expected: "float64",
		},
		{
			name: "less than operator returns bool",
			expr: &ast.BinaryExpression{
				Operator: "<",
				Left:     &ast.Identifier{Name: "low"},
				Right:    &ast.Identifier{Name: "support"},
			},
			expected: "bool",
		},
		{
			name: "not equal operator returns bool",
			expr: &ast.BinaryExpression{
				Operator: "!=",
				Left:     &ast.Identifier{Name: "status"},
				Right:    &ast.Literal{Value: 0.0},
			},
			expected: "bool",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := NewTypeInferenceEngine()
			result := engine.InferType(tt.expr)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestTypeInferenceEngine_InferType_LogicalExpression(t *testing.T) {
	tests := []struct {
		name     string
		expr     *ast.LogicalExpression
		expected string
	}{
		{
			name: "AND operator returns bool",
			expr: &ast.LogicalExpression{
				Operator: "&&",
				Left:     &ast.Identifier{Name: "cond1"},
				Right:    &ast.Identifier{Name: "cond2"},
			},
			expected: "bool",
		},
		{
			name: "OR operator returns bool",
			expr: &ast.LogicalExpression{
				Operator: "||",
				Left: &ast.BinaryExpression{
					Operator: ">",
					Left:     &ast.Identifier{Name: "close"},
					Right:    &ast.Literal{Value: 100.0},
				},
				Right: &ast.Identifier{Name: "enabled"},
			},
			expected: "bool",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := NewTypeInferenceEngine()
			result := engine.InferType(tt.expr)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestTypeInferenceEngine_InferType_CallExpression(t *testing.T) {
	tests := []struct {
		name     string
		expr     *ast.CallExpression
		expected string
	}{
		{
			name: "input.bool returns bool",
			expr: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "input"},
					Property: &ast.Identifier{Name: "bool"},
				},
				Arguments: []ast.Expression{
					&ast.Literal{Value: true},
				},
			},
			expected: "bool",
		},
		{
			name: "ta.crossover returns bool",
			expr: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ta"},
					Property: &ast.Identifier{Name: "crossover"},
				},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
					&ast.Identifier{Name: "sma"},
				},
			},
			expected: "bool",
		},
		{
			name: "ta.crossunder returns bool",
			expr: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ta"},
					Property: &ast.Identifier{Name: "crossunder"},
				},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "rsi"},
					&ast.Literal{Value: 30.0},
				},
			},
			expected: "bool",
		},
		{
			name: "ta.sma returns float64",
			expr: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ta"},
					Property: &ast.Identifier{Name: "sma"},
				},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
					&ast.Literal{Value: 20.0},
				},
			},
			expected: "float64",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := NewTypeInferenceEngine()
			result := engine.InferType(tt.expr)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestTypeInferenceEngine_RegisterVariable(t *testing.T) {
	tests := []struct {
		name        string
		varName     string
		varType     string
		checkIsBool bool
		expected    bool
	}{
		{
			name:        "register bool variable",
			varName:     "enabled",
			varType:     "bool",
			checkIsBool: true,
			expected:    true,
		},
		{
			name:        "register float64 variable",
			varName:     "price",
			varType:     "float64",
			checkIsBool: true,
			expected:    false,
		},
		{
			name:        "register string variable",
			varName:     "symbol",
			varType:     "string",
			checkIsBool: true,
			expected:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := NewTypeInferenceEngine()
			engine.RegisterVariable(tt.varName, tt.varType)

			if tt.checkIsBool {
				result := engine.IsBoolVariableByName(tt.varName)
				if result != tt.expected {
					t.Errorf("IsBoolVariableByName(%q) expected %v, got %v", tt.varName, tt.expected, result)
				}
			}

			varType, exists := engine.GetVariableType(tt.varName)
			if !exists {
				t.Errorf("GetVariableType(%q) expected to exist", tt.varName)
			}
			if varType != tt.varType {
				t.Errorf("GetVariableType(%q) expected %q, got %q", tt.varName, tt.varType, varType)
			}
		})
	}
}

func TestTypeInferenceEngine_RegisterConstant(t *testing.T) {
	tests := []struct {
		name        string
		constName   string
		value       interface{}
		checkIsBool bool
		expected    bool
	}{
		{
			name:        "register bool constant",
			constName:   "showTrades",
			value:       true,
			checkIsBool: true,
			expected:    true,
		},
		{
			name:        "register float constant",
			constName:   "multiplier",
			value:       1.5,
			checkIsBool: true,
			expected:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := NewTypeInferenceEngine()
			engine.RegisterConstant(tt.constName, tt.value)

			if tt.checkIsBool {
				result := engine.IsBoolConstant(tt.constName)
				if result != tt.expected {
					t.Errorf("IsBoolConstant(%q) expected %v, got %v", tt.constName, tt.expected, result)
				}
			}
		})
	}
}

func TestTypeInferenceEngine_IsBoolVariableByName(t *testing.T) {
	tests := []struct {
		name       string
		varName    string
		varType    string
		checkNames []string
		expected   []bool
	}{
		{
			name:    "bool variable recognized by name",
			varName: "longSignal",
			varType: "bool",
			checkNames: []string{
				"longSignal",
				"price",
			},
			expected: []bool{true, false},
		},
		{
			name:    "float64 variable not recognized as bool",
			varName: "sma",
			varType: "float64",
			checkNames: []string{
				"sma",
			},
			expected: []bool{false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := NewTypeInferenceEngine()
			engine.RegisterVariable(tt.varName, tt.varType)

			for i, name := range tt.checkNames {
				result := engine.IsBoolVariableByName(name)
				if result != tt.expected[i] {
					t.Errorf("IsBoolVariableByName(%q) expected %v, got %v", name, tt.expected[i], result)
				}
			}
		})
	}
}

func TestTypeInferenceEngine_EdgeCases(t *testing.T) {
	t.Run("nil expression returns float64", func(t *testing.T) {
		engine := NewTypeInferenceEngine()
		result := engine.InferType(nil)
		if result != "float64" {
			t.Errorf("expected float64 for nil expression, got %q", result)
		}
	})

	t.Run("unknown expression type returns float64", func(t *testing.T) {
		engine := NewTypeInferenceEngine()
		result := engine.InferType(&ast.Literal{Value: 42.0})
		if result != "float64" {
			t.Errorf("expected float64 for literal, got %q", result)
		}
	})

	t.Run("IsBoolVariableByName with unregistered variable returns false", func(t *testing.T) {
		engine := NewTypeInferenceEngine()
		result := engine.IsBoolVariableByName("nonexistent")
		if result {
			t.Error("expected false for unregistered variable")
		}
	})

	t.Run("IsBoolConstant with unregistered constant returns false", func(t *testing.T) {
		engine := NewTypeInferenceEngine()
		result := engine.IsBoolConstant("nonexistent")
		if result {
			t.Error("expected false for unregistered constant")
		}
	})

	t.Run("GetVariableType with unregistered variable returns not exists", func(t *testing.T) {
		engine := NewTypeInferenceEngine()
		_, exists := engine.GetVariableType("nonexistent")
		if exists {
			t.Error("expected not exists for unregistered variable")
		}
	})
}

func TestTypeInferenceEngine_MultipleVariables(t *testing.T) {
	engine := NewTypeInferenceEngine()

	engine.RegisterVariable("longCross", "bool")
	engine.RegisterVariable("shortCross", "bool")
	engine.RegisterVariable("sma20", "float64")
	engine.RegisterVariable("sma50", "float64")
	engine.RegisterConstant("enabled", true)
	engine.RegisterConstant("multiplier", 1.5)

	tests := []struct {
		name     string
		varName  string
		expected bool
	}{
		{"longCross is bool", "longCross", true},
		{"shortCross is bool", "shortCross", true},
		{"sma20 is not bool", "sma20", false},
		{"sma50 is not bool", "sma50", false},
		{"enabled const is bool", "enabled", true},
		{"multiplier const is not bool", "multiplier", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			varResult := engine.IsBoolVariableByName(tt.varName)
			constResult := engine.IsBoolConstant(tt.varName)
			result := varResult || constResult

			if result != tt.expected {
				t.Errorf("expected %v, got %v (var=%v, const=%v)", tt.expected, result, varResult, constResult)
			}
		})
	}
}
