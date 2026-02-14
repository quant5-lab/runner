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

/* Validates security functions classified as hoistable across callee forms */
func TestHoistableCallClassifier_SecurityFunctions(t *testing.T) {
	gen := newTestGenerator()
	classifier := NewHoistableCallClassifier(gen)

	tests := []struct {
		name     string
		callee   ast.Expression
		expected bool
	}{
		{"bare security", &ast.Identifier{Name: "security"}, true},
		{"namespaced request.security", &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "request"},
			Property: &ast.Identifier{Name: "security"},
		}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			call := &ast.CallExpression{Callee: tt.callee}
			if result := classifier.IsHoistable(call); result != tt.expected {
				funcName := gen.extractFunctionName(tt.callee)
				t.Errorf("IsHoistable(%s) = %v, want %v", funcName, result, tt.expected)
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

/* Validates TA functions with registry handler but no IIFE generator */
func TestHoistableCallClassifier_TAFunctionRegistryGate(t *testing.T) {
	gen := newTestGenerator()
	classifier := NewHoistableCallClassifier(gen)

	tests := []struct {
		name     string
		call     *ast.CallExpression
		expected bool
	}{
		{
			name: "ta.dev with literal period",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ta"},
					Property: &ast.Identifier{Name: "dev"},
				},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
					&ast.Literal{Value: float64(20)},
				},
			},
			expected: true,
		},
		{
			name: "bare dev with literal period",
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "dev"},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
					&ast.Literal{Value: float64(14)},
				},
			},
			expected: true,
		},
		{
			name: "ta.crossover with identifier args",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ta"},
					Property: &ast.Identifier{Name: "crossover"},
				},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
					&ast.Identifier{Name: "open"},
				},
			},
			expected: false,
		},
		{
			name: "ta.valuewhen with literal second arg",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ta"},
					Property: &ast.Identifier{Name: "valuewhen"},
				},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "condition"},
					&ast.Literal{Value: float64(1)},
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if result := classifier.IsHoistable(tt.call); result != tt.expected {
				t.Errorf("IsHoistable() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestHoistableCallClassifier_UnimplementedFunctions(t *testing.T) {
	gen := newTestGenerator()
	classifier := NewHoistableCallClassifier(gen)

	tests := []struct {
		name string
		call *ast.CallExpression
	}{
		{
			name: "ta.macd",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ta"},
					Property: &ast.Identifier{Name: "macd"},
				},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
					&ast.Literal{Value: float64(12)},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if classifier.IsHoistable(tt.call) {
				t.Error("Expected function without handler to NOT be hoistable")
			}
		})
	}
}

/* Validates classifier degrades gracefully with nil TAFunctionRegistry */
func TestHoistableCallClassifier_NilTAFunctionRegistry(t *testing.T) {
	gen := newTestGenerator()

	classifier := HoistableCallClassifier{
		gen:                    gen,
		inlineTARegistry:       NewInlineTAIIFERegistry(),
		taFunctionRegistry:     nil,
		statefulValueFunctions: defaultStatefulValueFunctions,
	}

	tests := []struct {
		name     string
		call     *ast.CallExpression
		expected bool
	}{
		{
			name: "IIFE-registered function still hoistable",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ta"},
					Property: &ast.Identifier{Name: "sma"},
				},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
					&ast.Literal{Value: float64(14)},
				},
			},
			expected: true,
		},
		{
			name: "stateful function still hoistable",
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "fixnan"},
			},
			expected: true,
		},
		{
			name: "registry-only function NOT hoistable when registry nil",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ta"},
					Property: &ast.Identifier{Name: "dev"},
				},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
					&ast.Literal{Value: float64(20)},
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if result := classifier.IsHoistable(tt.call); result != tt.expected {
				t.Errorf("IsHoistable() = %v, expected %v (nil TAFunctionRegistry)", result, tt.expected)
			}
		})
	}
}

/* Dynamic period must only hoist functions with codegen support — others silently miscompile */
func TestHoistableCallClassifier_PeriodTypeHoistability(t *testing.T) {
	gen := newTestGenerator()
	classifier := NewHoistableCallClassifier(gen)

	closeSrc := &ast.Identifier{Name: "close"}
	literalPeriod := &ast.Literal{Value: float64(14)}
	dynamicPeriod := &ast.Identifier{Name: "dynamicLength"}
	binaryDynamic := &ast.BinaryExpression{
		Left:     &ast.Identifier{Name: "baseLen"},
		Operator: "*",
		Right:    &ast.Literal{Value: float64(2)},
	}

	makeTACall := func(prop string, args ...ast.Expression) *ast.CallExpression {
		return &ast.CallExpression{
			Callee: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "ta"},
				Property: &ast.Identifier{Name: prop},
			},
			Arguments: args,
		}
	}

	makeBareCall := func(name string, args ...ast.Expression) *ast.CallExpression {
		return &ast.CallExpression{
			Callee:    &ast.Identifier{Name: name},
			Arguments: args,
		}
	}

	tests := []struct {
		name     string
		call     *ast.CallExpression
		expected bool
	}{
		/* Dynamic-period-supported functions: hoistable with any period form */
		{"ta.sma literal period", makeTACall("sma", closeSrc, literalPeriod), true},
		{"ta.sma dynamic identifier", makeTACall("sma", closeSrc, dynamicPeriod), true},
		{"ta.sma dynamic binary expr", makeTACall("sma", closeSrc, binaryDynamic), true},
		{"ta.ema dynamic period", makeTACall("ema", closeSrc, dynamicPeriod), true},
		{"ta.rsi dynamic period", makeTACall("rsi", closeSrc, dynamicPeriod), true},
		{"ta.stdev dynamic period", makeTACall("stdev", closeSrc, dynamicPeriod), true},
		{"ta.highest dynamic period", makeTACall("highest", closeSrc, dynamicPeriod), true},
		{"ta.lowest dynamic period", makeTACall("lowest", closeSrc, dynamicPeriod), true},

		/* ATR special case: single argument period extraction */
		{"ta.atr single dynamic arg", makeTACall("atr", dynamicPeriod), true},
		{"ta.atr single literal arg", makeTACall("atr", literalPeriod), true},

		/* Unsupported dynamic period: literal ok, dynamic blocked */
		{"ta.wma literal period", makeTACall("wma", closeSrc, literalPeriod), true},
		{"ta.wma dynamic period", makeTACall("wma", closeSrc, dynamicPeriod), false},
		{"ta.rma dynamic period", makeTACall("rma", closeSrc, dynamicPeriod), false},
		{"ta.linreg dynamic period", makeTACall("linreg", closeSrc, dynamicPeriod), false},
		{"ta.sum dynamic period", makeTACall("sum", closeSrc, dynamicPeriod), false},

		/* Bare form parity with namespaced */
		{"bare sma dynamic period", makeBareCall("sma", closeSrc, dynamicPeriod), true},
		{"bare ema dynamic period", makeBareCall("ema", closeSrc, dynamicPeriod), true},
		{"bare wma dynamic period", makeBareCall("wma", closeSrc, dynamicPeriod), false},
		{"bare rma dynamic period", makeBareCall("rma", closeSrc, dynamicPeriod), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if result := classifier.IsHoistable(tt.call); result != tt.expected {
				t.Errorf("IsHoistable() = %v, expected %v", result, tt.expected)
			}
		})
	}
}
