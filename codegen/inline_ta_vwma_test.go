package codegen

import (
	"strings"
	"testing"
)

func TestVWMAIIFEGenerator_Generate(t *testing.T) {
	registry := NewInlineTAIIFERegistry()

	accessor := NewArrowFunctionParameterAccessor("close")
	period := NewConstantPeriod(14)

	code, ok := registry.Generate("ta.vwma", accessor, period, "test_source")
	if !ok {
		t.Fatal("ta.vwma generator not found")
	}

	if code == "" {
		t.Fatal("Generated empty code")
	}

	expectedPatterns := []string{
		"weightedSum := 0.0",
		"volumeSum := 0.0",
		"for j := 0; j < 14; j++",
		"closeSeries.Get(j)",
		"ctx.Data[ctx.BarIndex-j].Volume",
		"math.IsNaN",
		"weightedSum / volumeSum",
	}

	for _, pattern := range expectedPatterns {
		if !strings.Contains(code, pattern) {
			t.Errorf("Generated code missing pattern: %s\nCode:\n%s", pattern, code)
		}
	}
}

func TestVWMAIIFEGenerator_DynamicPeriod(t *testing.T) {
	registry := NewInlineTAIIFERegistry()

	accessor := NewArrowFunctionParameterAccessor("high")
	period := NewRuntimePeriod("myPeriod")

	code, ok := registry.Generate("vwma", accessor, period, "test_source")
	if !ok {
		t.Fatal("vwma generator not found")
	}

	expectedPatterns := []string{
		"for j := 0; j < int(myPeriod); j++",
		"highSeries.Get(j)",
		"ctx.Data[ctx.BarIndex-j].Volume",
	}

	for _, pattern := range expectedPatterns {
		if !strings.Contains(code, pattern) {
			t.Errorf("Generated code missing pattern: %s\nCode:\n%s", pattern, code)
		}
	}
}

func TestVWMAIIFEGenerator_NaNCheck(t *testing.T) {
	registry := NewInlineTAIIFERegistry()

	accessor := NewArrowFunctionParameterAccessor("close")
	period := NewConstantPeriod(20)

	code, _ := registry.Generate("ta.vwma", accessor, period, "test_source")

	if !strings.Contains(code, "math.IsNaN(val)") {
		t.Errorf("Generated code must include NaN check")
	}

	if !strings.Contains(code, "!math.IsNaN(val)") {
		t.Errorf("Generated code must skip NaN values in accumulation")
	}
}

func TestVWMAIIFEGenerator_RegistryIntegration(t *testing.T) {
	registry := NewInlineTAIIFERegistry()

	tests := []string{"ta.vwma", "vwma"}
	for _, funcName := range tests {
		if !registry.IsSupported(funcName) {
			t.Errorf("Registry missing generator for %s", funcName)
		}
	}
}
