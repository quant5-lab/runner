package codegen

import (
	"strings"
	"testing"

	"github.com/borisquantlab/pinescript-go/ast"
)

func TestSeriesVariableDetection(t *testing.T) {
	// Program with sma20[1] access - should trigger Series storage
	program := &ast.Program{
		Body: []ast.Node{
			&ast.VariableDeclaration{
				Declarations: []ast.VariableDeclarator{
					{
						ID: ast.Identifier{Name: "sma20"},
						Init: &ast.CallExpression{
							Callee: &ast.MemberExpression{
								Object:   &ast.Identifier{Name: "ta"},
								Property: &ast.Identifier{Name: "sma"},
							},
							Arguments: []ast.Expression{
								&ast.MemberExpression{
									Object:   &ast.Identifier{Name: "close"},
									Property: &ast.Literal{Value: 0},
									Computed: true,
								},
								&ast.Literal{Value: 20},
							},
						},
					},
				},
			},
			&ast.VariableDeclaration{
				Declarations: []ast.VariableDeclarator{
					{
						ID: ast.Identifier{Name: "prev_sma"},
						Init: &ast.MemberExpression{
							Object:   &ast.Identifier{Name: "sma20"},
							Property: &ast.Literal{Value: 1}, // Historical access [1]
							Computed: true,
						},
					},
				},
			},
		},
	}

	code, err := GenerateStrategyCodeFromAST(program)
	if err != nil {
		t.Fatalf("Failed to generate code: %v", err)
	}

	// Should declare sma20 as Series
	if !strings.Contains(code.FunctionBody, "var sma20Series *series.Series") {
		t.Error("Expected sma20Series Series declaration, got:", code.FunctionBody)
	}

	// Should initialize Series
	if !strings.Contains(code.FunctionBody, "sma20Series = series.NewSeries(len(ctx.Data))") {
		t.Error("Expected Series initialization")
	}

	// Should use Series.Set() for sma20 assignment
	if !strings.Contains(code.FunctionBody, "sma20Series.Set(") {
		t.Error("Expected Series.Set() for sma20 assignment")
	}

	// Should use Series.Get(1) for prev_sma access
	if !strings.Contains(code.FunctionBody, "sma20Series.Get(1)") {
		t.Error("Expected sma20Series.Get(1) for historical access")
	}

	// Should advance cursor
	if !strings.Contains(code.FunctionBody, "sma20Series.Next()") {
		t.Error("Expected Series.Next() call")
	}
}

func TestBuiltinSeriesHistoricalAccess(t *testing.T) {
	// Program with close[1] - should use ctx.Data[i-1]
	program := &ast.Program{
		Body: []ast.Node{
			&ast.VariableDeclaration{
				Declarations: []ast.VariableDeclarator{
					{
						ID: ast.Identifier{Name: "prev_close"},
						Init: &ast.MemberExpression{
							Object:   &ast.Identifier{Name: "close"},
							Property: &ast.Literal{Value: 1},
							Computed: true,
						},
					},
				},
			},
		},
	}

	code, err := GenerateStrategyCodeFromAST(program)
	if err != nil {
		t.Fatalf("Failed to generate code: %v", err)
	}

	// Should NOT create Series for builtin close
	if strings.Contains(code.FunctionBody, "closeSeries") {
		t.Error("Should not create Series for builtin close")
	}

	// Should use ctx.Data[i-1].Close for historical access
	if !strings.Contains(code.FunctionBody, "ctx.Data[i-1].Close") {
		t.Error("Expected ctx.Data[i-1].Close for builtin historical access, got:", code.FunctionBody)
	}
}

