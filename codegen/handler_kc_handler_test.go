package codegen

import (
	"fmt"
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestKcHandler_CanHandle(t *testing.T) {
	h := &KcHandler{}
	accept := []string{"ta.kc", "kc"}
	reject := []string{"ta.kcw", "kcw", "ta.kc2", "", "ta.bb", "ta.supertrend"}

	for _, name := range accept {
		if !h.CanHandle(name) {
			t.Errorf("KcHandler must accept %q", name)
		}
	}
	for _, name := range reject {
		if h.CanHandle(name) {
			t.Errorf("KcHandler must not accept %q", name)
		}
	}
}

func TestKcHandler_InternalSeriesNames(t *testing.T) {
	h := &KcHandler{}
	got := h.InternalSeriesNames("upper", nil)
	if len(got) != 2 {
		t.Fatalf("expected 2 internal series, got %d", len(got))
	}
	if got[0] != "_upper_ema" {
		t.Errorf("expected _upper_ema, got %q", got[0])
	}
	if got[1] != "_upper_atr" {
		t.Errorf("expected _upper_atr, got %q", got[1])
	}
}

func TestKcHandler_InternalSeriesNamesReflectFirstOutputVar(t *testing.T) {
	h := &KcHandler{}
	got := h.InternalSeriesNames("kcUpper", nil)
	if got[0] != "_kcUpper_ema" || got[1] != "_kcUpper_atr" {
		t.Errorf("unexpected names: %v", got)
	}
}

func TestKcHandler_OutputVarCountValidation(t *testing.T) {
	h := &KcHandler{}
	call := kcCall(14, 2.0, nil)

	for _, count := range []int{0, 1, 2, 4} {
		count := count
		t.Run(fmt.Sprintf("count_%d_returns_error", count), func(t *testing.T) {
			vars := make([]string, count)
			for i := range vars {
				vars[i] = fmt.Sprintf("v%d", i)
			}
			_, err := h.GenerateTupleCode(newTestGenerator(), vars, call)
			if err == nil {
				t.Errorf("expected error for %d output variables, got nil", count)
			}
		})
	}

	t.Run("count_3_succeeds", func(t *testing.T) {
		_, err := h.GenerateTupleCode(newTestGenerator(), []string{"u", "b", "l"}, call)
		if err != nil {
			t.Errorf("unexpected error for 3 output variables: %v", err)
		}
	})
}

func TestKcHandler_WarmupBehavior(t *testing.T) {
	h := &KcHandler{}
	periods := []int{5, 14, 20}

	for _, period := range periods {
		period := period
		t.Run(fmt.Sprintf("period_%d", period), func(t *testing.T) {
			g := newTestGenerator()
			code, err := h.GenerateTupleCode(g, []string{"u", "b", "l"}, kcCall(period, 2.0, nil))
			if err != nil {
				t.Fatalf("GenerateTupleCode() error = %v", err)
			}
			expected := fmt.Sprintf("ctx.BarIndex < %d", period-1)
			if !strings.Contains(code, expected) {
				t.Errorf("warmup check %q not found\n%s", expected, code)
			}
			if !strings.Contains(code, "math.NaN()") {
				t.Error("warmup path must emit NaN")
			}
		})
	}
}

func TestKcHandler_TwoArgOverload_DefaultsCloseSourceAndATR(t *testing.T) {
	h := &KcHandler{}
	g := newTestGenerator()
	call := &ast.CallExpression{
		Callee:    kcCallee(),
		Arguments: []ast.Expression{&ast.Literal{Value: float64(20)}, &ast.Literal{Value: float64(2)}},
	}
	code, err := h.GenerateTupleCode(g, []string{"upper", "basis", "lower"}, call)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(code, "1.0 / float64(") {
		t.Error("2-arg form must default useTrueRange=true (ATR/RMA path)")
	}
	for _, want := range []string{"upperSeries.Set", "basisSeries.Set", "lowerSeries.Set"} {
		if !strings.Contains(code, want) {
			t.Errorf("expected %s in generated code", want)
		}
	}
}

func TestKcHandler_ThreeArgOverload_DefaultsUseTrueRangeTrue(t *testing.T) {
	h := &KcHandler{}
	g := newTestGenerator()
	code, err := h.GenerateTupleCode(g, []string{"u", "b", "l"}, kcCall(14, 2.0, &ast.Identifier{Name: "close"}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(code, "1.0 / float64(") {
		t.Error("3-arg form must default useTrueRange=true (ATR/RMA path)")
	}
}

/* TestKcHandler_UseTrueRangeCodegenBehavior verifies that useTrueRange selects
 * mutually exclusive code paths:
 *   true  → RMA of True Range (ATR semantics)
 *   false → windowed SMA of (High-Low)
 */
func TestKcHandler_UseTrueRangeCodegenBehavior(t *testing.T) {
	h := &KcHandler{}

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
				"math.Max(h-l",
				"math.Abs(h-pc)",
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
				"math.Max(h-l",
				"1.0 / float64(",
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			g := newTestGenerator()
			code, err := h.GenerateTupleCode(g, []string{"u", "b", "l"}, kcCallWithUseTrueRange(tt.useTrueRange))
			if err != nil {
				t.Fatalf("GenerateTupleCode() error = %v", err)
			}
			for _, want := range tt.mustContain {
				if !strings.Contains(code, want) {
					t.Errorf("expected %q in generated code\n%s", want, code)
				}
			}
			for _, banned := range tt.mustNotContain {
				if strings.Contains(code, banned) {
					t.Errorf("must NOT contain %q in generated code\n%s", banned, code)
				}
			}
		})
	}
}

