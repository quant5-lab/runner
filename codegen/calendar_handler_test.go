package codegen

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestCalendarHandler_CanHandle(t *testing.T) {
	h := NewCalendarHandler()

	tests := []struct {
		name     string
		funcName string
		want     bool
	}{
		{"year", "year", true},
		{"month", "month", true},
		{"dayofweek", "dayofweek", true},
		{"dayofmonth", "dayofmonth", true},
		{"hour", "hour", true},
		{"minute", "minute", true},
		{"second", "second", true},
		{"weekofyear", "weekofyear", true},
		{"timestamp", "timestamp", true},

		{"math.abs", "math.abs", false},
		{"ta.sma", "ta.sma", false},
		{"plot", "plot", false},
		{"close", "close", false},
		{"empty", "", false},
		{"case_sensitive", "Year", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := h.CanHandle(tt.funcName); got != tt.want {
				t.Errorf("CanHandle(%q) = %v, want %v", tt.funcName, got, tt.want)
			}
		})
	}
}

func TestCalendarHandler_ExtractionCodegen(t *testing.T) {
	h := NewCalendarHandler()
	g := newTestGenerator()

	tests := []struct {
		name        string
		funcName    string
		args        []ast.Expression
		wantContain []string
		wantAbsent  []string
	}{
		{
			"year_default_timezone",
			"year",
			[]ast.Expression{&ast.Identifier{Name: "time"}},
			[]string{"calendar.Year", "ctx.Timezone"},
			nil,
		},
		{
			"month_default_timezone",
			"month",
			[]ast.Expression{&ast.Identifier{Name: "time"}},
			[]string{"calendar.Month", "ctx.Timezone"},
			nil,
		},
		{
			"dayofweek_default_timezone",
			"dayofweek",
			[]ast.Expression{&ast.Identifier{Name: "time"}},
			[]string{"calendar.DayOfWeek", "ctx.Timezone"},
			nil,
		},
		{
			"dayofmonth_default_timezone",
			"dayofmonth",
			[]ast.Expression{&ast.Identifier{Name: "time"}},
			[]string{"calendar.DayOfMonth", "ctx.Timezone"},
			nil,
		},
		{
			"hour_default_timezone",
			"hour",
			[]ast.Expression{&ast.Identifier{Name: "time"}},
			[]string{"calendar.Hour", "ctx.Timezone"},
			nil,
		},
		{
			"minute_default_timezone",
			"minute",
			[]ast.Expression{&ast.Identifier{Name: "time"}},
			[]string{"calendar.Minute", "ctx.Timezone"},
			nil,
		},
		{
			"second_default_timezone",
			"second",
			[]ast.Expression{&ast.Identifier{Name: "time"}},
			[]string{"calendar.Second", "ctx.Timezone"},
			nil,
		},
		{
			"weekofyear_default_timezone",
			"weekofyear",
			[]ast.Expression{&ast.Identifier{Name: "time"}},
			[]string{"calendar.WeekOfYear", "ctx.Timezone"},
			nil,
		},
		{
			"explicit_timezone_overrides_default",
			"year",
			[]ast.Expression{
				&ast.Identifier{Name: "time"},
				&ast.Literal{Value: "America/New_York"},
			},
			[]string{"calendar.Year", "America/New_York"},
			[]string{"ctx.Timezone"},
		},
		{
			"variable_timezone",
			"hour",
			[]ast.Expression{
				&ast.Identifier{Name: "time"},
				&ast.Identifier{Name: "myTz"},
			},
			[]string{"calendar.Hour"},
			[]string{"ctx.Timezone"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := h.GenerateCalendarCall(tt.funcName, tt.args, g)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			for _, want := range tt.wantContain {
				if !contains(code, want) {
					t.Errorf("missing %q in: %s", want, code)
				}
			}
			for _, absent := range tt.wantAbsent {
				if contains(code, absent) {
					t.Errorf("unexpected %q in: %s", absent, code)
				}
			}
		})
	}
}

