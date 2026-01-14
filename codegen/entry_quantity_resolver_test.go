package codegen

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
)

/* TestEntryQuantityResolver_LessThanThreeArgs verifies default qty for insufficient arguments */
func TestEntryQuantityResolver_LessThanThreeArgs(t *testing.T) {
	resolver := NewEntryQuantityResolver()
	extractLiteral := func(expr ast.Expression) float64 {
		if lit, ok := expr.(*ast.Literal); ok {
			if val, ok := lit.Value.(float64); ok {
				return val
			}
		}
		return 0
	}

	tests := []struct {
		name        string
		args        []ast.Expression
		defaultQty  float64
		expectedQty float64
	}{
		{
			name:        "no arguments",
			args:        []ast.Expression{},
			defaultQty:  1.0,
			expectedQty: 1.0,
		},
		{
			name: "one argument",
			args: []ast.Expression{
				&ast.Literal{Value: "Buy"},
			},
			defaultQty:  2.0,
			expectedQty: 2.0,
		},
		{
			name: "two arguments",
			args: []ast.Expression{
				&ast.Literal{Value: "Buy"},
				&ast.Identifier{Name: "strategy.long"},
			},
			defaultQty:  3.0,
			expectedQty: 3.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qty := resolver.ResolveQuantity(tt.args, tt.defaultQty, extractLiteral)

			if qty != tt.expectedQty {
				t.Errorf("Expected qty %.2f, got %.2f", tt.expectedQty, qty)
			}
		})
	}
}

/* TestEntryQuantityResolver_ExplicitQuantity verifies explicit quantity extraction */
func TestEntryQuantityResolver_ExplicitQuantity(t *testing.T) {
	resolver := NewEntryQuantityResolver()
	extractLiteral := func(expr ast.Expression) float64 {
		if lit, ok := expr.(*ast.Literal); ok {
			if val, ok := lit.Value.(float64); ok {
				return val
			}
			if val, ok := lit.Value.(int); ok {
				return float64(val)
			}
		}
		return 0
	}

	tests := []struct {
		name        string
		thirdArg    ast.Expression
		defaultQty  float64
		expectedQty float64
	}{
		{
			name:        "float quantity",
			thirdArg:    &ast.Literal{Value: 5.5},
			defaultQty:  1.0,
			expectedQty: 5.5,
		},
		{
			name:        "integer quantity",
			thirdArg:    &ast.Literal{Value: 10},
			defaultQty:  1.0,
			expectedQty: 10.0,
		},
		{
			name:        "fractional quantity",
			thirdArg:    &ast.Literal{Value: 0.25},
			defaultQty:  1.0,
			expectedQty: 0.25,
		},
		{
			name:        "large quantity",
			thirdArg:    &ast.Literal{Value: 1000.0},
			defaultQty:  1.0,
			expectedQty: 1000.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := []ast.Expression{
				&ast.Literal{Value: "Buy"},
				&ast.Identifier{Name: "strategy.long"},
				tt.thirdArg,
			}

			qty := resolver.ResolveQuantity(args, tt.defaultQty, extractLiteral)

			if qty != tt.expectedQty {
				t.Errorf("Expected qty %.2f, got %.2f", tt.expectedQty, qty)
			}
		})
	}
}

/* TestEntryQuantityResolver_NamedParameter verifies named parameter detection */
func TestEntryQuantityResolver_NamedParameter(t *testing.T) {
	resolver := NewEntryQuantityResolver()
	extractLiteral := func(expr ast.Expression) float64 {
		if lit, ok := expr.(*ast.Literal); ok {
			if val, ok := lit.Value.(float64); ok {
				return val
			}
		}
		return 0
	}

	tests := []struct {
		name        string
		thirdArg    ast.Expression
		defaultQty  float64
		expectedQty float64
	}{
		{
			name: "empty object expression",
			thirdArg: &ast.ObjectExpression{
				Properties: []ast.Property{},
			},
			defaultQty:  2.0,
			expectedQty: 2.0,
		},
		{
			name: "object with properties",
			thirdArg: &ast.ObjectExpression{
				Properties: []ast.Property{
					{
						Key:   &ast.Identifier{Name: "stop"},
						Value: &ast.Literal{Value: 100.0},
					},
				},
			},
			defaultQty:  3.0,
			expectedQty: 3.0,
		},
		{
			name: "object with qty property",
			thirdArg: &ast.ObjectExpression{
				Properties: []ast.Property{
					{
						Key:   &ast.Identifier{Name: "qty"},
						Value: &ast.Literal{Value: 5.0},
					},
				},
			},
			defaultQty:  1.0,
			expectedQty: 1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := []ast.Expression{
				&ast.Literal{Value: "Buy"},
				&ast.Identifier{Name: "strategy.long"},
				tt.thirdArg,
			}

			qty := resolver.ResolveQuantity(args, tt.defaultQty, extractLiteral)

			if qty != tt.expectedQty {
				t.Errorf("Expected qty %.2f (default for object expr), got %.2f", tt.expectedQty, qty)
			}
		})
	}
}

