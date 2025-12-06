package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestExtractSeriesExpression(t *testing.T) {
	gen := &generator{
		imports:    make(map[string]bool),
		variables:  make(map[string]string),
		taRegistry: NewTAFunctionRegistry(),
	}

	tests := []struct {
		name     string
		expr     ast.Expression
		expected string
	}{
		{
			name: "close built-in series",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "close"},
				Property: &ast.Literal{Value: 0},
			},
			expected: "bar.Close",
		},
		{
			name: "open built-in series",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "open"},
				Property: &ast.Literal{Value: 0},
			},
			expected: "bar.Open",
		},
		{
			name:     "user variable identifier",
			expr:     &ast.Identifier{Name: "sma20"},
			expected: "sma20Series.GetCurrent()",
		},
		{
			name: "user variable with subscript",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "sma20"},
				Property: &ast.Literal{Value: 0},
			},
			expected: "sma20Series.Get(0)",
		},
		{
			name:     "float literal",
			expr:     &ast.Literal{Value: 100.50},
			expected: "100.50",
		},
		{
			name: "arithmetic expression",
			expr: &ast.BinaryExpression{
				Operator: "*",
				Left:     &ast.Identifier{Name: "sma20"},
				Right:    &ast.Literal{Value: 1.02},
			},
			expected: "(sma20Series.GetCurrent() * 1.02)",
		},
		{
			name: "complex arithmetic",
			expr: &ast.BinaryExpression{
				Operator: "+",
				Left: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "close"},
					Property: &ast.Literal{Value: 0},
				},
				Right: &ast.BinaryExpression{
					Operator: "*",
					Left:     &ast.Identifier{Name: "sma20"},
					Right:    &ast.Literal{Value: 0.05},
				},
			},
			expected: "(bar.Close + (sma20Series.GetCurrent() * 0.05))",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := gen.extractSeriesExpression(tt.expr)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestConvertSeriesAccessToPrev(t *testing.T) {
	gen := &generator{
		imports:    make(map[string]bool),
		variables:  make(map[string]string),
		taRegistry: NewTAFunctionRegistry(),
	}

	tests := []struct {
		name     string
		series   string
		expected string
	}{
		{
			name:     "bar.Close to previous",
			series:   "bar.Close",
			expected: "ctx.Data[i-1].Close",
		},
		{
			name:     "bar.Open to previous",
			series:   "bar.Open",
			expected: "ctx.Data[i-1].Open",
		},
		{
			name:     "bar.High to previous",
			series:   "bar.High",
			expected: "ctx.Data[i-1].High",
		},
		{
			name:     "bar.Low to previous",
			series:   "bar.Low",
			expected: "ctx.Data[i-1].Low",
		},
		{
			name:     "bar.Volume to previous",
			series:   "bar.Volume",
			expected: "ctx.Data[i-1].Volume",
		},
		{
			name:     "user variable (placeholder)",
			series:   "sma20",
			expected: "0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := gen.convertSeriesAccessToPrev(tt.series)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestCrossoverCodegenIntegration(t *testing.T) {
	// Test ta.crossover with close and sma20
	call := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "ta"},
			Property: &ast.Identifier{Name: "crossover"},
		},
		Arguments: []ast.Expression{
			&ast.MemberExpression{
				Object:   &ast.Identifier{Name: "close"},
				Property: &ast.Literal{Value: 0},
			},
			&ast.MemberExpression{
				Object:   &ast.Identifier{Name: "sma20"},
				Property: &ast.Literal{Value: 0},
			},
		},
	}

	gen := &generator{
		imports:    make(map[string]bool),
		variables:  make(map[string]string),
		taRegistry: NewTAFunctionRegistry(),
	}

	code, err := gen.generateVariableFromCall("longCross", call)
	if err != nil {
		t.Fatalf("generateVariableFromCall failed: %v", err)
	}

	// Verify generated code structure (ForwardSeriesBuffer paradigm)
	if !strings.Contains(code, "longCrossSeries.Set(0.0)") {
		t.Error("Missing initial Series.Set(0.0) assignment")
	}
	if !strings.Contains(code, "if i > 0") {
		t.Error("Missing warmup check")
	}
	if !strings.Contains(code, "ctx.Data[i-1].Close") {
		t.Error("Missing previous close access")
	}
	if !strings.Contains(code, "bar.Close > sma20Series.Get(0)") {
		t.Error("Missing crossover condition (current)")
	}
	if !strings.Contains(code, "&&") {
		t.Error("Missing AND operator")
	}
	if !strings.Contains(code, "<=") {
		t.Error("Missing previous comparison operator")
	}
	if !strings.Contains(code, "longCrossSeries.Set(func() float64") {
		t.Error("Missing Series.Set with bool→float64 conversion")
	}
}

