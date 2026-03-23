package codegen

import (
	"fmt"
	"strings"
	"testing"
)

func TestStatefulIndicatorBuilder_RMA_Structure(t *testing.T) {
	mockAccessor := &MockAccessGenerator{
		loopAccessFn: func(loopVar string) string {
			return "sourceSeries.Get(" + loopVar + ")"
		},
		initialAccessFn: func(period int) string {
			return "sourceSeries.Get(" + string(rune(period-1)) + ")"
		},
	}

	builder := NewStatefulIndicatorBuilder("ta.rma", "rma14", P(14), mockAccessor, false, NewTopLevelIndicatorContext())
	code := builder.BuildRMA()

	t.Run("HasWarmupPhase", func(t *testing.T) {
		if !strings.Contains(code, "if ctx.BarIndex < 13") {
			t.Error("Missing warmup phase check for period 14")
		}
		if !strings.Contains(code, "rma14Series.Set(math.NaN())") {
			t.Error("Missing NaN assignment during warmup")
		}
	})

	t.Run("HasInitializationPhase", func(t *testing.T) {
		if !strings.Contains(code, "if ctx.BarIndex == 13") {
			t.Error("Missing initialization phase check")
		}
		if !strings.Contains(code, "/* First valid value: calculate SMA as initial state */") {
			t.Error("Missing SMA initialization comment")
		}
		if !strings.Contains(code, "for j := 0; j < 14; j++") {
			t.Error("Missing forward loop for SMA calculation")
		}
		if !strings.Contains(code, "initialValue := _sma_accumulator / float64(14)") {
			t.Error("Missing SMA calculation")
		}
	})

	t.Run("HasRecursivePhase", func(t *testing.T) {
		if !strings.Contains(code, "} else {") {
			t.Error("Missing else block for recursive phase")
		}
		if !strings.Contains(code, "/* Recursive phase: use previous indicator value */") {
			t.Error("Missing recursive phase comment")
		}
		if !strings.Contains(code, "previousValue := rma14Series.Get(1)") {
			t.Error("Missing previous value retrieval")
		}
		if !strings.Contains(code, "currentSource := sourceSeries.Get(0)") {
			t.Error("Missing current source value retrieval")
		}
	})

	t.Run("HasCorrectFormula", func(t *testing.T) {
		if !strings.Contains(code, "alpha := 1.0 / float64(14)") {
			t.Error("Missing alpha calculation with correct RMA formula")
		}
		if !strings.Contains(code, "newValue := alpha*currentSource + (1-alpha)*previousValue") {
			t.Error("Missing correct RMA recursive formula")
		}
		if !strings.Contains(code, "rma14Series.Set(newValue)") {
			t.Error("Missing result assignment")
		}
	})

	t.Run("NoBackwardLoop", func(t *testing.T) {
		if strings.Contains(code, "j--") {
			t.Error("Should not contain backward loop - RMA is stateful")
		}
	})
}