func TestCalendarHandler_TimestampOverloads(t *testing.T) {
	h := NewCalendarHandler()
	g := newTestGenerator()

	tests := []struct {
		name        string
		args        []ast.Expression
		wantContain []string
		wantAbsent  []string
	}{
		{
			"1_arg_string_literal",
			[]ast.Expression{&ast.Literal{Value: "2024-01-15"}},
			[]string{"calendar.TimestampFromString"},
			nil,
		},
		{
			"1_arg_variable_passthrough",
			[]ast.Expression{&ast.Identifier{Name: "dateVar"}},
			nil,
			[]string{"calendar.TimestampFromString", "calendar.Timestamp"},
		},
		{
			"5_args_second_defaults_0",
			[]ast.Expression{
				&ast.Literal{Value: float64(2024)},
				&ast.Literal{Value: float64(1)},
				&ast.Literal{Value: float64(15)},
				&ast.Literal{Value: float64(9)},
				&ast.Literal{Value: float64(30)},
			},
			[]string{"calendar.Timestamp", "ctx.Timezone", ", 0,"},
			nil,
		},
		{
			"6_args_all_numeric",
			[]ast.Expression{
				&ast.Literal{Value: float64(2024)},
				&ast.Literal{Value: float64(1)},
				&ast.Literal{Value: float64(15)},
				&ast.Literal{Value: float64(14)},
				&ast.Literal{Value: float64(30)},
				&ast.Literal{Value: float64(0)},
			},
			[]string{"calendar.Timestamp", "ctx.Timezone"},
			nil,
		},
		{
			"6_args_string_literal_timezone",
			[]ast.Expression{
				&ast.Literal{Value: "America/New_York"},
				&ast.Literal{Value: float64(2024)},
				&ast.Literal{Value: float64(1)},
				&ast.Literal{Value: float64(15)},
				&ast.Literal{Value: float64(9)},
				&ast.Literal{Value: float64(30)},
			},
			[]string{"calendar.Timestamp", "America/New_York", ", 0,"},
			[]string{"ctx.Timezone"},
		},
		{
			"6_args_syminfo_timezone",
			[]ast.Expression{
				&ast.MemberExpression{
					Object:   &ast.Identifier{Name: "syminfo"},
					Property: &ast.Identifier{Name: "timezone"},
				},
				&ast.Literal{Value: float64(2024)},
				&ast.Literal{Value: float64(6)},
				&ast.Literal{Value: float64(19)},
				&ast.Literal{Value: float64(9)},
				&ast.Literal{Value: float64(30)},
			},
			[]string{"calendar.Timestamp", ", 0,"},
			nil,
		},
		{
			"7_args_with_timezone",
			[]ast.Expression{
				&ast.Literal{Value: "GMT+5"},
				&ast.Literal{Value: float64(2024)},
				&ast.Literal{Value: float64(1)},
				&ast.Literal{Value: float64(15)},
				&ast.Literal{Value: float64(0)},
				&ast.Literal{Value: float64(0)},
				&ast.Literal{Value: float64(0)},
			},
			[]string{"calendar.Timestamp", "GMT+5"},
			[]string{"ctx.Timezone"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := h.GenerateCalendarCall("timestamp", tt.args, g)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			for _, want := range tt.wantContain {
				if !contains(code, want) {
					t.Errorf("missing %q in: %s", want, code)
				}
			}
			for _, absent := range tt.wantAbsent {
				if contains(code, absent) {
					t.Errorf("unexpected %q in: %s", absent, code)
				}
			}
		})
	}
}

func TestCalendarHandler_TimezoneDisambiguation(t *testing.T) {
	h := NewCalendarHandler()
	g := newTestGenerator()

	/* 5 numeric components; last is 30 which becomes minute (timezone path)
	   or second (year path) depending on disambiguation */
	numericArgs := func() []ast.Expression {
		return []ast.Expression{
			&ast.Literal{Value: float64(2024)},
			&ast.Literal{Value: float64(1)},
			&ast.Literal{Value: float64(15)},
			&ast.Literal{Value: float64(9)},
			&ast.Literal{Value: float64(30)},
		}
	}

	tests := []struct {
		name         string
		firstArg     ast.Expression
		wantTimezone bool
	}{
		{
			"string_literal_is_timezone",
			&ast.Literal{Value: "UTC"},
			true,
		},
		{
			"syminfo_timezone_is_timezone",
			&ast.MemberExpression{
				Object:   &ast.Identifier{Name: "syminfo"},
				Property: &ast.Identifier{Name: "timezone"},
			},
			true,
		},
		{
			"numeric_literal_is_year",
			&ast.Literal{Value: float64(2024)},
			false,
		},
		{
			"identifier_defaults_to_year",
			&ast.Identifier{Name: "myTimezone"},
			false,
		},
		{
			"non_syminfo_member_defaults_to_year",
			&ast.MemberExpression{
				Object:   &ast.Identifier{Name: "input"},
				Property: &ast.Identifier{Name: "timezone"},
			},
			false,
		},
		{
			"syminfo_non_timezone_defaults_to_year",
			&ast.MemberExpression{
				Object:   &ast.Identifier{Name: "syminfo"},
				Property: &ast.Identifier{Name: "ticker"},
			},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := append([]ast.Expression{tt.firstArg}, numericArgs()...)
			code, err := h.GenerateCalendarCall("timestamp", args, g)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			/* Timezone path: 5 components → second defaults to 0
			   Year path: 6 components → second = 30 (the 6th arg) with ctx.Timezone */
			hasDefaultSecond := contains(code, ", 0,")
			if tt.wantTimezone && !hasDefaultSecond {
				t.Errorf("timezone path should default second to 0, got: %s", code)
			}
			if !tt.wantTimezone && hasDefaultSecond {
				t.Errorf("year path should not default second to 0, got: %s", code)
			}
			if !tt.wantTimezone && !contains(code, "ctx.Timezone") {
				t.Errorf("year path should use ctx.Timezone, got: %s", code)
			}
		})
	}
}

func TestCalendarHandler_ErrorCases(t *testing.T) {
	h := NewCalendarHandler()
	g := newTestGenerator()

	tests := []struct {
		name     string
		funcName string
		argCount int
	}{
		{"extraction_no_args", "year", 0},
		{"timestamp_0_args", "timestamp", 0},
		{"timestamp_2_args", "timestamp", 2},
		{"timestamp_3_args", "timestamp", 3},
		{"timestamp_4_args", "timestamp", 4},
		{"timestamp_8_args", "timestamp", 8},
		{"unsupported_function", "unknown", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := make([]ast.Expression, tt.argCount)
			for i := range args {
				args[i] = &ast.Literal{Value: float64(i)}
			}

			_, err := h.GenerateCalendarCall(tt.funcName, args, g)
			if err == nil {
				t.Errorf("%s() with %d args should error", tt.funcName, tt.argCount)
			}
		})
	}
}
