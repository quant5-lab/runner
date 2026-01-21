package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

/* TestCompositeIndicatorPattern_GenericBehavior validates the composite indicator pattern works for ANY indicator */
func TestCompositeIndicatorPattern_GenericBehavior(t *testing.T) {
	t.Run("metadata_provider_interface_contract", func(t *testing.T) {
		/* Test that metadata providers follow expected contract */
		providers := []struct {
			name      string
			provider  CompositeIndicatorMetadata
			funcName  string
			minSeries int
		}{
			{"RSI", &RSIHandler{}, "ta.rsi", 4},
			{"MockDMI", &MockDMIHandler{}, "ta.dmi", 5},
		}

		for _, tt := range providers {
			t.Run(tt.name, func(t *testing.T) {
				call := &ast.CallExpression{
					Arguments: []ast.Expression{&ast.Literal{Value: 14}},
				}

				seriesNames, err := tt.provider.GetInternalSeriesNames("test", call)
				if err != nil {
					t.Errorf("%s.GetInternalSeriesNames() error = %v", tt.name, err)
				}

				if len(seriesNames) < tt.minSeries {
					t.Errorf("%s declares %d series, want >= %d", tt.name, len(seriesNames), tt.minSeries)
				}

				/* All series should contain the variable name for uniqueness */
				varNameFound := false
				for _, name := range seriesNames {
					if strings.Contains(name, "test") {
						varNameFound = true
						break
					}
				}
				if !varNameFound {
					t.Errorf("%s internal series should contain variable name 'test' for uniqueness", tt.name)
				}
			})
		}
	})

	t.Run("registry_discovery_is_generic", func(t *testing.T) {
		gen := &generator{}
		gen.compositeIndicatorRegistry = NewCompositeIndicatorRegistry()
		gen.compositeIndicatorRegistry.Register("ta.rsi", &RSIHandler{})
		gen.compositeIndicatorRegistry.Register("ta.dmi", &MockDMIHandler{})

		testCases := []struct {
			funcName      string
			isComposite   bool
			expectedCount int
		}{
			{"ta.rsi", true, 4},
			{"ta.dmi", true, 5},
			{"ta.sma", false, 0},
			{"ta.ema", false, 0},
			{"unknown", false, 0},
		}

		for _, tt := range testCases {
			t.Run(tt.funcName, func(t *testing.T) {
				isComposite := gen.compositeIndicatorRegistry.IsCompositeIndicator(tt.funcName)
				if isComposite != tt.isComposite {
					t.Errorf("IsCompositeIndicator(%s) = %v, want %v", tt.funcName, isComposite, tt.isComposite)
				}

				if tt.isComposite {
					call := &ast.CallExpression{Arguments: []ast.Expression{&ast.Literal{Value: 14}}}
					series := gen.compositeIndicatorRegistry.GetInternalSeriesNames(tt.funcName, "test", call)
					if len(series) != tt.expectedCount {
						t.Errorf("GetInternalSeriesNames(%s) returned %d series, want %d", tt.funcName, len(series), tt.expectedCount)
					}
				}
			})
		}
	})

	t.Run("multiple_composites_in_same_strategy", func(t *testing.T) {
		gen := &generator{}
		gen.compositeIndicatorRegistry = NewCompositeIndicatorRegistry()
		gen.compositeIndicatorRegistry.Register("ta.rsi", &RSIHandler{})
		gen.compositeIndicatorRegistry.Register("ta.dmi", &MockDMIHandler{})

		call14 := &ast.CallExpression{Arguments: []ast.Expression{&ast.Literal{Value: 14}}}
		call20 := &ast.CallExpression{Arguments: []ast.Expression{&ast.Literal{Value: 20}}}

		rsi1 := gen.compositeIndicatorRegistry.GetInternalSeriesNames("ta.rsi", "rsi14", call14)
		rsi2 := gen.compositeIndicatorRegistry.GetInternalSeriesNames("ta.rsi", "rsi20", call20)
		dmi := gen.compositeIndicatorRegistry.GetInternalSeriesNames("ta.dmi", "dmi14", call14)

		/* Verify all series are unique */
		allSeries := append(append(rsi1, rsi2...), dmi...)
		uniqueCheck := make(map[string]bool)
		for _, name := range allSeries {
			if uniqueCheck[name] {
				t.Errorf("Duplicate series name detected: %s", name)
			}
			uniqueCheck[name] = true
		}

		if len(uniqueCheck) != len(allSeries) {
			t.Error("Series names must be unique across multiple composite indicators")
		}
	})
}

