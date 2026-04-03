package parser

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
)

/* Tests for tuple destructuring assignment parsing and conversion */

// TestTupleAssignment_ElementCounts verifies parsing with various element counts
func TestTupleAssignment_ElementCounts(t *testing.T) {
	tests := []struct {
		name          string
		source        string
		expectedNames []string
	}{
		{
			name:          "two elements",
			source:        `[a, b] = func()`,
			expectedNames: []string{"a", "b"},
		},
		{
			name:          "three elements - ADX pattern",
			source:        `[ADX, up, down] = adx(14, 16)`,
			expectedNames: []string{"ADX", "up", "down"},
		},
		{
			name:          "four elements",
			source:        `[w, x, y, z] = multi()`,
			expectedNames: []string{"w", "x", "y", "z"},
		},
		{
			name:          "five elements",
			source:        `[a, b, c, d, e] = func()`,
			expectedNames: []string{"a", "b", "c", "d", "e"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewParser()
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			script, err := p.ParseBytes("test.pine", []byte(tt.source))
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			if len(script.Statements) != 1 {
				t.Fatalf("Expected 1 statement, got %d", len(script.Statements))
			}

			stmt := script.Statements[0]
			if stmt.Core.TupleAssignment == nil {
				t.Fatal("Expected TupleAssignment, got nil")
			}

			if len(stmt.Core.TupleAssignment.Names) != len(tt.expectedNames) {
				t.Fatalf("Expected %d names, got %d", len(tt.expectedNames), len(stmt.Core.TupleAssignment.Names))
			}

			for i, expected := range tt.expectedNames {
				if stmt.Core.TupleAssignment.Names[i] != expected {
					t.Errorf("Expected name[%d] '%s', got '%s'", i, expected, stmt.Core.TupleAssignment.Names[i])
				}
			}
		})
	}
}

// TestTupleAssignment_WhitespaceVariations verifies parser handles whitespace correctly
func TestTupleAssignment_WhitespaceVariations(t *testing.T) {
	tests := []struct {
		name          string
		source        string
		expectedNames []string
	}{
		{
			name:          "no whitespace",
			source:        `[a,b,c]=func()`,
			expectedNames: []string{"a", "b", "c"},
		},
		{
			name:          "standard whitespace",
			source:        `[a, b, c] = func()`,
			expectedNames: []string{"a", "b", "c"},
		},
		{
			name:          "extra whitespace after commas",
			source:        `[a,  b,  c] = func()`,
			expectedNames: []string{"a", "b", "c"},
		},
		{
			name:          "whitespace inside brackets",
			source:        `[ a, b, c ] = func()`,
			expectedNames: []string{"a", "b", "c"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewParser()
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			script, err := p.ParseBytes("test.pine", []byte(tt.source))
			if err != nil {
				t.Fatalf("Parse failed for '%s': %v", tt.source, err)
			}

			stmt := script.Statements[0]
			if stmt.Core.TupleAssignment == nil {
				t.Fatal("Expected TupleAssignment, got nil")
			}

			if len(stmt.Core.TupleAssignment.Names) != len(tt.expectedNames) {
				t.Fatalf("Expected %d names, got %d", len(tt.expectedNames), len(stmt.Core.TupleAssignment.Names))
			}

			for i, expected := range tt.expectedNames {
				if stmt.Core.TupleAssignment.Names[i] != expected {
					t.Errorf("Expected name[%d] '%s', got '%s'", i, expected, stmt.Core.TupleAssignment.Names[i])
				}
			}
		})
	}
}