func TestCrossunderCodegenIntegration(t *testing.T) {
	// Test ta.crossunder with close and sma20
	call := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "ta"},
			Property: &ast.Identifier{Name: "crossunder"},
		},
		Arguments: []ast.Expression{
			&ast.MemberExpression{
				Object:   &ast.Identifier{Name: "close"},
				Property: &ast.Literal{Value: 0},
			},
			&ast.Identifier{Name: "sma50"},
		},
	}

	gen := &generator{
		imports:    make(map[string]bool),
		variables:  make(map[string]string),
		taRegistry: NewTAFunctionRegistry(),
	}

	code, err := gen.generateVariableFromCall("shortCross", call)
	if err != nil {
		t.Fatalf("generateVariableFromCall failed: %v", err)
	}

	t.Logf("Generated code:\n%s", code)

	// Verify generated code structure (ForwardSeriesBuffer paradigm)
	if !strings.Contains(code, "shortCrossSeries.Set(0.0)") {
		t.Error("Missing initial Series.Set(0.0) assignment")
	}
	if !strings.Contains(code, "if i > 0") {
		t.Error("Missing warmup check")
	}
	// sma50 is an Identifier (not MemberExpression), so it uses GetCurrent()
	if !strings.Contains(code, "bar.Close < sma50Series.GetCurrent()") && !strings.Contains(code, "bar.Close < sma50Series.Get(0)") {
		t.Error("Missing crossunder condition (current below)")
	}
	if !strings.Contains(code, ">=") {
		t.Error("Missing previous >= operator for crossunder")
	}
	if !strings.Contains(code, "shortCrossSeries.Set(func() float64") {
		t.Error("Missing Series.Set with bool→float64 conversion")
	}
}

func TestCrossoverWithArithmetic(t *testing.T) {
	// Test ta.crossover(close, sma20 * 1.02)
	call := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "ta"},
			Property: &ast.Identifier{Name: "crossover"},
		},
		Arguments: []ast.Expression{
			&ast.MemberExpression{
				Object:   &ast.Identifier{Name: "close"},
				Property: &ast.Literal{Value: 0},
			},
			&ast.BinaryExpression{
				Operator: "*",
				Left:     &ast.Identifier{Name: "sma20"},
				Right:    &ast.Literal{Value: 1.02},
			},
		},
	}

	gen := &generator{
		imports:    make(map[string]bool),
		variables:  make(map[string]string),
		taRegistry: NewTAFunctionRegistry(),
	}

	code, err := gen.generateVariableFromCall("crossAboveThreshold", call)
	if err != nil {
		t.Fatalf("generateVariableFromCall failed: %v", err)
	}

	// Verify arithmetic expression in generated code (ForwardSeriesBuffer paradigm)
	if !strings.Contains(code, "(sma20Series.GetCurrent() * 1.02)") {
		t.Error("Missing arithmetic expression in crossover")
	}
	if !strings.Contains(code, "bar.Close > (sma20Series.GetCurrent() * 1.02)") {
		t.Error("Missing arithmetic comparison")
	}
}

func TestBooleanTypeTracking(t *testing.T) {
	program := &ast.Program{
		Body: []ast.Node{
			&ast.VariableDeclaration{
				Declarations: []ast.VariableDeclarator{
					{
						ID: ast.Identifier{Name: "longCross"},
						Init: &ast.CallExpression{
							Callee: &ast.MemberExpression{
								Object:   &ast.Identifier{Name: "ta"},
								Property: &ast.Identifier{Name: "crossover"},
							},
							Arguments: []ast.Expression{
								&ast.Identifier{Name: "close"},
								&ast.Identifier{Name: "sma20"},
							},
						},
					},
				},
			},
			&ast.VariableDeclaration{
				Declarations: []ast.VariableDeclarator{
					{
						ID: ast.Identifier{Name: "sma50"},
						Init: &ast.CallExpression{
							Callee: &ast.MemberExpression{
								Object:   &ast.Identifier{Name: "ta"},
								Property: &ast.Identifier{Name: "sma"},
							},
							Arguments: []ast.Expression{
								&ast.Identifier{Name: "close"},
								&ast.Literal{Value: 50.0},
							},
						},
					},
				},
			},
		},
	}

	gen := &generator{
		imports:    make(map[string]bool),
		variables:  make(map[string]string),
		varInits:   make(map[string]ast.Expression),
		constants:  make(map[string]interface{}),
		taRegistry: NewTAFunctionRegistry(),
	}
	gen.tempVarMgr = NewTempVariableManager(gen)
	gen.exprAnalyzer = NewExpressionAnalyzer(gen)

	code, err := gen.generateProgram(program)
	if err != nil {
		t.Fatalf("generateProgram failed: %v", err)
	}

	// Verify ForwardSeriesBuffer paradigm (ALL variables are *series.Series)
	if !strings.Contains(code, "var longCrossSeries *series.Series") {
		t.Error("longCross should be declared as *series.Series")
	}
	if !strings.Contains(code, "var sma50Series *series.Series") {
		t.Error("sma50 should be declared as *series.Series")
	}
	// Verify type tracking in g.variables map
	if gen.variables["longCross"] != "bool" {
		t.Errorf("longCross should be tracked as bool type, got: %s", gen.variables["longCross"])
	}
	if gen.variables["sma50"] != "float64" {
		t.Errorf("sma50 should be tracked as float64 type, got: %s", gen.variables["sma50"])
	}
}
