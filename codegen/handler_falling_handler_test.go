package codegen

import (
	"fmt"
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestFallingHandler_TAFunctionHandlerInterface(t *testing.T) {
	handler := &FallingHandler{}

	t.Run("can_handle_ta_dot_falling", func(t *testing.T) {
		if !handler.CanHandle("ta.falling") {
			t.Error("FallingHandler must accept 'ta.falling'")
		}
	})

	t.Run("can_handle_falling", func(t *testing.T) {
		if !handler.CanHandle("falling") {
			t.Error("FallingHandler must accept 'falling' (v4 compatibility)")
		}
	})

	t.Run("cannot_handle_other_functions", func(t *testing.T) {
		for _, name := range []string{"ta.rising", "ta.sma", "ta.ema", ""} {
			if handler.CanHandle(name) {
				t.Errorf("FallingHandler must not accept %q", name)
			}
		}
	})
}

func TestFallingHandler_ArgumentValidation(t *testing.T) {
	handler := &FallingHandler{}
	gen := newTestGenerator()

	tests := []struct {
		name    string
		args    []ast.Expression
		wantErr bool
	}{
		{"no_args", []ast.Expression{}, true},
		{"one_arg_missing_period", []ast.Expression{&ast.Identifier{Name: "close"}}, true},
		{"two_args_valid", []ast.Expression{&ast.Identifier{Name: "close"}, &ast.Literal{Value: 3.0}}, false},
		{"two_args_with_hl2_source", []ast.Expression{&ast.Identifier{Name: "hl2"}, &ast.Literal{Value: 5.0}}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			call := &ast.CallExpression{Arguments: tt.args}
			_, err := handler.GenerateCode(gen, "x", call)
			if tt.wantErr && err == nil {
				t.Error("Expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestFallingHandler_WarmupBehavior(t *testing.T) {
	handler := &FallingHandler{}

	tests := []struct {
		name           string
		period         float64
		wantWarmupEdge int
	}{
		{"period_1", 1, 1},
		{"period_3", 3, 3},
		{"period_14", 14, 14},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := newTestGenerator()
			call := &ast.CallExpression{Arguments: []ast.Expression{
				&ast.Identifier{Name: "close"},
				&ast.Literal{Value: tt.period},
			}}

			code, err := handler.GenerateCode(gen, "f", call)
			if err != nil {
				t.Fatalf("GenerateCode() error = %v", err)
			}

			expectedCheck := fmt.Sprintf("ctx.BarIndex < %d", tt.wantWarmupEdge)
			if !strings.Contains(code, expectedCheck) {
				t.Errorf("Warmup check %q not found\nGot:\n%s", expectedCheck, code)
			}
		})
	}
}

func TestFallingHandler_AlgorithmCorrectness(t *testing.T) {
	handler := &FallingHandler{}
	gen := newTestGenerator()

	call := &ast.CallExpression{Arguments: []ast.Expression{
		&ast.Identifier{Name: "close"},
		&ast.Literal{Value: 3.0},
	}}

	code, err := handler.GenerateCode(gen, "f", call)
	if err != nil {
		t.Fatalf("GenerateCode() error = %v", err)
	}

	t.Run("violation_operator_is_greater_or_equal", func(t *testing.T) {
		if !strings.Contains(code, ">= prev") {
			t.Errorf("Falling must use '>= prev' as violation operator\nGot:\n%s", code)
		}
		if strings.Contains(code, "<= prev") {
			t.Errorf("Falling must not use '<= prev' (that is rising)\nGot:\n%s", code)
		}
	})

	t.Run("returns_zero_on_violation", func(t *testing.T) {
		if !strings.Contains(code, "return 0.0") {
			t.Errorf("Falling must return 0.0 when violation found\nGot:\n%s", code)
		}
	})

	t.Run("returns_one_when_all_pairs_pass", func(t *testing.T) {
		if !strings.Contains(code, "return 1.0") {
			t.Errorf("Falling must return 1.0 after all pairs verified\nGot:\n%s", code)
		}
	})

	t.Run("uses_curr_prev_variable_naming", func(t *testing.T) {
		if !strings.Contains(code, "curr") || !strings.Contains(code, "prev") {
			t.Errorf("Missing curr/prev variable names\nGot:\n%s", code)
		}
	})

	t.Run("loop_iterates_over_period_pairs", func(t *testing.T) {
		if !strings.Contains(code, "for j := 0; j < 3") {
			t.Errorf("Loop must iterate j from 0 to period\nGot:\n%s", code)
		}
	})
}

func TestFallingHandler_IIFEStructure(t *testing.T) {
	handler := &FallingHandler{}
	gen := newTestGenerator()

	call := &ast.CallExpression{Arguments: []ast.Expression{
		&ast.Identifier{Name: "close"},
		&ast.Literal{Value: 2.0},
	}}

	code, err := handler.GenerateCode(gen, "f", call)
	if err != nil {
		t.Fatalf("GenerateCode() error = %v", err)
	}

	t.Run("iife_wrapper", func(t *testing.T) {
		if !strings.Contains(code, "func() float64") {
			t.Errorf("Missing IIFE wrapper 'func() float64'\nGot:\n%s", code)
		}
		if !strings.Contains(code, "}())") {
			t.Errorf("Missing IIFE invocation '}())'\nGot:\n%s", code)
		}
	})

	t.Run("result_stored_in_series", func(t *testing.T) {
		if !strings.Contains(code, "fSeries.Set(") {
			t.Errorf("Missing 'fSeries.Set(' assignment\nGot:\n%s", code)
		}
	})

	t.Run("nan_before_warmup", func(t *testing.T) {
		if !strings.Contains(code, "math.NaN()") {
			t.Errorf("Missing NaN assignment before warmup period\nGot:\n%s", code)
		}
	})
}

func TestFallingHandler_SourceExpressions(t *testing.T) {
	handler := &FallingHandler{}

	tests := []struct {
		name   string
		source ast.Expression
	}{
		{"identifier_close", &ast.Identifier{Name: "close"}},
		{"identifier_hl2", &ast.Identifier{Name: "hl2"}},
		{"binary_expression", &ast.BinaryExpression{
			Left:     &ast.Identifier{Name: "high"},
			Operator: "+",
			Right:    &ast.Identifier{Name: "low"},
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := newTestGenerator()
			call := &ast.CallExpression{Arguments: []ast.Expression{
				tt.source,
				&ast.Literal{Value: 2.0},
			}}

			code, err := handler.GenerateCode(gen, "f", call)
			if err != nil {
				t.Fatalf("GenerateCode() error = %v", err)
			}

			if !strings.Contains(code, "fSeries.Set(") {
				t.Errorf("Missing series assignment\nGot:\n%s", code)
			}
		})
	}
}

func TestRisingFallingHandlers_ViolationOperatorsDiffer(t *testing.T) {
	/* Rising and falling must use opposite operators — catches copy-paste bugs */
	risingHandler := &RisingHandler{}
	fallingHandler := &FallingHandler{}
	gen := newTestGenerator()

	call := &ast.CallExpression{Arguments: []ast.Expression{
		&ast.Identifier{Name: "close"},
		&ast.Literal{Value: 3.0},
	}}

	risingCode, _ := risingHandler.GenerateCode(gen, "r", call)
	gen2 := newTestGenerator()
	fallingCode, _ := fallingHandler.GenerateCode(gen2, "f", call)

	if risingCode == fallingCode {
		t.Error("Rising and falling must produce different code (different violation operators)")
	}

	if !strings.Contains(risingCode, "<= prev") {
		t.Error("Rising must contain '<= prev'")
	}
	if !strings.Contains(fallingCode, ">= prev") {
		t.Error("Falling must contain '>= prev'")
	}
}
