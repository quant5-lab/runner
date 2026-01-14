package codegen

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
)

/* TestSignatureRegistrar_RegisterArrowFunction validates registration workflow */
func TestSignatureRegistrar_RegisterArrowFunction(t *testing.T) {
	tests := []struct {
		name       string
		funcName   string
		params     []ast.Identifier
		paramUsage map[string]ParameterUsageType
		returnType string
		verifyFunc func(*testing.T, *FunctionSignatureRegistry)
	}{
		{
			name:     "scalar-only function",
			funcName: "calc",
			params: []ast.Identifier{
				{Name: "len"},
				{Name: "mult"},
			},
			paramUsage: map[string]ParameterUsageType{
				"len":  ParameterUsageScalar,
				"mult": ParameterUsageScalar,
			},
			returnType: "float64",
			verifyFunc: func(t *testing.T, registry *FunctionSignatureRegistry) {
				sig, exists := registry.Get("calc")
				if !exists {
					t.Fatal("Function signature not found")
				}
				if sig.Name != "calc" {
					t.Errorf("Name mismatch: got %s, want calc", sig.Name)
				}
				if len(sig.Parameters) != 2 {
					t.Errorf("Parameter count mismatch: got %d, want 2", len(sig.Parameters))
				}
				if sig.Parameters[0] != ParamTypeScalar || sig.Parameters[1] != ParamTypeScalar {
					t.Error("Expected all scalar parameters")
				}
				if sig.ReturnType != "float64" {
					t.Errorf("Return type mismatch: got %s, want float64", sig.ReturnType)
				}
			},
		},
		{
			name:     "series-only function",
			funcName: "smoothed",
			params: []ast.Identifier{
				{Name: "src"},
			},
			paramUsage: map[string]ParameterUsageType{
				"src": ParameterUsageSeries,
			},
			returnType: "float64",
			verifyFunc: func(t *testing.T, registry *FunctionSignatureRegistry) {
				sig, exists := registry.Get("smoothed")
				if !exists {
					t.Fatal("Function signature not found")
				}
				if len(sig.Parameters) != 1 {
					t.Errorf("Parameter count mismatch: got %d, want 1", len(sig.Parameters))
				}
				if sig.Parameters[0] != ParamTypeSeries {
					t.Error("Expected series parameter")
				}
			},
		},
		{
			name:     "mixed parameter types",
			funcName: "bands",
			params: []ast.Identifier{
				{Name: "src"},
				{Name: "len"},
				{Name: "mult"},
			},
			paramUsage: map[string]ParameterUsageType{
				"src":  ParameterUsageSeries,
				"len":  ParameterUsageScalar,
				"mult": ParameterUsageScalar,
			},
			returnType: "float64",
			verifyFunc: func(t *testing.T, registry *FunctionSignatureRegistry) {
				sig, exists := registry.Get("bands")
				if !exists {
					t.Fatal("Function signature not found")
				}
				if len(sig.Parameters) != 3 {
					t.Errorf("Parameter count mismatch: got %d, want 3", len(sig.Parameters))
				}
				if sig.Parameters[0] != ParamTypeSeries {
					t.Error("First parameter should be series")
				}
				if sig.Parameters[1] != ParamTypeScalar || sig.Parameters[2] != ParamTypeScalar {
					t.Error("Last two parameters should be scalar")
				}
			},
		},
		{
			name:       "zero-parameter function",
			funcName:   "simple",
			params:     []ast.Identifier{},
			paramUsage: map[string]ParameterUsageType{},
			returnType: "float64",
			verifyFunc: func(t *testing.T, registry *FunctionSignatureRegistry) {
				sig, exists := registry.Get("simple")
				if !exists {
					t.Fatal("Function signature not found")
				}
				if len(sig.Parameters) != 0 {
					t.Errorf("Expected zero parameters, got %d", len(sig.Parameters))
				}
			},
		},
		{
			name:     "single series parameter",
			funcName: "transform",
			params: []ast.Identifier{
				{Name: "data"},
			},
			paramUsage: map[string]ParameterUsageType{
				"data": ParameterUsageSeries,
			},
			returnType: "float64",
			verifyFunc: func(t *testing.T, registry *FunctionSignatureRegistry) {
				paramType, exists := registry.GetParameterType("transform", 0)
				if !exists {
					t.Fatal("Parameter type not found")
				}
				if paramType != ParamTypeSeries {
					t.Errorf("Expected series parameter, got %v", paramType)
				}
			},
		},
		{
			name:     "single scalar parameter",
			funcName: "multiplier",
			params: []ast.Identifier{
				{Name: "factor"},
			},
			paramUsage: map[string]ParameterUsageType{
				"factor": ParameterUsageScalar,
			},
			returnType: "float64",
			verifyFunc: func(t *testing.T, registry *FunctionSignatureRegistry) {
				paramType, exists := registry.GetParameterType("multiplier", 0)
				if !exists {
					t.Fatal("Parameter type not found")
				}
				if paramType != ParamTypeScalar {
					t.Errorf("Expected scalar parameter, got %v", paramType)
				}
			},
		},
		{
			name:     "multiple series parameters",
			funcName: "compare",
			params: []ast.Identifier{
				{Name: "series1"},
				{Name: "series2"},
				{Name: "series3"},
			},
			paramUsage: map[string]ParameterUsageType{
				"series1": ParameterUsageSeries,
				"series2": ParameterUsageSeries,
				"series3": ParameterUsageSeries,
			},
			returnType: "float64",
			verifyFunc: func(t *testing.T, registry *FunctionSignatureRegistry) {
				sig, exists := registry.Get("compare")
				if !exists {
					t.Fatal("Function signature not found")
				}
				if len(sig.Parameters) != 3 {
					t.Fatalf("Expected 3 parameters, got %d", len(sig.Parameters))
				}
				for i, param := range sig.Parameters {
					if param != ParamTypeSeries {
						t.Errorf("Parameter %d should be series, got %v", i, param)
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := NewFunctionSignatureRegistry()
			registrar := NewSignatureRegistrar(registry)

			registrar.RegisterArrowFunction(tt.funcName, tt.params, tt.paramUsage, tt.returnType)

			tt.verifyFunc(t, registry)
		})
	}
}

/* TestSignatureRegistrar_ParameterOrdering validates order preservation */
func TestSignatureRegistrar_ParameterOrdering(t *testing.T) {
	registry := NewFunctionSignatureRegistry()
	registrar := NewSignatureRegistrar(registry)

	params := []ast.Identifier{
		{Name: "first"},
		{Name: "second"},
		{Name: "third"},
		{Name: "fourth"},
	}
	paramUsage := map[string]ParameterUsageType{
		"first":  ParameterUsageSeries,
		"second": ParameterUsageScalar,
		"third":  ParameterUsageSeries,
		"fourth": ParameterUsageScalar,
	}

	registrar.RegisterArrowFunction("ordered", params, paramUsage, "float64")

	expectedOrder := []FunctionParameterType{
		ParamTypeSeries,
		ParamTypeScalar,
		ParamTypeSeries,
		ParamTypeScalar,
	}

	for i, expected := range expectedOrder {
		paramType, exists := registry.GetParameterType("ordered", i)
		if !exists {
			t.Fatalf("Parameter at index %d not found", i)
		}
		if paramType != expected {
			t.Errorf("Order violation at index %d: got %v, want %v", i, paramType, expected)
		}
	}
}

/* TestSignatureRegistrar_EdgeCases validates boundary conditions */
func TestSignatureRegistrar_EdgeCases(t *testing.T) {
	t.Run("duplicate registration overwrites", func(t *testing.T) {
		registry := NewFunctionSignatureRegistry()
		registrar := NewSignatureRegistrar(registry)

		params1 := []ast.Identifier{{Name: "param1"}}
		usage1 := map[string]ParameterUsageType{"param1": ParameterUsageScalar}
		registrar.RegisterArrowFunction("func", params1, usage1, "float64")

		params2 := []ast.Identifier{{Name: "param2"}}
		usage2 := map[string]ParameterUsageType{"param2": ParameterUsageSeries}
		registrar.RegisterArrowFunction("func", params2, usage2, "float64")

		sig, exists := registry.Get("func")
		if !exists {
			t.Fatal("Function not found after duplicate registration")
		}
		if len(sig.Parameters) != 1 {
			t.Errorf("Expected 1 parameter, got %d", len(sig.Parameters))
		}
		if sig.Parameters[0] != ParamTypeSeries {
			t.Error("Expected series parameter from second registration")
		}
	})

	t.Run("empty function name", func(t *testing.T) {
		registry := NewFunctionSignatureRegistry()
		registrar := NewSignatureRegistrar(registry)

		params := []ast.Identifier{{Name: "param"}}
		usage := map[string]ParameterUsageType{"param": ParameterUsageScalar}
		registrar.RegisterArrowFunction("", params, usage, "float64")

		sig, exists := registry.Get("")
		if !exists {
			t.Error("Empty function name should be registered")
		}
		if sig == nil {
			t.Fatal("Expected signature for empty function name")
		}
	})

	t.Run("nil parameter usage map", func(t *testing.T) {
		registry := NewFunctionSignatureRegistry()
		registrar := NewSignatureRegistrar(registry)

		params := []ast.Identifier{
			{Name: "param1"},
			{Name: "param2"},
		}
		registrar.RegisterArrowFunction("nilUsage", params, nil, "float64")

		sig, exists := registry.Get("nilUsage")
		if !exists {
			t.Fatal("Function not registered with nil usage map")
		}
		if len(sig.Parameters) != 2 {
			t.Errorf("Expected 2 parameters, got %d", len(sig.Parameters))
		}
		for i, param := range sig.Parameters {
			if param != ParamTypeScalar {
				t.Errorf("Parameter %d should default to scalar, got %v", i, param)
			}
		}
	})

	t.Run("empty return type", func(t *testing.T) {
		registry := NewFunctionSignatureRegistry()
		registrar := NewSignatureRegistrar(registry)

		params := []ast.Identifier{{Name: "param"}}
		usage := map[string]ParameterUsageType{"param": ParameterUsageScalar}
		registrar.RegisterArrowFunction("noReturn", params, usage, "")

		sig, exists := registry.Get("noReturn")
		if !exists {
			t.Fatal("Function not registered with empty return type")
		}
		if sig.ReturnType != "" {
			t.Errorf("Expected empty return type, got %s", sig.ReturnType)
		}
	})
}

/* TestSignatureRegistrar_Integration validates full workflow */
func TestSignatureRegistrar_Integration(t *testing.T) {
	registry := NewFunctionSignatureRegistry()
	registrar := NewSignatureRegistrar(registry)

	params := []ast.Identifier{
		{Name: "src"},
		{Name: "len"},
	}
	paramUsage := map[string]ParameterUsageType{
		"src": ParameterUsageSeries,
		"len": ParameterUsageScalar,
	}

	registrar.RegisterArrowFunction("myFunc", params, paramUsage, "float64")

	paramType, exists := registry.GetParameterType("myFunc", 0)
	if !exists {
		t.Fatal("Parameter type not found for index 0")
	}
	if paramType != ParamTypeSeries {
		t.Errorf("First parameter should be series, got %v", paramType)
	}

	paramType, exists = registry.GetParameterType("myFunc", 1)
	if !exists {
		t.Fatal("Parameter type not found for index 1")
	}
	if paramType != ParamTypeScalar {
		t.Errorf("Second parameter should be scalar, got %v", paramType)
	}

	sig, exists := registry.Get("myFunc")
	if !exists {
		t.Fatal("Function signature not found")
	}
	if sig.Name != "myFunc" {
		t.Errorf("Function name mismatch: got %s, want myFunc", sig.Name)
	}
	if sig.ReturnType != "float64" {
		t.Errorf("Return type mismatch: got %s, want float64", sig.ReturnType)
	}
}

/* TestSignatureRegistrar_MultipleRegistrations validates registry accumulation */
func TestSignatureRegistrar_MultipleRegistrations(t *testing.T) {
	registry := NewFunctionSignatureRegistry()
	registrar := NewSignatureRegistrar(registry)

	registrar.RegisterArrowFunction("func1",
		[]ast.Identifier{{Name: "p1"}},
		map[string]ParameterUsageType{"p1": ParameterUsageScalar},
		"float64")

	registrar.RegisterArrowFunction("func2",
		[]ast.Identifier{{Name: "p2"}},
		map[string]ParameterUsageType{"p2": ParameterUsageSeries},
		"float64")

	registrar.RegisterArrowFunction("func3",
		[]ast.Identifier{{Name: "p3"}},
		map[string]ParameterUsageType{"p3": ParameterUsageScalar},
		"float64")

	_, exists1 := registry.Get("func1")
	_, exists2 := registry.Get("func2")
	_, exists3 := registry.Get("func3")

	if !exists1 || !exists2 || !exists3 {
		t.Error("All registered functions should be retrievable")
	}
}
