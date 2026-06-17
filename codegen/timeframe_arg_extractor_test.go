package codegen

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestTimeframeArgExtractor_GoExpr(t *testing.T) {
	e := NewTimeframeArgExtractor()

	tests := []struct {
		name string
		arg  ast.Expression
		want string
	}{
		{
			name: "nil_falls_back_to_primary",
			arg:  nil,
			want: "ctx.Timeframe",
		},

		{
			name: "timeframe_period_as_flat_identifier",
			arg:  &ast.Identifier{Name: "timeframe.period"},
			want: "ctx.Timeframe",
		},
		{
			name: "timeframe_period_as_member_expression",
			arg: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "timeframe"},
				Property: &ast.Identifier{Name: "period"},
			},
			want: "ctx.Timeframe",
		},

		{
			name: "intraday_numeric_60",
			arg:  &ast.Literal{Value: "60"},
			want: `"60"`,
		},
		{
			name: "intraday_numeric_240",
			arg:  &ast.Literal{Value: "240"},
			want: `"240"`,
		},
		{
			name: "intraday_suffixed_1h",
			arg:  &ast.Literal{Value: "1h"},
			want: `"1h"`,
		},
		{
			name: "intraday_suffixed_4h",
			arg:  &ast.Literal{Value: "4h"},
			want: `"4h"`,
		},
		{
			name: "daily",
			arg:  &ast.Literal{Value: "1D"},
			want: `"1D"`,
		},
		{
			name: "weekly",
			arg:  &ast.Literal{Value: "W"},
			want: `"W"`,
		},
		{
			name: "monthly",
			arg:  &ast.Literal{Value: "M"},
			want: `"M"`,
		},

		{
			name: "user_variable_tf",
			arg:  &ast.Identifier{Name: "tf"},
			want: "tf",
		},
		{
			name: "user_variable_alt_tf",
			arg:  &ast.Identifier{Name: "altTf"},
			want: "altTf",
		},

		{
			name: "numeric_literal_falls_back_to_primary",
			arg:  &ast.Literal{Value: 42},
			want: "ctx.Timeframe",
		},
		{
			name: "bool_literal_falls_back_to_primary",
			arg:  &ast.Literal{Value: true},
			want: "ctx.Timeframe",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := e.GoExpr(tt.arg)
			if got != tt.want {
				t.Errorf("GoExpr(%T) = %q, want %q", tt.arg, got, tt.want)
			}
		})
	}
}
