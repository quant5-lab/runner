package parser

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
)

/* Tests for IF statement indentation with INDENT/DEDENT */

// TestIfStatement_IndentationLevels verifies IF statements with various indentation
func TestIfStatement_IndentationLevels(t *testing.T) {
	tests := []struct {
		name          string
		source        string
		expectedStmts int
	}{
		{
			name: "single statement body",
			source: `if condition
    a = 1`,
			expectedStmts: 1,
		},
		{
			name: "two statement body",
			source: `if condition
    a = 1
    b = 2`,
			expectedStmts: 2,
		},
		{
			name: "three statement body",
			source: `if x > 0
    a = x + 1
    b = a * 2
    c = b - 3`,
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
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			if len(script.Statements) != 1 {
				t.Fatalf("Expected 1 statement, got %d", len(script.Statements))
			}

			ifStmt := script.Statements[0].If
			if ifStmt == nil {
				t.Fatal("Expected IfStatement, got nil")
			}

			if len(ifStmt.Body) != tt.expectedStmts {
				t.Errorf("Expected %d body statements, got %d", tt.expectedStmts, len(ifStmt.Body))
			}
		})
	}
}

// TestIfStatement_MultipleSequential verifies multiple IF statements in sequence
func TestIfStatement_MultipleSequential(t *testing.T) {
	tests := []struct {
		name        string
		source      string
		expectedIfs int
	}{
		{
			name: "two IFs - no blank line",
			source: `if condition1
    a = 1
if condition2
    b = 2`,
			expectedIfs: 2,
		},
		{
			name: "two IFs - with blank line",
			source: `if condition1
    a = 1

if condition2
    b = 2`,
			expectedIfs: 2,
		},
		{
			name: "three IFs",
			source: `if x > 0
    a = 1
if x < 0
    b = 2
if x == 0
    c = 3`,
			expectedIfs: 3,
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

			if len(script.Statements) != tt.expectedIfs {
				t.Fatalf("Expected %d IF statements, got %d", tt.expectedIfs, len(script.Statements))
			}

			for i := 0; i < tt.expectedIfs; i++ {
				if script.Statements[i].If == nil {
					t.Errorf("Statement %d: expected IF, got nil", i)
				}
			}
		})
	}
}

// TestIfStatement_WithEmptyLines verifies empty lines within IF bodies
func TestIfStatement_WithEmptyLines(t *testing.T) {
	tests := []struct {
		name          string
		source        string
		expectedStmts int
	}{
		{
			name: "single empty line in middle",
			source: `if condition
    a = 1

    b = 2`,
			expectedStmts: 2,
		},
		{
			name: "multiple empty lines",
			source: `if condition
    a = 1


    b = 2`,
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

			ifStmt := script.Statements[0].If
			if ifStmt == nil {
				t.Fatal("Expected IF statement")
			}

			if len(ifStmt.Body) != tt.expectedStmts {
				t.Errorf("Expected %d body statements, got %d", tt.expectedStmts, len(ifStmt.Body))
			}
		})
	}
}

// TestIfStatement_WithComments verifies comment handling in IF bodies
func TestIfStatement_WithComments(t *testing.T) {
	tests := []struct {
		name          string
		source        string
		expectedStmts int
	}{
		{
			name: "comment before body",
			source: `if condition
    // Set value
    a = 1`,
			expectedStmts: 1,
		},
		{
			name: "comments between statements",
			source: `if condition
    a = 1
    // Calculate b
    b = a * 2`,
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

			ifStmt := script.Statements[0].If
			if ifStmt == nil {
				t.Fatal("Expected IF statement")
			}

			if len(ifStmt.Body) != tt.expectedStmts {
				t.Errorf("Expected %d body statements, got %d", tt.expectedStmts, len(ifStmt.Body))
			}
		})
	}
}

// TestIfStatement_MixedWithOtherStatements verifies IF mixed with assignments/functions
func TestIfStatement_MixedWithOtherStatements(t *testing.T) {
	tests := []struct {
		name            string
		source          string
		expectedPattern []string
	}{
		{
			name: "assignment then IF",
			source: `value = 10
if value > 5
    result = value * 2`,
			expectedPattern: []string{"assign", "if"},
		},
		{
			name: "IF then assignment",
			source: `if condition
    temp = 1
result = temp + 1`,
			expectedPattern: []string{"if", "assign"},
		},
		{
			name: "multiple mixed",
			source: `a = 1
if a > 0
    b = 2
c = 3`,
			expectedPattern: []string{"assign", "if", "assign"},
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
				t.Fatalf("Expected %d statements, got %d", len(tt.expectedPattern), len(script.Statements))
			}

			for i, expected := range tt.expectedPattern {
				stmt := script.Statements[i]
				var actual string
				if stmt.If != nil {
					actual = "if"
				} else if stmt.Assignment != nil {
					actual = "assign"
				} else if stmt.Reassignment != nil {
					actual = "reassign"
				} else {
					actual = "other"
				}

				if actual != expected {
					t.Errorf("Statement %d: expected %s, got %s", i, expected, actual)
				}
			}
		})
	}
}

