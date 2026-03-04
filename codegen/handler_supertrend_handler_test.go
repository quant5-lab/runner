package codegen

import (
	"fmt"
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestSupertrendHandler_CanHandle(t *testing.T) {
	h := &SupertrendHandler{}
	accept := []string{"ta.supertrend", "supertrend"}
	reject := []string{"ta.kc", "ta.kcw", "", "supertrend2", "ta.atr"}

	for _, name := range accept {
		if !h.CanHandle(name) {
			t.Errorf("SupertrendHandler must accept %q", name)
		}
	}
	for _, name := range reject {
		if h.CanHandle(name) {
			t.Errorf("SupertrendHandler must not accept %q", name)
		}
	}
}

func TestSupertrendHandler_InternalSeriesNames(t *testing.T) {
	h := &SupertrendHandler{}
	got := h.InternalSeriesNames("st", nil)
	if len(got) != 4 {
		t.Fatalf("expected 4 internal series, got %d: %v", len(got), got)
	}
	expected := []string{"_st_atr", "_st_upper", "_st_lower", "_st_dir"}
	for i, want := range expected {
		if got[i] != want {
			t.Errorf("[%d] want %q, got %q", i, want, got[i])
		}
	}
}

func TestSupertrendHandler_InternalSeriesNamesReflectFirstOutputVar(t *testing.T) {
	h := &SupertrendHandler{}
	got := h.InternalSeriesNames("myTrend", nil)
	expected := []string{"_myTrend_atr", "_myTrend_upper", "_myTrend_lower", "_myTrend_dir"}
	for i, want := range expected {
		if got[i] != want {
			t.Errorf("[%d] want %q, got %q", i, want, got[i])
		}
	}
}

func TestSupertrendHandler_OutputVarCountValidation(t *testing.T) {
	h := &SupertrendHandler{}
	call := stCall(3.0, 10)

	for _, count := range []int{0, 1, 3} {
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

	t.Run("count_2_succeeds", func(t *testing.T) {
		_, err := h.GenerateTupleCode(newTestGenerator(), []string{"st", "dir"}, call)
		if err != nil {
			t.Errorf("unexpected error for 2 output variables: %v", err)
		}
	})
}

func TestSupertrendHandler_WarmupBehavior(t *testing.T) {
	h := &SupertrendHandler{}
	periods := []int{5, 10, 14}

	for _, period := range periods {
		period := period
		t.Run(fmt.Sprintf("atr_period_%d", period), func(t *testing.T) {
			g := newTestGenerator()
			code, err := h.GenerateTupleCode(g, []string{"st", "dir"}, stCall(3.0, period))
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

func TestSupertrendHandler_BandClampingGenerated(t *testing.T) {
	h := &SupertrendHandler{}
	g := newTestGenerator()
	code, err := h.GenerateTupleCode(g, []string{"st", "dir"}, stCall(3.0, 10))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(code, "_st_prevLower") {
		t.Error("expected _st_prevLower carry-forward in generated code")
	}
	if !strings.Contains(code, "_st_prevUpper") {
		t.Error("expected _st_prevUpper carry-forward in generated code")
	}
	if !strings.Contains(code, "math.IsNaN") {
		t.Error("expected NaN guard for first-bar clamping")
	}
}

func TestSupertrendHandler_DirectionFlipConditions(t *testing.T) {
	h := &SupertrendHandler{}
	g := newTestGenerator()
	code, err := h.GenerateTupleCode(g, []string{"st", "dir"}, stCall(3.0, 14))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(code, "_st_prevDir == -1.0") {
		t.Error("expected bearish-to-bullish flip check (_st_prevDir == -1.0)")
	}
	if !strings.Contains(code, "_st_prevDir == 1.0") {
		t.Error("expected bullish-to-bearish flip check (_st_prevDir == 1.0)")
	}
}

func TestSupertrendHandler_SupertrendValueSelection(t *testing.T) {
	h := &SupertrendHandler{}
	g := newTestGenerator()
	code, err := h.GenerateTupleCode(g, []string{"st", "dir"}, stCall(3.0, 10))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(code, "_st_value := _st_lower") {
		t.Error("expected _st_value := _st_lower (uptrend default)")
	}
	if !strings.Contains(code, "_st_value = _st_upper") {
		t.Error("expected _st_value = _st_upper (downtrend override)")
	}
}

func TestSupertrendHandler_HL2UsedAsMidpoint(t *testing.T) {
	h := &SupertrendHandler{}
	g := newTestGenerator()
	code, err := h.GenerateTupleCode(g, []string{"st", "dir"}, stCall(3.0, 10))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(code, "_st_hl2 := (ctx.Data[ctx.BarIndex].High + ctx.Data[ctx.BarIndex].Low) / 2.0") {
		t.Error("expected hl2 = (high + low) / 2.0 as supertrend midpoint")
	}
}

func TestSupertrendHandler_ATRUsesRMA(t *testing.T) {
	h := &SupertrendHandler{}
	periods := []int{3, 10, 14}

	for _, period := range periods {
		period := period
		t.Run(fmt.Sprintf("period_%d", period), func(t *testing.T) {
			code, err := h.GenerateTupleCode(newTestGenerator(), []string{"st", "dir"}, stCall(3.0, period))
			if err != nil {
				t.Fatalf("GenerateTupleCode() error = %v", err)
			}
			want := fmt.Sprintf("1.0 / float64(%d)", period)
			if !strings.Contains(code, want) {
				t.Errorf("ATR (RMA) must use alpha = 1/period: expected %q\n%s", want, code)
			}
			if !strings.Contains(code, "math.Max(h-l") {
				t.Error("ATR must include True Range components")
			}
		})
	}
}

func TestSupertrendHandler_FactorApplied(t *testing.T) {
	h := &SupertrendHandler{}
	tests := []struct {
		name       string
		factor     float64
		wantInCode string
	}{
		{"integer_renders_without_decimal", 3.0, "3*_st_atr"},
		{"fractional_renders_with_decimal", 3.7, "3.7*_st_atr"},
		{"unit_factor", 1.0, "1*_st_atr"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			code, err := h.GenerateTupleCode(newTestGenerator(), []string{"st", "dir"}, stCall(tt.factor, 10))
			if err != nil {
				t.Fatalf("GenerateTupleCode() error = %v", err)
			}
			if !strings.Contains(code, tt.wantInCode) {
				t.Errorf("expected %q in generated code\n%s", tt.wantInCode, code)
			}
		})
	}
}

func TestSupertrendHandler_PreviousCloseUsedForClamping(t *testing.T) {
	h := &SupertrendHandler{}
	g := newTestGenerator()
	code, err := h.GenerateTupleCode(g, []string{"st", "dir"}, stCall(3.0, 10))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(code, "closeSeries.Get(1)") {
		t.Error("expected closeSeries.Get(1) for previous close in band clamping")
	}
}

func TestSupertrendHandler_RegistrationInTupleHandler(t *testing.T) {
	h := NewTupleIndicatorHandler()
	if !h.CanHandle("ta.supertrend") {
		t.Error("TupleIndicatorHandler must route ta.supertrend")
	}
	if !h.CanHandle("supertrend") {
		t.Error("TupleIndicatorHandler must route bare supertrend")
	}
}

func TestExtractSupertrendArguments_Valid(t *testing.T) {
	tests := []struct {
		name       string
		factor     float64
		atrPeriod  int
		wantFactor string
	}{
		{"standard", 3.0, 10, "3"},
		{"non_integer_factor", 3.7, 14, "3.7"},
		{"small_period", 1.5, 3, "1.5"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			call := stCall(tt.factor, tt.atrPeriod)
			factor, period, err := extractSupertrendArguments(call)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if period != tt.atrPeriod {
				t.Errorf("period: want %d, got %d", tt.atrPeriod, period)
			}
			if factor != tt.wantFactor {
				t.Errorf("factor: want %q, got %q", tt.wantFactor, factor)
			}
		})
	}
}

func TestExtractSupertrendArguments_WrongArgCount(t *testing.T) {
	for _, argCount := range []int{0, 1, 3} {
		args := make([]ast.Expression, argCount)
		for i := range args {
			args[i] = &ast.Literal{Value: float64(1)}
		}
		call := &ast.CallExpression{Callee: stCallee(), Arguments: args}
		_, _, err := extractSupertrendArguments(call)
		if err == nil {
			t.Errorf("expected error for %d arguments, got nil", argCount)
		}
	}
}

func TestSupertrendHandler_InitialDirectionIsUptrend(t *testing.T) {
	h := &SupertrendHandler{}
	g := newTestGenerator()
	code, err := h.GenerateTupleCode(g, []string{"st", "dir"}, stCall(3.0, 10))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(code, "_st_dir := 1.0") {
		t.Errorf("initial direction must be 1.0 (uptrend)\n%s", code)
	}
}

func TestSupertrendHandler_BothOutputSeriesPopulated(t *testing.T) {
	h := &SupertrendHandler{}
	g := newTestGenerator()
	code, err := h.GenerateTupleCode(g, []string{"st", "dir"}, stCall(3.0, 10))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(code, "stSeries.Set(") {
		t.Error("supertrend value series (stSeries) must be assigned")
	}
	if !strings.Contains(code, "dirSeries.Set(") {
		t.Error("direction series (dirSeries) must be assigned")
	}
}

func TestSupertrendHandler_BandFormula(t *testing.T) {
	h := &SupertrendHandler{}
	code, err := h.GenerateTupleCode(newTestGenerator(), []string{"st", "dir"}, stCall(3.0, 10))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(code, "_st_hl2 + 3*_st_atr") {
		t.Error("raw upper band must be _st_hl2 + factor*_st_atr")
	}
	if !strings.Contains(code, "_st_hl2 - 3*_st_atr") {
		t.Error("raw lower band must be _st_hl2 - factor*_st_atr")
	}
}

func stCallee() ast.Expression {
	return &ast.MemberExpression{
		Object:   &ast.Identifier{Name: "ta"},
		Property: &ast.Identifier{Name: "supertrend"},
	}
}

func stCall(factor float64, atrPeriod int) *ast.CallExpression {
	return &ast.CallExpression{
		Callee: stCallee(),
		Arguments: []ast.Expression{
			&ast.Literal{Value: factor},
			&ast.Literal{Value: float64(atrPeriod)},
		},
	}
}
