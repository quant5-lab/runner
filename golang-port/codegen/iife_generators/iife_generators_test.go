package iife_generators

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/codegen"
	"github.com/quant5-lab/runner/codegen/series_naming"
)

/* TestRMAGenerator_BasicGeneration tests RMA IIFE code generation */
func TestRMAGenerator_BasicGeneration(t *testing.T) {
	namer := series_naming.NewStatefulIndicatorNamer()
	gen := NewRMAGenerator(namer)

	classifier := codegen.NewSeriesSourceClassifier()
	sourceInfo := classifier.Classify("ctx.Data[ctx.BarIndex].Close")
	accessor := codegen.CreateAccessGenerator(sourceInfo)

	code := gen.Generate(accessor, 14, "testhash")

	/* Should generate IIFE wrapper */
	if !strings.Contains(code, "func()") {
		t.Error("generated code should contain IIFE wrapper")
	}

	/* Should use naming strategy */
	if !strings.Contains(code, "_rma_") {
		t.Error("generated code should reference RMA series")
	}

	/* Should include hash in series name */
	if !strings.Contains(code, "testhash") {
		t.Error("generated code should include source hash in series name")
	}

	/* Should include period */
	if !strings.Contains(code, "14") {
		t.Error("generated code should include period")
	}
}

/* TestEMAGenerator_BasicGeneration tests EMA IIFE code generation */
func TestEMAGenerator_BasicGeneration(t *testing.T) {
	namer := series_naming.NewStatefulIndicatorNamer()
	gen := NewEMAGenerator(namer)

	classifier := codegen.NewSeriesSourceClassifier()
	sourceInfo := classifier.Classify("ctx.Data[ctx.BarIndex].Open")
	accessor := codegen.CreateAccessGenerator(sourceInfo)

	code := gen.Generate(accessor, 20, "emahash")

	/* Should generate valid Go code structure */
	if !strings.Contains(code, "func()") {
		t.Error("generated code should contain IIFE wrapper")
	}

	if !strings.Contains(code, "_ema_") {
		t.Error("generated code should reference EMA series")
	}

	if !strings.Contains(code, "emahash") {
		t.Error("generated code should include source hash")
	}
}

/* TestSMAGenerator_BasicGeneration tests SMA IIFE code generation */
func TestSMAGenerator_BasicGeneration(t *testing.T) {
	namer := series_naming.NewStatefulIndicatorNamer()
	gen := NewSMAGenerator(namer)

	classifier := codegen.NewSeriesSourceClassifier()
	sourceInfo := classifier.Classify("ctx.Data[ctx.BarIndex].High")
	accessor := codegen.CreateAccessGenerator(sourceInfo)

	code := gen.Generate(accessor, 50, "smahash")

	if !strings.Contains(code, "func()") {
		t.Error("generated code should contain IIFE wrapper")
	}

	if !strings.Contains(code, "_sma_") {
		t.Error("generated code should reference SMA series")
	}
}

/* TestGenerators_UniquenessAcrossDifferentSources tests collision prevention */
func TestGenerators_UniquenessAcrossDifferentSources(t *testing.T) {
	namer := series_naming.NewStatefulIndicatorNamer()
	gen := NewRMAGenerator(namer)

	classifier := codegen.NewSeriesSourceClassifier()
	sourceInfo := classifier.Classify("ctx.Data[ctx.BarIndex].Close")
	accessor := codegen.CreateAccessGenerator(sourceInfo)

	/* Same period, different source hashes */
	code1 := gen.Generate(accessor, 14, "source1")
	code2 := gen.Generate(accessor, 14, "source2")
	code3 := gen.Generate(accessor, 14, "source3")

	/* All should be different due to different hashes */
	if code1 == code2 {
		t.Error("different source hashes should produce different code")
	}
	if code2 == code3 {
		t.Error("different source hashes should produce different code")
	}
}

