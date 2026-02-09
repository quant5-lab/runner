package codegen

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/parser"
)

func TestControlFlowExpression_ForLoopAsExpression(t *testing.T) {
	tests := []struct {
		name    string
		source  string
		counter string
		hasStep bool
	}{
		{
			name: "simple for loop expression",
			source: `result = for i = 1 to 10
    i`,
			counter: "i",
			hasStep: false,
		},
		{
			name: "for loop expression with step",
			source: `result = for i = 0 to 100 by 5
    i`,
			counter: "i",
			hasStep: true,
		},
		{
			name: "for loop with custom counter name",
			source: `sum = for counter = 1 to 20
    counter`,
			counter: "counter",
			hasStep: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := parser.NewParser()
			if err != nil {
				t.Fatalf("Parser creation failed: %v", err)
			}

			script, err := p.ParseString("", tt.source)
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			converter := parser.NewConverter()
			program, err := converter.ToESTree(script)
			if err != nil {
				t.Fatalf("Conversion failed: %v", err)
			}

			varDecl := program.Body[0].(*ast.VariableDeclaration)
			forStmt, ok := varDecl.Declarations[0].Init.(*ast.ForStatement)
			if !ok {
				t.Fatalf("Expected ForStatement as init, got %T", varDecl.Declarations[0].Init)
			}

			if forStmt.Counter != tt.counter {
				t.Errorf("Counter: expected=%q got=%q", tt.counter, forStmt.Counter)
			}

			if tt.hasStep && forStmt.Step == nil {
				t.Error("Expected step expression, got nil")
			}
			if !tt.hasStep && forStmt.Step != nil {
				t.Error("Expected no step expression, found one")
			}
		})
	}
}

func TestControlFlowExpression_IfStatementAsExpression(t *testing.T) {
	tests := []struct {
		name        string
		source      string
		minBodySize int
	}{
		{
			name: "simple if expression",
			source: `result = if close > open
    close`,
			minBodySize: 1,
		},
		{
			name: "if expression with multiple statements",
			source: `result = if true
    a = 1
    b = 2
    b`,
			minBodySize: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := parser.NewParser()
			if err != nil {
				t.Fatalf("Parser creation failed: %v", err)
			}

			script, err := p.ParseString("", tt.source)
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			converter := parser.NewConverter()
			program, err := converter.ToESTree(script)
			if err != nil {
				t.Fatalf("Conversion failed: %v", err)
			}

			varDecl := program.Body[0].(*ast.VariableDeclaration)
			ifStmt, ok := varDecl.Declarations[0].Init.(*ast.IfStatement)
			if !ok {
				t.Fatalf("Expected IfStatement as init, got %T", varDecl.Declarations[0].Init)
			}

			if len(ifStmt.Consequent) < tt.minBodySize {
				t.Errorf("Consequent size: expected>=%d got=%d", tt.minBodySize, len(ifStmt.Consequent))
			}
		})
	}
}

func TestControlFlowExpression_Nesting(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name: "for inside if",
			source: `result = if close > open
    for i = 1 to 10
        i`,
		},
		{
			name: "if inside for",
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
			p, err := parser.NewParser()
			if err != nil {
				t.Fatalf("Parser creation failed: %v", err)
			}

			script, err := p.ParseString("", tt.source)
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			converter := parser.NewConverter()
			_, err = converter.ToESTree(script)
			if err != nil {
				t.Fatalf("Conversion failed: %v", err)
			}
		})
	}
}

func TestControlFlowExpression_BodyComplexity(t *testing.T) {
	tests := []struct {
		name        string
		source      string
		minBodySize int
	}{
		{
			name: "for with variable mutations",
			source: `result = for i = 1 to 10
    sum = 0
    sum := sum + i
    sum`,
			minBodySize: 3,
		},
		{
			name: "if with arithmetic operations",
			source: `result = if close > open
    diff = close - open
    ratio = diff / open
    ratio`,
			minBodySize: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := parser.NewParser()
			if err != nil {
				t.Fatalf("Parser creation failed: %v", err)
			}

			script, err := p.ParseString("", tt.source)
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			converter := parser.NewConverter()
			program, err := converter.ToESTree(script)
			if err != nil {
				t.Fatalf("Conversion failed: %v", err)
			}

			varDecl := program.Body[0].(*ast.VariableDeclaration)
			var bodyLen int

			switch init := varDecl.Declarations[0].Init.(type) {
			case *ast.ForStatement:
				bodyLen = len(init.Body)
			case *ast.IfStatement:
				bodyLen = len(init.Consequent)
			default:
				t.Fatalf("Unexpected init type: %T", init)
			}

			if bodyLen < tt.minBodySize {
				t.Errorf("Body size: expected>=%d got=%d", tt.minBodySize, bodyLen)
			}
		})
	}
}

/* If-expression na semantics: missing else returns math.NaN() per PineScript spec */
func TestControlFlowExpression_IfAlternateBehavior(t *testing.T) {
	tests := []struct {
		name         string
		source       string
		mustHaveNaN  bool
		mustHaveElse bool
	}{
		{
			name: "no else returns na",
			source: `x = if condition
    10`,
			mustHaveNaN:  true,
			mustHaveElse: false,
		},
		{
			name: "with else returns value",
			source: `x = if condition
    10
else
    20`,
			mustHaveNaN:  false,
			mustHaveElse: true,
		},
		{
			name: "else-if chain with final else",
			source: `x = if a
    1
else if b
    2
else
    3`,
			mustHaveNaN:  false,
			mustHaveElse: true,
		},
		{
			name: "else-if chain without final else returns na",
			source: `x = if a
    1
else if b
    2`,
			mustHaveNaN:  true,
			mustHaveElse: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := parser.NewParser()
			if err != nil {
				t.Fatalf("Parser creation failed: %v", err)
			}

			script, err := p.ParseString("", tt.source)
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			converter := parser.NewConverter()
			program, err := converter.ToESTree(script)
			if err != nil {
				t.Fatalf("Conversion failed: %v", err)
			}

			gen := newTestGenerator()
			code, err := gen.generateProgram(program)
			if err != nil {
				t.Fatalf("Generate failed: %v", err)
			}

			verifier := NewCodeVerifier(code, t)
			verifier.MustContain("func() float64")

			if tt.mustHaveNaN {
				verifier.MustContain("return math.NaN()")
			}

			if tt.mustHaveElse {
				verifier.MustContain("else")
			}
		})
	}
}
