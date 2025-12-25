package codegen

import (
	"strings"
	"testing"
)

func TestInlineLoopExpressionAccessor_GenerateLoopValueAccess(t *testing.T) {
	tests := []struct {
		name               string
		expressionCode     string
		loopVar            string
		expectedContains   []string
		expectedNotContain string
	}{
		{
			name:           "ternary with series access",
			expressionCode: "func() float64 { if upSeries.GetCurrent() > downSeries.GetCurrent() { return upSeries.GetCurrent() } else { return 0.0 } }()",
			loopVar:        "j",
			expectedContains: []string{
				"upSeries.Get(j)",
				"downSeries.Get(j)",
			},
			expectedNotContain: ".GetCurrent()",
		},
		{
			name:           "binary expression with series",
			expressionCode: "(upSeries.GetCurrent() + downSeries.GetCurrent())",
			loopVar:        "j",
			expectedContains: []string{
				"upSeries.Get(j)",
				"downSeries.Get(j)",
			},
			expectedNotContain: ".GetCurrent()",
		},
		{
			name:           "complex nested ternary",
			expressionCode: "func() float64 { if (upSeries.GetCurrent() > 0 && upSeries.GetCurrent() > downSeries.GetCurrent()) { return upSeries.GetCurrent() } else { return 0.0 } }()",
			loopVar:        "j",
			expectedContains: []string{
				"upSeries.Get(j) > 0",
				"upSeries.Get(j) > downSeries.Get(j)",
			},
			expectedNotContain: ".GetCurrent()",
		},
		{
			name:           "loop variable i",
			expressionCode: "valSeries.GetCurrent()",
			loopVar:        "i",
			expectedContains: []string{
				"valSeries.Get(i)",
			},
			expectedNotContain: ".GetCurrent()",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accessor := NewInlineLoopExpressionAccessor(tt.expressionCode)
			result := accessor.GenerateLoopValueAccess(tt.loopVar)

			for _, expected := range tt.expectedContains {
				if !strings.Contains(result, expected) {
					t.Errorf("Expected %q to contain %q", result, expected)
				}
			}

			if tt.expectedNotContain != "" && strings.Contains(result, tt.expectedNotContain) {
				t.Errorf("Expected %q to NOT contain %q", result, tt.expectedNotContain)
			}
		})
	}
}

func TestInlineLoopExpressionAccessor_GenerateInitialValueAccess(t *testing.T) {
	tests := []struct {
		name             string
		expressionCode   string
		period           int
		expectedContains []string
	}{
		{
			name:           "ternary initial value",
			expressionCode: "func() float64 { if upSeries.GetCurrent() > 0 { return upSeries.GetCurrent() } else { return 0.0 } }()",
			period:         20,
			expectedContains: []string{
				"upSeries.Get(0)",
				"return upSeries.Get(0)",
			},
		},
		{
			name:           "binary initial value",
			expressionCode: "(aSeries.GetCurrent() + bSeries.GetCurrent())",
			period:         10,
			expectedContains: []string{
				"aSeries.Get(0)",
				"bSeries.Get(0)",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accessor := NewInlineLoopExpressionAccessor(tt.expressionCode)
			result := accessor.GenerateInitialValueAccess(tt.period)

			for _, expected := range tt.expectedContains {
				if !strings.Contains(result, expected) {
					t.Errorf("Expected %q to contain %q", result, expected)
				}
			}

			if strings.Contains(result, ".GetCurrent()") {
				t.Errorf("Initial value access should use .Get(0), not .GetCurrent(): %q", result)
			}
		})
	}
}

func TestInlineLoopExpressionAccessor_GetPreamble(t *testing.T) {
	accessor := NewInlineLoopExpressionAccessor("test expression")
	preamble := accessor.GetPreamble()

	if preamble != "" {
		t.Errorf("Expected empty preamble, got %q", preamble)
	}
}

