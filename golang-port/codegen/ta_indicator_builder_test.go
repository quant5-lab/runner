package codegen

import (
	"strings"
	"testing"
)

func TestTAIndicatorBuilder_SMA(t *testing.T) {
	mockAccessor := &MockAccessGenerator{
		loopAccessFn: func(loopVar string) string {
			return "closeSeries.Get(" + loopVar + ")"
		},
	}
	
	builder := NewTAIndicatorBuilder("SMA", "sma20", 20, mockAccessor, false)
	builder.WithAccumulator(NewSumAccumulator())
	
	code := builder.Build()
	
	requiredElements := []string{
		"/* Inline SMA(20) */",
		"if ctx.BarIndex < 20-1",
		"sma20Series.Set(math.NaN())",
		"} else {",
		"sum := 0.0",
		"for j := 0; j < 20; j++",
		"closeSeries.Get(j)",
		"sum / 20.0",
		"sma20Series.Set",
	}
	
	for _, elem := range requiredElements {
		if !strings.Contains(code, elem) {
			t.Errorf("SMA builder missing %q\nGenerated code:\n%s", elem, code)
		}
	}
	
	// Verify structure
	if strings.Count(code, "if ctx.BarIndex") != 1 {
		t.Error("Should have exactly one warmup check")
	}
	
	if strings.Count(code, "for j :=") != 1 {
		t.Error("Should have exactly one loop")
	}
}

func TestTAIndicatorBuilder_SMAWithNaN(t *testing.T) {
	mockAccessor := &MockAccessGenerator{
		loopAccessFn: func(loopVar string) string {
			return "seriesVar.Get(" + loopVar + ")"
		},
	}
	
	builder := NewTAIndicatorBuilder("SMA", "smaTest", 10, mockAccessor, true)
	builder.WithAccumulator(NewSumAccumulator())
	
	code := builder.Build()
	
	nanCheckElements := []string{
		"hasNaN := false",
		"if math.IsNaN(val)",
		"hasNaN = true",
		"break",
		"if hasNaN",
		"smaTestSeries.Set(math.NaN())",
	}
	
	for _, elem := range nanCheckElements {
		if !strings.Contains(code, elem) {
			t.Errorf("SMA with NaN check missing %q\nGenerated code:\n%s", elem, code)
		}
	}
}

func TestTAIndicatorBuilder_EMA(t *testing.T) {
	mockAccessor := &MockAccessGenerator{
		loopAccessFn: func(loopVar string) string {
			return "closeSeries.Get(" + loopVar + ")"
		},
	}
	
	builder := NewTAIndicatorBuilder("EMA", "ema20", 20, mockAccessor, false)
	builder.WithAccumulator(NewEMAAccumulator(20))
	
	code := builder.Build()
	
	requiredElements := []string{
		"/* Inline EMA(20) */",
		"alpha := 2.0 / float64(20+1)",
		"for j := 0; j < 20; j++",
		"ema = alpha*closeSeries.Get(j) + (1-alpha)*ema",
		"ema20Series.Set(",
	}
	
	for _, elem := range requiredElements {
		if !strings.Contains(code, elem) {
			t.Errorf("EMA builder missing %q\nGenerated code:\n%s", elem, code)
		}
	}
}

func TestTAIndicatorBuilder_STDEV(t *testing.T) {
	mockAccessor := &MockAccessGenerator{
		loopAccessFn: func(loopVar string) string {
			return "closeSeries.Get(" + loopVar + ")"
		},
	}
	
	// STDEV requires two passes: mean calculation + variance calculation
	builder := NewTAIndicatorBuilder("STDEV", "stdev20", 20, mockAccessor, false)
	
	// First pass: calculate mean
	builder.WithAccumulator(NewSumAccumulator())
	meanCode := builder.Build()
	
	// Second pass: calculate variance (would need mean variable)
	builder2 := NewTAIndicatorBuilder("STDEV", "stdev20", 20, mockAccessor, false)
	builder2.WithAccumulator(NewVarianceAccumulator("mean"))
	varianceCode := builder2.Build()
	
	// Check mean calculation
	if !strings.Contains(meanCode, "sum := 0.0") {
		t.Error("STDEV mean pass missing sum initialization")
	}
	
	// Check variance calculation
	varianceElements := []string{
		"variance := 0.0",
		"closeSeries.Get(j) - mean",  // Actual accessor call
		"diff * diff",
		"variance",
	}
	
	for _, elem := range varianceElements {
		if !strings.Contains(varianceCode, elem) {
			t.Errorf("STDEV variance pass missing %q", elem)
		}
	}
}

