package codegen

import (
	"fmt"
	"strings"
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
		{"tr builtin", "tr", true},
		{"time builtin", "time", true},
		{"bar_index builtin", "bar_index", true},

		/* calendar builtins are series identifiers */
		{"dayofweek builtin", "dayofweek", true},
		{"hour builtin", "hour", true},
		{"year builtin", "year", true},
		{"weekofyear builtin", "weekofyear", true},

		/* derived prices are series identifiers */
		{"hl2 builtin", "hl2", true},
		{"hlc3 builtin", "hlc3", true},

		/* non-series */
		{"user variable", "my_var", false},
		{"na builtin", "na", false},
		{"last_bar_index not series", "last_bar_index", false},
		{"empty string", "", false},
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
		{"equity", "strategy", "equity", true},
		{"netprofit", "strategy", "netprofit", true},
		{"closedtrades", "strategy", "closedtrades", true},
		{"initial_capital", "strategy", "initial_capital", true},
		{"grossprofit", "strategy", "grossprofit", true},
		{"grossloss", "strategy", "grossloss", true},
		{"wintrades", "strategy", "wintrades", true},
		{"losstrades", "strategy", "losstrades", true},
		{"eventrades", "strategy", "eventrades", true},
		{"openprofit", "strategy", "openprofit", true},
		{"opentrades", "strategy", "opentrades", true},
		{"avg_trade", "strategy", "avg_trade", true},
		{"avg_winning_trade", "strategy", "avg_winning_trade", true},
		{"avg_losing_trade", "strategy", "avg_losing_trade", true},
		{"max_drawdown", "strategy", "max_drawdown", true},
		{"max_runup", "strategy", "max_runup", true},
		{"max_drawdown_percent", "strategy", "max_drawdown_percent", true},
		{"max_runup_percent", "strategy", "max_runup_percent", true},
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
		{"bar_index", "bar_index", "float64(i)"},
		{"time", "time", "float64(bar.Time * 1000)"},
		{"time_close", "time_close", "time_closeSeries.GetCurrent()"},
		{"time_tradingday", "time_tradingday", "time_tradingdaySeries.GetCurrent()"},

		/* derived prices */
		{"hl2", "hl2", "((bar.High + bar.Low) / 2)"},
		{"hlc3", "hlc3", "((bar.High + bar.Low + bar.Close) / 3)"},
		{"ohlc4", "ohlc4", "((bar.Open + bar.High + bar.Low + bar.Close) / 4)"},
		{"hlcc4", "hlcc4", "((bar.High + bar.Low + bar.Close + bar.Close) / 4)"},

		/* calendar builtins */
		{"dayofweek", "dayofweek", "dayofweekSeries.GetCurrent()"},
		{"dayofmonth", "dayofmonth", "dayofmonthSeries.GetCurrent()"},
		{"hour", "hour", "hourSeries.GetCurrent()"},
		{"minute", "minute", "minuteSeries.GetCurrent()"},
		{"month", "month", "monthSeries.GetCurrent()"},
		{"second", "second", "secondSeries.GetCurrent()"},
		{"year", "year", "yearSeries.GetCurrent()"},
		{"weekofyear", "weekofyear", "weekofyearSeries.GetCurrent()"},

		/* constant builtins */
		{"last_bar_index", "last_bar_index", "last_bar_index"},
		{"last_bar_time", "last_bar_time", "last_bar_time"},
		{"timenow", "timenow", "timenow"},

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

func TestBuiltinIdentifierHandler_GenerateCurrentBarAccess_TrueRange(t *testing.T) {
	handler := NewBuiltinIdentifierHandler()

	result := handler.GenerateCurrentBarAccess("tr")

	expectedComponents := []string{
		"bar.High", "bar.Low",
		"ctx.Data", "Close",
		"math.Max", "math.Abs",
		"if ctx.BarIndex < 1",
	}

	for _, component := range expectedComponents {
		if !contains(result, component) {
			t.Errorf("GenerateCurrentBarAccess(tr) missing expected component: %s\nGot: %s", component, result)
		}
	}

	if !contains(result, "func() float64") {
		t.Errorf("GenerateCurrentBarAccess(tr) should wrap in IIFE\nGot: %s", result)
	}
}

func TestBuiltinIdentifierHandler_GenerateSecurityContextAccess(t *testing.T) {
	handler := NewBuiltinIdentifierHandler()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"close in security", "close", "closeSeries.GetCurrent()"},
		{"open in security", "open", "openSeries.GetCurrent()"},
		{"high in security", "high", "highSeries.GetCurrent()"},
		{"low in security", "low", "lowSeries.GetCurrent()"},
		{"volume in security", "volume", "volumeSeries.GetCurrent()"},
		{"bar_index in security", "bar_index", "float64(ctx.BarIndex)"},
		{"time in security", "time", "timeSeries.GetCurrent()"},
		{"time_close in security", "time_close", "time_closeSeries.GetCurrent()"},
		{"time_tradingday in security", "time_tradingday", "time_tradingdaySeries.GetCurrent()"},

		/* derived prices in security */
		{"hl2 in security", "hl2", "((highSeries.GetCurrent() + lowSeries.GetCurrent()) / 2)"},
		{"hlc3 in security", "hlc3", "((highSeries.GetCurrent() + lowSeries.GetCurrent() + closeSeries.GetCurrent()) / 3)"},
		{"ohlc4 in security", "ohlc4", "((openSeries.GetCurrent() + highSeries.GetCurrent() + lowSeries.GetCurrent() + closeSeries.GetCurrent()) / 4)"},
		{"hlcc4 in security", "hlcc4", "((highSeries.GetCurrent() + lowSeries.GetCurrent() + closeSeries.GetCurrent() + closeSeries.GetCurrent()) / 4)"},

		/* calendar builtins in security */
		{"dayofweek in security", "dayofweek", "dayofweekSeries.GetCurrent()"},
		{"dayofmonth in security", "dayofmonth", "dayofmonthSeries.GetCurrent()"},
		{"hour in security", "hour", "hourSeries.GetCurrent()"},
		{"minute in security", "minute", "minuteSeries.GetCurrent()"},
		{"month in security", "month", "monthSeries.GetCurrent()"},
		{"second in security", "second", "secondSeries.GetCurrent()"},
		{"year in security", "year", "yearSeries.GetCurrent()"},
		{"weekofyear in security", "weekofyear", "weekofyearSeries.GetCurrent()"},

		/* constant builtins in security */
		{"last_bar_index in security", "last_bar_index", "last_bar_index"},
		{"last_bar_time in security", "last_bar_time", "last_bar_time"},
		{"timenow in security", "timenow", "timenow"},
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

func TestBuiltinIdentifierHandler_GenerateSecurityContextAccess_TrueRange(t *testing.T) {
	handler := NewBuiltinIdentifierHandler()

	result := handler.GenerateSecurityContextAccess("tr")

	expectedComponents := []string{
		"highSeries.GetCurrent()",
		"lowSeries.GetCurrent()",
		"Close",
		"math.Max", "math.Abs",
		"if ctx.BarIndex < 1",
	}

	for _, component := range expectedComponents {
		if !contains(result, component) {
			t.Errorf("GenerateSecurityContextAccess(tr) missing expected component: %s\nGot: %s", component, result)
		}
	}

	if !contains(result, "func() float64") {
		t.Errorf("GenerateSecurityContextAccess(tr) should wrap in IIFE\nGot: %s", result)
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
		{
			"bar_index[1]",
			"bar_index",
			1,
			"bar_indexSeries.Get(1)",
		},
		{
			"bar_index[5]",
			"bar_index",
			5,
			"bar_indexSeries.Get(5)",
		},
		{
			"time[1]",
			"time",
			1,
			"timeSeries.Get(1)",
		},
		{
			"time[5]",
			"time",
			5,
			"timeSeries.Get(5)",
		},

		/* calendar builtins historical */
		{"dayofweek[1]", "dayofweek", 1, "dayofweekSeries.Get(1)"},
		{"dayofmonth[3]", "dayofmonth", 3, "dayofmonthSeries.Get(3)"},
		{"hour[2]", "hour", 2, "hourSeries.Get(2)"},
		{"minute[1]", "minute", 1, "minuteSeries.Get(1)"},
		{"month[5]", "month", 5, "monthSeries.Get(5)"},
		{"second[1]", "second", 1, "secondSeries.Get(1)"},
		{"year[10]", "year", 10, "yearSeries.Get(10)"},
		{"weekofyear[4]", "weekofyear", 4, "weekofyearSeries.Get(4)"},

		/* constant builtins historical — returns constant regardless of offset */
		{"last_bar_index[1]", "last_bar_index", 1, "last_bar_index"},
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

func TestBuiltinIdentifierHandler_GenerateHistoricalAccess_TrueRange(t *testing.T) {
	handler := NewBuiltinIdentifierHandler()

	tests := []struct {
		name   string
		offset int
	}{
		{"tr[1]", 1},
		{"tr[5]", 5},
		{"tr[10]", 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handler.GenerateHistoricalAccess("tr", tt.offset)

			expectedComponents := []string{
				"func() float64",
				"ctx.Data",
				"math.Max",
				"math.Abs",
				"High", "Low", "Close",
			}

			for _, component := range expectedComponents {
				if !contains(result, component) {
					t.Errorf("GenerateHistoricalAccess(tr, %d) missing expected component: %s\nGot: %s", tt.offset, component, result)
				}
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
		{"position_avg_price", "position_avg_price", "strategy_position_avg_priceSeries.Get(0)"},
		{"position_size", "position_size", "strategy_position_sizeSeries.Get(0)"},
		{"position_entry_name", "position_entry_name", "strat.GetPositionEntryName()"},
		{"equity", "equity", "strategy_equitySeries.Get(0)"},
		{"netprofit", "netprofit", "strategy_netprofitSeries.Get(0)"},
		{"closedtrades", "closedtrades", "strategy_closedtradesSeries.Get(0)"},
		{"initial_capital", "initial_capital", "strategy_initial_capitalSeries.Get(0)"},
		{"grossprofit", "grossprofit", "strategy_grossprofitSeries.Get(0)"},
		{"grossloss", "grossloss", "strategy_grosslossSeries.Get(0)"},
		{"wintrades", "wintrades", "strategy_wintradesSeries.Get(0)"},
		{"losstrades", "losstrades", "strategy_losstradesSeries.Get(0)"},
		{"eventrades", "eventrades", "strategy_eventradesSeries.Get(0)"},
		{"openprofit", "openprofit", "strategy_openprofitSeries.Get(0)"},
		{"opentrades", "opentrades", "strategy_opentradesSeries.Get(0)"},
		{"avg_trade", "avg_trade", "strategy_avg_tradeSeries.Get(0)"},
		{"avg_winning_trade", "avg_winning_trade", "strategy_avg_winning_tradeSeries.Get(0)"},
		{"avg_losing_trade", "avg_losing_trade", "strategy_avg_losing_tradeSeries.Get(0)"},
		{"max_drawdown", "max_drawdown", "strategy_max_drawdownSeries.Get(0)"},
		{"max_runup", "max_runup", "strategy_max_runupSeries.Get(0)"},
		{"max_drawdown_percent", "max_drawdown_percent", "strategy_max_drawdown_percentSeries.Get(0)"},
		{"max_runup_percent", "max_runup_percent", "strategy_max_runup_percentSeries.Get(0)"},
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
		name             string
		identifier       string
		scope            AccessScope
		expectedCode     string
		expectedResolved bool
	}{
		{"na identifier", "na", BarLoopScope, "math.NaN()", true},
		{"close current bar", "close", BarLoopScope, "bar.Close", true},
		{"close in security", "close", SecurityScope, "closeSeries.GetCurrent()", true},
		{"time current bar", "time", BarLoopScope, "float64(bar.Time * 1000)", true},
		{"time in security", "time", SecurityScope, "timeSeries.GetCurrent()", true},

		/* calendar builtins */
		{"dayofweek current", "dayofweek", BarLoopScope, "dayofweekSeries.GetCurrent()", true},
		{"dayofweek in security", "dayofweek", SecurityScope, "dayofweekSeries.GetCurrent()", true},
		{"hour current", "hour", BarLoopScope, "hourSeries.GetCurrent()", true},
		{"year current", "year", BarLoopScope, "yearSeries.GetCurrent()", true},

		/* constant builtins */
		{"last_bar_index current", "last_bar_index", BarLoopScope, "last_bar_index", true},
		{"last_bar_index in security", "last_bar_index", SecurityScope, "last_bar_index", true},

		/* arrow scope — direct bar fields */
		{"close in arrow", "close", ArrowScope, "ctx.Data[ctx.BarIndex].Close", true},
		{"open in arrow", "open", ArrowScope, "ctx.Data[ctx.BarIndex].Open", true},
		{"high in arrow", "high", ArrowScope, "ctx.Data[ctx.BarIndex].High", true},
		{"low in arrow", "low", ArrowScope, "ctx.Data[ctx.BarIndex].Low", true},
		{"volume in arrow", "volume", ArrowScope, "ctx.Data[ctx.BarIndex].Volume", true},

		/* arrow scope — computed builtins */
		{"bar_index in arrow", "bar_index", ArrowScope, "float64(ctx.BarIndex)", true},
		{"time in arrow", "time", ArrowScope, "float64(ctx.Data[ctx.BarIndex].Time * 1000)", true},
		{"last_bar_index in arrow", "last_bar_index", ArrowScope, "float64(len(ctx.Data) - 1)", true},
		{"last_bar_time in arrow", "last_bar_time", ArrowScope, "float64(ctx.Data[len(ctx.Data)-1].Time * 1000)", true},
		{"timenow in arrow", "timenow", ArrowScope, "float64(ctx.Data[len(ctx.Data)-1].Time * 1000)", true},

		/* arrow scope — calendar builtins use LookupSeries */
		{"dayofweek in arrow", "dayofweek", ArrowScope,
			`func() float64 { if s, ok := ctx.LookupSeries("dayofweekSeries"); ok { return s.GetCurrent() }; return math.NaN() }()`, true},
		{"hour in arrow", "hour", ArrowScope,
			`func() float64 { if s, ok := ctx.LookupSeries("hourSeries"); ok { return s.GetCurrent() }; return math.NaN() }()`, true},

		/* arrow scope — time_close and time_tradingday use LookupSeries */
		{"time_close in arrow", "time_close", ArrowScope,
			`func() float64 { if s, ok := ctx.LookupSeries("time_closeSeries"); ok { return s.GetCurrent() }; return math.NaN() }()`, true},
		{"time_tradingday in arrow", "time_tradingday", ArrowScope,
			`func() float64 { if s, ok := ctx.LookupSeries("time_tradingdaySeries"); ok { return s.GetCurrent() }; return math.NaN() }()`, true},

		/* na is scope-independent */
		{"na in arrow", "na", ArrowScope, "math.NaN()", true},

		{"user variable", "my_var", BarLoopScope, "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr := &ast.Identifier{Name: tt.identifier}
			code, resolved := handler.TryResolveIdentifier(expr, tt.scope)
			if code != tt.expectedCode || resolved != tt.expectedResolved {
				t.Errorf("TryResolveIdentifier(%s, %v) = (%s, %v), want (%s, %v)",
					tt.identifier, tt.scope, code, resolved, tt.expectedCode, tt.expectedResolved)
			}
		})
	}
}

func TestBuiltinIdentifierHandler_TryResolveIdentifier_TrueRange(t *testing.T) {
	handler := NewBuiltinIdentifierHandler()

	tests := []struct {
		name             string
		scope            AccessScope
		expectedResolved bool
	}{
		{"tr current bar", BarLoopScope, true},
		{"tr in security", SecurityScope, true},
		{"tr in arrow", ArrowScope, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr := &ast.Identifier{Name: "tr"}
			code, resolved := handler.TryResolveIdentifier(expr, tt.scope)

			if resolved != tt.expectedResolved {
				t.Errorf("TryResolveIdentifier(tr, %v) resolved = %v, want %v", tt.scope, resolved, tt.expectedResolved)
			}

			if resolved {
				switch tt.scope {
				case SecurityScope:
					expectedComponents := []string{"math.Max", "highSeries.GetCurrent()", "lowSeries.GetCurrent()", "closeSeries.Get(1)"}
					for _, component := range expectedComponents {
						if !contains(code, component) {
							t.Errorf("TryResolveIdentifier(tr, %v) missing component: %s\nGot: %s", tt.scope, component, code)
						}
					}
				case ArrowScope:
					expectedComponents := []string{"math.Max", "ctx.Data[ctx.BarIndex]", "curBar", "prevClose"}
					for _, component := range expectedComponents {
						if !contains(code, component) {
							t.Errorf("TryResolveIdentifier(tr, %v) missing component: %s\nGot: %s", tt.scope, component, code)
						}
					}
				default:
					expectedComponents := []string{"math.Max", "bar.High", "bar.Low", "ctx.Data[ctx.BarIndex-1].Close"}
					for _, component := range expectedComponents {
						if !contains(code, component) {
							t.Errorf("TryResolveIdentifier(tr, %v) missing component: %s\nGot: %s", tt.scope, component, code)
						}
					}
				}
			}
		})
	}
}

func TestBuiltinIdentifierHandler_TryResolveMemberExpression(t *testing.T) {
	handler := NewBuiltinIdentifierHandler()

	tests := []struct {
		name             string
		obj              string
		prop             string
		computed         bool
		offset           int
		scope            AccessScope
		expectedCode     string
		expectedResolved bool
	}{
		{
			"strategy.position_avg_price",
			"strategy",
			"position_avg_price",
			false,
			0,
			BarLoopScope,
			"strategy_position_avg_priceSeries.Get(0)",
			true,
		},
		{
			"close[0] current bar",
			"close",
			"0",
			true,
			0,
			BarLoopScope,
			"bar.Close",
			true,
		},
		{
			"close[0] in security",
			"close",
			"0",
			true,
			0,
			SecurityScope,
			"closeSeries.GetCurrent()",
			true,
		},
		{
			"close[1] historical",
			"close",
			"1",
			true,
			1,
			BarLoopScope,
			"func() float64 { if i-1 >= 0 { return ctx.Data[i-1].Close }; return math.NaN() }()",
			true,
		},
		{
			"user variable member",
			"my_var",
			"field",
			false,
			0,
			BarLoopScope,
			"",
			false,
		},
		{
			"time[0] current bar",
			"time",
			"0",
			true,
			0,
			BarLoopScope,
			"float64(bar.Time * 1000)",
			true,
		},
		{
			"time[1] historical",
			"time",
			"1",
			true,
			1,
			BarLoopScope,
			"timeSeries.Get(1)",
			true,
		},
		{
			"namespace delegation - barstate.isfirst",
			"barstate",
			"isfirst",
			false,
			0,
			BarLoopScope,
			"(ctx.BarIndex == 0)",
			true,
		},
		{
			"namespace delegation - timeframe.period",
			"timeframe",
			"period",
			false,
			0,
			BarLoopScope,
			"ctx.Timeframe",
			true,
		},
		{
			"namespace delegation - syminfo.tickerid",
			"syminfo",
			"tickerid",
			false,
			0,
			BarLoopScope,
			"syminfo_tickerid",
			true,
		},
		{
			"namespace delegation - dayofweek.sunday",
			"dayofweek",
			"sunday",
			false,
			0,
			BarLoopScope,
			"1.0",
			true,
		},
		{
			"calendar dayofweek[0] current bar",
			"dayofweek",
			"0",
			true,
			0,
			BarLoopScope,
			"dayofweekSeries.GetCurrent()",
			true,
		},
		{
			"calendar dayofweek[1] historical",
			"dayofweek",
			"1",
			true,
			1,
			BarLoopScope,
			"dayofweekSeries.Get(1)",
			true,
		},
		{
			"calendar hour[3] historical",
			"hour",
			"3",
			true,
			3,
			BarLoopScope,
			"hourSeries.Get(3)",
			true,
		},
		/* arrow scope — strategy uses LookupSeries */
		{
			"strategy.position_avg_price in arrow",
			"strategy",
			"position_avg_price",
			false,
			0,
			ArrowScope,
			`func() float64 { if s, ok := ctx.LookupSeries("strategy_position_avg_priceSeries"); ok { return s.GetCurrent() }; return math.NaN() }()`,
			true,
		},
		{
			"strategy.equity in arrow",
			"strategy",
			"equity",
			false,
			0,
			ArrowScope,
			`func() float64 { if s, ok := ctx.LookupSeries("strategy_equitySeries"); ok { return s.GetCurrent() }; return math.NaN() }()`,
			true,
		},
		/* arrow scope — subscript current value */
		{
			"close[0] in arrow",
			"close",
			"0",
			true,
			0,
			ArrowScope,
			"ctx.Data[ctx.BarIndex].Close",
			true,
		},
		{
			"time[0] in arrow",
			"time",
			"0",
			true,
			0,
			ArrowScope,
			"float64(ctx.Data[ctx.BarIndex].Time * 1000)",
			true,
		},
		/* arrow scope — subscript historical uses ctx.BarIndex offset */
		{
			"close[1] in arrow",
			"close",
			"1",
			true,
			1,
			ArrowScope,
			"func() float64 { if ctx.BarIndex-1 < 0 { return math.NaN() }; return ctx.Data[ctx.BarIndex-1].Close }()",
			true,
		},
		{
			"time[1] in arrow",
			"time",
			"1",
			true,
			1,
			ArrowScope,
			"func() float64 { if ctx.BarIndex-1 < 0 { return math.NaN() }; return float64(ctx.Data[ctx.BarIndex-1].Time * 1000) }()",
			true,
		},
		/* arrow scope — calendar historical uses LookupSeries with offset */
		{
			"calendar dayofweek[1] in arrow",
			"dayofweek",
			"1",
			true,
			1,
			ArrowScope,
			`func() float64 { if s, ok := ctx.LookupSeries("dayofweekSeries"); ok { return s.Get(1) }; return math.NaN() }()`,
			true,
		},
		/* arrow scope — namespace delegation */
		{
			"barstate.isfirst in arrow",
			"barstate",
			"isfirst",
			false,
			0,
			ArrowScope,
			"(ctx.BarIndex == 0)",
			true,
		},
		{
			"syminfo.tickerid in arrow (overridden)",
			"syminfo",
			"tickerid",
			false,
			0,
			ArrowScope,
			"ctx.Symbol",
			true,
		},
		{
			"dayofweek.sunday in arrow",
			"dayofweek",
			"sunday",
			false,
			0,
			ArrowScope,
			"1.0",
			true,
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

			code, resolved := handler.TryResolveMemberExpression(expr, tt.scope)
			if code != tt.expectedCode || resolved != tt.expectedResolved {
				t.Errorf("TryResolveMemberExpression(%s.%s, %v) = (%s, %v), want (%s, %v)",
					tt.obj, tt.prop, tt.scope, code, resolved, tt.expectedCode, tt.expectedResolved)
			}
		})
	}
}

func TestBuiltinIdentifierHandler_CalendarBuiltinNames(t *testing.T) {
	handler := NewBuiltinIdentifierHandler()
	names := handler.CalendarBuiltinNames()

	expected := map[string]bool{
		"dayofweek": true, "dayofmonth": true, "hour": true, "minute": true,
		"month": true, "second": true, "year": true, "weekofyear": true,
	}

	if len(names) != len(expected) {
		t.Fatalf("CalendarBuiltinNames() returned %d names, want %d", len(names), len(expected))
	}
	for _, name := range names {
		if !expected[name] {
			t.Errorf("unexpected calendar name: %s", name)
		}
	}
}

func TestBuiltinIdentifierHandler_CalendarInfo(t *testing.T) {
	handler := NewBuiltinIdentifierHandler()

	tests := []struct {
		name      string
		input     string
		wantFound bool
	}{
		{"dayofweek found", "dayofweek", true},
		{"hour found", "hour", true},
		{"close not calendar", "close", false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, found := handler.CalendarInfo(tt.input)
			if found != tt.wantFound {
				t.Fatalf("CalendarInfo(%s) found = %v, want %v", tt.input, found, tt.wantFound)
			}
			if found && info.PineName != tt.input {
				t.Errorf("CalendarInfo(%s).PineName = %s", tt.input, info.PineName)
			}
		})
	}
}

func TestBuiltinIdentifierHandler_IsDerivedPrice(t *testing.T) {
	handler := NewBuiltinIdentifierHandler()

	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"hl2", "hl2", true},
		{"hlc3", "hlc3", true},
		{"ohlc4", "ohlc4", true},
		{"hlcc4", "hlcc4", true},
		{"close not derived", "close", false},
		{"dayofweek not derived", "dayofweek", false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if handler.IsDerivedPrice(tt.input) != tt.expected {
				t.Errorf("IsDerivedPrice(%s) = %v, want %v", tt.input, !tt.expected, tt.expected)
			}
		})
	}
}

