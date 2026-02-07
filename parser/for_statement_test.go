package parser

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestForStatement_BasicSyntax(t *testing.T) {
	tests := []struct {
		name          string
		source        string
		expectError   bool
		expectedStmts int
	}{
		{
			name: "simple ascending range",
			source: `for i = 0 to 9
    x = 1`,
			expectedStmts: 1,
		},
		{
			name: "descending range",
			source: `for i = 10 to 0
    x = 1`,
			expectedStmts: 1,
		},
		{
			name: "single iteration",
			source: `for i = 5 to 5
    x = 1`,
			expectedStmts: 1,
		},
		{
			name: "multiple statements in body",
			source: `for i = 0 to 5
    a = 1
    b = 2
    c = 3`,
			expectedStmts: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewParser()
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			script, err := p.ParseBytes("test.pine", []byte(tt.source))
			if tt.expectError {
				if err == nil {
					t.Fatal("Expected parse error but got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			if len(script.Statements) != 1 {
				t.Fatalf("Expected 1 top-level statement, got %d", len(script.Statements))
			}

			forStmt := script.Statements[0].Core.For
			if forStmt == nil {
				t.Fatal("Expected ForStatement, got nil")
			}

			if len(forStmt.Body) != tt.expectedStmts {
				t.Errorf("Expected %d body statements, got %d", tt.expectedStmts, len(forStmt.Body))
			}
		})
	}
}

func TestForStatement_WithStep(t *testing.T) {
	tests := []struct {
		name        string
		source      string
		hasStep     bool
		expectError bool
	}{
		{
			name: "positive step",
			source: `for i = 0 to 10 by 2
    x = 1`,
			hasStep: true,
		},
		{
			name: "negative step",
			source: `for i = 10 to 0 by -1
    x = 1`,
			hasStep: true,
		},
		{
			name: "no step specified",
			source: `for i = 0 to 10
    x = 1`,
			hasStep: false,
		},
		{
			name: "step with expression",
			source: `for i = 0 to 100 by 5
    x = 1`,
			hasStep: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewParser()
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			script, err := p.ParseBytes("test.pine", []byte(tt.source))
			if tt.expectError {
				if err == nil {
					t.Fatal("Expected parse error but got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			forStmt := script.Statements[0].Core.For
			if forStmt == nil {
				t.Fatal("Expected ForStatement, got nil")
			}

			if tt.hasStep && forStmt.Step == nil {
				t.Error("Expected step expression but got nil")
			}
			if !tt.hasStep && forStmt.Step != nil {
				t.Error("Expected no step expression but got one")
			}
		})
	}
}

func TestForStatement_ComplexBounds(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name: "arithmetic expression in from",
			source: `for i = bar_index - 10 to bar_index
    x = 1`,
		},
		{
			name: "arithmetic expression in to",
			source: `for i = 0 to bar_index + 5
    x = 1`,
		},
		{
			name: "both bounds are expressions",
			source: `for i = start - offset to end + offset
    x = 1`,
		},
		{
			name: "step is expression",
			source: `for i = 0 to 100 by stepSize * 2
    x = 1`,
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

			forStmt := script.Statements[0].Core.For
			if forStmt == nil {
				t.Fatal("Expected ForStatement, got nil")
			}

			if forStmt.From == nil {
				t.Error("Expected From expression but got nil")
			}
			if forStmt.To == nil {
				t.Error("Expected To expression but got nil")
			}
		})
	}
}

func TestForStatement_NestedLoops(t *testing.T) {
	tests := []struct {
		name        string
		source      string
		outerStmts  int
		hasNested   bool
		nestedDepth int
	}{
		{
			name: "simple nested loop",
			source: `for i = 0 to 2
    for j = 0 to 2
        x = 1`,
			outerStmts:  1,
			hasNested:   true,
			nestedDepth: 2,
		},
		{
			name: "triple nested loop",
			source: `for i = 0 to 2
    for j = 0 to 2
        for k = 0 to 2
            x = 1`,
			outerStmts:  1,
			hasNested:   true,
			nestedDepth: 3,
		},
		{
			name: "nested with multiple statements",
			source: `for i = 0 to 5
    a = i
    for j = 0 to 3
        b = j
    c = i + 1`,
			outerStmts:  3,
			hasNested:   true,
			nestedDepth: 2,
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

			outerFor := script.Statements[0].Core.For
			if outerFor == nil {
				t.Fatal("Expected outer ForStatement, got nil")
			}

			if len(outerFor.Body) != tt.outerStmts {
				t.Errorf("Expected %d outer body statements, got %d", tt.outerStmts, len(outerFor.Body))
			}

			if tt.hasNested {
				foundNested := false
				for _, stmt := range outerFor.Body {
					if stmt.Core.For != nil {
						foundNested = true
						break
					}
				}
				if !foundNested {
					t.Error("Expected nested for loop but found none")
				}
			}
		})
	}
}

