package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestInputSourceUsesSeriesStorage(t *testing.T) {
	tests := []struct {
		name          string
		sourceBuiltin string
	}{
		{name: "hl2", sourceBuiltin: "hl2"},
		{name: "hlc3", sourceBuiltin: "hlc3"},
		{name: "ohlc4", sourceBuiltin: "ohlc4"},
		{name: "close", sourceBuiltin: "close"},
		{name: "high", sourceBuiltin: "high"},
		{name: "low", sourceBuiltin: "low"},
		{name: "open", sourceBuiltin: "open"},
		{name: "volume", sourceBuiltin: "volume"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			program := &ast.Program{
				NodeType: ast.TypeProgram,
				Body: []ast.Node{
					&ast.VariableDeclaration{
						Kind: "let",
						Declarations: []ast.VariableDeclarator{
							{
								ID: &ast.Identifier{Name: "src"},
								Init: &ast.CallExpression{
									Callee: &ast.Identifier{Name: "input"},
									Arguments: []ast.Expression{
										&ast.Identifier{Name: tt.sourceBuiltin},
									},
								},
							},
						},
					},
				},
			}

			goCode, err := GenerateStrategyCodeFromAST(program)
			if err != nil {
				t.Fatalf("codegen error: %v", err)
			}

			fullCode := goCode.FunctionBody

			if !strings.Contains(fullCode, "srcSeries") {
				t.Error("input.source must use Series storage, not scalar const")
			}

			if !strings.Contains(fullCode, "srcSeries.Set(") {
				t.Error("input.source Series must be initialized with Set()")
			}

			if strings.Contains(fullCode, "const src =") {
				t.Error("input.source must NOT be scalar constant")
			}
		})
	}
}

func TestInputSourceInBinaryExpressions(t *testing.T) {
	program := &ast.Program{
		NodeType: ast.TypeProgram,
		Body: []ast.Node{
			&ast.VariableDeclaration{
				Kind: "let",
				Declarations: []ast.VariableDeclarator{
					{
						ID: &ast.Identifier{Name: "src"},
						Init: &ast.CallExpression{
							Callee: &ast.Identifier{Name: "input"},
							Arguments: []ast.Expression{
								&ast.Identifier{Name: "hl2"},
							},
						},
					},
				},
			},
			&ast.VariableDeclaration{
				Kind: "let",
				Declarations: []ast.VariableDeclarator{
					{
						ID: &ast.Identifier{Name: "value"},
						Init: &ast.BinaryExpression{
							Operator: "-",
							Left:     &ast.Identifier{Name: "src"},
							Right:    &ast.Literal{Value: 10.0},
						},
					},
				},
			},
		},
	}

	goCode, err := GenerateStrategyCodeFromAST(program)
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}

	fullCode := goCode.FunctionBody

	if !strings.Contains(fullCode, "srcSeries.GetCurrent()") {
		t.Error("input.source in expression must use srcSeries.GetCurrent(), not bare identifier")
	}

	if strings.Contains(fullCode, "Set((src -") || strings.Contains(fullCode, "Set((src +") ||
		strings.Contains(fullCode, "Set((src *") || strings.Contains(fullCode, "Set((src /") {
		t.Error("REGRESSION: input.source used as bare identifier instead of Series accessor")
	}
}

func TestInputSourceInConditionalExpressions(t *testing.T) {
	program := &ast.Program{
		NodeType: ast.TypeProgram,
		Body: []ast.Node{
			&ast.VariableDeclaration{
				Kind: "let",
				Declarations: []ast.VariableDeclarator{
					{
						ID: &ast.Identifier{Name: "src"},
						Init: &ast.CallExpression{
							Callee: &ast.Identifier{Name: "input"},
							Arguments: []ast.Expression{
								&ast.Identifier{Name: "hl2"},
							},
						},
					},
				},
			},
			&ast.VariableDeclaration{
				Kind: "let",
				Declarations: []ast.VariableDeclarator{
					{
						ID: &ast.Identifier{Name: "signal"},
						Init: &ast.BinaryExpression{
							Operator: ">",
							Left:     &ast.Identifier{Name: "src"},
							Right:    &ast.Literal{Value: 100.0},
						},
					},
				},
			},
		},
	}

	goCode, err := GenerateStrategyCodeFromAST(program)
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}

	fullCode := goCode.FunctionBody

	/* Must use srcSeries.GetCurrent() in condition */
	if !strings.Contains(fullCode, "srcSeries.GetCurrent()") {
		t.Error("input.source in condition must use srcSeries.GetCurrent()")
	}
}

