package codegen

import (
	"strings"
	"testing"

	"github.com/borisquantlab/pinescript-go/ast"
)

func TestTernaryCodegenIntegration(t *testing.T) {
	// Test: signal = close > close_avg ? 1 : 0
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
									Property: &ast.Literal{Value: float64(0)},
									Computed: true,
								},
								Right: &ast.Identifier{Name: "close_avg"},
							},
							Consequent: &ast.Literal{
								Value: float64(1),
							},
							Alternate: &ast.Literal{
								Value: float64(0),
							},
						},
					},
				},
			},
		},
	}

	gen := &generator{
		imports:   make(map[string]bool),
		variables: make(map[string]string),
	}

	code, err := gen.generateProgram(program)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Verify generated code structure (ForwardSeriesBuffer paradigm)
	if !strings.Contains(code, "var signalSeries *series.Series") {
		t.Errorf("Missing signal Series declaration: got %s", code)
	}

	if !strings.Contains(code, "if bar.Close > close_avgSeries.GetCurrent() { return 1") {
		t.Errorf("Missing ternary true branch: got %s", code)
	}

	if !strings.Contains(code, "} else { return 0") {
		t.Errorf("Missing ternary false branch: got %s", code)
	}
	
	if !strings.Contains(code, "signalSeries.Set(func() float64") {
		t.Errorf("Missing Series.Set with inline function: got %s", code)
	}
}

func TestTernaryWithArithmetic(t *testing.T) {
	// Test: volume_signal = volume > volume_avg * 1.5 ? 1 : 0
	program := &ast.Program{
		Body: []ast.Node{
			&ast.VariableDeclaration{
				Declarations: []ast.VariableDeclarator{
					{
						ID: ast.Identifier{Name: "volume_signal"},
						Init: &ast.ConditionalExpression{
							Test: &ast.BinaryExpression{
								Operator: ">",
								Left: &ast.MemberExpression{
									Object:   &ast.Identifier{Name: "volume"},
									Property: &ast.Literal{Value: float64(0)},
									Computed: true,
								},
								Right: &ast.BinaryExpression{
									Operator: "*",
									Left:     &ast.Identifier{Name: "volume_avg"},
									Right:    &ast.Literal{Value: float64(1.5)},
								},
							},
							Consequent: &ast.Literal{
								Value: float64(1),
							},
							Alternate: &ast.Literal{
								Value: float64(0),
							},
						},
					},
				},
			},
		},
	}

	gen := &generator{
		imports:   make(map[string]bool),
		variables: make(map[string]string),
	}

	code, err := gen.generateProgram(program)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Verify arithmetic in condition (ForwardSeriesBuffer paradigm)
	if !strings.Contains(code, "volume_avgSeries.GetCurrent() * 1.50") {
		t.Errorf("Missing arithmetic in ternary condition: got %s", code)
	}

	if !strings.Contains(code, "bar.Volume > volume_avgSeries.GetCurrent() * 1.50") {
		t.Errorf("Missing complete condition with arithmetic: got %s", code)
	}
}

func TestTernaryWithLogicalOperators(t *testing.T) {
	// Test: signal = close > open and volume > 1000 ? 1 : 0
	program := &ast.Program{
		Body: []ast.Node{
			&ast.VariableDeclaration{
				Declarations: []ast.VariableDeclarator{
					{
						ID: ast.Identifier{Name: "signal"},
						Init: &ast.ConditionalExpression{
							Test: &ast.LogicalExpression{
								Operator: "and",
								Left: &ast.BinaryExpression{
									Operator: ">",
									Left: &ast.MemberExpression{
										Object:   &ast.Identifier{Name: "close"},
										Property: &ast.Literal{Value: float64(0)},
										Computed: true,
									},
									Right: &ast.MemberExpression{
										Object:   &ast.Identifier{Name: "open"},
										Property: &ast.Literal{Value: float64(0)},
										Computed: true,
									},
								},
								Right: &ast.BinaryExpression{
									Operator: ">",
									Left: &ast.MemberExpression{
										Object:   &ast.Identifier{Name: "volume"},
										Property: &ast.Literal{Value: float64(0)},
										Computed: true,
									},
									Right: &ast.Literal{Value: float64(1000)},
								},
							},
							Consequent: &ast.Literal{
								Value: float64(1),
							},
							Alternate: &ast.Literal{
								Value: float64(0),
							},
						},
					},
				},
			},
		},
	}

	gen := &generator{
		imports:   make(map[string]bool),
		variables: make(map[string]string),
	}

	code, err := gen.generateProgram(program)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Verify logical operator in condition
	if !strings.Contains(code, "&&") {
		t.Errorf("Missing && operator: got %s", code)
	}

	if !strings.Contains(code, "bar.Close > bar.Open") {
		t.Errorf("Missing close > open comparison: got %s", code)
	}

	if !strings.Contains(code, "bar.Volume > 1000") {
		t.Errorf("Missing volume > 1000 comparison: got %s", code)
	}
}
