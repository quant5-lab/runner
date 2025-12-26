package codegen

import (
	"fmt"
	"testing"
)

// TestGeneratedRMACode_VisualInspection prints the generated RMA code
// for manual verification against Pine Script semantics
func TestGeneratedRMACode_VisualInspection(t *testing.T) {
	mockAccessor := &MockAccessGenerator{
		loopAccessFn: func(loopVar string) string {
			return fmt.Sprintf("closeSeries.Get(%s)", loopVar)
		},
		initialAccessFn: func(period int) string {
			return fmt.Sprintf("closeSeries.Get(%d)", period-1)
		},
	}

	builder := NewStatefulIndicatorBuilder("ta.rma", "rma14", 14, mockAccessor, false, NewTopLevelIndicatorContext())
	code := builder.BuildRMA()

	t.Log("Generated RMA(14) code:")
	t.Log("========================")
	t.Log(code)
	t.Log("========================")

	expectedFlow := `
Expected execution flow:
1. Bar 0-12: Set rma14Series to NaN (warmup)
2. Bar 13: Calculate SMA(close, 14) as initial RMA value
3. Bar 14+: Apply formula: rma[i] = (1/14)*close[i] + (13/14)*rma[i-1]
`
	t.Log(expectedFlow)
}

// TestGeneratedRMACode_WithNaN shows NaN-safe version
func TestGeneratedRMACode_WithNaN(t *testing.T) {
	mockAccessor := &MockAccessGenerator{
		loopAccessFn: func(loopVar string) string {
			return fmt.Sprintf("sourceSeries.Get(%s)", loopVar)
		},
	}

	builder := NewStatefulIndicatorBuilder("ta.rma", "rma20", 20, mockAccessor, true, NewTopLevelIndicatorContext())
	code := builder.BuildRMA()

	t.Log("Generated RMA(20) with NaN checking:")
	t.Log("====================================")
	t.Log(code)
	t.Log("====================================")
}
