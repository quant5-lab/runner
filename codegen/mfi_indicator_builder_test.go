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

func TestMFIIndicatorBuilder_RawMFAssignedInExactlyTwoBranches(t *testing.T) {
	sources := []struct {
		name     string
		accessor AccessGenerator
	}{
		{"close", NewOHLCVFieldAccessGenerator("Close")},
		{"hlc3", NewOHLCVFieldAccessGenerator("Close")},
		{"true_range", NewTrueRangeAccessGenerator()},
	}

	for _, src := range sources {
		t.Run(src.name, func(t *testing.T) {
			ctx := NewTopLevelIndicatorContext()
			builder := NewMFIIndicatorBuilder("mfi", NewConstantPeriod(14), src.accessor, ctx)
			code := builder.Build()

			if got := strings.Count(code, "= rawMF"); got != 2 {
				t.Errorf("rawMF must be assigned exactly 2 times (positive + negative branch), got %d", got)
			}
		})
	}
}

func TestMFIIndicatorBuilder_SplitBranchAssignsBothFlowVariables(t *testing.T) {
	ctx := NewTopLevelIndicatorContext()
	builder := NewMFIIndicatorBuilder("mfi", NewConstantPeriod(14), NewOHLCVFieldAccessGenerator("Close"), ctx)
	code := builder.Build()

	if got := strings.Count(code, "posMF = 0.0"); got < 3 {
		t.Errorf("posMF must be explicitly zeroed in at least 3 branches (NaN + negative + zero-change), got %d", got)
	}

	if got := strings.Count(code, "negMF = 0.0"); got < 3 {
		t.Errorf("negMF must be explicitly zeroed in at least 3 branches (NaN + positive + zero-change), got %d", got)
	}
}

func TestMFIIndicatorBuilder_DirectionalSplitFourBranchChain(t *testing.T) {
	ctx := NewTopLevelIndicatorContext()
	builder := NewMFIIndicatorBuilder("mfi", NewConstantPeriod(14), NewOHLCVFieldAccessGenerator("Close"), ctx)
	code := builder.Build()

	hasNaN := strings.Contains(code, "math.IsNaN")
	hasPositive := strings.Contains(code, "> 0")
	hasNegative := strings.Contains(code, "< 0")
	hasZero := strings.Contains(code, "} else {")
	hasChain := strings.Contains(code, "} else if")

	if !hasNaN {
		t.Error("missing NaN branch in directional split")
	}
	if !hasPositive {
		t.Error("missing positive-change branch in directional split")
	}
	if !hasNegative {
		t.Error("missing negative-change branch in directional split")
	}
	if !hasZero {
		t.Error("missing zero-change catch-all branch in directional split")
	}
	if !hasChain {
		t.Error("branches must be chained with else-if so they are mutually exclusive")
	}
}