func TestBuiltinIdentifierHandler_GenerateDerivedPriceFormula(t *testing.T) {
	handler := NewBuiltinIdentifierHandler()

	tests := []struct {
		name          string
		price         string
		wantAccessors []string
	}{
		{"hl2 uses high and low", "hl2", []string{"H", "L"}},
		{"hlc3 uses high low close", "hlc3", []string{"H", "L", "C"}},
		{"ohlc4 uses all four", "ohlc4", []string{"O", "H", "L", "C"}},
		{"hlcc4 uses high low close", "hlcc4", []string{"H", "L", "C"}},
		{"unknown produces empty", "unknown", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formula := handler.GenerateDerivedPriceFormula(tt.price, "H", "L", "C", "O")
			if tt.wantAccessors != nil && formula == "" {
				t.Errorf("GenerateDerivedPriceFormula(%s) returned empty", tt.price)
			}
			if tt.wantAccessors == nil && formula != "" {
				t.Errorf("GenerateDerivedPriceFormula(%s) should be empty, got: %s", tt.price, formula)
			}
			for _, accessor := range tt.wantAccessors {
				if !contains(formula, accessor) {
					t.Errorf("formula %q missing accessor %q", formula, accessor)
				}
			}
		})
	}
}

func TestBuiltinIdentifierHandler_ResolveCalendarBuiltins(t *testing.T) {
	handler := NewBuiltinIdentifierHandler()

	tests := []struct {
		name     string
		detected map[string]bool
		wantLen  int
	}{
		{"nil map", nil, 0},
		{"empty map", map[string]bool{}, 0},
		{"single calendar", map[string]bool{"dayofweek": true}, 1},
		{"multiple calendar", map[string]bool{"hour": true, "minute": true, "year": true}, 3},
		{"non-calendar filtered out", map[string]bool{"close": true, "unknown": true}, 0},
		{"mixed valid and invalid", map[string]bool{"close": true, "dayofweek": true, "unknown": true}, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolved := handler.ResolveCalendarBuiltins(tt.detected)
			if len(resolved) != tt.wantLen {
				t.Errorf("ResolveCalendarBuiltins() returned %d, want %d", len(resolved), tt.wantLen)
			}
			for _, info := range resolved {
				if info.PineName == "" || info.SeriesName == "" || info.StructField == "" {
					t.Errorf("resolved entry has empty fields: %+v", info)
				}
			}
		})
	}
}

