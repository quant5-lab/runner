package codegen

import (
	"fmt"
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestLowestbarsHandler_TAFunctionHandlerInterface(t *testing.T) {
	handler := &LowestbarsHandler{}

	t.Run("can_handle_ta_dot_lowestbars", func(t *testing.T) {
		if !handler.CanHandle("ta.lowestbars") {
			t.Error("LowestbarsHandler must accept 'ta.lowestbars'")
		}
	})

	t.Run("can_handle_lowestbars", func(t *testing.T) {
		if !handler.CanHandle("lowestbars") {
			t.Error("LowestbarsHandler must accept 'lowestbars' (v4 compatibility)")
		}
	})

	t.Run("cannot_handle_other_functions", func(t *testing.T) {
		for _, name := range []string{"ta.highestbars", "ta.highest", "ta.lowest", ""} {
			if handler.CanHandle(name) {
				t.Errorf("LowestbarsHandler must not accept %q", name)
			}
		}
	})
}

func TestLowestbarsHandler_ArgumentValidation(t *testing.T) {
	handler := &LowestbarsHandler{}
	gen := newTestGenerator()

	tests := []struct {
		name    string
		args    []ast.Expression
		wantErr bool
	}{
		{"no_args", []ast.Expression{}, true},
		{"one_arg_missing_period", []ast.Expression{&ast.Identifier{Name: "close"}}, true},
		{"two_args_valid", []ast.Expression{&ast.Identifier{Name: "close"}, &ast.Literal{Value: 5.0}}, false},
		{"two_args_with_low_source", []ast.Expression{&ast.Identifier{Name: "low"}, &ast.Literal{Value: 10.0}}, false},
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

func TestLowestbarsHandler_WarmupBehavior(t *testing.T) {
	handler := &LowestbarsHandler{}

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

			code, err := handler.GenerateCode(gen, "lb", call)
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

func TestLowestbarsHandler_AlgorithmCorrectness(t *testing.T) {
	handler := &LowestbarsHandler{}
	gen := newTestGenerator()

	call := &ast.CallExpression{Arguments: []ast.Expression{
		&ast.Identifier{Name: "close"},
		&ast.Literal{Value: 5.0},
	}}

	code, err := handler.GenerateCode(gen, "lb", call)
	if err != nil {
		t.Fatalf("GenerateCode() error = %v", err)
	}

	t.Run("comparison_operator_is_less_than", func(t *testing.T) {
		if !strings.Contains(code, "< _lb_extVal") {
			t.Errorf("Lowestbars must use '<' comparison to find minimum\nGot:\n%s", code)
		}
		if strings.Contains(code, "> _lb_extVal") {
			t.Errorf("Lowestbars must not use '>' (that is highestbars)\nGot:\n%s", code)
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
			t.Errorf("Lowestbars must return float64(-extIdx)\nGot:\n%s", code)
		}
	})

	t.Run("nan_guard_in_loop", func(t *testing.T) {
		if !strings.Contains(code, "math.IsNaN") {
			t.Errorf("Lowestbars loop must guard against NaN values\nGot:\n%s", code)
		}
	})

	t.Run("result_stored_in_series", func(t *testing.T) {
		if !strings.Contains(code, "lbSeries.Set(") {
			t.Errorf("Missing 'lbSeries.Set(' assignment\nGot:\n%s", code)
		}
	})
}

func TestHighestLowestbarsHandlers_ComparisonOperatorsDiffer(t *testing.T) {
	/* Highest and lowest must use opposite operators — catches copy-paste bugs */
	highHandler := &HighestbarsHandler{}
	lowHandler := &LowestbarsHandler{}

	call := &ast.CallExpression{Arguments: []ast.Expression{
		&ast.Identifier{Name: "close"},
		&ast.Literal{Value: 5.0},
	}}

	gen1 := newTestGenerator()
	highCode, _ := highHandler.GenerateCode(gen1, "hb", call)

	gen2 := newTestGenerator()
	lowCode, _ := lowHandler.GenerateCode(gen2, "lb", call)

	if highCode == lowCode {
		t.Error("Highestbars and lowestbars must produce different code")
	}

	if !strings.Contains(highCode, "> _hb_extVal") {
		t.Error("Highestbars must contain '> _hb_extVal'")
	}
	if !strings.Contains(lowCode, "< _lb_extVal") {
		t.Error("Lowestbars must contain '< _lb_extVal'")
	}
}
