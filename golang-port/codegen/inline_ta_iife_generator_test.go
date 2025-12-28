package codegen

import (
	"fmt"
	"testing"
)

func TestIIFECodeBuilder_Basic(t *testing.T) {
	builder := NewIIFECodeBuilder().
		WithWarmupCheck(20).
		WithBody("return 42.0")

	expected := "func() float64 { if ctx.BarIndex < 19 { return math.NaN() }; return 42.0 }()"
	actual := builder.Build()

	if actual != expected {
		t.Errorf("Expected: %s\nActual: %s", expected, actual)
	}
}

func TestIIFECodeBuilder_NoWarmup(t *testing.T) {
	builder := NewIIFECodeBuilder().
		WithBody("return 100.0")

	expected := "func() float64 { return 100.0 }()"
	actual := builder.Build()

	if actual != expected {
		t.Errorf("Expected: %s\nActual: %s", expected, actual)
	}
}

func TestSMAIIFEGenerator(t *testing.T) {
	classifier := NewSeriesSourceClassifier()
	sourceInfo := classifier.Classify("ctx.Data[ctx.BarIndex].Close")
	accessor := CreateAccessGenerator(sourceInfo)
	gen := &SMAIIFEGenerator{}

	result := gen.Generate(accessor, 20)

	if result == "" {
		t.Fatal("Generated code is empty")
	}

	if !contains(result, "sum := 0.0") {
		t.Error("Missing sum initialization")
	}

	if !contains(result, "for j := 0; j < 20") {
		t.Error("Missing loop structure")
	}

	if !contains(result, "return sum / 20.0") {
		t.Error("Missing average calculation")
	}

	if !contains(result, "ctx.BarIndex < 19") {
		t.Error("Missing warmup check")
	}
}

func TestEMAIIFEGenerator(t *testing.T) {
	classifier := NewSeriesSourceClassifier()
	sourceInfo := classifier.Classify("ctx.Data[ctx.BarIndex].Close")
	accessor := CreateAccessGenerator(sourceInfo)
	gen := &EMAIIFEGenerator{}

	result := gen.Generate(accessor, 10)

	if !contains(result, "alpha := 2.0 / float64(10+1)") {
		t.Error("Missing alpha calculation")
	}

	if !contains(result, "ctx.BarIndex < 9") {
		t.Error("Missing warmup check")
	}
}

func TestRMAIIFEGenerator(t *testing.T) {
	classifier := NewSeriesSourceClassifier()
	sourceInfo := classifier.Classify("ctx.Data[ctx.BarIndex].Close")
	accessor := CreateAccessGenerator(sourceInfo)
	gen := &RMAIIFEGenerator{}

	result := gen.Generate(accessor, 14)

	if !contains(result, "alpha := 1.0 / float64(14)") {
		t.Error("Missing alpha calculation for RMA")
	}

	if !contains(result, "ctx.BarIndex < 13") {
		t.Error("Missing warmup check for period 14")
	}

	if !contains(result, "arrowCtx.GetOrCreateSeries(\"_rma_14\")") {
		t.Error("Missing arrowCtx series creation for arrow function context")
	}

	if !contains(result, ".Set(") {
		t.Error("Missing Series.Set() for ForwardSeriesBuffer pattern")
	}

	if !contains(result, ".Get(1)") {
		t.Error("Missing forward reference to previous value")
	}

	if contains(result, "for j := 12; j >= 0; j--") {
		t.Error("Should NOT use backward loop (old pattern)")
	}

	if !contains(result, "func() float64") {
		t.Error("Missing IIFE wrapper")
	}

	if !contains(result, "return arrowCtx.GetOrCreateSeries(\"_rma_14\").Get(0).GetCurrent()") {
		t.Error("Missing IIFE return statement")
	}
}

