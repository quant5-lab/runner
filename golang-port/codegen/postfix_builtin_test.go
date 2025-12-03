package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/parser"
)

// TestBuiltinIdentifiers_InTAFunctions verifies built-ins generate correct code in TA functions
func TestBuiltinIdentifiers_InTAFunctions(t *testing.T) {
	tests := []struct {
		name     string
		script   string
		expected string // Expected in generated code
	}{
		{
			name: "close in sma",
			script: `//@version=5
indicator("Test")
x = ta.sma(close, 20)
`,
			expected: "ctx.Data[ctx.BarIndex-j].Close",
		},
		{
			name: "open in ema",
			script: `//@version=5
indicator("Test")
x = ta.ema(open, 10)
`,
			expected: "ctx.Data[ctx.BarIndex-j].Open",
		},
		{
			name: "high in stdev",
			script: `//@version=5
indicator("Test")
x = ta.stdev(high, 20)
`,
			expected: "ctx.Data[ctx.BarIndex-j].High",
		},

		{
			name: "volume in sma",
			script: `//@version=5
indicator("Test")
x = ta.sma(volume, 20)
`,
			expected: "ctx.Data[ctx.BarIndex-j].Volume",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := parser.NewParser()
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			script, err := p.ParseBytes("test.pine", []byte(tt.script))
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			converter := parser.NewConverter()
			program, err := converter.ToESTree(script)
			if err != nil {
				t.Fatalf("Conversion failed: %v", err)
			}

			result, err := GenerateStrategyCodeFromAST(program)
			if err != nil {
				t.Fatalf("Codegen failed: %v", err)
			}

			if !strings.Contains(result.FunctionBody, tt.expected) {
				t.Errorf("Expected built-in access %q not found in generated code", tt.expected)
			}
		})
	}
}

// TestBuiltinIdentifiers_InConditions verifies built-ins generate correct code in conditions
func TestBuiltinIdentifiers_InConditions(t *testing.T) {
	tests := []struct {
		name     string
		script   string
		expected string
	}{
		{
			name: "close in ternary",
			script: `//@version=5
indicator("Test")
x = close > 100 ? 1 : 0
`,
			expected: "bar.Close > 100",
		},
		{
			name: "open in comparison",
			script: `//@version=5
indicator("Test")
x = open < close ? 1 : 0
`,
			expected: "bar.Open < bar.Close",
		},
		{
			name: "high and low in condition",
			script: `//@version=5
indicator("Test")
x = high - low > 10 ? 1 : 0
`,
			expected: "bar.High - bar.Low",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := parser.NewParser()
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			script, err := p.ParseBytes("test.pine", []byte(tt.script))
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			converter := parser.NewConverter()
			program, err := converter.ToESTree(script)
			if err != nil {
				t.Fatalf("Conversion failed: %v", err)
			}

			result, err := GenerateStrategyCodeFromAST(program)
			if err != nil {
				t.Fatalf("Codegen failed: %v", err)
			}

			if !strings.Contains(result.FunctionBody, tt.expected) {
				t.Errorf("Expected condition code %q not found in:\n%s", tt.expected, result.FunctionBody)
			}
		})
	}
}

// TestPostfixExpr_Codegen verifies codegen for function()[subscript] pattern
func TestPostfixExpr_Codegen(t *testing.T) {
	script := `//@version=5
indicator("Test")
pivot = pivothigh(5, 5)[1]
filled = fixnan(pivot)
`
	p, err := parser.NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	parseResult, err := p.ParseBytes("test.pine", []byte(script))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	converter := parser.NewConverter()
	program, err := converter.ToESTree(parseResult)
	if err != nil {
		t.Fatalf("Conversion failed: %v", err)
	}

	result, err := GenerateStrategyCodeFromAST(program)
	if err != nil {
		t.Fatalf("Codegen failed: %v", err)
	}

	// Verify pivothigh()[1] generates proper series access
	expectedPatterns := []string{
		"pivothighSeries.Get(1)", // Access to pivothigh result with offset 1
		"fixnanState_filled",     // fixnan state variable
	}

	for _, pattern := range expectedPatterns {
		if !strings.Contains(result.FunctionBody, pattern) {
			t.Errorf("Expected pattern %q not found in generated code", pattern)
		}
	}
}

