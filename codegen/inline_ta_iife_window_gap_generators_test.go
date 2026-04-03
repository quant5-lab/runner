package codegen

import (
	"fmt"
	"strings"
	"testing"

	"github.com/quant5-lab/runner/codegen/series_naming"
)

func TestWindowGapIIFEGenerators_ImplementInterface(t *testing.T) {
	namer := series_naming.NewWindowBasedNamer()

	generators := []InlineTAIIFEGenerator{
		&MaxIIFEGenerator{namingStrategy: namer},
		&MinIIFEGenerator{namingStrategy: namer},
		&RangeIIFEGenerator{namingStrategy: namer},
		&VarianceIIFEGenerator{namingStrategy: namer},
		&DevIIFEGenerator{namingStrategy: namer},
		&MedianIIFEGenerator{namingStrategy: namer},
		&ModeIIFEGenerator{namingStrategy: namer},
	}

	for _, gen := range generators {
		accessor := NewArrowFunctionParameterAccessor("src")
		period := NewConstantPeriod(10)
		code := gen.Generate(accessor, period, "test")

		if code == "" {
			t.Errorf("%T generated empty code", gen)
		}
		if !strings.Contains(code, "func() float64") {
			t.Errorf("%T missing IIFE wrapper", gen)
		}
	}
}

func TestMaxIIFEGenerator_Body(t *testing.T) {
	gen := &MaxIIFEGenerator{}
	accessor := NewArrowFunctionParameterAccessor("src")
	period := NewConstantPeriod(5)

	code := gen.Generate(accessor, period, "test")

	for _, pattern := range []string{"maxVal", "srcSeries.Get(0)", "for j := 1", "return maxVal"} {
		if !strings.Contains(code, pattern) {
			t.Errorf("MaxIIFEGenerator missing %q\nGot:\n%s", pattern, code)
		}
	}
}

func TestMinIIFEGenerator_Body(t *testing.T) {
	gen := &MinIIFEGenerator{}
	accessor := NewArrowFunctionParameterAccessor("src")
	period := NewConstantPeriod(5)

	code := gen.Generate(accessor, period, "test")

	for _, pattern := range []string{"minVal", "srcSeries.Get(0)", "for j := 1", "return minVal"} {
		if !strings.Contains(code, pattern) {
			t.Errorf("MinIIFEGenerator missing %q\nGot:\n%s", pattern, code)
		}
	}
}

func TestVarianceIIFEGenerator_TwoPassCalculation(t *testing.T) {
	gen := &VarianceIIFEGenerator{}
	accessor := NewArrowFunctionParameterAccessor("src")
	period := NewConstantPeriod(14)

	code := gen.Generate(accessor, period, "test")

	for _, pattern := range []string{"sum", "mean", "varianceAcc", "diff"} {
		if !strings.Contains(code, pattern) {
			t.Errorf("VarianceIIFEGenerator missing %q\nGot:\n%s", pattern, code)
		}
	}
}

func TestMedianIIFEGenerator_UsesSortFloat64s(t *testing.T) {
	gen := &MedianIIFEGenerator{}
	accessor := NewArrowFunctionParameterAccessor("src")
	period := NewConstantPeriod(7)

	code := gen.Generate(accessor, period, "test")

	if !strings.Contains(code, "sort.Float64s") {
		t.Errorf("MedianIIFEGenerator must use sort.Float64s\nGot:\n%s", code)
	}
}

func TestWindowGapGenerators_InIIFERegistry(t *testing.T) {
	registry := NewInlineTAIIFERegistry()

	for _, name := range []string{"ta.max", "max", "ta.min", "min", "ta.range", "range", "ta.variance", "variance", "ta.dev", "dev", "ta.median", "median", "ta.mode", "mode"} {
		if !registry.IsSupported(name) {
			t.Errorf("InlineTAIIFERegistry should support %q", name)
		}
	}
}

