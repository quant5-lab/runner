package codegen

import (
	"strings"
	"testing"
)

func TestTASignatureRegistry_IsTupleFunction(t *testing.T) {
	registry := NewTASignatureRegistry()

	tests := []struct {
		name     string
		funcName string
		want     bool
	}{
		/* Tuple functions registered via appendTupleWithBareAlias — both forms */
		{"ta.macd", "ta.macd", true},
		{"macd bare", "macd", true},
		{"ta.bb", "ta.bb", true},
		{"bb bare", "bb", true},
		{"ta.stoch", "ta.stoch", true},
		{"stoch bare", "stoch", true},
		{"ta.dmi", "ta.dmi", true},
		{"dmi bare", "dmi", true},
		{"ta.supertrend", "ta.supertrend", true},
		{"supertrend bare", "supertrend", true},
		{"ta.kc", "ta.kc", true},
		{"kc bare", "kc", true},

		/* Non-tuple functions registered via appendWithBareAlias */
		{"ta.sma", "ta.sma", false},
		{"sma bare", "sma", false},
		{"ta.ema", "ta.ema", false},
		{"ta.rsi", "ta.rsi", false},
		{"ta.atr", "ta.atr", false},
		{"ta.highest", "ta.highest", false},

		/* Unregistered functions */
		{"unknown", "ta.nonexistent", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := registry.IsTupleFunction(tt.funcName)
			if got != tt.want {
				t.Errorf("IsTupleFunction(%q) = %v, want %v", tt.funcName, got, tt.want)
			}
		})
	}
}

/* Architectural invariant: tuple functions must NOT be claimed by TAIndicatorCallHandler */
func TestTASignatureRegistry_TupleExclusionFromHandler(t *testing.T) {
	registry := NewTASignatureRegistry()
	handler := &TAIndicatorCallHandler{}

	for _, name := range registry.AllFunctionNames() {
		if !registry.IsTupleFunction(name) {
			continue
		}
		t.Run(name, func(t *testing.T) {
			if handler.CanHandle(name) {
				t.Errorf("TAIndicatorCallHandler must not claim tuple function %q", name)
			}
		})
	}
}

/* Cross-registry consistency: codegen registry ⊆ signature registry */
func TestTASignatureRegistry_TupleRegistryConsistency(t *testing.T) {
	sigRegistry := NewTASignatureRegistry()
	tupleRegistry := NewTupleIndicatorRegistry()

	implementedTuples := []string{"ta.macd", "macd", "ta.bb", "bb", "ta.stoch", "stoch"}

	for _, name := range implementedTuples {
		t.Run(name, func(t *testing.T) {
			if !tupleRegistry.IsRegistered(name) {
				t.Fatalf("%s not in TupleIndicatorRegistry", name)
			}
			if !sigRegistry.IsTupleFunction(name) {
				t.Errorf("%s in TupleIndicatorRegistry but IsTupleFunction()=false in TASignatureRegistry", name)
			}
		})
	}
}

func TestTASignatureRegistry_BareAliasSymmetry(t *testing.T) {
	registry := NewTASignatureRegistry()

	pairs := []struct{ namespaced, bare string }{
		{"ta.sma", "sma"},
		{"ta.ema", "ema"},
		{"ta.rsi", "rsi"},
		{"ta.macd", "macd"},
		{"ta.bb", "bb"},
		{"ta.stoch", "stoch"},
		{"ta.highest", "highest"},
		{"ta.lowest", "lowest"},
		{"ta.change", "change"},
	}

	for _, pair := range pairs {
		t.Run(pair.namespaced, func(t *testing.T) {
			ns, nsOk := registry.Lookup(pair.namespaced)
			bare, bareOk := registry.Lookup(pair.bare)
			if !nsOk || !bareOk {
				t.Fatalf("Missing: namespaced=%v bare=%v", nsOk, bareOk)
			}
			if ns.IsTuple != bare.IsTuple {
				t.Errorf("IsTuple mismatch: %s=%v, %s=%v", pair.namespaced, ns.IsTuple, pair.bare, bare.IsTuple)
			}
			if ns.DefaultSource != bare.DefaultSource {
				t.Errorf("DefaultSource mismatch: %s=%q, %s=%q", pair.namespaced, ns.DefaultSource, pair.bare, bare.DefaultSource)
			}
			if len(ns.Overloads) != len(bare.Overloads) {
				t.Errorf("Overload count mismatch: %s=%d, %s=%d",
					pair.namespaced, len(ns.Overloads), pair.bare, len(bare.Overloads))
			}
		})
	}
}

func TestTASignatureRegistry_ContainsAllRegisteredFunctions(t *testing.T) {
	registry := NewTASignatureRegistry()

	for _, name := range registry.AllFunctionNames() {
		t.Run(name, func(t *testing.T) {
			if !registry.Contains(name) {
				t.Errorf("AllFunctionNames returned %q but Contains()=false", name)
			}
			meta, ok := registry.Lookup(name)
			if !ok {
				t.Errorf("AllFunctionNames returned %q but Lookup()=false", name)
			}
			if meta.FunctionName != name {
				t.Errorf("Lookup(%q).FunctionName = %q", name, meta.FunctionName)
			}
		})
	}
}

/* Cross-contamination guard: single-output functions never in TupleIndicatorRegistry */
func TestTASignatureRegistry_NonTupleNotInTupleRegistry(t *testing.T) {
	sigRegistry := NewTASignatureRegistry()
	tupleRegistry := NewTupleIndicatorRegistry()

	for _, name := range sigRegistry.AllFunctionNames() {
		if sigRegistry.IsTupleFunction(name) {
			continue
		}
		t.Run(name, func(t *testing.T) {
			if tupleRegistry.IsRegistered(name) {
				t.Errorf("Non-tuple function %q found in TupleIndicatorRegistry", name)
			}
		})
	}
}

func TestTASignatureRegistry_NeedsSourcePromotion(t *testing.T) {
	registry := NewTASignatureRegistry()

	tests := []struct {
		funcName string
		argCount int
		want     bool
	}{
		{"ta.sma", 2, true},
		{"ta.sma", 1, false},
		{"ta.ema", 2, true},
		{"ta.highest", 2, true},
		{"ta.highest", 1, false},
		{"ta.atr", 1, false},
		{"ta.nonexistent", 2, false},
	}

	for _, tt := range tests {
		name := tt.funcName + "_" + strings.Repeat("a", tt.argCount)
		t.Run(name, func(t *testing.T) {
			got := registry.NeedsSourcePromotion(tt.funcName, tt.argCount)
			if got != tt.want {
				t.Errorf("NeedsSourcePromotion(%q, %d) = %v, want %v", tt.funcName, tt.argCount, got, tt.want)
			}
		})
	}
}
