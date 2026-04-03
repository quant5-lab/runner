package codegen

import (
	"fmt"
	"strings"
	"testing"
)

/* TestStatefulIndicatorBuilder_PeriodBoundaries validates extreme period values
 *
 * Tests that stateful indicators handle boundary conditions for period parameter:
 * - Minimum period (1, 2)
 * - Typical periods (5, 10, 14, 20, 50)
 * - Large periods (100, 200, 500)
 *
 * Validates:
 * - Warmup threshold calculation (period-1)
 * - Initialization bar calculation (period-1)
 * - Loop range correctness
 * - Alpha formula correctness
 */
func TestStatefulIndicatorBuilder_PeriodBoundaries(t *testing.T) {
	testCases := []struct {
		name      string
		period    int
		warmupBar int // ctx.BarIndex < warmupBar
		initBar   int // ctx.BarIndex == initBar
		loopCount int // for j := 0; j < loopCount
		alphaRMA  string
		alphaEMA  string
	}{
		{
			name:      "Period 1 (minimum)",
			period:    1,
			warmupBar: 0,
			initBar:   0,
			loopCount: 1,
			alphaRMA:  "1.0 / float64(1)",
			alphaEMA:  "2.0 / float64(1+1)",
		},
		{
			name:      "Period 2 (edge)",
			period:    2,
			warmupBar: 1,
			initBar:   1,
			loopCount: 2,
			alphaRMA:  "1.0 / float64(2)",
			alphaEMA:  "2.0 / float64(2+1)",
		},
		{
			name:      "Period 5 (small)",
			period:    5,
			warmupBar: 4,
			initBar:   4,
			loopCount: 5,
			alphaRMA:  "1.0 / float64(5)",
			alphaEMA:  "2.0 / float64(5+1)",
		},
		{
			name:      "Period 14 (typical)",
			period:    14,
			warmupBar: 13,
			initBar:   13,
			loopCount: 14,
			alphaRMA:  "1.0 / float64(14)",
			alphaEMA:  "2.0 / float64(14+1)",
		},
		{
			name:      "Period 50 (medium)",
			period:    50,
			warmupBar: 49,
			initBar:   49,
			loopCount: 50,
			alphaRMA:  "1.0 / float64(50)",
			alphaEMA:  "2.0 / float64(50+1)",
		},
		{
			name:      "Period 100 (large)",
			period:    100,
			warmupBar: 99,
			initBar:   99,
			loopCount: 100,
			alphaRMA:  "1.0 / float64(100)",
			alphaEMA:  "2.0 / float64(100+1)",
		},
		{
			name:      "Period 200 (very large)",
			period:    200,
			warmupBar: 199,
			initBar:   199,
			loopCount: 200,
			alphaRMA:  "1.0 / float64(200)",
			alphaEMA:  "2.0 / float64(200+1)",
		},
		{
			name:      "Period 500 (extreme)",
			period:    500,
			warmupBar: 499,
			initBar:   499,
			loopCount: 500,
			alphaRMA:  "1.0 / float64(500)",
			alphaEMA:  "2.0 / float64(500+1)",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockAccessor := &MockAccessGenerator{
				loopAccessFn: func(loopVar string) string {
					return "src.Get(" + loopVar + ")"
				},
			}

			varName := fmt.Sprintf("test%d", tc.period)

			t.Run("RMA", func(t *testing.T) {
				builder := NewStatefulIndicatorBuilder("ta.rma", varName, P(tc.period), mockAccessor, false, NewTopLevelIndicatorContext())
				code := builder.BuildRMA()

				warmupCheck := fmt.Sprintf("if ctx.BarIndex < %d", tc.warmupBar)
				if !strings.Contains(code, warmupCheck) {
					t.Errorf("Missing warmup check: %s\nCode: %s", warmupCheck, code)
				}

				initCheck := fmt.Sprintf("if ctx.BarIndex == %d", tc.initBar)
				if !strings.Contains(code, initCheck) {
					t.Errorf("Missing initialization check: %s\nCode: %s", initCheck, code)
				}

				loopRange := fmt.Sprintf("for j := 0; j < %d; j++", tc.loopCount)
				if !strings.Contains(code, loopRange) {
					t.Errorf("Missing loop range: %s\nCode: %s", loopRange, code)
				}

				if !strings.Contains(code, tc.alphaRMA) {
					t.Errorf("Missing RMA alpha: %s\nCode: %s", tc.alphaRMA, code)
				}

				selfRef := fmt.Sprintf("%sSeries.Get(1)", varName)
				if !strings.Contains(code, selfRef) {
					t.Errorf("Missing self-reference: %s\nCode: %s", selfRef, code)
				}
			})

			t.Run("EMA", func(t *testing.T) {
				builder := NewStatefulIndicatorBuilder("ta.ema", varName, P(tc.period), mockAccessor, false, NewTopLevelIndicatorContext())
				code := builder.BuildEMA()

				if !strings.Contains(code, tc.alphaEMA) {
					t.Errorf("Missing EMA alpha: %s\nCode: %s", tc.alphaEMA, code)
				}

				if !strings.Contains(code, fmt.Sprintf("if ctx.BarIndex < %d", tc.warmupBar)) {
					t.Errorf("EMA missing warmup phase")
				}
				if !strings.Contains(code, fmt.Sprintf("if ctx.BarIndex == %d", tc.initBar)) {
					t.Errorf("EMA missing initialization phase")
				}
				if !strings.Contains(code, fmt.Sprintf("%sSeries.Get(1)", varName)) {
					t.Errorf("EMA missing self-reference")
				}
			})
		})
	}
}

