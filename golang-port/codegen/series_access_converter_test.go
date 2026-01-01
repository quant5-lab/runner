package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestSeriesAccessConverter(t *testing.T) {
	t.Run("scalar identifier unchanged", func(t *testing.T) {
		st := NewSymbolTable()
		st.Register("period", VariableTypeScalar)

		conv := NewSeriesAccessConverter(st, "0")
		expr := &ast.Identifier{Name: "period"}

		code, err := conv.ConvertExpression(expr)
		if err != nil {
			t.Fatalf("ConvertExpression failed: %v", err)
		}
		if code != "period" {
			t.Errorf("got %q, want %q", code, "period")
		}
	})

	t.Run("series identifier with offset 0 returns scalar", func(t *testing.T) {
		st := NewSymbolTable()
		st.Register("sum", VariableTypeSeries)

		conv := NewSeriesAccessConverter(st, "0")
		expr := &ast.Identifier{Name: "sum"}

		code, err := conv.ConvertExpression(expr)
		if err != nil {
			t.Fatalf("ConvertExpression failed: %v", err)
		}
		if code != "sum" {
			t.Errorf("got %q, want %q (offset 0 means current bar scalar)", code, "sum")
		}
	})

	t.Run("series with dynamic offset", func(t *testing.T) {
		st := NewSymbolTable()
		st.Register("plus", VariableTypeSeries)

		conv := NewSeriesAccessConverter(st, "j")
		expr := &ast.Identifier{Name: "plus"}

		code, err := conv.ConvertExpression(expr)
		if err != nil {
			t.Fatalf("ConvertExpression failed: %v", err)
		}
		if code != "plusSeries.Get(j)" {
			t.Errorf("got %q, want %q", code, "plusSeries.Get(j)")
		}
	})

	t.Run("builtin field converted", func(t *testing.T) {
		st := NewSymbolTable()
		conv := NewSeriesAccessConverter(st, "0")

		tests := []struct {
			field string
			want  string
		}{
			{"close", "ctx.Data[i-0].Close"},
			{"high", "ctx.Data[i-0].High"},
			{"volume", "ctx.Data[i-0].Volume"},
		}

		for _, tt := range tests {
			expr := &ast.Identifier{Name: tt.field}
			code, err := conv.ConvertExpression(expr)
			if err != nil {
				t.Errorf("field %s: %v", tt.field, err)
				continue
			}
			if code != tt.want {
				t.Errorf("field %s: got %q, want %q", tt.field, code, tt.want)
			}
		}
	})

	t.Run("binary expression with series", func(t *testing.T) {
		st := NewSymbolTable()
		st.Register("plus", VariableTypeSeries)
		st.Register("minus", VariableTypeSeries)

		conv := NewSeriesAccessConverter(st, "j")
		expr := &ast.BinaryExpression{
			Left:     &ast.Identifier{Name: "plus"},
			Operator: "+",
			Right:    &ast.Identifier{Name: "minus"},
		}

		code, err := conv.ConvertExpression(expr)
		if err != nil {
			t.Fatalf("ConvertExpression failed: %v", err)
		}

		want := "(plusSeries.Get(j) + minusSeries.Get(j))"
		if code != want {
			t.Errorf("got %q, want %q", code, want)
		}
	})

	t.Run("conditional with series", func(t *testing.T) {
		st := NewSymbolTable()
		st.Register("sum", VariableTypeSeries)

		conv := NewSeriesAccessConverter(st, "j")
		expr := &ast.ConditionalExpression{
			Test: &ast.BinaryExpression{
				Left:     &ast.Identifier{Name: "sum"},
				Operator: "==",
				Right:    &ast.Literal{Value: 0.0},
			},
			Consequent: &ast.Literal{Value: 1.0},
			Alternate:  &ast.Identifier{Name: "sum"},
		}

		code, err := conv.ConvertExpression(expr)
		if err != nil {
			t.Fatalf("ConvertExpression failed: %v", err)
		}

		if !strings.Contains(code, "sumSeries.Get(j)") {
			t.Errorf("code should contain sumSeries.Get(j), got: %s", code)
		}
	})

	t.Run("nested expressions", func(t *testing.T) {
		st := NewSymbolTable()
		st.Register("plus", VariableTypeSeries)
		st.Register("minus", VariableTypeSeries)
		st.Register("sum", VariableTypeSeries)

		conv := NewSeriesAccessConverter(st, "j")

		// abs(plus - minus) / sum
		expr := &ast.BinaryExpression{
			Left: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "math.Abs"},
				Arguments: []ast.Expression{
					&ast.BinaryExpression{
						Left:     &ast.Identifier{Name: "plus"},
						Operator: "-",
						Right:    &ast.Identifier{Name: "minus"},
					},
				},
			},
			Operator: "/",
			Right:    &ast.Identifier{Name: "sum"},
		}

		code, err := conv.ConvertExpression(expr)
		if err != nil {
			t.Fatalf("ConvertExpression failed: %v", err)
		}

		// Should contain all series access
		expected := []string{
			"plusSeries.Get(j)",
			"minusSeries.Get(j)",
			"sumSeries.Get(j)",
		}
		for _, exp := range expected {
			if !strings.Contains(code, exp) {
				t.Errorf("code should contain %q, got: %s", exp, code)
			}
		}
	})

	t.Run("literal values", func(t *testing.T) {
		st := NewSymbolTable()
		conv := NewSeriesAccessConverter(st, "0")

		tests := []struct {
			name  string
			value interface{}
			want  string
		}{
			{"float", 3.14, "3.14"},
			{"int", 42, "42"},
			{"bool", true, "true"},
			{"string", "test", `"test"`},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				expr := &ast.Literal{Value: tt.value}
				code, err := conv.ConvertExpression(expr)
				if err != nil {
					t.Fatalf("ConvertExpression failed: %v", err)
				}
				if code != tt.want {
					t.Errorf("got %q, want %q", code, tt.want)
				}
			})
		}
	})
}
