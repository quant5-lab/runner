package codegen

import "testing"

func TestArrayElementType_IsNumeric(t *testing.T) {
	tests := []struct {
		name     string
		elemType ArrayElementType
		expected bool
	}{
		{"Float64 is numeric", ArrayElementFloat64, true},
		{"String is not numeric", ArrayElementString, false},
		{"Zero value is numeric", ArrayElementType(0), true},
		{"Invalid type is not numeric", ArrayElementType(99), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.elemType.IsNumeric(); got != tt.expected {
				t.Errorf("IsNumeric() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestArrayElementType_IsString(t *testing.T) {
	tests := []struct {
		name     string
		elemType ArrayElementType
		expected bool
	}{
		{"String is string", ArrayElementString, true},
		{"Float64 is not string", ArrayElementFloat64, false},
		{"Zero value is not string", ArrayElementType(0), false},
		{"Invalid type is not string", ArrayElementType(99), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.elemType.IsString(); got != tt.expected {
				t.Errorf("IsString() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestArrayElementType_GoType(t *testing.T) {
	tests := []struct {
		name     string
		elemType ArrayElementType
		expected string
	}{
		{"Float64 maps to float64", ArrayElementFloat64, "float64"},
		{"String maps to string", ArrayElementString, "string"},
		{"Zero value defaults to float64", ArrayElementType(0), "float64"},
		{"Invalid type defaults to float64", ArrayElementType(99), "float64"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.elemType.GoType(); got != tt.expected {
				t.Errorf("GoType() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestArrayElementType_GoSliceType(t *testing.T) {
	tests := []struct {
		name     string
		elemType ArrayElementType
		expected string
	}{
		{"Float64 slice", ArrayElementFloat64, "[]float64"},
		{"String slice", ArrayElementString, "[]string"},
		{"Zero value slice", ArrayElementType(0), "[]float64"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.elemType.GoSliceType(); got != tt.expected {
				t.Errorf("GoSliceType() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestArrayElementType_SeriesType(t *testing.T) {
	tests := []struct {
		name     string
		elemType ArrayElementType
		expected string
	}{
		{"Float64 series pointer", ArrayElementFloat64, "*series.ArraySeries"},
		{"String series pointer", ArrayElementString, "*series.StringArraySeries"},
		{"Zero value defaults to ArraySeries", ArrayElementType(0), "*series.ArraySeries"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.elemType.SeriesType(); got != tt.expected {
				t.Errorf("SeriesType() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestArrayElementType_NewSeriesCall(t *testing.T) {
	tests := []struct {
		name     string
		elemType ArrayElementType
		capacity string
		expected string
	}{
		{"Float64 with capacity 10", ArrayElementFloat64, "10", "series.NewArraySeries(10)"},
		{"String with capacity 5", ArrayElementString, "5", "series.NewStringArraySeries(5)"},
		{"Float64 with expression capacity", ArrayElementFloat64, "n + 1", "series.NewArraySeries(n + 1)"},
		{"String with zero capacity", ArrayElementString, "0", "series.NewStringArraySeries(0)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.elemType.NewSeriesCall(tt.capacity); got != tt.expected {
				t.Errorf("NewSeriesCall(%q) = %q, want %q", tt.capacity, got, tt.expected)
			}
		})
	}
}

func TestArrayElementType_VariableSuffix(t *testing.T) {
	tests := []struct {
		name     string
		elemType ArrayElementType
		expected string
	}{
		{"Float64 suffix", ArrayElementFloat64, "ArraySeries"},
		{"String suffix", ArrayElementString, "StringArraySeries"},
		{"Zero value defaults to ArraySeries", ArrayElementType(0), "ArraySeries"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.elemType.VariableSuffix(); got != tt.expected {
				t.Errorf("VariableSuffix() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestArrayElementType_TypeTag(t *testing.T) {
	tests := []struct {
		name     string
		elemType ArrayElementType
		expected string
	}{
		{"Float64 type tag", ArrayElementFloat64, "array_series_float"},
		{"String type tag", ArrayElementString, "array_series_string"},
		{"Zero value defaults to float tag", ArrayElementType(0), "array_series_float"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.elemType.TypeTag(); got != tt.expected {
				t.Errorf("TypeTag() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestArrayElementType_MutatorConstructor(t *testing.T) {
	tests := []struct {
		name     string
		elemType ArrayElementType
		expected string
	}{
		{"Float64 mutator", ArrayElementFloat64, "arrayops.NewMutator()"},
		{"String mutator", ArrayElementString, "arrayops.NewStringMutator()"},
		{"Zero value defaults to float mutator", ArrayElementType(0), "arrayops.NewMutator()"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.elemType.MutatorConstructor(); got != tt.expected {
				t.Errorf("MutatorConstructor() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestArrayElementType_AccessorConstructor(t *testing.T) {
	tests := []struct {
		name     string
		elemType ArrayElementType
		expected string
	}{
		{"Float64 accessor", ArrayElementFloat64, "arrayops.NewAccessor()"},
		{"String accessor", ArrayElementString, "arrayops.NewStringAccessor()"},
		{"Zero value defaults to float accessor", ArrayElementType(0), "arrayops.NewAccessor()"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.elemType.AccessorConstructor(); got != tt.expected {
				t.Errorf("AccessorConstructor() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestArrayElementType_TransformerConstructor(t *testing.T) {
	tests := []struct {
		name     string
		elemType ArrayElementType
		expected string
	}{
		{"Float64 transformer", ArrayElementFloat64, "arrayops.NewTransformer()"},
		{"String transformer", ArrayElementString, "arrayops.NewStringTransformers()"},
		{"Zero value defaults to float transformer", ArrayElementType(0), "arrayops.NewTransformer()"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.elemType.TransformerConstructor(); got != tt.expected {
				t.Errorf("TransformerConstructor() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestArrayElementType_StatisticsConstructor(t *testing.T) {
	tests := []struct {
		name     string
		elemType ArrayElementType
		expected string
	}{
		{"Float64 statistics", ArrayElementFloat64, "arrayops.NewStatistics()"},
		{"String statistics", ArrayElementString, "arrayops.NewStatistics()"},
		{"Zero value statistics", ArrayElementType(0), "arrayops.NewStatistics()"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.elemType.StatisticsConstructor(); got != tt.expected {
				t.Errorf("StatisticsConstructor() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestArrayElementType_SearchConstructor(t *testing.T) {
	tests := []struct {
		name     string
		elemType ArrayElementType
		expected string
	}{
		{"Float64 search", ArrayElementFloat64, "arrayops.NewSearch()"},
		{"String search", ArrayElementString, "arrayops.NewSearch()"},
		{"Zero value search", ArrayElementType(0), "arrayops.NewSearch()"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.elemType.SearchConstructor(); got != tt.expected {
				t.Errorf("SearchConstructor() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestArrayElementType_PredicatesConstructor(t *testing.T) {
	tests := []struct {
		name     string
		elemType ArrayElementType
		expected string
	}{
		{"Float64 predicates", ArrayElementFloat64, "arrayops.NewPredicates()"},
		{"String predicates", ArrayElementString, "arrayops.NewPredicates()"},
		{"Zero value predicates", ArrayElementType(0), "arrayops.NewPredicates()"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.elemType.PredicatesConstructor(); got != tt.expected {
				t.Errorf("PredicatesConstructor() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestArrayElementType_FormattersConstructor(t *testing.T) {
	tests := []struct {
		name     string
		elemType ArrayElementType
		expected string
	}{
		{"Float64 formatters", ArrayElementFloat64, "arrayops.NewFormatters()"},
		{"String formatters", ArrayElementString, "arrayops.NewStringFormatters()"},
		{"Zero value defaults to float formatters", ArrayElementType(0), "arrayops.NewFormatters()"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.elemType.FormattersConstructor(); got != tt.expected {
				t.Errorf("FormattersConstructor() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestArrayElementType_SupportsStatistics(t *testing.T) {
	tests := []struct {
		name     string
		elemType ArrayElementType
		expected bool
	}{
		{"Float64 supports statistics", ArrayElementFloat64, true},
		{"String does not support statistics", ArrayElementString, false},
		{"Zero value supports statistics", ArrayElementType(0), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.elemType.SupportsStatistics(); got != tt.expected {
				t.Errorf("SupportsStatistics() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestParseArrayElementType(t *testing.T) {
	tests := []struct {
		name         string
		typeTag      string
		expectedType ArrayElementType
		expectedOk   bool
	}{
		{"Valid float tag", "array_series_float", ArrayElementFloat64, true},
		{"Valid string tag", "array_series_string", ArrayElementString, true},
		{"Invalid tag", "array_series_invalid", 0, false},
		{"Empty tag", "", 0, false},
		{"Old legacy tag", "array_series", 0, false},
		{"Partial match", "array_series_f", 0, false},
		{"Case sensitive", "Array_Series_Float", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotType, gotOk := ParseArrayElementType(tt.typeTag)
			if gotOk != tt.expectedOk {
				t.Errorf("ParseArrayElementType(%q) ok = %v, want %v", tt.typeTag, gotOk, tt.expectedOk)
			}
			if gotType != tt.expectedType {
				t.Errorf("ParseArrayElementType(%q) type = %v, want %v", tt.typeTag, gotType, tt.expectedType)
			}
		})
	}
}

func TestArrayElementType_RoundTrip(t *testing.T) {
	tests := []struct {
		name     string
		elemType ArrayElementType
	}{
		{"Float64 round trip", ArrayElementFloat64},
		{"String round trip", ArrayElementString},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tag := tt.elemType.TypeTag()
			parsed, ok := ParseArrayElementType(tag)
			if !ok {
				t.Errorf("ParseArrayElementType(%q) failed", tag)
			}
			if parsed != tt.elemType {
				t.Errorf("Round trip failed: started with %v, got %v", tt.elemType, parsed)
			}
		})
	}
}

func TestArrayElementType_ConsistencyAcrossMethods(t *testing.T) {
	types := []ArrayElementType{ArrayElementFloat64, ArrayElementString}

	for _, elemType := range types {
		t.Run(elemType.TypeTag(), func(t *testing.T) {
			isNumeric := elemType.IsNumeric()
			isString := elemType.IsString()

			if isNumeric == isString {
				t.Errorf("Type must be either numeric XOR string, got numeric=%v, string=%v", isNumeric, isString)
			}

			if isString && elemType.SupportsStatistics() {
				t.Error("String types must not support statistics")
			}

			if isNumeric && !elemType.SupportsStatistics() {
				t.Error("Numeric types must support statistics")
			}

			suffix := elemType.VariableSuffix()
			if isString && suffix != "StringArraySeries" {
				t.Errorf("String type has wrong suffix: %q", suffix)
			}
			if isNumeric && suffix != "ArraySeries" {
				t.Errorf("Numeric type has wrong suffix: %q", suffix)
			}
		})
	}
}

func TestArrayElementType_ConstructorStringFormat(t *testing.T) {
	tests := []struct {
		name        string
		elemType    ArrayElementType
		constructor string
	}{
		{"Mutator", ArrayElementFloat64, ArrayElementFloat64.MutatorConstructor()},
		{"Accessor", ArrayElementFloat64, ArrayElementFloat64.AccessorConstructor()},
		{"Transformer", ArrayElementFloat64, ArrayElementFloat64.TransformerConstructor()},
		{"Statistics", ArrayElementFloat64, ArrayElementFloat64.StatisticsConstructor()},
		{"Search", ArrayElementFloat64, ArrayElementFloat64.SearchConstructor()},
		{"Predicates", ArrayElementFloat64, ArrayElementFloat64.PredicatesConstructor()},
		{"Formatters", ArrayElementFloat64, ArrayElementFloat64.FormattersConstructor()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constructor == "" {
				t.Error("Constructor must not be empty")
			}
			if tt.constructor[:9] != "arrayops." {
				t.Errorf("Constructor must start with 'arrayops.', got %q", tt.constructor)
			}
			if tt.constructor[len(tt.constructor)-2:] != "()" {
				t.Errorf("Constructor must end with '()', got %q", tt.constructor)
			}
		})
	}
}
