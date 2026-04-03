package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestCumHandler_TAFunctionHandlerInterface(t *testing.T) {
	handler := &CumHandler{}

	t.Run("can_handle_ta_dot_cum", func(t *testing.T) {
		if !handler.CanHandle("ta.cum") {
			t.Error("CumHandler should handle 'ta.cum'")
		}
	})

	t.Run("can_handle_cum", func(t *testing.T) {
		if !handler.CanHandle("cum") {
			t.Error("CumHandler should handle 'cum' (bare name)")
		}
	})

	t.Run("cannot_handle_other_functions", func(t *testing.T) {
		invalidFuncs := []string{"ta.sma", "ta.sum", "cumulative", "ta.change", ""}
		for _, fn := range invalidFuncs {
			if handler.CanHandle(fn) {
				t.Errorf("CumHandler should not handle '%s'", fn)
			}
		}
	})
}

func TestCumHandler_ArgumentValidation(t *testing.T) {
	handler := &CumHandler{}
	gen := createTestGenerator()

	t.Run("no_arguments", func(t *testing.T) {
		call := &ast.CallExpression{
			Arguments: []ast.Expression{},
		}

		_, err := handler.GenerateCode(gen, "result", call)
		if err == nil {
			t.Error("Expected error for zero arguments")
		}
		if !strings.Contains(err.Error(), "requires exactly 1 argument") {
			t.Errorf("Error should mention 'requires exactly 1 argument', got: %v", err)
		}
	})

	t.Run("too_many_arguments", func(t *testing.T) {
		call := &ast.CallExpression{
			Arguments: []ast.Expression{
				&ast.Identifier{Name: "close"},
				&ast.Literal{Value: 10},
			},
		}

		_, err := handler.GenerateCode(gen, "result", call)
		if err == nil {
			t.Error("Expected error for too many arguments")
		}
		if !strings.Contains(err.Error(), "requires exactly 1 argument") {
			t.Errorf("Error should mention 'requires exactly 1 argument', got: %v", err)
		}
	})
}

func TestCumHandler_CodeGeneration(t *testing.T) {
	handler := &CumHandler{}

	t.Run("identifier_source", func(t *testing.T) {
		gen := createTestGenerator()
		call := &ast.CallExpression{
			Arguments: []ast.Expression{
				&ast.Identifier{Name: "close"},
			},
		}

		code, err := handler.GenerateCode(gen, "cumResult", call)
		if err != nil {
			t.Fatalf("GenerateCode() error = %v", err)
		}

		if !strings.Contains(code, "ta.cum") {
			t.Error("Generated code should contain 'ta.cum' comment")
		}
		if !strings.Contains(code, "cumResultSeries.Set") {
			t.Error("Generated code should call cumResultSeries.Set")
		}
		if !strings.Contains(code, "prevSum") {
			t.Error("Generated code should use prevSum for accumulation")
		}
	})

	t.Run("member_expression_source", func(t *testing.T) {
		gen := createTestGenerator()
		call := &ast.CallExpression{
			Arguments: []ast.Expression{
				&ast.MemberExpression{
					Object:   &ast.Identifier{Name: "bar"},
					Property: &ast.Identifier{Name: "Close"},
				},
			},
		}

		code, err := handler.GenerateCode(gen, "myCum", call)
		if err != nil {
			t.Fatalf("GenerateCode() error = %v", err)
		}

		if !strings.Contains(code, "myCumSeries.Set") {
			t.Error("Generated code should use provided variable name")
		}
	})

	t.Run("nan_handling", func(t *testing.T) {
		gen := createTestGenerator()
		call := &ast.CallExpression{
			Arguments: []ast.Expression{
				&ast.Identifier{Name: "volume"},
			},
		}

		code, err := handler.GenerateCode(gen, "cumVol", call)
		if err != nil {
			t.Fatalf("GenerateCode() error = %v", err)
		}

		if !strings.Contains(code, "math.IsNaN") {
			t.Error("Generated code should handle NaN values")
		}
		if !strings.Contains(code, "cumVolSeries.Set(math.NaN())") {
			t.Error("Generated code should set NaN when source is NaN")
		}
	})

	t.Run("cumulative_logic", func(t *testing.T) {
		gen := createTestGenerator()
		call := &ast.CallExpression{
			Arguments: []ast.Expression{
				&ast.Identifier{Name: "x"},
			},
		}

		code, err := handler.GenerateCode(gen, "cumX", call)
		if err != nil {
			t.Fatalf("GenerateCode() error = %v", err)
		}

		if !strings.Contains(code, "cumXSeries.Get(1)") {
			t.Error("Generated code should retrieve previous cumulative value")
		}
		if !strings.Contains(code, "prevSum + current") {
			t.Error("Generated code should add current to previous sum")
		}
		if !strings.Contains(code, "if i > 0") {
			t.Error("Generated code should check for first bar")
		}
	})
}

func TestCumHandler_IntegrationWithTARegistry(t *testing.T) {
	registry := NewTAFunctionRegistry()

	t.Run("registry_has_cum_handler", func(t *testing.T) {
		handler := registry.FindHandler("ta.cum")
		if handler == nil {
			t.Fatal("TAFunctionRegistry missing Cum handler")
		}

		if _, ok := handler.(*CumHandler); !ok {
			t.Errorf("Registry returned wrong handler type: %T", handler)
		}
	})

	t.Run("registry_supports_cum", func(t *testing.T) {
		if !registry.IsSupported("ta.cum") {
			t.Error("TAFunctionRegistry should support 'ta.cum'")
		}

		if !registry.IsSupported("cum") {
			t.Error("TAFunctionRegistry should support 'cum'")
		}
	})

	t.Run("registry_generates_code", func(t *testing.T) {
		gen := createTestGenerator()
		call := &ast.CallExpression{
			Arguments: []ast.Expression{
				&ast.Identifier{Name: "close"},
			},
		}

		code, err := registry.GenerateInlineTA(gen, "test", "ta.cum", call)
		if err != nil {
			t.Fatalf("GenerateInlineTA() error = %v", err)
		}

		if code == "" {
			t.Error("GenerateInlineTA() returned empty string")
		}
	})
}

func TestCumHandler_EdgeCases(t *testing.T) {
	handler := &CumHandler{}
	gen := createTestGenerator()

	t.Run("empty_variable_name", func(t *testing.T) {
		call := &ast.CallExpression{
			Arguments: []ast.Expression{
				&ast.Identifier{Name: "close"},
			},
		}

		code, err := handler.GenerateCode(gen, "", call)
		if err != nil {
			t.Fatalf("Should handle empty varName, got error: %v", err)
		}
		if code == "" {
			t.Error("Should generate code even with empty varName")
		}
	})

	t.Run("complex_expression_source", func(t *testing.T) {
		call := &ast.CallExpression{
			Arguments: []ast.Expression{
				&ast.BinaryExpression{
					Left:     &ast.Identifier{Name: "close"},
					Operator: "-",
					Right:    &ast.Identifier{Name: "open"},
				},
			},
		}

		code, err := handler.GenerateCode(gen, "cumDiff", call)
		if err != nil {
			t.Fatalf("GenerateCode() error = %v", err)
		}

		if !strings.Contains(code, "cumDiffSeries.Set") {
			t.Error("Should generate code for complex expressions")
		}
	})
}
