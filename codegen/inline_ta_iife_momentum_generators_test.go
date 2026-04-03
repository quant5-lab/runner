package codegen

import (
	"fmt"
	"strings"
	"testing"

	"github.com/quant5-lab/runner/codegen/series_naming"
)

func TestMomentumIIFEGenerators_ImplementInterface(t *testing.T) {
	namer := series_naming.NewWindowBasedNamer()

	generators := []InlineTAIIFEGenerator{
		&RisingIIFEGenerator{namingStrategy: namer},
		&FallingIIFEGenerator{namingStrategy: namer},
		&HighestbarsIIFEGenerator{namingStrategy: namer},
		&LowestbarsIIFEGenerator{namingStrategy: namer},
		&MomIIFEGenerator{namingStrategy: namer},
		&RocIIFEGenerator{namingStrategy: namer},
		&CmoIIFEGenerator{namingStrategy: namer},
		&WprIIFEGenerator{namingStrategy: namer},
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

func TestRisingIIFEGenerator_MonotoneCheck(t *testing.T) {
	gen := &RisingIIFEGenerator{}
	accessor := NewArrowFunctionParameterAccessor("src")
	period := NewConstantPeriod(3)

	code := gen.Generate(accessor, period, "test")

	for _, pattern := range []string{"for j := 0", "curr", "prev", "<= prev", "return 0.0", "return 1.0"} {
		if !strings.Contains(code, pattern) {
			t.Errorf("RisingIIFEGenerator missing %q\nGot:\n%s", pattern, code)
		}
	}
}

func TestFallingIIFEGenerator_MonotoneCheck(t *testing.T) {
	gen := &FallingIIFEGenerator{}
	accessor := NewArrowFunctionParameterAccessor("src")
	period := NewConstantPeriod(3)

	code := gen.Generate(accessor, period, "test")

	for _, pattern := range []string{"for j := 0", "curr", "prev", ">= prev", "return 0.0", "return 1.0"} {
		if !strings.Contains(code, pattern) {
			t.Errorf("FallingIIFEGenerator missing %q\nGot:\n%s", pattern, code)
		}
	}
}

func TestHighestbarsIIFEGenerator_ReturnsNegativeOffset(t *testing.T) {
	gen := &HighestbarsIIFEGenerator{}
	accessor := NewArrowFunctionParameterAccessor("src")
	period := NewConstantPeriod(5)

	code := gen.Generate(accessor, period, "test")

	if !strings.Contains(code, "float64(-extIdx)") {
		t.Errorf("HighestbarsIIFEGenerator must return negative offset\nGot:\n%s", code)
	}
}

func TestLowestbarsIIFEGenerator_ReturnsNegativeOffset(t *testing.T) {
	gen := &LowestbarsIIFEGenerator{}
	accessor := NewArrowFunctionParameterAccessor("src")
	period := NewConstantPeriod(5)

	code := gen.Generate(accessor, period, "test")

	if !strings.Contains(code, "float64(-extIdx)") {
		t.Errorf("LowestbarsIIFEGenerator must return negative offset\nGot:\n%s", code)
	}
}

func TestMomIIFEGenerator_DifferenceFormula(t *testing.T) {
	gen := &MomIIFEGenerator{}
	accessor := NewArrowFunctionParameterAccessor("src")
	period := NewConstantPeriod(10)

	code := gen.Generate(accessor, period, "test")

	if !strings.Contains(code, "return ") || !strings.Contains(code, " - ") {
		t.Errorf("MomIIFEGenerator must compute difference\nGot:\n%s", code)
	}
}

func TestRocIIFEGenerator_GuardsZeroDivision(t *testing.T) {
	gen := &RocIIFEGenerator{}
	accessor := NewArrowFunctionParameterAccessor("src")
	period := NewConstantPeriod(12)

	code := gen.Generate(accessor, period, "test")

	for _, pattern := range []string{"math.IsNaN", "past == 0.0", "100.0"} {
		if !strings.Contains(code, pattern) {
			t.Errorf("RocIIFEGenerator missing %q\nGot:\n%s", pattern, code)
		}
	}
}

func TestCmoIIFEGenerator_UpDownAccumulation(t *testing.T) {
	gen := &CmoIIFEGenerator{}
	accessor := NewArrowFunctionParameterAccessor("src")
	period := NewConstantPeriod(9)

	code := gen.Generate(accessor, period, "test")

	for _, pattern := range []string{"up", "down", "total", "100.0"} {
		if !strings.Contains(code, pattern) {
			t.Errorf("CmoIIFEGenerator missing %q\nGot:\n%s", pattern, code)
		}
	}
}

func TestWprIIFEGenerator_UsesOHLCDirectly(t *testing.T) {
	gen := &WprIIFEGenerator{}
	accessor := NewArrowFunctionParameterAccessor("ignored")
	period := NewConstantPeriod(14)

	code := gen.Generate(accessor, period, "test")

	for _, pattern := range []string{"ctx.Data", "High", "Low", "Close", "100.0"} {
		if !strings.Contains(code, pattern) {
			t.Errorf("WprIIFEGenerator missing %q\nGot:\n%s", pattern, code)
		}
	}

	if strings.Contains(code, "ignoredSeries") {
		t.Errorf("WprIIFEGenerator must ignore accessor\nGot:\n%s", code)
	}
}

func TestMomentumIIFEGenerators_InRegistry(t *testing.T) {
	registry := NewInlineTAIIFERegistry()

	for _, name := range []string{
		"ta.rising", "rising",
		"ta.falling", "falling",
		"ta.highestbars", "highestbars",
		"ta.lowestbars", "lowestbars",
		"ta.mom", "mom",
		"ta.roc", "roc",
		"ta.cmo", "cmo",
		"ta.wpr", "wpr",
	} {
		if !registry.IsSupported(name) {
			t.Errorf("InlineTAIIFERegistry should support %q", name)
		}
	}
}

func TestRisingFallingIIFE_ViolationOperatorsDistinct(t *testing.T) {
	risingGen := &RisingIIFEGenerator{}
	fallingGen := &FallingIIFEGenerator{}
	accessor := NewArrowFunctionParameterAccessor("src")
	period := NewConstantPeriod(3)

	risingCode := risingGen.Generate(accessor, period, "test")
	fallingCode := fallingGen.Generate(accessor, period, "test")

	if !strings.Contains(risingCode, "<= prev") {
		t.Errorf("Rising violation check must use '<= prev'\nGot:\n%s", risingCode)
	}
	if !strings.Contains(fallingCode, ">= prev") {
		t.Errorf("Falling violation check must use '>= prev'\nGot:\n%s", fallingCode)
	}
	if strings.Contains(risingCode, ">= prev") {
		t.Errorf("Rising must not contain falling operator '>= prev'\nGot:\n%s", risingCode)
	}
}

func TestHighestLowestBarsIIFE_ComparisonOperatorsDistinct(t *testing.T) {
	highestGen := &HighestbarsIIFEGenerator{}
	lowestGen := &LowestbarsIIFEGenerator{}
	accessor := NewArrowFunctionParameterAccessor("src")
	period := NewConstantPeriod(5)

	highestCode := highestGen.Generate(accessor, period, "test")
	lowestCode := lowestGen.Generate(accessor, period, "test")

	if !strings.Contains(highestCode, "v > extVal") {
		t.Errorf("HighestbarsIIFE must use 'v > extVal'\nGot:\n%s", highestCode)
	}
	if !strings.Contains(lowestCode, "v < extVal") {
		t.Errorf("LowestbarsIIFE must use 'v < extVal'\nGot:\n%s", lowestCode)
	}
	if strings.Contains(highestCode, "v < extVal") {
		t.Errorf("HighestbarsIIFE must not use '<' comparison\nGot:\n%s", highestCode)
	}
}

/* IIFECodeBuilder emits ctx.BarIndex check only when warmupPeriod > 0 (period >= 2). */
func TestRisingIIFEGenerator_WarmupBehavior(t *testing.T) {
	gen := &RisingIIFEGenerator{}
	tests := []struct {
		period         int
		expectCheck    bool
		wantWarmupEdge int
	}{
		{1, false, 0},
		{2, true, 1},
		{3, true, 2},
		{14, true, 13},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("period_%d", tt.period), func(t *testing.T) {
			accessor := NewArrowFunctionParameterAccessor("src")
			period := NewConstantPeriod(tt.period)
			code := gen.Generate(accessor, period, "test")

			hasCheck := strings.Contains(code, "ctx.BarIndex <")
			if tt.expectCheck && !hasCheck {
				t.Errorf("Period %d must emit warmup check\nGot:\n%s", tt.period, code)
			}
			if !tt.expectCheck && hasCheck {
				t.Errorf("Period %d must not emit warmup check\nGot:\n%s", tt.period, code)
			}
			if tt.expectCheck {
				expected := fmt.Sprintf("ctx.BarIndex < %d", tt.wantWarmupEdge)
				if !strings.Contains(code, expected) {
					t.Errorf("Expected warmup check %q\nGot:\n%s", expected, code)
				}
			}
		})
	}
}

func TestHighestbarsIIFEGenerator_WarmupBehavior(t *testing.T) {
	gen := &HighestbarsIIFEGenerator{}
	tests := []struct {
		period         int
		expectCheck    bool
		wantWarmupEdge int
	}{
		{1, false, 0},
		{2, true, 1},
		{14, true, 13},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("period_%d", tt.period), func(t *testing.T) {
			accessor := NewArrowFunctionParameterAccessor("src")
			period := NewConstantPeriod(tt.period)
			code := gen.Generate(accessor, period, "test")

			hasCheck := strings.Contains(code, "ctx.BarIndex <")
			if tt.expectCheck && !hasCheck {
				t.Errorf("Period %d must emit warmup check\nGot:\n%s", tt.period, code)
			}
			if !tt.expectCheck && hasCheck {
				t.Errorf("Period %d must not emit warmup check\nGot:\n%s", tt.period, code)
			}
			if tt.expectCheck {
				expected := fmt.Sprintf("ctx.BarIndex < %d", tt.wantWarmupEdge)
				if !strings.Contains(code, expected) {
					t.Errorf("Expected warmup check %q\nGot:\n%s", expected, code)
				}
			}
		})
	}
}

