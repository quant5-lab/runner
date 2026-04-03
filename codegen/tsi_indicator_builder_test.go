package codegen

import (
	"fmt"
	"strings"
	"testing"
)

func TestTSIIndicatorBuilder_TopLevelContext(t *testing.T) {
	ctx := NewTopLevelIndicatorContext()
	accessor := NewOHLCVFieldAccessGenerator("Close")
	short := NewConstantPeriod(5)
	long := NewConstantPeriod(13)

	builder := NewTSIIndicatorBuilder("myTSI", short, long, accessor, ctx)
	code := builder.Build()

	t.Run("momentum_step_generated", func(t *testing.T) {
		if !strings.Contains(code, "_myTSI_mom") {
			t.Errorf("Missing momentum series\nGot:\n%s", code)
		}
	})

	t.Run("absolute_momentum_step_generated", func(t *testing.T) {
		if !strings.Contains(code, "_myTSI_mom_abs") {
			t.Errorf("Missing absolute momentum series\nGot:\n%s", code)
		}
	})

	t.Run("first_ema_chain", func(t *testing.T) {
		if !strings.Contains(code, "_myTSI_ema1_mom") || !strings.Contains(code, "_myTSI_ema1_abs") {
			t.Errorf("Missing first EMA chain series\nGot:\n%s", code)
		}
	})

	t.Run("second_ema_chain", func(t *testing.T) {
		if !strings.Contains(code, "_myTSI_ema2_mom") || !strings.Contains(code, "_myTSI_ema2_abs") {
			t.Errorf("Missing second EMA chain series\nGot:\n%s", code)
		}
	})

	t.Run("nan_during_warmup", func(t *testing.T) {
		if !strings.Contains(code, "math.NaN()") {
			t.Errorf("Missing NaN assignment during warmup\nGot:\n%s", code)
		}
	})

	t.Run("hundred_multiplier", func(t *testing.T) {
		if !strings.Contains(code, "100.0") {
			t.Errorf("Missing 100.0 multiplier in TSI formula\nGot:\n%s", code)
		}
	})

	t.Run("zero_abs_division_guard", func(t *testing.T) {
		if !strings.Contains(code, "== 0.0") {
			t.Errorf("Missing zero absolute denominator guard\nGot:\n%s", code)
		}
	})

	t.Run("top_level_series_pattern", func(t *testing.T) {
		if !strings.Contains(code, "Series.Set") {
			t.Errorf("Missing Series.Set pattern for TopLevel context\nGot:\n%s", code)
		}
		if strings.Contains(code, "arrowCtx.GetOrCreateSeries") {
			t.Error("TopLevel context must not use arrow patterns")
		}
	})
}

func TestTSIIndicatorBuilder_ArrowContext(t *testing.T) {
	ctx := NewArrowFunctionIndicatorContext()
	accessor := NewInternalSeriesAccessor("source", ctx)
	shortVal, longVal := 5, 13
	short := NewConstantPeriod(shortVal)
	long := NewConstantPeriod(longVal)

	builder := NewTSIIndicatorBuilder("tsi", short, long, accessor, ctx)
	code := builder.Build()

	t.Run("arrow_context_series_access", func(t *testing.T) {
		if !strings.Contains(code, "arrowCtx") {
			t.Errorf("Arrow context must use arrowCtx references\nGot:\n%s", code)
		}
	})

	t.Run("warmup_guard_consistent_with_top_level", func(t *testing.T) {
		expected := fmt.Sprintf("ctx.BarIndex < %d", shortVal+longVal-1)
		if !strings.Contains(code, expected) {
			t.Errorf("Arrow context must emit warmup guard %q\nGot:\n%s", expected, code)
		}
	})
}

func TestTSIIndicatorBuilder_InternalSeriesNames(t *testing.T) {
	ctx := NewTopLevelIndicatorContext()
	accessor := NewOHLCVFieldAccessGenerator("Close")
	short := NewConstantPeriod(5)
	long := NewConstantPeriod(13)

	builder := NewTSIIndicatorBuilder("myTSI", short, long, accessor, ctx)
	builder.Build() /* must call Build to populate internal names */

	names := builder.GetInternalSeriesNames()

	if len(names) != 6 {
		t.Fatalf("Expected 6 internal series, got %d: %v", len(names), names)
	}

	expected := []string{
		"_myTSI_mom",
		"_myTSI_mom_abs",
		"_myTSI_ema1_mom",
		"_myTSI_ema1_abs",
		"_myTSI_ema2_mom",
		"_myTSI_ema2_abs",
	}

	for _, want := range expected {
		found := false
		for _, got := range names {
			if got == want {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected internal series %q not found in %v", want, names)
		}
	}
}

func TestTSIIndicatorBuilder_WarmupFormula(t *testing.T) {
	tests := []struct {
		name           string
		short          int
		long           int
		wantWarmupEdge int
	}{
		{"short1_long1", 1, 1, 1},
		{"short1_long13", 1, 13, 13},
		{"short5_long1", 5, 1, 5},
		{"short5_long13", 5, 13, 17},
		{"short13_long25", 13, 25, 37},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewTopLevelIndicatorContext()
			accessor := NewOHLCVFieldAccessGenerator("Close")
			builder := NewTSIIndicatorBuilder("r", NewConstantPeriod(tt.short), NewConstantPeriod(tt.long), accessor, ctx)
			code := builder.Build()

			expectedCheck := fmt.Sprintf("ctx.BarIndex < %d", tt.wantWarmupEdge)
			if !strings.Contains(code, expectedCheck) {
				t.Errorf("Expected warmup check %q in code\nGot:\n%s", expectedCheck, code)
			}
		})
	}
}

func TestTSIIndicatorBuilder_DifferentResultVarNames(t *testing.T) {
	ctx := NewTopLevelIndicatorContext()
	accessor := NewOHLCVFieldAccessGenerator("Close")

	b1 := NewTSIIndicatorBuilder("x", NewConstantPeriod(5), NewConstantPeriod(13), accessor, ctx)
	b2 := NewTSIIndicatorBuilder("y", NewConstantPeriod(5), NewConstantPeriod(13), accessor, ctx)

	code1 := b1.Build()
	code2 := b2.Build()

	/* Internal series names must be prefixed with result var to avoid collisions */
	if !strings.Contains(code1, "_x_mom") {
		t.Errorf("Builder for 'x' should use prefix _x_\nGot:\n%s", code1)
	}
	if !strings.Contains(code2, "_y_mom") {
		t.Errorf("Builder for 'y' should use prefix _y_\nGot:\n%s", code2)
	}
}
