package codegen

import (
	"fmt"
	"strings"
	"testing"
)

// TestEMANaNRecovery validates EMA NaN recovery behavior.
// Pine Script formula: sum := na(sum[1]) ? src : alpha * src + (1 - alpha) * nz(sum[1])
// When previous EMA is NaN, should use current source as starting point.
func TestEMANaNRecovery(t *testing.T) {
	tests := []struct {
		name              string
		wantPrevNaNBranch string
		wantElseBranch    string
		description       string
	}{
		{
			name:              "prevEMA NaN uses current source",
			wantPrevNaNBranch: "testEmaSeries.Set(currentSource)",
			wantElseBranch:    "alpha*currentSource",
			description:       "When prevEMA is NaN, EMA should use current source value (Pine: na(sum[1]) ? src)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accessor := NewOHLCVFieldAccessGenerator("Close")
			context := NewTopLevelIndicatorContext()
			builder := NewStatefulIndicatorBuilder("ta.ema", "testEma", P(20), accessor, true, context)

			code := builder.BuildEMA()

			if !strings.Contains(code, "if math.IsNaN(previousValue)") {
				t.Error("Missing previousValue NaN check")
			}

			if !strings.Contains(code, tt.wantPrevNaNBranch) {
				t.Errorf("Missing expected recovery pattern %q", tt.wantPrevNaNBranch)
			}

			if !strings.Contains(code, tt.wantElseBranch) {
				t.Errorf("Missing expected EMA formula pattern %q", tt.wantElseBranch)
			}
		})
	}
}

// TestEMASourceNaNHandling validates EMA propagates NaN when current source is NaN.
func TestEMASourceNaNHandling(t *testing.T) {
	accessor := NewOHLCVFieldAccessGenerator("Close")
	context := NewTopLevelIndicatorContext()
	builder := NewStatefulIndicatorBuilder("ta.ema", "testEma", P(20), accessor, true, context)

	code := builder.BuildEMA()

	if !strings.Contains(code, "math.IsNaN(currentSource)") {
		t.Error("Missing current source NaN check")
	}

	if idx := strings.Index(code, "if math.IsNaN(currentSource)"); idx >= 0 {
		if !strings.Contains(code[idx:], "Set(math.NaN())") {
			t.Error("Source NaN should propagate NaN to EMA")
		}
	}
}

// TestEMAAlphaCalculation validates alpha formula: 2 / (length + 1).
func TestEMAAlphaCalculation(t *testing.T) {
	periods := []int{9, 14, 20, 50, 200}

	for _, period := range periods {
		t.Run(fmt.Sprintf("%d_period", period), func(t *testing.T) {
			accessor := NewOHLCVFieldAccessGenerator("Close")
			context := NewTopLevelIndicatorContext()
			builder := NewStatefulIndicatorBuilder("ta.ema", "testEma", P(period), accessor, true, context)

			code := builder.BuildEMA()

			expectedAlpha := fmt.Sprintf("2.0 / float64(%d+1)", period)
			if !strings.Contains(code, expectedAlpha) {
				t.Errorf("Missing or incorrect alpha formula, expected %q", expectedAlpha)
			}

			warmupIdx := strings.Index(code, "ctx.BarIndex ==")
			recursiveIdx := strings.Index(code, "} else {")
			alphaIdx := strings.Index(code, "alpha :=")

			if warmupIdx > 0 && alphaIdx > 0 && alphaIdx < warmupIdx {
				t.Error("Alpha should be in recursive phase, not warmup")
			}

			if recursiveIdx > 0 && alphaIdx > 0 && alphaIdx < recursiveIdx {
				t.Error("Alpha should be after recursive phase starts")
			}
		})
	}
}

// TestEMAWarmupPhases validates three-phase initialization:
// pre-warmup (NaN), warmup (SMA seed), recursive (EMA formula).
func TestEMAWarmupPhases(t *testing.T) {
	accessor := NewOHLCVFieldAccessGenerator("Close")
	context := NewTopLevelIndicatorContext()
	period := 20
	builder := NewStatefulIndicatorBuilder("ta.ema", "testEma", P(period), accessor, true, context)

	code := builder.BuildEMA()

	expectedPreWarmup := fmt.Sprintf("ctx.BarIndex < %d", period-1)
	if !strings.Contains(code, expectedPreWarmup) {
		t.Errorf("Missing pre-warmup phase check: %q", expectedPreWarmup)
	}

	expectedWarmup := fmt.Sprintf("ctx.BarIndex == %d", period-1)
	if !strings.Contains(code, expectedWarmup) {
		t.Errorf("Missing warmup phase check: %q", expectedWarmup)
	}

	if !strings.Contains(code, "_sma_accumulator") {
		t.Error("Warmup missing SMA accumulation")
	}

	if !strings.Contains(code, "for j") {
		t.Error("Warmup missing SMA loop")
	}

	if !strings.Contains(code, "} else {") {
		t.Error("Missing recursive phase")
	}

	if !strings.Contains(code, "previousValue") {
		t.Error("Recursive phase missing previous EMA access")
	}

	if !strings.Contains(code, "alpha*") || !strings.Contains(code, "(1-alpha)") {
		t.Error("Recursive phase missing EMA formula")
	}
}

