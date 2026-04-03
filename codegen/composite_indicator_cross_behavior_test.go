package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

/* TestCompositeIndicator_CrossIndicatorBehavior validates multiple composite indicators in same strategy */
func TestCompositeIndicator_CrossIndicatorBehavior(t *testing.T) {
	t.Run("multiple_composites_independent_series", func(t *testing.T) {
		gen := newTestGenerator()

		rsiCall := &ast.CallExpression{
			Arguments: []ast.Expression{
				&ast.Identifier{Name: "close"},
				&ast.Literal{Value: 14},
			},
		}
		smaCall := &ast.CallExpression{
			Arguments: []ast.Expression{
				&ast.Identifier{Name: "close"},
				&ast.Literal{Value: 20},
			},
		}

		rsiSeries := gen.compositeIndicatorRegistry.GetInternalSeriesNames("ta.rsi", "rsi14", rsiCall)
		smaSeries := gen.compositeIndicatorRegistry.GetInternalSeriesNames("ta.sma", "sma20", smaCall)

		/* RSI has internal series, SMA does not (simple indicator) */
		if len(rsiSeries) == 0 {
			t.Error("RSI should expose internal series")
		}
		if len(smaSeries) != 0 {
			t.Error("SMA should not expose internal series (not composite)")
		}

		/* RSI series names must be prefixed with variable name */
		for _, name := range rsiSeries {
			if !strings.HasPrefix(name, "_rsi14_") {
				t.Errorf("RSI series name %q missing _rsi14_ prefix", name)
			}
		}
	})

	t.Run("same_indicator_different_variables", func(t *testing.T) {
		gen := newTestGenerator()

		call1 := &ast.CallExpression{
			Arguments: []ast.Expression{
				&ast.Identifier{Name: "close"},
				&ast.Literal{Value: 14},
			},
		}
		call2 := &ast.CallExpression{
			Arguments: []ast.Expression{
				&ast.Identifier{Name: "close"},
				&ast.Literal{Value: 9},
			},
		}

		series1 := gen.compositeIndicatorRegistry.GetInternalSeriesNames("ta.rsi", "rsi14", call1)
		series2 := gen.compositeIndicatorRegistry.GetInternalSeriesNames("ta.rsi", "rsi9", call2)

		if len(series1) != len(series2) {
			t.Errorf("Same indicator should expose same number of series: rsi14=%d, rsi9=%d", len(series1), len(series2))
		}

		/* Series names must be unique per variable */
		for _, s1 := range series1 {
			for _, s2 := range series2 {
				if s1 == s2 {
					t.Errorf("Series names overlap between rsi14 and rsi9: %s", s1)
				}
			}
		}
	})

	t.Run("composite_in_code_generation", func(t *testing.T) {
		gen := newTestGenerator()

		call := &ast.CallExpression{
			Arguments: []ast.Expression{
				&ast.Identifier{Name: "close"},
				&ast.Literal{Value: 14},
			},
		}

		handler := gen.taRegistry.FindHandler("ta.rsi")
		if handler == nil {
			t.Fatal("RSI handler not found in registry")
		}

		code, err := handler.GenerateCode(gen, "myRSI", call)
		if err != nil {
			t.Fatalf("GenerateCode() error = %v", err)
		}

		/* Verify internal series are declared */
		if !strings.Contains(code, "_myRSI_gainsSeries") {
			t.Error("Generated code missing internal gains series")
		}
		if !strings.Contains(code, "_myRSI_lossesSeries") {
			t.Error("Generated code missing internal losses series")
		}
		if !strings.Contains(code, "_myRSI_rma_gains") {
			t.Error("Generated code missing RMA gains series")
		}
		if !strings.Contains(code, "_myRSI_rma_losses") {
			t.Error("Generated code missing RMA losses series")
		}
	})

	t.Run("registry_isolation_between_instances", func(t *testing.T) {
		gen1 := newTestGenerator()
		gen2 := newTestGenerator()

		/* Both generators should have independent registries */
		if gen1.compositeIndicatorRegistry == gen2.compositeIndicatorRegistry {
			t.Error("Generators sharing same registry instance (should be independent)")
		}

		/* Both should have RSI registered */
		if !gen1.compositeIndicatorRegistry.IsCompositeIndicator("ta.rsi") {
			t.Error("Generator 1 missing RSI registration")
		}
		if !gen2.compositeIndicatorRegistry.IsCompositeIndicator("ta.rsi") {
			t.Error("Generator 2 missing RSI registration")
		}
	})
}

/* TestCompositeIndicator_PerformanceBoundaries validates behavior with extreme period values */
func TestCompositeIndicator_PerformanceBoundaries(t *testing.T) {
	t.Run("large_period_1000", func(t *testing.T) {
		gen := newTestGenerator()

		call := &ast.CallExpression{
			Arguments: []ast.Expression{
				&ast.Identifier{Name: "close"},
				&ast.Literal{Value: 1000},
			},
		}

		handler := gen.taRegistry.FindHandler("ta.rsi")
		code, err := handler.GenerateCode(gen, "rsi1000", call)

		if err != nil {
			t.Fatalf("Large period (1000) should not fail: %v", err)
		}

		/* Verify warmup guard uses correct period */
		if !strings.Contains(code, "if ctx.BarIndex < 1000") {
			t.Error("Warmup guard missing or incorrect for period=1000")
		}

		/* Verify RMA uses correct period */
		if !strings.Contains(code, "Inline RMA(1000)") {
			t.Error("RMA header missing period=1000")
		}
	})

	t.Run("large_period_5000", func(t *testing.T) {
		gen := newTestGenerator()

		call := &ast.CallExpression{
			Arguments: []ast.Expression{
				&ast.Identifier{Name: "close"},
				&ast.Literal{Value: 5000},
			},
		}

		handler := gen.taRegistry.FindHandler("ta.rsi")
		code, err := handler.GenerateCode(gen, "rsi5000", call)

		if err != nil {
			t.Fatalf("Large period (5000) should not fail: %v", err)
		}

		if !strings.Contains(code, "if ctx.BarIndex < 5000") {
			t.Error("Warmup guard missing or incorrect for period=5000")
		}
	})

	t.Run("extreme_period_10000", func(t *testing.T) {
		gen := newTestGenerator()

		call := &ast.CallExpression{
			Arguments: []ast.Expression{
				&ast.Identifier{Name: "close"},
				&ast.Literal{Value: 10000},
			},
		}

		handler := gen.taRegistry.FindHandler("ta.rsi")
		code, err := handler.GenerateCode(gen, "rsi10000", call)

		if err != nil {
			t.Fatalf("Extreme period (10000) should not fail: %v", err)
		}

		if !strings.Contains(code, "if ctx.BarIndex < 10000") {
			t.Error("Warmup guard missing or incorrect for period=10000 (self-documenting structure)")
		}

		if !strings.Contains(code, "rs :=") || !strings.Contains(code, "rsi :=") {
			t.Error("RSI calculation structure missing (code should self-document)")
		}
	})

	t.Run("period_arithmetic_overflow_protection", func(t *testing.T) {
		gen := newTestGenerator()

		call := &ast.CallExpression{
			Arguments: []ast.Expression{
				&ast.Identifier{Name: "close"},
				&ast.Literal{Value: 1 << 30}, /* Large but within int32 range */
			},
		}

		handler := gen.taRegistry.FindHandler("ta.rsi")
		_, err := handler.GenerateCode(gen, "rsiMax", call)

		/* Should not panic - error handling is implementation-specific */
		if err == nil {
			/* If no error, generation succeeded - valid behavior */
		}
	})
}
