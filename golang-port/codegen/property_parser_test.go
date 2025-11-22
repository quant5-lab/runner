package codegen

import (
	"testing"

	"github.com/borisquantlab/pinescript-go/ast"
)

func TestPropertyParser_ParseString_ValidString(t *testing.T) {
	parser := NewPropertyParser()
	obj := &ast.ObjectExpression{
		Properties: []ast.Property{
			{
				Key:   &ast.Identifier{Name: "title"},
				Value: &ast.Literal{Value: "My Title"},
			},
		},
	}

	result, ok := parser.ParseString(obj, "title")

	if !ok {
		t.Error("Expected parsing to succeed")
	}
	if result != "My Title" {
		t.Errorf("Expected 'My Title', got '%s'", result)
	}
}

func TestPropertyParser_ParseString_MissingKey(t *testing.T) {
	parser := NewPropertyParser()
	obj := &ast.ObjectExpression{
		Properties: []ast.Property{
			{
				Key:   &ast.Identifier{Name: "other"},
				Value: &ast.Literal{Value: "value"},
			},
		},
	}

	_, ok := parser.ParseString(obj, "title")

	if ok {
		t.Error("Expected parsing to fail for missing key")
	}
}

func TestPropertyParser_ParseString_WrongType(t *testing.T) {
	parser := NewPropertyParser()
	obj := &ast.ObjectExpression{
		Properties: []ast.Property{
			{
				Key:   &ast.Identifier{Name: "title"},
				Value: &ast.Literal{Value: 42},
			},
		},
	}

	_, ok := parser.ParseString(obj, "title")

	if ok {
		t.Error("Expected parsing to fail for wrong type")
	}
}

func TestPropertyParser_ParseString_EmptyObject(t *testing.T) {
	parser := NewPropertyParser()
	obj := &ast.ObjectExpression{
		Properties: []ast.Property{},
	}

	_, ok := parser.ParseString(obj, "title")

	if ok {
		t.Error("Expected parsing to fail for empty object")
	}
}

func TestPropertyParser_ParseInt_ValidInt(t *testing.T) {
	parser := NewPropertyParser()
	obj := &ast.ObjectExpression{
		Properties: []ast.Property{
			{
				Key:   &ast.Identifier{Name: "linewidth"},
				Value: &ast.Literal{Value: 2},
			},
		},
	}

	result, ok := parser.ParseInt(obj, "linewidth")

	if !ok {
		t.Error("Expected parsing to succeed")
	}
	if result != 2 {
		t.Errorf("Expected 2, got %d", result)
	}
}

func TestPropertyParser_ParseInt_Float(t *testing.T) {
	parser := NewPropertyParser()
	obj := &ast.ObjectExpression{
		Properties: []ast.Property{
			{
				Key:   &ast.Identifier{Name: "linewidth"},
				Value: &ast.Literal{Value: 2.7},
			},
		},
	}

	result, ok := parser.ParseInt(obj, "linewidth")

	if !ok {
		t.Error("Expected parsing to succeed")
	}
	if result != 2 {
		t.Errorf("Expected 2, got %d", result)
	}
}

func TestPropertyParser_ParseFloat_ValidFloat(t *testing.T) {
	parser := NewPropertyParser()
	obj := &ast.ObjectExpression{
		Properties: []ast.Property{
			{
				Key:   &ast.Identifier{Name: "transparency"},
				Value: &ast.Literal{Value: 0.5},
			},
		},
	}

	result, ok := parser.ParseFloat(obj, "transparency")

	if !ok {
		t.Error("Expected parsing to succeed")
	}
	if result != 0.5 {
		t.Errorf("Expected 0.5, got %f", result)
	}
}

