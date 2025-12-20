package codegen

import (
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

	if !contains(result, "alpha := 1.0 / 14.0") {
		t.Error("Missing alpha calculation for RMA")
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