// TestEMAPreviousValueAccess validates ForwardSeriesBuffer access pattern.
func TestEMAPreviousValueAccess(t *testing.T) {
	accessor := NewOHLCVFieldAccessGenerator("Close")
	context := NewTopLevelIndicatorContext()
	generator := NewStatefulEMAGenerator("testEma", 20, accessor, context)

	code := generator.GenerateEMA()

	if !strings.Contains(code, "Series.Get(1)") {
		t.Error("Missing Series.Get(1) for previous EMA")
	}

	if !strings.Contains(code, "Series.Set(") {
		t.Error("Missing Series.Set() for current EMA")
	}

	if strings.Contains(code, "[ctx.BarIndex-1]") {
		t.Error("Should not use array lookback pattern")
	}
}

// TestEMAAccessorCompatibility validates EMA works with OHLCV and series accessors.
func TestEMAAccessorCompatibility(t *testing.T) {
	tests := []struct {
		name            string
		accessor        AccessGenerator
		expectInWarmup  string
		expectRecursive string
		description     string
	}{
		{
			name:            "OHLCV field accessor",
			accessor:        NewOHLCVFieldAccessGenerator("High"),
			expectInWarmup:  "highSeries.Get(j)",
			expectRecursive: "currentSource := highSeries.Get(0)",
			description:     "OHLCV field access",
		},
		{
			name:            "Series variable accessor",
			accessor:        NewSeriesVariableAccessGenerator("cagr5"),
			expectInWarmup:  "cagr5Series.Get(j)",
			expectRecursive: "currentSource := cagr5Series.Get(0)",
			description:     "Series variable access",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			context := NewTopLevelIndicatorContext()
			builder := NewStatefulIndicatorBuilder("ta.ema", "testEma", P(20), tt.accessor, true, context)
			code := builder.BuildEMA()

			if !strings.Contains(code, tt.expectInWarmup) {
				t.Errorf("Warmup missing expected pattern %q (%s)", tt.expectInWarmup, tt.description)
			}

			if !strings.Contains(code, tt.expectRecursive) {
				t.Errorf("Recursive phase missing expected pattern %q (%s)", tt.expectRecursive, tt.description)
			}
		})
	}
}

// TestEMAContextTypes validates EMA works in top-level and arrow function contexts.
func TestEMAContextTypes(t *testing.T) {
	tests := []struct {
		name             string
		context          StatefulIndicatorContext
		expectSetPattern string
		description      string
	}{
		{
			name:             "top-level context",
			context:          NewTopLevelIndicatorContext(),
			expectSetPattern: "testEmaSeries.Set(",
			description:      "Top-level direct access",
		},
		{
			name:             "arrow function context",
			context:          NewArrowFunctionIndicatorContext(),
			expectSetPattern: "arrowCtx.GetOrCreateSeries(",
			description:      "Arrow function context access",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accessor := NewOHLCVFieldAccessGenerator("Close")
			generator := NewStatefulEMAGenerator("testEma", 20, accessor, tt.context)
			code := generator.GenerateEMA()

			if !strings.Contains(code, tt.expectSetPattern) {
				t.Errorf("Missing expected pattern %q (%s)", tt.expectSetPattern, tt.description)
			}
		})
	}
}

// TestEMAEdgeCasePeriods validates EMA correctness for edge case periods.
func TestEMAEdgeCasePeriods(t *testing.T) {
	edgePeriods := []int{1, 2, 3, 200, 500}

	for _, period := range edgePeriods {
		t.Run(string(rune('0'+period/100))+string(rune('0'+(period/10)%10))+string(rune('0'+period%10)), func(t *testing.T) {
			accessor := NewOHLCVFieldAccessGenerator("Close")
			context := NewTopLevelIndicatorContext()
			generator := NewStatefulEMAGenerator("testEma", period, accessor, context)

			code := generator.GenerateEMA()

			if !strings.Contains(code, "ctx.BarIndex") {
				t.Error("Missing bar index check")
			}

			if !strings.Contains(code, "alpha") {
				t.Error("Missing alpha calculation")
			}

			if strings.Count(code, "{") != strings.Count(code, "}") {
				t.Errorf("Unbalanced braces for period %d", period)
			}

			if period > 1 && !strings.Contains(code, "for j") {
				t.Error("Missing warmup loop for period > 1")
			}
		})
	}
}
