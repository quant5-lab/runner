package codegen

import (
	"strings"
	"testing"
)

func TestSeriesBackedComplexExpressionAccessor_GenerateLoopValueAccess(t *testing.T) {
	tests := []struct {
		name           string
		seriesName     string
		loopVar        string
		expectedOutput string
	}{
		{
			name:           "standard loop variable j",
			seriesName:     "ternary_source_tempSeries",
			loopVar:        "j",
			expectedOutput: "ternary_source_tempSeries.Get(j)",
		},
		{
			name:           "loop variable i",
			seriesName:     "binary_source_tempSeries",
			loopVar:        "i",
			expectedOutput: "binary_source_tempSeries.Get(i)",
		},
		{
			name:           "call expression series",
			seriesName:     "call_source_tempSeries",
			loopVar:        "j",
			expectedOutput: "call_source_tempSeries.Get(j)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accessor := NewSeriesBackedComplexExpressionAccessor(tt.seriesName, "")
			result := accessor.GenerateLoopValueAccess(tt.loopVar)

			if result != tt.expectedOutput {
				t.Errorf("GenerateLoopValueAccess(%q) = %q, want %q",
					tt.loopVar, result, tt.expectedOutput)
			}

			if !strings.Contains(result, ".Get(") {
				t.Errorf("Expected historical access .Get( in %q", result)
			}
		})
	}
}

func TestSeriesBackedComplexExpressionAccessor_GenerateInitialValueAccess(t *testing.T) {
	tests := []struct {
		name           string
		seriesName     string
		period         int
		expectedOutput string
	}{
		{
			name:           "RMA period 20",
			seriesName:     "ternary_source_tempSeries",
			period:         20,
			expectedOutput: "ternary_source_tempSeries.Get(0)",
		},
		{
			name:           "EMA period 12",
			seriesName:     "binary_source_tempSeries",
			period:         12,
			expectedOutput: "binary_source_tempSeries.Get(0)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accessor := NewSeriesBackedComplexExpressionAccessor(tt.seriesName, "")
			result := accessor.GenerateInitialValueAccess(tt.period)

			if result != tt.expectedOutput {
				t.Errorf("GenerateInitialValueAccess(%d) = %q, want %q",
					tt.period, result, tt.expectedOutput)
			}

			if !strings.Contains(result, ".Get(0)") {
				t.Errorf("Expected current bar access .Get(0) in %q", result)
			}
		})
	}
}

func TestSeriesBackedComplexExpressionAccessor_GetPreamble(t *testing.T) {
	tests := []struct {
		name           string
		seriesName     string
		expressionCode string
		expectedInit   string
		expectedSet    string
	}{
		{
			name:           "ternary expression",
			seriesName:     "ternary_source_tempSeries",
			expressionCode: "func() float64 { if cond { return a } else { return b } }()",
			expectedInit:   `ternary_source_tempSeries := arrowCtx.GetOrCreateSeries("ternary_source_temp")`,
			expectedSet:    "ternary_source_tempSeries.Set(func() float64 { if cond { return a } else { return b } }())",
		},
		{
			name:           "binary expression",
			seriesName:     "binary_source_tempSeries",
			expressionCode: "(a + b)",
			expectedInit:   `binary_source_tempSeries := arrowCtx.GetOrCreateSeries("binary_source_temp")`,
			expectedSet:    "binary_source_tempSeries.Set((a + b))",
		},
		{
			name:           "call expression",
			seriesName:     "call_source_tempSeries",
			expressionCode: "someFunc(arg1, arg2)",
			expectedInit:   `call_source_tempSeries := arrowCtx.GetOrCreateSeries("call_source_temp")`,
			expectedSet:    "call_source_tempSeries.Set(someFunc(arg1, arg2))",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accessor := NewSeriesBackedComplexExpressionAccessor(tt.seriesName, tt.expressionCode)
			result := accessor.GetPreamble()

			if !strings.Contains(result, tt.expectedInit) {
				t.Errorf("GetPreamble() missing initialization:\nGot:  %q\nWant: %q", result, tt.expectedInit)
			}

			if !strings.Contains(result, tt.expectedSet) {
				t.Errorf("GetPreamble() missing Set() call:\nGot:  %q\nWant: %q", result, tt.expectedSet)
			}

			if !strings.Contains(result, "GetOrCreateSeries") {
				t.Errorf("Expected GetOrCreateSeries() in preamble: %q", result)
			}

			if !strings.Contains(result, ".Set(") {
				t.Errorf("Expected Series.Set() call in preamble: %q", result)
			}
		})
	}
}

