package codegen

import (
	"strings"
	"testing"
)

func TestBuiltinIdentifierRegistry_IsBuiltinSeriesIdentifier(t *testing.T) {
	registry := NewBuiltinIdentifierRegistry()

	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"close is builtin", "close", true},
		{"open is builtin", "open", true},
		{"high is builtin", "high", true},
		{"low is builtin", "low", true},
		{"volume is builtin", "volume", true},
		{"tr is builtin", "tr", true},
		{"bar_index is builtin", "bar_index", true},
		{"hl2 is builtin", "hl2", true},
		{"hlc3 is builtin", "hlc3", true},
		{"ohlc4 is builtin", "ohlc4", true},
		{"hlcc4 is builtin", "hlcc4", true},
		{"time is builtin", "time", true},
		{"time_close is builtin", "time_close", true},
		{"time_tradingday is builtin", "time_tradingday", true},
		{"user_var not builtin", "user_var", false},
		{"CLOSE uppercase not builtin", "CLOSE", false},
		{"TIME uppercase not builtin", "TIME", false},
		{"HL2 uppercase not builtin", "HL2", false},
		{"empty string not builtin", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := registry.IsBuiltinSeriesIdentifier(tt.input)
			if result != tt.expected {
				t.Errorf("IsBuiltinSeriesIdentifier(%s) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestBuiltinIdentifierRegistry_IsDerivedPrice(t *testing.T) {
	registry := NewBuiltinIdentifierRegistry()

	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"hl2 is derived", "hl2", true},
		{"hlc3 is derived", "hlc3", true},
		{"ohlc4 is derived", "ohlc4", true},
		{"hlcc4 is derived", "hlcc4", true},
		{"close not derived", "close", false},
		{"open not derived", "open", false},
		{"high not derived", "high", false},
		{"low not derived", "low", false},
		{"volume not derived", "volume", false},
		{"tr not derived", "tr", false},
		{"time not derived", "time", false},
		{"user_var not derived", "user_var", false},
		{"HL2 uppercase not derived", "HL2", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := registry.IsDerivedPrice(tt.input)
			if result != tt.expected {
				t.Errorf("IsDerivedPrice(%s) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestBuiltinIdentifierRegistry_IsOHLCVField(t *testing.T) {
	registry := NewBuiltinIdentifierRegistry()

	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"close is OHLCV", "close", true},
		{"open is OHLCV", "open", true},
		{"high is OHLCV", "high", true},
		{"low is OHLCV", "low", true},
		{"volume is OHLCV", "volume", true},
		{"tr is OHLCV", "tr", true},
		{"bar_index is not OHLCV", "bar_index", false},
		{"hl2 not OHLCV", "hl2", false},
		{"hlc3 not OHLCV", "hlc3", false},
		{"ohlc4 not OHLCV", "ohlc4", false},
		{"hlcc4 not OHLCV", "hlcc4", false},
		{"time not OHLCV", "time", false},
		{"user_var not OHLCV", "user_var", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := registry.IsOHLCVField(tt.input)
			if result != tt.expected {
				t.Errorf("IsOHLCVField(%s) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestBuiltinIdentifierRegistry_IsIntegerSeriesBuiltin(t *testing.T) {
	registry := NewBuiltinIdentifierRegistry()

	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"bar_index is integer series", "bar_index", true},
		{"close is not integer series", "close", false},
		{"tr is not integer series", "tr", false},
		{"time is not integer series", "time", false},
		{"hl2 is not integer series", "hl2", false},
		{"last_bar_index is not integer series", "last_bar_index", false},
		{"n v3 alias is not integer series directly", "n", false},
		{"BAR_INDEX uppercase is not integer series", "BAR_INDEX", false},
		{"user_var is not integer series", "user_var", false},
		{"empty string is not integer series", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := registry.IsIntegerSeriesBuiltin(tt.input)
			if result != tt.expected {
				t.Errorf("IsIntegerSeriesBuiltin(%s) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestBuiltinIdentifierRegistry_MutualExclusivity(t *testing.T) {
	registry := NewBuiltinIdentifierRegistry()

	allBuiltins := []string{"close", "open", "high", "low", "volume", "tr", "bar_index", "hl2", "hlc3", "ohlc4", "hlcc4", "time",
		"time_close", "time_tradingday",
		"dayofweek", "dayofmonth", "hour", "minute", "month", "second", "year", "weekofyear"}

	for _, builtin := range allBuiltins {
		t.Run(builtin, func(t *testing.T) {
			isBuiltin := registry.IsBuiltinSeriesIdentifier(builtin)
			isDerived := registry.IsDerivedPrice(builtin)
			isOHLCV := registry.IsOHLCVField(builtin)
			isIntegerSeries := registry.IsIntegerSeriesBuiltin(builtin)
			isTimeSeries := registry.IsTimeSeriesBuiltin(builtin)
			isCalendar := registry.IsCalendarBuiltin(builtin)

			if !isBuiltin {
				t.Errorf("%s should be recognized as builtin", builtin)
			}

			categories := 0
			if isDerived {
				categories++
			}
			if isOHLCV {
				categories++
			}
			if isIntegerSeries {
				categories++
			}
			if isTimeSeries {
				categories++
			}
			if isCalendar {
				categories++
			}

			if categories != 1 {
				t.Errorf("%s must belong to exactly one category (derived=%v, ohlcv=%v, integerSeries=%v, timeSeries=%v, calendar=%v)",
					builtin, isDerived, isOHLCV, isIntegerSeries, isTimeSeries, isCalendar)
			}
		})
	}
}

func TestBuiltinIdentifierRegistry_IsTimeSeriesBuiltin(t *testing.T) {
	registry := NewBuiltinIdentifierRegistry()

	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"time is time-series", "time", true},
		{"time_close is time-series", "time_close", true},
		{"time_tradingday is time-series", "time_tradingday", true},
		{"close not time-series", "close", false},
		{"bar_index not time-series", "bar_index", false},
		{"hl2 not time-series", "hl2", false},
		{"TIME uppercase not time-series", "TIME", false},
		{"empty string not time-series", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := registry.IsTimeSeriesBuiltin(tt.input)
			if result != tt.expected {
				t.Errorf("IsTimeSeriesBuiltin(%s) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestBuiltinIdentifierRegistry_IsCalendarBuiltin(t *testing.T) {
	registry := NewBuiltinIdentifierRegistry()

	calendarNames := []string{"dayofweek", "dayofmonth", "hour", "minute", "month", "second", "year", "weekofyear"}
	for _, name := range calendarNames {
		t.Run(name+" is calendar", func(t *testing.T) {
			if !registry.IsCalendarBuiltin(name) {
				t.Errorf("IsCalendarBuiltin(%s) = false, want true", name)
			}
			if !registry.IsBuiltinSeriesIdentifier(name) {
				t.Errorf("IsBuiltinSeriesIdentifier(%s) = false for calendar builtin", name)
			}
		})
	}

	nonCalendar := []string{"close", "time", "bar_index", "hl2", "na", "unknown", ""}
	for _, name := range nonCalendar {
		t.Run(name+" not calendar", func(t *testing.T) {
			if registry.IsCalendarBuiltin(name) {
				t.Errorf("IsCalendarBuiltin(%s) = true, want false", name)
			}
		})
	}
}

func TestBuiltinIdentifierRegistry_CalendarInfo(t *testing.T) {
	registry := NewBuiltinIdentifierRegistry()

	tests := []struct {
		name       string
		pineName   string
		wantFound  bool
		wantSeries string
		wantField  string
	}{
		{"dayofweek", "dayofweek", true, "dayofweekSeries", "DayOfWeek"},
		{"dayofmonth", "dayofmonth", true, "dayofmonthSeries", "DayOfMonth"},
		{"hour", "hour", true, "hourSeries", "Hour"},
		{"minute", "minute", true, "minuteSeries", "Minute"},
		{"month", "month", true, "monthSeries", "Month"},
		{"second", "second", true, "secondSeries", "Second"},
		{"year", "year", true, "yearSeries", "Year"},
		{"weekofyear", "weekofyear", true, "weekofyearSeries", "WeekOfYear"},
		{"close not calendar", "close", false, "", ""},
		{"empty string", "", false, "", ""},
		{"bar_index not calendar", "bar_index", false, "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, ok := registry.CalendarInfo(tt.pineName)
			if ok != tt.wantFound {
				t.Fatalf("CalendarInfo(%s) found = %v, want %v", tt.pineName, ok, tt.wantFound)
			}
			if !ok {
				return
			}
			if info.SeriesName != tt.wantSeries {
				t.Errorf("SeriesName = %s, want %s", info.SeriesName, tt.wantSeries)
			}
			if info.StructField != tt.wantField {
				t.Errorf("StructField = %s, want %s", info.StructField, tt.wantField)
			}
			if info.PineName != tt.pineName {
				t.Errorf("PineName = %s, want %s", info.PineName, tt.pineName)
			}
			if info.ArrowExpression == "" {
				t.Errorf("ArrowExpression must not be empty for %s", tt.pineName)
			}
			if !strings.Contains(info.ArrowExpression, "float64(") {
				t.Errorf("ArrowExpression for %s must return float64, got: %s", tt.pineName, info.ArrowExpression)
			}
		})
	}
}

/* Pine convention: Sunday=1..Saturday=7. Go Weekday() returns Sunday=0. Offset by +1 */
func TestBuiltinIdentifierRegistry_DayOfWeekPineConvention(t *testing.T) {
	registry := NewBuiltinIdentifierRegistry()
	info, ok := registry.CalendarInfo("dayofweek")
	if !ok {
		t.Fatal("dayofweek not found in calendar registry")
	}
	if !strings.Contains(info.ArrowExpression, "Weekday()") {
		t.Errorf("dayofweek must derive from Go Weekday(), got: %s", info.ArrowExpression)
	}
	if !strings.Contains(info.ArrowExpression, "+ 1") {
		t.Errorf("dayofweek must add 1 for Pine Sunday=1 convention, got: %s", info.ArrowExpression)
	}
}

func TestBuiltinIdentifierRegistry_IsConstantBuiltin(t *testing.T) {
	registry := NewBuiltinIdentifierRegistry()

	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"last_bar_index is constant", "last_bar_index", true},
		{"last_bar_time is constant", "last_bar_time", true},
		{"timenow is constant", "timenow", true},
		{"close not constant", "close", false},
		{"dayofweek not constant", "dayofweek", false},
		{"bar_index not constant", "bar_index", false},
		{"time_close not constant", "time_close", false},
		{"empty string", "", false},
		{"unknown", "unknown", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if registry.IsConstantBuiltin(tt.input) != tt.expected {
				t.Errorf("IsConstantBuiltin(%s) = %v, want %v", tt.input, !tt.expected, tt.expected)
			}
		})
	}

}

func TestBuiltinIdentifierRegistry_ConstantBuiltinsExcludedFromSeriesCategories(t *testing.T) {
	registry := NewBuiltinIdentifierRegistry()

	constants := []string{"last_bar_index", "last_bar_time", "timenow"}

	for _, name := range constants {
		t.Run(name, func(t *testing.T) {
			if registry.IsBuiltinSeriesIdentifier(name) {
				t.Errorf("%s: IsBuiltinSeriesIdentifier = true, constant builtins must not be series", name)
			}
			if registry.IsOHLCVField(name) {
				t.Errorf("%s: IsOHLCVField = true, want false", name)
			}
			if registry.IsIntegerSeriesBuiltin(name) {
				t.Errorf("%s: IsIntegerSeriesBuiltin = true, want false", name)
			}
			if registry.IsTimeSeriesBuiltin(name) {
				t.Errorf("%s: IsTimeSeriesBuiltin = true, want false", name)
			}
			if registry.IsCalendarBuiltin(name) {
				t.Errorf("%s: IsCalendarBuiltin = true, want false", name)
			}
			if registry.IsDerivedPrice(name) {
				t.Errorf("%s: IsDerivedPrice = true, want false", name)
			}
		})
	}
}

func TestBuiltinIdentifierRegistry_IntegerSeriesBuiltinNames(t *testing.T) {
	registry := NewBuiltinIdentifierRegistry()

	names := registry.IntegerSeriesBuiltinNames()

	if len(names) == 0 {
		t.Fatal("IntegerSeriesBuiltinNames() returned empty — enumeration broken")
	}

	nameSet := make(map[string]bool, len(names))
	for _, name := range names {
		if nameSet[name] {
			t.Errorf("IntegerSeriesBuiltinNames() returned duplicate: %q", name)
		}
		nameSet[name] = true

		if !registry.IsIntegerSeriesBuiltin(name) {
			t.Errorf("%q returned by IntegerSeriesBuiltinNames() but IsIntegerSeriesBuiltin = false", name)
		}
		if !registry.IsBuiltinSeriesIdentifier(name) {
			t.Errorf("%q returned by IntegerSeriesBuiltinNames() but IsBuiltinSeriesIdentifier = false", name)
		}
	}

	if !nameSet["bar_index"] {
		t.Error("IntegerSeriesBuiltinNames() must contain bar_index")
	}
}

func TestBuiltinIdentifierRegistry_ResolveAlias(t *testing.T) {
	registry := NewBuiltinIdentifierRegistry()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"n resolves to bar_index (Pine v3 alias)", "n", "bar_index"},
		{"bar_index resolves to itself (no alias)", "bar_index", "bar_index"},
		{"close resolves to itself (no alias)", "close", "close"},
		{"open resolves to itself (no alias)", "open", "open"},
		{"user variable resolves to itself (not an alias)", "myVar", "myVar"},
		{"empty string resolves to itself", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := registry.ResolveAlias(tt.input)
			if result != tt.expected {
				t.Errorf("ResolveAlias(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestBuiltinIdentifierRegistry_AliasedIdentifierIsBuiltin(t *testing.T) {
	registry := NewBuiltinIdentifierRegistry()

	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"n is builtin via alias to bar_index", "n", true},
		{"bar_index is builtin directly", "bar_index", true},
		{"unaliased user var is not builtin", "userVar", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolved := registry.ResolveAlias(tt.input)
			result := registry.IsBuiltinSeriesIdentifier(resolved) || registry.IsConstantBuiltin(resolved)
			if result != tt.expected {
				t.Errorf("IsBuiltin(ResolveAlias(%q)) = %v, want %v (resolved=%q)",
					tt.input, result, tt.expected, resolved)
			}
		})
	}
}