func TestNoSeriesForSimpleVariable(t *testing.T) {
	// Variable never accessed with [offset > 0] - no Series needed
	program := &ast.Program{
		Body: []ast.Node{
			&ast.VariableDeclaration{
				Declarations: []ast.VariableDeclarator{
					{
						ID:   ast.Identifier{Name: "simple_var"},
						Init: &ast.Literal{Value: 100.0},
					},
				},
			},
		},
	}

	code, err := GenerateStrategyCodeFromAST(program)
	if err != nil {
		t.Fatalf("Failed to generate code: %v", err)
	}

	// Should declare as simple float64
	if !strings.Contains(code.FunctionBody, "var simple_var float64") {
		t.Error("Expected simple float64 declaration")
	}

	// Should NOT create Series
	if strings.Contains(code.FunctionBody, "simple_varSeries") {
		t.Error("Should not create Series for variable without historical access")
	}

	// Should NOT call Series.Next()
	if strings.Contains(code.FunctionBody, "simple_varSeries.Next()") {
		t.Error("Should not call Next() for non-Series variable")
	}
}

func TestSeriesInTernaryCondition(t *testing.T) {
	// close > close[1] ? 1 : 0
	program := &ast.Program{
		Body: []ast.Node{
			&ast.VariableDeclaration{
				Declarations: []ast.VariableDeclarator{
					{
						ID: ast.Identifier{Name: "signal"},
						Init: &ast.ConditionalExpression{
							Test: &ast.BinaryExpression{
								Operator: ">",
								Left: &ast.MemberExpression{
									Object:   &ast.Identifier{Name: "close"},
									Property: &ast.Literal{Value: 0},
									Computed: true,
								},
								Right: &ast.MemberExpression{
									Object:   &ast.Identifier{Name: "close"},
									Property: &ast.Literal{Value: 1},
									Computed: true,
								},
							},
							Consequent: &ast.Literal{Value: 1},
							Alternate:  &ast.Literal{Value: 0},
						},
					},
				},
			},
		},
	}

	code, err := GenerateStrategyCodeFromAST(program)
	if err != nil {
		t.Fatalf("Failed to generate code: %v", err)
	}

	// close is builtin, should use bar.Close and ctx.Data[i-1].Close
	if !strings.Contains(code.FunctionBody, "bar.Close") && !strings.Contains(code.FunctionBody, "ctx.Data[i]") {
		t.Error("Expected bar.Close or ctx.Data[i] for current close, got:", code.FunctionBody)
	}
	if !strings.Contains(code.FunctionBody, "ctx.Data[i-1].Close") {
		t.Error("Expected ctx.Data[i-1].Close for close[1], got:", code.FunctionBody)
	}
}

func TestMultipleSeriesVariables(t *testing.T) {
	// Multiple variables requiring Series
	program := &ast.Program{
		Body: []ast.Node{
			&ast.VariableDeclaration{
				Declarations: []ast.VariableDeclarator{
					{
						ID:   ast.Identifier{Name: "sma20"},
						Init: &ast.Literal{Value: 100.0},
					},
				},
			},
			&ast.VariableDeclaration{
				Declarations: []ast.VariableDeclarator{
					{
						ID:   ast.Identifier{Name: "ema50"},
						Init: &ast.Literal{Value: 110.0},
					},
				},
			},
			&ast.VariableDeclaration{
				Declarations: []ast.VariableDeclarator{
					{
						ID: ast.Identifier{Name: "cross"},
						Init: &ast.BinaryExpression{
							Operator: ">",
							Left: &ast.MemberExpression{
								Object:   &ast.Identifier{Name: "sma20"},
								Property: &ast.Literal{Value: 1},
								Computed: true,
							},
							Right: &ast.MemberExpression{
								Object:   &ast.Identifier{Name: "ema50"},
								Property: &ast.Literal{Value: 1},
								Computed: true,
							},
						},
					},
				},
			},
		},
	}

	code, err := GenerateStrategyCodeFromAST(program)
	if err != nil {
		t.Fatalf("Failed to generate code: %v", err)
	}

	// Should create Series for both sma20 and ema50
	if !strings.Contains(code.FunctionBody, "sma20Series") {
		t.Error("Expected sma20Series")
	}
	if !strings.Contains(code.FunctionBody, "ema50Series") {
		t.Error("Expected ema50Series")
	}

	// Should call Next() for both
	if !strings.Contains(code.FunctionBody, "sma20Series.Next()") {
		t.Error("Expected sma20Series.Next()")
	}
	if !strings.Contains(code.FunctionBody, "ema50Series.Next()") {
		t.Error("Expected ema50Series.Next()")
	}
}