func TestBuiltinIdentifierHandler_ResolveMemberExpressionGoType(t *testing.T) {
	handler := NewBuiltinIdentifierHandler()

	tests := []struct {
		name         string
		expr         *ast.MemberExpression
		expectedType GoValueType
		expectFound  bool
	}{
		/* bool properties */
		{"barstate.islast is bool", identMember("barstate", "islast"), GoBool, true},
		{"barstate.isfirst is bool", identMember("barstate", "isfirst"), GoBool, true},
		{"timeframe.isdaily is bool", identMember("timeframe", "isdaily"), GoBool, true},
		{"session.ismarket is bool", identMember("session", "ismarket"), GoBool, true},
		{"chart.is_standard is bool", identMember("chart", "is_standard"), GoBool, true},

		/* string properties */
		{"syminfo.type is string", identMember("syminfo", "type"), GoString, true},
		{"syminfo.tickerid is string", identMember("syminfo", "tickerid"), GoString, true},
		{"syminfo.currency is string", identMember("syminfo", "currency"), GoString, true},
		{"timeframe.period is string", identMember("timeframe", "period"), GoString, true},
		{"chart.bg_color is string", identMember("chart", "bg_color"), GoString, true},

		/* float64 properties */
		{"timeframe.multiplier is float64", identMember("timeframe", "multiplier"), GoFloat64, true},
		{"syminfo.mintick is float64", identMember("syminfo", "mintick"), GoFloat64, true},
		{"dayofweek.monday is float64", identMember("dayofweek", "monday"), GoFloat64, true},
		{"math.pi is float64", identMember("math", "pi"), GoFloat64, true},

		/* unknown namespace/property */
		{"unknown namespace", identMember("unknown", "prop"), GoFloat64, false},
		{"user variable", identMember("myVar", "field"), GoFloat64, false},

		/* guard clauses: non-Identifier AST node types */
		{"computed subscript property", &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "barstate"},
			Property: &ast.Literal{Value: float64(0)},
			Computed: true,
		}, GoFloat64, false},
		{"nested member as object", &ast.MemberExpression{
			Object: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "a"},
				Property: &ast.Identifier{Name: "b"},
			},
			Property: &ast.Identifier{Name: "c"},
		}, GoFloat64, false},
		{"call expression as object", &ast.MemberExpression{
			Object:   &ast.CallExpression{Callee: &ast.Identifier{Name: "getObj"}},
			Property: &ast.Identifier{Name: "field"},
		}, GoFloat64, false},
		{"literal as object", &ast.MemberExpression{
			Object:   &ast.Literal{Value: float64(42)},
			Property: &ast.Identifier{Name: "field"},
		}, GoFloat64, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			goType, found := handler.ResolveMemberExpressionGoType(tt.expr)
			if found != tt.expectFound {
				t.Fatalf("ResolveMemberExpressionGoType found = %v, want %v", found, tt.expectFound)
			}
			if found && goType != tt.expectedType {
				t.Errorf("ResolveMemberExpressionGoType type = %v, want %v", goType, tt.expectedType)
			}
		})
	}
}