func TestForStatement_WithVariableOperations(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name: "series subscript access",
			source: `for i = 0 to 10
    sum := sum + close[i]`,
		},
		{
			name: "multiple subscript accesses",
			source: `for i = 0 to 10
    result := high[i] - low[i]`,
		},
		{
			name: "local variable declaration",
			source: `for i = 0 to 10
    temp = i * 2`,
		},
		{
			name: "outer variable reassignment",
			source: `for i = 0 to 10
    counter := counter + 1`,
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

			forStmt := script.Statements[0].Core.For
			if forStmt == nil {
				t.Fatal("Expected ForStatement, got nil")
			}

			if len(forStmt.Body) == 0 {
				t.Error("Expected at least one statement in for body")
			}
		})
	}
}

func TestForStatement_MultipleSequential(t *testing.T) {
	tests := []struct {
		name         string
		source       string
		expectedFors int
	}{
		{
			name: "two sequential for loops",
			source: `for i = 0 to 5
    x = 1
for j = 0 to 3
    y = 2`,
			expectedFors: 2,
		},
		{
			name: "three sequential for loops",
			source: `for i = 0 to 5
    x = 1
for j = 0 to 3
    y = 2
for k = 0 to 10
    z = 3`,
			expectedFors: 3,
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

			if len(script.Statements) != tt.expectedFors {
				t.Errorf("Expected %d for statements, got %d", tt.expectedFors, len(script.Statements))
			}

			for i, stmt := range script.Statements {
				if stmt.Core.For == nil {
					t.Errorf("Statement %d is not a ForStatement", i)
				}
			}
		})
	}
}

func TestForStatement_MixedWithOtherStatements(t *testing.T) {
	tests := []struct {
		name            string
		source          string
		expectedPattern []string
	}{
		{
			name: "assignment then for",
			source: `total = 0
for i = 0 to 10
    total := total + i`,
			expectedPattern: []string{"assign", "for"},
		},
		{
			name: "for then assignment",
			source: `for i = 0 to 10
    temp = i
result = temp * 2`,
			expectedPattern: []string{"for", "assign"},
		},
		{
			name: "multiple mixed statements",
			source: `start = 0
for i = start to 10
    x = i
end = x + 1`,
			expectedPattern: []string{"assign", "for", "assign"},
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

			if len(script.Statements) != len(tt.expectedPattern) {
				t.Errorf("Expected %d statements, got %d", len(tt.expectedPattern), len(script.Statements))
			}

			for i, expected := range tt.expectedPattern {
				if i >= len(script.Statements) {
					break
				}
				stmt := script.Statements[i]
				switch expected {
				case "for":
					if stmt.Core.For == nil {
						t.Errorf("Statement %d: expected for, got other type", i)
					}
				case "assign":
					if stmt.Core.Assignment == nil && stmt.Core.Reassignment == nil {
						t.Errorf("Statement %d: expected assignment, got other type", i)
					}
				}
			}
		})
	}
}

func TestForStatement_Converter(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name: "simple for loop",
			source: `for i = 0 to 10
    x = 1`,
		},
		{
			name: "for with step",
			source: `for i = 0 to 100 by 10
    x = i`,
		},
		{
			name: "nested for loops",
			source: `for i = 0 to 5
    for j = 0 to 5
        result := result + close[i] * close[j]`,
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
				t.Fatalf("Conversion to ESTree failed: %v", err)
			}

			if len(program.Body) == 0 {
				t.Fatal("Expected at least one node in program body")
			}

			forNode, ok := program.Body[0].(*ast.ForStatement)
			if !ok {
				t.Fatalf("Expected ForStatement in AST, got %T", program.Body[0])
			}

			if forNode.Counter == "" {
				t.Error("Expected counter variable name but got empty string")
			}

			if len(forNode.Body) == 0 {
				t.Error("Expected at least one statement in for body")
			}
		})
	}
}

func TestForStatement_EmptyLinesAndComments(t *testing.T) {
	tests := []struct {
		name          string
		source        string
		expectedStmts int
	}{
		{
			name: "empty line in body",
			source: `for i = 0 to 5
    x = 1

    y = 2`,
			expectedStmts: 2,
		},
		{
			name: "comment in body",
			source: `for i = 0 to 5
    // Calculate value
    x = 1`,
			expectedStmts: 1,
		},
		{
			name: "multiple empty lines",
			source: `for i = 0 to 5
    x = 1


    y = 2`,
			expectedStmts: 2,
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

			forStmt := script.Statements[0].Core.For
			if forStmt == nil {
				t.Fatal("Expected ForStatement, got nil")
			}

			if len(forStmt.Body) != tt.expectedStmts {
				t.Errorf("Expected %d body statements, got %d", tt.expectedStmts, len(forStmt.Body))
			}
		})
	}
}