func TestStatefulIndicatorBuilder_RMA_WithNaNCheck(t *testing.T) {
	mockAccessor := &MockAccessGenerator{
		loopAccessFn: func(loopVar string) string {
			return "sourceSeries.Get(" + loopVar + ")"
		},
	}

	builder := NewStatefulIndicatorBuilder("ta.rma", "rma10", P(10), mockAccessor, true, NewTopLevelIndicatorContext())
	code := builder.BuildRMA()

	t.Run("InitPhase_FlagPatternDeclaredBeforeLoop", func(t *testing.T) {
		if !strings.Contains(code, "_sma_has_nan := false") {
			t.Error("Missing _sma_has_nan flag declaration before initialization loop")
		}
	})

	t.Run("InitPhase_LoopSetsFlag", func(t *testing.T) {
		if !strings.Contains(code, "val := sourceSeries.Get(j)") {
			t.Error("Missing value extraction in initialization loop")
		}
		if !strings.Contains(code, "if math.IsNaN(val)") {
			t.Error("Missing NaN check in initialization loop")
		}
		if !strings.Contains(code, "_sma_has_nan = true") {
			t.Error("Loop must set _sma_has_nan flag on NaN, not call Series.Set(NaN) inside loop")
		}
		if !strings.Contains(code, "break") {
			t.Error("Loop must break on NaN detection")
		}
	})

	t.Run("InitPhase_PostLoopFlagGuardsOutput", func(t *testing.T) {
		if !strings.Contains(code, "if _sma_has_nan {") {
			t.Error("Missing post-loop flag guard for NaN output")
		}
		loopEnd := strings.Index(code, "}")
		flagGuard := strings.Index(code, "if _sma_has_nan {")
		if flagGuard != -1 && loopEnd != -1 && flagGuard < loopEnd {
			t.Error("_sma_has_nan guard must appear after the initialization loop")
		}
	})

	t.Run("RecursivePhase_CurrentSourceNaNPropagates", func(t *testing.T) {
		if !strings.Contains(code, "if math.IsNaN(currentSource)") {
			t.Error("Missing NaN check for current source in recursive phase")
		}
	})

	t.Run("RecursivePhase_PrevNaNRecoversWithCurrentSource", func(t *testing.T) {
		if !strings.Contains(code, "else if math.IsNaN(previousValue)") {
			t.Error("Missing previous-NaN branch in recursive phase")
		}
		// Both EMA and RMA now recover from NaN previous by using currentSource
		prevNaNIdx := strings.Index(code, "else if math.IsNaN(previousValue)")
		if prevNaNIdx == -1 {
			t.Fatal("Missing else if math.IsNaN(previousValue)")
		}
		branchBody := code[prevNaNIdx:]
		nextElse := strings.Index(branchBody, "} else {")
		if nextElse == -1 {
			t.Fatal("Cannot find closing } else { of prev-NaN branch")
		}
		branchBody = branchBody[:nextElse]
		if !strings.Contains(branchBody, "currentSource") {
			t.Error("RMA prev-NaN branch must recover using currentSource")
		}
		if strings.Contains(branchBody, "math.NaN()") {
			t.Error("RMA prev-NaN branch must NOT propagate NaN — recovers with currentSource")
		}
	})
}

func TestStatefulIndicatorBuilder_isEMAVariant(t *testing.T) {
	tests := []struct {
		indicatorName string
		want          bool
	}{
		{"ta.ema", true},
		{"ema", true},
		{"ta.rma", false},
		{"rma", false},
		{"ta.atr", false},
		{"atr", false},
		{"ta.sma", false},
		{"sma", false},
		{"", false},
		{"ema_trend", false},
	}

	for _, tt := range tests {
		t.Run(tt.indicatorName, func(t *testing.T) {
			got := isEMAVariant(tt.indicatorName)
			if got != tt.want {
				t.Errorf("isEMAVariant(%q) = %v, want %v", tt.indicatorName, got, tt.want)
			}
		})
	}
}

func TestStatefulIndicatorBuilder_DifferentPeriods(t *testing.T) {
	testCases := []struct {
		name          string
		period        int
		warmupBar     int
		initBar       int
		loopCondition string
	}{
		{"Period5", 5, 4, 4, "for j := 0; j < 5; j++"},
		{"Period14", 14, 13, 13, "for j := 0; j < 14; j++"},
		{"Period50", 50, 49, 49, "for j := 0; j < 50; j++"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockAccessor := &MockAccessGenerator{
				loopAccessFn: func(loopVar string) string {
					return "src.Get(" + loopVar + ")"
				},
			}

			builder := NewStatefulIndicatorBuilder("ta.rma", "test", P(tc.period), mockAccessor, false, NewTopLevelIndicatorContext())
			code := builder.BuildRMA()

			warmupCheck := fmt.Sprintf("if ctx.BarIndex < %d", tc.warmupBar)
			if !strings.Contains(code, warmupCheck) {
				t.Errorf("Missing warmup check: %s", warmupCheck)
			}

			initCheck := fmt.Sprintf("if ctx.BarIndex == %d", tc.initBar)
			if !strings.Contains(code, initCheck) {
				t.Errorf("Missing initialization check: %s", initCheck)
			}

			if !strings.Contains(code, tc.loopCondition) {
				t.Errorf("Missing correct loop condition: %s", tc.loopCondition)
			}
		})
	}
}
