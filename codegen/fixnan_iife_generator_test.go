package codegen

import (
	"strings"
	"testing"
)

func TestFixnanIIFEGenerator_GenerateWithSelfReference(t *testing.T) {
	tests := []struct {
		name            string
		accessor        AccessGenerator
		targetSeriesVar string
		mustContain     []string
		mustNotContain  []string
	}{
		{
			name:            "OHLCV field accessor",
			accessor:        NewOHLCVFieldAccessGenerator("Close"),
			targetSeriesVar: "resultSeries",
			mustContain: []string{
				"func() float64",
				"val := closeSeries.Get(0)",
				"if math.IsNaN(val) { return 0.0 }",
				"return val",
			},
			mustNotContain: []string{
				"selfSeries",
				".Position()",
				"for j :=",
				"fixnanState",
			},
		},
		{
			name:            "arrow function parameter accessor",
			accessor:        NewArrowFunctionParameterAccessor("source"),
			targetSeriesVar: "outputSeries",
			mustContain: []string{
				"func() float64",
				"val := sourceSeries.Get(0)",
				"if math.IsNaN(val) { return 0.0 }",
				"return val",
			},
			mustNotContain: []string{
				"selfSeries",
				".Position()",
				"for j :=",
			},
		},
		{
			name:            "High field accessor",
			accessor:        NewOHLCVFieldAccessGenerator("High"),
			targetSeriesVar: "highFixedSeries",
			mustContain: []string{
				"func() float64",
				"highSeries.Get(0)",
				"if math.IsNaN(val)",
				"return 0.0",
			},
		},
		{
			name:            "Low field accessor",
			accessor:        NewOHLCVFieldAccessGenerator("Low"),
			targetSeriesVar: "lowFixedSeries",
			mustContain: []string{
				"func() float64",
				"lowSeries.Get(0)",
				"return val",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := &FixnanIIFEGenerator{}
			code := gen.GenerateWithSelfReference(tt.accessor, tt.targetSeriesVar)

			for _, pattern := range tt.mustContain {
				if !strings.Contains(code, pattern) {
					t.Errorf("Generated code missing pattern %q\nGot: %s", pattern, code)
				}
			}

			for _, pattern := range tt.mustNotContain {
				if strings.Contains(code, pattern) {
					t.Errorf("Generated code contains unwanted pattern %q\nGot: %s", pattern, code)
				}
			}

			if !strings.Contains(code, "func() float64") {
				t.Error("Generated code must be an IIFE returning float64")
			}
		})
	}
}

func TestFixnanCallExpressionAccessor(t *testing.T) {
	tests := []struct {
		name        string
		tempVarName string
		tempVarCode string
	}{
		{
			name:        "simple temp variable",
			tempVarName: "expr_temp",
			tempVarCode: "expr_temp := rma(...)\n",
		},
		{
			name:        "complex arithmetic expression",
			tempVarName: "calc_temp",
			tempVarCode: "calc_temp := 100 * rma(...) / truerange\n",
		},
		{
			name:        "nested function call",
			tempVarName: "nested_temp",
			tempVarCode: "nested_temp := func() float64 { return 42.0 }()\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accessor := &FixnanCallExpressionAccessor{
				tempVarName: tt.tempVarName,
				tempVarCode: tt.tempVarCode,
			}

			t.Run("GenerateLoopValueAccess", func(t *testing.T) {
				got := accessor.GenerateLoopValueAccess("j")
				if got != tt.tempVarName {
					t.Errorf("Got %q, expected %q", got, tt.tempVarName)
				}

				got0 := accessor.GenerateLoopValueAccess("0")
				if got0 != tt.tempVarName {
					t.Errorf("Got %q, expected %q", got0, tt.tempVarName)
				}
			})

			t.Run("GenerateInitialValueAccess", func(t *testing.T) {
				for _, period := range []int{1, 10, 100} {
					got := accessor.GenerateInitialValueAccess(period)
					if got != tt.tempVarName {
						t.Errorf("Period %d: Got %q, expected %q", period, got, tt.tempVarName)
					}
				}
			})

			t.Run("GetPreamble", func(t *testing.T) {
				got := accessor.GetPreamble()
				if got != tt.tempVarCode {
					t.Errorf("Got %q, expected %q", got, tt.tempVarCode)
				}

				if !strings.Contains(got, tt.tempVarName) {
					t.Errorf("Preamble missing temp variable name %q in: %s", tt.tempVarName, got)
				}
			})
		})
	}
}

