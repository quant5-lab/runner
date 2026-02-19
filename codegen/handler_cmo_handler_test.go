package codegen

import (
	"fmt"
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestCmoHandler_TAFunctionHandlerInterface(t *testing.T) {
	handler := &CmoHandler{}

	t.Run("can_handle_ta_dot_cmo", func(t *testing.T) {
		if !handler.CanHandle("ta.cmo") {
			t.Error("CmoHandler must accept 'ta.cmo'")
		}
	})

	t.Run("can_handle_cmo", func(t *testing.T) {
		if !handler.CanHandle("cmo") {
			t.Error("CmoHandler must accept 'cmo' (v4 compatibility)")
		}
	})

	t.Run("cannot_handle_other_functions", func(t *testing.T) {
		for _, name := range []string{"ta.roc", "ta.rsi", "ta.mom", ""} {
			if handler.CanHandle(name) {
				t.Errorf("CmoHandler must not accept %q", name)
			}
		}
	})
}

func TestCmoHandler_ArgumentValidation(t *testing.T) {
	handler := &CmoHandler{}
	gen := newTestGenerator()

	tests := []struct {
		name    string
		args    []ast.Expression
		wantErr bool
	}{
		{"no_args", []ast.Expression{}, true},
		{"one_arg_missing_period", []ast.Expression{&ast.Identifier{Name: "close"}}, true},
		{"two_args_valid", []ast.Expression{&ast.Identifier{Name: "close"}, &ast.Literal{Value: 9.0}}, false},
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

func TestCmoHandler_WarmupBehavior(t *testing.T) {
	handler := &CmoHandler{}

	tests := []struct {
		name           string
		period         float64
		wantWarmupEdge int
	}{
		{"period_1", 1, 1},
		{"period_9", 9, 9},
		{"period_14", 14, 14},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := newTestGenerator()
			call := &ast.CallExpression{Arguments: []ast.Expression{
				&ast.Identifier{Name: "close"},
				&ast.Literal{Value: tt.period},
			}}

			code, err := handler.GenerateCode(gen, "c", call)
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

func TestCmoHandler_AlgorithmCorrectness(t *testing.T) {
	handler := &CmoHandler{}
	gen := newTestGenerator()

	call := &ast.CallExpression{Arguments: []ast.Expression{
		&ast.Identifier{Name: "close"},
		&ast.Literal{Value: 9.0},
	}}

	code, err := handler.GenerateCode(gen, "c", call)
	if err != nil {
		t.Fatalf("GenerateCode() error = %v", err)
	}

	t.Run("up_accumulator", func(t *testing.T) {
		if !strings.Contains(code, "_c_up") {
			t.Errorf("CMO must use up accumulator variable\nGot:\n%s", code)
		}
	})

	t.Run("down_accumulator", func(t *testing.T) {
		if !strings.Contains(code, "_c_down") {
			t.Errorf("CMO must use down accumulator variable\nGot:\n%s", code)
		}
	})

	t.Run("total_sum", func(t *testing.T) {
		if !strings.Contains(code, "_c_total") {
			t.Errorf("CMO must compute total = up + down\nGot:\n%s", code)
		}
	})

	t.Run("zero_total_returns_zero_not_nan", func(t *testing.T) {
		/* When all bars unchanged, total=0 → return 0.0 (not NaN — defined behavior) */
		if !strings.Contains(code, "0.0") {
			t.Errorf("CMO must return 0.0 when total is zero\nGot:\n%s", code)
		}
	})

	t.Run("percentage_multiplier", func(t *testing.T) {
		if !strings.Contains(code, "100.0") {
			t.Errorf("CMO must multiply by 100.0 for percentage result\nGot:\n%s", code)
		}
	})

	t.Run("loop_over_period", func(t *testing.T) {
		if !strings.Contains(code, "for j := 0; j < 9") {
			t.Errorf("CMO loop must iterate j from 0 to period\nGot:\n%s", code)
		}
	})

	t.Run("nan_guard_on_pairs", func(t *testing.T) {
		/* NaN pairs are skipped to avoid contaminating accumulators */
		if !strings.Contains(code, "math.IsNaN") {
			t.Errorf("CMO must guard against NaN values in loop\nGot:\n%s", code)
		}
	})

	t.Run("result_stored_in_series", func(t *testing.T) {
		if !strings.Contains(code, "cSeries.Set(") {
			t.Errorf("Missing 'cSeries.Set(' assignment\nGot:\n%s", code)
		}
	})
}

func TestCmoHandler_SourceExpressions(t *testing.T) {
	handler := &CmoHandler{}

	tests := []struct {
		name   string
		source ast.Expression
	}{
		{"identifier_close", &ast.Identifier{Name: "close"}},
		{"identifier_hlc3", &ast.Identifier{Name: "hlc3"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := newTestGenerator()
			call := &ast.CallExpression{Arguments: []ast.Expression{
				tt.source,
				&ast.Literal{Value: 5.0},
			}}

			code, err := handler.GenerateCode(gen, "c", call)
			if err != nil {
				t.Fatalf("GenerateCode() error = %v", err)
			}

			if !strings.Contains(code, "cSeries.Set(") {
				t.Errorf("Missing series assignment\nGot:\n%s", code)
			}
		})
	}
}
