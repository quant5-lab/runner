package codegen

import (
	"fmt"
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestKcwHandler_TAFunctionHandlerInterface(t *testing.T) {
	handler := &KcwHandler{}

	t.Run("can_handle_ta_dot_kcw", func(t *testing.T) {
		if !handler.CanHandle("ta.kcw") {
			t.Error("KcwHandler must accept 'ta.kcw'")
		}
	})

	t.Run("can_handle_kcw", func(t *testing.T) {
		if !handler.CanHandle("kcw") {
			t.Error("KcwHandler must accept 'kcw' (v4 compatibility)")
		}
	})

	t.Run("cannot_handle_other_functions", func(t *testing.T) {
		for _, name := range []string{"ta.bbw", "ta.rsi", "ta.sma", "ta.atr", ""} {
			if handler.CanHandle(name) {
				t.Errorf("KcwHandler must not accept %q", name)
			}
		}
	})
}

func TestKcwHandler_CompositeIndicatorMetadataInterface(t *testing.T) {
	handler := &KcwHandler{}
	call := &ast.CallExpression{Arguments: []ast.Expression{
		&ast.Identifier{Name: "close"},
		&ast.Literal{Value: 20.0},
	}}

	names, err := handler.GetInternalSeriesNames("myVar", call)
	if err != nil {
		t.Fatalf("GetInternalSeriesNames error = %v", err)
	}

	t.Run("returns_two_internal_series", func(t *testing.T) {
		if len(names) != 2 {
			t.Errorf("KCW requires exactly 2 internal series, got %d", len(names))
		}
	})

	t.Run("ema_series_named_correctly", func(t *testing.T) {
		if len(names) < 1 || names[0] != "_myVar_ema" {
			t.Errorf("EMA series must be '_myVar_ema', got %q", names[0])
		}
	})

	t.Run("atr_series_named_correctly", func(t *testing.T) {
		if len(names) < 2 || names[1] != "_myVar_atr" {
			t.Errorf("ATR series must be '_myVar_atr', got %q", names[1])
		}
	})
}

func TestKcwHandler_ArgumentValidation(t *testing.T) {
	handler := &KcwHandler{}
	gen := newTestGenerator()

	tests := []struct {
		name    string
		args    []ast.Expression
		wantErr bool
	}{
		{"no_args", []ast.Expression{}, true},
		{"one_arg_missing_period", []ast.Expression{&ast.Identifier{Name: "close"}}, true},
		{"two_args_valid", []ast.Expression{&ast.Identifier{Name: "close"}, &ast.Literal{Value: 20.0}}, false},
		{"three_args_with_mult", []ast.Expression{&ast.Identifier{Name: "close"}, &ast.Literal{Value: 20.0}, &ast.Literal{Value: 2.0}}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			call := &ast.CallExpression{Arguments: tt.args}
			_, err := handler.GenerateCode(gen, "k", call)
			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestKcwHandler_WarmupBehavior(t *testing.T) {
	handler := &KcwHandler{}

	tests := []struct {
		name   string
		period int
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
				&ast.Literal{Value: float64(tt.period)},
			}}

			code, err := handler.GenerateCode(gen, "k", call)
			if err != nil {
				t.Fatalf("GenerateCode() error = %v", err)
			}

			expected := fmt.Sprintf("ctx.BarIndex < %d", tt.period-1)
			if !strings.Contains(code, expected) {
				t.Errorf("Warmup check %q not found\nGot:\n%s", expected, code)
			}
		})
	}
}

func TestKcwHandler_AlgorithmCorrectness(t *testing.T) {
	handler := &KcwHandler{}
	gen := newTestGenerator()

	call := &ast.CallExpression{Arguments: []ast.Expression{
		&ast.Identifier{Name: "close"},
		&ast.Literal{Value: 14.0},
	}}

	code, err := handler.GenerateCode(gen, "k", call)
	if err != nil {
		t.Fatalf("GenerateCode() error = %v", err)
	}

	t.Run("nan_during_warmup", func(t *testing.T) {
		if !strings.Contains(code, "math.NaN()") {
			t.Errorf("KCW must emit NaN during warmup\nGot:\n%s", code)
		}
	})

	t.Run("ema_intermediate_series", func(t *testing.T) {
		if !strings.Contains(code, "_k_ema") {
			t.Errorf("KCW must track EMA intermediate series\nGot:\n%s", code)
		}
	})

	t.Run("atr_intermediate_series", func(t *testing.T) {
		if !strings.Contains(code, "_k_atr") {
			t.Errorf("KCW must track ATR intermediate series\nGot:\n%s", code)
		}
	})

	t.Run("zero_ema_guard", func(t *testing.T) {
		if !strings.Contains(code, "== 0") {
			t.Errorf("KCW must guard against zero EMA (division by zero)\nGot:\n%s", code)
		}
	})

	t.Run("bandwidth_formula_shape", func(t *testing.T) {
		if !strings.Contains(code, "2.0 *") {
			t.Errorf("KCW formula must contain '2.0 *' factor\nGot:\n%s", code)
		}
	})

	t.Run("result_stored_in_series", func(t *testing.T) {
		if !strings.Contains(code, "kSeries.Set(") {
			t.Errorf("Missing 'kSeries.Set(' assignment\nGot:\n%s", code)
		}
	})
}

