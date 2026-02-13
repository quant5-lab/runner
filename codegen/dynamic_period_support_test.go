package codegen

import (
	"strings"
	"testing"
)

func TestSupportsDynamicPeriod_NamespacedAndBareForms(t *testing.T) {
	for funcName := range dynamicPeriodDispatch {
		bare := strings.TrimPrefix(funcName, "ta.")
		t.Run(funcName, func(t *testing.T) {
			if !SupportsDynamicPeriod(funcName) {
				t.Errorf("SupportsDynamicPeriod(%q) = false, want true", funcName)
			}
		})
		t.Run(bare, func(t *testing.T) {
			if !SupportsDynamicPeriod(bare) {
				t.Errorf("SupportsDynamicPeriod(%q) = false, want true", bare)
			}
		})
	}
}

func TestSupportsDynamicPeriod_Unsupported(t *testing.T) {
	tests := []struct {
		name     string
		funcName string
	}{
		{"ta.wma", "ta.wma"},
		{"ta.rma", "ta.rma"},
		{"ta.linreg", "ta.linreg"},
		{"ta.sum", "ta.sum"},
		{"ta.crossover", "ta.crossover"},
		{"ta.crossunder", "ta.crossunder"},
		{"ta.change", "ta.change"},
		{"math.abs", "math.abs"},
		{"strategy.entry", "strategy.entry"},
		{"empty string", ""},
		{"ta prefix only", "ta."},
		{"not a function", "notAFunction"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if SupportsDynamicPeriod(tt.funcName) {
				t.Errorf("SupportsDynamicPeriod(%q) = true, want false", tt.funcName)
			}
		})
	}
}
