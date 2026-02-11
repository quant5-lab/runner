package parser

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
)

/* TestVarAssignment_BasicParsing validates var/varip keyword recognition across all syntax variants */
func TestVarAssignment_BasicParsing(t *testing.T) {
	tests := []struct {
		name        string
		source      string
		varName     string
		persistence string
	}{
		{
			name:        "var simple integer",
			source:      `var x = 0`,
			varName:     "x",
			persistence: "var",
		},
		{
			name:        "var simple float literal",
			source:      `var x = 1.5`,
			varName:     "x",
			persistence: "var",
		},
		{
			name:        "varip simple integer",
			source:      `varip x = 0`,
			varName:     "x",
			persistence: "varip",
		},
		{
			name:        "var with identifier init",
			source:      `var x = close`,
			varName:     "x",
			persistence: "var",
		},
		{
			name:        "varip with identifier init",
			source:      `varip cumulative = close`,
			varName:     "cumulative",
			persistence: "varip",
		},
		{
			name:        "var with expression init",
			source:      `var total = close + open`,
			varName:     "total",
			persistence: "var",
		},
		{
			name:        "var with function call init",
			source:      `var avg = ta.sma(close, 14)`,
			varName:     "avg",
			persistence: "var",
		},
		{
			name:        "var with na init",
			source:      `var x = na`,
			varName:     "x",
			persistence: "var",
		},
		{
			name:        "var with true init",
			source:      `var flag = true`,
			varName:     "flag",
			persistence: "var",
		},
		{
			name:        "var with false init",
			source:      `var done = false`,
			varName:     "done",
			persistence: "var",
		},
		{
			name:        "var with negated expression",
			source:      `var x = -1`,
			varName:     "x",
			persistence: "var",
		},
		{
			name:        "var with ternary expression",
			source:      `var x = close > open ? 1 : 0`,
			varName:     "x",
			persistence: "var",
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

			decl := findVariableDeclaration(program, tt.varName)
			if decl == nil {
				t.Fatalf("Expected VariableDeclaration for %q, not found in AST", tt.varName)
			}

			if decl.Persistence != tt.persistence {
				t.Errorf("Persistence: expected %q, got %q", tt.persistence, decl.Persistence)
			}

			if decl.Kind != "let" {
				t.Errorf("Kind: expected \"let\" (declaration), got %q", decl.Kind)
			}

			if decl.Declarations[0].Init == nil {
				t.Error("Expected Init expression, got nil")
			}
		})
	}
}

/* TestVarAssignment_TypedDeclarations validates var/varip with explicit type hints */
func TestVarAssignment_TypedDeclarations(t *testing.T) {
	tests := []struct {
		name        string
		source      string
		varName     string
		persistence string
	}{
		{
			name:        "var float typed",
			source:      `var float x = 0.0`,
			varName:     "x",
			persistence: "var",
		},
		{
			name:        "var int typed",
			source:      `var int counter = 0`,
			varName:     "counter",
			persistence: "var",
		},
		{
			name:        "var bool typed",
			source:      `var bool flag = false`,
			varName:     "flag",
			persistence: "var",
		},
		{
			name:        "var string typed",
			source:      `var string label = "hello"`,
			varName:     "label",
			persistence: "var",
		},
		{
			name:        "var color typed",
			source:      `var color lineColor = na`,
			varName:     "lineColor",
			persistence: "var",
		},
		{
			name:        "varip float typed",
			source:      `varip float cumulative = 0.0`,
			varName:     "cumulative",
			persistence: "varip",
		},
		{
			name:        "varip int typed",
			source:      `varip int tickCount = 0`,
			varName:     "tickCount",
			persistence: "varip",
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

			decl := findVariableDeclaration(program, tt.varName)
			if decl == nil {
				t.Fatalf("Expected VariableDeclaration for %q, not found", tt.varName)
			}

			if decl.Persistence != tt.persistence {
				t.Errorf("Persistence: expected %q, got %q", tt.persistence, decl.Persistence)
			}

			if decl.Kind != "let" {
				t.Errorf("Kind: expected \"let\", got %q", decl.Kind)
			}
		})
	}
}