func identMember(obj, prop string) *ast.MemberExpression {
	return &ast.MemberExpression{
		Object:   &ast.Identifier{Name: obj},
		Property: &ast.Identifier{Name: prop},
	}
}

func TestTryResolveMemberExpression_TaBuiltinsNotSupported(t *testing.T) {
	handler := NewBuiltinIdentifierHandler()

	unsupported := []string{"close", "open", "high", "low", "volume"}

	for _, name := range unsupported {
		t.Run("ta."+name, func(t *testing.T) {
			expr := &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "ta"},
				Property: &ast.Identifier{Name: name},
				Computed: false,
			}
			_, resolved := handler.TryResolveMemberExpression(expr, BarLoopScope)
			if resolved {
				t.Errorf("ta.%s should not be resolved (update test if implemented)", name)
			}
		})
	}
}

func TestTryResolveMemberExpression_TrNestedSubscript(t *testing.T) {
	handler := NewBuiltinIdentifierHandler()

	offsets := []int{0, 1, 2, 5}

	scopes := []struct {
		name      string
		scope     AccessScope
		wantParts []string
		wantNot   []string
	}{
		{"bar_loop", BarLoopScope, []string{"math.Max", "math.Abs", "ctx.Data", "High", "Low", "Close"}, []string{"ctx.BarIndex"}},
		{"arrow", ArrowScope, []string{"math.Max", "math.Abs", "ctx.Data[barIdx]", "prevClose"}, []string{"i-"}},
	}

	for _, sc := range scopes {
		for _, offset := range offsets {
			name := fmt.Sprintf("ta.tr[%d]_%s", offset, sc.name)
			t.Run(name, func(t *testing.T) {
				nestedExpr := &ast.MemberExpression{
					Object: &ast.MemberExpression{
						Object:   &ast.Identifier{Name: "ta"},
						Property: &ast.Identifier{Name: "tr"},
						Computed: false,
					},
					Property: &ast.Literal{Value: offset},
					Computed: true,
				}

				code, resolved := handler.TryResolveMemberExpression(nestedExpr, sc.scope)
				if !resolved {
					t.Fatalf("ta.tr[%d] in %s should be resolved", offset, sc.name)
				}

				for _, part := range sc.wantParts {
					if !contains(code, part) {
						t.Errorf("ta.tr[%d] in %s missing %q, got: %s", offset, sc.name, part, code)
					}
				}

				for _, bad := range sc.wantNot {
					if contains(code, bad) {
						t.Errorf("ta.tr[%d] in %s must not contain %q, got: %s", offset, sc.name, bad, code)
					}
				}
			})
		}
	}
}

