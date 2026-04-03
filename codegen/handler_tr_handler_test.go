package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestTrHandler_TAFunctionHandlerInterface(t *testing.T) {
	handler := &TrHandler{}

	t.Run("can_handle_ta_dot_tr", func(t *testing.T) {
		if !handler.CanHandle("ta.tr") {
			t.Error("TrHandler must accept 'ta.tr'")
		}
	})

	t.Run("can_handle_bare_tr", func(t *testing.T) {
		if !handler.CanHandle("tr") {
			t.Error("TrHandler must accept 'tr' (v4 compatibility)")
		}
	})

	t.Run("cannot_handle_other_functions", func(t *testing.T) {
		for _, name := range []string{"ta.atr", "ta.rsi", "ta.ema", "ta.kcw", ""} {
			if handler.CanHandle(name) {
				t.Errorf("TrHandler must not accept %q", name)
			}
		}
	})
}

// TestTrHandler_HandleNAArgExtraction verifies that extractHandleNAArg correctly
// parses the boolean argument in all forms: omitted, false, true, non-literal.
func TestTrHandler_HandleNAArgExtraction(t *testing.T) {
	tests := []struct {
		name     string
		args     []ast.Expression
		expected bool
	}{
		{
			name:     "no_args_defaults_to_false",
			args:     nil,
			expected: false,
		},
		{
			name:     "explicit_false",
			args:     []ast.Expression{&ast.Literal{Value: false}},
			expected: false,
		},
		{
			name:     "explicit_true",
			args:     []ast.Expression{&ast.Literal{Value: true}},
			expected: true,
		},
		{
			name:     "non_literal_treated_as_false",
			args:     []ast.Expression{&ast.Identifier{Name: "someVar"}},
			expected: false,
		},
		{
			name:     "numeric_literal_treated_as_false",
			args:     []ast.Expression{&ast.Literal{Value: 1.0}},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			call := &ast.CallExpression{Arguments: tt.args}
			got := extractHandleNAArg(call)
			if got != tt.expected {
				t.Errorf("extractHandleNAArg() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestTrHandler_FirstBarPolicy verifies that generateTRExpression produces code
// with the correct bar-0 behaviour for each handleNA value:
//   - handleNA=false → math.NaN() on bar 0  (ta.tr variable semantics)
//   - handleNA=true  → High-Low on bar 0    (ATR-seed / ta.tr(true) semantics)
func TestTrHandler_FirstBarPolicy(t *testing.T) {
	tests := []struct {
		name          string
		handleNA      bool
		wantFirstBar  string
		wantNotInCode string
	}{
		{
			name:         "handleNA_false_emits_NaN_on_bar0",
			handleNA:     false,
			wantFirstBar: "math.NaN()",
		},
		{
			name:         "handleNA_true_emits_HL_on_bar0",
			handleNA:     true,
			wantFirstBar: "bar.High - bar.Low",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := generateTRExpression(tt.handleNA)

			if !strings.Contains(code, "ctx.BarIndex < 1") {
				t.Errorf("generated code missing first-bar guard 'ctx.BarIndex < 1'\nGot: %s", code)
			}

			if !strings.Contains(code, tt.wantFirstBar) {
				t.Errorf("first-bar body missing %q\nGot: %s", tt.wantFirstBar, code)
			}
		})
	}
}

// TestTrHandler_NormalBarFormula verifies that the generated code always
// includes the three-component true-range formula for non-first bars,
// regardless of the handleNA setting.
func TestTrHandler_NormalBarFormula(t *testing.T) {
	for _, handleNA := range []bool{false, true} {
		handleNA := handleNA
		name := "handleNA_false"
		if handleNA {
			name = "handleNA_true"
		}
		t.Run(name, func(t *testing.T) {
			code := generateTRExpression(handleNA)

			required := []string{
				"math.Max",
				"math.Abs",
				"prevClose",
				"bar.High",
				"bar.Low",
			}
			for _, want := range required {
				if !strings.Contains(code, want) {
					t.Errorf("normal-bar formula missing %q\nGot: %s", want, code)
				}
			}
		})
	}
}

// TestTrHandler_CodeGeneration verifies that TrHandler.GenerateCode produces
// a series assignment wrapping the correct inline IIFE.
func TestTrHandler_CodeGeneration(t *testing.T) {
	handler := &TrHandler{}
	gen := newTestGenerator()

	tests := []struct {
		name         string
		args         []ast.Expression
		wantContains []string
	}{
		{
			name:         "no_args_stores_NaN_on_bar0",
			args:         nil,
			wantContains: []string{"trSeries.Set(", "math.NaN()", "func() float64"},
		},
		{
			name:         "handleNA_false_stores_NaN_on_bar0",
			args:         []ast.Expression{&ast.Literal{Value: false}},
			wantContains: []string{"trSeries.Set(", "math.NaN()", "func() float64"},
		},
		{
			name:         "handleNA_true_stores_HL_on_bar0",
			args:         []ast.Expression{&ast.Literal{Value: true}},
			wantContains: []string{"trSeries.Set(", "bar.High - bar.Low", "func() float64"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			call := &ast.CallExpression{Arguments: tt.args}
			code, err := handler.GenerateCode(gen, "tr", call)
			if err != nil {
				t.Fatalf("GenerateCode() error = %v", err)
			}

			for _, want := range tt.wantContains {
				if !strings.Contains(code, want) {
					t.Errorf("expected %q in generated code\nGot: %s", want, code)
				}
			}
		})
	}
}
