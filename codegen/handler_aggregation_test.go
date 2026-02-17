package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

/* MaxHandler tests */

func TestMaxHandler_TAFunctionHandlerInterface(t *testing.T) {
	handler := &MaxHandler{}

	t.Run("can_handle_ta_dot_max", func(t *testing.T) {
		if !handler.CanHandle("ta.max") {
			t.Error("MaxHandler should handle 'ta.max'")
		}
	})

	t.Run("cannot_handle_bare_max", func(t *testing.T) {
		if handler.CanHandle("max") {
			t.Error("MaxHandler should not handle 'max' (conflicts with math.max)")
		}
	})

	t.Run("cannot_handle_other_functions", func(t *testing.T) {
		invalidFuncs := []string{"ta.min", "math.max", "maximum", "ta.highest", ""}
		for _, fn := range invalidFuncs {
			if handler.CanHandle(fn) {
				t.Errorf("MaxHandler should not handle '%s'", fn)
			}
		}
	})
}

func TestMaxHandler_ArgumentValidation(t *testing.T) {
	handler := &MaxHandler{}
	gen := createTestGenerator()

	t.Run("valid_two_arguments", func(t *testing.T) {
		call := &ast.CallExpression{
			Arguments: []ast.Expression{
				&ast.Identifier{Name: "close"},
				&ast.Literal{Value: float64(14)},
			},
		}

		code, err := handler.GenerateCode(gen, "maxResult", call)
		if err != nil {
			t.Errorf("Expected no error for valid arguments, got: %v", err)
		}
		if code == "" {
			t.Error("Expected generated code, got empty string")
		}
	})

	t.Run("missing_arguments", func(t *testing.T) {
		call := &ast.CallExpression{
			Arguments: []ast.Expression{},
		}

		_, err := handler.GenerateCode(gen, "maxResult", call)
		if err == nil {
			t.Error("Expected error for missing arguments")
		}
	})
}

func TestMaxHandler_CodeGeneration(t *testing.T) {
	handler := &MaxHandler{}

	t.Run("identifier_source", func(t *testing.T) {
		gen := createTestGenerator()
		call := &ast.CallExpression{
			Arguments: []ast.Expression{
				&ast.Identifier{Name: "close"},
				&ast.Literal{Value: float64(20)},
			},
		}

		code, err := handler.GenerateCode(gen, "maxVal", call)
		if err != nil {
			t.Fatalf("GenerateCode() error = %v", err)
		}

		if !strings.Contains(code, "maxValSeries.Set") {
			t.Error("Generated code should call maxValSeries.Set")
		}
	})

	t.Run("warmup_period", func(t *testing.T) {
		gen := createTestGenerator()
		call := &ast.CallExpression{
			Arguments: []ast.Expression{
				&ast.Identifier{Name: "high"},
				&ast.Literal{Value: float64(14)},
			},
		}

		code, err := handler.GenerateCode(gen, "maxHigh", call)
		if err != nil {
			t.Fatalf("GenerateCode() error = %v", err)
		}

		if !strings.Contains(code, "ctx.BarIndex < 13") {
			t.Error("Generated code should check warmup period (period-1)")
		}

		if !strings.Contains(code, "math.NaN()") {
			t.Error("Generated code should return NaN during warmup")
		}
	})
}

/* MinHandler tests */

func TestMinHandler_TAFunctionHandlerInterface(t *testing.T) {
	handler := &MinHandler{}

	t.Run("can_handle_ta_dot_min", func(t *testing.T) {
		if !handler.CanHandle("ta.min") {
			t.Error("MinHandler should handle 'ta.min'")
		}
	})

	t.Run("cannot_handle_bare_min", func(t *testing.T) {
		if handler.CanHandle("min") {
			t.Error("MinHandler should not handle 'min' (conflicts with math.min)")
		}
	})

	t.Run("cannot_handle_other_functions", func(t *testing.T) {
		invalidFuncs := []string{"ta.max", "math.min", "minimum", "ta.lowest", ""}
		for _, fn := range invalidFuncs {
			if handler.CanHandle(fn) {
				t.Errorf("MinHandler should not handle '%s'", fn)
			}
		}
	})
}

func TestMinHandler_CodeGeneration(t *testing.T) {
	handler := &MinHandler{}

	t.Run("generates_valid_code", func(t *testing.T) {
		gen := createTestGenerator()
		call := &ast.CallExpression{
			Arguments: []ast.Expression{
				&ast.Identifier{Name: "low"},
				&ast.Literal{Value: float64(10)},
			},
		}

		code, err := handler.GenerateCode(gen, "minLow", call)
		if err != nil {
			t.Fatalf("GenerateCode() error = %v", err)
		}

		if !strings.Contains(code, "minLowSeries.Set") {
			t.Error("Generated code should call minLowSeries.Set")
		}

		if !strings.Contains(code, "ctx.BarIndex < 9") {
			t.Error("Generated code should check warmup period (period-1)")
		}
	})
}

