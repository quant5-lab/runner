package parser

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestBreakContinue_InForLoop(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		bodyLen  int
		breakIdx int
		contIdx  int
	}{
		{
			name: "break as only body statement",
			source: `for i = 0 to 10
    break`,
			bodyLen:  1,
			breakIdx: 0,
			contIdx:  -1,
		},
		{
			name: "continue as only body statement",
			source: `for i = 0 to 10
    continue`,
			bodyLen:  1,
			breakIdx: -1,
			contIdx:  0,
		},
		{
			name: "break after assignment",
			source: `for i = 0 to 10
    x = i
    break`,
			bodyLen:  2,
			breakIdx: 1,
			contIdx:  -1,
		},
		{
			name: "continue before expression",
			source: `for i = 0 to 10
    continue
    x = i`,
			bodyLen:  2,
			breakIdx: -1,
			contIdx:  0,
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

			if len(forStmt.Body) != tt.bodyLen {
				t.Fatalf("Expected %d body statements, got %d", tt.bodyLen, len(forStmt.Body))
			}

			if tt.breakIdx >= 0 {
				if forStmt.Body[tt.breakIdx].Core.Break == nil {
					t.Errorf("Expected break at index %d", tt.breakIdx)
				}
			}
			if tt.contIdx >= 0 {
				if forStmt.Body[tt.contIdx].Core.Continue == nil {
					t.Errorf("Expected continue at index %d", tt.contIdx)
				}
			}
		})
	}
}

func TestBreakContinue_InForInLoop(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		bodyLen  int
		breakIdx int
		contIdx  int
	}{
		{
			name: "break in for-in single form",
			source: `for val in myArray
    break`,
			bodyLen:  1,
			breakIdx: 0,
			contIdx:  -1,
		},
		{
			name: "continue in for-in tuple form",
			source: `for [i, val] in myArray
    continue`,
			bodyLen:  1,
			breakIdx: -1,
			contIdx:  0,
		},
		{
			name: "break and continue in for-in body",
			source: `for val in prices
    x = val
    continue
    break`,
			bodyLen:  3,
			breakIdx: 2,
			contIdx:  1,
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

			forIn := script.Statements[0].Core.ForIn
			if forIn == nil {
				t.Fatal("Expected ForInStatement, got nil")
			}

			if len(forIn.Body) != tt.bodyLen {
				t.Fatalf("Expected %d body statements, got %d", tt.bodyLen, len(forIn.Body))
			}

			if tt.breakIdx >= 0 {
				if forIn.Body[tt.breakIdx].Core.Break == nil {
					t.Errorf("Expected break at index %d", tt.breakIdx)
				}
			}
			if tt.contIdx >= 0 {
				if forIn.Body[tt.contIdx].Core.Continue == nil {
					t.Errorf("Expected continue at index %d", tt.contIdx)
				}
			}
		})
	}
}

func TestBreakContinue_InsideIfInLoop(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name: "break inside if in for loop",
			source: `for i = 0 to 10
    if i == 5
        break`,
		},
		{
			name: "continue inside if in for-in loop",
			source: `for val in myArray
    if val == 0
        continue`,
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
				t.Fatalf("Expected 1 top-level statement, got %d", len(script.Statements))
			}
		})
	}
}

func TestBreakContinue_ASTConversion(t *testing.T) {
	tests := []struct {
		name         string
		source       string
		expectedType ast.NodeType
	}{
		{
			name: "break converts to BreakStatement",
			source: `for i = 0 to 10
    break`,
			expectedType: ast.TypeBreakStatement,
		},
		{
			name: "continue converts to ContinueStatement",
			source: `for i = 0 to 10
    continue`,
			expectedType: ast.TypeContinueStatement,
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

			forStmt, ok := program.Body[0].(*ast.ForStatement)
			if !ok {
				t.Fatalf("Expected ForStatement, got %T", program.Body[0])
			}

			if len(forStmt.Body) != 1 {
				t.Fatalf("Expected 1 body node, got %d", len(forStmt.Body))
			}

			if forStmt.Body[0].Type() != tt.expectedType {
				t.Errorf("Expected %s, got %s", tt.expectedType, forStmt.Body[0].Type())
			}
		})
	}
}

func TestBreakContinue_ConditionalBreakInForIn(t *testing.T) {
	source := `for [i, val] in prices
    if val > 100
        break
    total := total + val
    total`

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

	forIn, ok := program.Body[0].(*ast.ForInStatement)
	if !ok {
		t.Fatalf("Expected ForInStatement, got %T", program.Body[0])
	}

	if len(forIn.Body) != 3 {
		t.Fatalf("Expected 3 body nodes, got %d", len(forIn.Body))
	}

	ifStmt, ok := forIn.Body[0].(*ast.IfStatement)
	if !ok {
		t.Fatalf("Expected IfStatement as first body node, got %T", forIn.Body[0])
	}

	if len(ifStmt.Consequent) != 1 {
		t.Fatalf("Expected 1 consequent node, got %d", len(ifStmt.Consequent))
	}

	if ifStmt.Consequent[0].Type() != ast.TypeBreakStatement {
		t.Errorf("Expected BreakStatement in if body, got %s", ifStmt.Consequent[0].Type())
	}
}

func TestBreakContinue_ContinueSkipEvenInForIn(t *testing.T) {
	source := `for [i, val] in prices
    if i == 2
        continue
    total := total + val`

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

	forIn, ok := program.Body[0].(*ast.ForInStatement)
	if !ok {
		t.Fatalf("Expected ForInStatement, got %T", program.Body[0])
	}

	ifStmt, ok := forIn.Body[0].(*ast.IfStatement)
	if !ok {
		t.Fatalf("Expected IfStatement as first body node, got %T", forIn.Body[0])
	}

	if ifStmt.Consequent[0].Type() != ast.TypeContinueStatement {
		t.Errorf("Expected ContinueStatement in if body, got %s", ifStmt.Consequent[0].Type())
	}
}

func TestBreakContinue_BothInSameLoop(t *testing.T) {
	source := `for i = 0 to 100
    if i == 50
        break
    if i == 3
        continue
    x = i`

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

	forStmt, ok := program.Body[0].(*ast.ForStatement)
	if !ok {
		t.Fatalf("Expected ForStatement, got %T", program.Body[0])
	}

	if len(forStmt.Body) != 3 {
		t.Fatalf("Expected 3 body nodes, got %d", len(forStmt.Body))
	}

	breakIf, ok := forStmt.Body[0].(*ast.IfStatement)
	if !ok {
		t.Fatalf("Expected IfStatement at index 0, got %T", forStmt.Body[0])
	}
	if breakIf.Consequent[0].Type() != ast.TypeBreakStatement {
		t.Errorf("Expected BreakStatement in first if, got %s", breakIf.Consequent[0].Type())
	}

	contIf, ok := forStmt.Body[1].(*ast.IfStatement)
	if !ok {
		t.Fatalf("Expected IfStatement at index 1, got %T", forStmt.Body[1])
	}
	if contIf.Consequent[0].Type() != ast.TypeContinueStatement {
		t.Errorf("Expected ContinueStatement in second if, got %s", contIf.Consequent[0].Type())
	}
}
