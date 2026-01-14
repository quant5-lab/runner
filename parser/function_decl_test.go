package parser

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
)

/* Tests for arrow function declaration parsing with INDENT/DEDENT */

func getBodyLength(funcDecl *FunctionDecl) int {
	if funcDecl.InlineBody != nil {
		return 1
	}
	if funcDecl.MultiLineBody != nil {
		return len(funcDecl.MultiLineBody)
	}
	return 0
}

// TestFunctionDecl_StatementCounts verifies functions with varying body sizes
func TestFunctionDecl_StatementCounts(t *testing.T) {
	tests := []struct {
		name          string
		source        string
		funcName      string
		expectedStmts int
	}{
		{
			name: "single statement",
			source: `simple(x) =>
    x + 1`,
			funcName:      "simple",
			expectedStmts: 1,
		},
		{
			name: "two statements",
			source: `calc(x) =>
    a = x * 2
    a + 1`,
			funcName:      "calc",
			expectedStmts: 2,
		},
		{
			name: "three statements",
			source: `multi(x) =>
    a = x + 1
    b = a * 2
    b - 3`,
			funcName:      "multi",
			expectedStmts: 3,
		},
		{
			name: "complex body - BB7 dirmov pattern",
			source: `dirmov(len) =>
    up = change(high)
    down = -change(low)
    truerange = rma(tr, len)
    plus = fixnan(100 * rma(up > down and up > 0 ? up : 0, len) / truerange)
    minus = fixnan(100 * rma(down > up and down > 0 ? down : 0, len) / truerange)
    [plus, minus]`,
			funcName:      "dirmov",
			expectedStmts: 6,
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

			funcDecl := script.Statements[0].FunctionDecl
			if funcDecl == nil {
				t.Fatal("Expected FunctionDecl, got nil")
			}

			if funcDecl.Name != tt.funcName {
				t.Errorf("Expected function name '%s', got '%s'", tt.funcName, funcDecl.Name)
			}

			if getBodyLength(funcDecl) != tt.expectedStmts {
				t.Errorf("Expected %d body statements, got %d", tt.expectedStmts, getBodyLength(funcDecl))
			}
		})
	}
}

// TestFunctionDecl_ParameterCounts verifies functions with varying parameter counts
func TestFunctionDecl_ParameterCounts(t *testing.T) {
	tests := []struct {
		name           string
		source         string
		expectedParams []string
	}{
		{
			name: "no parameters",
			source: `noparams() =>
    42`,
			expectedParams: []string{},
		},
		{
			name: "single parameter",
			source: `single(x) =>
    x + 1`,
			expectedParams: []string{"x"},
		},
		{
			name: "two parameters",
			source: `double(a, b) =>
    a + b`,
			expectedParams: []string{"a", "b"},
		},
		{
			name: "three parameters - BB7 pattern",
			source: `adx(LWdilength, LWadxlength, extra) =>
    [plus, minus] = dirmov(LWdilength)
    100 * rma(abs(plus - minus), LWadxlength)`,
			expectedParams: []string{"LWdilength", "LWadxlength", "extra"},
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
				t.Fatal("Expected FunctionDecl")
			}

			if len(funcDecl.Params) != len(tt.expectedParams) {
				t.Fatalf("Expected %d params, got %d", len(tt.expectedParams), len(funcDecl.Params))
			}

			for i, expected := range tt.expectedParams {
				if funcDecl.Params[i] != expected {
					t.Errorf("Expected param[%d] '%s', got '%s'", i, expected, funcDecl.Params[i])
				}
			}
		})
	}
}

// TestFunctionDecl_MultipleFunctions verifies multiple functions in sequence
func TestFunctionDecl_MultipleFunctions(t *testing.T) {
	tests := []struct {
		name          string
		source        string
		expectedFuncs []string
	}{
		{
			name: "two functions - no blank line",
			source: `first(x) =>
    x + 1
second(y) =>
    y * 2`,
			expectedFuncs: []string{"first", "second"},
		},
		{
			name: "two functions - with blank line",
			source: `first(x) =>
    x + 1

second(y) =>
    y * 2`,
			expectedFuncs: []string{"first", "second"},
		},
		{
			name: "three functions - BB7 pattern",
			source: `dirmov(len) =>
    up = change(high)
    [up, 0]

adx(a, b) =>
    [x, y] = dirmov(a)
    x + y

helper(z) =>
    z * 2`,
			expectedFuncs: []string{"dirmov", "adx", "helper"},
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

			if len(script.Statements) != len(tt.expectedFuncs) {
				t.Fatalf("Expected %d functions, got %d statements", len(tt.expectedFuncs), len(script.Statements))
			}

			for i, expectedName := range tt.expectedFuncs {
				funcDecl := script.Statements[i].FunctionDecl
				if funcDecl == nil {
					t.Fatalf("Statement %d: expected FunctionDecl, got nil", i)
				}
				if funcDecl.Name != expectedName {
					t.Errorf("Function %d: expected name '%s', got '%s'", i, expectedName, funcDecl.Name)
				}
			}
		})
	}
}

