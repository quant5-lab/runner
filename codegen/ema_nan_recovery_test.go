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
	accessor := NewOHLCVFieldAccessGenerator("Close")
	context := NewTopLevelIndicatorContext()
	builder := NewStatefulIndicatorBuilder("ta.ema", "testEma", P(20), accessor, true, context)

	code := builder.BuildEMA()

	if !strings.Contains(code, "if math.IsNaN(previousValue)") {
		t.Error("Missing previousValue NaN check")
	}
	if !strings.Contains(code, "testEmaSeries.Set(currentSource)") {
		t.Error("EMA prev-NaN recovery must use current source (Pine: na(sum[1]) ? src)")
	}
	if !strings.Contains(code, "alpha*currentSource") {
		t.Error("EMA recursive formula must use alpha*currentSource")
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

// TestRMAvsEMA_PrevNaNBranchContract validates the prev-NaN branch of the recursive phase
// for both the needsNaN=true (NaN guards present) and needsNaN=false (no guards) cases.
//
// Pine semantics:
//   - EMA: na(sum[1]) ? src — restarts from current source (single-value defined recovery)
//   - RMA: na(sum[1]) ? ta.sma(source, length) — re-seeds via SMA every bar until the
//     full window is NaN-free (RMA has no single-value fallback; SMA must succeed first)
//
// needsNaN=false (OHLCV source — always finite): neither indicator emits NaN guards.
func TestRMAvsEMA_PrevNaNBranchContract(t *testing.T) {
	const (
		prevNaNSourceRestart = "source_restart" // EMA: Set(currentSource)
		prevNaNSMASeed       = "sma_seed"       // RMA: re-seed via SMA accumulation loop
		prevNaNNone          = "none"           // needsNaN=false: no guards emitted
	)

	tests := []struct {
		name            string
		indicator       string
		buildFn         func(*StatefulIndicatorBuilder) string
		needsNaN        bool
		prevNaNStrategy string
	}{
		{
			name:            "EMA restarts from current source when previous is NaN",
			indicator:       "ta.ema",
			buildFn:         func(b *StatefulIndicatorBuilder) string { return b.BuildEMA() },
			needsNaN:        true,
			prevNaNStrategy: prevNaNSourceRestart,
		},
		{
			name:            "EMA bare alias restarts from current source when previous is NaN",
			indicator:       "ema",
			buildFn:         func(b *StatefulIndicatorBuilder) string { return b.BuildEMA() },
			needsNaN:        true,
			prevNaNStrategy: prevNaNSourceRestart,
		},
		{
			name:            "RMA re-seeds via SMA when previous is NaN (Pine: na(sum[1]) ? ta.sma)",
			indicator:       "ta.rma",
			buildFn:         func(b *StatefulIndicatorBuilder) string { return b.BuildRMA() },
			needsNaN:        true,
			prevNaNStrategy: prevNaNSMASeed,
		},
		{
			name:            "RMA bare alias re-seeds via SMA when previous is NaN",
			indicator:       "rma",
			buildFn:         func(b *StatefulIndicatorBuilder) string { return b.BuildRMA() },
			needsNaN:        true,
			prevNaNStrategy: prevNaNSMASeed,
		},
		{
			name:            "EMA needsNaN=false emits no NaN guards",
			indicator:       "ta.ema",
			buildFn:         func(b *StatefulIndicatorBuilder) string { return b.BuildEMA() },
			needsNaN:        false,
			prevNaNStrategy: prevNaNNone,
		},
		{
			name:            "RMA needsNaN=false emits no NaN guards",
			indicator:       "ta.rma",
			buildFn:         func(b *StatefulIndicatorBuilder) string { return b.BuildRMA() },
			needsNaN:        false,
			prevNaNStrategy: prevNaNNone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accessor := NewOHLCVFieldAccessGenerator("Close")
			context := NewTopLevelIndicatorContext()
			builder := NewStatefulIndicatorBuilder(tt.indicator, "testIndicator", P(14), accessor, tt.needsNaN, context)
			code := tt.buildFn(builder)

			if tt.prevNaNStrategy == prevNaNNone {
				if strings.Contains(code, "math.IsNaN(currentSource)") {
					t.Errorf("%s needsNaN=false must not emit currentSource NaN guard", tt.indicator)
				}
				if strings.Contains(code, "math.IsNaN(previousValue)") {
					t.Errorf("%s needsNaN=false must not emit previousValue NaN guard", tt.indicator)
				}
				return
			}

			prevNaNIdx := strings.Index(code, "else if math.IsNaN(previousValue)")
			if prevNaNIdx == -1 {
				t.Fatal("Missing 'else if math.IsNaN(previousValue)' branch in recursive phase")
			}

			afterBranch := code[prevNaNIdx:]
			openBrace := strings.Index(afterBranch, "{")
			if openBrace == -1 {
				t.Fatal("Cannot find opening brace of prev-NaN branch")
			}
			closingElse := strings.Index(afterBranch[openBrace:], "} else {")
			if closingElse == -1 {
				t.Fatal("Cannot find '} else {' closing the prev-NaN branch")
			}
			branchBody := afterBranch[openBrace : openBrace+closingElse]

			switch tt.prevNaNStrategy {
			case prevNaNSourceRestart:
				if !strings.Contains(branchBody, "currentSource") {
					t.Errorf("%s prev-NaN branch must use currentSource\nBranch: %s", tt.indicator, branchBody)
				}
			case prevNaNSMASeed:
				if !strings.Contains(branchBody, "_sma_accumulator") {
					t.Errorf("%s prev-NaN branch must use SMA accumulation loop\nBranch: %s", tt.indicator, branchBody)
				}
				if strings.Contains(branchBody, "currentSource") && !strings.Contains(branchBody, "_sma_accumulator") {
					t.Errorf("%s prev-NaN branch must not fall back to currentSource (Pine uses SMA)\nBranch: %s", tt.indicator, branchBody)
				}
			}
		})
	}
}