/* TestStatefulIndicatorBuilder_NaNPropagation validates NaN handling behavior
 *
 * Tests that NaN checks are properly integrated when needsNaN is enabled:
 * - NaN detection during SMA initialization phase
 * - NaN propagation in recursive phase
 * - Early return on NaN source values
 * - NaN previous value handling
 *
 * Edge cases:
 * - All NaN input → NaN output
 * - Partial NaN input → NaN propagation
 * - NaN in middle of series → stops calculation
 */
func TestStatefulIndicatorBuilder_NaNPropagation(t *testing.T) {
	testCases := []struct {
		name          string
		needsNaN      bool
		shouldHave    []string
		shouldNotHave []string
	}{
		{
			name:     "NaN checks enabled",
			needsNaN: true,
			shouldHave: []string{
				"val := ",
				"if math.IsNaN(val)",
				"break", // Break loop on NaN (not return which exits function)
				"if math.IsNaN(currentSource)",
				"else if math.IsNaN(previousValue)",
			},
			shouldNotHave: []string{},
		},
		{
			name:     "NaN checks disabled",
			needsNaN: false,
			shouldHave: []string{
				"_sma_accumulator += ",
			},
			shouldNotHave: []string{
				"val := ",
				"if math.IsNaN(val)",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockAccessor := &MockAccessGenerator{
				loopAccessFn: func(loopVar string) string {
					return "data.Get(" + loopVar + ")"
				},
			}

			builder := NewStatefulIndicatorBuilder("ta.rma", "rma10", P(10), mockAccessor, tc.needsNaN, NewTopLevelIndicatorContext())
			code := builder.BuildRMA()

			for _, expected := range tc.shouldHave {
				if !strings.Contains(code, expected) {
					t.Errorf("Expected to find %q in code:\n%s", expected, code)
				}
			}

			for _, unexpected := range tc.shouldNotHave {
				if strings.Contains(code, unexpected) {
					t.Errorf("Should not find %q in code:\n%s", unexpected, code)
				}
			}
		})
	}
}

/* TestStatefulIndicatorBuilder_AlgorithmCorrectness validates algorithmic properties
 *
 * Tests that generated code follows correct stateful recursive algorithm:
 * - Three distinct phases (warmup, initialization, recursive)
 * - Forward loops only (no backward iteration)
 * - Self-reference for previous value (series.Get(1))
 * - No recalculation from scratch each bar
 * - Correct phase transitions
 *
 * This is algorithm validation, not bug-specific testing.
 */