func TestFixnanIIFEGenerator_PreambleExtraction(t *testing.T) {
	tests := []struct {
		name            string
		accessor        *FixnanCallExpressionAccessor
		targetSeriesVar string
		mustContain     []string
		mustNotContain  []string
	}{
		{
			name: "arithmetic expression - preamble extracted separately",
			accessor: &FixnanCallExpressionAccessor{
				tempVarName: "expr_temp",
				tempVarCode: "expr_temp := 100 * rma(...) / truerange\n",
			},
			targetSeriesVar: "plusSeries",
			mustContain: []string{
				"func() float64",
				"val := expr_temp",
				"if math.IsNaN(val) { return 0.0 }",
				"return val",
			},
			mustNotContain: []string{
				"expr_temp := 100 * rma(...) / truerange",
				"plusSeries.Position()",
				"plusSeries.Get(j)",
				"for j :=",
			},
		},
		{
			name: "nested function call - preamble not embedded",
			accessor: &FixnanCallExpressionAccessor{
				tempVarName: "nested_result",
				tempVarCode: "nested_result := func() float64 { return sma(...) }()\n",
			},
			targetSeriesVar: "indicatorSeries",
			mustContain: []string{
				"func() float64",
				"val := nested_result",
				"if math.IsNaN(val) { return 0.0 }",
			},
			mustNotContain: []string{
				"nested_result := func() float64",
				"selfSeries",
			},
		},
		{
			name: "division operation - clean IIFE only",
			accessor: &FixnanCallExpressionAccessor{
				tempVarName: "ratio_temp",
				tempVarCode: "ratio_temp := numerator / denominator\n",
			},
			targetSeriesVar: "ratioSeries",
			mustContain: []string{
				"func() float64",
				"val := ratio_temp",
				"return 0.0",
			},
			mustNotContain: []string{
				"ratio_temp := numerator / denominator",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := &FixnanIIFEGenerator{}
			code := gen.GenerateWithSelfReference(tt.accessor, tt.targetSeriesVar)

			for _, pattern := range tt.mustContain {
				if !strings.Contains(code, pattern) {
					t.Errorf("Code missing pattern %q\nGot: %s", pattern, code)
				}
			}

			for _, pattern := range tt.mustNotContain {
				if strings.Contains(code, pattern) {
					t.Errorf("Code contains unwanted pattern %q (preambles should be extracted separately)\nGot: %s", pattern, code)
				}
			}

			if !strings.HasPrefix(code, "func() float64 {") {
				t.Errorf("Expected IIFE to start with 'func() float64 {', got: %s", code[:min(50, len(code))])
			}
		})
	}
}

func TestPreambleExtractor_ExtractsFromAccessor(t *testing.T) {
	tests := []struct {
		name             string
		accessor         AccessGenerator
		expectedPreamble string
	}{
		{
			name: "accessor with preamble",
			accessor: &FixnanCallExpressionAccessor{
				tempVarName: "expr_temp",
				tempVarCode: "expr_temp := 100 * rma(...) / truerange\n",
			},
			expectedPreamble: "expr_temp := 100 * rma(...) / truerange\n",
		},
		{
			name: "accessor without preamble",
			accessor: &FixnanCallExpressionAccessor{
				tempVarName: "simple",
				tempVarCode: "",
			},
			expectedPreamble: "",
		},
		{
			name: "non-preamble accessor",
			accessor: &struct {
				AccessGenerator
			}{},
			expectedPreamble: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			extractor := NewPreambleExtractor()
			preamble := extractor.ExtractFromAccessor(tt.accessor)

			if preamble != tt.expectedPreamble {
				t.Errorf("Expected preamble %q, got %q", tt.expectedPreamble, preamble)
			}
		})
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func TestFixnanIIFEGenerator_EdgeCases(t *testing.T) {
	tests := []struct {
		name            string
		accessor        AccessGenerator
		targetSeriesVar string
		description     string
	}{
		{
			name:            "empty target series variable name",
			accessor:        NewOHLCVFieldAccessGenerator("Close"),
			targetSeriesVar: "",
			description:     "Should handle empty series name",
		},
		{
			name:            "long series variable name",
			accessor:        NewOHLCVFieldAccessGenerator("Volume"),
			targetSeriesVar: "veryLongSeriesVariableNameThatExceedsTypicalLength",
			description:     "Should handle long variable names",
		},
		{
			name:            "series name with underscores",
			accessor:        NewArrowFunctionParameterAccessor("my_custom_source"),
			targetSeriesVar: "result_output_series",
			description:     "Should handle underscores in names",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := &FixnanIIFEGenerator{}
			code := gen.GenerateWithSelfReference(tt.accessor, tt.targetSeriesVar)

			if code == "" {
				t.Errorf("%s: Generated empty code", tt.description)
			}

			if !strings.Contains(code, "func() float64") {
				t.Errorf("%s: Missing IIFE structure", tt.description)
			}

			if !strings.Contains(code, "if math.IsNaN(val) { return 0.0 }") {
				t.Errorf("%s: Missing NaN check", tt.description)
			}

			if !strings.Contains(code, "return val") {
				t.Errorf("%s: Missing value return", tt.description)
			}
		})
	}
}

func TestFixnanIIFEGenerator_CodeStructure(t *testing.T) {
	accessors := []struct {
		name     string
		accessor AccessGenerator
	}{
		{"OHLCV Close", NewOHLCVFieldAccessGenerator("Close")},
		{"OHLCV High", NewOHLCVFieldAccessGenerator("High")},
		{"OHLCV Low", NewOHLCVFieldAccessGenerator("Low")},
		{"OHLCV Open", NewOHLCVFieldAccessGenerator("Open")},
		{"OHLCV Volume", NewOHLCVFieldAccessGenerator("Volume")},
		{"Parameter x", NewArrowFunctionParameterAccessor("x")},
		{"Parameter data", NewArrowFunctionParameterAccessor("data")},
	}

	for _, tt := range accessors {
		t.Run(tt.name, func(t *testing.T) {
			gen := &FixnanIIFEGenerator{}
			code := gen.GenerateWithSelfReference(tt.accessor, "testSeries")

			requiredPatterns := []string{
				"func() float64",
				"val :=",
				"if math.IsNaN(val)",
				"return 0.0",
				"return val",
			}

			for _, pattern := range requiredPatterns {
				if !strings.Contains(code, pattern) {
					t.Errorf("Missing required pattern %q in generated code:\n%s", pattern, code)
				}
			}

			forbiddenPatterns := []string{
				"selfSeries",
				".Position()",
				"for j :=",
				"fixnanState",
				"lastValidValue",
				"Series.Get(j)",
			}

			for _, pattern := range forbiddenPatterns {
				if strings.Contains(code, pattern) {
					t.Errorf("Found forbidden pattern %q in generated code:\n%s", pattern, code)
				}
			}
		})
	}
}
