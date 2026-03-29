package codegen

import (
	"strings"
	"testing"
)

func TestArrayTypedIntegration_FloatArrayDeclarationInitialization(t *testing.T) {
	source := `//@version=5
indicator("test")
var prices = array.new_float(10, 0.0)
plot(0)`

	code, err := compilePineScript(source)
	if err != nil {
		t.Fatalf("compilation failed: %v", err)
	}

	if !strings.Contains(code, "var pricesArraySeries *series.ArraySeries") {
		t.Error("missing float array series declaration")
	}
	if !strings.Contains(code, "pricesArraySeries = series.NewArraySeries(len(ctx.Data))") {
		t.Error("missing float array series initialization")
	}
	if !strings.Contains(code, "pricesArraySeries.Next()") {
		t.Error("missing float array series Next() advancement")
	}
}

func TestArrayTypedIntegration_StringArrayDeclarationInitialization(t *testing.T) {
	source := `//@version=5
indicator("test")
var labels = array.new_string(5, "")
plot(0)`

	code, err := compilePineScript(source)
	if err != nil {
		t.Fatalf("compilation failed: %v", err)
	}

	if !strings.Contains(code, "var labelsStringArraySeries *series.StringArraySeries") {
		t.Error("missing string array series declaration")
	}
	if !strings.Contains(code, "labelsStringArraySeries = series.NewStringArraySeries(len(ctx.Data))") {
		t.Error("missing string array series initialization")
	}
	if !strings.Contains(code, "labelsStringArraySeries.Next()") {
		t.Error("missing string array series Next() advancement")
	}
}

func TestArrayTypedIntegration_MixedFloatAndStringArrays(t *testing.T) {
	source := `//@version=5
indicator("test")
var prices = array.new_float(10)
var labels = array.new_string(5)
plot(0)`

	code, err := compilePineScript(source)
	if err != nil {
		t.Fatalf("compilation failed: %v", err)
	}

	if !strings.Contains(code, "var pricesArraySeries *series.ArraySeries") {
		t.Error("missing float array declaration")
	}
	if !strings.Contains(code, "var labelsStringArraySeries *series.StringArraySeries") {
		t.Error("missing string array declaration")
	}

	if !strings.Contains(code, "pricesArraySeries = series.NewArraySeries(len(ctx.Data))") {
		t.Error("missing float array initialization")
	}
	if !strings.Contains(code, "labelsStringArraySeries = series.NewStringArraySeries(len(ctx.Data))") {
		t.Error("missing string array initialization")
	}

	floatNext := strings.Contains(code, "pricesArraySeries.Next()")
	stringNext := strings.Contains(code, "labelsStringArraySeries.Next()")
	if !floatNext || !stringNext {
		t.Errorf("missing Next() calls: float=%v, string=%v", floatNext, stringNext)
	}
}

func TestArrayTypedIntegration_TypeInferenceFromConstructor(t *testing.T) {
	tests := []struct {
		name           string
		constructor    string
		expectedType   ArrayElementType
		expectedSuffix string
		expectedGoType string
	}{
		{
			name:           "array.new_float infers float",
			constructor:    "array.new_float",
			expectedType:   ArrayElementFloat64,
			expectedSuffix: "ArraySeries",
			expectedGoType: "*series.ArraySeries",
		},
		{
			name:           "array.new_int infers float",
			constructor:    "array.new_int",
			expectedType:   ArrayElementFloat64,
			expectedSuffix: "ArraySeries",
			expectedGoType: "*series.ArraySeries",
		},
		{
			name:           "array.new_string infers string",
			constructor:    "array.new_string",
			expectedType:   ArrayElementString,
			expectedSuffix: "StringArraySeries",
			expectedGoType: "*series.StringArraySeries",
		},
		{
			name:           "array.new_bool infers float",
			constructor:    "array.new_bool",
			expectedType:   ArrayElementFloat64,
			expectedSuffix: "ArraySeries",
			expectedGoType: "*series.ArraySeries",
		},
	}

	resolver := NewArrayConstructorTypeResolver()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotType, ok := resolver.ResolveVariableType(tt.constructor)
			if !ok {
				t.Errorf("ResolveVariableType(%q) returned ok=false", tt.constructor)
			}
			if gotType != tt.expectedType {
				t.Errorf("ResolveVariableType(%q) = %v, want %v", tt.constructor, gotType, tt.expectedType)
			}

			if gotType.VariableSuffix() != tt.expectedSuffix {
				t.Errorf("VariableSuffix() = %q, want %q", gotType.VariableSuffix(), tt.expectedSuffix)
			}

			if gotType.SeriesType() != tt.expectedGoType {
				t.Errorf("SeriesType() = %q, want %q", gotType.SeriesType(), tt.expectedGoType)
			}
		})
	}
}

func TestArrayTypedIntegration_DrawingArraysMapToFloat(t *testing.T) {
	constructors := []string{
		"array.new_label",
		"array.new_line",
		"array.new_box",
		"array.new_table",
		"array.new_linefill",
	}

	resolver := NewArrayConstructorTypeResolver()

	for _, constructor := range constructors {
		t.Run(constructor, func(t *testing.T) {
			gotType, ok := resolver.ResolveVariableType(constructor)
			if !ok {
				t.Errorf("ResolveVariableType(%q) returned ok=false", constructor)
			}
			if gotType != ArrayElementFloat64 {
				t.Errorf("drawing array %q should map to ArrayElementFloat64, got %v", constructor, gotType)
			}
		})
	}
}

func TestArrayTypedIntegration_RegistryIsolation(t *testing.T) {
	registry1 := NewArrayVariableRegistry()
	registry2 := NewArrayVariableRegistry()

	registry1.Register("data", ArrayElementFloat64)
	registry2.Register("data", ArrayElementString)

	type1, _ := registry1.Lookup("data")
	type2, _ := registry2.Lookup("data")

	if type1 == type2 {
		t.Error("separate registries should not share state")
	}

	if type1 != ArrayElementFloat64 {
		t.Errorf("registry1 should have Float64, got %v", type1)
	}
	if type2 != ArrayElementString {
		t.Errorf("registry2 should have String, got %v", type2)
	}
}

func TestArrayTypedIntegration_InvalidTypeGracefulDegradation(t *testing.T) {
	resolver := NewArrayConstructorTypeResolver()

	invalidConstructors := []string{
		"array.push",
		"array.get",
		"ta.sma",
		"unknown",
		"",
	}

	for _, constructor := range invalidConstructors {
		t.Run(constructor, func(t *testing.T) {
			_, ok := resolver.ResolveVariableType(constructor)
			if ok {
				t.Errorf("ResolveVariableType(%q) should return ok=false for non-constructor", constructor)
			}
		})
	}
}
