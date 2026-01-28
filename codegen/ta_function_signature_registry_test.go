package codegen

import (
	"testing"
)

func TestTAFunctionSignatureRegistry_ArgumentPatternClassification(t *testing.T) {
	registry := NewTAFunctionSignatureRegistry()

	tests := []struct {
		name            string
		functionName    string
		expectedPattern TAArgumentPattern
		expectedSource  string
	}{
		{"highest unprefixed", "highest", TAPatternSingleArgIsLength, "high"},
		{"highest prefixed", "ta.highest", TAPatternSingleArgIsLength, "high"},
		{"lowest unprefixed", "lowest", TAPatternSingleArgIsLength, "low"},
		{"lowest prefixed", "ta.lowest", TAPatternSingleArgIsLength, "low"},
		{"highestbars unprefixed", "highestbars", TAPatternSingleArgIsLength, "high"},
		{"highestbars prefixed", "ta.highestbars", TAPatternSingleArgIsLength, "high"},
		{"lowestbars unprefixed", "lowestbars", TAPatternSingleArgIsLength, "low"},
		{"lowestbars prefixed", "ta.lowestbars", TAPatternSingleArgIsLength, "low"},
		{"change unprefixed", "change", TAPatternSingleArgIsSource, ""},
		{"change prefixed", "ta.change", TAPatternSingleArgIsSource, ""},
		{"sma unprefixed", "sma", TAPatternExplicitSourceAndLength, ""},
		{"sma prefixed", "ta.sma", TAPatternExplicitSourceAndLength, ""},
		{"ema unprefixed", "ema", TAPatternExplicitSourceAndLength, ""},
		{"ema prefixed", "ta.ema", TAPatternExplicitSourceAndLength, ""},
		{"rma unprefixed", "rma", TAPatternExplicitSourceAndLength, ""},
		{"rma prefixed", "ta.rma", TAPatternExplicitSourceAndLength, ""},
		{"wma unprefixed", "wma", TAPatternExplicitSourceAndLength, ""},
		{"wma prefixed", "ta.wma", TAPatternExplicitSourceAndLength, ""},
		{"stdev unprefixed", "stdev", TAPatternExplicitSourceAndLength, ""},
		{"stdev prefixed", "ta.stdev", TAPatternExplicitSourceAndLength, ""},
		{"rsi unprefixed", "rsi", TAPatternExplicitSourceAndLength, ""},
		{"rsi prefixed", "ta.rsi", TAPatternExplicitSourceAndLength, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sig, exists := registry.GetSignature(tt.functionName)
			if !exists {
				t.Fatalf("GetSignature(%q) returned exists=false", tt.functionName)
			}

			if sig.ArgumentPattern != tt.expectedPattern {
				t.Errorf("GetSignature(%q).ArgumentPattern = %v, want %v",
					tt.functionName, sig.ArgumentPattern, tt.expectedPattern)
			}

			if sig.DefaultSource != tt.expectedSource {
				t.Errorf("GetSignature(%q).DefaultSource = %q, want %q",
					tt.functionName, sig.DefaultSource, tt.expectedSource)
			}
		})
	}
}

func TestTAFunctionSignatureRegistry_UnknownFunctionHandling(t *testing.T) {
	registry := NewTAFunctionSignatureRegistry()

	unknownFunctions := []string{
		"unknown_function",
		"custom_indicator",
		"",
		"ta.nonexistent",
		"strategy.entry",
		"plot",
		"math.abs",
	}

	for _, funcName := range unknownFunctions {
		t.Run(funcName, func(t *testing.T) {
			sig, exists := registry.GetSignature(funcName)
			if exists {
				t.Errorf("GetSignature(%q) returned exists=true for unknown function", funcName)
			}

			if sig.DefaultSource != "" || sig.ArgumentPattern != 0 {
				t.Errorf("GetSignature(%q) returned non-zero signature: %+v", funcName, sig)
			}
		})
	}
}

func TestTAFunctionSignatureRegistry_HasOptionalSourceBehavior(t *testing.T) {
	registry := NewTAFunctionSignatureRegistry()

	tests := []struct {
		functionName string
		hasOptional  bool
	}{
		{"highest", true},
		{"ta.highest", true},
		{"lowest", true},
		{"ta.lowest", true},
		{"highestbars", true},
		{"ta.highestbars", true},
		{"lowestbars", true},
		{"ta.lowestbars", true},
		{"change", false},
		{"ta.change", false},
		{"sma", false},
		{"ta.sma", false},
		{"ema", false},
		{"rma", false},
		{"wma", false},
		{"stdev", false},
		{"rsi", false},
		{"unknown", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.functionName, func(t *testing.T) {
			result := registry.HasOptionalSource(tt.functionName)
			if result != tt.hasOptional {
				t.Errorf("HasOptionalSource(%q) = %v, want %v",
					tt.functionName, result, tt.hasOptional)
			}
		})
	}
}
