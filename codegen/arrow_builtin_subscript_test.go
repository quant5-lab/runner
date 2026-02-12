package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestArrowBuiltinSubscript(t *testing.T) {
	gen := &generator{
		builtinHandler: NewBuiltinIdentifierHandler(),
	}
	resolver := NewArrowSeriesAccessResolver()
	exprGen := NewArrowExpressionGeneratorImpl(gen, resolver)

	tests := []struct {
		name        string
		builtin     string
		mustHave    []string
		mustNotHave []string
	}{
		/* FSB-backed series: calendar builtins use Series.Get() */
		{"dayofweek", "dayofweek", []string{"dayofweekSeries.Get("}, nil},
		{"dayofmonth", "dayofmonth", []string{"dayofmonthSeries.Get("}, nil},
		{"hour", "hour", []string{"hourSeries.Get("}, nil},
		{"minute", "minute", []string{"minuteSeries.Get("}, nil},
		{"month", "month", []string{"monthSeries.Get("}, nil},
		{"second", "second", []string{"secondSeries.Get("}, nil},
		{"year", "year", []string{"yearSeries.Get("}, nil},
		{"weekofyear", "weekofyear", []string{"weekofyearSeries.Get("}, nil},

		/* FSB-backed series: bar_index and time */
		{"bar_index", "bar_index", []string{"bar_indexSeries.Get("}, nil},
		{"time", "time", []string{"timeSeries.Get("}, nil},

		/* Derived prices: formula generation at data offset */
		{"hl2", "hl2", []string{"ctx.Data[barIdx].High", "ctx.Data[barIdx].Low"}, []string{"Series.Get("}},
		{"hlc3", "hlc3", []string{"ctx.Data[barIdx].High", "ctx.Data[barIdx].Low", "ctx.Data[barIdx].Close"}, nil},
		{"ohlc4", "ohlc4", []string{"ctx.Data[barIdx].Open", "ctx.Data[barIdx].High"}, nil},
		{"hlcc4", "hlcc4", []string{"ctx.Data[barIdx].Close"}, nil},

		/* OHLCV: direct ctx.Data field access */
		{"close", "close", []string{"ctx.Data[barIdx].Close"}, []string{"Series.Get("}},
		{"open", "open", []string{"ctx.Data[barIdx].Open"}, nil},
		{"high", "high", []string{"ctx.Data[barIdx].High"}, nil},
		{"low", "low", []string{"ctx.Data[barIdx].Low"}, nil},
		{"volume", "volume", []string{"ctx.Data[barIdx].Volume"}, nil},
		{"tr", "tr", []string{"ctx.Data[barIdx].Tr"}, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := exprGen.resolveArrowSubscript(tt.builtin, &ast.Literal{Value: float64(1)})
			if err != nil {
				t.Fatalf("resolveArrowSubscript(%s) error: %v", tt.builtin, err)
			}
			for _, must := range tt.mustHave {
				if !strings.Contains(code, must) {
					t.Errorf("resolveArrowSubscript(%s) missing %q\ngot: %s", tt.builtin, must, code)
				}
			}
			for _, mustNot := range tt.mustNotHave {
				if strings.Contains(code, mustNot) {
					t.Errorf("resolveArrowSubscript(%s) should not contain %q\ngot: %s", tt.builtin, mustNot, code)
				}
			}
		})
	}
}

func TestArrowBuiltinSubscript_UserVariableUsesGenericSeriesGet(t *testing.T) {
	gen := &generator{
		builtinHandler: NewBuiltinIdentifierHandler(),
	}
	resolver := NewArrowSeriesAccessResolver()
	exprGen := NewArrowExpressionGeneratorImpl(gen, resolver)

	userVars := []string{"my_var", "sma_result", "custom_series", "x"}
	for _, name := range userVars {
		t.Run(name, func(t *testing.T) {
			code, err := exprGen.resolveArrowSubscript(name, &ast.Literal{Value: float64(1)})
			if err != nil {
				t.Fatalf("resolveArrowSubscript(%s) error: %v", name, err)
			}
			if !strings.Contains(code, name+"Series.Get(") {
				t.Errorf("user variable %q should use generic Series.Get() path, got: %s", name, code)
			}
			if strings.Contains(code, "ctx.Data[barIdx]") {
				t.Errorf("user variable %q should not use builtin ctx.Data path, got: %s", name, code)
			}
		})
	}
}
