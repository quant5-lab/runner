package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestCrossHandler_TAFunctionHandlerInterface(t *testing.T) {
	handler := &CrossHandler{}

	t.Run("can_handle_ta_dot_cross", func(t *testing.T) {
		if !handler.CanHandle("ta.cross") {
			t.Error("CrossHandler must accept 'ta.cross'")
		}
	})

	t.Run("can_handle_cross", func(t *testing.T) {
		if !handler.CanHandle("cross") {
			t.Error("CrossHandler must accept 'cross' (v4 compatibility)")
		}
	})

	t.Run("cannot_handle_crossover_or_crossunder", func(t *testing.T) {
		for _, name := range []string{"ta.crossover", "ta.crossunder", "crossover", "crossunder", ""} {
			if handler.CanHandle(name) {
				t.Errorf("CrossHandler must not accept %q (use specific directional handlers)", name)
			}
		}
	})

	t.Run("cannot_handle_other_functions", func(t *testing.T) {
		for _, name := range []string{"ta.sma", "ta.ema", "ta.rsi"} {
			if handler.CanHandle(name) {
				t.Errorf("CrossHandler must not accept %q", name)
			}
		}
	})
}

func TestCrossHandler_ArgumentValidation(t *testing.T) {
	handler := &CrossHandler{}

	tests := []struct {
		name    string
		args    []ast.Expression
		wantErr bool
	}{
		{"no_args", []ast.Expression{}, true},
		{"one_arg", []ast.Expression{&ast.Identifier{Name: "close"}}, true},
		{"two_args_valid", []ast.Expression{
			&ast.Identifier{Name: "close"},
			&ast.Identifier{Name: "open"},
		}, false},
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

func TestCrossHandler_AlgorithmCorrectness(t *testing.T) {
	handler := &CrossHandler{}
	gen := newTestGenerator()

	call := &ast.CallExpression{Arguments: []ast.Expression{
		&ast.Identifier{Name: "close"},
		&ast.Identifier{Name: "open"},
	}}

	code, err := handler.GenerateCode(gen, "c", call)
	if err != nil {
		t.Fatalf("GenerateCode() error = %v", err)
	}

	t.Run("bar_index_guard_uses_i_not_ctx", func(t *testing.T) {
		if !strings.Contains(code, "i > 0") {
			t.Errorf("Cross must guard first bar with 'i > 0'\nGot:\n%s", code)
		}
	})

	t.Run("first_bar_returns_zero", func(t *testing.T) {
		if !strings.Contains(code, "0.0") {
			t.Errorf("Cross must emit 0.0 for first bar\nGot:\n%s", code)
		}
	})

	t.Run("crossover_condition", func(t *testing.T) {
		if !strings.Contains(code, ">") {
			t.Errorf("Missing greater-than operator for crossover\nGot:\n%s", code)
		}
		if !strings.Contains(code, "<=") {
			t.Errorf("Missing '<=' for crossover previous bar comparison\nGot:\n%s", code)
		}
	})

	t.Run("crossunder_condition", func(t *testing.T) {
		if !strings.Contains(code, "<") {
			t.Errorf("Missing less-than operator for crossunder\nGot:\n%s", code)
		}
		if !strings.Contains(code, ">=") {
			t.Errorf("Missing '>=' for crossunder previous bar comparison\nGot:\n%s", code)
		}
	})

	t.Run("conditions_combined_with_or", func(t *testing.T) {
		if !strings.Contains(code, "||") {
			t.Errorf("Crossover and crossunder conditions must be combined with '||'\nGot:\n%s", code)
		}
	})

	t.Run("result_is_binary_float", func(t *testing.T) {
		if !strings.Contains(code, "1.0") || !strings.Contains(code, "0.0") {
			t.Errorf("Cross result must be 1.0 or 0.0\nGot:\n%s", code)
		}
	})

	t.Run("result_stored_in_series", func(t *testing.T) {
		if !strings.Contains(code, "cSeries.Set(") {
			t.Errorf("Missing 'cSeries.Set(' assignment\nGot:\n%s", code)
		}
	})
}

func TestCrossHandler_NoWarmupPeriod(t *testing.T) {
	handler := &CrossHandler{}
	gen := newTestGenerator()

	call := &ast.CallExpression{Arguments: []ast.Expression{
		&ast.Identifier{Name: "close"},
		&ast.Identifier{Name: "open"},
	}}

	code, err := handler.GenerateCode(gen, "c", call)
	if err != nil {
		t.Fatalf("GenerateCode() error = %v", err)
	}

	if strings.Contains(code, "ctx.BarIndex <") {
		t.Errorf("CrossHandler must not emit period-based warmup check\nGot:\n%s", code)
	}
}

func TestCrossHandler_TwoSeriesInputTypes(t *testing.T) {
	handler := &CrossHandler{}

	tests := []struct {
		name    string
		series1 ast.Expression
		series2 ast.Expression
	}{
		{
			"two_identifiers",
			&ast.Identifier{Name: "close"},
			&ast.Identifier{Name: "open"},
		},
		{
			"identifier_and_binary",
			&ast.Identifier{Name: "close"},
			&ast.BinaryExpression{
				Left:     &ast.Identifier{Name: "high"},
				Operator: "+",
				Right:    &ast.Identifier{Name: "low"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := newTestGenerator()
			call := &ast.CallExpression{Arguments: []ast.Expression{tt.series1, tt.series2}}
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