func TestWprIIFEGenerator_AccessorIgnoredInOutput(t *testing.T) {
	gen := &WprIIFEGenerator{}
	accessor := NewArrowFunctionParameterAccessor("ignored")
	period := NewConstantPeriod(14)

	code := gen.Generate(accessor, period, "test")

	if strings.Contains(code, "ignored") {
		t.Errorf("WPR must not reference accessor name 'ignored' in output\nGot:\n%s", code)
	}
	if !strings.Contains(code, "ctx.Data") {
		t.Errorf("WPR must use ctx.Data directly\nGot:\n%s", code)
	}
}

func TestCmoIIFEGenerator_ZeroTotalGuard(t *testing.T) {
	gen := &CmoIIFEGenerator{}
	accessor := NewArrowFunctionParameterAccessor("src")
	period := NewConstantPeriod(9)

	code := gen.Generate(accessor, period, "test")

	t.Run("total_zero_guard_present", func(t *testing.T) {
		if !strings.Contains(code, "total == 0.0") {
			t.Errorf("CMO must guard against total == 0.0\nGot:\n%s", code)
		}
	})

	t.Run("returns_zero_not_nan_on_zero_total", func(t *testing.T) {
		if !strings.Contains(code, "return 0.0") {
			t.Errorf("CMO zero-total guard must return 0.0\nGot:\n%s", code)
		}
	})
}
