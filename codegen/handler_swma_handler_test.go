package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestSwmaHandler_TAFunctionHandlerInterface(t *testing.T) {
	handler := &SwmaHandler{}

	t.Run("can_handle_ta_dot_swma", func(t *testing.T) {
		if !handler.CanHandle("ta.swma") {
			t.Error("SwmaHandler must accept 'ta.swma'")
		}
	})

	t.Run("can_handle_swma", func(t *testing.T) {
		if !handler.CanHandle("swma") {
			t.Error("SwmaHandler must accept 'swma' (v4 compatibility)")
		}
	})

	t.Run("cannot_handle_other_functions", func(t *testing.T) {
		for _, name := range []string{"ta.sma", "ta.ema", "ta.cci", "ta.bbw", ""} {
			if handler.CanHandle(name) {
				t.Errorf("SwmaHandler must not accept %q", name)
			}
		}
	})
}

func TestSwmaHandler_ArgumentValidation(t *testing.T) {
	handler := &SwmaHandler{}
	gen := newTestGenerator()

	tests := []struct {
		name    string
		args    []ast.Expression
		wantErr bool
	}{
		{"no_args", []ast.Expression{}, true},
		{"one_arg_valid", []ast.Expression{&ast.Identifier{Name: "close"}}, false},
		{"two_args_still_valid", []ast.Expression{&ast.Identifier{Name: "close"}, &ast.Literal{Value: 4.0}}, false},
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

func TestSwmaHandler_WarmupBehavior(t *testing.T) {
	handler := &SwmaHandler{}
	gen := newTestGenerator()

	call := &ast.CallExpression{Arguments: []ast.Expression{
		&ast.Identifier{Name: "close"},
	}}

	code, err := handler.GenerateCode(gen, "s", call)
	if err != nil {
		t.Fatalf("GenerateCode() error = %v", err)
	}

	/* SWMA has fixed 4-bar window: warmup is always BarIndex < 3 */
	if !strings.Contains(code, "ctx.BarIndex < 3") {
		t.Errorf("SWMA warmup must check BarIndex < 3\nGot:\n%s", code)
	}
}

func TestSwmaHandler_AlgorithmCorrectness(t *testing.T) {
	handler := &SwmaHandler{}
	gen := newTestGenerator()

	call := &ast.CallExpression{Arguments: []ast.Expression{
		&ast.Identifier{Name: "close"},
	}}

	code, err := handler.GenerateCode(gen, "s", call)
	if err != nil {
		t.Fatalf("GenerateCode() error = %v", err)
	}

	t.Run("weight_one_sixth", func(t *testing.T) {
		if !strings.Contains(code, "1.0/6.0") {
			t.Errorf("SWMA must use weight 1/6\nGot:\n%s", code)
		}
	})

	t.Run("weight_two_sixths", func(t *testing.T) {
		if !strings.Contains(code, "2.0/6.0") {
			t.Errorf("SWMA must use weight 2/6\nGot:\n%s", code)
		}
	})

	t.Run("result_stored_in_series", func(t *testing.T) {
		if !strings.Contains(code, "sSeries.Set(") {
			t.Errorf("Missing 'sSeries.Set(' assignment\nGot:\n%s", code)
		}
	})

	t.Run("nan_during_warmup", func(t *testing.T) {
		if !strings.Contains(code, "math.NaN()") {
			t.Errorf("SWMA must emit NaN during warmup\nGot:\n%s", code)
		}
	})
}

func TestSwmaHandler_SourceExpressions(t *testing.T) {
	handler := &SwmaHandler{}

	tests := []struct {
		name   string
		source ast.Expression
	}{
		{"identifier_close", &ast.Identifier{Name: "close"}},
		{"identifier_hlc3", &ast.Identifier{Name: "hlc3"}},
		{"identifier_hl2", &ast.Identifier{Name: "hl2"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := newTestGenerator()
			call := &ast.CallExpression{Arguments: []ast.Expression{tt.source}}

			code, err := handler.GenerateCode(gen, "s", call)
			if err != nil {
				t.Fatalf("GenerateCode() error = %v", err)
			}

			if !strings.Contains(code, "sSeries.Set(") {
				t.Errorf("Missing series assignment\nGot:\n%s", code)
			}
		})
	}
}
