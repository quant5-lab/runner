package codegen

import (
	"fmt"
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestMomHandler_TAFunctionHandlerInterface(t *testing.T) {
	handler := &MomHandler{}

	t.Run("can_handle_ta_dot_mom", func(t *testing.T) {
		if !handler.CanHandle("ta.mom") {
			t.Error("MomHandler must accept 'ta.mom'")
		}
	})

	t.Run("can_handle_mom", func(t *testing.T) {
		if !handler.CanHandle("mom") {
			t.Error("MomHandler must accept 'mom' (v4 compatibility)")
		}
	})

	t.Run("cannot_handle_other_functions", func(t *testing.T) {
		for _, name := range []string{"ta.roc", "ta.cmo", "ta.rsi", ""} {
			if handler.CanHandle(name) {
				t.Errorf("MomHandler must not accept %q", name)
			}
		}
	})
}

func TestMomHandler_ArgumentValidation(t *testing.T) {
	handler := &MomHandler{}
	gen := newTestGenerator()

	tests := []struct {
		name    string
		args    []ast.Expression
		wantErr bool
	}{
		{"no_args", []ast.Expression{}, true},
		{"one_arg_missing_period", []ast.Expression{&ast.Identifier{Name: "close"}}, true},
		{"two_args_valid", []ast.Expression{&ast.Identifier{Name: "close"}, &ast.Literal{Value: 10.0}}, false},
		{"two_args_with_hlc3_source", []ast.Expression{&ast.Identifier{Name: "hlc3"}, &ast.Literal{Value: 5.0}}, false},
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

func TestMomHandler_WarmupBehavior(t *testing.T) {
	handler := &MomHandler{}

	tests := []struct {
		name           string
		period         float64
		wantWarmupEdge int
	}{
		{"period_1", 1, 1},
		{"period_10", 10, 10},
		{"period_14", 14, 14},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := newTestGenerator()
			call := &ast.CallExpression{Arguments: []ast.Expression{
				&ast.Identifier{Name: "close"},
				&ast.Literal{Value: tt.period},
			}}

			code, err := handler.GenerateCode(gen, "m", call)
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

func TestMomHandler_AlgorithmCorrectness(t *testing.T) {
	handler := &MomHandler{}
	gen := newTestGenerator()

	call := &ast.CallExpression{Arguments: []ast.Expression{
		&ast.Identifier{Name: "close"},
		&ast.Literal{Value: 10.0},
	}}

	code, err := handler.GenerateCode(gen, "m", call)
	if err != nil {
		t.Fatalf("GenerateCode() error = %v", err)
	}

	t.Run("subtraction_formula", func(t *testing.T) {
		if !strings.Contains(code, " - ") {
			t.Errorf("Mom must compute subtraction (source - source[period])\nGot:\n%s", code)
		}
	})

	t.Run("no_division", func(t *testing.T) {
		/* mom is raw difference — not normalized like roc */
		if strings.Contains(code, " / ") {
			t.Errorf("Mom must not divide (use roc for percentage change)\nGot:\n%s", code)
		}
	})

	t.Run("no_loop", func(t *testing.T) {
		/* mom is O(1) — single subtraction, no window iteration */
		if strings.Contains(code, "for j") {
			t.Errorf("Mom must not use loops (it is a direct difference)\nGot:\n%s", code)
		}
	})

	t.Run("nan_before_warmup", func(t *testing.T) {
		if !strings.Contains(code, "math.NaN()") {
			t.Errorf("Mom must return NaN before warmup period\nGot:\n%s", code)
		}
	})

	t.Run("result_stored_in_series", func(t *testing.T) {
		if !strings.Contains(code, "mSeries.Set(") {
			t.Errorf("Missing 'mSeries.Set(' assignment\nGot:\n%s", code)
		}
	})
}

func TestMomHandler_SourceExpressions(t *testing.T) {
	handler := &MomHandler{}

	tests := []struct {
		name   string
		source ast.Expression
	}{
		{"identifier_close", &ast.Identifier{Name: "close"}},
		{"identifier_hlc3", &ast.Identifier{Name: "hlc3"}},
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
				&ast.Literal{Value: 5.0},
			}}

			code, err := handler.GenerateCode(gen, "m", call)
			if err != nil {
				t.Fatalf("GenerateCode() error = %v", err)
			}

			if !strings.Contains(code, " - ") {
				t.Errorf("Missing subtraction in generated code\nGot:\n%s", code)
			}
		})
	}
}
