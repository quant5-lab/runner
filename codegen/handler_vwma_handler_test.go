package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestVWMAHandler_CanHandle(t *testing.T) {
	handler := &VWMAHandler{}

	tests := []struct {
		funcName string
		expected bool
	}{
		{"ta.vwma", true},
		{"vwma", true},
		{"ta.sma", false},
		{"ta.wma", false},
		{"vwap", false},
	}

	for _, tt := range tests {
		t.Run(tt.funcName, func(t *testing.T) {
			if got := handler.CanHandle(tt.funcName); got != tt.expected {
				t.Errorf("CanHandle(%q) = %v, want %v", tt.funcName, got, tt.expected)
			}
		})
	}
}

func TestVWMAHandler_BasicGeneration(t *testing.T) {
	gen := createTestGenerator()
	handler := &VWMAHandler{}

	call := &ast.CallExpression{
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "close"},
			&ast.Literal{Value: 20},
		},
	}

	code, err := handler.GenerateCode(gen, "vwma20", call)
	if err != nil {
		t.Fatalf("GenerateCode() error = %v", err)
	}

	requiredPatterns := []string{
		"weightedSum",
		"volumeSum",
		"bar.Volume",
	}

	for _, pattern := range requiredPatterns {
		if !strings.Contains(code, pattern) {
			t.Errorf("Generated code missing pattern: %q\nCode:\n%s", pattern, code)
		}
	}
}

func TestVWMAHandler_VolumeWeighting(t *testing.T) {
	gen := createTestGenerator()
	handler := &VWMAHandler{}

	call := &ast.CallExpression{
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "close"},
			&ast.Literal{Value: 10},
		},
	}

	code, err := handler.GenerateCode(gen, "result", call)
	if err != nil {
		t.Fatalf("GenerateCode() error = %v", err)
	}

	if !strings.Contains(code, "bar.Volume") {
		t.Errorf("VWMA must access bar.Volume:\n%s", code)
	}

	if !strings.Contains(code, "volumeSum") {
		t.Errorf("VWMA must accumulate volumes:\n%s", code)
	}
}

func TestVWMAHandler_NaNHandling(t *testing.T) {
	gen := createTestGenerator()
	handler := &VWMAHandler{}

	call := &ast.CallExpression{
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "close"},
			&ast.Literal{Value: 5},
		},
	}

	code, err := handler.GenerateCode(gen, "v", call)
	if err != nil {
		t.Fatalf("GenerateCode() error = %v", err)
	}

	if !strings.Contains(code, "hasNaN") {
		t.Errorf("VWMA must check for NaN values:\n%s", code)
	}

	if !strings.Contains(code, "math.IsNaN") {
		t.Errorf("VWMA must use math.IsNaN for validation:\n%s", code)
	}
}

func TestVWMAHandler_BareAlias(t *testing.T) {
	gen := createTestGenerator()
	handler := &VWMAHandler{}

	if !handler.CanHandle("vwma") {
		t.Error("Handler must accept bare vwma alias")
	}

	call := &ast.CallExpression{
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "close"},
			&ast.Literal{Value: 14},
		},
	}

	code, err := handler.GenerateCode(gen, "v", call)
	if err != nil {
		t.Fatalf("GenerateCode() with bare alias error = %v", err)
	}

	if !strings.Contains(code, "weightedSum") || !strings.Contains(code, "volumeSum") {
		t.Errorf("Bare vwma() must generate same code as ta.vwma():\n%s", code)
	}
}

func TestVWMAHandler_RegistryIntegration(t *testing.T) {
	registry := NewTAFunctionRegistry()

	if !registry.IsSupported("ta.vwma") {
		t.Error("TAFunctionRegistry must support ta.vwma")
	}

	if !registry.IsSupported("vwma") {
		t.Error("TAFunctionRegistry must support bare vwma")
	}

	handler := registry.FindHandler("ta.vwma")
	if handler == nil {
		t.Fatal("FindHandler(ta.vwma) returned nil")
	}

	if _, ok := handler.(*VWMAHandler); !ok {
		t.Errorf("FindHandler(ta.vwma) returned wrong type: %T", handler)
	}
}

func TestVWMAHandler_ArgumentValidation(t *testing.T) {
	gen := createTestGenerator()
	handler := &VWMAHandler{}

	t.Run("no_arguments", func(t *testing.T) {
		call := &ast.CallExpression{
			Arguments: []ast.Expression{},
		}

		_, err := handler.GenerateCode(gen, "result", call)
		if err == nil {
			t.Error("Expected error for zero arguments")
		}
	})

	t.Run("one_argument_only", func(t *testing.T) {
		call := &ast.CallExpression{
			Arguments: []ast.Expression{
				&ast.Identifier{Name: "close"},
			},
		}

		_, err := handler.GenerateCode(gen, "result", call)
		if err == nil {
			t.Error("Expected error for only one argument (period missing)")
		}
	})
}