func TestTAIndicatorBuilder_EdgeCases(t *testing.T) {
	mockAccessor := &MockAccessGenerator{
		loopAccessFn: func(loopVar string) string {
			return "data.Get(" + loopVar + ")"
		},
	}
	
	t.Run("Period 1", func(t *testing.T) {
		builder := NewTAIndicatorBuilder("SMA", "sma1", 1, mockAccessor, false)
		builder.WithAccumulator(NewSumAccumulator())
		code := builder.Build()
		
		if !strings.Contains(code, "for j := 0; j < 1; j++") {
			t.Error("Period 1 should still have loop")
		}
	})
	
	t.Run("Large Period", func(t *testing.T) {
		builder := NewTAIndicatorBuilder("SMA", "sma200", 200, mockAccessor, false)
		builder.WithAccumulator(NewSumAccumulator())
		code := builder.Build()
		
		if !strings.Contains(code, "if ctx.BarIndex < 200-1") {
			t.Error("Large period should have correct warmup check")
		}
		
		if !strings.Contains(code, "for j := 0; j < 200; j++") {
			t.Error("Large period should have correct loop")
		}
		
		if !strings.Contains(code, "sum / 200.0") {
			t.Error("Large period should have correct finalization")
		}
	})
	
	t.Run("Variable Names with Underscores", func(t *testing.T) {
		builder := NewTAIndicatorBuilder("EMA", "ema_20_close", 20, mockAccessor, false)
		builder.WithAccumulator(NewEMAAccumulator(20))
		code := builder.Build()
		
		if !strings.Contains(code, "ema_20_closeSeries.Set") {
			t.Error("Variable name with underscores should be preserved")
		}
	})
}

func TestTAIndicatorBuilder_BuildStep(t *testing.T) {
	mockAccessor := &MockAccessGenerator{
		loopAccessFn: func(loopVar string) string {
			return "test.Get(" + loopVar + ")"
		},
	}
	
	builder := NewTAIndicatorBuilder("TEST", "test", 10, mockAccessor, false)
	builder.WithAccumulator(NewSumAccumulator())
	
	t.Run("BuildHeader", func(t *testing.T) {
		header := builder.BuildHeader()
		if !strings.Contains(header, "/* Inline TEST(10) */") {
			t.Errorf("Header incorrect: %s", header)
		}
	})
	
	t.Run("BuildWarmupCheck", func(t *testing.T) {
		warmup := builder.BuildWarmupCheck()
		if !strings.Contains(warmup, "if ctx.BarIndex < 10-1") {
			t.Errorf("Warmup check incorrect: %s", warmup)
		}
	})
	
	t.Run("BuildInitialization", func(t *testing.T) {
		init := builder.BuildInitialization()
		if !strings.Contains(init, "sum := 0.0") {
			t.Errorf("Initialization incorrect: %s", init)
		}
	})
	
	t.Run("BuildLoop", func(t *testing.T) {
		loop := builder.BuildLoop(func(val string) string {
			return "sum += " + val
		})
		if !strings.Contains(loop, "for j := 0; j < 10; j++") {
			t.Errorf("Loop structure incorrect: %s", loop)
		}
		if !strings.Contains(loop, "test.Get(j)") {
			t.Errorf("Loop body incorrect: %s", loop)
		}
	})
	
	t.Run("BuildFinalization", func(t *testing.T) {
		final := builder.BuildFinalization("sum / 10.0")
		if !strings.Contains(final, "testSeries.Set(sum / 10.0)") {
			t.Errorf("Finalization incorrect: %s", final)
		}
	})
}

func TestTAIndicatorBuilder_Integration(t *testing.T) {
	// Test that the builder integrates all components correctly
	mockAccessor := &MockAccessGenerator{
		loopAccessFn: func(loopVar string) string {
			return "closeSeries.Get(" + loopVar + ")"
		},
	}
	
	// Build SMA with all components
	builder := NewTAIndicatorBuilder("SMA", "sma20", 20, mockAccessor, true)
	builder.WithAccumulator(NewSumAccumulator())
	
	code := builder.Build()
	
	// Verify complete structure
	tests := []struct {
		name     string
		contains string
		count    int
	}{
		{"Header comment", "/* Inline SMA(20) */", 1},
		{"Warmup check", "if ctx.BarIndex < 20-1", 1},
		{"NaN set in warmup", "sma20Series.Set(math.NaN())", 2}, // warmup + NaN check
		{"Initialization", "sum := 0.0", 1},
		{"NaN flag", "hasNaN := false", 1},
		{"Loop", "for j :=", 1},
		{"Value access", "closeSeries.Get(j)", 1},
		{"NaN check", "if math.IsNaN(val)", 1},
		{"Accumulation", "sum += val", 1},
		{"Final NaN check", "if hasNaN", 1},
		{"Result calculation", "sum / 20.0", 1},
		{"Result set", "sma20Series.Set(", 3}, // warmup NaN + hasNaN NaN + actual result
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count := strings.Count(code, tt.contains)
			if count != tt.count {
				t.Errorf("Expected %d occurrences of %q, got %d\nCode:\n%s", 
					tt.count, tt.contains, count, code)
			}
		})
	}
	
	// Verify indentation is consistent
	lines := strings.Split(code, "\n")
	for i, line := range lines {
		if line == "" {
			continue
		}
		// Check that lines don't have inconsistent indentation
		if strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") {
			t.Errorf("Line %d has space indentation instead of tabs: %q", i+1, line)
		}
	}
}
