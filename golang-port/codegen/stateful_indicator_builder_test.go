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

	builder := NewStatefulIndicatorBuilder("ta.rma", "rma14", 14, mockAccessor, false)
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
		if !strings.Contains(code, "initialValue := sum / float64(14)") {
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

	builder := NewStatefulIndicatorBuilder("ta.rma", "rma10", 10, mockAccessor, true)
	code := builder.BuildRMA()

	t.Run("HasNaNCheckInInitialization", func(t *testing.T) {
		if !strings.Contains(code, "val := sourceSeries.Get(j)") {
			t.Error("Missing value extraction in initialization loop")
		}
		if !strings.Contains(code, "if math.IsNaN(val)") {
			t.Error("Missing NaN check in initialization")
		}
		if !strings.Contains(code, "return") {
			t.Error("Missing early return on NaN in initialization")
		}
	})

	t.Run("HasNaNCheckInRecursivePhase", func(t *testing.T) {
		if !strings.Contains(code, "if math.IsNaN(currentSource) || math.IsNaN(previousValue)") {
			t.Error("Missing NaN check for current and previous values")
		}
	})
}

func TestStatefulIndicatorBuilder_EMA_Structure(t *testing.T) {
	mockAccessor := &MockAccessGenerator{
		loopAccessFn: func(loopVar string) string {
			return "priceSeries.Get(" + loopVar + ")"
		},
	}

	builder := NewStatefulIndicatorBuilder("ta.ema", "ema20", 20, mockAccessor, false)
	code := builder.BuildEMA()

	t.Run("HasCorrectAlpha", func(t *testing.T) {
		if !strings.Contains(code, "alpha := 2.0 / float64(20+1)") {
			t.Error("EMA must use alpha = 2/(period+1), not 1/period")
		}
	})

	t.Run("HasSameStructureAsRMA", func(t *testing.T) {
		if !strings.Contains(code, "/* Inline EMA(20) - Stateful recursive calculation */") {
			t.Error("Missing EMA header comment")
		}
		if !strings.Contains(code, "previousValue := ema20Series.Get(1)") {
			t.Error("EMA must reference its own previous value")
		}
		if !strings.Contains(code, "newValue := alpha*currentSource + (1-alpha)*previousValue") {
			t.Error("Missing EMA recursive formula")
		}
	})
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

			builder := NewStatefulIndicatorBuilder("ta.rma", "test", tc.period, mockAccessor, false)
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