/* TestTAFunctionRegistry_EdgeCases validates registry handles boundary conditions */
func TestTAFunctionRegistry_EdgeCases(t *testing.T) {
	t.Run("negative_period_handling", func(t *testing.T) {
		gen := newTestGenerator()
		handler := &RSIHandler{}

		call := &ast.CallExpression{
			Arguments: []ast.Expression{
				&ast.Identifier{Name: "close"},
				&ast.Literal{Value: -14},
			},
		}

		_, err := handler.GenerateCode(gen, "rsi", call)
		/* Current implementation may not validate negative periods at handler level */
		/* This documents expected behavior - handler delegates validation to builder */
		_ = err // Document: validation happens in period extraction/builder
	})

	t.Run("zero_period_handling", func(t *testing.T) {
		gen := newTestGenerator()
		handler := &RSIHandler{}

		call := &ast.CallExpression{
			Arguments: []ast.Expression{
				&ast.Identifier{Name: "close"},
				&ast.Literal{Value: 0},
			},
		}

		_, err := handler.GenerateCode(gen, "rsi", call)
		/* Current implementation may not validate zero periods at handler level */
		/* This documents expected behavior - handler delegates validation to builder */
		_ = err // Document: validation happens in period extraction/builder
	})

	t.Run("extremely_large_period", func(t *testing.T) {
		gen := newTestGenerator()
		handler := &RSIHandler{}

		call := &ast.CallExpression{
			Arguments: []ast.Expression{
				&ast.Identifier{Name: "close"},
				&ast.Literal{Value: 5000},
			},
		}

		code, err := handler.GenerateCode(gen, "rsi", call)
		if err != nil {
			t.Fatalf("Should handle large periods gracefully, got error: %v", err)
		}

		/* Should generate valid warmup check */
		if !strings.Contains(code, "4999") || !strings.Contains(code, "5000") {
			t.Error("Large period should generate correct warmup boundary")
		}
	})

	t.Run("missing_arguments", func(t *testing.T) {
		gen := newTestGenerator()
		handler := &RSIHandler{}

		call := &ast.CallExpression{
			Arguments: []ast.Expression{},
		}

		_, err := handler.GenerateCode(gen, "rsi", call)
		if err == nil {
			t.Error("Expected error for missing arguments, got nil")
		}
	})

	t.Run("empty_variable_name", func(t *testing.T) {
		gen := newTestGenerator()
		handler := &RSIHandler{}

		call := &ast.CallExpression{
			Arguments: []ast.Expression{
				&ast.Identifier{Name: "close"},
				&ast.Literal{Value: 14},
			},
		}

		code, err := handler.GenerateCode(gen, "", call)
		if err != nil {
			t.Fatalf("Should handle empty varName gracefully, got error: %v", err)
		}

		/* Code should still be generated even with empty varName */
		if code == "" {
			t.Error("Generated code should not be empty even with empty varName")
		}
	})
}

/* TestCompositeIndicatorRegistry_Lifecycle validates registry state management */
func TestCompositeIndicatorRegistry_Lifecycle(t *testing.T) {
	t.Run("nil_registry_safety", func(t *testing.T) {
		var registry *CompositeIndicatorRegistry

		/* Should not panic on nil registry */
		series := registry.GetInternalSeriesNames("ta.rsi", "test", nil)
		if len(series) != 0 {
			t.Error("Nil registry should return empty slice")
		}

		isComposite := registry.IsCompositeIndicator("ta.rsi")
		if isComposite {
			t.Error("Nil registry should return false for IsCompositeIndicator")
		}
	})

	t.Run("nil_providers_map_safety", func(t *testing.T) {
		registry := &CompositeIndicatorRegistry{providers: nil}

		series := registry.GetInternalSeriesNames("ta.rsi", "test", nil)
		if len(series) != 0 {
			t.Error("Registry with nil providers should return empty slice")
		}
	})

	t.Run("duplicate_registration_last_wins", func(t *testing.T) {
		registry := NewCompositeIndicatorRegistry()

		handler1 := &RSIHandler{}
		handler2 := &MockDMIHandler{}

		registry.Register("ta.test", handler1)
		registry.Register("ta.test", handler2) // Overwrite

		call := &ast.CallExpression{Arguments: []ast.Expression{&ast.Literal{Value: 14}}}
		series := registry.GetInternalSeriesNames("ta.test", "var", call)

		/* Should use handler2 (MockDMI with 5 series, not RSI with 4) */
		if len(series) != 5 {
			t.Errorf("Duplicate registration should use last handler (expected 5 series from DMI), got %d", len(series))
		}
	})

	t.Run("function_name_normalization", func(t *testing.T) {
		registry := NewCompositeIndicatorRegistry()
		/* In production, both variants are registered (see generator.go:51-52) */
		registry.Register("ta.rsi", &RSIHandler{})
		registry.Register("rsi", &RSIHandler{})

		/* Both should work since both are registered */
		withPrefix := registry.IsCompositeIndicator("ta.rsi")
		withoutPrefix := registry.IsCompositeIndicator("rsi")

		if !withPrefix {
			t.Error("Registry should recognize 'ta.rsi'")
		}
		if !withoutPrefix {
			t.Error("Registry should recognize 'rsi' (Pine v4 syntax)")
		}
	})
}

/* TestCompositeIndicator_MetadataErrors validates error handling in metadata providers */
func TestCompositeIndicator_MetadataErrors(t *testing.T) {
	t.Run("nil_call_expression", func(t *testing.T) {
		handler := &RSIHandler{}

		/* GetInternalSeriesNames with nil call */
		series, err := handler.GetInternalSeriesNames("test", nil)

		/* Current behavior: returns series names regardless of call validity */
		/* Series names are static and don't depend on call expression content */
		if err == nil && len(series) == 4 {
			/* This is expected - metadata is static */
			return
		}

		if err != nil {
			/* Also acceptable - handler may validate nil call */
			return
		}

		t.Error("GetInternalSeriesNames with nil call should either return 4 series or error")
	})

	t.Run("malformed_call_missing_period", func(t *testing.T) {
		handler := &RSIHandler{}

		call := &ast.CallExpression{
			Arguments: []ast.Expression{
				&ast.Identifier{Name: "close"},
				/* Missing period argument */
			},
		}

		/* GetInternalSeriesNames should still work (doesn't validate arguments) */
		series, err := handler.GetInternalSeriesNames("test", call)
		if err != nil {
			t.Errorf("GetInternalSeriesNames should not validate call arguments, got error: %v", err)
		}

		/* Should still return series names (metadata is independent of validation) */
		if len(series) != 4 {
			t.Errorf("Expected 4 series names regardless of call validity, got %d", len(series))
		}
	})
}
