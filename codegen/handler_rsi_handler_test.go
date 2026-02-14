package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestRSIHandler_TAFunctionHandlerInterface(t *testing.T) {
	handler := &RSIHandler{}

	t.Run("can_handle_ta_dot_rsi", func(t *testing.T) {
		if !handler.CanHandle("ta.rsi") {
			t.Error("RSIHandler should handle 'ta.rsi'")
		}
	})

	t.Run("can_handle_rsi", func(t *testing.T) {
		if !handler.CanHandle("rsi") {
			t.Error("RSIHandler should handle 'rsi' (Pine v4 syntax)")
		}
	})

	t.Run("cannot_handle_other_functions", func(t *testing.T) {
		invalidFuncs := []string{"ta.sma", "ta.ema", "rma", "ta.macd", ""}
		for _, fn := range invalidFuncs {
			if handler.CanHandle(fn) {
				t.Errorf("RSIHandler should not handle '%s'", fn)
			}
		}
	})
}

func TestRSIHandler_CodeGenerationDispatch(t *testing.T) {
	handler := &RSIHandler{}
	gen := newTestGenerator()

	call := &ast.CallExpression{
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "close"},
			&ast.Literal{Value: 14},
		},
	}

	code, err := handler.GenerateCode(gen, "myRSI", call)
	if err != nil {
		t.Fatalf("GenerateCode() error = %v", err)
	}

	if code == "" {
		t.Fatal("GenerateCode() returned empty string")
	}

	t.Run("contains_rsi_calculation", func(t *testing.T) {
		if !strings.Contains(code, "rs :=") && !strings.Contains(code, "rsi :=") {
			t.Error("missing RSI calculation logic")
		}
	})

	t.Run("contains_series_storage", func(t *testing.T) {
		if !strings.Contains(code, "Series.Set") {
			t.Error("Generated code missing Series.Set() for final RSI value")
		}
	})
}

func TestRSIHandler_IntegrationWithTARegistry(t *testing.T) {
	registry := NewTAFunctionRegistry()

	t.Run("registry_has_rsi_handler", func(t *testing.T) {
		handler := registry.FindHandler("ta.rsi")
		if handler == nil {
			t.Fatal("TAFunctionRegistry missing RSI handler")
		}

		if _, ok := handler.(*RSIHandler); !ok {
			t.Errorf("Registry returned wrong handler type: %T", handler)
		}
	})

	t.Run("registry_supports_rsi", func(t *testing.T) {
		if !registry.IsSupported("ta.rsi") {
			t.Error("TAFunctionRegistry should support 'ta.rsi'")
		}

		if !registry.IsSupported("rsi") {
			t.Error("TAFunctionRegistry should support 'rsi'")
		}
	})

	t.Run("registry_generates_code", func(t *testing.T) {
		gen := newTestGenerator()
		call := &ast.CallExpression{
			Arguments: []ast.Expression{
				&ast.Identifier{Name: "close"},
				&ast.Literal{Value: 14},
			},
		}

		code, err := registry.GenerateInlineTA(gen, "testRSI", "ta.rsi", call)
		if err != nil {
			t.Fatalf("GenerateInlineTA() error = %v", err)
		}

		if code == "" {
			t.Error("Registry dispatched to handler but got empty code")
		}

		if !strings.Contains(code, "RSI") {
			t.Error("Registry-generated code missing RSI logic")
		}
	})
}

