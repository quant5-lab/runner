package codegen

import (
	"strings"
	"testing"
)

/* Zero overhead for scripts without calendar access */
func TestCalendarSeriesLifecycle(t *testing.T) {
	tests := []struct {
		name     string
		builtins []CalendarBuiltinInfo
		hasUsage bool
	}{
		{"nil slice", nil, false},
		{"empty slice", []CalendarBuiltinInfo{}, false},
		{"single builtin", []CalendarBuiltinInfo{
			{PineName: "dayofweek", SeriesName: "dayofweekSeries", StructField: "DayOfWeek"},
		}, true},
		{"multiple builtins", []CalendarBuiltinInfo{
			{PineName: "hour", SeriesName: "hourSeries", StructField: "Hour"},
			{PineName: "minute", SeriesName: "minuteSeries", StructField: "Minute"},
			{PineName: "dayofweek", SeriesName: "dayofweekSeries", StructField: "DayOfWeek"},
		}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lifecycle := NewCalendarSeriesLifecycle(tt.builtins)

			if lifecycle.HasCalendarUsage() != tt.hasUsage {
				t.Errorf("HasCalendarUsage() = %v, want %v", lifecycle.HasCalendarUsage(), tt.hasUsage)
			}

			methods := map[string]string{
				"declarations":    lifecycle.GenerateDeclarations("\t"),
				"initializations": lifecycle.GenerateInitializations("\t"),
				"timezone":        lifecycle.GenerateTimezoneSetup("\t"),
				"bar population":  lifecycle.GenerateBarPopulation("\t"),
				"advancement":     lifecycle.GenerateAdvancement("\t", "i"),
				"registrations":   lifecycle.GenerateRegistrations("\t"),
				"suppress":        lifecycle.GenerateSuppressUnused("\t"),
			}

			for method, output := range methods {
				if tt.hasUsage && output == "" {
					t.Errorf("%s should produce output when builtins are used", method)
				}
				if !tt.hasUsage && output != "" {
					t.Errorf("%s should be empty when no builtins used, got: %q", method, output)
				}
			}
		})
	}
}

func TestCalendarSeriesLifecycle_CodeContent(t *testing.T) {
	builtins := []CalendarBuiltinInfo{
		{PineName: "dayofweek", SeriesName: "dayofweekSeries", StructField: "DayOfWeek"},
		{PineName: "hour", SeriesName: "hourSeries", StructField: "Hour"},
	}
	lifecycle := NewCalendarSeriesLifecycle(builtins)

	t.Run("declarations", func(t *testing.T) {
		code := lifecycle.GenerateDeclarations("\t")
		for _, s := range []string{"var dayofweekSeries *series.Series", "var hourSeries *series.Series"} {
			if !strings.Contains(code, s) {
				t.Errorf("missing %q in:\n%s", s, code)
			}
		}
	})

	t.Run("initializations", func(t *testing.T) {
		code := lifecycle.GenerateInitializations("\t")
		for _, s := range []string{"dayofweekSeries = series.NewSeries(len(ctx.Data))", "hourSeries = series.NewSeries(len(ctx.Data))"} {
			if !strings.Contains(code, s) {
				t.Errorf("missing %q in:\n%s", s, code)
			}
		}
	})

	t.Run("timezone setup", func(t *testing.T) {
		code := lifecycle.GenerateTimezoneSetup("\t")
		for _, required := range []string{
			"exchangeLoc",
			"time.LoadLocation(ctx.Timezone)",
			"panic(",
		} {
			if !strings.Contains(code, required) {
				t.Errorf("missing %q in:\n%s", required, code)
			}
		}
	})

	t.Run("bar population single decompose call", func(t *testing.T) {
		code := lifecycle.GenerateBarPopulation("\t")
		if cnt := strings.Count(code, "DecomposeBarTime"); cnt != 1 {
			t.Errorf("expected exactly 1 DecomposeBarTime call, got %d", cnt)
		}
		for _, s := range []string{"dayofweekSeries.Set(barCal.DayOfWeek)", "hourSeries.Set(barCal.Hour)"} {
			if !strings.Contains(code, s) {
				t.Errorf("missing %q in:\n%s", s, code)
			}
		}
	})

	t.Run("registrations", func(t *testing.T) {
		code := lifecycle.GenerateRegistrations("\t")
		for _, s := range []string{`ctx.RegisterSeries("dayofweekSeries", dayofweekSeries)`, `ctx.RegisterSeries("hourSeries", hourSeries)`} {
			if !strings.Contains(code, s) {
				t.Errorf("missing %q in:\n%s", s, code)
			}
		}
	})

	t.Run("advancement uses iterVar", func(t *testing.T) {
		code := lifecycle.GenerateAdvancement("\t", "idx")
		if !strings.Contains(code, "idx < barCount-1") {
			t.Errorf("advancement should use custom iterVar, got:\n%s", code)
		}
		for _, s := range []string{"dayofweekSeries.Next()", "hourSeries.Next()"} {
			if !strings.Contains(code, s) {
				t.Errorf("missing %q in:\n%s", s, code)
			}
		}
	})

	t.Run("suppress unused", func(t *testing.T) {
		code := lifecycle.GenerateSuppressUnused("\t")
		for _, s := range []string{"_ = dayofweekSeries", "_ = hourSeries"} {
			if !strings.Contains(code, s) {
				t.Errorf("missing %q in:\n%s", s, code)
			}
		}
	})
}

