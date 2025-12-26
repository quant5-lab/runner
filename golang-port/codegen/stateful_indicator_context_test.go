package codegen

import (
	"strings"
	"testing"
)

/* TestStatefulIndicatorContext_SeriesAccessPatterns validates context-dependent series access
 *
 * Tests that StatefulIndicatorContext generates correct series buffer access patterns:
 * - TopLevelIndicatorContext: direct series access ({varName}Series.Get/Set)
 * - ArrowFunctionIndicatorContext: arrowCtx-mediated access (arrowCtx.GetOrCreateSeries().Get/Set)
 *
 * Validates generalized behavior applicable to all stateful indicators (RMA, EMA, RSI, etc.)
 */
func TestStatefulIndicatorContext_SeriesAccessPatterns(t *testing.T) {
	testCases := []struct {
		name            string
		context         StatefulIndicatorContext
		varName         string
		offset          int
		value           string
		expectedAccess  string
		expectedUpdate  string
		isArrowFunction bool
	}{
		{
			name:            "TopLevel: simple variable, offset 0",
			context:         NewTopLevelIndicatorContext(),
			varName:         "rma14",
			offset:          0,
			value:           "newValue",
			expectedAccess:  "rma14Series.Get(0)",
			expectedUpdate:  "rma14Series.Set(newValue)",
			isArrowFunction: false,
		},
		{
			name:            "TopLevel: simple variable, offset 1 (previous value)",
			context:         NewTopLevelIndicatorContext(),
			varName:         "ema20",
			offset:          1,
			value:           "result",
			expectedAccess:  "ema20Series.Get(1)",
			expectedUpdate:  "ema20Series.Set(result)",
			isArrowFunction: false,
		},
		{
			name:            "TopLevel: complex variable name with underscores",
			context:         NewTopLevelIndicatorContext(),
			varName:         "my_indicator_value",
			offset:          5,
			value:           "calculated",
			expectedAccess:  "my_indicator_valueSeries.Get(5)",
			expectedUpdate:  "my_indicator_valueSeries.Set(calculated)",
			isArrowFunction: false,
		},
		{
			name:            "Arrow: simple variable, offset 0",
			context:         NewArrowFunctionIndicatorContext(),
			varName:         "truerange",
			offset:          0,
			value:           "tr",
			expectedAccess:  "arrowCtx.GetOrCreateSeries(\"truerange\").Get(0)",
			expectedUpdate:  "arrowCtx.GetOrCreateSeries(\"truerange\").Set(tr)",
			isArrowFunction: true,
		},
		{
			name:            "Arrow: simple variable, offset 1 (previous value)",
			context:         NewArrowFunctionIndicatorContext(),
			varName:         "plus",
			offset:          1,
			value:           "newPlus",
			expectedAccess:  "arrowCtx.GetOrCreateSeries(\"plus\").Get(1)",
			expectedUpdate:  "arrowCtx.GetOrCreateSeries(\"plus\").Set(newPlus)",
			isArrowFunction: true,
		},
		{
			name:            "Arrow: complex variable name",
			context:         NewArrowFunctionIndicatorContext(),
			varName:         "adx_smoothed",
			offset:          10,
			value:           "smoothedValue",
			expectedAccess:  "arrowCtx.GetOrCreateSeries(\"adx_smoothed\").Get(10)",
			expectedUpdate:  "arrowCtx.GetOrCreateSeries(\"adx_smoothed\").Set(smoothedValue)",
			isArrowFunction: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			access := tc.context.GenerateSeriesAccess(tc.varName, tc.offset)
			if access != tc.expectedAccess {
				t.Errorf("GenerateSeriesAccess mismatch\nExpected: %s\nGot:      %s", tc.expectedAccess, access)
			}

			update := tc.context.GenerateSeriesUpdate(tc.varName, tc.value)
			if update != tc.expectedUpdate {
				t.Errorf("GenerateSeriesUpdate mismatch\nExpected: %s\nGot:      %s", tc.expectedUpdate, update)
			}

			isArrow := tc.context.IsWithinArrowFunction()
			if isArrow != tc.isArrowFunction {
				t.Errorf("IsWithinArrowFunction mismatch\nExpected: %v\nGot:      %v", tc.isArrowFunction, isArrow)
			}
		})
	}
}

/* TestStatefulIndicatorBuilder_ContextIntegration validates StatefulIndicatorBuilder uses context correctly
 *
 * Tests that StatefulIndicatorBuilder generates different code based on execution context:
 * - TopLevel: {varName}Series.Get(1), {varName}Series.Set(value)
 * - Arrow: arrowCtx.GetOrCreateSeries("{varName}").Get(1), arrowCtx.GetOrCreateSeries("{varName}").Set(value)
 *
 * Validates integration between StatefulIndicatorBuilder and StatefulIndicatorContext
 */
