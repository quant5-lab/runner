package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestArrayConstructorHandler_CanHandle(t *testing.T) {
	handler := NewArrayConstructorHandler()

	tests := []struct {
		funcName string
		expected bool
	}{
		{"array.new_float", true},
		{"array.new_int", true},
		{"array.new_bool", true},
		{"array.new_color", true},
		{"array.from", true},
		{"array.new_label", true},
		{"array.new_line", true},
		{"array.new_box", true},
		{"array.new_table", true},
		{"array.new_linefill", true},
		{"array.new_string", true},
		{"array.push", false},
		{"array.get", false},
		{"ta.sma", false},
	}

	for _, tt := range tests {
		t.Run(tt.funcName, func(t *testing.T) {
			result := handler.CanHandle(tt.funcName)
			if result != tt.expected {
				t.Errorf("CanHandle(%q) = %v, want %v", tt.funcName, result, tt.expected)
			}
		})
	}
}

func TestArrayConstructorHandler_GenerateNewFloat_Empty(t *testing.T) {
	handler := NewArrayConstructorHandler()
	g := createMockGenerator()

	call := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "array"},
			Property: &ast.Identifier{Name: "new_float"},
		},
		Arguments: []ast.Expression{},
	}

	code, err := handler.GenerateCode(g, call)
	if err != nil {
		t.Fatalf("GenerateCode failed: %v", err)
	}

	if code != "[]float64{}" {
		t.Errorf("Expected []float64{}, got %q", code)
	}
}

func TestArrayConstructorHandler_GenerateNewFloat_SizeOnly(t *testing.T) {
	handler := NewArrayConstructorHandler()
	g := createMockGenerator()

	call := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "array"},
			Property: &ast.Identifier{Name: "new_float"},
		},
		Arguments: []ast.Expression{
			&ast.Literal{Value: 10.0},
		},
	}

	code, err := handler.GenerateCode(g, call)
	if err != nil {
		t.Fatalf("GenerateCode failed: %v", err)
	}

	if !strings.Contains(code, "make([]float64, int(10))") {
		t.Errorf("Expected make([]float64, int(10)), got %q", code)
	}
}

func TestArrayConstructorHandler_GenerateNewFloat_SizeAndInitial(t *testing.T) {
	handler := NewArrayConstructorHandler()
	g := createMockGenerator()

	call := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "array"},
			Property: &ast.Identifier{Name: "new_float"},
		},
		Arguments: []ast.Expression{
			&ast.Literal{Value: 5.0},
			&ast.Literal{Value: 99.0},
		},
	}

	code, err := handler.GenerateCode(g, call)
	if err != nil {
		t.Fatalf("GenerateCode failed: %v", err)
	}

	if !strings.Contains(code, "arrayops.NewArrayWithValue") {
		t.Errorf("Expected arrayops.NewArrayWithValue call, got %q", code)
	}
	if !strings.Contains(code, "int(5)") {
		t.Errorf("Expected size parameter int(5), got %q", code)
	}
	if !strings.Contains(code, "99") {
		t.Errorf("Expected initial value 99, got %q", code)
	}
}

func TestArrayConstructorHandler_GenerateFrom(t *testing.T) {
	handler := NewArrayConstructorHandler()
	g := createMockGenerator()

	call := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "array"},
			Property: &ast.Identifier{Name: "from"},
		},
		Arguments: []ast.Expression{
			&ast.Literal{Value: 1.0},
			&ast.Literal{Value: 2.0},
			&ast.Literal{Value: 3.0},
		},
	}

	code, err := handler.GenerateCode(g, call)
	if err != nil {
		t.Fatalf("GenerateCode failed: %v", err)
	}

	if !strings.Contains(code, "[]float64{") {
		t.Errorf("Expected []float64{...}, got %q", code)
	}
	if !strings.Contains(code, "1") || !strings.Contains(code, "2") || !strings.Contains(code, "3") {
		t.Errorf("Expected elements 1, 2, 3, got %q", code)
	}
}

func TestArrayConstructorHandler_GenerateNewColor(t *testing.T) {
	handler := NewArrayConstructorHandler()
	g := createMockGenerator()

	call := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "array"},
			Property: &ast.Identifier{Name: "new_color"},
		},
		Arguments: []ast.Expression{
			&ast.Literal{Value: 3.0},
			&ast.Identifier{Name: "color.red"},
		},
	}

	code, err := handler.GenerateCode(g, call)
	if err != nil {
		t.Fatalf("GenerateCode failed: %v", err)
	}

	if !strings.Contains(code, "arrayops.NewArrayWithValue") {
		t.Errorf("Expected arrayops.NewArrayWithValue call, got %q", code)
	}
	if !strings.Contains(code, "int(3)") {
		t.Errorf("Expected size parameter int(3), got %q", code)
	}
}

func TestArrayConstructorHandler_GenerateDrawingArrayStub(t *testing.T) {
	handler := NewArrayConstructorHandler()
	g := createMockGenerator()

	tests := []struct {
		name     string
		property string
		args     []ast.Expression
		want     string
	}{
		{"empty label", "new_label", nil, "[]float64{}"},
		{"empty line", "new_line", nil, "[]float64{}"},
		{"empty box", "new_box", nil, "[]float64{}"},
		{"empty table", "new_table", nil, "[]float64{}"},
		{"empty linefill", "new_linefill", nil, "[]float64{}"},
		{"label with size", "new_label", []ast.Expression{&ast.Literal{Value: 5.0}}, "make([]float64, int(5))"},
		{"line with zero size", "new_line", []ast.Expression{&ast.Literal{Value: 0.0}}, "make([]float64, int(0))"},
		{"box with large size", "new_box", []ast.Expression{&ast.Literal{Value: 500.0}}, "make([]float64, int(500))"},
		{"table with identifier", "new_table", []ast.Expression{&ast.Identifier{Name: "maxCount"}}, "make([]float64, int(maxCount))"},
		{"initial value ignored", "new_label", []ast.Expression{&ast.Literal{Value: 10.0}, &ast.Identifier{Name: "na"}}, "make([]float64, int(10))"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			call := &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "array"},
					Property: &ast.Identifier{Name: tt.property},
				},
				Arguments: tt.args,
			}

			code, err := handler.GenerateCode(g, call)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !strings.Contains(code, tt.want) {
				t.Errorf("got %q, want substring %q", code, tt.want)
			}
		})
	}
}

func createMockGenerator() *generator {
	g := &generator{
		variables: make(map[string]string),
		indent:    1,
	}
	return g
}
