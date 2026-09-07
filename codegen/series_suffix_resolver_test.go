package codegen

import "testing"

func TestSeriesSuffixResolver_Fixed(t *testing.T) {
	r := &FixedSeriesSuffixResolver{}

	for _, name := range []string{"price", "levels", "x", "anything"} {
		if got := r.ResolveSuffix(name); got != "Series" {
			t.Errorf("FixedSeriesSuffixResolver.ResolveSuffix(%q) = %q, want %q", name, got, "Series")
		}
	}
}

func TestSeriesSuffixResolver_Array(t *testing.T) {
	registry := NewArrayVariableRegistry()
	registry.Register("floatArr", ArrayElementFloat64)
	registry.Register("strArr", ArrayElementString)

	r := NewArrayRegistrySuffixResolver(registry)

	tests := []struct {
		varName string
		want    string
	}{
		{"floatArr", "ArraySeries"},
		{"strArr", "StringArraySeries"},
		{"unregistered", "Series"},
		{"", "Series"},
	}

	for _, tt := range tests {
		t.Run(tt.varName+"→"+tt.want, func(t *testing.T) {
			if got := r.ResolveSuffix(tt.varName); got != tt.want {
				t.Errorf("ResolveSuffix(%q) = %q, want %q", tt.varName, got, tt.want)
			}
		})
	}
}

// TestSeriesSuffixResolver_AllArrayElementTypes ensures every concrete ArrayElementType
// maps to a distinct, non-empty suffix and does not fall back to "Series".
func TestSeriesSuffixResolver_AllArrayElementTypes(t *testing.T) {
	elementTypes := []struct {
		elem   ArrayElementType
		suffix string
	}{
		{ArrayElementFloat64, "ArraySeries"},
		{ArrayElementString, "StringArraySeries"},
	}

	for _, tt := range elementTypes {
		t.Run(tt.suffix, func(t *testing.T) {
			registry := NewArrayVariableRegistry()
			registry.Register("v", tt.elem)
			r := NewArrayRegistrySuffixResolver(registry)

			got := r.ResolveSuffix("v")
			if got != tt.suffix {
				t.Errorf("element type %v: ResolveSuffix = %q, want %q", tt.elem, got, tt.suffix)
			}
			if got == "Series" {
				t.Errorf("element type %v fell back to plain 'Series' — wrong suffix", tt.elem)
			}
		})
	}
}

// TestSeriesSuffixResolver_InterfaceCompliance verifies both implementations satisfy
// SeriesSuffixResolver so they can be used interchangeably in VarPersistenceEmitter.
func TestSeriesSuffixResolver_InterfaceCompliance(t *testing.T) {
	var _ SeriesSuffixResolver = &FixedSeriesSuffixResolver{}
	var _ SeriesSuffixResolver = NewArrayRegistrySuffixResolver(NewArrayVariableRegistry())
}