// TestNestedPostfixExpr_Codegen verifies nested function()[subscript] in arguments
func TestNestedPostfixExpr_Codegen(t *testing.T) {
	script := `//@version=5
indicator("Test")
filled = fixnan(pivothigh(5, 5)[1])
`
	p, err := parser.NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	parseResult, err := p.ParseBytes("test.pine", []byte(script))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	converter := parser.NewConverter()
	program, err := converter.ToESTree(parseResult)
	if err != nil {
		t.Fatalf("Conversion failed: %v", err)
	}

	result, err := GenerateStrategyCodeFromAST(program)
	if err != nil {
		t.Fatalf("Codegen failed: %v", err)
	}

	// Verify nested pattern generates correct code
	expectedPatterns := []string{
		"pivothighSeries.Get(1)", // Subscripted function call
		"fixnanState_filled",     // fixnan state tracking
		"if !math.IsNaN",         // fixnan forward-fill check
	}

	for _, pattern := range expectedPatterns {
		if !strings.Contains(result.FunctionBody, pattern) {
			t.Errorf("Expected pattern %q not found in generated code", pattern)
		}
	}
}

// TestPostfixExpr_RegressionSafety ensures previous patterns still work
func TestPostfixExpr_RegressionSafety(t *testing.T) {
	tests := []struct {
		name     string
		script   string
		mustHave []string
	}{
		{
			name: "simple variable subscript",
			script: `//@version=5
indicator("Test")
x = close[1]
`,
			mustHave: []string{"ctx.Data[i-1].Close"},
		},
		{
			name: "ta function without subscript",
			script: `//@version=5
indicator("Test")
x = ta.sma(close, 20)
`,
			mustHave: []string{"ta.sma", "ctx.Data[ctx.BarIndex-j].Close"},
		},
		{
			name: "security with ta function",
			script: `//@version=5
indicator("Test")
x = request.security(syminfo.tickerid, "1D", ta.sma(close, 20))
`,
			mustHave: []string{"security", "ta.sma", "ctx.Data"},
		},
		{
			name: "plain identifier",
			script: `//@version=5
indicator("Test")
x = close
`,
			mustHave: []string{"closeSeries.GetCurrent()"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := parser.NewParser()
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			parseResult, err := p.ParseBytes("test.pine", []byte(tt.script))
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			converter := parser.NewConverter()
			program, err := converter.ToESTree(parseResult)
			if err != nil {
				t.Fatalf("Conversion failed: %v", err)
			}

			result, err := GenerateStrategyCodeFromAST(program)
			if err != nil {
				t.Fatalf("Codegen failed: %v", err)
			}

			for _, pattern := range tt.mustHave {
				if !strings.Contains(result.FunctionBody, pattern) {
					t.Errorf("Regression: Expected pattern %q not found", pattern)
				}
			}
		})
	}
}

// TestInputConstants_NotConfusedWithBuiltins verifies input constants aren't treated as built-ins
func TestInputConstants_NotConfusedWithBuiltins(t *testing.T) {
	// Create a program with input constant named 'close' (edge case)
	program := &ast.Program{
		Body: []ast.Node{
			&ast.ExpressionStatement{
				Expression: &ast.CallExpression{
					Callee: &ast.Identifier{Name: "indicator"},
					Arguments: []ast.Expression{
						&ast.Literal{Value: "Test"},
					},
				},
			},
			&ast.VariableDeclaration{
				Declarations: []ast.VariableDeclarator{
					{
						ID: ast.Identifier{Name: "myInput"},
						Init: &ast.CallExpression{
							Callee: &ast.MemberExpression{
								Object:   &ast.Identifier{Name: "input"},
								Property: &ast.Identifier{Name: "float"},
							},
							Arguments: []ast.Expression{
								&ast.Literal{Value: float64(10)},
							},
						},
					},
				},
			},
			&ast.VariableDeclaration{
				Declarations: []ast.VariableDeclarator{
					{
						ID: ast.Identifier{Name: "x"},
						Init: &ast.BinaryExpression{
							Operator: "+",
							Left:     &ast.Identifier{Name: "myInput"},
							Right:    &ast.Identifier{Name: "close"}, // Built-in
						},
					},
				},
			},
		},
	}

	result, err := GenerateStrategyCodeFromAST(program)
	if err != nil {
		t.Fatalf("Codegen failed: %v", err)
	}

	// myInput should be treated as constant (not bar.myInput)
	if !strings.Contains(result.FunctionBody, "myInput + bar.Close") {
		t.Error("Input constant not properly distinguished from built-in")
	}
}
