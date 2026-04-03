package codegen

import (
	"strings"
	"testing"
)

func TestMFIIndicatorBuilder_Build_ConstantPeriod(t *testing.T) {
	context := NewTopLevelIndicatorContext()
	accessor := NewOHLCVFieldAccessGenerator("Close")
	builder := NewMFIIndicatorBuilder("mfi14", NewConstantPeriod(14), accessor, context)

	code := builder.Build()

	expectedFragments := []string{
		"_mfi14_change",
		"bar.Volume",
		"_mfi14_positive_mfSeries.Set(",
		"_mfi14_negative_mfSeries.Set(",
		"posSum",
		"negSum",
		"mfr := posSum / negSum",
		"100.0 - (100.0 / (1.0 + mfr))",
		"mfi14Series.Set(",
		"ctx.BarIndex < 14",
		"math.NaN()",
		"negSum == 0",
	}

	for _, frag := range expectedFragments {
		if !strings.Contains(code, frag) {
			t.Errorf("missing expected fragment %q in generated code:\n%s", frag, code)
		}
	}
}

func TestMFIIndicatorBuilder_GetInternalSeriesNames(t *testing.T) {
	context := NewTopLevelIndicatorContext()
	accessor := NewOHLCVFieldAccessGenerator("Close")
	builder := NewMFIIndicatorBuilder("mfi14", NewConstantPeriod(14), accessor, context)

	/* Build must be called to populate internal series names */
	builder.Build()

	names := builder.GetInternalSeriesNames()
	if len(names) != 2 {
		t.Fatalf("expected 2 internal series, got %d", len(names))
	}
	if names[0] != "_mfi14_positive_mf" {
		t.Errorf("expected _mfi14_positive_mf, got %q", names[0])
	}
	if names[1] != "_mfi14_negative_mf" {
		t.Errorf("expected _mfi14_negative_mf, got %q", names[1])
	}
}

func TestMFIIndicatorBuilder_Build_LookbackLoopStructure(t *testing.T) {
	context := NewTopLevelIndicatorContext()
	accessor := NewOHLCVFieldAccessGenerator("Close")
	builder := NewMFIIndicatorBuilder("mfi14", NewConstantPeriod(14), accessor, context)

	code := builder.Build()

	/* Verify lookback loop sums both positive and negative money flows */
	if !strings.Contains(code, "for j := 0; j < 14; j++") {
		t.Errorf("expected lookback loop with period 14 in generated code:\n%s", code)
	}

	posSumCount := strings.Count(code, "posSum +=")
	negSumCount := strings.Count(code, "negSum +=")
	if posSumCount != 1 {
		t.Errorf("expected 1 posSum accumulation, got %d", posSumCount)
	}
	if negSumCount != 1 {
		t.Errorf("expected 1 negSum accumulation, got %d", negSumCount)
	}
}
