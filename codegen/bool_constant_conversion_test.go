package codegen

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
)

/* Validates typeBasedRule identifies bool constants and skips conversion */
func TestBoolConstantConversionRuleLogic(t *testing.T) {
	engine := NewTypeInferenceEngine()

	engine.RegisterConstant("show_trades", true)
	engine.RegisterVariable("signal", "bool")

	rule := NewTypeBasedRule(engine)

	tests := []struct {
		name          string
		identifier    string
		shouldConvert bool
		description   string
	}{
		{
			name:          "bool constant no conversion",
			identifier:    "show_trades",
			shouldConvert: false,
			description:   "Bool constants from input.bool are already bool",
		},
		{
			name:          "bool variable no conversion",
			identifier:    "signal",
			shouldConvert: false,
			description:   "Bool variables are already bool type",
		},
		{
			name:          "unknown identifier no conversion",
			identifier:    "unknown",
			shouldConvert: false,
			description:   "Unknown identifiers conservative - don't convert",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr := &ast.Identifier{Name: tt.identifier}
			result := rule.ShouldConvert(expr, tt.identifier)

			if result != tt.shouldConvert {
				t.Errorf("%s: expected ShouldConvert=%v, got %v",
					tt.description, tt.shouldConvert, result)
			}
		})
	}
}

/* Validates IsBoolConstant differentiates bool from other types */
func TestBoolConstantDetection(t *testing.T) {
	tests := []struct {
		name            string
		setupFunc       func(*TypeInferenceEngine)
		identifier      string
		expectBoolConst bool
		description     string
	}{
		{
			name: "true bool is bool constant",
			setupFunc: func(engine *TypeInferenceEngine) {
				engine.RegisterConstant("enabled", true)
			},
			identifier:      "enabled",
			expectBoolConst: true,
			description:     "True bool constant detected",
		},
		{
			name: "false bool is bool constant",
			setupFunc: func(engine *TypeInferenceEngine) {
				engine.RegisterConstant("disabled", false)
			},
			identifier:      "disabled",
			expectBoolConst: true,
			description:     "False bool constant detected",
		},
		{
			name: "nil value not bool constant",
			setupFunc: func(engine *TypeInferenceEngine) {
				engine.RegisterConstant("nullable", nil)
			},
			identifier:      "nullable",
			expectBoolConst: false,
			description:     "Nil constants are not bool",
		},
		{
			name: "float constant not bool",
			setupFunc: func(engine *TypeInferenceEngine) {
				engine.RegisterConstant("threshold", 0.5)
			},
			identifier:      "threshold",
			expectBoolConst: false,
			description:     "Float constants are not bool",
		},
		{
			name: "int constant not bool",
			setupFunc: func(engine *TypeInferenceEngine) {
				engine.RegisterConstant("count", 10)
			},
			identifier:      "count",
			expectBoolConst: false,
			description:     "Int constants are not bool",
		},
		{
			name: "string constant not bool",
			setupFunc: func(engine *TypeInferenceEngine) {
				engine.RegisterConstant("direction", "long")
			},
			identifier:      "direction",
			expectBoolConst: false,
			description:     "String constants are not bool",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := NewTypeInferenceEngine()
			tt.setupFunc(engine)

			result := engine.IsBoolConstant(tt.identifier)

			if result != tt.expectBoolConst {
				t.Errorf("%s: expected IsBoolConstant=%v, got %v",
					tt.description, tt.expectBoolConst, result)
			}
		})
	}
}

/* Validates addBoolConversionIfNeeded skips conversion for bool constants */
func TestAddBoolConversionIfNeeded(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func(*generator)
		expr        ast.Expression
		code        string
		expectCode  string
		description string
	}{
		{
			name: "bool constant no conversion",
			setupFunc: func(g *generator) {
				g.typeSystem.RegisterConstant("show_trades", true)
			},
			expr:        &ast.Identifier{Name: "show_trades"},
			code:        "show_trades",
			expectCode:  "show_trades",
			description: "Bool constant used directly without != 0",
		},
		{
			name: "bool variable gets conversion",
			setupFunc: func(g *generator) {
				g.typeSystem.RegisterVariable("signal", "bool")
			},
			expr:        &ast.Identifier{Name: "signal"},
			code:        "signalSeries.GetCurrent()",
			expectCode:  "value.IsTrue(signalSeries.GetCurrent())",
			description: "Bool variable gets != 0 conversion",
		},
		{
			name: "comparison already has operator",
			setupFunc: func(g *generator) {
				/* No registration needed */
			},
			expr: &ast.BinaryExpression{
				Operator: ">",
				Left:     &ast.Identifier{Name: "close"},
				Right:    &ast.Identifier{Name: "open"},
			},
			code:        "closeSeries.GetCurrent() > openSeries.GetCurrent()",
			expectCode:  "closeSeries.GetCurrent() > openSeries.GetCurrent()",
			description: "Comparison expressions skip conversion",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &generator{
				variables:     make(map[string]string),
				constants:     make(map[string]interface{}),
				typeSystem:    NewTypeInferenceEngine(),
				boolConverter: NewBooleanConverter(NewTypeInferenceEngine()),
			}

			/* Re-create boolConverter with same typeSystem instance */
			g.boolConverter = NewBooleanConverter(g.typeSystem)

			tt.setupFunc(g)

			result := g.addBoolConversionIfNeeded(tt.expr, tt.code)

			if result != tt.expectCode {
				t.Errorf("%s:\nexpected: %s\ngot:      %s",
					tt.description, tt.expectCode, result)
			}
		})
	}
}
