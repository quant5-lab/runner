package codegen

import (
	"fmt"
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestWprHandler_TAFunctionHandlerInterface(t *testing.T) {
	handler := &WprHandler{}

	t.Run("can_handle_ta_dot_wpr", func(t *testing.T) {
		if !handler.CanHandle("ta.wpr") {
			t.Error("WprHandler must accept 'ta.wpr'")
		}
	})

	t.Run("can_handle_wpr", func(t *testing.T) {
		if !handler.CanHandle("wpr") {
			t.Error("WprHandler must accept 'wpr' (v4 compatibility)")
		}
	})

	t.Run("cannot_handle_other_functions", func(t *testing.T) {
		for _, name := range []string{"ta.rsi", "ta.mfi", "ta.stoch", ""} {
			if handler.CanHandle(name) {
				t.Errorf("WprHandler must not accept %q", name)
			}
		}
	})
}

func TestWprHandler_ArgumentValidation(t *testing.T) {
	handler := &WprHandler{}

	tests := []struct {
		name    string
		args    []ast.Expression
		wantErr bool
	}{
		{"no_args", []ast.Expression{}, true},
		{"one_arg_period_valid", []ast.Expression{&ast.Literal{Value: 14.0}}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := newTestGenerator()
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

func TestWprHandler_WarmupBehavior(t *testing.T) {
	handler := &WprHandler{}

	tests := []struct {
		name           string
		period         float64
		wantWarmupEdge int
	}{
		{"period_1_warmup_0", 1, 0},
		{"period_2_warmup_1", 2, 1},
		{"period_14_warmup_13", 14, 13},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := newTestGenerator()
			call := &ast.CallExpression{Arguments: []ast.Expression{
				&ast.Literal{Value: tt.period},
			}}

			code, err := handler.GenerateCode(gen, "w", call)
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

func TestWprHandler_AlgorithmCorrectness(t *testing.T) {
	handler := &WprHandler{}
	gen := newTestGenerator()

	call := &ast.CallExpression{Arguments: []ast.Expression{
		&ast.Literal{Value: 14.0},
	}}

	code, err := handler.GenerateCode(gen, "w", call)
	if err != nil {
		t.Fatalf("GenerateCode() error = %v", err)
	}

	t.Run("uses_ctx_data_for_ohlc", func(t *testing.T) {
		if !strings.Contains(code, "ctx.Data") {
			t.Errorf("WPR must access OHLC via ctx.Data\nGot:\n%s", code)
		}
	})

	t.Run("uses_high_low_close_fields", func(t *testing.T) {
		if !strings.Contains(code, ".High") {
			t.Errorf("WPR must use .High field\nGot:\n%s", code)
		}
		if !strings.Contains(code, ".Low") {
			t.Errorf("WPR must use .Low field\nGot:\n%s", code)
		}
		if !strings.Contains(code, ".Close") {
			t.Errorf("WPR must use .Close field\nGot:\n%s", code)
		}
	})

	t.Run("denominator_guard_returns_zero", func(t *testing.T) {
		/* denom == 0.0 → 0.0 (not NaN — price range collapsed) */
		if !strings.Contains(code, "== 0.0") {
			t.Errorf("WPR must guard denom == 0.0\nGot:\n%s", code)
		}
	})

	t.Run("percentage_multiplier", func(t *testing.T) {
		if !strings.Contains(code, "100.0") {
			t.Errorf("WPR must multiply by 100.0\nGot:\n%s", code)
		}
	})

	t.Run("highest_high_tracking", func(t *testing.T) {
		if !strings.Contains(code, "_w_hh") {
			t.Errorf("WPR must track highest high\nGot:\n%s", code)
		}
	})

	t.Run("lowest_low_tracking", func(t *testing.T) {
		if !strings.Contains(code, "_w_ll") {
			t.Errorf("WPR must track lowest low\nGot:\n%s", code)
		}
	})

	t.Run("result_stored_in_series", func(t *testing.T) {
		if !strings.Contains(code, "wSeries.Set(") {
			t.Errorf("Missing 'wSeries.Set(' assignment\nGot:\n%s", code)
		}
	})
}

func TestWprHandler_NoSourceArgument(t *testing.T) {
	handler := &WprHandler{}
	gen := newTestGenerator()

	call := &ast.CallExpression{Arguments: []ast.Expression{
		&ast.Literal{Value: 14.0},
	}}

	code, err := handler.GenerateCode(gen, "w", call)
	if err != nil {
		t.Fatalf("GenerateCode() error = %v", err)
	}

	t.Run("no_arbitrary_series_access", func(t *testing.T) {
		if strings.Contains(code, "srcSeries") {
			t.Errorf("WPR must not reference srcSeries (uses OHLC directly)\nGot:\n%s", code)
		}
	})
}