func TestMaxMinIIFEGenerators_ComparisonOperatorsDiffer(t *testing.T) {
	maxGen := &MaxIIFEGenerator{}
	minGen := &MinIIFEGenerator{}
	accessor := NewArrowFunctionParameterAccessor("src")
	period := NewConstantPeriod(5)

	maxCode := maxGen.Generate(accessor, period, "test")
	minCode := minGen.Generate(accessor, period, "test")

	if !strings.Contains(maxCode, "v > maxVal") {
		t.Errorf("MaxIIFEGenerator must use 'v > maxVal'\nGot:\n%s", maxCode)
	}
	if !strings.Contains(minCode, "v < minVal") {
		t.Errorf("MinIIFEGenerator must use 'v < minVal'\nGot:\n%s", minCode)
	}
	if strings.Contains(maxCode, "v < minVal") {
		t.Errorf("MaxIIFEGenerator must not use '<' comparison\nGot:\n%s", maxCode)
	}
}

func TestVarianceIIFEGenerator_TwoPassStructure(t *testing.T) {
	gen := &VarianceIIFEGenerator{}
	accessor := NewArrowFunctionParameterAccessor("src")
	period := NewConstantPeriod(10)

	code := gen.Generate(accessor, period, "test")

	t.Run("first_pass_sums_values", func(t *testing.T) {
		if !strings.Contains(code, "sum") {
			t.Errorf("Missing sum accumulator in first pass\nGot:\n%s", code)
		}
	})

	t.Run("computes_mean", func(t *testing.T) {
		if !strings.Contains(code, "mean") {
			t.Errorf("Missing mean computation\nGot:\n%s", code)
		}
	})

	t.Run("second_pass_accumulates_squared_diffs", func(t *testing.T) {
		if !strings.Contains(code, "diff") || !strings.Contains(code, "varianceAcc") {
			t.Errorf("Missing diff/varianceAcc in second pass\nGot:\n%s", code)
		}
	})

	t.Run("divides_by_period", func(t *testing.T) {
		if !strings.Contains(code, " / ") {
			t.Errorf("Variance must divide accumulator by period\nGot:\n%s", code)
		}
	})
}

func TestMedianIIFEGenerator_EvenOddIndexFormula(t *testing.T) {
	gen := &MedianIIFEGenerator{}
	accessor := NewArrowFunctionParameterAccessor("src")
	period := NewConstantPeriod(7)

	code := gen.Generate(accessor, period, "test")

	t.Run("odd_branch_uses_midpoint", func(t *testing.T) {
		if !strings.Contains(code, "medVals[n/2]") {
			t.Errorf("Missing odd-n median access 'medVals[n/2]'\nGot:\n%s", code)
		}
	})

	t.Run("even_branch_averages_two_midpoints", func(t *testing.T) {
		if !strings.Contains(code, "n/2-1") {
			t.Errorf("Missing even-n left midpoint 'n/2-1'\nGot:\n%s", code)
		}
	})

	t.Run("empty_check_returns_nan", func(t *testing.T) {
		if !strings.Contains(code, "math.NaN()") {
			t.Errorf("Median must return NaN for empty slice\nGot:\n%s", code)
		}
	})
}

func TestModeIIFEGenerator_FrequencyMapStructure(t *testing.T) {
	gen := &ModeIIFEGenerator{}
	accessor := NewArrowFunctionParameterAccessor("src")
	period := NewConstantPeriod(10)

	code := gen.Generate(accessor, period, "test")

	t.Run("frequency_map_present", func(t *testing.T) {
		if !strings.Contains(code, "modeFreq") {
			t.Errorf("Mode must use frequency map\nGot:\n%s", code)
		}
	})

	t.Run("iterates_frequency_map", func(t *testing.T) {
		if !strings.Contains(code, "for v, f := range modeFreq") {
			t.Errorf("Mode must iterate over frequency map\nGot:\n%s", code)
		}
	})

	t.Run("returns_mode_value", func(t *testing.T) {
		if !strings.Contains(code, "return modeVal") {
			t.Errorf("Mode must return modeVal\nGot:\n%s", code)
		}
	})
}