/* MedianHandler tests */

func TestMedianHandler_TAFunctionHandlerInterface(t *testing.T) {
	handler := &MedianHandler{}

	t.Run("can_handle_ta_dot_median", func(t *testing.T) {
		if !handler.CanHandle("ta.median") {
			t.Error("MedianHandler should handle 'ta.median'")
		}
	})

	t.Run("can_handle_bare_median", func(t *testing.T) {
		if !handler.CanHandle("median") {
			t.Error("MedianHandler should handle 'median' (no namespace conflict)")
		}
	})

	t.Run("cannot_handle_other_functions", func(t *testing.T) {
		invalidFuncs := []string{"ta.mean", "average", "mid", ""}
		for _, fn := range invalidFuncs {
			if handler.CanHandle(fn) {
				t.Errorf("MedianHandler should not handle '%s'", fn)
			}
		}
	})
}

func TestMedianHandler_RequiresSortImport(t *testing.T) {
	handler := &MedianHandler{}
	gen := createTestGenerator()

	call := &ast.CallExpression{
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "close"},
			&ast.Literal{Value: float64(14)},
		},
	}

	_, err := handler.GenerateCode(gen, "medianVal", call)
	if err != nil {
		t.Fatalf("GenerateCode() error = %v", err)
	}

	if !gen.hasSortUsage {
		t.Error("MedianHandler should set hasSortUsage flag for sort import")
	}
}

func TestMedianHandler_CodeGeneration(t *testing.T) {
	handler := &MedianHandler{}

	t.Run("generates_sorting_logic", func(t *testing.T) {
		gen := createTestGenerator()
		call := &ast.CallExpression{
			Arguments: []ast.Expression{
				&ast.Identifier{Name: "close"},
				&ast.Literal{Value: float64(5)},
			},
		}

		code, err := handler.GenerateCode(gen, "med", call)
		if err != nil {
			t.Fatalf("GenerateCode() error = %v", err)
		}

		if !strings.Contains(code, "sort.Float64s") {
			t.Error("Generated code should use sort.Float64s for median calculation")
		}

		if !strings.Contains(code, "func()") {
			t.Error("Generated code should use IIFE for median calculation")
		}
	})
}

/* VarianceHandler tests */

func TestVarianceHandler_TAFunctionHandlerInterface(t *testing.T) {
	handler := &VarianceHandler{}

	t.Run("can_handle_ta_dot_variance", func(t *testing.T) {
		if !handler.CanHandle("ta.variance") {
			t.Error("VarianceHandler should handle 'ta.variance'")
		}
	})

	t.Run("can_handle_bare_variance", func(t *testing.T) {
		if !handler.CanHandle("variance") {
			t.Error("VarianceHandler should handle 'variance'")
		}
	})
}

func TestVarianceHandler_CodeGeneration(t *testing.T) {
	handler := &VarianceHandler{}

	t.Run("generates_two_pass_algorithm", func(t *testing.T) {
		gen := createTestGenerator()
		call := &ast.CallExpression{
			Arguments: []ast.Expression{
				&ast.Identifier{Name: "close"},
				&ast.Literal{Value: float64(10)},
			},
		}

		code, err := handler.GenerateCode(gen, "variance", call)
		if err != nil {
			t.Fatalf("GenerateCode() error = %v", err)
		}

		if !strings.Contains(code, "sum") && !strings.Contains(code, "mean") {
			t.Error("Generated code should calculate mean for variance")
		}

		if !strings.Contains(code, "func()") {
			t.Error("Generated code should use IIFE for variance calculation")
		}
	})
}

/* RangeHandler tests */

func TestRangeHandler_TAFunctionHandlerInterface(t *testing.T) {
	handler := &RangeHandler{}

	t.Run("can_handle_ta_dot_range", func(t *testing.T) {
		if !handler.CanHandle("ta.range") {
			t.Error("RangeHandler should handle 'ta.range'")
		}
	})

	t.Run("can_handle_bare_range", func(t *testing.T) {
		if !handler.CanHandle("range") {
			t.Error("RangeHandler should handle 'range'")
		}
	})
}