/* TestVarAssignment_TupleDestructuring validates var/varip with tuple pattern */
func TestVarAssignment_TupleDestructuring(t *testing.T) {
	tests := []struct {
		name        string
		source      string
		varNames    []string
		persistence string
	}{
		{
			name:        "var tuple two elements",
			source:      `var [upper, lower] = ta.bb(close, 20, 2)`,
			varNames:    []string{"upper", "lower"},
			persistence: "var",
		},
		{
			name:        "varip tuple two elements",
			source:      `varip [a, b] = someFunc()`,
			varNames:    []string{"a", "b"},
			persistence: "varip",
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
				t.Fatal("Expected at least one statement in AST")
			}

			varDecl, ok := program.Body[0].(*ast.VariableDeclaration)
			if !ok {
				t.Fatalf("Expected VariableDeclaration, got %T", program.Body[0])
			}

			if varDecl.Persistence != tt.persistence {
				t.Errorf("Persistence: expected %q, got %q", tt.persistence, varDecl.Persistence)
			}

			if len(varDecl.Declarations) == 0 {
				t.Fatal("Expected at least one declarator")
			}

			arrayPattern, ok := varDecl.Declarations[0].ID.(*ast.ArrayPattern)
			if !ok {
				t.Fatalf("Expected ArrayPattern, got %T", varDecl.Declarations[0].ID)
			}

			if len(arrayPattern.Elements) != len(tt.varNames) {
				t.Fatalf("Expected %d tuple elements, got %d", len(tt.varNames), len(arrayPattern.Elements))
			}

			for i, expected := range tt.varNames {
				if arrayPattern.Elements[i].Name != expected {
					t.Errorf("Element[%d]: expected %q, got %q", i, expected, arrayPattern.Elements[i].Name)
				}
			}
		})
	}
}

/* TestVarAssignment_WithReassignment validates var + reassignment pattern */
func TestVarAssignment_WithReassignment(t *testing.T) {
	source := `var x = 0
x := x + 1`

	p, err := NewParser()
	if err != nil {
		t.Fatalf("Parser creation failed: %v", err)
	}

	script, err := p.ParseString("", source)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	converter := NewConverter()
	program, err := converter.ToESTree(script)
	if err != nil {
		t.Fatalf("Conversion failed: %v", err)
	}

	if len(program.Body) != 2 {
		t.Fatalf("Expected 2 statements (var decl + reassignment), got %d", len(program.Body))
	}

	/* First statement: var x = 0 → VariableDeclaration{Kind:"let", Persistence:"var"} */
	varDecl, ok := program.Body[0].(*ast.VariableDeclaration)
	if !ok {
		t.Fatalf("Statement 0: expected VariableDeclaration, got %T", program.Body[0])
	}
	if varDecl.Kind != "let" {
		t.Errorf("Statement 0 Kind: expected \"let\", got %q", varDecl.Kind)
	}
	if varDecl.Persistence != "var" {
		t.Errorf("Statement 0 Persistence: expected \"var\", got %q", varDecl.Persistence)
	}

	/* Second statement: x := x + 1 → VariableDeclaration{Kind:"var", Persistence:""} */
	reassign, ok := program.Body[1].(*ast.VariableDeclaration)
	if !ok {
		t.Fatalf("Statement 1: expected VariableDeclaration, got %T", program.Body[1])
	}
	if reassign.Kind != "var" {
		t.Errorf("Statement 1 Kind: expected \"var\" (reassignment), got %q", reassign.Kind)
	}
	if reassign.Persistence != "" {
		t.Errorf("Statement 1 Persistence: expected \"\" (not persisted), got %q", reassign.Persistence)
	}
}

/* TestVarAssignment_StatementCount validates var is parsed as single statement (not split) */
func TestVarAssignment_StatementCount(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		expected int
	}{
		{
			name:     "var alone produces 1 statement",
			source:   `var x = 0`,
			expected: 1,
		},
		{
			name:     "var + reassignment produces 2 statements",
			source:   "var x = 0\nx := 1",
			expected: 2,
		},
		{
			name:     "plain assignment still produces 1 statement",
			source:   `x = 0`,
			expected: 1,
		},
		{
			name:     "reassignment still produces 1 statement",
			source:   `x := 1`,
			expected: 1,
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

			if len(program.Body) != tt.expected {
				t.Errorf("Expected %d statement(s), got %d", tt.expected, len(program.Body))
			}
		})
	}
}

