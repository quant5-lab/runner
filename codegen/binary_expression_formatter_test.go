package codegen

import (
	"fmt"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestBinaryExpressionFormatter_OperatorPrecedence(t *testing.T) {
	extract := func(expr ast.Expression) (string, error) {
		switch e := expr.(type) {
		case *ast.Identifier:
			return e.Name, nil
		case *ast.Literal:
			return fmt.Sprintf("%v", e.Value), nil
		}
		return "?", fmt.Errorf("unexpected expression type: %T", expr)
	}

	tests := []struct {
		name string
		ast  *ast.BinaryExpression
		want string
	}{
		{
			name: "(a - b) / b * c",
			ast: &ast.BinaryExpression{
				Operator: "*",
				Left: &ast.BinaryExpression{
					Operator: "/",
					Left: &ast.BinaryExpression{
						Operator: "-",
						Left:     &ast.Identifier{Name: "a"},
						Right:    &ast.Identifier{Name: "b"},
					},
					Right: &ast.Identifier{Name: "b"},
				},
				Right: &ast.Identifier{Name: "c"},
			},
			want: "((a - b) / b * c)",
		},
		{
			name: "(a - b) / b * 100",
			ast: &ast.BinaryExpression{
				Operator: "*",
				Left: &ast.BinaryExpression{
					Operator: "/",
					Left: &ast.BinaryExpression{
						Operator: "-",
						Left:     &ast.Identifier{Name: "a"},
						Right:    &ast.Identifier{Name: "b"},
					},
					Right: &ast.Identifier{Name: "b"},
				},
				Right: &ast.Literal{Value: float64(100)},
			},
			want: "((a - b) / b * 100)",
		},
		{
			name: "a + b * c",
			ast: &ast.BinaryExpression{
				Operator: "+",
				Left:     &ast.Identifier{Name: "a"},
				Right: &ast.BinaryExpression{
					Operator: "*",
					Left:     &ast.Identifier{Name: "b"},
					Right:    &ast.Identifier{Name: "c"},
				},
			},
			want: "(a + b * c)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := NewBinaryExpressionFormatter(extract)
			got, err := formatter.Format(tt.ast)
			if err != nil {
				t.Fatalf("Format() error = %v", err)
			}
			if got != tt.want {
				t.Errorf("Format() = %q, want %q", got, tt.want)
			}
		})
	}
}
