package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestArrayConstructorHandler_StringArrayEmpty(t *testing.T) {
	handler := NewArrayConstructorHandler()
	g := createMockGenerator()

	call := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "array"},
			Property: &ast.Identifier{Name: "new_string"},
		},
		Arguments: []ast.Expression{},
	}

	code, err := handler.GenerateCode(g, call)
	if err != nil {
		t.Fatalf("GenerateCode error: %v", err)
	}

	expected := "[]string{}"
	if code != expected {
		t.Errorf("empty string array: expected '%s', got '%s'", expected, code)
	}
}

func TestArrayConstructorHandler_StringArraySizeOnly(t *testing.T) {
	handler := NewArrayConstructorHandler()
	g := createMockGenerator()

	call := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "array"},
			Property: &ast.Identifier{Name: "new_string"},
		},
		Arguments: []ast.Expression{
			&ast.Literal{Value: 5.0},
		},
	}

	code, err := handler.GenerateCode(g, call)
	if err != nil {
		t.Fatalf("GenerateCode error: %v", err)
	}

	expected := "make([]string, int(5))"
	if code != expected {
		t.Errorf("string array with size: expected '%s', got '%s'", expected, code)
	}
}

func TestArrayConstructorHandler_StringArrayWithInitial(t *testing.T) {
	handler := NewArrayConstructorHandler()
	g := createMockGenerator()

	call := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "array"},
			Property: &ast.Identifier{Name: "new_string"},
		},
		Arguments: []ast.Expression{
			&ast.Literal{Value: 3.0},
			&ast.Identifier{Name: `"hello"`},
		},
	}

	code, err := handler.GenerateCode(g, call)
	if err != nil {
		t.Fatalf("GenerateCode error: %v", err)
	}

	if !strings.Contains(code, "arrayops.NewStringArrayWithValue") {
		t.Errorf("Expected NewStringArrayWithValue call, got %q", code)
	}
	if !strings.Contains(code, "int(3)") {
		t.Errorf("Expected size int(3), got %q", code)
	}
}