func TestTryResolveMemberExpression_TrSimpleSubscriptAllScopes(t *testing.T) {
	handler := NewBuiltinIdentifierHandler()

	scopes := []struct {
		name      string
		scope     AccessScope
		wantParts []string
		wantNot   []string
	}{
		{"bar_loop", BarLoopScope, []string{"math.Max", "math.Abs", "ctx.Data", "High", "Low", "Close"}, []string{"ctx.BarIndex"}},
		{"arrow", ArrowScope, []string{"math.Max", "math.Abs", "ctx.Data[barIdx]", "prevClose"}, []string{"i-"}},
	}

	for _, sc := range scopes {
		for _, offset := range []int{1, 3} {
			name := fmt.Sprintf("tr[%d]_%s", offset, sc.name)
			t.Run(name, func(t *testing.T) {
				expr := &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "tr"},
					Property: &ast.Literal{Value: offset},
					Computed: true,
				}

				code, resolved := handler.TryResolveMemberExpression(expr, sc.scope)
				if !resolved {
					t.Fatalf("tr[%d] in %s should be resolved", offset, sc.name)
				}

				for _, part := range sc.wantParts {
					if !contains(code, part) {
						t.Errorf("tr[%d] in %s missing %q, got: %s", offset, sc.name, part, code)
					}
				}

				for _, bad := range sc.wantNot {
					if contains(code, bad) {
						t.Errorf("tr[%d] in %s must not contain %q, got: %s", offset, sc.name, bad, code)
					}
				}
			})
		}
	}
}