/* Map iteration order must not affect output */
func TestCalendarSeriesLifecycle_DeterministicOrder(t *testing.T) {
	builtins := []CalendarBuiltinInfo{
		{PineName: "year", SeriesName: "yearSeries", StructField: "Year"},
		{PineName: "hour", SeriesName: "hourSeries", StructField: "Hour"},
		{PineName: "dayofweek", SeriesName: "dayofweekSeries", StructField: "DayOfWeek"},
	}
	lifecycle := NewCalendarSeriesLifecycle(builtins)

	code := lifecycle.GenerateDeclarations("\t")
	dayPos := strings.Index(code, "dayofweek")
	hourPos := strings.Index(code, "hour")
	yearPos := strings.Index(code, "year")

	if dayPos > hourPos || hourPos > yearPos {
		t.Errorf("output not alphabetically ordered: %q", code)
	}
}

func TestCalendarSeriesLifecycle_NilReceiver(t *testing.T) {
	var lifecycle *CalendarSeriesLifecycle

	if lifecycle.HasCalendarUsage() {
		t.Error("nil receiver should report no usage")
	}

	nilSafeMethods := map[string]string{
		"declarations":    lifecycle.GenerateDeclarations("\t"),
		"initializations": lifecycle.GenerateInitializations("\t"),
		"timezone":        lifecycle.GenerateTimezoneSetup("\t"),
		"bar population":  lifecycle.GenerateBarPopulation("\t"),
		"advancement":     lifecycle.GenerateAdvancement("\t", "i"),
		"registrations":   lifecycle.GenerateRegistrations("\t"),
		"suppress":        lifecycle.GenerateSuppressUnused("\t"),
	}
	for name, output := range nilSafeMethods {
		if output != "" {
			t.Errorf("nil receiver %s should be empty, got: %q", name, output)
		}
	}
}

func TestCalendarSeriesLifecycle_SymbolTableRegistration(t *testing.T) {
	builtins := []CalendarBuiltinInfo{
		{PineName: "dayofweek", SeriesName: "dayofweekSeries", StructField: "DayOfWeek"},
		{PineName: "hour", SeriesName: "hourSeries", StructField: "Hour"},
	}
	lifecycle := NewCalendarSeriesLifecycle(builtins)

	st := NewSymbolTable()
	lifecycle.GenerateSymbolTableRegistrations(st)

	for _, name := range []string{"dayofweek", "hour"} {
		if !st.IsSeries(name) {
			t.Errorf("%s should be registered as series in symbol table", name)
		}
	}
}

func TestCalendarSeriesLifecycle_SymbolTableRegistration_NilReceiver(t *testing.T) {
	var lifecycle *CalendarSeriesLifecycle
	st := NewSymbolTable()

	lifecycle.GenerateSymbolTableRegistrations(st)

	if st.IsSeries("dayofweek") {
		t.Error("nil lifecycle should not register anything")
	}
}

func TestCalendarSeriesLifecycle_SymbolTableRegistration_NilTable(t *testing.T) {
	builtins := []CalendarBuiltinInfo{
		{PineName: "hour", SeriesName: "hourSeries", StructField: "Hour"},
	}
	lifecycle := NewCalendarSeriesLifecycle(builtins)

	lifecycle.GenerateSymbolTableRegistrations(nil)
}

func TestCalendarSeriesLifecycle_InputSliceIsolation(t *testing.T) {
	input := []CalendarBuiltinInfo{
		{PineName: "year", SeriesName: "yearSeries", StructField: "Year"},
		{PineName: "hour", SeriesName: "hourSeries", StructField: "Hour"},
	}
	lifecycle := NewCalendarSeriesLifecycle(input)

	input[0] = CalendarBuiltinInfo{PineName: "mutated", SeriesName: "mutatedSeries", StructField: "Mutated"}

	code := lifecycle.GenerateDeclarations("\t")
	if strings.Contains(code, "mutated") {
		t.Error("constructor must copy input — mutation propagated")
	}
	if !strings.Contains(code, "yearSeries") {
		t.Error("original data lost after input mutation")
	}
}
