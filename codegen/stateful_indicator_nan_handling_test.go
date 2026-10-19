package codegen

import (
	"fmt"
	"strings"
	"testing"
)

// TestStatefulIndicatorBuilder_NaNHandling verifies the loop-exit contract for the
// NaN-checked initialization phase of stateful indicators.
//
// When needsNaN=true the initialization loop must exit with 'break' on a NaN value,
// not with 'return', because 'return' inside a generated closure exits the enclosing
// executeStrategy function instead of breaking the loop — a compile-time structural
// error in the generated Go code.
//
// When needsNaN=false no NaN infrastructure is emitted at all.
func TestStatefulIndicatorBuilder_NaNHandling(t *testing.T) {
	tests := []struct {
		name       string
		needsNaN   bool
		wantBreak  bool
		wantReturn bool
	}{
		{
			name:       "with NaN check enabled",
			needsNaN:   true,
			wantBreak:  true,
			wantReturn: false,
		},
		{
			name:       "without NaN check",
			needsNaN:   false,
			wantBreak:  false,
			wantReturn: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accessor := NewOHLCVFieldAccessGenerator("Close")
			context := NewTopLevelIndicatorContext()
			builder := NewStatefulIndicatorBuilder("ta.rma", "testRma", P(14), accessor, tt.needsNaN, context)

			code := builder.BuildRMA()

			hasBreak := strings.Contains(code, "break")
			if hasBreak != tt.wantBreak {
				t.Errorf("Code contains 'break' = %v, want %v", hasBreak, tt.wantBreak)
			}

			// Detect naked 'return' (no value) — 'return value' is legitimate in generated code
			hasNakedReturn := strings.Contains(code, "return\n") || strings.Contains(code, "return\t")
			if hasNakedReturn != tt.wantReturn {
				if hasNakedReturn {
					t.Errorf("Code contains naked 'return' in loop (causes compilation error in generated Go — use 'break' instead)")
					t.Logf("Generated code:\n%s", code)
				}
			}

			if tt.needsNaN {
				if !strings.Contains(code, "math.IsNaN") {
					t.Error("NaN checking enabled but code missing 'math.IsNaN' check")
				}
				if !strings.Contains(code, "break") {
					t.Error("NaN checking enabled but missing 'break' statement to exit loop")
				}
				if strings.Contains(code, "return\n") || strings.Contains(code, "return\t") {
					t.Error("Code has naked 'return' in warmup loop - this causes compilation errors")
				}
				if !strings.Contains(code, "_sma_has_nan := false") {
					t.Error("NaN checking must declare _sma_has_nan flag before loop")
				}
				if !strings.Contains(code, "_sma_has_nan = true") {
					t.Error("NaN checking must set flag in loop body on NaN value")
				}
				if !strings.Contains(code, "if _sma_has_nan {") {
					t.Error("NaN checking must check deferred flag after loop to emit NaN series update")
				}
			}
		})
	}
}

func TestStatefulIndicatorBuilder_WarmupPhases(t *testing.T) {
	accessor := NewOHLCVFieldAccessGenerator("Close")
	context := NewTopLevelIndicatorContext()

	for _, period := range []int{9, 14, 20, 50, 200} {
		t.Run(fmt.Sprintf("period_%d", period), func(t *testing.T) {
			builder := NewStatefulIndicatorBuilder("ta.rma", "testRma", P(period), accessor, true, context)
			code := builder.BuildRMA()

			if !strings.Contains(code, "ctx.BarIndex <") {
				t.Error("Missing pre-warmup phase check (ctx.BarIndex < period-1)")
			}
			if !strings.Contains(code, "ctx.BarIndex ==") {
				t.Error("Missing warmup phase check (ctx.BarIndex == period-1)")
			}
			if !strings.Contains(code, "_sma_accumulator") {
				t.Error("Warmup phase missing SMA calculation (_sma_accumulator)")
			}
			if !strings.Contains(code, "for j") {
				t.Error("Warmup phase missing accumulation loop for SMA seed")
			}
			if !strings.Contains(code, "} else {") {
				t.Error("Missing recursive phase (else block)")
			}
			if !strings.Contains(code, "alpha") {
				t.Error("Recursive phase missing alpha calculation")
			}
			if !strings.Contains(code, ".Get(1)") {
				t.Error("Recursive phase missing previous value access")
			}
			if !strings.Contains(code, "alpha*") && !strings.Contains(code, "(1-alpha)") {
				t.Error("Recursive phase missing RMA formula")
			}
		})
	}
}

