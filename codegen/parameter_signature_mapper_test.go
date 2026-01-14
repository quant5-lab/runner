package codegen

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
)

/* TestParameterSignatureMapper_MapUsageToSignatureTypes validates type conversion */
func TestParameterSignatureMapper_MapUsageToSignatureTypes(t *testing.T) {
	tests := []struct {
		name          string
		params        []ast.Identifier
		usageTypes    map[string]ParameterUsageType
		expectedTypes []FunctionParameterType
	}{
		{
			name: "all scalar parameters",
			params: []ast.Identifier{
				{Name: "len"},
				{Name: "mult"},
			},
			usageTypes: map[string]ParameterUsageType{
				"len":  ParameterUsageScalar,
				"mult": ParameterUsageScalar,
			},
			expectedTypes: []FunctionParameterType{
				ParamTypeScalar,
				ParamTypeScalar,
			},
		},
		{
			name: "all series parameters",
			params: []ast.Identifier{
				{Name: "src"},
				{Name: "baseline"},
			},
			usageTypes: map[string]ParameterUsageType{
				"src":      ParameterUsageSeries,
				"baseline": ParameterUsageSeries,
			},
			expectedTypes: []FunctionParameterType{
				ParamTypeSeries,
				ParamTypeSeries,
			},
		},
		{
			name: "mixed series and scalar",
			params: []ast.Identifier{
				{Name: "src"},
				{Name: "len"},
				{Name: "mult"},
			},
			usageTypes: map[string]ParameterUsageType{
				"src":  ParameterUsageSeries,
				"len":  ParameterUsageScalar,
				"mult": ParameterUsageScalar,
			},
			expectedTypes: []FunctionParameterType{
				ParamTypeSeries,
				ParamTypeScalar,
				ParamTypeScalar,
			},
		},
		{
			name:          "empty parameters",
			params:        []ast.Identifier{},
			usageTypes:    map[string]ParameterUsageType{},
			expectedTypes: []FunctionParameterType{},
		},
		{
			name: "single series parameter",
			params: []ast.Identifier{
				{Name: "source"},
			},
			usageTypes: map[string]ParameterUsageType{
				"source": ParameterUsageSeries,
			},
			expectedTypes: []FunctionParameterType{
				ParamTypeSeries,
			},
		},
		{
			name: "single scalar parameter",
			params: []ast.Identifier{
				{Name: "period"},
			},
			usageTypes: map[string]ParameterUsageType{
				"period": ParameterUsageScalar,
			},
			expectedTypes: []FunctionParameterType{
				ParamTypeScalar,
			},
		},
		{
			name: "series-scalar-series pattern",
			params: []ast.Identifier{
				{Name: "src1"},
				{Name: "len"},
				{Name: "src2"},
			},
			usageTypes: map[string]ParameterUsageType{
				"src1": ParameterUsageSeries,
				"len":  ParameterUsageScalar,
				"src2": ParameterUsageSeries,
			},
			expectedTypes: []FunctionParameterType{
				ParamTypeSeries,
				ParamTypeScalar,
				ParamTypeSeries,
			},
		},
		{
			name: "multiple consecutive series",
			params: []ast.Identifier{
				{Name: "src1"},
				{Name: "src2"},
				{Name: "src3"},
			},
			usageTypes: map[string]ParameterUsageType{
				"src1": ParameterUsageSeries,
				"src2": ParameterUsageSeries,
				"src3": ParameterUsageSeries,
			},
			expectedTypes: []FunctionParameterType{
				ParamTypeSeries,
				ParamTypeSeries,
				ParamTypeSeries,
			},
		},
		{
			name: "multiple consecutive scalars",
			params: []ast.Identifier{
				{Name: "len"},
				{Name: "mult"},
				{Name: "offset"},
				{Name: "period"},
			},
			usageTypes: map[string]ParameterUsageType{
				"len":    ParameterUsageScalar,
				"mult":   ParameterUsageScalar,
				"offset": ParameterUsageScalar,
				"period": ParameterUsageScalar,
			},
			expectedTypes: []FunctionParameterType{
				ParamTypeScalar,
				ParamTypeScalar,
				ParamTypeScalar,
				ParamTypeScalar,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mapper := NewParameterSignatureMapper()
			result := mapper.MapUsageToSignatureTypes(tt.params, tt.usageTypes)

			if len(result) != len(tt.expectedTypes) {
				t.Fatalf("Length mismatch: got %d, want %d", len(result), len(tt.expectedTypes))
			}

			for i, expected := range tt.expectedTypes {
				if result[i] != expected {
					t.Errorf("Type mismatch at index %d: got %v, want %v", i, result[i], expected)
				}
			}
		})
	}
}

