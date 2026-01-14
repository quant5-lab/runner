package codegen

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestStringExtractor_Extract_ValidString(t *testing.T) {
	extractor := StringExtractor{}
	lit := &ast.Literal{Value: "test"}

	result, ok := extractor.Extract(lit)

	if !ok {
		t.Error("Expected extraction to succeed")
	}
	if result != "test" {
		t.Errorf("Expected 'test', got '%v'", result)
	}
}

func TestStringExtractor_Extract_NonLiteral(t *testing.T) {
	extractor := StringExtractor{}
	id := &ast.Identifier{Name: "test"}

	_, ok := extractor.Extract(id)

	if ok {
		t.Error("Expected extraction to fail for non-literal")
	}
}

func TestStringExtractor_Extract_WrongType(t *testing.T) {
	extractor := StringExtractor{}
	lit := &ast.Literal{Value: 42}

	_, ok := extractor.Extract(lit)

	if ok {
		t.Error("Expected extraction to fail for non-string literal")
	}
}

func TestIntExtractor_Extract_Int(t *testing.T) {
	extractor := IntExtractor{}
	lit := &ast.Literal{Value: 42}

	result, ok := extractor.Extract(lit)

	if !ok {
		t.Error("Expected extraction to succeed")
	}
	if result != 42 {
		t.Errorf("Expected 42, got %v", result)
	}
}

func TestIntExtractor_Extract_Float(t *testing.T) {
	extractor := IntExtractor{}
	lit := &ast.Literal{Value: 42.7}

	result, ok := extractor.Extract(lit)

	if !ok {
		t.Error("Expected extraction to succeed")
	}
	if result != 42 {
		t.Errorf("Expected 42, got %v", result)
	}
}

func TestIntExtractor_Extract_String(t *testing.T) {
	extractor := IntExtractor{}
	lit := &ast.Literal{Value: "42"}

	_, ok := extractor.Extract(lit)

	if ok {
		t.Error("Expected extraction to fail for string")
	}
}

func TestFloatExtractor_Extract_Float(t *testing.T) {
	extractor := FloatExtractor{}
	lit := &ast.Literal{Value: 3.14}

	result, ok := extractor.Extract(lit)

	if !ok {
		t.Error("Expected extraction to succeed")
	}
	if result != 3.14 {
		t.Errorf("Expected 3.14, got %v", result)
	}
}

func TestFloatExtractor_Extract_Int(t *testing.T) {
	extractor := FloatExtractor{}
	lit := &ast.Literal{Value: 42}

	result, ok := extractor.Extract(lit)

	if !ok {
		t.Error("Expected extraction to succeed")
	}
	if result != 42.0 {
		t.Errorf("Expected 42.0, got %v", result)
	}
}

func TestFloatExtractor_Extract_String(t *testing.T) {
	extractor := FloatExtractor{}
	lit := &ast.Literal{Value: "3.14"}

	_, ok := extractor.Extract(lit)

	if ok {
		t.Error("Expected extraction to fail for string")
	}
}

func TestBoolExtractor_Extract_True(t *testing.T) {
	extractor := BoolExtractor{}
	lit := &ast.Literal{Value: true}

	result, ok := extractor.Extract(lit)

	if !ok {
		t.Error("Expected extraction to succeed")
	}
	if result != true {
		t.Errorf("Expected true, got %v", result)
	}
}

func TestBoolExtractor_Extract_False(t *testing.T) {
	extractor := BoolExtractor{}
	lit := &ast.Literal{Value: false}

	result, ok := extractor.Extract(lit)

	if !ok {
		t.Error("Expected extraction to succeed")
	}
	if result != false {
		t.Errorf("Expected false, got %v", result)
	}
}

func TestBoolExtractor_Extract_NonBool(t *testing.T) {
	extractor := BoolExtractor{}
	lit := &ast.Literal{Value: 1}

	_, ok := extractor.Extract(lit)

	if ok {
		t.Error("Expected extraction to fail for non-bool")
	}
}

func TestIdentifierExtractor_Extract_Valid(t *testing.T) {
	extractor := IdentifierExtractor{}
	id := &ast.Identifier{Name: "myVar"}

	result, ok := extractor.Extract(id)

	if !ok {
		t.Error("Expected extraction to succeed")
	}
	if result != "myVar" {
		t.Errorf("Expected 'myVar', got '%v'", result)
	}
}

func TestIdentifierExtractor_Extract_Literal(t *testing.T) {
	extractor := IdentifierExtractor{}
	lit := &ast.Literal{Value: "myVar"}

	_, ok := extractor.Extract(lit)

	if ok {
		t.Error("Expected extraction to fail for literal")
	}
}

func TestIdentifierExtractor_Extract_Empty(t *testing.T) {
	extractor := IdentifierExtractor{}
	id := &ast.Identifier{Name: ""}

	result, ok := extractor.Extract(id)

	if !ok {
		t.Error("Expected extraction to succeed")
	}
	if result != "" {
		t.Errorf("Expected empty string, got '%v'", result)
	}
}