func TestStatefulIndicatorBuilder_AccessorTypes(t *testing.T) {
	tests := []struct {
		name          string
		accessor      AccessGenerator
		expectInLoop  string
		expectRecurse string
	}{
		{
			name:          "OHLCV field accessor",
			accessor:      NewOHLCVFieldAccessGenerator("Close"),
			expectInLoop:  "closeSeries.Get(j)",
			expectRecurse: "closeSeries.Get(0)",
		},
		{
			name:          "Series variable accessor",
			accessor:      NewSeriesVariableAccessGenerator("myVar"),
			expectInLoop:  "myVarSeries.Get(j)",
			expectRecurse: "myVarSeries.Get(0)",
		},
		{
			name:          "Series with base offset",
			accessor:      NewSeriesVariableAccessGeneratorWithOffset("myVar", 2),
			expectInLoop:  "myVarSeries.Get(j+2)",
			expectRecurse: "myVarSeries.Get(0+2)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			context := NewTopLevelIndicatorContext()
			builder := NewStatefulIndicatorBuilder("ta.rma", "testRma", P(14), tt.accessor, true, context)
			code := builder.BuildRMA()

			loopAccess := tt.accessor.GenerateLoopValueAccess("j")
			if !strings.Contains(code, loopAccess) {
				t.Errorf("Warmup loop missing expected access pattern %q\nGenerated code:\n%s", loopAccess, code)
			}

			recurseAccess := tt.accessor.GenerateLoopValueAccess("0")
			if !strings.Contains(code, recurseAccess) {
				t.Errorf("Recursive phase missing expected access pattern %q", recurseAccess)
			}
		})
	}
}

func TestStatefulIndicatorBuilder_ContextTypes(t *testing.T) {
	tests := []struct {
		name             string
		context          StatefulIndicatorContext
		expectSetPattern string
	}{
		{
			name:             "top-level context",
			context:          NewTopLevelIndicatorContext(),
			expectSetPattern: "testRmaSeries.Set(",
		},
		{
			name:             "arrow function context",
			context:          NewArrowFunctionIndicatorContext(),
			expectSetPattern: "arrowCtx.GetOrCreateSeries(",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accessor := NewOHLCVFieldAccessGenerator("Close")
			builder := NewStatefulIndicatorBuilder("ta.rma", "testRma", P(14), accessor, false, tt.context)
			code := builder.BuildRMA()

			if !strings.Contains(code, tt.expectSetPattern) {
				t.Errorf("Code missing expected Set pattern %q\nGenerated code:\n%s", tt.expectSetPattern, code)
			}
		})
	}
}

// TestStatefulIndicatorBuilder_NaNSeedFlagAbsent verifies that when needsNaN is false,
// the initialization loop accumulates directly without NaN flag infrastructure.
func TestStatefulIndicatorBuilder_NaNSeedFlagAbsent(t *testing.T) {
	accessor := NewOHLCVFieldAccessGenerator("Close")
	context := NewTopLevelIndicatorContext()
	builder := NewStatefulIndicatorBuilder("ta.rma", "testRma", P(14), accessor, false, context)
	code := builder.BuildRMA()

	if !strings.Contains(code, "_sma_accumulator +=") {
		t.Errorf("Expected direct accumulation pattern\nCode:\n%s", code)
	}
	for _, absent := range []string{"_sma_has_nan", "val :="} {
		if strings.Contains(code, absent) {
			t.Errorf("Unexpected %q in generated code (needsNaN=false)\nCode:\n%s", absent, code)
		}
	}
}

func TestStatefulIndicatorBuilder_EdgeCasePeriods(t *testing.T) {
	for _, period := range []int{1, 2, 3, 200, 500} {
		t.Run(fmt.Sprintf("period_%d", period), func(t *testing.T) {
			accessor := NewOHLCVFieldAccessGenerator("Close")
			context := NewTopLevelIndicatorContext()
			builder := NewStatefulIndicatorBuilder("ta.rma", "testRma", P(period), accessor, false, context)
			code := builder.BuildRMA()

			if !strings.Contains(code, "ctx.BarIndex") {
				t.Error("Missing bar index check for warmup")
			}
			if period > 1 && !strings.Contains(code, "for j") {
				t.Error("Missing accumulation loop for period > 1")
			}
			if !strings.Contains(code, "alpha") {
				t.Error("Missing alpha calculation in recursive phase")
			}
			if strings.Count(code, "{") != strings.Count(code, "}") {
				t.Errorf("Unbalanced braces in generated code (period=%d)", period)
			}
		})
	}
}
