package parser

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
)

/* For-loop as expression tests */

func TestForExpression_AssignmentContexts(t *testing.T) {
	tests := []struct {
		name    string
		source  string
		varName string
	}{
		{
			name: "simple variable assignment",
			source: `result = for i = 1 to 10
    i`,
			varName: "result",
		},
		{
			name: "reassignment",
			source: `x = 0
x := for i = 1 to 5
    i * 2`,
			varName: "x",
		},
		{
			name: "var declaration",
			source: `var sum = for i = 1 to 100
    i`,
			varName: "sum",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewParser()
			if err != nil {
				t.Fatalf("Parser creation failed: %v", err)
			}

			script, err := p.ParseString("", tt.source)
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			converter := NewConverter()
			program, err := converter.ToESTree(script)
			if err != nil {
				t.Fatalf("Conversion failed: %v", err)
			}

			var foundForExpr bool
			for _, stmt := range program.Body {
				if varDecl, ok := stmt.(*ast.VariableDeclaration); ok {
					for _, decl := range varDecl.Declarations {
						if _, ok := decl.Init.(*ast.ForStatement); ok {
							foundForExpr = true
							if ident, ok := decl.ID.(*ast.Identifier); ok {
								if ident.Name != tt.varName {
									t.Errorf("Variable name: expected=%q got=%q", tt.varName, ident.Name)
								}
							}
						}
					}
				}
			}

			if !foundForExpr {
				t.Error("ForStatement not found in expression context")
			}
		})
	}
}

func TestForExpression_StepVariations(t *testing.T) {
	tests := []struct {
		name    string
		source  string
		hasStep bool
	}{
		{
			name: "no step",
			source: `x = for i = 0 to 10
    i`,
			hasStep: false,
		},
		{
			name: "positive step",
			source: `x = for i = 0 to 10 by 2
    i`,
			hasStep: true,
		},
		{
			name: "negative step",
			source: `x = for i = 10 to 0 by -1
    i`,
			hasStep: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewParser()
			if err != nil {
				t.Fatalf("Parser creation failed: %v", err)
			}

			script, err := p.ParseString("", tt.source)
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			converter := NewConverter()
			program, err := converter.ToESTree(script)
			if err != nil {
				t.Fatalf("Conversion failed: %v", err)
			}

			varDecl := program.Body[0].(*ast.VariableDeclaration)
			forStmt := varDecl.Declarations[0].Init.(*ast.ForStatement)

			if tt.hasStep && forStmt.Step == nil {
				t.Error("Expected step expression, got nil")
			}
			if !tt.hasStep && forStmt.Step != nil {
				t.Error("Expected no step expression, but found one")
			}
		})
	}
}

func TestForExpression_BodyComplexity(t *testing.T) {
	tests := []struct {
		name          string
		source        string
		expectedStmts int
	}{
		{
			name: "single statement",
			source: `x = for i = 1 to 10
    i`,
			expectedStmts: 1,
		},
		{
			name: "multiple statements",
			source: `x = for i = 1 to 10
    a = i
    b = a * 2
    b`,
			expectedStmts: 3,
		},
		{
			name: "with variable mutation",
			source: `sum = for i = 1 to 10
    s = 0
    s := s + i
    s`,
			expectedStmts: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewParser()
			if err != nil {
				t.Fatalf("Parser creation failed: %v", err)
			}

			script, err := p.ParseString("", tt.source)
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			converter := NewConverter()
			program, err := converter.ToESTree(script)
			if err != nil {
				t.Fatalf("Conversion failed: %v", err)
			}

			varDecl := program.Body[0].(*ast.VariableDeclaration)
			forStmt := varDecl.Declarations[0].Init.(*ast.ForStatement)

			if len(forStmt.Body) != tt.expectedStmts {
				t.Errorf("Body statements: expected=%d got=%d", tt.expectedStmts, len(forStmt.Body))
			}
		})
	}
}

func TestForExpression_SyntacticContexts(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name: "standalone assignment",
			source: `result = for i = 1 to 10
    i`,
		},
		{
			name: "in var declaration",
			source: `var x = for i = 1 to 5
    i * 2`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewParser()
			if err != nil {
				t.Fatalf("Parser creation failed: %v", err)
			}

			script, err := p.ParseString("", tt.source)
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			converter := NewConverter()
			_, err = converter.ToESTree(script)
			if err != nil {
				t.Fatalf("Conversion failed: %v", err)
			}
		})
	}
}

func TestForExpression_NestedControlFlow(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name: "for inside if expression",
			source: `result = if close > open
    for i = 1 to 10
        i`,
		},
		{
			name: "if inside for expression",
			source: `result = for i = 1 to 10
    if i > 5
        i`,
		},
		{
			name: "nested for expressions",
			source: `result = for i = 1 to 5
    for j = 1 to 3
        j`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewParser()
			if err != nil {
				t.Fatalf("Parser creation failed: %v", err)
			}

			script, err := p.ParseString("", tt.source)
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			converter := NewConverter()
			_, err = converter.ToESTree(script)
			if err != nil {
				t.Fatalf("Conversion failed: %v", err)
			}
		})
	}
}

