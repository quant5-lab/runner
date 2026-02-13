package codegen

import (
	"strings"
	"testing"
)

func TestTimeSeriesLifecycle(t *testing.T) {
	tests := []struct {
		name              string
		hasTimeClose      bool
		hasTimeTradingday bool
		hasUsage          bool
	}{
		{"neither", false, false, false},
		{"time_close only", true, false, true},
		{"time_tradingday only", false, true, true},
		{"both", true, true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lifecycle := NewTimeSeriesLifecycle(tt.hasTimeClose, tt.hasTimeTradingday)

			if lifecycle.HasUsage() != tt.hasUsage {
				t.Errorf("HasUsage() = %v, want %v", lifecycle.HasUsage(), tt.hasUsage)
			}

			methods := map[string]string{
				"declarations":    lifecycle.GenerateDeclarations("\t"),
				"initializations": lifecycle.GenerateInitializations("\t"),
				"bar population":  lifecycle.GenerateBarPopulation("\t", "i"),
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

func TestTimeSeriesLifecycle_CodeContent(t *testing.T) {
	lifecycle := NewTimeSeriesLifecycle(true, true)

	t.Run("declarations", func(t *testing.T) {
		code := lifecycle.GenerateDeclarations("\t")
		for _, s := range []string{"var time_closeSeries *series.Series", "var time_tradingdaySeries *series.Series"} {
			if !strings.Contains(code, s) {
				t.Errorf("missing %q in:\n%s", s, code)
			}
		}
	})

	t.Run("initializations", func(t *testing.T) {
		code := lifecycle.GenerateInitializations("\t")
		for _, s := range []string{"time_closeSeries = series.NewSeries(len(ctx.Data))", "time_tradingdaySeries = series.NewSeries(len(ctx.Data))"} {
			if !strings.Contains(code, s) {
				t.Errorf("missing %q in:\n%s", s, code)
			}
		}
	})

	t.Run("bar population time_close uses next bar", func(t *testing.T) {
		code := lifecycle.GenerateBarPopulation("\t", "i")
		for _, s := range []string{
			"i < barCount-1",
			"ctx.Data[i+1].Time * 1000",
			"time_closeSeries.Set(",
			"TimeframeToSeconds(ctx.Timeframe)",
		} {
			if !strings.Contains(code, s) {
				t.Errorf("missing %q in:\n%s", s, code)
			}
		}
	})

	t.Run("bar population time_tradingday uses timezone", func(t *testing.T) {
		code := lifecycle.GenerateBarPopulation("\t", "i")
		for _, s := range []string{
			"exchangeLoc",
			"tradingDayStart",
			"time_tradingdaySeries.Set(",
		} {
			if !strings.Contains(code, s) {
				t.Errorf("missing %q in:\n%s", s, code)
			}
		}
	})

	t.Run("registrations", func(t *testing.T) {
		code := lifecycle.GenerateRegistrations("\t")
		for _, s := range []string{
			`ctx.RegisterSeries("time_closeSeries", time_closeSeries)`,
			`ctx.RegisterSeries("time_tradingdaySeries", time_tradingdaySeries)`,
		} {
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
		for _, s := range []string{"time_closeSeries.Next()", "time_tradingdaySeries.Next()"} {
			if !strings.Contains(code, s) {
				t.Errorf("missing %q in:\n%s", s, code)
			}
		}
	})

	t.Run("suppress unused", func(t *testing.T) {
		code := lifecycle.GenerateSuppressUnused("\t")
		for _, s := range []string{"_ = time_closeSeries", "_ = time_tradingdaySeries"} {
			if !strings.Contains(code, s) {
				t.Errorf("missing %q in:\n%s", s, code)
			}
		}
	})
}

/* time_close only lifecycle should not emit time_tradingday code */
func TestTimeSeriesLifecycle_Isolation(t *testing.T) {
	t.Run("time_close only excludes time_tradingday", func(t *testing.T) {
		lifecycle := NewTimeSeriesLifecycle(true, false)
		code := lifecycle.GenerateDeclarations("\t")
		if strings.Contains(code, "time_tradingday") {
			t.Errorf("time_close-only should not contain time_tradingday: %s", code)
		}
	})

	t.Run("time_tradingday only excludes time_close", func(t *testing.T) {
		lifecycle := NewTimeSeriesLifecycle(false, true)
		code := lifecycle.GenerateDeclarations("\t")
		if strings.Contains(code, "time_close") {
			t.Errorf("time_tradingday-only should not contain time_close: %s", code)
		}
	})
}

func TestTimeSeriesLifecycle_NeedsTimezone(t *testing.T) {
	tests := []struct {
		name              string
		hasTimeClose      bool
		hasTimeTradingday bool
		needsTimezone     bool
	}{
		{"neither", false, false, false},
		{"time_close only", true, false, false},
		{"time_tradingday only", false, true, true},
		{"both", true, true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lifecycle := NewTimeSeriesLifecycle(tt.hasTimeClose, tt.hasTimeTradingday)
			if lifecycle.NeedsTimezone() != tt.needsTimezone {
				t.Errorf("NeedsTimezone() = %v, want %v", lifecycle.NeedsTimezone(), tt.needsTimezone)
			}
		})
	}
}

func TestTimeSeriesLifecycle_NilReceiver(t *testing.T) {
	var lifecycle *TimeSeriesLifecycle

	if lifecycle.HasUsage() {
		t.Error("nil receiver should report no usage")
	}
	if lifecycle.NeedsTimezone() {
		t.Error("nil receiver should not need timezone")
	}

	nilSafeMethods := map[string]string{
		"declarations":    lifecycle.GenerateDeclarations("\t"),
		"initializations": lifecycle.GenerateInitializations("\t"),
		"bar population":  lifecycle.GenerateBarPopulation("\t", "i"),
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

func TestTimeSeriesLifecycle_SymbolTableRegistration(t *testing.T) {
	lifecycle := NewTimeSeriesLifecycle(true, true)
	st := NewSymbolTable()
	lifecycle.GenerateSymbolTableRegistrations(st)

	for _, name := range []string{"time_close", "time_tradingday"} {
		if !st.IsSeries(name) {
			t.Errorf("%s should be registered as series in symbol table", name)
		}
	}
}

func TestTimeSeriesLifecycle_SymbolTableRegistration_NilReceiver(t *testing.T) {
	var lifecycle *TimeSeriesLifecycle
	st := NewSymbolTable()

	lifecycle.GenerateSymbolTableRegistrations(st)

	for _, name := range []string{"time_close", "time_tradingday"} {
		if st.IsSeries(name) {
			t.Errorf("nil lifecycle should not register %s", name)
		}
	}
}

func TestTimeSeriesLifecycle_SymbolTableRegistration_NilTable(t *testing.T) {
	lifecycle := NewTimeSeriesLifecycle(true, true)
	lifecycle.GenerateSymbolTableRegistrations(nil)
}