// TestRMAPrevNaNInitSMALoop validates that the RMA init phase emits a NaN-aware SMA
// accumulation loop using the deferred-flag pattern: flag set inside loop, output
// guarded after loop so that no series update occurs inside the loop body.
func TestRMAPrevNaNInitSMALoop(t *testing.T) {
	tests := []struct {
		name      string
		indicator string
		period    int
	}{
		{"ta.rma period 2", "ta.rma", 2},
		{"ta.rma period 14", "ta.rma", 14},
		{"ta.rma period 50", "ta.rma", 50},
		{"rma bare alias period 14", "rma", 14},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accessor := NewOHLCVFieldAccessGenerator("Close")
			context := NewTopLevelIndicatorContext()
			builder := NewStatefulIndicatorBuilder(tt.indicator, "testRMA", P(tt.period), accessor, true, context)
			code := builder.BuildRMA()

			if !strings.Contains(code, "_sma_accumulator := 0.0") {
				t.Error("Missing SMA accumulation in init phase")
			}
			if !strings.Contains(code, "_sma_has_nan := false") {
				t.Error("Missing _sma_has_nan flag declaration")
			}
			if !strings.Contains(code, "if math.IsNaN(val)") {
				t.Error("Init phase SMA loop must check for NaN")
			}
			if !strings.Contains(code, "_sma_has_nan = true") {
				t.Error("Loop must set flag on NaN detection")
			}
			if !strings.Contains(code, "break") {
				t.Error("Init phase SMA loop must break on NaN")
			}
			if !strings.Contains(code, "if _sma_has_nan {") {
				t.Error("Missing post-loop flag guard")
			}
		})
	}
}

// TestEMAEdgeCasePeriods validates EMA correctness for edge case periods.
func TestEMAEdgeCasePeriods(t *testing.T) {
	edgePeriods := []int{1, 2, 3, 200, 500}

	for _, period := range edgePeriods {
		t.Run(fmt.Sprintf("period_%d", period), func(t *testing.T) {
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