func TestKcwHandler_MultExpressions(t *testing.T) {
	handler := &KcwHandler{}

	tests := []struct {
		name         string
		multArg      ast.Expression // nil = omitted (default)
		wantContains []string
	}{
		{
			name:         "default_when_omitted",
			multArg:      nil,
			wantContains: []string{"1.5"},
		},
		{
			name:         "literal_integer",
			multArg:      &ast.Literal{Value: 2.0},
			wantContains: []string{"2.0 * 2"},
		},
		{
			name:         "literal_fractional",
			multArg:      &ast.Literal{Value: 2.5},
			wantContains: []string{"2.5"},
		},
		{
			name:         "literal_unit",
			multArg:      &ast.Literal{Value: 1.0},
			wantContains: []string{"2.0 * 1"},
		},
		{
			name:         "identifier_renders_as_series_access",
			multArg:      &ast.Identifier{Name: "myMult"},
			wantContains: []string{"myMultSeries.GetCurrent()"},
		},
		{
			name: "binary_expression_renders_both_operands",
			multArg: &ast.BinaryExpression{
				Operator: "*",
				Left:     &ast.Identifier{Name: "a"},
				Right:    &ast.Identifier{Name: "b"},
			},
			wantContains: []string{"aSeries.GetCurrent()", "bSeries.GetCurrent()"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := newTestGenerator()
			args := []ast.Expression{
				&ast.Identifier{Name: "close"},
				&ast.Literal{Value: 20.0},
			}
			if tt.multArg != nil {
				args = append(args, tt.multArg)
			}
			call := &ast.CallExpression{Arguments: args}

			code, err := handler.GenerateCode(gen, "k", call)
			if err != nil {
				t.Fatalf("GenerateCode() error = %v", err)
			}

			for _, want := range tt.wantContains {
				if !strings.Contains(code, want) {
					t.Errorf("expected %q in generated code\nGot:\n%s", want, code)
				}
			}
		})
	}
}

func TestKcwHandler_SourceExpressions(t *testing.T) {
	handler := &KcwHandler{}

	tests := []struct {
		name   string
		source ast.Expression
	}{
		{"identifier_close", &ast.Identifier{Name: "close"}},
		{"identifier_hlc3", &ast.Identifier{Name: "hlc3"}},
		{"identifier_high", &ast.Identifier{Name: "high"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := newTestGenerator()
			call := &ast.CallExpression{Arguments: []ast.Expression{
				tt.source,
				&ast.Literal{Value: 20.0},
			}}

			code, err := handler.GenerateCode(gen, "k", call)
			if err != nil {
				t.Fatalf("GenerateCode() error = %v", err)
			}

			if !strings.Contains(code, "kSeries.Set(") {
				t.Errorf("Missing series assignment for source %T\nGot:\n%s", tt.source, code)
			}
		})
	}
}

/* TestExtractUseTrueRangeArg covers all paths through the bool-extraction helper.
 *
 * Invariants:
 *   - Missing arg → true  (ATR is the PineScript default)
 *   - Non-literal arg → true  (conservative fallback)
 *   - Non-bool literal → true  (conservative fallback)
 *   - Explicit true → true
 *   - Explicit false → false
 */
func TestExtractUseTrueRangeArg(t *testing.T) {
	boolLiteral := func(v bool) *ast.Literal { return &ast.Literal{Value: v} }
	intLiteral := func(v float64) *ast.Literal { return &ast.Literal{Value: v} }

	tests := []struct {
		name     string
		args     []ast.Expression
		argIndex int
		want     bool
	}{
		{
			name:     "no_args_returns_true",
			args:     nil,
			argIndex: 3,
			want:     true,
		},
		{
			name:     "arg_index_beyond_length_returns_true",
			args:     []ast.Expression{boolLiteral(false)},
			argIndex: 3,
			want:     true,
		},
		{
			name:     "explicit_false_at_index_returns_false",
			args:     []ast.Expression{nil, nil, nil, boolLiteral(false)},
			argIndex: 3,
			want:     false,
		},
		{
			name:     "explicit_true_at_index_returns_true",
			args:     []ast.Expression{nil, nil, nil, boolLiteral(true)},
			argIndex: 3,
			want:     true,
		},
		{
			name:     "non_literal_arg_returns_true",
			args:     []ast.Expression{nil, nil, nil, &ast.Identifier{Name: "useATR"}},
			argIndex: 3,
			want:     true,
		},
		{
			name:     "numeric_literal_not_bool_returns_true",
			args:     []ast.Expression{nil, nil, nil, intLiteral(0)},
			argIndex: 3,
			want:     true,
		},
		{
			name:     "index_zero_explicit_false",
			args:     []ast.Expression{boolLiteral(false)},
			argIndex: 0,
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			call := &ast.CallExpression{Arguments: tt.args}
			got := extractUseTrueRangeArg(call, tt.argIndex)
			if got != tt.want {
				t.Errorf("extractUseTrueRangeArg(argIndex=%d) = %v, want %v", tt.argIndex, got, tt.want)
			}
		})
	}
}

