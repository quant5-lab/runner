package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestArrayMutatorCodegen_TypeAwareRouting(t *testing.T) {
	tests := []struct {
		name         string
		funcName     string
		elemType     ArrayElementType
		expectedCtor string
		expectedSfx  string
	}{
		{"push on float array", "array.push", ArrayElementFloat64, "arrayops.NewMutator()", "ArraySeries"},
		{"push on string array", "array.push", ArrayElementString, "arrayops.NewStringMutator()", "StringArraySeries"},
		{"pop on float array", "array.pop", ArrayElementFloat64, "arrayops.NewMutator()", "ArraySeries"},
		{"pop on string array", "array.pop", ArrayElementString, "arrayops.NewStringMutator()", "StringArraySeries"},
		{"reverse on float array", "array.reverse", ArrayElementFloat64, "arrayops.NewTransformer()", "ArraySeries"},
		{"reverse on string array", "array.reverse", ArrayElementString, "arrayops.NewStringTransformers()", "StringArraySeries"},
		{"sort on float array", "array.sort", ArrayElementFloat64, "arrayops.NewTransformer()", "ArraySeries"},
		{"sort on string array", "array.sort", ArrayElementString, "arrayops.NewStringTransformers()", "StringArraySeries"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newTestGenerator()
			g.arrayVariableRegistry.Register("data", tt.elemType)

			h := NewArrayMutatorCodegen()
			call := buildMutatorCall(tt.funcName, "data")
			code, err := h.GenerateCode(g, call)
			if err != nil {
				t.Fatalf("GenerateCode() error = %v", err)
			}

			if !strings.Contains(code, tt.expectedCtor) {
				t.Errorf("Expected constructor %q not found in: %s", tt.expectedCtor, code)
			}
			if !strings.Contains(code, "data"+tt.expectedSfx) {
				t.Errorf("Expected suffix %q not found in: %s", tt.expectedSfx, code)
			}
		})
	}
}

func TestArrayReaderCodegen_TypeAwareRouting(t *testing.T) {
	tests := []struct {
		name         string
		funcName     string
		elemType     ArrayElementType
		expectedCtor string
		expectedSfx  string
	}{
		{"first on float array", "array.first", ArrayElementFloat64, "arrayops.NewAccessor()", "ArraySeries"},
		{"first on string array", "array.first", ArrayElementString, "arrayops.NewStringAccessor()", "StringArraySeries"},
		{"last on float array", "array.last", ArrayElementFloat64, "arrayops.NewAccessor()", "ArraySeries"},
		{"last on string array", "array.last", ArrayElementString, "arrayops.NewStringAccessor()", "StringArraySeries"},
		{"includes on float array", "array.includes", ArrayElementFloat64, "arrayops.NewAccessor()", "ArraySeries"},
		{"includes on string array", "array.includes", ArrayElementString, "arrayops.NewStringAccessor()", "StringArraySeries"},
		{"slice on float array", "array.slice", ArrayElementFloat64, "arrayops.NewTransformer()", "ArraySeries"},
		{"slice on string array", "array.slice", ArrayElementString, "arrayops.NewStringTransformers()", "StringArraySeries"},
		{"join on float array", "array.join", ArrayElementFloat64, "arrayops.NewFormatters()", "ArraySeries"},
		{"join on string array", "array.join", ArrayElementString, "arrayops.NewStringFormatters()", "StringArraySeries"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newTestGenerator()
			g.arrayVariableRegistry.Register("data", tt.elemType)

			h := NewArrayReaderCodegen()
			call := buildReaderCall(tt.funcName, "data")
			code, err := h.GenerateCode(g, call)
			if err != nil {
				t.Fatalf("GenerateCode() error = %v", err)
			}

			if !strings.Contains(code, tt.expectedCtor) {
				t.Errorf("Expected constructor %q not found in: %s", tt.expectedCtor, code)
			}
			if !strings.Contains(code, "data"+tt.expectedSfx) {
				t.Errorf("Expected suffix %q not found in: %s", tt.expectedSfx, code)
			}
		})
	}
}