func TestInlineLoopExpressionAccessor_RMAIntegration(t *testing.T) {
	/* Simulate RMA inline generation with loop expression accessor */
	ternaryExpr := "func() float64 { if (upSeries.GetCurrent() > downSeries.GetCurrent() && upSeries.GetCurrent() > 0.00) { return upSeries.GetCurrent() } else { return 0.00 } }()"

	accessor := NewInlineLoopExpressionAccessor(ternaryExpr)

	/* No preamble needed - inline evaluation */
	preamble := accessor.GetPreamble()
	if preamble != "" {
		t.Errorf("Inline expression should have no preamble, got: %q", preamble)
	}

	/* Initial value uses current bar (Get(0)) */
	initialAccess := accessor.GenerateInitialValueAccess(20)
	if !strings.Contains(initialAccess, "upSeries.Get(0)") {
		t.Errorf("Initial access must use Get(0): %q", initialAccess)
	}

	/* Loop iterations use historical access Get(j) */
	loopAccess := accessor.GenerateLoopValueAccess("j")
	if !strings.Contains(loopAccess, "upSeries.Get(j)") {
		t.Errorf("Loop access must use Get(j): %q", loopAccess)
	}
	if !strings.Contains(loopAccess, "downSeries.Get(j)") {
		t.Errorf("Loop access must transform all series: %q", loopAccess)
	}

	/* Verify NO .GetCurrent() remains */
	if strings.Contains(loopAccess, ".GetCurrent()") {
		t.Errorf("Loop access must not contain .GetCurrent(): %q", loopAccess)
	}
}

func TestInlineLoopExpressionAccessor_MultipleSeriesAccess(t *testing.T) {
	/* Test expression with multiple series references */
	expr := "func() float64 { return math.Max(highSeries.GetCurrent(), lowSeries.GetCurrent()) + closeSeries.GetCurrent() }()"

	accessor := NewInlineLoopExpressionAccessor(expr)
	result := accessor.GenerateLoopValueAccess("j")

	expectedTransforms := []string{
		"highSeries.Get(j)",
		"lowSeries.Get(j)",
		"closeSeries.Get(j)",
	}

	for _, expected := range expectedTransforms {
		if !strings.Contains(result, expected) {
			t.Errorf("Expected all series transformed to Get(j):\nGot: %q\nMissing: %q", result, expected)
		}
	}

	if strings.Contains(result, ".GetCurrent()") {
		t.Errorf("All .GetCurrent() should be transformed: %q", result)
	}
}

func TestInlineLoopExpressionAccessor_ComparisonWithSeriesBacked(t *testing.T) {
	/* Demonstrate difference: InlineLoop vs SeriesBacked */

	expr := "func() float64 { if aSeries.GetCurrent() > 0 { return aSeries.GetCurrent() } else { return 0 } }()"

	inlineLoop := NewInlineLoopExpressionAccessor(expr)
	seriesBacked := NewSeriesBackedComplexExpressionAccessor("tempSeries", expr)

	/* InlineLoop: transforms expression per iteration */
	inlineResult := inlineLoop.GenerateLoopValueAccess("j")
	if !strings.Contains(inlineResult, "aSeries.Get(j)") {
		t.Error("InlineLoop must transform series access to Get(j)")
	}

	/* SeriesBacked: accesses pre-computed temp series */
	seriesResult := seriesBacked.GenerateLoopValueAccess("j")
	if !strings.Contains(seriesResult, "tempSeries.Get(j)") {
		t.Error("SeriesBacked must access temp series")
	}

	/* KEY DIFFERENCE: InlineLoop re-evaluates, SeriesBacked accesses stored values */
	if inlineResult == seriesResult {
		t.Error("InlineLoop and SeriesBacked must generate different code")
	}

	/* InlineLoop: no preamble */
	if inlineLoop.GetPreamble() != "" {
		t.Error("InlineLoop should have empty preamble")
	}

	/* SeriesBacked: has preamble for initialization */
	if seriesBacked.GetPreamble() == "" {
		t.Error("SeriesBacked should have non-empty preamble")
	}
}

func TestInlineLoopExpressionAccessor_LocalVariableTransformation(t *testing.T) {
	/* Test transformation of local variables with series backing */

	/* Create mock resolver */
	accessResolver := NewArrowSeriesAccessResolver()
	accessResolver.RegisterLocalVariable("sum")
	identifierResolver := NewArrowIdentifierResolver(accessResolver)

	/* Expression using local variable 'sum' */
	expr := "func() float64 { if (sum == 0) { return 1 } else { return sum } }()"

	accessor := NewInlineLoopExpressionAccessorWithResolver(expr, identifierResolver)

	/* Loop access should transform 'sum' to 'sumSeries.Get(j)' */
	loopAccess := accessor.GenerateLoopValueAccess("j")

	if !strings.Contains(loopAccess, "sumSeries.Get(j)") {
		t.Errorf("Expected 'sumSeries.Get(j)' in transformed expression\nGot: %s", loopAccess)
	}

	/* Verify transformation happened at both occurrences */
	occurrences := strings.Count(loopAccess, "sumSeries.Get(j)")
	if occurrences != 2 {
		t.Errorf("Expected 2 occurrences of 'sumSeries.Get(j)', got %d\nFull output: %s", occurrences, loopAccess)
	}
}
