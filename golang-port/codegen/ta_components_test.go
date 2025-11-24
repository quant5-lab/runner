package codegen

import (
	"strings"
	"testing"
)

func TestWarmupChecker(t *testing.T) {
	tests := []struct {
		name           string
		period         int
		varName        string
		expectedInCode []string
	}{
		{
			name:    "Period 20",
			period:  20,
			varName: "sma20",
			expectedInCode: []string{
				"if ctx.BarIndex < 20-1",
				"sma20Series.Set(math.NaN())",
				"} else {",
			},
		},
		{
			name:    "Period 5",
			period:  5,
			varName: "ema5",
			expectedInCode: []string{
				"if ctx.BarIndex < 5-1",
				"ema5Series.Set(math.NaN())",
			},
		},
		{
			name:    "Period 1",
			period:  1,
			varName: "test",
			expectedInCode: []string{
				"if ctx.BarIndex < 1-1",
				"testSeries.Set(math.NaN())",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := NewWarmupChecker(tt.period)

			if checker.MinimumBarsRequired() != tt.period {
				t.Errorf("MinimumBarsRequired() = %d, want %d",
					checker.MinimumBarsRequired(), tt.period)
			}

			indenter := NewCodeIndenter()
			code := checker.GenerateCheck(tt.varName, &indenter)

			for _, expected := range tt.expectedInCode {
				if !strings.Contains(code, expected) {
					t.Errorf("Generated code missing %q\nGot:\n%s", expected, code)
				}
			}
		})
	}
}

func TestSumAccumulator(t *testing.T) {
	acc := NewSumAccumulator()

	t.Run("Initialize", func(t *testing.T) {
		init := acc.Initialize()
		if !strings.Contains(init, "sum := 0.0") {
			t.Errorf("Initialize() missing sum initialization, got: %s", init)
		}
		if !strings.Contains(init, "hasNaN := false") {
			t.Errorf("Initialize() missing hasNaN initialization, got: %s", init)
		}
	})

	t.Run("Accumulate", func(t *testing.T) {
		result := acc.Accumulate("value")
		expected := "sum += value"
		if result != expected {
			t.Errorf("Accumulate() = %q, want %q", result, expected)
		}
	})

	t.Run("Finalize", func(t *testing.T) {
		result := acc.Finalize(20)
		expected := "sum / 20.0"
		if result != expected {
			t.Errorf("Finalize(20) = %q, want %q", result, expected)
		}
	})

	t.Run("NeedsNaNGuard", func(t *testing.T) {
		if !acc.NeedsNaNGuard() {
			t.Error("NeedsNaNGuard() = false, want true")
		}
	})
}

func TestVarianceAccumulator(t *testing.T) {
	acc := NewVarianceAccumulator("mean")

	t.Run("Initialize", func(t *testing.T) {
		init := acc.Initialize()
		expected := "variance := 0.0"
		if init != expected {
			t.Errorf("Initialize() = %q, want %q", init, expected)
		}
	})

	t.Run("Accumulate", func(t *testing.T) {
		result := acc.Accumulate("val")
		if !strings.Contains(result, "diff := val - mean") {
			t.Errorf("Accumulate() missing diff calculation, got: %s", result)
		}
		if !strings.Contains(result, "variance += diff * diff") {
			t.Errorf("Accumulate() missing variance calculation, got: %s", result)
		}
	})

	t.Run("Finalize", func(t *testing.T) {
		result := acc.Finalize(20)
		expected := "variance /= 20.0"
		if result != expected {
			t.Errorf("Finalize(20) = %q, want %q", result, expected)
		}
	})

	t.Run("NeedsNaNGuard", func(t *testing.T) {
		if acc.NeedsNaNGuard() {
			t.Error("NeedsNaNGuard() = true, want false")
		}
	})
}

func TestEMAAccumulator(t *testing.T) {
	acc := NewEMAAccumulator(20)

	t.Run("Initialize", func(t *testing.T) {
		init := acc.Initialize()
		if !strings.Contains(init, "alpha := 2.0 / float64(20+1)") {
			t.Errorf("Initialize() missing alpha calculation, got: %s", init)
		}
	})

	t.Run("Accumulate", func(t *testing.T) {
		result := acc.Accumulate("val")
		if !strings.Contains(result, "ema = alpha*val + (1-alpha)*ema") {
			t.Errorf("Accumulate() wrong formula, got: %s", result)
		}
	})

	t.Run("GetResultVariable", func(t *testing.T) {
		result := acc.GetResultVariable()
		expected := "ema"
		if result != expected {
			t.Errorf("GetResultVariable() = %q, want %q", result, expected)
		}
	})

	t.Run("NeedsNaNGuard", func(t *testing.T) {
		if !acc.NeedsNaNGuard() {
			t.Error("NeedsNaNGuard() = false, want true")
		}
	})
}

