package codegen

import (
	"strings"
	"testing"
)

func TestStochSingleValueHandler_Registered(t *testing.T) {
	registry := NewTAFunctionRegistry()
	if !registry.IsSupported("stoch") {
		t.Fatal("stoch not registered in TAFunctionRegistry")
	}
	if !registry.IsSupported("ta.stoch") {
		t.Fatal("ta.stoch not registered in TAFunctionRegistry")
	}
}

func TestStochSingleValueHandler_AsSource(t *testing.T) {
	src := `//@version=4
strategy("test")
K=input(title="K",type=input.integer,defval=14)
D=input(title="D",type=input.integer,defval=3)
k_val=sma(stoch(close,high,low,K),D)
`
	code, err := compilePineScript(src)
	if err != nil {
		t.Fatalf("compilePineScript error: %v", err)
	}
	// Should NOT be just a NaN stub — expect actual stoch computation (_hh/_ll pattern)
	if !strings.Contains(code, "_hh") {
		t.Errorf("expected stoch window computation in generated code:\n%s", code)
	}
}
