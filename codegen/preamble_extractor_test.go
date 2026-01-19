package codegen

import (
	"testing"
)

func TestPreambleExtractor_ExtractFromAccessor(t *testing.T) {
	tests := []struct {
		name             string
		accessor         AccessGenerator
		expectedPreamble string
		description      string
	}{
		{
			name: "accessor implementing PreambleProvider",
			accessor: &FixnanCallExpressionAccessor{
				tempVarName: "expr_temp",
				tempVarCode: "expr_temp := 100 * rma(...) / truerange\n",
			},
			expectedPreamble: "expr_temp := 100 * rma(...) / truerange\n",
			description:      "Standard preamble extraction from FixnanCallExpressionAccessor",
		},
		{
			name: "accessor with empty preamble",
			accessor: &FixnanCallExpressionAccessor{
				tempVarName: "simple",
				tempVarCode: "",
			},
			expectedPreamble: "",
			description:      "Empty preamble should return empty string",
		},
		{
			name: "accessor without PreambleProvider interface",
			accessor: &mockAccessorWithoutPreamble{
				value: "closeSeries.Get(0)",
			},
			expectedPreamble: "",
			description:      "Non-preamble accessor returns empty string",
		},
		{
			name: "accessor with multiline preamble",
			accessor: &FixnanCallExpressionAccessor{
				tempVarName: "complex_temp",
				tempVarCode: "temp1 := a + b\ntemp2 := temp1 * c\ncomplex_temp := temp2 / d\n",
			},
			expectedPreamble: "temp1 := a + b\ntemp2 := temp1 * c\ncomplex_temp := temp2 / d\n",
			description:      "Multiline preamble extraction",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			extractor := NewPreambleExtractor()
			preamble := extractor.ExtractFromAccessor(tt.accessor)

			if preamble != tt.expectedPreamble {
				t.Errorf("[%s]\nExpected preamble:\n%s\nGot:\n%s",
					tt.description, tt.expectedPreamble, preamble)
			}
		})
	}
}

func TestPreambleExtractor_TypeSafety(t *testing.T) {
	extractor := NewPreambleExtractor()

	t.Run("nil accessor", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("ExtractFromAccessor panicked on nil accessor: %v", r)
			}
		}()

		preamble := extractor.ExtractFromAccessor(nil)
		if preamble != "" {
			t.Errorf("Expected empty preamble for nil accessor, got %q", preamble)
		}
	})

	t.Run("accessor with PreambleProvider", func(t *testing.T) {
		accessor := &FixnanCallExpressionAccessor{
			tempVarName: "test",
			tempVarCode: "test := value\n",
		}

		preamble := extractor.ExtractFromAccessor(accessor)
		if preamble != "test := value\n" {
			t.Errorf("Failed to extract preamble from valid PreambleProvider")
		}
	})

	t.Run("accessor without PreambleProvider", func(t *testing.T) {
		accessor := &mockAccessorWithoutPreamble{value: "data"}

		preamble := extractor.ExtractFromAccessor(accessor)
		if preamble != "" {
			t.Errorf("Expected empty preamble for non-PreambleProvider, got %q", preamble)
		}
	})
}

func TestPreambleExtractor_EdgeCases(t *testing.T) {
	extractor := NewPreambleExtractor()

	t.Run("very long preamble", func(t *testing.T) {
		longCode := ""
		for i := 0; i < 100; i++ {
			longCode += "temp := expr\n"
		}

		accessor := &FixnanCallExpressionAccessor{
			tempVarName: "final",
			tempVarCode: longCode,
		}

		preamble := extractor.ExtractFromAccessor(accessor)
		if preamble != longCode {
			t.Error("Large preamble not preserved")
		}
	})

	t.Run("preamble with special characters", func(t *testing.T) {
		specialCode := "temp := \"\\n\\t\\r\\\"\"\n"
		accessor := &FixnanCallExpressionAccessor{
			tempVarName: "special",
			tempVarCode: specialCode,
		}

		preamble := extractor.ExtractFromAccessor(accessor)
		if preamble != specialCode {
			t.Error("Special characters not preserved in preamble")
		}
	})

	t.Run("preamble with unicode", func(t *testing.T) {
		unicodeCode := "temp := \"日本語 中文 한글\"\n"
		accessor := &FixnanCallExpressionAccessor{
			tempVarName: "unicode",
			tempVarCode: unicodeCode,
		}

		preamble := extractor.ExtractFromAccessor(accessor)
		if preamble != unicodeCode {
			t.Error("Unicode not preserved in preamble")
		}
	})
}