func TestStatefulIndicatorBuilder_AlgorithmCorrectness(t *testing.T) {
	mockAccessor := &MockAccessGenerator{
		loopAccessFn: func(loopVar string) string {
			return "source.Get(" + loopVar + ")"
		},
	}

	t.Run("Three-phase structure", func(t *testing.T) {
		builder := NewStatefulIndicatorBuilder("ta.rma", "rma20", P(20), mockAccessor, false, NewTopLevelIndicatorContext())
		code := builder.BuildRMA()

		if !strings.Contains(code, "if ctx.BarIndex < 19") {
			t.Error("Missing warmup condition")
		}
		if !strings.Contains(code, "rma20Series.Set(math.NaN())") {
			t.Error("Missing NaN assignment in warmup")
		}

		if !strings.Contains(code, "/* First valid value: calculate SMA as initial state */") {
			t.Error("Missing initialization phase documentation")
		}
		if !strings.Contains(code, "if ctx.BarIndex == 19") {
			t.Error("Missing initialization condition")
		}
		if !strings.Contains(code, "_sma_accumulator := 0.0") {
			t.Error("Missing accumulator initialization")
		}

		if !strings.Contains(code, "/* Recursive phase: use previous indicator value */") {
			t.Error("Missing recursive phase documentation")
		}
		if !strings.Contains(code, "} else {") {
			t.Error("Missing else block for recursive phase")
		}
		if !strings.Contains(code, "previousValue := rma20Series.Get(1)") {
			t.Error("Missing previous value retrieval")
		}
	})

	t.Run("Forward loops only", func(t *testing.T) {
		builder := NewStatefulIndicatorBuilder("ta.rma", "rma30", P(30), mockAccessor, false, NewTopLevelIndicatorContext())
		code := builder.BuildRMA()

		if strings.Contains(code, "j--") {
			t.Error("Should not contain backward loop - violates stateful principle")
		}
		if strings.Contains(code, "j >= 0") {
			t.Error("Should not contain reverse iteration condition")
		}
		if !strings.Contains(code, "for j := 0; j <") {
			t.Error("Must use forward loop starting from 0")
		}
	})

	t.Run("Self-reference pattern", func(t *testing.T) {
		builder := NewStatefulIndicatorBuilder("ta.rma", "rma14", P(14), mockAccessor, false, NewTopLevelIndicatorContext())
		code := builder.BuildRMA()

		if !strings.Contains(code, "rma14Series.Get(1)") {
			t.Error("Missing self-reference to previous indicator value")
		}

		if !strings.Contains(code, "currentSource := ") {
			t.Error("Missing current source value extraction")
		}

		if !strings.Contains(code, "alpha*currentSource + (1-alpha)*previousValue") {
			t.Error("Missing correct recursive formula")
		}
	})

	t.Run("No recalculation from scratch", func(t *testing.T) {
		builder := NewStatefulIndicatorBuilder("ta.rma", "rma25", P(25), mockAccessor, false, NewTopLevelIndicatorContext())
		code := builder.BuildRMA()

		loopCount := strings.Count(code, "for j :=")
		if loopCount != 1 {
			t.Errorf("Should have exactly 1 loop (SMA initialization), found %d", loopCount)
		}

		firstElse := strings.Index(code, "} else {")
		if firstElse == -1 {
			t.Fatal("Missing first else block")
		}
		secondElse := strings.Index(code[firstElse+1:], "} else {")
		if secondElse != -1 {
			recursiveSection := code[firstElse+secondElse:]
			if strings.Contains(recursiveSection, "for j :=") {
				t.Error("Recursive phase should not contain loops - violates stateful principle")
			}
		}
	})
}

/* TestStatefulIndicatorBuilder_RMA_vs_EMA_Distinction validates alpha formula differences
 *
 * Tests that RMA and EMA use correct, distinct alpha formulas:
 * - RMA: alpha = 1 / period
 * - EMA: alpha = 2 / (period + 1)
 *
 * This validates the fundamental mathematical difference between indicators.
 */
func TestStatefulIndicatorBuilder_RMA_vs_EMA_Distinction(t *testing.T) {
	mockAccessor := &MockAccessGenerator{
		loopAccessFn: func(loopVar string) string {
			return "data.Get(" + loopVar + ")"
		},
	}

	testPeriods := []int{2, 10, 14, 20, 50, 100, 200}

	for _, period := range testPeriods {
		t.Run(fmt.Sprintf("Period %d", period), func(t *testing.T) {
			varName := fmt.Sprintf("test%d", period)

			rmaBuilder := NewStatefulIndicatorBuilder("ta.rma", varName, P(period), mockAccessor, false, NewTopLevelIndicatorContext())
			rmaCode := rmaBuilder.BuildRMA()

			emaBuilder := NewStatefulIndicatorBuilder("ta.ema", varName, P(period), mockAccessor, false, NewTopLevelIndicatorContext())
			emaCode := emaBuilder.BuildEMA()

			rmaAlpha := fmt.Sprintf("alpha := 1.0 / float64(%d)", period)
			if !strings.Contains(rmaCode, rmaAlpha) {
				t.Errorf("RMA missing correct alpha formula: %s\nCode: %s", rmaAlpha, rmaCode)
			}

			emaAlpha := fmt.Sprintf("alpha := 2.0 / float64(%d+1)", period)
			if !strings.Contains(emaCode, emaAlpha) {
				t.Errorf("EMA missing correct alpha formula: %s\nCode: %s", emaAlpha, emaCode)
			}

			wrongRmaAlpha := fmt.Sprintf("alpha := 2.0 / float64(%d+1)", period)
			if strings.Contains(rmaCode, wrongRmaAlpha) {
				t.Error("RMA should not use EMA alpha formula")
			}

			wrongEmaAlpha := fmt.Sprintf("alpha := 1.0 / float64(%d)", period)
			if strings.Contains(emaCode, wrongEmaAlpha) {
				t.Error("EMA should not use RMA alpha formula")
			}

			// Both should share same three-phase structure
			sharedStructure := []string{
				fmt.Sprintf("if ctx.BarIndex < %d", period-1),
				fmt.Sprintf("if ctx.BarIndex == %d", period-1),
				"} else {",
				".Get(1)", // Self-reference
				"newValue := alpha*currentSource + (1-alpha)*previousValue",
			}

			for _, pattern := range sharedStructure {
				if !strings.Contains(rmaCode, pattern) {
					t.Errorf("RMA missing shared pattern: %s", pattern)
				}
				if !strings.Contains(emaCode, pattern) {
					t.Errorf("EMA missing shared pattern: %s", pattern)
				}
			}
		})
	}
}