/* TestVarAssignment_NoRegressionPlainAssignment ensures existing plain/typed/tuple assignments still work */
func TestVarAssignment_NoRegressionPlainAssignment(t *testing.T) {
	tests := []struct {
		name        string
		source      string
		varName     string
		kind        string
		persistence string
	}{
		{
			name:        "plain assignment",
			source:      `x = close`,
			varName:     "x",
			kind:        "let",
			persistence: "",
		},
		{
			name:        "typed assignment",
			source:      `float x = close`,
			varName:     "x",
			kind:        "let",
			persistence: "",
		},
		{
			name:        "reassignment",
			source:      `x := 1`,
			varName:     "x",
			kind:        "var",
			persistence: "",
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

			decl := findVariableDeclaration(program, tt.varName)
			if decl == nil {
				t.Fatalf("Expected VariableDeclaration for %q", tt.varName)
			}

			if decl.Kind != tt.kind {
				t.Errorf("Kind: expected %q, got %q", tt.kind, decl.Kind)
			}

			if decl.Persistence != tt.persistence {
				t.Errorf("Persistence: expected %q, got %q", tt.persistence, decl.Persistence)
			}
		})
	}
}

/* TestVarAssignment_WithControlFlowInit validates var with for/if/switch as init expression */
func TestVarAssignment_WithControlFlowInit(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		varName  string
		initType string
	}{
		{
			name: "var with for expression",
			source: `var sum = for i = 1 to 10
    i`,
			varName:  "sum",
			initType: "ForStatement",
		},
		{
			name: "var with if expression",
			source: `var signal = if close > open
    1`,
			varName:  "signal",
			initType: "IfStatement",
		},
		{
			name: "var with switch expression",
			source: `var result = switch mode
    1 =>
        10
    2 =>
        20`,
			varName:  "result",
			initType: "IfStatement", // switch lowers to if
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

			decl := findVariableDeclaration(program, tt.varName)
			if decl == nil {
				t.Fatalf("Expected VariableDeclaration for %q", tt.varName)
			}

			if decl.Persistence != "var" {
				t.Errorf("Persistence: expected \"var\", got %q", decl.Persistence)
			}

			if decl.Declarations[0].Init == nil {
				t.Error("Expected Init expression, got nil")
			}
		})
	}
}

/* TestVarAssignment_InFunctionBody validates var/varip inside user-defined function bodies */
func TestVarAssignment_InFunctionBody(t *testing.T) {
	source := `myFunc() =>
    var x = 0
    x := x + 1
    x`

	p, err := NewParser()
	if err != nil {
		t.Fatalf("Parser creation failed: %v", err)
	}

	script, err := p.ParseString("", source)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	converter := NewConverter()
	program, err := converter.ToESTree(script)
	if err != nil {
		t.Fatalf("Conversion failed: %v", err)
	}

	/* The function declaration wraps var x = 0 inside its body */
	if len(program.Body) == 0 {
		t.Fatal("Expected at least one statement")
	}

	funcDecl, ok := program.Body[0].(*ast.VariableDeclaration)
	if !ok {
		t.Fatalf("Expected VariableDeclaration (function), got %T", program.Body[0])
	}

	arrowFunc, ok := funcDecl.Declarations[0].Init.(*ast.ArrowFunctionExpression)
	if !ok {
		t.Fatalf("Expected ArrowFunctionExpression, got %T", funcDecl.Declarations[0].Init)
	}

	/* Find var x = 0 inside the arrow function body */
	var foundVarDecl bool
	for _, bodyNode := range arrowFunc.Body {
		if varDecl, ok := bodyNode.(*ast.VariableDeclaration); ok {
			if varDecl.Persistence == "var" {
				foundVarDecl = true
				if len(varDecl.Declarations) > 0 {
					if id, ok := varDecl.Declarations[0].ID.(*ast.Identifier); ok {
						if id.Name != "x" {
							t.Errorf("Expected var declaration for 'x', got %q", id.Name)
						}
					}
				}
			}
		}
	}

	if !foundVarDecl {
		t.Error("Expected to find var-persisted declaration inside function body")
	}
}