func TestTryResolveMemberExpression_SessionNestedSubscript(t *testing.T) {
	handler := NewBuiltinIdentifierHandler()

	tests := []struct {
		name         string
		sessionProp  string
		offset       int
		scope        AccessScope
		wantContains []string
	}{
		{
			"session.isfirstbar[0] bar loop",
			"isfirstbar", 0, BarLoopScope,
			[]string{"session_isfirstbarSeries", "GetCurrent()", "== 1.0"},
		},
		{
			"session.isfirstbar[1] bar loop",
			"isfirstbar", 1, BarLoopScope,
			[]string{"session_isfirstbarSeries", "Get(1)", "== 1.0"},
		},
		{
			"session.isfirstbar[0] arrow",
			"isfirstbar", 0, ArrowScope,
			[]string{"ctx.LookupSeries", "session_isfirstbarSeries", "GetCurrent()"},
		},
		{
			"session.isfirstbar[1] arrow",
			"isfirstbar", 1, ArrowScope,
			[]string{"ctx.LookupSeries", "session_isfirstbarSeries", "Get(1)"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr := &ast.MemberExpression{
				Object: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "session"},
					Property: &ast.Identifier{Name: tt.sessionProp},
					Computed: false,
				},
				Property: &ast.Literal{Value: tt.offset},
				Computed: true,
			}

			code, resolved := handler.TryResolveMemberExpression(expr, tt.scope)
			if !resolved {
				t.Fatalf("%s should be resolved", tt.name)
			}

			for _, want := range tt.wantContains {
				if !contains(code, want) {
					t.Errorf("%s should contain %q, got: %s", tt.name, want, code)
				}
			}
		})
	}
}

