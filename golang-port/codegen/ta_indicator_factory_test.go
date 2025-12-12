package codegen

import (
	"strings"
	"testing"
)

func TestTAIndicatorFactory_CreateBuilder_SMA(t *testing.T) {
	factory := NewTAIndicatorFactory()
	accessor := &MockAccessGenerator{
		loopAccessFn: func(loopVar string) string {
			return "data.Get(" + loopVar + ")"
		},
	}

	builder, err := factory.CreateBuilder("ta.sma", "sma20", 20, accessor)
	if err != nil {
		t.Fatalf("Failed to create SMA builder: %v", err)
	}

	if builder == nil {
		t.Fatal("Builder is nil")
	}

	// Verify builder generates code
	code := builder.Build()
	if code == "" {
		t.Error("Generated code is empty")
	}

	// Check for key SMA elements
	if !strings.Contains(code, "sum := 0.0") {
		t.Error("SMA code missing sum initialization")
	}

	if !strings.Contains(code, "sum / 20.0") {
		t.Error("SMA code missing average calculation")
	}
}

func TestTAIndicatorFactory_CreateBuilder_EMA(t *testing.T) {
	factory := NewTAIndicatorFactory()
	accessor := &MockAccessGenerator{
		loopAccessFn: func(loopVar string) string {
			return "data.Get(" + loopVar + ")"
		},
	}

	builder, err := factory.CreateBuilder("ta.ema", "ema20", 20, accessor)
	if err != nil {
		t.Fatalf("Failed to create EMA builder: %v", err)
	}

	if builder == nil {
		t.Fatal("Builder is nil")
	}

	// Verify builder generates code
	code := builder.Build()
	if code == "" {
		t.Error("Generated code is empty")
	}

	// Check for key EMA elements
	if !strings.Contains(code, "alpha") {
		t.Error("EMA code missing alpha calculation")
	}
}

func TestTAIndicatorFactory_CreateBuilder_UnsupportedIndicator(t *testing.T) {
	factory := NewTAIndicatorFactory()
	accessor := &MockAccessGenerator{
		loopAccessFn: func(loopVar string) string {
			return "data.Get(" + loopVar + ")"
		},
	}

	_, err := factory.CreateBuilder("ta.macd", "macd", 20, accessor)
	if err == nil {
		t.Error("Expected error for unsupported indicator")
	}

	if !strings.Contains(err.Error(), "unsupported indicator type") {
		t.Errorf("Unexpected error message: %v", err)
	}
}

func TestTAIndicatorFactory_CreateSTDEVBuilders(t *testing.T) {
	factory := NewTAIndicatorFactory()
	accessor := &MockAccessGenerator{
		loopAccessFn: func(loopVar string) string {
			return "data.Get(" + loopVar + ")"
		},
	}

	meanBuilder, varianceBuilder, err := factory.CreateSTDEVBuilders("stdev20", 20, accessor)
	if err != nil {
		t.Fatalf("Failed to create STDEV builders: %v", err)
	}

	if meanBuilder == nil {
		t.Fatal("Mean builder is nil")
	}

	if varianceBuilder == nil {
		t.Fatal("Variance builder is nil")
	}

	// Verify mean builder uses SumAccumulator
	meanCode := meanBuilder.Build()
	if !strings.Contains(meanCode, "sum := 0.0") {
		t.Error("Mean builder missing sum initialization")
	}

	// Verify variance builder uses VarianceAccumulator
	varianceCode := varianceBuilder.Build()
	if !strings.Contains(varianceCode, "variance := 0.0") {
		t.Error("Variance builder missing variance initialization")
	}

	if !strings.Contains(varianceCode, "diff") {
		t.Error("Variance builder missing diff calculation")
	}
}

func TestTAIndicatorFactory_ShouldCheckNaN_SeriesVariable(t *testing.T) {
	factory := NewTAIndicatorFactory()

	// Create a Series variable accessor using the existing classifier
	classifier := NewSeriesSourceClassifier()
	sourceInfo := classifier.Classify("sma20Series.Get(0)")
	seriesAccessor := CreateAccessGenerator(sourceInfo)

	needsNaN := factory.shouldCheckNaN(seriesAccessor)
	if !needsNaN {
		t.Error("Should check NaN for Series variables")
	}
}

func TestTAIndicatorFactory_ShouldCheckNaN_OHLCV(t *testing.T) {
	factory := NewTAIndicatorFactory()

	// Create an OHLCV accessor using the existing classifier
	classifier := NewSeriesSourceClassifier()
	sourceInfo := classifier.Classify("close")
	ohlcvAccessor := CreateAccessGenerator(sourceInfo)

	needsNaN := factory.shouldCheckNaN(ohlcvAccessor)
	if needsNaN {
		t.Error("Should not check NaN for OHLCV fields")
	}
}

