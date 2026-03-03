package codegen

import (
	"fmt"
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestTsiHandler_TAFunctionHandlerInterface(t *testing.T) {
	handler := &TsiHandler{}

	t.Run("can_handle_ta_dot_tsi", func(t *testing.T) {
		if !handler.CanHandle("ta.tsi") {
			t.Error("TsiHandler must accept 'ta.tsi'")
		}
	})

	t.Run("can_handle_tsi", func(t *testing.T) {
		if !handler.CanHandle("tsi") {
			t.Error("TsiHandler must accept 'tsi' (v4 compatibility)")
		}
	})

	t.Run("cannot_handle_other_functions", func(t *testing.T) {
		for _, name := range []string{"ta.cci", "ta.rsi", "ta.sma", "ta.cog", ""} {
			if handler.CanHandle(name) {
				t.Errorf("TsiHandler must not accept %q", name)
			}
		}
	})
}

func TestTsiHandler_ArgumentValidation(t *testing.T) {
	handler := &TsiHandler{}
	gen := newTestGenerator()

	tests := []struct {
		name    string
		args    []ast.Expression
		wantErr bool
	}{
		{"no_args", []ast.Expression{}, true},
		{"one_arg", []ast.Expression{&ast.Identifier{Name: "close"}}, true},
		{"two_args", []ast.Expression{&ast.Identifier{Name: "close"}, &ast.Literal{Value: 5.0}}, true},
		{"three_args_valid", []ast.Expression{
			&ast.Identifier{Name: "close"},
			&ast.Literal{Value: 5.0},
			&ast.Literal{Value: 13.0},
		}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			call := &ast.CallExpression{Arguments: tt.args}
			_, err := handler.GenerateCode(gen, "t", call)
			if tt.wantErr && err == nil {
				t.Error("Expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestTsiHandler_InternalSeriesNames(t *testing.T) {
	handler := &TsiHandler{}

	names, err := handler.GetInternalSeriesNames("myTSI", &ast.CallExpression{})
	if err != nil {
		t.Fatalf("GetInternalSeriesNames() error = %v", err)
	}

	if len(names) != 6 {
		t.Fatalf("Expected 6 internal series, got %d: %v", len(names), names)
	}

	requiredSuffixes := []string{"_mom", "_mom_abs", "_ema1_mom", "_ema1_abs", "_ema2_mom", "_ema2_abs"}
	for _, suffix := range requiredSuffixes {
		found := false
		for _, name := range names {
			if strings.HasSuffix(name, suffix) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Internal series with suffix %q not found in %v", suffix, names)
		}
	}
}

func TestTsiHandler_GeneratedCode(t *testing.T) {
	handler := &TsiHandler{}
	gen := newTestGenerator()

	call := &ast.CallExpression{Arguments: []ast.Expression{
		&ast.Identifier{Name: "close"},
		&ast.Literal{Value: 5.0},
		&ast.Literal{Value: 13.0},
	}}

	code, err := handler.GenerateCode(gen, "t", call)
	if err != nil {
		t.Fatalf("GenerateCode() error = %v", err)
	}

	t.Run("momentum_series", func(t *testing.T) {
		if !strings.Contains(code, "_t_mom") {
			t.Errorf("TSI must generate momentum series\nGot:\n%s", code)
		}
	})

	t.Run("abs_momentum_series", func(t *testing.T) {
		if !strings.Contains(code, "_t_mom_abs") {
			t.Errorf("TSI must generate absolute momentum series\nGot:\n%s", code)
		}
	})

	t.Run("ema_chain_series", func(t *testing.T) {
		if !strings.Contains(code, "_t_ema1_mom") || !strings.Contains(code, "_t_ema2_mom") {
			t.Errorf("TSI must generate double EMA chain series\nGot:\n%s", code)
		}
	})

	t.Run("warmup_guard_present", func(t *testing.T) {
		if !strings.Contains(code, "ctx.BarIndex <") {
			t.Errorf("TSI must emit a warmup guard\nGot:\n%s", code)
		}
	})

	t.Run("hundred_multiplier", func(t *testing.T) {
		if !strings.Contains(code, "100.0") {
			t.Errorf("TSI formula must scale by 100.0\nGot:\n%s", code)
		}
	})

	t.Run("zero_abs_guard", func(t *testing.T) {
		if !strings.Contains(code, "0.0") {
			t.Errorf("TSI must guard against zero absolute denominator\nGot:\n%s", code)
		}
	})
}

func TestTsiHandler_SourceExpressions(t *testing.T) {
	handler := &TsiHandler{}

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
				&ast.Literal{Value: 13.0},
			}}

			_, err := handler.GenerateCode(gen, "t", call)
			if err != nil {
				t.Fatalf("GenerateCode() error = %v", err)
			}
		})
	}
}

func TestTsiHandler_WarmupBehavior(t *testing.T) {
	handler := &TsiHandler{}

	tests := []struct {
		name           string
		short          float64
		long           float64
		wantWarmupEdge int
	}{
		{"short1_long1", 1, 1, 1},
		{"short1_long13", 1, 13, 13},
		{"short5_long1", 5, 1, 5},
		{"short5_long13", 5, 13, 17},
		{"short13_long25", 13, 25, 37},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := newTestGenerator()
			call := &ast.CallExpression{Arguments: []ast.Expression{
				&ast.Identifier{Name: "close"},
				&ast.Literal{Value: tt.short},
				&ast.Literal{Value: tt.long},
			}}

			code, err := handler.GenerateCode(gen, "t", call)
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