func TestRangeHandler_CodeGeneration(t *testing.T) {
	handler := &RangeHandler{}

	t.Run("generates_max_minus_min", func(t *testing.T) {
		gen := createTestGenerator()
		call := &ast.CallExpression{
			Arguments: []ast.Expression{
				&ast.Identifier{Name: "close"},
				&ast.Literal{Value: float64(14)},
			},
		}

		code, err := handler.GenerateCode(gen, "rangeVal", call)
		if err != nil {
			t.Fatalf("GenerateCode() error = %v", err)
		}

		if !strings.Contains(code, "math.Inf(-1)") {
			t.Error("Generated code should track max with -Inf initialization")
		}

		if !strings.Contains(code, "math.Inf(1)") {
			t.Error("Generated code should track min with +Inf initialization")
		}

		// RangeAccumulator uses inline calculation, not IIFE
		if !strings.Contains(code, "rangeValSeries.Set") {
			t.Error("Generated code should set range value")
		}
	})
}

/* ModeHandler tests */

func TestModeHandler_TAFunctionHandlerInterface(t *testing.T) {
	handler := &ModeHandler{}

	t.Run("can_handle_ta_dot_mode", func(t *testing.T) {
		if !handler.CanHandle("ta.mode") {
			t.Error("ModeHandler should handle 'ta.mode'")
		}
	})

	t.Run("can_handle_bare_mode", func(t *testing.T) {
		if !handler.CanHandle("mode") {
			t.Error("ModeHandler should handle 'mode'")
		}
	})
}

func TestModeHandler_RequiresSortImport(t *testing.T) {
	handler := &ModeHandler{}
	gen := createTestGenerator()

	call := &ast.CallExpression{
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "close"},
			&ast.Literal{Value: float64(10)},
		},
	}

	_, err := handler.GenerateCode(gen, "modeVal", call)
	if err != nil {
		t.Fatalf("GenerateCode() error = %v", err)
	}

	if !gen.hasSortUsage {
		t.Error("ModeHandler should set hasSortUsage flag for sort import")
	}
}

func TestModeHandler_CodeGeneration(t *testing.T) {
	handler := &ModeHandler{}

	t.Run("generates_frequency_counting", func(t *testing.T) {
		gen := createTestGenerator()
		call := &ast.CallExpression{
			Arguments: []ast.Expression{
				&ast.Identifier{Name: "volume"},
				&ast.Literal{Value: float64(20)},
			},
		}

		code, err := handler.GenerateCode(gen, "modeVol", call)
		if err != nil {
			t.Fatalf("GenerateCode() error = %v", err)
		}

		if !strings.Contains(code, "make(map[float64]int)") {
			t.Error("Generated code should use frequency map")
		}

		if !strings.Contains(code, "func()") {
			t.Error("Generated code should use IIFE for mode calculation")
		}
	})
}

/* Registry integration tests */

func TestAggregationHandlers_IntegrationWithTARegistry(t *testing.T) {
	registry := NewTAFunctionRegistry()

	t.Run("registry_has_all_aggregation_handlers", func(t *testing.T) {
		functions := []string{"ta.max", "ta.min", "ta.median", "ta.variance", "ta.range", "ta.mode"}
		for _, fn := range functions {
			handler := registry.FindHandler(fn)
			if handler == nil {
				t.Errorf("TAFunctionRegistry missing handler for %s", fn)
			}
		}
	})

	t.Run("registry_supports_all_aggregation_functions", func(t *testing.T) {
		functions := []string{"ta.max", "ta.min", "ta.median", "ta.variance", "ta.range", "ta.mode"}
		for _, fn := range functions {
			if !registry.IsSupported(fn) {
				t.Errorf("TAFunctionRegistry should support '%s'", fn)
			}
		}
	})
}

func TestAggregationHandlers_EdgeCases(t *testing.T) {
	t.Run("max_with_literal_period", func(t *testing.T) {
		handler := &MaxHandler{}
		gen := createTestGenerator()

		call := &ast.CallExpression{
			Arguments: []ast.Expression{
				&ast.Identifier{Name: "high"},
				&ast.Literal{Value: float64(14)},
			},
		}

		code, err := handler.GenerateCode(gen, "maxHigh", call)
		if err != nil {
			t.Fatalf("GenerateCode() with literal period error = %v", err)
		}

		if code == "" {
			t.Error("Should generate code for literal period")
		}
	})

	t.Run("median_with_complex_source_expression", func(t *testing.T) {
		handler := &MedianHandler{}
		gen := createTestGenerator()

		call := &ast.CallExpression{
			Arguments: []ast.Expression{
				&ast.BinaryExpression{
					Left:     &ast.Identifier{Name: "close"},
					Operator: "-",
					Right:    &ast.Identifier{Name: "open"},
				},
				&ast.Literal{Value: float64(5)},
			},
		}

		code, err := handler.GenerateCode(gen, "medianDiff", call)
		if err != nil {
			t.Fatalf("GenerateCode() with complex expression error = %v", err)
		}

		if code == "" {
			t.Error("Should generate code for complex source expression")
		}
	})
}