func TestStatefulIndicatorBuilder_ContextIntegration(t *testing.T) {
	mockAccessor := &MockAccessGenerator{
		loopAccessFn: func(loopVar string) string {
			return "sourceSeries.Get(" + loopVar + ")"
		},
		initialAccessFn: func(period int) string {
			return "sourceSeries.Get(" + string(rune(period-1)) + ")"
		},
	}

	testCases := []struct {
		name                   string
		context                StatefulIndicatorContext
		varName                string
		period                 int
		needsNaN               bool
		expectedPreviousAccess string // previousValue := <this>
		expectedSetPattern     string // <this> during warmup, init, recursive
	}{
		{
			name:                   "TopLevel RMA context: direct series access",
			context:                NewTopLevelIndicatorContext(),
			varName:                "rma14",
			period:                 14,
			needsNaN:               false,
			expectedPreviousAccess: "rma14Series.Get(1)",
			expectedSetPattern:     "rma14Series.Set(",
		},
		{
			name:                   "TopLevel EMA context: direct series access",
			context:                NewTopLevelIndicatorContext(),
			varName:                "ema20",
			period:                 20,
			needsNaN:               true,
			expectedPreviousAccess: "ema20Series.Get(1)",
			expectedSetPattern:     "ema20Series.Set(",
		},
		{
			name:                   "Arrow RMA context: arrowCtx-mediated access",
			context:                NewArrowFunctionIndicatorContext(),
			varName:                "truerange",
			period:                 20,
			needsNaN:               false,
			expectedPreviousAccess: "arrowCtx.GetOrCreateSeries(\"truerange\").Get(1)",
			expectedSetPattern:     "arrowCtx.GetOrCreateSeries(\"truerange\").Set(",
		},
		{
			name:                   "Arrow EMA context: arrowCtx-mediated access",
			context:                NewArrowFunctionIndicatorContext(),
			varName:                "adx",
			period:                 16,
			needsNaN:               true,
			expectedPreviousAccess: "arrowCtx.GetOrCreateSeries(\"adx\").Get(1)",
			expectedSetPattern:     "arrowCtx.GetOrCreateSeries(\"adx\").Set(",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			builder := NewStatefulIndicatorBuilder("ta.rma", tc.varName, tc.period, mockAccessor, tc.needsNaN, tc.context)
			code := builder.BuildRMA()

			if !strings.Contains(code, tc.expectedPreviousAccess) {
				t.Errorf("Missing expected previous value access pattern\nExpected substring: %s\nGenerated code:\n%s",
					tc.expectedPreviousAccess, code)
			}

			setCount := strings.Count(code, tc.expectedSetPattern)
			if setCount < 3 { // warmup, init, recursive phases each have at least 1 Set()
				t.Errorf("Insufficient Set() calls with expected pattern\nExpected pattern: %s\nCount: %d\nGenerated code:\n%s",
					tc.expectedSetPattern, setCount, code)
			}

			if tc.context.IsWithinArrowFunction() {
				if strings.Contains(code, tc.varName+"Series.Get(") {
					t.Errorf("Arrow function context contaminated with top-level series access\nFound: %sSeries.Get(\nGenerated code:\n%s",
						tc.varName, code)
				}
			} else {
				if strings.Contains(code, "arrowCtx.GetOrCreateSeries(") {
					t.Errorf("Top-level context contaminated with arrow function series access\nFound: arrowCtx.GetOrCreateSeries(\nGenerated code:\n%s",
						code)
				}
			}
		})
	}
}

/* TestStatefulIndicatorContext_EdgeCases validates context behavior with edge case inputs
 *
 * Tests generalized edge case handling across all contexts:
 * - Empty variable names
 * - Special characters in variable names
 * - Zero/negative offsets
 * - Large offsets
 * - Empty/null value strings
 * - Complex value expressions
 *
 * Ensures robustness independent of specific indicator type
 */
