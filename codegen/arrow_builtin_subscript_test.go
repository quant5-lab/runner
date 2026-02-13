package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

/* Builtins must generate ctx.Data IIFE patterns; user variables must use Series.Get() */
func TestArrowBuiltinSubscript_RoutingCategories(t *testing.T) {
	gen := &generator{
		builtinHandler: NewBuiltinIdentifierHandler(),
	}
	resolver := NewArrowSeriesAccessResolver()
	exprGen := NewArrowExpressionGeneratorImpl(gen, resolver)

	tests := []struct {
		name     string
		builtin  string
		category string
		mustHave []string
		mustNot  []string
	}{
		{"dayofweek", "dayofweek", "calendar", []string{"time.LoadLocation(ctx.Timezone)", "Weekday()"}, []string{"dayofweekSeries.Get("}},
		{"dayofmonth", "dayofmonth", "calendar", []string{"time.LoadLocation(ctx.Timezone)", "barTime.Day()"}, []string{"dayofmonthSeries.Get("}},
		{"hour", "hour", "calendar", []string{"time.LoadLocation(ctx.Timezone)", "barTime.Hour()"}, []string{"hourSeries.Get("}},
		{"minute", "minute", "calendar", []string{"time.LoadLocation(ctx.Timezone)", "barTime.Minute()"}, []string{"minuteSeries.Get("}},
		{"month", "month", "calendar", []string{"time.LoadLocation(ctx.Timezone)", "barTime.Month()"}, []string{"monthSeries.Get("}},
		{"second", "second", "calendar", []string{"time.LoadLocation(ctx.Timezone)", "barTime.Second()"}, []string{"secondSeries.Get("}},
		{"year", "year", "calendar", []string{"time.LoadLocation(ctx.Timezone)", "barTime.Year()"}, []string{"yearSeries.Get("}},
		{"weekofyear", "weekofyear", "calendar", []string{"time.LoadLocation(ctx.Timezone)", "ISOWeek()"}, []string{"weekofyearSeries.Get("}},

		{"time", "time", "time", []string{"float64(ctx.Data[barIdx].Time * 1000)"}, []string{"timeSeries.Get("}},

		{"bar_index", "bar_index", "bar_index", []string{"float64(barIdx)"}, []string{"bar_indexSeries.Get("}},

		{"hl2", "hl2", "derived", []string{"ctx.Data[barIdx].High", "ctx.Data[barIdx].Low"}, []string{"Series.Get("}},
		{"hlc3", "hlc3", "derived", []string{"ctx.Data[barIdx].High", "ctx.Data[barIdx].Low", "ctx.Data[barIdx].Close"}, nil},
		{"ohlc4", "ohlc4", "derived", []string{"ctx.Data[barIdx].Open", "ctx.Data[barIdx].High"}, nil},
		{"hlcc4", "hlcc4", "derived", []string{"ctx.Data[barIdx].Close"}, nil},

		{"close", "close", "ohlcv", []string{"ctx.Data[barIdx].Close"}, []string{"closeSeries.Get("}},
		{"open", "open", "ohlcv", []string{"ctx.Data[barIdx].Open"}, []string{"openSeries.Get("}},
		{"high", "high", "ohlcv", []string{"ctx.Data[barIdx].High"}, []string{"highSeries.Get("}},
		{"low", "low", "ohlcv", []string{"ctx.Data[barIdx].Low"}, []string{"lowSeries.Get("}},
		{"volume", "volume", "ohlcv", []string{"ctx.Data[barIdx].Volume"}, []string{"volumeSeries.Get("}},
		{"tr", "tr", "ohlcv", []string{"ctx.Data[barIdx].Tr"}, []string{"trSeries.Get("}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := exprGen.resolveArrowSubscript(tt.builtin, &ast.Literal{Value: float64(1)})
			if err != nil {
				t.Fatalf("resolveArrowSubscript(%s) error: %v", tt.builtin, err)
			}

			if !strings.Contains(code, "func() float64 {") {
				t.Errorf("builtin %q (%s) must generate IIFE, got: %s", tt.builtin, tt.category, code)
			}

			for _, must := range tt.mustHave {
				if !strings.Contains(code, must) {
					t.Errorf("%q missing routing pattern %q\ngot: %s", tt.builtin, must, code)
				}
			}
			for _, mustNot := range tt.mustNot {
				if strings.Contains(code, mustNot) {
					t.Errorf("%q should not contain %q\ngot: %s", tt.builtin, mustNot, code)
				}
			}
		})
	}
}

