package codegen

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestBuiltinIdentifierHandler_IsBuiltinSeriesIdentifier(t *testing.T) {
	handler := NewBuiltinIdentifierHandler()

	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"close builtin", "close", true},
		{"open builtin", "open", true},
		{"high builtin", "high", true},
		{"low builtin", "low", true},
		{"volume builtin", "volume", true},
		{"user variable", "my_var", false},
		{"na builtin", "na", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handler.IsBuiltinSeriesIdentifier(tt.input)
			if result != tt.expected {
				t.Errorf("IsBuiltinSeriesIdentifier(%s) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestBuiltinIdentifierHandler_IsStrategyRuntimeValue(t *testing.T) {
	handler := NewBuiltinIdentifierHandler()

	tests := []struct {
		name     string
		obj      string
		prop     string
		expected bool
	}{
		{"position_avg_price", "strategy", "position_avg_price", true},
		{"position_size", "strategy", "position_size", true},
		{"position_entry_name", "strategy", "position_entry_name", true},
		{"strategy.long constant", "strategy", "long", false},
		{"strategy.short constant", "strategy", "short", false},
		{"non-strategy object", "other", "position_avg_price", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handler.IsStrategyRuntimeValue(tt.obj, tt.prop)
			if result != tt.expected {
				t.Errorf("IsStrategyRuntimeValue(%s, %s) = %v, want %v", tt.obj, tt.prop, result, tt.expected)
			}
		})
	}
}

func TestBuiltinIdentifierHandler_GenerateCurrentBarAccess(t *testing.T) {
	handler := NewBuiltinIdentifierHandler()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"close", "close", "bar.Close"},
		{"open", "open", "bar.Open"},
		{"high", "high", "bar.High"},
		{"low", "low", "bar.Low"},
		{"volume", "volume", "bar.Volume"},
		{"unknown", "unknown", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handler.GenerateCurrentBarAccess(tt.input)
			if result != tt.expected {
				t.Errorf("GenerateCurrentBarAccess(%s) = %s, want %s", tt.input, result, tt.expected)
			}
		})
	}
}

func TestBuiltinIdentifierHandler_GenerateSecurityContextAccess(t *testing.T) {
	handler := NewBuiltinIdentifierHandler()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"close in security", "close", "ctx.Data[ctx.BarIndex].Close"},
		{"open in security", "open", "ctx.Data[ctx.BarIndex].Open"},
		{"high in security", "high", "ctx.Data[ctx.BarIndex].High"},
		{"low in security", "low", "ctx.Data[ctx.BarIndex].Low"},
		{"volume in security", "volume", "ctx.Data[ctx.BarIndex].Volume"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handler.GenerateSecurityContextAccess(tt.input)
			if result != tt.expected {
				t.Errorf("GenerateSecurityContextAccess(%s) = %s, want %s", tt.input, result, tt.expected)
			}
		})
	}
}

func TestBuiltinIdentifierHandler_GenerateHistoricalAccess(t *testing.T) {
	handler := NewBuiltinIdentifierHandler()

	tests := []struct {
		name     string
		builtin  string
		offset   int
		expected string
	}{
		{
			"close[1]",
			"close",
			1,
			"func() float64 { if i-1 >= 0 { return ctx.Data[i-1].Close }; return math.NaN() }()",
		},
		{
			"open[5]",
			"open",
			5,
			"func() float64 { if i-5 >= 0 { return ctx.Data[i-5].Open }; return math.NaN() }()",
		},
		{
			"high[10]",
			"high",
			10,
			"func() float64 { if i-10 >= 0 { return ctx.Data[i-10].High }; return math.NaN() }()",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handler.GenerateHistoricalAccess(tt.builtin, tt.offset)
			if result != tt.expected {
				t.Errorf("GenerateHistoricalAccess(%s, %d) = %s, want %s", tt.builtin, tt.offset, result, tt.expected)
			}
		})
	}
}

func TestBuiltinIdentifierHandler_GenerateStrategyRuntimeAccess(t *testing.T) {
	handler := NewBuiltinIdentifierHandler()

	tests := []struct {
		name     string
		property string
		expected string
	}{
		{"position_avg_price", "position_avg_price", "strat.GetPositionAvgPrice()"},
		{"position_size", "position_size", "strat.GetPositionSize()"},
		{"position_entry_name", "position_entry_name", "strat.GetPositionEntryName()"},
		{"unknown property", "unknown", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handler.GenerateStrategyRuntimeAccess(tt.property)
			if result != tt.expected {
				t.Errorf("GenerateStrategyRuntimeAccess(%s) = %s, want %s", tt.property, result, tt.expected)
			}
		})
	}
}

func TestBuiltinIdentifierHandler_TryResolveIdentifier(t *testing.T) {
	handler := NewBuiltinIdentifierHandler()

	tests := []struct {
		name              string
		identifier        string
		inSecurityContext bool
		expectedCode      string
		expectedResolved  bool
	}{
		{"na identifier", "na", false, "math.NaN()", true},
		{"close current bar", "close", false, "bar.Close", true},
		{"close in security", "close", true, "ctx.Data[ctx.BarIndex].Close", true},
		{"user variable", "my_var", false, "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr := &ast.Identifier{Name: tt.identifier}
			code, resolved := handler.TryResolveIdentifier(expr, tt.inSecurityContext)
			if code != tt.expectedCode || resolved != tt.expectedResolved {
				t.Errorf("TryResolveIdentifier(%s, %v) = (%s, %v), want (%s, %v)",
					tt.identifier, tt.inSecurityContext, code, resolved, tt.expectedCode, tt.expectedResolved)
			}
		})
	}
}

func TestBuiltinIdentifierHandler_TryResolveMemberExpression(t *testing.T) {
	handler := NewBuiltinIdentifierHandler()

	tests := []struct {
		name              string
		obj               string
		prop              string
		computed          bool
		offset            int
		inSecurityContext bool
		expectedCode      string
		expectedResolved  bool
	}{
		{
			"strategy.position_avg_price",
			"strategy",
			"position_avg_price",
			false,
			0,
			false,
			"strat.GetPositionAvgPrice()",
			true,
		},
		{
			"close[0] current bar",
			"close",
			"0",
			true,
			0,
			false,
			"bar.Close",
			true,
		},
		{
			"close[0] in security",
			"close",
			"0",
			true,
			0,
			true,
			"ctx.Data[ctx.BarIndex].Close",
			true,
		},
		{
			"close[1] historical",
			"close",
			"1",
			true,
			1,
			false,
			"func() float64 { if i-1 >= 0 { return ctx.Data[i-1].Close }; return math.NaN() }()",
			true,
		},
		{
			"user variable member",
			"my_var",
			"field",
			false,
			0,
			false,
			"",
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := &ast.Identifier{Name: tt.obj}
			var prop ast.Expression
			if tt.computed {
				prop = &ast.Literal{Value: tt.offset}
			} else {
				prop = &ast.Identifier{Name: tt.prop}
			}

			expr := &ast.MemberExpression{
				Object:   obj,
				Property: prop,
				Computed: tt.computed,
			}

			code, resolved := handler.TryResolveMemberExpression(expr, tt.inSecurityContext)
			if code != tt.expectedCode || resolved != tt.expectedResolved {
				t.Errorf("TryResolveMemberExpression(%s.%s, %v) = (%s, %v), want (%s, %v)",
					tt.obj, tt.prop, tt.inSecurityContext, code, resolved, tt.expectedCode, tt.expectedResolved)
			}
		})
	}
}