func TestKcHandler_OutputBandsAreSymmetric(t *testing.T) {
	h := &KcHandler{}
	g := newTestGenerator()
	call := kcCall(20, 2.5, nil)
	code, err := h.GenerateTupleCode(g, []string{"upper", "basis", "lower"}, call)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(code, "_kc_ema + _kc_band") {
		t.Error("upper band must be ema + band")
	}
	if !strings.Contains(code, "_kc_ema - _kc_band") {
		t.Error("lower band must be ema - band")
	}
	if !strings.Contains(code, "basisSeries.Set(_kc_ema)") {
		t.Error("basis must equal EMA")
	}
}

func TestKcHandler_SourceExpressions(t *testing.T) {
	h := &KcHandler{}
	sources := []struct {
		name string
		expr ast.Expression
	}{
		{"close", &ast.Identifier{Name: "close"}},
		{"hlc3", &ast.Identifier{Name: "hlc3"}},
		{"high", &ast.Identifier{Name: "high"}},
	}
	for _, src := range sources {
		src := src
		t.Run(src.name, func(t *testing.T) {
			g := newTestGenerator()
			code, err := h.GenerateTupleCode(g, []string{"u", "b", "l"}, kcCall(14, 2.0, src.expr))
			if err != nil {
				t.Fatalf("GenerateTupleCode() error = %v", err)
			}
			if !strings.Contains(code, "uSeries.Set(") {
				t.Errorf("expected uSeries.Set for source %s\n%s", src.name, code)
			}
		})
	}
}

func TestKcHandler_MultExprRendering(t *testing.T) {
	h := &KcHandler{}
	tests := []struct {
		name       string
		multArg    ast.Expression
		wantInBand string
	}{
		{
			name:       "integer_literal",
			multArg:    &ast.Literal{Value: 2.0},
			wantInBand: "2 * _kc_atr",
		},
		{
			name:       "fractional_literal",
			multArg:    &ast.Literal{Value: 1.5},
			wantInBand: "1.5 * _kc_atr",
		},
		{
			name:       "identifier_renders_as_name",
			multArg:    &ast.Identifier{Name: "myMult"},
			wantInBand: "myMult * _kc_atr",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			call := &ast.CallExpression{
				Callee: kcCallee(),
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
					&ast.Literal{Value: float64(20)},
					tt.multArg,
				},
			}
			code, err := h.GenerateTupleCode(newTestGenerator(), []string{"u", "b", "l"}, call)
			if err != nil {
				t.Fatalf("GenerateTupleCode() error = %v", err)
			}
			if !strings.Contains(code, tt.wantInBand) {
				t.Errorf("expected %q in generated code\n%s", tt.wantInBand, code)
			}
		})
	}
}

/* TestKcHandler_UseTrueRangeDefaultIsTrue verifies that the 2-arg and 3-arg overloads
 * (which omit the useTrueRange argument) produce the same ATR code path as the
 * 4-arg form with explicit true — ATR is the PineScript default for ta.kc.
 */