func TestTryResolveMemberExpression_NamespaceDelegation(t *testing.T) {
	handler := NewBuiltinIdentifierHandler()
	resolver := NewBuiltinNamespaceResolver()

	namespaces := map[string]string{
		"barstate":  "isfirst",
		"timeframe": "period",
		"syminfo":   "tickerid",
	}

	for ns, prop := range namespaces {
		t.Run(ns+"."+prop, func(t *testing.T) {
			expected, found := resolver.Resolve(ns, prop)
			if !found {
				t.Fatalf("resolver.Resolve(%s, %s) not found", ns, prop)
			}

			expr := &ast.MemberExpression{
				Object:   &ast.Identifier{Name: ns},
				Property: &ast.Identifier{Name: prop},
			}
			handlerCode, handlerFound := handler.TryResolveMemberExpression(expr, BarLoopScope)
			if !handlerFound {
				t.Fatalf("handler.TryResolveMemberExpression(%s.%s) not found", ns, prop)
			}

			if handlerCode != expected.Code {
				t.Errorf("handler returned %q, resolver returned %q", handlerCode, expected.Code)
			}
		})
	}
}

func TestTryResolveMemberExpression_SessionNamespaceArrow(t *testing.T) {
	handler := NewBuiltinIdentifierHandler()

	tests := []struct {
		name         string
		prop         string
		scope        AccessScope
		expectedCode string
	}{
		{"session.isfirstbar bar loop", "isfirstbar", BarLoopScope, "session_isfirstbarSeries.GetCurrent() == 1.0"},
		{"session.isfirstbar arrow", "isfirstbar", ArrowScope,
			`func() bool { if s, ok := ctx.LookupSeries("session_isfirstbarSeries"); ok { return s.GetCurrent() == 1.0 }; return false }()`},
		{"session.ismarket bar loop", "ismarket", BarLoopScope, "true"},
		{"session.ismarket arrow", "ismarket", ArrowScope, "true"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr := &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "session"},
				Property: &ast.Identifier{Name: tt.prop},
			}
			code, resolved := handler.TryResolveMemberExpression(expr, tt.scope)
			if !resolved {
				t.Fatalf("%s should be resolved", tt.name)
			}
			if code != tt.expectedCode {
				t.Errorf("%s = %q, want %q", tt.name, code, tt.expectedCode)
			}
		})
	}
}

func TestTryResolveMemberExpression_TaTrAllScopes(t *testing.T) {
	handler := NewBuiltinIdentifierHandler()

	scopes := []struct {
		name      string
		scope     AccessScope
		wantParts []string
	}{
		{"bar loop", BarLoopScope, []string{"bar.High", "bar.Low", "math.Max"}},
		{"security", SecurityScope, []string{"highSeries.GetCurrent()", "lowSeries.GetCurrent()", "math.Max"}},
		{"arrow", ArrowScope, []string{"ctx.Data[ctx.BarIndex]", "curBar", "math.Max"}},
	}

	for _, tt := range scopes {
		t.Run(tt.name, func(t *testing.T) {
			expr := &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "ta"},
				Property: &ast.Identifier{Name: "tr"},
				Computed: false,
			}
			code, resolved := handler.TryResolveMemberExpression(expr, tt.scope)
			if !resolved {
				t.Fatalf("ta.tr in %s should be resolved", tt.name)
			}
			for _, part := range tt.wantParts {
				if !strings.Contains(code, part) {
					t.Errorf("ta.tr in %s missing %q, got: %s", tt.name, part, code)
				}
			}
		})
	}
}