/* TestKcwHandler_UseTrueRangeCodegenBehavior verifies that the useTrueRange flag
 * selects the correct range-measurement sub-algorithm in generated code.
 *
 * useTrueRange=true  → RMA (Wilder smoothing) of True Range (ATR semantics)
 * useTrueRange=false → windowed SMA of (High – Low), no gap adjustment
 *
 * The two paths are mutually exclusive: code distinguishing markers must appear
 * in exactly one branch and must NOT appear in the other.
 */
func TestKcwHandler_UseTrueRangeCodegenBehavior(t *testing.T) {
	handler := &KcwHandler{}

	baseArgs := func(useTrueRange bool) []ast.Expression {
		return []ast.Expression{
			&ast.Identifier{Name: "close"},
			&ast.Literal{Value: 20.0},
			&ast.Literal{Value: 1.5},
			&ast.Literal{Value: useTrueRange},
		}
	}

	tests := []struct {
		name           string
		useTrueRange   bool
		mustContain    []string
		mustNotContain []string
	}{
		{
			name:         "use_true_range_generates_atr_rma_path",
			useTrueRange: true,
			mustContain: []string{
				// True Range inline lambda — only in ATR (RMA) path
				"math.Max(h-l",
				"math.Abs(h-pc)",
				// Wilder smoothing alpha = 1/period — RMA only; EMA uses 2/(period+1)
				"1.0 / float64(",
			},
			mustNotContain: []string{
				"_hlSum",
				"ctx.Data[ctx.BarIndex-_j].High - ctx.Data[ctx.BarIndex-_j].Low",
			},
		},
		{
			name:         "use_high_low_sma_generates_loop_without_true_range",
			useTrueRange: false,
			mustContain: []string{
				"_hlSum",
				"ctx.Data[ctx.BarIndex-_j].High - ctx.Data[ctx.BarIndex-_j].Low",
				"_hlSum / float64(",
			},
			mustNotContain: []string{
				// True Range lambda absent in H-L SMA mode
				"math.Max(h-l",
				// RMA alpha (1/period) absent; EMA uses 2/(period+1), not "1.0 / float64("
				"1.0 / float64(",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := newTestGenerator()
			call := &ast.CallExpression{Arguments: baseArgs(tt.useTrueRange)}

			code, err := handler.GenerateCode(gen, "k", call)
			if err != nil {
				t.Fatalf("GenerateCode() error = %v", err)
			}

			for _, want := range tt.mustContain {
				if !strings.Contains(code, want) {
					t.Errorf("expected %q in generated code\nGot:\n%s", want, code)
				}
			}
			for _, banned := range tt.mustNotContain {
				if strings.Contains(code, banned) {
					t.Errorf("must NOT contain %q in generated code\nGot:\n%s", banned, code)
				}
			}
		})
	}
}

/* TestKcwHandler_UseTrueRangeDefaultIsTrue verifies that omitting the 4th argument
 * produces the same code as explicitly passing true — ATR path is the default.
 */
func TestKcwHandler_UseTrueRangeDefaultIsTrue(t *testing.T) {
	handler := &KcwHandler{}
	gen := newTestGenerator()

	argsDefault := []ast.Expression{
		&ast.Identifier{Name: "close"},
		&ast.Literal{Value: 14.0},
	}
	argsExplicitTrue := []ast.Expression{
		&ast.Identifier{Name: "close"},
		&ast.Literal{Value: 14.0},
		&ast.Literal{Value: 1.5},
		&ast.Literal{Value: true},
	}

	codeDefault, err := handler.GenerateCode(gen, "k", &ast.CallExpression{Arguments: argsDefault})
	if err != nil {
		t.Fatalf("default args: GenerateCode() error = %v", err)
	}

	gen2 := newTestGenerator()
	codeExplicit, err := handler.GenerateCode(gen2, "k", &ast.CallExpression{Arguments: argsExplicitTrue})
	if err != nil {
		t.Fatalf("explicit true: GenerateCode() error = %v", err)
	}

	for _, code := range []string{codeDefault, codeExplicit} {
		if !strings.Contains(code, "math.Max(h-l") {
			t.Error("default/true useTrueRange must generate ATR (True Range lambda)")
		}
		if strings.Contains(code, "_hlSum") {
			t.Error("default/true useTrueRange must NOT generate SMA-HL loop")
		}
	}
}
