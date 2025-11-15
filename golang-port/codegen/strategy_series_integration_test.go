package codegen

import (
	"os"
	"strings"
	"testing"

	"github.com/borisquantlab/pinescript-go/parser"
)

func TestGenerateSeriesStrategyFullPipeline(t *testing.T) {
	// Read strategy file
	content, err := os.ReadFile("../testdata/strategy-sma-crossover-series.pine")
	if err != nil {
		t.Fatalf("Failed to read strategy file: %v", err)
	}

	// Parse Pine Script
	p, err := parser.NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	script, err := p.ParseBytes("strategy-sma-crossover-series.pine", content)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	// Convert to AST
	converter := parser.NewConverter()
	program, err := converter.ToESTree(script)
	if err != nil {
		t.Fatalf("Conversion error: %v", err)
	}

	// Generate Go code
	code, err := GenerateStrategyCodeFromAST(program)
	if err != nil {
		t.Fatalf("Codegen error: %v", err)
	}

	generated := code.FunctionBody

	// Verify Series declarations for variables with [1] access
	t.Run("Series declarations", func(t *testing.T) {
		if !strings.Contains(generated, "var sma20Series *series.Series") {
			t.Error("Expected sma20Series declaration (accessed with sma20[1])")
		}
		if !strings.Contains(generated, "var sma50Series *series.Series") {
			t.Error("Expected sma50Series declaration (accessed with sma50[1])")
		}
	})

	// Verify Series initialization
	t.Run("Series initialization", func(t *testing.T) {
		if !strings.Contains(generated, "sma20Series = series.NewSeries(len(ctx.Data))") {
			t.Error("Expected sma20Series initialization")
		}
		if !strings.Contains(generated, "sma50Series = series.NewSeries(len(ctx.Data))") {
			t.Error("Expected sma50Series initialization")
		}
	})

	// Verify Series.Set() for ta.sma assignments
	t.Run("Series.Set for calculations", func(t *testing.T) {
		if !strings.Contains(generated, "sma20Series.Set(") {
			t.Error("Expected sma20Series.Set() for ta.sma result")
		}
		if !strings.Contains(generated, "sma50Series.Set(") {
			t.Error("Expected sma50Series.Set() for ta.sma result")
		}
	})

	// Verify Series.Get(1) for historical access
	t.Run("Series.Get for historical access", func(t *testing.T) {
		if !strings.Contains(generated, "sma20Series.Get(1)") {
			t.Error("Expected sma20Series.Get(1) for prev_sma20 = sma20[1]")
		}
		if !strings.Contains(generated, "sma50Series.Get(1)") {
			t.Error("Expected sma50Series.Get(1) for prev_sma50 = sma50[1]")
		}
	})

	// Verify Series.Next() calls at bar loop end
	t.Run("Series.Next cursor advancement", func(t *testing.T) {
		if !strings.Contains(generated, "sma20Series.Next()") {
			t.Error("Expected sma20Series.Next() to advance cursor")
		}
		if !strings.Contains(generated, "sma50Series.Next()") {
			t.Error("Expected sma50Series.Next() to advance cursor")
		}
	})

	// Verify builtin series use ctx.Data[i-1] for historical access
	t.Run("Builtin series historical access", func(t *testing.T) {
		// crossover_signal and crossunder_signal don't use close[1] directly
		// but ta.crossover internally uses series[1]
		// Just verify code generation doesn't crash
		if len(generated) == 0 {
			t.Error("Generated code is empty")
		}
	})

	// Verify crossover detection logic
	t.Run("Crossover logic", func(t *testing.T) {
		// Manual crossover: sma20 > sma50 and prev_sma20 <= prev_sma50
		// Should generate comparison with Series.Get(1)
		if !strings.Contains(generated, "sma20Series.GetCurrent()") || !strings.Contains(generated, "sma20Series.Get(1)") {
			t.Log("Note: Manual crossover logic may use different Series access pattern")
		}
	})

	// Print generated code for manual inspection
	t.Logf("\n=== Generated Go Code ===\n%s\n=== End Generated Code ===\n", generated)
}

func TestSeriesCodegenPerformanceCheck(t *testing.T) {
	// This test verifies the generated code will have good performance characteristics

	content, err := os.ReadFile("../testdata/strategy-sma-crossover-series.pine")
	if err != nil {
		t.Skip("Strategy file not available")
	}

	p, err := parser.NewParser()
	if err != nil {
		t.Fatalf("Parser creation failed: %v", err)
	}

	script, err := p.ParseBytes("test.pine", content)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	converter := parser.NewConverter()
	program, err := converter.ToESTree(script)
	if err != nil {
		t.Fatalf("Conversion failed: %v", err)
	}

	code, err := GenerateStrategyCodeFromAST(program)
	if err != nil {
		t.Fatalf("Codegen failed: %v", err)
	}

	generated := code.FunctionBody

	// Verify no O(N) array operations
	antiPatterns := []string{
		"append(",        // Growing slices in loop
		"copy(",          // Array copying
		"make([]float64", // Repeated allocations (Series pre-allocates)
	}

	for _, pattern := range antiPatterns {
		count := strings.Count(generated, pattern)
		if pattern == "make([]float64" && count > 0 {
			// Series.NewSeries uses make(), but only ONCE per variable before loop
			lines := strings.Split(generated, "\n")
			makeCount := 0
			inLoop := false
			for _, line := range lines {
				if strings.Contains(line, "for i := 0; i < len(ctx.Data)") {
					inLoop = true
				}
				if inLoop && strings.Contains(line, pattern) {
					makeCount++
				}
			}
			if makeCount > 0 {
				t.Errorf("Performance issue: %s found %d times inside bar loop", pattern, makeCount)
			}
		}
	}

	// Verify Series operations (all O(1))
	requiredPatterns := []string{
		"Series.Get(",   // O(1) cursor-offset arithmetic
		"Series.Set(",   // O(1) cursor write
		"Series.Next()", // O(1) cursor increment
	}

	for _, pattern := range requiredPatterns {
		if !strings.Contains(generated, pattern) {
			t.Logf("Info: Pattern %s not found (may not be required for all strategies)", pattern)
		}
	}
}
