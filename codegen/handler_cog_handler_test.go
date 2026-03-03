package codegen

import (
	"fmt"
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestCogHandler_TAFunctionHandlerInterface(t *testing.T) {
	handler := &CogHandler{}

	t.Run("can_handle_ta_dot_cog", func(t *testing.T) {
		if !handler.CanHandle("ta.cog") {
			t.Error("CogHandler must accept 'ta.cog'")
		}
	})

	t.Run("can_handle_cog", func(t *testing.T) {
		if !handler.CanHandle("cog") {
			t.Error("CogHandler must accept 'cog' (v4 compatibility)")
		}
	})

	t.Run("cannot_handle_other_functions", func(t *testing.T) {
		for _, name := range []string{"ta.cci", "ta.bbw", "ta.sma", "ta.rsi", ""} {
			if handler.CanHandle(name) {
				t.Errorf("CogHandler must not accept %q", name)
			}
		}
	})
}

func TestCogHandler_ArgumentValidation(t *testing.T) {
	handler := &CogHandler{}
	gen := newTestGenerator()

	tests := []struct {
		name    string
		args    []ast.Expression
		wantErr bool
	}{
		{"no_args", []ast.Expression{}, true},
		{"one_arg_missing_period", []ast.Expression{&ast.Identifier{Name: "close"}}, true},
		{"two_args_valid", []ast.Expression{&ast.Identifier{Name: "close"}, &ast.Literal{Value: 10.0}}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			call := &ast.CallExpression{Arguments: tt.args}
			_, err := handler.GenerateCode(gen, "g", call)
			if tt.wantErr && err == nil {
				t.Error("Expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestCogHandler_WarmupBehavior(t *testing.T) {
	handler := &CogHandler{}

	tests := []struct {
		name   string
		period float64
	}{
		{"period_5", 5},
		{"period_10", 10},
		{"period_14", 14},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := newTestGenerator()
			call := &ast.CallExpression{Arguments: []ast.Expression{
				&ast.Identifier{Name: "close"},
				&ast.Literal{Value: tt.period},
			}}

			code, err := handler.GenerateCode(gen, "g", call)
			if err != nil {
				t.Fatalf("GenerateCode() error = %v", err)
			}

			expected := fmt.Sprintf("ctx.BarIndex < %d", int(tt.period))
			if !strings.Contains(code, expected) {
				t.Errorf("Warmup check %q not found\nGot:\n%s", expected, code)
			}
		})
	}
}

func TestCogHandler_AlgorithmCorrectness(t *testing.T) {
	handler := &CogHandler{}
	gen := newTestGenerator()

	call := &ast.CallExpression{Arguments: []ast.Expression{
		&ast.Identifier{Name: "close"},
		&ast.Literal{Value: 10.0},
	}}

	code, err := handler.GenerateCode(gen, "g", call)
	if err != nil {
		t.Fatalf("GenerateCode() error = %v", err)
	}

	t.Run("numerator_variable", func(t *testing.T) {
		if !strings.Contains(code, "_g_num") {
			t.Errorf("COG must use numerator variable\nGot:\n%s", code)
		}
	})

	t.Run("denominator_variable", func(t *testing.T) {
		if !strings.Contains(code, "_g_den") {
			t.Errorf("COG must use denominator variable\nGot:\n%s", code)
		}
	})

	t.Run("negative_sign_in_formula", func(t *testing.T) {
		if !strings.Contains(code, "-_g_num") {
			t.Errorf("COG formula must negate numerator\nGot:\n%s", code)
		}
	})

	t.Run("weight_increments_with_j_plus_1", func(t *testing.T) {
		if !strings.Contains(code, "float64(j+1)") {
			t.Errorf("COG must weight values by (j+1)\nGot:\n%s", code)
		}
	})

	t.Run("zero_denominator_guard", func(t *testing.T) {
		if !strings.Contains(code, "== 0.0") {
			t.Errorf("COG must guard against zero denominator\nGot:\n%s", code)
		}
	})

	t.Run("loop_over_period", func(t *testing.T) {
		if !strings.Contains(code, "for j := 0; j < 10") {
			t.Errorf("COG loop must iterate over period\nGot:\n%s", code)
		}
	})

	t.Run("result_stored_in_series", func(t *testing.T) {
		if !strings.Contains(code, "gSeries.Set(") {
			t.Errorf("Missing 'gSeries.Set(' assignment\nGot:\n%s", code)
		}
	})

	t.Run("nan_during_warmup", func(t *testing.T) {
		if !strings.Contains(code, "math.NaN()") {
			t.Errorf("COG must emit NaN during warmup\nGot:\n%s", code)
		}
	})
}

func TestCogHandler_SourceExpressions(t *testing.T) {
	handler := &CogHandler{}

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
				&ast.Literal{Value: 10.0},
			}}

			code, err := handler.GenerateCode(gen, "g", call)
			if err != nil {
				t.Fatalf("GenerateCode() error = %v", err)
			}

			if !strings.Contains(code, "gSeries.Set(") {
				t.Errorf("Missing series assignment\nGot:\n%s", code)
			}
		})
	}
}
