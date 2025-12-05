package codegen

import "testing"

func TestInlineTAIIFERegistry_IsSupported(t *testing.T) {
	registry := NewInlineTAIIFERegistry()

	tests := []struct {
		name     string
		funcName string
		want     bool
	}{
		{"ta.sma", "ta.sma", true},
		{"sma", "sma", true},
		{"ta.ema", "ta.ema", true},
		{"ema", "ema", true},
		{"ta.rma", "ta.rma", true},
		{"rma", "rma", true},
		{"ta.wma", "ta.wma", true},
		{"wma", "wma", true},
		{"ta.stdev", "ta.stdev", true},
		{"stdev", "stdev", true},
		{"unsupported", "ta.unsupported", false},
		{"random", "random_func", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := registry.IsSupported(tt.funcName)
			if got != tt.want {
				t.Errorf("IsSupported(%q) = %v, want %v", tt.funcName, got, tt.want)
			}
		})
	}
}

func TestInlineTAIIFERegistry_Generate(t *testing.T) {
	registry := NewInlineTAIIFERegistry()
	classifier := NewSeriesSourceClassifier()
	sourceInfo := classifier.Classify("ctx.Data[ctx.BarIndex].Close")
	accessor := CreateAccessGenerator(sourceInfo)

	tests := []struct {
		name     string
		funcName string
		period   int
		wantOk   bool
	}{
		{"sma_20", "ta.sma", 20, true},
		{"ema_10", "ta.ema", 10, true},
		{"rma_14", "ta.rma", 14, true},
		{"wma_9", "ta.wma", 9, true},
		{"stdev_20", "ta.stdev", 20, true},
		{"unsupported", "ta.unsupported", 10, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, ok := registry.Generate(tt.funcName, accessor, tt.period)
			if ok != tt.wantOk {
				t.Errorf("Generate(%q) ok = %v, want %v", tt.funcName, ok, tt.wantOk)
			}
			if ok && code == "" {
				t.Errorf("Generate(%q) returned empty code", tt.funcName)
			}
			if !ok && code != "" {
				t.Errorf("Generate(%q) returned code when it should not: %q", tt.funcName, code)
			}
		})
	}
}

func TestInlineTAIIFERegistry_CustomRegistration(t *testing.T) {
	registry := NewInlineTAIIFERegistry()

	customGen := &SMAIIFEGenerator{}
	registry.Register("custom.ta", customGen)

	if !registry.IsSupported("custom.ta") {
		t.Error("Custom registration failed: not supported")
	}

	classifier := NewSeriesSourceClassifier()
	sourceInfo := classifier.Classify("ctx.Data[ctx.BarIndex].Close")
	accessor := CreateAccessGenerator(sourceInfo)

	code, ok := registry.Generate("custom.ta", accessor, 10)
	if !ok {
		t.Error("Custom generator not executed")
	}
	if code == "" {
		t.Error("Custom generator returned empty code")
	}
}
