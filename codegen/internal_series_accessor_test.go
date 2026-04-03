package codegen

import (
	"strings"
	"testing"
)

/* TestInternalSeriesAccessor_ContextAwareAccess validates series access patterns across execution contexts */
func TestInternalSeriesAccessor_ContextAwareAccess(t *testing.T) {
	tests := []struct {
		name         string
		context      StatefulIndicatorContext
		seriesName   string
		loopVar      string
		offset       int
		expectedLoop string
		expectedInit string
		expectedCurr string
	}{
		{
			name:         "TopLevel simple loop var",
			context:      NewTopLevelIndicatorContext(),
			seriesName:   "_internal",
			loopVar:      "j",
			offset:       14,
			expectedLoop: "_internalSeries.Get(j)",
			expectedInit: "_internalSeries.Get(13)",
			expectedCurr: "_internalSeries.Get(0)",
		},
		{
			name:         "Arrow simple loop var",
			context:      NewArrowFunctionIndicatorContext(),
			seriesName:   "_internal",
			loopVar:      "j",
			offset:       14,
			expectedLoop: "arrowCtx.GetOrCreateSeries(\"_internal\").Get(j)",
			expectedInit: "arrowCtx.GetOrCreateSeries(\"_internal\").Get(13)",
			expectedCurr: "arrowCtx.GetOrCreateSeries(\"_internal\").Get(0)",
		},
		{
			name:         "TopLevel expression passthrough",
			context:      NewTopLevelIndicatorContext(),
			seriesName:   "_temp",
			loopVar:      "i-1",
			offset:       10,
			expectedLoop: "_tempSeries.Get(i-1)",
			expectedInit: "_tempSeries.Get(9)",
			expectedCurr: "_tempSeries.Get(0)",
		},
		{
			name:         "Arrow expression passthrough",
			context:      NewArrowFunctionIndicatorContext(),
			seriesName:   "_temp",
			loopVar:      "idx+offset",
			offset:       5,
			expectedLoop: "arrowCtx.GetOrCreateSeries(\"_temp\").Get(idx+offset)",
			expectedInit: "arrowCtx.GetOrCreateSeries(\"_temp\").Get(4)",
			expectedCurr: "arrowCtx.GetOrCreateSeries(\"_temp\").Get(0)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accessor := NewInternalSeriesAccessor(tt.seriesName, tt.context)

			loopCode := accessor.GenerateLoopValueAccess(tt.loopVar)
			if loopCode != tt.expectedLoop {
				t.Errorf("GenerateLoopValueAccess(%q) = %q, want %q", tt.loopVar, loopCode, tt.expectedLoop)
			}

			initCode := accessor.GenerateInitialValueAccess(tt.offset)
			if initCode != tt.expectedInit {
				t.Errorf("GenerateInitialValueAccess(%d) = %q, want %q", tt.offset, initCode, tt.expectedInit)
			}

			currCode := accessor.GenerateCurrentValueAccess()
			if currCode != tt.expectedCurr {
				t.Errorf("GenerateCurrentValueAccess() = %q, want %q", currCode, tt.expectedCurr)
			}
		})
	}
}

/* TestInternalSeriesAccessor_OffsetArithmetic validates offset-to-index conversion behavior */
func TestInternalSeriesAccessor_OffsetArithmetic(t *testing.T) {
	tests := []struct {
		name       string
		offset     int
		wantIndex  string
		wantOffset int
	}{
		{"zero offset yields zero index", 0, "Get(-1)", 0},
		{"offset 1 yields index 0", 1, "Get(0)", 0},
		{"offset 14 yields index 13", 14, "Get(13)", 0},
		{"offset 100 yields index 99", 100, "Get(99)", 0},
	}

	context := NewTopLevelIndicatorContext()
	accessor := NewInternalSeriesAccessor("_test", context)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := accessor.GenerateInitialValueAccess(tt.offset)
			if !strings.Contains(code, tt.wantIndex) {
				t.Errorf("GenerateInitialValueAccess(%d) = %q, want to contain %q", tt.offset, code, tt.wantIndex)
			}

			baseOffset := accessor.GetBaseOffset()
			if baseOffset != tt.wantOffset {
				t.Errorf("GetBaseOffset() = %d, want %d", baseOffset, tt.wantOffset)
			}
		})
	}
}

/* TestInternalSeriesAccessor_ExpressionPassthrough validates loop variable expression handling */
func TestInternalSeriesAccessor_ExpressionPassthrough(t *testing.T) {
	tests := []struct {
		name        string
		loopVar     string
		mustContain string
	}{
		{"simple identifier", "j", "Get(j)"},
		{"arithmetic expression", "j+1", "Get(j+1)"},
		{"complex expression", "i*2-offset", "Get(i*2-offset)"},
	}

	context := NewTopLevelIndicatorContext()
	accessor := NewInternalSeriesAccessor("_calc", context)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := accessor.GenerateLoopValueAccess(tt.loopVar)

			if !strings.Contains(code, tt.mustContain) {
				t.Errorf("GenerateLoopValueAccess(%q) = %q, must contain %q", tt.loopVar, code, tt.mustContain)
			}
		})
	}
}

/* TestInternalSeriesAccessor_InterfaceCompliance ensures AccessGenerator contract */
func TestInternalSeriesAccessor_InterfaceCompliance(t *testing.T) {
	var _ AccessGenerator = (*InternalSeriesAccessor)(nil)
}