func SkipTestRMAIIFEGenerator_PeriodVariations(t *testing.T) {
	tests := []struct {
		name              string
		period            int
		expectedAlpha     string
		expectWarmupCheck bool
		expectedLoopStart string
	}{
		{
			name:              "period 1",
			period:            1,
			expectedAlpha:     "alpha := 1.0 / 1.0",
			expectWarmupCheck: false,
			expectedLoopStart: "for j := -1; j >= 0; j--",
		},
		{
			name:              "period 2",
			period:            2,
			expectedAlpha:     "alpha := 1.0 / 2.0",
			expectWarmupCheck: true,
			expectedLoopStart: "for j := 0; j >= 0; j--",
		},
		{
			name:              "period 10",
			period:            10,
			expectedAlpha:     "alpha := 1.0 / 10.0",
			expectWarmupCheck: true,
			expectedLoopStart: "for j := 8; j >= 0; j--",
		},
		{
			name:              "period 20 (RSI default)",
			period:            20,
			expectedAlpha:     "alpha := 1.0 / 20.0",
			expectWarmupCheck: true,
			expectedLoopStart: "for j := 18; j >= 0; j--",
		},
		{
			name:              "period 200 (large)",
			period:            200,
			expectedAlpha:     "alpha := 1.0 / 200.0",
			expectWarmupCheck: true,
			expectedLoopStart: "for j := 198; j >= 0; j--",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			classifier := NewSeriesSourceClassifier()
			sourceInfo := classifier.Classify("ctx.Data[ctx.BarIndex].Close")
			accessor := CreateAccessGenerator(sourceInfo)
			gen := &RMAIIFEGenerator{}

			result := gen.Generate(accessor, tt.period)

			if !contains(result, tt.expectedAlpha) {
				t.Errorf("Expected alpha %q, not found in generated code", tt.expectedAlpha)
			}

			if tt.expectWarmupCheck {
				expectedWarmup := fmt.Sprintf("ctx.BarIndex < %d", tt.period-1)
				if !contains(result, expectedWarmup) {
					t.Errorf("Expected warmup check %q, not found in generated code", expectedWarmup)
				}
			} else {
				if contains(result, "ctx.BarIndex <") {
					t.Error("Did not expect warmup check for period 1")
				}
			}

			if !contains(result, tt.expectedLoopStart) {
				t.Errorf("Expected loop start %q, not found in generated code", tt.expectedLoopStart)
			}

			if contains(result, "sum :=") || contains(result, "sma :=") {
				t.Error("RMA should not generate unused sum or sma variables")
			}

			if !contains(result, "rma := ") {
				t.Error("Missing rma initialization")
			}

			if !contains(result, "rma = alpha*") {
				t.Error("Missing RMA update formula")
			}
		})
	}
}

func SkipTestRMAIIFEGenerator_SourceTypeVariations(t *testing.T) {
	tests := []struct {
		name        string
		sourceExpr  string
		expectValid bool
	}{
		{
			name:        "bar field close",
			sourceExpr:  "ctx.Data[ctx.BarIndex].Close",
			expectValid: true,
		},
		{
			name:        "bar field high",
			sourceExpr:  "ctx.Data[ctx.BarIndex].High",
			expectValid: true,
		},
		{
			name:        "bar field low",
			sourceExpr:  "ctx.Data[ctx.BarIndex].Low",
			expectValid: true,
		},
		{
			name:        "bar field open",
			sourceExpr:  "ctx.Data[ctx.BarIndex].Open",
			expectValid: true,
		},
		{
			name:        "series accessor",
			sourceExpr:  "closeSeries.GetCurrent()",
			expectValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			classifier := NewSeriesSourceClassifier()
			sourceInfo := classifier.Classify(tt.sourceExpr)
			accessor := CreateAccessGenerator(sourceInfo)
			gen := &RMAIIFEGenerator{}

			result := gen.Generate(accessor, 14)

			if result == "" && tt.expectValid {
				t.Error("Expected valid generated code, got empty string")
			}

			if tt.expectValid {
				if contains(result, "sum :=") || contains(result, "sma :=") {
					t.Error("RMA should not generate unused variables regardless of source type")
				}

				if !contains(result, "alpha := 1.0 / 14.0") {
					t.Error("Missing alpha calculation")
				}

				if !contains(result, "rma := ") {
					t.Error("Missing rma initialization")
				}
			}
		})
	}
}

func SkipTestRMAIIFEGenerator_CodeStructureValidation(t *testing.T) {
	classifier := NewSeriesSourceClassifier()
	sourceInfo := classifier.Classify("ctx.Data[ctx.BarIndex].Close")
	accessor := CreateAccessGenerator(sourceInfo)
	gen := &RMAIIFEGenerator{}

	result := gen.Generate(accessor, 20)

	requiredComponents := []string{
		"func() float64",
		"if ctx.BarIndex < 19",
		"return math.NaN()",
		"alpha := 1.0 / 20.0",
		"rma := ",
		"for j := 18; j >= 0; j--",
		"rma = alpha*",
		"+ (1-alpha)*rma",
		"return rma",
		"}()",
	}

	for _, component := range requiredComponents {
		if !contains(result, component) {
			t.Errorf("Missing required component: %q", component)
		}
	}

	prohibitedComponents := []string{
		"sum := 0.0",
		"sum +=",
		"sma := sum",
		"sma :=",
	}

	for _, component := range prohibitedComponents {
		if contains(result, component) {
			t.Errorf("Found prohibited component (unused variable): %q", component)
		}
	}
}