func TestPreambleExtractor_RealWorldScenarios(t *testing.T) {
	extractor := NewPreambleExtractor()

	scenarios := []struct {
		name        string
		accessor    *FixnanCallExpressionAccessor
		pattern     string
		description string
	}{
		{
			name: "fixnan with nested TA call",
			accessor: &FixnanCallExpressionAccessor{
				tempVarName: "fixnan_source_temp",
				tempVarCode: "fixnan_source_temp := (100 * rma(upSeries.Get(j), 20) / trueSeries.Get(j))\n",
			},
			pattern:     "rma(upSeries.Get(j), 20)",
			description: "Nested TA function in arithmetic expression",
		},
		{
			name: "conditional expression preamble",
			accessor: &FixnanCallExpressionAccessor{
				tempVarName: "ternary_source_temp",
				tempVarCode: "ternary_source_temp := func() float64 { if (cond) { return a } else { return b } }()\n",
			},
			pattern:     "func() float64",
			description: "IIFE from conditional expression",
		},
		{
			name: "binary expression preamble",
			accessor: &FixnanCallExpressionAccessor{
				tempVarName: "binary_source_temp",
				tempVarCode: "binary_source_temp := (math.Abs(plusSeries.Get(0) - minusSeries.Get(0)))\n",
			},
			pattern:     "math.Abs",
			description: "Binary expression with math function",
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			preamble := extractor.ExtractFromAccessor(scenario.accessor)

			if preamble != scenario.accessor.tempVarCode {
				t.Errorf("[%s] Preamble mismatch\nExpected: %s\nGot: %s",
					scenario.description, scenario.accessor.tempVarCode, preamble)
			}

			if !containsPattern(preamble, scenario.pattern) {
				t.Errorf("[%s] Expected pattern %q not found in preamble: %s",
					scenario.description, scenario.pattern, preamble)
			}
		})
	}
}

func TestPreambleProvider_Interface(t *testing.T) {
	t.Run("FixnanCallExpressionAccessor implements PreambleProvider", func(t *testing.T) {
		var _ PreambleProvider = (*FixnanCallExpressionAccessor)(nil)
	})

	t.Run("GetPreamble returns tempVarCode", func(t *testing.T) {
		expectedCode := "test := value\n"
		accessor := &FixnanCallExpressionAccessor{
			tempVarName: "test",
			tempVarCode: expectedCode,
		}

		var provider PreambleProvider = accessor
		preamble := provider.GetPreamble()

		if preamble != expectedCode {
			t.Errorf("GetPreamble() returned %q, expected %q", preamble, expectedCode)
		}
	})
}

type mockAccessorWithoutPreamble struct {
	value string
}

func (m *mockAccessorWithoutPreamble) GenerateLoopValueAccess(loopVar string) string {
	return m.value
}

func (m *mockAccessorWithoutPreamble) GenerateInitialValueAccess(period int) string {
	return m.value
}

func (m *mockAccessorWithoutPreamble) GenerateCurrentValueAccess() string {
	return m.value
}

func (m *mockAccessorWithoutPreamble) GetBaseOffset() int {
	return 0
}

func containsPattern(text, pattern string) bool {
	return len(pattern) > 0 && len(text) > 0 && containsSubstring(text, pattern)
}

func containsSubstring(text, substr string) bool {
	for i := 0; i <= len(text)-len(substr); i++ {
		if text[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
