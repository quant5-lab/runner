package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestMFIHandler_CanHandle(t *testing.T) {
	handler := &MFIHandler{}

	tests := []struct {
		funcName string
		want     bool
	}{
		{"ta.mfi", true},
		{"mfi", true},
		{"ta.rsi", false},
		{"ta.sma", false},
	}

	for _, tt := range tests {
		t.Run(tt.funcName, func(t *testing.T) {
			if got := handler.CanHandle(tt.funcName); got != tt.want {
				t.Errorf("CanHandle(%q) = %v, want %v", tt.funcName, got, tt.want)
			}
		})
	}
}

func TestMFIHandler_GenerateCode_ArgumentValidation(t *testing.T) {
	handler := &MFIHandler{}
	g := newTestGenerator()

	tests := []struct {
		name    string
		args    []ast.Expression
		wantErr string
	}{
		{
			name:    "no arguments",
			args:    []ast.Expression{},
			wantErr: "requires at least 2 arguments",
		},
		{
			name: "one argument",
			args: []ast.Expression{
				&ast.Identifier{Name: "close"},
			},
			wantErr: "requires at least 2 arguments",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			call := &ast.CallExpression{
				Callee:    &ast.Identifier{Name: "ta.mfi"},
				Arguments: tt.args,
			}

			_, err := handler.GenerateCode(g, "test", call)
			if err == nil {
				t.Error("expected error, got nil")
				return
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("error = %q, want substring %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestMFIHandler_GenerateCode_OHLCVSource(t *testing.T) {
	handler := &MFIHandler{}
	g := newTestGenerator()

	call := &ast.CallExpression{
		Callee: &ast.Identifier{Name: "ta.mfi"},
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "close"},
			&ast.Literal{Value: float64(14)},
		},
	}

	code, err := handler.GenerateCode(g, "mfi14", call)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedFragments := []string{
		"_mfi14_change",
		"bar.Volume",
		"_mfi14_posMF",
		"_mfi14_negMF",
		"_mfi14_positive_mfSeries",
		"_mfi14_negative_mfSeries",
		"posSum",
		"negSum",
		"mfr := posSum / negSum",
		"100.0 - (100.0 / (1.0 + mfr))",
		"mfi14Series",
		"ctx.BarIndex < 14",
		"math.NaN()",
	}

	for _, frag := range expectedFragments {
		if !strings.Contains(code, frag) {
			t.Errorf("missing expected fragment %q in generated code:\n%s", frag, code)
		}
	}
}

func TestMFIHandler_GenerateCode_DerivedPriceSource(t *testing.T) {
	handler := &MFIHandler{}
	g := newTestGenerator()

	call := &ast.CallExpression{
		Callee: &ast.Identifier{Name: "ta.mfi"},
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "hlc3"},
			&ast.Literal{Value: float64(14)},
		},
	}

	code, err := handler.GenerateCode(g, "mfi_hlc3", call)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(code, "high") || !strings.Contains(code, "low") || !strings.Contains(code, "close") {
		t.Errorf("expected derived price formula components in generated code:\n%s", code)
	}

	if !strings.Contains(code, "bar.Volume") {
		t.Errorf("expected bar.Volume access in generated code:\n%s", code)
	}
}

func TestMFIHandler_GenerateCode_ThreeBranchMoneyFlowSplit(t *testing.T) {
	handler := &MFIHandler{}
	g := newTestGenerator()

	call := &ast.CallExpression{
		Callee: &ast.Identifier{Name: "ta.mfi"},
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "close"},
			&ast.Literal{Value: float64(14)},
		},
	}

	code, err := handler.GenerateCode(g, "mfi14", call)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(code, "math.IsNaN(") {
		t.Errorf("expected NaN check for change in money flow split:\n%s", code)
	}
	if !strings.Contains(code, "> 0") {
		t.Errorf("expected positive change check in money flow split:\n%s", code)
	}
}

func TestMFIHandler_GenerateCode_ZeroDenominatorProtection(t *testing.T) {
	handler := &MFIHandler{}
	g := newTestGenerator()

	call := &ast.CallExpression{
		Callee: &ast.Identifier{Name: "ta.mfi"},
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "close"},
			&ast.Literal{Value: float64(14)},
		},
	}

	code, err := handler.GenerateCode(g, "mfi14", call)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(code, "negSum == 0") {
		t.Errorf("expected zero-denominator protection in generated code:\n%s", code)
	}
	if !strings.Contains(code, "100.0") {
		t.Errorf("expected 100.0 for all-positive flows:\n%s", code)
	}
}

func TestMFIHandler_GetInternalSeriesNames(t *testing.T) {
	handler := &MFIHandler{}
	call := &ast.CallExpression{
		Callee: &ast.Identifier{Name: "ta.mfi"},
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "close"},
			&ast.Literal{Value: float64(14)},
		},
	}

	names, err := handler.GetInternalSeriesNames("mfi14", call)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(names) != 2 {
		t.Fatalf("expected 2 internal series, got %d", len(names))
	}
	if names[0] != "_mfi14_positive_mf" {
		t.Errorf("expected _mfi14_positive_mf, got %q", names[0])
	}
	if names[1] != "_mfi14_negative_mf" {
		t.Errorf("expected _mfi14_negative_mf, got %q", names[1])
	}
}

func TestMFIHandler_RegisteredInTAFunctionRegistry(t *testing.T) {
	registry := NewTAFunctionRegistry()

	for _, name := range []string{"ta.mfi", "mfi"} {
		handler := registry.FindHandler(name)
		if handler == nil {
			t.Errorf("expected handler for %q in registry, got nil", name)
		}
	}
}
