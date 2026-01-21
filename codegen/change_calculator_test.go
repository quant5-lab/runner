package codegen

import (
	"strings"
	"testing"
)

/* TestChangeCalculator_SourceTypes validates change calculation across different data sources */
func TestChangeCalculator_SourceTypes(t *testing.T) {
	tests := []struct {
		name               string
		accessor           AccessGenerator
		varName            string
		mustContainCurrent string
		mustContainPrev    string
	}{
		{
			name:               "OHLCV field source",
			accessor:           NewOHLCVFieldAccessGenerator("Close"),
			varName:            "priceChange",
			mustContainCurrent: "ctx.Data[ctx.BarIndex].Close",
			mustContainPrev:    "ctx.Data[ctx.BarIndex-1].Close",
		},
		{
			name:               "Series variable source",
			accessor:           NewSeriesVariableAccessGenerator("myIndicator"),
			varName:            "indicatorChange",
			mustContainCurrent: "myIndicatorSeries.GetCurrent()",
			mustContainPrev:    "myIndicatorSeries.Get(1)",
		},
		{
			name:               "Internal series TopLevel",
			accessor:           NewInternalSeriesAccessor("_intermediate", NewTopLevelIndicatorContext()),
			varName:            "deltaValue",
			mustContainCurrent: "_intermediateSeries.Get(0)",
			mustContainPrev:    "_intermediateSeries.Get(1)",
		},
		{
			name:               "Internal series Arrow",
			accessor:           NewInternalSeriesAccessor("_gain", NewArrowFunctionIndicatorContext()),
			varName:            "change",
			mustContainCurrent: "arrowCtx.GetOrCreateSeries(\"_gain\").Get(0)",
			mustContainPrev:    "arrowCtx.GetOrCreateSeries(\"_gain\").Get(1)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calc := NewChangeCalculator(tt.accessor)
			code := calc.GenerateChangeCode(tt.varName)

			/* Validate change variable declaration (var declaration, not := assignment) */
			expectedDecl := "var " + tt.varName + " float64"
			if !strings.Contains(code, expectedDecl) {
				t.Errorf("Missing change variable declaration %q in:\n%s", expectedDecl, code)
			}

			/* Validate current bar access */
			if !strings.Contains(code, tt.mustContainCurrent) {
				t.Errorf("Missing current bar access %q in:\n%s", tt.mustContainCurrent, code)
			}

			/* Validate previous bar access (NO FUTURE PEEK) */
			if !strings.Contains(code, tt.mustContainPrev) {
				t.Errorf("Missing previous bar access %q in:\n%s", tt.mustContainPrev, code)
			}
		})
	}
}

/* TestChangeCalculator_WarmupBehavior validates first bar handling (no previous value) */
func TestChangeCalculator_WarmupBehavior(t *testing.T) {
	tests := []struct {
		name             string
		accessor         AccessGenerator
		mustContainGuard string
		mustContainNaN   bool
	}{
		{
			name:             "OHLCV warmup",
			accessor:         NewOHLCVFieldAccessGenerator("High"),
			mustContainGuard: "if ctx.BarIndex < 1",
			mustContainNaN:   true,
		},
		{
			name:             "Series variable warmup",
			accessor:         NewSeriesVariableAccessGenerator("volume"),
			mustContainGuard: "if ctx.BarIndex < 1",
			mustContainNaN:   true,
		},
		{
			name:             "Internal series warmup",
			accessor:         NewInternalSeriesAccessor("_temp", NewTopLevelIndicatorContext()),
			mustContainGuard: "if ctx.BarIndex < 1",
			mustContainNaN:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calc := NewChangeCalculator(tt.accessor)
			code := calc.GenerateChangeCode("testChange")

			/* Validate warmup guard exists */
			if !strings.Contains(code, tt.mustContainGuard) {
				t.Errorf("Missing warmup guard %q in:\n%s", tt.mustContainGuard, code)
			}

			/* Validate NaN assignment during warmup */
			if tt.mustContainNaN && !strings.Contains(code, "math.NaN()") {
				t.Errorf("Missing NaN handling during warmup in:\n%s", code)
			}
		})
	}
}

/* TestChangeCalculator_DebugInstrumentation validates code readability through structure */
func TestChangeCalculator_DebugInstrumentation(t *testing.T) {
	accessor := NewOHLCVFieldAccessGenerator("Close")
	calc := NewChangeCalculator(accessor)
	code := calc.GenerateChangeCode("change")

	/* Code self-documents through variable naming and structure */
	if !strings.Contains(code, "var change float64") {
		t.Error("Missing self-explanatory variable declaration")
	}
	if !strings.Contains(code, "ctx.BarIndex < 1") {
		t.Error("Missing warmup guard (code should speak for itself)")
	}
}

/* TestChangeCalculator_CodeStructure validates algorithmic properties */
func TestChangeCalculator_CodeStructure(t *testing.T) {
	accessor := NewOHLCVFieldAccessGenerator("Close")
	calc := NewChangeCalculator(accessor)
	code := calc.GenerateChangeCode("delta")

	/* Algorithmic property: Change must be current minus previous */
	if !strings.Contains(code, "-") {
		t.Error("Change calculation missing subtraction operator")
	}

	/* Algorithmic property: Must handle warmup period */
	if !strings.Contains(code, "ctx.BarIndex") {
		t.Error("Change calculation missing bar index check")
	}

	/* Algorithmic property: Two branches (warmup + calculation) */
	ifCount := strings.Count(code, "if ")
	if ifCount < 1 {
		t.Error("Change calculation missing conditional branching")
	}
}
