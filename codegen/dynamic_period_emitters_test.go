package codegen

import (
	"strings"
	"testing"
)

func TestDynamicPeriodEmitter_DispatchCompleteness(t *testing.T) {
	expectedFunctions := []string{
		"ta.sma", "ta.ema", "ta.rsi", "ta.stdev",
		"ta.highest", "ta.lowest", "ta.atr",
	}

	for _, fn := range expectedFunctions {
		t.Run(fn, func(t *testing.T) {
			emitter, ok := dynamicPeriodDispatch[fn]
			if !ok {
				t.Fatalf("dispatch map missing entry for %s", fn)
			}
			if emitter == nil {
				t.Fatalf("dispatch map has nil emitter for %s", fn)
			}
		})
	}

	t.Run("no_extra_entries", func(t *testing.T) {
		if len(dynamicPeriodDispatch) != len(expectedFunctions) {
			t.Errorf("dispatch map has %d entries, expected %d", len(dynamicPeriodDispatch), len(expectedFunctions))
		}
	})
}

func TestDynamicPeriodEmitter_WarmupThresholdClassification(t *testing.T) {
	g := newMinimalGenerator()

	t.Run("period_minus_1_warmup_group", func(t *testing.T) {
		emitters := map[string]DynamicPeriodEmitter{
			"SMA":     DynamicSMAEmitter{},
			"STDEV":   DynamicSTDEVEmitter{},
			"Highest": DynamicHighestEmitter{},
			"Lowest":  DynamicLowestEmitter{},
		}
		for name, emitter := range emitters {
			code := emitter.EmitCalculation(g, "test", "closeSeries")
			g.indent = 0
			if !strings.Contains(code, "ctx.BarIndex < period-1") {
				t.Errorf("%s should use period-1 warmup threshold", name)
			}
		}
	})

	t.Run("period_warmup_group", func(t *testing.T) {
		emitters := map[string]DynamicPeriodEmitter{
			"RSI": DynamicRSIEmitter{},
			"ATR": DynamicATREmitter{},
		}
		for name, emitter := range emitters {
			code := emitter.EmitCalculation(g, "test", "closeSeries")
			g.indent = 0
			if !strings.Contains(code, "ctx.BarIndex < period") {
				t.Errorf("%s should use period warmup threshold", name)
			}
			if strings.Contains(code, "ctx.BarIndex < period-1") {
				t.Errorf("%s should NOT use period-1 warmup threshold", name)
			}
		}
	})

	t.Run("ema_unique_warmup", func(t *testing.T) {
		code := DynamicEMAEmitter{}.EmitCalculation(g, "test", "closeSeries")
		g.indent = 0
		if strings.Contains(code, "ctx.BarIndex < period-1") || strings.Contains(code, "ctx.BarIndex < period") {
			t.Error("EMA should not use standard warmup guard threshold")
		}
		if !strings.Contains(code, "ctx.BarIndex == 0") {
			t.Error("EMA should have bar-0 seeding logic")
		}
	})
}

func TestDynamicPeriodEmitter_VarNamePropagation(t *testing.T) {
	g := newMinimalGenerator()

	allEmitters := map[string]DynamicPeriodEmitter{
		"SMA":     DynamicSMAEmitter{},
		"EMA":     DynamicEMAEmitter{},
		"RSI":     DynamicRSIEmitter{},
		"STDEV":   DynamicSTDEVEmitter{},
		"Highest": DynamicHighestEmitter{},
		"Lowest":  DynamicLowestEmitter{},
		"ATR":     DynamicATREmitter{},
	}

	for name, emitter := range allEmitters {
		t.Run(name, func(t *testing.T) {
			code := emitter.EmitCalculation(g, "customIndicator", "closeSeries")
			g.indent = 0

			if !strings.Contains(code, "customIndicatorSeries.Set(") {
				t.Errorf("emitter must write to customIndicatorSeries.Set(), got:\n%s", code)
			}
		})
	}
}