func TestRangeIIFEGenerator_MaxMinDifference(t *testing.T) {
	gen := &RangeIIFEGenerator{}
	accessor := NewArrowFunctionParameterAccessor("src")
	period := NewConstantPeriod(5)

	code := gen.Generate(accessor, period, "test")

	t.Run("tracks_max_and_min", func(t *testing.T) {
		if !strings.Contains(code, "maxVal") || !strings.Contains(code, "minVal") {
			t.Errorf("Range must track both maxVal and minVal\nGot:\n%s", code)
		}
	})

	t.Run("returns_max_minus_min", func(t *testing.T) {
		if !strings.Contains(code, "return maxVal - minVal") {
			t.Errorf("Range must return 'maxVal - minVal'\nGot:\n%s", code)
		}
	})
}

func TestDevIIFEGenerator_MeanAbsoluteDeviation(t *testing.T) {
	gen := &DevIIFEGenerator{}
	accessor := NewArrowFunctionParameterAccessor("src")
	period := NewConstantPeriod(10)

	code := gen.Generate(accessor, period, "test")

	t.Run("uses_math_abs", func(t *testing.T) {
		if !strings.Contains(code, "math.Abs") {
			t.Errorf("Dev must use math.Abs for absolute deviation\nGot:\n%s", code)
		}
	})

	t.Run("computes_mean", func(t *testing.T) {
		if !strings.Contains(code, "devMean") {
			t.Errorf("Dev must compute mean for deviation calculation\nGot:\n%s", code)
		}
	})

	t.Run("accumulates_absolute_deviations", func(t *testing.T) {
		if !strings.Contains(code, "devAcc") {
			t.Errorf("Dev must accumulate absolute deviations\nGot:\n%s", code)
		}
	})
}

/* IIFECodeBuilder emits ctx.BarIndex check only when warmupPeriod > 0 (period >= 2). */
func TestWindowGapIIFEGenerators_WarmupBehavior(t *testing.T) {
	tests := []struct {
		name              string
		period            int
		expectWarmupCheck bool
		wantWarmupEdge    int
	}{
		{"period_1_no_warmup", 1, false, 0},
		{"period_2_warmup_1", 2, true, 1},
		{"period_10_warmup_9", 10, true, 9},
		{"period_14_warmup_13", 14, true, 13},
	}

	generators := []struct {
		name string
		gen  InlineTAIIFEGenerator
	}{
		{"MaxIIFE", &MaxIIFEGenerator{}},
		{"MinIIFE", &MinIIFEGenerator{}},
		{"VarianceIIFE", &VarianceIIFEGenerator{}},
		{"MedianIIFE", &MedianIIFEGenerator{}},
	}

	for _, g := range generators {
		for _, tt := range tests {
			t.Run(g.name+"_"+tt.name, func(t *testing.T) {
				accessor := NewArrowFunctionParameterAccessor("src")
				period := NewConstantPeriod(tt.period)
				code := g.gen.Generate(accessor, period, "test")

				hasCheck := strings.Contains(code, "ctx.BarIndex <")

				if tt.expectWarmupCheck && !hasCheck {
					t.Errorf("Missing warmup check for period %d\nGot:\n%s", tt.period, code)
				}
				if !tt.expectWarmupCheck && hasCheck {
					t.Errorf("Period %d should not emit warmup check\nGot:\n%s", tt.period, code)
				}

				if tt.expectWarmupCheck {
					expected := fmt.Sprintf("ctx.BarIndex < %d", tt.wantWarmupEdge)
					if !strings.Contains(code, expected) {
						t.Errorf("Expected warmup check %q\nGot:\n%s", expected, code)
					}
				}
			})
		}
	}
}