/* TestParameterSignatureMapper_ParameterOrdering validates order preservation */
func TestParameterSignatureMapper_ParameterOrdering(t *testing.T) {
	mapper := NewParameterSignatureMapper()

	params := []ast.Identifier{
		{Name: "first"},
		{Name: "second"},
		{Name: "third"},
	}
	usageTypes := map[string]ParameterUsageType{
		"first":  ParameterUsageScalar,
		"second": ParameterUsageSeries,
		"third":  ParameterUsageScalar,
	}

	result := mapper.MapUsageToSignatureTypes(params, usageTypes)

	expected := []FunctionParameterType{
		ParamTypeScalar,
		ParamTypeSeries,
		ParamTypeScalar,
	}

	if len(result) != len(expected) {
		t.Fatalf("Length mismatch: got %d, want %d", len(result), len(expected))
	}

	for i, exp := range expected {
		if result[i] != exp {
			t.Errorf("Order violation at index %d: got %v, want %v", i, result[i], exp)
		}
	}
}

/* TestParameterSignatureMapper_EdgeCases validates boundary conditions */
func TestParameterSignatureMapper_EdgeCases(t *testing.T) {
	t.Run("parameter missing from usage map defaults to scalar", func(t *testing.T) {
		mapper := NewParameterSignatureMapper()
		params := []ast.Identifier{
			{Name: "existing"},
			{Name: "missing"},
		}
		usageTypes := map[string]ParameterUsageType{
			"existing": ParameterUsageSeries,
		}

		result := mapper.MapUsageToSignatureTypes(params, usageTypes)

		if len(result) != 2 {
			t.Fatalf("Expected 2 parameters, got %d", len(result))
		}
		if result[0] != ParamTypeSeries {
			t.Errorf("First parameter should be series, got %v", result[0])
		}
		if result[1] != ParamTypeScalar {
			t.Errorf("Missing parameter should default to scalar, got %v", result[1])
		}
	})

	t.Run("nil usage map treats all as scalar", func(t *testing.T) {
		mapper := NewParameterSignatureMapper()
		params := []ast.Identifier{
			{Name: "param1"},
			{Name: "param2"},
		}

		result := mapper.MapUsageToSignatureTypes(params, nil)

		if len(result) != 2 {
			t.Fatalf("Expected 2 parameters, got %d", len(result))
		}
		for i, paramType := range result {
			if paramType != ParamTypeScalar {
				t.Errorf("Parameter %d should default to scalar, got %v", i, paramType)
			}
		}
	})

	t.Run("empty parameter name", func(t *testing.T) {
		mapper := NewParameterSignatureMapper()
		params := []ast.Identifier{
			{Name: ""},
		}
		usageTypes := map[string]ParameterUsageType{}

		result := mapper.MapUsageToSignatureTypes(params, usageTypes)

		if len(result) != 1 {
			t.Fatalf("Expected 1 parameter, got %d", len(result))
		}
		if result[0] != ParamTypeScalar {
			t.Errorf("Empty name should default to scalar, got %v", result[0])
		}
	})
}

/* TestParameterSignatureMapper_Idempotency validates consistent mapping */
func TestParameterSignatureMapper_Idempotency(t *testing.T) {
	mapper := NewParameterSignatureMapper()

	params := []ast.Identifier{
		{Name: "src"},
		{Name: "len"},
	}
	usageTypes := map[string]ParameterUsageType{
		"src": ParameterUsageSeries,
		"len": ParameterUsageScalar,
	}

	result1 := mapper.MapUsageToSignatureTypes(params, usageTypes)
	result2 := mapper.MapUsageToSignatureTypes(params, usageTypes)

	if len(result1) != len(result2) {
		t.Fatalf("Length mismatch between calls: %d vs %d", len(result1), len(result2))
	}

	for i := range result1 {
		if result1[i] != result2[i] {
			t.Errorf("Mapping changed at index %d: %v vs %v", i, result1[i], result2[i])
		}
	}
}