// TestTupleAssignment_ConverterValidation verifies ESTree conversion creates correct AST
func TestTupleAssignment_ConverterValidation(t *testing.T) {
	tests := []struct {
		name          string
		source        string
		expectedNames []string
		expectedKind  string
	}{
		{
			name:          "basic two-element",
			source:        `[a, b] = func()`,
			expectedNames: []string{"a", "b"},
			expectedKind:  "let",
		},
		{
			name:          "three-element with realistic names",
			source:        `[ADX, plusDI, minusDI] = ta.dmi(14, 16)`,
			expectedNames: []string{"ADX", "plusDI", "minusDI"},
			expectedKind:  "let",
		},
		{
			name:          "four-element",
			source:        `[open, high, low, close] = security("AAPL", "D")`,
			expectedNames: []string{"open", "high", "low", "close"},
			expectedKind:  "let",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewParser()
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			script, err := p.ParseBytes("test.pine", []byte(tt.source))
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			converter := NewConverter()
			program, err := converter.ToESTree(script)
			if err != nil {
				t.Fatalf("Conversion failed: %v", err)
			}

			if len(program.Body) != 1 {
				t.Fatalf("Expected 1 statement in body, got %d", len(program.Body))
			}

			varDecl, ok := program.Body[0].(*ast.VariableDeclaration)
			if !ok {
				t.Fatalf("Expected VariableDeclaration, got %T", program.Body[0])
			}

			if varDecl.Kind != tt.expectedKind {
				t.Errorf("Expected Kind '%s', got '%s'", tt.expectedKind, varDecl.Kind)
			}

			if len(varDecl.Declarations) != 1 {
				t.Fatalf("Expected 1 declarator, got %d", len(varDecl.Declarations))
			}

			declarator := varDecl.Declarations[0]
			arrayPattern, ok := declarator.ID.(*ast.ArrayPattern)
			if !ok {
				t.Fatalf("Expected ArrayPattern as ID, got %T", declarator.ID)
			}

			if len(arrayPattern.Elements) != len(tt.expectedNames) {
				t.Fatalf("Expected %d elements in ArrayPattern, got %d", len(tt.expectedNames), len(arrayPattern.Elements))
			}

			for i, expected := range tt.expectedNames {
				if arrayPattern.Elements[i].Name != expected {
					t.Errorf("Expected element[%d] name '%s', got '%s'", i, expected, arrayPattern.Elements[i].Name)
				}
			}

			if declarator.Init == nil {
				t.Fatal("Expected Init expression, got nil")
			}
		})
	}
}

// TestTupleAssignment_BackwardCompatibility ensures regular assignments still work
func TestTupleAssignment_BackwardCompatibility(t *testing.T) {
	tests := []struct {
		name         string
		source       string
		expectedName string
	}{
		{
			name:         "simple assignment",
			source:       `a = func()`,
			expectedName: "a",
		},
		{
			name:         "assignment with arguments",
			source:       `sma20 = ta.sma(close, 20)`,
			expectedName: "sma20",
		},
		{
			name:         "assignment with subscript",
			source:       `prev = close[1]`,
			expectedName: "prev",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewParser()
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			script, err := p.ParseBytes("test.pine", []byte(tt.source))
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			if len(script.Statements) != 1 {
				t.Fatalf("Expected 1 statement, got %d", len(script.Statements))
			}

			stmt := script.Statements[0]
			if stmt.Core.Assignment == nil {
				t.Fatal("Expected Assignment (not TupleAssignment), got nil")
			}

			if stmt.Core.TupleAssignment != nil {
				t.Fatal("Expected nil TupleAssignment, but got non-nil")
			}

			converter := NewConverter()
			program, err := converter.ToESTree(script)
			if err != nil {
				t.Fatalf("Conversion failed: %v", err)
			}

			varDecl := program.Body[0].(*ast.VariableDeclaration)
			declarator := varDecl.Declarations[0]

			ident, ok := declarator.ID.(*ast.Identifier)
			if !ok {
				t.Fatalf("Expected ID to be Identifier, got %T", declarator.ID)
			}

			if ident.Name != tt.expectedName {
				t.Errorf("Expected name '%s', got '%s'", tt.expectedName, ident.Name)
			}
		})
	}
}

// TestTupleAssignment_MixedStatements verifies tuple assignments work alongside other statements
func TestTupleAssignment_MixedStatements(t *testing.T) {
	source := `//@version=5
indicator("Test")
a = 10
[b, c] = func1()
d = 20
[e, f, g] = func2()
h = 30`

	p, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	script, err := p.ParseBytes("test.pine", []byte(source))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	converter := NewConverter()
	program, err := converter.ToESTree(script)
	if err != nil {
		t.Fatalf("Conversion failed: %v", err)
	}

	/* indicator() + 5 assignments = 6 statements */
	if len(program.Body) < 6 {
		t.Fatalf("Expected at least 6 statements, got %d", len(program.Body))
	}

	/* Verify statement 2 is regular assignment (a = 10) */
	varDecl1 := program.Body[1].(*ast.VariableDeclaration)
	if ident, ok := varDecl1.Declarations[0].ID.(*ast.Identifier); !ok || ident.Name != "a" {
		t.Error("Statement 2 should be regular assignment 'a'")
	}

	/* Verify statement 3 is tuple assignment ([b, c] = func1()) */
	varDecl2 := program.Body[2].(*ast.VariableDeclaration)
	if arrayPattern, ok := varDecl2.Declarations[0].ID.(*ast.ArrayPattern); !ok {
		t.Error("Statement 3 should be tuple assignment")
	} else if len(arrayPattern.Elements) != 2 {
		t.Errorf("Statement 3 should have 2 elements, got %d", len(arrayPattern.Elements))
	}

	/* Verify statement 4 is regular assignment (d = 20) */
	varDecl3 := program.Body[3].(*ast.VariableDeclaration)
	if ident, ok := varDecl3.Declarations[0].ID.(*ast.Identifier); !ok || ident.Name != "d" {
		t.Error("Statement 4 should be regular assignment 'd'")
	}

	/* Verify statement 5 is tuple assignment ([e, f, g] = func2()) */
	varDecl4 := program.Body[4].(*ast.VariableDeclaration)
	if arrayPattern, ok := varDecl4.Declarations[0].ID.(*ast.ArrayPattern); !ok {
		t.Error("Statement 5 should be tuple assignment")
	} else if len(arrayPattern.Elements) != 3 {
		t.Errorf("Statement 5 should have 3 elements, got %d", len(arrayPattern.Elements))
	}

	/* Verify statement 6 is regular assignment (h = 30) */
	varDecl5 := program.Body[5].(*ast.VariableDeclaration)
	if ident, ok := varDecl5.Declarations[0].ID.(*ast.Identifier); !ok || ident.Name != "h" {
		t.Error("Statement 6 should be regular assignment 'h'")
	}
}