/* TestGenerators_DeterministicGeneration tests generation consistency */
func TestGenerators_DeterministicGeneration(t *testing.T) {
	tests := []struct {
		name      string
		generator Generator
	}{
		{"RMA", NewRMAGenerator(series_naming.NewStatefulIndicatorNamer())},
		{"EMA", NewEMAGenerator(series_naming.NewStatefulIndicatorNamer())},
		{"SMA", NewSMAGenerator(series_naming.NewStatefulIndicatorNamer())},
		{"WMA", NewWMAGenerator(series_naming.NewStatefulIndicatorNamer())},
		{"STDEV", NewSTDEVGenerator(series_naming.NewStatefulIndicatorNamer())},
	}

	classifier := codegen.NewSeriesSourceClassifier()
	sourceInfo := classifier.Classify("ctx.Data[ctx.BarIndex].Close")
	accessor := codegen.CreateAccessGenerator(sourceInfo)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			/* Generate same code multiple times */
			code1 := tt.generator.Generate(accessor, 14, "consistency")
			code2 := tt.generator.Generate(accessor, 14, "consistency")
			code3 := tt.generator.Generate(accessor, 14, "consistency")

			/* All should be identical */
			if code1 != code2 {
				t.Errorf("%s generation not deterministic", tt.name)
			}
			if code2 != code3 {
				t.Errorf("%s generation not deterministic", tt.name)
			}
		})
	}
}

/* TestHighestGenerator_WindowBased tests window-based indicator generation */
func TestHighestGenerator_WindowBased(t *testing.T) {
	namer := series_naming.NewWindowBasedNamer()
	gen := NewHighestGenerator(namer)

	classifier := codegen.NewSeriesSourceClassifier()
	sourceInfo := classifier.Classify("ctx.Data[ctx.BarIndex].High")
	accessor := codegen.CreateAccessGenerator(sourceInfo)

	code := gen.Generate(accessor, 10, "shouldbeignored")

	/* Should not include hash (window-based) */
	if strings.Contains(code, "shouldbeignored") {
		t.Error("window-based generator should not include source hash in generated code")
	}

	/* Should contain window logic */
	if !strings.Contains(code, "for j :=") {
		t.Error("highest should use loop-based window logic")
	}

	if !strings.Contains(code, "highest") {
		t.Error("highest logic should reference 'highest' variable")
	}
}

/* TestLowestGenerator_WindowBased tests lowest value generation */
func TestLowestGenerator_WindowBased(t *testing.T) {
	namer := series_naming.NewWindowBasedNamer()
	gen := NewLowestGenerator(namer)

	classifier := codegen.NewSeriesSourceClassifier()
	sourceInfo := classifier.Classify("ctx.Data[ctx.BarIndex].Low")
	accessor := codegen.CreateAccessGenerator(sourceInfo)

	code := gen.Generate(accessor, 5, "ignored")

	if strings.Contains(code, "ignored") {
		t.Error("window-based generator should not include source hash")
	}

	if !strings.Contains(code, "for j :=") {
		t.Error("lowest should use loop-based window logic")
	}

	if !strings.Contains(code, "lowest") {
		t.Error("lowest logic should reference 'lowest' variable")
	}
}

/* TestChangeGenerator_OffsetHandling tests change calculation with offsets */
func TestChangeGenerator_OffsetHandling(t *testing.T) {
	namer := series_naming.NewWindowBasedNamer()
	gen := NewChangeGenerator(namer)

	classifier := codegen.NewSeriesSourceClassifier()
	sourceInfo := classifier.Classify("ctx.Data[ctx.BarIndex].Close")
	accessor := codegen.CreateAccessGenerator(sourceInfo)

	tests := []struct {
		name   string
		offset int
	}{
		{"offset 1", 1},
		{"offset 2", 2},
		{"offset 5", 5},
		{"zero offset", 0},
		{"negative offset", -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			/* Should not panic for any offset */
			code := gen.Generate(accessor, tt.offset, "hash")

			/* Should generate subtraction */
			if !strings.Contains(code, "-") {
				t.Error("change should include subtraction")
			}

			/* Should reference current and previous */
			if !strings.Contains(code, "current") && !strings.Contains(code, "previous") {
				t.Error("change should reference current and previous values")
			}
		})
	}
}

/* TestGenerators_PeriodVariations tests generation across period range */
func TestGenerators_PeriodVariations(t *testing.T) {
	namer := series_naming.NewStatefulIndicatorNamer()
	gen := NewRMAGenerator(namer)

	classifier := codegen.NewSeriesSourceClassifier()
	sourceInfo := classifier.Classify("ctx.Data[ctx.BarIndex].Close")
	accessor := codegen.CreateAccessGenerator(sourceInfo)

	periods := []int{1, 5, 10, 14, 20, 50, 100, 200}
	generated := make(map[string]bool)

	for _, period := range periods {
		code := gen.Generate(accessor, period, "constanthash")

		/* Each period should produce unique code */
		if generated[code] {
			t.Errorf("period %d produced duplicate code", period)
		}
		generated[code] = true
	}
}

