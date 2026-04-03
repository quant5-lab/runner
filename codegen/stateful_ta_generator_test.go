package codegen

import (
	"strings"
	"testing"
)

func TestStatefulRMAGenerator_GeneratesForwardSeriesPattern(t *testing.T) {
	accessor := NewBuiltinIdentifierAccessor("closeSeries.GetCurrent()")
	context := NewTopLevelIndicatorContext()
	generator := NewStatefulRMAGenerator("rma14", 14, accessor, context)

	code := generator.GenerateRMA()

	if !strings.Contains(code, "/* Inline RMA(14)") {
		t.Error("missing RMA header comment")
	}

	if !strings.Contains(code, "if ctx.BarIndex < 13") {
		t.Error("missing warmup period check")
	}

	if !strings.Contains(code, "rma14Series.Set(math.NaN())") {
		t.Error("missing warmup NaN assignment")
	}

	if !strings.Contains(code, "if ctx.BarIndex == 13") {
		t.Error("missing initialization phase")
	}

	if !strings.Contains(code, "previousValue := rma14Series.Get(1)") {
		t.Error("missing forward reference to previous value")
	}

	if !strings.Contains(code, "alpha := 1.0 / float64(14)") {
		t.Error("missing RMA alpha calculation")
	}

	if !strings.Contains(code, "rma14Series.Set(newValue)") {
		t.Error("missing final series update")
	}

	if strings.Contains(code, "arrowCtx") {
		t.Error("should NOT use arrowCtx in main context")
	}

	if strings.Contains(code, "IIFE") || strings.Contains(code, "func()") {
		t.Error("should NOT generate IIFE wrapper for compile-time period")
	}
}

func TestStatefulEMAGenerator_GeneratesForwardSeriesPattern(t *testing.T) {
	accessor := NewBuiltinIdentifierAccessor("highSeries.GetCurrent()")
	context := NewTopLevelIndicatorContext()
	generator := NewStatefulEMAGenerator("ema20", 20, accessor, context)

	code := generator.GenerateEMA()

	if !strings.Contains(code, "/* Inline EMA(20)") {
		t.Error("missing EMA header comment")
	}

	if !strings.Contains(code, "if ctx.BarIndex < 19") {
		t.Error("missing warmup period check")
	}

	if !strings.Contains(code, "alpha := 2.0 / float64(20+1)") {
		t.Error("missing EMA alpha calculation")
	}

	if !strings.Contains(code, "ema20Series.Set(newValue)") {
		t.Error("missing final series update")
	}
}

func TestStatefulRMAGenerator_DirectSeriesAccess(t *testing.T) {
	accessor := NewBuiltinIdentifierAccessor("closeSeries.GetCurrent()")
	context := NewTopLevelIndicatorContext()
	generator := NewStatefulRMAGenerator("test", 10, accessor, context)

	code := generator.GenerateRMA()

	t.Logf("Generated code:\n%s", code)

	seriesSetCount := strings.Count(code, "testSeries.Set(")
	if seriesSetCount < 2 {
		t.Errorf("expected at least 2 direct Series.Set() calls, found %d", seriesSetCount)
	}

	seriesGetCount := strings.Count(code, "testSeries.Get(1)")
	if seriesGetCount < 1 {
		t.Error("expected at least 1 testSeries.Get(1) for previous value access")
	}
}