func TestDynamicPeriodEmitter_SourceAccessorUsage(t *testing.T) {
	g := newMinimalGenerator()

	t.Run("source_dependent_emitters_use_accessor", func(t *testing.T) {
		emitters := map[string]DynamicPeriodEmitter{
			"SMA":     DynamicSMAEmitter{},
			"EMA":     DynamicEMAEmitter{},
			"RSI":     DynamicRSIEmitter{},
			"STDEV":   DynamicSTDEVEmitter{},
			"Highest": DynamicHighestEmitter{},
			"Lowest":  DynamicLowestEmitter{},
		}
		for name, emitter := range emitters {
			code := emitter.EmitCalculation(g, "test", "mySrcSeries")
			g.indent = 0
			if !strings.Contains(code, "mySrcSeries.Get(") {
				t.Errorf("%s should reference mySrcSeries.Get() in calculation", name)
			}
		}
	})

	t.Run("atr_ignores_source_accessor", func(t *testing.T) {
		code := DynamicATREmitter{}.EmitCalculation(g, "test", "shouldBeIgnored")
		g.indent = 0
		if strings.Contains(code, "shouldBeIgnored") {
			t.Error("ATR should ignore sourceAccessor parameter")
		}
	})

	t.Run("atr_uses_ohlc_series", func(t *testing.T) {
		code := DynamicATREmitter{}.EmitCalculation(g, "test", "")
		g.indent = 0
		mustContain := []string{"highSeries.Get(", "lowSeries.Get(", "closeSeries.Get("}
		for _, pattern := range mustContain {
			if !strings.Contains(code, pattern) {
				t.Errorf("ATR should reference %s", pattern)
			}
		}
	})
}

func TestDynamicPeriodEmitter_SMAAlgorithm(t *testing.T) {
	g := newMinimalGenerator()
	code := DynamicSMAEmitter{}.EmitCalculation(g, "mySma", "closeSeries")

	requiredComponents := []struct {
		pattern string
		desc    string
	}{
		{"sum := 0.0", "accumulator initialization"},
		{"for j := 0; j < period; j++", "iteration over window"},
		{"sum += closeSeries.Get(j)", "source value accumulation"},
		{"mySmaSeries.Set(sum / float64(period))", "mean calculation"},
	}

	for _, rc := range requiredComponents {
		t.Run(rc.desc, func(t *testing.T) {
			if !strings.Contains(code, rc.pattern) {
				t.Errorf("SMA missing %s: %q", rc.desc, rc.pattern)
			}
		})
	}
}

func TestDynamicPeriodEmitter_EMAAlgorithm(t *testing.T) {
	g := newMinimalGenerator()
	code := DynamicEMAEmitter{}.EmitCalculation(g, "myEma", "closeSeries")

	t.Run("smoothing_factor", func(t *testing.T) {
		if !strings.Contains(code, "alpha := 2.0 / (float64(period) + 1.0)") {
			t.Errorf("EMA missing alpha calculation")
		}
	})

	t.Run("source_read", func(t *testing.T) {
		if !strings.Contains(code, "src := closeSeries.Get(0)") {
			t.Errorf("EMA missing current source read")
		}
	})

	t.Run("bar_zero_seeding", func(t *testing.T) {
		if !strings.Contains(code, "if ctx.BarIndex == 0") {
			t.Errorf("EMA missing bar-0 seeding check")
		}
	})

	t.Run("nan_fallback", func(t *testing.T) {
		if !strings.Contains(code, "math.IsNaN(prev)") {
			t.Errorf("EMA missing NaN-prev fallback")
		}
	})

	t.Run("recursive_formula", func(t *testing.T) {
		if !strings.Contains(code, "alpha*src + (1.0-alpha)*prev") {
			t.Errorf("EMA missing recursive smoothing formula")
		}
	})

	t.Run("previous_value_access", func(t *testing.T) {
		if !strings.Contains(code, "prev := myEmaSeries.Get(1)") {
			t.Errorf("EMA missing self-referential previous value access")
		}
	})

	t.Run("no_loop", func(t *testing.T) {
		if strings.Contains(code, "for j") {
			t.Error("EMA should not use a loop — it is recursive")
		}
	})
}