func TestSeriesBackedComplexExpressionAccessor_GetSeriesName(t *testing.T) {
	seriesName := "ternary_source_tempSeries"
	accessor := NewSeriesBackedComplexExpressionAccessor(seriesName, "")

	result := accessor.GetSeriesName()
	if result != seriesName {
		t.Errorf("GetSeriesName() = %q, want %q", result, seriesName)
	}
}

func TestSeriesBackedComplexExpressionAccessor_NeedsAdvance(t *testing.T) {
	accessor := NewSeriesBackedComplexExpressionAccessor("test", "")

	if !accessor.NeedsAdvance() {
		t.Error("Expected NeedsAdvance() = true, temp series must advance cursor")
	}
}

func TestSeriesBackedComplexExpressionAccessor_RMAIntegration(t *testing.T) {
	/* Simulate RMA inline generation with series-backed accessor */
	seriesName := "ternary_source_tempSeries"
	expressionCode := "func() float64 { if (up > down && up > 0) { return up } else { return 0 } }()"

	accessor := NewSeriesBackedComplexExpressionAccessor(seriesName, expressionCode)

	/* Test preamble generates both GetOrCreateSeries and Set() */
	preamble := accessor.GetPreamble()

	if !strings.Contains(preamble, "GetOrCreateSeries") {
		t.Errorf("RMA preamble missing GetOrCreateSeries:\n%q", preamble)
	}

	if !strings.Contains(preamble, ".Set(") {
		t.Errorf("RMA preamble missing Set() call:\n%q", preamble)
	}

	/* Test initial value access for RMA start */
	initialAccess := accessor.GenerateInitialValueAccess(20)
	expectedInitial := "ternary_source_tempSeries.Get(0)"
	if initialAccess != expectedInitial {
		t.Errorf("RMA initial access = %q, want %q", initialAccess, expectedInitial)
	}

	/* Test loop value access for historical lookback */
	loopAccess := accessor.GenerateLoopValueAccess("j")
	expectedLoop := "ternary_source_tempSeries.Get(j)"
	if loopAccess != expectedLoop {
		t.Errorf("RMA loop access = %q, want %q", loopAccess, expectedLoop)
	}

	/* Verify historical access pattern (positive offset) */
	if !strings.Contains(loopAccess, ".Get(j)") {
		t.Error("RMA loop access must use positive offset for historical values")
	}
}

func TestSeriesBackedComplexExpressionAccessor_ComparisonWithFixnan(t *testing.T) {
	/* Demonstrate difference: SeriesBacked vs FixnanCallExpression */

	seriesBacked := NewSeriesBackedComplexExpressionAccessor("temp_series", "expr()")
	fixnan := &FixnanCallExpressionAccessor{
		tempVarName: "temp_var",
		tempVarCode: "temp_var := expr()",
	}

	/* SeriesBacked: returns historical access */
	seriesLoop := seriesBacked.GenerateLoopValueAccess("j")
	if !strings.Contains(seriesLoop, ".Get(j)") {
		t.Errorf("SeriesBacked must provide historical access: %q", seriesLoop)
	}

	/* Fixnan: returns same scalar (no historical access) */
	fixnanLoop := fixnan.GenerateLoopValueAccess("j")
	if fixnanLoop != "temp_var" {
		t.Errorf("Fixnan returns scalar: got %q, want 'temp_var'", fixnanLoop)
	}

	/* KEY DIFFERENCE: SeriesBacked enables RMA/EMA, Fixnan does not */
	if seriesLoop == fixnanLoop {
		t.Error("SeriesBacked and Fixnan must generate different access patterns")
	}
}