func TestMultipleInputSourceVariables(t *testing.T) {
	program := &ast.Program{
		NodeType: ast.TypeProgram,
		Body: []ast.Node{
			&ast.VariableDeclaration{
				Kind: "let",
				Declarations: []ast.VariableDeclarator{
					{
						ID: &ast.Identifier{Name: "src1"},
						Init: &ast.CallExpression{
							Callee: &ast.Identifier{Name: "input"},
							Arguments: []ast.Expression{
								&ast.Identifier{Name: "hl2"},
							},
						},
					},
				},
			},
			&ast.VariableDeclaration{
				Kind: "let",
				Declarations: []ast.VariableDeclarator{
					{
						ID: &ast.Identifier{Name: "src2"},
						Init: &ast.CallExpression{
							Callee: &ast.Identifier{Name: "input"},
							Arguments: []ast.Expression{
								&ast.Identifier{Name: "close"},
							},
						},
					},
				},
			},
			&ast.VariableDeclaration{
				Kind: "let",
				Declarations: []ast.VariableDeclarator{
					{
						ID: &ast.Identifier{Name: "combined"},
						Init: &ast.BinaryExpression{
							Operator: "+",
							Left:     &ast.Identifier{Name: "src1"},
							Right:    &ast.Identifier{Name: "src2"},
						},
					},
				},
			},
		},
	}

	goCode, err := GenerateStrategyCodeFromAST(program)
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}

	fullCode := goCode.FunctionBody

	if !strings.Contains(fullCode, "src1Series") {
		t.Error("src1 must use Series storage")
	}
	if !strings.Contains(fullCode, "src2Series") {
		t.Error("src2 must use Series storage")
	}

	src1Count := strings.Count(fullCode, "src1Series.GetCurrent()")
	src2Count := strings.Count(fullCode, "src2Series.GetCurrent()")

	if src1Count < 1 {
		t.Error("src1 must use GetCurrent() accessor in expression")
	}
	if src2Count < 1 {
		t.Error("src2 must use GetCurrent() accessor in expression")
	}
}

func TestInputSourceInNestedExpressions(t *testing.T) {
	program := &ast.Program{
		NodeType: ast.TypeProgram,
		Body: []ast.Node{
			&ast.VariableDeclaration{
				Kind: "let",
				Declarations: []ast.VariableDeclarator{
					{
						ID: &ast.Identifier{Name: "src"},
						Init: &ast.CallExpression{
							Callee: &ast.Identifier{Name: "input"},
							Arguments: []ast.Expression{
								&ast.Identifier{Name: "hl2"},
							},
						},
					},
				},
			},
			&ast.VariableDeclaration{
				Kind: "let",
				Declarations: []ast.VariableDeclarator{
					{
						ID: &ast.Identifier{Name: "complex"},
						Init: &ast.BinaryExpression{
							Operator: "/",
							Left: &ast.BinaryExpression{
								Operator: "*",
								Left:     &ast.Identifier{Name: "src"},
								Right:    &ast.Literal{Value: 2.0},
							},
							Right: &ast.BinaryExpression{
								Operator: "+",
								Left:     &ast.Identifier{Name: "src"},
								Right:    &ast.Literal{Value: 10.0},
							},
						},
					},
				},
			},
		},
	}

	goCode, err := GenerateStrategyCodeFromAST(program)
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}

	fullCode := goCode.FunctionBody

	count := strings.Count(fullCode, "srcSeries.GetCurrent()")
	if count < 2 {
		t.Errorf("expected at least 2 occurrences of srcSeries.GetCurrent(), found %d", count)
	}
}