func TestDynamicPeriodEmitter_RSIAlgorithm(t *testing.T) {
	g := newMinimalGenerator()
	code := DynamicRSIEmitter{}.EmitCalculation(g, "myRsi", "closeSeries")

	requiredComponents := []struct {
		pattern string
		desc    string
	}{
		{"var gains, losses float64", "gain/loss accumulators"},
		{"for j := 0; j < period; j++", "iteration over window"},
		{"closeSeries.Get(j) - closeSeries.Get(j+1)", "price change calculation"},
		{"if change > 0", "gain/loss classification"},
		{"gains += change", "gain accumulation"},
		{"losses -= change", "loss accumulation (absolute value)"},
		{"avgGain := gains / float64(period)", "average gain"},
		{"avgLoss := losses / float64(period)", "average loss"},
		{"if avgLoss == 0", "zero-loss edge case"},
		{"myRsiSeries.Set(100.0)", "RSI=100 when no losses"},
		{"rs := avgGain / avgLoss", "relative strength"},
		{"myRsiSeries.Set(100.0 - 100.0/(1.0+rs))", "RSI formula"},
	}

	for _, rc := range requiredComponents {
		t.Run(rc.desc, func(t *testing.T) {
			if !strings.Contains(code, rc.pattern) {
				t.Errorf("RSI missing %s: %q", rc.desc, rc.pattern)
			}
		})
	}
}

func TestDynamicPeriodEmitter_STDEVAlgorithm(t *testing.T) {
	g := newMinimalGenerator()
	code := DynamicSTDEVEmitter{}.EmitCalculation(g, "myStdev", "closeSeries")

	requiredComponents := []struct {
		pattern string
		desc    string
	}{
		{"sum := 0.0", "sum accumulator"},
		{"mean := sum / float64(period)", "mean calculation"},
		{"variance := 0.0", "variance accumulator"},
		{"diff := closeSeries.Get(j) - mean", "deviation from mean"},
		{"variance += diff * diff", "squared deviation accumulation"},
		{"myStdevSeries.Set(math.Sqrt(variance / float64(period)))", "population stdev formula"},
	}

	for _, rc := range requiredComponents {
		t.Run(rc.desc, func(t *testing.T) {
			if !strings.Contains(code, rc.pattern) {
				t.Errorf("STDEV missing %s: %q", rc.desc, rc.pattern)
			}
		})
	}

	t.Run("two_pass_algorithm", func(t *testing.T) {
		sumIdx := strings.Index(code, "sum := 0.0")
		varianceIdx := strings.Index(code, "variance := 0.0")
		if sumIdx >= varianceIdx {
			t.Error("STDEV should compute sum before variance (two-pass)")
		}
	})
}

func TestDynamicPeriodEmitter_HighestAlgorithm(t *testing.T) {
	g := newMinimalGenerator()
	code := DynamicHighestEmitter{}.EmitCalculation(g, "myHigh", "highSeries")

	requiredComponents := []struct {
		pattern string
		desc    string
	}{
		{"maxVal := highSeries.Get(0)", "initial max from first element"},
		{"for j := 1; j < period; j++", "iteration starting from second element"},
		{"val := highSeries.Get(j)", "element access"},
		{"if val > maxVal", "max comparison"},
		{"maxVal = val", "max update"},
		{"myHighSeries.Set(maxVal)", "result assignment"},
	}

	for _, rc := range requiredComponents {
		t.Run(rc.desc, func(t *testing.T) {
			if !strings.Contains(code, rc.pattern) {
				t.Errorf("Highest missing %s: %q", rc.desc, rc.pattern)
			}
		})
	}
}