func TestTAIndicatorFactory_SupportedIndicators(t *testing.T) {
	factory := NewTAIndicatorFactory()

	supported := factory.SupportedIndicators()
	if len(supported) == 0 {
		t.Error("No supported indicators returned")
	}

	// Check that expected indicators are present
	expectedIndicators := []string{"ta.sma", "ta.ema", "ta.stdev"}
	for _, expected := range expectedIndicators {
		found := false
		for _, actual := range supported {
			if actual == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected indicator %s not in supported list", expected)
		}
	}
}

func TestTAIndicatorFactory_IsSupported(t *testing.T) {
	factory := NewTAIndicatorFactory()

	tests := []struct {
		indicator string
		supported bool
	}{
		{"ta.sma", true},
		{"ta.ema", true},
		{"ta.stdev", true},
		{"ta.macd", false},
		{"ta.rsi", false},
		{"sma", false},
	}

	for _, tt := range tests {
		t.Run(tt.indicator, func(t *testing.T) {
			result := factory.IsSupported(tt.indicator)
			if result != tt.supported {
				t.Errorf("IsSupported(%s) = %v, want %v", tt.indicator, result, tt.supported)
			}
		})
	}
}

func TestTAIndicatorFactory_Integration_SMA(t *testing.T) {
	factory := NewTAIndicatorFactory()
	classifier := NewSeriesSourceClassifier()
	sourceInfo := classifier.Classify("close")
	accessor := CreateAccessGenerator(sourceInfo)

	builder, err := factory.CreateBuilder("ta.sma", "sma50", 50, accessor)
	if err != nil {
		t.Fatalf("Failed to create builder: %v", err)
	}

	code := builder.Build()

	// Debug: print generated code
	t.Logf("Generated SMA code:\n%s", code)

	// Verify complete SMA code structure
	requiredElements := []string{
		"ta.sma(50)",
		"ctx.BarIndex < 50-1",
		"sma50Series.Set(math.NaN())",
		"sum := 0.0",
		"for j := 0; j < 50; j++",
		"sum / 50.0",
	}

	for _, elem := range requiredElements {
		if !strings.Contains(code, elem) {
			t.Errorf("SMA code missing: %s", elem)
		}
	}
}

func TestTAIndicatorFactory_Integration_EMA(t *testing.T) {
	factory := NewTAIndicatorFactory()
	classifier := NewSeriesSourceClassifier()
	sourceInfo := classifier.Classify("close")
	accessor := CreateAccessGenerator(sourceInfo)

	builder, err := factory.CreateBuilder("ta.ema", "ema21", 21, accessor)
	if err != nil {
		t.Fatalf("Failed to create builder: %v", err)
	}

	code := builder.Build()

	// Verify complete EMA code structure
	requiredElements := []string{
		"ta.ema(21)",
		"ctx.BarIndex < 21-1",
		"ema21Series.Set(math.NaN())",
		"alpha := 2.0 / float64(21+1)",
	}

	for _, elem := range requiredElements {
		if !strings.Contains(code, elem) {
			t.Errorf("EMA code missing: %s", elem)
		}
	}
}

func TestTAIndicatorFactory_Integration_STDEV(t *testing.T) {
	factory := NewTAIndicatorFactory()
	classifier := NewSeriesSourceClassifier()
	sourceInfo := classifier.Classify("close")
	accessor := CreateAccessGenerator(sourceInfo)

	meanBuilder, varianceBuilder, err := factory.CreateSTDEVBuilders("stdev30", 30, accessor)
	if err != nil {
		t.Fatalf("Failed to create STDEV builders: %v", err)
	}

	meanCode := meanBuilder.Build()
	varianceCode := varianceBuilder.Build()

	// Verify mean calculation
	if !strings.Contains(meanCode, "sum := 0.0") {
		t.Error("Mean code missing sum initialization")
	}

	if !strings.Contains(meanCode, "for j := 0; j < 30; j++") {
		t.Error("Mean code missing loop")
	}

	// Verify variance calculation
	if !strings.Contains(varianceCode, "variance := 0.0") {
		t.Error("Variance code missing variance initialization")
	}

	if !strings.Contains(varianceCode, "diff") {
		t.Error("Variance code missing diff calculation")
	}

	if !strings.Contains(varianceCode, "variance += diff * diff") {
		t.Error("Variance code missing variance accumulation")
	}
}