/* TestGenerators_NamingStrategyInjection tests dependency injection pattern */
func TestGenerators_NamingStrategyInjection(t *testing.T) {
	/* Test with stateful namer */
	statefulNamer := series_naming.NewStatefulIndicatorNamer()
	statefulGen := NewRMAGenerator(statefulNamer)

	/* Test with window-based namer */
	windowNamer := series_naming.NewWindowBasedNamer()
	windowGen := NewRMAGenerator(windowNamer)

	classifier := codegen.NewSeriesSourceClassifier()
	sourceInfo := classifier.Classify("ctx.Data[ctx.BarIndex].Close")
	accessor := codegen.CreateAccessGenerator(sourceInfo)

	/* Different naming strategies should produce different results */
	code1 := statefulGen.Generate(accessor, 14, "testhash")
	code2 := windowGen.Generate(accessor, 14, "testhash")

	/* Stateful should include hash, window should not */
	hasHash1 := strings.Contains(code1, "testhash")
	hasHash2 := strings.Contains(code2, "testhash")

	if !hasHash1 {
		t.Error("stateful namer should include hash in generated code")
	}
	if hasHash2 {
		t.Error("window namer should not include hash in generated code")
	}
}

/* TestGenerators_AccessorIntegration tests accessor pattern usage */
func TestGenerators_AccessorIntegration(t *testing.T) {
	namer := series_naming.NewStatefulIndicatorNamer()
	gen := NewRMAGenerator(namer)

	classifier := codegen.NewSeriesSourceClassifier()

	accessorTests := []struct {
		name   string
		source string
	}{
		{"close field", "ctx.Data[ctx.BarIndex].Close"},
		{"open field", "ctx.Data[ctx.BarIndex].Open"},
		{"high field", "ctx.Data[ctx.BarIndex].High"},
		{"low field", "ctx.Data[ctx.BarIndex].Low"},
	}

	for _, tt := range accessorTests {
		t.Run(tt.name, func(t *testing.T) {
			sourceInfo := classifier.Classify(tt.source)
			accessor := codegen.CreateAccessGenerator(sourceInfo)

			/* Should generate valid code for any accessor */
			code := gen.Generate(accessor, 14, "hash")

			if code == "" {
				t.Error("generator should produce non-empty code")
			}

			if !strings.Contains(code, "func()") {
				t.Error("generated code should be valid IIFE")
			}
		})
	}
}

/* TestGenerators_EmptyHashHandling tests behavior with empty hash */
func TestGenerators_EmptyHashHandling(t *testing.T) {
	namer := series_naming.NewStatefulIndicatorNamer()
	gen := NewRMAGenerator(namer)

	classifier := codegen.NewSeriesSourceClassifier()
	sourceInfo := classifier.Classify("ctx.Data[ctx.BarIndex].Close")
	accessor := codegen.CreateAccessGenerator(sourceInfo)

	/* Should handle empty hash gracefully */
	code := gen.Generate(accessor, 14, "")

	if code == "" {
		t.Error("generator should produce code even with empty hash")
	}

	/* Should be deterministic */
	code2 := gen.Generate(accessor, 14, "")
	if code != code2 {
		t.Error("generation with empty hash should be deterministic")
	}
}

/* TestGenerators_InterfaceCompliance tests all generators implement interface */
func TestGenerators_InterfaceCompliance(t *testing.T) {
	namer := series_naming.NewStatefulIndicatorNamer()

	/* Verify all generators implement Generator interface */
	var _ Generator = NewRMAGenerator(namer)
	var _ Generator = NewEMAGenerator(namer)
	var _ Generator = NewSMAGenerator(namer)
	var _ Generator = NewWMAGenerator(namer)
	var _ Generator = NewSTDEVGenerator(namer)
	var _ Generator = NewHighestGenerator(namer)
	var _ Generator = NewLowestGenerator(namer)
	var _ Generator = NewChangeGenerator(namer)
}