// TestIfStatement_InsideFunctionBody verifies IF statements within functions
func TestIfStatement_InsideFunctionBody(t *testing.T) {
	tests := []struct {
		name           string
		source         string
		expectedIfInFn bool
	}{
		{
			name: "single IF in function",
			source: `func(x) =>
    result = 0
    if x > 0
        result := x
    result`,
			expectedIfInFn: true,
		},
		{
			name: "multiple IFs in function",
			source: `func(x) =>
    result = 0
    if x > 10
        result := 10
    if x < 0
        result := 0
    result`,
			expectedIfInFn: true,
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

			funcDecl := script.Statements[0].FunctionDecl
			if funcDecl == nil {
				t.Fatal("Expected function declaration")
			}

			body := funcDecl.MultiLineBody
			if body == nil {
				body = []*Statement{}
			}

			hasIf := false
			for _, stmt := range body {
				if stmt.If != nil {
					hasIf = true
					break
				}
			}

			if hasIf != tt.expectedIfInFn {
				t.Errorf("Expected hasIf=%v, got %v", tt.expectedIfInFn, hasIf)
			}
		})
	}
}

// TestIfStatement_NestedReassignments verifies reassignments in IF bodies
func TestIfStatement_NestedReassignments(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name: "single reassignment",
			source: `result = 0
if condition
    result := 1`,
		},
		{
			name: "multiple reassignments",
			source: `a = 0
b = 0
if condition
    a := 1
    b := 2`,
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

			// Find IF statement
			var ifStmt *IfStatement
			for _, stmt := range script.Statements {
				if stmt.If != nil {
					ifStmt = stmt.If
					break
				}
			}

			if ifStmt == nil {
				t.Fatal("Expected IF statement")
			}

			// Verify at least one reassignment in body
			hasReassign := false
			for _, stmt := range ifStmt.Body {
				if stmt.Reassignment != nil {
					hasReassign = true
					break
				}
			}

			if !hasReassign {
				t.Error("Expected at least one reassignment in IF body")
			}
		})
	}
}

// TestIfStatement_Converter verifies AST conversion to ESTree
func TestIfStatement_Converter(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name: "simple IF",
			source: `if condition
    a = 1`,
		},
		{
			name: "IF with multiple statements",
			source: `if x > 0
    a = x + 1
    b = a * 2`,
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

			if len(program.Body) == 0 {
				t.Fatal("Program body is empty")
			}

			// Should be IfStatement
			ifNode, ok := program.Body[0].(*ast.IfStatement)
			if !ok {
				t.Fatalf("Expected IfStatement, got %T", program.Body[0])
			}

			if len(ifNode.Consequent) == 0 {
				t.Error("IF consequent is empty")
			}
		})
	}
}

// TestIfStatement_RealWorldPatterns verifies actual PineScript patterns
func TestIfStatement_RealWorldPatterns(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name: "strategy entry condition",
			source: `longCondition = crossover(sma(close, 50), sma(close, 200))
if longCondition
    strategy.entry("Long", strategy.long)`,
		},
		{
			name: "variable state update",
			source: `isUptrend = false
if close > sma(close, 20)
    isUptrend := true`,
		},
		{
			name: "conditional calculation",
			source: `result = 0
if high - low > atr
    diff = high - low
    result := diff * 100`,
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

			if len(program.Body) == 0 {
				t.Fatal("Program body is empty")
			}
		})
	}
}

// TestIfStatement_EdgeCases verifies error handling
func TestIfStatement_EdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		source    string
		shouldErr bool
	}{
		{
			name: "IF without indent - parser strict",
			source: `if condition
a = 1`,
			shouldErr: true, // Proper INDENT/DEDENT required for if body
		},
		{
			name: "IF with inconsistent indent - lexer lenient",
			source: `if condition
    a = 1
  b = 2`,
			shouldErr: false, // Lexer treats any dedent as valid
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewParser()
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			_, err = p.ParseBytes("test.pine", []byte(tt.source))
			if tt.shouldErr && err == nil {
				t.Error("Expected parse error, got nil")
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}
		})
	}
}