// TestTupleAssignment_ComplexRHS verifies various right-hand side expression types
func TestTupleAssignment_ComplexRHS(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name:   "function with multiple arguments",
			source: `[a, b] = ta.bb(close, 20, 2.0)`,
		},
		{
			name:   "function with nested calls",
			source: `[x, y] = func(ta.sma(close, 20))`,
		},
		{
			name:   "function with arithmetic in arguments",
			source: `[high, low] = range(close * 1.1, close * 0.9)`,
		},
		{
			name:   "function with identifiers as arguments",
			source: `[ADX, plusDI, minusDI] = ta.dmi(length, adxSmoothing)`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewParser()
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			script, err := p.ParseBytes("test.pine", []byte(tt.source))
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			converter := NewConverter()
			program, err := converter.ToESTree(script)
			if err != nil {
				t.Fatalf("Conversion failed: %v", err)
			}

			varDecl := program.Body[0].(*ast.VariableDeclaration)
			if _, ok := varDecl.Declarations[0].ID.(*ast.ArrayPattern); !ok {
				t.Fatalf("Expected ArrayPattern, got %T", varDecl.Declarations[0].ID)
			}

			if varDecl.Declarations[0].Init == nil {
				t.Fatal("Expected non-nil Init expression")
			}
		})
	}
}

// TestTupleAssignment_IdentifierNamingPatterns verifies various identifier naming conventions
func TestTupleAssignment_IdentifierNamingPatterns(t *testing.T) {
	tests := []struct {
		name          string
		source        string
		expectedNames []string
	}{
		{
			name:          "lowercase names",
			source:        `[low, high] = range()`,
			expectedNames: []string{"low", "high"},
		},
		{
			name:          "UPPERCASE names",
			source:        `[ADX, RSI, MACD] = indicators()`,
			expectedNames: []string{"ADX", "RSI", "MACD"},
		},
		{
			name:          "camelCase names",
			source:        `[fastMA, slowMA] = movingAverages()`,
			expectedNames: []string{"fastMA", "slowMA"},
		},
		{
			name:          "snake_case names",
			source:        `[upper_band, lower_band] = bb()`,
			expectedNames: []string{"upper_band", "lower_band"},
		},
		{
			name:          "mixed naming conventions",
			source:        `[ADX, plus_DI, minusDI] = dmi()`,
			expectedNames: []string{"ADX", "plus_DI", "minusDI"},
		},
		{
			name:          "names with numbers",
			source:        `[sma20, ema50, rsi14] = combo()`,
			expectedNames: []string{"sma20", "ema50", "rsi14"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewParser()
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			script, err := p.ParseBytes("test.pine", []byte(tt.source))
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			stmt := script.Statements[0]
			if stmt.Core.TupleAssignment == nil {
				t.Fatal("Expected TupleAssignment, got nil")
			}

			if len(stmt.Core.TupleAssignment.Names) != len(tt.expectedNames) {
				t.Fatalf("Expected %d names, got %d", len(tt.expectedNames), len(stmt.Core.TupleAssignment.Names))
			}

			for i, expected := range tt.expectedNames {
				if stmt.Core.TupleAssignment.Names[i] != expected {
					t.Errorf("Expected name[%d] '%s', got '%s'", i, expected, stmt.Core.TupleAssignment.Names[i])
				}
			}
		})
	}
}