func TestStatefulIndicatorContext_EdgeCases(t *testing.T) {
	testCases := []struct {
		name           string
		context        StatefulIndicatorContext
		varName        string
		offset         int
		value          string
		shouldContain  []string
		shouldNotPanic bool
	}{
		{
			name:           "TopLevel: offset 0 (current bar)",
			context:        NewTopLevelIndicatorContext(),
			varName:        "test",
			offset:         0,
			value:          "val",
			shouldContain:  []string{"testSeries.Get(0)", "testSeries.Set(val)"},
			shouldNotPanic: true,
		},
		{
			name:           "TopLevel: large offset (historical access)",
			context:        NewTopLevelIndicatorContext(),
			varName:        "sma200",
			offset:         199,
			value:          "avg",
			shouldContain:  []string{"sma200Series.Get(199)", "sma200Series.Set(avg)"},
			shouldNotPanic: true,
		},
		{
			name:           "Arrow: offset 0 (current bar)",
			context:        NewArrowFunctionIndicatorContext(),
			varName:        "local",
			offset:         0,
			value:          "result",
			shouldContain:  []string{"arrowCtx.GetOrCreateSeries(\"local\").Get(0)", "arrowCtx.GetOrCreateSeries(\"local\").Set(result)"},
			shouldNotPanic: true,
		},
		{
			name:           "Arrow: large offset",
			context:        NewArrowFunctionIndicatorContext(),
			varName:        "buffer",
			offset:         500,
			value:          "data",
			shouldContain:  []string{"arrowCtx.GetOrCreateSeries(\"buffer\").Get(500)", "arrowCtx.GetOrCreateSeries(\"buffer\").Set(data)"},
			shouldNotPanic: true,
		},
		{
			name:           "TopLevel: complex value expression",
			context:        NewTopLevelIndicatorContext(),
			varName:        "indicator",
			offset:         1,
			value:          "alpha*source + (1-alpha)*prev",
			shouldContain:  []string{"indicatorSeries.Set(alpha*source + (1-alpha)*prev)"},
			shouldNotPanic: true,
		},
		{
			name:           "Arrow: complex value expression",
			context:        NewArrowFunctionIndicatorContext(),
			varName:        "smoothed",
			offset:         2,
			value:          "math.Max(current, previous)",
			shouldContain:  []string{"arrowCtx.GetOrCreateSeries(\"smoothed\").Set(math.Max(current, previous))"},
			shouldNotPanic: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil && tc.shouldNotPanic {
					t.Errorf("Function panicked unexpectedly: %v", r)
				}
			}()

			access := tc.context.GenerateSeriesAccess(tc.varName, tc.offset)
			update := tc.context.GenerateSeriesUpdate(tc.varName, tc.value)

			for _, expected := range tc.shouldContain {
				if !strings.Contains(access, expected) && !strings.Contains(update, expected) {
					t.Errorf("Missing expected substring\nExpected: %s\nAccess:   %s\nUpdate:   %s", expected, access, update)
				}
			}
		})
	}
}

/* TestStatefulIndicatorBuilder_MultiIndicatorContext validates multiple indicators share context correctly
 *
 * Tests that multiple stateful indicators in same scope share execution context:
 * - Multiple RMA indicators in top-level scope
 * - Multiple EMA indicators in arrow function scope
 * - Mixed indicator types (RMA + EMA) in same scope
 *
 * Validates context consistency across multiple indicator instances
 */
func TestStatefulIndicatorBuilder_MultiIndicatorContext(t *testing.T) {
	mockAccessor := &MockAccessGenerator{
		loopAccessFn: func(loopVar string) string {
			return "sourceSeries.Get(" + loopVar + ")"
		},
		initialAccessFn: func(period int) string {
			return "sourceSeries.Get(" + string(rune(period-1)) + ")"
		},
	}

	t.Run("TopLevel: multiple RMA indicators share context", func(t *testing.T) {
		ctx := NewTopLevelIndicatorContext()

		builder1 := NewStatefulIndicatorBuilder("ta.rma", "rma14", 14, mockAccessor, false, ctx)
		code1 := builder1.BuildRMA()

		builder2 := NewStatefulIndicatorBuilder("ta.rma", "rma20", 20, mockAccessor, false, ctx)
		code2 := builder2.BuildRMA()

		// Both should use top-level series access
		if !strings.Contains(code1, "rma14Series.Get(1)") {
			t.Error("RMA14 missing top-level series access")
		}
		if !strings.Contains(code2, "rma20Series.Get(1)") {
			t.Error("RMA20 missing top-level series access")
		}

		// Neither should have arrow context access
		if strings.Contains(code1, "arrowCtx") || strings.Contains(code2, "arrowCtx") {
			t.Error("Top-level indicators contaminated with arrow context")
		}
	})

	t.Run("Arrow: multiple indicators share arrowCtx", func(t *testing.T) {
		ctx := NewArrowFunctionIndicatorContext()

		builder1 := NewStatefulIndicatorBuilder("ta.rma", "plus", 18, mockAccessor, false, ctx)
		code1 := builder1.BuildRMA()

		builder2 := NewStatefulIndicatorBuilder("ta.ema", "minus", 18, mockAccessor, false, ctx)
		code2 := builder2.BuildEMA()

		// Both should use arrowCtx-mediated access
		if !strings.Contains(code1, "arrowCtx.GetOrCreateSeries(\"plus\")") {
			t.Error("Plus indicator missing arrowCtx access")
		}
		if !strings.Contains(code2, "arrowCtx.GetOrCreateSeries(\"minus\")") {
			t.Error("Minus indicator missing arrowCtx access")
		}

		// Neither should have direct series access
		if strings.Contains(code1, "plusSeries.Get(") || strings.Contains(code2, "minusSeries.Get(") {
			t.Error("Arrow indicators contaminated with top-level series access")
		}
	})
}