func TestRSIHandler_CompositeIndicatorMetadataInterface(t *testing.T) {
	handler := &RSIHandler{}

	call := &ast.CallExpression{
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "close"},
			&ast.Literal{Value: 14},
		},
	}

	t.Run("returns_correct_series_count", func(t *testing.T) {
		seriesNames, err := handler.GetInternalSeriesNames("test", call)
		if err != nil {
			t.Fatalf("GetInternalSeriesNames() error = %v", err)
		}

		if len(seriesNames) != 4 {
			t.Errorf("Expected 4 internal series, got %d", len(seriesNames))
		}
	})

	t.Run("series_names_use_varname_prefix", func(t *testing.T) {
		seriesNames, err := handler.GetInternalSeriesNames("myRSI", call)
		if err != nil {
			t.Fatalf("GetInternalSeriesNames() error = %v", err)
		}

		expectedPrefix := "_myRSI_"
		for _, name := range seriesNames {
			if !strings.HasPrefix(name, expectedPrefix) {
				t.Errorf("Series name %q missing prefix %q", name, expectedPrefix)
			}
		}
	})

	t.Run("series_names_are_unique", func(t *testing.T) {
		seriesNames, err := handler.GetInternalSeriesNames("test", call)
		if err != nil {
			t.Fatalf("GetInternalSeriesNames() error = %v", err)
		}

		seen := make(map[string]bool)
		for _, name := range seriesNames {
			if seen[name] {
				t.Errorf("Duplicate series name: %s", name)
			}
			seen[name] = true
		}
	})

	t.Run("expected_series_names", func(t *testing.T) {
		seriesNames, err := handler.GetInternalSeriesNames("rsi", call)
		if err != nil {
			t.Fatalf("GetInternalSeriesNames() error = %v", err)
		}

		expected := []string{
			"_rsi_gains",
			"_rsi_losses",
			"_rsi_rma_gains",
			"_rsi_rma_losses",
		}

		for i, exp := range expected {
			if i >= len(seriesNames) || seriesNames[i] != exp {
				t.Errorf("Series[%d]: expected %q, got %q", i, exp, seriesNames[i])
			}
		}
	})
}

func TestRSIHandler_ErrorHandling(t *testing.T) {
	handler := &RSIHandler{}
	gen := newTestGenerator()

	t.Run("missing_arguments", func(t *testing.T) {
		call := &ast.CallExpression{
			Arguments: []ast.Expression{},
		}

		_, err := handler.GenerateCode(gen, "rsi", call)
		if err == nil {
			t.Error("Expected error for missing arguments, got nil")
		}
	})

	t.Run("single_argument", func(t *testing.T) {
		call := &ast.CallExpression{
			Arguments: []ast.Expression{
				&ast.Identifier{Name: "close"},
			},
		}

		_, err := handler.GenerateCode(gen, "rsi", call)
		if err == nil {
			t.Error("Expected error for single argument (missing period), got nil")
		}
	})

	t.Run("nil_call", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic for nil call expression, got no panic")
			}
		}()
		_, _ = handler.GenerateCode(gen, "rsi", nil)
	})

	t.Run("metadata_with_nil_call", func(t *testing.T) {
		seriesNames, err := handler.GetInternalSeriesNames("test", nil)
		if err != nil {
			t.Errorf("GetInternalSeriesNames() with nil call should not error, got: %v", err)
		}

		if len(seriesNames) != 4 {
			t.Errorf("Expected 4 series even with nil call, got %d", len(seriesNames))
		}
	})
}

func TestRSIHandler_DynamicPeriodDelegation(t *testing.T) {
	handler := &RSIHandler{}
	gen := newTestGenerator()

	t.Run("delegates_to_dynamic_generator", func(t *testing.T) {
		call := &ast.CallExpression{
			Arguments: []ast.Expression{
				&ast.Identifier{Name: "close"},
				&ast.Identifier{Name: "dynamicLen"},
			},
		}

		code, err := handler.GenerateCode(gen, "myRsi", call)
		if err != nil {
			t.Fatalf("GenerateCode() error = %v", err)
		}

		if code == "" {
			t.Fatal("GenerateCode() returned empty string for dynamic period")
		}

		if !strings.Contains(code, "myRsiSeries.Set(") {
			t.Error("dynamic RSI should write to myRsiSeries")
		}
	})

	t.Run("static_period_uses_standard_path", func(t *testing.T) {
		call := &ast.CallExpression{
			Arguments: []ast.Expression{
				&ast.Identifier{Name: "close"},
				&ast.Literal{Value: 14},
			},
		}

		code, err := handler.GenerateCode(gen, "myRsi", call)
		if err != nil {
			t.Fatalf("GenerateCode() error = %v", err)
		}

		if strings.Contains(code, "period := int(") {
			t.Error("static period RSI should not contain dynamic period conversion")
		}
	})
}