func TestKcHandler_UseTrueRangeDefaultIsTrue(t *testing.T) {
	h := &KcHandler{}

	tests := []struct {
		name string
		call *ast.CallExpression
	}{
		{
			name: "two_arg_omits_utr",
			call: &ast.CallExpression{
				Callee:    kcCallee(),
				Arguments: []ast.Expression{&ast.Literal{Value: float64(14)}, &ast.Literal{Value: float64(2)}},
			},
		},
		{
			name: "three_arg_omits_utr",
			call: kcCall(14, 2.0, nil),
		},
		{
			name: "four_arg_explicit_true",
			call: kcCallWithUseTrueRange(true),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			code, err := h.GenerateTupleCode(newTestGenerator(), []string{"u", "b", "l"}, tt.call)
			if err != nil {
				t.Fatalf("GenerateTupleCode() error = %v", err)
			}
			if !strings.Contains(code, "math.Max(h-l") {
				t.Error("must generate ATR (True Range lambda) when useTrueRange is true or omitted")
			}
			if strings.Contains(code, "_hlSum") {
				t.Error("must NOT generate SMA-HL loop when useTrueRange is true or omitted")
			}
		})
	}
}

func TestKcHandler_RegistrationInTupleHandler(t *testing.T) {
	h := NewTupleIndicatorHandler()
	if !h.CanHandle("ta.kc") {
		t.Error("TupleIndicatorHandler must route ta.kc")
	}
	if !h.CanHandle("kc") {
		t.Error("TupleIndicatorHandler must route bare kc")
	}
}

func TestExtractKCArguments_AllOverloads(t *testing.T) {
	tests := []struct {
		name       string
		args       []ast.Expression
		wantPeriod int
		wantMult   string
		wantUTR    bool
		wantErr    bool
	}{
		{
			name:       "two_args_no_source",
			args:       []ast.Expression{&ast.Literal{Value: float64(20)}, &ast.Literal{Value: float64(2)}},
			wantPeriod: 20, wantMult: "2", wantUTR: true,
		},
		{
			name: "three_args_with_source",
			args: []ast.Expression{
				&ast.Identifier{Name: "close"},
				&ast.Literal{Value: float64(14)},
				&ast.Literal{Value: float64(1.5)},
			},
			wantPeriod: 14, wantMult: "1.5", wantUTR: true,
		},
		{
			name: "four_args_utr_false",
			args: []ast.Expression{
				&ast.Identifier{Name: "close"},
				&ast.Literal{Value: float64(20)},
				&ast.Literal{Value: float64(2)},
				&ast.Literal{Value: false},
			},
			wantPeriod: 20, wantMult: "2", wantUTR: false,
		},
		{
			name: "four_args_utr_true",
			args: []ast.Expression{
				&ast.Identifier{Name: "hlc3"},
				&ast.Literal{Value: float64(10)},
				&ast.Literal{Value: float64(3)},
				&ast.Literal{Value: true},
			},
			wantPeriod: 10, wantMult: "3", wantUTR: true,
		},
		{
			name:    "zero_args_error",
			args:    []ast.Expression{},
			wantErr: true,
		},
		{
			name:    "one_arg_error",
			args:    []ast.Expression{&ast.Literal{Value: float64(14)}},
			wantErr: true,
		},
		{
			name: "five_args_error",
			args: []ast.Expression{
				&ast.Identifier{Name: "close"},
				&ast.Literal{Value: float64(14)},
				&ast.Literal{Value: float64(2)},
				&ast.Literal{Value: true},
				&ast.Literal{Value: float64(0)},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			call := &ast.CallExpression{Callee: kcCallee(), Arguments: tt.args}
			_, period, multExpr, utr, err := extractKCArguments(call)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if period != tt.wantPeriod {
				t.Errorf("period: want %d, got %d", tt.wantPeriod, period)
			}
			if multExpr != tt.wantMult {
				t.Errorf("multExpr: want %q, got %q", tt.wantMult, multExpr)
			}
			if utr != tt.wantUTR {
				t.Errorf("useTrueRange: want %v, got %v", tt.wantUTR, utr)
			}
		})
	}
}

func kcCallee() ast.Expression {
	return &ast.MemberExpression{
		Object:   &ast.Identifier{Name: "ta"},
		Property: &ast.Identifier{Name: "kc"},
	}
}

func kcCall(period int, mult float64, source ast.Expression) *ast.CallExpression {
	var args []ast.Expression
	if source != nil {
		args = append(args, source)
	} else {
		args = append(args, &ast.Identifier{Name: "close"})
	}
	args = append(args,
		&ast.Literal{Value: float64(period)},
		&ast.Literal{Value: mult},
	)
	return &ast.CallExpression{Callee: kcCallee(), Arguments: args}
}

func kcCallWithUseTrueRange(utr bool) *ast.CallExpression {
	return &ast.CallExpression{
		Callee: kcCallee(),
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "close"},
			&ast.Literal{Value: float64(20)},
			&ast.Literal{Value: float64(2)},
			&ast.Literal{Value: utr},
		},
	}
}
