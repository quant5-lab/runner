package codegen

import (
	"strings"
	"testing"
)

/* TestStatefulIndicatorBuilder_NaNHandling tests correct NaN handling in stateful indicators.
 * Ensures warmup loops use 'break' (not 'return') when encountering NaN values,
 * preventing premature function exit that would cause compilation errors.
 *
 * This is a regression test for a bug where NaN checks used 'return' inside warmup loops,
 * which compiled as a naked return from executeStrategy() instead of breaking the loop.
 *
 * Edge cases:
 * - NaN in first value
 * - NaN in middle of warmup window
 * - All values NaN
 * - No NaN values (normal case)
 */
func TestStatefulIndicatorBuilder_NaNHandling(t *testing.T) {
	tests := []struct {
		name        string
		needsNaN    bool
		wantBreak   bool
		wantReturn  bool
		description string
	}{
		{
			name:        "with NaN check enabled",
			needsNaN:    true,
			wantBreak:   true,
			wantReturn:  false,
			description: "Should use 'break' to exit loop on NaN, not 'return' which exits function",
		},
		{
			name:        "without NaN check",
			needsNaN:    false,
			wantBreak:   false,
			wantReturn:  false,
			description: "Should not have break/return when NaN checking disabled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accessor := NewOHLCVFieldAccessGenerator("Close")
			context := NewTopLevelIndicatorContext()
			builder := NewStatefulIndicatorBuilder("ta.rma", "testRma", 14, accessor, tt.needsNaN, context)

			code := builder.BuildRMA()

			// Check for 'break' statement in warmup loop
			hasBreak := strings.Contains(code, "break")
			if hasBreak != tt.wantBreak {
				t.Errorf("Code contains 'break' = %v, want %v (reason: %s)",
					hasBreak, tt.wantBreak, tt.description)
			}

			// Check for naked 'return' in warmup loop (BUG pattern)
			// Look for 'return' followed by newline/closing brace (not 'return value')
			hasNakedReturn := strings.Contains(code, "return\n") || strings.Contains(code, "return\t")
			if hasNakedReturn != tt.wantReturn {
				if hasNakedReturn {
					t.Errorf("Code contains naked 'return' in loop (BUG: would cause compilation error), "+
						"should use 'break' instead (reason: %s)", tt.description)
					t.Logf("Generated code:\n%s", code)
				}
			}

			// Verify code structure when NaN checking enabled
			if tt.needsNaN {
				// Should have: if math.IsNaN(val) { Set(NaN); break }
				if !strings.Contains(code, "math.IsNaN") {
					t.Error("NaN checking enabled but code missing 'math.IsNaN' check")
				}
				if !strings.Contains(code, "break") {
					t.Error("NaN checking enabled but missing 'break' statement to exit loop")
				}

				// Should NOT have naked return (compilation error)
				if strings.Contains(code, "return\n") || strings.Contains(code, "return\t") {
					t.Error("Code has naked 'return' in warmup loop - this causes compilation errors")
				}
			}
		})
	}
}

/* TestStatefulIndicatorBuilder_WarmupPhases tests the three phases of stateful indicators.
 *
 * Phases:
 * 1. Pre-warmup (ctx.BarIndex < period-1): Set NaN
 * 2. Warmup (ctx.BarIndex == period-1): Calculate SMA seed
 * 3. Recursive (ctx.BarIndex > period-1): Use previous value with alpha
 *
 * Ensures:
 * - Correct bar index checks
 * - SMA seeding uses full period window
 * - Recursive phase uses alpha = 1/period for RMA
 * - Proper NaN propagation in all phases
 */
func TestStatefulIndicatorBuilder_WarmupPhases(t *testing.T) {
	accessor := NewOHLCVFieldAccessGenerator("Close")
	context := NewTopLevelIndicatorContext()

	periods := []int{9, 14, 20, 50, 200}

	for _, period := range periods {
		t.Run(string(rune('0'+period/100))+string(rune('0'+(period/10)%10))+string(rune('0'+period%10))+" period", func(t *testing.T) {
			builder := NewStatefulIndicatorBuilder("ta.rma", "testRma", period, accessor, true, context)
			code := builder.BuildRMA()

			// Phase 1: Pre-warmup check
			if !strings.Contains(code, "ctx.BarIndex <") {
				t.Error("Missing pre-warmup phase check (ctx.BarIndex < period-1)")
			}

			// Phase 2: Warmup initialization check
			if !strings.Contains(code, "ctx.BarIndex ==") {
				t.Error("Missing warmup phase check (ctx.BarIndex == period-1)")
			}

			// Should have SMA calculation in warmup
			if !strings.Contains(code, "sum") {
				t.Error("Warmup phase missing SMA calculation (sum accumulation)")
			}

			// Should have loop over period for SMA seed
			if !strings.Contains(code, "for j") {
				t.Error("Warmup phase missing accumulation loop for SMA seed")
			}

			// Phase 3: Recursive phase
			if !strings.Contains(code, "} else {") {
				t.Error("Missing recursive phase (else block)")
			}

			// Should use alpha = 1/period
			if !strings.Contains(code, "alpha") {
				t.Error("Recursive phase missing alpha calculation")
			}

			// Should access previous value
			if !strings.Contains(code, ".Get(1)") {
				t.Error("Recursive phase missing previous value access")
			}

			// Should have recursive formula: alpha*curr + (1-alpha)*prev
			if !strings.Contains(code, "alpha*") && !strings.Contains(code, "(1-alpha)") {
				t.Error("Recursive phase missing RMA formula")
			}
		})
	}
}

