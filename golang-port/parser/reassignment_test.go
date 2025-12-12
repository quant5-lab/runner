package parser

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestReassignment_Simple(t *testing.T) {
	script := `//@version=5
x = 0.0
x := 10.0`

	p, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	parseResult, err := p.ParseBytes("test.pine", []byte(script))
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	converter := NewConverter()
	program, err := converter.ToESTree(parseResult)
	if err != nil {
		t.Fatalf("Conversion error: %v", err)
	}

	if len(program.Body) != 2 {
		t.Fatalf("Expected 2 statements, got %d", len(program.Body))
	}

	// First statement: declaration (=)
	varDecl1, ok := program.Body[0].(*ast.VariableDeclaration)
	if !ok {
		t.Fatalf("First statement is not VariableDeclaration, got %T", program.Body[0])
	}
	if varDecl1.Kind != "let" {
		t.Errorf("First statement Kind = %s, want 'let'", varDecl1.Kind)
	}
	if varDecl1.Declarations[0].ID.Name != "x" {
		t.Errorf("First statement variable name = %s, want 'x'", varDecl1.Declarations[0].ID.Name)
	}

	// Second statement: reassignment (:=)
	varDecl2, ok := program.Body[1].(*ast.VariableDeclaration)
	if !ok {
		t.Fatalf("Second statement is not VariableDeclaration, got %T", program.Body[1])
	}
	if varDecl2.Kind != "var" {
		t.Errorf("Second statement Kind = %s, want 'var'", varDecl2.Kind)
	}
	if varDecl2.Declarations[0].ID.Name != "x" {
		t.Errorf("Second statement variable name = %s, want 'x'", varDecl2.Declarations[0].ID.Name)
	}
}

func TestReassignment_WithTernary(t *testing.T) {
	script := `//@version=5
sr_xup = 0.0
sr_sup = true
sr_xup := sr_sup ? low : sr_xup[1]`

	p, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	parseResult, err := p.ParseBytes("test.pine", []byte(script))
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	converter := NewConverter()
	program, err := converter.ToESTree(parseResult)
	if err != nil {
		t.Fatalf("Conversion error: %v", err)
	}

	if len(program.Body) != 3 {
		t.Fatalf("Expected 3 statements, got %d", len(program.Body))
	}

	// Third statement: reassignment with ternary
	varDecl, ok := program.Body[2].(*ast.VariableDeclaration)
	if !ok {
		t.Fatalf("Third statement is not VariableDeclaration, got %T", program.Body[2])
	}
	if varDecl.Kind != "var" {
		t.Errorf("Reassignment Kind = %s, want 'var'", varDecl.Kind)
	}
	if varDecl.Declarations[0].ID.Name != "sr_xup" {
		t.Errorf("Variable name = %s, want 'sr_xup'", varDecl.Declarations[0].ID.Name)
	}

	// Verify the init is a ConditionalExpression
	_, ok = varDecl.Declarations[0].Init.(*ast.ConditionalExpression)
	if !ok {
		t.Errorf("Init is not ConditionalExpression, got %T", varDecl.Declarations[0].Init)
	}
}

func TestReassignment_MultiLine(t *testing.T) {
	script := `//@version=5
result = 0
condition = true
result := condition ? 
   100 : 
   200`

	p, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	parseResult, err := p.ParseBytes("test.pine", []byte(script))
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	converter := NewConverter()
	program, err := converter.ToESTree(parseResult)
	if err != nil {
		t.Fatalf("Conversion error: %v", err)
	}

	if len(program.Body) != 3 {
		t.Fatalf("Expected 3 statements, got %d", len(program.Body))
	}

	// Third statement: multi-line reassignment
	varDecl, ok := program.Body[2].(*ast.VariableDeclaration)
	if !ok {
		t.Fatalf("Third statement is not VariableDeclaration, got %T", program.Body[2])
	}
	if varDecl.Kind != "var" {
		t.Errorf("Multi-line reassignment Kind = %s, want 'var'", varDecl.Kind)
	}
}