func TestLoopGenerator(t *testing.T) {
	mockAccessor := &MockAccessGenerator{
		loopAccessFn: func(loopVar string) string {
			return "testSeries.Get(" + loopVar + ")"
		},
	}

	t.Run("ForwardLoop", func(t *testing.T) {
		gen := NewLoopGenerator(20, mockAccessor, true)
		indenter := NewCodeIndenter()
		code := gen.GenerateForwardLoop(&indenter)

		expected := "for j := 0; j < 20; j++ {"
		if !strings.Contains(code, expected) {
			t.Errorf("GenerateForwardLoop() missing %q, got: %s", expected, code)
		}
	})

	t.Run("BackwardLoop", func(t *testing.T) {
		gen := NewLoopGenerator(20, mockAccessor, true)
		indenter := NewCodeIndenter()
		code := gen.GenerateBackwardLoop(&indenter)

		expected := "for j := 20-2; j >= 0; j-- {"
		if !strings.Contains(code, expected) {
			t.Errorf("GenerateBackwardLoop() missing %q, got: %s", expected, code)
		}
	})

	t.Run("GenerateValueAccess", func(t *testing.T) {
		gen := NewLoopGenerator(10, mockAccessor, true)
		access := gen.GenerateValueAccess()

		expected := "testSeries.Get(j)"
		if access != expected {
			t.Errorf("GenerateValueAccess() = %q, want %q", access, expected)
		}
	})

	t.Run("RequiresNaNCheck", func(t *testing.T) {
		gen := NewLoopGenerator(10, mockAccessor, true)
		if !gen.RequiresNaNCheck() {
			t.Error("RequiresNaNCheck() = false, want true")
		}

		genNoNaN := NewLoopGenerator(10, mockAccessor, false)
		if genNoNaN.RequiresNaNCheck() {
			t.Error("RequiresNaNCheck() = true, want false")
		}
	})
}

func TestCodeIndenter(t *testing.T) {
	t.Run("Line with no indentation", func(t *testing.T) {
		indenter := NewCodeIndenter()
		line := indenter.Line("test")

		if line != "test\n" {
			t.Errorf("Line() = %q, want %q", line, "test\n")
		}
	})

	t.Run("Line with indentation", func(t *testing.T) {
		indenter := NewCodeIndenter()
		indenter.IncreaseIndent()
		line := indenter.Line("test")

		if line != "\ttest\n" {
			t.Errorf("Line() = %q, want %q", line, "\ttest\n")
		}
	})

	t.Run("Nested indentation", func(t *testing.T) {
		indenter := NewCodeIndenter()
		indenter.IncreaseIndent()
		indenter.IncreaseIndent()
		line := indenter.Line("test")

		if line != "\t\ttest\n" {
			t.Errorf("Line() = %q, want %q", line, "\t\ttest\n")
		}
	})

	t.Run("Decrease indentation", func(t *testing.T) {
		indenter := NewCodeIndenter()
		indenter.IncreaseIndent()
		indenter.IncreaseIndent()
		indenter.DecreaseIndent()
		line := indenter.Line("test")

		if line != "\ttest\n" {
			t.Errorf("Line() = %q, want %q", line, "\ttest\n")
		}
	})

	t.Run("Decrease below zero", func(t *testing.T) {
		indenter := NewCodeIndenter()
		indenter.DecreaseIndent()
		indenter.DecreaseIndent()
		line := indenter.Line("test")

		if line != "test\n" {
			t.Errorf("Line() = %q, want %q", line, "test\n")
		}
	})

	t.Run("CurrentLevel", func(t *testing.T) {
		indenter := NewCodeIndenter()
		if indenter.CurrentLevel() != 0 {
			t.Errorf("CurrentLevel() = %d, want 0", indenter.CurrentLevel())
		}

		indenter.IncreaseIndent()
		if indenter.CurrentLevel() != 1 {
			t.Errorf("CurrentLevel() = %d, want 1", indenter.CurrentLevel())
		}
	})
}

// MockAccessGenerator for testing
type MockAccessGenerator struct {
	loopAccessFn func(loopVar string) string
}

func (m *MockAccessGenerator) GenerateLoopValueAccess(loopVar string) string {
	if m.loopAccessFn != nil {
		return m.loopAccessFn(loopVar)
	}
	return "mockAccess"
}

func (m *MockAccessGenerator) GenerateInitialValueAccess(period int) string {
	return "mockInitialAccess"
}