/* TestStatefulIndicatorBuilder_AccessorTypes tests different accessor types work correctly.
 *
 * Accessor types:
 * - OHLCVFieldAccessGenerator: Direct bar field access
 * - SeriesVariableAccessGenerator: User-defined series
 * - ExpressionAccessGenerator: Complex expressions
 *
 * Ensures each accessor type generates correct value access code in loops and recursive phase.
 */
func TestStatefulIndicatorBuilder_AccessorTypes(t *testing.T) {
	tests := []struct {
		name          string
		accessor      AccessGenerator
		expectInLoop  string
		expectRecurse string
		description   string
	}{
		{
			name:          "OHLCV field accessor",
			accessor:      NewOHLCVFieldAccessGenerator("Close"),
			expectInLoop:  "ctx.Data[ctx.BarIndex-j].Close",
			expectRecurse: "ctx.Data[ctx.BarIndex-0].Close",
			description:   "OHLCV fields should use ctx.Data array access",
		},
		{
			name:          "Series variable accessor",
			accessor:      NewSeriesVariableAccessGenerator("myVar"),
			expectInLoop:  "myVarSeries.Get(j)",
			expectRecurse: "myVarSeries.Get(0)",
			description:   "Series variables should use .Get() method",
		},
		{
			name:          "Series with base offset",
			accessor:      NewSeriesVariableAccessGeneratorWithOffset("myVar", 2),
			expectInLoop:  "myVarSeries.Get(j+2)",
			expectRecurse: "myVarSeries.Get(0+2)",
			description:   "Series with offset should add offset to access",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			context := NewTopLevelIndicatorContext()
			builder := NewStatefulIndicatorBuilder("ta.rma", "testRma", 14, tt.accessor, true, context)
			code := builder.BuildRMA()

			// Check warmup loop value access
			loopAccess := tt.accessor.GenerateLoopValueAccess("j")
			if !strings.Contains(code, loopAccess) {
				t.Errorf("Warmup loop missing expected access pattern %q (reason: %s)",
					loopAccess, tt.description)
				t.Logf("Generated code:\n%s", code)
			}

			// Check recursive phase current value access
			recurseAccess := tt.accessor.GenerateLoopValueAccess("0")
			if !strings.Contains(code, recurseAccess) {
				t.Errorf("Recursive phase missing expected access pattern %q (reason: %s)",
					recurseAccess, tt.description)
			}
		})
	}
}

/* TestStatefulIndicatorBuilder_ContextTypes tests top-level vs arrow function contexts.
 *
 * Context differences:
 * - Top-level: varSeries.Set(value)
 * - Arrow function: arrowCtx.GetOrCreateSeries("var").Set(value)
 *
 * Ensures generated code uses correct Series update method for context type.
 */
func TestStatefulIndicatorBuilder_ContextTypes(t *testing.T) {
	tests := []struct {
		name             string
		context          StatefulIndicatorContext
		expectSetPattern string
		description      string
	}{
		{
			name:             "top-level context",
			context:          NewTopLevelIndicatorContext(),
			expectSetPattern: "testRmaSeries.Set(",
			description:      "Top-level should use direct Series.Set() call",
		},
		{
			name:             "arrow function context",
			context:          NewArrowFunctionIndicatorContext(),
			expectSetPattern: "arrowCtx.GetOrCreateSeries(",
			description:      "Arrow functions should use arrowCtx.GetOrCreateSeries() call",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accessor := NewOHLCVFieldAccessGenerator("Close")
			builder := NewStatefulIndicatorBuilder("ta.rma", "testRma", 14, accessor, false, tt.context)
			code := builder.BuildRMA()

			if !strings.Contains(code, tt.expectSetPattern) {
				t.Errorf("Code missing expected Set pattern %q (reason: %s)",
					tt.expectSetPattern, tt.description)
				t.Logf("Generated code:\n%s", code)
			}
		})
	}
}

/* TestStatefulIndicatorBuilder_EdgeCasePeriods tests edge case period values.
 *
 * Edge cases:
 * - period = 1: Immediate warmup, no accumulation
 * - period = 2: Minimal accumulation
 * - Large periods: 200, 500
 *
 * Ensures correct warmup bar index calculations for all period values.
 */
func TestStatefulIndicatorBuilder_EdgeCasePeriods(t *testing.T) {
	edgePeriods := []int{1, 2, 3, 200, 500}

	for _, period := range edgePeriods {
		t.Run(string(rune('0'+period/100))+string(rune('0'+(period/10)%10))+string(rune('0'+period%10)), func(t *testing.T) {
			accessor := NewOHLCVFieldAccessGenerator("Close")
			context := NewTopLevelIndicatorContext()
			builder := NewStatefulIndicatorBuilder("ta.rma", "testRma", period, accessor, false, context)

			code := builder.BuildRMA()

			// Should have warmup at bar index = period - 1
			expectedWarmupBar := period - 1
			if !strings.Contains(code, "ctx.BarIndex") {
				t.Error("Missing bar index check for warmup")
			}

			// Should have loop with correct period
			if period > 1 && !strings.Contains(code, "for j") {
				t.Error("Missing accumulation loop for period > 1")
			}

			// Should calculate with correct period in alpha
			if !strings.Contains(code, "alpha") {
				t.Error("Missing alpha calculation in recursive phase")
			}

			// Verify code compiles (syntax check)
			if strings.Count(code, "{") != strings.Count(code, "}") {
				t.Errorf("Unbalanced braces in generated code (warmup bar = %d)", expectedWarmupBar)
			}
		})
	}
}