func TestArrayReaderCodegen_StatisticsGuard(t *testing.T) {
	statisticsMethods := []string{
		"array.sum", "array.avg", "array.min", "array.max",
		"array.median", "array.mode", "array.stdev", "array.variance",
		"array.range", "array.percentile_linear_interpolation",
		"array.percentile_nearest_rank", "array.percentrank",
		"array.standardize", "array.abs",
	}

	for _, method := range statisticsMethods {
		t.Run(method+" on string array", func(t *testing.T) {
			g := newTestGenerator()
			g.arrayVariableRegistry.Register("labels", ArrayElementString)

			h := NewArrayReaderCodegen()
			call := buildReaderCall(method, "labels")
			_, err := h.GenerateCode(g, call)
			if err == nil {
				t.Errorf("Expected error for %s on string array, got nil", method)
			}
			if err != nil && !strings.Contains(err.Error(), "string array") {
				t.Errorf("Error message should mention 'string array', got: %v", err)
			}
		})

		t.Run(method+" on float array", func(t *testing.T) {
			g := newTestGenerator()
			g.arrayVariableRegistry.Register("prices", ArrayElementFloat64)

			h := NewArrayReaderCodegen()
			call := buildReaderCall(method, "prices")
			code, err := h.GenerateCode(g, call)
			if err != nil {
				t.Errorf("Unexpected error for %s on float array: %v", method, err)
			}
			if code == "" {
				t.Error("Expected non-empty code for float array statistics")
			}
		})
	}
}

func TestArrayReaderCodegen_CovarianceWithMixedTypes(t *testing.T) {
	g := newTestGenerator()
	g.arrayVariableRegistry.Register("prices", ArrayElementFloat64)
	g.arrayVariableRegistry.Register("labels", ArrayElementString)

	h := NewArrayReaderCodegen()

	t.Run("covariance float + float", func(t *testing.T) {
		g.arrayVariableRegistry.Register("volumes", ArrayElementFloat64)
		call := &ast.CallExpression{
			Callee: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "array"},
				Property: &ast.Identifier{Name: "covariance"},
			},
			Arguments: []ast.Expression{
				&ast.Identifier{Name: "prices"},
				&ast.Identifier{Name: "volumes"},
			},
		}
		code, err := h.GenerateCode(g, call)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if !strings.Contains(code, "pricesArraySeries") {
			t.Error("Should contain pricesArraySeries")
		}
		if !strings.Contains(code, "volumesArraySeries") {
			t.Error("Should contain volumesArraySeries")
		}
	})

	t.Run("covariance string + float rejects string", func(t *testing.T) {
		call := &ast.CallExpression{
			Callee: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "array"},
				Property: &ast.Identifier{Name: "covariance"},
			},
			Arguments: []ast.Expression{
				&ast.Identifier{Name: "labels"},
				&ast.Identifier{Name: "prices"},
			},
		}
		_, err := h.GenerateCode(g, call)
		if err == nil {
			t.Error("Expected error for covariance with string array")
		}
	})
}

func TestArrayCallHandler_GetWithTypedSuffix(t *testing.T) {
	tests := []struct {
		name        string
		elemType    ArrayElementType
		expectedSfx string
	}{
		{"get on float array", ArrayElementFloat64, "ArraySeries"},
		{"get on string array", ArrayElementString, "StringArraySeries"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newTestGenerator()
			g.arrayVariableRegistry.Register("data", tt.elemType)

			h := NewArrayMethodCallHandler()
			call := &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "array"},
					Property: &ast.Identifier{Name: "get"},
				},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "data"},
					&ast.Literal{Value: 0},
				},
			}
			code, err := h.GenerateCode(g, call)
			if err != nil {
				t.Fatalf("GenerateCode() error = %v", err)
			}

			expectedVar := "data" + tt.expectedSfx
			if !strings.Contains(code, expectedVar) {
				t.Errorf("Expected %q in code, got: %s", expectedVar, code)
			}
		})
	}
}

