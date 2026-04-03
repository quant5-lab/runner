package codegen

import (
	"fmt"
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

// TestATRHandler_CanHandle verifies the handler accepts "ta.atr" and bare "atr"
// (v4 alias) and rejects all other function names.
func TestATRHandler_CanHandle(t *testing.T) {
	handler := &ATRHandler{}

	for _, name := range []string{"ta.atr", "atr"} {
		if !handler.CanHandle(name) {
			t.Errorf("ATRHandler.CanHandle(%q) = false, want true", name)
		}
	}

	for _, name := range []string{"ta.rma", "ta.tr", "ta.ema", "ta.sma", "ta.rsi", ""} {
		if handler.CanHandle(name) {
			t.Errorf("ATRHandler.CanHandle(%q) = true, want false", name)
		}
	}
}

// TestATRHandler_MissingPeriodArgError verifies that an ATR call with no arguments
// returns an error rather than panicking or producing incomplete code.
func TestATRHandler_MissingPeriodArgError(t *testing.T) {
	handler := &ATRHandler{}
	g := newTestGenerator()

	call := &ast.CallExpression{
		Callee:    &ast.MemberExpression{Object: &ast.Identifier{Name: "ta"}, Property: &ast.Identifier{Name: "atr"}},
		Arguments: nil,
	}

	_, err := handler.GenerateCode(g, "atr_var", call)
	if err == nil {
		t.Error("expected error for missing period argument, got nil")
	}
}

// TestATRHandler_StaticPeriod_ThreePhaseRMAStructure verifies that ATR with a
// compile-time-constant period generates the canonical Pine three-phase RMA:
//
//	Phase 1 — NaN warmup:    ctx.BarIndex < period-1
//	Phase 2 — SMA seed:      ctx.BarIndex == period-1, loop 0..period-1
//	Phase 3 — RMA recursive: alpha=1/period, newValue = alpha*src + (1-alpha)*prev
func TestATRHandler_StaticPeriod_ThreePhaseRMAStructure(t *testing.T) {
	for _, period := range []int{1, 2, 5, 14, 50} {
		period := period
		t.Run(fmt.Sprintf("period_%d", period), func(t *testing.T) {
			handler := &ATRHandler{}
			g := newTestGenerator()

			call := &ast.CallExpression{
				Callee:    &ast.MemberExpression{Object: &ast.Identifier{Name: "ta"}, Property: &ast.Identifier{Name: "atr"}},
				Arguments: []ast.Expression{&ast.Literal{Value: float64(period)}},
			}

			code, err := handler.GenerateCode(g, "atr_out", call)
			if err != nil {
				t.Fatalf("GenerateCode() error = %v", err)
			}

			for _, want := range []string{
				fmt.Sprintf("ctx.BarIndex < %d", period-1),
				fmt.Sprintf("ctx.BarIndex == %d", period-1),
				fmt.Sprintf("alpha := 1.0 / float64(%d)", period),
				fmt.Sprintf("initialValue := _sma_accumulator / float64(%d)", period),
				"newValue := alpha*currentSource + (1-alpha)*previousValue",
			} {
				if !strings.Contains(code, want) {
					t.Errorf("missing %q\n%s", want, code)
				}
			}
		})
	}
}

// TestATRHandler_TRHandleNATrue verifies that the TR accessor embedded in the
// ATR codegen uses handle_na=true: bar 0 returns high-low so the SMA seed window
// covering bars 0..period-1 is always fully valid and never produces NaN output.
func TestATRHandler_TRHandleNATrue(t *testing.T) {
	handler := &ATRHandler{}
	g := newTestGenerator()

	call := &ast.CallExpression{
		Callee:    &ast.MemberExpression{Object: &ast.Identifier{Name: "ta"}, Property: &ast.Identifier{Name: "atr"}},
		Arguments: []ast.Expression{&ast.Literal{Value: float64(14)}},
	}

	code, err := handler.GenerateCode(g, "atr_out", call)
	if err != nil {
		t.Fatalf("GenerateCode() error = %v", err)
	}

	if !strings.Contains(code, "if idx == 0 { return h - l }") {
		t.Errorf("missing handleNA=true branch 'if idx == 0 { return h - l }'\n%s", code)
	}
	if strings.Contains(code, "return math.NaN()") {
		t.Errorf("ATR TR accessor must not return NaN on bar 0 (handle_na=true)\n%s", code)
	}
}

// TestATRHandler_SeriesAccessByContext verifies that ATR uses direct varSeries.Set
// in top-level scope and arrowCtx.GetOrCreateSeries inside an arrow function body,
// ensuring series state is correctly scoped in each execution context.
func TestATRHandler_SeriesAccessByContext(t *testing.T) {
	tests := []struct {
		name         string
		inArrow      bool
		wantContains string
		wantAbsent   string
	}{
		{
			name:         "top_level",
			inArrow:      false,
			wantContains: "atr_outSeries.Set(",
			wantAbsent:   "arrowCtx.GetOrCreateSeries",
		},
		{
			name:         "arrow_function_body",
			inArrow:      true,
			wantContains: "arrowCtx.GetOrCreateSeries",
			wantAbsent:   "atr_outSeries.Set(",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			handler := &ATRHandler{}
			g := newTestGenerator()
			g.inArrowFunctionBody = tt.inArrow

			call := &ast.CallExpression{
				Callee:    &ast.MemberExpression{Object: &ast.Identifier{Name: "ta"}, Property: &ast.Identifier{Name: "atr"}},
				Arguments: []ast.Expression{&ast.Literal{Value: float64(14)}},
			}

			code, err := handler.GenerateCode(g, "atr_out", call)
			if err != nil {
				t.Fatalf("GenerateCode() error = %v", err)
			}

			if !strings.Contains(code, tt.wantContains) {
				t.Errorf("missing %q\n%s", tt.wantContains, code)
			}
			if strings.Contains(code, tt.wantAbsent) {
				t.Errorf("unexpected %q\n%s", tt.wantAbsent, code)
			}
		})
	}
}

// TestATRHandler_NoOuterIIFE verifies that ATR is never wrapped in an outer
// arrow-function IIFE.  TR computation internally uses func() float64 lambdas,
// but the ATR indicator itself must remain top-level stateful code.
func TestATRHandler_NoOuterIIFE(t *testing.T) {
	handler := &ATRHandler{}
	g := newTestGenerator()

	call := &ast.CallExpression{
		Callee:    &ast.MemberExpression{Object: &ast.Identifier{Name: "ta"}, Property: &ast.Identifier{Name: "atr"}},
		Arguments: []ast.Expression{&ast.Literal{Value: float64(14)}},
	}

	code, err := handler.GenerateCode(g, "atr_out", call)
	if err != nil {
		t.Fatalf("GenerateCode() error = %v", err)
	}

	if strings.Contains(code, "return func") {
		t.Errorf("ATR must not be wrapped in an outer IIFE\n%s", code)
	}
}
