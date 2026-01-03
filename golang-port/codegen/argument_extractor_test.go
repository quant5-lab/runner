package codegen

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestExtractNamedArgument_Found(t *testing.T) {
	g := &generator{}
	extractor := &ArgumentExtractor{generator: g}

	// Named args are packed into ObjectExpression by converter
	args := []ast.Expression{
		&ast.Literal{Value: "positional1"},
		&ast.ObjectExpression{
			NodeType: ast.TypeObjectExpression,
			Properties: []ast.Property{
				{
					NodeType: ast.TypeProperty,
					Key:      &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "stop"},
					Value:    &ast.Literal{NodeType: ast.TypeLiteral, Value: 48000.0},
				},
			},
		},
	}

	code, found := extractor.ExtractNamedArgument(args, "stop")
	if !found {
		t.Fatal("Expected stop argument to be found")
	}
	if code != "48000.00" {
		t.Errorf("Expected '48000.00', got %q", code)
	}
}

func TestExtractNamedArgument_NotFound(t *testing.T) {
	g := &generator{}
	extractor := &ArgumentExtractor{generator: g}

	// ObjectExpression with different named arg
	args := []ast.Expression{
		&ast.ObjectExpression{
			NodeType: ast.TypeObjectExpression,
			Properties: []ast.Property{
				{
					NodeType: ast.TypeProperty,
					Key:      &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "limit"},
					Value:    &ast.Literal{NodeType: ast.TypeLiteral, Value: 58000.0},
				},
			},
		},
	}

	_, found := extractor.ExtractNamedArgument(args, "stop")
	if found {
		t.Error("Expected stop argument to not be found")
	}
}

func TestExtractPositionalArgument_Valid(t *testing.T) {
	g := &generator{}
	extractor := &ArgumentExtractor{generator: g}

	args := []ast.Expression{
		&ast.Literal{Value: 48000.0},
		&ast.Literal{Value: 58000.0},
	}

	code, found := extractor.ExtractPositionalArgument(args, 0)
	if !found {
		t.Fatal("Expected positional argument at index 0")
	}
	if code != "48000.00" {
		t.Errorf("Expected '48000.00', got %q", code)
	}
}

func TestExtractPositionalArgument_OutOfBounds(t *testing.T) {
	g := &generator{}
	extractor := &ArgumentExtractor{generator: g}

	args := []ast.Expression{
		&ast.Literal{Value: 48000.0},
	}

	_, found := extractor.ExtractPositionalArgument(args, 5)
	if found {
		t.Error("Expected out of bounds to return not found")
	}
}

func TestExtractNamedOrPositional_NamedTakesPrecedence(t *testing.T) {
	g := &generator{}
	extractor := &ArgumentExtractor{generator: g}

	// Positional + ObjectExpression with named args
	args := []ast.Expression{
		&ast.Literal{Value: 99999.0}, // Positional at index 0
		&ast.ObjectExpression{
			NodeType: ast.TypeObjectExpression,
			Properties: []ast.Property{
				{
					NodeType: ast.TypeProperty,
					Key:      &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "stop"},
					Value:    &ast.Literal{NodeType: ast.TypeLiteral, Value: 48000.0},
				},
			},
		},
	}

	code := extractor.ExtractNamedOrPositional(args, "stop", 0, "math.NaN()")
	if code != "48000.00" {
		t.Errorf("Expected named argument (48000.00), got %q", code)
	}
}

func TestExtractNamedOrPositional_FallbackToPositional(t *testing.T) {
	g := &generator{}
	extractor := &ArgumentExtractor{generator: g}

	args := []ast.Expression{
		&ast.Literal{Value: 48000.0}, // Positional at index 0
	}

	code := extractor.ExtractNamedOrPositional(args, "stop", 0, "math.NaN()")
	if code != "48000.00" {
		t.Errorf("Expected positional fallback (48000.00), got %q", code)
	}
}

func TestExtractNamedOrPositional_UseDefault(t *testing.T) {
	g := &generator{}
	extractor := &ArgumentExtractor{generator: g}

	args := []ast.Expression{}

	code := extractor.ExtractNamedOrPositional(args, "stop", 0, "math.NaN()")
	if code != "math.NaN()" {
		t.Errorf("Expected default value, got %q", code)
	}
}