/* Every series identifier in the registry must produce a bounds-checked IIFE */
func TestArrowBuiltinSubscript_RegistryCompleteness(t *testing.T) {
	registry := NewBuiltinIdentifierRegistry()

	allSeriesBuiltins := make(map[string]bool)
	for _, name := range registry.OHLCVFieldNames() {
		allSeriesBuiltins[name] = true
	}
	for _, name := range registry.DerivedPriceNames() {
		allSeriesBuiltins[name] = true
	}
	for _, name := range registry.TimeSeriesBuiltinNames() {
		allSeriesBuiltins[name] = true
	}
	for _, name := range registry.CalendarBuiltinNames() {
		allSeriesBuiltins[name] = true
	}

	if len(allSeriesBuiltins) == 0 {
		t.Fatal("registry returned zero series builtins — enumeration methods broken")
	}

	gen := &generator{builtinHandler: NewBuiltinIdentifierHandler()}
	resolver := NewArrowSeriesAccessResolver()
	exprGen := NewArrowExpressionGeneratorImpl(gen, resolver)

	for name := range allSeriesBuiltins {
		t.Run(name, func(t *testing.T) {
			if !registry.IsBuiltinSeriesIdentifier(name) {
				t.Fatalf("%q returned by enumeration but IsBuiltinSeriesIdentifier=false", name)
			}
			code, err := exprGen.resolveArrowSubscript(name, &ast.Literal{Value: float64(1)})
			if err != nil {
				t.Fatalf("resolveArrowSubscript(%s) error: %v", name, err)
			}
			if !strings.Contains(code, "func() float64 {") {
				t.Errorf("registered builtin %q must produce IIFE, got: %s", name, code)
			}
		})
	}
}

/* Constant builtins (last_bar_index) are time-invariant — subscript returns scalar directly */
func TestArrowBuiltinSubscript_ConstantBuiltinNotSeries(t *testing.T) {
	registry := NewBuiltinIdentifierRegistry()
	if !registry.IsConstantBuiltin("last_bar_index") {
		t.Fatal("last_bar_index should be a constant builtin")
	}
	if registry.IsBuiltinSeriesIdentifier("last_bar_index") {
		t.Fatal("last_bar_index should NOT be a series identifier")
	}

	gen := &generator{builtinHandler: NewBuiltinIdentifierHandler()}
	resolver := NewArrowSeriesAccessResolver()
	exprGen := NewArrowExpressionGeneratorImpl(gen, resolver)

	code, err := exprGen.resolveArrowSubscript("last_bar_index", &ast.Literal{Value: float64(1)})
	if err != nil {
		t.Fatalf("resolveArrowSubscript(last_bar_index) error: %v", err)
	}
	if code != "last_bar_index" {
		t.Errorf("constant builtin should return scalar identifier, got: %s", code)
	}
	if strings.Contains(code, "Series.Get(") {
		t.Errorf("constant builtin must NOT use Series.Get(), got: %s", code)
	}
	if strings.Contains(code, "func() float64 {") {
		t.Errorf("constant builtin must NOT produce IIFE, got: %s", code)
	}
}

func TestArrowBuiltinSubscript_CompoundIndexExpression(t *testing.T) {
	gen := &generator{builtinHandler: NewBuiltinIdentifierHandler()}
	resolver := NewArrowSeriesAccessResolver()
	exprGen := NewArrowExpressionGeneratorImpl(gen, resolver)

	code, err := exprGen.resolveArrowSubscript("close", &ast.BinaryExpression{
		Left:     &ast.Identifier{Name: "a"},
		Operator: "+",
		Right:    &ast.Literal{Value: float64(1)},
	})
	if err != nil {
		t.Fatalf("compound index error: %v", err)
	}
	if !strings.Contains(code, "ctx.BarIndex-int(") {
		t.Errorf("compound index must use ctx.BarIndex arithmetic, got: %s", code)
	}
	if !strings.Contains(code, "ctx.Data[barIdx].Close") {
		t.Errorf("compound index must access ctx.Data, got: %s", code)
	}
}

func TestArrowBuiltinSubscript_UserVariableUsesSeriesGet(t *testing.T) {
	gen := &generator{builtinHandler: NewBuiltinIdentifierHandler()}
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
				t.Errorf("user variable %q should use Series.Get(), got: %s", name, code)
			}
			if strings.Contains(code, "ctx.Data[barIdx]") {
				t.Errorf("user variable %q should not use ctx.Data path, got: %s", name, code)
			}
		})
	}
}
