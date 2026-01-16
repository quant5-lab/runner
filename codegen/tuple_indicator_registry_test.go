package codegen

import (
	"testing"
)

func TestTupleIndicatorRegistry_MACDRegistration(t *testing.T) {
	registry := NewTupleIndicatorRegistry()

	spec := registry.Lookup("ta.macd")
	if spec == nil {
		t.Fatal("ta.macd not registered")
	}

	if spec.OutputCount != 3 {
		t.Errorf("Expected 3 outputs, got %d", spec.OutputCount)
	}

	if spec.RuntimeFunction != "ta.Macd" {
		t.Errorf("Expected runtime function ta.Macd, got %s", spec.RuntimeFunction)
	}
}

func TestTupleIndicatorRegistry_BBRegistration(t *testing.T) {
	registry := NewTupleIndicatorRegistry()

	spec := registry.Lookup("ta.bb")
	if spec == nil {
		t.Fatal("ta.bb not registered")
	}

	if spec.OutputCount != 3 {
		t.Errorf("Expected 3 outputs, got %d", spec.OutputCount)
	}

	if spec.RuntimeFunction != "ta.BBands" {
		t.Errorf("Expected runtime function ta.BBands, got %s", spec.RuntimeFunction)
	}
}

func TestTupleIndicatorRegistry_StochRegistration(t *testing.T) {
	registry := NewTupleIndicatorRegistry()

	spec := registry.Lookup("ta.stoch")
	if spec == nil {
		t.Fatal("ta.stoch not registered")
	}

	if spec.OutputCount != 2 {
		t.Errorf("Expected 2 outputs, got %d", spec.OutputCount)
	}
}

func TestTupleIndicatorRegistry_UnregisteredFunction(t *testing.T) {
	registry := NewTupleIndicatorRegistry()

	spec := registry.Lookup("ta.nonexistent")
	if spec != nil {
		t.Error("Expected nil for unregistered function")
	}

	if registry.IsRegistered("ta.nonexistent") {
		t.Error("IsRegistered should return false for unregistered function")
	}
}

func TestTupleIndicatorRegistry_PineV4Syntax(t *testing.T) {
	registry := NewTupleIndicatorRegistry()

	spec := registry.Lookup("macd")
	if spec == nil {
		t.Fatal("macd (Pine v4 syntax) not registered")
	}

	if spec.RuntimeFunction != "ta.Macd" {
		t.Errorf("Expected runtime function ta.Macd, got %s", spec.RuntimeFunction)
	}
}

func TestTupleIndicatorRegistry_ComprehensiveRegistration(t *testing.T) {
	tests := []struct {
		name            string
		functionName    string
		outputCount     int
		runtimeFunction string
		sourceArgIndex  int
		periodArgCount  int
	}{
		{
			name:            "ta.macd",
			functionName:    "ta.macd",
			outputCount:     3,
			runtimeFunction: "ta.Macd",
			sourceArgIndex:  0,
			periodArgCount:  3,
		},
		{
			name:            "macd (v4)",
			functionName:    "macd",
			outputCount:     3,
			runtimeFunction: "ta.Macd",
			sourceArgIndex:  0,
			periodArgCount:  3,
		},
		{
			name:            "ta.bb",
			functionName:    "ta.bb",
			outputCount:     3,
			runtimeFunction: "ta.BBands",
			sourceArgIndex:  0,
			periodArgCount:  2,
		},
		{
			name:            "ta.stoch",
			functionName:    "ta.stoch",
			outputCount:     2,
			runtimeFunction: "ta.Stoch",
			sourceArgIndex:  -1, // Stoch uses multiple sources
			periodArgCount:  2,  // Actual period count in registry
		},
	}

	registry := NewTupleIndicatorRegistry()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec := registry.Lookup(tt.functionName)
			if spec == nil {
				t.Fatalf("%s not registered", tt.functionName)
			}

			if spec.FunctionName != tt.functionName {
				t.Errorf("FunctionName: expected %s, got %s", tt.functionName, spec.FunctionName)
			}

			if spec.OutputCount != tt.outputCount {
				t.Errorf("OutputCount: expected %d, got %d", tt.outputCount, spec.OutputCount)
			}

			if spec.RuntimeFunction != tt.runtimeFunction {
				t.Errorf("RuntimeFunction: expected %s, got %s", tt.runtimeFunction, spec.RuntimeFunction)
			}

			if spec.SourceArgIndex != tt.sourceArgIndex {
				t.Errorf("SourceArgIndex: expected %d, got %d", tt.sourceArgIndex, spec.SourceArgIndex)
			}

			if spec.PeriodArgCount != tt.periodArgCount {
				t.Errorf("PeriodArgCount: expected %d, got %d", tt.periodArgCount, spec.PeriodArgCount)
			}
		})
	}
}

func TestTupleIndicatorRegistry_IsRegistered(t *testing.T) {
	tests := []struct {
		name         string
		functionName string
		registered   bool
	}{
		{"ta.macd registered", "ta.macd", true},
		{"macd v4 registered", "macd", true},
		{"ta.bb registered", "ta.bb", true},
		{"ta.stoch registered", "ta.stoch", true},
		{"bb v4 not registered", "bb", false},
		{"stoch v4 not registered", "stoch", false},
		{"ta.nonexistent not registered", "ta.nonexistent", false},
		{"random not registered", "randomIndicator", false},
		{"empty string not registered", "", false},
	}

	registry := NewTupleIndicatorRegistry()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := registry.IsRegistered(tt.functionName)
			if result != tt.registered {
				t.Errorf("IsRegistered(%s): expected %v, got %v", tt.functionName, tt.registered, result)
			}
		})
	}
}

func TestTupleIndicatorRegistry_EdgeCases(t *testing.T) {
	registry := NewTupleIndicatorRegistry()

	tests := []struct {
		name        string
		lookup      string
		shouldBeNil bool
	}{
		{"Nil for unregistered", "ta.unknown", true},
		{"Nil for empty string", "", true},
		{"Nil for partial match", "ta.mac", true},
		{"Nil for case mismatch", "ta.MACD", true},
		{"Valid for exact match", "ta.macd", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec := registry.Lookup(tt.lookup)
			if tt.shouldBeNil && spec != nil {
				t.Errorf("Expected nil for %q, got spec", tt.lookup)
			}
			if !tt.shouldBeNil && spec == nil {
				t.Errorf("Expected spec for %q, got nil", tt.lookup)
			}
		})
	}
}

func TestTupleIndicatorRegistry_SpecValidation(t *testing.T) {
	registry := NewTupleIndicatorRegistry()

	indicators := []string{"ta.macd", "macd", "ta.bb", "ta.stoch"}

	for _, name := range indicators {
		t.Run(name, func(t *testing.T) {
			spec := registry.Lookup(name)
			if spec == nil {
				t.Fatalf("Indicator %s not found", name)
			}

			err := spec.Validate()
			if err != nil {
				t.Errorf("Spec validation failed: %v", err)
			}
		})
	}
}

func TestTupleIndicatorRegistry_Immutability(t *testing.T) {
	registry := NewTupleIndicatorRegistry()

	spec1 := registry.Lookup("ta.macd")
	spec2 := registry.Lookup("ta.macd")

	if spec1 == nil || spec2 == nil {
		t.Fatal("Lookup returned nil")
	}

	if spec1.FunctionName != spec2.FunctionName {
		t.Error("Registry returned inconsistent specs")
	}

	if spec1.OutputCount != spec2.OutputCount {
		t.Error("Registry returned inconsistent specs")
	}
}