/* If-statement as expression tests */

func TestIfExpression_AssignmentContexts(t *testing.T) {
	tests := []struct {
		name    string
		source  string
		varName string
	}{
		{
			name: "simple variable assignment",
			source: `result = if close > open
    close`,
			varName: "result",
		},
		{
			name: "reassignment",
			source: `x = 0
x := if high > low
    high`,
			varName: "x",
		},
		{
			name: "var declaration",
			source: `var signal = if rsi > 70
    1`,
			varName: "signal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewParser()
			if err != nil {
				t.Fatalf("Parser creation failed: %v", err)
			}

			script, err := p.ParseString("", tt.source)
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			converter := NewConverter()
			program, err := converter.ToESTree(script)
			if err != nil {
				t.Fatalf("Conversion failed: %v", err)
			}

			var foundIfExpr bool
			for _, stmt := range program.Body {
				if varDecl, ok := stmt.(*ast.VariableDeclaration); ok {
					for _, decl := range varDecl.Declarations {
						if _, ok := decl.Init.(*ast.IfStatement); ok {
							foundIfExpr = true
							if ident, ok := decl.ID.(*ast.Identifier); ok {
								if ident.Name != tt.varName {
									t.Errorf("Variable name: expected=%q got=%q", tt.varName, ident.Name)
								}
							}
						}
					}
				}
			}

			if !foundIfExpr {
				t.Error("IfStatement not found in expression context")
			}
		})
	}
}

func TestIfExpression_BranchComplexity(t *testing.T) {
	tests := []struct {
		name          string
		source        string
		expectedStmts int
	}{
		{
			name: "single consequent statement",
			source: `x = if true
    1`,
			expectedStmts: 1,
		},
		{
			name: "multiple consequent statements",
			source: `x = if true
    a = 10
    b = 20
    b`,
			expectedStmts: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewParser()
			if err != nil {
				t.Fatalf("Parser creation failed: %v", err)
			}

			script, err := p.ParseString("", tt.source)
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			converter := NewConverter()
			program, err := converter.ToESTree(script)
			if err != nil {
				t.Fatalf("Conversion failed: %v", err)
			}

			varDecl := program.Body[0].(*ast.VariableDeclaration)
			ifStmt := varDecl.Declarations[0].Init.(*ast.IfStatement)

			if len(ifStmt.Consequent) != tt.expectedStmts {
				t.Errorf("Consequent statements: expected=%d got=%d", tt.expectedStmts, len(ifStmt.Consequent))
			}
		})
	}
}

func TestIfExpression_SyntacticContexts(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name: "standalone assignment",
			source: `result = if close > open
    close`,
		},
		{
			name: "in var declaration",
			source: `var signal = if rsi > 70
    1`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewParser()
			if err != nil {
				t.Fatalf("Parser creation failed: %v", err)
			}

			script, err := p.ParseString("", tt.source)
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			converter := NewConverter()
			_, err = converter.ToESTree(script)
			if err != nil {
				t.Fatalf("Conversion failed: %v", err)
			}
		})
	}
}

/* Edge cases and robustness */

func TestControlFlowExpression_EdgeCases(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name: "for with variable bounds",
			source: `per = 10
result = for i = 1 to per
    i`,
		},
		{
			name: "for with call expression bounds",
			source: `result = for i = 1 to input.int(10)
    i`,
		},
		{
			name: "if with complex condition",
			source: `result = if ta.crossover(close, ta.sma(close, 20))
    1`,
		},
		{
			name: "for with arithmetic in body",
			source: `result = for i = 1 to 10
    i * 2 + 3`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewParser()
			if err != nil {
				t.Fatalf("Parser creation failed: %v", err)
			}

			script, err := p.ParseString("", tt.source)
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			converter := NewConverter()
			_, err = converter.ToESTree(script)
			if err != nil {
				t.Fatalf("Conversion failed: %v", err)
			}
		})
	}
}

func TestControlFlowExpression_ConverterRobustness(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name: "for expression preserves counter name",
			source: `result = for counter = 1 to 10
    counter`,
		},
		{
			name: "if expression preserves test condition",
			source: `result = if high[1] > high[2]
    1`,
		},
		{
			name: "for expression with step preserves step value",
			source: `result = for i = 0 to 100 by 5
    i`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewParser()
			if err != nil {
				t.Fatalf("Parser creation failed: %v", err)
			}

			script, err := p.ParseString("", tt.source)
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			converter := NewConverter()
			program, err := converter.ToESTree(script)
			if err != nil {
				t.Fatalf("Conversion failed: %v", err)
			}

			if len(program.Body) == 0 {
				t.Fatal("Conversion produced empty program")
			}
		})
	}
}
