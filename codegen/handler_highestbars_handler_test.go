package codegen

import (
	"fmt"
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestHighestbarsHandler_TAFunctionHandlerInterface(t *testing.T) {
	handler := &HighestbarsHandler{}

	t.Run("can_handle_ta_dot_highestbars", func(t *testing.T) {
		if !handler.CanHandle("ta.highestbars") {
			t.Error("HighestbarsHandler must accept 'ta.highestbars'")
		}
	})

	t.Run("can_handle_highestbars", func(t *testing.T) {
		if !handler.CanHandle("highestbars") {
			t.Error("HighestbarsHandler must accept 'highestbars' (v4 compatibility)")
		}
	})

	t.Run("cannot_handle_other_functions", func(t *testing.T) {
		for _, name := range []string{"ta.lowestbars", "ta.highest", "ta.lowest", ""} {
			if handler.CanHandle(name) {
				t.Errorf("HighestbarsHandler must not accept %q", name)
			}
		}
	})
}

func TestHighestbarsHandler_ArgumentValidation(t *testing.T) {
	handler := &HighestbarsHandler{}
	gen := newTestGenerator()

	tests := []struct {
		name    string
		args    []ast.Expression
		wantErr bool
	}{
		{"no_args", []ast.Expression{}, true},
		{"one_arg_missing_period", []ast.Expression{&ast.Identifier{Name: "close"}}, true},
		{"two_args_valid", []ast.Expression{&ast.Identifier{Name: "close"}, &ast.Literal{Value: 5.0}}, false},
		{"two_args_with_high_source", []ast.Expression{&ast.Identifier{Name: "high"}, &ast.Literal{Value: 10.0}}, false},
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

func TestHighestbarsHandler_WarmupBehavior(t *testing.T) {
	handler := &HighestbarsHandler{}

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
				&ast.Identifier{Name: "close"},
				&ast.Literal{Value: tt.period},
			}}

			code, err := handler.GenerateCode(gen, "hb", call)
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

func TestHighestbarsHandler_AlgorithmCorrectness(t *testing.T) {
	handler := &HighestbarsHandler{}
	gen := newTestGenerator()

	call := &ast.CallExpression{Arguments: []ast.Expression{
		&ast.Identifier{Name: "close"},
		&ast.Literal{Value: 5.0},
	}}

	code, err := handler.GenerateCode(gen, "hb", call)
	if err != nil {
		t.Fatalf("GenerateCode() error = %v", err)
	}

	t.Run("comparison_operator_is_greater_than", func(t *testing.T) {
		if !strings.Contains(code, "> ") && !strings.Contains(code, ">_hb") {
			t.Errorf("Highestbars must use '>' comparison to find maximum\nGot:\n%s", code)
		}
		if strings.Contains(code, "< _hb_extVal") {
			t.Errorf("Highestbars must not use '<' (that is lowestbars)\nGot:\n%s", code)
		}
	})

	t.Run("tracks_extremum_index", func(t *testing.T) {
		if !strings.Contains(code, "extIdx") {
			t.Errorf("Must track extremum index\nGot:\n%s", code)
		}
		if !strings.Contains(code, "extVal") {
			t.Errorf("Must track extremum value\nGot:\n%s", code)
		}
	})

	t.Run("returns_negative_offset", func(t *testing.T) {
		if !strings.Contains(code, "float64(-") {
			t.Errorf("Highestbars must return float64(-extIdx)\nGot:\n%s", code)
		}
	})

	t.Run("nan_guard_in_loop", func(t *testing.T) {
		if !strings.Contains(code, "math.IsNaN") {
			t.Errorf("Highestbars loop must guard against NaN values\nGot:\n%s", code)
		}
	})

	t.Run("result_stored_in_series", func(t *testing.T) {
		if !strings.Contains(code, "hbSeries.Set(") {
			t.Errorf("Missing 'hbSeries.Set(' assignment\nGot:\n%s", code)
		}
	})
}

func TestHighestbarsHandler_SourceExpressions(t *testing.T) {
	handler := &HighestbarsHandler{}

	tests := []struct {
		name   string
		source ast.Expression
	}{
		{"identifier_high", &ast.Identifier{Name: "high"}},
		{"identifier_close", &ast.Identifier{Name: "close"}},
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
				&ast.Literal{Value: 3.0},
			}}

			code, err := handler.GenerateCode(gen, "hb", call)
			if err != nil {
				t.Fatalf("GenerateCode() error = %v", err)
			}

			if !strings.Contains(code, "hbSeries.Set(") {
				t.Errorf("Missing series assignment\nGot:\n%s", code)
			}
			if !strings.Contains(code, "float64(-") {
				t.Errorf("Missing negative offset return\nGot:\n%s", code)
			}
		})
	}
}
