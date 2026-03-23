package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestBarsSinceHandler_CanHandle(t *testing.T) {
	handler := &BarsSinceHandler{}

	tests := []struct {
		funcName string
		want     bool
	}{
		{"ta.barssince", true},
		{"barssince", true},
		{"ta.sma", false},
		{"ta.change", false},
		{"barsSince", false},
	}

	for _, tt := range tests {
		t.Run(tt.funcName, func(t *testing.T) {
			if got := handler.CanHandle(tt.funcName); got != tt.want {
				t.Errorf("CanHandle(%q) = %v, want %v", tt.funcName, got, tt.want)
			}
		})
	}
}

func TestBarsSinceHandler_GenerateCode_ArgumentValidation(t *testing.T) {
	handler := &BarsSinceHandler{}
	g := newTestGenerator()

	call := &ast.CallExpression{
		Callee:    &ast.Identifier{Name: "ta.barssince"},
		Arguments: []ast.Expression{},
	}

	_, err := handler.GenerateCode(g, "K1", call)
	if err == nil {
		t.Fatal("expected error for zero arguments, got nil")
	}
	if !strings.Contains(err.Error(), "requires 1 argument") {
		t.Errorf("error = %q, want substring %q", err.Error(), "requires 1 argument")
	}
}

func TestBarsSinceHandler_GenerateCode_IdentifierCondition(t *testing.T) {
	handler := &BarsSinceHandler{}
	g := newTestGenerator()
	g.variables["buySignal"] = "series"

	call := &ast.CallExpression{
		Callee: &ast.Identifier{Name: "ta.barssince"},
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "buySignal"},
		},
	}

	code, err := handler.GenerateCode(g, "K1", call)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedFragments := []string{
		"value.IsTrue(buySignalSeries.GetCurrent())",
		"K1Series.Set(0.0)",
		"K1Series.Get(1)",
		"K1Series.Set(K1Series.Get(1) + 1.0)",
		"math.IsNaN(K1Series.Get(1))",
		"K1Series.Set(math.NaN())",
	}

	for _, frag := range expectedFragments {
		if !strings.Contains(code, frag) {
			t.Errorf("missing expected fragment %q in generated code:\n%s", frag, code)
		}
	}
}

func TestBarsSinceHandler_GenerateCode_BarFieldCondition(t *testing.T) {
	handler := &BarsSinceHandler{}
	g := newTestGenerator()

	call := &ast.CallExpression{
		Callee: &ast.Identifier{Name: "ta.barssince"},
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "close"},
		},
	}

	code, err := handler.GenerateCode(g, "result", call)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(code, "value.IsTrue(") {
		t.Errorf("expected value.IsTrue() wrapper in generated code:\n%s", code)
	}
	if !strings.Contains(code, "resultSeries.Set(0.0)") {
		t.Errorf("expected resultSeries.Set(0.0) in generated code:\n%s", code)
	}
}

func TestBarsSinceHandler_GenerateCode_Structure(t *testing.T) {
	handler := &BarsSinceHandler{}
	g := newTestGenerator()
	g.variables["cond"] = "series"

	call := &ast.CallExpression{
		Callee: &ast.Identifier{Name: "barssince"},
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "cond"},
		},
	}

	code, err := handler.GenerateCode(g, "bs", call)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	/* Verify three-branch structure: condition true, counter increment, NaN */
	ifCount := strings.Count(code, "} else if")
	elseCount := strings.Count(code, "} else {")
	if ifCount != 1 {
		t.Errorf("expected 1 'else if' branch, got %d in:\n%s", ifCount, code)
	}
	if elseCount != 1 {
		t.Errorf("expected 1 'else' branch, got %d in:\n%s", elseCount, code)
	}
}

func TestBarsSinceHandler_RegisteredInTAFunctionRegistry(t *testing.T) {
	registry := NewTAFunctionRegistry()

	for _, name := range []string{"ta.barssince", "barssince"} {
		handler := registry.FindHandler(name)
		if handler == nil {
			t.Errorf("expected handler for %q in registry, got nil", name)
		}
	}
}

func TestBarsSinceHandler_HistoricalSubscriptCondition(t *testing.T) {
	handler := &BarsSinceHandler{}

	tests := []struct {
		name   string
		offset int
		want   string
	}{
		{"offset_1", 1, "signalSeries.Get(1)"},
		{"offset_2", 2, "signalSeries.Get(2)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newTestGenerator()
			g.variables["signal"] = "series"

			call := &ast.CallExpression{
				Callee: &ast.Identifier{Name: "ta.barssince"},
				Arguments: []ast.Expression{
					&ast.MemberExpression{
						Object:   &ast.Identifier{Name: "signal"},
						Property: &ast.Literal{Value: float64(tt.offset)},
						Computed: true,
					},
				},
			}

			code, err := handler.GenerateCode(g, "K1", call)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !strings.Contains(code, "value.IsTrue("+tt.want+")") {
				t.Errorf("want value.IsTrue(%s), got:\n%s", tt.want, code)
			}
			if !strings.Contains(code, "K1Series.Set(0.0)") {
				t.Errorf("missing counter reset in:\n%s", code)
			}
			if !strings.Contains(code, "K1Series.Set(math.NaN())") {
				t.Errorf("missing NaN sentinel in:\n%s", code)
			}
		})
	}
}
