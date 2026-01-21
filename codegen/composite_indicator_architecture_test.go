package codegen

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestCompositeIndicatorRegistry_BasicRegistration(t *testing.T) {
	registry := NewCompositeIndicatorRegistry()

	t.Run("register_and_check", func(t *testing.T) {
		handler := &RSIHandler{}
		registry.Register("ta.rsi", handler)

		if !registry.IsCompositeIndicator("ta.rsi") {
			t.Error("Registered indicator not found in registry")
		}
	})

	t.Run("unregistered_indicator", func(t *testing.T) {
		if registry.IsCompositeIndicator("ta.unknown") {
			t.Error("Unregistered indicator should return false")
		}
	})

	t.Run("nil_safety", func(t *testing.T) {
		var nilRegistry *CompositeIndicatorRegistry
		seriesNames := nilRegistry.GetInternalSeriesNames("ta.rsi", "test", &ast.CallExpression{})
		if len(seriesNames) != 0 {
			t.Errorf("Nil registry should return empty slice, got %d entries", len(seriesNames))
		}
	})
}

func TestCompositeIndicatorRegistry_MultiIndicator(t *testing.T) {
	registry := NewCompositeIndicatorRegistry()

	registry.Register("ta.rsi", &RSIHandler{})
	registry.Register("ta.dmi", &MockDMIHandler{})

	t.Run("both_indicators_registered", func(t *testing.T) {
		if !registry.IsCompositeIndicator("ta.rsi") {
			t.Error("RSI not found after multi-indicator registration")
		}
		if !registry.IsCompositeIndicator("ta.dmi") {
			t.Error("DMI not found after multi-indicator registration")
		}
	})

	t.Run("rsi_series_discovery", func(t *testing.T) {
		seriesNames := registry.GetInternalSeriesNames("ta.rsi", "myRSI", &ast.CallExpression{})
		if len(seriesNames) != 4 {
			t.Errorf("RSI should expose 4 internal series, got %d", len(seriesNames))
		}
	})

	t.Run("dmi_series_discovery", func(t *testing.T) {
		seriesNames := registry.GetInternalSeriesNames("ta.dmi", "myDMI", &ast.CallExpression{})
		if len(seriesNames) != 5 {
			t.Errorf("MockDMI should expose 5 internal series, got %d", len(seriesNames))
		}
	})

	t.Run("independent_indicator_series", func(t *testing.T) {
		rsiSeries := registry.GetInternalSeriesNames("ta.rsi", "test", &ast.CallExpression{})
		dmiSeries := registry.GetInternalSeriesNames("ta.dmi", "test", &ast.CallExpression{})

		for _, rsiName := range rsiSeries {
			for _, dmiName := range dmiSeries {
				if rsiName == dmiName {
					t.Errorf("RSI and DMI series overlap: %s", rsiName)
				}
			}
		}
	})
}

func TestCompositeIndicatorRegistry_GeneratorIntegration(t *testing.T) {
	gen := newTestGenerator()

	t.Run("test_generator_has_registry", func(t *testing.T) {
		if gen.compositeIndicatorRegistry == nil {
			t.Fatal("Test generator missing composite indicator registry")
		}
	})

	t.Run("test_generator_has_rsi_registered", func(t *testing.T) {
		if !gen.compositeIndicatorRegistry.IsCompositeIndicator("ta.rsi") {
			t.Error("Test generator should have RSI pre-registered")
		}
	})

	t.Run("generator_discovers_rsi_series", func(t *testing.T) {
		seriesNames := gen.compositeIndicatorRegistry.GetInternalSeriesNames(
			"ta.rsi", "testRSI", &ast.CallExpression{})

		if len(seriesNames) == 0 {
			t.Error("Generator should discover RSI internal series via registry")
		}

		expectedCount := 4
		if len(seriesNames) != expectedCount {
			t.Errorf("Generator discovered %d series, expected %d", len(seriesNames), expectedCount)
		}
	})
}