// TestFunctionDecl_WithEmptyLines verifies empty lines within function bodies
func TestFunctionDecl_WithEmptyLines(t *testing.T) {
	tests := []struct {
		name          string
		source        string
		expectedStmts int
	}{
		{
			name: "single empty line in middle",
			source: `func(x) =>
    a = x + 1

    b = a * 2`,
			expectedStmts: 2,
		},
		{
			name: "multiple empty lines",
			source: `func(x) =>
    a = x + 1


    b = a * 2`,
			expectedStmts: 2,
		},
		{
			name: "empty line before return",
			source: `func(x) =>
    a = x + 1

    a`,
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

			funcDecl := script.Statements[0].FunctionDecl
			if funcDecl == nil {
				t.Fatal("Expected FunctionDecl")
			}

			if getBodyLength(funcDecl) != tt.expectedStmts {
				t.Errorf("Expected %d body statements, got %d", tt.expectedStmts, getBodyLength(funcDecl))
			}
		})
	}
}

// TestFunctionDecl_WithComments verifies comment handling in function bodies
func TestFunctionDecl_WithComments(t *testing.T) {
	tests := []struct {
		name          string
		source        string
		expectedStmts int
	}{
		{
			name: "comment before body",
			source: `func(x) =>
    // Calculate result
    x + 1`,
			expectedStmts: 1,
		},
		{
			name: "comments between statements",
			source: `func(x) =>
    a = x + 1
    // Multiply by 2
    b = a * 2
    // Return result
    b`,
			expectedStmts: 3,
		},
		{
			name: "inline comment",
			source: `func(x) =>
    a = x + 1  // increment
    a * 2`,
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

			funcDecl := script.Statements[0].FunctionDecl
			if funcDecl == nil {
				t.Fatal("Expected FunctionDecl")
			}

			if getBodyLength(funcDecl) != tt.expectedStmts {
				t.Errorf("Expected %d body statements, got %d", tt.expectedStmts, getBodyLength(funcDecl))
			}
		})
	}
}

// TestFunctionDecl_ReturnValues verifies various return value patterns
func TestFunctionDecl_ReturnValues(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name: "simple expression",
			source: `func(x) =>
    x + 1`,
		},
		{
			name: "identifier",
			source: `func(x) =>
    result = x + 1
    result`,
		},
		{
			name: "tuple literal",
			source: `func(x) =>
    a = x + 1
    [a, x]`,
		},
		{
			name: "function call",
			source: `func(x) =>
    sma(x, 20)`,
		},
		{
			name: "ternary expression",
			source: `func(x) =>
    x > 0 ? x : 0`,
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
				t.Fatal("Expected FunctionDecl")
			}

			if getBodyLength(funcDecl) == 0 {
				t.Fatal("Function body is empty")
			}

			if funcDecl.MultiLineBody == nil {
				t.Skip("Skipping multi-line body test for inline function")
			}

			lastStmt := funcDecl.MultiLineBody[len(funcDecl.MultiLineBody)-1]
			if lastStmt.Expression == nil && lastStmt.TupleAssignment == nil && lastStmt.Assignment == nil {
				t.Error("Last statement should be an expression, tuple, or assignment")
			}
		})
	}
}

// TestFunctionDecl_MixedWithStatements verifies functions mixed with other statements
func TestFunctionDecl_MixedWithStatements(t *testing.T) {
	tests := []struct {
		name            string
		source          string
		expectedPattern []string // "func", "assign", "expr", etc.
	}{
		{
			name: "function then assignment",
			source: `helper(x) =>
    x + 1
result = helper(10)`,
			expectedPattern: []string{"func", "assign"},
		},
		{
			name: "assignment then function",
			source: `value = 10
helper(x) =>
    x + value`,
			expectedPattern: []string{"assign", "func"},
		},
		{
			name: "function, assignment, function",
			source: `first(x) =>
    x + 1
value = 10
second(y) =>
    y * value`,
			expectedPattern: []string{"func", "assign", "func"},
		},
		{
			name: "real world - BB7 pattern",
			source: `LWadxlength = input(16)
dirmov(len) =>
    up = change(high)
    [up, 0]
[ADX, up, down] = dirmov(LWadxlength)`,
			expectedPattern: []string{"assign", "func", "tuple"},
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
				switch {
				case stmt.FunctionDecl != nil:
					actual = "func"
				case stmt.Assignment != nil:
					actual = "assign"
				case stmt.TupleAssignment != nil:
					actual = "tuple"
				case stmt.Expression != nil:
					actual = "expr"
				case stmt.If != nil:
					actual = "if"
				case stmt.Reassignment != nil:
					actual = "reassign"
				default:
					actual = "unknown"
				}

				if actual != expected {
					t.Errorf("Statement %d: expected %s, got %s", i, expected, actual)
				}
			}
		})
	}
}

