package codegen

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/parser"
)

func parseArrowFunction(t *testing.T, source string) *ast.ArrowFunctionExpression {
	p, err := parser.NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	script, err := p.ParseBytes("test.pine", []byte(source))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	converter := parser.NewConverter()
	program, err := converter.ToESTree(script)
	if err != nil {
		t.Fatalf("Conversion failed: %v", err)
	}

	var arrowFunc *ast.ArrowFunctionExpression
	for _, stmt := range program.Body {
		if varDecl, ok := stmt.(*ast.VariableDeclaration); ok {
			for _, decl := range varDecl.Declarations {
				if arrow, ok := decl.Init.(*ast.ArrowFunctionExpression); ok {
					arrowFunc = arrow
					break
				}
			}
		}
	}

	if arrowFunc == nil {
		t.Fatal("Arrow function not found in parsed AST")
	}

	return arrowFunc
}

func TestLocalSeriesAnalyzer_RmaOnLocalVariable(t *testing.T) {
	source := `
dirmov(len) =>
    up = change(high)
    plus = rma(up, len)
    [plus, up]
`
	arrowFunc := parseArrowFunction(t, source)
	analyzer := NewLocalSeriesAnalyzer()
	needsSeries := analyzer.Analyze(arrowFunc)

	if !needsSeries["up"] {
		t.Error("Expected 'up' to need Series storage (used in rma)")
	}

	if needsSeries["plus"] {
		t.Error("Expected 'plus' to NOT need Series storage (only assigned once)")
	}
}

func TestLocalSeriesAnalyzer_MultipleLocalVars(t *testing.T) {
	source := `
complex(len) =>
    a = close
    b = rma(a, len)
    c = sma(b, len)
    d = c * 2
    [b, c, d]
`
	arrowFunc := parseArrowFunction(t, source)
	analyzer := NewLocalSeriesAnalyzer()
	needsSeries := analyzer.Analyze(arrowFunc)

	if !needsSeries["a"] {
		t.Error("Expected 'a' to need Series storage (used in rma)")
	}

	if !needsSeries["b"] {
		t.Error("Expected 'b' to need Series storage (used in sma)")
	}

	if needsSeries["c"] {
		t.Error("Expected 'c' to NOT need Series storage (only used in arithmetic)")
	}

	if needsSeries["d"] {
		t.Error("Expected 'd' to NOT need Series storage (only assigned once)")
	}
}

func TestLocalSeriesAnalyzer_NoLocalVars(t *testing.T) {
	source := `//@version=5
study("Test")
simple(a, b) => 
    a + b
x = simple(10, 20)
`
	arrowFunc := parseArrowFunction(t, source)
	analyzer := NewLocalSeriesAnalyzer()
	needsSeries := analyzer.Analyze(arrowFunc)

	if len(needsSeries) != 0 {
		t.Errorf("Expected no Series needed, got %d variables", len(needsSeries))
	}
}

/* TestLocalSeriesAnalyzer_EdgeCases validates boundary conditions and complex patterns */
func TestLocalSeriesAnalyzer_EdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		source    string
		expected  map[string]bool
		expectLen int
	}{
		{
			name: "empty arrow body",
			source: `//@version=5
study("Test")
empty() => 
    0
x = empty()
`,
			expected:  map[string]bool{},
			expectLen: 0,
		},
		{
			name: "deeply nested TA calls",
			source: `
nested(len) =>
    a = close
    b = rma(a, len)
    c = ema(rma(b, len), len)
    [c]
`,
			expected: map[string]bool{
				"a": true,
				"b": true,
			},
			expectLen: 2,
		},
		{
			name: "mixed TA prefixes",
			source: `
mixed(len) =>
    x = close
    y = ta.sma(x, len)
    z = sma(y, len)
    [z]
`,
			expected: map[string]bool{
				"x": true,
				"y": true,
			},
			expectLen: 2,
		},
		{
			name: "historical access patterns",
			source: `
historical(len) =>
    a = close
    b = a[1] + a[2]
    c = b * 2
    [c]
`,
			expected: map[string]bool{
				"a": true,
			},
			expectLen: 1,
		},
		{
			name: "parameter shadowing",
			source: `
shadow(len, close) =>
    a = close
    b = rma(a, len)
    [b]
`,
			expected: map[string]bool{
				"a": true,
			},
			expectLen: 1,
		},
		{
			name: "ternary with TA",
			source: `
ternary(len, cond) =>
    a = close
    b = cond ? rma(a, len) : sma(a, len)
    [b]
`,
			expected: map[string]bool{
				"a": true,
			},
			expectLen: 1,
		},
		{
			name: "multiple TA on same var",
			source: `
multiple(len) =>
    x = close
    y = rma(x, len) + sma(x, len) + ema(x, len)
    [y]
`,
			expected: map[string]bool{
				"x": true,
			},
			expectLen: 1,
		},
		{
			name: "no local vars just params",
			source: `
params(a, b) =>
    rma(a, b) + sma(a, b)
`,
			expected:  map[string]bool{},
			expectLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			arrowFunc := parseArrowFunction(t, tt.source)
			analyzer := NewLocalSeriesAnalyzer()
			needsSeries := analyzer.Analyze(arrowFunc)

			if len(needsSeries) != tt.expectLen {
				t.Errorf("Expected %d variables, got %d: %v", tt.expectLen, len(needsSeries), needsSeries)
			}

			for varName, shouldNeed := range tt.expected {
				if needsSeries[varName] != shouldNeed {
					t.Errorf("Variable %q: expected needsSeries=%v, got %v", varName, shouldNeed, needsSeries[varName])
				}
			}
		})
	}
}