func TestPropertyParser_ParseFloat_Int(t *testing.T) {
	parser := NewPropertyParser()
	obj := &ast.ObjectExpression{
		Properties: []ast.Property{
			{
				Key:   &ast.Identifier{Name: "transparency"},
				Value: &ast.Literal{Value: 1},
			},
		},
	}

	result, ok := parser.ParseFloat(obj, "transparency")

	if !ok {
		t.Error("Expected parsing to succeed")
	}
	if result != 1.0 {
		t.Errorf("Expected 1.0, got %f", result)
	}
}

func TestPropertyParser_ParseBool_True(t *testing.T) {
	parser := NewPropertyParser()
	obj := &ast.ObjectExpression{
		Properties: []ast.Property{
			{
				Key:   &ast.Identifier{Name: "display"},
				Value: &ast.Literal{Value: true},
			},
		},
	}

	result, ok := parser.ParseBool(obj, "display")

	if !ok {
		t.Error("Expected parsing to succeed")
	}
	if result != true {
		t.Errorf("Expected true, got %v", result)
	}
}

func TestPropertyParser_ParseBool_False(t *testing.T) {
	parser := NewPropertyParser()
	obj := &ast.ObjectExpression{
		Properties: []ast.Property{
			{
				Key:   &ast.Identifier{Name: "display"},
				Value: &ast.Literal{Value: false},
			},
		},
	}

	result, ok := parser.ParseBool(obj, "display")

	if !ok {
		t.Error("Expected parsing to succeed")
	}
	if result != false {
		t.Errorf("Expected false, got %v", result)
	}
}

func TestPropertyParser_ParseIdentifier_ValidIdentifier(t *testing.T) {
	parser := NewPropertyParser()
	obj := &ast.ObjectExpression{
		Properties: []ast.Property{
			{
				Key:   &ast.Identifier{Name: "color"},
				Value: &ast.Identifier{Name: "blue"},
			},
		},
	}

	result, ok := parser.ParseIdentifier(obj, "color")

	if !ok {
		t.Error("Expected parsing to succeed")
	}
	if result != "blue" {
		t.Errorf("Expected 'blue', got '%s'", result)
	}
}

func TestPropertyParser_ParseIdentifier_Literal(t *testing.T) {
	parser := NewPropertyParser()
	obj := &ast.ObjectExpression{
		Properties: []ast.Property{
			{
				Key:   &ast.Identifier{Name: "color"},
				Value: &ast.Literal{Value: "blue"},
			},
		},
	}

	_, ok := parser.ParseIdentifier(obj, "color")

	if ok {
		t.Error("Expected parsing to fail for literal value")
	}
}

func TestPropertyParser_FindProperty_MultipleProperties(t *testing.T) {
	parser := NewPropertyParser()
	obj := &ast.ObjectExpression{
		Properties: []ast.Property{
			{
				Key:   &ast.Identifier{Name: "title"},
				Value: &ast.Literal{Value: "Title"},
			},
			{
				Key:   &ast.Identifier{Name: "linewidth"},
				Value: &ast.Literal{Value: 2},
			},
			{
				Key:   &ast.Identifier{Name: "color"},
				Value: &ast.Identifier{Name: "red"},
			},
		},
	}

	title, ok1 := parser.ParseString(obj, "title")
	linewidth, ok2 := parser.ParseInt(obj, "linewidth")
	color, ok3 := parser.ParseIdentifier(obj, "color")

	if !ok1 || title != "Title" {
		t.Error("Failed to parse title")
	}
	if !ok2 || linewidth != 2 {
		t.Error("Failed to parse linewidth")
	}
	if !ok3 || color != "red" {
		t.Error("Failed to parse color")
	}
}

func TestPropertyParser_FindProperty_NonIdentifierKey(t *testing.T) {
	parser := NewPropertyParser()
	obj := &ast.ObjectExpression{
		Properties: []ast.Property{
			{
				Key:   &ast.Literal{Value: "title"},
				Value: &ast.Literal{Value: "My Title"},
			},
		},
	}

	_, ok := parser.ParseString(obj, "title")

	if ok {
		t.Error("Expected parsing to fail for non-identifier key")
	}
}