// TestFunctionDecl_NestedIfStatements verifies IF statements inside function bodies
func TestFunctionDecl_NestedIfStatements(t *testing.T) {
	tests := []struct {
		name          string
		source        string
		expectedStmts int
	}{
		{
			name: "single if in function",
			source: `func(x) =>
    result = 0
    if x > 0
        result := x
    result`,
			expectedStmts: 3,
		},
		{
			name: "multiple ifs in function",
			source: `func(x) =>
    result = 0
    if x > 10
        result := 10
    if x < 0
        result := 0
    result`,
			expectedStmts: 4,
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
				t.Fatal("Expected FunctionDecl")
			}

			if getBodyLength(funcDecl) != tt.expectedStmts {
				t.Errorf("Expected %d body statements, got %d", tt.expectedStmts, getBodyLength(funcDecl))
			}

			if funcDecl.MultiLineBody == nil {
				t.Skip("Skipping multi-line body test for inline function")
			}

			// Verify at least one IF statement exists
			hasIf := false
			for _, stmt := range funcDecl.MultiLineBody {
				if stmt.If != nil {
					hasIf = true
					break
				}
			}
			if !hasIf {
				t.Error("Expected at least one IF statement in function body")
			}
		})
	}
}

// TestFunctionDecl_Converter verifies AST conversion to ESTree
func TestFunctionDecl_Converter(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		funcName string
	}{
		{
			name: "simple function",
			source: `simple(x) =>
    x + 1`,
			funcName: "simple",
		},
		{
			name: "BB7 dirmov pattern",
			source: `dirmov(len) =>
    up = change(high)
    down = -change(low)
    [up, down]`,
			funcName: "dirmov",
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

			// Should be VariableDeclaration with ArrowFunctionExpression
			varDecl, ok := program.Body[0].(*ast.VariableDeclaration)
			if !ok {
				t.Fatalf("Expected VariableDeclaration, got %T", program.Body[0])
			}

			if len(varDecl.Declarations) == 0 {
				t.Fatal("No declarations in VariableDeclaration")
			}

			idNode, ok := varDecl.Declarations[0].ID.(*ast.Identifier)
			if !ok {
				t.Fatalf("Expected Identifier, got %T", varDecl.Declarations[0].ID)
			}

			if idNode.Name != tt.funcName {
				t.Errorf("Expected function name '%s', got '%s'", tt.funcName, idNode.Name)
			}

			arrowFunc, ok := varDecl.Declarations[0].Init.(*ast.ArrowFunctionExpression)
			if !ok {
				t.Fatalf("Expected ArrowFunctionExpression, got %T", varDecl.Declarations[0].Init)
			}

			if len(arrowFunc.Body) == 0 {
				t.Error("ArrowFunctionExpression body is empty")
			}
		})
	}
}

// TestFunctionDecl_EdgeCases verifies error handling and edge cases
func TestFunctionDecl_EdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		source    string
		shouldErr bool
	}{
		{
			name: "function without indent",
			source: `func(x) =>
x + 1`,
			shouldErr: false, // Inline expression now supported
		},
		{
			name: "empty function body",
			source: `func(x) =>
`,
			shouldErr: true, // No expression after =>
		},
		{
			name: "inconsistent indentation - lexer lenient",
			source: `func(x) =>
    a = 1
  b = 2`, // Different indent levels
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

// TestFunctionDecl_RealWorldPatterns verifies actual PineScript patterns from BB7
func TestFunctionDecl_RealWorldPatterns(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name: "BB7 dirmov - complete",
			source: `dirmov(len) =>
    up = change(high)
    down = -change(low)
    truerange = rma(tr, len)
    plus = fixnan(100 * rma(up > down and up > 0 ? up : 0, len) / truerange)
    minus = fixnan(100 * rma(down > up and down > 0 ? down : 0, len) / truerange)
    [plus, minus]`,
		},
		{
			name: "BB7 adx - complete",
			source: `adx(LWdilength, LWadxlength) =>
    [plus, minus] = dirmov(LWdilength)
    sum = plus + minus
    adx = 100 * rma(abs(plus - minus) / (sum == 0 ? 1 : sum), LWadxlength)
    [adx, plus, minus]`,
		},
		{
			name: "function calling function",
			source: `helper(x) =>
    x * 2
main(y) =>
    result = helper(y)
    result + 1`,
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

			// Verify it's a valid function declaration
			varDecl, ok := program.Body[0].(*ast.VariableDeclaration)
			if !ok {
				t.Fatalf("Expected VariableDeclaration, got %T", program.Body[0])
			}

			arrowFunc, ok := varDecl.Declarations[0].Init.(*ast.ArrowFunctionExpression)
			if !ok {
				t.Fatalf("Expected ArrowFunctionExpression, got %T", varDecl.Declarations[0].Init)
			}

			if len(arrowFunc.Body) == 0 {
				t.Error("Function body is empty")
			}
		})
	}
}