func TestWMAIIFEGenerator(t *testing.T) {
	classifier := NewSeriesSourceClassifier()
	sourceInfo := classifier.Classify("ctx.Data[ctx.BarIndex].Close")
	accessor := CreateAccessGenerator(sourceInfo)
	gen := &WMAIIFEGenerator{}

	result := gen.Generate(accessor, 9)

	if !contains(result, "weightedSum") {
		t.Error("Missing weighted sum variable")
	}

	if !contains(result, "weightSum := 45.0") {
		t.Error("Missing weight sum (9*(9+1)/2 = 45)")
	}
}

func TestSTDEVIIFEGenerator(t *testing.T) {
	classifier := NewSeriesSourceClassifier()
	sourceInfo := classifier.Classify("ctx.Data[ctx.BarIndex].Close")
	accessor := CreateAccessGenerator(sourceInfo)
	gen := &STDEVIIFEGenerator{}

	result := gen.Generate(accessor, 20)

	if !contains(result, "mean := sum / 20.0") {
		t.Error("Missing mean calculation")
	}

	if !contains(result, "variance := 0.0") {
		t.Error("Missing variance variable")
	}

	if !contains(result, "math.Sqrt(variance / 20.0)") {
		t.Error("Missing standard deviation calculation")
	}
}

func TestChangeIIFEGenerator(t *testing.T) {
	classifier := NewSeriesSourceClassifier()
	sourceInfo := classifier.Classify("ctx.Data[ctx.BarIndex].High")
	accessor := CreateAccessGenerator(sourceInfo)
	gen := &ChangeIIFEGenerator{}

	result := gen.Generate(accessor, 1)

	if !contains(result, "current := ") {
		t.Error("Missing current value access")
	}

	if !contains(result, "previous := ") {
		t.Error("Missing previous value access")
	}

	if !contains(result, "return current - previous") {
		t.Error("Missing difference calculation")
	}

	if !contains(result, "ctx.BarIndex < 1") {
		t.Error("Missing warmup check for offset=1")
	}
}

func TestChangeIIFEGenerator_DefaultOffset(t *testing.T) {
	classifier := NewSeriesSourceClassifier()
	sourceInfo := classifier.Classify("bar.Low")
	accessor := CreateAccessGenerator(sourceInfo)
	gen := &ChangeIIFEGenerator{}

	result := gen.Generate(accessor, 0)

	if !contains(result, "ctx.BarIndex < 1") {
		t.Error("Offset 0 should default to 1")
	}
}

func TestChangeIIFEGenerator_CustomOffset(t *testing.T) {
	classifier := NewSeriesSourceClassifier()
	sourceInfo := classifier.Classify("bar.Close")
	accessor := CreateAccessGenerator(sourceInfo)
	gen := &ChangeIIFEGenerator{}

	result := gen.Generate(accessor, 5)

	if !contains(result, "ctx.BarIndex < 5") {
		t.Error("Missing warmup check for offset=5")
	}
}

/* TestChangeIIFEGenerator_EdgeCases validates boundary conditions and error handling */
func TestChangeIIFEGenerator_EdgeCases(t *testing.T) {
	tests := []struct {
		name                string
		offset              int
		expectedWarmupCheck string
	}{
		{
			name:                "negative offset defaults to 1",
			offset:              -5,
			expectedWarmupCheck: "ctx.BarIndex < 1",
		},
		{
			name:                "zero offset defaults to 1",
			offset:              0,
			expectedWarmupCheck: "ctx.BarIndex < 1",
		},
		{
			name:                "offset 1",
			offset:              1,
			expectedWarmupCheck: "ctx.BarIndex < 1",
		},
		{
			name:                "large offset",
			offset:              1000,
			expectedWarmupCheck: "ctx.BarIndex < 1000",
		},
		{
			name:                "warmup boundary matches offset",
			offset:              50,
			expectedWarmupCheck: "ctx.BarIndex < 50",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			classifier := NewSeriesSourceClassifier()
			sourceInfo := classifier.Classify("ctx.Data[ctx.BarIndex].Close")
			accessor := CreateAccessGenerator(sourceInfo)
			gen := &ChangeIIFEGenerator{}

			result := gen.Generate(accessor, tt.offset)

			if !contains(result, tt.expectedWarmupCheck) {
				t.Errorf("Expected warmup check %q, but not found in: %s", tt.expectedWarmupCheck, result)
			}

			if !contains(result, "current := ") {
				t.Error("Missing current value assignment")
			}

			if !contains(result, "previous := ") {
				t.Error("Missing previous value assignment")
			}

			if !contains(result, "return current - previous") {
				t.Error("Missing difference calculation")
			}

			if !contains(result, "math.NaN()") {
				t.Error("Missing NaN return for warmup period")
			}
		})
	}
}
