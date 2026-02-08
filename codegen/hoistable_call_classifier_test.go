package codegen

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
)

/* Validates TA functions classified as hoistable */
func TestHoistableCallClassifier_TAFunctions(t *testing.T) {
	gen := newTestGenerator()
	classifier := NewHoistableCallClassifier(gen)

	tests := []struct {
		name     string
		funcName string
		expected bool
	}{
		{"ta.sma", "sma", true},
		{"ta.ema", "ema", true},
		{"ta.rsi", "rsi", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			call := &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ta"},
					Property: &ast.Identifier{Name: tt.funcName},
				},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
					&ast.Literal{Value: float64(14)},
				},
			}

			if result := classifier.IsHoistable(call); result != tt.expected {
				t.Errorf("IsHoistable(ta.%s) = %v, expected %v", tt.funcName, result, tt.expected)
			}
		})
	}
}

/* Validates stateful value functions classified as hoistable */
func TestHoistableCallClassifier_StatefulValueFunctions(t *testing.T) {
	gen := newTestGenerator()
	classifier := NewHoistableCallClassifier(gen)

	tests := []struct {
		name     string
		funcName string
		expected bool
	}{
		{"fixnan", "fixnan", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			call := &ast.CallExpression{
				Callee: &ast.Identifier{Name: tt.funcName},
			}

			if result := classifier.IsHoistable(call); result != tt.expected {
				t.Errorf("IsHoistable(%s) = %v, expected %v", tt.funcName, result, tt.expected)
			}
		})
	}
}

/* Validates IsStatefulValueFunction exported method */
func TestHoistableCallClassifier_IsStatefulValueFunction(t *testing.T) {
	gen := newTestGenerator()
	classifier := NewHoistableCallClassifier(gen)

	tests := []struct {
		name     string
		funcName string
		expected bool
	}{
		{"fixnan is stateful", "fixnan", true},
		{"ta.sma not stateful value", "ta.sma", false},
		{"nz not stateful value", "nz", false},
		{"math.abs not stateful value", "math.abs", false},
		{"unknown not stateful value", "unknownFunc", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if result := classifier.IsStatefulValueFunction(tt.funcName); result != tt.expected {
				t.Errorf("IsStatefulValueFunction(%s) = %v, expected %v", tt.funcName, result, tt.expected)
			}
		})
	}
}

/* Validates non-hoistable functions rejected */
func TestHoistableCallClassifier_NonHoistable(t *testing.T) {
	gen := newTestGenerator()
	classifier := NewHoistableCallClassifier(gen)

	tests := []struct {
		name     string
		funcName string
		expected bool
	}{
		{"math.abs", "abs", false},
		{"math.max", "max", false},
		{"plot", "plot", false},
		{"plotshape", "plotshape", false},
		{"strategy.entry", "entry", false},
		{"nz", "nz", false},
		{"na", "na", false},
		{"user function", "myCustomFunc", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			call := &ast.CallExpression{
				Callee: &ast.Identifier{Name: tt.funcName},
			}

			if result := classifier.IsHoistable(call); result != tt.expected {
				t.Errorf("IsHoistable(%s) = %v, expected %v", tt.funcName, result, tt.expected)
			}
		})
	}
}

/* Validates boundary conditions */
func TestHoistableCallClassifier_BoundaryConditions(t *testing.T) {
	gen := newTestGenerator()
	classifier := NewHoistableCallClassifier(gen)

	tests := []struct {
		name     string
		funcName string
		expected bool
	}{
		{"empty string", "", false},
		{"single character", "x", false},
		{"very long name", "thisIsAVeryLongFunctionNameThatDoesNotExist", false},
		{"numeric only", "12345", false},
		{"special characters", "@#$%", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			call := &ast.CallExpression{
				Callee: &ast.Identifier{Name: tt.funcName},
			}

			if result := classifier.IsHoistable(call); result != tt.expected {
				t.Errorf("IsHoistable(%s) = %v, expected %v", tt.funcName, result, tt.expected)
			}
		})
	}
}

/* Validates case sensitivity */
func TestHoistableCallClassifier_CaseSensitivity(t *testing.T) {
	gen := newTestGenerator()
	classifier := NewHoistableCallClassifier(gen)

	tests := []struct {
		name     string
		funcName string
		expected bool
	}{
		{"lowercase fixnan", "fixnan", true},
		{"uppercase FIXNAN", "FIXNAN", false},
		{"mixed FixNan", "FixNan", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			call := &ast.CallExpression{
				Callee: &ast.Identifier{Name: tt.funcName},
			}

			if result := classifier.IsHoistable(call); result != tt.expected {
				t.Errorf("IsHoistable(%s) = %v, expected %v", tt.funcName, result, tt.expected)
			}
		})
	}
}

/* Validates extensibility of stateful value functions */
func TestHoistableCallClassifier_CustomStatefulFunctions(t *testing.T) {
	gen := newTestGenerator()

	customStateful := map[string]bool{
		"fixnan":       true,
		"customState1": true,
		"customState2": true,
	}

	classifier := HoistableCallClassifier{
		gen:                    gen,
		inlineTARegistry:       NewInlineTAIIFERegistry(),
		statefulValueFunctions: customStateful,
	}

	tests := []struct {
		name     string
		funcName string
		expected bool
	}{
		{"default fixnan", "fixnan", true},
		{"custom state 1", "customState1", true},
		{"custom state 2", "customState2", true},
		{"non-stateful", "regularFunc", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if result := classifier.IsStatefulValueFunction(tt.funcName); result != tt.expected {
				t.Errorf("IsStatefulValueFunction(%s) = %v, expected %v", tt.funcName, result, tt.expected)
			}
		})
	}
}

/* Validates multiple calls to same function */
func TestHoistableCallClassifier_ConsistencyAcrossMultipleCalls(t *testing.T) {
	gen := newTestGenerator()
	classifier := NewHoistableCallClassifier(gen)

	call := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "ta"},
			Property: &ast.Identifier{Name: "sma"},
		},
	}

	firstResult := classifier.IsHoistable(call)
	for i := 0; i < 100; i++ {
		result := classifier.IsHoistable(call)
		if result != firstResult {
			t.Errorf("Inconsistent result at iteration %d: got %v, expected %v", i, result, firstResult)
		}
	}
}

/* Validates TA functions with member expression callee */
func TestHoistableCallClassifier_MemberExpressionCallee(t *testing.T) {
	gen := newTestGenerator()
	classifier := NewHoistableCallClassifier(gen)

	call := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "ta"},
			Property: &ast.Identifier{Name: "sma"},
		},
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "close"},
			&ast.Literal{Value: float64(14)},
		},
	}

	if !classifier.IsHoistable(call) {
		t.Error("Expected ta.sma with valid period to be hoistable")
	}
}