/* TestEntryQuantityResolver_ZeroQuantity verifies zero quantity falls back to default */
func TestEntryQuantityResolver_ZeroQuantity(t *testing.T) {
	resolver := NewEntryQuantityResolver()
	extractLiteral := func(expr ast.Expression) float64 {
		if lit, ok := expr.(*ast.Literal); ok {
			if val, ok := lit.Value.(float64); ok {
				return val
			}
		}
		return 0
	}

	args := []ast.Expression{
		&ast.Literal{Value: "Buy"},
		&ast.Identifier{Name: "strategy.long"},
		&ast.Literal{Value: 0.0},
	}

	qty := resolver.ResolveQuantity(args, 5.0, extractLiteral)

	if qty != 5.0 {
		t.Errorf("Zero explicit qty should use default, expected 5.0, got %.2f", qty)
	}
}

/* TestEntryQuantityResolver_NegativeQuantity verifies negative quantity falls back to default */
func TestEntryQuantityResolver_NegativeQuantity(t *testing.T) {
	resolver := NewEntryQuantityResolver()
	extractLiteral := func(expr ast.Expression) float64 {
		if lit, ok := expr.(*ast.Literal); ok {
			if val, ok := lit.Value.(float64); ok {
				return val
			}
		}
		return 0
	}

	args := []ast.Expression{
		&ast.Literal{Value: "Buy"},
		&ast.Identifier{Name: "strategy.long"},
		&ast.Literal{Value: -2.0},
	}

	qty := resolver.ResolveQuantity(args, 3.0, extractLiteral)

	if qty != 3.0 {
		t.Errorf("Negative explicit qty should use default, expected 3.0, got %.2f", qty)
	}
}

/* TestEntryQuantityResolver_NonLiteralThirdArg verifies non-literal argument handling */
func TestEntryQuantityResolver_NonLiteralThirdArg(t *testing.T) {
	resolver := NewEntryQuantityResolver()
	extractLiteral := func(expr ast.Expression) float64 {
		if lit, ok := expr.(*ast.Literal); ok {
			if val, ok := lit.Value.(float64); ok {
				return val
			}
		}
		return 0
	}

	tests := []struct {
		name        string
		thirdArg    ast.Expression
		defaultQty  float64
		expectedQty float64
	}{
		{
			name:        "identifier",
			thirdArg:    &ast.Identifier{Name: "qtyVariable"},
			defaultQty:  2.0,
			expectedQty: 2.0,
		},
		{
			name: "binary expression",
			thirdArg: &ast.BinaryExpression{
				Left:     &ast.Literal{Value: 2.0},
				Operator: "*",
				Right:    &ast.Literal{Value: 3.0},
			},
			defaultQty:  1.0,
			expectedQty: 1.0,
		},
		{
			name: "call expression",
			thirdArg: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "calculateQty"},
			},
			defaultQty:  4.0,
			expectedQty: 4.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := []ast.Expression{
				&ast.Literal{Value: "Buy"},
				&ast.Identifier{Name: "strategy.long"},
				tt.thirdArg,
			}

			qty := resolver.ResolveQuantity(args, tt.defaultQty, extractLiteral)

			if qty != tt.expectedQty {
				t.Errorf("Non-literal should use default, expected %.2f, got %.2f", tt.expectedQty, qty)
			}
		})
	}
}

/* TestEntryQuantityResolver_DifferentDefaults verifies resolver uses provided default */
func TestEntryQuantityResolver_DifferentDefaults(t *testing.T) {
	resolver := NewEntryQuantityResolver()
	extractLiteral := func(expr ast.Expression) float64 {
		return 0
	}

	defaults := []float64{0.5, 1.0, 2.5, 5.0, 10.0, 100.0}

	for _, defaultQty := range defaults {
		args := []ast.Expression{
			&ast.Literal{Value: "Buy"},
			&ast.Identifier{Name: "strategy.long"},
		}

		qty := resolver.ResolveQuantity(args, defaultQty, extractLiteral)

		if qty != defaultQty {
			t.Errorf("For default %.2f, expected qty %.2f, got %.2f", defaultQty, defaultQty, qty)
		}
	}
}

/* TestEntryQuantityResolver_ExtractLiteralFailure verifies fallback when extractor returns zero */
func TestEntryQuantityResolver_ExtractLiteralFailure(t *testing.T) {
	resolver := NewEntryQuantityResolver()
	extractLiteral := func(expr ast.Expression) float64 {
		return 0
	}

	args := []ast.Expression{
		&ast.Literal{Value: "Buy"},
		&ast.Identifier{Name: "strategy.long"},
		&ast.Literal{Value: "not-a-number"},
	}

	qty := resolver.ResolveQuantity(args, 7.0, extractLiteral)

	if qty != 7.0 {
		t.Errorf("Failed extraction should use default, expected 7.0, got %.2f", qty)
	}
}

/* TestEntryQuantityResolver_MoreThanThreeArgs verifies behavior with extra arguments */
func TestEntryQuantityResolver_MoreThanThreeArgs(t *testing.T) {
	resolver := NewEntryQuantityResolver()
	extractLiteral := func(expr ast.Expression) float64 {
		if lit, ok := expr.(*ast.Literal); ok {
			if val, ok := lit.Value.(float64); ok {
				return val
			}
		}
		return 0
	}

	args := []ast.Expression{
		&ast.Literal{Value: "Buy"},
		&ast.Identifier{Name: "strategy.long"},
		&ast.Literal{Value: 8.0},
		&ast.ObjectExpression{},
		&ast.Literal{Value: 10.0},
	}

	qty := resolver.ResolveQuantity(args, 1.0, extractLiteral)

	if qty != 8.0 {
		t.Errorf("Should extract from third arg, expected 8.0, got %.2f", qty)
	}
}