/* TestStatefulIndicatorBuilder_VariableNaming validates correct variable naming
 *
 * Tests that generated code uses consistent, collision-free variable names:
 * - Series variable naming
 * - Temporary variable naming (sum, alpha, previousValue, currentSource)
 * - Special character handling (underscores, numbers)
 */
func TestStatefulIndicatorBuilder_VariableNaming(t *testing.T) {
	testCases := []struct {
		name           string
		varName        string
		expectedSeries string
	}{
		{
			name:           "Simple name",
			varName:        "rma14",
			expectedSeries: "rma14Series",
		},
		{
			name:           "Name with underscore",
			varName:        "rma_14_close",
			expectedSeries: "rma_14_closeSeries",
		},
		{
			name:           "Name with number",
			varName:        "rma20v2",
			expectedSeries: "rma20v2Series",
		},
		{
			name:           "Short name",
			varName:        "r",
			expectedSeries: "rSeries",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockAccessor := &MockAccessGenerator{
				loopAccessFn: func(loopVar string) string {
					return "src.Get(" + loopVar + ")"
				},
			}

			builder := NewStatefulIndicatorBuilder("ta.rma", tc.varName, P(10), mockAccessor, false, NewTopLevelIndicatorContext())
			code := builder.BuildRMA()

			if !strings.Contains(code, tc.expectedSeries+".Set(math.NaN())") {
				t.Errorf("Missing series variable: %s\nCode: %s", tc.expectedSeries, code)
			}

			requiredVars := []string{
				"_sma_accumulator := 0.0",
				"alpha := ",
				"previousValue := ",
				"currentSource := ",
				"newValue := ",
			}

			for _, varDecl := range requiredVars {
				if !strings.Contains(code, varDecl) {
					t.Errorf("Missing variable declaration: %s\nCode: %s", varDecl, code)
				}
			}
		})
	}
}

/* TestStatefulIndicatorBuilder_CodeStructure validates generated code quality
 *
 * Tests structural properties of generated code:
 * - Proper indentation
 * - Comment placement
 * - No code duplication
 * - Logical flow
 */
func TestStatefulIndicatorBuilder_CodeStructure(t *testing.T) {
	mockAccessor := &MockAccessGenerator{
		loopAccessFn: func(loopVar string) string {
			return "data.Get(" + loopVar + ")"
		},
	}

	builder := NewStatefulIndicatorBuilder("ta.rma", "rma20", P(20), mockAccessor, true, NewTopLevelIndicatorContext())
	code := builder.BuildRMA()

	t.Run("Has documentation comments", func(t *testing.T) {
		expectedComments := []string{
			"/* Inline RMA(20) - Stateful recursive calculation */",
			"/* First valid value: calculate SMA as initial state */",
			"/* Recursive phase: use previous indicator value */",
		}

		for _, comment := range expectedComments {
			if !strings.Contains(code, comment) {
				t.Errorf("Missing documentation comment: %s", comment)
			}
		}
	})

	t.Run("Proper block structure", func(t *testing.T) {
		// Should have proper if-else structure
		if !strings.Contains(code, "if ctx.BarIndex") {
			t.Error("Missing if block")
		}
		if !strings.Contains(code, "} else {") {
			t.Error("Missing else block")
		}

		// Should have nested if for initialization
		initBlock := "if ctx.BarIndex == 19"
		if !strings.Contains(code, initBlock) {
			t.Error("Missing initialization if block")
		}
	})

	t.Run("No duplicate code patterns", func(t *testing.T) {
		// Alpha calculation should appear exactly once
		alphaCount := strings.Count(code, "alpha := ")
		if alphaCount != 1 {
			t.Errorf("Alpha calculation should appear once, found %d times", alphaCount)
		}

		// Formula should appear exactly once
		formulaCount := strings.Count(code, "alpha*currentSource + (1-alpha)*previousValue")
		if formulaCount != 1 {
			t.Errorf("Recursive formula should appear once, found %d times", formulaCount)
		}
	})

	t.Run("Logical phase ordering", func(t *testing.T) {
		headerPos := strings.Index(code, "/* Inline RMA(20)")
		initPos := strings.Index(code, "/* First valid value")
		recursivePos := strings.Index(code, "/* Recursive phase")

		if headerPos == -1 || initPos == -1 || recursivePos == -1 {
			t.Fatal("Missing phase comments")
		}

		if !(headerPos < initPos && initPos < recursivePos) {
			t.Error("Phases are not in correct order: header -> init -> recursive")
		}
	})
}
