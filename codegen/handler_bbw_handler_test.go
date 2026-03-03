package codegen

import (
	"fmt"
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestBbwHandler_TAFunctionHandlerInterface(t *testing.T) {
	handler := &BbwHandler{}

	t.Run("can_handle_ta_dot_bbw", func(t *testing.T) {
		if !handler.CanHandle("ta.bbw") {
			t.Error("BbwHandler must accept 'ta.bbw'")
		}
	})

	t.Run("can_handle_bbw", func(t *testing.T) {
		if !handler.CanHandle("bbw") {
			t.Error("BbwHandler must accept 'bbw' (v4 compatibility)")
		}
	})

	t.Run("cannot_handle_other_functions", func(t *testing.T) {
		for _, name := range []string{"ta.cci", "ta.rsi", "ta.sma", "ta.bbands", ""} {
			if handler.CanHandle(name) {
				t.Errorf("BbwHandler must not accept %q", name)
			}
		}
	})
}

func TestBbwHandler_ArgumentValidation(t *testing.T) {
	handler := &BbwHandler{}
	gen := newTestGenerator()

	tests := []struct {
		name    string
		args    []ast.Expression
		wantErr bool
	}{
		{"no_args", []ast.Expression{}, true},
		{"one_arg_missing_period", []ast.Expression{&ast.Identifier{Name: "close"}}, true},
		{"two_args_valid", []ast.Expression{&ast.Identifier{Name: "close"}, &ast.Literal{Value: 20.0}}, false},
		{"three_args_with_mult", []ast.Expression{&ast.Identifier{Name: "close"}, &ast.Literal{Value: 20.0}, &ast.Literal{Value: 2.5}}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			call := &ast.CallExpression{Arguments: tt.args}
			_, err := handler.GenerateCode(gen, "b", call)
			if tt.wantErr && err == nil {
				t.Error("Expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestBbwHandler_WarmupBehavior(t *testing.T) {
	handler := &BbwHandler{}

	tests := []struct {
		name   string
		period float64
	}{
		{"period_5", 5},
		{"period_20", 20},
		{"period_50", 50},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := newTestGenerator()
			call := &ast.CallExpression{Arguments: []ast.Expression{
				&ast.Identifier{Name: "close"},
				&ast.Literal{Value: tt.period},
			}}

			code, err := handler.GenerateCode(gen, "b", call)
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

func TestBbwHandler_AlgorithmCorrectness(t *testing.T) {
	handler := &BbwHandler{}
	gen := newTestGenerator()

	call := &ast.CallExpression{Arguments: []ast.Expression{
		&ast.Identifier{Name: "close"},
		&ast.Literal{Value: 20.0},
	}}

	code, err := handler.GenerateCode(gen, "b", call)
	if err != nil {
		t.Fatalf("GenerateCode() error = %v", err)
	}

	t.Run("sma_variable", func(t *testing.T) {
		if !strings.Contains(code, "_b_sma") {
			t.Errorf("BBW must compute SMA variable\nGot:\n%s", code)
		}
	})

	t.Run("stdev_variable", func(t *testing.T) {
		if !strings.Contains(code, "_b_sd") {
			t.Errorf("BBW must compute standard deviation variable\nGot:\n%s", code)
		}
	})

	t.Run("sqrt_for_stdev", func(t *testing.T) {
		if !strings.Contains(code, "math.Sqrt(") {
			t.Errorf("BBW must use math.Sqrt for standard deviation\nGot:\n%s", code)
		}
	})

	t.Run("default_multiplier_2", func(t *testing.T) {
		if !strings.Contains(code, "2.0 *") || !strings.Contains(code, "2") {
			t.Errorf("BBW default formula must include multiplier 2\nGot:\n%s", code)
		}
	})

	t.Run("zero_sma_guard", func(t *testing.T) {
		if !strings.Contains(code, "== 0.0") {
			t.Errorf("BBW must guard against zero SMA\nGot:\n%s", code)
		}
	})

	t.Run("result_stored_in_series", func(t *testing.T) {
		if !strings.Contains(code, "bSeries.Set(") {
			t.Errorf("Missing 'bSeries.Set(' assignment\nGot:\n%s", code)
		}
	})
}

func TestBbwHandler_CustomMultiplier(t *testing.T) {
	handler := &BbwHandler{}
	gen := newTestGenerator()

	call := &ast.CallExpression{Arguments: []ast.Expression{
		&ast.Identifier{Name: "close"},
		&ast.Literal{Value: 20.0},
		&ast.Literal{Value: 1.5},
	}}

	code, err := handler.GenerateCode(gen, "b", call)
	if err != nil {
		t.Fatalf("GenerateCode() error = %v", err)
	}

	if !strings.Contains(code, "1.5") {
		t.Errorf("BBW must use custom multiplier 1.5 when provided\nGot:\n%s", code)
	}
}

func TestBbwHandler_SourceExpressions(t *testing.T) {
	handler := &BbwHandler{}

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
				&ast.Literal{Value: 20.0},
			}}

			code, err := handler.GenerateCode(gen, "b", call)
			if err != nil {
				t.Fatalf("GenerateCode() error = %v", err)
			}

			if !strings.Contains(code, "bSeries.Set(") {
				t.Errorf("Missing series assignment\nGot:\n%s", code)
			}
		})
	}
}