func TestArrayCallHandler_SizeWithTypedSuffix(t *testing.T) {
	tests := []struct {
		name        string
		elemType    ArrayElementType
		expectedSfx string
	}{
		{"size on float array", ArrayElementFloat64, "ArraySeries"},
		{"size on string array", ArrayElementString, "StringArraySeries"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newTestGenerator()
			g.arrayVariableRegistry.Register("data", tt.elemType)

			h := NewArrayMethodCallHandler()
			call := &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "array"},
					Property: &ast.Identifier{Name: "size"},
				},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "data"},
				},
			}
			code, err := h.GenerateCode(g, call)
			if err != nil {
				t.Fatalf("GenerateCode() error = %v", err)
			}

			expectedVar := "data" + tt.expectedSfx
			if !strings.Contains(code, expectedVar) {
				t.Errorf("Expected %q in code, got: %s", expectedVar, code)
			}
		})
	}
}

func TestArrayMutatorCodegen_UnregisteredVariableError(t *testing.T) {
	g := newTestGenerator()
	h := NewArrayMutatorCodegen()

	call := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "array"},
			Property: &ast.Identifier{Name: "push"},
		},
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "unregistered"},
			&ast.Literal{Value: 1.0},
		},
	}

	_, err := h.GenerateCode(g, call)
	if err == nil {
		t.Error("Expected error for unregistered variable")
	}
	if err != nil && !strings.Contains(err.Error(), "not an array") {
		t.Errorf("Error should mention 'not an array', got: %v", err)
	}
}

func TestArrayReaderCodegen_UnregisteredVariableError(t *testing.T) {
	g := newTestGenerator()
	h := NewArrayReaderCodegen()

	call := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "array"},
			Property: &ast.Identifier{Name: "first"},
		},
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "unregistered"},
		},
	}

	_, err := h.GenerateCode(g, call)
	if err == nil {
		t.Error("Expected error for unregistered variable")
	}
	if err != nil && !strings.Contains(err.Error(), "not an array") {
		t.Errorf("Error should mention 'not an array', got: %v", err)
	}
}

func TestArrayTypedCodegen_ConcatPreservesBothTypes(t *testing.T) {
	g := newTestGenerator()
	g.arrayVariableRegistry.Register("arr1", ArrayElementFloat64)
	g.arrayVariableRegistry.Register("arr2", ArrayElementFloat64)

	h := NewArrayMutatorCodegen()
	call := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "array"},
			Property: &ast.Identifier{Name: "concat"},
		},
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "arr1"},
			&ast.Identifier{Name: "arr2"},
		},
	}

	code, err := h.GenerateCode(g, call)
	if err != nil {
		t.Fatalf("GenerateCode() error = %v", err)
	}

	if !strings.Contains(code, "arr1ArraySeries") {
		t.Error("Should contain arr1ArraySeries")
	}
	if !strings.Contains(code, "arr2ArraySeries") {
		t.Error("Should contain arr2ArraySeries")
	}
}

func buildMutatorCall(funcName, varName string) *ast.CallExpression {
	parts := strings.Split(funcName, ".")
	args := []ast.Expression{&ast.Identifier{Name: varName}}

	switch parts[1] {
	case "push", "unshift":
		args = append(args, &ast.Literal{Value: 1.0})
	case "set", "insert":
		args = append(args, &ast.Literal{Value: 0}, &ast.Literal{Value: 1.0})
	case "fill":
		args = append(args, &ast.Literal{Value: 1.0})
	}

	return &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: parts[0]},
			Property: &ast.Identifier{Name: parts[1]},
		},
		Arguments: args,
	}
}

func buildReaderCall(funcName, varName string) *ast.CallExpression {
	parts := strings.Split(funcName, ".")
	args := []ast.Expression{&ast.Identifier{Name: varName}}

	switch parts[1] {
	case "includes", "indexof", "lastindexof":
		args = append(args, &ast.Literal{Value: 1.0})
	case "percentile_linear_interpolation", "percentile_nearest_rank", "percentrank", "binary_search":
		args = append(args, &ast.Literal{Value: 50.0})
	}

	return &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: parts[0]},
			Property: &ast.Identifier{Name: parts[1]},
		},
		Arguments: args,
	}
}