func TestTryResolveMemberExpression_StrategyHistoricalSubscript(t *testing.T) {
	handler := NewBuiltinIdentifierHandler()

	tests := []struct {
		name         string
		property     string
		offset       int
		scope        AccessScope
		wantContains []string
	}{
		{
			"equity[0] bar loop",
			"equity", 0, BarLoopScope,
			[]string{StrategyEquitySeriesName, "Get(0)"},
		},
		{
			"equity[1] bar loop",
			"equity", 1, BarLoopScope,
			[]string{StrategyEquitySeriesName, "Get(1)"},
		},
		{
			"equity[0] arrow",
			"equity", 0, ArrowScope,
			[]string{"ctx.LookupSeries", StrategyEquitySeriesName, "GetCurrent()"},
		},
		{
			"equity[1] arrow",
			"equity", 1, ArrowScope,
			[]string{"ctx.LookupSeries", StrategyEquitySeriesName, "Get(1)"},
		},
		{
			"closedtrades[2] bar loop",
			"closedtrades", 2, BarLoopScope,
			[]string{StrategyClosedTradesSeriesName, "Get(2)"},
		},
		{
			"wintrades[1] bar loop",
			"wintrades", 1, BarLoopScope,
			[]string{StrategyWinTradesSeriesName, "Get(1)"},
		},
		{
			"avg_trade[1] bar loop",
			"avg_trade", 1, BarLoopScope,
			[]string{StrategyAvgTradeSeriesName, "Get(1)"},
		},
		{
			"max_drawdown[1] bar loop",
			"max_drawdown", 1, BarLoopScope,
			[]string{StrategyMaxDrawdownSeriesName, "Get(1)"},
		},
		{
			"max_runup[1] arrow",
			"max_runup", 1, ArrowScope,
			[]string{"ctx.LookupSeries", StrategyMaxRunupSeriesName, "Get(1)"},
		},
		{
			"avg_winning_trade[2] arrow",
			"avg_winning_trade", 2, ArrowScope,
			[]string{"ctx.LookupSeries", StrategyAvgWinningTradeSeriesName, "Get(2)"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr := &ast.MemberExpression{
				Object: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "strategy"},
					Property: &ast.Identifier{Name: tt.property},
					Computed: false,
				},
				Property: &ast.Literal{Value: tt.offset},
				Computed: true,
			}

			code, resolved := handler.TryResolveMemberExpression(expr, tt.scope)
			if !resolved {
				t.Fatalf("strategy.%s[%d] in %v should be resolved", tt.property, tt.offset, tt.scope)
			}

			for _, want := range tt.wantContains {
				if !contains(code, want) {
					t.Errorf("strategy.%s[%d] in %v missing %q, got: %s", tt.property, tt.offset, tt.scope, want, code)
				}
			}
		})
	}
}

func TestBuiltinIdentifierHandler_StrategyNestedConstants(t *testing.T) {
	handler := NewBuiltinIdentifierHandler()

	tests := []struct {
		name         string
		outerObj     string
		innerProp    string
		leafProp     string
		expectedCode string
		expectFound  bool
	}{
		{"strategy.commission.percent", "strategy", "commission", "percent", `"percent"`, true},
		{"strategy.commission.cash_per_order", "strategy", "commission", "cash_per_order", `"cash_per_order"`, true},
		{"strategy.commission.cash_per_contract", "strategy", "commission", "cash_per_contract", `"cash_per_contract"`, true},
		{"strategy.direction.long", "strategy", "direction", "long", `"long"`, true},
		{"strategy.direction.short", "strategy", "direction", "short", `"short"`, true},
		{"strategy.direction.all", "strategy", "direction", "all", `"all"`, true},
		{"strategy.oca.cancel", "strategy", "oca", "cancel", `"cancel"`, true},
		{"strategy.oca.reduce", "strategy", "oca", "reduce", `"reduce"`, true},
		{"strategy.oca.none", "strategy", "oca", "none", `"none"`, true},
		{"unknown sub-namespace not resolved", "strategy", "unknown", "anything", "", false},
		{"non-strategy 3-level not resolved", "ta", "commission", "percent", "", false},
	}

	/* constants are scope-independent — verify both scopes return identical results */
	for _, scope := range []AccessScope{BarLoopScope, ArrowScope} {
		scope := scope
		for _, tt := range tests {
			t.Run(fmt.Sprintf("%s/%v", tt.name, scope), func(t *testing.T) {
				expr := &ast.MemberExpression{
					Object: &ast.MemberExpression{
						Object:   &ast.Identifier{Name: tt.outerObj},
						Property: &ast.Identifier{Name: tt.innerProp},
						Computed: false,
					},
					Property: &ast.Identifier{Name: tt.leafProp},
					Computed: false,
				}

				code, found := handler.TryResolveMemberExpression(expr, scope)

				if found != tt.expectFound {
					t.Errorf("expected found=%v, got %v (code=%q)", tt.expectFound, found, code)
				}
				if tt.expectFound && code != tt.expectedCode {
					t.Errorf("expected code %q, got %q", tt.expectedCode, code)
				}
			})
		}
	}
}

func TestBuiltinIdentifierHandler_StrategyFlatConstants(t *testing.T) {
	handler := NewBuiltinIdentifierHandler()

	tests := []struct {
		prop         string
		expectedCode string
	}{
		{"long", `"long"`},
		{"short", `"short"`},
		{"both", `"both"`},
		{"cash", `"cash"`},
		{"fixed", `"fixed"`},
		{"percent_of_equity", `"percent_of_equity"`},
		{"account_currency", `"USD"`},
		{"margin_liquidation_price", "math.NaN()"},
	}

	/* constants are scope-independent */
	for _, scope := range []AccessScope{BarLoopScope, ArrowScope} {
		scope := scope
		for _, tt := range tests {
			t.Run(fmt.Sprintf("strategy.%s/%v", tt.prop, scope), func(t *testing.T) {
				expr := &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "strategy"},
					Property: &ast.Identifier{Name: tt.prop},
					Computed: false,
				}

				code, found := handler.TryResolveMemberExpression(expr, scope)

				if !found {
					t.Errorf("strategy.%s should be resolved in %v", tt.prop, scope)
				}
				if code != tt.expectedCode {
					t.Errorf("strategy.%s in %v: expected %q, got %q", tt.prop, scope, tt.expectedCode, code)
				}
			})
		}
	}
}