func TestDynamicPeriodEmitter_LowestAlgorithm(t *testing.T) {
	g := newMinimalGenerator()
	code := DynamicLowestEmitter{}.EmitCalculation(g, "myLow", "lowSeries")

	requiredComponents := []struct {
		pattern string
		desc    string
	}{
		{"minVal := lowSeries.Get(0)", "initial min from first element"},
		{"for j := 1; j < period; j++", "iteration starting from second element"},
		{"val := lowSeries.Get(j)", "element access"},
		{"if val < minVal", "min comparison"},
		{"minVal = val", "min update"},
		{"myLowSeries.Set(minVal)", "result assignment"},
	}

	for _, rc := range requiredComponents {
		t.Run(rc.desc, func(t *testing.T) {
			if !strings.Contains(code, rc.pattern) {
				t.Errorf("Lowest missing %s: %q", rc.desc, rc.pattern)
			}
		})
	}
}

func TestDynamicPeriodEmitter_HighestLowestSymmetry(t *testing.T) {
	g := newMinimalGenerator()

	highCode := DynamicHighestEmitter{}.EmitCalculation(g, "test", "srcSeries")
	g.indent = 0
	lowCode := DynamicLowestEmitter{}.EmitCalculation(g, "test", "srcSeries")
	g.indent = 0

	t.Run("same_warmup", func(t *testing.T) {
		if strings.Contains(highCode, "period-1") != strings.Contains(lowCode, "period-1") {
			t.Error("Highest and Lowest should use the same warmup threshold")
		}
	})

	t.Run("same_loop_structure", func(t *testing.T) {
		if strings.Contains(highCode, "for j := 1") != strings.Contains(lowCode, "for j := 1") {
			t.Error("Highest and Lowest should use the same loop structure")
		}
	})

	t.Run("opposite_comparisons", func(t *testing.T) {
		if !strings.Contains(highCode, "val > maxVal") {
			t.Error("Highest should use > comparison")
		}
		if !strings.Contains(lowCode, "val < minVal") {
			t.Error("Lowest should use < comparison")
		}
	})
}

func TestDynamicPeriodEmitter_ATRAlgorithm(t *testing.T) {
	g := newMinimalGenerator()
	code := DynamicATREmitter{}.EmitCalculation(g, "myAtr", "")

	requiredComponents := []struct {
		pattern string
		desc    string
	}{
		{"sum := 0.0", "TR accumulator"},
		{"for j := 0; j < period; j++", "iteration over window"},
		{"high := highSeries.Get(j)", "high price access"},
		{"low := lowSeries.Get(j)", "low price access"},
		{"prevClose := closeSeries.Get(j + 1)", "previous close access"},
		{"math.Max(high-low,", "true range: high-low component"},
		{"math.Abs(high-prevClose)", "true range: high-prevClose component"},
		{"math.Abs(low-prevClose)", "true range: low-prevClose component"},
		{"myAtrSeries.Set(sum / float64(period))", "average true range"},
	}

	for _, rc := range requiredComponents {
		t.Run(rc.desc, func(t *testing.T) {
			if !strings.Contains(code, rc.pattern) {
				t.Errorf("ATR missing %s: %q", rc.desc, rc.pattern)
			}
		})
	}
}

func TestDynamicPeriodEmitter_InvalidPeriodGuard(t *testing.T) {
	g := newMinimalGenerator()

	allEmitters := map[string]DynamicPeriodEmitter{
		"SMA":     DynamicSMAEmitter{},
		"EMA":     DynamicEMAEmitter{},
		"RSI":     DynamicRSIEmitter{},
		"STDEV":   DynamicSTDEVEmitter{},
		"Highest": DynamicHighestEmitter{},
		"Lowest":  DynamicLowestEmitter{},
		"ATR":     DynamicATREmitter{},
	}

	for name, emitter := range allEmitters {
		t.Run(name, func(t *testing.T) {
			code := emitter.EmitCalculation(g, "test", "closeSeries")
			g.indent = 0

			if !strings.Contains(code, "period <= 0") {
				t.Errorf("%s must guard against non-positive period values", name)
			}
			if !strings.Contains(code, "testSeries.Set(math.NaN())") {
				t.Errorf("%s must emit NaN for invalid period", name)
			}
		})
	}
}
