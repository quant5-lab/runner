package codegen

import (
	"fmt"
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestRocHandler_TAFunctionHandlerInterface(t *testing.T) {
	handler := &RocHandler{}

	t.Run("can_handle_ta_dot_roc", func(t *testing.T) {
		if !handler.CanHandle("ta.roc") {
			t.Error("RocHandler must accept 'ta.roc'")
		}
	})

	t.Run("can_handle_roc", func(t *testing.T) {
		if !handler.CanHandle("roc") {
			t.Error("RocHandler must accept 'roc' (v4 compatibility)")
		}
	})

	t.Run("cannot_handle_other_functions", func(t *testing.T) {
		for _, name := range []string{"ta.mom", "ta.cmo", "ta.rsi", ""} {
			if handler.CanHandle(name) {
				t.Errorf("RocHandler must not accept %q", name)
			}
		}
	})
}

func TestRocHandler_ArgumentValidation(t *testing.T) {
	handler := &RocHandler{}
	gen := newTestGenerator()

	tests := []struct {
		name    string
		args    []ast.Expression
		wantErr bool
	}{
		{"no_args", []ast.Expression{}, true},
		{"one_arg_missing_period", []ast.Expression{&ast.Identifier{Name: "close"}}, true},
		{"two_args_valid", []ast.Expression{&ast.Identifier{Name: "close"}, &ast.Literal{Value: 12.0}}, false},
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

func TestRocHandler_WarmupBehavior(t *testing.T) {
	handler := &RocHandler{}

	tests := []struct {
		name           string
		period         float64
		wantWarmupEdge int
	}{
		{"period_1", 1, 1},
		{"period_12", 12, 12},
		{"period_14", 14, 14},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := newTestGenerator()
			call := &ast.CallExpression{Arguments: []ast.Expression{
				&ast.Identifier{Name: "close"},
				&ast.Literal{Value: tt.period},
			}}

			code, err := handler.GenerateCode(gen, "r", call)
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

func TestRocHandler_AlgorithmCorrectness(t *testing.T) {
	handler := &RocHandler{}
	gen := newTestGenerator()

	call := &ast.CallExpression{Arguments: []ast.Expression{
		&ast.Identifier{Name: "close"},
		&ast.Literal{Value: 12.0},
	}}

	code, err := handler.GenerateCode(gen, "r", call)
	if err != nil {
		t.Fatalf("GenerateCode() error = %v", err)
	}

	t.Run("division_guard_for_zero", func(t *testing.T) {
		if !strings.Contains(code, "== 0.0") {
			t.Errorf("Roc must guard against past == 0.0 (division by zero)\nGot:\n%s", code)
		}
	})

	t.Run("division_guard_for_nan", func(t *testing.T) {
		if !strings.Contains(code, "math.IsNaN") {
			t.Errorf("Roc must guard against math.IsNaN(past)\nGot:\n%s", code)
		}
	})

	t.Run("zero_or_nan_past_returns_nan", func(t *testing.T) {
		if !strings.Contains(code, "math.NaN()") {
			t.Errorf("Roc must return NaN when past is zero or NaN\nGot:\n%s", code)
		}
	})

	t.Run("percentage_multiplier", func(t *testing.T) {
		if !strings.Contains(code, "100.0") {
			t.Errorf("Roc must multiply by 100.0 for percentage result\nGot:\n%s", code)
		}
	})

	t.Run("division_present", func(t *testing.T) {
		if !strings.Contains(code, " / ") {
			t.Errorf("Roc must divide by past value\nGot:\n%s", code)
		}
	})

	t.Run("subtraction_present", func(t *testing.T) {
		if !strings.Contains(code, " - ") {
			t.Errorf("Roc must subtract past from current\nGot:\n%s", code)
		}
	})

	t.Run("result_stored_in_series", func(t *testing.T) {
		if !strings.Contains(code, "rSeries.Set(") {
			t.Errorf("Missing 'rSeries.Set(' assignment\nGot:\n%s", code)
		}
	})
}

func TestRocHandler_NestedBranchStructure(t *testing.T) {
	handler := &RocHandler{}
	gen := newTestGenerator()

	call := &ast.CallExpression{Arguments: []ast.Expression{
		&ast.Identifier{Name: "close"},
		&ast.Literal{Value: 5.0},
	}}

	code, err := handler.GenerateCode(gen, "r", call)
	if err != nil {
		t.Fatalf("GenerateCode() error = %v", err)
	}

	t.Run("outer_warmup_branch", func(t *testing.T) {
		if !strings.Contains(code, "ctx.BarIndex <") {
			t.Errorf("Missing outer warmup check\nGot:\n%s", code)
		}
	})

	t.Run("inner_guard_branch", func(t *testing.T) {
		if !strings.Contains(code, "math.IsNaN") {
			t.Errorf("Missing inner NaN guard\nGot:\n%s", code)
		}
	})

	t.Run("both_nan_paths_present", func(t *testing.T) {
		nanCount := strings.Count(code, "math.NaN()")
		if nanCount < 2 {
			t.Errorf("Expected at least 2 math.NaN() occurrences (warmup + guard), got %d\nGot:\n%s", nanCount, code)
		}
	})
}
