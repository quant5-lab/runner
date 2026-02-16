package codegen

import (
	"testing"
)

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
			sourceArgIndex:  -1,
			periodArgCount:  2,
		},
		{
			name:            "bb (v4)",
			functionName:    "bb",
			outputCount:     3,
			runtimeFunction: "ta.BBands",
			sourceArgIndex:  0,
			periodArgCount:  2,
		},
		{
			name:            "stoch (v4)",
			functionName:    "stoch",
			outputCount:     2,
			runtimeFunction: "ta.Stoch",
			sourceArgIndex:  -1,
			periodArgCount:  2,
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
		/* Implemented tuple indicators — namespaced */
		{"ta.macd registered", "ta.macd", true},
		{"ta.bb registered", "ta.bb", true},
		{"ta.stoch registered", "ta.stoch", true},
		{"ta.dmi registered", "ta.dmi", true},

		/* Implemented tuple indicators — bare v4 aliases */
		{"macd v4 registered", "macd", true},
		{"bb v4 registered", "bb", true},
		{"stoch v4 registered", "stoch", true},
		{"dmi v4 registered", "dmi", true},

		/* Unregistered tuple signatures (no runtime implementation) */
		{"ta.kc unregistered", "ta.kc", false},
		{"kc unregistered", "kc", false},
		{"ta.supertrend unregistered", "ta.supertrend", false},
		{"supertrend unregistered", "supertrend", false},

		/* Non-tuple functions */
		{"ta.nonexistent not registered", "ta.nonexistent", false},
		{"random not registered", "randomIndicator", false},
		{"empty string not registered", "", false},

		/* Single-output TA functions must not be in tuple registry */
		{"ta.sma not tuple", "ta.sma", false},
		{"ta.ema not tuple", "ta.ema", false},
		{"ta.rsi not tuple", "ta.rsi", false},
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

	indicators := []string{"ta.macd", "macd", "ta.bb", "bb", "ta.stoch", "stoch"}

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

func TestTupleIndicatorRegistry_BareAliasSymmetry(t *testing.T) {
	registry := NewTupleIndicatorRegistry()

	pairs := []struct{ namespaced, bare string }{
		{"ta.macd", "macd"},
		{"ta.bb", "bb"},
		{"ta.stoch", "stoch"},
	}

	for _, pair := range pairs {
		t.Run(pair.namespaced, func(t *testing.T) {
			ns := registry.Lookup(pair.namespaced)
			bare := registry.Lookup(pair.bare)
			if ns == nil || bare == nil {
				t.Fatalf("Missing spec: namespaced=%v bare=%v", ns != nil, bare != nil)
			}
			if ns.OutputCount != bare.OutputCount {
				t.Errorf("OutputCount mismatch: %d vs %d", ns.OutputCount, bare.OutputCount)
			}
			if ns.RuntimeFunction != bare.RuntimeFunction {
				t.Errorf("RuntimeFunction mismatch: %s vs %s", ns.RuntimeFunction, bare.RuntimeFunction)
			}
			if ns.SourceArgIndex != bare.SourceArgIndex {
				t.Errorf("SourceArgIndex mismatch: %d vs %d", ns.SourceArgIndex, bare.SourceArgIndex)
			}
			if ns.PeriodArgCount != bare.PeriodArgCount {
				t.Errorf("PeriodArgCount mismatch: %d vs %d", ns.PeriodArgCount, bare.PeriodArgCount)
			}
		})
	}
}
