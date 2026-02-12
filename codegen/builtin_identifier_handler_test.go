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
		{"time", "time", "float64(bar.Time * 1000)"},

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
		{"time in security", "time", "timeSeries.GetCurrent()"},

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
		{"close in security", "close", true, "closeSeries.GetCurrent()", true},
		{"time current bar", "time", false, "float64(bar.Time * 1000)", true},
		{"time in security", "time", true, "timeSeries.GetCurrent()", true},

		/* calendar builtins */
		{"dayofweek current", "dayofweek", false, "dayofweekSeries.GetCurrent()", true},
		{"dayofweek in security", "dayofweek", true, "dayofweekSeries.GetCurrent()", true},
		{"hour current", "hour", false, "hourSeries.GetCurrent()", true},
		{"year current", "year", false, "yearSeries.GetCurrent()", true},

		/* constant builtins */
		{"last_bar_index current", "last_bar_index", false, "last_bar_index", true},
		{"last_bar_index in security", "last_bar_index", true, "last_bar_index", true},

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

func TestBuiltinIdentifierHandler_TryResolveIdentifier_TrueRange(t *testing.T) {
	handler := NewBuiltinIdentifierHandler()

	tests := []struct {
		name              string
		inSecurityContext bool
		expectedResolved  bool
	}{
		{"tr current bar", false, true},
		{"tr in security", true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr := &ast.Identifier{Name: "tr"}
			code, resolved := handler.TryResolveIdentifier(expr, tt.inSecurityContext)

			if resolved != tt.expectedResolved {
				t.Errorf("TryResolveIdentifier(tr, %v) resolved = %v, want %v", tt.inSecurityContext, resolved, tt.expectedResolved)
			}

			if resolved {
				if tt.inSecurityContext {
					expectedComponents := []string{"math.Max", "highSeries.GetCurrent()", "lowSeries.GetCurrent()", "closeSeries.Get(1)"}
					for _, component := range expectedComponents {
						if !contains(code, component) {
							t.Errorf("TryResolveIdentifier(tr, %v) missing component: %s\nGot: %s", tt.inSecurityContext, component, code)
						}
					}
				} else {
					expectedComponents := []string{"math.Max", "bar.High", "bar.Low", "ctx.Data[ctx.BarIndex-1].Close"}
					for _, component := range expectedComponents {
						if !contains(code, component) {
							t.Errorf("TryResolveIdentifier(tr, %v) missing component: %s\nGot: %s", tt.inSecurityContext, component, code)
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
			"strategy_position_avg_priceSeries.Get(0)",
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
			"closeSeries.GetCurrent()",
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
		{
			"time[0] current bar",
			"time",
			"0",
			true,
			0,
			false,
			"float64(bar.Time * 1000)",
			true,
		},
		{
			"time[1] historical",
			"time",
			"1",
			true,
			1,
			false,
			"timeSeries.Get(1)",
			true,
		},
		{
			"namespace delegation - barstate.isfirst",
			"barstate",
			"isfirst",
			false,
			0,
			false,
			"(ctx.BarIndex == 0)",
			true,
		},
		{
			"namespace delegation - timeframe.period",
			"timeframe",
			"period",
			false,
			0,
			false,
			"ctx.Timeframe",
			true,
		},
		{
			"namespace delegation - syminfo.tickerid",
			"syminfo",
			"tickerid",
			false,
			0,
			false,
			"syminfo_tickerid",
			true,
		},
		{
			"namespace delegation - dayofweek.sunday",
			"dayofweek",
			"sunday",
			false,
			0,
			false,
			"1.0",
			true,
		},
		{
			"calendar dayofweek[0] current bar",
			"dayofweek",
			"0",
			true,
			0,
			false,
			"dayofweekSeries.GetCurrent()",
			true,
		},
		{
			"calendar dayofweek[1] historical",
			"dayofweek",
			"1",
			true,
			1,
			false,
			"dayofweekSeries.Get(1)",
			true,
		},
		{
			"calendar hour[3] historical",
			"hour",
			"3",
			true,
			3,
			false,
			"hourSeries.Get(3)",
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

			code, resolved := handler.TryResolveMemberExpression(expr, tt.inSecurityContext)
			if code != tt.expectedCode || resolved != tt.expectedResolved {
				t.Errorf("TryResolveMemberExpression(%s.%s, %v) = (%s, %v), want (%s, %v)",
					tt.obj, tt.prop, tt.inSecurityContext, code, resolved, tt.expectedCode, tt.expectedResolved)
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
