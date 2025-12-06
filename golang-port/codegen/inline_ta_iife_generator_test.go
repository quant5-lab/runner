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
