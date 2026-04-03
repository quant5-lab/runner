package codegen

import (
	"fmt"
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestCciHandler_TAFunctionHandlerInterface(t *testing.T) {
	handler := &CciHandler{}

	t.Run("can_handle_ta_dot_cci", func(t *testing.T) {
		if !handler.CanHandle("ta.cci") {
			t.Error("CciHandler must accept 'ta.cci'")
		}
	})

	t.Run("can_handle_cci", func(t *testing.T) {
		if !handler.CanHandle("cci") {
			t.Error("CciHandler must accept 'cci' (v4 compatibility)")
		}
	})

	t.Run("cannot_handle_other_functions", func(t *testing.T) {
		for _, name := range []string{"ta.sma", "ta.rsi", "ta.swma", "ta.bbw", ""} {
			if handler.CanHandle(name) {
				t.Errorf("CciHandler must not accept %q", name)
			}
		}
	})
}

func TestCciHandler_ArgumentValidation(t *testing.T) {
	handler := &CciHandler{}
	gen := newTestGenerator()

	tests := []struct {
		name    string
		args    []ast.Expression
		wantErr bool
	}{
		{"no_args", []ast.Expression{}, true},
		{"one_arg_missing_period", []ast.Expression{&ast.Identifier{Name: "close"}}, true},
		{"two_args_valid", []ast.Expression{&ast.Identifier{Name: "close"}, &ast.Literal{Value: 20.0}}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			call := &ast.CallExpression{Arguments: tt.args}
			_, err := handler.GenerateCode(gen, "c", call)
			if tt.wantErr && err == nil {
				t.Error("Expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestCciHandler_WarmupBehavior(t *testing.T) {
	handler := &CciHandler{}

	tests := []struct {
		name   string
		period float64
	}{
		{"period_5", 5},
		{"period_14", 14},
		{"period_20", 20},
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

			expected := fmt.Sprintf("ctx.BarIndex < %d", int(tt.period))
			if !strings.Contains(code, expected) {
				t.Errorf("Warmup check %q not found\nGot:\n%s", expected, code)
			}
		})
	}
}

func TestCciHandler_AlgorithmCorrectness(t *testing.T) {
	handler := &CciHandler{}
	gen := newTestGenerator()

	call := &ast.CallExpression{Arguments: []ast.Expression{
		&ast.Identifier{Name: "close"},
		&ast.Literal{Value: 20.0},
	}}

	code, err := handler.GenerateCode(gen, "c", call)
	if err != nil {
		t.Fatalf("GenerateCode() error = %v", err)
	}

	t.Run("sma_variable", func(t *testing.T) {
		if !strings.Contains(code, "_c_sma") {
			t.Errorf("CCI must compute SMA variable\nGot:\n%s", code)
		}
	})

	t.Run("deviation_variable", func(t *testing.T) {
		if !strings.Contains(code, "_c_dev") {
			t.Errorf("CCI must compute mean absolute deviation variable\nGot:\n%s", code)
		}
	})

	t.Run("cci_constant_0015", func(t *testing.T) {
		if !strings.Contains(code, "0.015") {
			t.Errorf("CCI formula must include 0.015 constant\nGot:\n%s", code)
		}
	})

	t.Run("zero_deviation_guard", func(t *testing.T) {
		if !strings.Contains(code, "== 0.0") {
			t.Errorf("CCI must guard against zero deviation\nGot:\n%s", code)
		}
	})

	t.Run("loop_over_period", func(t *testing.T) {
		if !strings.Contains(code, "for j := 0; j < 20") {
			t.Errorf("CCI loop must iterate over period\nGot:\n%s", code)
		}
	})

	t.Run("result_stored_in_series", func(t *testing.T) {
		if !strings.Contains(code, "cSeries.Set(") {
			t.Errorf("Missing 'cSeries.Set(' assignment\nGot:\n%s", code)
		}
	})

	t.Run("nan_during_warmup", func(t *testing.T) {
		if !strings.Contains(code, "math.NaN()") {
			t.Errorf("CCI must emit NaN during warmup\nGot:\n%s", code)
		}
	})
}

func TestCciHandler_SourceExpressions(t *testing.T) {
	handler := &CciHandler{}

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
